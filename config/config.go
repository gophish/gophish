package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// SMTPServer represents the SMTP configuration details
type SMTPServer struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
}

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
	SMTPConf       SMTPServer  `json:"smtp"`
	DBPath         string      `json:"db_path"`
	MigrationsPath string      `json:"migrations_path"`
}

// Conf contains the initialized configuration struct
var Conf Config

func init() {
	// Get the config file
	config_file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
	}
	json.Unmarshal(config_file, &Conf)
}
