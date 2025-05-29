package controllers

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gophish/gophish/config"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/controllers/api"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jordan-wright/unindexed"
)

// ErrInvalidRequest is thrown when a request with an invalid structure is
// received
var ErrInvalidRequest = errors.New("Invalid request")

// ErrCampaignComplete is thrown when an event is received for a campaign that
// has already been marked as complete.
var ErrCampaignComplete = errors.New("Event received on completed campaign")

// TransparencyResponse is the JSON response provided when a third-party
// makes a request to the transparency handler.
type TransparencyResponse struct {
	Server         string    `json:"server"`
	ContactAddress string    `json:"contact_address"`
	SendDate       time.Time `json:"send_date"`
}

// TransparencySuffix (when appended to a valid result ID), will cause Gophish
// to return a transparency response.
const TransparencySuffix = "+"

// PhishingServerOption is a functional option that is used to configure the
// the phishing server
type PhishingServerOption func(*PhishingServer)

// PhishingServer is an HTTP server that implements the campaign event
// handlers, such as email open tracking, click tracking, and more.
type PhishingServer struct {
	server         *http.Server
	config         config.PhishServer
	contactAddress string
}

// NewPhishingServer returns a new instance of the phishing server with
// provided options applied.
func NewPhishingServer(config config.PhishServer, options ...PhishingServerOption) *PhishingServer {
	defaultServer := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         config.ListenURL,
	}
	ps := &PhishingServer{
		server: defaultServer,
		config: config,
	}
	for _, opt := range options {
		opt(ps)
	}
	ps.registerRoutes()
	return ps
}

// WithContactAddress sets the contact address used by the transparency
// handlers
func WithContactAddress(addr string) PhishingServerOption {
	return func(ps *PhishingServer) {
		ps.contactAddress = addr
	}
}

// Start launches the phishing server, listening on the configured address.
func (ps *PhishingServer) Start() {
	if ps.config.UseTLS {
		// Only support TLS 1.2 and above - ref #1691, #1689
		ps.server.TLSConfig = defaultTLSConfig
		err := util.CheckAndCreateSSL(ps.config.CertPath, ps.config.KeyPath)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Starting phishing server at https://%s", ps.config.ListenURL)
		log.Fatal(ps.server.ListenAndServeTLS(ps.config.CertPath, ps.config.KeyPath))
	}
	// If TLS isn't configured, just listen on HTTP
	log.Infof("Starting phishing server at http://%s", ps.config.ListenURL)
	log.Fatal(ps.server.ListenAndServe())
}

// Shutdown attempts to gracefully shutdown the server.
func (ps *PhishingServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return ps.server.Shutdown(ctx)
}

// CreatePhishingRouter creates the router that handles phishing connections.
func (ps *PhishingServer) registerRoutes() {
	router := mux.NewRouter()
	fileServer := http.FileServer(unindexed.Dir("./static/endpoint/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
	router.HandleFunc("/track", ps.TrackHandler)
	router.HandleFunc("/robots.txt", ps.RobotsHandler)
	router.HandleFunc("/{path:.*}/track", ps.TrackHandler)
	router.HandleFunc("/{path:.*}/report", ps.ReportHandler)
	router.HandleFunc("/report", ps.ReportHandler)
	router.HandleFunc("/{path:.*}", ps.PhishHandler)

	// Setup GZIP compression
	gzipWrapper, _ := gziphandler.NewGzipLevelHandler(gzip.BestCompression)
	phishHandler := gzipWrapper(router)

	// Respect X-Forwarded-For and X-Real-IP headers in case we're behind a
	// reverse proxy.
	phishHandler = handlers.ProxyHeaders(phishHandler)

	// Setup logging
	phishHandler = handlers.CombinedLoggingHandler(log.Writer(), phishHandler)
	ps.server.Handler = phishHandler
}

// TrackHandler tracks emails as they are opened, updating the status for the given Result
func (ps *PhishingServer) TrackHandler(w http.ResponseWriter, r *http.Request) {
	r, err := setupContext(r)
	if err != nil {
		// Log the error if it wasn't something we can safely ignore
		if err != ErrInvalidRequest && err != ErrCampaignComplete {
			log.Error(err)
		}
		http.NotFound(w, r)
		return
	}
	// Check for a preview
	if _, ok := ctx.Get(r, "result").(models.EmailRequest); ok {
		http.ServeFile(w, r, "static/images/pixel.png")
		return
	}
	rs := ctx.Get(r, "result").(models.Result)
	rid := ctx.Get(r, "rid").(string)
	d := ctx.Get(r, "details").(models.EventDetails)

	// Check for a transparency request
	if strings.HasSuffix(rid, TransparencySuffix) {
		ps.TransparencyHandler(w, r)
		return
	}

	err = rs.HandleEmailOpened(d)
	if err != nil {
		log.Error(err)
	}
	http.ServeFile(w, r, "static/images/pixel.png")
}

// ReportHandler tracks emails as they are reported, updating the status for the given Result
func (ps *PhishingServer) ReportHandler(w http.ResponseWriter, r *http.Request) {
	r, err := setupContext(r)
	w.Header().Set("Access-Control-Allow-Origin", "*") // To allow Chrome extensions (or other pages) to report a campaign without violating CORS
	if err != nil {
		// Log the error if it wasn't something we can safely ignore
		if err != ErrInvalidRequest && err != ErrCampaignComplete {
			log.Error(err)
		}
		http.NotFound(w, r)
		return
	}
	// Check for a preview
	if _, ok := ctx.Get(r, "result").(models.EmailRequest); ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	rs := ctx.Get(r, "result").(models.Result)
	rid := ctx.Get(r, "rid").(string)
	d := ctx.Get(r, "details").(models.EventDetails)

	// Check for a transparency request
	if strings.HasSuffix(rid, TransparencySuffix) {
		ps.TransparencyHandler(w, r)
		return
	}

	err = rs.HandleEmailReport(d)
	if err != nil {
		log.Error(err)
	}
	w.WriteHeader(http.StatusNoContent)
}

// PhishHandler handles incoming client connections and registers the associated actions performed
// (such as clicked link, etc.)
func (ps *PhishingServer) PhishHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received request to phishing handler")
	r, err := setupContext(r)
	if err != nil {
		// Log the error if it wasn't something we can safely ignore
		if err != ErrInvalidRequest && err != ErrCampaignComplete {
			log.Error("Error setting up phishing context:", err)
		}
		http.NotFound(w, r)
		return
	}
	
	w.Header().Set("X-Server", config.ServerName) // Useful for checking if this is a GoPhish server (e.g. for campaign reporting plugins)
	var ptx models.PhishingTemplateContext
	
	// Check for a preview
	if preview, ok := ctx.Get(r, "result").(models.EmailRequest); ok {
		log.Debug("Processing preview request")
		ptx, err = models.NewPhishingTemplateContext(&preview, preview.BaseRecipient, preview.RId)
		if err != nil {
			log.Error("Error creating template context for preview:", err)
			http.NotFound(w, r)
			return
		}
		p, err := models.GetPage(preview.PageId, preview.UserId)
		if err != nil {
			log.Error("Error getting page for preview:", err)
			http.NotFound(w, r)
			return
		}
		renderPhishResponse(w, r, ptx, p)
		return
	}
	
	rs := ctx.Get(r, "result").(models.Result)
	rid := ctx.Get(r, "rid").(string)
	c := ctx.Get(r, "campaign").(models.Campaign)
	d := ctx.Get(r, "details").(models.EventDetails)
	
	log.Debug("Processing request for campaign:", c.Name, "RId:", rid)
	log.Debug("Request method:", r.Method)
	if r.Method == "POST" {
		log.Debug("Form data received:", r.Form)
	}

	// Check for a transparency request
	if strings.HasSuffix(rid, TransparencySuffix) {
		log.Debug("Processing transparency request")
		ps.TransparencyHandler(w, r)
		return
	}

	p, err := models.GetPage(c.PageId, c.UserId)
	if err != nil {
		log.Error("Error getting page for campaign:", err)
		http.NotFound(w, r)
		return
	}
	
	switch {
	case r.Method == "GET":
		log.Debug("Processing GET request - clicked link")
		err = rs.HandleClickedLink(d)
		if err != nil {
			log.Error("Error handling clicked link:", err)
		}
	case r.Method == "POST":
		log.Debug("Processing POST request - form submission")
		log.Debug("Form data:", d.Payload)
		err = rs.HandleFormSubmit(d)
		if err != nil {
			log.Error("Error handling form submit:", err)
		}
	}
	
	ptx, err = models.NewPhishingTemplateContext(&c, rs.BaseRecipient, rs.RId)
	if err != nil {
		log.Error("Error creating template context:", err)
		http.NotFound(w, r)
		return
	}
	renderPhishResponse(w, r, ptx, p)
}

// renderPhishResponse handles rendering the correct response to the phishing
// connection. This usually involves writing out the page HTML or redirecting
// the user to the correct URL.
func renderPhishResponse(w http.ResponseWriter, r *http.Request, ptx models.PhishingTemplateContext, p models.Page) {
	// If the request was a form submit and a redirect URL was specified, we
	// should send the user to that URL
	if r.Method == "POST" {
		if p.RedirectURL != "" {
			redirectURL, err := models.ExecuteTemplate(p.RedirectURL, ptx)
			if err != nil {
				log.Error(err)
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
	}
	// Otherwise, we just need to write out the templated HTML
	html, err := models.ExecuteTemplate(p.HTML, ptx)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}
	w.Write([]byte(html))
}

// RobotsHandler prevents search engines, etc. from indexing phishing materials
func (ps *PhishingServer) RobotsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User-agent: *\nDisallow: /")
}

// TransparencyHandler returns a TransparencyResponse for the provided result
// and campaign.
func (ps *PhishingServer) TransparencyHandler(w http.ResponseWriter, r *http.Request) {
	rs := ctx.Get(r, "result").(models.Result)
	tr := &TransparencyResponse{
		Server:         config.ServerName,
		SendDate:       rs.SendDate,
		ContactAddress: ps.contactAddress,
	}
	api.JSONResponse(w, tr, http.StatusOK)
}

// setupContext handles some of the administrative work around receiving a new
// request, such as checking the result ID, the campaign, etc.
func setupContext(r *http.Request) (*http.Request, error) {
	log.Debug("Setting up request context")
	err := r.ParseForm()
	if err != nil {
		log.Error("Error parsing form:", err)
		return r, err
	}
	log.Debug("Form data parsed:", r.Form)
	
	rid := r.Form.Get(models.RecipientParameter)
	if rid == "" {
		log.Debug("No recipient parameter found in request")
		return r, ErrInvalidRequest
	}
	log.Debug("Found recipient ID:", rid)
	
	// Since we want to support the common case of adding a "+" to indicate a
	// transparency request, we need to take care to handle the case where the
	// request ends with a space, since a "+" is technically reserved for use
	// as a URL encoding of a space.
	if strings.HasSuffix(rid, " ") {
		// We'll trim off the space
		rid = strings.TrimRight(rid, " ")
		// Then we'll add the transparency suffix
		rid = fmt.Sprintf("%s%s", rid, TransparencySuffix)
		log.Debug("Modified recipient ID for transparency:", rid)
	}
	
	// Finally, if this is a transparency request, we'll need to verify that
	// a valid rid has been provided, so we'll look up the result with a
	// trimmed parameter.
	id := strings.TrimSuffix(rid, TransparencySuffix)
	
	// Check to see if this is a preview or a real result
	if strings.HasPrefix(id, models.PreviewPrefix) {
		log.Debug("Processing preview request for ID:", id)
		rs, err := models.GetEmailRequestByResultId(id)
		if err != nil {
			log.Error("Error getting email request:", err)
			return r, err
		}
		r = ctx.Set(r, "result", rs)
		return r, nil
	}
	
	rs, err := models.GetResult(id)
	if err != nil {
		log.Error("Error getting result:", err)
		return r, err
	}
	log.Debug("Found result for ID:", id)
	
	c, err := models.GetCampaign(rs.CampaignId, rs.UserId)
	if err != nil {
		log.Error("Error getting campaign:", err)
		return r, err
	}
	log.Debug("Found campaign:", c.Name)
	
	// Don't process events for completed campaigns
	if c.Status == models.CampaignComplete {
		log.Debug("Campaign is complete, not processing event")
		return r, ErrCampaignComplete
	}
	
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	log.Debug("Client IP:", ip)
	
	// Handle post processing such as GeoIP
	err = rs.UpdateGeo(ip)
	if err != nil {
		log.Error("Error updating geo location:", err)
	}
	
	d := models.EventDetails{
		Payload: r.Form,
		Browser: make(map[string]string),
	}
	d.Browser["address"] = ip
	d.Browser["user-agent"] = r.Header.Get("User-Agent")
	
	log.Debug("Event details:", d)

	r = ctx.Set(r, "rid", rid)
	r = ctx.Set(r, "result", rs)
	r = ctx.Set(r, "campaign", c)
	r = ctx.Set(r, "details", d)
	return r, nil
}
