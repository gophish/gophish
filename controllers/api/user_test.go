package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"golang.org/x/crypto/bcrypt"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
)

func (s *APISuite) createUnpriviledgedUser(slug string) *models.User {
	role, err := models.GetRoleBySlug(slug)
	s.Nil(err)
	unauthorizedUser := &models.User{
		Username: "foo",
		Hash:     "bar",
		ApiKey:   "12345",
		Role:     role,
		RoleID:   role.ID,
	}
	err = models.PutUser(unauthorizedUser)
	s.Nil(err)
	return unauthorizedUser
}

func (s *APISuite) TestGetUsers() {
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	r = ctx.Set(r, "user", s.admin)
	w := httptest.NewRecorder()

	s.apiServer.Users(w, r)
	s.Equal(w.Code, http.StatusOK)

	got := []models.User{}
	err := json.NewDecoder(w.Body).Decode(&got)
	s.Nil(err)

	// We only expect one user
	s.Equal(1, len(got))
	// And it should be the admin user
	s.Equal(s.admin.Id, got[0].Id)
}

func (s *APISuite) TestCreateUser() {
	payload := &userRequest{
		Username: "foo",
		Password: "bar",
		Role:     models.RoleUser,
	}
	body, err := json.Marshal(payload)
	s.Nil(err)

	r := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	r = ctx.Set(r, "user", s.admin)
	w := httptest.NewRecorder()

	s.apiServer.Users(w, r)
	s.Equal(w.Code, http.StatusOK)

	got := &models.User{}
	err = json.NewDecoder(w.Body).Decode(got)
	s.Nil(err)
	s.Equal(got.Username, payload.Username)
	s.Equal(got.Role.Slug, payload.Role)
}

// TestModifyUser tests that a user with the appropriate access is able to
// modify their username and password.
func (s *APISuite) TestModifyUser() {
	unpriviledgedUser := s.createUnpriviledgedUser(models.RoleUser)
	newPassword := "new-password"
	newUsername := "new-username"
	payload := userRequest{
		Username: newUsername,
		Password: newPassword,
		Role:     unpriviledgedUser.Role.Slug,
	}
	body, err := json.Marshal(payload)
	s.Nil(err)
	url := fmt.Sprintf("/api/users/%d", unpriviledgedUser.Id)
	r := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unpriviledgedUser.ApiKey))
	w := httptest.NewRecorder()

	s.apiServer.ServeHTTP(w, r)
	response := &models.User{}
	err = json.NewDecoder(w.Body).Decode(response)
	s.Nil(err)
	s.Equal(w.Code, http.StatusOK)
	s.Equal(response.Username, newUsername)
	got, err := models.GetUser(unpriviledgedUser.Id)
	s.Nil(err)
	s.Equal(response.Username, got.Username)
	s.Equal(newUsername, got.Username)
	err = bcrypt.CompareHashAndPassword([]byte(got.Hash), []byte(newPassword))
	s.Nil(err)
}

// TestUnauthorizedListUsers ensures that users without the ModifySystem
// permission are unable to list the users registered in Gophish.
func (s *APISuite) TestUnauthorizedListUsers() {
	// First, let's create a standard user which doesn't
	// have ModifySystem permissions.
	unauthorizedUser := s.createUnpriviledgedUser(models.RoleUser)
	// We'll try to make a request to the various users API endpoints to
	// ensure that they fail. Previously, we could hit the handlers directly
	// but we need to go through the router for this test to ensure the
	// middleware gets applied.
	r := httptest.NewRequest(http.MethodGet, "/api/users/", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedUser.ApiKey))
	w := httptest.NewRecorder()

	s.apiServer.ServeHTTP(w, r)
	s.Equal(w.Code, http.StatusForbidden)
}

// TestUnauthorizedModifyUsers verifies that users without ModifySystem
// permission (a "standard" user) can only get or modify their own information.
func (s *APISuite) TestUnauthorizedGetUser() {
	// First, we'll make sure that a user with the "user" role is unable to
	// get the information of another user (in this case, the main admin).
	unauthorizedUser := s.createUnpriviledgedUser(models.RoleUser)
	url := fmt.Sprintf("/api/users/%d", s.admin.Id)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedUser.ApiKey))
	w := httptest.NewRecorder()

	s.apiServer.ServeHTTP(w, r)
	s.Equal(w.Code, http.StatusForbidden)
}

// TestUnauthorizedModifyRole ensures that users without the ModifySystem
// privilege are unable to modify their own role, preventing a potential
// privilege escalation issue.
func (s *APISuite) TestUnauthorizedSetRole() {
	unauthorizedUser := s.createUnpriviledgedUser(models.RoleUser)
	url := fmt.Sprintf("/api/users/%d", unauthorizedUser.Id)
	payload := &userRequest{
		Username: unauthorizedUser.Username,
		Role:     models.RoleAdmin,
	}
	body, err := json.Marshal(payload)
	s.Nil(err)
	r := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedUser.ApiKey))
	w := httptest.NewRecorder()

	s.apiServer.ServeHTTP(w, r)
	s.Equal(w.Code, http.StatusBadRequest)
	response := &models.Response{}
	err = json.NewDecoder(w.Body).Decode(response)
	s.Nil(err)
	s.Equal(response.Message, ErrInsufficientPermission.Error())
}

// TestModifyWithExistingUsername verifies that it's not possible to modify
// an user's username to one which already exists.
func (s *APISuite) TestModifyWithExistingUsername() {
	unauthorizedUser := s.createUnpriviledgedUser(models.RoleUser)
	payload := &userRequest{
		Username: s.admin.Username,
		Role:     unauthorizedUser.Role.Slug,
	}
	body, err := json.Marshal(payload)
	s.Nil(err)
	url := fmt.Sprintf("/api/users/%d", unauthorizedUser.Id)
	r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedUser.ApiKey))
	w := httptest.NewRecorder()

	s.apiServer.ServeHTTP(w, r)
	s.Equal(w.Code, http.StatusBadRequest)
	expected := &models.Response{
		Message: ErrUsernameTaken.Error(),
		Success: false,
	}
	got := &models.Response{}
	err = json.NewDecoder(w.Body).Decode(got)
	s.Nil(err)
	s.Equal(got.Message, expected.Message)
}
