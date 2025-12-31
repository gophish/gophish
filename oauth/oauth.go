package oauth

import (
	"context"
	"errors"
	"time"

	"golang.org/x/oauth2"
)

// Common errors returned by the oauth package
var (
	// ErrProviderNotFound is returned when a requested provider is not configured
	ErrProviderNotFound = errors.New("oauth provider not found")

	// ErrInvalidState is returned when the OAuth state parameter is invalid
	ErrInvalidState = errors.New("invalid oauth state parameter")

	// ErrTokenExchange is returned when the authorization code exchange fails
	ErrTokenExchange = errors.New("failed to exchange authorization code for token")

	// ErrTokenVerification is returned when ID token verification fails
	ErrTokenVerification = errors.New("failed to verify ID token")

	// ErrMissingClaims is returned when required claims are missing from the ID token
	ErrMissingClaims = errors.New("required claims missing from ID token")

	// ErrOAuthDisabled is returned when OAuth is not enabled in configuration
	ErrOAuthDisabled = errors.New("oauth is not enabled")
)

// Provider defines the interface for OAuth2/OIDC provider implementations.
// This abstraction allows supporting multiple identity providers through
// a unified interface.
type Provider interface {
	// Name returns the provider identifier (e.g., "google", "azure")
	Name() string

	// DisplayName returns a human-readable name for UI display
	DisplayName() string

	// AuthCodeURL returns the URL for the authorization request.
	// The state parameter should be a cryptographically secure random string
	// for CSRF protection.
	AuthCodeURL(state string) string

	// Exchange converts an authorization code into an access token.
	// Returns the raw OAuth2 token and any ID token if available (OIDC).
	Exchange(ctx context.Context, code string) (*Token, error)

	// VerifyIDToken verifies and parses an OIDC ID token.
	// Returns the verified claims or an error if verification fails.
	VerifyIDToken(ctx context.Context, rawIDToken string) (*IDTokenClaims, error)

	// RefreshToken exchanges a refresh token for a new access token.
	// Returns nil error if refresh tokens are not supported.
	RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
}

// Token represents an OAuth2 token response.
// This wraps the standard oauth2.Token with additional fields.
type Token struct {
	// AccessToken is the token for making authenticated API requests
	AccessToken string

	// TokenType is the type of token (usually "Bearer")
	TokenType string

	// RefreshToken is used to obtain new access tokens (optional)
	RefreshToken string

	// Expiry is when the access token expires
	Expiry time.Time

	// IDToken is the raw OIDC ID token (optional, for OIDC providers)
	IDToken string
}

// FromOAuth2Token converts an oauth2.Token to our Token type
func FromOAuth2Token(t *oauth2.Token) *Token {
	if t == nil {
		return nil
	}
	token := &Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
	}
	// Extract ID token if present (OIDC)
	if idToken, ok := t.Extra("id_token").(string); ok {
		token.IDToken = idToken
	}
	return token
}

// Valid returns true if the token is valid (not expired)
func (t *Token) Valid() bool {
	return t != nil && t.AccessToken != "" && !t.Expiry.IsZero() && time.Now().Before(t.Expiry)
}

// IDTokenClaims represents verified claims from an OIDC ID token.
// These are the standard OIDC claims that Gophish uses for user identity.
type IDTokenClaims struct {
	// Subject is the unique identifier for the user at the provider
	Subject string

	// Email is the user's email address (if available)
	Email string

	// EmailVerified indicates if the email has been verified by the provider
	EmailVerified bool

	// Name is the user's full name (if available)
	Name string

	// PreferredUsername is the user's preferred username (if available)
	PreferredUsername string

	// Issuer is the provider's issuer URL
	Issuer string

	// Audience is the client ID(s) this token is intended for
	Audience []string

	// IssuedAt is when the token was issued
	IssuedAt time.Time

	// Expiry is when the token expires
	Expiry time.Time
}

// Valid returns true if the token claims are currently valid
func (c *IDTokenClaims) Valid() bool {
	if c == nil {
		return false
	}
	now := time.Now()
	// Check expiry
	if !c.Expiry.IsZero() && now.After(c.Expiry) {
		return false
	}
	// Basic validation: must have subject and issuer
	return c.Subject != "" && c.Issuer != ""
}

// GetPreferredUsername returns the best available username claim
func (c *IDTokenClaims) GetPreferredUsername() string {
	if c.PreferredUsername != "" {
		return c.PreferredUsername
	}
	if c.Email != "" {
		return c.Email
	}
	return c.Subject
}
