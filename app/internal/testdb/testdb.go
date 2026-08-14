// Package testdb boots the real application against a live Postgres.
//
// The unit tests elsewhere in this repository run against a nil database
// handle, which means they can only reach the validation that happens before a
// query. Everything the database itself decides - the row locks, the sequence
// allocation, the unique indexes that back them, the SQL the dashboard is built
// from, and the migrations - is invisible to them.
//
// These helpers build the application exactly as cmd/server does, through the
// module registry, so the tests exercise the real wiring rather than a
// hand-assembled imitation of it. A test that constructed the services itself
// would not notice if module registration stopped connecting them.
package testdb

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/extensions"
	"github.com/annaselh/gorbio/modules"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DSNEnv names the environment variable holding the test database URL. The
// tests skip when it is unset rather than failing, so `go test ./...` stays
// usable without a database while CI can supply one.
const DSNEnv = "GORBIO_TEST_DATABASE_URL"

var (
	once     sync.Once
	shared   *core.App
	setupErr error
)

// App returns the booted application, skipping the test when no database is
// configured. Migration and registration run once per test binary: they are
// idempotent, but repeating them for every test would dominate the runtime.
func App(t *testing.T) *core.App {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv(DSNEnv))
	if dsn == "" {
		t.Skipf("set %s to run integration tests, for example:\n"+
			"  docker run --rm -e POSTGRES_PASSWORD=gorbio -e POSTGRES_USER=gorbio "+
			"-e POSTGRES_DB=gorbio_test -p 5432:5432 -d postgres:17\n"+
			"  %s='postgres://gorbio:gorbio@localhost:5432/gorbio_test?sslmode=disable' go test ./...",
			DSNEnv, DSNEnv)
	}

	once.Do(func() { shared, setupErr = boot(dsn) })
	if setupErr != nil {
		t.Fatalf("boot the integration application: %v", setupErr)
	}
	return shared
}

// DB is the shorthand for tests that only need the handle.
func DB(t *testing.T) *gorm.DB {
	t.Helper()
	return App(t).DB
}

func boot(dsn string) (*core.App, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// The query log would bury the test output; a failing assertion says
		// more than the statement that produced it.
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", DSNEnv, err)
	}

	moduleRegistry := core.NewRegistry()
	extensionRegistry := core.NewExtensionRegistry()

	app := core.NewApp(db, core.NewRouter(), moduleRegistry, extensionRegistry).
		WithSettings(core.Settings{
			Env:        core.EnvDevelopment,
			SessionTTL: time.Hour,
			BaseURL:    "http://localhost",
		})

	if err := modules.RegisterAll(moduleRegistry); err != nil {
		return nil, fmt.Errorf("add modules: %w", err)
	}
	if err := extensions.RegisterAll(extensionRegistry); err != nil {
		return nil, fmt.Errorf("add extensions: %w", err)
	}
	if err := app.Migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := app.Register(); err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	if err := app.Boot(); err != nil {
		return nil, fmt.Errorf("boot: %w", err)
	}
	return app, nil
}

// Tenant creates a fresh tenant and returns its id.
//
// Isolation is by tenant rather than by truncating tables between tests. Every
// query in this codebase is tenant-scoped, so a private tenant gives each test
// a private view of the data, and tests can run in parallel without one
// clearing another's rows out from under it. It also means a test that leaks
// across the tenant boundary fails here rather than passing quietly.
func Tenant(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()

	tenant := base.Tenant{
		ID:       uuid.New(),
		Slug:     "it-" + uuid.NewString(),
		Name:     "Integration Tenant",
		Status:   base.TenantStatusActive,
		Plan:     "starter",
		Settings: []byte("{}"),
	}
	if err := db.Create(&tenant).Error; err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	return tenant.ID
}

// User creates an active user, for the membership rules that need one.
func User(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()

	user := base.User{
		ID:           uuid.New(),
		Email:        "it-" + uuid.NewString() + "@example.test",
		DisplayName:  "Integration User",
		PasswordHash: "not-a-real-hash",
		Status:       base.UserStatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}
