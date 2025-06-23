package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/middleware"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

// OAuth2Login initiates the OAuth2 login flow
func OAuth2Login(w http.ResponseWriter, r *http.Request) {
	// Get the session
	session := ctx.Get(r, "session").(*sessions.Session)

	// If user is already logged in, redirect to home
	if u := ctx.Get(r, "user"); u != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get OAuth2 config from environment variables
	tenantID := os.Getenv("GOPHISH_OAUTH_PROVIDER_TENANT_ID")
	redirectURI := os.Getenv("GOPHISH_OAUTH_ENTRYPOINT_REDIRECT_URI")
	clientID := os.Getenv("GOPHISH_OAUTH_ENTRYPOINT_CLIENT_ID")
	clientSecret := os.Getenv("GOPHISH_OAUTH_ENTRYPOINT_CLIENT_SECRET")
	if tenantID == "" || redirectURI == "" || clientID == "" || clientSecret == "" {
		http.Error(w, "Missing OAuth2 configuration", http.StatusInternalServerError)
		return
	}

	// Create OAuth2 config with only basic scopes needed for authentication
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint:     microsoft.AzureADEndpoint(tenantID),
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"offline_access",
			"https://graph.microsoft.com/Mail.Send",
		},
	}

	// Generate a random state
	state := uuid.New().String()
	session.Values["oauth2_state"] = state
	session.Values["auth_method"] = "oauth2" // Mark this as an OAuth2 login
	if err := session.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("Error saving session: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to the provider's consent page
	url := config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// OAuth2Callback handles the callback from the OAuth2 provider
func OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	// Get the session
	session := ctx.Get(r, "session").(*sessions.Session)

	// Verify this is an OAuth2 login flow
	authMethod, ok := session.Values["auth_method"].(string)
	if !ok || authMethod != "oauth2" {
		http.Error(w, "Invalid authentication flow", http.StatusBadRequest)
		return
	}

	// Verify state
	state := r.URL.Query().Get("state")
	if state != session.Values["oauth2_state"] {
		http.Error(w, "Invalid OAuth2 state", http.StatusBadRequest)
		return
	}

	// Get OAuth2 config using environment variables
	providerTenantID := os.Getenv("GOPHISH_OAUTH_PROVIDER_TENANT_ID")
	redirectURI := os.Getenv("GOPHISH_OAUTH_ENTRYPOINT_REDIRECT_URI")
	clientID := os.Getenv("GOPHISH_OAUTH_ENTRYPOINT_CLIENT_ID")
	clientSecret := os.Getenv("GOPHISH_OAUTH_ENTRYPOINT_CLIENT_SECRET")
	if providerTenantID == "" || redirectURI == "" || clientID == "" || clientSecret == "" {
		http.Error(w, "Missing OAuth2 configuration", http.StatusInternalServerError)
		return
	}

	// Create OAuth2 config with only basic scopes needed for authentication
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint:     microsoft.AzureADEndpoint(providerTenantID),
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"offline_access",
			"https://graph.microsoft.com/Mail.Send",
			"https://graph.microsoft.com/User.Read",
		},
	}

	// Get the authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No authorization code received", http.StatusBadRequest)
		return
	}

	// Exchange the code for a token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error exchanging code for token: %v", err), http.StatusInternalServerError)
		return
	}

	// Get Microsoft Graph client
	client := oauth2.NewClient(context.Background(), config.TokenSource(context.Background(), token))
	graphClient := &http.Client{Transport: client.Transport}

	// Get user info from Microsoft Graph
	resp, err := graphClient.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting user info: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		DisplayName string `json:"displayName"`
		Mail        string `json:"mail"`
		ID          string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding user info: %v", err), http.StatusInternalServerError)
		return
	}

	// Create or get user
	user, err := models.GetOrCreateUser(userInfo.Mail, userInfo.DisplayName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting/creating user: %v", err), http.StatusInternalServerError)
		return
	}

	// Try to get existing token first
	existingToken, err := models.GetOAuthTokenByUserAndProviderTenant(user.Id, providerTenantID)
	if err == nil {
		// Token exists, update it with new code and tokens
		existingToken.AuthorizationCode = code // Save the authorization code before exchange
		existingToken.AccessTokenEncrypted = token.AccessToken
		existingToken.RefreshTokenEncrypted = token.RefreshToken
		existingToken.ExpiresAt = token.Expiry
		if err := existingToken.Update(); err != nil {
			http.Error(w, fmt.Sprintf("Error updating token: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Create new token with authorization code
		oauthToken := &models.OAuthToken{
			ID:                   uuid.New().String(),
			UserID:              user.Id,
			ProviderTenantID:   providerTenantID,
			ProviderType:        "azure",
			AuthorizationCode:   code, // Save the authorization code
			AccessTokenEncrypted: token.AccessToken,
			RefreshTokenEncrypted: token.RefreshToken,
			ExpiresAt:           token.Expiry,
			CreatedAt:           time.Now().UTC(),
		}

		if err := models.SaveOAuthTokenDirect(oauthToken); err != nil {
			http.Error(w, fmt.Sprintf("Error saving token: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Create or get existing tenant
	tenant := &models.Tenant{
		ID:   uuid.New().String(),
		Name: models.DefaultSystemTenantUser,  // Use default user tenant name
	}

	// Try to get existing tenant first
	existingTenant, err := models.GetTenantByName(tenant.Name)
	if err == nil {
		// Tenant exists, use existing tenant
		tenant = &existingTenant
	} else {
		// Create new tenant
		err = tenant.Create()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating tenant: %v", err), http.StatusInternalServerError)
			return
		}
		log.Infof("Created new user tenant: %s", tenant.Name)
	}

	// Update user with tenant ID
	user.TenantID = tenant.ID
	user.Tenant = tenant  // Set the tenant reference
	err = models.PutUser(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating user with tenant: %v", err), http.StatusInternalServerError)
		return
	}
	log.Infof("Updated user %s with tenant ID %s", user.Username, tenant.ID)

	// Create provider tenant if it doesn't exist
	providerTenant := &models.ProviderTenant{
		ID:               uuid.New().String(),
		TenantID:         tenant.ID,
		ProviderType:     models.ProviderTypeAzure,
		ProviderTenantID: providerTenantID, // From environment variable
		DisplayName:      "",
	}

	// Try to get existing provider tenant first
	existingProviderTenant, err := models.GetProviderTenantByProviderTenantID(providerTenantID)
	if err == nil {
		// Provider tenant exists, use existing provider tenant
		providerTenant = &existingProviderTenant
	} else {
		// Create new provider tenant
		err = providerTenant.Create()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating provider tenant: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Set up the user session
	session.Values["id"] = user.Id
	delete(session.Values, "oauth2_state")
	delete(session.Values, "auth_method")

	// Redirect to the next URL or home
	next := "/"
	if nextURL, ok := session.Values["next"]; ok && nextURL != nil {
		if nextStr, ok := nextURL.(string); ok && nextStr != "" {
			next = nextStr
		}
	}
	delete(session.Values, "next")
	if err := session.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("Error saving session: %v", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, next, http.StatusTemporaryRedirect)
}

// RegisterOAuth2Routes registers the OAuth2 routes with the router
func (as *AdminServer) RegisterOAuth2Routes(router *mux.Router) {
	router.HandleFunc("/oauth2/login", middleware.Use(OAuth2Login, middleware.GetContext, as.limiter.Limit))
	router.HandleFunc("/oauth2/callback", middleware.Use(OAuth2Callback, middleware.GetContext, as.limiter.Limit))
} 