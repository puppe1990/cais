package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	nav, err := os.ReadFile(filepath.Join(appDir, "web/src/components/AppLayout.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(nav)
	if !strings.Contains(body, "<!-- cais:nav -->") {
		t.Fatal("AppLayout.svelte missing <!-- cais:nav --> marker")
	}
	markerIdx := strings.Index(body, "<!-- cais:nav -->")
	linkIdx := strings.Index(body, `href="/products"`)
	if linkIdx == -1 {
		t.Fatal("AppLayout.svelte missing public products nav link")
	}
	if linkIdx < markerIdx {
		t.Error("nav link should appear after <!-- cais:nav --> marker")
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

	nav, err := os.ReadFile(filepath.Join(appDir, "web/src/components/AppLayout.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(nav)
	if !strings.Contains(body, `href="/books"`) {
		t.Error("AppLayout.svelte nav should include public books list link")
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

	svelte, err := os.ReadFile(filepath.Join(appDir, "web/src/pages/Tasks.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(svelte)
	if !strings.Contains(body, `Tasks`) {
		t.Error("public svelte page should use plural resource name Tasks")
	}
	if !strings.Contains(body, `{#each items`) {
		t.Error("public list should iterate items")
	}
	if !strings.Contains(body, `item.Title`) {
		t.Error("public list should render title field")
	}
	for _, needle := range []string{`use:inertia`, `export let items`} {
		if !strings.Contains(body, needle) {
			t.Errorf("public list missing HTMX UX attribute %q", needle)
		}
	}
}
