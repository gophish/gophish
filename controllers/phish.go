package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"strings"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

// ErrInvalidRequest is thrown when a request with an invalid structure is
// received
var ErrInvalidRequest = errors.New("Invalid request")

// ErrCampaignComplete is thrown when an event is received for a campaign that
// has already been marked as complete.
var ErrCampaignComplete = errors.New("Event received on completed campaign")

// eventDetails is a struct that wraps common attributes we want to store
// in an event
type eventDetails struct {
	Payload url.Values        `json:"payload"`
	Browser map[string]string `json:"browser"`
}

// CreatePhishingRouter creates the router that handles phishing connections.
func CreatePhishingRouter() http.Handler {
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/endpoint/"))))
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
			Logger.Println(err)
		}
		http.NotFound(w, r)
		return
	}
	rs := ctx.Get(r, "result").(models.Result)
	c := ctx.Get(r, "campaign").(models.Campaign)
	rj := ctx.Get(r, "details").([]byte)
	c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_OPENED, Details: string(rj)})
	// Don't update the status if the user already clicked the link
	// or submitted data to the campaign
	if rs.Status == models.EVENT_CLICKED || rs.Status == models.EVENT_DATA_SUBMIT {
		http.ServeFile(w, r, "static/images/pixel.png")
		return
	}
	err = rs.UpdateStatus(models.EVENT_OPENED)
	if err != nil {
		Logger.Println(err)
	}
	http.ServeFile(w, r, "static/images/pixel.png")
}

// PhishReporter tracks emails as they are reported, updating the status for the given Result
func PhishReporter(w http.ResponseWriter, r *http.Request) {
	err, r := setupContext(r)
	if err != nil {
		// Log the error if it wasn't something we can safely ignore
		if err != ErrInvalidRequest && err != ErrCampaignComplete {
			Logger.Println(err)
		}
		http.NotFound(w, r)
		return
	}
	rs := ctx.Get(r, "result").(models.Result)
	c := ctx.Get(r, "campaign").(models.Campaign)
	rj := ctx.Get(r, "details").([]byte)
	c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_REPORTED, Details: string(rj)})

	err = rs.UpdateReported(true)
	if err != nil {
		Logger.Println(err)
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
			Logger.Println(err)
		}
		http.NotFound(w, r)
		return
	}
	rs := ctx.Get(r, "result").(models.Result)
	c := ctx.Get(r, "campaign").(models.Campaign)
	rj := ctx.Get(r, "details").([]byte)
	p, err := models.GetPage(c.PageId, c.UserId)
	if err != nil {
		Logger.Println(err)
		http.NotFound(w, r)
		return
	}
	switch {
	case r.Method == "GET":
		if rs.Status != models.EVENT_CLICKED && rs.Status != models.EVENT_DATA_SUBMIT {
			rs.UpdateStatus(models.EVENT_CLICKED)
		}
		err = c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_CLICKED, Details: string(rj)})
		if err != nil {
			Logger.Println(err)
		}
	case r.Method == "POST":
		// If data was POST'ed, let's record it
		rs.UpdateStatus(models.EVENT_DATA_SUBMIT)
		// Store the data in an event
		c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_DATA_SUBMIT, Details: string(rj)})
		if err != nil {
			Logger.Println(err)
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
		Logger.Println(err)
		http.NotFound(w, r)
		return
	}
	f, err := mail.ParseAddress(c.SMTP.FromAddress)
	if err != nil {
		Logger.Println(err)
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
		Logger.Println(err)
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
		Logger.Println(err)
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
		Logger.Println(err)
		return err, r
	}
	// Don't process events for completed campaigns
	if c.Status == models.CAMPAIGN_COMPLETE {
		return ErrCampaignComplete, r
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		Logger.Println(err)
		return err, r
	}
	// Respect X-Forwarded headers
	if fips := r.Header.Get("X-Forwarded-For"); fips != "" {
		ip = strings.Split(fips, ", ")[0]
	}
	// Handle post processing such as GeoIP
	err = rs.UpdateGeo(ip)
	if err != nil {
		Logger.Println(err)
	}
	d := eventDetails{
		Payload: r.Form,
		Browser: make(map[string]string),
	}
	d.Browser["address"] = ip
	d.Browser["user-agent"] = r.Header.Get("User-Agent")
	rj, err := json.Marshal(d)

	r = ctx.Set(r, "result", rs)
	r = ctx.Set(r, "campaign", c)
	r = ctx.Set(r, "details", rj)
	return nil, r
}
