package controllers

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
	"github.com/stretchr/testify/suite"
)

// ControllersSuite is a suite of tests to cover API related functions
type ControllersSuite struct {
	suite.Suite
	apiKey      string
	config      *config.Config
	adminServer *httptest.Server
	phishServer *httptest.Server
}

func (s *ControllersSuite) SetupSuite() {
	conf := &config.Config{
		DBName:         "sqlite3",
		DBPath:         ":memory:",
		MigrationsPath: "../db/db_sqlite3/migrations/",
	}
	err := models.Setup(conf)
	if err != nil {
		s.T().Fatalf("Failed creating database: %v", err)
	}
	s.config = conf
	s.Nil(err)
	// Setup the admin server for use in testing
	s.adminServer = httptest.NewUnstartedServer(NewAdminServer(s.config.AdminConf).server.Handler)
	s.adminServer.Config.Addr = s.config.AdminConf.ListenURL
	s.adminServer.Start()
	// Get the API key to use for these tests
	u, err := models.GetUser(1)
	s.Nil(err)
	s.apiKey = u.ApiKey
	// Start the phishing server
	s.phishServer = httptest.NewUnstartedServer(NewPhishingServer(s.config.PhishConf).server.Handler)
	s.phishServer.Config.Addr = s.config.PhishConf.ListenURL
	s.phishServer.Start()
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
	c.UpdateStatus(models.CampaignEmailsSent)
}

func (s *ControllersSuite) TearDownSuite() {
	// Tear down the admin and phishing servers
	s.adminServer.Close()
	s.phishServer.Close()
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllersSuite))
}
