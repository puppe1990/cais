package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldResource_forceUpgradesRouteConstructors(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "legacyroutes")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "legacyroutes",
		ModulePath: "github.com/puppe1990/legacyroutes",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	opts := resourceOpts{Fields: "name:string,price:int", Public: true, Seed: false}
	if err := scaffoldResource(appDir, "product", opts); err != nil {
		t.Fatal(err)
	}

	routesPath := filepath.Join(appDir, "internal/app/routes.go")
	body, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatal(err)
	}
	legacy := strings.ReplaceAll(string(body),
		"handlers.NewProductsHandler(deps.Renderer, deps.Store, deps.Site, cfg)",
		"handlers.NewProductsHandler(deps.Renderer, deps.Store, cfg)",
	)
	legacy = strings.ReplaceAll(legacy,
		"handlers.NewAdminProductsHandler(deps.Renderer, deps.Store, deps.Site, cfg)",
		"handlers.NewAdminProductsHandler(deps.Renderer, deps.Store, cfg)",
	)
	if err := os.WriteFile(routesPath, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	opts.Force = true
	if err := scaffoldResource(appDir, "product", opts); err != nil {
		t.Fatal(err)
	}

	updated, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(updated)
	for _, want := range []string{
		"handlers.NewProductsHandler(deps.Renderer, deps.Store, deps.Site, cfg)",
		"handlers.NewAdminProductsHandler(deps.Renderer, deps.Store, deps.Site, cfg)",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("routes.go missing %q after --force:\n%s", want, s)
		}
	}
	if strings.Contains(s, "handlers.NewProductsHandler(deps.Renderer, deps.Store, cfg)") {
		t.Error("routes.go still has legacy public handler constructor")
	}
	if strings.Contains(s, "handlers.NewAdminProductsHandler(deps.Renderer, deps.Store, cfg)") {
		t.Error("routes.go still has legacy admin handler constructor")
	}
}
