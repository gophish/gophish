package controllers

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gophish/gophish/config"
	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
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

// CreatePhishingRouter creates the router that handles phishing connections.
func CreatePhishingRouter() http.Handler {
	router := mux.NewRouter()
	fileServer := http.FileServer(UnindexedFileSystem{http.Dir("./static/endpoint/")})
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
	router.HandleFunc("/track", PhishTracker)
	router.HandleFunc("/robots.txt", RobotsHandler)
	router.HandleFunc("/{path:.*}/track", PhishTracker)
	router.HandleFunc("/{path:.*}/report", PhishReporter)
	router.HandleFunc("/report", PhishReporter)
	router.HandleFunc("/{path:.*}", PhishHandler)
	return router
}

// PhishTracker tracks emails as they are opened, updating the status for the given Result
func PhishTracker(w http.ResponseWriter, r *http.Request) {
	err, r := setupContext(r)
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
		TransparencyHandler(w, r)
		return
	}

	err = rs.HandleEmailOpened(d)
	if err != nil {
		log.Error(err)
	}
	http.ServeFile(w, r, "static/images/pixel.png")
}

// PhishReporter tracks emails as they are reported, updating the status for the given Result
func PhishReporter(w http.ResponseWriter, r *http.Request) {
	err, r := setupContext(r)
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
		TransparencyHandler(w, r)
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
func PhishHandler(w http.ResponseWriter, r *http.Request) {
	err, r := setupContext(r)
	if err != nil {
		// Log the error if it wasn't something we can safely ignore
		if err != ErrInvalidRequest && err != ErrCampaignComplete {
			log.Error(err)
		}
		http.NotFound(w, r)
		return
	}
	var ptx models.PhishingTemplateContext
	// Check for a preview
	if preview, ok := ctx.Get(r, "result").(models.EmailRequest); ok {
		ptx, err = models.NewPhishingTemplateContext(&preview, preview.BaseRecipient, preview.RId)
		if err != nil {
			log.Error(err)
			http.NotFound(w, r)
			return
		}
		p, err := models.GetPage(preview.PageId, preview.UserId)
		if err != nil {
			log.Error(err)
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

	// Check for a transparency request
	if strings.HasSuffix(rid, TransparencySuffix) {
		TransparencyHandler(w, r)
		return
	}

	p, err := models.GetPage(c.PageId, c.UserId)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}
	switch {
	case r.Method == "GET":
		err = rs.HandleClickedLink(d)
		if err != nil {
			log.Error(err)
		}
	case r.Method == "POST":
		err = rs.HandleFormSubmit(d)
		if err != nil {
			log.Error(err)
		}
	}
	ptx, err = models.NewPhishingTemplateContext(&c, rs.BaseRecipient, rs.RId)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
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
			http.Redirect(w, r, p.RedirectURL, 302)
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
func RobotsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User-agent: *\nDisallow: /")
}

// TransparencyHandler returns a TransparencyResponse for the provided result
// and campaign.
func TransparencyHandler(w http.ResponseWriter, r *http.Request) {
	rs := ctx.Get(r, "result").(models.Result)
	tr := &TransparencyResponse{
		Server:         config.ServerName,
		SendDate:       rs.SendDate,
		ContactAddress: config.Conf.ContactAddress,
	}
	JSONResponse(w, tr, http.StatusOK)
}

// setupContext handles some of the administrative work around receiving a new request, such as checking the result ID, the campaign, etc.
func setupContext(r *http.Request) (error, *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		return err, r
	}
	rid := r.Form.Get(models.RecipientParameter)
	if rid == "" {
		return ErrInvalidRequest, r
	}
	// Since we want to support the common case of adding a "+" to indicate a
	// transparency request, we need to take care to handle the case where the
	// request ends with a space, since a "+" is technically reserved for use
	// as a URL encoding of a space.
	if strings.HasSuffix(rid, " ") {
		// We'll trim off the space
		rid = strings.TrimRight(rid, " ")
		// Then we'll add the transparency suffix
		rid = fmt.Sprintf("%s%s", rid, TransparencySuffix)
	}
	// Finally, if this is a transparency request, we'll need to verify that
	// a valid rid has been provided, so we'll look up the result with a
	// trimmed parameter.
	id := strings.TrimSuffix(rid, TransparencySuffix)
	// Check to see if this is a preview or a real result
	if strings.HasPrefix(id, models.PreviewPrefix) {
		rs, err := models.GetEmailRequestByResultId(id)
		if err != nil {
			return err, r
		}
		r = ctx.Set(r, "result", rs)
		return nil, r
	}
	rs, err := models.GetResult(id)
	if err != nil {
		return err, r
	}
	c, err := models.GetCampaign(rs.CampaignId, rs.UserId)
	if err != nil {
		log.Error(err)
		return err, r
	}
	// Don't process events for completed campaigns
	if c.Status == models.CAMPAIGN_COMPLETE {
		return ErrCampaignComplete, r
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Error(err)
		return err, r
	}
	// Respect X-Forwarded headers
	if fips := r.Header.Get("X-Forwarded-For"); fips != "" {
		ip = strings.Split(fips, ", ")[0]
	}
	// Handle post processing such as GeoIP
	err = rs.UpdateGeo(ip)
	if err != nil {
		log.Error(err)
	}
	d := models.EventDetails{
		Payload: r.Form,
		Browser: make(map[string]string),
	}
	d.Browser["address"] = ip
	d.Browser["user-agent"] = r.Header.Get("User-Agent")

	r = ctx.Set(r, "rid", rid)
	r = ctx.Set(r, "result", rs)
	r = ctx.Set(r, "campaign", c)
	r = ctx.Set(r, "details", d)
	return nil, r
}
