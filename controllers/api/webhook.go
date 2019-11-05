package api

//TODO

import (
  "net/http"
  "strconv"

  "github.com/gophish/gophish/models"
  "github.com/gorilla/mux"
  log "github.com/gophish/gophish/logger"
)

func (as *Server) Webhooks(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "GET":
    whs, err := models.GetWebhooks()
    if err != nil {
      log.Error(err)
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }
    JSONResponse(w, whs, http.StatusOK)
  }
}

func (as *Server) Webhook(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, _ := strconv.ParseInt(vars["id"], 0, 64)
  wh, err := models.GetWebhook(id)
  if err != nil {
    JSONResponse(w, models.Response{Success: false, Message: "Webhook not found"}, http.StatusNotFound)
    return
  }


  switch {
  case r.Method == "GET":
    JSONResponse(w, wh, http.StatusOK)
  case r.Method == "DELETE":
    err = models.DeleteWebhook(id)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }
    log.Infof("Deleted webhook with id: %d", id)
    JSONResponse(w, models.Response{Success: true, Message: "Webhook deleted Successfully!"}, http.StatusOK)
  case r.Method == "PUT":
    //TODO
    JSONResponse(w, wh, http.StatusOK)

 
  }
}

func (as *Server) ValidateWebhook(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, _ := strconv.ParseInt(vars["id"], 0, 64)
  wh, err := models.GetWebhook(id)
  if err != nil {
    log.Error(err)
    JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
    return
  }
  // err = wh.Send("") //TODO empty data
  // if err != nil {
  //   log.Error(err)
  //   JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadGateway)
  //   return
  // }
  JSONResponse(w, wh, http.StatusOK)
}
