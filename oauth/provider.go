package oauth

import (
	"context"
	"fmt"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gophish/gophish/config"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"golang.org/x/oauth2"
)

// OIDCProvider implements the Provider interface for OpenID Connect providers.
// This implementation supports any OIDC-compliant provider (Google, Azure AD,
// Okta, Keycloak, etc.) through OIDC Discovery.
type OIDCProvider struct {
	name        string
	displayName string
	config      *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	provider    *oidc.Provider
}

// NewOIDCProvider creates a new OIDC provider from configuration.
// It performs OIDC Discovery to fetch the provider's configuration.
func NewOIDCProvider(ctx context.Context, cfg config.OAuthProviderConfig, callbackURL string) (*OIDCProvider, error) {
	// Perform OIDC Discovery
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC provider %s: %w", cfg.Name, err)
	}

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  callbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       cfg.Scopes,
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	return &OIDCProvider{
		name:        cfg.Name,
		displayName: cfg.DisplayName,
		config:      oauth2Config,
		verifier:    verifier,
		provider:    provider,
	}, nil
}

// Name returns the provider identifier
func (p *OIDCProvider) Name() string {
	return p.name
}

// DisplayName returns the human-readable provider name
func (p *OIDCProvider) DisplayName() string {
	return p.displayName
}

// AuthCodeURL generates the authorization URL with the given state parameter
func (p *OIDCProvider) AuthCodeURL(state string) string {
	// Use offline_access to request refresh token if needed
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange converts an authorization code into tokens
func (p *OIDCProvider) Exchange(ctx context.Context, code string) (*Token, error) {
	oauth2Token, err := p.config.Exchange(ctx, code)
	if err != nil {
		log.Errorf("OAuth token exchange failed for provider %s: %v", p.name, err)
		return nil, fmt.Errorf("%w: %v", ErrTokenExchange, err)
	}

	return FromOAuth2Token(oauth2Token), nil
}

// VerifyIDToken verifies and parses an OIDC ID token
func (p *OIDCProvider) VerifyIDToken(ctx context.Context, rawIDToken string) (*IDTokenClaims, error) {
	// Verify the ID token
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Errorf("ID token verification failed for provider %s: %v", p.name, err)
		return nil, fmt.Errorf("%w: %v", ErrTokenVerification, err)
	}

	// Extract standard claims
	var claims struct {
		Subject           string `json:"sub"`
		Email             string `json:"email"`
		EmailVerified     bool   `json:"email_verified"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}

	if err := idToken.Claims(&claims); err != nil {
		log.Errorf("Failed to parse ID token claims for provider %s: %v", p.name, err)
		return nil, fmt.Errorf("%w: %v", ErrMissingClaims, err)
	}

	// Build our claims structure
	tokenClaims := &IDTokenClaims{
		Subject:           claims.Subject,
		Email:             claims.Email,
		EmailVerified:     claims.EmailVerified,
		Name:              claims.Name,
		PreferredUsername: claims.PreferredUsername,
		Issuer:            idToken.Issuer,
		Audience:          idToken.Audience,
		IssuedAt:          idToken.IssuedAt,
		Expiry:            idToken.Expiry,
	}

	// Validate required claims
	if tokenClaims.Subject == "" {
		log.Errorf("ID token missing 'sub' claim for provider %s", p.name)
		return nil, ErrMissingClaims
	}

	return tokenClaims, nil
}

// RefreshToken exchanges a refresh token for a new access token
func (p *OIDCProvider) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	tokenSource := p.config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	newToken, err := tokenSource.Token()
	if err != nil {
		log.Errorf("Token refresh failed for provider %s: %v", p.name, err)
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return FromOAuth2Token(newToken), nil
}

// ProviderRegistry manages configured OAuth providers
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	enabled   bool
}

// NewProviderRegistry creates a registry from the OAuth configuration.
// It will first try to load providers from the database. If no providers are found
// in the database and the config has providers defined, it will use those as fallback.
func NewProviderRegistry(ctx context.Context, cfg *config.OAuthConfig) (*ProviderRegistry, error) {
	if cfg == nil {
		return &ProviderRegistry{
			providers: make(map[string]Provider),
			enabled:   false,
		}, nil
	}

	registry := &ProviderRegistry{
		providers: make(map[string]Provider),
		enabled:   cfg.Enabled || hasProvidersInDB(),
	}

	if !registry.enabled {
		return registry, nil
	}

	// Try to load providers from database first
	dbProviders, err := getDBProviders()
	if err == nil && len(dbProviders) > 0 {
		// Use database providers
		for _, providerCfg := range dbProviders {
			provider, err := NewOIDCProvider(ctx, providerCfg, cfg.CallbackURL)
			if err != nil {
				log.Warnf("Failed to initialize OAuth provider %s: %v", providerCfg.Name, err)
				continue
			}
			registry.providers[providerCfg.Name] = provider
			log.Infof("Initialized OAuth provider: %s (from database)", providerCfg.DisplayName)
		}
	} else {
		// Fallback to config file providers
		for _, providerCfg := range cfg.Providers {
			if !providerCfg.Enabled {
				continue
			}

			provider, err := NewOIDCProvider(ctx, providerCfg, cfg.CallbackURL)
			if err != nil {
				log.Warnf("Failed to initialize OAuth provider %s: %v", providerCfg.Name, err)
				continue
			}

			registry.providers[providerCfg.Name] = provider
			log.Infof("Initialized OAuth provider: %s (from config)", providerCfg.DisplayName)
		}
	}

	return registry, nil
}

// GetProvider retrieves a provider by name
func (r *ProviderRegistry) GetProvider(name string) (Provider, error) {
	if !r.enabled {
		return nil, ErrOAuthDisabled
	}

	r.mu.RLock()
	provider, ok := r.providers[name]
	r.mu.RUnlock()

	if !ok {
		return nil, ErrProviderNotFound
	}

	return provider, nil
}

// ListProviders returns all available providers
func (r *ProviderRegistry) ListProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}

// IsEnabled returns whether OAuth is enabled
func (r *ProviderRegistry) IsEnabled() bool {
	return r.enabled
}

// Reload refreshes the provider registry from the database
func (r *ProviderRegistry) Reload(ctx context.Context, callbackURL string) error {
	dbProviders, err := getDBProviders()
	if err != nil {
		return fmt.Errorf("failed to load providers from database: %w", err)
	}

	// Build new providers map before replacing
	newProviders := make(map[string]Provider)
	for _, providerCfg := range dbProviders {
		provider, err := NewOIDCProvider(ctx, providerCfg, callbackURL)
		if err != nil {
			log.Warnf("Failed to initialize OAuth provider %s: %v", providerCfg.Name, err)
			continue
		}
		newProviders[providerCfg.Name] = provider
		log.Infof("Reloaded OAuth provider: %s", providerCfg.DisplayName)
	}

	// Replace providers atomically
	r.mu.Lock()
	r.providers = newProviders
	r.mu.Unlock()

	return nil
}

// hasProvidersInDB checks if there are any providers in the database
func hasProvidersInDB() bool {
	providers, err := models.GetEnabledOAuthProviders()
	return err == nil && len(providers) > 0
}

// getDBProviders loads OAuth providers from the database and converts them to config format
func getDBProviders() ([]config.OAuthProviderConfig, error) {
	dbProviders, err := models.GetEnabledOAuthProviders()
	if err != nil {
		return nil, err
	}

	configs := make([]config.OAuthProviderConfig, 0, len(dbProviders))
	for _, p := range dbProviders {
		configs = append(configs, config.OAuthProviderConfig{
			Name:         p.Name,
			DisplayName:  p.DisplayName,
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			IssuerURL:    p.IssuerURL,
			Scopes:       p.ScopesList,
			Enabled:      p.Enabled,
		})
	}

	return configs, nil
}
