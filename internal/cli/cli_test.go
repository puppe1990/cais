package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Help(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "cais new") {
		t.Error("help missing cais new")
	}
}

func TestNames(t *testing.T) {
	data := dataForHandler("user_settings")
	if data.Pascal != "UserSettings" {
		t.Errorf("Pascal = %q", data.Pascal)
	}
	if data.Snake != "user_settings" {
		t.Errorf("Snake = %q", data.Snake)
	}
}

func TestCLI_Help_IncludesResource(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "g resource") {
		t.Error("help missing g resource")
	}
}

func TestCLI_NewMinimalCreatesSlimApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "slim")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "slim",
		ModulePath: "github.com/puppe1990/slim",
	}, true); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/handlers/home.go",
		"go.mod",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
		"internal/store/migrations/001_contacts.sql",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("minimal app should not have %s", path)
		}
	}
}

func TestScaffoldResource_CreatesCRUD(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "shop")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "shop",
		ModulePath: "github.com/puppe1990/shop",
	}, true); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldResource(appDir, "product", resourceOpts{}); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/models/product.go",
		"internal/handlers/admin_products.go",
		"internal/handlers/admin_products_test.go",
		"web/templates/pages/admin_products.html",
		"web/templates/pages/admin_product_form.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	storeBody, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(storeBody), "InsertProduct") {
		t.Error("store.go missing InsertProduct")
	}

	routesBody, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routesBody), "/admin/products") {
		t.Error("routes.go missing /admin/products")
	}
	if !strings.Contains(string(routesBody), "r.Group(middleware.TokenAuth") {
		t.Error("routes.go missing r.Group(middleware.TokenAuth")
	}
}

func TestScaffoldResource_PublicWithFields(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "links")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "links",
		ModulePath: "github.com/puppe1990/links",
	}, true); err != nil {
		t.Fatal(err)
	}

	opts := resourceOpts{Fields: "title:string,url:url,notes:text?", Public: true, Seed: true}
	if err := scaffoldResource(appDir, "bookmark", opts); err != nil {
		t.Fatal(err)
	}

	model, err := os.ReadFile(filepath.Join(appDir, "internal/models/bookmark.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(model), "URL") {
		t.Error("model missing URL field")
	}

	if _, err := os.Stat(filepath.Join(appDir, "internal/handlers/bookmarks.go")); err != nil {
		t.Error("missing public handler")
	}

	routes, _ := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if !strings.Contains(string(routes), `r.Get("/bookmarks"`) {
		t.Error("routes missing public list")
	}
}

func TestScaffoldResource_PluralPascal_ListAllMethod(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "recipes")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "recipes",
		ModulePath: "github.com/puppe1990/recipes",
	}, true); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "recipe", resourceOpts{Public: true}); err != nil {
		t.Fatal(err)
	}
	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_recipes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(admin), "ListAllRecipes()") {
		t.Errorf("admin handler wrong ListAll method: %s", admin)
	}

	publicHTML, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/recipes.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(publicHTML), `{{"{{"}}`) {
		t.Error("recipes.html has escaped template syntax")
	}
	if !strings.Contains(string(publicHTML), `{{ range .Items }}`) {
		t.Error("recipes.html missing valid template range")
	}
}

func TestCLI_NewIncludesHTMXAndAir(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "full")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "full",
		ModulePath: "github.com/puppe1990/full",
	}, false); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"web/static/js/htmx.min.js",
		".air.toml",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}
}

func TestScaffoldResource_IntFields(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "menu")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "menu",
		ModulePath: "github.com/puppe1990/menu",
	}, true); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "meal", resourceOpts{
		Fields: "title:string,prep_minutes:int,servings:int?",
		Seed:   true,
	}); err != nil {
		t.Fatal(err)
	}

	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_meals.go"))
	if err != nil {
		t.Fatal(err)
	}
	adminBody := string(admin)
	if !strings.Contains(adminBody, "strconv.ParseInt") {
		t.Error("admin handler missing strconv.ParseInt for int fields")
	}
	if strings.Contains(adminBody, `PrepMinutes: strings.TrimSpace`) {
		t.Error("admin handler should not assign int field from string TrimSpace")
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(store), "PrepMinutes: 30") {
		t.Error("seed data should use numeric literal for int fields")
	}

	adminTest, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_meals_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adminTest), "prep_minutes=30") {
		t.Error("admin test form body should use numeric value for int fields")
	}
}

func TestParseFields(t *testing.T) {
	fields, err := parseFields("title:string,url:url,notes:text?")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 {
		t.Fatalf("len = %d", len(fields))
	}
	if fields[2].Required {
		t.Error("notes should be optional")
	}
}

func TestPatchGoModReplace(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	parent := t.TempDir()
	appDir := filepath.Join(parent, "demo")
	caisDir := filepath.Join(parent, "Cais")
	if err := os.MkdirAll(caisDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(caisDir, "go.mod"), []byte("module github.com/puppe1990/cais\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "demo",
		ModulePath: "github.com/puppe1990/demo",
	}, true); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "replace github.com/puppe1990/cais => ../Cais") {
		t.Errorf("go.mod missing replace: %s", body)
	}
}

func TestCLI_NewCreatesApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "myapp")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "myapp",
		ModulePath: "github.com/puppe1990/myapp",
	}, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"go.mod",
		"cmd/server/main.go",
		"internal/handlers/dashboard.go",
		"web/templates/pages/dashboard.html",
		"web/static/manifest.webmanifest",
		"web/static/js/sw.js",
		"web/static/icons/icon-192.png",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}
}
