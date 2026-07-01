package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scaffoldResource(dir, name string, opts resourceOpts) error {
	fields, err := parseFields(opts.Fields)
	if err != nil {
		return err
	}

	data := dataForResource(name)
	data.ModulePath = readModulePath(dir)
	data.Fields = fields
	data.Public = opts.Public
	data.Seed = opts.Seed
	data.AdminAuth = opts.AdminAuth

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	if !opts.dryRun {
		if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}
	sqlCount := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlCount++
		}
	}
	data.MigrationNum = fmt.Sprintf("%03d", sqlCount+1)

	files := map[string]string{
		filepath.Join("internal/models", data.Snake+".go"):                                   buildResourceModel(data),
		filepath.Join("internal/handlers", "admin_"+data.Plural+".go"):                       buildResourceAdminHandler(data),
		filepath.Join("internal/handlers", "admin_"+data.Plural+"_test.go"):                  buildResourceAdminTest(data),
		filepath.Join("web/templates/pages", "admin_"+data.Plural+".html"):                   buildAdminIndexHTML(data),
		filepath.Join("web/templates/pages", "admin_"+data.Snake+"_form.html"):               buildAdminFormHTML(data),
		filepath.Join("internal/store/migrations", data.MigrationNum+"_"+data.Plural+".sql"): buildResourceMigration(data),
	}

	if data.Public {
		files[filepath.Join("internal/handlers", data.Plural+".go")] = buildResourcePublicHandler(data)
		files[filepath.Join("internal/handlers", data.Plural+"_test.go")] = buildResourcePublicTest(data)
		files[filepath.Join("web/templates/pages", data.Plural+".html")] = buildPublicListHTML(data)
		togglePartial := buildPublicTogglePartial(data)
		if togglePartial != "" {
			files[filepath.Join("web/templates/partials", data.Plural+"_toggle.html")] = togglePartial
		}
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := writeScaffoldFile(full, []byte(content), 0o644, path, opts.dryRun); err != nil {
			return err
		}
	}

	if err := patchStoreForResource(dir, data, opts.dryRun); err != nil {
		return err
	}
	if err := patchStoreTestForResource(dir, data, opts.dryRun); err != nil {
		return err
	}
	if err := patchRoutesForResource(dir, data, opts.dryRun); err != nil {
		return err
	}
	var finalErr error
	if data.Seed {
		finalErr = patchMainForSeed(dir, data, opts.dryRun)
	} else {
		finalErr = patchLayoutNav(dir, data, opts.dryRun)
	}
	if finalErr != nil {
		return finalErr
	}
	if opts.dryRun {
		return nil
	}
	return gofmtGoFiles(dir)
}

func buildResourceAdminTest(data scaffoldData) string {
	first := data.Fields[0]
	formBody := buildAdminTestFormBody(data.Fields)
	return fmt.Sprintf(`package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"%s/pkg/cais"
	"%s/pkg/cais/testutil"
	"%s/internal/models"
)

func TestAdmin%sHandler_Index(t *testing.T) {
	s := setupTestStore(t)
	h := NewAdmin%sHandler(setupTestRenderer(t), s, cais.Config{})
	rr := httptest.NewRecorder()
	h.Index(rr, httptest.NewRequest(http.MethodGet, "/admin/%s", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("status = %%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "admin-%s") {
		t.Error("missing admin table")
	}
}

func TestAdmin%sHandler_Create(t *testing.T) {
	s := setupTestStore(t)
	h := NewAdmin%sHandler(setupTestRenderer(t), s, cais.Config{})
	req := httptest.NewRequest(http.MethodPost, "/admin/%s", strings.NewReader(%q))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %%d", rr.Code)
	}
}

func TestAdmin%sHandler_Delete(t *testing.T) {
	s := setupTestStore(t)
	id, err := s.Insert%s(models.%s{%s: "x"%s})
	if err != nil {
		t.Fatal(err)
	}
	h := NewAdmin%sHandler(setupTestRenderer(t), s, cais.Config{})
	rr := httptest.NewRecorder()
	h.Delete(rr, testutil.NewRequest(http.MethodPost, "/admin/%s/1/delete", testutil.PathValue("id", "1")), id)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %%d", rr.Code)
	}
}
`,
		frameworkModule, frameworkModule, data.ModulePath,
		data.PluralPascal, data.PluralPascal, data.Plural, data.Plural,
		data.PluralPascal, data.PluralPascal, data.Plural, formBody,
		data.PluralPascal, data.Pascal, data.Pascal, first.Pascal, urlFieldTestExtra(data),
		data.PluralPascal, data.Plural,
	)
}

func buildAdminTestFormBody(fields []FieldDef) string {
	var parts []string
	for _, f := range fields {
		if !f.Required || f.GoType == "bool" {
			continue
		}
		val := "Demo"
		switch f.GoType {
		case "int64":
			val = "30"
		default:
			if f.HTMLType == "url" {
				val = "https://example.com"
			}
			if f.Widget == "textarea" {
				val = "Sample " + f.Pascal
			}
		}
		parts = append(parts, f.Name+"="+val)
	}
	if len(parts) == 0 && len(fields) > 0 {
		return fields[0].Name + "=Demo"
	}
	return strings.Join(parts, "&")
}

func urlFieldTestExtra(data scaffoldData) string {
	for _, f := range data.Fields {
		if f.HTMLType == "url" {
			return fmt.Sprintf(", %s: \"https://example.com\"", f.Pascal)
		}
	}
	return ""
}

func buildResourcePublicTest(data scaffoldData) string {
	seedCall := ""
	if data.Seed {
		seedCall = `	if err := s.SeedDemo` + data.PluralPascal + `(); err != nil {
		t.Fatal(err)
	}
`
	}
	return fmt.Sprintf(`package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"%s/pkg/cais"
)

func Test%sHandler_List(t *testing.T) {
	s := setupTestStore(t)
%s
	h := New%sHandler(setupTestRenderer(t), s, cais.Config{})
	rr := httptest.NewRecorder()
	h.List(rr, httptest.NewRequest(http.MethodGet, "/%s", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("status = %%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "%s-list") {
		t.Error("missing public list")
	}
}
`, frameworkModule, data.PluralPascal, seedCall, data.PluralPascal, data.Plural, data.Plural)
}

func patchStoreForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "Insert"+data.Pascal) {
		return nil
	}

	ifaceMarker := "\n\tClose() error"
	if !strings.Contains(content, ifaceMarker) {
		return fmt.Errorf("could not patch store interface")
	}
	ifaceInsert := fmt.Sprintf(
		"\n\tInsert%s(models.%s) (int64, error)\n\tUpdate%s(models.%s) error\n\tDelete%s(id int64) error\n\tFind%sByID(id int64) (models.%s, error)\n\tListAll%s() ([]models.%s, error)",
		data.Pascal, data.Pascal,
		data.Pascal, data.Pascal,
		data.Pascal,
		data.Pascal, data.Pascal,
		data.PluralPascal, data.Pascal,
	)
	if data.Seed {
		ifaceInsert += fmt.Sprintf("\n\tSeedDemo%s() error", data.PluralPascal)
	}
	content = strings.Replace(content, ifaceMarker, ifaceInsert+ifaceMarker, 1)

	implMarker := "\nfunc (s *SQLiteStore) Close()"
	implInsert := buildResourceStoreMethods(data)
	if data.Seed {
		implInsert += buildResourceSeed(data)
	}
	if hasBoolField(data.Fields) && !strings.Contains(content, "func boolInt(") {
		implInsert = "\nfunc boolInt(v bool) int {\n\tif v {\n\t\treturn 1\n\t}\n\treturn 0\n}\n" + implInsert
	}
	content = strings.Replace(content, implMarker, implInsert+implMarker, 1)

	if !strings.Contains(content, data.ModulePath+"/internal/models") {
		content = strings.Replace(content,
			`_ "modernc.org/sqlite"`,
			`"`+data.ModulePath+`/internal/models"
	_ "modernc.org/sqlite"`,
			1,
		)
	}

	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func patchStoreTestForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store_test.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "TestStore_Insert"+data.Pascal) {
		return nil
	}

	insertArgs := buildInsertTestLiteral(data.Fields)
	insert := fmt.Sprintf(`
func TestStore_Insert%s(t *testing.T) {
	s := newTestStore(t)
	id, err := s.Insert%s(models.%s{%s})
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Error("id = 0")
	}
}
`, data.Pascal, data.Pascal, data.Pascal, insertArgs)

	if !strings.Contains(content, data.ModulePath+"/internal/models") {
		content = strings.Replace(content,
			`import "testing"`,
			`import (
	"testing"

	"`+data.ModulePath+`/internal/models"
)`,
			1,
		)
	}
	content = strings.TrimRight(content, "\n") + "\n" + insert
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func buildInsertTestLiteral(fields []FieldDef) string {
	var parts []string
	for _, f := range fields {
		if !f.Required {
			continue
		}
		parts = append(parts, f.Pascal+": "+seedValueForField(f))
	}
	if len(parts) == 0 && len(fields) > 0 {
		return fields[0].Pascal + ": " + seedValueForField(fields[0])
	}
	return strings.Join(parts, ", ")
}

func patchRoutesForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "/admin/"+data.Plural) {
		return nil
	}

	if !strings.Contains(content, frameworkModule+"/pkg/cais/middleware") {
		content = strings.Replace(content,
			`"`+frameworkModule+`/pkg/cais"`,
			`"`+frameworkModule+`/pkg/cais"
	"`+frameworkModule+`/pkg/cais/middleware"`,
			1,
		)
	}

	adminVar := "admin" + data.PluralPascal
	var insert strings.Builder
	if data.Public {
		pubVar := lowerFirst(data.PluralPascal)
		fmt.Fprintf(&insert, "\t%s := handlers.New%sHandler(deps.Renderer, deps.Store, cfg)\n", pubVar, data.PluralPascal)
		fmt.Fprintf(&insert, "\tr.Get(\"/%s\", %s.List)\n", data.Plural, pubVar)
		if firstBoolField(data.Fields) != nil {
			fmt.Fprintf(&insert, "\tr.Post(\"/%s/{id}/toggle\", cais.IntParam(\"id\", %s.Toggle))\n", data.Plural, pubVar)
		}
	}
	fmt.Fprintf(&insert, "\t%s := handlers.NewAdmin%sHandler(deps.Renderer, deps.Store, cfg)\n", adminVar, data.PluralPascal)
	if data.AdminAuth == "bearer" {
		fmt.Fprintf(&insert, "\tr.Group(middleware.AdminAuth(cfg), func(g *cais.Router) {\n")
	} else {
		fmt.Fprintf(&insert, "\tr.Group(middleware.RequireAuth(\"/login\"), func(g *cais.Router) {\n")
	}
	fmt.Fprintf(&insert, "\t\tg.Get(\"/admin/%s\", %s.Index)\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Get(\"/admin/%s/new\", %s.New)\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Post(\"/admin/%s\", %s.Create)\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Get(\"/admin/%s/{id}/edit\", cais.IntParam(\"id\", %s.Edit))\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Post(\"/admin/%s/{id}\", cais.IntParam(\"id\", %s.Update))\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Post(\"/admin/%s/{id}/delete\", cais.IntParam(\"id\", %s.Delete))\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t})\n")

	var err2 error
	content, err2 = insertBeforeFunctionEnd(content, "registerRoutes", insert.String())
	if err2 != nil {
		return fmt.Errorf("could not patch routes.go: %w", err2)
	}
	if err := updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun); err != nil {
		return err
	}
	return patchLayoutNav(dir, data, dryRun)
}

const layoutNavMarker = "<!-- cais:nav -->"

func patchLayoutNav(dir string, data scaffoldData, dryRun bool) error {
	if !data.Public {
		return nil
	}
	path := filepath.Join(dir, "web/templates/layouts/base.html")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	linkHref := `href="/` + data.Plural + `"`
	if strings.Contains(content, linkHref) {
		return nil
	}
	link := fmt.Sprintf(`          <a href="/%s" class="text-slate-600 hover:text-indigo-600 transition">%s</a>
`, data.Plural, toTitle(data.Plural))
	switch {
	case strings.Contains(content, layoutNavMarker):
		content = strings.Replace(content, layoutNavMarker, layoutNavMarker+"\n"+link, 1)
	case strings.Contains(content, "</nav>"):
		content = strings.Replace(content, "</nav>", link+"        </nav>", 1)
	default:
		return fmt.Errorf("%s: missing %s marker and </nav> element", path, layoutNavMarker)
	}
	content = patchLayoutLogoHref(dir, content, data)
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func patchLayoutLogoHref(dir, content string, data scaffoldData) string {
	routes, err := os.ReadFile(filepath.Join(dir, "internal/app/routes.go"))
	if err != nil {
		return content
	}
	if strings.Contains(string(routes), `r.Get("/", home`) {
		return content
	}
	return strings.Replace(content,
		`<a href="/" class="font-bold`,
		fmt.Sprintf(`<a href="/%s" class="font-bold`, data.Plural),
		1,
	)
}

func patchMainForSeed(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "cmd/server/main.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "SeedDemo"+data.PluralPascal) {
		return patchLayoutNav(dir, data, dryRun)
	}
	marker := "\n\tstaticDir, err := findWebDir(\"static\")"
	seed := fmt.Sprintf(`
	if err := s.SeedDemo%s(); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("seed: %%w", err)
	}
`, data.PluralPascal)
	if !strings.Contains(content, marker) {
		return fmt.Errorf("could not patch main.go for seed")
	}
	content = strings.Replace(content, marker, seed+marker, 1)
	if err := updateScaffoldFile(path, []byte(content), "cmd/server/main.go", dryRun); err != nil {
		return err
	}
	return patchLayoutNav(dir, data, dryRun)
}
