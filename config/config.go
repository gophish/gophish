package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jordan-wright/gophish/models"
)

var Conf models.Config

func init() {
	// Get the config file
	config_file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(config_file, &Conf)
}
