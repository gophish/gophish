package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"os"

	"github.com/gophish/gophish/models"
	log "github.com/gophish/gophish/logger"
	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

// init registers the necessary models and configures session store
func init() {
	gob.Register(&models.User{})
	gob.Register(&models.Flash{})

	// Initialize session store with persistent keys
	Store = initSessionStore()
}

// initSessionStore creates a session store with persistent keys from environment variables
func initSessionStore() *sessions.CookieStore {
	authKey := getSessionKey("GOPHISH_SESSION_AUTH_KEY", 64)
	encKey := getSessionKey("GOPHISH_SESSION_ENC_KEY", 32)

	store := sessions.NewCookieStore(authKey, encKey)

	// Security settings
	store.Options.HttpOnly = true
	store.Options.Secure = true  // Force HTTPS (will be overridden in main if not using TLS)
	store.Options.SameSite = 3   // SameSiteStrictMode
	store.MaxAge(86400)          // 1 day instead of 5 for better security

	return store
}

// getSessionKey retrieves session key from environment variable or generates a new one
func getSessionKey(envVar string, length int) []byte {
	// Try to get from environment variable
	keyStr := os.Getenv(envVar)

	if keyStr != "" {
		key, err := base64.StdEncoding.DecodeString(keyStr)
		if err == nil && len(key) == length {
			log.Infof("Loaded %s from environment", envVar)
			return key
		}
		log.Warnf("Invalid %s in environment (expected %d bytes base64-encoded), generating new key", envVar, length)
	}

	// Generate new key
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Failed to generate session key: %v", err)
	}

	// Encode to base64 for easy storage
	keyStr = base64.StdEncoding.EncodeToString(key)

	log.Warnf("==============================================")
	log.Warnf("IMPORTANT: Generated new %s", envVar)
	log.Warnf("Please save this to your environment to persist sessions across restarts:")
	log.Warnf("export %s=\"%s\"", envVar, keyStr)
	log.Warnf("Or add to .env file or systemd service configuration")
	log.Warnf("==============================================")

	return key
}
