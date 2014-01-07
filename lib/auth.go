package main

/*
gophish - Open-Source Phishing Framework

The MIT License (MIT)

Copyright (c) 2013 Jordan Wright

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
	"database/sql"
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
	ctx "github.com/gorilla/context"
)

func CheckLogin(r *http.Request) (bool, error) {
	username, password := r.FormValue("username"), r.FormValue("password")
	session, _ := store.Get(r, "gophish")
	stmt, err := db.Prepare("SELECT * FROM Users WHERE username=?")
	if err != nil {
		return false, err
	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
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
	ctx.Set(r, User, u)
	session.Values["id"] = GetUser(r).Id
	return true, nil
}

func GetUser(r *http.Request) User {
	if rv := ctx.Get(r, User); rv != nil {
		return rv.(User)
	}
	return nil
}
