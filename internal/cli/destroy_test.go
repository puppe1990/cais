package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldMinimalApp_hasNoAuthOrphans(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "minapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "minapp",
		ModulePath: "github.com/puppe1990/minapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"internal/handlers/auth.go",
		"internal/models/user.go",
		"internal/store/migrations/002_auth.sql",
		"web/templates/pages/login.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("minimal app should not include %s", path)
		}
	}
}

func TestDestroyResource_removesGeneratedFiles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "destapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "destapp",
		ModulePath: "github.com/puppe1990/destapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields: "title:string,url:url",
		Seed:   false,
	}); err != nil {
		t.Fatal(err)
	}
	if err := destroyResource(appDir, "bookmark", false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/models/bookmark.go",
		"internal/handlers/admin_bookmarks.go",
		"web/templates/pages/admin_bookmarks.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("expected %s removed", path)
		}
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(routes), "/admin/bookmarks") {
		t.Error("routes.go still references admin bookmarks")
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(store), "InsertBookmark") {
		t.Error("store.go still has InsertBookmark")
	}
	if strings.Contains(string(routes), "adminBookmarks") {
		t.Error("routes.go still references adminBookmarks handler")
	}
	if strings.Contains(string(routes), "r.Group(") && strings.Contains(string(routes), "RequireAuth") {
		t.Error("routes.go still has orphan admin auth group")
	}
}

func TestDestroyModel_removesGeneratedFiles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "destmodel")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "destmodel",
		ModulePath: "github.com/puppe1990/destmodel",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldModel(appDir, "tag", modelOpts{Fields: "name:string"}); err != nil {
		t.Fatal(err)
	}
	if err := destroyModel(appDir, "tag", false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/models/tag.go",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("expected %s removed", path)
		}
	}

	migrationsDir := filepath.Join(appDir, "internal/store/migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_tags.sql") {
			t.Errorf("expected migration removed, still have %s", e.Name())
		}
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	storeBody := string(store)
	if strings.Contains(storeBody, "InsertTag") {
		t.Error("store.go still has InsertTag")
	}
	if strings.Contains(storeBody, "models.") {
		t.Error("store.go should drop unused models import after destroy model")
	}
}

func TestDestroyModel_dryRunWritesNothing(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "destmodeldry")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "destmodeldry",
		ModulePath: "github.com/puppe1990/destmodeldry",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldModel(appDir, "label", modelOpts{Fields: "name:string"}); err != nil {
		t.Fatal(err)
	}
	storeBefore, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}

	modelPath := filepath.Join(appDir, "internal/models/label.go")
	if _, err := os.Stat(modelPath); err != nil {
		t.Fatalf("model file should exist before dry-run: %v", err)
	}

	if err := destroyModel(appDir, "label", true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(modelPath); err != nil {
		t.Errorf("dry-run should not remove model file: %v", err)
	}
	storeAfter, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(storeAfter) != string(storeBefore) {
		t.Error("dry-run should not modify store.go")
	}
}

func TestDestroyAuth_removesGeneratedFiles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "destauth")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "destauth",
		ModulePath: "github.com/puppe1990/destauth",
	}, false, true); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldAuth(appDir, scaffoldData{
		AppName:    "destauth",
		ModulePath: "github.com/puppe1990/destauth",
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := destroyAuth(appDir, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/handlers/auth.go",
		"internal/models/user.go",
		"web/templates/pages/login.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("expected %s removed", path)
		}
	}

	store, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(store), "FindUserByEmail") {
		t.Error("store.go still has FindUserByEmail")
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	routesBody := string(routes)
	if strings.Contains(routesBody, "/login") || strings.Contains(routesBody, "NewAuthHandler") {
		t.Error("routes.go still references auth routes")
	}
	if strings.Contains(routesBody, "RequireAuthFunc") {
		t.Error("routes.go should not protect dashboard after destroy auth")
	}

	appGo, err := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(appGo), "LoadSession") {
		t.Error("blank app should keep LoadSession after destroy auth (baseline middleware)")
	}
}

func TestDestroyMigration_removesSQLFile(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "internal", "store", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_contacts.sql"), []byte("-- up\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldMigration(dir, "add_tags", false); err != nil {
		t.Fatal(err)
	}

	if err := destroyMigration(dir, "add_tags", false); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(migrationsDir, "001_contacts.sql")); err != nil {
		t.Fatal("existing migration should remain")
	}
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_add_tags.sql") {
			t.Errorf("expected add_tags migration removed, still have %s", e.Name())
		}
	}
}

func TestDestroyMigration_dryRun(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "internal", "store", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldMigration(dir, "posts", false); err != nil {
		t.Fatal(err)
	}
	migrationPath := filepath.Join(migrationsDir, "001_posts.sql")
	if _, err := os.Stat(migrationPath); err != nil {
		t.Fatal(err)
	}

	if err := destroyMigration(dir, "posts", true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(migrationPath); err != nil {
		t.Errorf("dry-run should not remove migration: %v", err)
	}
}

func TestRemoveMethodsFromStoreInterface_onlyTouchesStoreIface(t *testing.T) {
	in := `package store

type Other interface {
	InsertTag(models.Tag) (int64, error)
}

type Store interface {
	InsertTag(models.Tag) (int64, error)
	ListAllTags() ([]models.Tag, error)
	Ping() error
	Close() error
}
`
	out := removeMethodsFromStoreInterface(in, []string{"InsertTag", "ListAllTags"})
	if strings.Contains(out, "type Store interface {\n\tInsertTag") {
		t.Error("InsertTag should be removed from Store interface")
	}
	if strings.Contains(out, "type Store interface {\n\tListAllTags") {
		t.Error("ListAllTags should be removed from Store interface")
	}
	if !strings.Contains(out, "type Other interface {\n\tInsertTag") {
		t.Error("Other interface methods should remain")
	}
	if !strings.Contains(out, "Ping() error") {
		t.Error("unrelated Store methods should remain")
	}
}
