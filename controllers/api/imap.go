package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
)

// ImapServerValidate handles requests for the /api/imapserver/validate endpoint
func (as *Server) ImapServerValidate(w http.ResponseWriter, r *http.Request) {
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
		err = models.ValidateIMAP(&s)
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
		if err != nil {
			log.Error(err)
		}
		if len(ss) > 0 {
			ss[0].LastLoginFriendly = humanize.Time(ss[0].LastLogin)
			delta := time.Now().Sub(ss[0].LastLogin).Hours() // Default value if never logged in is "0001-01-01T00:00:00Z"
			if delta > 87600 {
				ss[0].LastLoginFriendly = "Never" //Well, either Never or > 10 years ago.
			}

		}

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
