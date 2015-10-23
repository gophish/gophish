package controllers

import (
	"fmt"
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

// Logger is used to send logging messages to stdout.
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

// CreateAdminRouter creates the routes for handling requests to the web interface.
// This function returns an http.Handler to be used in http.ListenAndServe().
func CreateAdminRouter() http.Handler {
	router := mux.NewRouter()
	// Base Front-end routes
	router.HandleFunc("/", Use(Base, mid.RequireLogin))
	router.HandleFunc("/login", Login)
	router.HandleFunc("/logout", Use(Logout, mid.RequireLogin))
	router.HandleFunc("/campaigns", Use(Campaigns, mid.RequireLogin))
	router.HandleFunc("/campaigns/{id:[0-9]+}", Use(CampaignID, mid.RequireLogin))
	router.HandleFunc("/templates", Use(Templates, mid.RequireLogin))
	router.HandleFunc("/users", Use(Users, mid.RequireLogin))
	router.HandleFunc("/landing_pages", Use(LandingPages, mid.RequireLogin))
	router.HandleFunc("/register", Register)
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
	api.HandleFunc("/pages/", Use(API_Pages, mid.RequireAPIKey))
	api.HandleFunc("/pages/{id:[0-9]+}", Use(API_Pages_Id, mid.RequireAPIKey))
	api.HandleFunc("/import/group", API_Import_Group)
	api.HandleFunc("/import/email", API_Import_Email)
	api.HandleFunc("/import/site", API_Import_Site)

	// Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// Setup CSRF Protection
	csrfHandler := nosurf.New(router)
	// Exempt API routes and Static files
	csrfHandler.ExemptGlob("/api/campaigns")
	csrfHandler.ExemptGlob("/api/campaigns/*")
	csrfHandler.ExemptGlob("/api/groups")
	csrfHandler.ExemptGlob("/api/groups/*")
	csrfHandler.ExemptGlob("/api/templates")
	csrfHandler.ExemptGlob("/api/templates/*")
	csrfHandler.ExemptGlob("/api/pages")
	csrfHandler.ExemptGlob("/api/pages/*")
	csrfHandler.ExemptGlob("/api/import/*")
	csrfHandler.ExemptGlob("/static/*")
	return Use(csrfHandler.ServeHTTP, mid.GetContext)
}

// CreatePhishingRouter creates the router that handles phishing connections.
func CreatePhishingRouter() http.Handler {
	router := mux.NewRouter()
	router.PathPrefix("/static").Handler(http.FileServer(http.Dir("./static/endpoint/")))
	router.HandleFunc("/track", PhishTracker)
	router.HandleFunc("/{path:.*}", PhishHandler)
	return router
}

// PhishTracker tracks emails as they are opened, updating the status for the given Result
func PhishTracker(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.Form.Get("rid")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	rs, err := models.GetResult(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	c, err := models.GetCampaign(rs.CampaignId, rs.UserId)
	if err != nil {
		Logger.Println(err)
	}
	c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_OPENED})
	err = rs.UpdateStatus(models.EVENT_OPENED)
	if err != nil {
		Logger.Println(err)
	}
	w.Write([]byte(""))
}

// PhishHandler handles incoming client connections and registers the associated actions performed
// (such as clicked link, etc.)
func PhishHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.Form.Get("rid")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	rs, err := models.GetResult(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rs.UpdateStatus(models.STATUS_SUCCESS)
	c, err := models.GetCampaign(rs.CampaignId, rs.UserId)
	if err != nil {
		Logger.Println(err)
	}
	p, err := models.GetPage(c.PageId, c.UserId)
	if err != nil {
		Logger.Println(err)
	}
	c.AddEvent(models.Event{Email: rs.Email, Message: models.EVENT_CLICKED})
	w.Write([]byte(p.HTML))
}

// Use allows us to stack middleware to process the request
// Example taken from https://github.com/gorilla/mux/pull/36#issuecomment-25849172
func Use(handler http.HandlerFunc, mid ...func(http.Handler) http.HandlerFunc) http.HandlerFunc {
	for _, m := range mid {
		handler = m(handler)
	}
	return handler
}

// Register creates a new user
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
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/register.html", "templates/flashes.html")
		if err != nil {
			Logger.Println(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
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

// Base handles the default path and template execution
func Base(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r)}
	getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
}

// Campaigns handles the default path and template execution
func Campaigns(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Campaigns", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r)}
	getTemplate(w, "campaigns").ExecuteTemplate(w, "base", params)
}

// CampaignID handles the default path and template execution
func CampaignID(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r)}
	getTemplate(w, "campaign_results").ExecuteTemplate(w, "base", params)
}

// Templates handles the default path and template execution
func Templates(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r)}
	getTemplate(w, "templates").ExecuteTemplate(w, "base", params)
}

// Users handles the default path and template execution
func Users(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r)}
	getTemplate(w, "users").ExecuteTemplate(w, "base", params)
}

// LandingPages handles the default path and template execution
func LandingPages(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r)}
	getTemplate(w, "landing_pages").ExecuteTemplate(w, "base", params)
}

// Settings handles the changing of settings
func Settings(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		params := struct {
			User    models.User
			Title   string
			Flashes []interface{}
			Token   string
		}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: nosurf.Token(r)}
		getTemplate(w, "settings").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		err := auth.ChangePassword(r)
		msg := models.Response{Success: true, Message: "Settings Updated Successfully"}
		if err == auth.ErrInvalidPassword {
			msg.Message = "Invalid Password"
			msg.Success = false
			JSONResponse(w, msg, http.StatusBadRequest)
			return
		} else if err != nil {
			msg.Message = "Unknown Error Occured"
			msg.Success = false
			JSONResponse(w, msg, http.StatusBadRequest)
			return
		}
		JSONResponse(w, msg, http.StatusOK)
	}
}

// Login handles the authentication flow for a user. If credentials are valid,
// a session is created
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

// Logout destroys the current user session
func Logout(w http.ResponseWriter, r *http.Request) {
	// If it is a post request, attempt to register the account
	// Now that we are all registered, we can log the user in
	session := ctx.Get(r, "session").(*sessions.Session)
	delete(session.Values, "id")
	Flash(w, r, "success", "You have successfully logged out")
	http.Redirect(w, r, "/login", 302)
}

// Preview allows for the viewing of page html in a separate browser window
func Preview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	}
	fmt.Fprintf(w, "%s", r.FormValue("html"))
}

// Clone takes a URL as a POST parameter and returns the site HTML
func Clone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	}
	if url, ok := vars["url"]; ok {
		Logger.Println(url)
	}
	http.Error(w, "No URL given.", http.StatusBadRequest)
}

func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template")
	_, err := templates.ParseFiles("templates/base.html", "templates/"+tmpl+".html", "templates/flashes.html")
	if err != nil {
		Logger.Println(err)
	}
	return template.Must(templates, err)
}

// Flash handles the rendering flash messages
func Flash(w http.ResponseWriter, r *http.Request, t string, m string) {
	session := ctx.Get(r, "session").(*sessions.Session)
	session.AddFlash(models.Flash{
		Type:    t,
		Message: m,
	})
	session.Save(r, w)
}
