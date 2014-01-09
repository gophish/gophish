package auth

import (
	"database/sql"
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
	ctx "github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/gophish/models"
)

var Store = sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(64)))

// CheckLogin attempts to request a SQL record with the given username.
// If successful, it then compares the received bcrypt hash.
// If all checks pass, this function sets the session id for later use.
func CheckLogin(r *http.Request) (bool, error) {
	username, password := r.FormValue("username"), r.FormValue("password")
	session, _ := Store.Get(r, "gophish")
	stmt, err := db.Prepare("SELECT * FROM Users WHERE username=?")
	if err != nil {
		return false, err
	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	u := models.User{}
	err = stmt.QueryRow(username).Scan(&u.Id, &u.Username, &u.Hash, &u.APIKey)
	if err == sql.ErrNoRows {
		return false, err
	}
	//If we've made it here, we should have a valid user stored in u
	//Let's check the password
	err = bcrypt.CompareHashAndPassword(u.Hash, hash)
	if err != nil {
		ctx.Set(r, User, nil)
		//Return false, but don't return an error
		return false, nil
	}
	ctx.Set(r, models.User, u)
	session.Values["id"] = GetUser(r).Id
	return true, nil
}

func GetUser(r *http.Request) User {
	if rv := ctx.Get(r, models.User); rv != nil {
		return rv.(models.User)
	}
	return nil
}
