package cli

import "fmt"

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
