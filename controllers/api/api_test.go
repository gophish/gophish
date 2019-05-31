package api

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
	"github.com/stretchr/testify/suite"
)

type APISuite struct {
	suite.Suite
	apiKey    string
	config    *config.Config
	apiServer *Server
	admin     models.User
}

func (s *APISuite) SetupSuite() {
	conf := &config.Config{
		DBName:         "sqlite3",
		DBPath:         ":memory:",
		MigrationsPath: "../../db/db_sqlite3/migrations/",
	}
	err := models.Setup(conf)
	if err != nil {
		s.T().Fatalf("Failed creating database: %v", err)
	}
	s.config = conf
	s.Nil(err)
	// Get the API key to use for these tests
	u, err := models.GetUser(1)
	s.Nil(err)
	s.apiKey = u.ApiKey
	s.admin = u
	// Move our cwd up to the project root for help with resolving
	// static assets
	err = os.Chdir("../")
	s.Nil(err)
	s.apiServer = NewServer()
}

func (s *APISuite) TearDownTest() {
	campaigns, _ := models.GetCampaigns(1)
	for _, campaign := range campaigns {
		models.DeleteCampaign(campaign.Id)
	}
	// Cleanup all users except the original admin
	users, _ := models.GetUsers()
	for _, user := range users {
		if user.Id == 1 {
			continue
		}
		err := models.DeleteUser(user.Id)
		s.Nil(err)
	}
}

func (s *APISuite) SetupTest() {
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

func (s *APISuite) TestSiteImportBaseHref() {
	h := "<html><head></head><body><img src=\"/test.png\"/></body></html>"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, h)
	}))
	hr := fmt.Sprintf("<html><head><base href=\"%s\"/></head><body><img src=\"/test.png\"/>\n</body></html>", ts.URL)
	defer ts.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/import/site",
		bytes.NewBuffer([]byte(fmt.Sprintf(`
			{
				"url" : "%s",
				"include_resources" : false
			}
		`, ts.URL))))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	s.apiServer.ImportSite(response, req)
	cs := cloneResponse{}
	err := json.NewDecoder(response.Body).Decode(&cs)
	s.Nil(err)
	s.Equal(cs.HTML, hr)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}
