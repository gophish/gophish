package models

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/check.v1"
)

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

func (s *ModelsSuite) TestGeneratedAPIKey(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(u.ApiKey, check.Not(check.Equals), "12345678901234567890123456789012")
}
