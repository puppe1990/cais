package cli

const tplGoMod = `module {{.ModulePath}}

go 1.26

require (
	github.com/matheuspuppe/cais v0.0.0
	modernc.org/sqlite v1.53.0
)

replace github.com/matheuspuppe/cais => ../Cais
`

const tplMain = `package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"{{.ModulePath}}/internal/app"
	"{{.ModulePath}}/internal/store"
	"github.com/matheuspuppe/cais/pkg/cais"
	"{{.ModulePath}}/web"
)

func main() {
	cfg := cais.Load()
	a, err := bootstrapWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("{{.AppName}} rodando na porta %s...", cfg.Port)
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

	renderer, err := cais.NewRenderer(tmplFS)
	if err != nil {
		return nil, fmt.Errorf("renderer: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath)
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

const tplApp = `package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"{{.ModulePath}}/internal/store"
	"github.com/matheuspuppe/cais/pkg/cais"
	"github.com/matheuspuppe/cais/pkg/cais/middleware"
)

type Deps struct {
	Renderer  *cais.Renderer
	Store     store.Store
	StaticDir string
}

type App struct {
	config cais.Config
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

	r := cais.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recover)
	r.Static("/static", deps.StaticDir)

	registerRoutes(r, deps)
	r.Get("/health", healthHandler)

	return &App{
		config: cfg,
		router: r,
		server: &http.Server{
			Addr:    cfg.Port,
			Handler: r,
		},
	}, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
			return err
		}
		<-errCh
		return nil
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
`

const tplRoutes = `package app

import (
	"{{.ModulePath}}/internal/handlers"
	"github.com/matheuspuppe/cais/pkg/cais"
)

func registerRoutes(r *cais.Router, deps Deps) {
	home := handlers.NewHomeHandler(deps.Renderer)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store)
	dashboard := handlers.NewDashboardHandler(deps.Renderer, deps.Store)

	r.Get("/", home.ServeHTTP)
	r.Get("/contact", contact.Get)
	r.Post("/contact", contact.Post)
	r.Get("/dashboard", dashboard.ServeHTTP)
}
`

const tplHomeHandler = `package handlers

import (
	"net/http"

	"github.com/matheuspuppe/cais/pkg/cais"
)

type PageData struct {
	Nome string
}

type HomeHandler struct {
	renderer *cais.Renderer
}

func NewHomeHandler(renderer *cais.Renderer) *HomeHandler {
	return &HomeHandler{renderer: renderer}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", "home", PageData{Nome: "{{.AppName}}"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	h := NewHomeHandler(setupTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_ContainsWelcome(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "Bem-vindo, {{.AppName}}!") {
		t.Errorf("body missing welcome message, got: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t))

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
	"github.com/matheuspuppe/cais/pkg/cais"
)

type ContactHandler struct {
	renderer *cais.Renderer
	store    store.Store
}

type contactErrorData struct {
	Message string
}

func NewContactHandler(renderer *cais.Renderer, s store.Store) *ContactHandler {
	return &ContactHandler{renderer: renderer, store: s}
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", "contact", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ContactHandler) Post(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))

	if email == "" {
		h.renderContactResponse(w, r, http.StatusUnprocessableEntity, "contact_errors", contactErrorData{
			Message: "O campo email é obrigatório.",
		})
		return
	}

	_, err := h.store.InsertContact(models.Contact{Name: name, Email: email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderContactResponse(w, r, http.StatusOK, "contact_success", nil)
}

func (h *ContactHandler) renderContactResponse(w http.ResponseWriter, r *http.Request, status int, tmpl string, data any) {
	w.WriteHeader(status)
	if cais.IsHTMX(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.renderer.RenderPartial(w, tmpl, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", "contact", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
`

const tplContactTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"{{.ModulePath}}/internal/store"
)

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestContactHandler_Get_ReturnsForm(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t))

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

func TestContactHandler_Post_InvalidEmail_Returns422(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t))

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
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t))

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("expected partial HTML, got full page")
	}
	if !strings.Contains(body, "email") {
		t.Errorf("body missing error message, got: %s", body)
	}
}

func TestContactHandler_Post_Valid_SavesAndReturnsSuccess(t *testing.T) {
	s := setupTestStore(t)
	h := NewContactHandler(setupTestRenderer(t), s)

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=alice@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "sucesso") {
		t.Errorf("body missing success message, got: %s", rr.Body.String())
	}
}
`

const tplDashboardHandler = `package handlers

import (
	"net/http"

	"{{.ModulePath}}/internal/store"
	"github.com/matheuspuppe/cais/pkg/cais"
)

type DashboardData struct {
	TotalContacts int64
	Env           string
}

type DashboardHandler struct {
	renderer *cais.Renderer
	store    store.Store
}

func NewDashboardHandler(renderer *cais.Renderer, s store.Store) *DashboardHandler {
	return &DashboardHandler{renderer: renderer, store: s}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.CountContacts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		TotalContacts: count,
		Env:           cais.Load().Env,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", "dashboard", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
`

const tplDashboardTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardHandler_Returns200(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestDashboardHandler_ContainsDashboard(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t))

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
	"os"
	"path/filepath"
	"testing"

	"github.com/matheuspuppe/cais/pkg/cais"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}

func setupTestRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	templatesDir := filepath.Join(projectRoot(t), "web", "templates")
	r, err := cais.NewRendererFromDir(templatesDir)
	if err != nil {
		t.Fatal(err)
	}
	return r
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

	"{{.ModulePath}}/internal/models"
)

type Store interface {
	InsertContact(contact models.Contact) (int64, error)
	FindContact(id int64) (models.Contact, error)
	CountContacts() (int64, error)
	Close() error
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
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

	return &SQLiteStore{db: db}, nil
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
	s, err := NewSQLiteStore(":memory:")
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
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func applyMigrations(db *sql.DB) error {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		sqlBytes, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}

	return nil
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

const tplLayout = `{{"{{"}} define "base" {{"}}"}}
<!doctype html>
<html lang="pt-BR">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{"{{"}} block "title" . {{"}}"}}}{{.AppName}}{{"{{"}} end {{"}}"}}</title>
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#4f46e5" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="default" />
    <meta name="apple-mobile-web-app-title" content="{{.AppName}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon-192.png" />
    <link rel="icon" href="/static/icons/icon.svg" type="image/svg+xml" />
    <script src="/static/js/htmx.min.js" defer></script>
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

const tplPageHome = `{{"{{"}} define "title" {{"}}"}}Página Inicial{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-2">Bem-vindo, {{"{{"}} .Nome {{"}}"}}!</h2>
  <p class="text-slate-600 mb-4">Mini app Go com HTMX, Tailwind e SQLite.</p>
  <a
    href="/contact"
    class="block w-full text-center bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition shadow-sm"
  >
    Contato
  </a>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageContact = `{{"{{"}} define "title" {{"}}"}}Contato{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">Fale conosco</h2>
  <form id="contact-form" hx-post="/contact" hx-target="#form-errors" hx-swap="innerHTML">
    <div id="form-errors"></div>
    <label class="block mb-2 text-sm font-medium text-slate-700" for="name">Nome</label>
    <input
      class="w-full border border-slate-300 rounded-lg px-3 py-2 mb-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
      type="text"
      id="name"
      name="name"
      required
    />
    <label class="block mb-2 text-sm font-medium text-slate-700" for="email">Email</label>
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
      Enviar
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
<div class="text-green-600 text-sm mb-4">Mensagem enviada com sucesso!</div>
{{"{{- "}}end -{{"}}"}}
`

const tplInputCSS = `@tailwind base;
@tailwind components;
@tailwind utilities;
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
    "format:check": "prettier --check ."
  }
}
`

const tplMakefile = `.PHONY: dev build test css css-watch

BIN := bin/server
CSS_IN := input.css
CSS_OUT := web/static/css/styles.css

test:
	go test ./... -race -count=1

css:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --minify

css-watch:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --watch

build: css
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/server

AIR := $(shell command -v $(HOME)/go/bin/air 2>/dev/null || command -v air 2>/dev/null)

dev: css
	$(MAKE) css-watch &
	$(AIR) -c .air.toml
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
	"Full-stack Go app built with [Cais](https://github.com/matheuspuppe/cais): server-side HTML, HTMX, Tailwind, and SQLite.\n\n" +
	"## Stack\n\n" +
	"- Go 1.26 (net/http stdlib)\n" +
	"- html/template + HTMX 2.x\n" +
	"- Tailwind CSS 3.x\n" +
	"- SQLite (modernc.org/sqlite, no CGO)\n\n" +
	"## Quick start\n\n" +
	"```bash\n" +
	"npm install\n" +
	"make dev      # http://localhost:8080\n" +
	"make test     # full test suite\n" +
	"make build    # builds bin/server\n" +
	"```\n\n" +
	"## Cais CLI\n\n" +
	"This app was scaffolded with the Cais CLI. Useful commands:\n\n" +
	"```bash\n" +
	"cais g handler <name>      # handler + test + page template\n" +
	"cais g page <name>         # page template only\n" +
	"cais g migration <name>    # SQL migration file\n" +
	"cais server                # go run ./cmd/server\n" +
	"cais test                  # go test ./...\n" +
	"```\n\n" +
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

	"github.com/matheuspuppe/cais/pkg/cais"
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
