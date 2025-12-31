package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// stateTokenValidityDuration is how long a state token remains valid
	stateTokenValidityDuration = 10 * time.Minute

	// stateTokenLength is the number of random bytes in the nonce
	stateTokenLength = 32
)

var (
	// ErrExpiredState is returned when the state token has expired
	ErrExpiredState = errors.New("state token has expired")

	// ErrInvalidSignature is returned when the state signature is invalid
	ErrInvalidSignature = errors.New("invalid state signature")
)

// StateManager handles generation and validation of OAuth state tokens.
// State tokens protect against CSRF attacks by ensuring the callback
// matches a legitimate authorization request.
type StateManager struct {
	secret []byte
}

// statePayload represents the internal structure of a state token
type statePayload struct {
	Nonce      string    `json:"n"` // Random nonce
	Provider   string    `json:"p"` // Provider name
	RedirectTo string    `json:"r"` // Where to redirect after login (optional)
	Timestamp  time.Time `json:"t"` // When the token was created
}

// NewStateManager creates a new state manager with the given secret.
// The secret should be a high-entropy byte slice (at least 32 bytes).
func NewStateManager(secret []byte) *StateManager {
	if len(secret) < 32 {
		// Generate a random secret if none provided (logged as a warning)
		secret = make([]byte, 32)
		rand.Read(secret)
	}
	return &StateManager{
		secret: secret,
	}
}

// GenerateState creates a new state token for the given provider.
// The redirectTo parameter specifies where to redirect after successful login.
func (sm *StateManager) GenerateState(provider string, redirectTo string) (string, error) {
	// Generate random nonce
	nonce := make([]byte, stateTokenLength)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	payload := statePayload{
		Nonce:      base64.RawURLEncoding.EncodeToString(nonce),
		Provider:   provider,
		RedirectTo: redirectTo,
		Timestamp:  time.Now().UTC(),
	}

	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state payload: %w", err)
	}

	// Encode payload
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Generate HMAC signature
	signature := sm.sign(encodedPayload)

	// Combine payload and signature
	state := encodedPayload + "." + signature

	return state, nil
}

// ValidateState validates a state token and returns the provider name and redirect URL.
// Returns an error if the token is invalid, expired, or has been tampered with.
func (sm *StateManager) ValidateState(state string) (provider string, redirectTo string, err error) {
	// Split state into payload and signature
	var encodedPayload, signature string
	for i := len(state) - 1; i >= 0; i-- {
		if state[i] == '.' {
			encodedPayload = state[:i]
			signature = state[i+1:]
			break
		}
	}

	if encodedPayload == "" || signature == "" {
		return "", "", ErrInvalidState
	}

	// Verify signature
	expectedSignature := sm.sign(encodedPayload)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", "", ErrInvalidSignature
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode state payload: %w", err)
	}

	// Unmarshal payload
	var payload statePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal state payload: %w", err)
	}

	// Check expiration
	if time.Since(payload.Timestamp) > stateTokenValidityDuration {
		return "", "", ErrExpiredState
	}

	return payload.Provider, payload.RedirectTo, nil
}

// sign generates an HMAC-SHA256 signature for the given data
func (sm *StateManager) sign(data string) string {
	h := hmac.New(sha256.New, sm.secret)
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
