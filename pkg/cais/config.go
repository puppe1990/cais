package cais

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port              string
	DBPath            string
	Env               string
	AppURL            string
	AdminToken        string
	Locale            string
	LogFormat         string
	TrustedProxies    []string
	PermissionsPolicy string
	CSPStyleSrc       string
	CSPConnectSrc     string
	CSPMediaSrc       string
}

func Load() Config {
	cfg := Config{
		Port:   ":8080",
		DBPath: "./data/app.db",
		Env:    "development",
		Locale: "en",
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
	if v := os.Getenv("LOCALE"); v != "" {
		cfg.Locale = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.LogFormat = v
	}
	if v := os.Getenv("TRUSTED_PROXIES"); v != "" {
		for _, ip := range strings.Split(v, ",") {
			if ip = strings.TrimSpace(ip); ip != "" {
				cfg.TrustedProxies = append(cfg.TrustedProxies, ip)
			}
		}
	}
	if v := os.Getenv("PERMISSIONS_POLICY"); v != "" {
		cfg.PermissionsPolicy = v
	} else if cfg.Env == "development" {
		// Barcode scan in dev needs camera=(self); camera=() blocks the prompt entirely.
		cfg.PermissionsPolicy = "camera=(self), microphone=(), geolocation=()"
	} else {
		cfg.PermissionsPolicy = "camera=(), microphone=(), geolocation=()"
	}
	if v := os.Getenv("CSP_STYLE_SRC"); v != "" {
		cfg.CSPStyleSrc = v
	}
	if v := os.Getenv("CSP_CONNECT_SRC"); v != "" {
		cfg.CSPConnectSrc = v
	}
	if v := os.Getenv("CSP_MEDIA_SRC"); v != "" {
		cfg.CSPMediaSrc = v
	} else if cfg.Env == "development" {
		cfg.CSPMediaSrc = "blob:"
	}

	return cfg
}

func (c Config) CookieSecure() bool {
	return c.Env == "production"
}

func (c Config) SanitizeErrors() bool {
	return c.Env == "production"
}

// LogJSON reports whether request/SQL logs should emit structured JSON lines.
// Default: JSON in development and production; LOG_FORMAT=text opts out; LOG_FORMAT=json forces JSON.
func (c Config) LogJSON() bool {
	switch strings.ToLower(strings.TrimSpace(c.LogFormat)) {
	case "json":
		return true
	case "text":
		return false
	default:
		return c.Env == "development" || c.Env == "production"
	}
}

// Validate checks required settings for the active environment.
func (c Config) Validate() error {
	if c.Env == "production" && c.AdminToken == "" {
		return fmt.Errorf("ADMIN_TOKEN is required when ENV=production")
	}
	if c.Env == "production" && c.AppURL == "" {
		return fmt.Errorf("APP_URL is required when ENV=production")
	}
	return nil
}
