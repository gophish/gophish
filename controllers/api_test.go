package controllers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/handlers"
	"github.com/jordan-wright/gophish/config"
	"github.com/jordan-wright/gophish/models"
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
	config.Conf.DBPath = ":memory:"
	err := models.Setup()
	if err != nil {
		s.T().Fatalf("Failed creating database: %v", err)
	}
	s.Nil(err)
	// Setup the admin server for use in testing
	as.Config.Addr = config.Conf.AdminURL
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
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	fmt.Printf("%s", body)
}

func (s *ControllersSuite) TearDownSuite() {
	// Tear down the admin server
	as.Close()
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllersSuite))
}
