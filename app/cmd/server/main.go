package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/extensions"
	"github.com/annaselh/gorbio/internal/config"
	"github.com/annaselh/gorbio/internal/database"
	"github.com/annaselh/gorbio/internal/mailer"
	"github.com/annaselh/gorbio/modules"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// buildMailer returns the SMTP transport when one is configured. Config
// validation already rejects a production start without SMTP_HOST, so the
// logging fallback can only be reached in development, where it prints reset
// links to the console instead of dropping them.
func buildMailer(cfg *config.Config) core.Mailer {
	if !cfg.SMTP.Configured() {
		slog.Warn("no SMTP transport configured; transactional email will be logged, not delivered")
		return core.LogMailer{}
	}

	return mailer.New(mailer.Config{
		Host:        cfg.SMTP.Host,
		Port:        cfg.SMTP.Port,
		Username:    cfg.SMTP.Username,
		Password:    cfg.SMTP.Password,
		From:        cfg.SMTP.From,
		FromName:    cfg.SMTP.FromName,
		ImplicitTLS: cfg.SMTP.ImplicitTLS,
	})
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	db, err := database.Connect(cfg.DB)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database handle: %w", err)
	}
	defer sqlDB.Close()

	router := core.NewRouter()

	// Global middleware must be installed before modules register routes.
	if len(cfg.CORSOrigins) > 0 {
		router.Use(core.CORS(cfg.CORSOrigins))
	}

	moduleRegistry := core.NewRegistry()

	extensionRegistry := core.NewExtensionRegistry()

	app := core.NewApp(
		db,
		router,
		moduleRegistry,
		extensionRegistry,
	).WithSettings(core.Settings{
		Env: cfg.Env,
		// Secure cookies are dropped by browsers over plain HTTP, which would
		// make login silently fail on a local http://localhost dev server.
		CookieSecure: cfg.IsProduction(),
		SessionTTL:   cfg.SessionTTL,
		BaseURL:      cfg.AppBaseURL,
	}).WithMailer(buildMailer(cfg))

	if err := modules.RegisterAll(moduleRegistry); err != nil {
		return fmt.Errorf("add modules: %w", err)
	}

	if err := extensions.RegisterAll(extensionRegistry); err != nil {
		return fmt.Errorf("add extensions: %w", err)
	}

	if cfg.AutoMigrate {
		if err := app.Migrate(); err != nil {
			return fmt.Errorf("migrate application: %w", err)
		}
	}

	if err := app.Register(); err != nil {
		return err
	}

	if err := app.Boot(); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
	case <-signalContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		if err := app.Shutdown(shutdownContext); err != nil {
			return err
		}
	}

	return nil
}
