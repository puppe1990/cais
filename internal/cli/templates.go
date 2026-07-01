package cli

const tplGoMod = `module {{.ModulePath}}

go 1.26

require (
	github.com/puppe1990/cais v0.1.0
	modernc.org/sqlite v1.53.0
)
`

const tplEmptyCSS = `/* Run: cais css */\n`

const tplAir = `root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  entrypoint = ["./tmp/main"]
  delay = 1000
  exclude_dir = ["tmp", "data", "bin", "node_modules"]
  include_ext = ["go", "html"]
  stop_on_error = true

[log]
  time = false
  main_only = true

[misc]
  clean_on_exit = true
  startup_banner = ""
`

const tplMain = `package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/boot"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"{{.ModulePath}}/internal/app"
	appi18n "{{.ModulePath}}/internal/i18n"
	"{{.ModulePath}}/internal/store"
	"{{.ModulePath}}/web"
)

func main() {
	cfg := cais.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	preferredPort := cfg.Port
	port, shifted, err := cais.ResolvePort(cfg.Port, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Port = port

	a, err := bootstrapWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	shiftedFrom := ""
	if shifted {
		shiftedFrom = preferredPort
	}
	boot.Print(os.Stdout, boot.Options{
		AppName:         "{{.AppName}}",
		Config:          cfg,
		Version:         boot.CaisVersion(),
		PortShiftedFrom: shiftedFrom,
	})
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func bootstrap() (*app.App, error) {
	return bootstrapWithConfig(cais.Load())
}

func bootstrapWithConfig(cfg cais.Config) (*app.App, error) {
	tmplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	catalog := appi18n.NewCatalog(cfg.Locale)
	renderer, err := cais.NewRenderer(tmplFS, catalog)
	if err != nil {
		return nil, fmt.Errorf("renderer: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	staticDir, err := findWebDir("static")
	if err != nil {
		_ = s.Close()
		return nil, err
	}

	return app.New(cfg, app.Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: staticDir,
		Site:      meta.SiteFrom("{{.AppName}}", cfg.AppURL),
		Catalog:   catalog,
	})
}

func findWebDir(subpath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(wd, "web", subpath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("web/%s not found", subpath)
		}
		wd = parent
	}
}
`

const tplConsole = `package main

import (
	"log"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/console"
	"{{.ModulePath}}/internal/store"
)

func openStore(cfg cais.Config) (*store.SQLiteStore, error) {
	return store.NewSQLiteStore(cfg.DBPath, cfg.Env)
}

func bindings(s *store.SQLiteStore) map[string]any {
	return map[string]any{
		"store": s,
		"db":    s.DB(),
	}
}

func main() {
	cfg := cais.Load()
	s, err := openStore(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	active := s
	if err := console.Run(console.Options{
		AppName:  "{{.AppName}}",
		Config:   cfg,
		Bindings: bindings(active),
		Reload: func() (map[string]any, error) {
			_ = active.Close()
			next, err := openStore(cfg)
			if err != nil {
				return nil, err
			}
			active = next
			return bindings(active), nil
		},
	}); err != nil {
		log.Fatal(err)
	}
}
`

const tplApp = `package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"{{.ModulePath}}/internal/store"
)

type Deps struct {
	Renderer  *cais.Renderer
	Store     store.Store
	StaticDir string
	Site      meta.Site
	Catalog   *i18n.Catalog
}

type App struct {
	config cais.Config
	store  store.Store
	router *cais.Router
	server *http.Server
}

func New(cfg cais.Config, deps Deps) (*App, error) {
	if deps.Renderer == nil {
		return nil, fmt.Errorf("renderer is required")
	}
	if deps.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	site := deps.Site
	if site.AppName == "" {
		site = meta.SiteFrom("{{.AppName}}", cfg.AppURL)
	}
	deps.Site = site

	r := cais.NewRouter()
	r.Use(middleware.CSRF(cfg))
	r.Use(middleware.LoadSession(deps.Store.Sessions()))
	r.Use(middleware.Flash)
	buf := devlog.Prepare(cfg.Env)
	if buf != nil {
		r.Use(middleware.LoggerTo(devlog.MirrorDefault(log.Writer())))
	} else {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recover)
	r.Use(middleware.SecurityHeaders(cfg))
	r.Static("/static", deps.StaticDir)

	registerRoutes(r, deps, cfg)
	devlog.Register(r, cfg.Env, buf)
	r.Get("/health", healthHandler(deps.Store))

	return &App{
		config: cfg,
		store:  deps.Store,
		router: r,
		server: &http.Server{
			Addr:              cfg.Port,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}, nil
}

func healthHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if err := s.Ping(); err != nil {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
}

func (a *App) Handler() http.Handler {
	return a.router
}

func (a *App) Run() error {
	return a.RunContext(context.Background())
}

func (a *App) RunContext(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			_ = a.store.Close()
			return err
		}
		<-errCh
		_ = a.store.Close()
		return nil
	case err := <-errCh:
		_ = a.store.Close()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
`

const tplRoutes = `package app

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"{{.ModulePath}}/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store, deps.Site, deps.Catalog)
	dashboard := handlers.NewDashboardHandler(deps.Renderer, deps.Store, deps.Site, cfg)
	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, deps.Site, deps.Store.Sessions(), cfg, deps.Catalog)

	loginLimit := middleware.NewRateLimiter(10)
	contactLimit := middleware.NewRateLimiter(20)

	r.Get("/", home.ServeHTTP)
	r.Get("/contact", contact.Get)
	r.Post("/contact", contactLimit.Middleware(http.HandlerFunc(contact.Post)).ServeHTTP)
	r.Get("/login", auth.Login)
	r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
	r.Post("/logout", auth.LogoutPost)
	r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
}
`

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
}

func NewHomeHandler(renderer *cais.Renderer, site meta.Site, catalog *i18n.Catalog) *HomeHandler {
	return &HomeHandler{renderer: renderer, site: site, catalog: catalog}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "welcome", "home", PageData{
		Site: meta.ForRequest(h.site, r),
	})
}
`

const tplHomeTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

)

func TestHomeHandler_Returns200(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_ContainsWelcome(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "on Cais!") {
		t.Errorf("body missing welcome message, got: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog())

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

)

func TestHomeHandler_Returns200(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_ContainsWelcome(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "on Cais!") {
		t.Errorf("body missing welcome message, got: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog())

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
}

type contactErrorData struct {
	Message string
}

func NewContactHandler(renderer *cais.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog) *ContactHandler {
	return &ContactHandler{renderer: renderer, store: s, site: site, catalog: catalog}
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "base", "contact", meta.ForRequest(h.site, r))
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
	})
}
`

const tplContactTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

)

func TestContactHandler_Get_ReturnsForm(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog())

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
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog())

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
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog())

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
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog())

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
	h := NewContactHandler(setupTestRenderer(t), s, testSite(), testCatalog())

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
	})
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

const tplContactModel = `package models

import "time"

type Contact struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
}
`

const tplStore = `package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/session"
	caissqlite "github.com/puppe1990/cais/pkg/cais/sqlite"
	"github.com/puppe1990/cais/pkg/cais/sqllog"
	"{{.ModulePath}}/internal/models"
)

type Store interface {
	InsertContact(contact models.Contact) (int64, error)
	FindContact(id int64) (models.Contact, error)
	CountContacts() (int64, error)
	FindUserByEmail(email string) (models.User, error)
	Sessions() session.Store
	Ping() error
	Close() error
}

type SQLiteStore struct {
	db *sqllog.DB
}

func NewSQLiteStore(dsn string, env string) (*SQLiteStore, error) {
	if dsn != ":memory:" {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := caissqlite.Configure(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	cfg := sqllog.Config{Enabled: sqllog.EnabledForEnv(env)}
	if cfg.Enabled {
		cfg.Writer = devlog.MirrorDefault(os.Stdout)
	}
	wrapped := sqllog.Wrap(db, cfg)
	if err := seedAuthData(wrapped.Raw(), env); err != nil {
		_ = wrapped.Close()
		return nil, err
	}
	return &SQLiteStore{db: wrapped}, nil
}

func seedAuthData(db *sql.DB, env string) error {
	if env != "development" {
		return nil
	}
	if err := session.EnsureSQLiteSchema(db); err != nil {
		return err
	}
	hash, err := session.HashPassword("password")
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT OR IGNORE INTO users (email, password_hash) VALUES (?, ?)", "demo@example.com", hash)
	return err
}

func (s *SQLiteStore) InsertContact(contact models.Contact) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO contacts (name, email) VALUES (?, ?)",
		contact.Name, contact.Email,
	)
	if err != nil {
		return 0, fmt.Errorf("insert contact: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) FindContact(id int64) (models.Contact, error) {
	var c models.Contact
	err := s.db.QueryRow(
		"SELECT id, name, email, created_at FROM contacts WHERE id = ?",
		id,
	).Scan(&c.ID, &c.Name, &c.Email, &c.CreatedAt)
	if err != nil {
		return models.Contact{}, fmt.Errorf("find contact: %w", err)
	}
	return c, nil
}

func (s *SQLiteStore) CountContacts() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count contacts: %w", err)
	}
	return count, nil
}

func (s *SQLiteStore) FindUserByEmail(email string) (models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (s *SQLiteStore) Sessions() session.Store {
	return session.NewSQLiteStore(s.db.Raw())
}

func (s *SQLiteStore) Ping() error {
	return s.db.Raw().Ping()
}

func (s *SQLiteStore) DB() *sql.DB {
	return s.db.Raw()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
`

const tplStoreTest = `package store

import (
	"testing"

	"{{.ModulePath}}/internal/models"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStore_Migrations(t *testing.T) {
	s := newTestStore(t)

	var name string
	err := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='contacts'").Scan(&name)
	if err != nil {
		t.Fatalf("contacts table not found: %v", err)
	}
}

func TestStore_InsertContact(t *testing.T) {
	s := newTestStore(t)

	contact := models.Contact{Name: "Alice", Email: "alice@example.com"}
	id, err := s.InsertContact(contact)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Error("id = 0, want non-zero")
	}
}

func TestStore_FindContact(t *testing.T) {
	s := newTestStore(t)

	contact := models.Contact{Name: "Bob", Email: "bob@example.com"}
	id, err := s.InsertContact(contact)
	if err != nil {
		t.Fatal(err)
	}

	found, err := s.FindContact(id)
	if err != nil {
		t.Fatal(err)
	}
	if found.Name != "Bob" {
		t.Errorf("Name = %q, want %q", found.Name, "Bob")
	}
	if found.Email != "bob@example.com" {
		t.Errorf("Email = %q, want %q", found.Email, "bob@example.com")
	}
}

func TestStore_CountContacts(t *testing.T) {
	s := newTestStore(t)

	count, err := s.CountContacts()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	_, err = s.InsertContact(models.Contact{Name: "Alice", Email: "alice@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.InsertContact(models.Contact{Name: "Bob", Email: "bob@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	count, err = s.CountContacts()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}
`

const tplMigrations = `package store

import (
	"database/sql"
	"embed"

	"github.com/puppe1990/cais/pkg/cais/migrate"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func applyMigrations(db *sql.DB) error {
	return migrate.Apply(db, migrationFS, "migrations")
}
`

const tplMigration001 = `CREATE TABLE IF NOT EXISTS contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const tplWebEmbed = `package web

import "embed"

//go:embed templates/*
var Templates embed.FS
`

const tplLayout = `{{"{{"}} define "title" {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{.AppName}} — powered by Cais{{"{{"}} end {{"}}"}}
{{"{{"}} define "base" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} htmlLang {{"}}"}}">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
    {{"{{"}} if .CSRFToken {{"}}"}}<meta name="csrf-token" content="{{"{{"}} .CSRFToken {{"}}"}}" />{{"{{"}} end {{"}}"}}
    <title>{{"{{"}} template "title" . {{"}}"}}</title>
    <meta name="description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="{{.AppName}}" />
    <meta property="og:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta property="og:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <meta property="og:locale" content="{{"{{"}} ogLocale {{"}}"}}" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta name="twitter:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta name="twitter:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#4f46e5" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="{{.AppName}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />
    <script src="/static/js/htmx.min.js" defer></script>
    <script src="/static/js/cais.js" defer></script>
  </head>
  <body class="bg-slate-50 text-slate-900 min-h-screen flex flex-col">
    <header class="bg-white border-b border-slate-200 p-4 shadow-sm">
      <div class="max-w-5xl mx-auto flex justify-between items-center">
        <a href="/" class="font-bold text-xl text-indigo-600 hover:text-indigo-700 transition">{{.AppName}}</a>
        <nav class="flex items-center gap-6 text-sm font-medium">
          <a href="/" class="text-slate-600 hover:text-indigo-600 transition">Home</a>
          <a href="/contact" class="text-slate-600 hover:text-indigo-600 transition">Contact</a>
          <a href="/dashboard" class="text-slate-600 hover:text-indigo-600 transition">Dashboard</a>
        </nav>
      </div>
    </header>
    <main class="flex-grow max-w-5xl w-full mx-auto p-6">{{"{{"}} template "content" . {{"}}"}}</main>
    <footer class="border-t border-slate-200 p-4 text-center text-sm text-slate-500">
      {{.AppName}} — powered by Cais
    </footer>
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
  </body>
</html>
{{"{{"}} end {{"}}"}}
`

const tplLayoutWelcome = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} if .AppName {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} else {{"}}"}}Cais{{"{{"}} end {{"}}"}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{"{{"}} if .AppName {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} else {{"}}"}}Cais{{"{{"}} end {{"}}"}} — powered by Cais{{"{{"}} end {{"}}"}}
{{"{{"}} define "welcome" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} htmlLang {{"}}"}}">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
    {{"{{"}} if .CSRFToken {{"}}"}}<meta name="csrf-token" content="{{"{{"}} .CSRFToken {{"}}"}}" />{{"{{"}} end {{"}}"}}
    <title>{{"{{"}} template "title" . {{"}}"}}</title>
    <meta name="description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="{{"{{"}} if .AppName {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} else {{"}}"}}Cais{{"{{"}} end {{"}}"}}" />
    <meta property="og:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta property="og:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <meta property="og:locale" content="{{"{{"}} ogLocale {{"}}"}}" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta name="twitter:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta name="twitter:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#D4A574" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="default" />
    <meta name="apple-mobile-web-app-title" content="{{"{{"}} if .AppName {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} else {{"}}"}}Cais{{"{{"}} end {{"}}"}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />
    <script src="/static/js/htmx.min.js" defer></script>
    <script src="/static/js/cais.js" defer></script>
  </head>
  <body class="min-h-screen bg-gradient-to-b from-[#FAF3E8] via-[#EDCFA8] to-[#C9895E] text-stone-800 antialiased">
    <main>{{"{{"}} template "content" . {{"}}"}}</main>
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
  </body>
</html>
{{"{{"}} end {{"}}"}}
`

const tplCaisLogo = `{{"{{"}} define "cais_logo" {{"}}"}}
<img
  src="/static/img/go-on-cais.jpg"
  alt="Go on Cais"
  width="1024"
  height="683"
  class="w-full max-w-lg rounded-2xl shadow-xl shadow-amber-950/15 ring-1 ring-amber-900/10"
/>
{{"{{"}} end {{"}}"}}
`

const tplPageHome = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} .AppName {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-screen flex-col items-center justify-center px-6 py-14 text-center">
  {{"{{"}} template "cais_logo" . {{"}}"}}
  <h1 class="mt-10 font-serif text-4xl font-semibold tracking-tight text-stone-800 md:text-5xl">{{"{{"}} t "home.rails_heading" {{"}}"}}</h1>
  <p class="mt-3 max-w-md text-lg text-stone-600">{{"{{"}} t "home.rails_subtitle" .AppName {{"}}"}}</p>
  <p class="mt-6 text-sm font-medium uppercase tracking-[0.2em] text-amber-900/60">{{"{{"}} t "home.stack" {{"}}"}}</p>
  <div class="mt-12 w-full max-w-lg rounded-2xl border border-amber-900/10 bg-white/45 p-8 text-left shadow-xl shadow-amber-950/5 backdrop-blur-sm">
    <h2 class="mb-5 text-xs font-semibold uppercase tracking-wider text-stone-500">{{"{{"}} t "home.next_steps" {{"}}"}}</h2>
    <ol class="space-y-5 text-stone-700">
      <li class="flex gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-amber-800/10 text-xs font-bold text-amber-950">1</span>
        <div>
          <p class="font-medium text-stone-800">{{"{{"}} t "home.step_resource" {{"}}"}}</p>
          <code class="mt-1.5 block rounded-lg bg-stone-100/90 px-3 py-2 font-mono text-xs text-stone-600">cais g resource item --fields name:string --public</code>
        </div>
      </li>
      <li class="flex gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-amber-800/10 text-xs font-bold text-amber-950">2</span>
        <div>
          <p class="font-medium text-stone-800">{{"{{"}} t "home.step_dev" {{"}}"}}</p>
          <code class="mt-1.5 block rounded-lg bg-stone-100/90 px-3 py-2 font-mono text-xs text-stone-600">cais dev</code>
        </div>
      </li>
      <li class="flex gap-3">
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-amber-800/10 text-xs font-bold text-amber-950">3</span>
        <div>
          <p class="font-medium text-stone-800">{{"{{"}} t "home.step_docs" {{"}}"}}</p>
          <a href="https://github.com/puppe1990/cais" class="mt-1 inline-block text-sm text-amber-900 underline decoration-amber-700/40 underline-offset-2 hover:decoration-amber-800">github.com/puppe1990/cais</a>
        </div>
      </li>
    </ol>
  </div>
  <p class="mt-10 text-xs text-stone-500/90">{{"{{"}} t "home.powered_by" {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageContact = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "contact.title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "contact.heading" {{"}}"}}</h2>
  <form
    id="contact-form"
    hx-post="/contact"
    hx-target="#form-errors"
    hx-swap="innerHTML swap:150ms"
    hx-indicator="#contact-spinner"
    hx-disabled-elt="button[type='submit']"
  >
    <div id="form-errors"></div>
    <label class="block mb-2 text-sm font-medium text-slate-700" for="name">{{"{{"}} t "contact.name_label" {{"}}"}}</label>
    <input
      class="w-full border border-slate-300 rounded-lg px-3 py-2 mb-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
      type="text"
      id="name"
      name="name"
      required
    />
    <label class="block mb-2 text-sm font-medium text-slate-700" for="email">{{"{{"}} t "contact.email_label" {{"}}"}}</label>
    <input
      class="w-full border border-slate-300 rounded-lg px-3 py-2 mb-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
      type="email"
      id="email"
      name="email"
      required
    />
    <button
      class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition"
      type="submit"
    >
      <span class="htmx-indicator" id="contact-spinner">{{"{{"}} t "contact.sending" {{"}}"}}</span>
      <span class="htmx-request-hide">{{"{{"}} t "contact.submit" {{"}}"}}</span>
    </button>
  </form>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageDashboard = `{{"{{"}} define "title" {{"}}"}}Dashboard{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="space-y-8">
  <div>
    <h2 class="text-3xl font-bold text-slate-800">Dashboard</h2>
    <p class="text-slate-500 mt-1">Visão geral do seu app {{.AppName}}</p>
  </div>
  <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
    <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-6 hover:shadow-md transition">
      <div class="flex items-center justify-between">
        <p class="text-sm font-semibold text-slate-500 uppercase tracking-wide">Total Contacts</p>
        <span class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-indigo-100 text-indigo-600">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </span>
      </div>
      <p class="mt-4 text-4xl font-bold text-indigo-600">{{"{{"}} .TotalContacts {{"}}"}}</p>
      <p class="mt-1 text-sm text-slate-400">contatos cadastrados</p>
    </div>
    <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-6 hover:shadow-md transition">
      <div class="flex items-center justify-between">
        <p class="text-sm font-semibold text-slate-500 uppercase tracking-wide">Environment</p>
        <span class="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-emerald-100 text-emerald-600">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </span>
      </div>
      <p class="mt-4 text-4xl font-bold text-emerald-600 capitalize">{{"{{"}} .Env {{"}}"}}</p>
      <p class="mt-1 text-sm text-slate-400">ambiente atual</p>
    </div>
  </div>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPartialErrors = `{{"{{- "}}define "contact_errors" -{{"}}"}}
<div class="text-red-600 text-sm mb-4">{{"{{"}} .Message {{"}}"}}</div>
{{"{{- "}}end -{{"}}"}}
`

const tplPartialSuccess = `{{"{{- "}}define "contact_success" -{{"}}"}}
<div class="text-green-600 text-sm mb-4">{{"{{"}} t "contact.success" {{"}}"}}</div>
{{"{{- "}}end -{{"}}"}}
`

const tplInputCSS = `@tailwind base;
@tailwind components;
@tailwind utilities;

@layer components {
  .htmx-swapping {
    opacity: 0;
    transition: opacity 150ms ease-out;
  }

  .htmx-settling {
    opacity: 1;
    transition: opacity 150ms ease-in;
  }

  form.htmx-request button[type="submit"] {
    @apply opacity-60 pointer-events-none;
  }

  .htmx-indicator {
    @apply hidden;
  }

  .htmx-request .htmx-indicator {
    @apply inline-block;
  }

  .htmx-request .htmx-request-hide {
    @apply hidden;
  }
}
`

const tplTailwind = `/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html"],
  theme: {
    extend: {},
  },
  plugins: [],
};
`

const tplPackageJSON = `{
  "private": true,
  "devDependencies": {
    "prettier": "^3.5.3",
    "tailwindcss": "^3.4.17"
  },
  "scripts": {
    "format": "prettier --write .",
    "format:check": "prettier --check .",
    "test": "npm run format:check"
  }
}
`

const tplMakefile = `.PHONY: dev build test css css-watch lint format format-check pre-commit-install ci

CAIS := $(shell command -v cais 2>/dev/null || command -v $(HOME)/go/bin/cais 2>/dev/null)

BIN := bin/server
CSS_IN := input.css
CSS_OUT := web/static/css/styles.css

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

format:
	npm run format

format-check:
	npm run format:check

pre-commit-install:
	pre-commit install

ci: test lint format-check

css:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --minify

css-watch:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --watch

build: css
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/server

dev: css
	$(MAKE) css-watch &
	$(CAIS) dev
`

const tplCIWorkflow = `name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Run tests
        run: go test ./... -race -count=1 -v

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v7
        with:
          version: v2.12.2

  js:
    name: JS
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: npm

      - run: npm ci

      - name: Prettier
        run: npx prettier --check .

      - name: npm test
        run: npm test
`

const tplPreCommitConfig = `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
        exclude: ^web/static/
      - id: end-of-file-fixer
        exclude: ^web/static/
      - id: check-yaml
      - id: check-added-large-files

  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v4.0.0-alpha.8
    hooks:
      - id: prettier
        exclude: ^web/static/

  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: go fmt ./...
        language: system
        pass_filenames: false
        types: [go]

      - id: go-test
        name: go test
        entry: go test ./... -race -count=1
        language: system
        pass_filenames: false
        types: [go]

      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint run ./...
        language: system
        pass_filenames: false
        types: [go]

      - id: npm-test
        name: npm test
        entry: npm test
        language: system
        pass_filenames: false
        files: \.(js|json|css|html|md|ya?ml)$
`

const tplGolangci = `version: "2"

linters:
  default: none
  enable:
    - errcheck
    - gocritic
    - govet
    - ineffassign
    - staticcheck
    - unused
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$

formatters:
  enable:
    - gofmt
    - goimports
  settings:
    goimports:
      local-prefixes:
        - {{.ModulePath}}
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$
`

const tplPrettierrc = `{
  "printWidth": 100,
  "tabWidth": 2,
  "useTabs": false,
  "semi": true,
  "singleQuote": false,
  "trailingComma": "es5",
  "bracketSameLine": false,
  "htmlWhitespaceSensitivity": "css",
  "overrides": [
    {
      "files": "*.html",
      "options": {
        "parser": "html"
      }
    }
  ]
}
`

const tplPrettierignore = `node_modules/
bin/
tmp/
data/
web/templates/
web/static/css/styles.css
web/static/js/htmx.min.js
package-lock.json
go.sum
`

const tplGitignore = `bin/
data/
web/static/css/styles.css
node_modules/
tmp/
.air/
*.db
.DS_Store
`

const tplREADME = "# {{.AppName}}\n\n" +
	"Full-stack Go app built with [Cais](https://github.com/puppe1990/cais): server-side HTML, HTMX, Tailwind, and SQLite.\n\n" +
	"## Stack\n\n" +
	"- Go 1.26 (net/http stdlib)\n" +
	"- html/template + HTMX 2.x\n" +
	"- Tailwind CSS 3.x\n" +
	"- SQLite (modernc.org/sqlite, no CGO)\n\n" +
	"## Quick start\n\n" +
	"```bash\n" +
	"cais install  # npm install + go mod tidy\n" +
	"cais dev        # http://localhost:8080\n" +
	"cais test       # full test suite\n" +
	"cais build      # bin/server\n" +
	"```\n\n" +
	"## Cais CLI\n\n" +
	"This app was scaffolded with the Cais CLI. Useful commands:\n\n" +
	"```bash\n" +
	"cais install               # npm install + go mod tidy\n" +
	"cais css                   # build Tailwind\n" +
	"cais dev                   # hot reload + tailwind watch\n" +
	"cais server                # go run ./cmd/server\n" +
	"cais console               # interactive Go REPL + SQL\n" +
	"cais g handler <name>      # handler + test + page template\n" +
	"cais g resource <name>     # model + migration + admin CRUD\n" +
	"cais g page <name>         # page template only\n" +
	"cais g migration <name>    # SQL migration file\n" +
	"cais test                  # go test ./...\n" +
	"cais doctor                # verify setup\n" +
	"```\n\n" +
	"## CI and pre-commit\n\n" +
	"GitHub Actions runs Go tests, `golangci-lint`, Prettier, and `npm test` on every push/PR to `main`.\n\n" +
	"```bash\n" +
	"make pre-commit-install   # once: installs git hooks\n" +
	"make ci                   # test + lint + format-check locally\n" +
	"```\n\n" +
	"Pre-commit hooks run: trailing whitespace, Prettier, `go fmt`, `go test`, `golangci-lint`, and `npm test`.\n\n" +
	"## Structure\n\n" +
	"```\n" +
	"pkg/cais/          → framework (via dependency)\n" +
	"internal/app/      → bootstrap and routes\n" +
	"internal/handlers/ → HTTP handlers\n" +
	"internal/store/    → SQLite + migrations\n" +
	"web/templates/     → HTML\n" +
	"web/static/        → CSS + JS\n" +
	"cmd/server/        → entry point\n" +
	"```\n\n" +
	"## Environment variables\n\n" +
	"| Variable  | Default         | Description      |\n" +
	"| --------- | --------------- | ---------------- |\n" +
	"| PORT      | :8080           | Server port      |\n" +
	"| DB_PATH   | ./data/app.db   | SQLite file path |\n" +
	"| ENV       | development     | Environment      |\n\n" +
	"Health check: GET /health → {\"status\":\"ok\"}\n"

const tplGenericHandler = `package handlers

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
)

type {{.Pascal}}Handler struct {
	renderer *cais.Renderer
}

func New{{.Pascal}}Handler(renderer *cais.Renderer) *{{.Pascal}}Handler {
	return &{{.Pascal}}Handler{renderer: renderer}
}

func (h *{{.Pascal}}Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", "{{.Snake}}", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
`

const tplGenericHandlerTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test{{.Pascal}}Handler_Returns200(t *testing.T) {
	h := New{{.Pascal}}Handler(setupTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/{{.Snake}}", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func Test{{.Pascal}}Handler_ContainsTitle(t *testing.T) {
	h := New{{.Pascal}}Handler(setupTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/{{.Snake}}", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "{{.Title}}") {
		t.Errorf("body missing title, got: %s", rr.Body.String())
	}
}
`

const tplGenericPage = `{{"{{"}} define "title" {{"}}"}}{{.Title}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-2">{{.Title}}</h2>
  <p class="text-slate-600">{{.Title}} page — customize this template.</p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplRoutesMinimal = `package app

import (
	"github.com/puppe1990/cais/pkg/cais"
	"{{.ModulePath}}/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog)
	r.Get("/", home.ServeHTTP)
}
`

const tplStoreMinimal = `package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/sqllog"
	_ "modernc.org/sqlite"
)

type Store interface {
	Ping() error
	Close() error
}

type SQLiteStore struct {
	db *sqllog.DB
}

func NewSQLiteStore(dsn string, env string) (*SQLiteStore, error) {
	if dsn != ":memory:" {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	cfg := sqllog.Config{Enabled: sqllog.EnabledForEnv(env)}
	if cfg.Enabled {
		cfg.Writer = devlog.MirrorDefault(os.Stdout)
	}
	return &SQLiteStore{db: sqllog.Wrap(db, cfg)}, nil
}

func (s *SQLiteStore) Ping() error {
	return s.db.Raw().Ping()
}

func (s *SQLiteStore) DB() *sql.DB {
	return s.db.Raw()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
`

const tplStoreTestMinimal = `package store

import "testing"

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStore_Migrations(t *testing.T) {
	_ = newTestStore(t)
}
`

const tplLayoutMinimal = `{{"{{"}} define "title" {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{.AppName}} — powered by Cais{{"{{"}} end {{"}}"}}
{{"{{"}} define "base" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} htmlLang {{"}}"}}">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
    {{"{{"}} if .CSRFToken {{"}}"}}<meta name="csrf-token" content="{{"{{"}} .CSRFToken {{"}}"}}" />{{"{{"}} end {{"}}"}}
    <title>{{"{{"}} template "title" . {{"}}"}}</title>
    <meta name="description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="{{.AppName}}" />
    <meta property="og:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta property="og:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <meta property="og:locale" content="{{"{{"}} ogLocale {{"}}"}}" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta name="twitter:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta name="twitter:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#4f46e5" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="{{.AppName}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />
    <script src="/static/js/htmx.min.js" defer></script>
    <script src="/static/js/cais.js" defer></script>
  </head>
  <body class="bg-slate-50 text-slate-900 min-h-screen flex flex-col">
    <header class="bg-white border-b border-slate-200 p-4 shadow-sm">
      <div class="max-w-5xl mx-auto flex justify-between items-center">
        <a href="/" class="font-bold text-xl text-indigo-600 hover:text-indigo-700 transition">{{.AppName}}</a>
      </div>
    </header>
    <main class="flex-grow max-w-5xl w-full mx-auto p-6">{{"{{"}} template "content" . {{"}}"}}</main>
    <footer class="border-t border-slate-200 p-4 text-center text-sm text-slate-500">
      {{.AppName}} — powered by Cais
    </footer>
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
  </body>
</html>
{{"{{"}} end {{"}}"}}
`

const tplMainBlank = `package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/boot"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"{{.ModulePath}}/internal/app"
	appi18n "{{.ModulePath}}/internal/i18n"
	"{{.ModulePath}}/internal/store"
	"{{.ModulePath}}/web"
)

func main() {
	cfg := cais.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	preferredPort := cfg.Port
	port, shifted, err := cais.ResolvePort(cfg.Port, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Port = port

	a, err := bootstrapWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	shiftedFrom := ""
	if shifted {
		shiftedFrom = preferredPort
	}
	boot.Print(os.Stdout, boot.Options{
		AppName:         "{{.AppName}}",
		Config:          cfg,
		Version:         boot.CaisVersion(),
		PortShiftedFrom: shiftedFrom,
	})
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func bootstrap() (*app.App, error) {
	return bootstrapWithConfig(cais.Load())
}

func bootstrapWithConfig(cfg cais.Config) (*app.App, error) {
	tmplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	catalog := appi18n.NewCatalog(cfg.Locale)
	renderer, err := cais.NewRenderer(tmplFS, catalog)
	if err != nil {
		return nil, fmt.Errorf("renderer: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	staticDir, err := findWebDir("static")
	if err != nil {
		_ = s.Close()
		return nil, err
	}

	return app.New(cfg, app.Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: staticDir,
		Site:      meta.SiteFrom("{{.AppName}}", cfg.AppURL),
		Catalog:   catalog,
	})
}

func findWebDir(subpath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(wd, "web", subpath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("web/%s not found", subpath)
		}
		wd = parent
	}
}
`

const tplAppBlank = `package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"{{.ModulePath}}/internal/store"
)

type Deps struct {
	Renderer  *cais.Renderer
	Store     store.Store
	StaticDir string
	Site      meta.Site
	Catalog   *i18n.Catalog
}

type App struct {
	config cais.Config
	store  store.Store
	router *cais.Router
	server *http.Server
}

func New(cfg cais.Config, deps Deps) (*App, error) {
	if deps.Renderer == nil {
		return nil, fmt.Errorf("renderer is required")
	}
	if deps.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	site := deps.Site
	if site.AppName == "" {
		site = meta.SiteFrom("{{.AppName}}", cfg.AppURL)
	}
	deps.Site = site

	r := cais.NewRouter()
	r.Use(middleware.CSRF(cfg))
	buf := devlog.Prepare(cfg.Env)
	if buf != nil {
		r.Use(middleware.LoggerTo(devlog.MirrorDefault(log.Writer())))
	} else {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recover)
	r.Use(middleware.SecurityHeaders(cfg))
	r.Static("/static", deps.StaticDir)

	registerRoutes(r, deps, cfg)
	devlog.Register(r, cfg.Env, buf)
	r.Get("/health", healthHandler(deps.Store))

	return &App{
		config: cfg,
		store:  deps.Store,
		router: r,
		server: &http.Server{
			Addr:              cfg.Port,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}, nil
}

func healthHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if err := s.Ping(); err != nil {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
}

func (a *App) Handler() http.Handler {
	return a.router
}

func (a *App) Run() error {
	return a.RunContext(context.Background())
}

func (a *App) RunContext(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			_ = a.store.Close()
			return err
		}
		<-errCh
		_ = a.store.Close()
		return nil
	case err := <-errCh:
		_ = a.store.Close()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
`

const tplRoutesBlank = `package app

import (
	"github.com/puppe1990/cais/pkg/cais"
	"{{.ModulePath}}/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog)
	r.Get("/", home.ServeHTTP)
}
`

const tplLayoutBlank = `{{"{{"}} define "title" {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{.AppName}} — powered by Cais{{"{{"}} end {{"}}"}}
{{"{{"}} define "base" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} htmlLang {{"}}"}}">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
    {{"{{"}} if .CSRFToken {{"}}"}}<meta name="csrf-token" content="{{"{{"}} .CSRFToken {{"}}"}}" />{{"{{"}} end {{"}}"}}
    <title>{{"{{"}} template "title" . {{"}}"}}</title>
    <meta name="description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="{{.AppName}}" />
    <meta property="og:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta property="og:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta property="og:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <meta property="og:locale" content="{{"{{"}} ogLocale {{"}}"}}" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="{{"{{"}} template "title" . {{"}}"}}" />
    <meta name="twitter:description" content="{{"{{"}} template "description" . {{"}}"}}" />
    <meta name="twitter:image" content="{{"{{"}} absURL .AppURL "/static/og.png" {{"}}"}}" />
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#4f46e5" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="{{.AppName}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />
    <script src="/static/js/htmx.min.js" defer></script>
    <script src="/static/js/cais.js" defer></script>
  </head>
  <body class="bg-slate-50 text-slate-900 min-h-screen flex flex-col">
    <header class="bg-white border-b border-slate-200 p-4 shadow-sm">
      <div class="max-w-5xl mx-auto flex justify-between items-center">
        <a href="/" class="font-bold text-xl text-indigo-600 hover:text-indigo-700 transition">{{.AppName}}</a>
      </div>
    </header>
    <main class="flex-grow max-w-5xl w-full mx-auto p-6">{{"{{"}} template "content" . {{"}}"}}</main>
    <footer class="border-t border-slate-200 p-4 text-center text-sm text-slate-500">
      {{.AppName}} — powered by Cais
    </footer>
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
  </body>
</html>
{{"{{"}} end {{"}}"}}
`

const tplREADMEBlank = "# {{.AppName}}\n\n" +
	"Full-stack Go app built with [Cais](https://github.com/puppe1990/cais): server-side HTML, HTMX, Tailwind, and SQLite.\n\n" +
	"## Quick start\n\n" +
	"```bash\n" +
	"cais install  # npm install + go mod tidy\n" +
	"cais dev        # http://localhost:8080\n" +
	"cais test       # full test suite\n" +
	"make ci         # test + lint + format-check\n" +
	"```\n\n" +
	"## CI and pre-commit\n\n" +
	"```bash\n" +
	"make pre-commit-install   # once: installs git hooks\n" +
	"make ci                   # test + lint + format-check locally\n" +
	"```\n\n" +
	"## Add your first resource\n\n" +
	"```bash\n" +
	"cais g resource bookmark --fields title:string,url:url,notes:text?\n" +
	"```\n\n" +
	"This generates:\n" +
	"- Model, migration, admin CRUD, and public list page\n" +
	"- Tests for handlers and store\n" +
	"- Routes with admin protection\n"

const tplEnvExample = `# Server
PORT=:8080
ENV=development
APP_URL=http://localhost:8080
LOCALE=en

# Database
DB_PATH=./data/app.db

# Security (required when ENV=production)
ADMIN_TOKEN=
`

const tplI18nCatalog = `package i18n

import (
	caisi18n "github.com/puppe1990/cais/pkg/cais/i18n"
)

var locales = map[string]map[string]string{
	"en": enMessages,
	"pt": ptMessages,
}

// NewCatalog returns a catalog for the given locale (en default, pt for pt-BR).
func NewCatalog(locale string) *caisi18n.Catalog {
	return caisi18n.NewCatalogFrom(locale, locales)
}

// DefaultCatalog returns the English catalog.
func DefaultCatalog() *caisi18n.Catalog {
	return NewCatalog(caisi18n.DefaultLocale)
}
`

const tplI18nEn = `package i18n

var enMessages = map[string]string{
	"auth.invalid_credentials": "Invalid email or password.",
	"auth.welcome":             "Welcome!",
	"auth.login_title":         "Sign in",
	"auth.login_submit":        "Sign in",
	"auth.password_label":      "Password",
	"auth.logout":              "Sign out",

	"contact.title":          "Contact",
	"contact.heading":        "Get in touch",
	"contact.name_label":     "Name",
	"contact.name_required":  "Name is required.",
	"contact.email_label":    "Email",
	"contact.email_required": "Email is required.",
	"contact.email_invalid":  "Enter a valid email.",
	"contact.submit":         "Send",
	"contact.sending":        "Sending…",
	"contact.success":        "Message sent successfully!",

	"home.title":            "Home",
	"home.welcome":          "Welcome, %s!",
	"home.tagline":          "Mini Go app with HTMX, Tailwind, and SQLite.",
	"home.contact_link":     "Contact",
	"home.default_name":     "Developer",
	"home.rails_heading":    "You're on Cais!",
	"home.rails_subtitle":   "%s is ready to sail.",
	"home.stack":            "Go · HTMX · Tailwind · SQLite",
	"home.next_steps":       "Next steps",
	"home.step_resource":    "Generate your first resource",
	"home.step_dev":         "Start the dev server",
	"home.step_docs":        "Explore the framework",
	"home.powered_by":       "Powered by Cais — lightweight apps on Lightsail",
	"home.minimal.tagline":  "Go app with HTMX, Tailwind, and SQLite — powered by Cais.",
	"home.minimal.hint":     "Use ` + "`cais g resource <name> --public`" + ` to get started.",

	"dashboard.title":    "Dashboard",
	"dashboard.contacts": "Contacts:",
	"dashboard.env":      "Environment:",

	"layout.footer": "Running light on Lightsail",
}
`

const tplI18nPt = `package i18n

var ptMessages = map[string]string{
	"auth.invalid_credentials": "Email ou senha inválidos.",
	"auth.welcome":             "Bem-vindo!",
	"auth.login_title":         "Entrar",
	"auth.login_submit":        "Entrar",
	"auth.password_label":      "Senha",
	"auth.logout":              "Sair",

	"contact.title":          "Contato",
	"contact.heading":        "Fale conosco",
	"contact.name_label":     "Nome",
	"contact.name_required":  "O campo nome é obrigatório.",
	"contact.email_label":    "Email",
	"contact.email_required": "O campo email é obrigatório.",
	"contact.email_invalid":  "Informe um email válido.",
	"contact.submit":         "Enviar",
	"contact.sending":        "Enviando…",
	"contact.success":        "Mensagem enviada com sucesso!",

	"home.title":            "Página Inicial",
	"home.welcome":          "Bem-vindo, %s!",
	"home.tagline":          "Mini app Go com HTMX, Tailwind e SQLite.",
	"home.contact_link":     "Contato",
	"home.default_name":     "Desenvolvedor",
	"home.rails_heading":    "Você está no Cais!",
	"home.rails_subtitle":   "%s está pronto para navegar.",
	"home.stack":            "Go · HTMX · Tailwind · SQLite",
	"home.next_steps":       "Próximos passos",
	"home.step_resource":    "Gere seu primeiro resource",
	"home.step_dev":         "Suba o servidor de desenvolvimento",
	"home.step_docs":        "Explore o framework",
	"home.powered_by":       "Powered by Cais — apps leves no Lightsail",
	"home.minimal.tagline":  "App Go com HTMX, Tailwind e SQLite — powered by Cais.",
	"home.minimal.hint":     "Use ` + "`cais g resource <name> --public`" + ` para começar.",

	"dashboard.title":    "Dashboard",
	"dashboard.contacts": "Contatos:",
	"dashboard.env":      "Ambiente:",

	"layout.footer": "Rodando leve no Lightsail",
}
`

const tplI18nTest = `package i18n

import "testing"

func TestDefaultCatalog_english(t *testing.T) {
	c := DefaultCatalog()
	if got := c.T("auth.welcome"); got != "Welcome!" {
		t.Errorf("T(auth.welcome) = %q", got)
	}
}

func TestNewCatalog_portuguese(t *testing.T) {
	c := NewCatalog("pt-BR")
	if got := c.T("auth.welcome"); got != "Bem-vindo!" {
		t.Errorf("T(auth.welcome) = %q", got)
	}
	if c.HTMLLang() != "pt-BR" {
		t.Errorf("HTMLLang() = %q, want pt-BR", c.HTMLLang())
	}
}
`
