package controllers

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
	"html/template"
	"net/http"

	ctx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/gophish/auth"
	"github.com/jordan-wright/gophish/models"
)

func CreateRouter() http.Handler {
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/", Base)
	router.HandleFunc("/login", Login)
	router.HandleFunc("/register", Register)
	router.HandleFunc("/campaigns", Base_Campaigns)
	router.HandleFunc("/users", Users)
	router.HandleFunc("/settings", Settings)

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
	// Example of using session - will be removed.
	getTemplate(w, "dashboard").ExecuteTemplate(w, "base", nil)
}

func Users(w http.ResponseWriter, r *http.Request) {
	getTemplate(w, "users").ExecuteTemplate(w, "base", nil)
}

func Settings(w http.ResponseWriter, r *http.Request) {
	getTemplate(w, "settings").ExecuteTemplate(w, "base", nil)
}

func Base_Campaigns(w http.ResponseWriter, r *http.Request) {
	//session, _ := auth.Store.Get(r, "gophish")
	getTemplate(w, "dashboard").ExecuteTemplate(w, "base", nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
	}{}
	session := ctx.Get(r, "session").(*sessions.Session)
	params.Title = "Login"
	switch {
	case r.Method == "GET":
		getTemplate(w, "login").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to login
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing request", http.StatusInternalServerError)
		}
		succ, err := auth.CheckLogin(r)
		if err != nil {
			http.Error(w, "Error logging in", http.StatusInternalServerError)
		}
		//If we've logged in, save the session and redirect to the dashboard
		if succ {
			session.Save(r, w)
			http.Redirect(w, r, "/", 302)
		} else {
			session.AddFlash(models.Flash{
				Type:    "danger",
				Message: "Invalid Username/Password",
			})
			params.Flashes = session.Flashes()
			getTemplate(w, "login").ExecuteTemplate(w, "base", params)
		}
	}
}

func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	return template.Must(template.New("template").ParseFiles("templates/base.html", "templates/nav.html", "templates/"+tmpl+".html", "templates/flashes.html"))
}
