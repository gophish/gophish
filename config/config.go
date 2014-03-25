package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type SMTPServer struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Config represents the configuration information.
type Config struct {
	URL    string     `json:"url"`
	SMTP   SMTPServer `json:"smtp"`
	DBPath string     `json:"dbpath"`
}

var Conf Config

func init() {
	// Get the config file
	config_file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(config_file, &Conf)
}
