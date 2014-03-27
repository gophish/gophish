package models

import (
	"os"
	"testing"

	"github.com/jordan-wright/gophish/config"
	"launchpad.net/gocheck"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type ModelsSuite struct{}

var _ = gocheck.Suite(&ModelsSuite{})

func (s *ModelsSuite) SetUpSuite(c *gocheck.C) {
	config.Conf.DBPath = "../gophish_test.db"
	err := Setup()
	if err != nil {
		c.Fatalf("Failed creating database: %v", err)
	}
}

func (s *ModelsSuite) TestGetUser(c *gocheck.C) {
	u, err := GetUser(1)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(u.Username, gocheck.Equals, "admin")
}

func (s *ModelsSuite) TestPutUser(c *gocheck.C) {
	u, err := GetUser(1)
	u.Username = "admin_changed"
	err = PutUser(&u)
	c.Assert(err, gocheck.Equals, nil)
	u, err = GetUser(1)
	c.Assert(u.Username, gocheck.Equals, "admin_changed")
}

func (s *ModelsSuite) TearDownSuite(c *gocheck.C) {
	db.DB().Close()
	err := os.Remove(config.Conf.DBPath)
	if err != nil {
		c.Fatalf("Failed deleting test database: %v", err)
	}
}
