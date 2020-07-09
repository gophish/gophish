package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

/*
// ReportedEmailsSave handles requests for the /api/reportedemails/save endpoint
func (as *Server) ReportedEmailsSave(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		em := models.ReportedEmail{}
		err := json.NewDecoder(r.Body).Decode(&em)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid email data."}, http.StatusBadRequest)
			return
		}

		err = models.SaveReportedEmail(&em)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Successfully saved reported email."}, http.StatusCreated)

	}

}*/

// ReportedEmailAttachment handles requests for the /api/reported/attachments endpoint
func (as *Server) ReportedEmailAttachment(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)

	att, err := models.GetReportedEmailAttachment(ctx.Get(r, "user_id").(int64), id)

	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	//JSONResponse(w, ems, http.StatusOK)
	data, err := base64.StdEncoding.DecodeString(att.Content)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", att.Header)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

// ReportedEmails handles requests for the /api/reported endpoint
func (as *Server) ReportedEmails(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	emailid := int64(-1)
	offset := int64(-1)
	limit := int64(-1)

	if _, ok := vars["id"]; ok {
		emailid, _ = strconv.ParseInt(vars["id"], 0, 64)
	}

	if _, ok := vars["range"]; ok {
		r := strings.Split(vars["range"], ",")
		offset, _ = strconv.ParseInt(r[0], 0, 64)
		limit, _ = strconv.ParseInt(r[1], 0, 64)
	}

	switch {
	// GET: Return all emails
	case r.Method == "GET":

		ems, err := models.GetReportedEmails(ctx.Get(r, "user_id").(int64), emailid, limit, offset)

		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, ems, http.StatusOK)

	// PUT: Update an email
	case r.Method == "PUT":
		// Get existing email by id
		ems, err := models.GetReportedEmail(ctx.Get(r, "user_id").(int64), emailid)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		if len(ems) > 0 {
			em := ems[0]
			err := json.NewDecoder(r.Body).Decode(&em)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: "Invalid data"}, http.StatusBadRequest)
				return
			}
			err = models.SaveReportedEmail(em)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: "Failed to update email"}, http.StatusBadRequest)
				return
			}

			JSONResponse(w, models.Response{Success: true, Message: "Email record udpated"}, http.StatusCreated)
		} else {
			JSONResponse(w, models.Response{Success: false, Message: "Unable to locate email"}, http.StatusCreated)
		}
	case r.Method == "DELETE":
		ems, err := models.GetReportedEmail(ctx.Get(r, "user_id").(int64), emailid)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		if len(ems) > 0 {
			err := models.DeleteReportedEmail(emailid)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: "Failed to delete email"}, http.StatusBadRequest)
				return
			}
			JSONResponse(w, models.Response{Success: true, Message: "Email deleted"}, http.StatusCreated)
		} else {
			JSONResponse(w, models.Response{Success: false, Message: "Unable to locate email"}, http.StatusCreated)
		}
	}
}
