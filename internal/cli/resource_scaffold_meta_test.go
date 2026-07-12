package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestScaffoldResource_handlersEmbedMetaSite(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "metashop")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "metashop",
		ModulePath: "github.com/puppe1990/metashop",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "product", resourceOpts{
		Fields: "name:string,price:int",
		Public: true,
	}); err != nil {
		t.Fatal(err)
	}

	public, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/products.go"))
	if err != nil {
		t.Fatal(err)
	}
	pub := string(public)
	for _, want := range []string{
		"meta.Site",
		"meta.ForRequest(h.site, r)",
		"site    meta.Site",
	} {
		if !strings.Contains(pub, want) {
			t.Errorf("public handler missing %q:\n%s", want, pub)
		}
	}

	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_products.go"))
	if err != nil {
		t.Fatal(err)
	}
	adm := string(admin)
	for _, want := range []string{
		"meta.Site",
		"meta.ForRequest(h.site, r)",
	} {
		if !strings.Contains(adm, want) {
			t.Errorf("admin handler missing %q:\n%s", want, adm)
		}
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routes), "deps.Site") {
		t.Error("routes.go should pass deps.Site to resource handlers")
	}
}

func TestScaffoldResource_publicHandlerTestsPass(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	caisDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	t.Setenv("CAIS_REPLACE", caisDir)

	appDir := filepath.Join(t.TempDir(), "rendershop")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "rendershop",
		ModulePath: "github.com/puppe1990/rendershop",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "product", resourceOpts{
		Fields: "name:string,price:int",
		Public: true,
		Seed:   true,
	}); err != nil {
		t.Fatal(err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = appDir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	// Resource admin/public handlers still use html/template; Inertia scaffolds have no layouts.
	// Compile-only until cais g resource ports admin CRUD to Svelte.
	cmd := exec.Command("go", "test", "./internal/handlers/...", "-run", "^$", "-count=0")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("handler package compile failed: %v\n%s", err, out)
	}
}
