package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestScaffoldResource_FloatFields(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "markets")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "markets",
		ModulePath: "github.com/puppe1990/markets",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "store", resourceOpts{
		Fields: "name:string,lat:float,lng:float?",
		Seed:   true,
	}); err != nil {
		t.Fatal(err)
	}

	migrations, err := filepath.Glob(filepath.Join(appDir, "internal/store/migrations", "*_stores.sql"))
	if err != nil || len(migrations) == 0 {
		t.Fatal("missing stores migration")
	}
	migration, err := os.ReadFile(migrations[0])
	if err != nil {
		t.Fatal(err)
	}
	mig := string(migration)
	if !strings.Contains(mig, "lat REAL NOT NULL") {
		t.Errorf("migration missing lat REAL: %s", mig)
	}
	if !strings.Contains(mig, "lng REAL") || strings.Contains(mig, "lng REAL NOT NULL") {
		t.Errorf("migration lng should be nullable REAL: %s", mig)
	}

	model, err := os.ReadFile(filepath.Join(appDir, "internal/models/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	modelBody := string(model)
	if !strings.Contains(modelBody, "Lat") || !strings.Contains(modelBody, "float64") ||
		!strings.Contains(modelBody, "Lng") || !strings.Contains(modelBody, "*float64") {
		t.Errorf("model missing float fields: %s", model)
	}

	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_stores.go"))
	if err != nil {
		t.Fatal(err)
	}
	adminBody := string(admin)
	if !strings.Contains(adminBody, "strconv.ParseFloat") {
		t.Error("admin handler missing strconv.ParseFloat for float fields")
	}
	if strings.Contains(adminBody, `Lat: strings.TrimSpace`) {
		t.Error("admin handler should not assign float field from string TrimSpace")
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(store), "Lat: -25.4284") {
		t.Error("seed data should use float literal for lat")
	}

	form, err := os.ReadFile(filepath.Join(appDir, "web/src/pages/AdminStoreForm.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	formBody := string(form)
	if !strings.Contains(formBody, `bind:value={$form.Lat}`) || !strings.Contains(formBody, `type="float"`) {
		t.Error("admin svelte form should include lat float field")
	}
	if !strings.Contains(formBody, `$form.Lng`) {
		t.Error("admin svelte form should include lng float field")
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
	if !strings.Contains(adminBody, "ListCategoryOptions()") {
		t.Error("admin handler missing ListCategoryOptions call")
	}
	if !strings.Contains(adminBody, `"categoryOptions"`) {
		t.Error("admin handler should pass categoryOptions in inertia props")
	}

	form, err := os.ReadFile(filepath.Join(appDir, "web/src/pages/AdminBookmarkForm.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(form), "categoryOptions") {
		t.Error("admin svelte form missing category select options")
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
