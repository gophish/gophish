package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"strings"

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
	rs := ctx.Get(r, "result").(models.Result)
	d := ctx.Get(r, "details").(models.EventDetails)
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
	rs := ctx.Get(r, "result").(models.Result)
	d := ctx.Get(r, "details").(models.EventDetails)

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
	rs := ctx.Get(r, "result").(models.Result)
	c := ctx.Get(r, "campaign").(models.Campaign)
	d := ctx.Get(r, "details").(models.EventDetails)
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
		// Redirect to the desired page
		if p.RedirectURL != "" {
			http.Redirect(w, r, p.RedirectURL, 302)
			return
		}
	}
	var htmlBuff bytes.Buffer
	tmpl, err := template.New("html_template").Parse(p.HTML)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}
	f, err := mail.ParseAddress(c.SMTP.FromAddress)
	if err != nil {
		log.Error(err)
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}

	phishURL, _ := url.Parse(c.URL)
	q := phishURL.Query()
	q.Set(models.RecipientParameter, rs.RId)
	phishURL.RawQuery = q.Encode()

	rsf := struct {
		models.Result
		URL  string
		From string
	}{
		rs,
		phishURL.String(),
		fn,
	}
	err = tmpl.Execute(&htmlBuff, rsf)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}
	w.Write(htmlBuff.Bytes())
}

// RobotsHandler prevents search engines, etc. from indexing phishing materials
func RobotsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User-agent: *\nDisallow: /")
}

// setupContext handles some of the administrative work around receiving a new request, such as checking the result ID, the campaign, etc.
func setupContext(r *http.Request) (error, *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		return err, r
	}
	id := r.Form.Get(models.RecipientParameter)
	if id == "" {
		return ErrInvalidRequest, r
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

	r = ctx.Set(r, "result", rs)
	r = ctx.Set(r, "campaign", c)
	r = ctx.Set(r, "details", d)
	return nil, r
}
