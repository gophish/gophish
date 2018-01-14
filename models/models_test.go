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
	"github.com/jinzhu/gorm"
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

func (s *ModelsSuite) TestGetUserExists(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(u.Username, check.Equals, "admin")
}

func (s *ModelsSuite) TestGetUserByUsernameWithExistingUser(c *check.C) {
	u, err := GetUserByUsername("admin")
	c.Assert(err, check.Equals, nil)
	c.Assert(u.Username, check.Equals, "admin")
}

func (s *ModelsSuite) TestGetUserDoesNotExist(c *check.C) {
	u, err := GetUser(100)
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
	c.Assert(u.Username, check.Equals, "")
}

func (s *ModelsSuite) TestGetUserByAPIKeyWithExistingAPIKey(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)

	u, err = GetUserByAPIKey(u.ApiKey)
}

func (s *ModelsSuite) TestGetUserByAPIKeyWithNotExistingAPIKey(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)

	u, err = GetUserByAPIKey(u.ApiKey + "test")
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
	c.Assert(u.Username, check.Equals, "")
}

func (s *ModelsSuite) TestGetUserByUsernameWithNotExistingUser(c *check.C) {
	u, err := GetUserByUsername("test user does not exist")
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
	c.Assert(u.Username, check.Equals, "")
}

func (s *ModelsSuite) TestPutUser(c *check.C) {
	u, err := GetUser(1)
	u.Username = "admin_changed"
	err = PutUser(&u)
	c.Assert(err, check.Equals, nil)
	u, err = GetUser(1)
	c.Assert(u.Username, check.Equals, "admin_changed")
}

func (s *ModelsSuite) TestDeleteUser(c *check.C) {
	// Create user for test
	username := "test user for deleteUser"
	testUser := User{
		Username: username,
		Hash:     "test",
	}
	c.Assert(db.Save(&testUser).Error, check.Equals, nil)

	// Assert user exists
	u, err := GetUserByUsername(username)
	c.Assert(u.Username, check.Equals, username)
	c.Assert(err, check.Equals, nil)

	// Delete user
	c.Assert(DeleteUser(u.Id), check.Equals, nil)

	// Assert user does not exists
	u, err = GetUserByUsername(username)
	c.Assert(u.Username, check.Equals, "")
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestGeneratedAPIKey(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(u.ApiKey, check.Not(check.Equals), "12345678901234567890123456789012")
}

func (s *ModelsSuite) TestPostGroup(c *check.C) {
	g := Group{Name: "Test Group"}
	g.Targets = []Target{Target{Email: "test@example.com"}}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, nil)
	c.Assert(g.Name, check.Equals, "Test Group")
	c.Assert(g.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestPostGroupNoName(c *check.C) {
	g := Group{Name: ""}
	g.Targets = []Target{Target{Email: "test@example.com"}}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, ErrGroupNameNotSpecified)
}

func (s *ModelsSuite) TestPostGroupNoTargets(c *check.C) {
	g := Group{Name: "No Target Group"}
	g.Targets = []Target{}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, ErrNoTargetsSpecified)
}

func (s *ModelsSuite) TestGetGroups(c *check.C) {
	// Add groups.
	PostGroup(&Group{
		Name:    "Test Group 1",
		Targets: []Target{Target{Email: "test1@example.com"}},
		UserId:  1,
	})
	PostGroup(&Group{
		Name:    "Test Group 2",
		Targets: []Target{Target{Email: "test2@example.com"}},
		UserId:  1,
	})

	// Get groups and test result.
	groups, err := GetGroups(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(groups), check.Equals, 2)
	c.Assert(len(groups[0].Targets), check.Equals, 1)
	c.Assert(len(groups[1].Targets), check.Equals, 1)
	c.Assert(groups[0].Name, check.Equals, "Test Group 1")
	c.Assert(groups[1].Name, check.Equals, "Test Group 2")
	c.Assert(groups[0].Targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(groups[1].Targets[0].Email, check.Equals, "test2@example.com")
}

func (s *ModelsSuite) TestGetGroupsNoGroups(c *check.C) {
	groups, err := GetGroups(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(groups), check.Equals, 0)
}

func (s *ModelsSuite) TestGetGroup(c *check.C) {
	// Add group.
	originalGroup := &Group{
		Name:    "Test Group",
		Targets: []Target{Target{Email: "test@example.com"}},
		UserId:  1,
	}
	c.Assert(PostGroup(originalGroup), check.Equals, nil)

	// Get group and test result.
	group, err := GetGroup(originalGroup.Id, 1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(group.Targets), check.Equals, 1)
	c.Assert(group.Name, check.Equals, "Test Group")
	c.Assert(group.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestGetGroupNoGroups(c *check.C) {
	_, err := GetGroup(1, 1)
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestGetGroupByName(c *check.C) {
	// Add group.
	PostGroup(&Group{
		Name:    "Test Group",
		Targets: []Target{Target{Email: "test@example.com"}},
		UserId:  1,
	})

	// Get group and test result.
	group, err := GetGroupByName("Test Group", 1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(group.Targets), check.Equals, 1)
	c.Assert(group.Name, check.Equals, "Test Group")
	c.Assert(group.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestGetGroupByNameNoGroups(c *check.C) {
	_, err := GetGroupByName("Test Group", 1)
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestPutGroup(c *check.C) {
	// Add test group.
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{Email: "test1@example.com", FirstName: "First", LastName: "Example"},
		Target{Email: "test2@example.com", FirstName: "Second", LastName: "Example"},
	}
	group.UserId = 1
	PostGroup(&group)

	// Update one of group's targets.
	group.Targets[0].FirstName = "Updated"
	err := PutGroup(&group)
	c.Assert(err, check.Equals, nil)

	// Verify updated target information.
	targets, _ := GetTargets(group.Id)
	c.Assert(targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(targets[0].FirstName, check.Equals, "Updated")
	c.Assert(targets[0].LastName, check.Equals, "Example")
	c.Assert(targets[1].Email, check.Equals, "test2@example.com")
	c.Assert(targets[1].FirstName, check.Equals, "Second")
	c.Assert(targets[1].LastName, check.Equals, "Example")
}

func (s *ModelsSuite) TestPutGroupEmptyAttribute(c *check.C) {
	// Add test group.
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{Email: "test1@example.com", FirstName: "First", LastName: "Example"},
		Target{Email: "test2@example.com", FirstName: "Second", LastName: "Example"},
	}
	group.UserId = 1
	PostGroup(&group)

	// Update one of group's targets.
	group.Targets[0].FirstName = ""
	err := PutGroup(&group)
	c.Assert(err, check.Equals, nil)

	// Verify updated empty attribute was saved.
	targets, _ := GetTargets(group.Id)
	c.Assert(targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(targets[0].FirstName, check.Equals, "")
	c.Assert(targets[0].LastName, check.Equals, "Example")
	c.Assert(targets[1].Email, check.Equals, "test2@example.com")
	c.Assert(targets[1].FirstName, check.Equals, "Second")
	c.Assert(targets[1].LastName, check.Equals, "Example")
}

func (s *ModelsSuite) TestDeleteGroupsByUserId(c *check.C) {
	// Add groups.
	testGroup1 := &Group{
		Name:    "Test Group 1",
		Targets: []Target{Target{Email: "test@example.com"}},
		UserId:  1,
	}
	testGroup2 := &Group{
		Name:    "Test Group 2",
		Targets: []Target{Target{Email: "test@example.com"}},
		UserId:  1,
	}

	c.Assert(PostGroup(testGroup1), check.Equals, nil)
	c.Assert(PostGroup(testGroup2), check.Equals, nil)

	//	Delete groups just created.
	c.Assert(DeleteGroupsByUserId(1), check.Equals, nil)

	// Assert groups are deleted.
	_, err := GetGroup(testGroup1.Id, 1)
	c.Assert(err, check.ErrorMatches, "record not found")

	_, err = GetGroup(testGroup2.Id, 1)
	c.Assert(err, check.ErrorMatches, "record not found")

	// Assert targets are deleted either.
	ts1, err := GetTargets(testGroup1.Id)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ts1), check.Equals, 0)

	ts2, err := GetTargets(testGroup2.Id)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ts2), check.Equals, 0)
}

func (s *ModelsSuite) TestDeleteCampaignsByUserId(ch *check.C) {
	// Create campaigns
	c1 := s.createCampaignDependencies(ch)
	c1.Events = []Event{
		{Email: "test1@test.com"},
		{Email: "test2@test.com"},
	}
	c1.Results = []Result{
		{Email: "test1@test.com"},
		{Email: "test2@test.com"},
	}
	ch.Assert(PostCampaign(&c1, c1.UserId), check.Equals, nil)

	c2 := s.createCampaignDependencies(ch)
	c2.Events = []Event{
		{Email: "test1@test.com"},
		{Email: "test2@test.com"},
	}
	c2.Results = []Result{
		{Email: "test1@test.com"},
		{Email: "test2@test.com"},
	}
	ch.Assert(PostCampaign(&c2, c2.UserId), check.Equals, nil)

	// Assert Campaigns are created
	cs, err := GetCampaigns(1)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(len(cs), check.Equals, 2)

	// Assert Events are created
	var es []Event
	err = db.Where("campaign_id in (?)", []int64{c1.Id, c2.Id}).Find(&es).Error
	ch.Assert(err, check.Equals, nil)
	ch.Assert(len(es), check.Not(check.Equals), 0)

	// Assert Results are created
	var rs []Result
	err = db.Where("campaign_id in (?)", []int64{c1.Id, c2.Id}).Find(&rs).Error
	ch.Assert(err, check.Equals, nil)
	ch.Assert(len(rs), check.Not(check.Equals), 0)

	DeleteCampaignsByUserId(1)

	// Assert Campaigns are deleted
	cs, err = GetCampaigns(1)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(len(cs), check.Equals, 0)

	// Assert Events are deleted
	err = db.Where("campaign_id in (?)", []int64{c1.Id, c2.Id}).Find(&es).Error
	ch.Assert(err, check.Equals, nil)
	ch.Assert(len(es), check.Equals, 0)

	// Assert Results are deleted
	err = db.Where("campaign_id in (?)", []int64{c1.Id, c2.Id}).Find(&rs).Error
	ch.Assert(err, check.Equals, nil)
	ch.Assert(len(rs), check.Equals, 0)
}

func (s *ModelsSuite) TestPostSMTP(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
	}
	err = PostSMTP(&smtp)
	c.Assert(err, check.Equals, nil)
	ss, err := GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 1)
}

func (s *ModelsSuite) TestPostSMTPNoHost(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
	}
	err = PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrHostNotSpecified)
}

func (s *ModelsSuite) TestPostSMTPNoFrom(c *check.C) {
	smtp := SMTP{
		Name:   "Test SMTP",
		UserId: 1,
		Host:   "1.1.1.1:25",
	}
	err = PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrFromAddressNotSpecified)
}

func (s *ModelsSuite) TestPostSMTPValidHeader(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
		Headers: []Header{
			Header{Key: "Reply-To", Value: "test@example.com"},
			Header{Key: "X-Mailer", Value: "gophish"},
		},
	}
	err = PostSMTP(&smtp)
	c.Assert(err, check.Equals, nil)
	ss, err := GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 1)
}

func (s *ModelsSuite) TestDeleteSMTPsByUserId(c *check.C) {
	smtp1 := SMTP{
		Name:        "Test SMTP 1",
		Host:        "1.1.1.1:25",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
		Headers: []Header{
			{Key: "Reply-To", Value: "test@example.com"},
			{Key: "X-Mailer", Value: "gophish"},
		},
	}
	smtp2 := SMTP{
		Name:        "Test SMTP 2",
		Host:        "1.1.1.1:25",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
		Headers: []Header{
			{Key: "Reply-To", Value: "test@example.com"},
			{Key: "X-Mailer", Value: "gophish"},
		},
	}

	c.Assert(PostSMTP(&smtp1), check.Equals, nil)
	c.Assert(PostSMTP(&smtp2), check.Equals, nil)

	// Assert SMTPs are successfully created
	ss, err := GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 2)

	// Assert Headers are successfully created
	var hs []Header
	err = db.Where("smtp_id in (?)", []int64{smtp1.Id, smtp2.Id}).Find(&hs).Error
	c.Assert(err, check.Equals, nil)
	c.Assert(len(hs), check.Equals, 4)

	c.Assert(DeleteSMTPsByUserId(1), check.Equals, nil)

	// Assert SMTPs are successfully deleted
	ss, err = GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 0)

	// Assert Headers are successfully deleted
	err = db.Where("smtp_id in (?)", []int64{smtp1.Id, smtp2.Id}).Find(&hs).Error
	c.Assert(err, check.Equals, nil)
	c.Assert(len(hs), check.Equals, 0)
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

func (s *ModelsSuite) TestDeletePagesByUserId(c *check.C) {
	// Create pages
	p1 := Page{
		Name:        "Test Page 1",
		HTML:        "<html></html>",
		RedirectURL: "http://example.com",
		UserId:      1,
	}
	p2 := Page{
		Name:        "Test Page 2",
		HTML:        "<html></html>",
		RedirectURL: "http://example.com",
		UserId:      1,
	}

	c.Assert(PostPage(&p1), check.Equals, nil)
	c.Assert(PostPage(&p2), check.Equals, nil)

	// Assert pages are successfully created
	ps, err := GetPages(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ps), check.Equals, 2)

	// Delete pages
	c.Assert(DeletePagesByUserId(1), check.Equals, nil)

	// Assert pages are successfully deleted
	ps, err = GetPages(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ps), check.Equals, 0)

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

func (s *ModelsSuite) TestDeleteTemplatesByUserId(c *check.C) {
	template1 := Template{
		Name:   "Test Template 1",
		UserId: 2,
		HTML:   "<body></body>",
		Attachments: []Attachment{
			{Name: "Some attachment"},
			{Name: "Some attachment"},
		},
	}
	template2 := Template{
		Name:   "Test Template 2",
		UserId: 2,
		HTML:   "<body></body>",
		Attachments: []Attachment{
			{Name: "Some attachment"},
			{Name: "Some attachment"},
		},
	}

	c.Assert(PostTemplate(&template1), check.Equals, nil)
	c.Assert(PostTemplate(&template2), check.Equals, nil)

	// Assert Templates are successfully created
	ts, err := GetTemplates(2)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ts), check.Equals, 2)

	// Assert Attachments are successfully created
	var as []Attachment
	err = db.Where("template_id in (?)", []int64{template1.Id, template2.Id}).Find(&as).Error
	c.Assert(err, check.Equals, nil)
	c.Assert(len(as), check.Equals, 4)

	c.Assert(DeleteTemplatesByUserId(2), check.Equals, nil)

	// Assert Templates are successfully deleted
	ts, err = GetTemplates(2)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ts), check.Equals, 0)

	// Assert Attachments are successfully deleted
	err = db.Where("template_id in (?)", []int64{template1.Id, template2.Id}).Find(&as).Error
	c.Assert(err, check.Equals, nil)
	c.Assert(len(as), check.Equals, 0)
}
