package auth

import (
	"errors"
	"net/http"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidPassword is thrown when a user provides an incorrect password.
var ErrInvalidPassword = errors.New("Invalid Password")

// ErrPasswordMismatch is thrown when a user provides a blank password to the register
// or change password functions
var ErrPasswordMismatch = errors.New("Password cannot be blank")

// ErrEmptyPassword is thrown when a user provides a blank password to the register
// or change password functions
var ErrEmptyPassword = errors.New("No password provided")

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

// ChangePassword verifies the current password provided in the request and,
// if it's valid, changes the password for the authenticated user.
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
