package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kick/sigma-connected/pkg/response"
)

func TestDish_Search(t *testing.T) {
	_, email, password := seedTenantAndUser(t, db)
	defer truncateAll(t, db)
	defer flushRedis(t)

	loginBody := map[string]string{
		"email":       email,
		"password":    password,
	}
	b, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(b))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()

	loginRouter := buildLoginRouter(t)
	loginRouter.ServeHTTP(loginRec, loginReq)

	var loginResp response.APIResponse
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))
	require.True(t, loginResp.Success)

	data := loginResp.Data.(map[string]interface{})
	token := data["token"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dishes", nil)
	req.Header.Set("Authorization", "Bearer " + token)
	rec := httptest.NewRecorder()

	r := buildDishRouter(t)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var dishSearchResp response.PaginatedResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &dishSearchResp))

	assert.Equal(t, dishSearchResp.Success, true)
	assert.Equal(t, dishSearchResp.Data, []interface{}{})
	assert.Equal(t, dishSearchResp.Page, 1)
	assert.Equal(t, dishSearchResp.Limit, 20)
	assert.Equal(t, dishSearchResp.Total, 0)
}
