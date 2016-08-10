package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/handlers"
	"github.com/stretchr/testify/suite"
)

// ControllersSuite is a suite of tests to cover API related functions
type ControllersSuite struct {
	suite.Suite
	ApiKey string
}

// as is the Admin Server for our API calls
var as *httptest.Server = httptest.NewUnstartedServer(handlers.CombinedLoggingHandler(os.Stdout, CreateAdminRouter()))

func (s *ControllersSuite) SetupSuite() {
	config.Conf.DBName = "sqlite3"
	config.Conf.DBPath = ":memory:"
	config.Conf.MigrationsPath = "../db/db_sqlite3/migrations/"
	err := models.Setup()
	if err != nil {
		s.T().Fatalf("Failed creating database: %v", err)
	}
	s.Nil(err)
	// Setup the admin server for use in testing
	as.Config.Addr = config.Conf.AdminConf.ListenURL
	as.Start()
	// Get the API key to use for these tests
	u, err := models.GetUser(1)
	s.Nil(err)
	s.ApiKey = u.ApiKey
}

func (s *ControllersSuite) TestSiteImportBaseHref() {
	h := "<html><head></head><body><img src=\"/test.png\"/></body></html>"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, h)
	}))
	hr := fmt.Sprintf("<html><head><base href=\"%s\"/></head><body><img src=\"/test.png\"/>\n</body></html>", ts.URL)
	defer ts.Close()
	resp, err := http.Post(fmt.Sprintf("%s/api/import/site?api_key=%s", as.URL, s.ApiKey), "application/json",
		bytes.NewBuffer([]byte(fmt.Sprintf(`
			{
				"url" : "%s",
				"include_resources" : false
			}
		`, ts.URL))))
	s.Nil(err)
	defer resp.Body.Close()
	cs := cloneResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cs)
	s.Nil(err)
	s.Equal(cs.HTML, hr)
}

func (s *ControllersSuite) TearDownSuite() {
	// Tear down the admin server
	as.Close()
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllersSuite))
}
