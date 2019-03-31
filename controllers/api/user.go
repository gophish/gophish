package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/auth"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (as *Server) Users(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		us, err := models.GetUsers()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, us, http.StatusOK)
		return
	case r.Method == "POST":
		ur := &UserRequest{}
		err := json.NewDecoder(r.Body).Decode(ur)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		// Register the user
		hash, err := auth.NewHash(ur.Password)
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
			ApiKey:   auth.GenerateSecureKey(),
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
}

// User contains functions to retrieve or delete a single user. Users with
// the ModifySystem permission can view and modify any user. Otherwise, users
// may only view or delete their own account.
func (as *Server) User(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	// If the user doesn't have ModifySystem permissions, verify that they're
	// only taking action on their account
	currentUser := ctx.Get(r, "user").(models.User)
	ok, err := currentUser.HasPermission(models.PermissionModifyObjects)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	if !ok && currentUser.Id != id {
		JSONResponse(w, models.Response{Success: false, Message: http.StatusText(http.StatusForbidden)}, http.StatusForbidden)
		return
	}
	u, err := models.GetUser(id)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "User not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, u, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteUser(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting user"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "User Delete Successfully!"}, http.StatusOK)
	}
}
