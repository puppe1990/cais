package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldResource_CreatesCRUD(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "shop")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "shop",
		ModulePath: "github.com/puppe1990/shop",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldResource(appDir, "product", resourceOpts{}); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/models/product.go",
		"internal/handlers/admin_products.go",
		"internal/handlers/admin_products_test.go",
		"web/src/pages/AdminProducts.svelte",
		"web/src/pages/AdminProductForm.svelte",
		"web/src/pages/AdminProductShow.svelte",
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
	if !strings.Contains(string(routesBody), `cais.IntParam("id", adminProducts.Show)`) {
		t.Error("routes.go missing admin show route")
	}
	if !strings.Contains(string(routesBody), `middleware.RequireAuth("/login")`) {
		t.Error("routes.go missing middleware.RequireAuth(\"/login\")")
	}
	if strings.Contains(string(routesBody), "\n\n\n") {
		t.Error("routes.go has triple newlines (formatting issue)")
	}
}

func TestScaffoldResource_adminInertiaDefaults(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "inertiashop2")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "inertiashop2",
		ModulePath: "github.com/puppe1990/inertiashop2",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "widget", resourceOpts{}); err != nil {
		t.Fatal(err)
	}

	indexSvelte, err := os.ReadFile(filepath.Join(appDir, "web/src/pages/AdminWidgets.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	indexBody := string(indexSvelte)
	for _, want := range []string{`use:inertia`, `/admin/widgets/new`, `deleteItem`} {
		if !strings.Contains(indexBody, want) {
			t.Errorf("admin index missing %q", want)
		}
	}

	formSvelte, err := os.ReadFile(filepath.Join(appDir, "web/src/pages/AdminWidgetForm.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	formBody := string(formSvelte)
	if !strings.Contains(formBody, `useForm`) || !strings.Contains(formBody, `errors.`) {
		t.Error("admin form should use useForm and errors prop")
	}

	adminGo, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_widgets.go"))
	if err != nil {
		t.Fatal(err)
	}
	adminBody := string(adminGo)
	if !strings.Contains(adminBody, "inertia.SetValidationErrors") {
		t.Error("admin handler should use inertia.SetValidationErrors on validation errors")
	}
	if strings.Contains(adminBody, "IsHTMX") {
		t.Error("inertia admin should not reference IsHTMX")
	}
}

func TestScaffoldResource_DryRunWritesNothing(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "dryrun")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "dryrun",
		ModulePath: "github.com/puppe1990/dryrun",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	storeBefore, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	routesBefore, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}

	opts := resourceOpts{Fields: "name:string", dryRun: true}
	if err := scaffoldResource(appDir, "item", opts); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(appDir, "internal/models/item.go")); !os.IsNotExist(err) {
		t.Error("dry-run should not create item.go")
	}

	storeAfter, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(storeAfter) != string(storeBefore) {
		t.Error("dry-run should not modify store.go")
	}

	routesAfter, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(routesAfter) != string(routesBefore) {
		t.Error("dry-run should not modify routes.go")
	}
}

func TestScaffoldResource_DefaultAdminAuthUsesRequireAuth(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "items")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "items", ModulePath: "github.com/puppe1990/items",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "item", resourceOpts{Fields: "name:string"}); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	s := string(body)
	if !strings.Contains(s, `middleware.RequireAuth("/login")`) {
		t.Errorf("routes should use RequireAuth for session admin: %s", s)
	}
	if strings.Contains(s, "middleware.AdminAuth(cfg)") {
		t.Error("default should not use AdminAuth")
	}
}

func TestScaffoldResource_AdminAuthBearerFlag(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "apiitems")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "apiitems", ModulePath: "github.com/puppe1990/apiitems",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "item", resourceOpts{
		Fields: "name:string", AdminAuth: "bearer",
	}); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if !strings.Contains(string(body), "middleware.AdminAuth(cfg)") {
		t.Error("bearer flag should use AdminAuth")
	}
}

func TestScaffoldResource_PluralPascal_ListAllMethod(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "recipes")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "recipes",
		ModulePath: "github.com/puppe1990/recipes",
	}, true, false); err != nil {
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

	publicSvelte, err := os.ReadFile(filepath.Join(appDir, "web/src/pages/Recipes.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(publicSvelte), `{#each items`) {
		t.Error("Recipes.svelte missing items loop")
	}
}

func TestParseResourceOpts_Paginate(t *testing.T) {
	opts, err := parseResourceOpts([]string{"--paginate", "--fields", "title:string"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.Paginate {
		t.Error("Paginate should be true with --paginate flag")
	}
	if opts.Fields != "title:string" {
		t.Errorf("Fields = %q", opts.Fields)
	}

	opts, err = parseResourceOpts([]string{"--fields", "title:string"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Paginate {
		t.Error("Paginate should default to false")
	}
}

func TestScaffoldResource_DishPluralization(t *testing.T) {
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
	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_dishes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(admin), "ListAllDishes()") {
		t.Error("dish resource should pluralize to dishes, not dishs")
	}
}
