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
	"github.com/jordan-wright/gophish/db"
	"github.com/jordan-wright/gophish/models"
)

const (
	IN_PROGRESS string = "In progress"
	WAITING     string = "Waiting"
	COMPLETE    string = "Completed"
	ERROR       string = "Error"
)

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
		u.APIKey = auth.GenerateSecureKey()
		err := db.PutUser(&u)
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
		cs, err := db.GetCampaigns(ctx.Get(r, "user_id").(int64))
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
		// Fill in the details
		c.CreatedDate = time.Now()
		c.CompletedDate = time.Time{}
		c.Status = IN_PROGRESS
		c.Uid = ctx.Get(r, "user_id").(int64)
		// Insert into the DB
		err = db.Conn.Insert(&c)
		if checkError(err, w, "Cannot insert campaign into database", http.StatusInternalServerError) {
			return
		}
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
		c, err := db.GetCampaign(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No campaign found", http.StatusNotFound) {
			return
		}
		cj, err := json.MarshalIndent(c, "", "  ")
		if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, cj)
	case r.Method == "DELETE":
		//c := models.Campaign{}
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
		gs, err := db.GetGroups(ctx.Get(r, "user_id").(int64))
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
		err = db.PostGroup(&g, ctx.Get(r, "user_id").(int64))
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
		g, err := db.GetGroup(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No group found", http.StatusNotFound) {
			return
		}
		gj, err := json.MarshalIndent(g, "", "  ")
		if checkError(err, w, "Error creating JSON response", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, gj)
	case r.Method == "DELETE":
		_, err := db.GetGroup(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No group found", http.StatusNotFound) {
			return
		}
		err = db.DeleteGroup(id)
		if checkError(err, w, "Error deleting group", http.StatusInternalServerError) {
			return
		}
		writeJSON(w, []byte("{\"success\" : \"true\"}"))
	case r.Method == "PUT":
		_, err := db.GetGroup(id, ctx.Get(r, "user_id").(int64))
		if checkError(err, w, "No group found", http.StatusNotFound) {
			return
		}
		g := models.Group{}
		err = json.NewDecoder(r.Body).Decode(&g)
		if g.Id != id {
			http.Error(w, "Error: /:id and group_id mismatch", http.StatusBadRequest)
			return
		}
		err = db.PutGroup(&g, ctx.Get(r, "user_id").(int64))
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

func writeJSON(w http.ResponseWriter, c []byte) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", c)
}
