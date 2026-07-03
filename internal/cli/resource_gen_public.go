package cli

import "fmt"

func buildResourcePublicHandler(data scaffoldData) string {
	boolField := firstBoolField(data.Fields)
	intField := firstIntField(data.Fields)

	listDataExtra := ""
	paginationFields := ""
	if data.Paginate {
		paginationFields = `
	Page     int
	Total    int
	PerPage  int
	HasPrev  bool
	HasNext  bool
	PrevPage int
	NextPage int`
	}
	sumField := "Total"
	if data.Paginate && intField != nil {
		sumField = "Sum"
	}
	if intField != nil {
		listDataExtra = fmt.Sprintf("\n\t%s int64", sumField)
	}
	listSum := ""
	if intField != nil {
		listSum = fmt.Sprintf(`
	var %s int64
	for _, item := range items {
		%s += item.%s
	}
`, sumField, sumField, intField.Pascal)
	}

	listMethod := buildPublicListMethod(data, sumField, listSum)

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

	extraImports := ""
	if data.Paginate {
		extraImports = fmt.Sprintf("\t\"strconv\"\n\t\"%s/pkg/cais/pagination\"\n", frameworkModule)
	}

	return fmt.Sprintf(`package handlers

import (
	"net/http"
%s
	"%s/pkg/cais"
	"%s/pkg/cais/httpx"
	"%s/pkg/cais/meta"
	"%s/internal/models"
	"%s/internal/store"
)

type %sHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
}

type %sListData struct {
	meta.Site
	Items []models.%s%s%s
}

func New%sHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config) *%sHandler {
	return &%sHandler{renderer: renderer, store: s, site: site, cfg: cfg}
}

%s%s`,
		extraImports,
		frameworkModule, frameworkModule, frameworkModule, data.ModulePath, data.ModulePath,
		data.PluralPascal,
		data.PluralPascal, data.Pascal, listDataExtra, paginationFields,
		data.PluralPascal, data.PluralPascal, data.PluralPascal,
		listMethod,
		toggleMethod,
	)
}

func buildPublicListMethod(data scaffoldData, sumField, listSum string) string {
	sumArg := ""
	if listSum != "" {
		sumArg = fmt.Sprintf(", %s: %s", sumField, sumField)
	}
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
	}%s
	pg := pagination.New(page, perPage, total)
	listData := %sListData{
		Site:     meta.ForRequest(h.site, r),
		Items:    items,
		Page:     pg.Page,
		Total:    pg.Total,
		PerPage:  pg.PerPage,
		HasPrev:  pg.HasPrev,
		HasNext:  pg.HasNext,
		PrevPage: pg.PrevPage,
		NextPage: pg.NextPage%s,
	}
	httpx.RenderPageOrPartial(w, r, h.renderer, httpx.RenderOptions{
		Layout:  "base",
		Page:    "%s",
		Partial: "%s_list",
		Data:    listData,
	}, h.cfg)
}
`, data.PluralPascal, data.PluralPascal, listSum, data.PluralPascal, sumArg, data.Plural, data.Plural)
	}
	return fmt.Sprintf(`func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAll%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}%s
	httpx.RenderOrError(w, h.renderer, "base", "%s", %sListData{
		Site:  meta.ForRequest(h.site, r),
		Items: items%s,
	}, h.cfg)
}
`, data.PluralPascal, data.PluralPascal, listSum, data.Plural, data.PluralPascal, sumArg)
}
