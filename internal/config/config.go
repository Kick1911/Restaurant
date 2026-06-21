package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	App      AppConfig
}

type ServerConfig struct {
	Host            string        `env:"SERVER_HOST" env-default:"0.0.0.0"`
	Port            string        `env:"SERVER_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" env-default:"30s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" env-default:"15s"`
}

type DatabaseConfig struct {
	URL string `env:"DATABASE_URL" env-default:"postgres://postgres:postgres@localhost:5432/restaurant?sslmode=disable"`
}

type RedisConfig struct {
	URL string `env:"REDIS_URL" env-default:"localhost:6379"`
}

type JWTConfig struct {
	Secret     string        `env:"JWT_SECRET" env-default:"super-secret-key-change-in-production"`
	Expiration time.Duration `env:"JWT_EXPIRATION" env-default:"24h"`
}

type AppConfig struct {
	Environment string `env:"APP_ENV" env-default:"development"`
}

func (c ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return &cfg, nil
}
