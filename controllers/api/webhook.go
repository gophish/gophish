package api

//TODO

import (
  "net/http"

  ctx "github.com/gophish/gophish/context"
  "github.com/gophish/gophish/models"
)



func (as *Server) Webhooks(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "GET":
    whs, err := models.GetWebhooks(ctx.Get(r, "user_id").(int64))
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }
    JSONResponse(w, whs, http.StatusOK)
    return
  }
}

func (as *Server) ValidateWebhook(w http.ResponseWriter, r *http.Request) {
  // vars := mux.Vars(r)
  // id, _ := strconv.ParseInt(vars["id"], 0, 64)

  JSONResponse(w, "TODO", http.StatusOK)
  return
}
