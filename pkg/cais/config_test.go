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
	cfg := Config{Env: "production", AdminToken: "", AppURL: "https://example.com"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error without ADMIN_TOKEN in production")
	}
}

func TestConfig_Validate_requiresAppURLInProduction(t *testing.T) {
	cfg := Config{Env: "production", AdminToken: "secret", AppURL: ""}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error without APP_URL in production")
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

func TestConfig_SanitizeErrors_trueInProduction(t *testing.T) {
	cfg := Config{Env: "production"}
	if !cfg.SanitizeErrors() {
		t.Error("SanitizeErrors() = false, want true in production")
	}
}

func TestConfig_SanitizeErrors_falseInDevelopment(t *testing.T) {
	cfg := Config{Env: "development"}
	if cfg.SanitizeErrors() {
		t.Error("SanitizeErrors() = true, want false in development")
	}
}

func TestConfig_DefaultLocale(t *testing.T) {
	t.Setenv("LOCALE", "")

	cfg := Load()

	if cfg.Locale != "en" {
		t.Errorf("Locale = %q, want en", cfg.Locale)
	}
}

func TestConfig_LoadLocaleFromEnv(t *testing.T) {
	t.Setenv("LOCALE", "pt_BR")

	cfg := Load()

	if cfg.Locale != "pt_BR" {
		t.Errorf("Locale = %q, want pt_BR", cfg.Locale)
	}
}

func TestConfig_TrustedProxies_emptyByDefault(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "")

	cfg := Load()

	if len(cfg.TrustedProxies) != 0 {
		t.Errorf("TrustedProxies = %v, want empty", cfg.TrustedProxies)
	}
}

func TestConfig_LogJSON_developmentAndProductionDefault(t *testing.T) {
	for _, env := range []string{"development", "production"} {
		if !(Config{Env: env}).LogJSON() {
			t.Errorf("LogJSON() = false for %q, want true", env)
		}
	}
}

func TestConfig_LogJSON_textOverride(t *testing.T) {
	cfg := Config{Env: "production", LogFormat: "text"}
	if cfg.LogJSON() {
		t.Fatal("LogJSON() = true with LOG_FORMAT=text")
	}
}

func TestConfig_LogJSON_jsonOverride(t *testing.T) {
	cfg := Config{Env: "staging", LogFormat: "json"}
	if !cfg.LogJSON() {
		t.Fatal("LogJSON() = false with LOG_FORMAT=json")
	}
}

func TestConfig_LoadLogFormatFromEnv(t *testing.T) {
	t.Setenv("LOG_FORMAT", "text")
	cfg := Load()
	if cfg.LogFormat != "text" {
		t.Errorf("LogFormat = %q, want text", cfg.LogFormat)
	}
}

func TestConfig_TrustedProxies_loadFromEnv(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "127.0.0.1, 10.0.0.1")

	cfg := Load()

	want := []string{"127.0.0.1", "10.0.0.1"}
	if len(cfg.TrustedProxies) != len(want) {
		t.Fatalf("TrustedProxies = %v, want %v", cfg.TrustedProxies, want)
	}
	for i, ip := range want {
		if cfg.TrustedProxies[i] != ip {
			t.Errorf("TrustedProxies[%d] = %q, want %q", i, cfg.TrustedProxies[i], ip)
		}
	}
}
