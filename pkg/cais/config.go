package cais

import (
	"fmt"
	"os"
)

type Config struct {
	Port       string
	DBPath     string
	Env        string
	AppURL     string
	AdminToken string
}

func Load() Config {
	cfg := Config{
		Port:   ":8080",
		DBPath: "./data/app.db",
		Env:    "development",
	}

	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("ENV"); v != "" {
		cfg.Env = v
	}
	if v := os.Getenv("APP_URL"); v != "" {
		cfg.AppURL = v
	}
	if v := os.Getenv("ADMIN_TOKEN"); v != "" {
		cfg.AdminToken = v
	}

	return cfg
}

func (c Config) CookieSecure() bool {
	return c.Env == "production"
}

// Validate checks required settings for the active environment.
func (c Config) Validate() error {
	if c.Env == "production" && c.AdminToken == "" {
		return fmt.Errorf("ADMIN_TOKEN is required when ENV=production")
	}
	return nil
}
