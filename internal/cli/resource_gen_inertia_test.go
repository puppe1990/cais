package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestScaffoldResource_Inertia_generatesSvelteAdmin(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	caisDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	t.Setenv("CAIS_REPLACE", caisDir)

	appDir := filepath.Join(t.TempDir(), "inertiashop")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "inertiashop",
		ModulePath: "github.com/puppe1990/inertiashop",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "product", resourceOpts{}); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"web/src/pages/AdminProducts.svelte",
		"web/src/pages/AdminProductForm.svelte",
		"web/src/pages/AdminProductShow.svelte",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"web/templates/pages/admin_products.html",
		"web/templates/pages/admin_product_form.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("legacy HTMX template should not exist: %s", path)
		}
	}

	adminGo, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_products.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(adminGo)
	for _, want := range []string{
		"gonertia/v3",
		`inertia.Render(w, r, "AdminProducts"`,
		`inertia.Render(w, r, "AdminProductForm"`,
		"inertia.SetValidationErrors",
		"h.inertia.Redirect",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("admin handler missing %q", want)
		}
	}
	if strings.Contains(body, "RenderPageOrPartial") {
		t.Error("inertia admin should not use RenderPageOrPartial")
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routes), "deps.Inertia") {
		t.Error("routes.go should pass deps.Inertia to admin handler")
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = appDir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	cmd := exec.Command("go", "test", "./internal/handlers/...", "-run", "^$", "-count=0")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("handler package compile failed: %v\n%s", err, out)
	}
}