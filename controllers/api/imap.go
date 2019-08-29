package api

import (
	"encoding/json"
	"net/http"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
)

// ImapServerTest handles requests for the /api/imapserver/test endpoint
func (as *Server) ImapServerTest(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		JSONResponse(w, models.Response{Success: false, Message: "Only POSTs allowed"}, http.StatusBadRequest)
	case r.Method == "POST":
		s := models.IMAP{}
		// Put the request into a page
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		err = models.TestIMAP(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusOK)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Successful login."}, http.StatusCreated)
	}
}

// ImapServer handles requests for the /api/imapserver/ endpoint
func (as *Server) ImapServer(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ss, err := models.GetIMAP(ctx.Get(r, "user_id").(int64))
		//err := models.DeleteIMAP(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		//JSONResponse(w, models.Response{Success: true, Message: "IMAP Deleted Successfully"}, http.StatusOK)
		JSONResponse(w, ss, http.StatusOK)
	//POST: Update database
	case r.Method == "POST":
		s := models.IMAP{}
		// Put the request into a page
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid data. Please check your IMAP settings."}, http.StatusBadRequest)
			return
		}
		s.ModifiedDate = time.Now().UTC()
		s.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostIMAP(&s, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Successfully saved IMAP settings."}, http.StatusCreated)
	}
}

// IMAPProfile contains functions to handle the GET'ing, DELETE'ing, and PUT'ing
// of an IMAP object (Used to be SendingProfile)
func (as *Server) IMAPProfile(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	s, err := models.GetIMAP(ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "IMAP not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, s, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteIMAP(ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting IMAP"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "IMAP Deleted Successfully"}, http.StatusOK)
	}
}
