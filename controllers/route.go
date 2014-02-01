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
	"fmt"
	"html/template"
	"net/http"

	ctx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/gophish/auth"
	mid "github.com/jordan-wright/gophish/middleware"
	"github.com/jordan-wright/gophish/models"
)

var templateDelims = []string{"{{%", "%}}"}

func CreateRouter() *mux.Router {
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/login", Login)
	router.HandleFunc("/register", Register)
	router.HandleFunc("/", Use(Base, mid.RequireLogin))
	router.HandleFunc("/campaigns/{id}", Use(Campaigns_Id, mid.RequireLogin))
	router.HandleFunc("/users", Use(Users, mid.RequireLogin))
	router.HandleFunc("/settings", Use(Settings, mid.RequireLogin))

	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", Use(API, mid.RequireLogin))
	api.HandleFunc("/campaigns", Use(API_Campaigns, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id}", Use(API_Campaigns_Id, mid.RequireAPIKey))

	//Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	return router
}

// Use allows us to stack middleware to process the request
// Example taken from https://github.com/gorilla/mux/pull/36#issuecomment-25849172
func Use(handler http.HandlerFunc, mid ...func(http.Handler) http.HandlerFunc) http.HandlerFunc {
	for _, m := range mid {
		handler = m(handler)
	}
	return handler
}

func Register(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	Login(w, r)
}

func Base(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User)}
	getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
}

func Users(w http.ResponseWriter, r *http.Request) {
	getTemplate(w, "users").ExecuteTemplate(w, "base", nil)
}

func Settings(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User  models.User
		Title string
	}{Title: "Settings", User: ctx.Get(r, "user").(models.User)}
	getTemplate(w, "settings").ExecuteTemplate(w, "base", params)
}

func Campaigns_Id(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
	}{Title: "Results", User: ctx.Get(r, "user").(models.User)}
	getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
}

func Login(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
	}{Title: "Login"}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		getTemplate(w, "login").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to login
		err := r.ParseForm()
		if checkError(err, w, "Error parsing request") {
			return
		}
		succ, err := auth.Login(r)
		if checkError(err, w, "Error logging in") {
			return
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
	templates := template.New("template")
	templates.Delims(templateDelims[0], templateDelims[1])
	_, err := templates.ParseFiles("templates/base.html", "templates/nav.html", "templates/"+tmpl+".html", "templates/flashes.html")
	if err != nil {
		fmt.Println(err)
	}
	return template.Must(templates, err)
}

func checkError(e error, w http.ResponseWriter, m string) bool {
	if e != nil {
		fmt.Println(e)
		http.Error(w, "Error: "+m, http.StatusInternalServerError)
		return true
	}
	return false
}
