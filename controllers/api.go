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
	"github.com/jordan-wright/gophish/auth"
	"github.com/jordan-wright/gophish/models"
	"github.com/jordan-wright/gophish/worker"
)

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
			Flash(w, r, "danger", "Error resetting API Key")
		} else {
			Flash(w, r, "success", "API Key Successfully Reset")
		}
		http.Redirect(w, r, "/settings", 302)
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
		cj, err := json.MarshalIndent(cs, "", "  ")
		if checkError(err, w, "Error looking up campaigns", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, cj)
	//POST: Create a new campaign and return it as JSON
	case r.Method == "POST":
		c := models.Campaign{}
		// Put the request into a campaign
		err := json.NewDecoder(r.Body).Decode(&c)
		if checkError(err, w, "Invalid Request", http.StatusBadRequest) {
			return
		}
		if m, ok := models.ValidateCampaign(&c); !ok {
			http.Error(w, "Error: "+m, http.StatusBadRequest)
			return
		}
		// Fill in the details
		c.CreatedDate = time.Now()
		c.CompletedDate = time.Time{}
		c.Status = models.QUEUED
		c.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostCampaign(&c, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "Cannot insert campaign into database", http.StatusInternalServerError) {
			return
		}
		Worker.Queue <- &c
		cj, err := json.MarshalIndent(c, "", "  ")
		if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, cj)
	}
}

// API_Campaigns_Id returns details about the requested campaign. If the campaign is not
// valid, API_Campaigns_Id returns null.
func API_Campaigns_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		c := models.Campaign{}
		c, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No campaign found", http.StatusNotFound) {
			return
		}
		cj, err := json.MarshalIndent(c, "", "  ")
		if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, cj)
	case r.Method == "DELETE":
		_, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No campaign found", http.StatusNotFound) {
			return
		}
		err = models.DeleteCampaign(id)
		if checkError(err, w, "Error deleting campaign", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, []byte("{\"success\" : \"true\"}"))
	}
}

func API_Campaigns_Id_Launch(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 302)
}

// API_Groups returns details about the requested group. If the campaign is not
// valid, API_Groups returns null.
// Example:
/*
POST	/api/groups
		{ "name" : "Test Group",
		  "targets" : [
		  {
		  	"email" : "test@example.com"
		  },
		  { "email" : test2@example.com"
		  }]
		}

RESULT { "name" : "Test Group",
		  "targets" : [
		  {
		  	"email" : "test@example.com"
		  },
		  { "email" : test2@example.com"
		  }]
		}
*/
func API_Groups(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		gs, err := models.GetGroups(ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "Groups not found", http.StatusNotFound) {
			return
		}
		gj, err := json.MarshalIndent(gs, "", "  ")
		if checkError(err, w, "Error marshaling group information", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, gj)
	//POST: Create a new group and return it as JSON
	case r.Method == "POST":
		g := models.Group{}
		// Put the request into a group
		err := json.NewDecoder(r.Body).Decode(&g)
		if checkError(err, w, "Invalid Request", http.StatusBadRequest) {
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
		gj, err := json.MarshalIndent(g, "", "  ")
		if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, gj)
	}
}

// API_Groups_Id returns details about the requested campaign. If the campaign is not
// valid, API_Campaigns_Id returns null.
func API_Groups_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		g, err := models.GetGroup(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No group found", http.StatusNotFound) {
			return
		}
		gj, err := json.MarshalIndent(g, "", "  ")
		if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, gj)
	case r.Method == "DELETE":
		g, err := models.GetGroup(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No group found", http.StatusNotFound) {
			return
		}
		err = models.DeleteGroup(&g)
		if checkError(err, w, "Error deleting group", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, []byte("{\"success\" : \"true\"}"))
	case r.Method == "PUT":
		_, err := models.GetGroup(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No group found", http.StatusNotFound) {
			return
		}
		g := models.Group{}
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
		gj, err := json.MarshalIndent(g, "", "  ")
		if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, gj)
	}
}

func API_Templates(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ts, err := models.GetTemplates(ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "Templates not found", http.StatusNotFound) {
			return
		}
		tj, err := json.MarshalIndent(ts, "", "  ")
		if checkError(err, w, "Error marshaling template information", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, tj)
	//POST: Create a new group and return it as JSON
	case r.Method == "POST":
		t := models.Template{}
		// Put the request into a group
		err := json.NewDecoder(r.Body).Decode(&t)
		if checkError(err, w, "Invalid Request", http.StatusBadRequest) {
			return
		}
		t.ModifiedDate = time.Now()
		err = models.PostTemplate(&t, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "Error inserting template", http.StatusInternalServerError) {
			return
		}
		tj, err := json.MarshalIndent(t, "", "  ")
		if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, tj)
	}
}

func API_Templates_Id(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 302)
}

func writeJSON(w http.ResponseWriter, c []byte) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", c)
}
