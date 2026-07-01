package cais

import "testing"

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
	t.Setenv("ENV", "")

	cfg := Load()

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
}

func TestConfig_AppURL(t *testing.T) {
	t.Setenv("APP_URL", "https://pulsefit.gestaobem.com")

	cfg := Load()

	if cfg.AppURL != "https://pulsefit.gestaobem.com" {
		t.Errorf("AppURL = %q, want https://pulsefit.gestaobem.com", cfg.AppURL)
	}
}
