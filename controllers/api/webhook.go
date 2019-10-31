package api

//TODO

import (
  // "encoding/json"
  // "errors"
  "net/http"

  // "github.com/gorilla/mux"
)



func (as *Server) Webhooks(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "GET":
    us, err := models.GetWebhooks()
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }
    JSONResponse(w, us, http.StatusOK)
    return
  case r.Method == "POST":
    ur := &userRequest{}
    err := json.NewDecoder(r.Body).Decode(ur)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
      return
    }
    err = ur.Validate(nil)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
      return
    }
    if ur.Password == "" {
      JSONResponse(w, models.Response{Success: false, Message: ErrEmptyPassword.Error()}, http.StatusBadRequest)
      return
    }
    hash, err := util.NewHash(ur.Password)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }
    role, err := models.GetRoleBySlug(ur.Role)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }
    user := models.User{
      Username: ur.Username,
      Hash:     hash,
      ApiKey:   util.GenerateSecureKey(),
      Role:     role,
      RoleID:   role.ID,
    }
    err = models.PutUser(&user)
    if err != nil {
      JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
      return
    }
    JSONResponse(w, user, http.StatusOK)
    return
  }
  JSONResponse(w, "TODO", http.StatusOK)
  return
}

func (as *Server) PingWebhook(w http.ResponseWriter, r *http.Request) {
  // vars := mux.Vars(r)
  // id, _ := strconv.ParseInt(vars["id"], 0, 64)

  JSONResponse(w, "TODO", http.StatusOK)
  return
}
