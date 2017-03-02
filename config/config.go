package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
}

// Conf contains the initialized configuration struct
var Conf Config

// Version contains the current gophish version
var Version = "0.3"

func init() {
	// Get the config file
	config_file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
	}
	json.Unmarshal(config_file, &Conf)
	// Choosing the migrations directory based on the database used.
	Conf.MigrationsPath = Conf.MigrationsPath + Conf.DBName
}
