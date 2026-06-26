.PHONY: build run migrate-up migrate-down lint test clean

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

test:
	go test -v -count=1 -timeout=180s ./internal/tests

clean:
	rm -rf bin/

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

.PHONY: sqlc
sqlc:
	sqlc generate
