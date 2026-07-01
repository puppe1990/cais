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
}
