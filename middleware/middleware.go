package middleware

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// CSRFExemptPrefixes are a list of routes that are exempt from CSRF protection
var CSRFExemptPrefixes = []string{
	"/api",
}

// CSRFExceptions is a middleware that prevents CSRF checks on routes listed in
// CSRFExemptPrefixes.
func CSRFExceptions(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, prefix := range CSRFExemptPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				r = csrf.UnsafeSkipCheck(r)
				break
			}
		}
		handler.ServeHTTP(w, r)
	}
}

// Use allows us to stack middleware to process the request
// Example taken from https://github.com/gorilla/mux/pull/36#issuecomment-25849172
func Use(handler http.HandlerFunc, mid ...func(http.Handler) http.HandlerFunc) http.HandlerFunc {
	for _, m := range mid {
		handler = m(handler)
	}
	return handler
}

// init registers the necessary models to be saved in the session later
func init() {
	gob.Register(&models.User{})
	gob.Register(&models.Flash{})
	gob.Register(&oauth2.Token{})
	gob.Register(&models.Tenant{})
	gob.Register(&models.ProviderTenant{})
	Store.Options.HttpOnly = true
	// This sets the maxAge to 1 hour to match Microsoft token expiration
	Store.MaxAge(3600)
	// Set logger to debug level
	log.Logger.SetLevel(logrus.DebugLevel)
}

// Store contains the session information for the request
var Store = sessions.NewCookieStore(
	[]byte(securecookie.GenerateRandomKey(64)), //Signing key
	[]byte(securecookie.GenerateRandomKey(32)))

// GetContext wraps each request in a function which fills in the context for a given request.
// This includes setting the User, Session, Tenant, and ProviderTenants keys and values as necessary for use in later functions.
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
		session, _ := Store.Get(r, "gophish")
		// Put the session in the context so that we can
		// reuse the values in different handlers
		r = ctx.Set(r, "session", session)
		if id, ok := session.Values["id"]; ok {
			u, err := models.GetUser(id.(int64))
			if err != nil {
				r = ctx.Set(r, "user", nil)
				r = ctx.Set(r, "tenant", nil)
				r = ctx.Set(r, "provider_tenants", nil)
				log.Error(err)
			} else {
				r = ctx.Set(r, "user", u)
				
				// Set tenant in context if it exists
				if u.Tenant != nil {
					r = ctx.Set(r, "tenant", u.Tenant)
				}

				// Set provider tenants in context if they exist
				if len(u.ProviderTenants) > 0 {
					r = ctx.Set(r, "provider_tenants", u.ProviderTenants)
				}

			}
		} else {
			r = ctx.Set(r, "user", nil)
			r = ctx.Set(r, "tenant", nil)
			r = ctx.Set(r, "provider_tenants", nil)
		}
		handler.ServeHTTP(w, r)
		// Remove context contents
		ctx.Clear(r)
	}
}

// RequireAPIKey ensures that a valid API key is set as either the api_key GET
// parameter, or a Bearer token.
func RequireAPIKey(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Max-Age", "1000")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			return
		}
		r.ParseForm()
		ak := r.Form.Get("api_key")
		// If we can't get the API key, we'll also check for the
		// Authorization Bearer token
		if ak == "" {
			tokens, ok := r.Header["Authorization"]
			if ok && len(tokens) >= 1 {
				ak = tokens[0]
				ak = strings.TrimPrefix(ak, "Bearer ")
			}
		}
		if ak == "" {
			JSONError(w, http.StatusUnauthorized, "API Key not set")
			return
		}
		// Get user by API key
		u, err := models.GetUserByAPIKey(ak)
		if err != nil {
			JSONError(w, http.StatusUnauthorized, "Invalid API Key")
			return
		}

		// Now get complete user with tenant information
		userWithTenant, err := models.GetUser(u.Id)
		if err != nil {
			log.Warnf("Failed to get user with tenant info: %v", err)
			// Still proceed with the basic user we have
		} else {
			// Use the complete user with tenant information
			u = userWithTenant
			log.Infof("Loaded user %d with tenant ID %s and %d provider tenants", 
				u.Id, u.TenantID, len(u.ProviderTenants))
		}

		r = ctx.Set(r, "user", u)
		r = ctx.Set(r, "user_id", u.Id)
		r = ctx.Set(r, "api_key", ak)

		// Set tenant in context if it exists
		if u.Tenant != nil {
			r = ctx.Set(r, "tenant", u.Tenant)
		}

		// Set provider tenants in context if they exist
		if len(u.ProviderTenants) > 0 {
			r = ctx.Set(r, "provider_tenants", u.ProviderTenants)
		}

		handler.ServeHTTP(w, r)
	})
}

// RequireLogin checks to see if the user is currently logged in.
// If not, the function returns a 302 redirect to the login page.
func RequireLogin(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if u := ctx.Get(r, "user"); u != nil {
			// If a password change is required for the user, then redirect them
			// to the login page
			currentUser := u.(models.User)
			if currentUser.PasswordChangeRequired && r.URL.Path != "/reset_password" {
				q := r.URL.Query()
				q.Set("next", r.URL.Path)
				http.Redirect(w, r, fmt.Sprintf("/reset_password?%s", q.Encode()), http.StatusTemporaryRedirect)
				return
			}

			// User is authenticated, proceed with the request
			handler.ServeHTTP(w, r)
			return
		}

		// User is not authenticated, redirect to login page
		q := r.URL.Query()
		q.Set("next", r.URL.Path)
		http.Redirect(w, r, fmt.Sprintf("/login?%s", q.Encode()), http.StatusTemporaryRedirect)
	}
}

// EnforceViewOnly is a global middleware that limits the ability to edit
// objects to accounts with the PermissionModifyObjects permission.
func EnforceViewOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request is for any non-GET HTTP method, e.g. POST, PUT,
		// or DELETE, we need to ensure the user has the appropriate
		// permission.
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			user := ctx.Get(r, "user").(models.User)
			access, err := user.HasPermission(models.PermissionModifyObjects)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if !access {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks to see if the user has the requested permission
// before executing the handler. If the request is unauthorized, a JSONError
// is returned.
func RequirePermission(perm string) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user := ctx.Get(r, "user").(models.User)
			access, err := user.HasPermission(perm)
			if err != nil {
				JSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if !access {
				JSONError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
				return
			}
			next.ServeHTTP(w, r)
		}
	}
}

// ApplySecurityHeaders applies various security headers according to best-
// practices.
func ApplySecurityHeaders(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csp := "frame-ancestors 'none';"
		w.Header().Set("Content-Security-Policy", csp)
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	}
}

// JSONError returns an error in JSON format with the given
// status code and message
func JSONError(w http.ResponseWriter, c int, m string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(models.Response{Success: false, Message: m})
}

// GetToken retrieves a valid OAuth2 token
func GetToken(userId int64, appRegID string) (*oauth2.Token, error) {
	ctx := context.Background()

	// If no app registration ID is provided, get the default one
	if appRegID == "" {
		defaultAppReg, err := models.GetDefaultAppRegistration()
		if err != nil {
			return nil, fmt.Errorf("failed to get default app registration: %v", err)
		}
		appRegID = defaultAppReg.ID
	}

	token, err := models.GetUserOAuth2Token(ctx, appRegID, userId)
	if err != nil {
		return nil, err
	}

	if token.Expiry.Before(time.Now()) {
		config, err := models.GetOAuth2Config(appRegID)
		if err != nil {
			return nil, err
		}

		// Use the refresh token to get a new access token
		tokenSource := config.TokenSource(ctx, token)
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, err
		}

		// Save the new token
		err = models.SaveOAuth2Token(ctx, appRegID, userId, newToken)
		if err != nil {
			return nil, err
		}

		return newToken, nil
	}

	return token, nil
}
