package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"text/template"
	"time"

	ctx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/jordan-wright/gophish/auth"
	"github.com/jordan-wright/gophish/models"
	"github.com/jordan-wright/gophish/util"
	"github.com/jordan-wright/gophish/worker"
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
		_, err := templates.ParseFiles("templates/api-docs.html")
		if err != nil {
			fmt.Println(err)
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
			JSONResponse(w, models.Response{Success: true, Message: "API Key Successfully Reset", Data: u.ApiKey}, http.StatusOK)
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
			fmt.Println(err)
		}
		JSONResponse(w, cs, http.StatusOK)
	//POST: Create a new campaign and return it as JSON
	case r.Method == "POST":
		c := models.Campaign{}
		// Put the request into a campaign
		err := json.NewDecoder(r.Body).Decode(&c)
		if checkError(err, w, "Invalid Request", http.StatusBadRequest) {
			return
		}
		if m, ok := c.Validate(); !ok {
			http.Error(w, "Error: "+m, http.StatusBadRequest)
			return
		}
		err = models.PostCampaign(&c, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "Cannot insert campaign into database", http.StatusInternalServerError) {
			return
		}
		Worker.Queue <- &c
		JSONResponse(w, c, http.StatusCreated)
	}
}

// API_Campaigns_Id returns details about the requested campaign. If the campaign is not
// valid, API_Campaigns_Id returns null.
func API_Campaigns_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	c, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
	if checkError(err, w, "Campaign not found", http.StatusNotFound) {
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, c, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteCampaign(id)
		if checkError(err, w, "Error deleting campaign", http.StatusInternalServerError) {
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign Deleted Successfully!"}, http.StatusOK)
	}
}

// API_Groups returns details about the requested group. If the campaign is not
// valid, API_Groups returns null.
func API_Groups(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		gs, err := models.GetGroups(ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "Groups not found", http.StatusNotFound) {
			return
		}
		JSONResponse(w, gs, http.StatusOK)
	//POST: Create a new group and return it as JSON
	case r.Method == "POST":
		g := models.Group{}
		// Put the request into a group
		err := json.NewDecoder(r.Body).Decode(&g)
		if checkError(err, w, "Invalid Request", http.StatusBadRequest) {
			return
		}
		_, err = models.GetGroupByName(g.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.RecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Group name already in use"}, http.StatusConflict)
			return
		}
		// Check to make sure targets were specified
		if len(g.Targets) == 0 {
			http.Error(w, "Error: No targets specified", http.StatusBadRequest)
			return
		}
		g.ModifiedDate = time.Now()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostGroup(&g)
		if checkError(err, w, "Error inserting group", http.StatusInternalServerError) {
			return
		}
		w.Header().Set("Location", "http://localhost:3333/api/groups/"+string(g.Id))
		JSONResponse(w, g, http.StatusCreated)
	}
}

// API_Groups_Id returns details about the requested campaign. If the campaign is not
// valid, API_Campaigns_Id returns null.
func API_Groups_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	g, err := models.GetGroup(id, ctx.Get(r, "user_id").(int64))
	if checkError(err, w, "Group not found", http.StatusNotFound) {
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, g, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteGroup(&g)
		if checkError(err, w, "Error deleting group", http.StatusInternalServerError) {
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Group Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		// Change this to get from URL and uid (don't bother with id in r.Body)
		g = models.Group{}
		err = json.NewDecoder(r.Body).Decode(&g)
		if g.Id != id {
			http.Error(w, "Error: /:id and group_id mismatch", http.StatusBadRequest)
			return
		}
		// Check to make sure targets were specified
		if len(g.Targets) == 0 {
			http.Error(w, "Error: No targets specified", http.StatusBadRequest)
			return
		}
		g.ModifiedDate = time.Now()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutGroup(&g)
		if checkError(err, w, "Error updating group", http.StatusInternalServerError) {
			return
		}
		JSONResponse(w, g, http.StatusOK)
	}
}

func API_Templates(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ts, err := models.GetTemplates(ctx.Get(r, "user_id").(int64))
		if err != nil {
			fmt.Println(err)
		}
		JSONResponse(w, ts, http.StatusOK)
	//POST: Create a new template and return it as JSON
	case r.Method == "POST":
		t := models.Template{}
		// Put the request into a template
		err := json.NewDecoder(r.Body).Decode(&t)
		if checkError(err, w, "Invalid Request", http.StatusBadRequest) {
			return
		}
		_, err = models.GetTemplateByName(t.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.RecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Template name already in use"}, http.StatusConflict)
			return
		}
		t.ModifiedDate = time.Now()
		t.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostTemplate(&t)
		if checkError(err, w, "Error inserting template", http.StatusInternalServerError) {
			return
		}
		JSONResponse(w, t, http.StatusCreated)
	}
}

func API_Templates_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	t, err := models.GetTemplate(id, ctx.Get(r, "user_id").(int64))
	if checkError(err, w, "Template not found", http.StatusNotFound) {
		Logger.Println(err)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, t, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteTemplate(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "Error deleting template", http.StatusInternalServerError) {
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Template Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		t = models.Template{}
		err = json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			Logger.Println(err)
		}
		if t.Id != id {
			http.Error(w, "Error: /:id and template_id mismatch", http.StatusBadRequest)
			return
		}
		err = t.Validate()
		/*		if checkError(err, w, http.StatusBadRequest) {
				return
			}*/
		t.ModifiedDate = time.Now()
		t.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutTemplate(&t)
		if checkError(err, w, "Error updating group", http.StatusInternalServerError) {
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
			fmt.Println(err)
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
		_, err = models.GetPageByName(p.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.RecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Page name already in use"}, http.StatusConflict)
			Logger.Println(err)
			return
		}
		p.ModifiedDate = time.Now()
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostPage(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error inserting page"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusCreated)
	}
}

func API_Pages_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	p, err := models.GetPage(id, ctx.Get(r, "user_id").(int64))
	if checkError(err, w, "Page not found", http.StatusNotFound) {
		Logger.Println(err)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, p, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeletePage(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "Error deleting page", http.StatusInternalServerError) {
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Page Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		p = models.Page{}
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			Logger.Println(err)
		}
		if p.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "/:id and /:page_id mismatch"}, http.StatusBadRequest)
			return
		}
		err = p.Validate()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid attributes given"}, http.StatusBadRequest)
			return
		}
		p.ModifiedDate = time.Now()
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutPage(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error updating page"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusOK)
	}
}

func API_Import_Group(w http.ResponseWriter, r *http.Request) {
	Logger.Println("Parsing CSV....")
	ts, err := util.ParseCSV(r)
	if checkError(err, w, "Error deleting template", http.StatusInternalServerError) {
		return
	}
	JSONResponse(w, ts, http.StatusOK)
}

// JSONResponse attempts to set the status code, c, and marshal the given interface, d, into a response that
// is written to the given ResponseWriter.
func JSONResponse(w http.ResponseWriter, d interface{}, c int) {
	dj, err := json.MarshalIndent(d, "", "  ")
	if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", dj)
}
