package controllers

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
	"github.com/justinas/nosurf"
)

var templateDelims = []string{"{{%", "%}}"}

func CreateRouter() *nosurf.CSRFHandler {
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/login", Login)
	router.HandleFunc("/logout", Use(Logout, mid.RequireLogin))
	router.HandleFunc("/register", Register)
	router.HandleFunc("/", Use(Base, mid.RequireLogin))
	router.HandleFunc("/campaigns/{id:[0-9]+}", Use(Campaigns_Id, mid.RequireLogin))
	router.HandleFunc("/users", Use(Users, mid.RequireLogin))
	router.HandleFunc("/settings", Use(Settings, mid.RequireLogin))

	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", Use(API, mid.RequireLogin))
	api.HandleFunc("/reset", Use(API_Reset, mid.RequireLogin))
	api.HandleFunc("/campaigns", Use(API_Campaigns, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}", Use(API_Campaigns_Id, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/id:[0-9]+}", Use(API_Campaigns_Id_Launch, mid.RequireAPIKey))
	api.HandleFunc("/groups", Use(API_Groups, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}", Use(API_Groups_Id, mid.RequireAPIKey))

	//Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	//Setup CSRF Protection
	csrfHandler := nosurf.New(router)
	csrfHandler.ExemptGlob("/api/*/*")
	csrfHandler.ExemptGlob("/static/*")
	return csrfHandler
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

func Logout(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	session := ctx.Get(r, "session").(*sessions.Session)
	delete(session.Values, "id")
	session.AddFlash(models.Flash{
		Type:    "success",
		Message: "You have successfully logged out.",
	})
	session.Save(r, w)
	http.Redirect(w, r, "login", 302)
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
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
	}{Title: "Users & Groups", User: ctx.Get(r, "user").(models.User)}
	getTemplate(w, "users").ExecuteTemplate(w, "base", params)
}

func Settings(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Settings", User: ctx.Get(r, "user").(models.User)}
	session := ctx.Get(r, "session").(*sessions.Session)
	params.Token = nosurf.Token(r)
	params.Flashes = session.Flashes()
	session.Save(r, w)
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
		Token   string
	}{Title: "Login", Token: nosurf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
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
			session.Save(r, w)
			http.Redirect(w, r, "/login", 302)
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
