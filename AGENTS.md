# sigma-connected

REST API for "The Orc Shack" restaurant — a multi-tenant dish catalog with ratings, JWT auth, brute-force protection, and rate limiting.

## Stack

- **Go 1.26**, `chi` router, `sqlx`+`pgx` (Postgres), `go-redis`
- Config via env vars (`cleanenv`), no `.env` file required (sensible defaults)
- Validation via `go-playground/validator`
- Metrics at `/metrics` (Prometheus), structured logging via `slog`

## Commands

```sh
make build                  # go build -o bin/api ./cmd/api
make run                    # go run ./cmd/api
make test                   # go test ./... -v  (requires Docker)
make lint                   # golangci-lint run ./...
make sqlc                   # sqlc generate  (see sqlc.yaml)
make docker-build           # docker compose build
make docker-up              # docker compose up -d
make migrate-up             # migrate -path migrations -database "$DATABASE_URL" up
```

## Architecture

| Layer | Directory | Notes |
|---|---|---|
| Entrypoint | `cmd/api/main.go` | Wires everything; graceful shutdown via signals |
| Config | `internal/config` | All env-var driven |
| Handlers | `internal/handler` | Validate via `dto.Validate()`, call services, use `pkg/response` helpers |
| Services | `internal/service` | Business logic; `UserService` owns brute-force + JWT |
| Repositories | `internal/repository` | Raw sqlx queries (no ORM) |
| Middleware | `internal/middleware` | `Auth` (JWT), `RequireRole`, `RateLimiter` (Redis sorted set, 20 req/s sliding window), `BruteForceProtector` (Redis, 5 attempts → 15m lockout), `Logging` |
| Auth | `internal/auth` | `jwt.go` (HS256, issuer `the-orc-shack`), `password.go` (bcrypt) |
| DTOs | `internal/dto` | Request/response structs with `validate` tags |
| Response helpers | `pkg/response` | `JSON`, `Error`, `ValidationError`, `Paginated` — all wrapped in `{"success": bool, ...}` |
| Migrations | `migrations/` | 4 migrations: tenants → users → dishes → ratings |
| sqlc codegen | `internal/db/` | **Not yet generated** — `sqlc.yaml` targets this dir (pgx/v5) |

## API routes

All under `/api/v1`:

```
POST   /register               # public
POST   /login                  # public
GET    /dishes                 # auth + rate-limit (search, any role)
GET    /dishes/{id}            # auth + rate-limit (any role)
POST   /dishes                 # auth + rate-limit + admin
PUT    /dishes/{id}            # auth + rate-limit + admin
DELETE /dishes/{id}            # auth + rate-limit + admin
GET    /dishes/{id}/ratings    # auth + rate-limit (any role)
POST   /dishes/{id}/ratings    # auth + rate-limit (any role)
```

## Testing

- **Requires Docker** — `TestMain` spins up Postgres 16 + Redis 7 via `testcontainers-go`
- Tests live in `internal/tests/` as `package handler_test` (black-box)
- Each test seeds data with `seedTenantAndUser`, `seedDishes`, and cleans up with `truncateAll` + `flushRedis`
- `TestMain` applies migrations directly (reads `.up.sql` files), not via `golang-migrate`

## Multi-tenant quirks

- Tenancy is header-based: **`X-Tenant-ID`** is required. The middleware at `internal/middleware/tenant.go` exists but is **not wired in `main.go`** — tenancy currently comes from JWT claims (`claims.TenantID`) instead.
- User registration auto-creates a tenant by `tenant_slug` if it doesn't exist.
- Slug `"admin"` grants `admin` role on registration.

## Gotchas

- `internal/db/` is the intended sqlc output dir but doesn't exist — run `make sqlc` first if adding sqlc-generated queries.
- `Login` in `UserService.FindByEmail` queries **without tenant scoping** (searches globally by email), unlike other queries.

