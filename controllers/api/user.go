package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/auth"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// ErrEmptyPassword is thrown when a user provides a blank password to the register
// or change password functions
var ErrEmptyPassword = errors.New("No password provided")

// ErrUsernameTaken is thrown when a user attempts to register a username that is taken.
var ErrUsernameTaken = errors.New("Username already taken")

// ErrEmptyUsername is thrown when a user attempts to register a username that is taken.
var ErrEmptyUsername = errors.New("No username provided")

// ErrEmptyRole is throws when no role is provided when creating or modifying a user.
var ErrEmptyRole = errors.New("No role specified")

// userRequest is the payload which represents the creation of a new user.
type userRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (ur *userRequest) Validate(existingUser *models.User) error {
	switch {
	case ur.Username == "":
		return ErrEmptyUsername
	case ur.Role == "":
		return ErrEmptyRole
	}
	// Verify that the username isn't already taken. We consider two cases:
	// * We creating a new user, in which case any match is a conflict
	// * We're modifying a user, in which case any match with a different ID is
	//   a conflict.
	possibleConflict, err := GetUserByUsername(ur.Username)
	if err == nil {
		if existingUser == nil {
			return ErrUsernameTaken
		}
		if possibleConflict.Id != existingUser.Id {
			return ErrUsernameTaken
		}
	}
	// If we have an error which is not simply indicating that no user was found, report it
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

// Users contains functions to retrieve a list of existing users or create a
// new user. Users with the ModifySystem permissions can view and create users.
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
			JSONResponse(w, models.Response{Success: false, Message: ErrEmptyPassword}, http.StatusBadRequest)
			return
		}
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
	// If the user doesn't have ModifySystem permissions, we need to verify
	// that they're only taking action on their account.
	currentUser := ctx.Get(r, "user").(models.User)
	ok, err := currentUser.HasPermission(models.PermissionModifySystem)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	if !ok && currentUser.Id != id {
		JSONResponse(w, models.Response{Success: false, Message: http.StatusText(http.StatusForbidden)}, http.StatusForbidden)
		return
	}
	existingUser, err := models.GetUser(id)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "User not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, existingUser, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteUser(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting user"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "User deleted Successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		ur := &userRequest{}
		err = json.NewDecoder(r.Body).Decode(ur)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		err = ur.Validate(existingUser)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		existingUser.Username = ur.Username
		role, err := models.GetRoleBySlug(ur.Role)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		existingUser.Role = role
		existingUser.RoleID = role.ID
		// We don't force the password to be provided, since it may be an admin
		// managing the user's account, and making a simple change like
		// updating the username or role. However, if it _is_ provided, we'll
		// update the stored hash.
		if ur.Password != "" {
			hash, err := auth.NewHash(ur.Password)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
				return
			}
			existingUser.Hash = hash
		}
		err = models.PutUser(&existingUser)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, user, http.StatusOK)
	}
}