package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldNewApp_includesAuth(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "authapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "authapp",
		ModulePath: "github.com/puppe1990/authapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/handlers/auth.go",
		"internal/models/user.go",
		"internal/store/migrations/002_auth.sql",
		"web/templates/pages/login.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(routes)
	if !strings.Contains(body, "RequireAuthFunc") {
		t.Error("routes.go missing protected dashboard")
	}
	if !strings.Contains(body, "/login") {
		t.Error("routes.go missing login routes")
	}

	appGo, err := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(appGo), "LoadSession") {
		t.Error("app.go missing LoadSession middleware")
	}
	if !strings.Contains(string(appGo), "middleware.Flash") {
		t.Error("app.go missing Flash middleware")
	}
	if !strings.Contains(string(appGo), "SecurityHeaders") {
		t.Error("app.go missing SecurityHeaders middleware")
	}

	authHandler, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/auth.go"))
	if err != nil {
		t.Fatal(err)
	}
	authBody := string(authHandler)
	if !strings.Contains(authBody, "CookieOptionsFromConfig") {
		t.Error("auth.go missing CookieOptionsFromConfig")
	}
	if !strings.Contains(authBody, "flash.Set") {
		t.Error("auth.go missing flash.Set after login")
	}
	if !strings.Contains(authBody, "meta.ForRequest") {
		t.Error("auth.go missing meta.ForRequest")
	}
	if !strings.Contains(body, "NewRateLimiter") {
		t.Error("routes.go missing rate limiter on login")
	}
}

func TestScaffoldAuth_migrationIncludesExpiresAt(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "authmigrate")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "authmigrate",
		ModulePath: "github.com/puppe1990/authmigrate",
	}, false, true); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldAuth(appDir, scaffoldData{
		AppName:    "authmigrate",
		ModulePath: "github.com/puppe1990/authmigrate",
	}, false); err != nil {
		t.Fatal(err)
	}

	migration, err := os.ReadFile(filepath.Join(appDir, "internal/store/migrations/002_auth.sql"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(migration)
	if !strings.Contains(body, "expires_at") {
		t.Errorf("002_auth.sql missing expires_at:\n%s", body)
	}
	if !strings.Contains(body, `expires_at DATETIME NOT NULL DEFAULT (datetime('now', '+7 days'))`) {
		t.Errorf("002_auth.sql missing expires_at default:\n%s", body)
	}
}
