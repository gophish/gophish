package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
)

func createUnpriviledgedUser(t *testing.T, slug string) *models.User {
	role, err := models.GetRoleBySlug(slug)
	if err != nil {
		t.Fatalf("error getting role by slug: %v", err)
	}
	unauthorizedUser := &models.User{
		Username: "foo",
		Hash:     "bar",
		ApiKey:   "12345",
		Role:     role,
		RoleID:   role.ID,
	}
	err = models.PutUser(unauthorizedUser)
	if err != nil {
		t.Fatalf("error saving unpriviledged user: %v", err)
	}
	return unauthorizedUser
}

func TestGetUsers(t *testing.T) {
	testCtx := setupTest(t)
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	r = ctx.Set(r, "user", testCtx.admin)
	w := httptest.NewRecorder()

	testCtx.apiServer.Users(w, r)
	expected := http.StatusOK
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}

	got := []models.User{}
	err := json.NewDecoder(w.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding users data: %v", err)
	}

	// We only expect one user
	expectedUsers := 1
	if len(got) != expectedUsers {
		t.Fatalf("unexpected number of users returned. expected %d got %d", expectedUsers, len(got))
	}
	// And it should be the admin user
	if testCtx.admin.Id != got[0].Id {
		t.Fatalf("unexpected user received. expected %d got %d", testCtx.admin.Id, got[0].Id)
	}
}

func TestCreateUser(t *testing.T) {
	testCtx := setupTest(t)
	payload := &userRequest{
		Username: "foo",
		Password: "validpassword",
		Role:     models.RoleUser,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling userRequest payload: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	r = ctx.Set(r, "user", testCtx.admin)
	w := httptest.NewRecorder()

	testCtx.apiServer.Users(w, r)
	expected := http.StatusOK
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}

	got := &models.User{}
	err = json.NewDecoder(w.Body).Decode(got)
	if err != nil {
		t.Fatalf("error decoding user payload: %v", err)
	}
	if got.Username != payload.Username {
		t.Fatalf("unexpected username received. expected %s got %s", payload.Username, got.Username)
	}
	if got.Role.Slug != payload.Role {
		t.Fatalf("unexpected role received. expected %s got %s", payload.Role, got.Role.Slug)
	}
}

// TestModifyUser tests that a user with the appropriate access is able to
// modify their username and password.
func TestModifyUser(t *testing.T) {
	testCtx := setupTest(t)
	unpriviledgedUser := createUnpriviledgedUser(t, models.RoleUser)
	newPassword := "new-password"
	newUsername := "new-username"
	payload := userRequest{
		Username: newUsername,
		Password: newPassword,
		Role:     unpriviledgedUser.Role.Slug,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling userRequest payload: %v", err)
	}
	url := fmt.Sprintf("/api/users/%d", unpriviledgedUser.Id)
	r := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unpriviledgedUser.ApiKey))
	w := httptest.NewRecorder()

	testCtx.apiServer.ServeHTTP(w, r)
	response := &models.User{}
	err = json.NewDecoder(w.Body).Decode(response)
	if err != nil {
		t.Fatalf("error decoding user payload: %v", err)
	}
	expected := http.StatusOK
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}
	if response.Username != newUsername {
		t.Fatalf("unexpected username received. expected %s got %s", newUsername, response.Username)
	}
	got, err := models.GetUser(unpriviledgedUser.Id)
	if err != nil {
		t.Fatalf("error getting unpriviledged user: %v", err)
	}
	if response.Username != got.Username {
		t.Fatalf("unexpected username received. expected %s got %s", response.Username, got.Username)
	}
	err = bcrypt.CompareHashAndPassword([]byte(got.Hash), []byte(newPassword))
	if err != nil {
		t.Fatalf("incorrect hash received for created user. expected %s got %s", []byte(newPassword), []byte(got.Hash))
	}
}

// TestUnauthorizedListUsers ensures that users without the ModifySystem
// permission are unable to list the users registered in Gophish.
func TestUnauthorizedListUsers(t *testing.T) {
	testCtx := setupTest(t)
	// First, let's create a standard user which doesn't
	// have ModifySystem permissions.
	unauthorizedUser := createUnpriviledgedUser(t, models.RoleUser)
	// We'll try to make a request to the various users API endpoints to
	// ensure that they fail. Previously, we could hit the handlers directly
	// but we need to go through the router for this test to ensure the
	// middleware gets applied.
	r := httptest.NewRequest(http.MethodGet, "/api/users/", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedUser.ApiKey))
	w := httptest.NewRecorder()

	testCtx.apiServer.ServeHTTP(w, r)
	expected := http.StatusForbidden
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}
}

// TestUnauthorizedModifyUsers verifies that users without ModifySystem
// permission (a "standard" user) can only get or modify their own information.
func TestUnauthorizedGetUser(t *testing.T) {
	testCtx := setupTest(t)
	// First, we'll make sure that a user with the "user" role is unable to
	// get the information of another user (in this case, the main admin).
	unauthorizedUser := createUnpriviledgedUser(t, models.RoleUser)
	url := fmt.Sprintf("/api/users/%d", testCtx.admin.Id)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedUser.ApiKey))
	w := httptest.NewRecorder()

	testCtx.apiServer.ServeHTTP(w, r)
	expected := http.StatusForbidden
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}
}

// TestUnauthorizedModifyRole ensures that users without the ModifySystem
// privilege are unable to modify their own role, preventing a potential
// privilege escalation issue.
func TestUnauthorizedSetRole(t *testing.T) {
	testCtx := setupTest(t)
	unauthorizedUser := createUnpriviledgedUser(t, models.RoleUser)
	url := fmt.Sprintf("/api/users/%d", unauthorizedUser.Id)
	payload := &userRequest{
		Username: unauthorizedUser.Username,
		Role:     models.RoleAdmin,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling userRequest payload: %v", err)
	}
	r := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedUser.ApiKey))
	w := httptest.NewRecorder()

	testCtx.apiServer.ServeHTTP(w, r)
	expected := http.StatusBadRequest
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}
	response := &models.Response{}
	err = json.NewDecoder(w.Body).Decode(response)
	if err != nil {
		t.Fatalf("error decoding response payload: %v", err)
	}
	if response.Message != ErrInsufficientPermission.Error() {
		t.Fatalf("incorrect error received when setting role. expected %s got %s", ErrInsufficientPermission.Error(), response.Message)
	}
}

// TestModifyWithExistingUsername verifies that it's not possible to modify
// an user's username to one which already exists.
func TestModifyWithExistingUsername(t *testing.T) {
	testCtx := setupTest(t)
	unauthorizedUser := createUnpriviledgedUser(t, models.RoleUser)
	payload := &userRequest{
		Username: testCtx.admin.Username,
		Role:     unauthorizedUser.Role.Slug,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling userRequest payload: %v", err)
	}
	url := fmt.Sprintf("/api/users/%d", unauthorizedUser.Id)
	r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedUser.ApiKey))
	w := httptest.NewRecorder()

	testCtx.apiServer.ServeHTTP(w, r)
	expected := http.StatusBadRequest
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}
	expectedResponse := &models.Response{
		Message: ErrUsernameTaken.Error(),
		Success: false,
	}
	got := &models.Response{}
	err = json.NewDecoder(w.Body).Decode(got)
	if err != nil {
		t.Fatalf("error decoding response payload: %v", err)
	}
	if got.Message != expectedResponse.Message {
		t.Fatalf("incorrect error received when setting role. expected %s got %s", expectedResponse.Message, got.Message)
	}
}
