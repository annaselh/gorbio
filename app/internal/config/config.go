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
	AppBaseURL  string
	CORSOrigins []string
	SessionTTL  time.Duration
	DB          DBConfig
	SMTP        SMTPConfig
	AutoMigrate bool
}

// SMTPConfig describes the outbound mail transport. Host being empty means no
// transport is configured, which is allowed only in development.
type SMTPConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	From        string
	FromName    string
	ImplicitTLS bool
}

func (s SMTPConfig) Configured() bool {
	return strings.TrimSpace(s.Host) != ""
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

	implicitTLS, err := strconv.ParseBool(envOrDefault("SMTP_IMPLICIT_TLS", "false"))
	if err != nil {
		return nil, fmt.Errorf("parse SMTP_IMPLICIT_TLS: %w", err)
	}

	config := &Config{
		Env:         envOrDefault("APP_ENV", EnvProduction),
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		AppBaseURL:  strings.TrimRight(envOrDefault("APP_BASE_URL", "http://localhost:5173"), "/"),
		CORSOrigins: splitAndTrim(os.Getenv("CORS_ORIGINS")),
		SessionTTL:  sessionTTL,
		SMTP: SMTPConfig{
			Host:        strings.TrimSpace(os.Getenv("SMTP_HOST")),
			Port:        envOrDefault("SMTP_PORT", "587"),
			Username:    os.Getenv("SMTP_USERNAME"),
			Password:    os.Getenv("SMTP_PASSWORD"),
			From:        strings.TrimSpace(os.Getenv("SMTP_FROM")),
			FromName:    envOrDefault("SMTP_FROM_NAME", "Orbio"),
			ImplicitTLS: implicitTLS,
		},
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

	if !strings.HasPrefix(c.AppBaseURL, "http://") && !strings.HasPrefix(c.AppBaseURL, "https://") {
		return fmt.Errorf("APP_BASE_URL %q must include a scheme; it is the origin of password reset links", c.AppBaseURL)
	}

	// Password reset and email verification are useless without a transport, so
	// production must configure one rather than silently drop the mail.
	if c.IsProduction() && !c.SMTP.Configured() {
		return fmt.Errorf("SMTP_HOST must be set in production; email delivery has no fallback")
	}
	if c.SMTP.Configured() {
		if strings.TrimSpace(c.SMTP.From) == "" {
			return fmt.Errorf("SMTP_FROM must be set when SMTP_HOST is configured")
		}
		if !strings.Contains(c.SMTP.From, "@") {
			return fmt.Errorf("SMTP_FROM %q must be an email address", c.SMTP.From)
		}
		if strings.TrimSpace(c.SMTP.Port) == "" {
			return fmt.Errorf("SMTP_PORT must be set when SMTP_HOST is configured")
		}
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
