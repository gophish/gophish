package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

// isValidProviderName checks if provider name contains only allowed characters
func isValidProviderName(name string) bool {
	matched, _ := regexp.MatchString("^[a-z0-9_-]+$", name)
	return matched
}

// OAuthProviders handles GET (list all) and POST (create new) for OAuth providers
func (as *Server) OAuthProviders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		providers, err := models.GetOAuthProviders()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		// Sanitize secrets before returning
		for i := range providers {
			providers[i].Sanitize()
		}
		JSONResponse(w, providers, http.StatusOK)

	case "POST":
		p := models.OAuthProvider{}
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if p.Name == "" || p.DisplayName == "" || p.ClientID == "" || p.ClientSecret == "" || p.IssuerURL == "" {
			JSONResponse(w, models.Response{Success: false, Message: "Missing required fields"}, http.StatusBadRequest)
			return
		}

		// Validate provider name format (lowercase alphanumeric, hyphens, underscores only)
		if !isValidProviderName(p.Name) {
			JSONResponse(w, models.Response{Success: false, Message: "Provider name must contain only lowercase letters, numbers, hyphens, and underscores"}, http.StatusBadRequest)
			return
		}

		// Validate issuer URL uses HTTPS
		if !strings.HasPrefix(p.IssuerURL, "https://") {
			JSONResponse(w, models.Response{Success: false, Message: "Issuer URL must use HTTPS"}, http.StatusBadRequest)
			return
		}

		// Default scopes if not provided
		if len(p.ScopesList) == 0 {
			p.ScopesList = []string{"openid", "profile", "email"}
		}

		err = models.PostOAuthProvider(&p)
		if err != nil {
			if err == models.ErrOAuthProviderNameExists {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusConflict)
				return
			}
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: "Error creating OAuth provider"}, http.StatusInternalServerError)
			return
		}

		// Trigger OAuth provider reload
		as.reloadOAuthProviders()

		// Sanitize before returning
		p.Sanitize()
		JSONResponse(w, p, http.StatusCreated)
	}
}

// OAuthProvider handles GET (fetch), PUT (update), and DELETE for a single OAuth provider
func (as *Server) OAuthProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 0, 64)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid ID"}, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		p, err := models.GetOAuthProvider(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "OAuth provider not found"}, http.StatusNotFound)
			return
		}
		p.Sanitize()
		JSONResponse(w, p, http.StatusOK)

	case "PUT":
		p, err := models.GetOAuthProvider(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "OAuth provider not found"}, http.StatusNotFound)
			return
		}

		// Decode the update
		updateData := models.OAuthProvider{}
		err = json.NewDecoder(r.Body).Decode(&updateData)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}

		// Validate and update fields
		if updateData.Name != "" {
			if !isValidProviderName(updateData.Name) {
				JSONResponse(w, models.Response{Success: false, Message: "Provider name must contain only lowercase letters, numbers, hyphens, and underscores"}, http.StatusBadRequest)
				return
			}
			p.Name = updateData.Name
		}
		if updateData.DisplayName != "" {
			p.DisplayName = updateData.DisplayName
		}
		if updateData.ClientID != "" {
			p.ClientID = updateData.ClientID
		}
		// Only update client secret if provided and not the masked placeholder
		if updateData.ClientSecret != "" && updateData.ClientSecret != models.MaskedSecret {
			p.ClientSecret = updateData.ClientSecret
		}
		if updateData.IssuerURL != "" {
			// Validate issuer URL uses HTTPS
			if !strings.HasPrefix(updateData.IssuerURL, "https://") {
				JSONResponse(w, models.Response{Success: false, Message: "Issuer URL must use HTTPS"}, http.StatusBadRequest)
				return
			}
			p.IssuerURL = updateData.IssuerURL
		}
		if len(updateData.ScopesList) > 0 {
			p.ScopesList = updateData.ScopesList
		}
		p.Enabled = updateData.Enabled

		err = models.PutOAuthProvider(&p)
		if err != nil {
			if err == models.ErrOAuthProviderNameExists {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusConflict)
				return
			}
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: "Error updating OAuth provider"}, http.StatusInternalServerError)
			return
		}

		// Trigger OAuth provider reload
		as.reloadOAuthProviders()

		p.Sanitize()
		JSONResponse(w, p, http.StatusOK)

	case "DELETE":
		err := models.DeleteOAuthProvider(id)
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting OAuth provider"}, http.StatusInternalServerError)
			return
		}

		// Trigger OAuth provider reload
		as.reloadOAuthProviders()

		JSONResponse(w, models.Response{Success: true, Message: "OAuth provider deleted successfully"}, http.StatusOK)
	}
}

// OAuthProvidersReload triggers a reload of OAuth providers without restarting the server
func (as *Server) OAuthProvidersReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := as.reloadOAuthProviders()
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}

	JSONResponse(w, models.Response{Success: true, Message: "OAuth providers reloaded successfully"}, http.StatusOK)
}

// reloadOAuthProviders is a helper that triggers OAuth provider reload in the admin server
func (as *Server) reloadOAuthProviders() error {
	// This will be called through a channel or callback to the admin server
	if as.reloadOAuthCallback != nil {
		return as.reloadOAuthCallback()
	}
	return nil
}
