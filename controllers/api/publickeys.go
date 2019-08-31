package api

import (
	"encoding/json"
	"net/http"

	"strconv"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// PublicKeys contains functions to handle the getting of all public keys related to this user, and adding of a new public key to user
// of a public key object
func (as *Server) PublicKeys(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ps, err := models.GetPublicKeys(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, ps, http.StatusOK)

	case r.Method == "POST":
		p := models.PublicKey{}
		// Put the request into a public key
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		// Check to make sure the name is unique
		_, err = models.GetPublicKeyByName(p.FriendlyName, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Public key name already in use"}, http.StatusConflict)
			log.Error(err)
			return
		}
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostPublicKey(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusCreated)

	}
}

//PublicKey contains functions to handle the GET'ing, DELETING'ing and PUT'ing
// of a public key object
func (as *Server) PublicKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	p, err := models.GetPublicKey(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Public key not found"}, http.StatusNotFound)
		return
	}

	switch {
	case r.Method == "GET":
		JSONResponse(w, p, http.StatusOK)

	case r.Method == "DELETE":
		c, err := models.GetCampaignByPublicKey(id, ctx.Get(r, "user_id").(int64))
		if err == nil {
			log.Error("Could not delete public key as it was associated with a campaign")
			JSONResponse(w, models.Response{Success: false, Message: "Public key associated with campaign " + c.Name}, http.StatusBadRequest)
			return
		}
		err = models.DeletePublicKey(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting public key"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Public key Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		p = models.PublicKey{}
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			log.Error(err)
		}
		if p.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "/:id and /:public_key_id mismatch"}, http.StatusBadRequest)
			return
		}
		p.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PutPublicKey(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, p, http.StatusOK)

	}

}
