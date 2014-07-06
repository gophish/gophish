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

// Config represents the configuration information.
type Config struct {
	AdminURL string     `json:"admin_url"`
	PhishURL string     `json:"phish_url"`
	SMTP     SMTPServer `json:"smtp"`
	DBPath   string     `json:"dbpath"`
}

var Conf Config

func init() {
	// Get the config file
	config_file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
	}
	json.Unmarshal(config_file, &Conf)
}
