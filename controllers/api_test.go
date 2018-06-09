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

// ps is the Phishing Server
var ps *httptest.Server = httptest.NewUnstartedServer(handlers.CombinedLoggingHandler(os.Stdout, CreatePhishingRouter()))

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
	// Start the phishing server
	ps.Config.Addr = config.Conf.PhishConf.ListenURL
	ps.Start()
	// Move our cwd up to the project root for help with resolving
	// static assets
	err = os.Chdir("../")
	s.Nil(err)
}

func (s *ControllersSuite) TearDownTest() {
	campaigns, _ := models.GetCampaigns(1)
	for _, campaign := range campaigns {
		models.DeleteCampaign(campaign.Id)
	}
}

func (s *ControllersSuite) SetupTest() {
	// Add a group
	group := models.Group{Name: "Test Group"}
	group.Targets = []models.Target{
		models.Target{BaseRecipient: models.BaseRecipient{Email: "test1@example.com", FirstName: "First", LastName: "Example"}},
		models.Target{BaseRecipient: models.BaseRecipient{Email: "test2@example.com", FirstName: "Second", LastName: "Example"}},
	}
	group.UserId = 1
	models.PostGroup(&group)

	// Add a template
	t := models.Template{Name: "Test Template"}
	t.Subject = "Test subject"
	t.Text = "Text text"
	t.HTML = "<html>Test</html>"
	t.UserId = 1
	models.PostTemplate(&t)

	// Add a landing page
	p := models.Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	models.PostPage(&p)

	// Add a sending profile
	smtp := models.SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	models.PostSMTP(&smtp)

	// Setup and "launch" our campaign
	// Set the status such that no emails are attempted
	c := models.Campaign{Name: "Test campaign"}
	c.UserId = 1
	c.Template = t
	c.Page = p
	c.SMTP = smtp
	c.Groups = []models.Group{group}
	models.PostCampaign(&c, c.UserId)
	c.UpdateStatus(models.CAMPAIGN_EMAILS_SENT)
}

func (s *ControllersSuite) TestRequireAPIKey() {
	resp, err := http.Post(fmt.Sprintf("%s/api/import/site", as.URL), "application/json", nil)
	s.Nil(err)
	defer resp.Body.Close()
	s.Equal(resp.StatusCode, http.StatusBadRequest)
}

func (s *ControllersSuite) TestInvalidAPIKey() {
	resp, err := http.Get(fmt.Sprintf("%s/api/groups/?api_key=%s", as.URL, "bogus-api-key"))
	s.Nil(err)
	defer resp.Body.Close()
	s.Equal(resp.StatusCode, http.StatusBadRequest)
}

func (s *ControllersSuite) TestBearerToken() {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/groups/", as.URL), nil)
	s.Nil(err)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.ApiKey))
	resp, err := http.DefaultClient.Do(req)
	s.Nil(err)
	defer resp.Body.Close()
	s.Equal(resp.StatusCode, http.StatusOK)
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
	// Tear down the admin and phishing servers
	as.Close()
	ps.Close()
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllersSuite))
}
