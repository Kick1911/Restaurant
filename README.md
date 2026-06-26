# sigma-connected

REST API for "The Orc Shack" restaurant — a multi-tenant dish catalog with ratings, JWT auth, brute-force protection, and rate limiting.

Built with Go 1.26, `chi` router, `sqlx`+`pgx` (Postgres), `go-redis`, and Prometheus metrics.

## Quick start

```sh
make docker-build
make docker-up       # start Postgres, Redis, and the API
docker compose exec api make migrate-up      # apply migrations (set DATABASE_URL first)
```

The API listens on `:8080`. Health check at `GET /health`, metrics at `GET /metrics`.

## API

All routes under `/api/v1`:

```
POST   /register              # register a new user (public)
POST   /login                 # login (public)
GET    /dishes                # search dishes (auth + rate-limit)
GET    /dishes/{id}           # get dish by ID (auth + rate-limit)
POST   /dishes                # create dish (auth + rate-limit + admin)
PUT    /dishes/{id}           # update dish (auth + rate-limit + admin)
DELETE /dishes/{id}           # delete dish (auth + rate-limit + admin)
GET    /dishes/{id}/ratings   # get ratings (auth + rate-limit)
POST   /dishes/{id}/ratings   # rate a dish (auth + rate-limit)
```

Auth uses `Authorization: Bearer <token>` headers. Tenancy is scoped via JWT claims — registering with `tenant_slug` auto-creates a tenant.

## Configuration

All via environment variables with sensible defaults (no `.env` file needed):

| Variable | Default | Description |
|---|---|---|
| `SERVER_HOST` | `0.0.0.0` | Bind address |
| `SERVER_PORT` | `8080` | Port |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/restaurant?sslmode=disable` | Postgres DSN |
| `REDIS_URL` | `localhost:6379` | Redis address |
| `JWT_SECRET` | `super-secret-key-change-in-production` | HS256 signing key |
| `JWT_EXPIRATION` | `24h` | Token lifetime |
| `APP_ENV` | `development` | Sets log format (text vs JSON) |

## Commands

```sh
make build          # go build -o bin/api ./cmd/api
make test           # go test ...  (requires Docker)
make lint           # golangci-lint run ./...
make sqlc           # sqlc generate
make docker-up      # docker compose up -d
make migrate-up     # golang-migrate up
```

## Testing

Integration tests spin up Postgres 16 + Redis 7 via `testcontainers-go` — Docker is required. Tests live in `internal/tests/` as a black-box package (`handler_test`). Each test seeds its data and cleans up via `TRUNCATE ... CASCADE` and `FLUSHALL`.
