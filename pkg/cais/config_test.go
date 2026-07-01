package cais

import (
	"os"
	"testing"
)

func TestConfig_DefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DB_PATH", "")
	t.Setenv("ENV", "")

	cfg := Load()

	if cfg.Port != ":8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, ":8080")
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	t.Setenv("PORT", ":3000")
	t.Setenv("DB_PATH", "")
	t.Setenv("ENV", "")

	cfg := Load()

	if cfg.Port != ":3000" {
		t.Errorf("Port = %q, want %q", cfg.Port, ":3000")
	}
}

func TestConfig_DBPath(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DB_PATH", "/tmp/test.db")
	t.Setenv("ENV", "")

	cfg := Load()

	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/test.db")
	}
}

func TestConfig_DefaultEnv(t *testing.T) {
	os.Unsetenv("ENV")

	cfg := Load()

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
}