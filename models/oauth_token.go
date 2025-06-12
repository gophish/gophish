package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// OAuthToken represents an OAuth token for a specific provider tenant
type OAuthToken struct {
	ID                    string    `gorm:"type:text;primary_key"`
	UserID               int64     `gorm:"index:idx_user_provider,unique:user_provider"` // Unique constraint with provider_tenant_id
	ProviderTenantID     string    `gorm:"type:text;index:idx_user_provider,unique:user_provider"` // Unique constraint with user_id
	ProviderType         string    `gorm:"type:text"`
	AuthorizationCode    string    `gorm:"type:text"` // Authorization code from OAuth flow
	AccessTokenEncrypted string    `gorm:"type:text"`
	RefreshTokenEncrypted string   `gorm:"type:text"`
	ExpiresAt            time.Time `gorm:"type:timestamp"`
	CreatedAt            time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt            time.Time `gorm:"type:timestamp"`
}

// TableName specifies the table name for the OAuthToken model
func (OAuthToken) TableName() string {
	return "oauth_tokens"
}

// BeforeCreate will set a UUID rather than numeric ID.
func (t *OAuthToken) BeforeCreate() error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// BeforeUpdate will update the UpdatedAt timestamp
func (t *OAuthToken) BeforeUpdate() error {
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Validate checks if the OAuth token has valid data
func (t *OAuthToken) Validate() error {
	if t.ID == "" {
		return errors.New("token ID cannot be empty")
	}
	if t.ProviderTenantID == "" {
		return errors.New("provider tenant ID cannot be empty")
	}
	if t.UserID == 0 {
		return errors.New("user ID cannot be empty")
	}
	if len(t.AccessTokenEncrypted) == 0 {
		return errors.New("token cannot be empty")
	}
	return nil
}

// Create inserts a new OAuth token into the database
func (t *OAuthToken) Create() error {
	if err := t.Validate(); err != nil {
		return err
	}
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = time.Now().UTC()
	return db.Create(t).Error
}

// Update updates an existing OAuth token in the database
func (t *OAuthToken) Update() error {
	if err := t.Validate(); err != nil {
		return fmt.Errorf("invalid token: %v", err)
	}
	return db.Model(&OAuthToken{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
		"access_token_encrypted": t.AccessTokenEncrypted,
		"refresh_token_encrypted": t.RefreshTokenEncrypted,
		"expires_at": t.ExpiresAt,
		"authorization_code": t.AuthorizationCode,
	}).Error
}

// Delete removes an OAuth token from the database
func (t *OAuthToken) Delete() error {
	// Check if the record exists first
	var count int64
	if err := db.Model(&OAuthToken{}).Where("id = ?", t.ID).Count(&count).Error; err != nil {
		return fmt.Errorf("OAuth token not found: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("OAuth token not found: record does not exist")
	}

	err := db.Delete(t).Error
	if err != nil {
		return fmt.Errorf("failed to delete OAuth token: %v", err)
	}
	return nil
}

// GetOAuthToken retrieves an OAuth token by ID
func GetOAuthToken(id string) (*OAuthToken, error) {
	if id == "" {
		return nil, errors.New("invalid OAuth token ID")
	}
	
	token := &OAuthToken{}
	err := db.Where("id = ?", id).First(token).Error
	if err != nil {
		return nil, fmt.Errorf("OAuth token not found: %v", err)
	}
	return token, nil
}

// GetOAuthTokensByProviderTenant retrieves all OAuth tokens for a given provider tenant
func GetOAuthTokensByProviderTenant(providerTenantID string) ([]*OAuthToken, error) {
	if providerTenantID == "" {
		return nil, errors.New("invalid provider tenant ID")
	}
	
	var tokens []*OAuthToken
	err := db.Where("provider_tenant_id = ?", providerTenantID).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetOAuthTokenByUserAndProviderTenant retrieves an OAuth token for a specific user and provider tenant
func GetOAuthTokenByUserAndProviderTenant(userID int64, providerTenantID string) (*OAuthToken, error) {
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}
	if providerTenantID == "" {
		return nil, errors.New("invalid provider tenant ID")
	}
	
	token := &OAuthToken{}
	err := db.Where("user_id = ? AND provider_tenant_id = ?", userID, providerTenantID).First(token).Error
	if err != nil {
		return nil, fmt.Errorf("OAuth token not found: %v", err)
	}
	return token, nil
}

// IsExpired checks if the token is expired or about to expire (within 5 minutes)
func (t *OAuthToken) IsExpired() bool {
	return time.Now().UTC().Add(5 * time.Minute).After(t.ExpiresAt)
}

// RefreshToken refreshes the OAuth token using the refresh token
func (t *OAuthToken) RefreshToken(config *oauth2.Config) error {
	if t.RefreshTokenEncrypted == "" {
		return errors.New("no refresh token available")
	}

	token := &oauth2.Token{
		RefreshToken: t.RefreshTokenEncrypted,
	}

	// Get new token
	newToken, err := config.TokenSource(context.Background(), token).Token()
	if err != nil {
		return fmt.Errorf("error refreshing token: %v", err)
	}

	// Update token fields
	t.AccessTokenEncrypted = newToken.AccessToken
	if newToken.RefreshToken != "" {
		t.RefreshTokenEncrypted = newToken.RefreshToken
	}
	t.ExpiresAt = newToken.Expiry

	// Save to database
	return t.Update()
}

// GetValidToken returns a valid OAuth2 token, refreshing if necessary
func (t *OAuthToken) GetValidToken(config *oauth2.Config) (*oauth2.Token, error) {
	if t.IsExpired() {
		if err := t.RefreshToken(config); err != nil {
			return nil, fmt.Errorf("error refreshing expired token: %v", err)
		}
	}

	return &oauth2.Token{
		AccessToken:  t.AccessTokenEncrypted,
		RefreshToken: t.RefreshTokenEncrypted,
		Expiry:      t.ExpiresAt,
		TokenType:   "Bearer",
	}, nil
}

// GetTokenForGraph returns a token suitable for Graph API calls using authorization code flow
func (t *OAuthToken) GetTokenForGraph(config *oauth2.Config) (*oauth2.Token, error) {
	// If we have a valid access token, use it
	if !t.IsExpired() {
		return &oauth2.Token{
			AccessToken:  t.AccessTokenEncrypted,
			RefreshToken: t.RefreshTokenEncrypted,
			Expiry:      t.ExpiresAt,
			TokenType:   "Bearer",
		}, nil
	}

	// If we have a refresh token, try to use it
	if t.RefreshTokenEncrypted != "" {
		token, err := t.GetValidToken(config)
		if err == nil {
			return token, nil
		}
		// If refresh fails, fall through to authorization code
	}

	// If we have an authorization code, exchange it for a new token
	if t.AuthorizationCode != "" {
		token, err := config.Exchange(context.Background(), t.AuthorizationCode)
		if err != nil {
			return nil, fmt.Errorf("error exchanging authorization code: %v", err)
		}

		// Update token fields
		t.AccessTokenEncrypted = token.AccessToken
		t.RefreshTokenEncrypted = token.RefreshToken
		t.ExpiresAt = token.Expiry

		// Save to database
		if err := t.Update(); err != nil {
			return nil, fmt.Errorf("error saving new token: %v", err)
		}

		return token, nil
	}

	return nil, errors.New("no valid token or authorization code available")
}