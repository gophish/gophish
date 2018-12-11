package auth

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"crypto/rand"

	ctx "github.com/binodlamsal/gophish/context"
	log "github.com/binodlamsal/gophish/logger"
	"github.com/binodlamsal/gophish/models"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

//init registers the necessary models to be saved in the session later
func init() {
	gob.Register(&models.User{})
	gob.Register(&models.Flash{})
	Store.Options.HttpOnly = true
	// This sets the maxAge to 5 days for all cookies
	Store.MaxAge(86400 * 5)
}

// Store contains the session information for the request
var Store = sessions.NewCookieStore(
	[]byte(securecookie.GenerateRandomKey(64)), //Signing key
	[]byte(securecookie.GenerateRandomKey(32)))

// ErrInvalidPassword is thrown when a user provides an incorrect password.
var ErrInvalidPassword = errors.New("Invalid Password")

// ErrEmptyPassword is thrown when a user provides a blank password to the register
// or change password functions
var ErrEmptyPassword = errors.New("Password cannot be blank")

// ErrPasswordMismatch is thrown when a user provides passwords that do not match
var ErrPasswordMismatch = errors.New("Passwords must match")

// ErrUsernameTaken is thrown when a user attempts to register a username that is taken.
var ErrUsernameTaken = errors.New("Username already taken")

// Login attempts to login the user given a request.
func Login(r *http.Request) (bool, models.User, error) {
	username, password := r.FormValue("username"), r.FormValue("password")
	u, err := models.GetUserByUsername(username)
	if err != nil {
		return false, models.User{}, err
	}
	//If we've made it here, we should have a valid user stored in u
	//Let's check the password
	err = bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password))
	if err != nil {
		return false, models.User{}, ErrInvalidPassword
	}
	return true, u, nil
}

// Register attempts to register the user given a request.
func Register(r *http.Request) (bool, error) {
	username := r.FormValue("username")
	newEmail := r.FormValue("email")
	newPassword := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	role := r.FormValue("roles")
	rid, _ := strconv.ParseInt(role, 10, 0)

	u, err := models.GetUserByUsername(username)
	// If the given username already exists, throw an error and return false
	if err == nil {
		return false, ErrUsernameTaken
	}

	// If we have an error which is not simply indicating that no user was found, report it
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Warn(err)
		return false, err
	}

	u = models.User{}
	ur := models.UserRole{}
	// If we've made it here, we should have a valid username given
	// Check that the passsword isn't blank
	if newPassword == "" {
		return false, ErrEmptyPassword
	}
	// Make sure passwords match
	if newPassword != confirmPassword {
		return false, ErrPasswordMismatch
	}
	// Let's create the password hash
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	u.Username = username
	u.Email = newEmail
	u.Hash = string(h)
	u.ApiKey = GenerateSecureKey()

	currentUser := ctx.Get(r, "user").(models.User)
	currentRole, err := models.GetUserRole(currentUser.Id)

	if err != nil {
		log.Error(err)
	}

	if currentRole.Is(models.Administrator) || currentRole.Is(models.Partner) {
		if rid == models.Customer || rid == models.ChildUser {
			u.Partner = ctx.Get(r, "user").(models.User).Id
		}
	} else if currentRole.Is(models.ChildUser) {
		if rid == models.Customer {
			u.Partner = currentUser.Partner
		}
	}

	err = models.PutUser(&u)

	//Getting the inserted U after inserted
	iu, err := models.GetUserByUsername(username)

	ur.Uid = iu.Id
	ur.Rid = rid

	err = models.PutUserRole(&ur)
	return true, nil
}

// GenerateSecureKey creates a secure key to use
// as an API key
func GenerateSecureKey() string {
	// Inspired from gorilla/securecookie
	k := make([]byte, 32)
	io.ReadFull(rand.Reader, k)
	return fmt.Sprintf("%x", k)
}

func ChangePassword(r *http.Request) error {
	u := ctx.Get(r, "user").(models.User)
	//currentPw := r.FormValue("current_password")
	//newPassword := r.FormValue("new_password")
	//confirmPassword := r.FormValue("confirm_new_password")

	r.ParseForm()                               // Parses the request body
	currentPw := r.Form.Get("current_password") // x will be "" if parameter is not set
	newPassword := r.Form.Get("new_password")
	confirmPassword := r.Form.Get("confirm_new_password")

	// Check the current password
	err := bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(currentPw))
	if err != nil {
		return ErrInvalidPassword
	}
	// Check that the new password isn't blank
	if newPassword == "" {
		return ErrEmptyPassword
	}
	// Check that new passwords match
	if newPassword != confirmPassword {
		return ErrPasswordMismatch
	}
	// Generate the new hash
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Hash = string(h)
	if err = models.PutUser(&u); err != nil {
		return err
	}
	return nil
}

func ChangePasswordByadmin(r *http.Request) error {
	u := ctx.Get(r, "user").(models.User)
	type Usersdata struct {
		Id                   int64  `json:"id"`
		Username             string `json:"username"`
		Email                string `json:"email" `
		New_password         string `json:"new_password" `
		Confirm_new_password string `json:"confirm_new_password" `
		Role                 int64  `json:"role" `
		Hash                 string `json:"-"`
		ApiKey               string `json:"api_key"`
		Partner              int64  `json:"partner" `
	}

	var ud = new(Usersdata)
	err := json.NewDecoder(r.Body).Decode(&ud)

	newPassword := ud.New_password
	confirmPassword := ud.Confirm_new_password

	u.Id = ud.Id
	u.Email = ud.Email
	u.Username = ud.Username
	u.ApiKey = ud.ApiKey
	u.Partner = ud.Partner

	// Check the current password

	// Check that new passwords match  //since this is going to do by admin no longer need to check
	if newPassword != "" && confirmPassword != "" {

		// Check that the new password isn't blank
		if newPassword == "" {
			return ErrEmptyPassword
		}

		if newPassword != confirmPassword {
			return ErrPasswordMismatch
		}

		// Generate the new hash
		h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Hash = string(h)
	}

	// Unset partner for non-customers
	if ud.Role != models.Customer && ud.Role != models.ChildUser {
		u.Partner = 0
	}

	if err = models.PutUser(&u); err != nil {
		return err
	}

	ur := models.UserRole{}
	ur.Uid = ud.Id
	ur.Rid = ud.Role

	//first delete the users roles in update
	if err = models.DeleteUserRoles(ur.Uid); err != nil {
		return err
	}

	//Second save the user roles again
	err = models.PutUserRole(&ur)

	return nil
}
