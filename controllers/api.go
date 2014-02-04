package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	ctx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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
		getTemplate(w, "api_doc").ExecuteTemplate(w, "base", nil)
	}
}

// API (/api/reset) resets a user's API key
func API_Reset(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "POST":
		u := ctx.Get(r, "user").(models.User)
		// Inspired from gorilla/securecookie
		k := make([]byte, 32)
		_, err := io.ReadFull(rand.Reader, k)
		checkError(err, w, "Error setting new API key")
		u.APIKey = fmt.Sprintf("%x", k)
		db.Conn.Exec("UPDATE users SET api_key=? WHERE id=?", u.APIKey, u.Id)
		session := ctx.Get(r, "session").(*sessions.Session)
		session.AddFlash(models.Flash{
			Type:    "success",
			Message: "API Key Successfully Reset",
		})
		session.Save(r, w)
		http.Redirect(w, r, "/settings", 302)
	}
}

// API_Campaigns returns a list of campaigns if requested via GET.
// If requested via POST, API_Campaigns creates a new campaign and returns a reference to it.
func API_Campaigns(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs := []models.Campaign{}
		_, err := db.Conn.Select(&cs, "SELECT c.id, name, created_date, completed_date, status, template FROM campaigns c, users u WHERE c.uid=u.id AND u.api_key=?", ctx.Get(r, "api_key"))
		if err != nil {
			fmt.Println(err)
		}
		/*for c := range cs {
			_, err := db.Conn.Select(&cs.Results, "SELECT r.id ")
		}*/
		cj, err := json.MarshalIndent(cs, "", "  ")
		if checkError(err, w, "Error looking up campaigns") {
			return
		}
		writeJSON(w, cj)
	//POST: Create a new campaign and return it as JSON
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
		if c.Uid == 0 {
			http.Error(w, "Error: Invalid API Key", http.StatusInternalServerError)
			return
		}
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

// API_Campaigns_Id returns details about the requested campaign. If the campaign is not
// valid, API_Campaigns_Id returns null.
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
		cj, err := json.MarshalIndent(c, "", "  ")
		if checkError(err, w, "Error creating JSON response") {
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
func API_Groups(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 302)
}

// API_Campaigns_Id returns details about the requested campaign. If the campaign is not
// valid, API_Campaigns_Id returns null.
func API_Groups_Id(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 302)
}

func writeJSON(w http.ResponseWriter, c []byte) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", c)
}
