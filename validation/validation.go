package validation

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Limits for input validation
const (
	MaxNameLength        = 255
	MaxEmailLength       = 255
	MaxURLLength         = 2048
	MaxDescriptionLength = 10000
	MaxJSONDepth         = 10
)

// Error definitions
var (
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidURL      = errors.New("invalid URL")
	ErrNameTooLong     = errors.New("name exceeds maximum length")
	ErrInvalidFormat   = errors.New("invalid format")
	ErrContainsHTML    = errors.New("HTML content not allowed in this field")
	ErrExceedsMaxLength = errors.New("input exceeds maximum length")
)

// Email regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if email is valid
func ValidateEmail(email string) error {
	if len(email) > MaxEmailLength {
		return ErrInvalidEmail
	}

	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateURL checks if URL is valid and safe
func ValidateURL(rawURL string) error {
	if len(rawURL) > MaxURLLength {
		return ErrInvalidURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	// Require HTTP/HTTPS for web URLs
	if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use HTTP or HTTPS protocol, got: %s", parsed.Scheme)
	}

	// Reject dangerous protocols
	dangerous := []string{"file", "javascript", "data", "vbscript"}
	for _, proto := range dangerous {
		if strings.EqualFold(parsed.Scheme, proto) {
			return fmt.Errorf("dangerous URL protocol: %s", parsed.Scheme)
		}
	}

	return nil
}

// SanitizeString removes potentially dangerous characters
// This is a basic implementation - for production consider using bluemonday
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// ValidateName checks if name is valid (no HTML, reasonable length)
func ValidateName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return errors.New("name cannot be empty")
	}

	if len(name) > MaxNameLength {
		return ErrNameTooLong
	}

	// Check for HTML injection attempts
	if containsHTML(name) {
		return ErrContainsHTML
	}

	return nil
}

// containsHTML checks if string contains HTML tags
func containsHTML(s string) bool {
	// Simple check for < and > characters
	// For production, consider more sophisticated HTML detection
	return strings.Contains(s, "<") || strings.Contains(s, ">")
}

// ValidateLength checks if input exceeds maximum length
func ValidateLength(input string, maxLength int) error {
	if len(input) > maxLength {
		return fmt.Errorf("%w: %d characters (max %d)", ErrExceedsMaxLength, len(input), maxLength)
	}
	return nil
}

// ValidateJSONDepth checks if data structure is nested too deeply
// This helps prevent JSON bomb attacks
func ValidateJSONDepth(data interface{}, maxDepth int) error {
	return checkDepth(data, 0, maxDepth)
}

func checkDepth(data interface{}, currentDepth, maxDepth int) error {
	if currentDepth > maxDepth {
		return fmt.Errorf("JSON nesting too deep (max %d levels)", maxDepth)
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for _, val := range v {
			if err := checkDepth(val, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, val := range v {
			if err := checkDepth(val, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidateNoSQLInjection checks for common SQL injection patterns
// Note: This is defense in depth - parameterized queries are the primary defense
func ValidateNoSQLInjection(input string) error {
	// Check for common SQL injection patterns
	dangerous := []string{
		"--",         // SQL comment
		";",          // Statement terminator
		"/*",         // Block comment start
		"*/",         // Block comment end
		"xp_",        // Extended stored procedures
		"sp_",        // System stored procedures
		"0x",         // Hex values
		"@@",         // Global variables
		"EXEC",       // Execute command
		"EXECUTE",    // Execute command
		"UNION",      // Union queries
		"DROP",       // Drop table
		"INSERT",     // Insert data
		"UPDATE",     // Update data
		"DELETE",     // Delete data
		"TRUNCATE",   // Truncate table
	}

	upperInput := strings.ToUpper(input)
	for _, pattern := range dangerous {
		if strings.Contains(upperInput, pattern) {
			return fmt.Errorf("potentially dangerous SQL pattern detected: %s", pattern)
		}
	}

	return nil
}

// ValidateNoXSS checks for common XSS patterns
// Note: This is defense in depth - proper output encoding is the primary defense
func ValidateNoXSS(input string) error {
	// Check for common XSS patterns
	dangerous := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"onmouseover=",
		"<iframe",
		"<embed",
		"<object",
		"eval(",
		"expression(",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range dangerous {
		if strings.Contains(lowerInput, pattern) {
			return fmt.Errorf("potentially dangerous XSS pattern detected: %s", pattern)
		}
	}

	return nil
}
