package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
	"github.com/stretchr/testify/suite"
)

type logMailer struct {
	queue chan []mailer.Mail
}

func (m *logMailer) Start(ctx context.Context) {
	return
}

func (m *logMailer) Queue(ms []mailer.Mail) {
	m.queue <- ms
}

// WorkerSuite is a suite of tests to cover API related functions
type WorkerSuite struct {
	suite.Suite
	config *config.Config
}

func (s *WorkerSuite) SetupSuite() {
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
	s.setupCampaignDependencies()
}

func (s *WorkerSuite) TearDownTest() {
	campaigns, _ := models.GetCampaigns(1)
	for _, campaign := range campaigns {
		models.DeleteCampaign(campaign.Id)
	}
}

func (s *WorkerSuite) setupCampaignDependencies() {
	s.config.TestFlag = true
	// Add a group
	group := models.Group{Name: "Test Group"}
	for i := 0; i < 10; i++ {
		group.Targets = append(group.Targets, models.Target{
			BaseRecipient: models.BaseRecipient{
				Email:     fmt.Sprintf("test%d@example.com", i),
				FirstName: "First",
				LastName:  "Example"}})
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
}

func (s *WorkerSuite) setupCampaign(id int) (*models.Campaign, error) {
	// Setup and "launch" our campaign
	// Set the status such that no emails are attempted
	c := models.Campaign{Name: fmt.Sprintf("Test campaign - %d", id)}
	c.UserId = 1
	template, err := models.GetTemplate(1, 1)
	if err != nil {
		return nil, err
	}
	c.Template = template

	page, err := models.GetPage(1, 1)
	if err != nil {
		return nil, err
	}
	c.Page = page

	smtp, err := models.GetSMTP(1, 1)
	if err != nil {
		return nil, err
	}
	c.SMTP = smtp

	group, err := models.GetGroup(1, 1)
	if err != nil {
		return nil, err
	}
	c.Groups = []models.Group{group}
	err = models.PostCampaign(&c, c.UserId)
	if err != nil {
		return nil, err
	}
	err = c.UpdateStatus(models.CampaignEmailsSent)
	return &c, err
}

func (s *WorkerSuite) TestMailLogGrouping() {
	// Create the campaigns and unlock the maillogs so that they're picked up
	// by the worker
	for i := 0; i < 10; i++ {
		campaign, err := s.setupCampaign(i)
		s.Nil(err)
		ms, err := models.GetMailLogsByCampaign(campaign.Id)
		s.Nil(err)
		for _, m := range ms {
			m.Unlock()
		}
	}

	lm := &logMailer{queue: make(chan []mailer.Mail)}
	worker := &DefaultWorker{}
	worker.mailer = lm

	// Trigger the worker, generating the maillogs and sending them to the
	// mailer
	worker.processCampaigns(time.Now())

	// Verify that each slice of maillogs received belong to the same campaign
	for i := 0; i < 10; i++ {
		ms := <-lm.queue
		maillog, ok := ms[0].(*models.MailLog)
		if !ok {
			s.T().Fatalf("unable to cast mail to models.MailLog")
		}
		expected := maillog.CampaignId
		for _, m := range ms {
			maillog, ok = m.(*models.MailLog)
			if !ok {
				s.T().Fatalf("unable to cast mail to models.MailLog")
			}
			got := maillog.CampaignId
			s.Equal(expected, got)
		}
	}
}

func TestMailerSuite(t *testing.T) {
	suite.Run(t, new(WorkerSuite))
}
