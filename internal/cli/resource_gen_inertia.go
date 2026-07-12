// Inertia + Svelte generation for cais g resource (default on Inertia scaffolds).
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func appUsesInertia(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "web/src/pages/Home.svelte"))
	return err == nil
}

func buildAdminIndexSvelte(data scaffoldData) string {
	display := adminIndexDisplayField(data.Fields)
	pagination := ""
	if data.Paginate {
		pagination = `
  {#if hasPrev}<a href="/admin/` + data.Plural + `?page={prevPage}" use:inertia>← Previous</a>{/if}
  <span>Page {page}</span>
  {#if hasNext}<a href="/admin/` + data.Plural + `?page={nextPage}" use:inertia>Next →</a>{/if}`
	}
	return fmt.Sprintf(`<script>
  import { inertia } from '@inertiajs/svelte'
  export let items = []
  export let site = {}
%s  function deleteItem(id) {
    if (!confirm('Delete?')) return
    const f = document.createElement('form')
    f.method = 'POST'
    f.action = '/admin/%s/' + id + '/delete'
    document.body.appendChild(f)
    f.submit()
  }
</script>

<div class="max-w-3xl mx-auto p-6">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-semibold">%s</h1>
    <a href="/admin/%s/new" use:inertia class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm">New</a>
  </div>
  <table class="w-full text-sm border rounded-lg overflow-hidden">
    <thead class="bg-slate-50"><tr><th class="px-4 py-2 text-left">%s</th><th class="px-4 py-2 text-right">Actions</th></tr></thead>
    <tbody>
      {#each items as item}
      <tr class="border-t">
        <td class="px-4 py-2">{item.%s}</td>
        <td class="px-4 py-2 text-right space-x-2">
          <a href="/admin/%s/{item.id}" use:inertia class="text-indigo-600">View</a>
          <a href="/admin/%s/{item.id}/edit" use:inertia class="text-indigo-600">Edit</a>
          <button type="button" class="text-red-600" on:click={() => deleteItem(item.id)}>Delete</button>
        </td>
      </tr>
      {/each}
    </tbody>
  </table>%s
</div>
`, inertiaPaginationProps(data), data.Plural, data.Title, data.Plural, display.Pascal, display.Pascal, data.Plural, data.Plural, pagination)
}

func inertiaPaginationProps(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return `
  export let page = 1
  export let hasPrev = false
  export let hasNext = false
  export let prevPage = 1
  export let nextPage = 1`
}

func buildAdminShowSvelte(data scaffoldData) string {
	display := adminIndexDisplayField(data.Fields)
	return fmt.Sprintf(`<script>
  import { inertia } from '@inertiajs/svelte'
  export let item = {}
  export let site = {}
</script>

<div class="max-w-md mx-auto p-6">
  <a href="/admin/%s" use:inertia class="text-sm text-indigo-600">← Back</a>
  <h1 class="text-2xl font-semibold mt-4">{item.%s}</h1>
  <a href="/admin/%s/{item.id}/edit" use:inertia class="mt-4 inline-block text-indigo-600">Edit</a>
</div>
`, data.Plural, display.Pascal, data.Plural)
}

func buildSvelteFormField(f FieldDef) string {
	label := f.Pascal
	errBlock := fmt.Sprintf(`{#if errors.%s}<p class="text-red-600 text-sm">{errors.%s}</p>{/if}`, f.Name, f.Name)
	switch f.Widget {
	case "textarea":
		return fmt.Sprintf(`    <label class="block text-sm font-medium">%s</label>
    <textarea bind:value={$form.%s} class="block w-full border rounded p-2"></textarea>
    %s
`, label, f.Pascal, errBlock)
	case "checkbox":
		return fmt.Sprintf(`    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" bind:checked={$form.%s} class="rounded" /> %s
    </label>
    %s
`, f.Pascal, label, errBlock)
	case "select":
		optVar := f.RefPascal + "Options"
		return fmt.Sprintf(`    <label class="block text-sm font-medium">%s</label>
    <select bind:value={$form.%s} class="block w-full border rounded p-2">
      {#each %s as opt}<option value={opt.value}>{opt.label}</option>{/each}
    </select>
    %s
`, label, f.Pascal, optVar, errBlock)
	default:
		inputType := f.HTMLType
		if inputType == "" {
			inputType = "text"
		}
		return fmt.Sprintf(`    <label class="block text-sm font-medium">%s</label>
    <input type="%s" bind:value={$form.%s} class="block w-full border rounded p-2" />
    %s
`, label, inputType, f.Pascal, errBlock)
	}
}

func buildAdminFormSvelte(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		fields.WriteString(buildSvelteFormField(f))
		fields.WriteString("\n")
	}
	optsExports := ""
	optsInit := ""
	for _, f := range data.Fields {
		if f.RefTable == "" {
			continue
		}
		optVar := lowerFirst(f.RefPascal) + "Options"
		optsExports += fmt.Sprintf("\n  export let %s = []", optVar)
		optsInit += fmt.Sprintf("\n    %s: item.%s != null ? String(item.%s) : '',", f.Pascal, f.Pascal, f.Pascal)
	}
	formInit := buildSvelteFormInit(data.Fields)
	if optsInit != "" {
		formInit = strings.TrimSuffix(formInit, "  })") + optsInit + "\n  })"
	}
	return fmt.Sprintf(`<script>
  import { useForm } from '@inertiajs/svelte'
  export let item = {}
  export let isNew = true
  export let errors = {}
  export let site = {}
%s
  let form = useForm({
%s
  })
  function submit() {
    if (isNew) $form.post('/admin/%s')
    else $form.post('/admin/%s/' + item.id)
  }
</script>

<div class="max-w-md mx-auto p-6">
  <a href="/admin/%s" use:inertia class="text-sm text-indigo-600">← Back</a>
  <h1 class="text-2xl font-semibold mt-4 mb-4">{isNew ? 'New %s' : 'Edit %s'}</h1>
  <form on:submit|preventDefault={submit} class="space-y-3">
%s
    <button type="submit" class="w-full bg-indigo-600 text-white py-2 rounded-lg">{isNew ? 'Create' : 'Save'}</button>
  </form>
</div>
`, optsExports, formInit, data.Plural, data.Plural, data.Plural, data.Title, data.Title, fields.String())
}

func buildSvelteFormInit(fields []FieldDef) string {
	var lines []string
	for _, f := range fields {
		switch f.GoType {
		case "bool":
			lines = append(lines, fmt.Sprintf("    %s: item.%s ?? false,", f.Pascal, f.Pascal))
		case "int64", "float64":
			lines = append(lines, fmt.Sprintf("    %s: item.%s != null ? String(item.%s) : '',", f.Pascal, f.Pascal, f.Pascal))
		case "*int64", "*float64", "*string":
			lines = append(lines, fmt.Sprintf("    %s: item.%s ?? '',", f.Pascal, f.Pascal))
		default:
			lines = append(lines, fmt.Sprintf("    %s: item.%s ?? '',", f.Pascal, f.Pascal))
		}
	}
	return strings.Join(lines, "\n")
}

func buildInertiaIndexMethod(data scaffoldData) string {
	component := "Admin" + data.PluralPascal
	if data.Paginate {
		return fmt.Sprintf(`func (h *Admin%sHandler) Index(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	perPage := 25
	items, total, err := h.store.List%s(page, perPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pg := pagination.New(page, perPage, total)
	_ = h.inertia.Render(w, r, %q, h.indexProps(r, items, pg))
}`, data.PluralPascal, data.PluralPascal, component)
	}
	return fmt.Sprintf(`func (h *Admin%sHandler) Index(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAll%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.inertia.Render(w, r, %q, h.indexProps(r, items))
}`, data.PluralPascal, data.PluralPascal, component)
}

func buildInertiaShowMethod(data scaffoldData) string {
	component := "Admin" + data.Pascal + "Show"
	return fmt.Sprintf(`func (h *Admin%sHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.inertia.Render(w, r, %q, inertia.Props{
		"site": meta.ForRequest(h.site, r),
		"item": item,
	})
}`, data.PluralPascal, data.Pascal, component)
}

func buildInertiaFormPropsMethod(data scaffoldData) string {
	var optionLoads []string
	for _, f := range data.Fields {
		if f.RefTable == "" {
			continue
		}
		optVar := f.RefPascal + "Options"
		propKey := lowerFirst(f.RefPascal) + "Options"
		optionLoads = append(optionLoads, fmt.Sprintf(`	if raw%s, err := h.store.List%sOptions(); err == nil {
		opts := make([]map[string]string, 0, len(raw%s))
		for _, o := range raw%s {
			opts = append(opts, map[string]string{"value": strconv.FormatInt(o.ID, 10), "label": o.Label})
		}
		props[%q] = opts
	}`, optVar, f.RefPascal, optVar, optVar, propKey))
	}
	optsBlock := ""
	if len(optionLoads) > 0 {
		optsBlock = "\n" + strings.Join(optionLoads, "\n")
	}
	return fmt.Sprintf(`func (h *Admin%sHandler) formProps(r *http.Request, item models.%s, isNew bool) inertia.Props {
	props := inertia.Props{
		"site":  meta.ForRequest(h.site, r),
		"item":  item,
		"isNew": isNew,
	}%s
	return props
}`, data.PluralPascal, data.Pascal, optsBlock)
}

func buildInertiaIndexPropsMethod(data scaffoldData) string {
	if data.Paginate {
		return fmt.Sprintf(`func (h *Admin%sHandler) indexProps(r *http.Request, items []models.%s, pg pagination.Page) inertia.Props {
	return inertia.Props{
		"site":     meta.ForRequest(h.site, r),
		"items":    items,
		"page":     pg.Page,
		"total":    pg.Total,
		"perPage":  pg.PerPage,
		"hasPrev":  pg.HasPrev,
		"hasNext":  pg.HasNext,
		"prevPage": pg.PrevPage,
		"nextPage": pg.NextPage,
	}
}`, data.PluralPascal, data.Pascal)
	}
	return fmt.Sprintf(`func (h *Admin%sHandler) indexProps(r *http.Request, items []models.%s) inertia.Props {
	return inertia.Props{
		"site":  meta.ForRequest(h.site, r),
		"items": items,
	}
}`, data.PluralPascal, data.Pascal)
}

func buildInertiaValidationRender(data scaffoldData, isNew bool) string {
	component := "Admin" + data.Pascal + "Form"
	isNewLit := "false"
	if isNew {
		isNewLit = "true"
	}
	return fmt.Sprintf(`		ve := make(inertia.ValidationErrors)
		for k, v := range errs {
			ve[k] = v
		}
		ctx := inertia.SetValidationErrors(r.Context(), ve)
		_ = h.inertia.Render(w, r.WithContext(ctx), %q, h.formProps(r, item, %s))`, component, isNewLit)
}

func buildResourceAdminInertiaHandler(data scaffoldData) string {
	parse := buildAdminParseForm(data)
	hasStrconv := needsStrconv(data.Fields) || data.Paginate || hasReferenceFields(data.Fields)
	paginationImport := ""
	if data.Paginate {
		paginationImport = "\t\"" + frameworkModule + "/pkg/cais/pagination\"\n"
	}
	formComponent := "Admin" + data.Pascal + "Form"
	strconvImport := ""
	if hasStrconv {
		strconvImport = "\t\"strconv\"\n"
	}
	return fmt.Sprintf(`package handlers

import (
	"net/http"
%s	"strings"

	inertia "github.com/romsar/gonertia/v3"
	"%s/pkg/cais/validate"
%s
	"%s/pkg/cais/meta"
	"%s/internal/models"
	"%s/internal/store"
)

type Admin%sHandler struct {
	store   store.Store
	site    meta.Site
	inertia *inertia.Inertia
}

func NewAdmin%sHandler(s store.Store, site meta.Site, i *inertia.Inertia) *Admin%sHandler {
	return &Admin%sHandler{store: s, site: site, inertia: i}
}

%s

%s

%s

%s

func (h *Admin%sHandler) New(w http.ResponseWriter, r *http.Request) {
	_ = h.inertia.Render(w, r, %q, h.formProps(r, models.%s{}, true))
}

func (h *Admin%sHandler) Edit(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.inertia.Render(w, r, %q, h.formProps(r, item, false))
}

func (h *Admin%sHandler) Create(w http.ResponseWriter, r *http.Request) {
	item, errs := h.parseForm(r)
	if errs.Any() {
%s
		return
	}
	if _, err := h.store.Insert%s(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.inertia.Redirect(w, r, "/admin/%s", http.StatusSeeOther)
}

func (h *Admin%sHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	item, errs := h.parseForm(r)
	item.ID = id
	if errs.Any() {
%s
		return
	}
	if err := h.store.Update%s(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.inertia.Redirect(w, r, "/admin/%s", http.StatusSeeOther)
}

func (h *Admin%sHandler) Delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.Delete%s(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.inertia.Redirect(w, r, "/admin/%s", http.StatusSeeOther)
}

func (h *Admin%sHandler) parseForm(r *http.Request) (models.%s, validate.FieldErrors) {
	%s
}
`,
		strconvImport,
		frameworkModule,
		paginationImport,
		frameworkModule,
		data.ModulePath, data.ModulePath,
		data.PluralPascal,
		data.PluralPascal, data.PluralPascal, data.PluralPascal,
		buildInertiaIndexMethod(data),
		buildInertiaShowMethod(data),
		buildInertiaFormPropsMethod(data),
		buildInertiaIndexPropsMethod(data),
		data.PluralPascal, formComponent, data.Pascal,
		data.PluralPascal, data.Pascal, formComponent,
		data.PluralPascal, buildInertiaValidationRender(data, true),
		data.Pascal, data.Plural,
		data.PluralPascal, buildInertiaValidationRender(data, false),
		data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, parse,
	)
}

func resourceFilesInertia(dir string, data scaffoldData, migrationPath string) map[string]string {
	files := map[string]string{
		filepath.Join("internal/models", data.Snake+".go"):                  buildResourceModel(data),
		filepath.Join("internal/handlers", "admin_"+data.Plural+".go"):      buildResourceAdminInertiaHandler(data),
		filepath.Join("internal/handlers", "admin_"+data.Plural+"_test.go"): buildResourceAdminInertiaTest(data),
		filepath.Join("web/src/pages", "Admin"+data.PluralPascal+".svelte"): buildAdminIndexSvelte(data),
		filepath.Join("web/src/pages", "Admin"+data.Pascal+"Form.svelte"):   buildAdminFormSvelte(data),
		filepath.Join("web/src/pages", "Admin"+data.Pascal+"Show.svelte"):   buildAdminShowSvelte(data),
		migrationPath: buildResourceMigration(data),
	}
	if data.Public {
		files[filepath.Join("internal/handlers", data.Plural+".go")] = buildResourcePublicInertiaHandler(data)
		files[filepath.Join("internal/handlers", data.Plural+"_test.go")] = buildResourcePublicInertiaTest(data)
		files[filepath.Join("web/src/pages", data.PluralPascal+".svelte")] = buildPublicListSvelte(data)
	}
	return files
}

func buildResourceAdminInertiaTest(data scaffoldData) string {
	first := data.Fields[0]
	formBody := buildAdminTestFormBody(data.Fields)
	return fmt.Sprintf(`package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"%s/internal/models"
)

func TestAdmin%sHandler_Index_Inertia(t *testing.T) {
	s := setupTestStore(t)
	if _, err := s.Insert%s(models.%s{%s: "idx"}); err != nil {
		t.Fatal(err)
	}
	h := NewAdmin%sHandler(s, testSite(), setupTestInertia(t))
	req := inertiaRequest(http.MethodGet, "/admin/%s", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	assertInertiaComponent(t, rr, "Admin%s")
}

func TestAdmin%sHandler_Create_Inertia(t *testing.T) {
	s := setupTestStore(t)
	h := NewAdmin%sHandler(s, testSite(), setupTestInertia(t))
	req := inertiaRequest(http.MethodPost, "/admin/%s", strings.NewReader(%q))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %%d", rr.Code)
	}
}
`, data.ModulePath, data.PluralPascal, data.Pascal, data.Pascal, first.Pascal, data.PluralPascal, data.Plural, data.PluralPascal, data.PluralPascal, data.PluralPascal, data.Plural, formBody)
}

func buildResourcePublicInertiaHandler(data scaffoldData) string {
	listMethod := buildPublicListInertiaMethod(data)
	extraImports := ""
	if data.Paginate {
		extraImports = fmt.Sprintf("\t\"strconv\"\n\t\"%s/pkg/cais/pagination\"\n", frameworkModule)
	}
	return fmt.Sprintf(`package handlers

import (
	"net/http"
%s
	inertia "github.com/romsar/gonertia/v3"
	"%s/pkg/cais/meta"
	"%s/internal/store"
)

type %sHandler struct {
	store   store.Store
	site    meta.Site
	inertia *inertia.Inertia
}

func New%sHandler(s store.Store, site meta.Site, i *inertia.Inertia) *%sHandler {
	return &%sHandler{store: s, site: site, inertia: i}
}

%s
`, extraImports, frameworkModule, data.ModulePath, data.PluralPascal, data.PluralPascal, data.PluralPascal, data.PluralPascal, listMethod)
}

func buildPublicListInertiaMethod(data scaffoldData) string {
	component := data.PluralPascal
	if data.Paginate {
		return fmt.Sprintf(`func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	perPage := 25
	items, total, err := h.store.List%s(page, perPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pg := pagination.New(page, perPage, total)
	_ = h.inertia.Render(w, r, %q, inertia.Props{
		"site": meta.ForRequest(h.site, r), "items": items,
		"page": pg.Page, "hasPrev": pg.HasPrev, "hasNext": pg.HasNext,
		"prevPage": pg.PrevPage, "nextPage": pg.NextPage,
	})
}`, data.PluralPascal, data.PluralPascal, component)
	}
	return fmt.Sprintf(`func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAll%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.inertia.Render(w, r, %q, inertia.Props{
		"site": meta.ForRequest(h.site, r), "items": items,
	})
}`, data.PluralPascal, data.PluralPascal, component)
}

func buildPublicListSvelte(data scaffoldData) string {
	display := adminIndexDisplayField(data.Fields)
	pagination := ""
	paginationProps := ""
	if data.Paginate {
		paginationProps = `
  export let page = 1
  export let hasPrev = false
  export let hasNext = false
  export let prevPage = 1
  export let nextPage = 1`
		pagination = fmt.Sprintf(`
  <div class="flex justify-between mt-4 text-sm">
    {#if hasPrev}<a href="/%s?page={prevPage}" use:inertia>← Previous</a>{/if}
    <span>Page {page}</span>
    {#if hasNext}<a href="/%s?page={nextPage}" use:inertia>Next →</a>{/if}
  </div>`, data.Plural, data.Plural)
	}
	return fmt.Sprintf(`<script>
  import { inertia } from '@inertiajs/svelte'
  export let items = []
  export let site = {}
%s
</script>

<div class="max-w-2xl mx-auto p-6">
  <h1 class="text-2xl font-semibold mb-4">%s</h1>
  <a href="/" use:inertia class="sr-only">Home</a>
  <ul class="space-y-2">
    {#each items as item}
    <li class="border rounded p-3">{item.%s}</li>
    {/each}
  </ul>%s
</div>
`, paginationProps, data.PluralPascal, display.Pascal, pagination)
}

func buildResourcePublicInertiaTest(data scaffoldData) string {
	return fmt.Sprintf(`package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test%sHandler_List_Inertia(t *testing.T) {
	s := setupTestStore(t)
	h := New%sHandler(s, testSite(), setupTestInertia(t))
	req := inertiaRequest(http.MethodGet, "/%s", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)
	assertInertiaComponent(t, rr, %q)
}
`, data.PluralPascal, data.PluralPascal, data.Plural, data.PluralPascal)
}
