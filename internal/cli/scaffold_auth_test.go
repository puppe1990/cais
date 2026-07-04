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
		"internal/store/password_reset.go",
		"web/templates/pages/login.html",
		"web/templates/pages/signup.html",
		"web/templates/pages/forgot_password.html",
		"web/templates/pages/reset_password.html",
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
	if !strings.Contains(body, "/forgot-password") {
		t.Error("routes.go missing forgot-password routes")
	}
	if !strings.Contains(authBody, "ForgotPasswordPost") {
		t.Error("auth.go missing password reset handlers")
	}
	if !strings.Contains(body, "/signup") {
		t.Error("routes.go missing signup routes")
	}
	if !strings.Contains(authBody, "SignUpPost") {
		t.Error("auth.go missing signup handlers")
	}

	for _, page := range []string{"login.html", "signup.html", "reset_password.html"} {
		body, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages", page))
		if err != nil {
			t.Fatalf("read %s: %v", page, err)
		}
		s := string(body)
		if !strings.Contains(s, "fieldPassword") {
			t.Errorf("%s missing fieldPassword helper", page)
		}
	}
}

func TestScaffoldAuth_patchesBlankAppStore(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "blankauth")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "blankauth",
		ModulePath: "github.com/puppe1990/blankauth",
	}, false, true); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldAuth(appDir, scaffoldData{
		AppName:    "blankauth",
		ModulePath: "github.com/puppe1990/blankauth",
	}, false); err != nil {
		t.Fatal(err)
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(store)
	for _, needle := range []string{
		"FindUserByEmail",
		"CreateUser",
		"ErrEmailTaken",
		"Sessions() session.Store",
		`"github.com/puppe1990/blankauth/internal/models"`,
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("store.go missing %q:\n%s", needle, body)
		}
	}
}

func TestScaffoldAuth_usesNextMigrationNumber(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "authnum")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "authnum",
		ModulePath: "github.com/puppe1990/authnum",
	}, false, true); err != nil {
		t.Fatal(err)
	}

	migrationsDir := filepath.Join(appDir, "internal/store/migrations")
	if err := os.WriteFile(filepath.Join(migrationsDir, "002_existing.sql"), []byte("-- up\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldAuth(appDir, scaffoldData{
		AppName:    "authnum",
		ModulePath: "github.com/puppe1990/authnum",
	}, false); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(migrationsDir, "003_auth.sql")); err != nil {
		t.Fatalf("expected 003_auth.sql, stat: %v", err)
	}
	if _, err := os.Stat(filepath.Join(migrationsDir, "002_auth.sql")); err == nil {
		t.Fatal("should not create 002_auth.sql when 002 is taken")
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

	migration, err := os.ReadFile(filepath.Join(appDir, "internal/store/migrations/001_auth.sql"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(migration)
	if !strings.Contains(body, "expires_at") {
		t.Errorf("001_auth.sql missing expires_at:\n%s", body)
	}
	if !strings.Contains(body, `expires_at DATETIME NOT NULL DEFAULT (datetime('now', '+7 days'))`) {
		t.Errorf("001_auth.sql missing expires_at default:\n%s", body)
	}
}

func TestScaffoldNewApp_includesSeeds(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "seedapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "seedapp",
		ModulePath: "github.com/puppe1990/seedapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	seeds, err := os.ReadFile(filepath.Join(appDir, "internal/db/seeds.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(seeds)
	if !strings.Contains(body, "func RunSeeds") {
		t.Error("seeds.go missing RunSeeds")
	}
	if !strings.Contains(body, "InsertContact") {
		t.Error("full app seeds.go should seed demo contact")
	}
}
