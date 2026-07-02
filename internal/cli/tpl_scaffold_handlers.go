package cli

const tplHomeHandler = `package handlers

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

type PageData struct {
	meta.Site
	Nome string
}

type HomeHandler struct {
	renderer *cais.Renderer
	site     meta.Site
	catalog  *i18n.Catalog
	cfg      cais.Config
}

func NewHomeHandler(renderer *cais.Renderer, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *HomeHandler {
	return &HomeHandler{renderer: renderer, site: site, catalog: catalog, cfg: cfg}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "welcome", "home", PageData{
		Site: meta.ForRequest(h.site, r),
	}, h.cfg)
}
`

const tplHomeTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestHomeHandler_Returns200(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_ContainsWelcome(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "on Cais!") {
		t.Errorf("body missing welcome message, got: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}
`

const tplHomeTestMinimal = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestHomeHandler_Returns200(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_ContainsWelcome(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "on Cais!") {
		t.Errorf("body missing welcome message, got: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}
`

const tplContactHandler = `package handlers

import (
	"net/http"
	"strings"

	"{{.ModulePath}}/internal/models"
	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/validate"
)

type ContactHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	catalog  *i18n.Catalog
	cfg      cais.Config
}

type contactErrorData struct {
	Message string
}

func NewContactHandler(renderer *cais.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *ContactHandler {
	return &ContactHandler{renderer: renderer, store: s, site: site, catalog: catalog, cfg: cfg}
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "base", "contact", meta.ForRequest(h.site, r), h.cfg)
}

func (h *ContactHandler) Post(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))

	var errs validate.FieldErrors
	if name == "" {
		errs.Add("name", h.catalog.T("contact.name_required"))
	}
	if err := validate.Email(email); err != nil {
		msg := h.catalog.T("contact.email_required")
		if email != "" {
			msg = h.catalog.T("contact.email_invalid")
		}
		errs.Add("email", msg)
	}
	if errs.Any() {
		h.renderContactResponse(w, r, http.StatusUnprocessableEntity, "contact_errors", contactErrorData{Message: errs.First()})
		return
	}

	_, err := h.store.InsertContact(models.Contact{Name: name, Email: email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderContactResponse(w, r, http.StatusOK, "contact_success", nil)
}

func (h *ContactHandler) renderContactResponse(w http.ResponseWriter, r *http.Request, status int, partial string, data any) {
	httpx.RenderPageOrPartial(w, r, h.renderer, httpx.RenderOptions{
		Layout:  "base",
		Page:    "contact",
		Partial: partial,
		Data:    data,
		Status:  status,
	}, h.cfg)
}
`

const tplContactTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestContactHandler_Get_ReturnsForm(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "contact-form") {
		t.Errorf("body missing form, got: %s", rr.Body.String())
	}
}

func TestContactHandler_Post_MissingName_Returns422(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=&email=alice@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(rr.Body.String(), "Name is required") {
		t.Errorf("body missing name validation: %s", rr.Body.String())
	}
}

func TestContactHandler_Post_InvalidEmail_Returns422(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestContactHandler_Post_InvalidEmail_ReturnsPartial(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("expected partial HTML, got full page")
	}
	if !strings.Contains(body, "Email is required") {
		t.Errorf("body missing error message, got: %s", body)
	}
}

func TestContactHandler_Post_Valid_SavesAndReturnsSuccess(t *testing.T) {
	s := setupTestStore(t)
	h := NewContactHandler(setupTestRenderer(t), s, testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=alice@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "successfully") {
		t.Errorf("body missing success message, got: %s", rr.Body.String())
	}
}
`

const tplDashboardHandler = `package handlers

import (
	"net/http"

	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

type DashboardData struct {
	meta.Site
	TotalContacts int64
	Env           string
}

type DashboardHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
}

func NewDashboardHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config) *DashboardHandler {
	return &DashboardHandler{renderer: renderer, store: s, site: site, cfg: cfg}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.CountContacts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpx.RenderOrError(w, h.renderer, "base", "dashboard", DashboardData{
		Site:          meta.ForRequest(h.site, r),
		TotalContacts: count,
		Env:           h.cfg.Env,
	}, h.cfg)
}
`

const tplDashboardTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestDashboardHandler_Returns200(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestDashboardHandler_ContainsDashboard(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "Dashboard") {
		t.Errorf("body missing Dashboard, got: %s", rr.Body.String())
	}
}
`

const tplHelpersTest = `package handlers

import (
	"testing"

	appi18n "{{.ModulePath}}/internal/i18n"
	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	caisi18n "github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/testutil"
)

func setupTestRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	return testutil.NewRenderer(t)
}

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func testSite() meta.Site {
	return meta.Site{AppName: "{{.AppName}}", AppURL: "https://example.com"}
}

func testCatalog() *caisi18n.Catalog {
	return appi18n.DefaultCatalog()
}
`

const tplGenericHandler = `package handlers

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

type {{.Pascal}}PageData struct {
	meta.Site
}

type {{.Pascal}}Handler struct {
	renderer *cais.Renderer
	site     meta.Site
	catalog  *i18n.Catalog
	cfg      cais.Config
}

func New{{.Pascal}}Handler(renderer *cais.Renderer, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *{{.Pascal}}Handler {
	return &{{.Pascal}}Handler{renderer: renderer, site: site, catalog: catalog, cfg: cfg}
}

func (h *{{.Pascal}}Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "base", "{{.Snake}}", {{.Pascal}}PageData{
		Site: meta.ForRequest(h.site, r),
	}, h.cfg)
}
`

const tplGenericHandlerTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func Test{{.Pascal}}Handler_Returns200(t *testing.T) {
	h := New{{.Pascal}}Handler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/{{.Snake}}", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func Test{{.Pascal}}Handler_ContainsTitle(t *testing.T) {
	h := New{{.Pascal}}Handler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/{{.Snake}}", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "{{.Title}}") {
		t.Errorf("body missing title, got: %s", rr.Body.String())
	}
}
`
