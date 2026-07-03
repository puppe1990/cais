// Resource admin handler Go code generation for cais g resource.
package cli

import (
	"fmt"
	"strings"
)

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

func buildAdminShowDataStruct(data scaffoldData) string {
	return fmt.Sprintf(`type Admin%sShowData struct {
	meta.Site
	Item models.%s
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
		Site: meta.ForRequest(h.site, r),
		Item: item,
	}, h.cfg)
}`, data.PluralPascal, data.Pascal, data.Snake, data.PluralPascal)
}

func buildAdminIndexDataStruct(data scaffoldData) string {
	if data.Paginate {
		return fmt.Sprintf(`type Admin%sIndexData struct {
	meta.Site
	Items    []models.%s
	Page     int
	Total    int
	PerPage  int
	HasPrev  bool
	HasNext  bool
	PrevPage int
	NextPage int
}`, data.PluralPascal, data.Pascal)
	}
	return fmt.Sprintf(`type Admin%sIndexData struct {
	meta.Site
	Items []models.%s
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
		Site:     meta.ForRequest(h.site, r),
		Items:    items,
		Page:     pg.Page,
		Total:    pg.Total,
		PerPage:  pg.PerPage,
		HasPrev:  pg.HasPrev,
		HasNext:  pg.HasNext,
		PrevPage: pg.PrevPage,
		NextPage: pg.NextPage,
	}, h.cfg)
}`, data.PluralPascal, data.PluralPascal, data.Plural, data.PluralPascal)
	}
	return fmt.Sprintf(`func (h *Admin%sHandler) Index(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAll%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "admin_%s", Admin%sIndexData{
		Site:  meta.ForRequest(h.site, r),
		Items: items,
	}, h.cfg)
}`, data.PluralPascal, data.PluralPascal, data.Plural, data.PluralPascal)
}

func adminFormRender(data scaffoldData, itemExpr, isNewExpr, errsExpr string) string {
	if hasReferenceFields(data.Fields) {
		return fmt.Sprintf("h.formData(r, %s, %s, %s)", itemExpr, isNewExpr, errsExpr)
	}
	if errsExpr == "nil" {
		return fmt.Sprintf("Admin%sFormData{Site: meta.ForRequest(h.site, r), Item: %s, IsNew: %s}", data.PluralPascal, itemExpr, isNewExpr)
	}
	return fmt.Sprintf(`Admin%sFormData{
			Site:   meta.ForRequest(h.site, r),
			Item:   %s,
			IsNew:  %s,
			Errors: %s,
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
	"%s/pkg/cais/httpx"
	"%s/pkg/cais/meta"
	"%s/internal/models"
	"%s/internal/store"
)

type Admin%sHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
}

%s

%s

%s

func NewAdmin%sHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config) *Admin%sHandler {
	return &Admin%sHandler{renderer: renderer, store: s, site: site, cfg: cfg}
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
	meta.Site
	Item   models.%s
	IsNew  bool
	Errors validate.FieldErrors%s
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
		Site:   meta.ForRequest(h.site, r),
		Item:   item,
		IsNew:  isNew,
		Errors: errs,
	}
%s	return data
}`, data.PluralPascal, data.Pascal, data.PluralPascal, data.PluralPascal, loader)
}
