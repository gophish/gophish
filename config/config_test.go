package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
	ConfigFile *os.File
}

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

func (s *ConfigSuite) SetupTest() {
	f, err := ioutil.TempFile("", "gophish-config")
	s.Nil(err)
	s.ConfigFile = f
}

func (s *ConfigSuite) TearDownTest() {
	err := s.ConfigFile.Close()
	s.Nil(err)
}

func (s *ConfigSuite) TestLoadConfig() {
	_, err := s.ConfigFile.Write(validConfig)
	s.Nil(err)
	// Load the valid config
	conf, err := LoadConfig(s.ConfigFile.Name())
	s.Nil(err)

	expectedConfig := &Config{}
	err = json.Unmarshal(validConfig, &expectedConfig)
	s.Nil(err)
	expectedConfig.MigrationsPath = expectedConfig.MigrationsPath + expectedConfig.DBName
	expectedConfig.TestFlag = false
	s.Equal(expectedConfig, conf)

	// Load an invalid config
	conf, err = LoadConfig("bogusfile")
	s.NotNil(err)
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
