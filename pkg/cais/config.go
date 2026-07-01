package cais

import "os"

type Config struct {
	Port   string
	DBPath string
	Env    string
	AppURL string
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

	return cfg
}
