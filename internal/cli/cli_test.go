package cli

import (
	"bytes"
	"io"
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
	if !strings.Contains(buf.String(), "[--dry-run] resource") {
		t.Error("help missing g resource")
	}
}

func TestCLI_NewMinimalCreatesSlimApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "slim")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "slim",
		ModulePath: "github.com/puppe1990/slim",
	}, true, false); err != nil {
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

func TestScaffoldModel_CreatesModelAndStore(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "links")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "links",
		ModulePath: "github.com/puppe1990/links",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	opts := modelOpts{Fields: "title:string,url:url"}
	if err := scaffoldModel(appDir, "bookmark", opts); err != nil {
		t.Fatal(err)
	}

	model, err := os.ReadFile(filepath.Join(appDir, "internal/models/bookmark.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(model)
	if !strings.Contains(body, "Title") || !strings.Contains(body, "URL") {
		t.Error("model missing title or url fields")
	}

	migFiles, err := filepath.Glob(filepath.Join(appDir, "internal/store/migrations/*_bookmarks.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(migFiles) != 1 {
		t.Fatalf("expected 1 bookmarks migration, got %d", len(migFiles))
	}
	migBody, err := os.ReadFile(migFiles[0])
	if err != nil {
		t.Fatal(err)
	}
	mig := string(migBody)
	if !strings.Contains(mig, "CREATE TABLE") || !strings.Contains(mig, "url") {
		t.Error("migration missing expected SQL")
	}

	storeBody, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	store := string(storeBody)
	for _, method := range []string{
		"InsertBookmark",
		"UpdateBookmark",
		"DeleteBookmark",
		"FindBookmarkByID",
		"ListAllBookmarks",
	} {
		if !strings.Contains(store, method) {
			t.Errorf("store.go missing %s", method)
		}
	}

	for _, path := range []string{
		"internal/handlers/admin_bookmarks.go",
		"internal/handlers/bookmarks.go",
		"web/templates/pages/admin_bookmarks.html",
		"web/templates/pages/bookmarks.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); !os.IsNotExist(err) {
			t.Errorf("should not create %s", path)
		}
	}

	routesBody, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(routesBody), "/admin/bookmarks") {
		t.Error("routes should not be patched for model generator")
	}
}

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
	if !strings.Contains(string(routesBody), `middleware.RequireAuth("/login")`) {
		t.Error("routes.go missing middleware.RequireAuth(\"/login\")")
	}
	if strings.Contains(string(routesBody), "\n\n\n") {
		t.Error("routes.go has triple newlines (formatting issue)")
	}
}

func TestCLI_GenerateConsoleDryRun(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "cliconsoledry")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "cliconsoledry",
		ModulePath: "github.com/puppe1990/cliconsoledry",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(appDir, "cmd/console/main.go")); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	c := &CLI{Out: io.Discard}
	if err := c.Run([]string{"g", "--dry-run", "console"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(appDir, "cmd/console/main.go")); !os.IsNotExist(err) {
		t.Error("CLI --dry-run should not create cmd/console/main.go")
	}
}

func TestCLI_GenerateCIDryRun(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "clicidry")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "clicidry",
		ModulePath: "github.com/puppe1990/clicidry",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".github/workflows/ci.yml", ".golangci.yml"} {
		if err := os.Remove(filepath.Join(appDir, path)); err != nil {
			t.Fatal(err)
		}
	}
	makefileBefore, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	c := &CLI{Out: io.Discard}
	if err := c.Run([]string{"g", "--dry-run", "ci"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(appDir, ".github/workflows/ci.yml")); !os.IsNotExist(err) {
		t.Error("CLI --dry-run should not create ci.yml")
	}
	makefileAfter, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	if string(makefileAfter) != string(makefileBefore) {
		t.Error("CLI --dry-run should not modify Makefile")
	}
}

func TestCLI_GenerateResourceDryRun(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "clidryrun")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "clidryrun",
		ModulePath: "github.com/puppe1990/clidryrun",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	c := &CLI{Out: io.Discard}
	if err := c.Run([]string{"g", "--dry-run", "resource", "post", "--fields", "title:string"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(appDir, "internal/models/post.go")); !os.IsNotExist(err) {
		t.Error("CLI --dry-run should not create post.go")
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

func TestScaffoldResource_PublicWithFields(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "links")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "links",
		ModulePath: "github.com/puppe1990/links",
	}, true, false); err != nil {
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

func TestScaffoldResource_PublicInsertsNavAfterMarker(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "shop")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "shop",
		ModulePath: "github.com/puppe1990/shop",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "product", resourceOpts{Public: true}); err != nil {
		t.Fatal(err)
	}

	layout, err := os.ReadFile(filepath.Join(appDir, "web/templates/layouts/base.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(layout)
	if !strings.Contains(body, "<!-- cais:nav -->") {
		t.Fatal("layout missing <!-- cais:nav --> marker")
	}
	markerIdx := strings.Index(body, "<!-- cais:nav -->")
	linkIdx := strings.Index(body, `href="/products"`)
	if linkIdx == -1 {
		t.Fatal("layout missing public products nav link")
	}
	if linkIdx < markerIdx {
		t.Error("nav link should appear after <!-- cais:nav --> marker")
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
	}, false, false); err != nil {
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
	}, true, false); err != nil {
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
	if strings.Contains(adminBody, "PrepMinutes, err := strconv.ParseInt") {
		t.Error("admin handler should use lowercase variable name for strconv result, not PascalCase")
	}
	if !strings.Contains(adminBody, "prep_minutesVal") {
		t.Error("admin handler should use camelCase variable name for strconv result")
	}
	if !strings.Contains(adminBody, `"strconv"`) {
		t.Error("admin handler missing strconv import")
	}
	lines := strings.Split(adminBody, "\n")
	var stdlibImports []string
	inImport := false
	for _, line := range lines {
		if strings.HasPrefix(line, "import (") {
			inImport = true
			continue
		}
		if inImport {
			if strings.HasPrefix(line, ")") {
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, `"github.com`) && !strings.HasPrefix(trimmed, `"modernc.org`) {
				stdlibImports = append(stdlibImports, trimmed)
			}
		}
	}
	for i := 0; i < len(stdlibImports)-1; i++ {
		if stdlibImports[i] > stdlibImports[i+1] {
			t.Errorf("stdlib imports not sorted: %q > %q", stdlibImports[i], stdlibImports[i+1])
		}
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(store), "PrepMinutes: 30") {
		t.Error("seed data should use numeric literal for int fields")
	}
	if strings.Contains(string(store), "Demo ") {
		t.Error("seed data should use realistic values, not 'Demo X' pattern")
	}

	adminTest, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_meals_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adminTest), "prep_minutes=30") {
		t.Error("admin test form body should use numeric value for int fields")
	}
}

func TestScaffoldResource_BlankAppLogoLinksToPublicList(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "library")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "library",
		ModulePath: "github.com/puppe1990/library",
	}, false, true); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "book", resourceOpts{
		Fields: "title:string,url:url,pages:int,read:bool",
		Public: true,
		Seed:   true,
	}); err != nil {
		t.Fatal(err)
	}

	layout, err := os.ReadFile(filepath.Join(appDir, "web/templates/layouts/base.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(layout)
	if !strings.Contains(body, `<a href="/" class="font-bold`) {
		t.Error("blank app logo should link to welcome screen at /")
	}
	if !strings.Contains(body, `href="/books"`) {
		t.Error("layout nav should include public books list link")
	}
}

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

func TestScaffoldResource_PublicListRichFields(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "tasks")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "tasks",
		ModulePath: "github.com/puppe1990/tasks",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "task", resourceOpts{
		Fields: "title:string,done:bool,priority:int?,notes:text?",
		Public: true,
		Seed:   true,
	}); err != nil {
		t.Fatal(err)
	}

	html, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/tasks.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(html)
	if !strings.Contains(body, `{{ define "title" }}Tasks{{ end }}`) {
		t.Error("public page title should use plural resource name Tasks")
	}
	if !strings.Contains(body, `<h1 class="text-3xl font-bold text-slate-900 mb-6">Tasks</h1>`) {
		t.Error("public page h1 should use plural resource name")
	}
	if !strings.Contains(body, ".Done") {
		t.Error("public list should render done bool field")
	}
	if !strings.Contains(body, ".Priority") {
		t.Error("public list should render priority int field")
	}
	if !strings.Contains(body, ".Notes") {
		t.Error("public list should render notes text field")
	}
	for _, needle := range []string{"swap:150ms", `data-cais-optimistic="toggle"`} {
		if !strings.Contains(body, needle) {
			t.Errorf("public list missing HTMX UX attribute %q", needle)
		}
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

func TestScaffoldResource_Paginate(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "pages")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "pages",
		ModulePath: "github.com/puppe1990/pages",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "article", resourceOpts{
		Fields:   "title:string",
		Paginate: true,
	}); err != nil {
		t.Fatal(err)
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	storeBody := string(store)
	if !strings.Contains(storeBody, "ListArticles(page, perPage int) ([]models.Article, int, error)") {
		t.Error("store.go missing paginated ListArticles method")
	}
	if !strings.Contains(storeBody, "SELECT COUNT(*) FROM articles") {
		t.Error("store.go missing count query for pagination")
	}
	if !strings.Contains(storeBody, "LIMIT ? OFFSET ?") {
		t.Error("store.go missing LIMIT/OFFSET for pagination")
	}
	if !strings.Contains(storeBody, "ListAllArticles()") {
		t.Error("paginated resource should still include ListAllArticles for public handlers")
	}

	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_articles.go"))
	if err != nil {
		t.Fatal(err)
	}
	adminBody := string(admin)
	if strings.Contains(adminBody, "ListAllArticles()") {
		t.Error("paginated admin handler should not call ListAllArticles")
	}
	if !strings.Contains(adminBody, "ListArticles(page, perPage)") {
		t.Error("admin handler should call ListArticles with page and perPage")
	}
	if !strings.Contains(adminBody, `r.URL.Query().Get("page")`) {
		t.Error("admin handler should read page query param")
	}
	if !strings.Contains(adminBody, "perPage := 25") {
		t.Error("admin handler should default perPage to 25")
	}
	for _, needle := range []string{"Page", "Total", "PerPage", "HasPrev", "HasNext"} {
		if !strings.Contains(adminBody, needle) {
			t.Errorf("admin index data missing field %s", needle)
		}
	}

	html, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/admin_articles.html"))
	if err != nil {
		t.Fatal(err)
	}
	htmlBody := string(html)
	for _, needle := range []string{`{{ if .HasPrev }}`, `{{ if .HasNext }}`, `?page={{ .PrevPage }}`, `?page={{ .NextPage }}`} {
		if !strings.Contains(htmlBody, needle) {
			t.Errorf("admin template missing pagination control %q", needle)
		}
	}
}

func TestScaffoldResource_NoPaginate_UsesListAll(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "nopage")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "nopage",
		ModulePath: "github.com/puppe1990/nopage",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "note", resourceOpts{Fields: "title:string"}); err != nil {
		t.Fatal(err)
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	storeBody := string(store)
	if strings.Contains(storeBody, "ListNotes(page, perPage int)") {
		t.Error("non-paginated store should not have ListNotes(page, perPage)")
	}
	if !strings.Contains(storeBody, "ListAllNotes()") {
		t.Error("non-paginated store should have ListAllNotes")
	}

	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_notes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(admin), "ListAllNotes()") {
		t.Error("non-paginated admin handler should call ListAllNotes")
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

func TestScaffoldResource_BoolFields(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "tasks")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "tasks",
		ModulePath: "github.com/puppe1990/tasks",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "task", resourceOpts{
		Fields: "title:string,done:bool",
		Seed:   true,
	}); err != nil {
		t.Fatal(err)
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(store)
	if strings.Contains(body, "\n\tpublished int\n") || strings.Contains(body, "\tpublished int\n") {
		t.Error("bool scan temp must use var declaration, not bare published int")
	}
	if !strings.Contains(body, "var doneInt int") {
		t.Error("bool scan temp should be named after field: var doneInt int")
	}
	if !strings.Contains(body, "c.Done = doneInt == 1") {
		t.Error("bool assign should use field-specific temp var")
	}
	if strings.Contains(body, "published") {
		t.Error("should not hardcode published variable name for non-published bool fields")
	}
}

func TestPatchGoModReplace_CaisAppsLayout(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	caisDir := filepath.Join(root, "Cais")
	appsDir := filepath.Join(root, "Cais-apps", "demo")
	for _, d := range []string{caisDir, filepath.Dir(appsDir)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(caisDir, "go.mod"), []byte("module github.com/puppe1990/cais\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldNewApp(appsDir, scaffoldData{
		AppName:    "demo",
		ModulePath: "github.com/puppe1990/demo",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	mod, err := os.ReadFile(filepath.Join(appsDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "replace github.com/puppe1990/cais => ../../Cais") {
		t.Errorf("go.mod missing sibling Cais replace: %s", mod)
	}
}

func TestPatchGoModReplace_RemoteAppDirFromCwd(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	caisDir := filepath.Join(root, "Cais")
	appsDir := filepath.Join(root, "Cais-apps")
	appDir := filepath.Join(root, "remote", "testapp")
	for _, d := range []string{caisDir, appsDir, filepath.Dir(appDir)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(caisDir, "go.mod"), []byte("module github.com/puppe1990/cais\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(appsDir)
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "testapp",
		ModulePath: "github.com/puppe1990/testapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	mod, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	wantRel := filepath.Join("..", "..", "Cais")
	want := "replace github.com/puppe1990/cais => " + wantRel
	if !strings.Contains(string(mod), want) {
		t.Errorf("go.mod missing Cais replace from cwd layout:\nwant substring %q\ngot:\n%s", want, mod)
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

func TestParseFields_DateType(t *testing.T) {
	fields, err := parseFields("title:string,due_date:date")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("len = %d", len(fields))
	}
	if fields[1].GoType != "string" {
		t.Errorf("date GoType = %q, want string", fields[1].GoType)
	}
	if fields[1].HTMLType != "date" {
		t.Errorf("date HTMLType = %q, want date", fields[1].HTMLType)
	}
	if fields[1].SQLType != "TEXT NOT NULL" {
		t.Errorf("date SQLType = %q", fields[1].SQLType)
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
	}, true, false); err != nil {
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
	}, false, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"go.mod",
		"cmd/server/main.go",
		"internal/i18n/en.go",
		"internal/i18n/pt.go",
		".env.example",
		"internal/handlers/dashboard.go",
		"web/templates/pages/dashboard.html",
		"web/static/manifest.webmanifest",
		"web/static/js/sw.js",
		"web/static/js/cais.js",
		"web/static/img/go-on-cais.jpg",
		"web/static/og.png",
		"web/static/icons/icon.png",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}
}

func TestScaffold_InputCSSIncludesHTMXStyles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "styles")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "styles",
		ModulePath: "github.com/puppe1990/styles",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	css, err := os.ReadFile(filepath.Join(appDir, "input.css"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(css)
	for _, needle := range []string{".htmx-swapping", ".htmx-settling", ".htmx-indicator"} {
		if !strings.Contains(body, needle) {
			t.Errorf("input.css missing %q", needle)
		}
	}
}

func TestScaffold_IncludesQualityTooling(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")

	for _, tc := range []struct {
		name           string
		minimal, blank bool
	}{
		{"full", false, false},
		{"minimal", true, false},
		{"blank", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appDir := filepath.Join(t.TempDir(), tc.name)
			if err := scaffoldNewApp(appDir, scaffoldData{
				AppName:    tc.name,
				ModulePath: "github.com/puppe1990/" + tc.name,
			}, tc.minimal, tc.blank); err != nil {
				t.Fatal(err)
			}

			for _, path := range []string{
				".github/workflows/ci.yml",
				".pre-commit-config.yaml",
				".golangci.yml",
				".prettierrc.json",
				".prettierignore",
			} {
				if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
					t.Errorf("missing %s: %v", path, err)
				}
			}

			makefile, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
			if err != nil {
				t.Fatal(err)
			}
			body := string(makefile)
			for _, target := range []string{"lint:", "format-check:", "pre-commit-install:", "ci:"} {
				if !strings.Contains(body, target) {
					t.Errorf("Makefile missing target %s", target)
				}
			}

			golangci, err := os.ReadFile(filepath.Join(appDir, ".golangci.yml"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(golangci), "github.com/puppe1990/"+tc.name) {
				t.Error(".golangci.yml missing module local-prefix")
			}

			ci, err := os.ReadFile(filepath.Join(appDir, ".github/workflows/ci.yml"))
			if err != nil {
				t.Fatal(err)
			}
			ciBody := string(ci)
			for _, needle := range []string{"go test", "golangci-lint", "prettier", "npm test"} {
				if !strings.Contains(ciBody, needle) {
					t.Errorf("ci.yml missing %q", needle)
				}
			}

			pkg, err := os.ReadFile(filepath.Join(appDir, "package.json"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(pkg), `"test"`) {
				t.Error("package.json missing test script")
			}
		})
	}
}

func TestScaffoldNewApp_ContactHandlerValidatesName(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "contactapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "contactapp",
		ModulePath: "github.com/puppe1990/contactapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/contact.go"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, `errs.Add("name"`) {
		t.Errorf("contact handler missing name validation: %s", s)
	}
	if !strings.Contains(s, `contact.name_required`) {
		t.Errorf("contact handler missing name_required i18n key: %s", s)
	}
}

func TestScaffoldBlankApp_IncludesSecurityMiddleware(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "blankapp")
	if err := scaffoldNewApp(appDir, scaffoldData{AppName: "blankapp", ModulePath: "github.com/puppe1990/blankapp"}, false, true); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	s := string(body)
	for _, want := range []string{
		"middleware.Recover",
		"middleware.SecurityHeaders(cfg)",
		"ReadHeaderTimeout",
		"ReadTimeout",
		"r.Static",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("blank app missing %q in app.go", want)
		}
	}
}

func TestCLI_NewBlankCreatesEmptyApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "empty")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "empty",
		ModulePath: "github.com/puppe1990/empty",
	}, false, true); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"go.mod",
		"cmd/server/main.go",
		"internal/app/app.go",
		"internal/app/routes.go",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/home.go",
		"web/templates/pages/home.html",
		"web/templates/layouts/welcome.html",
		"web/templates/partials/cais_logo.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("blank app missing welcome screen file %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
		"internal/models/contact.go",
		"internal/store/migrations/001_contacts.sql",
		"web/templates/pages/contact.html",
		"web/templates/pages/dashboard.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("blank app should not have %s", path)
		}
	}

	routesBody, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routesBody), "home.ServeHTTP") {
		t.Error("blank app routes should register welcome home handler")
	}
}

func TestScaffoldNewApp_CustomModule(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "myapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "myapp",
		ModulePath: "github.com/acme/myapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/acme/myapp") {
		t.Errorf("go.mod missing custom module path: %s", body)
	}
}

func TestCLI_New_CustomModule(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	appDir := filepath.Join(root, "myapp")

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"new", "myapp", appDir, "--module", "github.com/acme/myapp"}); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/acme/myapp") {
		t.Errorf("go.mod missing custom module path: %s", body)
	}
}

func TestCLI_New_CustomModule_DefaultWhenOmitted(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	appDir := filepath.Join(root, "cool-app")

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"new", "cool-app", appDir}); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/puppe1990/coolapp") {
		t.Errorf("go.mod missing default module path: %s", body)
	}
}

func TestCLI_New_ModuleRequiresValue(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"new", "myapp", "--module"}); err == nil {
		t.Fatal("expected error for --module without value")
	}
}

func TestCLI_Help_IncludesModuleFlag(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "--module") {
		t.Error("help missing --module flag")
	}
}

func TestScaffoldMigration_usesMaxMigrationNumber(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "internal", "store", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"001_contacts.sql", "003_other.sql"} {
		if err := os.WriteFile(filepath.Join(migrationsDir, name), []byte("-- up\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := scaffoldMigration(dir, "posts", false); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(migrationsDir, "004_posts.sql")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected %s: %v", want, err)
	}
	if _, err := os.Stat(filepath.Join(migrationsDir, "003_posts.sql")); err == nil {
		t.Fatal("should not create 003_posts.sql when 003 is taken")
	}
}

func TestScaffoldMigration_numbersAfterSQLOnly(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "internal", "store", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_contacts.sql"), []byte("CREATE TABLE contacts (id INTEGER PRIMARY KEY);"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(migrationsDir, ".gitkeep"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldMigration(dir, "posts", false); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(migrationsDir, "002_posts.sql")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected %s: %v", want, err)
	}
}
