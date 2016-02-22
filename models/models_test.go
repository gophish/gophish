package models

import (
	"testing"

	"github.com/gophish/gophish/config"
	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type ModelsSuite struct{}

var _ = check.Suite(&ModelsSuite{})

func (s *ModelsSuite) SetUpSuite(c *check.C) {
	config.Conf.DBPath = ":memory:"
	config.Conf.MigrationsPath = "../db/migrations/"
	err := Setup()
	if err != nil {
		c.Fatalf("Failed creating database: %v", err)
	}
}

func (s *ModelsSuite) TestGetUser(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(u.Username, check.Equals, "admin")
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

func (s *ModelsSuite) TestPutUser(c *check.C) {
	u, err := GetUser(1)
	u.Username = "admin_changed"
	err = PutUser(&u)
	c.Assert(err, check.Equals, nil)
	u, err = GetUser(1)
	c.Assert(u.Username, check.Equals, "admin_changed")
}
