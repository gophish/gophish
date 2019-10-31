package api

//TODO

import (
  // "encoding/json"
  // "errors"
  "net/http"

  // "github.com/gorilla/mux"
)



func (as *Server) Webhooks(w http.ResponseWriter, r *http.Request) {

  JSONResponse(w, "TODO", http.StatusOK)
  return
}

func (as *Server) PingWebhook(w http.ResponseWriter, r *http.Request) {
  // vars := mux.Vars(r)
  // id, _ := strconv.ParseInt(vars["id"], 0, 64)

  JSONResponse(w, "TODO", http.StatusOK)
  return
}
