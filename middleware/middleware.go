package middleware

import (
	"net/http"

	ctx "github.com/gorilla/context"
	"github.com/jordan-wright/gophish/auth"
	"github.com/jordan-wright/gophish/models"
)

// GetContext wraps each request in a function which fills in the context for a given request.
// This includes setting the User and Session keys and values as necessary for use in later functions.
func GetContext(handler http.Handler) http.HandlerFunc {
	// Set the context here
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request form
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing request", http.StatusInternalServerError)
		}
		// Set the context appropriately here.
		// Set the session
		session, _ := auth.Store.Get(r, "gophish")
		// Put the session in the context so that
		ctx.Set(r, "session", session)
		if id, ok := session.Values["id"]; ok {
			u, err := models.GetUser(id.(int64))
			if err != nil {
				ctx.Set(r, "user", nil)
			}
			ctx.Set(r, "user", u)
		} else {
			ctx.Set(r, "user", nil)
		}
		handler.ServeHTTP(w, r)
		// Remove context contents
		ctx.Clear(r)
	}
}

func RequireAPIKey(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		ak := r.Form.Get("api_key")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "1000")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			return
		}
		if ak == "" {
			JSONError(w, 400, "API Key not set")
		} else {
			u, err := models.GetUserByAPIKey(ak)
			/*			id, err := models.Conn.SelectInt("SELECT id FROM users WHERE api_key=?", ak)
			 */if err != nil {
				JSONError(w, 400, "Invalid API Key")
				return
			}
			ctx.Set(r, "user_id", u.Id)
			ctx.Set(r, "api_key", ak)
			handler.ServeHTTP(w, r)
		}
	}
}

// RequireLogin is a simple middleware which checks to see if the user is currently logged in.
// If not, the function returns a 302 redirect to the login page.
func RequireLogin(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if u := ctx.Get(r, "user"); u != nil {
			handler.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/login", 302)
		}
	}
}

func JSONError(w http.ResponseWriter, c int, m string) {
	http.Error(w, m, c)
}
