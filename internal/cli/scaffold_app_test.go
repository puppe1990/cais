package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldApp_supermarket(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "prices")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "prices",
		ModulePath: "github.com/puppe1990/prices",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldApp(appDir, "supermarket", false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/handlers/supermarket.go",
		"internal/handlers/supermarket_test.go",
		"web/templates/layouts/base.html",
		"web/templates/pages/scan.html",
		"web/templates/pages/map.html",
		"web/templates/pages/feed.html",
		"web/templates/pages/achievements.html",
		"web/templates/pages/nfce.html",
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
	for _, want := range []string{
		"NewSupermarketHandler",
		`r.Get("/", super.Scan)`,
		`r.Get("/map", super.Map)`,
		`r.Get("/feed", super.Feed)`,
		`r.Get("/achievements", super.Achievements)`,
		`r.Get("/nfce", super.NFCe)`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("routes.go missing %q", want)
		}
	}
	if strings.Contains(body, `r.Get("/", home.ServeHTTP)`) {
		t.Error("routes.go should replace home with supermarket scan page")
	}

	layout, err := os.ReadFile(filepath.Join(appDir, "web/templates/layouts/base.html"))
	if err != nil {
		t.Fatal(err)
	}
	layoutBody := string(layout)
	for _, want := range []string{"navTab", "makeNavTab", "ActiveNav", "cais-toast-host", "Escanear Preço"} {
		if !strings.Contains(layoutBody, want) {
			t.Errorf("base layout missing %q", want)
		}
	}

	handler, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/supermarket.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(handler), "demoScans") {
		t.Error("supermarket handler should ship demo data")
	}
}

func TestScaffoldApp_unknownTemplate(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "x")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "x",
		ModulePath: "github.com/puppe1990/x",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	err := scaffoldApp(appDir, "notreal", false)
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "notreal") {
		t.Errorf("error = %v", err)
	}
}

func TestScaffoldApp_supermarketIdempotentGuard(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "dup")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "dup",
		ModulePath: "github.com/puppe1990/dup",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldApp(appDir, "supermarket", false); err != nil {
		t.Fatal(err)
	}
	err := scaffoldApp(appDir, "supermarket", false)
	if err == nil {
		t.Fatal("expected error when supermarket already installed")
	}
}

func TestListAppTemplates_includesSupermarket(t *testing.T) {
	list := listAppTemplates()
	if len(list) == 0 {
		t.Fatal("expected at least one app template")
	}
	found := false
	for _, name := range list {
		if name == "supermarket" {
			found = true
		}
	}
	if !found {
		t.Errorf("listAppTemplates() = %v, want supermarket", list)
	}
}
