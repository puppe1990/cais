package cli

// Generic handler generator templates (Inertia + Svelte).
const tplGenericHandler = `package handlers

import (
	"net/http"

	inertia "github.com/romsar/gonertia/v3"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

type {{.Pascal}}Handler struct {
	site    meta.Site
	catalog *i18n.Catalog
	inertia *inertia.Inertia
}

func New{{.Pascal}}Handler(site meta.Site, catalog *i18n.Catalog, i *inertia.Inertia) *{{.Pascal}}Handler {
	return &{{.Pascal}}Handler{site: site, catalog: catalog, inertia: i}
}

func (h *{{.Pascal}}Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = h.inertia.Render(w, r, "{{.Pascal}}", inertia.Props{
		"site": meta.ForRequest(h.site, r),
	})
}
`

const tplGenericHandlerTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appi18n "{{.ModulePath}}/internal/i18n"
)

func Test{{.Pascal}}Handler_InertiaComponent(t *testing.T) {
	h := New{{.Pascal}}Handler(testSite(), appi18n.DefaultCatalog(), setupTestInertia(t))

	req := inertiaRequest(http.MethodGet, "/{{.Snake}}", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	assertInertiaComponent(t, rr, "{{.Pascal}}")
}
`

const tplGenericPage = `<script>
  export let site = {}
</script>

<div class="p-8">
  <h1 class="text-2xl font-bold">{{.Title}}</h1>
  <p>{{.Title}} page — customize this Svelte component.</p>
</div>
`
