package models

import (
	"gopkg.in/check.v1"
)

func (s *ModelsSuite) createCampaign() Campaign {
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{Email: "test1@example.com", FirstName: "First", LastName: "Example"},
		Target{Email: "test2@example.com", FirstName: "Second", LastName: "Example"},
	}
	group.UserId = 1
	PostGroup(&group)

	// Add a template
	t := Template{Name: "Test Template"}
	t.Subject = "Test subject"
	t.Text = "Text text"
	t.HTML = "<html>Test</html>"
	t.UserId = 1
	PostTemplate(&t)

	// Add a landing page
	p := Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	PostPage(&p)

	// Add a sending profile
	smtp := SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	PostSMTP(&smtp)

	// Setup and "launch" our campaign
	// Set the status such that no emails are attempted
	c := Campaign{Name: "Test campaign"}
	c.UserId = 1
	c.Template = t
	c.Page = p
	c.SMTP = smtp
	c.Groups = []Group{group}
	PostCampaign(&c, c.UserId)
	c.UpdateStatus(CAMPAIGN_EMAILS_SENT)
	return c
}

func (s *ModelsSuite) TestPutMailLog(c *check.C) {
	c := s.createCampaign()
	for _, r := range c.Results {
		m := MailLog{}
		err := db.Where("rid=? && campaign_id=?", r.RId, c.Id).
			Find(&m).Error
		c.Assert(err, check.Equals, nil)
		c.Assert(m.RId, check.Equals, r.RId)
		c.Assert(m.CampaignId, check.Equals, c.Id)
		c.Assert(m.SendAt, check.Equals, c.LaunchDate)
		c.Assert(m.UserId, check.Equals, 1)
		c.Assert(m.SendAttempt, check.Equals, 0)
	}
}
