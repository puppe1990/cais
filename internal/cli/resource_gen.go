package cli

import (
	"fmt"
	"strings"
)

type resourceOpts struct {
	Fields    string
	Public    bool
	Seed      bool
	Paginate  bool
	AdminAuth string
	Force     bool
	dryRun    bool
}

func parseResourceOpts(args []string) (resourceOpts, error) {
	opts := resourceOpts{Seed: true, AdminAuth: "session"}
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
		case "--paginate":
			opts.Paginate = true
		case "--no-seed":
			opts.Seed = false
		case "--force":
			opts.Force = true
		case "--admin-auth":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--admin-auth requires a value")
			}
			i++
			switch args[i] {
			case "session", "bearer":
				opts.AdminAuth = args[i]
			default:
				return opts, fmt.Errorf("--admin-auth must be session or bearer")
			}
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

const tplSelectOptionModel = `package models

// SelectOption is a label/value pair for foreign-key select fields.
type SelectOption struct {
	ID    int64
	Label string
}
`

func buildReferenceStoreMethods(fields []FieldDef, existing string) string {
	var b strings.Builder
	for _, f := range uniqueReferenceFields(fields) {
		if strings.Contains(existing, "List"+f.RefPascal+"Options()") {
			continue
		}
		fmt.Fprintf(&b, `
func (s *SQLiteStore) List%sOptions() ([]models.SelectOption, error) {
	rows, err := s.db.Query(
		"SELECT id, COALESCE(NULLIF(name, ''), NULLIF(title, ''), CAST(id AS TEXT)) FROM %s ORDER BY 2",
	)
	if err != nil {
		return nil, fmt.Errorf("list %s options: %%w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.SelectOption
	for rows.Next() {
		var opt models.SelectOption
		if err := rows.Scan(&opt.ID, &opt.Label); err != nil {
			return nil, fmt.Errorf("scan %s option: %%w", err)
		}
		items = append(items, opt)
	}
	return items, rows.Err()
}
`, f.RefPascal, f.RefTable, f.RefTable, f.RefTable)
	}
	return b.String()
}

func buildResourceMigration(data scaffoldData) string {
	var cols []string
	for _, f := range data.Fields {
		cols = append(cols, fmt.Sprintf("    %s %s", f.Name, f.SQLType))
	}
	create := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n    id INTEGER PRIMARY KEY AUTOINCREMENT,\n%s,\n    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n);",
		data.Plural, strings.Join(cols, ",\n"))
	return fmt.Sprintf("-- up\n%s\n\n-- down\nDROP TABLE IF EXISTS %s;\n", create, data.Plural)
}

func nullableStoreHelpers(fields []FieldDef) string {
	var b strings.Builder
	if fieldNeedsStrPtr(fields) {
		b.WriteString(`
func strPtr(s string) *string { return &s }
`)
	}
	if fieldNeedsInt64Ptr(fields) {
		b.WriteString(`
func int64Ptr(n int64) *int64 { return &n }
`)
	}
	return b.String()
}

func buildResourceStoreMethods(data scaffoldData) string {
	cols, ph := insertColumns(data.Fields)
	args := insertArgs(data.Fields)
	sets := updateSets(data.Fields)
	updArgs := insertArgs(data.Fields) + ", c.ID"
	sel := selectColumns(data.Fields)

	return nullableStoreHelpers(data.Fields) + fmt.Sprintf(`
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

func buildResourcePaginatedStoreMethod(data scaffoldData) string {
	sel := selectColumns(data.Fields)
	return fmt.Sprintf(`
func (s *SQLiteStore) List%s(page, perPage int) ([]models.%s, int, error) {
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM %s").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count %s: %%w", err)
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}
	offset := pagination.Offset(page, perPage)
	rows, err := s.db.Query(
		"SELECT id, %s, created_at FROM %s ORDER BY id DESC LIMIT ? OFFSET ?",
		perPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list %s: %%w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.%s
	for rows.Next() {
		var c models.%s
%s
		if err := rows.Scan(%s); err != nil {
			return nil, 0, fmt.Errorf("scan %s: %%w", err)
		}
%s
		items = append(items, c)
	}
	return items, total, rows.Err()
}
`,
		data.PluralPascal, data.Pascal,
		data.Plural, data.Plural,
		sel, data.Plural, data.Plural,
		data.Pascal, data.Pascal,
		scanLoopDeclare(data.Fields), scanVars(data.Fields), data.Snake, scanLoopAssign(data.Fields),
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
	case "*string":
		return "strPtr(" + seedStringValue(f, name) + ")"
	case "*int64":
		return "int64Ptr(" + seedIntValue(f, name) + ")"
	case "bool":
		if strings.Contains(name, "active") || strings.Contains(name, "enabled") {
			return "true"
		}
		return "false"
	case "int64":
		return seedIntValue(f, name)
	default:
		return seedStringValue(f, name)
	}
}

func seedIntValue(f FieldDef, name string) string {
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
}

func seedStringValue(f FieldDef, name string) string {
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
			after = append(after, fmt.Sprintf(`raw%s := strings.TrimSpace(r.FormValue(%q))
	if raw%s == "" {
		errs.Add(%q, %q)
	} else if %sVal, err := strconv.ParseInt(raw%s, 10, 64); err != nil {
		errs.Add(%q, %q)
	} else {
		item.%s = %sVal
	}`, f.Pascal, f.Name, f.Pascal, f.Name, f.Name+" is required", f.Name, f.Pascal, f.Name, f.Name+" must be a number", f.Pascal, f.Name))
		case "*int64":
			after = append(after, fmt.Sprintf(`if raw%s := strings.TrimSpace(r.FormValue(%q)); raw%s != "" {
		if %sVal, err := strconv.ParseInt(raw%s, 10, 64); err != nil {
			errs.Add(%q, %q)
		} else {
			item.%s = &%sVal
		}
	}`, f.Pascal, f.Name, f.Pascal, f.Name, f.Pascal, f.Name, f.Name+" must be a number", f.Pascal, f.Name))
		case "*string":
			if f.HTMLType == "url" {
				after = append(after, fmt.Sprintf(`if raw%s := strings.TrimSpace(r.FormValue(%q)); raw%s != "" {
		if err := validate.URL(raw%s); err != nil {
			errs.Add(%q, err.Error())
		} else {
			item.%s = &raw%s
		}
	}`, f.Pascal, f.Name, f.Pascal, f.Pascal, f.Name, f.Pascal, f.Pascal))
			} else {
				after = append(after, fmt.Sprintf(`if raw%s := strings.TrimSpace(r.FormValue(%q)); raw%s != "" {
		item.%s = &raw%s
	}`, f.Pascal, f.Name, f.Pascal, f.Pascal, f.Pascal))
			}
		default:
			literal = append(literal, fmt.Sprintf("%s: strings.TrimSpace(r.FormValue(%q))", f.Pascal, f.Name))
			if f.Required {
				if f.HTMLType == "url" {
					validations = append(validations, fmt.Sprintf("if item.%s == \"\" {\n\t\terrs.Add(%q, %q)\n\t} else if err := validate.URL(item.%s); err != nil {\n\t\terrs.Add(%q, err.Error())\n\t}", f.Pascal, f.Name, f.Name+" is required", f.Pascal, f.Name))
				} else {
					validations = append(validations, fmt.Sprintf("if item.%s == \"\" {\n\t\terrs.Add(%q, %q)\n\t}", f.Pascal, f.Name, f.Name+" is required"))
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
	return fmt.Sprintf(`var errs validate.FieldErrors
	if err := r.ParseForm(); err != nil {
		errs.Add("_form", err.Error())
		return models.%s{}, errs
	}
	item := models.%s{%s}%s%s	return item, errs`, data.Pascal, data.Pascal, strings.Join(literal, ", "), validateBlock, afterBlock)
}

func needsStrconv(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "int64" || f.GoType == "*int64" {
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

func boolImport(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}

func buildAdminShowDataStruct(data scaffoldData) string {
	return fmt.Sprintf(`type Admin%sShowData struct {
	CSRFToken string
	Item      models.%s
}`, data.PluralPascal, data.Pascal)
}

func buildAdminShowMethod(data scaffoldData) string {
	return fmt.Sprintf(`func (h *Admin%sHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s_show", Admin%sShowData{
		CSRFToken: csrf.TokenFromRequest(r),
		Item:      item,
	}, h.cfg)
}`, data.PluralPascal, data.Pascal, data.Snake, data.PluralPascal)
}

func buildAdminIndexDataStruct(data scaffoldData) string {
	if data.Paginate {
		return fmt.Sprintf(`type Admin%sIndexData struct {
	CSRFToken string
	Items     []models.%s
	Page      int
	Total     int
	PerPage   int
	HasPrev   bool
	HasNext   bool
	PrevPage  int
	NextPage  int
}`, data.PluralPascal, data.Pascal)
	}
	return fmt.Sprintf(`type Admin%sIndexData struct {
	CSRFToken string
	Items     []models.%s
}`, data.PluralPascal, data.Pascal)
}

func buildAdminIndexMethod(data scaffoldData) string {
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
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s", Admin%sIndexData{
		CSRFToken: csrf.TokenFromRequest(r),
		Items:     items,
		Page:      pg.Page,
		Total:     pg.Total,
		PerPage:   pg.PerPage,
		HasPrev:   pg.HasPrev,
		HasNext:   pg.HasNext,
		PrevPage:  pg.PrevPage,
		NextPage:  pg.NextPage,
	}, h.cfg)
}`, data.PluralPascal, data.PluralPascal, data.Plural, data.PluralPascal)
	}
	return fmt.Sprintf(`func (h *Admin%sHandler) Index(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAll%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s", Admin%sIndexData{CSRFToken: csrf.TokenFromRequest(r), Items: items}, h.cfg)
}`, data.PluralPascal, data.PluralPascal, data.Plural, data.PluralPascal)
}

func adminFormRender(data scaffoldData, itemExpr, isNewExpr, errsExpr string) string {
	if hasReferenceFields(data.Fields) {
		return fmt.Sprintf("h.formData(r, %s, %s, %s)", itemExpr, isNewExpr, errsExpr)
	}
	if errsExpr == "nil" {
		return fmt.Sprintf("Admin%sFormData{CSRFToken: csrf.TokenFromRequest(r), Item: %s, IsNew: %s}", data.PluralPascal, itemExpr, isNewExpr)
	}
	return fmt.Sprintf(`Admin%sFormData{
			CSRFToken: csrf.TokenFromRequest(r),
			Item:      %s,
			IsNew:     %s,
			Errors:    %s,
		}`, data.PluralPascal, itemExpr, isNewExpr, errsExpr)
}

func buildResourceAdminHandler(data scaffoldData) string {
	parse := buildAdminParseForm(data)
	hasStrconv := needsStrconv(data.Fields) || data.Paginate || hasReferenceFields(data.Fields)
	hasRefs := hasReferenceFields(data.Fields)
	indexDataStruct := buildAdminIndexDataStruct(data)
	showDataStruct := buildAdminShowDataStruct(data)
	formDataStruct := buildAdminFormDataStruct(data)
	indexMethod := buildAdminIndexMethod(data)
	showMethod := buildAdminShowMethod(data)
	formDataMethod := ""
	if hasReferenceFields(data.Fields) {
		formDataMethod = buildAdminFormDataMethod(data) + "\n\n"
	}
	paginationImport := ""
	if data.Paginate {
		paginationImport = "\t\"" + frameworkModule + "/pkg/cais/pagination\"\n"
	}
	formsImport := ""
	if hasRefs {
		formsImport = "\t\"" + frameworkModule + "/pkg/cais/forms\"\n"
	}
	newRender := adminFormRender(data, "models."+data.Pascal+"{}", "true", "nil")
	editRender := adminFormRender(data, "item", "false", "nil")
	createErrRender := adminFormRender(data, "item", "true", "errs")
	updateErrRender := adminFormRender(data, "item", "false", "errs")
	return fmt.Sprintf(`package handlers

import (
	"net/http"
%s	"strings"
	"%s/pkg/cais/validate"
%s%s
	"%s/pkg/cais"
	"%s/pkg/cais/csrf"
	"%s/pkg/cais/httpx"
	"%s/internal/models"
	"%s/internal/store"
)

type Admin%sHandler struct {
	renderer *cais.Renderer
	store    store.Store
	cfg      cais.Config
}

%s

%s

%s

func NewAdmin%sHandler(renderer *cais.Renderer, s store.Store, cfg cais.Config) *Admin%sHandler {
	return &Admin%sHandler{renderer: renderer, store: s, cfg: cfg}
}

%s

%s

%sfunc (h *Admin%sHandler) New(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s_form", %s, h.cfg)
}

func (h *Admin%sHandler) Edit(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s_form", %s, h.cfg)
}

func (h *Admin%sHandler) Create(w http.ResponseWriter, r *http.Request) {
	item, errs := h.parseForm(r)
	if errs.Any() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		httpx.RenderOrError(w, h.renderer, "base", "admin_%s_form", %s, h.cfg)
		return
	}
	if _, err := h.store.Insert%s(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	item, errs := h.parseForm(r)
	item.ID = id
	if errs.Any() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		httpx.RenderOrError(w, h.renderer, "base", "admin_%s_form", %s, h.cfg)
		return
	}
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

func (h *Admin%sHandler) parseForm(r *http.Request) (models.%s, validate.FieldErrors) {
	%s
}
`,
		boolImport(hasStrconv, "\t\"strconv\"\n"),
		frameworkModule,
		paginationImport,
		formsImport,
		frameworkModule, frameworkModule, frameworkModule, data.ModulePath, data.ModulePath,
		data.PluralPascal,
		indexDataStruct,
		showDataStruct,
		formDataStruct,
		data.PluralPascal, data.PluralPascal, data.PluralPascal,
		indexMethod,
		showMethod,
		formDataMethod,
		data.PluralPascal, data.Snake, newRender,
		data.PluralPascal, data.Pascal, data.Snake, editRender,
		data.PluralPascal, data.Snake, createErrRender,
		data.Pascal, data.Plural,
		data.PluralPascal, data.Snake, updateErrRender,
		data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, parse,
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
	cfg      cais.Config
}

type %sListData struct {
	CSRFToken string
	Items     []models.%s%s
}

func New%sHandler(renderer *cais.Renderer, s store.Store, cfg cais.Config) *%sHandler {
	return &%sHandler{renderer: renderer, store: s, cfg: cfg}
}

func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAll%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}%s
	httpx.RenderOrError(w, h.renderer, "base", "%s", %sListData{CSRFToken: csrf.TokenFromRequest(r), Items: items%s}, h.cfg)
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
		if f.Widget == "select" {
			continue
		}
		if f.GoType == "int64" || f.GoType == "*int64" {
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

func buildAdminFormDataStruct(data scaffoldData) string {
	var extra []string
	for _, f := range data.Fields {
		if f.RefTable != "" {
			extra = append(extra, fmt.Sprintf("\t%sOptions []forms.SelectOption", f.RefPascal))
		}
	}
	extraBlock := ""
	if len(extra) > 0 {
		extraBlock = "\n" + strings.Join(extra, "\n")
	}
	return fmt.Sprintf(`type Admin%sFormData struct {
	CSRFToken string
	Item      models.%s
	IsNew     bool
	Errors    validate.FieldErrors%s
}`, data.PluralPascal, data.Pascal, extraBlock)
}

func buildAdminFormDataLoader(data scaffoldData) string {
	var lines []string
	for _, f := range data.Fields {
		if f.RefTable == "" {
			continue
		}
		rawVar := "raw" + f.RefPascal + "Opts"
		lines = append(lines, fmt.Sprintf(`	if %s, err := h.store.List%sOptions(); err == nil {
		for _, opt := range %s {
			data.%sOptions = append(data.%sOptions, forms.SelectOption{
				Value: strconv.FormatInt(opt.ID, 10),
				Label: opt.Label,
			})
		}
	}`, rawVar, f.RefPascal, rawVar, f.RefPascal, f.RefPascal))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func buildAdminFormDataMethod(data scaffoldData) string {
	loader := buildAdminFormDataLoader(data)
	return fmt.Sprintf(`func (h *Admin%sHandler) formData(r *http.Request, item models.%s, isNew bool, errs validate.FieldErrors) Admin%sFormData {
	data := Admin%sFormData{
		CSRFToken: csrf.TokenFromRequest(r),
		Item:      item,
		IsNew:     isNew,
		Errors:    errs,
	}
%s	return data
}`, data.PluralPascal, data.Pascal, data.PluralPascal, data.PluralPascal, loader)
}

func buildAdminFormHTML(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		switch f.Widget {
		case "select":
			if f.GoType == "*int64" {
				fmt.Fprintf(&fields, `    {{ fieldSelect (makeSelectFieldPtr "%s" "%s" .Item.%s .%sOptions %t .Errors) }}
`, f.Name, f.RefPascal, f.Pascal, f.RefPascal, f.Required)
			} else {
				fmt.Fprintf(&fields, `    {{ fieldSelect (makeSelectField "%s" "%s" .Item.%s .%sOptions %t .Errors) }}
`, f.Name, f.RefPascal, f.Pascal, f.RefPascal, f.Required)
			}
		case "textarea":
			fmt.Fprintf(&fields, `    {{ fieldInput (makeField "%s" "%s" .Item.%s "textarea" %t .Errors) }}
`, f.Name, f.Pascal, f.Pascal, f.Required)
		case "checkbox":
			fmt.Fprintf(&fields, `    <label class="flex items-center gap-2 text-sm text-slate-700">
      <input type="checkbox" name="%s" class="rounded border-slate-300 text-indigo-600" {{ if .Item.%s }}checked{{ end }} />
      %s
    </label>
`, f.Name, f.Pascal, f.Pascal)
		default:
			fmt.Fprintf(&fields, `    {{ fieldInput (makeField "%s" "%s" .Item.%s "%s" %t .Errors) }}
`, f.Name, f.Pascal, f.Pascal, f.HTMLType, f.Required)
		}
	}
	return fmt.Sprintf(`{{ define "title" }}{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}{{ end }} {{ define "content" }}
<div class="max-w-md mx-auto">
  <a href="/admin/%s" class="text-sm text-indigo-600 hover:underline mb-4 inline-block">← Back</a>
  <h1 class="text-3xl font-bold text-slate-900 mb-6">{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}</h1>
  <form class="bg-white rounded-2xl border border-slate-200 p-6 shadow-sm space-y-4" method="post"
    action="{{ if .IsNew }}/admin/%s{{ else }}/admin/%s/{{ .Item.ID }}{{ end }}">
    {{ csrfField .CSRFToken }}
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
	paginationBlock := ""
	if data.Paginate {
		paginationBlock = fmt.Sprintf(`    <div class="flex items-center justify-between px-6 py-4 border-t bg-slate-50">
      {{ if .HasPrev }}
      <a href="/admin/%s?page={{ .PrevPage }}" class="text-indigo-600 hover:underline">← Previous</a>
      {{ else }}
      <span></span>
      {{ end }}
      <span class="text-sm text-slate-500">Page {{ .Page }}</span>
      {{ if .HasNext }}
      <a href="/admin/%s?page={{ .NextPage }}" class="text-indigo-600 hover:underline">Next →</a>
      {{ else }}
      <span></span>
      {{ end }}
    </div>
`, data.Plural, data.Plural)
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
            <a href="/admin/%s/{{ .ID }}" class="text-indigo-600 hover:underline">View</a>
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
%s  </div>
</div>
{{ end }}
`, data.Title, data.Title, data.Plural, data.Plural, displayField.Pascal, displayField.Pascal, data.Plural, data.Plural, data.Plural, paginationBlock)
}

func buildAdminShowHTML(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		if f.GoType == "bool" {
			fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm font-medium text-slate-500">%s</dt>
      <dd class="mt-1 text-slate-900">{{ if .Item.%s }}Yes{{ else }}No{{ end }}</dd>
    </div>
`, f.Pascal, f.Pascal)
			continue
		}
		fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm font-medium text-slate-500">%s</dt>
      <dd class="mt-1 text-slate-900">{{ .Item.%s }}</dd>
    </div>
`, f.Pascal, f.Pascal)
	}
	return fmt.Sprintf(`{{ define "title" }}%s{{ end }} {{ define "content" }}
<div class="max-w-md mx-auto">
  <a href="/admin/%s" class="text-sm text-indigo-600 hover:underline mb-4 inline-block">← Back</a>
  <h1 class="text-3xl font-bold text-slate-900 mb-6">%s</h1>
  <dl class="bg-white rounded-2xl border border-slate-200 p-6 shadow-sm space-y-4">
%s  </dl>
  <div class="mt-6 flex gap-3">
    <a href="/admin/%s/{{ .Item.ID }}/edit" class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition shadow-sm">Edit</a>
    <form method="post" action="/admin/%s/{{ .Item.ID }}/delete" onsubmit="return confirm('Delete?')">
      <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}" />
      <button type="submit" class="text-red-600 hover:underline py-2 px-4">Delete</button>
    </form>
  </div>
</div>
{{ end }}
`, data.Title, data.Plural, data.Title, fields.String(), data.Plural, data.Plural)
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
		case "int64", "*int64":
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
