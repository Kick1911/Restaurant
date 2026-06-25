package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kick/sigma-connected/internal/dto"
	"github.com/kick/sigma-connected/pkg/response"
)

func TestLogin_Success(t *testing.T) {
	slug, email, password := seedTenantAndUser(t, db)
	defer truncateAll(t, db)
	defer flushRedis(t)

	r := buildLoginRouter(t)

	body := dto.LoginRequest{
		TenantSlug: slug,
		Email:      email,
		Password:   password,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp response.APIResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	token, ok := data["token"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, token)

	userData, ok := data["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, email, userData["email"])
	assert.Equal(t, "Test User", userData["name"])
	assert.Equal(t, "customer", userData["role"])
}

func TestLogin_InvalidPassword(t *testing.T) {
	slug, email, _ := seedTenantAndUser(t, db)
	defer truncateAll(t, db)
	defer flushRedis(t)

	r := buildLoginRouter(t)

	body := dto.LoginRequest{
		TenantSlug: slug,
		Email:      email,
		Password:   "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp response.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid email or password")
}

func TestLogin_NonExistentUser(t *testing.T) {
	slug, _, _ := seedTenantAndUser(t, db)
	defer truncateAll(t, db)
	defer flushRedis(t)

	r := buildLoginRouter(t)

	body := dto.LoginRequest{
		TenantSlug: slug,
		Email:      "nonexistent@example.com",
		Password:   "password123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp response.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid email or password")
}

func TestLogin_InvalidJSON(t *testing.T) {
	defer flushRedis(t)

	r := buildLoginRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader([]byte(`{bad json`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp response.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "invalid request body", resp.Error)
}

func TestLogin_ValidationError(t *testing.T) {
	defer flushRedis(t)

	r := buildLoginRouter(t)

	body := dto.LoginRequest{
		TenantSlug: "",
		Email:      "",
		Password:   "",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp response.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "validation failed", resp.Error)
}
