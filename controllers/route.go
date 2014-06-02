package controllers

import (
	"html/template"
	"log"
	"net/http"
	"os"

	ctx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/gophish/auth"
	mid "github.com/jordan-wright/gophish/middleware"
	"github.com/jordan-wright/gophish/models"
	"github.com/justinas/nosurf"
)

var templateDelims = []string{"{{%", "%}}"}
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

func CreateRouter() *nosurf.CSRFHandler {
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/login", Login)
	router.HandleFunc("/logout", Use(Logout, mid.RequireLogin))
	router.HandleFunc("/register", Register)
	router.HandleFunc("/", Use(Base, mid.RequireLogin))
	router.HandleFunc("/settings", Use(Settings, mid.RequireLogin))

	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api = api.StrictSlash(true)
	api.HandleFunc("/", Use(API, mid.RequireLogin))
	api.HandleFunc("/reset", Use(API_Reset, mid.RequireLogin))
	api.HandleFunc("/campaigns/", Use(API_Campaigns, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}", Use(API_Campaigns_Id, mid.RequireAPIKey))
	api.HandleFunc("/groups/", Use(API_Groups, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}", Use(API_Groups_Id, mid.RequireAPIKey))
	api.HandleFunc("/templates/", Use(API_Templates, mid.RequireAPIKey))
	api.HandleFunc("/templates/{id:[0-9]+}", Use(API_Templates_Id, mid.RequireAPIKey))

	// Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// Setup CSRF Protection
	csrfHandler := nosurf.New(router)
	// Exempt API routes and Static files
	csrfHandler.ExemptGlob("/api/campaigns/*")
	csrfHandler.ExemptGlob("/api/groups/*")
	csrfHandler.ExemptGlob("/api/templates/*")
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
	params := struct {
		Title   string
		Flashes []interface{}
		User    models.User
		Token   string
	}{Title: "Register", Token: nosurf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		getTemplate(w, "register").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to register
		succ, err := auth.Register(r)
		//If we've registered, redirect to the login page
		if succ {
			session.AddFlash(models.Flash{
				Type:    "success",
				Message: "Registration successful!.",
			})
			session.Save(r, w)
			http.Redirect(w, r, "/login", 302)
		} else {
			// Check the error
			m := ""
			if err == models.ErrUsernameTaken {
				m = "Username already taken"
			} else {
				m = "Unknown error - please try again"
				Logger.Println(err)
			}
			session.AddFlash(models.Flash{
				Type:    "danger",
				Message: m,
			})
			session.Save(r, w)
			http.Redirect(w, r, "/register", 302)
		}

	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	session := ctx.Get(r, "session").(*sessions.Session)
	delete(session.Values, "id")
	Flash(w, r, "success", "You have successfully logged out")
	http.Redirect(w, r, "login", 302)
}

func Base(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r)}
	getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
}

func Settings(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "POST":
		err := auth.ChangePassword(r)
		msg := models.Response{Success: true, Message: "Settings Updated Successfully"}
		if err == auth.ErrInvalidPassword {
			msg.Message = "Invalid Password"
			msg.Success = false
		} else if err != nil {
			msg.Message = "Unknown Error Occured"
			msg.Success = false
		}
		writeJSON(w, msg)
	}
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
		templates := template.New("template")
		templates.Delims(templateDelims[0], templateDelims[1])
		_, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")
		if err != nil {
			Logger.Println(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to login
		succ, err := auth.Login(r)
		if err != nil {
			Logger.Println(err)
		}
		//If we've logged in, save the session and redirect to the dashboard
		if succ {
			session.Save(r, w)
			http.Redirect(w, r, "/", 302)
		} else {
			Flash(w, r, "danger", "Invalid Username/Password")
			http.Redirect(w, r, "/login", 302)
		}
	}
}

func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template")
	templates.Delims(templateDelims[0], templateDelims[1])
	_, err := templates.ParseFiles("templates/base.html", "templates/"+tmpl+".html", "templates/flashes.html")
	if err != nil {
		Logger.Println(err)
	}
	return template.Must(templates, err)
}

func checkError(e error, w http.ResponseWriter, m string, c int) bool {
	if e != nil {
		Logger.Println(e)
		w.WriteHeader(c)
		writeJSON(w, models.Response{Success: false, Message: m})
		return true
	}
	return false
}

func Flash(w http.ResponseWriter, r *http.Request, t string, m string) {
	session := ctx.Get(r, "session").(*sessions.Session)
	session.AddFlash(models.Flash{
		Type:    t,
		Message: m,
	})
	session.Save(r, w)
}
