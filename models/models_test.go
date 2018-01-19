package models

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gophish/gophish/config"
	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type ModelsSuite struct{}

var _ = check.Suite(&ModelsSuite{})

func (s *ModelsSuite) SetUpSuite(c *check.C) {
	config.Conf.DBName = "sqlite3"
	config.Conf.DBPath = ":memory:"
	config.Conf.MigrationsPath = "../db/db_sqlite3/migrations/"
	err := Setup()
	if err != nil {
		c.Fatalf("Failed creating database: %v", err)
	}
}

func (s *ModelsSuite) TearDownTest(c *check.C) {
	// Clear database tables between each test. If new tables are
	// used in this test suite they will need to be cleaned up here.
	db.Delete(Group{})
	db.Delete(Target{})
	db.Delete(GroupTarget{})
	db.Delete(SMTP{})
	db.Delete(Page{})
	db.Delete(Result{})
	db.Delete(MailLog{})
	db.Delete(Campaign{})

	// Reset users table to default state.
	db.Not("id", 1).Delete(User{})
	db.Model(User{}).Update("username", "admin")
}

func (s *ModelsSuite) createCampaignDependencies(ch *check.C) Campaign {
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{Email: "test1@example.com", FirstName: "First", LastName: "Example"},
		Target{Email: "test2@example.com", FirstName: "Second", LastName: "Example"},
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
	return c
}

func (s *ModelsSuite) createCampaign(ch *check.C) Campaign {
	c := s.createCampaignDependencies(ch)
	// Setup and "launch" our campaign
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	return c
}

func (s *ModelsSuite) TestPostPage(c *check.C) {
	html := `<html>
			<head></head>
			<body><form action="example.com">
				<input name="username"/>
				<input name="password" type="password"/>
			</form></body>
		  </html>`
	p := Page{
		Name:        "Test Page",
		HTML:        html,
		RedirectURL: "http://example.com",
	}
	// Check the capturing credentials and passwords
	p.CaptureCredentials = true
	p.CapturePasswords = true
	err := PostPage(&p)
	c.Assert(err, check.Equals, nil)
	c.Assert(p.RedirectURL, check.Equals, "http://example.com")
	d, err := goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms := d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password still has a name
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, true)
		// Check the username is still correct
		u, ok := f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, true)
		c.Assert(u, check.Equals, "username")
	})
	// Check what happens when we don't capture passwords
	p.CapturePasswords = false
	p.HTML = html
	p.RedirectURL = ""
	err = PutPage(&p)
	c.Assert(err, check.Equals, nil)
	c.Assert(p.RedirectURL, check.Equals, "")
	d, err = goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms = d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password still has a name
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, false)
		// Check the username is still correct
		u, ok := f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, true)
		c.Assert(u, check.Equals, "username")
	})
	// Finally, check when we don't capture credentials
	p.CaptureCredentials = false
	p.HTML = html
	err = PutPage(&p)
	c.Assert(err, check.Equals, nil)
	d, err = goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms = d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password still has a name
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, false)
		// Check the username is still correct
		_, ok = f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, false)
	})
}

func (s *ModelsSuite) TestGenerateResultId(c *check.C) {
	r := Result{}
	r.GenerateId()
	match, err := regexp.Match("[a-zA-Z0-9]{7}", []byte(r.RId))
	c.Assert(err, check.Equals, nil)
	c.Assert(match, check.Equals, true)
}

func (s *ModelsSuite) TestFormatAddress(c *check.C) {
	r := Result{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
	}
	expected := &mail.Address{
		Name:    "John Doe",
		Address: "johndoe@example.com",
	}
	c.Assert(r.FormatAddress(), check.Equals, expected.String())

	r = Result{
		Email: "johndoe@example.com",
	}
	c.Assert(r.FormatAddress(), check.Equals, r.Email)
}

func (s *ModelsSuite) TestResultSendingStatus(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	// This campaign wasn't scheduled, so we expect the status to
	// be sending
	fmt.Println("Campaign STATUS")
	fmt.Println(c.Status)
	for _, r := range c.Results {
		ch.Assert(r.Status, check.Equals, STATUS_SENDING)
	}
}
func (s *ModelsSuite) TestResultScheduledStatus(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	c.LaunchDate = time.Now().UTC().Add(time.Hour * time.Duration(1))
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	// This campaign wasn't scheduled, so we expect the status to
	// be sending
	for _, r := range c.Results {
		ch.Assert(r.Status, check.Equals, STATUS_SCHEDULED)
	}
}

func (s *ModelsSuite) TestDuplicateResults(ch *check.C) {
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{Email: "test1@example.com", FirstName: "First", LastName: "Example"},
		Target{Email: "test1@example.com", FirstName: "Duplicate", LastName: "Duplicate"},
		Target{Email: "test2@example.com", FirstName: "Second", LastName: "Example"},
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
