package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/auth"
	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// ErrUsernameTaken is thrown when a user attempts to register a username that is taken.
var ErrUsernameTaken = errors.New("Username already taken")

// ErrEmptyUsername is thrown when a user attempts to register a username that is taken.
var ErrEmptyUsername = errors.New("No username provided")

// ErrEmptyRole is throws when no role is provided when creating or modifying a user.
var ErrEmptyRole = errors.New("No role specified")

// ErrInsufficientPermission is thrown when a user attempts to change an
// attribute (such as the role) for which they don't have permission.
var ErrInsufficientPermission = errors.New("Permission denied")

// userRequest is the payload which represents the creation of a new user.
type userRequest struct {
	Username               string `json:"username"`
	Password               string `json:"password"`
	Role                   string `json:"role"`
	PasswordChangeRequired bool   `json:"password_change_required"`
	AccountLocked          bool   `json:"account_locked"`
}

func (ur *userRequest) Validate(existingUser *models.User) error {
	switch {
	case ur.Username == "":
		return ErrEmptyUsername
	case ur.Role == "":
		return ErrEmptyRole
	}
	// Verify that the username isn't already taken. We consider two cases:
	// * We're creating a new user, in which case any match is a conflict
	// * We're modifying a user, in which case any match with a different ID is
	//   a conflict.
	possibleConflict, err := models.GetUserByUsername(ur.Username)
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
		err = auth.CheckPasswordPolicy(ur.Password)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		hash, err := auth.GeneratePasswordHash(ur.Password)
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
			Username:               ur.Username,
			Hash:                   hash,
			ApiKey:                 auth.GenerateSecureKey(auth.APIKeyLength),
			Role:                   role,
			RoleID:                 role.ID,
			PasswordChangeRequired: ur.PasswordChangeRequired,
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
	hasSystem, err := currentUser.HasPermission(models.PermissionModifySystem)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	if !hasSystem && currentUser.Id != id {
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
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		log.Infof("Deleted user account for %s", existingUser.Username)
		JSONResponse(w, models.Response{Success: true, Message: "User deleted Successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		ur := &userRequest{}
		err = json.NewDecoder(r.Body).Decode(ur)
		if err != nil {
			log.Errorf("error decoding user request: %v", err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		err = ur.Validate(&existingUser)
		if err != nil {
			log.Errorf("invalid user request received: %v", err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		existingUser.Username = ur.Username
		// Only users with the ModifySystem permission are able to update a
		// user's role. This prevents a privilege escalation letting users
		// upgrade their own account.
		if !hasSystem && ur.Role != existingUser.Role.Slug {
			JSONResponse(w, models.Response{Success: false, Message: ErrInsufficientPermission.Error()}, http.StatusBadRequest)
			return
		}
		role, err := models.GetRoleBySlug(ur.Role)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		// If our user is trying to change the role of an admin, we need to
		// ensure that it isn't the last user account with the Admin role.
		if existingUser.Role.Slug == models.RoleAdmin && existingUser.Role.ID != role.ID {
			err = models.EnsureEnoughAdmins()
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
				return
			}
		}
		existingUser.Role = role
		existingUser.RoleID = role.ID
		// We don't force the password to be provided, since it may be an admin
		// managing the user's account, and making a simple change like
		// updating the username or role. However, if it _is_ provided, we'll
		// update the stored hash after validating the new password meets our
		// password policy.
		//
		// Note that we don't force the current password to be provided. The
		// assumption here is that the API key is a proper bearer token proving
		// authenticated access to the account.
		existingUser.PasswordChangeRequired = ur.PasswordChangeRequired
		if ur.Password != "" {
			err = auth.CheckPasswordPolicy(ur.Password)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
				return
			}
			hash, err := auth.GeneratePasswordHash(ur.Password)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
				return
			}
			existingUser.Hash = hash
		}
		existingUser.AccountLocked = ur.AccountLocked
		err = models.PutUser(&existingUser)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, existingUser, http.StatusOK)
	}
}
