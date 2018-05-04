package config

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/gophish/gophish/logger"
)

// AdminServer represents the Admin server configuration details
type AdminServer struct {
	ListenURL string `json:"listen_url"`
	UseTLS    bool   `json:"use_tls"`
	CertPath  string `json:"cert_path"`
	KeyPath   string `json:"key_path"`
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
	MigrationsPath string      `json:"migrations_prefix"`
	TestFlag       bool        `json:"test_flag"`
}

// Conf contains the initialized configuration struct
var Conf Config

// Version contains the current gophish version
var Version = ""

// LoadConfig loads the configuration from the specified filepath
func LoadConfig(filepath string) {
	// Get the config file
	configFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf("File error: %v\n", err)
	}
	json.Unmarshal(configFile, &Conf)

	// Choosing the migrations directory based on the database used.
	Conf.MigrationsPath = Conf.MigrationsPath + Conf.DBName
	// Explicitly set the TestFlag to false to prevent config.json overrides
	Conf.TestFlag = false
}
