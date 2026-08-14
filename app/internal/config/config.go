package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

type Config struct {
	Env         string
	HTTPAddr    string
	CORSOrigins []string
	SessionTTL  time.Duration
	DB          DBConfig
	AutoMigrate bool
}

type DBConfig struct {
	Driver   string
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

// IsProduction reports whether the process runs with production defaults. It
// drives cookie hardening, so anything that is not explicitly "development" is
// treated as production - failing closed rather than open.
func (c *Config) IsProduction() bool {
	return c.Env != EnvDevelopment
}

func Load() (*Config, error) {

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	autoMigrate, err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE"))
	if os.Getenv("AUTO_MIGRATE") != "" && err != nil {
		return nil, fmt.Errorf("parse AUTO_MIGRATE: %w", err)
	}

	sessionTTL := 12 * time.Hour
	if raw := os.Getenv("SESSION_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parse SESSION_TTL: %w", err)
		}
		sessionTTL = parsed
	}

	config := &Config{
		Env:         envOrDefault("APP_ENV", EnvProduction),
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		CORSOrigins: splitAndTrim(os.Getenv("CORS_ORIGINS")),
		SessionTTL:  sessionTTL,
		DB: DBConfig{
			Driver:   os.Getenv("DB_DRIVER"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Database: os.Getenv("DB_DATABASE"),
			Username: os.Getenv("DB_USERNAME"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		},
		AutoMigrate: autoMigrate,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.Env != EnvDevelopment && c.Env != EnvProduction {
		return fmt.Errorf("unsupported APP_ENV %q; use %q or %q", c.Env, EnvDevelopment, EnvProduction)
	}

	if c.DB.Driver != "postgres" {
		return fmt.Errorf("unsupported DB_DRIVER %q; only postgres is supported", c.DB.Driver)
	}

	if c.SessionTTL <= 0 {
		return fmt.Errorf("SESSION_TTL must be positive")
	}

	// A cross-origin browser session needs an explicit allowlist: credentialed
	// CORS forbids the "*" wildcard, so an empty list in production means the
	// SPA is expected to be served same-origin (behind a reverse proxy).
	for _, origin := range c.CORSOrigins {
		if origin == "*" {
			return fmt.Errorf("CORS_ORIGINS must not contain %q; credentialed requests require explicit origins", "*")
		}
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return fmt.Errorf("CORS_ORIGINS entry %q must include a scheme", origin)
		}
	}

	for name, value := range map[string]string{
		"DB_HOST":     c.DB.Host,
		"DB_PORT":     c.DB.Port,
		"DB_DATABASE": c.DB.Database,
		"DB_USERNAME": c.DB.Username,
		"DB_SSLMODE":  c.DB.SSLMode,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s must be set", name)
		}
	}

	return nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitAndTrim(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
