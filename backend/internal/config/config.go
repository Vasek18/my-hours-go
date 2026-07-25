package config

import (
	"log"
	"os"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	AppPort       string // port the HTTP server listens on
	AppEnv        string // "dev" or "production"
	AppURL        string // public base URL of the web app (for building links)
	DatabaseURL   string // Postgres connection string
	RedisAddr     string // host:port of Redis
	SessionSecret string // key used to authenticate/encrypt the session cookie
}

// Load reads configuration from the environment, applying sensible local defaults.
func Load() Config {
	cfg := Config{
		// Railway injects PORT; fall back to APP_PORT, then a local default.
		AppPort:       env("APP_PORT", env("PORT", "8080")),
		AppEnv:        env("APP_ENV", "dev"),
		AppURL:        env("APP_URL", "http://localhost:5173"),
		DatabaseURL:   env("DATABASE_URL", "postgres://my_hours:my_hours@localhost:5432/my_hours?sslmode=disable"),
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		SessionSecret: env("SESSION_SECRET", "change-me-in-production-please-32b"),
	}

	if cfg.IsProduction() && cfg.SessionSecret == "change-me-in-production-please-32b" {
		log.Fatal("SESSION_SECRET must be set to a strong value in production")
	}

	return cfg
}

// IsProduction reports whether the app is running in production mode.
func (c Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
