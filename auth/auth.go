package auth

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"

	"crypto/rand"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
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
	newPassword := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	u, err := models.GetUserByUsername(username)
	// If the given username already exists, throw an error and return false
	if err == nil {
		return false, ErrUsernameTaken
	}

	// If we have an error which is not simply indicating that no user was found, report it
	if err != nil && err != gorm.ErrRecordNotFound {
		fmt.Println(err)
		return false, err
	}

	u = models.User{}
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
	u.Hash = string(h)
	u.ApiKey = GenerateSecureKey()
	err = models.PutUser(&u)
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
	currentPw := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_new_password")
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
