package database

import (
	"fmt"
	"net"
	"net/url"

	"github.com/annaselh/gorbio/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.DBConfig) (*gorm.DB, error) {
	if cfg.Driver != "postgres" {
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}

	dsn := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
		Path:   cfg.Database,
		RawQuery: url.Values{
			"sslmode": []string{cfg.SSLMode},
		}.Encode(),
	}).String()

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)

	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	return db, nil
}
