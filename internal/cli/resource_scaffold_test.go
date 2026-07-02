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
		"web/templates/pages/admin_products.html",
		"web/templates/pages/admin_product_show.html",
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

func TestScaffoldResource_ReferencesField(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "library")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "library",
		ModulePath: "github.com/puppe1990/library",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields: "title:string,category_id:references",
	}); err != nil {
		t.Fatal(err)
	}

	migrations, err := filepath.Glob(filepath.Join(appDir, "internal/store/migrations", "*_bookmarks.sql"))
	if err != nil || len(migrations) == 0 {
		t.Fatal("missing bookmarks migration")
	}
	migration, err := os.ReadFile(migrations[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(migration), "REFERENCES categories(id)") {
		t.Errorf("migration missing FK:\n%s", migration)
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(store), "ListCategoryOptions()") {
		t.Error("store missing ListCategoryOptions for references field")
	}

	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_bookmarks.go"))
	if err != nil {
		t.Fatal(err)
	}
	adminBody := string(admin)
	if !strings.Contains(adminBody, "CategoryOptions []forms.SelectOption") {
		t.Error("admin form data missing CategoryOptions")
	}
	if !strings.Contains(adminBody, "ListCategoryOptions()") {
		t.Error("admin handler missing ListCategoryOptions call")
	}

	form, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/admin_bookmark_form.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(form), "fieldSelect") {
		t.Error("admin form template missing fieldSelect for references field")
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
