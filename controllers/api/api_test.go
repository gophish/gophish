package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
)

type testContext struct {
	apiKey    string
	config    *config.Config
	apiServer *Server
	admin     models.User
}

func setupTest(t *testing.T) *testContext {
	conf := &config.Config{
		DBName:         "sqlite3",
		DBPath:         ":memory:",
		MigrationsPath: "../../db/db_sqlite3/migrations/",
	}
	err := models.Setup(conf)
	if err != nil {
		t.Fatalf("Failed creating database: %v", err)
	}
	ctx := &testContext{}
	ctx.config = conf
	// Get the API key to use for these tests
	u, err := models.GetUser(1)
	if err != nil {
		t.Fatalf("error getting admin user: %v", err)
	}
	ctx.apiKey = u.ApiKey
	ctx.admin = u
	ctx.apiServer = NewServer()
	return ctx
}

func createTestData(t *testing.T) {
	// Add a group
	group := models.Group{Name: "Test Group"}
	group.Targets = []models.Target{
		models.Target{BaseRecipient: models.BaseRecipient{Email: "test1@example.com", FirstName: "First", LastName: "Example"}},
		models.Target{BaseRecipient: models.BaseRecipient{Email: "test2@example.com", FirstName: "Second", LastName: "Example"}},
	}
	group.UserId = 1
	models.PostGroup(&group)

	// Add a template
	template := models.Template{Name: "Test Template"}
	template.Subject = "Test subject"
	template.Text = "Text text"
	template.HTML = "<html>Test</html>"
	template.UserId = 1
	models.PostTemplate(&template)

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
	c.Template = template
	c.Page = p
	c.SMTP = smtp
	c.Groups = []models.Group{group}
	models.PostCampaign(&c, c.UserId)
	c.UpdateStatus(models.CampaignEmailsSent)
}

func TestSiteImportBaseHref(t *testing.T) {
	ctx := setupTest(t)
	h := "<html><head></head><body><img src=\"/test.png\"/></body></html>"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, h)
	}))
	expected := fmt.Sprintf("<html><head><base href=\"%s\"/></head><body><img src=\"/test.png\"/>\n</body></html>", ts.URL)
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
	ctx.apiServer.ImportSite(response, req)
	cs := cloneResponse{}
	err := json.NewDecoder(response.Body).Decode(&cs)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	if cs.HTML != expected {
		t.Fatalf("unexpected response received. expected %s got %s", expected, cs.HTML)
	}
}

func TestResendCampaign(t *testing.T) {
	ctx := setupTest(t)
	createTestData(t)

	t.Run("Test ResendAll Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/campaigns/1/resendall", nil)
		req.Header.Set("Authorization", "Bearer "+ctx.apiKey)

		rr := httptest.NewRecorder()
		ctx.apiServer.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		count, _ := models.CountMailLogs(1)
		assert.Equal(t, int64(4), count, "Expected 4 total mail logs after resend")
	})

	t.Run("Test ResendAll Authorization Failure", func(t *testing.T) {
		otherUser := models.User{Username: "other", Role: models.Role{Name: models.RoleUser}}
		models.PutUser(&otherUser)
		otherCampaign := models.Campaign{Name: "Other Campaign", UserId: otherUser.Id}
		models.PostCampaign(&otherCampaign, otherUser.Id)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/campaigns/%d/resendall", otherCampaign.Id), nil)
		req.Header.Set("Authorization", "Bearer "+ctx.apiKey)

		rr := httptest.NewRecorder()
		ctx.apiServer.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestResendResult(t *testing.T) {
	ctx := setupTest(t)
	createTestData(t)

	t.Run("Test Resend Single Result Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/results/1/resend", nil)
		req.Header.Set("Authorization", "Bearer "+ctx.apiKey)

		rr := httptest.NewRecorder()
		ctx.apiServer.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected Status OK")

		count, _ := models.CountMailLogs(1)
		assert.Equal(t, int64(3), count, "Expected 3 total mail logs after single resend")
	})

	t.Run("Test Resend Single Result Authorization Failure", func(t *testing.T) {
		otherUser := models.User{Username: "other2", Role: models.Role{Name: models.RoleUser}}
		models.PutUser(&otherUser)

		otherGroup := models.Group{Name: "Other Group", UserId: otherUser.Id}
		otherGroup.Targets = []models.Target{{BaseRecipient: models.BaseRecipient{Email: "someone@else.com"}}}
		models.PostGroup(&otherGroup)

		otherCampaign := models.Campaign{
			Name:     "Other Campaign 2",
			UserId:   otherUser.Id,
			Groups:   []models.Group{otherGroup},
			Template: models.Template{Id: 1},
			Page:     models.Page{Id: 1},
			SMTP:     models.SMTP{Id: 1},
		}
		models.PostCampaign(&otherCampaign, otherUser.Id)

		resultToTest, _ := models.GetFirstResultForCampaign(otherCampaign.Id)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/results/%d/resend", resultToTest.Id), nil)
		req.Header.Set("Authorization", "Bearer "+ctx.apiKey)

		rr := httptest.NewRecorder()
		ctx.apiServer.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
