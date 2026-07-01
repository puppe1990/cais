package cais

import "testing"

func TestConfig_DefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DB_PATH", "")
	t.Setenv("ENV", "")
	t.Setenv("ADMIN_TOKEN", "")

	cfg := Load()

	if cfg.Port != ":8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, ":8080")
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	t.Setenv("PORT", ":3000")
	t.Setenv("DB_PATH", "")
	t.Setenv("ENV", "")
	t.Setenv("ADMIN_TOKEN", "")

	cfg := Load()

	if cfg.Port != ":3000" {
		t.Errorf("Port = %q, want %q", cfg.Port, ":3000")
	}
}

func TestConfig_DBPath(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DB_PATH", "/tmp/test.db")
	t.Setenv("ENV", "")
	t.Setenv("ADMIN_TOKEN", "")

	cfg := Load()

	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/test.db")
	}
}

func TestConfig_DefaultEnv(t *testing.T) {
	t.Setenv("ENV", "")
	t.Setenv("ADMIN_TOKEN", "")

	cfg := Load()

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
}

func TestConfig_AppURL(t *testing.T) {
	t.Setenv("APP_URL", "https://pulsefit.gestaobem.com")
	t.Setenv("ADMIN_TOKEN", "")

	cfg := Load()

	if cfg.AppURL != "https://pulsefit.gestaobem.com" {
		t.Errorf("AppURL = %q, want https://pulsefit.gestaobem.com", cfg.AppURL)
	}
}

func TestConfig_AdminToken(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret-token")

	cfg := Load()

	if cfg.AdminToken != "secret-token" {
		t.Errorf("AdminToken = %q, want secret-token", cfg.AdminToken)
	}
}

func TestConfig_Validate_requiresAdminTokenInProduction(t *testing.T) {
	cfg := Config{Env: "production", AdminToken: ""}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error without ADMIN_TOKEN in production")
	}
}

func TestConfig_Validate_allowsEmptyTokenInDevelopment(t *testing.T) {
	cfg := Config{Env: "development", AdminToken: ""}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_CookieSecure_trueInProduction(t *testing.T) {
	cfg := Config{Env: "production"}
	if !cfg.CookieSecure() {
		t.Error("CookieSecure() = false, want true in production")
	}
}

func TestConfig_CookieSecure_falseInDevelopment(t *testing.T) {
	cfg := Config{Env: "development"}
	if cfg.CookieSecure() {
		t.Error("CookieSecure() = true, want false in development")
	}
}
