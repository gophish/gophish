package api

//TODO

import (
  "net/http"
  "strconv"
  "encoding/json"

  "github.com/gophish/gophish/models"
  // "github.com/gophish/gophish/webhook"
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

  case r.Method == "POST":
    wh := models.Webhook{}
    err := json.NewDecoder(r.Body).Decode(&wh)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
      return
    }
    err = models.PostWebhook(&wh)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
      return
    }
    JSONResponse(w, wh, http.StatusCreated)
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
    wh = models.Webhook{}
    err = json.NewDecoder(r.Body).Decode(&wh)
    wh.Id = id;
    wh.IsActive = false;
    err = models.PutWebhook(&wh)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
      return
    }
    JSONResponse(w, wh, http.StatusOK)
  }
}

func (as *Server) PingWebhook(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "POST":
    vars := mux.Vars(r)
    id, _ := strconv.ParseInt(vars["id"], 0, 64)
    wh, err := models.GetWebhook(id)
    if err != nil {
      log.Error(err)
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }




    // httpCnt := &http.Client{} //TODO add timeout and other stuff
    // whTr := &webhook.Transport{Client: httpCnt}
    // err = whTr.Send(wh.Url, wh.Secret, "{}")
    // if err != nil {
    //   log.Error(err)
    //   JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadGateway)
    //   return
    // }


    //TODO update it here inplace?
    if !wh.IsActive {
      wh.IsActive = true;
      err = models.PutWebhook(&wh)
      if err != nil {
        JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
        return
      }
    }

    JSONResponse(w, wh, http.StatusOK)
  }
}
