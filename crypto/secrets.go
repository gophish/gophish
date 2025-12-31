package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	log "github.com/gophish/gophish/logger"
)

var (
	// ErrDecryptionFailed indicates the ciphertext could not be decrypted
	ErrDecryptionFailed = errors.New("decryption failed")

	// secretKey holds the encryption key (32 bytes for AES-256)
	secretKey []byte
)

// InitSecretEncryption initializes encryption with a key from the environment.
// This must be called once at application startup before any encryption operations.
//
// The GOPHISH_SECRET_KEY environment variable should contain a base64-encoded
// 32-byte key. Generate one with: openssl rand -base64 32
//
// If the key is not set, secrets will NOT be encrypted (development mode).
func InitSecretEncryption() error {
	keyB64 := os.Getenv("GOPHISH_SECRET_KEY")

	if keyB64 == "" {
		log.Warn("GOPHISH_SECRET_KEY not set - OAuth secrets will be stored in plaintext")
		log.Warn("For production, generate a key: openssl rand -base64 32")
		return nil
	}

	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return fmt.Errorf("invalid GOPHISH_SECRET_KEY: must be base64-encoded: %w", err)
	}

	if len(key) != 32 {
		return fmt.Errorf("invalid GOPHISH_SECRET_KEY: must be 32 bytes, got %d", len(key))
	}

	secretKey = key
	log.Info("Secret encryption initialized with AES-256-GCM")
	return nil
}

// EncryptSecret encrypts a plaintext string using AES-256-GCM.
// Returns base64-encoded ciphertext with embedded IV.
//
// If encryption is not initialized, returns plaintext unchanged.
func EncryptSecret(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// No encryption key = development mode
	if len(secretKey) == 0 {
		return plaintext, nil
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", fmt.Errorf("cipher creation failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM creation failed: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce generation failed: %w", err)
	}

	// Encrypt: nonce || ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts a base64-encoded ciphertext using AES-256-GCM.
//
// If encryption is not initialized, returns ciphertext unchanged (assumes plaintext).
// If decryption fails, logs a warning and returns the input unchanged for backward compatibility.
func DecryptSecret(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// No encryption key = development mode
	if len(secretKey) == 0 {
		return ciphertext, nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		// Not base64 = likely plaintext from before encryption was enabled
		log.Debug("Secret not base64-encoded, treating as plaintext")
		return ciphertext, nil
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", fmt.Errorf("cipher creation failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM creation failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		// Too short = likely plaintext
		log.Debug("Secret too short for encrypted format, treating as plaintext")
		return ciphertext, nil
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		// Decryption failed = likely plaintext from before encryption
		log.Debug("Decryption failed, treating as plaintext")
		return ciphertext, nil
	}

	return string(plaintext), nil
}

// IsEncryptionEnabled returns true if secret encryption is active.
func IsEncryptionEnabled() bool {
	return len(secretKey) == 32
}
