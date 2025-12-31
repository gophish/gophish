package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gophish/gophish/crypto"
	"github.com/jinzhu/gorm"
)

// OAuthProvider represents an OAuth/OIDC provider configuration stored in the database
type OAuthProvider struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name" sql:"not null;unique"`                    // Provider identifier (e.g., "google", "azure")
	DisplayName  string    `json:"display_name" sql:"not null"`                   // Human-readable name for UI
	ClientID     string    `json:"client_id" sql:"not null"`                      // OAuth client ID
	ClientSecret string    `json:"client_secret,omitempty" sql:"not null"`        // OAuth client secret (not returned in JSON by default)
	IssuerURL    string    `json:"issuer_url" sql:"not null"`                     // OIDC issuer URL
	Scopes       string    `json:"-" sql:"type:text"`                             // JSON array of scopes (internal storage)
	ScopesList   []string  `json:"scopes" gorm:"-"`                               // Scopes as array (for API)
	Enabled      bool      `json:"enabled" sql:"not null;default:true"`           // Whether this provider is active
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

// TableName specifies the table name for GORM
func (OAuthProvider) TableName() string {
	return "oauth_providers"
}

// ErrOAuthProviderNameExists is returned when attempting to create a provider with a duplicate name
var ErrOAuthProviderNameExists = errors.New("OAuth provider with this name already exists")

// BeforeSave is a GORM hook that runs before saving to convert ScopesList to JSON
func (p *OAuthProvider) BeforeSave(scope *gorm.DB) error {
	// Encrypt client secret if present and not masked
	if p.ClientSecret != "" && p.ClientSecret != MaskedSecret {
		encrypted, err := crypto.EncryptSecret(p.ClientSecret)
		if err != nil {
			return err
		}
		p.ClientSecret = encrypted
	}

	if len(p.ScopesList) > 0 {
		scopesJSON, err := json.Marshal(p.ScopesList)
		if err != nil {
			return err
		}
		p.Scopes = string(scopesJSON)
	} else if p.Scopes == "" {
		// Set default scopes if none provided
		p.ScopesList = []string{"openid", "profile", "email"}
		scopesJSON, _ := json.Marshal(p.ScopesList)
		p.Scopes = string(scopesJSON)
	}
	p.ModifiedAt = time.Now().UTC()
	return nil
}

// AfterFind is a GORM hook that runs after loading from DB to parse Scopes JSON
func (p *OAuthProvider) AfterFind(scope *gorm.DB) error {
	// Decrypt client secret
	if p.ClientSecret != "" {
		decrypted, err := crypto.DecryptSecret(p.ClientSecret)
		if err != nil {
			return err
		}
		p.ClientSecret = decrypted
	}

	if p.Scopes != "" {
		return json.Unmarshal([]byte(p.Scopes), &p.ScopesList)
	}
	p.ScopesList = []string{}
	return nil
}

// GetOAuthProviders returns all OAuth providers
func GetOAuthProviders() ([]OAuthProvider, error) {
	providers := []OAuthProvider{}
	err := db.Find(&providers).Error
	return providers, err
}

// GetOAuthProvider returns an OAuth provider by ID
func GetOAuthProvider(id int64) (OAuthProvider, error) {
	p := OAuthProvider{}
	err := db.Where("id = ?", id).First(&p).Error
	return p, err
}

// GetOAuthProviderByName returns an OAuth provider by name
func GetOAuthProviderByName(name string) (OAuthProvider, error) {
	p := OAuthProvider{}
	err := db.Where("name = ?", name).First(&p).Error
	return p, err
}

// GetEnabledOAuthProviders returns only enabled OAuth providers
func GetEnabledOAuthProviders() ([]OAuthProvider, error) {
	providers := []OAuthProvider{}
	err := db.Where("enabled = ?", true).Find(&providers).Error
	return providers, err
}

// PostOAuthProvider creates a new OAuth provider
func PostOAuthProvider(p *OAuthProvider) error {
	// Check if provider with this name already exists
	_, err := GetOAuthProviderByName(p.Name)
	if err == nil {
		return ErrOAuthProviderNameExists
	}

	p.CreatedAt = time.Now().UTC()
	p.ModifiedAt = time.Now().UTC()

	err = db.Create(p).Error
	return err
}

// PutOAuthProvider updates an existing OAuth provider
func PutOAuthProvider(p *OAuthProvider) error {
	// Fetch the existing provider to preserve CreatedAt
	original, err := GetOAuthProvider(p.Id)
	if err != nil {
		return err
	}

	// Check if another provider has the same name
	if p.Name != original.Name {
		existing, err := GetOAuthProviderByName(p.Name)
		if err == nil && existing.Id != p.Id {
			return ErrOAuthProviderNameExists
		} else if err != nil && err.Error() != "record not found" {
			// Return database errors, but ignore "not found"
			return err
		}
	}

	// Preserve CreatedAt
	p.CreatedAt = original.CreatedAt

	err = db.Save(p).Error
	return err
}

// DeleteOAuthProvider deletes an OAuth provider by ID
func DeleteOAuthProvider(id int64) error {
	return db.Where("id = ?", id).Delete(&OAuthProvider{}).Error
}

const (
	// MaskedSecret is the placeholder shown for masked client secrets
	MaskedSecret = "__MASKED__"
)

// Sanitize removes sensitive data (client secret) before returning to API
func (p *OAuthProvider) Sanitize() {
	// Mask client secret
	if p.ClientSecret != "" {
		p.ClientSecret = MaskedSecret
	}
}
