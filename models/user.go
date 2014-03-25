package models

import "database/sql"

// User represents the user model for gophish.
type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Hash     string `json:"-"`
	APIKey   string `json:"api_key" db:"api_key"`
}

// GetUser returns the user that the given id corresponds to. If no user is found, an
// error is thrown.
func GetUser(id int64) (User, error) {
	u := User{}
	err := Conn.SelectOne(&u, "SELECT * FROM Users WHERE id=?", id)
	if err != nil {
		return u, err
	}
	return u, nil
}

// GetUserByAPIKey returns the user that the given API Key corresponds to. If no user is found, an
// error is thrown.
func GetUserByAPIKey(key []byte) (User, error) {
	u := User{}
	err := Conn.SelectOne(&u, "SELECT id, username, api_key FROM Users WHERE apikey=?", key)
	if err != nil {
		return u, err
	}
	return u, nil
}

// GetUserByUsername returns the user that the given username corresponds to. If no user is found, an
// error is thrown.
func GetUserByUsername(username string) (User, error) {
	u := User{}
	err := Conn.SelectOne(&u, "SELECT * FROM Users WHERE username=?", username)
	if err != sql.ErrNoRows {
		return u, ErrUsernameTaken
	} else if err != nil {
		return u, err
	}
	return u, nil
}

// PutUser updates the given user
func PutUser(u *User) error {
	_, err := Conn.Update(u)
	return err
}
