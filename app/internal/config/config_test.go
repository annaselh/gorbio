package config

import (
	"strings"
	"testing"
	"time"
)

func validConfig() *Config {
	return &Config{
		Env:        EnvProduction,
		HTTPAddr:   ":8080",
		SessionTTL: time.Hour,
		DB: DBConfig{
			Driver: "postgres", Host: "localhost", Port: "5432",
			Database: "gorbio", Username: "gorbio", SSLMode: "disable",
		},
	}
}

func TestValidateAcceptsCompleteConfig(t *testing.T) {
	if err := validConfig().Validate(); err != nil {
		t.Fatalf("expected a complete config to validate, got %v", err)
	}
}

func TestValidateRejectsWildcardCORSOrigin(t *testing.T) {
	config := validConfig()
	config.CORSOrigins = []string{"*"}

	err := config.Validate()
	if err == nil {
		t.Fatal("a wildcard origin is invalid for credentialed CORS and must be rejected")
	}
	if !strings.Contains(err.Error(), "credentialed") {
		t.Fatalf("error should explain why the wildcard is refused, got %v", err)
	}
}

func TestValidateRejectsSchemelessCORSOrigin(t *testing.T) {
	config := validConfig()
	config.CORSOrigins = []string{"localhost:5173"}

	if err := config.Validate(); err == nil {
		t.Fatal("an origin without a scheme never matches a browser Origin header")
	}
}

func TestValidateAcceptsExplicitOrigins(t *testing.T) {
	config := validConfig()
	config.CORSOrigins = []string{"http://localhost:5173", "https://erp.example.com"}

	if err := config.Validate(); err != nil {
		t.Fatalf("explicit origins should validate, got %v", err)
	}
}

func TestValidateRejectsUnknownEnv(t *testing.T) {
	config := validConfig()
	config.Env = "staging"

	if err := config.Validate(); err == nil {
		t.Fatal("an unrecognised APP_ENV must be rejected rather than silently treated as production")
	}
}

func TestValidateRejectsNonPositiveSessionTTL(t *testing.T) {
	config := validConfig()
	config.SessionTTL = 0

	if err := config.Validate(); err == nil {
		t.Fatal("a non-positive session TTL would issue already-expired sessions")
	}
}

func TestValidateRejectsMissingDatabaseFields(t *testing.T) {
	config := validConfig()
	config.DB.Host = "   "

	if err := config.Validate(); err == nil {
		t.Fatal("a blank DB_HOST must not pass validation")
	}
}

// Anything that is not explicitly "development" must harden cookies, so a typo
// in APP_ENV fails closed rather than shipping insecure cookies to production.
func TestIsProductionFailsClosed(t *testing.T) {
	if (&Config{Env: EnvDevelopment}).IsProduction() {
		t.Fatal("development must not be treated as production")
	}
	if !(&Config{Env: EnvProduction}).IsProduction() {
		t.Fatal("production must be treated as production")
	}
	if !(&Config{Env: ""}).IsProduction() {
		t.Fatal("an unset environment must default to production hardening")
	}
}

func TestSplitAndTrim(t *testing.T) {
	got := splitAndTrim(" http://a.test , http://b.test ,, ")
	want := []string{"http://a.test", "http://b.test"}

	if len(got) != len(want) {
		t.Fatalf("expected %d origins, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("origin %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestSplitAndTrimEmpty(t *testing.T) {
	if got := splitAndTrim("   "); got != nil {
		t.Fatalf("a blank list should yield nil so CORS stays disabled, got %v", got)
	}
}
