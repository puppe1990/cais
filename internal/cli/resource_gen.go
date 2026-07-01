package cli

import (
	"fmt"
	"strings"
)

type resourceOpts struct {
	Fields string
	Public bool
	Seed   bool
}

func parseResourceOpts(args []string) (resourceOpts, error) {
	opts := resourceOpts{Seed: true}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--fields":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--fields requires a value")
			}
			i++
			opts.Fields = args[i]
		case "--public":
			opts.Public = true
		case "--no-seed":
			opts.Seed = false
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opts, nil
}

func buildResourceModel(data scaffoldData) string {
	var b strings.Builder
	b.WriteString("package models\n\nimport \"time\"\n\n")
	fmt.Fprintf(&b, "type %s struct {\n\tID int64\n", data.Pascal)
	for _, f := range data.Fields {
		fmt.Fprintf(&b, "\t%s %s\n", f.Pascal, f.GoType)
	}
	b.WriteString("\tCreatedAt time.Time\n}\n")
	return b.String()
}

func buildResourceMigration(data scaffoldData) string {
	var cols []string
	for _, f := range data.Fields {
		cols = append(cols, fmt.Sprintf("    %s %s", f.Name, f.SQLType))
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n    id INTEGER PRIMARY KEY AUTOINCREMENT,\n%s,\n    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n);\n",
		data.Plural, strings.Join(cols, ",\n"))
}

func buildResourceStoreMethods(data scaffoldData) string {
	cols, ph := insertColumns(data.Fields)
	args := insertArgs(data.Fields)
	sets := updateSets(data.Fields)
	updArgs := insertArgs(data.Fields) + ", c.ID"
	sel := selectColumns(data.Fields)

	return fmt.Sprintf(`
func (s *SQLiteStore) Insert%s(c models.%s) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO %s (%s) VALUES (%s)",
		%s,
	)
	if err != nil {
		return 0, fmt.Errorf("insert %s: %%w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) Update%s(c models.%s) error {
	_, err := s.db.Exec(
		"UPDATE %s SET %s WHERE id = ?",
		%s,
	)
	if err != nil {
		return fmt.Errorf("update %s: %%w", err)
	}
	return nil
}

func (s *SQLiteStore) Delete%s(id int64) error {
	_, err := s.db.Exec("DELETE FROM %s WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete %s: %%w", err)
	}
	return nil
}

func (s *SQLiteStore) Find%sByID(id int64) (models.%s, error) {
	var c models.%s
%s
	err := s.db.QueryRow(
		"SELECT id, %s, created_at FROM %s WHERE id = ?",
		id,
	).Scan(%s)
	if err != nil {
		return models.%s{}, fmt.Errorf("find %s: %%w", err)
	}
%s
	return c, nil
}

func (s *SQLiteStore) ListAll%s() ([]models.%s, error) {
	rows, err := s.db.Query("SELECT id, %s, created_at FROM %s ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("list %s: %%w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.%s
	for rows.Next() {
		var c models.%s
%s
		if err := rows.Scan(%s); err != nil {
			return nil, fmt.Errorf("scan %s: %%w", err)
		}
%s
		items = append(items, c)
	}
	return items, rows.Err()
}
`,
		data.Pascal, data.Pascal, data.Plural, cols, ph, args, data.Snake,
		data.Pascal, data.Pascal, data.Plural, sets, updArgs, data.Snake,
		data.Pascal, data.Plural, data.Snake,
		data.Pascal, data.Pascal, data.Pascal, scanDeclare(data.Fields), sel, data.Plural, scanVars(data.Fields), data.Pascal, data.Snake, scanAssign(data.Fields),
		data.PluralPascal, data.Pascal, sel, data.Plural, data.Plural,
		data.Pascal, data.Pascal, scanLoopDeclare(data.Fields), scanVars(data.Fields), data.Snake, scanLoopAssign(data.Fields),
	)
}

func buildResourceSeed(data scaffoldData) string {
	if !data.Seed {
		return ""
	}
	var inserts []string
	for _, f := range data.Fields {
		inserts = append(inserts, fmt.Sprintf("%s: %s", f.Pascal, seedValueForField(f)))
	}
	body := fmt.Sprintf("models.%s{%s}", data.Pascal, strings.Join(inserts, ", "))
	return fmt.Sprintf(`
func (s *SQLiteStore) SeedDemo%s() error {
	count, err := s.count%s()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err = s.Insert%s(%s)
	return err
}

func (s *SQLiteStore) count%s() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM %s").Scan(&count)
	return count, err
}
`, data.PluralPascal, data.PluralPascal, data.Pascal, body, data.PluralPascal, data.Plural)
}

func seedValueForField(f FieldDef) string {
	name := strings.ToLower(f.Name)
	switch f.GoType {
	case "bool":
		if strings.Contains(name, "active") || strings.Contains(name, "enabled") {
			return "true"
		}
		return "false"
	case "int64":
		if strings.Contains(name, "count") || strings.Contains(name, "total") {
			return "10"
		}
		if strings.Contains(name, "price") || strings.Contains(name, "amount") {
			return "99"
		}
		if strings.Contains(name, "age") || strings.Contains(name, "year") {
			return "25"
		}
		if strings.Contains(name, "rating") || strings.Contains(name, "score") {
			return "5"
		}
		if strings.Contains(name, "minute") || strings.Contains(name, "hour") || strings.Contains(name, "second") || strings.Contains(name, "duration") || strings.Contains(name, "calorie") {
			return "30"
		}
		if strings.Contains(name, "quantity") || strings.Contains(name, "qty") || strings.Contains(name, "servings") {
			return "4"
		}
		return "1"
	default:
		if f.HTMLType == "url" {
			if strings.Contains(name, "github") {
				return `"https://github.com/example"`
			}
			if strings.Contains(name, "twitter") || strings.Contains(name, "x") {
				return `"https://twitter.com/example"`
			}
			return `"https://example.com"`
		}
		if f.Widget == "textarea" {
			if strings.Contains(name, "description") {
				return `"A detailed description of this item."`
			}
			if strings.Contains(name, "notes") || strings.Contains(name, "comment") {
				return `"Some notes about this entry."`
			}
			return `"Lorem ipsum dolor sit amet, consectetur adipiscing elit."`
		}
		if f.HTMLType == "date" {
			return `"2024-01-15"`
		}
		if strings.Contains(name, "email") {
			return `"user@example.com"`
		}
		if strings.Contains(name, "name") || strings.Contains(name, "title") {
			return `"Sample Item"`
		}
		if strings.Contains(name, "status") {
			return `"active"`
		}
		if strings.Contains(name, "category") {
			return `"general"`
		}
		return `"Sample"`
	}
}

func insertColumns(fields []FieldDef) (cols, placeholders string) {
	names := fieldNames(fields)
	ph := make([]string, len(names))
	for i := range names {
		ph[i] = "?"
	}
	return strings.Join(names, ", "), strings.Join(ph, ", ")
}

func insertArgs(fields []FieldDef) string {
	var args []string
	for _, f := range fields {
		if f.GoType == "bool" {
			args = append(args, "boolInt(c."+f.Pascal+")")
		} else {
			args = append(args, "c."+f.Pascal)
		}
	}
	return strings.Join(args, ", ")
}

func updateSets(fields []FieldDef) string {
	var sets []string
	for _, f := range fields {
		sets = append(sets, f.Name+" = ?")
	}
	return strings.Join(sets, ", ")
}

func selectColumns(fields []FieldDef) string {
	return strings.Join(fieldNames(fields), ", ")
}

func fieldNames(fields []FieldDef) []string {
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	return names
}

func boolScanTemp(f FieldDef) string {
	return f.Name + "Int"
}

func scanDeclare(fields []FieldDef) string {
	var extra []string
	for _, f := range fields {
		if f.GoType == "bool" {
			extra = append(extra, "\tvar "+boolScanTemp(f)+" int")
		}
	}
	if len(extra) == 0 {
		return ""
	}
	return strings.Join(extra, "\n") + "\n"
}

func scanLoopDeclare(fields []FieldDef) string {
	var extra []string
	for _, f := range fields {
		if f.GoType == "bool" {
			extra = append(extra, "\t\tvar "+boolScanTemp(f)+" int")
		}
	}
	if len(extra) == 0 {
		return ""
	}
	return strings.Join(extra, "\n") + "\n"
}

func scanVars(fields []FieldDef) string {
	var vars []string
	for _, f := range fields {
		if f.GoType == "bool" {
			vars = append(vars, "&"+boolScanTemp(f))
		} else {
			vars = append(vars, "&c."+f.Pascal)
		}
	}
	vars = append(vars, "&c.CreatedAt")
	return "&c.ID, " + strings.Join(vars, ", ")
}

func scanAssign(fields []FieldDef) string {
	var lines []string
	for _, f := range fields {
		if f.GoType == "bool" {
			lines = append(lines, "\tc."+f.Pascal+" = "+boolScanTemp(f)+" == 1")
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func scanLoopAssign(fields []FieldDef) string {
	var lines []string
	for _, f := range fields {
		if f.GoType == "bool" {
			lines = append(lines, "\t\tc."+f.Pascal+" = "+boolScanTemp(f)+" == 1")
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func buildAdminParseForm(data scaffoldData) string {
	var literal []string
	var after []string
	var validations []string
	for _, f := range data.Fields {
		switch f.GoType {
		case "bool":
			literal = append(literal, fmt.Sprintf("%s: r.FormValue(%q) == \"on\"", f.Pascal, f.Name))
		case "int64":
			if f.Required {
				after = append(after, fmt.Sprintf(`raw%s := strings.TrimSpace(r.FormValue(%q))
	if raw%s == "" {
		return models.%s{}, fmt.Errorf(%q)
	}
	%sVal, err := strconv.ParseInt(raw%s, 10, 64)
	if err != nil {
		return models.%s{}, fmt.Errorf(%q)
	}
	item.%s = %sVal`, f.Pascal, f.Name, f.Pascal, data.Pascal, f.Name+" is required", f.Name, f.Pascal, data.Pascal, f.Name+" must be a number", f.Pascal, f.Name))
			} else {
				after = append(after, fmt.Sprintf(`if raw%s := strings.TrimSpace(r.FormValue(%q)); raw%s != "" {
		%sVal, err := strconv.ParseInt(raw%s, 10, 64)
		if err != nil {
			return models.%s{}, fmt.Errorf(%q)
		}
		item.%s = %sVal
	}`, f.Pascal, f.Name, f.Pascal, f.Name, f.Pascal, data.Pascal, f.Name+" must be a number", f.Pascal, f.Name))
			}
		default:
			literal = append(literal, fmt.Sprintf("%s: strings.TrimSpace(r.FormValue(%q))", f.Pascal, f.Name))
			if f.Required {
				if f.HTMLType == "url" {
					validations = append(validations, fmt.Sprintf("if err := validate.URL(item.%s); err != nil {\n\t\treturn models.%s{}, err\n\t}", f.Pascal, data.Pascal))
				} else {
					validations = append(validations, fmt.Sprintf("if item.%s == \"\" {\n\t\treturn models.%s{}, fmt.Errorf(%q)\n\t}", f.Pascal, data.Pascal, f.Name+" is required"))
				}
			}
		}
	}
	validateBlock := ""
	if len(validations) > 0 {
		validateBlock = "\n\t" + strings.Join(validations, "\n\t") + "\n"
	}
	afterBlock := ""
	if len(after) > 0 {
		afterBlock = "\n\t" + strings.Join(after, "\n\t") + "\n"
	}
	return fmt.Sprintf(`item := models.%s{%s}%s%s	return item, nil`, data.Pascal, strings.Join(literal, ", "), validateBlock, afterBlock)
}

func needsStrconv(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "int64" {
			return true
		}
	}
	return false
}

func hasBoolField(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "bool" {
			return true
		}
	}
	return false
}

func needsValidate(fields []FieldDef) bool {
	for _, f := range fields {
		if f.HTMLType == "url" && f.Required {
			return true
		}
	}
	return false
}

func boolImport(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}

func buildResourceAdminHandler(data scaffoldData) string {
	parse := buildAdminParseForm(data)
	hasValidate := needsValidate(data.Fields)
	hasStrconv := needsStrconv(data.Fields)
	return fmt.Sprintf(`package handlers

import (
	"fmt"
	"net/http"
%s	"strings"
%s
	"%s/pkg/cais"
	"%s/pkg/cais/csrf"
	"%s/pkg/cais/httpx"
	"%s/internal/models"
	"%s/internal/store"
)

type Admin%sHandler struct {
	renderer *cais.Renderer
	store    store.Store
}

type Admin%sIndexData struct {
	CSRFToken string
	Items     []models.%s
}

type Admin%sFormData struct {
	CSRFToken string
	Item      models.%s
	IsNew     bool
}

func NewAdmin%sHandler(renderer *cais.Renderer, s store.Store) *Admin%sHandler {
	return &Admin%sHandler{renderer: renderer, store: s}
}

func (h *Admin%sHandler) Index(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAll%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s", Admin%sIndexData{CSRFToken: csrf.TokenFromRequest(r), Items: items})
}

func (h *Admin%sHandler) New(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s_form", Admin%sFormData{CSRFToken: csrf.TokenFromRequest(r), IsNew: true})
}

func (h *Admin%sHandler) Edit(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s_form", Admin%sFormData{CSRFToken: csrf.TokenFromRequest(r), Item: item})
}

func (h *Admin%sHandler) Create(w http.ResponseWriter, r *http.Request) {
	item, err := h.parseForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := h.store.Insert%s(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.parseForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item.ID = id
	if err := h.store.Update%s(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) Delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.Delete%s(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) parseForm(r *http.Request) (models.%s, error) {
	if err := r.ParseForm(); err != nil {
		return models.%s{}, err
	}
	%s
}
`,
		boolImport(hasStrconv, "\t\"strconv\"\n"),
		boolImport(hasValidate, "\t\""+frameworkModule+"/pkg/cais/validate\"\n"),
		frameworkModule, frameworkModule, frameworkModule, data.ModulePath, data.ModulePath,
		data.PluralPascal,
		data.PluralPascal, data.Pascal,
		data.Pascal, data.Pascal,
		data.PluralPascal, data.PluralPascal, data.PluralPascal,
		data.PluralPascal,
		data.PluralPascal, data.Plural, data.PluralPascal,
		data.PluralPascal, data.Snake, data.Pascal,
		data.PluralPascal, data.Pascal, data.Snake, data.Pascal,
		data.PluralPascal, data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, data.Pascal, parse,
	)
}

func buildResourcePublicHandler(data scaffoldData) string {
	boolField := firstBoolField(data.Fields)
	intField := firstIntField(data.Fields)

	listDataExtra := ""
	listSum := ""
	if intField != nil {
		listDataExtra = "\n\tTotal int64"
		listSum = fmt.Sprintf(`
	var total int64
	for _, item := range items {
		total += item.%s
	}
`, intField.Pascal)
	}

	toggleMethod := ""
	if boolField != nil {
		toggleMethod = fmt.Sprintf(`

func (h *%sHandler) Toggle(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	item.%s = !item.%s
	if err := h.store.Update%s(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.RenderPartial(w, h.renderer, "%s_toggle", item)
}
`, data.PluralPascal, data.Pascal, boolField.Pascal, boolField.Pascal, data.Pascal, data.Plural)
	}

	return fmt.Sprintf(`package handlers

import (
	"net/http"

	"%s/pkg/cais"
	"%s/pkg/cais/csrf"
	"%s/pkg/cais/httpx"
	"%s/internal/models"
	"%s/internal/store"
)

type %sHandler struct {
	renderer *cais.Renderer
	store    store.Store
}

type %sListData struct {
	CSRFToken string
	Items     []models.%s%s
}

func New%sHandler(renderer *cais.Renderer, s store.Store) *%sHandler {
	return &%sHandler{renderer: renderer, store: s}
}

func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAll%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}%s
	httpx.RenderOrError(w, h.renderer, "base", "%s", %sListData{CSRFToken: csrf.TokenFromRequest(r), Items: items%s})
}
%s`,
		frameworkModule, frameworkModule, frameworkModule, data.ModulePath, data.ModulePath,
		data.PluralPascal,
		data.PluralPascal, data.Pascal, listDataExtra,
		data.PluralPascal, data.PluralPascal, data.PluralPascal,
		data.PluralPascal, data.PluralPascal,
		listSum,
		data.Plural, data.PluralPascal,
		sumArg(intField),
		toggleMethod,
	)
}

func firstBoolField(fields []FieldDef) *FieldDef {
	for i, f := range fields {
		if f.GoType == "bool" {
			return &fields[i]
		}
	}
	return nil
}

func firstIntField(fields []FieldDef) *FieldDef {
	for i, f := range fields {
		if f.GoType == "int64" {
			return &fields[i]
		}
	}
	return nil
}

func sumArg(intField *FieldDef) string {
	if intField == nil {
		return ""
	}
	return ", Total: total"
}

func buildAdminFormHTML(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		switch f.Widget {
		case "textarea":
			fmt.Fprintf(&fields, `    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="%s">%s</label>
      <textarea class="w-full border border-slate-300 rounded-lg px-3 py-2 min-h-[80px] focus:ring-2 focus:ring-indigo-500 outline-none" id="%s" name="%s">{{ .Item.%s }}</textarea>
    </div>
`, f.Name, f.Pascal, f.Name, f.Name, f.Pascal)
		case "checkbox":
			fmt.Fprintf(&fields, `    <label class="flex items-center gap-2 text-sm text-slate-700">
      <input type="checkbox" name="%s" class="rounded border-slate-300 text-indigo-600" {{ if .Item.%s }}checked{{ end }} />
      %s
    </label>
`, f.Name, f.Pascal, f.Pascal)
		default:
			req := ""
			if f.Required {
				req = ` required`
			}
			fmt.Fprintf(&fields, `    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="%s">%s</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-indigo-500 outline-none" type="%s" id="%s" name="%s" value="{{ .Item.%s }}"%s />
    </div>
`, f.Name, f.Pascal, f.HTMLType, f.Name, f.Name, f.Pascal, req)
		}
	}
	return fmt.Sprintf(`{{ define "title" }}{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}{{ end }} {{ define "content" }}
<div class="max-w-md mx-auto">
  <a href="/admin/%s" class="text-sm text-indigo-600 hover:underline mb-4 inline-block">← Back</a>
  <h1 class="text-3xl font-bold text-slate-900 mb-6">{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}</h1>
  <form class="bg-white rounded-2xl border border-slate-200 p-6 shadow-sm space-y-4" method="post"
    action="{{ if .IsNew }}/admin/%s{{ else }}/admin/%s/{{ .Item.ID }}{{ end }}">
    <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}" />
%s
    <button type="submit" class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition">
      {{ if .IsNew }}Create{{ else }}Save{{ end }}
    </button>
  </form>
</div>
{{ end }}
`, data.Title, data.Title, data.Plural, data.Title, data.Title, data.Plural, data.Plural, fields.String())
}

func buildAdminIndexHTML(data scaffoldData) string {
	displayField := data.Fields[0]
	for _, f := range data.Fields {
		if f.Name == "title" || f.Name == "name" {
			displayField = f
			break
		}
	}
	return fmt.Sprintf(`{{ define "title" }}Admin — %s{{ end }} {{ define "content" }}
<div class="max-w-3xl mx-auto">
  <div class="flex items-center justify-between mb-8">
    <h1 class="text-3xl font-bold text-slate-900">%s</h1>
    <a href="/admin/%s/new" class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition shadow-sm">+ New</a>
  </div>
  <div id="admin-%s" class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
    <table class="w-full text-left text-sm">
      <thead class="bg-slate-50 border-b"><tr><th class="px-6 py-3">%s</th><th class="px-6 py-3 text-right">Actions</th></tr></thead>
      <tbody class="divide-y">
        {{ range .Items }}
        <tr class="hover:bg-slate-50">
          <td class="px-6 py-4 font-medium">{{ .%s }}</td>
          <td class="px-6 py-4 text-right space-x-3">
            <a href="/admin/%s/{{ .ID }}/edit" class="text-slate-600 hover:underline">Edit</a>
            <form class="inline" method="post" action="/admin/%s/{{ .ID }}/delete" onsubmit="return confirm('Delete?')">
              <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}" />
              <button type="submit" class="text-red-600 hover:underline">Delete</button>
            </form>
          </td>
        </tr>
        {{ else }}
        <tr><td colspan="2" class="px-6 py-8 text-center text-slate-500">No items yet.</td></tr>
        {{ end }}
      </tbody>
    </table>
  </div>
</div>
{{ end }}
`, data.Title, data.Title, data.Plural, data.Plural, displayField.Pascal, displayField.Pascal, data.Plural, data.Plural)
}

func displayFieldForList(fields []FieldDef) FieldDef {
	for _, f := range fields {
		if f.Name == "title" || f.Name == "name" {
			return f
		}
	}
	return fields[0]
}

func buildPublicListItemHTML(data scaffoldData) string {
	display := displayFieldForList(data.Fields)
	var linkField *FieldDef
	for i, f := range data.Fields {
		if f.HTMLType == "url" {
			linkField = &data.Fields[i]
			break
		}
	}

	var b strings.Builder
	if linkField != nil {
		fmt.Fprintf(&b, `<a href="{{ .%s }}" target="_blank" rel="noopener" class="text-lg font-semibold text-indigo-600 hover:underline">{{ .%s }}</a>`, linkField.Pascal, display.Pascal)
	} else {
		fmt.Fprintf(&b, `<p class="text-lg font-semibold text-slate-800">{{ .%s }}</p>`, display.Pascal)
	}

	var meta []string
	for _, f := range data.Fields {
		if f.Pascal == display.Pascal {
			continue
		}
		switch f.GoType {
		case "bool":
			meta = append(meta, fmt.Sprintf(`<span hx-post="/%s/{{ .ID }}/toggle" hx-swap="outerHTML swap:150ms" hx-target="this" data-cais-optimistic="toggle" class="cursor-pointer inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {{ if .%s }}bg-green-50 text-green-700{{ else }}bg-slate-100 text-slate-600{{ end }}">{{ if .%s }}%s{{ else }}Pending{{ end }}</span>`, data.Plural, f.Pascal, f.Pascal, f.Pascal))
		case "int64":
			meta = append(meta, fmt.Sprintf(`<span class="text-sm text-slate-500">%s: {{ .%s }}</span>`, f.Pascal, f.Pascal))
		}
	}
	if len(meta) > 0 {
		b.WriteString(`<div class="mt-2 flex flex-wrap items-center gap-2">`)
		b.WriteString(strings.Join(meta, "\n"))
		b.WriteString(`</div>`)
	}

	for _, f := range data.Fields {
		if f.Pascal == display.Pascal || f.Widget != "textarea" {
			continue
		}
		fmt.Fprintf(&b, `{{ if .%s }}<p class="mt-2 text-sm text-slate-600 line-clamp-2">{{ .%s }}</p>{{ end }}`, f.Pascal, f.Pascal)
	}

	return b.String()
}

func buildPublicListHTML(data scaffoldData) string {
	pluralTitle := toTitle(data.Plural)
	itemBlock := buildPublicListItemHTML(data)
	intField := firstIntField(data.Fields)
	totalBlock := ""
	if intField != nil {
		totalBlock = fmt.Sprintf(`  <div class="bg-white rounded-2xl border border-slate-200 p-4 shadow-sm mb-6 flex items-center justify-between">
    <span class="text-sm font-medium text-slate-500">Total %s</span>
    <span class="text-2xl font-bold text-indigo-600">{{ .Total }}</span>
  </div>
`, intField.Pascal)
	}
	return fmt.Sprintf(`{{ define "title" }}%s{{ end }} {{ define "content" }}
<div class="max-w-2xl mx-auto">
  <h1 class="text-3xl font-bold text-slate-900 mb-6">%s</h1>
%s  <ul id="%s-list" class="space-y-3">
    {{ range .Items }}
    <li class="bg-white rounded-2xl border border-slate-200 p-5 shadow-sm">%s</li>
    {{ else }}
    <li class="text-center text-slate-500 py-8">No items yet.</li>
    {{ end }}
  </ul>
</div>
{{ end }}
`, pluralTitle, pluralTitle, totalBlock, data.Plural, itemBlock)
}

func buildPublicTogglePartial(data scaffoldData) string {
	boolField := firstBoolField(data.Fields)
	if boolField == nil {
		return ""
	}
	return fmt.Sprintf(`{{- define "%s_toggle" -}}<span hx-post="/%s/{{ .ID }}/toggle" hx-swap="outerHTML swap:150ms" hx-target="this" data-cais-optimistic="toggle" class="cursor-pointer inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {{ if .%s }}bg-green-50 text-green-700{{ else }}bg-slate-100 text-slate-600{{ end }}">{{ if .%s }}%s{{ else }}Pending{{ end }}</span>{{- end -}}
`, data.Plural, data.Plural, boolField.Pascal, boolField.Pascal, boolField.Pascal)
}
