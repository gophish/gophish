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

var Store = sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(64)))

// CheckLogin attempts to request a SQL record with the given username.
// If successful, it then compares the received bcrypt hash.
// If all checks pass, this function sets the session id for later use.
func CheckLogin(r *http.Request) (bool, error) {
	username, password := r.FormValue("username"), r.FormValue("password")
	session, _ := Store.Get(r, "gophish")
	stmt, err := db.Conn.Prepare("SELECT * FROM Users WHERE username=?")
	if err != nil {
		return false, err
	}
	u := models.User{}
	err = stmt.QueryRow(username).Scan(&u.Id, &u.Username, &u.Hash, &u.APIKey)
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

func GetUser(id int) (models.User, error) {
	u := models.User{}
	stmt, err := db.Conn.Prepare("SELECT * FROM Users WHERE id=?")
	if err != nil {
		return u, err
	}
	err = stmt.QueryRow(id).Scan(&u.Id, &u.Username, &u.Hash, &u.APIKey)
	if err != nil {
		//Return false, but don't return an error
		return u, err
	}
	return u, nil
}
