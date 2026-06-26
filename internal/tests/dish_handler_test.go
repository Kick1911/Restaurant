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
	"github.com/kick/sigma-connected/internal/dto"
)

func TestDish_Search(t *testing.T) {
	user, password := seedTenantAndUser(t, db)
	defer truncateAll(t, db)
	defer flushRedis(t)

	loginBody := map[string]string{
		"email":       user.Email,
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

	data := loginResp.Data.(map[string]any)
	token := data["token"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dishes", nil)
	req.Header.Set("Authorization", "Bearer " + token)
	rec := httptest.NewRecorder()

	r := buildDishRouter(t)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var dishSearchResp response.PaginatedResponse[[]dto.DishResponse]
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&dishSearchResp))

	assert.Equal(t, dishSearchResp.Success, true)
	assert.Empty(t, dishSearchResp.Data)
	assert.Equal(t, dishSearchResp.Page, 1)
	assert.Equal(t, dishSearchResp.Limit, 20)
	assert.Equal(t, dishSearchResp.Total, 0)

	dishes := seedDishes(t, db, user, 5)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	// fmt.Printf("Body: %s\n", rec.Body.Bytes())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &dishSearchResp))

	assert.Equal(t, dishSearchResp.Success, true)
	assert.Equal(t, dishSearchResp.Page, 1)
	assert.Equal(t, dishSearchResp.Limit, 20)
	assert.Equal(t, dishSearchResp.Total, 5)
	assert.Equal(t, len(dishSearchResp.Data), 5)
	assert.Equal(t, dishSearchResp.Data[0].ID, dishes[0].ID.String())
	assert.Equal(t, dishSearchResp.Data[0].Name, dishes[0].Name)
	assert.Equal(t, dishSearchResp.Data[0].Description, dishes[0].Description)
	assert.Equal(t, dishSearchResp.Data[0].Price, dishes[0].Price)
	assert.Equal(t, dishSearchResp.Data[0].ImageURL, dishes[0].ImageURL)
	assert.Equal(t, dishSearchResp.Data[0].AvgRating, 0.0)
}
