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

	got, err := GetUserByAPIKey(u.ApiKey)
	c.Assert(err, check.Equals, nil)
	c.Assert(got.Id, check.Equals, u.Id)
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
	u, _ := GetUser(1)
	u.Username = "admin_changed"
	err := PutUser(&u)
	c.Assert(err, check.Equals, nil)
	u, err = GetUser(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(u.Username, check.Equals, "admin_changed")
}

func (s *ModelsSuite) TestGeneratedAPIKey(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(u.ApiKey, check.Not(check.Equals), "12345678901234567890123456789012")
}

func (s *ModelsSuite) verifyRoleCount(c *check.C, roleID, expected int64) {
	var adminCount int64
	err := db.Model(&User{}).Where("role_id=?", roleID).Count(&adminCount).Error
	c.Assert(err, check.Equals, nil)
	c.Assert(adminCount, check.Equals, expected)
}

func (s *ModelsSuite) TestDeleteLastAdmin(c *check.C) {
	// Create a new admin user
	role, err := GetRoleBySlug(RoleAdmin)
	c.Assert(err, check.Equals, nil)
	newAdmin := User{
		Username: "new-admin",
		Hash:     "123456",
		ApiKey:   "123456",
		Role:     role,
		RoleID:   role.ID,
	}
	err = PutUser(&newAdmin)
	c.Assert(err, check.Equals, nil)

	// Ensure that there are two admins
	s.verifyRoleCount(c, role.ID, 2)

	// Delete the newly created admin - this should work since we have more
	// than one current admin.
	err = DeleteUser(newAdmin.Id)
	c.Assert(err, check.Equals, nil)

	// Verify that we now have one admin
	s.verifyRoleCount(c, role.ID, 1)

	// Try to delete the last admin - this should fail since we always want at
	// least one admin active in Gophish.
	err = DeleteUser(1)
	c.Assert(err, check.Equals, ErrModifyingOnlyAdmin)

	// Verify that the admin wasn't deleted
	s.verifyRoleCount(c, role.ID, 1)
}
