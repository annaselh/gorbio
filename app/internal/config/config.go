package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
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

func Load() (*Config, error) {

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	autoMigrate, err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE"))
	if os.Getenv("AUTO_MIGRATE") != "" && err != nil {
		return nil, fmt.Errorf("parse AUTO_MIGRATE: %w", err)
	}

	config := &Config{
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
	if c.DB.Driver != "postgres" {
		return fmt.Errorf("unsupported DB_DRIVER %q; only postgres is supported", c.DB.Driver)
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
