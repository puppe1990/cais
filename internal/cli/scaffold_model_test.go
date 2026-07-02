package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
