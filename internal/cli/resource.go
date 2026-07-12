// cais g resource orchestration: writes generated files then delegates patches to resource_patch.go.
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
	data.Paginate = opts.Paginate
	data.AdminAuth = opts.AdminAuth

	migrationPath, migrationNum, err := nextMigrationFile(dir, data.Plural, opts.dryRun)
	if err != nil {
		return err
	}
	data.MigrationNum = migrationNum

	var files map[string]string
	if appUsesInertia(dir) {
		files = resourceFilesInertia(dir, data, migrationPath)
	} else {
		files = map[string]string{
			filepath.Join("internal/models", data.Snake+".go"):                               buildResourceModel(data),
			filepath.Join("internal/handlers", "admin_"+data.Plural+".go"):                   buildResourceAdminHandler(data),
			filepath.Join("internal/handlers", "admin_"+data.Plural+"_test.go"):              buildResourceAdminTest(data),
			filepath.Join("web/templates/pages", "admin_"+data.Plural+".html"):               buildAdminIndexHTML(data),
			filepath.Join("web/templates/pages", "admin_"+data.Snake+"_show.html"):           buildAdminShowHTML(data),
			filepath.Join("web/templates/pages", "admin_"+data.Snake+"_form.html"):           buildAdminFormHTML(data),
			filepath.Join("web/templates/partials", "admin_"+data.Snake+"_form_errors.html"): buildAdminFormErrorsPartial(data),
			migrationPath: buildResourceMigration(data),
		}
		if data.Paginate {
			files[filepath.Join("web/templates/partials", "admin_"+data.Plural+"_index.html")] = buildAdminIndexPartial(data)
		}
		if data.Public {
			files[filepath.Join("internal/handlers", data.Plural+".go")] = buildResourcePublicHandler(data)
			files[filepath.Join("internal/handlers", data.Plural+"_test.go")] = buildResourcePublicTest(data)
			files[filepath.Join("web/templates/pages", data.Plural+".html")] = buildPublicListHTML(data)
			togglePartial := buildPublicTogglePartial(data)
			if togglePartial != "" {
				files[filepath.Join("web/templates/partials", data.Plural+"_toggle.html")] = togglePartial
			}
			if data.Paginate {
				files[filepath.Join("web/templates/partials", data.Plural+"_list.html")] = buildPublicListPartial(data)
			}
		}
	}

	if hasReferenceFields(data.Fields) {
		selectPath := filepath.Join(dir, "internal/models/select_option.go")
		if _, err := os.Stat(selectPath); os.IsNotExist(err) {
			files["internal/models/select_option.go"] = tplSelectOptionModel
		}
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			if !opts.Force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", path)
			}
			if opts.dryRun {
				printfScaffold("update", path)
				continue
			}
		}
		if err := writeScaffoldFile(full, []byte(content), 0o644, path, opts.dryRun); err != nil {
			return err
		}
	}

	if err := patchStoreForResource(dir, data, opts.dryRun, opts.Force); err != nil {
		return err
	}
	if err := patchStoreTestForResource(dir, data, opts.dryRun); err != nil {
		return err
	}
	if err := patchRoutesForResource(dir, data, opts.dryRun, opts.Force); err != nil {
		return err
	}
	var finalErr error
	if data.Seed {
		if err := patchSeedsForResource(dir, data, opts.dryRun); err != nil {
			return err
		}
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

func TestAdmin%sHandler_Show(t *testing.T) {
	s := setupTestStore(t)
	id, err := s.Insert%s(models.%s{%s: "show-me"%s})
	if err != nil {
		t.Fatal(err)
	}
	h := NewAdmin%sHandler(setupTestRenderer(t), s, testSite(), cais.Config{})
	rr := httptest.NewRecorder()
	h.Show(rr, testutil.NewRequest(http.MethodGet, "/admin/%s/1", testutil.PathValue("id", "1")), id)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "show-me") {
		t.Error("missing item detail in show page")
	}
}

func TestAdmin%sHandler_Index(t *testing.T) {
	s := setupTestStore(t)
	h := NewAdmin%sHandler(setupTestRenderer(t), s, testSite(), cais.Config{})
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
	h := NewAdmin%sHandler(setupTestRenderer(t), s, testSite(), cais.Config{})
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
	h := NewAdmin%sHandler(setupTestRenderer(t), s, testSite(), cais.Config{})
	rr := httptest.NewRecorder()
	h.Delete(rr, testutil.NewRequest(http.MethodPost, "/admin/%s/1/delete", testutil.PathValue("id", "1")), id)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %%d", rr.Code)
	}
}
`,
		frameworkModule, frameworkModule, data.ModulePath,
		data.PluralPascal, data.Pascal, data.Pascal, first.Pascal, urlFieldTestExtra(data),
		data.PluralPascal, data.Plural,
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
		case "float64":
			val = "25.49"
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
	h := New%sHandler(setupTestRenderer(t), s, testSite(), cais.Config{})
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
