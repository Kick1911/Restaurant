package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"

	"github.com/kick/sigma-connected/internal/dto"
	"github.com/kick/sigma-connected/internal/handler"
	"github.com/kick/sigma-connected/internal/middleware"
	"github.com/kick/sigma-connected/internal/repository"
	"github.com/kick/sigma-connected/internal/service"
	"github.com/kick/sigma-connected/pkg/response"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	db  *sqlx.DB
	rdb *redis.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection string: %v", err)
	}

	db, err = sqlx.Connect("pgx", connStr)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	redisContainer, err := tcRedis.RunContainer(
		ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(60 * time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start redis container: %v", err)
	}

	redisURL, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get redis connection string: %v", err)
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("failed to parse redis URL: %v", err)
	}

	rdb = redis.NewClient(opts)

	code := m.Run()

	db.Close()
	rdb.Close()

	if err := pgContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate postgres container: %v", err)
	}
	if err := redisContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate redis container: %v", err)
	}

	os.Exit(code)
}

func runMigrations(db *sqlx.DB) error {
	migrations := []string{
		"000001_create_tenants",
		"000002_create_users",
		"000003_create_dishes",
		"000004_create_ratings",
	}
	for _, m := range migrations {
		path := filepath.Join("..", "..", "migrations", m+".up.sql")
		sql, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(sql)); err != nil {
			return err
		}
	}
	return nil
}

func truncateAll(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE TABLE dishes, ratings, users, tenants CASCADE")
	require.NoError(t, err)
}

func flushRedis(t *testing.T) {
	t.Helper()
	require.NoError(t, rdb.FlushAll(context.Background()).Err())
}

func seedTenantAndUser(t *testing.T, db *sqlx.DB) (slug, email, password string) {
	t.Helper()

	tenantID := uuid.New()
	slug = "test-tenant"

	_, err := db.Exec(`
		INSERT INTO tenants (id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())`,
		tenantID, "Test Tenant", slug)
	require.NoError(t, err)

	password = "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	email = "test@example.com"
	_, err = db.Exec(`
		INSERT INTO users (id, tenant_id, name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
		uuid.New(), tenantID, "Test User", email, string(hash), "customer")
	require.NoError(t, err)

	return
}

func buildLoginRouter(t *testing.T) *chi.Mux {
	t.Helper()

	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	bfProtector := middleware.NewBruteForceProtector(rdb)
	svc := service.NewUserService(userRepo, tenantRepo, bfProtector, "test-secret", 24 * time.Hour, db)
	h := handler.NewUserHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/login", h.Login)
	return r
}

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
