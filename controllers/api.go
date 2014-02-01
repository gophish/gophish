package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	ctx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jordan-wright/gophish/db"
	"github.com/jordan-wright/gophish/models"
)

const (
	IN_PROGRESS string = "In progress"
	WAITING     string = "Waiting"
	COMPLETE    string = "Completed"
	ERROR       string = "Error"
)

func API(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

	}
	if r.Method == "POST" {
		//Add a new campaign
		//v :=
	}
	if u, err := json.Marshal(ctx.Get(r, "user")); err == nil {
		writeJSON(w, u)
	} else {
		http.Error(w, "Server Error", 500)
	}
}

//API_Campaigns returns a list of campaigns if requested via GET.
//If requested via POST, API_Campaigns creates a new campaign and returns a reference to it.
func API_Campaigns(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs := []models.Campaign{}
		_, err := db.Conn.Select(&cs, "SELECT campaigns.id, name, created_date, completed_date, status, template FROM campaigns, users WHERE campaigns.uid=users.id AND users.api_key=?", ctx.Get(r, "api_key"))
		if err != nil {
			fmt.Println(err)
		}
		cj, err := json.MarshalIndent(cs, "", "  ")
		if checkError(err, w, "Error looking up campaigns") {
			return
		}
		writeJSON(w, cj)
	case r.Method == "POST":
		c := models.Campaign{}
		// Put the request into a campaign
		err := json.NewDecoder(r.Body).Decode(&c)
		checkError(err, w, "Invalid Request")
		// Fill in the details
		c.CreatedDate = time.Now()
		c.CompletedDate = time.Time{}
		c.Status = IN_PROGRESS
		c.Uid, err = db.Conn.SelectInt("SELECT id FROM users WHERE api_key=?", ctx.Get(r, "api_key"))
		if checkError(err, w, "Invalid API Key") {
			return
		}
		// Insert into the DB
		err = db.Conn.Insert(&c)
		if checkError(err, w, "Cannot insert campaign into database") {
			return
		}
		cj, err := json.MarshalIndent(c, "", "  ")
		if checkError(err, w, "Error creating JSON response") {
			return
		}
		writeJSON(w, cj)
	}
}

//API_Campaigns_Id returns details about the requested campaign. If the campaign is not
//valid, API_Campaigns_Id returns null.
func API_Campaigns_Id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 0, 64)
	if checkError(err, w, "Invalid Int") {
		return
	}
	switch {
	case r.Method == "GET":
		c := models.Campaign{}
		err := db.Conn.SelectOne(&c, "SELECT campaigns.id, name, created_date, completed_date, status, template FROM campaigns, users WHERE campaigns.uid=users.id AND campaigns.id =? AND users.api_key=?", id, ctx.Get(r, "api_key"))
		if checkError(err, w, "No campaign found") {
			return
		}
		fmt.Printf("%v\n", c)
		cj, err := json.MarshalIndent(c, "", "  ")
		if checkError(err, w, "Error creating JSON response") {
			return
		}
		writeJSON(w, cj)
	case r.Method == "DELETE":
		//c := models.Campaign{}
	}
}

//API_Doc renders a template describing the API documentation.
func API_Doc(w http.ResponseWriter, r *http.Request) {
	getTemplate(w, "api_doc").ExecuteTemplate(w, "base", nil)
}

func writeJSON(w http.ResponseWriter, c []byte) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", c)
}
