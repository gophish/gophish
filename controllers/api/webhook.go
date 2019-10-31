package api

//TODO

import (
  // "encoding/json"
  // "errors"
  "net/http"

  "github.com/gophish/gophish/models"
  // "github.com/gorilla/mux"
)



func (as *Server) Webhooks(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "GET":
    wh_s, err := models.GetWebhooks()
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }
    JSONResponse(w, wh_s, http.StatusOK)
    return
  }
}

func (as *Server) PingWebhook(w http.ResponseWriter, r *http.Request) {
  // vars := mux.Vars(r)
  // id, _ := strconv.ParseInt(vars["id"], 0, 64)

  JSONResponse(w, "TODO", http.StatusOK)
  return
}
