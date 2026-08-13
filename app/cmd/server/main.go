package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/extensions"
	"github.com/annaselh/gorbio/internal/config"
	"github.com/annaselh/gorbio/internal/database"
	"github.com/annaselh/gorbio/modules"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
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

	moduleRegistry := core.NewRegistry()

	extensionRegistry := core.NewExtensionRegistry()

	app := core.NewApp(
		db,
		router,
		moduleRegistry,
		extensionRegistry,
	)

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
		Addr:              ":8080",
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
