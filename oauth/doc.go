/*
Package oauth provides OAuth2 and OpenID Connect (OIDC) authentication for Gophish.

This package implements a clean abstraction for OAuth2/OIDC providers, allowing
Gophish to support multiple identity providers (Google, Azure AD, Okta, etc.)
through a unified interface.

# Architecture

The package is organized around several key components:

  - Provider interface: Abstracts OAuth2/OIDC provider operations
  - State management: CSRF protection using signed state tokens
  - Handlers: HTTP handlers for OAuth login flow
  - User mapping: Links OAuth identity to Gophish users

# OAuth Flow

1. User initiates login via /oauth/login?provider=<name>
2. System generates a state token and redirects to provider
3. User authenticates at the OAuth provider
4. Provider redirects to /oauth/callback with authorization code
5. System validates state, exchanges code for tokens
6. System verifies ID token and extracts user claims
7. System finds or creates user, establishes session
8. User is redirected to the dashboard

# Security

The implementation follows OAuth2/OIDC security best practices:

  - CSRF protection via HMAC-signed state parameters
  - Strict redirect URI validation
  - Proper ID token verification (signature, issuer, audience, expiry)
  - Secure session management
  - No sensitive data in logs

# Configuration

OAuth providers are configured in config.json:

	{
	  "oauth": {
	    "enabled": true,
	    "callback_url": "https://gophish.example.com/oauth/callback",
	    "providers": [
	      {
	        "name": "google",
	        "display_name": "Google",
	        "client_id": "xxx.apps.googleusercontent.com",
	        "client_secret": "secret",
	        "issuer_url": "https://accounts.google.com",
	        "scopes": ["openid", "profile", "email"],
	        "enabled": true
	      }
	    ]
	  }
	}
*/
package oauth
