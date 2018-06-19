package controllers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gophish/gophish/auth"
	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util"
	"github.com/gophish/gophish/worker"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
)

// Worker is the worker that processes phishing events and updates campaigns.
var Worker *worker.Worker

func init() {
	Worker = worker.New()
	go Worker.Start()
}

// API (/api) provides access to api documentation
func API(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/docs.html")
		if err != nil {
			log.Error(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", nil)
	}
}

// API (/api/reset) resets a user's API key
func API_Reset(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "POST":
		u := ctx.Get(r, "user").(models.User)
		u.ApiKey = auth.GenerateSecureKey()
		err := models.PutUser(&u)
		if err != nil {
			http.Error(w, "Error setting API Key", http.StatusInternalServerError)
		} else {
			JSONResponse(w, models.Response{Success: true, Message: "API Key successfully reset!", Data: u.ApiKey}, http.StatusOK)
		}
	}
}

// API_Campaigns returns a list of campaigns if requested via GET.
// If requested via POST, API_Campaigns creates a new campaign and returns a reference to it.
func API_Campaigns(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, cs, http.StatusOK)
	//POST: Create a new campaign and return it as JSON
	case r.Method == "POST":
		c := models.Campaign{}
		// Put the request into a campaign
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		err = models.PostCampaign(&c, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		// If the campaign is scheduled to launch immediately, send it to the worker.
		// Otherwise, the worker will pick it up at the scheduled time
		if c.Status == models.CAMPAIGN_IN_PROGRESS {
			go Worker.LaunchCampaign(c)
		}
		JSONResponse(w, c, http.StatusCreated)
	}
}

// API_Campaigns_Summary returns the summary for the current user's campaigns
func API_Campaigns_Summary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaignSummaries(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, cs, http.StatusOK)
	}
}

// API_Campaigns_Id returns details about the requested campaign. If the campaign is not
// valid, API_Campaigns_Id returns null.
func API_Campaigns_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	c, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, c, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteCampaign(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting campaign"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign deleted successfully!"}, http.StatusOK)
	}
}

// API_Campaigns_Id_Results returns just the results for a given campaign to
// significantly reduce the information returned.
func API_Campaigns_Id_Results(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	cr, err := models.GetCampaignResults(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}
	if r.Method == "GET" {
		JSONResponse(w, cr, http.StatusOK)
		return
	}
}

// API_Campaigns_Id_Summary returns just the summary for a given campaign.
func API_Campaign_Id_Summary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaignSummary(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
			} else {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			}
			log.Error(err)
			return
		}
		JSONResponse(w, cs, http.StatusOK)
	}
}

// API_Campaigns_Id_Complete effectively "ends" a campaign.
// Future phishing emails clicked will return a simple "404" page.
func API_Campaigns_Id_Complete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		err := models.CompleteCampaign(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error completing campaign"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign completed successfully!"}, http.StatusOK)
	}
}

// API_Groups returns a list of groups if requested via GET.
// If requested via POST, API_Groups creates a new group and returns a reference to it.
func API_Groups(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		gs, err := models.GetGroups(ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "No groups found"}, http.StatusNotFound)
			return
		}
		JSONResponse(w, gs, http.StatusOK)
	//POST: Create a new group and return it as JSON
	case r.Method == "POST":
		g := models.Group{}
		// Put the request into a group
		err := json.NewDecoder(r.Body).Decode(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		_, err = models.GetGroupByName(g.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Group name already in use"}, http.StatusConflict)
			return
		}
		g.ModifiedDate = time.Now().UTC()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, g, http.StatusCreated)
	}
}

// API_Groups_Summary returns a summary of the groups owned by the current user.
func API_Groups_Summary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		gs, err := models.GetGroupSummaries(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, gs, http.StatusOK)
	}
}

// API_Groups_Id returns details about the requested group.
// If the group is not valid, API_Groups_Id returns null.
func API_Groups_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	g, err := models.GetGroup(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, g, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting group"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Group deleted successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		// Change this to get from URL and uid (don't bother with id in r.Body)
		g = models.Group{}
		err = json.NewDecoder(r.Body).Decode(&g)
		if g.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "Error: /:id and group_id mismatch"}, http.StatusInternalServerError)
			return
		}
		g.ModifiedDate = time.Now().UTC()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, g, http.StatusOK)
	}
}

// API_Groups_Id_Summary returns a summary of the groups owned by the current user.
func API_Groups_Id_Summary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 0, 64)
		g, err := models.GetGroupSummary(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
			return
		}
		JSONResponse(w, g, http.StatusOK)
	}
}

// API_Templates handles the functionality for the /api/templates endpoint
func API_Templates(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ts, err := models.GetTemplates(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, ts, http.StatusOK)
	//POST: Create a new template and return it as JSON
	case r.Method == "POST":
		t := models.Template{}
		// Put the request into a template
		err := json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		_, err = models.GetTemplateByName(t.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Template name already in use"}, http.StatusConflict)
			return
		}
		t.ModifiedDate = time.Now().UTC()
		t.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostTemplate(&t)
		if err == models.ErrTemplateNameNotSpecified {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		if err == models.ErrTemplateMissingParameter {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error inserting template into database"}, http.StatusInternalServerError)
			log.Error(err)
			return
		}
		JSONResponse(w, t, http.StatusCreated)
	}
}

// API_Templates_Id handles the functions for the /api/templates/:id endpoint
func API_Templates_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	t, err := models.GetTemplate(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Template not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, t, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteTemplate(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting template"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Template deleted successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		t = models.Template{}
		err = json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			log.Error(err)
		}
		if t.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "Error: /:id and template_id mismatch"}, http.StatusBadRequest)
			return
		}
		t.ModifiedDate = time.Now().UTC()
		t.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutTemplate(&t)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, t, http.StatusOK)
	}
}

// API_Pages handles requests for the /api/pages/ endpoint
func API_Pages(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ps, err := models.GetPages(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, ps, http.StatusOK)
	//POST: Create a new page and return it as JSON
	case r.Method == "POST":
		p := models.Page{}
		// Put the request into a page
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		// Check to make sure the name is unique
		_, err = models.GetPageByName(p.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Page name already in use"}, http.StatusConflict)
			log.Error(err)
			return
		}
		p.ModifiedDate = time.Now().UTC()
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostPage(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusCreated)
	}
}

// API_Pages_Id contains functions to handle the GET'ing, DELETE'ing, and PUT'ing
// of a Page object
func API_Pages_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	p, err := models.GetPage(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Page not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, p, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeletePage(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting page"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Page Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		p = models.Page{}
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			log.Error(err)
		}
		if p.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "/:id and /:page_id mismatch"}, http.StatusBadRequest)
			return
		}
		p.ModifiedDate = time.Now().UTC()
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutPage(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error updating page: " + err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusOK)
	}
}

// API_SMTP handles requests for the /api/smtp/ endpoint
func API_SMTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ss, err := models.GetSMTPs(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, ss, http.StatusOK)
	//POST: Create a new SMTP and return it as JSON
	case r.Method == "POST":
		s := models.SMTP{}
		// Put the request into a page
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		// Check to make sure the name is unique
		_, err = models.GetSMTPByName(s.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "SMTP name already in use"}, http.StatusConflict)
			log.Error(err)
			return
		}
		s.ModifiedDate = time.Now().UTC()
		s.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostSMTP(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, s, http.StatusCreated)
	}
}

// API_SMTP_Id contains functions to handle the GET'ing, DELETE'ing, and PUT'ing
// of a SMTP object
func API_SMTP_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	s, err := models.GetSMTP(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "SMTP not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, s, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteSMTP(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting SMTP"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "SMTP Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		s = models.SMTP{}
		err = json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			log.Error(err)
		}
		if s.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "/:id and /:smtp_id mismatch"}, http.StatusBadRequest)
			return
		}
		err = s.Validate()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		s.ModifiedDate = time.Now().UTC()
		s.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutSMTP(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error updating page"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, s, http.StatusOK)
	}
}

// API_Import_Group imports a CSV of group members
func API_Import_Group(w http.ResponseWriter, r *http.Request) {
	ts, err := util.ParseCSV(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Error parsing CSV"}, http.StatusInternalServerError)
		return
	}
	JSONResponse(w, ts, http.StatusOK)
	return
}

// API_Import_Email allows for the importing of email.
// Returns a Message object
func API_Import_Email(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusBadRequest)
		return
	}
	ir := struct {
		Content      string `json:"content"`
		ConvertLinks bool   `json:"convert_links"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&ir)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Error decoding JSON Request"}, http.StatusBadRequest)
		return
	}
	e, err := email.NewEmailFromReader(strings.NewReader(ir.Content))
	if err != nil {
		log.Error(err)
	}
	// If the user wants to convert links to point to
	// the landing page, let's make it happen by changing up
	// e.HTML
	if ir.ConvertLinks {
		d, err := goquery.NewDocumentFromReader(bytes.NewReader(e.HTML))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		d.Find("a").Each(func(i int, a *goquery.Selection) {
			a.SetAttr("href", "{{.URL}}")
		})
		h, err := d.Html()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		e.HTML = []byte(h)
	}
	er := emailResponse{
		Subject: e.Subject,
		Text:    string(e.Text),
		HTML:    string(e.HTML),
	}
	JSONResponse(w, er, http.StatusOK)
	return
}

// API_Import_Site allows for the importing of HTML from a website
// Without "include_resources" set, it will merely place a "base" tag
// so that all resources can be loaded relative to the given URL.
func API_Import_Site(w http.ResponseWriter, r *http.Request) {
	cr := cloneRequest{}
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Error decoding JSON Request"}, http.StatusBadRequest)
		return
	}
	if err = cr.validate(); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(cr.URL)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	// Insert the base href tag to better handle relative resources
	d, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	// Assuming we don't want to include resources, we'll need a base href
	if d.Find("head base").Length() == 0 {
		d.Find("head").PrependHtml(fmt.Sprintf("<base href=\"%s\">", cr.URL))
	}
	forms := d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// We'll want to store where we got the form from
		// (the current URL)
		url := f.AttrOr("action", cr.URL)
		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("%s%s", cr.URL, url)
		}
		f.PrependHtml(fmt.Sprintf("<input type=\"hidden\" name=\"__original_url\" value=\"%s\"/>", url))
	})
	h, err := d.Html()
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	cs := cloneResponse{HTML: h}
	JSONResponse(w, cs, http.StatusOK)
	return
}

// API_Send_Test_Email sends a test email using the template name
// and Target given.
func API_Send_Test_Email(w http.ResponseWriter, r *http.Request) {
	s := &models.EmailRequest{
		ErrorChan: make(chan error),
		UserId:    ctx.Get(r, "user_id").(int64),
	}
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(s)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Error decoding JSON Request"}, http.StatusBadRequest)
		return
	}

	storeRequest := false

	// If a Template is not specified use a default
	if s.Template.Name == "" {
		//default message body
		text := "It works!\n\nThis is an email letting you know that your gophish\nconfiguration was successful.\n" +
			"Here are the details:\n\nWho you sent from: {{.From}}\n\nWho you sent to: \n" +
			"{{if .FirstName}} First Name: {{.FirstName}}\n{{end}}" +
			"{{if .LastName}} Last Name: {{.LastName}}\n{{end}}" +
			"{{if .Position}} Position: {{.Position}}\n{{end}}" +
			"\nNow go send some phish!"
		t := models.Template{
			Subject: "Default Email from Gophish",
			Text:    text,
		}
		s.Template = t
	} else {
		// Get the Template requested by name
		s.Template, err = models.GetTemplateByName(s.Template.Name, s.UserId)
		if err == gorm.ErrRecordNotFound {
			log.WithFields(logrus.Fields{
				"template": s.Template.Name,
			}).Error("Template does not exist")
			JSONResponse(w, models.Response{Success: false, Message: models.ErrTemplateNotFound.Error()}, http.StatusBadRequest)
			return
		} else if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		s.TemplateId = s.Template.Id
		// We'll only save the test request to the database if there is a
		// user-specified template to use.
		storeRequest = true
	}

	if s.Page.Name != "" {
		s.Page, err = models.GetPageByName(s.Page.Name, s.UserId)
		if err == gorm.ErrRecordNotFound {
			log.WithFields(logrus.Fields{
				"page": s.Page.Name,
			}).Error("Page does not exist")
			JSONResponse(w, models.Response{Success: false, Message: models.ErrPageNotFound.Error()}, http.StatusBadRequest)
			return
		} else if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		s.PageId = s.Page.Id
	}

	// If a complete sending profile is provided use it
	if err := s.SMTP.Validate(); err != nil {
		// Otherwise get the SMTP requested by name
		smtp, lookupErr := models.GetSMTPByName(s.SMTP.Name, s.UserId)
		// If the Sending Profile doesn't exist, let's err on the side
		// of caution and assume that the validation failure was more important.
		if lookupErr != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		s.SMTP = smtp
	}
	s.FromAddress = s.SMTP.FromAddress

	// Validate the given request
	if err = s.Validate(); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	// Store the request if this wasn't the default template
	if storeRequest {
		err = models.PostEmailRequest(s)
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
	}
	// Send the test email
	err = Worker.SendTestEmail(s)
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	JSONResponse(w, models.Response{Success: true, Message: "Email Sent"}, http.StatusOK)
	return
}

// JSONResponse attempts to set the status code, c, and marshal the given interface, d, into a response that
// is written to the given ResponseWriter.
func JSONResponse(w http.ResponseWriter, d interface{}, c int) {
	dj, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", dj)
}

type cloneRequest struct {
	URL              string `json:"url"`
	IncludeResources bool   `json:"include_resources"`
}

func (cr *cloneRequest) validate() error {
	if cr.URL == "" {
		return errors.New("No URL Specified")
	}
	return nil
}

type cloneResponse struct {
	HTML string `json:"html"`
}

type emailResponse struct {
	Text    string `json:"text"`
	HTML    string `json:"html"`
	Subject string `json:"subject"`
}
