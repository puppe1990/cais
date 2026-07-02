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
