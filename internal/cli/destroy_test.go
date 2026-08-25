package cli

import (
	"encoding/json"
	"go/parser"
	"go/token"
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
	if err := destroyResource(appDir, "bookmark", false, false); err != nil {
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
	if err := destroyModel(appDir, "tag", false, false); err != nil {
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

	if err := destroyModel(appDir, "label", true, false); err != nil {
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
		"web/src/pages/Login.svelte",
		"web/src/pages/Signup.svelte",
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

// #168: line-based unpatch dropped ANY line containing /admin/<plural> —
// including user comments and custom routes like /admin/bookmarks-export —
// and prefix matching removed lookalike methods (countBookmarksByUser).
// Removal must be statement/name exact via go/ast.
func TestUnpatchResourceRoutes_keepsUserLinesMentioningAdminPath(t *testing.T) {
	data := dataForResource("bookmark")
	content := `package app

import (
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"

	"example.com/app/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	r.Get("/", home.ServeHTTP)
	bookmarks := handlers.NewBookmarksHandler(deps.Store, deps.Site, deps.Inertia)
	r.Get("/bookmarks", bookmarks.List)
	adminBookmarks := handlers.NewAdminBookmarksHandler(deps.Store, deps.Site, deps.Inertia)
	r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
		g.Get("/admin/bookmarks", adminBookmarks.Index)
		g.Post("/admin/bookmarks/{id}", cais.IntParam("id", adminBookmarks.Update))
	})
	// TODO(#42): move /admin/bookmarks export behind a queue
	exporter := handlers.NewExportHandler(deps.Store)
	r.Get("/admin/bookmarks-export", exporter.ServeHTTP)
}
`
	out, err := unpatchResourceRoutes(content, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"/admin/bookmarks-export",
		"// TODO(#42): move /admin/bookmarks export behind a queue",
		"NewExportHandler",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("user code removed: %q missing from output:\n%s", want, out)
		}
	}
	for _, gone := range []string{"r.Group(", "adminBookmarks := ", "g.Get(\"/admin/bookmarks\",", `"bookmarks", bookmarks.List`} {
		if strings.Contains(out, gone) {
			t.Errorf("generated statement survived: %q in output:\n%s", gone, out)
		}
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "routes.go", out, parser.SkipObjectResolution); err != nil {
		t.Errorf("output does not parse: %v\n%s", err, out)
	}
}

func TestRemoveStoreResourceMethods_exactNamesOnly(t *testing.T) {
	data := dataForResource("bookmark")
	content := `package store

import "database/sql"

type Store interface {
	Ping() error
	Close() error
	InsertBookmark(b models.Bookmark) (int64, error)
	countBookmarks() (int, error)
}

type SQLiteStore struct{ db *sql.DB }

func (s *SQLiteStore) InsertBookmark(b models.Bookmark) (int64, error) { return 1, nil }

func (s *SQLiteStore) countBookmarks() (int, error) { return 0, nil }

func (s *SQLiteStore) countBookmarksByUser(userID int64) (int, error) { return 0, nil }
`
	out, err := removeStoreResourceMethods(content, data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "countBookmarksByUser") {
		t.Error("lookalike user method countBookmarksByUser must survive:\n" + out)
	}
	if strings.Contains(out, "countBookmarks()") || strings.Contains(out, "InsertBookmark") {
		t.Errorf("generated methods must be removed:\n%s", out)
	}
	if !strings.Contains(out, "Ping() error") || !strings.Contains(out, "Close() error") {
		t.Error("unrelated Store methods must survive:\n" + out)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "store.go", out, parser.SkipObjectResolution); err != nil {
		t.Errorf("output does not parse: %v\n%s", err, out)
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
	out, err := removeInterfaceMethods(in, "Store", map[string]bool{"InsertTag": true, "ListAllTags": true})
	if err != nil {
		t.Fatal(err)
	}
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

// #174: ignored os.ReadDir/read errors made partial destroys report success.
// Missing paths stay tolerated; real I/O errors must surface.
func TestDestroyResource_propagesReadErrors(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission errors unreliable as root")
	}
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "destrerr")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "destrerr",
		ModulePath: "github.com/puppe1990/destrerr",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{Fields: "title:string", Seed: false}); err != nil {
		t.Fatal(err)
	}

	migDir := filepath.Join(appDir, "internal/store/migrations")
	if err := os.Chmod(migDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(migDir, 0o755) })

	if err := destroyResource(appDir, "bookmark", false, false); err == nil {
		t.Error("destroyResource ignored unreadable migrations dir")
	}
}

func TestUnpatchSeedsForResource_propagatesReadErrors(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission errors unreliable as root")
	}
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "destrseed")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "destrseed",
		ModulePath: "github.com/puppe1990/destrseed",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	seedsPath := filepath.Join(appDir, "internal/db/seeds.go")
	if err := os.Chmod(seedsPath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(seedsPath, 0o644) })

	data := dataForResource("bookmark")
	if err := unpatchSeedsForResource(appDir, data, false); err == nil {
		t.Error("unpatchSeedsForResource ignored unreadable seeds.go")
	} else if os.IsNotExist(err) {
		t.Errorf("should tolerate missing file, got: %v", err)
	}
}

func TestDestroyResource_toleratesMissingOptionalFiles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "destrmiss")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "destrmiss",
		ModulePath: "github.com/puppe1990/destrmiss",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	data := dataForResource("bookmark")
	// No migrations dir, seeds or main seed block exist yet — must not error.
	if err := os.RemoveAll(filepath.Join(appDir, "internal/store/migrations")); err != nil {
		t.Fatal(err)
	}
	if err := unpatchSeedsForResource(appDir, data, false); err != nil {
		t.Errorf("missing seeds.go should be tolerated: %v", err)
	}
	if err := unpatchMainForSeed(appDir, data, false); err != nil {
		t.Errorf("missing main.go should be tolerated: %v", err)
	}
}

// #169: destroy wiped user-modified files silently. The generator records
// content hashes in .cais-generated.json; destroy warns and skips modified
// files unless --force is passed.
func TestScaffoldResource_recordsManifestHashes(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "manifestapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "manifestapp",
		ModulePath: "github.com/puppe1990/manifestapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{Fields: "title:string", Seed: false}); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(filepath.Join(appDir, ".cais-generated.json"))
	if err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
	var entries map[string]string
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("manifest not valid JSON map: %v", err)
	}
	for _, want := range []string{
		"internal/handlers/admin_bookmarks.go",
		"internal/models/bookmark.go",
	} {
		if _, ok := entries[want]; !ok {
			t.Errorf("manifest missing %q: %s", want, raw)
		}
	}
}

func TestDestroyResource_skipsModifiedFilesWithoutForce(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "modwarn")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "modwarn",
		ModulePath: "github.com/puppe1990/modwarn",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{Fields: "title:string", Seed: false}); err != nil {
		t.Fatal(err)
	}

	adminHandler := filepath.Join(appDir, "internal/handlers/admin_bookmarks.go")
	f, err := os.OpenFile(adminHandler, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("\n// custom export endpoint (#42)\n"); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	// Without force: modified file survives.
	if err := destroyResource(appDir, "bookmark", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(adminHandler); err != nil {
		t.Error("modified admin handler deleted without --force")
	}
	// Unmodified files are still removed.
	if _, err := os.Stat(filepath.Join(appDir, "internal/models/bookmark.go")); !os.IsNotExist(err) {
		t.Error("unmodified model should have been removed")
	}

	// With force: everything goes.
	if err := destroyResource(appDir, "bookmark", false, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(adminHandler); !os.IsNotExist(err) {
		t.Error("modified file survived --force destroy")
	}
}
