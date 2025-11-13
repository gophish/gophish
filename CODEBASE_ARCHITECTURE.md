# Gophish Codebase Architecture Overview

## Project Overview
**Gophish** is an open-source phishing framework and security awareness training tool written in Go. It allows businesses and penetration testers to quickly design, execute, and track phishing campaigns and email security awareness training exercises.

- **Language:** Go 1.13+
- **License:** MIT
- **Total Lines of Code:** ~10,525 lines across 86 Go files
- **Module:** github.com/gophish/gophish

---

## 1. Overall Project Structure

The Gophish project is organized as a monolithic application with clear separation between:
- **Admin Server** - Web UI and REST API for campaign management
- **Phishing Server** - Lightweight server for tracking phishing email interactions
- **Background Worker** - Campaign execution and email sending
- **Database Layer** - GORM-based ORM with GORM migrations (Goose)

### Key Architectural Patterns:
- **Multi-mode operation** - Can run as "admin", "phish", or "all" (combined)
- **Multi-user/Multi-tenant** - Role-based access control (RBAC) with User and Admin roles
- **Database agnostic** - Supports SQLite3 (default) and MySQL via GORM and database drivers
- **Modular design** - Separate packages for models, controllers, middleware, etc.

---

## 2. Key Directories and Their Purposes

```
/home/user/gophish/
├── auth/                      # Authentication & password management (bcrypt, API keys)
├── config/                    # Configuration management (JSON-based config parsing)
├── context/                   # Request context utilities (Go 1.7+ context wrapper)
├── controllers/               # HTTP handlers for both admin UI and API
│   ├── api/                   # REST API endpoints and routing
│   ├── route.go               # Admin server routes and page handlers
│   └── phish.go               # Phishing server email tracking handlers
├── db/                        # Database layer
│   ├── db_sqlite3/
│   │   └── migrations/        # SQL migration files (Goose format)
│   └── db_mysql/
│       └── migrations/        # MySQL-specific migrations
├── dialer/                    # Network dialing with SSRF protection
├── doc/                       # Documentation
├── docker/                    # Docker-related files
├── imap/                      # IMAP monitoring for reported emails
├── logger/                    # Logging infrastructure (logrus-based)
├── mailer/                    # Email sending/queuing interface
├── middleware/                # HTTP middleware (auth, logging, CSRF, rate limiting)
│   └── ratelimit/             # Rate limiting for POST requests
├── models/                    # Data models (GORM ORM models)
├── static/                    # Frontend assets (JS, CSS, images)
├── templates/                 # HTML templates for admin UI
├── util/                      # Utility functions (CSV parsing, SSL cert generation)
├── webhook/                   # Webhook event notifications
├── worker/                    # Background worker for campaign processing
├── gophish.go                 # Main entry point
├── go.mod / go.sum            # Go module dependencies
└── config.json                # Default configuration file
```

---

## 3. Main Entry Points and Core Functionality

### Primary Entry Point: `gophish.go`

```go
func main() {
    // 1. Load version
    // 2. Parse CLI flags (config path, mode: admin/phish/all, disable-mailer)
    // 3. Load configuration from JSON
    // 4. Setup logging
    // 5. Initialize models and database
    // 6. Create servers based on mode:
    //    - AdminServer (if admin/all mode)
    //    - PhishingServer (if phish/all mode)
    // 7. Start IMAP monitor (if admin mode)
    // 8. Handle graceful shutdown on SIGINT
}
```

### CLI Flags Available:
- `--config` - Path to config.json (default: "./config.json")
- `--disable-mailer` - Disable built-in mailer (for distributed deployments)
- `--mode` - Operation mode: "admin", "phish", or "all" (default: "all")

### Core Components Initialized:
1. **Models Package** - Database setup, migrations, user authentication
2. **Admin Server** - HTTP server on configured port (default: :3333)
3. **Phishing Server** - HTTP server for tracking (default: :80)
4. **IMAP Monitor** - Background monitoring for email reports
5. **Background Worker** - Email campaign execution

---

## 4. Technology Stack

### Go Dependencies (Key Libraries)

**ORM & Database:**
- `github.com/jinzhu/gorm` (v1.9.12) - Object-relational mapping
- `bitbucket.org/liamstask/goose` - Database migrations
- `github.com/go-sql-driver/mysql` (v1.5.0) - MySQL driver
- `github.com/mattn/go-sqlite3` (v2.0.3) - SQLite3 driver

**Web Framework & Routing:**
- `github.com/gorilla/mux` (v1.7.3) - HTTP router
- `github.com/gorilla/csrf` (v1.6.2) - CSRF protection
- `github.com/gorilla/sessions` (v1.2.0) - Session management
- `github.com/gorilla/securecookie` (v1.1.1) - Encrypted cookies
- `github.com/gorilla/context` (v1.1.1) - Request context
- `github.com/gorilla/handlers` (v1.4.2) - HTTP middleware (logging, CORS, etc.)

**Security & Encryption:**
- `golang.org/x/crypto` - Bcrypt password hashing, TLS configuration
- `github.com/NYTimes/gziphandler` (v1.1.1) - Response compression

**Email & Communication:**
- `github.com/gophish/gomail` - Email sending (custom fork)
- `github.com/jordan-wright/email` (v4.0.1) - Email utilities
- `github.com/emersion/go-imap` (v1.0.4) - IMAP client for email monitoring
- `github.com/emersion/go-message` (v0.12.0) - RFC822 message parsing

**Other:**
- `github.com/PuerkitoBio/goquery` (v1.5.0) - HTML parsing/manipulation
- `github.com/oschwald/maxminddb-golang` (v1.6.0) - GeoIP lookups
- `github.com/sirupsen/logrus` (v1.4.2) - Logging
- `gopkg.in/alecthomas/kingpin.v2` (v2.2.6) - CLI argument parsing
- `golang.org/x/time` - Rate limiting

### Go Version Requirement:
- **Minimum:** Go 1.10
- **Used in Project:** Go 1.13+

---

## 5. API Endpoints and Routing

### Admin API Base Path: `/api/`

All API endpoints require **API Key authentication** via either:
- GET parameter: `?api_key=<key>`
- HTTP Header: `Authorization: Bearer <api_key>`

### REST API Endpoints Structure:

**Campaigns:**
```
GET    /api/campaigns/                    # List user's campaigns
POST   /api/campaigns/                    # Create campaign
GET    /api/campaigns/summary             # Campaign summaries
GET    /api/campaigns/{id}                # Get campaign details
GET    /api/campaigns/{id}/results        # Get campaign results
GET    /api/campaigns/{id}/summary        # Campaign summary
PUT    /api/campaigns/{id}                # Update campaign
DELETE /api/campaigns/{id}                # Delete campaign
POST   /api/campaigns/{id}/complete       # Complete campaign
```

**Groups (Target Lists):**
```
GET    /api/groups/                       # List groups
POST   /api/groups/                       # Create group
GET    /api/groups/summary                # Group summaries
GET    /api/groups/{id}                   # Get group details
GET    /api/groups/{id}/summary           # Group summary
PUT    /api/groups/{id}                   # Update group
DELETE /api/groups/{id}                   # Delete group
```

**Templates (Email):**
```
GET    /api/templates/                    # List templates
POST   /api/templates/                    # Create template
GET    /api/templates/{id}                # Get template
PUT    /api/templates/{id}                # Update template
DELETE /api/templates/{id}                # Delete template
```

**Landing Pages:**
```
GET    /api/pages/                        # List pages
POST   /api/pages/                        # Create page
GET    /api/pages/{id}                    # Get page
PUT    /api/pages/{id}                    # Update page
DELETE /api/pages/{id}                    # Delete page
```

**Sending Profiles (SMTP):**
```
GET    /api/smtp/                         # List SMTP profiles
POST   /api/smtp/                         # Create SMTP profile
GET    /api/smtp/{id}                     # Get SMTP profile
PUT    /api/smtp/{id}                     # Update SMTP profile
DELETE /api/smtp/{id}                     # Delete SMTP profile
```

**Users (Admin Only):**
```
GET    /api/users/                        # List users (requires PermissionModifySystem)
POST   /api/users/                        # Create user
GET    /api/users/{id}                    # Get user
PUT    /api/users/{id}                    # Update user
DELETE /api/users/{id}                    # Delete user
```

**Utilities:**
```
POST   /api/util/send_test_email          # Send test email
GET    /api/import/group                  # Import group from CSV
POST   /api/import/email                  # Import emails
POST   /api/import/site                   # Import landing page from URL
```

**Email Monitoring (IMAP):**
```
GET    /api/imap/                         # Get IMAP settings
POST   /api/imap/                         # Update IMAP settings
POST   /api/imap/validate                 # Validate IMAP connection
```

**Webhooks (Admin Only):**
```
GET    /api/webhooks/                     # List webhooks
POST   /api/webhooks/                     # Create webhook
GET    /api/webhooks/{id}                 # Get webhook
PUT    /api/webhooks/{id}                 # Update webhook
DELETE /api/webhooks/{id}                 # Delete webhook
POST   /api/webhooks/{id}/validate        # Test webhook
```

### Admin UI Routes (Web Interface):
```
GET  /                           # Dashboard
GET  /login                      # Login page
POST /login                      # Login submission (rate-limited)
POST /logout                     # Logout
GET  /campaigns                  # Campaigns page
GET  /campaigns/{id}             # Campaign details
GET  /templates                  # Email templates page
GET  /groups                     # Target groups page
GET  /landing_pages              # Landing pages page
GET  /sending_profiles           # SMTP profiles page
GET  /settings                   # User settings
GET  /users                      # User management (admin only)
GET  /webhooks                   # Webhooks page (admin only)
GET  /impersonate                # User impersonation (admin only)
```

### Phishing Server Endpoints:
```
GET  /{rid}                      # Track email open (1x1 pixel tracking)
POST /{rid}                      # Capture form submissions (credentials)
GET  /{rid}+                     # Transparency response
```

---

## 6. Database Models and ORM Usage

### ORM Framework: GORM (v1.9.12)

**Key Models:**

**User Model:**
```go
type User struct {
    Id                     int64     // Primary key
    Username               string    // Unique username
    Hash                   string    // Bcrypt password hash
    ApiKey                 string    // Unique API key (32 chars)
    Role                   Role      // Associated role (Admin/User)
    RoleID                 int64     // Foreign key to Role
    PasswordChangeRequired bool      // Force password reset on login
    AccountLocked          bool      // Account locked status
    LastLogin              time.Time // Last successful login
}
```

**Campaign Model:**
```go
type Campaign struct {
    Id            int64      // Campaign ID
    UserId        int64      // Owner user ID
    Name          string     // Campaign name
    CreatedDate   time.Time  // Creation timestamp
    LaunchDate    time.Time  // When campaign starts
    SendByDate    time.Time  // Deadline for sending
    CompletedDate time.Time  // Completion timestamp
    TemplateId    int64      // Email template FK
    Template      Template   // Email template reference
    PageId        int64      // Landing page FK
    Page          Page       // Landing page reference
    Status        string     // "Created", "In progress", "Completed", etc.
    Results       []Result   // Campaign results
    Groups        []Group    // Target groups
    Events        []Event    // Timeline events
    SMTPId        int64      // Sending profile FK
    SMTP          SMTP       // SMTP profile reference
    URL           string     // Campaign URL
}
```

**Result Model:**
```go
type Result struct {
    Id           int64     // Result ID
    CampaignId   int64     // Campaign FK
    UserId       int64     // User FK
    RId          string    // Unique tracking ID
    Status       string    // "Sent", "Opened", "Clicked Link", "Submitted Data", etc.
    IP           string    // IP address on interaction
    Latitude     float64   // GeoIP latitude
    Longitude    float64   // GeoIP longitude
    SendDate     time.Time // When email was sent
    Reported     bool      // Whether user reported email
    ModifiedDate time.Time // Last update
    BaseRecipient         // Email, first_name, last_name, position
}
```

**Template Model:**
```go
type Template struct {
    Id             int64        // Template ID
    UserId         int64        // Owner user ID
    Name           string       // Template name
    EnvelopeSender string       // Envelope sender (optional)
    Subject        string       // Email subject
    Text           string       // Plain text body
    HTML           string       // HTML body
    ModifiedDate   time.Time    // Last modified
    Attachments    []Attachment // File attachments
}
```

**Page Model:**
```go
type Page struct {
    Id                 int64     // Page ID
    UserId             int64     // Owner user ID
    Name               string    // Page name
    HTML               string    // HTML content
    CaptureCredentials bool      // Capture form inputs
    CapturePasswords   bool      // Capture password fields
    RedirectURL        string    // Post-submission redirect
    ModifiedDate       time.Time // Last modified
}
```

**Group Model:**
```go
type Group struct {
    Id           int64    // Group ID
    UserId       int64    // Owner user ID
    Name         string   // Group name
    ModifiedDate time.Time
    Targets      []Target // Email targets
}

type Target struct {
    Id   int64
    BaseRecipient // Email, first_name, last_name, position
}
```

**SMTP Profile Model:**
```go
type SMTP struct {
    Id               int64     // SMTP ID
    UserId           int64     // Owner user ID
    Interface        string    // "SMTP" type
    Name             string    // Profile name
    Host             string    // SMTP server host:port
    Username         string    // SMTP username (optional)
    Password         string    // SMTP password (optional)
    FromAddress      string    // From email address
    IgnoreCertErrors bool      // Skip TLS verification
    Headers          []Header  // Custom email headers
    ModifiedDate     time.Time
}
```

**MailLog Model:**
```go
type MailLog struct {
    Id         int64     // Log ID
    UserId     int64     // User FK
    CampaignId int64     // Campaign FK
    RId        string    // Result ID (tracking ID)
    SendDate   time.Time // Scheduled send date
    SendAttempt int      // Number of retry attempts
    Processing bool      // Currently being processed
}
```

**Event Model:**
```go
type Event struct {
    Id         int64     // Event ID
    CampaignId int64     // Campaign FK
    Email      string    // Target email
    Time       time.Time // Event timestamp
    Message    string    // Event type (e.g., "Email Sent", "Clicked Link")
    Details    string    // JSON details object
}
```

**IMAP Model:**
```go
type IMAP struct {
    UserId                      int64     // User FK
    Enabled                     bool      // Is IMAP monitoring enabled
    Host                        string    // IMAP server host
    Port                        uint16    // IMAP port
    Username                    string    // IMAP username
    Password                    string    // IMAP password
    TLS                         bool      // Use TLS
    IgnoreCertErrors            bool      // Skip cert verification
    Folder                      string    // Email folder to monitor (default: INBOX)
    RestrictDomain              string    // Restrict reports to domain
    DeleteReportedCampaignEmail bool      // Auto-delete reported emails
    LastLogin                   time.Time // Last successful connection
    IMAPFreq                    uint32    // Polling frequency in seconds (default: 60)
}
```

**Webhook Model:**
```go
type Webhook struct {
    Id       int64  // Webhook ID
    Name     string // Webhook name
    URL      string // Webhook endpoint URL
    Secret   string // HMAC secret for signing
    IsActive bool   // Is webhook enabled
}
```

**Role & Permission Models:**
```go
type Role struct {
    ID          int64        // Role ID
    Slug        string       // "admin" or "user"
    Name        string       // Display name
    Description string       // Description
    Permissions []Permission // Associated permissions
}

type Permission struct {
    ID          int64  // Permission ID
    Slug        string // Permission identifier
    Name        string // Display name
    Description string // Description
}

// Permission Slugs:
// - "view_objects" - Can view campaigns, groups, etc.
// - "modify_objects" - Can create/modify campaigns, groups, etc.
// - "modify_system" - Can manage users and system settings
```

### Database Migrations:
- **Tool:** Goose (bitbucket.org/liamstask/goose)
- **Location:** `/db/db_sqlite3/migrations/` and `/db/db_mysql/migrations/`
- **Format:** SQL files with timestamp prefixes
- **Migration Examples:**
  - Initial schema creation
  - Event details field addition
  - Campaign scheduling support
  - Email header support
  - MailLogs for queuing
  - User reporting feature
  - Result modified date tracking

---

## 7. Authentication and Authorization

### Authentication Mechanisms:

**Session-Based (Admin UI):**
- Gorilla sessions with secure cookies
- Session key stored in encrypted cookie
- User ID stored in session
- CSRF tokens for form submissions (Gorilla CSRF)
- Minimum password length: 8 characters
- Password hashing: bcrypt (golang.org/x/crypto/bcrypt)

**API Key Authentication:**
- 32-character hex strings (generated via crypto/rand)
- Can be provided via:
  - GET parameter: `?api_key=<key>`
  - HTTP Header: `Authorization: Bearer <key>`
- Unique per user
- Persisted in User.ApiKey field

**Initial Admin User:**
- Auto-created on first startup
- Temporary password generated and logged (or from GOPHISH_INITIAL_ADMIN_PASSWORD env var)
- User must change password on first login
- API token can be set via GOPHISH_INITIAL_ADMIN_API_TOKEN env var

### Authorization (RBAC):

**Two Built-in Roles:**

1. **Admin Role** (`admin`):
   - PermissionViewObjects - View campaigns, groups, templates, pages, etc.
   - PermissionModifyObjects - Create and modify campaigns, groups, templates, pages, etc.
   - PermissionModifySystem - Manage users, system configuration, webhooks

2. **User Role** (`user`):
   - PermissionViewObjects - View their own campaigns, groups, templates, pages, etc.
   - PermissionModifyObjects - Create and modify their own campaigns, groups, templates, pages, etc.
   - No PermissionModifySystem access

**Permission Checks:**
- Implemented via `User.HasPermission(slug string)` method
- Checks user's role and its permissions
- Used in API endpoints and page handlers

**User Management Constraints:**
- Only admins can create/delete users
- Cannot delete the last admin user (ErrModifyingOnlyAdmin)
- Users with forced password change cannot perform other actions

### Security Features:

**CSRF Protection:**
- Gorilla CSRF middleware on admin routes
- CSRF token in session cookies
- Token included in forms and required for POST requests
- API routes exempt from CSRF (they use API keys instead)

**Password Policy:**
- Minimum 8 characters
- Cannot reuse previous password
- Validated with bcrypt comparison

**Rate Limiting:**
- Login endpoint: 5 POST requests per minute per IP
- PostLimiter with per-IP buckets using golang.org/x/time/rate
- Stale entries cleaned up after 10 minutes

**TLS/HTTPS:**
- Configurable TLS support for both admin and phishing servers
- Strong TLS 1.2+ with secure cipher suites
- Auto-generates self-signed certs if not provided

**SSRF Protection:**
- RestrictedDialer prevents outbound connections to internal ranges
- Blocks by default: 169.254.0.0/16 (metadata service)
- Can whitelist allowed internal hosts via config
- Blocks all internal ranges if whitelist is specified

**CORS & Headers:**
- Security headers set: X-Frame-Options, X-Content-Type-Options, etc.
- X-Forwarded-For and X-Real-IP header support for reverse proxies

---

## 8. Configuration Management

### Configuration File: `config.json`

**Structure:**
```json
{
  "admin_server": {
    "listen_url": "0.0.0.0:3333",
    "use_tls": true,
    "cert_path": "gophish_admin.crt",
    "key_path": "gophish_admin.key",
    "csrf_key": "",
    "allowed_internal_hosts": [],
    "trusted_origins": []
  },
  "phish_server": {
    "listen_url": "0.0.0.0:80",
    "use_tls": false,
    "cert_path": "gophish_phish.crt",
    "key_path": "gophish_phish.key"
  },
  "db_name": "sqlite3",
  "db_path": "gophish.db",
  "db_sslca_path": "",
  "migrations_prefix": "./db/",
  "contact_address": "contact@example.com",
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

**Configuration Fields:**

**Admin Server:**
- `listen_url` - Address and port to listen on (default: 0.0.0.0:3333)
- `use_tls` - Enable HTTPS (default: true)
- `cert_path` - Path to SSL certificate
- `key_path` - Path to SSL private key
- `csrf_key` - CSRF token key (auto-generated if empty)
- `allowed_internal_hosts` - IP ranges allowed for webhooks (SSRF allowlist)
- `trusted_origins` - Origins allowed for CORS requests

**Phishing Server:**
- `listen_url` - Address and port to listen on (default: 0.0.0.0:80)
- `use_tls` - Enable HTTPS
- `cert_path` - Path to SSL certificate
- `key_path` - Path to SSL private key

**Database:**
- `db_name` - Database type: "sqlite3" (default) or "mysql"
- `db_path` - SQLite: file path; MySQL: connection string (user:password@host:port/dbname?params)
- `db_sslca_path` - CA certificate for MySQL TLS
- `migrations_prefix` - Base path for migrations directory

**Application:**
- `contact_address` - Contact email in transparency responses
- `logging` - Logging configuration (level, format)

### Loading Configuration:

1. Configuration loaded from JSON file specified by `--config` flag
2. Defaults to `./config.json` if not specified
3. Database migrations path constructed as: `{migrations_prefix}/{db_name}`
4. TestFlag always set to false (prevents config overrides)

### Environment Variables:

- `GOPHISH_INITIAL_ADMIN_PASSWORD` - Set initial admin password
- `GOPHISH_INITIAL_ADMIN_API_TOKEN` - Set initial admin API token

---

## Summary of Key Components

| Component | Purpose | Language | Key Tech |
|-----------|---------|----------|----------|
| **Admin Server** | Web UI & REST API | Go | Gorilla Mux, GORM |
| **Phishing Server** | Email tracking | Go | Gorilla Mux, GeoIP |
| **Worker** | Campaign execution | Go | Gorilla mail, SMTP |
| **IMAP Monitor** | Email reporting | Go | go-imap |
| **Database** | Persistent storage | SQL | GORM, SQLite3/MySQL |
| **Frontend** | User interface | HTML/JS | Bootstrap |
| **Authentication** | User login & API | Go | bcrypt, Sessions |
| **Middleware** | Request processing | Go | Gorilla, CSRF, Rate Limiting |

---

## Data Flow Examples

### Campaign Execution Flow:
1. Admin creates campaign via UI/API (POST /api/campaigns/)
2. Campaign stored in database with Status="Created"
3. Admin launches campaign (updates Status to "In progress")
4. Worker polls for queued MailLogs
5. Worker creates email messages from template + group targets
6. Worker sends via SMTP profile
7. Email opens tracked via GET /{tracking_id}
8. Link clicks tracked via POST /{tracking_id}
9. Form submissions captured via POST form submission
10. Results aggregated and displayed in UI

### Email Reporting Flow:
1. Recipient marks email as spam/phishing
2. IMAP monitor checks configured mailbox periodically
3. Monitor identifies reported emails via headers/subject
4. Updates Result.Reported = true for matching targets
5. Event logged as "Email Reported"
6. Webhook triggered (if configured)
7. Admin views in campaign results

### Authentication Flow (UI):
1. User accesses /login
2. Enters username and password
3. POST /login with rate limiting
4. System validates against User.Hash (bcrypt)
5. Creates session with User.Id
6. Sets secure cookie
7. Redirects to dashboard
8. Subsequent requests include session cookie
9. Middleware validates session and loads User object
10. User object available in request context

---

## Performance Considerations

- **Database:** Max open connections set to 1 for SQLite (concurrent access protection)
- **Email Sending:** Grouped by campaign/SMTP profile to reuse connections
- **MailLogs:** Exponential backoff on retries (max 8 attempts = ~4.2 hours)
- **Rate Limiting:** Per-IP tracking with cleanup every minute
- **Gzip Compression:** Response compression at best level
- **Caching:** Campaign cache during bulk mail processing

---

## Security Considerations Implemented

1. **Input Validation** - All user inputs validated before storage
2. **SQL Injection Prevention** - GORM parameterized queries
3. **Password Security** - bcrypt hashing with salt
4. **Session Management** - Secure cookies with encryption
5. **CSRF Protection** - Token validation on form submissions
6. **Rate Limiting** - Prevent brute force attacks on login
7. **SSRF Protection** - Restricted dialer for outbound connections
8. **TLS** - Configurable HTTPS on both servers
9. **Access Control** - Role-based permission checks
10. **Audit Logging** - Events tracked in database

