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
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"html/template"
	"net/http"
)

var store = sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(64)))

func createRouter() http.Handler {
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/", Base)
	router.HandleFunc("/login", Login)
	router.HandleFunc("/register", Register)
	router.HandleFunc("/campaigns", Base_Campaigns)

	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", API)
	api.HandleFunc("/campaigns", API_Campaigns)
	api.HandleFunc("/campaigns/{id}", API_Campaigns_Id)
	api.HandleFunc("/doc", API_Doc)

	//Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	return router
}

func Register(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	Login(w, r)
}

func Base(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "gophish")
	// Example of saving session - will be removed.
	session.Save(r, w)
	renderTemplate(w, "dashboard")
}

func Base_Campaigns(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "gophish")
	renderTemplate(w, "dashboard")
}

func Login(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login")
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
	t := template.Must(template.New("template").ParseFiles("templates/base.html", "templates/nav.html", "templates/"+tmpl+".html"))
	t.ExecuteTemplate(w, "base", "T")
}
