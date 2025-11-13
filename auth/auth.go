package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Password policy constants
const (
	MinPasswordLength = 12  // Increased from 8 for better security
	MaxPasswordLength = 128
)

// APIKeyLength is the length of Gophish API keys
const APIKeyLength = 32

// Common passwords to blacklist (expandable list)
var commonPasswords = map[string]bool{
	"password":     true,
	"password123":  true,
	"admin":        true,
	"admin123":     true,
	"gophish":      true,
	"changeme":     true,
	"welcome":      true,
	"123456":       true,
	"12345678":     true,
	"123456789":    true,
	"qwerty":       true,
	"letmein":      true,
	"trustno1":     true,
}

// Enhanced error messages
var (
	ErrInvalidPassword      = errors.New("Invalid Password")
	ErrPasswordMismatch     = errors.New("Passwords do not match")
	ErrReusedPassword       = errors.New("Cannot reuse existing password")
	ErrEmptyPassword        = errors.New("No password provided")
	ErrPasswordTooShort     = fmt.Errorf("Password must be at least %d characters", MinPasswordLength)
	ErrPasswordTooLong      = errors.New("Password exceeds maximum length")
	ErrPasswordNoUppercase  = errors.New("Password must contain at least one uppercase letter")
	ErrPasswordNoLowercase  = errors.New("Password must contain at least one lowercase letter")
	ErrPasswordNoNumber     = errors.New("Password must contain at least one number")
	ErrPasswordNoSpecial    = errors.New("Password must contain at least one special character")
	ErrPasswordCommon       = errors.New("Password is too common and easily guessable")
	ErrPasswordRepeating    = errors.New("Password contains too many repeating characters")
)

// GenerateSecureKey returns the hex representation of key generated from n
// random bytes
func GenerateSecureKey(n int) string {
	k := make([]byte, n)
	io.ReadFull(rand.Reader, k)
	return fmt.Sprintf("%x", k)
}

// GeneratePasswordHash returns the bcrypt hash for the provided password using
// the default bcrypt cost.
func GeneratePasswordHash(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPasswordPolicy ensures the provided password is valid according to security best practices.
// Requirements:
// - At least 12 characters long
// - Contains uppercase letter
// - Contains lowercase letter
// - Contains number
// - Contains special character
// - Not in common passwords list
// - No excessive repeating characters
func CheckPasswordPolicy(password string) error {
	// Check length
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	// Check for common passwords (case-insensitive)
	if commonPasswords[strings.ToLower(password)] {
		return ErrPasswordCommon
	}

	// Check character requirements
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUppercase
	}
	if !hasLower {
		return ErrPasswordNoLowercase
	}
	if !hasNumber {
		return ErrPasswordNoNumber
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	// Check for repeating characters (e.g., "aaaa")
	if hasRepeatingChars(password, 4) {
		return ErrPasswordRepeating
	}

	return nil
}

// hasRepeatingChars checks if password has n or more consecutive repeating characters
func hasRepeatingChars(s string, n int) bool {
	if len(s) < n {
		return false
	}

	for i := 0; i <= len(s)-n; i++ {
		allSame := true
		for j := 1; j < n; j++ {
			if s[i+j] != s[i] {
				allSame = false
				break
			}
		}
		if allSame {
			return true
		}
	}
	return false
}

// ValidatePassword validates that the provided password matches the provided
// bcrypt hash.
func ValidatePassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePasswordChange validates that the new password matches the
// configured password policy, that the new password and confirmation
// password match.
//
// Note that this assumes the current password has been confirmed by the
// caller.
//
// If all of the provided data is valid, then the hash of the new password is
// returned.
func ValidatePasswordChange(currentHash, newPassword, confirmPassword string) (string, error) {
	// Ensure the new password passes our password policy
	if err := CheckPasswordPolicy(newPassword); err != nil {
		return "", err
	}
	// Check that new passwords match
	if newPassword != confirmPassword {
		return "", ErrPasswordMismatch
	}
	// Make sure that the new password isn't the same as the old one
	err := ValidatePassword(newPassword, currentHash)
	if err == nil {
		return "", ErrReusedPassword
	}
	// Generate the new hash
	return GeneratePasswordHash(newPassword)
}
