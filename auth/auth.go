package auth

import (
	"database/sql"
	"encoding/gob"
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
	ctx "github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/gophish/db"
	"github.com/jordan-wright/gophish/models"
)

//init registers the necessary models to be saved in the session later
func init() {
	gob.Register(&models.User{})
}

var Store = sessions.NewCookieStore(
	[]byte(securecookie.GenerateRandomKey(64)), //Signing key
	[]byte(securecookie.GenerateRandomKey(32)))

// CheckLogin attempts to request a SQL record with the given username.
// If successful, it then compares the received bcrypt hash.
// If all checks pass, this function sets the session id for later use.
func Login(r *http.Request) (bool, error) {
	username, password := r.FormValue("username"), r.FormValue("password")
	session, _ := Store.Get(r, "gophish")
	u := models.User{}
	err := db.Conn.SelectOne(&u, "SELECT * FROM Users WHERE username=?", username)
	if err == sql.ErrNoRows {
		//Return false, but don't return an error
		return false, nil
	} else if err != nil {
		return false, err
	}
	//If we've made it here, we should have a valid user stored in u
	//Let's check the password
	err = bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password))
	if err != nil {
		ctx.Set(r, "user", nil)
		//Return false, but don't return an error
		return false, nil
	}
	ctx.Set(r, "user", u)
	session.Values["id"] = u.Id
	return true, nil
}

// GetUserById returns the user that the given id corresponds to. If no user is found, an
// error is thrown.
func GetUserById(id int) (models.User, error) {
	u := models.User{}
	err := db.Conn.SelectOne(&u, "SELECT id, username, apikey FROM Users WHERE id=?", id)
	if err != nil {
		return u, err
	}
	return u, nil
}

// GetUserByAPIKey returns the user that the given API Key corresponds to. If no user is found, an
// error is thrown.
func GetUserByAPIKey(key []byte) (models.User, error) {
	u := models.User{}
	err := db.Conn.SelectOne(&u, "SELECT id, username, apikey FROM Users WHERE apikey=?", key)
	if err != nil {
		return u, err
	}
	return u, nil
}
