package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scaffoldAuth(dir string, data scaffoldData) error {
	if _, err := os.Stat(filepath.Join(dir, "internal/handlers/auth.go")); err == nil {
		return fmt.Errorf("auth already exists — remove internal/handlers/auth.go first")
	}

	files := map[string]string{
		"internal/models/user.go":                tplUserModel,
		"internal/handlers/auth.go":              tplAuthHandler,
		"internal/handlers/auth_test.go":         tplAuthTest,
		"internal/store/migrations/002_auth.sql": tplMigration002Auth,
		"web/templates/pages/login.html":         tplPageLogin,
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := writeTemplate(full, content, data); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		_, _ = fmt.Printf("  create %s\n", path)
	}

	if err := patchStoreForAuth(dir); err != nil {
		return err
	}
	if err := patchAppForAuth(dir); err != nil {
		return err
	}
	if err := patchRoutesForAuth(dir); err != nil {
		return err
	}

	return nil
}

func patchStoreForAuth(dir string) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "FindUserByEmail") {
		return nil
	}

	if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/session") {
		content = strings.Replace(content,
			`"github.com/puppe1990/cais/pkg/cais/sqllog"`,
			`"github.com/puppe1990/cais/pkg/cais/session"
	"github.com/puppe1990/cais/pkg/cais/sqllog"`,
			1,
		)
	}

	content = strings.Replace(content,
		"\tCountContacts() (int64, error)\n\tClose() error",
		"\tCountContacts() (int64, error)\n\tFindUserByEmail(email string) (models.User, error)\n\tSessions() session.Store\n\tClose() error",
		1,
	)

	insert := `
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
`
	content = strings.Replace(content, "\nfunc (s *SQLiteStore) Close()", insert+"\nfunc (s *SQLiteStore) Close()", 1)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update internal/store/store.go")
	return nil
}

func patchAppForAuth(dir string) error {
	path := filepath.Join(dir, "internal/app/app.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	changed := false

	if !strings.Contains(content, "LoadSession") {
		if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/session") {
			content = strings.Replace(content,
				`"github.com/puppe1990/cais/pkg/cais/middleware"`,
				`"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/cais/pkg/cais/session"`,
				1,
			)
		}
		content = strings.Replace(content,
			"r.Use(middleware.CSRF)\n",
			"r.Use(middleware.CSRF)\n\tr.Use(middleware.LoadSession(deps.Store.Sessions()))\n\tr.Use(middleware.Flash)\n",
			1,
		)
		changed = true
	} else if !strings.Contains(content, "middleware.Flash") {
		content = strings.Replace(content,
			"r.Use(middleware.LoadSession(deps.Store.Sessions()))\n",
			"r.Use(middleware.LoadSession(deps.Store.Sessions()))\n\tr.Use(middleware.Flash)\n",
			1,
		)
		changed = true
	}

	if !strings.Contains(content, "SecurityHeaders") {
		content = strings.Replace(content,
			"r.Use(middleware.Recover)\n",
			"r.Use(middleware.Recover)\n\tr.Use(middleware.SecurityHeaders(cfg))\n",
			1,
		)
		changed = true
	}

	if !changed {
		return nil
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update internal/app/app.go")
	return nil
}

func patchRoutesForAuth(dir string) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "NewAuthHandler") {
		return nil
	}

	if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/middleware") {
		content = strings.Replace(content,
			`"github.com/puppe1990/cais/pkg/cais"`,
			`"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"`,
			1,
		)
	}
	if !strings.Contains(content, `"net/http"`) {
		content = strings.Replace(content,
			`import (
	"github.com/puppe1990/cais/pkg/cais"`,
			`import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"`,
			1,
		)
	}

	insert := `	loginLimit := middleware.NewRateLimiter(10)

	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, deps.Site, deps.Store.Sessions(), cfg)
	r.Get("/login", auth.Login)
	r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
	r.Post("/logout", auth.LogoutPost)

`
	content = strings.Replace(content, "func registerRoutes", insert+"func registerRoutes", 1)

	content = strings.Replace(content,
		`r.Get("/dashboard", dashboard.ServeHTTP)`,
		`r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))`,
		1,
	)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update internal/app/routes.go")
	return nil
}

const tplUserModel = `package models

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
`

const tplMigration002Auth = `CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY NOT NULL,
    user_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const tplAuthHandler = `package handlers

import (
	"net/http"
	"strings"

	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/session"
)

type AuthHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	sessions session.Store
	cfg      cais.Config
}

type loginData struct {
	meta.Site
	Error string
}

func NewAuthHandler(renderer *cais.Renderer, s store.Store, site meta.Site, sessions session.Store, cfg cais.Config) *AuthHandler {
	return &AuthHandler{renderer: renderer, store: s, site: site, sessions: sessions, cfg: cfg}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "login", loginData{Site: meta.ForRequest(h.site, r)})
}

func (h *AuthHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	user, err := h.store.FindUserByEmail(email)
	if err != nil || !session.VerifyPassword(user.PasswordHash, password) {
		httpx.RenderOrError(w, h.renderer, "base", "login", loginData{
			Site:  meta.ForRequest(h.site, r),
			Error: "Email ou senha inválidos.",
		})
		return
	}

	if err := session.SignIn(w, h.sessions, user.ID, session.CookieOptionsFromConfig(h.cfg)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flash.Set(w, "notice", "Bem-vindo!")
	httpx.SeeOther(w, r, "/dashboard")
}

func (h *AuthHandler) LogoutPost(w http.ResponseWriter, r *http.Request) {
	session.SignOut(w, h.sessions, r)
	httpx.SeeOther(w, r, "/login")
}
`

const tplAuthTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestAuth_Login_redirectsWhenAuthenticated(t *testing.T) {
	s := setupTestStore(t)
	sessions := s.Sessions()
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), sessions, cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestAuth_LoginPost_invalidCredentials(t *testing.T) {
	s := setupTestStore(t)
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), s.Sessions(), cais.Config{})

	form := url.Values{"email": {"nobody@example.com"}, "password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "inválidos") {
		t.Errorf("body missing error: %s", rr.Body.String())
	}
}
`

const tplPageLogin = `{{"{{"}} define "title" {{"}}"}}Login{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">Entrar</h2>
  {{"{{"}} if .Error {{"}}"}}<p class="text-red-600 text-sm mb-4">{{"{{"}} .Error {{"}}"}}</p>{{"{{"}} end {{"}}"}}
  <form method="post" action="/login" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="email">Email</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="email" id="email" name="email" required />
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password">Senha</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password" name="password" required />
    </div>
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
      Entrar
    </button>
  </form>
</div>
{{"{{"}} end {{"}}"}}
`
