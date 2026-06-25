package handler_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"

	"github.com/kick/sigma-connected/internal/auth"
	"github.com/kick/sigma-connected/internal/handler"
	"github.com/kick/sigma-connected/internal/middleware"
	"github.com/kick/sigma-connected/internal/repository"
	"github.com/kick/sigma-connected/internal/service"

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

func generateTestToken(t *testing.T) string {
	t.Helper()

	userID := uuid.New()
	tenantID := uuid.New()
	token, err := auth.GenerateToken("test-secret", userID, tenantID, "test@example.com", "customer", 24*time.Hour)
	require.NoError(t, err)
	return token
}

func buildLoginRouter(t *testing.T) *chi.Mux {
	t.Helper()

	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	bfProtector := middleware.NewBruteForceProtector(rdb)
	svc := service.NewUserService(userRepo, tenantRepo, bfProtector, "test-secret", 24*time.Hour, db)
	h := handler.NewUserHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/login", h.Login)
	return r
}

func buildDishRouter(t *testing.T) *chi.Mux {
	t.Helper()

	dishRepo := repository.NewDishRepository(db)
	dishService := service.NewDishService(dishRepo)
	dishHandler := handler.NewDishHandler(dishService)

	ratingRepo := repository.NewRatingRepository(db)
	ratingService := service.NewRatingService(ratingRepo)
	ratingHandler := handler.NewRatingHandler(ratingService)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth("test-secret"))
			r.Use(middleware.RateLimit)

			r.Get("/dishes", dishHandler.Search)
			r.Get("/dishes/{id}", dishHandler.GetByID)

			r.Route("/dishes/{id}/ratings", func(r chi.Router) {
				r.Get("/", ratingHandler.GetByDishID)
				r.Post("/", ratingHandler.Create)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))

				r.Post("/dishes", dishHandler.Create)
				r.Put("/dishes/{id}", dishHandler.Update)
				r.Delete("/dishes/{id}", dishHandler.Delete)
			})
		})
	})
	return r
}
