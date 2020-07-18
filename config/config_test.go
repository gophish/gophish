package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	log "github.com/gophish/gophish/logger"
)

var validConfig = []byte(`{
	"admin_server": {
		"listen_url": "127.0.0.1:3333",
		"use_tls": true,
		"cert_path": "gophish_admin.crt",
		"key_path": "gophish_admin.key"
	},
	"phish_server": {
		"listen_url": "0.0.0.0:8080",
		"use_tls": false,
		"cert_path": "example.crt",
		"key_path": "example.key"
	},
	"db_name": "sqlite3",
	"db_path": "gophish.db",
	"migrations_prefix": "db/db_",
	"contact_address": ""
}`)

func createTemporaryConfig(t *testing.T) *os.File {
	f, err := ioutil.TempFile("", "gophish-config")
	if err != nil {
		t.Fatalf("unable to create temporary config: %v", err)
	}
	return f
}

func removeTemporaryConfig(t *testing.T, f *os.File) {
	err := f.Close()
	if err != nil {
		t.Fatalf("unable to remove temporary config: %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	_, err := f.Write(validConfig)
	if err != nil {
		t.Fatalf("error writing config to temporary file: %v", err)
	}
	// Load the valid config
	conf, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("error loading config from temporary file: %v", err)
	}

	expectedConfig := &Config{}
	err = json.Unmarshal(validConfig, &expectedConfig)
	if err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}
	expectedConfig.MigrationsPath = expectedConfig.MigrationsPath + expectedConfig.DBName
	expectedConfig.TestFlag = false
	expectedConfig.AdminConf.CSRFKey = ""
	expectedConfig.Logging = &log.Config{}
	if !reflect.DeepEqual(expectedConfig, conf) {
		t.Fatalf("invalid config received. expected %#v got %#v", expectedConfig, conf)
	}

	// Load an invalid config
	_, err = LoadConfig("bogusfile")
	if err == nil {
		t.Fatalf("expected error when loading invalid config, but got %v", err)
	}
}
