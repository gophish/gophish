package api

import (
	"net/http"

	mid "github.com/gophish/gophish/middleware"
	"github.com/gophish/gophish/middleware/ratelimit"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/worker"
	"github.com/gorilla/mux"
)

// ServerOption is an option to apply to the API server.
type ServerOption func(*Server)

// Server represents the routes and functionality of the Gophish API.
// It's not a server in the traditional sense, in that it isn't started and
// stopped. Rather, it's meant to be used as an http.Handler in the
// AdminServer.
type Server struct {
	handler         http.Handler
	worker          worker.Worker
	limiter         *ratelimit.PostLimiter
	readLimiter     *ratelimit.APILimiter
	writeLimiter    *ratelimit.APILimiter
	campaignLimiter *ratelimit.APILimiter
}

// NewServer returns a new instance of the API handler with the provided
// options applied.
func NewServer(options ...ServerOption) *Server {
	defaultWorker, _ := worker.New()
	defaultLimiter := ratelimit.NewPostLimiter()

	// Create tiered rate limiters for API endpoints
	// Read operations: 100 requests/minute, burst 10
	readLimiter := ratelimit.NewAPILimiter(100.0/60.0, 10)

	// Write operations: 20 requests/minute, burst 5
	writeLimiter := ratelimit.NewAPILimiter(20.0/60.0, 5)

	// Campaign operations: 5 requests/hour, burst 2
	campaignLimiter := ratelimit.NewAPILimiter(5.0/3600.0, 2)

	as := &Server{
		worker:          defaultWorker,
		limiter:         defaultLimiter,
		readLimiter:     readLimiter,
		writeLimiter:    writeLimiter,
		campaignLimiter: campaignLimiter,
	}
	for _, opt := range options {
		opt(as)
	}
	as.registerRoutes()
	return as
}

// WithWorker is an option that sets the background worker.
func WithWorker(w worker.Worker) ServerOption {
	return func(as *Server) {
		as.worker = w
	}
}

func WithLimiter(limiter *ratelimit.PostLimiter) ServerOption {
	return func(as *Server) {
		as.limiter = limiter
	}
}

func (as *Server) registerRoutes() {
	root := mux.NewRouter()
	root = root.StrictSlash(true)
	router := root.PathPrefix("/api/").Subrouter()

	// Apply global middleware
	router.Use(mid.MaxBodySize(10 * 1024 * 1024)) // 10MB max request size
	router.Use(mid.RequireAPIKey)
	router.Use(mid.EnforceViewOnly)

	// IMAP endpoints
	router.Handle("/imap/", as.readLimiter.Limit(http.HandlerFunc(as.IMAPServer)))
	router.Handle("/imap/validate", as.writeLimiter.Limit(http.HandlerFunc(as.IMAPServerValidate)))

	// Reset endpoint (critical operation - heavily rate limited)
	router.Handle("/reset", as.campaignLimiter.Limit(http.HandlerFunc(as.Reset)))

	// Campaign endpoints (tiered based on operation)
	router.HandleFunc("/campaigns/", as.campaignHandler())
	router.Handle("/campaigns/summary", as.readLimiter.Limit(http.HandlerFunc(as.CampaignsSummary)))
	router.HandleFunc("/campaigns/{id:[0-9]+}", as.campaignDetailHandler())
	router.Handle("/campaigns/{id:[0-9]+}/results", as.readLimiter.Limit(http.HandlerFunc(as.CampaignResults)))
	router.Handle("/campaigns/{id:[0-9]+}/summary", as.readLimiter.Limit(http.HandlerFunc(as.CampaignSummary)))
	router.Handle("/campaigns/{id:[0-9]+}/complete", as.writeLimiter.Limit(http.HandlerFunc(as.CampaignComplete)))

	// Group endpoints
	router.HandleFunc("/groups/", as.resourceHandler(as.Groups))
	router.Handle("/groups/summary", as.readLimiter.Limit(http.HandlerFunc(as.GroupsSummary)))
	router.HandleFunc("/groups/{id:[0-9]+}", as.resourceHandler(as.Group))
	router.Handle("/groups/{id:[0-9]+}/summary", as.readLimiter.Limit(http.HandlerFunc(as.GroupSummary)))

	// Template endpoints
	router.HandleFunc("/templates/", as.resourceHandler(as.Templates))
	router.HandleFunc("/templates/{id:[0-9]+}", as.resourceHandler(as.Template))

	// Page endpoints
	router.HandleFunc("/pages/", as.resourceHandler(as.Pages))
	router.HandleFunc("/pages/{id:[0-9]+}", as.resourceHandler(as.Page))

	// SMTP endpoints
	router.HandleFunc("/smtp/", as.resourceHandler(as.SendingProfiles))
	router.HandleFunc("/smtp/{id:[0-9]+}", as.resourceHandler(as.SendingProfile))

	// User endpoints (require special permission)
	router.Handle("/users/", as.writeLimiter.Limit(mid.Use(as.Users, mid.RequirePermission(models.PermissionModifySystem))))
	router.Handle("/users/{id:[0-9]+}", as.writeLimiter.Limit(http.HandlerFunc(as.User)))

	// Utility endpoints
	router.Handle("/util/send_test_email", as.writeLimiter.Limit(http.HandlerFunc(as.SendTestEmail)))

	// Import endpoints
	router.Handle("/import/group", as.writeLimiter.Limit(http.HandlerFunc(as.ImportGroup)))
	router.Handle("/import/email", as.writeLimiter.Limit(http.HandlerFunc(as.ImportEmail)))
	router.Handle("/import/site", as.writeLimiter.Limit(http.HandlerFunc(as.ImportSite)))

	// Webhook endpoints (require special permission)
	router.Handle("/webhooks/", as.readLimiter.Limit(mid.Use(as.Webhooks, mid.RequirePermission(models.PermissionModifySystem))))
	router.Handle("/webhooks/{id:[0-9]+}/validate", as.writeLimiter.Limit(mid.Use(as.ValidateWebhook, mid.RequirePermission(models.PermissionModifySystem))))
	router.Handle("/webhooks/{id:[0-9]+}", as.writeLimiter.Limit(mid.Use(as.Webhook, mid.RequirePermission(models.PermissionModifySystem))))

	as.handler = router
}

// resourceHandler applies appropriate rate limiting based on HTTP method
func (as *Server) resourceHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply read limiter for GET, write limiter for others
		if r.Method == http.MethodGet {
			as.readLimiter.Limit(http.HandlerFunc(handler)).ServeHTTP(w, r)
		} else {
			as.writeLimiter.Limit(http.HandlerFunc(handler)).ServeHTTP(w, r)
		}
	}
}

// campaignHandler applies stricter rate limiting for campaign operations
func (as *Server) campaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// POST (create/launch campaign) is heavily rate limited
		if r.Method == http.MethodPost {
			as.campaignLimiter.Limit(http.HandlerFunc(as.Campaigns)).ServeHTTP(w, r)
		} else {
			as.readLimiter.Limit(http.HandlerFunc(as.Campaigns)).ServeHTTP(w, r)
		}
	}
}

// campaignDetailHandler applies rate limiting for specific campaign operations
func (as *Server) campaignDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			as.readLimiter.Limit(http.HandlerFunc(as.Campaign)).ServeHTTP(w, r)
		} else {
			as.writeLimiter.Limit(http.HandlerFunc(as.Campaign)).ServeHTTP(w, r)
		}
	}
}

func (as *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	as.handler.ServeHTTP(w, r)
}
