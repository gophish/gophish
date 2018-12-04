package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/binodlamsal/gophish/auth"
	"github.com/binodlamsal/gophish/config"
	ctx "github.com/binodlamsal/gophish/context"
	log "github.com/binodlamsal/gophish/logger"
	mid "github.com/binodlamsal/gophish/middleware"
	"github.com/binodlamsal/gophish/models"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

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
	router.HandleFunc("/sending_profiles", Use(SendingProfiles, mid.RequireLogin))
	router.HandleFunc("/our_domains", Use(SendingDomains, mid.RequireLogin))
	router.HandleFunc("/register", Use(Register, mid.RequireLogin))
	router.HandleFunc("/settings", Use(Settings, mid.RequireLogin))
	router.HandleFunc("/people", Use(People, mid.RequireLogin))
	router.HandleFunc("/roles", Use(Roles, mid.RequireLogin))
	// Create the API routes
	api := router.PathPrefix("/api").Subrouter()
	api = api.StrictSlash(true)
	api.HandleFunc("/reset", Use(API_Reset, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/", Use(API_Campaigns, mid.RequireAPIKey))
	api.HandleFunc("/people", Use(API_Users, mid.RequireAPIKey))
	api.HandleFunc("/people/partner", Use(API_User_Partners, mid.RequireAPIKey))
	api.HandleFunc("/roles", Use(API_Roles, mid.RequireAPIKey))
	api.HandleFunc("/roles/{id:[0-9]+}", Use(API_Roles_Id, mid.RequireAPIKey))
	api.HandleFunc("/people/{id:[0-9]+}", Use(API_Users_Id, mid.RequireAPIKey))
	api.HandleFunc("/tags", Use(API_Tags, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/summary", Use(API_Campaigns_Summary, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}", Use(API_Campaigns_Id, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}/results", Use(API_Campaigns_Id_Results, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}/summary", Use(API_Campaign_Id_Summary, mid.RequireAPIKey))
	api.HandleFunc("/campaigns/{id:[0-9]+}/complete", Use(API_Campaigns_Id_Complete, mid.RequireAPIKey))
	api.HandleFunc("/groups/", Use(API_Groups, mid.RequireAPIKey))
	api.HandleFunc("/groups/summary", Use(API_Groups_Summary, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}", Use(API_Groups_Id, mid.RequireAPIKey))
	api.HandleFunc("/groups/{id:[0-9]+}/summary", Use(API_Groups_Id_Summary, mid.RequireAPIKey))
	api.HandleFunc("/templates/", Use(API_Templates, mid.RequireAPIKey))
	api.HandleFunc("/templates/{id:[0-9]+}", Use(API_Templates_Id, mid.RequireAPIKey))
	api.HandleFunc("/pages/", Use(API_Pages, mid.RequireAPIKey))
	api.HandleFunc("/pages/{id:[0-9]+}", Use(API_Pages_Id, mid.RequireAPIKey))
	api.HandleFunc("/smtp/", Use(API_SMTP, mid.RequireAPIKey))
	api.HandleFunc("/sendingdomains", Use(API_SMTP_domains, mid.RequireAPIKey))
	api.HandleFunc("/smtp/{id:[0-9]+}", Use(API_SMTP_Id, mid.RequireAPIKey))
	api.HandleFunc("/util/send_test_email", Use(API_Send_Test_Email, mid.RequireAPIKey))
	api.HandleFunc("/import/group", Use(API_Import_Group, mid.RequireAPIKey))
	api.HandleFunc("/import/email", Use(API_Import_Email, mid.RequireAPIKey))
	api.HandleFunc("/import/site", Use(API_Import_Site, mid.RequireAPIKey))

	// Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(UnindexedFileSystem{http.Dir("./static/")}))

	// Setup CSRF Protection
	csrfHandler := csrf.Protect([]byte(auth.GenerateSecureKey()),
		csrf.FieldName("csrf_token"),
		csrf.Secure(config.Conf.AdminConf.UseTLS))
	csrfRouter := csrfHandler(router)
	return Use(csrfRouter.ServeHTTP, mid.CSRFExceptions, mid.GetContext)
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
		Roles   []models.Roles
		Token   string
	}{Title: "Register", Token: csrf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		rs, err := models.GetRoles(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		params.Flashes = session.Flashes()
		params.Roles = rs
		session.Save(r, w)
		templates := template.New("template")

		_, errs := templates.ParseFiles("templates/register.html", "templates/flashes.html")
		if errs != nil {
			log.Error(errs)
		}
		template.Must(templates, errs).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to register
		succ, err := auth.Register(r)
		//If we've registered, redirect to the login page
		if succ {
			Flash(w, r, "success", "Registration successful!")
			session.Save(r, w)
			http.Redirect(w, r, "/login", 302)
			return
		}
		// Check the error
		m := err.Error()
		log.Error(err)
		Flash(w, r, "danger", m)
		session.Save(r, w)
		http.Redirect(w, r, "/register", 302)
		return
	}
}

// Base handles the default path and template execution
func Base(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Dashboard", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
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
	}{Title: "Campaigns", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "campaigns").ExecuteTemplate(w, "base", params)
}

// People handles the default path and template execution
func People(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "People", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "people").ExecuteTemplate(w, "base", params)
}

// Roles handles the default path and template execution
func Roles(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Roles", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "roles").ExecuteTemplate(w, "base", params)
}

// CampaignID handles the default path and template execution
func CampaignID(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Campaign Results", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
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
	}{Title: "Email Templates", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
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
	}{Title: "Users & Groups", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
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
	}{Title: "Landing Pages", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "landing_pages").ExecuteTemplate(w, "base", params)
}

// SendingProfiles handles the default path and template execution
func SendingProfiles(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Sending Profiles", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "sending_profiles").ExecuteTemplate(w, "base", params)
}

// Replancememnt of SendingProfiles by sendingdomains in our application a nornal user can use the profile/domains created by
// the administrator handles the default path and template execution
func SendingDomains(w http.ResponseWriter, r *http.Request) {
	// Example of using session - will be removed.
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Sending Domains", User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
	getTemplate(w, "sending_domains").ExecuteTemplate(w, "base", params)
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
			Version string
		}{Title: "Settings", Version: config.Version, User: ctx.Get(r, "user").(models.User), Token: csrf.Token(r)}
		getTemplate(w, "settings").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		err := auth.ChangePassword(r)
		msg := models.Response{Success: true, Message: "Settings Updated Successfully"}
		if err == auth.ErrInvalidPassword {
			msg.Message = "Invalid Password"
			msg.Success = false
			JSONResponse(w, msg, http.StatusBadRequest)
			return
		}
		if err != nil {
			msg.Message = err.Error()
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
	}{Title: "Login", Token: csrf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)
	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")
		if err != nil {
			log.Error(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		//Attempt to login
		succ, u, err := auth.Login(r)
		if err != nil {
			log.Error(err)
		}
		//If we've logged in, save the session and redirect to the dashboard
		if succ {
			session.Values["id"] = u.Id
			session.Save(r, w)
			next := "/"
			url, err := url.Parse(r.FormValue("next"))
			if err == nil {
				path := url.Path
				if path != "" {
					next = path
				}
			}
			http.Redirect(w, r, next, 302)
		} else {
			Flash(w, r, "danger", "Invalid Username/Password")
			params.Flashes = session.Flashes()
			session.Save(r, w)
			templates := template.New("template")
			_, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")
			if err != nil {
				log.Error(err)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			template.Must(templates, err).ExecuteTemplate(w, "base", params)
		}
	}
}

// Logout destroys the current user session
func Logout(w http.ResponseWriter, r *http.Request) {
	session := ctx.Get(r, "session").(*sessions.Session)
	delete(session.Values, "id")
	Flash(w, r, "success", "You have successfully logged out")
	session.Save(r, w)
	http.Redirect(w, r, "/login", 302)
}

// Preview allows for the viewing of page html in a separate browser window
func Preview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%s", r.FormValue("html"))
}

// Clone takes a URL as a POST parameter and returns the site HTML
func Clone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	if url, ok := vars["url"]; ok {
		log.Error(url)
	}
	http.Error(w, "No URL given.", http.StatusBadRequest)
}

func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template")
	_, err := templates.ParseFiles("templates/base.html", "templates/"+tmpl+".html", "templates/flashes.html")
	if err != nil {
		log.Error(err)
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
}
