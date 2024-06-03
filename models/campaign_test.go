package models

import (
	"fmt"
	"testing"
	"time"

	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) TestGenerateSendDate(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	// Test that if no launch date is provided, the campaign's creation date
	// is used.
	err := PostCampaign(&campaign, campaign.UserId)
	c.Assert(err, check.Equals, nil)
	c.Assert(campaign.LaunchDate, check.Equals, campaign.CreatedDate)

	// For comparing the dates, we need to fetch the campaign again. This is
	// to solve an issue where the campaign object right now has time down to
	// the microsecond, while in MySQL it's rounded down to the second.
	campaign, _ = GetCampaign(campaign.Id, campaign.UserId)

	ms, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)
	for _, m := range ms {
		c.Assert(m.SendDate, check.Equals, campaign.CreatedDate)
	}

	// Test that if no send date is provided, all the emails are sent at the
	// campaign's launch date
	campaign = s.createCampaignDependencies(c)
	campaign.LaunchDate = time.Now().UTC()
	err = PostCampaign(&campaign, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	campaign, _ = GetCampaign(campaign.Id, campaign.UserId)

	ms, err = GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)
	for _, m := range ms {
		c.Assert(m.SendDate, check.Equals, campaign.LaunchDate)
	}

	// Finally, test that if a send date is provided, the emails are staggered
	// correctly.
	campaign = s.createCampaignDependencies(c)
	campaign.LaunchDate = time.Now().UTC()
	campaign.SendByDate = campaign.LaunchDate.Add(2 * time.Minute)
	err = PostCampaign(&campaign, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	campaign, _ = GetCampaign(campaign.Id, campaign.UserId)

	ms, err = GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)
	sendingOffset := 2 / float64(len(ms))
	for i, m := range ms {
		expectedOffset := int(sendingOffset * float64(i))
		expectedDate := campaign.LaunchDate.Add(time.Duration(expectedOffset) * time.Minute)
		c.Assert(m.SendDate, check.Equals, expectedDate)
	}
}

func (s *ModelsSuite) TestCampaignDateValidation(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	// If both are zero, then the campaign should start immediately with no
	// send by date
	err := campaign.Validate()
	c.Assert(err, check.Equals, nil)

	// If the launch date is specified, then the send date is optional
	campaign = s.createCampaignDependencies(c)
	campaign.LaunchDate = time.Now().UTC()
	err = campaign.Validate()
	c.Assert(err, check.Equals, nil)

	// If the send date is greater than the launch date, then there's no
	//problem
	campaign = s.createCampaignDependencies(c)
	campaign.LaunchDate = time.Now().UTC()
	campaign.SendByDate = campaign.LaunchDate.Add(1 * time.Minute)
	err = campaign.Validate()
	c.Assert(err, check.Equals, nil)

	// If the send date is less than the launch date, then there's an issue
	campaign = s.createCampaignDependencies(c)
	campaign.LaunchDate = time.Now().UTC()
	campaign.SendByDate = campaign.LaunchDate.Add(-1 * time.Minute)
	err = campaign.Validate()
	c.Assert(err, check.Equals, ErrInvalidSendByDate)
}

func (s *ModelsSuite) TestLaunchCampaignMaillogStatus(c *check.C) {
	// For the first test, ensure that campaigns created with the zero date
	// (and therefore are set to launch immediately) have maillogs that are
	// locked to prevent race conditions.
	campaign := s.createCampaign(c)
	ms, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)

	for _, m := range ms {
		c.Assert(m.Processing, check.Equals, true)
	}

	// Next, verify that campaigns scheduled in the future do not lock the
	// maillogs so that they can be picked up by the background worker.
	campaign = s.createCampaignDependencies(c)
	campaign.Name = "New Campaign"
	campaign.LaunchDate = time.Now().Add(1 * time.Hour)
	c.Assert(PostCampaign(&campaign, campaign.UserId), check.Equals, nil)
	ms, err = GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)

	for _, m := range ms {
		c.Assert(m.Processing, check.Equals, false)
	}
}

func (s *ModelsSuite) TestDeleteCampaignAlsoDeletesMailLogs(c *check.C) {
	campaign := s.createCampaign(c)
	ms, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ms), check.Equals, len(campaign.Results))

	err = DeleteCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)

	ms, err = GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ms), check.Equals, 0)
}

func (s *ModelsSuite) TestCompleteCampaignAlsoDeletesMailLogs(c *check.C) {
	campaign := s.createCampaign(c)
	ms, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ms), check.Equals, len(campaign.Results))

	err = CompleteCampaign(campaign.Id, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	ms, err = GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ms), check.Equals, 0)
}

func (s *ModelsSuite) TestCampaignGetResults(c *check.C) {
	campaign := s.createCampaign(c)
	got, err := GetCampaign(campaign.Id, campaign.UserId)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(campaign.Results), check.Equals, len(got.Results))
}

func setupCampaignDependencies(b *testing.B, size int) {
	group := Group{Name: "Test Group"}
	// Create a large group of 5000 members
	for i := 0; i < size; i++ {
		group.Targets = append(group.Targets, Target{BaseRecipient: BaseRecipient{Email: fmt.Sprintf("test%d@example.com", i), FirstName: "User", LastName: fmt.Sprintf("%d", i)}})
	}
	group.UserId = 1
	err := PostGroup(&group)
	if err != nil {
		b.Fatalf("error posting group: %v", err)
	}

	// Add a template
	template := Template{Name: "Test Template"}
	template.Subject = "{{.RId}} - Subject"
	template.Text = "{{.RId}} - Text"
	template.HTML = "{{.RId}} - HTML"
	template.UserId = 1
	err = PostTemplate(&template)
	if err != nil {
		b.Fatalf("error posting template: %v", err)
	}

	// Add a landing page
	p := Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	err = PostPage(&p)
	if err != nil {
		b.Fatalf("error posting page: %v", err)
	}

	// Add a sending profile
	smtp := SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	err = PostSMTP(&smtp)
	if err != nil {
		b.Fatalf("error posting smtp: %v", err)
	}
}

// setupCampaign sets up the campaign dependencies as well as posting the
// actual campaign
func setupCampaign(b *testing.B, size int) Campaign {
	setupCampaignDependencies(b, size)
	campaign := Campaign{Name: "Test campaign"}
	campaign.UserId = 1
	campaign.Template = Template{Name: "Test Template"}
	campaign.Page = Page{Name: "Test Page"}
	campaign.SMTP = SMTP{Name: "Test Page"}
	campaign.Groups = []Group{Group{Name: "Test Group"}}
	PostCampaign(&campaign, 1)
	return campaign
}

func BenchmarkCampaign100(b *testing.B) {
	setupBenchmark(b)
	setupCampaignDependencies(b, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		campaign := Campaign{Name: "Test campaign"}
		campaign.UserId = 1
		campaign.Template = Template{Name: "Test Template"}
		campaign.Page = Page{Name: "Test Page"}
		campaign.SMTP = SMTP{Name: "Test Page"}
		campaign.Groups = []Group{Group{Name: "Test Group"}}

		b.StartTimer()
		err := PostCampaign(&campaign, 1)
		if err != nil {
			b.Fatalf("error posting campaign: %v", err)
		}
		b.StopTimer()
		db.Delete(Result{})
		db.Delete(MailLog{})
		db.Delete(Campaign{})
	}
	tearDownBenchmark(b)
}

func BenchmarkCampaign1000(b *testing.B) {
	setupBenchmark(b)
	setupCampaignDependencies(b, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		campaign := Campaign{Name: "Test campaign"}
		campaign.UserId = 1
		campaign.Template = Template{Name: "Test Template"}
		campaign.Page = Page{Name: "Test Page"}
		campaign.SMTP = SMTP{Name: "Test Page"}
		campaign.Groups = []Group{Group{Name: "Test Group"}}

		b.StartTimer()
		err := PostCampaign(&campaign, 1)
		if err != nil {
			b.Fatalf("error posting campaign: %v", err)
		}
		b.StopTimer()
		db.Delete(Result{})
		db.Delete(MailLog{})
		db.Delete(Campaign{})
	}
	tearDownBenchmark(b)
}

func BenchmarkCampaign10000(b *testing.B) {
	setupBenchmark(b)
	setupCampaignDependencies(b, 10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		campaign := Campaign{Name: "Test campaign"}
		campaign.UserId = 1
		campaign.Template = Template{Name: "Test Template"}
		campaign.Page = Page{Name: "Test Page"}
		campaign.SMTP = SMTP{Name: "Test Page"}
		campaign.Groups = []Group{Group{Name: "Test Group"}}

		b.StartTimer()
		err := PostCampaign(&campaign, 1)
		if err != nil {
			b.Fatalf("error posting campaign: %v", err)
		}
		b.StopTimer()
		db.Delete(Result{})
		db.Delete(MailLog{})
		db.Delete(Campaign{})
	}
	tearDownBenchmark(b)
}

func BenchmarkGetCampaign100(b *testing.B) {
	setupBenchmark(b)
	campaign := setupCampaign(b, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCampaign(campaign.Id, campaign.UserId)
		if err != nil {
			b.Fatalf("error getting campaign: %v", err)
		}
	}
	tearDownBenchmark(b)
}

func BenchmarkGetCampaign1000(b *testing.B) {
	setupBenchmark(b)
	campaign := setupCampaign(b, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCampaign(campaign.Id, campaign.UserId)
		if err != nil {
			b.Fatalf("error getting campaign: %v", err)
		}
	}
	tearDownBenchmark(b)
}

func BenchmarkGetCampaign5000(b *testing.B) {
	setupBenchmark(b)
	campaign := setupCampaign(b, 5000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCampaign(campaign.Id, campaign.UserId)
		if err != nil {
			b.Fatalf("error getting campaign: %v", err)
		}
	}
	tearDownBenchmark(b)
}

func BenchmarkGetCampaign10000(b *testing.B) {
	setupBenchmark(b)
	campaign := setupCampaign(b, 10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCampaign(campaign.Id, campaign.UserId)
		if err != nil {
			b.Fatalf("error getting campaign: %v", err)
		}
	}
	tearDownBenchmark(b)
}
