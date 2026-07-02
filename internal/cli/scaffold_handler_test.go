package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldHandler_AfterResourceRoutesCompile(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "menu")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "menu",
		ModulePath: "github.com/puppe1990/menu",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "dish", resourceOpts{Public: true}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldHandler(appDir, "about", false); err != nil {
		t.Fatal(err)
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(routes)
	if strings.Contains(body, "})about") || strings.Contains(body, "})\tabout") {
		t.Errorf("handler route insert must start on new line after resource group: %s", body)
	}
	if !strings.Contains(body, `r.Get("/about", about.ServeHTTP)`) {
		t.Error("missing about route")
	}
}
