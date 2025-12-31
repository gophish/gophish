package config

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/gophish/gophish/logger"
)

// AdminServer represents the Admin server configuration details
type AdminServer struct {
	ListenURL            string   `json:"listen_url"`
	UseTLS               bool     `json:"use_tls"`
	CertPath             string   `json:"cert_path"`
	KeyPath              string   `json:"key_path"`
	CSRFKey              string   `json:"csrf_key"`
	AllowedInternalHosts []string `json:"allowed_internal_hosts"`
	TrustedOrigins       []string `json:"trusted_origins"`
}

// OAuthProviderConfig represents a single OAuth/OIDC provider configuration
type OAuthProviderConfig struct {
	Name         string   `json:"name"`          // Provider identifier (e.g., "google", "azure", "okta")
	DisplayName  string   `json:"display_name"`  // Human-readable name for UI
	ClientID     string   `json:"client_id"`     // OAuth client ID
	ClientSecret string   `json:"client_secret"` // OAuth client secret
	IssuerURL    string   `json:"issuer_url"`    // OIDC issuer URL (for discovery)
	Scopes       []string `json:"scopes"`        // OAuth scopes to request
	Enabled      bool     `json:"enabled"`       // Whether this provider is active
}

// OAuthConfig represents OAuth/OIDC configuration
type OAuthConfig struct {
	Enabled     bool                  `json:"enabled"`      // Master switch for OAuth
	CallbackURL string                `json:"callback_url"` // OAuth callback URL (e.g., "https://gophish.example.com/oauth/callback")
	Providers   []OAuthProviderConfig `json:"providers"`    // List of configured providers
}

// PhishServer represents the Phish server configuration details
type PhishServer struct {
	ListenURL string `json:"listen_url"`
	UseTLS    bool   `json:"use_tls"`
	CertPath  string `json:"cert_path"`
	KeyPath   string `json:"key_path"`
}

// Config represents the configuration information.
type Config struct {
	AdminConf      AdminServer `json:"admin_server"`
	PhishConf      PhishServer `json:"phish_server"`
	DBName         string      `json:"db_name"`
	DBPath         string      `json:"db_path"`
	DBSSLCaPath    string      `json:"db_sslca_path"`
	MigrationsPath string      `json:"migrations_prefix"`
	TestFlag       bool        `json:"test_flag"`
	ContactAddress string      `json:"contact_address"`
	Logging        *log.Config `json:"logging"`
	OAuth          *OAuthConfig `json:"oauth"` // OAuth/OIDC configuration
}

// Version contains the current gophish version
var Version = ""

// ServerName is the server type that is returned in the transparency response.
const ServerName = "gophish"

// LoadConfig loads the configuration from the specified filepath
func LoadConfig(filepath string) (*Config, error) {
	// Get the config file
	configFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(configFile, config)
	if err != nil {
		return nil, err
	}
	if config.Logging == nil {
		config.Logging = &log.Config{}
	}
	// Choosing the migrations directory based on the database used.
	config.MigrationsPath = config.MigrationsPath + config.DBName
	// Explicitly set the TestFlag to false to prevent config.json overrides
	config.TestFlag = false
	return config, nil
}
