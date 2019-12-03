package models

import (
	"net/mail"
	"regexp"
	"time"

	"gopkg.in/check.v1"
)

func (s *ModelsSuite) TestGenerateResultId(c *check.C) {
	r := Result{}
	r.GenerateId(db)
	match, err := regexp.Match("[a-zA-Z0-9]{7}", []byte(r.RId))
	c.Assert(err, check.Equals, nil)
	c.Assert(match, check.Equals, true)
}

func (s *ModelsSuite) TestFormatAddress(c *check.C) {
	r := Result{
		BaseRecipient: BaseRecipient{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "johndoe@example.com",
		},
	}
	expected := &mail.Address{
		Name:    "John Doe",
		Address: "johndoe@example.com",
	}
	c.Assert(r.FormatAddress(), check.Equals, expected.String())

	r = Result{
		BaseRecipient: BaseRecipient{Email: "johndoe@example.com"},
	}
	c.Assert(r.FormatAddress(), check.Equals, r.Email)
}

func (s *ModelsSuite) TestResultSendingStatus(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	// This campaign wasn't scheduled, so we expect the status to
	// be sending
	for _, r := range c.Results {
		ch.Assert(r.Status, check.Equals, StatusSending)
		ch.Assert(r.ModifiedDate, check.Equals, c.CreatedDate)
	}
}
func (s *ModelsSuite) TestResultScheduledStatus(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	c.LaunchDate = time.Now().UTC().Add(time.Hour * time.Duration(1))
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	// This campaign wasn't scheduled, so we expect the status to
	// be sending
	for _, r := range c.Results {
		ch.Assert(r.Status, check.Equals, StatusScheduled)
		ch.Assert(r.ModifiedDate, check.Equals, c.CreatedDate)
	}
}

func (s *ModelsSuite) TestResultVariableStatus(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	c.LaunchDate = time.Now().UTC()
	c.SendByDate = c.LaunchDate.Add(2 * time.Minute)
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)

	// The campaign has a window smaller than our group size, so we expect some
	// emails to be sent immediately, while others will be scheduled
	for _, r := range c.Results {
		if r.SendDate.Before(c.CreatedDate) || r.SendDate.Equal(c.CreatedDate) {
			ch.Assert(r.Status, check.Equals, StatusSending)
		} else {
			ch.Assert(r.Status, check.Equals, StatusScheduled)
		}
	}
}

func (s *ModelsSuite) TestDuplicateResults(ch *check.C) {
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{BaseRecipient: BaseRecipient{Email: "test1@example.com", FirstName: "First", LastName: "Example"}},
		Target{BaseRecipient: BaseRecipient{Email: "test1@example.com", FirstName: "Duplicate", LastName: "Duplicate"}},
		Target{BaseRecipient: BaseRecipient{Email: "test2@example.com", FirstName: "Second", LastName: "Example"}},
	}
	group.UserId = 1
	ch.Assert(PostGroup(&group), check.Equals, nil)

	// Add a template
	t := Template{Name: "Test Template"}
	t.Subject = "{{.RId}} - Subject"
	t.Text = "{{.RId}} - Text"
	t.HTML = "{{.RId}} - HTML"
	t.UserId = 1
	ch.Assert(PostTemplate(&t), check.Equals, nil)

	// Add a landing page
	p := Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	ch.Assert(PostPage(&p), check.Equals, nil)

	// Add a sending profile
	smtp := SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	ch.Assert(PostSMTP(&smtp), check.Equals, nil)

	c := Campaign{Name: "Test campaign"}
	c.UserId = 1
	c.Template = t
	c.Page = p
	c.SMTP = smtp
	c.Groups = []Group{group}

	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	ch.Assert(len(c.Results), check.Equals, 2)
	ch.Assert(c.Results[0].Email, check.Equals, group.Targets[0].Email)
	ch.Assert(c.Results[1].Email, check.Equals, group.Targets[2].Email)
}
