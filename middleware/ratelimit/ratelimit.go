package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/gophish/gophish/logger"
	"golang.org/x/time/rate"
)

// DefaultRequestsPerMinute is the number of requests to allow per minute.
// Any requests over this interval will return a HTTP 429 error.
const DefaultRequestsPerMinute = 5

// DefaultCleanupInterval determines how frequently the cleanup routine
// executes.
const DefaultCleanupInterval = 1 * time.Minute

// DefaultExpiry is the amount of time to track a bucket for a particular
// visitor.
const DefaultExpiry = 10 * time.Minute

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// PostLimiter is a simple rate limiting middleware which only allows n POST
// requests per minute.
type PostLimiter struct {
	visitors        map[string]*bucket
	requestLimit    int
	cleanupInterval time.Duration
	expiry          time.Duration
	sync.RWMutex
}

// PostLimiterOption is a functional option that allows callers to configure
// the rate limiter.
type PostLimiterOption func(*PostLimiter)

// WithRequestsPerMinute sets the number of requests to allow per minute.
func WithRequestsPerMinute(requestLimit int) PostLimiterOption {
	return func(p *PostLimiter) {
		p.requestLimit = requestLimit
	}
}

// WithCleanupInterval sets the interval between cleaning up stale entries in
// the rate limit client list
func WithCleanupInterval(interval time.Duration) PostLimiterOption {
	return func(p *PostLimiter) {
		p.cleanupInterval = interval
	}
}

// WithExpiry sets the amount of time to store client entries before they are
// considered stale.
func WithExpiry(expiry time.Duration) PostLimiterOption {
	return func(p *PostLimiter) {
		p.expiry = expiry
	}
}

// NewPostLimiter returns a new instance of a PostLimiter
func NewPostLimiter(opts ...PostLimiterOption) *PostLimiter {
	limiter := &PostLimiter{
		visitors:        make(map[string]*bucket),
		requestLimit:    DefaultRequestsPerMinute,
		cleanupInterval: DefaultCleanupInterval,
		expiry:          DefaultExpiry,
	}
	for _, opt := range opts {
		opt(limiter)
	}
	go limiter.pollCleanup()
	return limiter
}

func (limiter *PostLimiter) pollCleanup() {
	ticker := time.NewTicker(time.Duration(limiter.cleanupInterval) * time.Second)
	for range ticker.C {
		limiter.Cleanup()
	}
}

// Cleanup removes any buckets that were last seen past the configured expiry.
func (limiter *PostLimiter) Cleanup() {
	limiter.Lock()
	defer limiter.Unlock()
	for ip, bucket := range limiter.visitors {
		if time.Since(bucket.lastSeen) >= limiter.expiry {
			delete(limiter.visitors, ip)
		}
	}
}

func (limiter *PostLimiter) addBucket(ip string) *bucket {
	limiter.Lock()
	defer limiter.Unlock()
	limit := rate.NewLimiter(rate.Every(time.Minute/time.Duration(limiter.requestLimit)), limiter.requestLimit)
	b := &bucket{
		limiter: limit,
	}
	limiter.visitors[ip] = b
	return b
}

func (limiter *PostLimiter) allow(ip string) bool {
	// Check if we have a limiter already active for this clientIP
	limiter.RLock()
	bucket, exists := limiter.visitors[ip]
	limiter.RUnlock()
	if !exists {
		bucket = limiter.addBucket(ip)
	}
	// Update the lastSeen for this bucket to assist with cleanup
	limiter.Lock()
	defer limiter.Unlock()
	bucket.lastSeen = time.Now()
	return bucket.limiter.Allow()
}

// Limit enforces the configured rate limit for POST requests.
//
// TODO: Change the return value to an http.Handler when we clean up the
// way Gophish routing is done.
func (limiter *PostLimiter) Limit(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		}
		if r.Method == http.MethodPost && !limiter.allow(clientIP) {
			log.Error("")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
