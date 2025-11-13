# Gophish Security Improvement Implementation Plan
**Date**: 2025-11-12
**Priority**: CRITICAL
**Estimated Timeline**: 8-12 weeks
**Risk Level**: Medium (changes require testing)

---

## PHASE 1: CRITICAL FIXES (Week 1-2) - P0 Priority

### 1.1 Upgrade Go Runtime (1-2 days)

**Current**: Go 1.13 (2019)
**Target**: Go 1.23+ (2024/2025)

#### Steps:
```bash
# 1. Update go.mod
go mod edit -go=1.23

# 2. Update dependencies
go get -u ./...
go mod tidy

# 3. Fix breaking changes (if any)
go build ./...

# 4. Run all tests
go test ./...

# 5. Update Docker/CI configurations
# Edit Dockerfile, .github/workflows, etc.
```

#### Breaking Changes to Address:
- `io/ioutil` deprecated → use `os` and `io` packages
- `context` handling improvements
- Type inference changes

#### Validation:
```bash
go version  # Verify 1.23+
go vet ./...
staticcheck ./...
```

**Risk**: LOW - Go maintains excellent backwards compatibility
**Rollback**: Keep Go 1.13 binary as backup

---

### 1.2 Fix CORS Wildcard (1 hour)

**File**: `middleware/middleware.go:78`

**Current Code**:
```go
func RequireAPIKey(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")  // ❌ VULNERABLE
```

**Fixed Code**:
```go
func RequireAPIKey(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get allowed origins from config
        allowedOrigins := getConfig().AdminConf.TrustedOrigins
        origin := r.Header.Get("Origin")

        // Only set CORS if origin is trusted
        if isOriginAllowed(origin, allowedOrigins) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Credentials", "true")
        }

        if r.Method == "OPTIONS" {
            w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
            w.Header().Set("Access-Control-Max-Age", "1000")
            w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
            w.WriteHeader(http.StatusOK)
            return
        }
```

**Add Helper Function**:
```go
func isOriginAllowed(origin string, allowedOrigins []string) bool {
    if len(allowedOrigins) == 0 {
        return false  // Deny all if not configured
    }

    for _, allowed := range allowedOrigins {
        if origin == allowed {
            return true
        }
    }
    return false
}
```

**Update Config**:
```go
// config/config.go
type AdminServer struct {
    // ... existing fields ...
    TrustedOrigins []string `json:"trusted_origins"`  // Add this
}
```

**Update config.json**:
```json
{
  "admin_server": {
    "trusted_origins": [
      "https://admin.yourdomain.com",
      "https://yourdomain.com"
    ]
  }
}
```

**Testing**:
```bash
# Test 1: Trusted origin
curl -H "Origin: https://yourdomain.com" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  https://localhost:3333/api/campaigns

# Test 2: Untrusted origin (should fail)
curl -H "Origin: https://evil.com" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  https://localhost:3333/api/campaigns
```

---

### 1.3 Strengthen Password Policy (2-3 hours)

**File**: `auth/auth.go`

**New Implementation**:
```go
package auth

import (
    "crypto/rand"
    "errors"
    "fmt"
    "io"
    "regexp"
    "strings"
    "unicode"

    "golang.org/x/crypto/bcrypt"
)

// Password policy constants
const (
    MinPasswordLength = 12  // Increased from 8
    MaxPasswordLength = 128
)

// Common passwords to blacklist (expand this list)
var commonPasswords = map[string]bool{
    "password": true,
    "password123": true,
    "admin": true,
    "admin123": true,
    "gophish": true,
    "changeme": true,
    "welcome": true,
    "123456": true,
    "12345678": true,
    "qwerty": true,
}

// Enhanced error messages
var (
    ErrPasswordTooShort = fmt.Errorf("Password must be at least %d characters", MinPasswordLength)
    ErrPasswordTooLong = errors.New("Password exceeds maximum length")
    ErrPasswordNoUppercase = errors.New("Password must contain at least one uppercase letter")
    ErrPasswordNoLowercase = errors.New("Password must contain at least one lowercase letter")
    ErrPasswordNoNumber = errors.New("Password must contain at least one number")
    ErrPasswordNoSpecial = errors.New("Password must contain at least one special character")
    ErrPasswordCommon = errors.New("Password is too common and easily guessable")
    ErrPasswordRepeating = errors.New("Password contains too many repeating characters")
)

// CheckPasswordPolicy validates password strength according to security best practices
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

    // Check for common passwords
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

// hasRepeatingChars checks if password has n or more repeating characters
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

// CalculatePasswordStrength returns a score from 0-100
// Can be used for password strength meters in UI
func CalculatePasswordStrength(password string) int {
    score := 0

    // Length scoring
    if len(password) >= 12 {
        score += 20
    }
    if len(password) >= 16 {
        score += 10
    }
    if len(password) >= 20 {
        score += 10
    }

    // Character diversity scoring
    var hasUpper, hasLower, hasNumber, hasSpecial bool
    charTypes := 0

    for _, char := range password {
        if !hasUpper && unicode.IsUpper(char) {
            hasUpper = true
            charTypes++
        }
        if !hasLower && unicode.IsLower(char) {
            hasLower = true
            charTypes++
        }
        if !hasNumber && unicode.IsNumber(char) {
            hasNumber = true
            charTypes++
        }
        if !hasSpecial && (unicode.IsPunct(char) || unicode.IsSymbol(char)) {
            hasSpecial = true
            charTypes++
        }
    }

    score += charTypes * 15  // 15 points per character type

    // Entropy bonus (character variety)
    uniqueChars := make(map[rune]bool)
    for _, char := range password {
        uniqueChars[char] = true
    }
    entropyRatio := float64(len(uniqueChars)) / float64(len(password))
    score += int(entropyRatio * 20)

    // Cap at 100
    if score > 100 {
        score = 100
    }

    return score
}
```

**Migration Script** for existing users:
```go
// migrations/20250112_enforce_password_policy.go
package main

import (
    "github.com/gophish/gophish/models"
    log "github.com/gophish/gophish/logger"
)

func migratePasswordPolicy() error {
    users, err := models.GetUsers()
    if err != nil {
        return err
    }

    for _, user := range users {
        // Set password change required for all users
        user.PasswordChangeRequired = true
        err = models.PutUser(&user)
        if err != nil {
            log.Errorf("Failed to update user %d: %v", user.Id, err)
            continue
        }
        log.Infof("User %s will be required to change password on next login", user.Username)
    }

    return nil
}
```

**Testing**:
```go
// auth/auth_test.go
func TestPasswordPolicy(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  error
    }{
        {"Too short", "Pass1!", ErrPasswordTooShort},
        {"No uppercase", "password123!", ErrPasswordNoUppercase},
        {"No lowercase", "PASSWORD123!", ErrPasswordNoLowercase},
        {"No number", "PasswordTest!", ErrPasswordNoNumber},
        {"No special", "Password123", ErrPasswordNoSpecial},
        {"Common password", "password123", ErrPasswordCommon},
        {"Repeating chars", "Passssssword1!", ErrPasswordRepeating},
        {"Valid strong", "MyStr0ng!Pass", nil},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CheckPasswordPolicy(tt.password)
            if err != tt.wantErr {
                t.Errorf("CheckPasswordPolicy() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

### 1.4 Fix Session Key Persistence (2-3 hours)

**Problem**: Session keys regenerate on restart, logging out all users

**Solution**: Store session keys in config or environment variables

**File**: `middleware/session.go`

**New Implementation**:
```go
package middleware

import (
    "crypto/rand"
    "encoding/base64"
    "encoding/gob"
    "fmt"
    "os"

    "github.com/gophish/gophish/models"
    "github.com/gorilla/securecookie"
    "github.com/gorilla/sessions"
    log "github.com/gophish/gophish/logger"
)

var Store *sessions.CookieStore

// init registers the necessary models and configures session store
func init() {
    gob.Register(&models.User{})
    gob.Register(&models.Flash{})

    // Initialize session store with persistent keys
    Store = initSessionStore()
}

// initSessionStore creates a session store with persistent keys
func initSessionStore() *sessions.CookieStore {
    authKey := getSessionKey("GOPHISH_SESSION_AUTH_KEY", 64)
    encKey := getSessionKey("GOPHISH_SESSION_ENC_KEY", 32)

    store := sessions.NewCookieStore(authKey, encKey)

    // Security settings
    store.Options.HttpOnly = true
    store.Options.Secure = true  // Force HTTPS
    store.Options.SameSite = 3   // SameSiteStrictMode
    store.MaxAge(86400)  // 1 day instead of 5

    return store
}

// getSessionKey retrieves or generates a session key
func getSessionKey(envVar string, length int) []byte {
    // Try to get from environment variable
    keyStr := os.Getenv(envVar)

    if keyStr != "" {
        key, err := base64.StdEncoding.DecodeString(keyStr)
        if err == nil && len(key) == length {
            log.Infof("Loaded %s from environment", envVar)
            return key
        }
        log.Warnf("Invalid %s in environment, generating new key", envVar)
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
    log.Warnf("Generated new %s:", envVar)
    log.Warnf("Please save this to your environment:")
    log.Warnf("export %s=\"%s\"", envVar, keyStr)
    log.Warnf("==============================================")

    return key
}

// RotateSessionKeys allows rotating session keys
// Call this during maintenance windows
func RotateSessionKeys() error {
    log.Info("Rotating session keys...")

    newAuthKey := securecookie.GenerateRandomKey(64)
    newEncKey := securecookie.GenerateRandomKey(32)

    if newAuthKey == nil || newEncKey == nil {
        return fmt.Errorf("failed to generate new session keys")
    }

    // TODO: Implement graceful key rotation
    // 1. Keep old keys valid for grace period
    // 2. Migrate active sessions
    // 3. Invalidate after grace period

    Store = sessions.NewCookieStore(newAuthKey, newEncKey)
    Store.Options.HttpOnly = true
    Store.Options.Secure = true
    Store.Options.SameSite = 3
    Store.MaxAge(86400)

    log.Info("Session keys rotated successfully")
    log.Warn("Users will need to re-login")

    return nil
}
```

**Setup Instructions**:
```bash
# Generate session keys
openssl rand -base64 64
openssl rand -base64 32

# Add to .env file or systemd service
cat >> .env <<EOF
GOPHISH_SESSION_AUTH_KEY="<generated-64-byte-key>"
GOPHISH_SESSION_ENC_KEY="<generated-32-byte-key>"
EOF

# Load in startup script
source .env
export GOPHISH_SESSION_AUTH_KEY
export GOPHISH_SESSION_ENC_KEY
./gophish
```

---

### 1.5 Add Comprehensive Security Headers (1 hour)

**File**: `middleware/middleware.go`

**Enhanced Implementation**:
```go
// ApplySecurityHeaders applies comprehensive security headers
func ApplySecurityHeaders(next http.Handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Content Security Policy
        csp := []string{
            "default-src 'self'",
            "script-src 'self' 'unsafe-inline' 'unsafe-eval'",  // TODO: Remove unsafe-* after CSP audit
            "style-src 'self' 'unsafe-inline'",
            "img-src 'self' data: https:",
            "font-src 'self' data:",
            "connect-src 'self'",
            "form-action 'self'",
            "frame-ancestors 'none'",
            "base-uri 'self'",
            "upgrade-insecure-requests",
        }
        w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))

        // Frame protection
        w.Header().Set("X-Frame-Options", "DENY")

        // HSTS (HTTP Strict Transport Security)
        // Only set if using HTTPS
        if r.TLS != nil {
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        }

        // Prevent MIME sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")

        // XSS Protection (legacy but still useful)
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        // Referrer Policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        // Permissions Policy (limit browser features)
        permissions := []string{
            "geolocation=()",
            "microphone=()",
            "camera=()",
            "payment=()",
            "usb=()",
            "magnetometer=()",
            "accelerometer=()",
            "gyroscope=()",
        }
        w.Header().Set("Permissions-Policy", strings.Join(permissions, ", "))

        // Remove server identification
        w.Header().Del("Server")
        w.Header().Del("X-Powered-By")

        next.ServeHTTP(w, r)
    }
}
```

**Testing**:
```bash
# Use Mozilla Observatory
# https://observatory.mozilla.org/

# Or curl
curl -I https://localhost:3333 | grep -E "Content-Security-Policy|Strict-Transport-Security|X-Frame-Options"
```

---

## PHASE 2: HIGH PRIORITY FIXES (Week 3-4) - P1 Priority

### 2.1 Implement API Rate Limiting (2-3 days)

**New File**: `middleware/ratelimit/api_limiter.go`

```go
package ratelimit

import (
    "net/http"
    "sync"
    "time"

    "golang.org/x/time/rate"
    "github.com/gophish/gophish/middleware"
)

// APILimiter provides rate limiting for API endpoints
type APILimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

// NewAPILimiter creates a new API rate limiter
// rate: requests per second
// burst: maximum burst size
func NewAPILimiter(r float64, b int) *APILimiter {
    return &APILimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     rate.Limit(r),
        burst:    b,
    }
}

// Limit is middleware that enforces rate limits per IP
func (al *APILimiter) Limit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := getIPAddress(r)

        limiter := al.getLimiter(ip)
        if !limiter.Allow() {
            middleware.JSONError(w, http.StatusTooManyRequests,
                "Rate limit exceeded. Please try again later.")
            return
        }

        next.ServeHTTP(w, r)
    })
}

// getLimiter returns the rate limiter for the given IP
func (al *APILimiter) getLimiter(ip string) *rate.Limiter {
    al.mu.RLock()
    limiter, exists := al.limiters[ip]
    al.mu.RUnlock()

    if exists {
        return limiter
    }

    al.mu.Lock()
    defer al.mu.Unlock()

    // Double-check after acquiring write lock
    if limiter, exists := al.limiters[ip]; exists {
        return limiter
    }

    limiter = rate.NewLimiter(al.rate, al.burst)
    al.limiters[ip] = limiter

    // Cleanup old limiters (optional, prevents memory leak)
    go al.cleanup()

    return limiter
}

// cleanup removes stale rate limiters
func (al *APILimiter) cleanup() {
    time.Sleep(10 * time.Minute)

    al.mu.Lock()
    defer al.mu.Unlock()

    // Remove limiters that haven't been used recently
    // This is a simple cleanup strategy
    if len(al.limiters) > 10000 {
        al.limiters = make(map[string]*rate.Limiter)
    }
}

// getIPAddress extracts IP from request
func getIPAddress(r *http.Request) string {
    // Check X-Forwarded-For header (if behind proxy)
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        return forwarded
    }

    // Check X-Real-IP header
    realIP := r.Header.Get("X-Real-IP")
    if realIP != "" {
        return realIP
    }

    // Fall back to RemoteAddr
    return r.RemoteAddr
}
```

**Apply to API Routes**:
```go
// controllers/api/server.go
func NewServer(options ...ServerOption) *Server {
    // ... existing code ...

    // Create rate limiters
    readLimiter := ratelimit.NewAPILimiter(100.0/60.0, 10)   // 100/min, burst 10
    writeLimiter := ratelimit.NewAPILimiter(20.0/60.0, 5)     // 20/min, burst 5
    campaignLimiter := ratelimit.NewAPILimiter(5.0/3600.0, 2) // 5/hour, burst 2

    // Apply to routes
    router.Handle("/api/campaigns", writeLimiter.Limit(
        mid.RequireAPIKey(http.HandlerFunc(as.Campaigns)))).Methods("POST")

    router.Handle("/api/campaigns", readLimiter.Limit(
        mid.RequireAPIKey(http.HandlerFunc(as.Campaigns)))).Methods("GET")

    router.Handle("/api/campaigns/{id:[0-9]+}", campaignLimiter.Limit(
        mid.RequireAPIKey(http.HandlerFunc(as.Campaign)))).Methods("POST", "PUT")

    // ... apply to other routes ...
}
```

---

### 2.2 Enhance Input Validation (3-4 days)

**New File**: `validation/validation.go`

```go
package validation

import (
    "errors"
    "fmt"
    "html"
    "net/url"
    "regexp"
    "strings"
)

const (
    MaxNameLength        = 255
    MaxEmailLength       = 255
    MaxURLLength         = 2048
    MaxDescriptionLength = 10000
    MaxJSONDepth         = 10
)

var (
    ErrInvalidEmail    = errors.New("invalid email address")
    ErrInvalidURL      = errors.New("invalid URL")
    ErrNameTooLong     = errors.New("name exceeds maximum length")
    ErrInvalidFormat   = errors.New("invalid format")
    ErrContainsHTML    = errors.New("HTML content not allowed in this field")
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

    // Require HTTP/HTTPS
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return fmt.Errorf("URL must use HTTP or HTTPS protocol")
    }

    // Reject file:// and other dangerous protocols
    if parsed.Scheme == "file" || parsed.Scheme == "javascript" {
        return fmt.Errorf("dangerous URL protocol: %s", parsed.Scheme)
    }

    return nil
}

// SanitizeHTML removes dangerous HTML while keeping safe formatting
func SanitizeHTML(input string) string {
    // Use bluemonday for production: github.com/microcosm-cc/bluemonday
    // For now, simple escape
    return html.EscapeString(input)
}

// ValidateName checks if name is valid
func ValidateName(name string) error {
    name = strings.TrimSpace(name)

    if len(name) == 0 {
        return errors.New("name cannot be empty")
    }

    if len(name) > MaxNameLength {
        return ErrNameTooLong
    }

    // Check for HTML injection attempts
    if strings.Contains(name, "<") || strings.Contains(name, ">") {
        return ErrContainsHTML
    }

    return nil
}

// ValidateJSONDepth checks if JSON is nested too deeply (DoS protection)
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
```

**Apply to Models**:
```go
// models/template.go
func (t *Template) Validate() error {
    // Validate name
    if err := validation.ValidateName(t.Name); err != nil {
        return err
    }

    // Validate subject
    if len(t.Subject) == 0 {
        return errors.New("subject cannot be empty")
    }
    if len(t.Subject) > validation.MaxNameLength {
        return errors.New("subject too long")
    }

    // Validate HTML (allow HTML but sanitize)
    // Use bluemonday.StrictPolicy() for production

    // Validate envelope sender
    if t.EnvelopeSender != "" {
        if err := validation.ValidateEmail(t.EnvelopeSender); err != nil {
            return fmt.Errorf("invalid envelope sender: %w", err)
        }
    }

    return nil
}

// Call Validate() before PostTemplate
func PostTemplate(t *Template) error {
    if err := t.Validate(); err != nil {
        return err
    }

    // ... rest of implementation
}
```

---

### 2.3 Implement Security Audit Logging (2-3 days)

**New File**: `audit/audit.go`

```go
package audit

import (
    "encoding/json"
    "time"

    log "github.com/gophish/gophish/logger"
)

// EventType represents different audit event types
type EventType string

const (
    EventLogin              EventType = "auth.login"
    EventLoginFailed        EventType = "auth.login_failed"
    EventLogout             EventType = "auth.logout"
    EventPasswordChange     EventType = "auth.password_change"
    EventPasswordReset      EventType = "auth.password_reset"
    EventAPIKeyUsed         EventType = "auth.api_key_used"
    EventAPIKeyInvalid      EventType = "auth.api_key_invalid"
    EventUserCreated        EventType = "user.created"
    EventUserDeleted        EventType = "user.deleted"
    EventUserModified       EventType = "user.modified"
    EventCampaignCreated    EventType = "campaign.created"
    EventCampaignLaunched   EventType = "campaign.launched"
    EventCampaignDeleted    EventType = "campaign.deleted"
    EventConfigChanged      EventType = "config.changed"
    EventPermissionDenied   EventType = "permission.denied"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    EventType   EventType              `json:"event_type"`
    UserID      int64                  `json:"user_id,omitempty"`
    Username    string                 `json:"username,omitempty"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent,omitempty"`
    Success     bool                   `json:"success"`
    Message     string                 `json:"message"`
    Details     map[string]interface{} `json:"details,omitempty"`
}

// Log records an audit event
func Log(event AuditEvent) {
    event.Timestamp = time.Now().UTC()

    // Marshal to JSON for structured logging
    eventJSON, err := json.Marshal(event)
    if err != nil {
        log.Errorf("Failed to marshal audit event: %v", err)
        return
    }

    // Log to file (configure separate audit log file)
    log.WithFields(log.Fields{
        "audit": true,
        "event": event.EventType,
    }).Info(string(eventJSON))

    // Optional: Send to SIEM/log aggregation service
    // sendToSIEM(eventJSON)
}

// LogLogin records a login attempt
func LogLogin(username, ip, userAgent string, success bool, reason string) {
    Log(AuditEvent{
        EventType: EventLogin,
        Username:  username,
        IPAddress: ip,
        UserAgent: userAgent,
        Success:   success,
        Message:   reason,
    })
}

// LogPasswordChange records a password change
func LogPasswordChange(userID int64, username, ip string, success bool) {
    Log(AuditEvent{
        EventType: EventPasswordChange,
        UserID:    userID,
        Username:  username,
        IPAddress: ip,
        Success:   success,
        Message:   "Password changed",
    })
}

// LogCampaignLaunch records a campaign launch
func LogCampaignLaunch(userID int64, username string, campaignID int64, campaignName, ip string) {
    Log(AuditEvent{
        EventType: EventCampaignLaunched,
        UserID:    userID,
        Username:  username,
        IPAddress: ip,
        Success:   true,
        Message:   "Campaign launched",
        Details: map[string]interface{}{
            "campaign_id":   campaignID,
            "campaign_name": campaignName,
        },
    })
}

// Additional logging functions...
```

**Integrate into controllers**:
```go
// controllers/route.go
func (as *AdminServer) Login(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...

    if err == auth.ErrInvalidPassword {
        audit.LogLogin(username, r.RemoteAddr, r.UserAgent(), false, "Invalid password")
        // ... rest of error handling
    }

    if u.AccountLocked {
        audit.LogLogin(username, r.RemoteAddr, r.UserAgent(), false, "Account locked")
        // ... rest of error handling
    }

    // Successful login
    audit.LogLogin(username, r.RemoteAddr, r.UserAgent(), true, "Login successful")

    // ... rest of implementation
}
```

---

## PHASE 3: DEPENDENCY UPGRADES (Week 5-7) - P1 Priority

### 3.1 Migrate to GORM v2 (5-7 days)

**This is the most complex migration**

#### Step 1: Install GORM v2
```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite
go get -u gorm.io/driver/mysql
```

#### Step 2: Update imports
```go
// OLD
import "github.com/jinzhu/gorm"
import _ "github.com/jinzhu/gorm/dialects/mysql"

// NEW
import "gorm.io/gorm"
import "gorm.io/driver/mysql"
import "gorm.io/driver/sqlite"
```

#### Step 3: Update database connection

**File**: `models/models.go`

**OLD**:
```go
db, err = gorm.Open("sqlite3", dbPath)
```

**NEW**:
```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
    "gorm.io/driver/mysql"
)

// For SQLite
db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
    NowFunc: func() time.Time {
        return time.Now().UTC()
    },
})

// For MySQL
dsn := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=UTC", dbPath)
db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```

#### Step 4: Update query syntax

**OLD GORM v1**:
```go
db.Where("id = ?", id).First(&user)
db.Find(&users)
db.Create(&campaign)
db.Save(&template)
db.Delete(&group)
```

**NEW GORM v2** (mostly compatible, but check errors):
```go
// Most queries work the same, but error handling changed
result := db.Where("id = ?", id).First(&user)
if result.Error != nil {
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        // Handle not found
    }
}

// Create with context
db.WithContext(ctx).Create(&campaign)

// Batch operations improved
db.CreateInBatches(users, 100)
```

#### Step 5: Update associations

**OLD**:
```go
db.Model(&campaign).Association("Results").Find(&results)
```

**NEW**:
```go
db.Model(&campaign).Association("Results").Find(&results)  // Same API!
```

#### Step 6: Test extensively
```bash
# Run all tests
go test ./models/... -v

# Test migrations
go run migrations/migrate.go

# Test in staging environment
```

---

### 3.2 Replace Gorilla Toolkit (7-10 days)

This is a significant refactor. Prioritize based on risk:

#### Option A: Gradual Migration (Recommended)
- Keep Gorilla for now, monitor for security issues
- Plan migration for next major version
- Document technical debt

#### Option B: Immediate Migration
Replace with modern alternatives:

**Router**: gorilla/mux → go-chi/chi
**Sessions**: gorilla/sessions → alexedwards/scs
**CSRF**: gorilla/csrf → justinas/nosurf

**Example for chi**:
```go
// go get github.com/go-chi/chi/v5

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func (as *AdminServer) registerRoutes() {
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(ApplySecurityHeaders)

    // Routes
    r.Get("/", mid.RequireLogin(as.Base))
    r.Post("/login", as.limiter.Limit(as.Login))

    // API routes
    r.Route("/api", func(r chi.Router) {
        r.Use(mid.RequireAPIKey)
        r.Get("/campaigns", as.Campaigns)
        r.Post("/campaigns", as.Campaigns)
    })
}
```

---

## PHASE 4: MEDIUM PRIORITY (Week 8-10) - P2 Priority

### 4.1 Implement Multi-Factor Authentication (MFA)

**Add TOTP support using github.com/pquerna/otp**

```go
go get github.com/pquerna/otp
go get github.com/pquerna/otp/totp
```

**Database migration**:
```sql
ALTER TABLE users ADD COLUMN mfa_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN mfa_secret VARCHAR(255);
```

**Implementation** (summary - full code available on request):
- Add MFA setup page
- Generate TOTP secret
- Store encrypted secret in database
- Validate TOTP during login
- Provide backup codes

### 4.2 Add Account Lockout

**After N failed login attempts, lock account**:

```go
// models/user.go
type User struct {
    // ... existing fields ...
    FailedLoginAttempts int       `json:"-"`
    LastFailedLogin     time.Time `json:"-"`
}

const MaxFailedAttempts = 5
const LockoutDuration = 15 * time.Minute

func (u *User) RecordFailedLogin() error {
    u.FailedLoginAttempts++
    u.LastFailedLogin = time.Now()

    if u.FailedLoginAttempts >= MaxFailedAttempts {
        u.AccountLocked = true
    }

    return db.Save(u).Error
}

func (u *User) ResetFailedLogins() error {
    u.FailedLoginAttempts = 0
    u.LastFailedLogin = time.Time{}
    return db.Save(u).Error
}
```

### 4.3 Add Request Size Limits

**Middleware**:
```go
func MaxBodySize(maxSize int64) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            r.Body = http.MaxBytesReader(w, r.Body, maxSize)
            next.ServeHTTP(w, r)
        })
    }
}

// Apply to routes
router.Use(MaxBodySize(10 * 1024 * 1024))  // 10MB max
```

---

## TESTING STRATEGY

### Security Testing Checklist

```bash
# 1. Dependency scanning
go get -u github.com/securego/gosec/v2/cmd/gosec
gosec ./...

# 2. Static analysis
go get honnef.co/go/tools/cmd/staticcheck
staticcheck ./...

# 3. Vulnerability scanning
go get golang.org/x/vuln/cmd/govulncheck
govulncheck ./...

# 4. OWASP ZAP automated scan
docker run -t owasp/zap2docker-stable zap-baseline.py -t https://localhost:3333

# 5. Manual penetration testing
- SQL injection attempts
- XSS payloads
- CSRF attacks
- Session fixation
- Brute force testing
- Rate limit bypass attempts

# 6. Load testing
go get -u github.com/tsenart/vegeta
echo "GET https://localhost:3333/api/campaigns" | vegeta attack -duration=60s -rate=100 | vegeta report
```

---

## ROLLBACK PROCEDURES

### If issues arise:

1. **Git rollback**:
```bash
git revert HEAD
git push origin main
```

2. **Database rollback**:
```bash
# Keep database backups before migrations
sqlite3 gophish.db ".backup backup-$(date +%Y%m%d).db"

# Rollback migration
go run migrations/rollback.go
```

3. **Binary rollback**:
```bash
# Keep previous binary
cp gophish gophish.backup
# Restore if needed
mv gophish.backup gophish
```

---

## TIMELINE SUMMARY

| Phase | Duration | Priority | Tasks |
|-------|----------|----------|-------|
| 1 | Week 1-2 | P0 | Go upgrade, CORS fix, password policy, session keys, headers |
| 2 | Week 3-4 | P1 | Rate limiting, input validation, audit logging |
| 3 | Week 5-7 | P1 | GORM v2 migration, Gorilla replacement planning |
| 4 | Week 8-10 | P2 | MFA, account lockout, additional hardening |
| 5 | Week 11-12 | P3 | Documentation, training, security review |

**Total**: 8-12 weeks for complete security overhaul

---

## SUCCESS METRICS

- ✓ All CRITICAL and HIGH vulnerabilities remediated
- ✓ OWASP Top 10 compliance score: 8/10 or better
- ✓ Mozilla Observatory score: B+ or better
- ✓ Zero outdated dependencies with known CVEs
- ✓ Comprehensive audit logging implemented
- ✓ Security test suite passing
- ✓ Penetration test report with no critical findings

---

## RESOURCES & REFERENCES

1. **GORM v2 Migration Guide**: https://gorm.io/docs/v2_release_note.html
2. **Go Security Best Practices**: https://go.dev/security/
3. **OWASP Cheat Sheets**: https://cheatsheetseries.owasp.org/
4. **Chi Router**: https://github.com/go-chi/chi
5. **SCS Sessions**: https://github.com/alexedwards/scs
6. **Gosec**: https://github.com/securego/gosec

---

**Document Version**: 1.0
**Last Updated**: 2025-11-12
**Next Review**: After Phase 1 completion
