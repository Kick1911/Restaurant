.PHONY: build run migrate-up migrate-down lint test clean

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

lint:
	golangci-lint run ./...

test:
	go test ./... -v

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
