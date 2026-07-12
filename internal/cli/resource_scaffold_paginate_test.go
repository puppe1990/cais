package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	for _, needle := range []string{`"page"`, `"total"`, `"perPage"`, `"hasPrev"`, `"hasNext"`} {
		if !strings.Contains(adminBody, needle) {
			t.Errorf("admin index props missing %s", needle)
		}
	}
	for _, needle := range []string{`inertia.Render(w, r, "AdminArticles"`, `"hasPrev"`, `"nextPage"`} {
		if !strings.Contains(adminBody, needle) {
			t.Errorf("paginated admin handler missing %q", needle)
		}
	}

	svelte, err := os.ReadFile(filepath.Join(appDir, "web/src/pages/AdminArticles.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	svelteBody := string(svelte)
	for _, needle := range []string{`<table`, `hasPrev`, `nextPage`, `use:inertia`} {
		if !strings.Contains(svelteBody, needle) {
			t.Errorf("admin svelte page missing %q", needle)
		}
	}
}

func TestScaffoldResource_PublicPaginate(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "pubpages")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "pubpages",
		ModulePath: "github.com/puppe1990/pubpages",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "post", resourceOpts{
		Fields:   "title:string",
		Public:   true,
		Paginate: true,
	}); err != nil {
		t.Fatal(err)
	}

	handler, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/posts.go"))
	if err != nil {
		t.Fatal(err)
	}
	handlerBody := string(handler)
	if strings.Contains(handlerBody, "ListAllPosts()") {
		t.Error("paginated public handler should not call ListAllPosts")
	}
	for _, needle := range []string{"ListPosts(page, perPage)", `inertia.Render(w, r, "Posts"`} {
		if !strings.Contains(handlerBody, needle) {
			t.Errorf("public handler missing %q", needle)
		}
	}

	svelte, err := os.ReadFile(filepath.Join(appDir, "web/src/pages/Posts.svelte"))
	if err != nil {
		t.Fatal(err)
	}
	svelteBody := string(svelte)
	for _, needle := range []string{`{#each items`, `hasPrev`, `use:inertia`} {
		if !strings.Contains(svelteBody, needle) {
			t.Errorf("public svelte page missing %q", needle)
		}
	}

	if !strings.Contains(svelteBody, "nextPage") {
		t.Error("public svelte page should include nextPage pagination")
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
