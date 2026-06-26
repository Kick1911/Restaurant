FROM golang:1.26.4-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

FROM alpine:3.20
RUN apk --no-cache add ca-certificates make

WORKDIR /app

COPY --from=builder /app/api /
COPY --from=builder /go/bin/migrate /usr/local/bin
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
