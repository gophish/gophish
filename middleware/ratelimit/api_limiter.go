package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/gophish/gophish/middleware"
)

// APILimiter provides tiered rate limiting for API endpoints
type APILimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewAPILimiter creates a new API rate limiter
// rate: requests per second (e.g., 100.0/60.0 = 100 per minute)
// burst: maximum burst size
func NewAPILimiter(r float64, b int) *APILimiter {
	return &APILimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(r),
		burst:    b,
	}
}

// Limit is middleware that enforces rate limits per IP address
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

	// Cleanup old limiters periodically (prevents memory leak)
	go al.cleanupAfterDelay(ip, 10*time.Minute)

	return limiter
}

// cleanupAfterDelay removes a limiter after a delay if it hasn't been used
func (al *APILimiter) cleanupAfterDelay(ip string, delay time.Duration) {
	time.Sleep(delay)

	al.mu.Lock()
	defer al.mu.Unlock()

	// Only remove if we have too many entries (prevents aggressive cleanup)
	if len(al.limiters) > 10000 {
		delete(al.limiters, ip)
	}
}

// getIPAddress extracts the real IP address from the request
// Handles proxied requests correctly
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy/load balancer)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
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

// CleanupAll removes all rate limiters (useful for testing)
func (al *APILimiter) CleanupAll() {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.limiters = make(map[string]*rate.Limiter)
}
