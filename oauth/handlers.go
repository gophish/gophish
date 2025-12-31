package oauth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/sessions"
)

// Handler manages OAuth HTTP handlers with access to the provider registry
// and state manager.
type Handler struct {
	registry     *ProviderRegistry
	stateManager *StateManager
}

// NewHandler creates a new OAuth handler
func NewHandler(registry *ProviderRegistry, stateManager *StateManager) *Handler {
	return &Handler{
		registry:     registry,
		stateManager: stateManager,
	}
}

// Login initiates the OAuth login flow.
// GET /oauth/login?provider=<name>&next=<url>
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.registry.IsEnabled() {
		log.Warn("OAuth login attempted but OAuth is disabled")
		http.Error(w, "OAuth is not enabled", http.StatusNotFound)
		return
	}

	// Get provider name from query parameter
	providerName := r.URL.Query().Get("provider")
	if providerName == "" {
		log.Warn("OAuth login attempted without provider parameter")
		http.Error(w, "Missing provider parameter", http.StatusBadRequest)
		return
	}

	// Get the provider
	provider, err := h.registry.GetProvider(providerName)
	if err != nil {
		log.Warnf("OAuth login attempted with unknown provider: %s", providerName)
		http.Error(w, "Unknown provider", http.StatusNotFound)
		return
	}

	// Get optional redirect URL and validate to prevent open redirect
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	} else {
		// Only allow relative URLs to prevent open redirect attacks
		if !strings.HasPrefix(next, "/") || strings.Contains(next, "://") {
			log.Warnf("OAuth login attempted with invalid redirect URL: %s", next)
			next = "/"
		}
	}

	// Generate state token
	state, err := h.stateManager.GenerateState(providerName, next)
	if err != nil {
		log.Errorf("Failed to generate OAuth state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get authorization URL
	authURL := provider.AuthCodeURL(state)

	log.Infof("OAuth login initiated for provider %s (%s)", provider.DisplayName(), providerName)

	// Redirect to provider's authorization URL
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth callback from the provider.
// GET /oauth/callback?code=<code>&state=<state>
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	if !h.registry.IsEnabled() {
		log.Warn("OAuth callback received but OAuth is disabled")
		http.Error(w, "OAuth is not enabled", http.StatusNotFound)
		return
	}

	// Extract authorization code and state
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		log.Warn("OAuth callback received without authorization code")
		h.handleCallbackError(w, r, "Missing authorization code")
		return
	}

	if state == "" {
		log.Warn("OAuth callback received without state parameter")
		h.handleCallbackError(w, r, "Missing state parameter")
		return
	}

	// Validate state token
	providerName, redirectTo, err := h.stateManager.ValidateState(state)
	if err != nil {
		log.Warnf("OAuth callback received with invalid state: %v", err)
		h.handleCallbackError(w, r, "Invalid or expired state parameter")
		return
	}

	// Get the provider
	provider, err := h.registry.GetProvider(providerName)
	if err != nil {
		log.Errorf("OAuth callback for unknown provider: %s", providerName)
		h.handleCallbackError(w, r, "Unknown provider")
		return
	}

	// Exchange authorization code for tokens
	token, err := provider.Exchange(r.Context(), code)
	if err != nil {
		log.Errorf("OAuth token exchange failed: %v", err)
		h.handleCallbackError(w, r, "Failed to exchange authorization code")
		return
	}

	// Verify ID token (OIDC)
	if token.IDToken == "" {
		log.Errorf("OAuth callback for provider %s did not include ID token", providerName)
		h.handleCallbackError(w, r, "Provider did not return ID token")
		return
	}

	claims, err := provider.VerifyIDToken(r.Context(), token.IDToken)
	if err != nil {
		log.Errorf("OAuth ID token verification failed: %v", err)
		h.handleCallbackError(w, r, "Failed to verify ID token")
		return
	}

	// Find or create user
	user, err := h.findOrCreateUser(providerName, claims)
	if err != nil {
		log.Errorf("Failed to find or create user from OAuth: %v", err)
		h.handleCallbackError(w, r, "Failed to process user account")
		return
	}

	// Check if account is locked
	if user.AccountLocked {
		log.Warnf("OAuth login attempted for locked account: %s", user.Username)
		h.handleCallbackError(w, r, "Account is locked")
		return
	}

	// Update last login time
	user.LastLogin = time.Now().UTC()
	if err := models.PutUser(&user); err != nil {
		log.Errorf("Failed to update user last login: %v", err)
	}

	// Create session
	session := ctx.Get(r, "session").(*sessions.Session)
	session.Values["id"] = user.Id
	if err := session.Save(r, w); err != nil {
		log.Errorf("Failed to save session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	log.Infof("OAuth login successful for user %s via provider %s", user.Username, providerName)

	// Redirect to the original destination
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// findOrCreateUser finds an existing user or creates a new one based on OAuth claims
func (h *Handler) findOrCreateUser(providerName string, claims *IDTokenClaims) (models.User, error) {
	// Try to find existing user by OAuth subject
	user, err := models.GetUserByOAuthSubject(providerName, claims.Subject)
	if err == nil {
		// User found
		return user, nil
	}

	// If user not found by OAuth subject, try to find by email (if available and verified)
	if claims.Email != "" && claims.EmailVerified {
		user, err := models.GetUserByUsername(claims.Email)
		if err == nil {
			// Link this OAuth identity to existing user
			user.OAuthProvider = providerName
			user.OAuthSubject = claims.Subject
			if err := models.PutUser(&user); err != nil {
				return models.User{}, fmt.Errorf("failed to link OAuth identity: %w", err)
			}
			log.Infof("Linked OAuth identity %s:%s to existing user %s", providerName, claims.Subject, user.Username)
			return user, nil
		}
	}

	// Create new user
	username := claims.GetPreferredUsername()
	if username == "" {
		username = fmt.Sprintf("%s_%s", providerName, claims.Subject)
	}

	// Ensure username is unique
	originalUsername := username
	for i := 1; i < 100; i++ {
		_, err := models.GetUserByUsername(username)
		if err != nil {
			// Username is available
			break
		}
		// Username taken, try with suffix
		username = fmt.Sprintf("%s_%d", originalUsername, i)
	}

	// Get the default user role
	defaultRole, err := models.GetRoleBySlug(models.RoleUser)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get default role: %w", err)
	}

	newUser := models.User{
		Username:      username,
		Hash:          "", // No password for OAuth users
		ApiKey:        "", // Will be generated by models.PutUser
		RoleID:        defaultRole.ID,
		OAuthProvider: providerName,
		OAuthSubject:  claims.Subject,
	}

	// Generate API key
	if err := models.PostUser(&newUser); err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	log.Infof("Created new user %s from OAuth provider %s", newUser.Username, providerName)
	return newUser, nil
}

// handleCallbackError redirects to login page with an error message
func (h *Handler) handleCallbackError(w http.ResponseWriter, r *http.Request, message string) {
	session := ctx.Get(r, "session").(*sessions.Session)
	session.AddFlash(models.Flash{
		Type:    "danger",
		Message: message,
	})
	if err := session.Save(r, w); err != nil {
		log.Errorf("Failed to save session with error flash: %v", err)
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
