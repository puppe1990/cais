# Cais Framework Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden Cais for production (sessions, security headers, rate limiting) and reduce handler/CLI boilerplate (HTMX render helper, field validation, flash messages, DB tooling).

**Architecture:** Four independently shippable phases. Each phase adds packages under `pkg/cais/` with tests first, then wires into the reference app (`internal/app`, `internal/handlers`). Generators (`internal/cli`) are updated last so scaffolded apps inherit the improvements. No new runtime dependencies — stdlib only.

**Tech Stack:** Go 1.22+, `html/template`, HTMX, SQLite (`modernc.org/sqlite`), existing Cais middleware/router/render stack.

**Prerequisite:** Run from a dedicated git worktree (`superpowers:using-git-worktrees`) before implementation.

---

## File Map

| File                               | Responsibility                                     |
| ---------------------------------- | -------------------------------------------------- |
| `pkg/cais/config.go`               | `CookieSecure()`, extended `Validate()`            |
| `pkg/cais/session/sqlite.go`       | `expires_at`, `PruneExpired`, reject expired `Get` |
| `pkg/cais/session/cookie.go`       | `CookieOptionsFromConfig(cfg)`                     |
| `pkg/cais/middleware/security.go`  | CSP, HSTS, X-Frame-Options, Referrer-Policy        |
| `pkg/cais/middleware/ratelimit.go` | In-memory per-IP token bucket                      |
| `pkg/cais/middleware/logger.go`    | `clientIP` with `X-Forwarded-For`                  |
| `pkg/cais/httpx/httpx.go`          | `RenderPageOrPartial`                              |
| `pkg/cais/validate/errors.go`      | `FieldErrors` map + `First()`                      |
| `pkg/cais/flash/flash.go`          | Set-once cookie flash messages                     |
| `pkg/cais/migrate/migrate.go`      | `RollbackLast`                                     |
| `internal/cli/db.go`               | `cais db rollback`, `cais db prune-sessions`       |
| `internal/app/app.go`              | Wire new middleware                                |
| `internal/handlers/auth.go`        | Rate limit + secure cookies                        |
| `internal/handlers/contact.go`     | Use `RenderPageOrPartial`                          |
| `internal/cli/templates.go`        | Generator output uses new helpers                  |

**Out of scope (separate plans):** resource pagination/search, new field types, docs site, structured JSON logging, `cais routes` command.

---

## Phase 1 — Security Foundation

### Task 1: Session expiry in SQLite store

**Files:**

- Modify: `pkg/cais/session/sqlite.go`
- Modify: `pkg/cais/session/sqlite_test.go`
- Create: `internal/store/migrations/004_session_expiry.sql` (reference app only if schema lives in migrations; otherwise schema is in `EnsureSQLiteSchema`)

- [ ] **Step 1: Write the failing test**

Add to `pkg/cais/session/sqlite_test.go`:

```go
func TestSQLiteStore_Get_rejectsExpiredSession(t *testing.T) {
	db := testDB(t)
	store := NewSQLiteStore(db)

	token, err := store.Create(42)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("UPDATE sessions SET expires_at = datetime('now', '-1 hour') WHERE token = ?", token)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := store.Get(token); ok {
		t.Fatal("expected expired session to be rejected")
	}
}

func TestSQLiteStore_PruneExpired_removesOldRows(t *testing.T) {
	db := testDB(t)
	store := NewSQLiteStore(db)

	token, err := store.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("UPDATE sessions SET expires_at = datetime('now', '-1 day') WHERE token = ?", token)
	if err != nil {
		t.Fatal(err)
	}

	n, err := store.PruneExpired()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("PruneExpired = %d, want 1", n)
	}
	if _, ok := store.Get(token); ok {
		t.Fatal("pruned session should not be found")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/cais/session/... -v -run 'TestSQLiteStore_Get_rejectsExpired|TestSQLiteStore_PruneExpired'`

Expected: FAIL — `PruneExpired` undefined or `Get` returns true for expired token.

- [ ] **Step 3: Write minimal implementation**

Update schema in `pkg/cais/session/sqlite.go`:

```go
const sqliteSchema = `CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY NOT NULL,
  user_id INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  expires_at TEXT NOT NULL DEFAULT (datetime('now', '+7 days'))
);`

const sessionTTL = 7 * 24 * time.Hour
```

Add import `"time"`.

Update `Create`:

```go
func (s *SQLiteStore) Create(userID int64) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	expires := time.Now().UTC().Add(sessionTTL).Format("2006-01-02 15:04:05")
	if _, err := s.db.Exec(
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expires,
	); err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return token, nil
}
```

Update `Get`:

```go
func (s *SQLiteStore) Get(token string) (int64, bool) {
	var id int64
	err := s.db.QueryRow(
		"SELECT user_id FROM sessions WHERE token = ? AND expires_at > datetime('now')",
		token,
	).Scan(&id)
	if err != nil {
		return 0, false
	}
	return id, true
}
```

Add `PruneExpired`:

```go
func (s *SQLiteStore) PruneExpired() (int64, error) {
	res, err := s.db.Exec("DELETE FROM sessions WHERE expires_at <= datetime('now')")
	if err != nil {
		return 0, fmt.Errorf("prune sessions: %w", err)
	}
	return res.RowsAffected()
}
```

Update `EnsureSQLiteSchema` to migrate existing tables (idempotent):

```go
func EnsureSQLiteSchema(db *sql.DB) error {
	if _, err := db.Exec(sqliteSchema); err != nil {
		return fmt.Errorf("sessions schema: %w", err)
	}
	// Add column on existing installs without breaking fresh installs.
	_, _ = db.Exec(`ALTER TABLE sessions ADD COLUMN expires_at TEXT NOT NULL DEFAULT (datetime('now', '+7 days'))`)
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/cais/session/... -v -run 'TestSQLiteStore_Get_rejectsExpired|TestSQLiteStore_PruneExpired'`

Expected: PASS

- [ ] **Step 5: Run full session package**

Run: `make test` or `go test ./pkg/cais/session/... -race`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/cais/session/sqlite.go pkg/cais/session/sqlite_test.go
git commit -m "feat(session): expire SQLite sessions and add PruneExpired"
```

---

### Task 2: Secure cookies in production

**Files:**

- Modify: `pkg/cais/config.go`
- Modify: `pkg/cais/config_test.go`
- Modify: `pkg/cais/session/cookie.go`
- Create: `pkg/cais/session/cookie_config_test.go`
- Modify: `internal/handlers/auth.go`

- [ ] **Step 1: Write the failing test**

Add to `pkg/cais/config_test.go`:

```go
func TestConfig_CookieSecure_trueInProduction(t *testing.T) {
	cfg := Config{Env: "production"}
	if !cfg.CookieSecure() {
		t.Error("CookieSecure() = false, want true in production")
	}
}

func TestConfig_CookieSecure_falseInDevelopment(t *testing.T) {
	cfg := Config{Env: "development"}
	if cfg.CookieSecure() {
		t.Error("CookieSecure() = true, want false in development")
	}
}
```

Add `pkg/cais/session/cookie_config_test.go`:

```go
package session

import (
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestCookieOptionsFromConfig_productionSecure(t *testing.T) {
	opts := CookieOptionsFromConfig(cais.Config{Env: "production"})
	if !opts.Secure {
		t.Error("Secure = false, want true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/cais/... ./pkg/cais/session/... -v -run 'CookieSecure|CookieOptionsFromConfig'`

Expected: FAIL — methods not defined.

- [ ] **Step 3: Write minimal implementation**

In `pkg/cais/config.go`:

```go
func (c Config) CookieSecure() bool {
	return c.Env == "production"
}
```

In `pkg/cais/session/cookie.go`:

```go
import "github.com/puppe1990/cais/pkg/cais"

func CookieOptionsFromConfig(cfg cais.Config) CookieOptions {
	return CookieOptions{Secure: cfg.CookieSecure()}
}
```

In `internal/handlers/auth.go`, change `SignIn` call:

```go
if err := session.SignIn(w, h.sessions, user.ID, session.CookieOptionsFromConfig(h.cfg)); err != nil {
```

Add `cfg cais.Config` field to `AuthHandler` and update constructor + wiring in `internal/app/routes.go` (or wherever `NewAuthHandler` is called).

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/cais/... ./internal/handlers/... -v -run 'CookieSecure|CookieOptionsFromConfig|Auth'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/cais/config.go pkg/cais/config_test.go pkg/cais/session/cookie.go pkg/cais/session/cookie_config_test.go internal/handlers/auth.go internal/app/
git commit -m "feat(session): enable Secure cookies in production"
```

---

### Task 3: Security headers middleware

**Files:**

- Create: `pkg/cais/middleware/security.go`
- Create: `pkg/cais/middleware/security_test.go`
- Modify: `internal/app/app.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/cais/middleware/security_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestSecurityHeaders_production(t *testing.T) {
	cfg := cais.Config{Env: "production", AppURL: "https://app.example.com"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, key := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"Content-Security-Policy",
		"Strict-Transport-Security",
	} {
		if rr.Header().Get(key) == "" {
			t.Errorf("missing header %s", key)
		}
	}
}

func TestSecurityHeaders_development_noHSTS(t *testing.T) {
	cfg := cais.Config{Env: "development"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS should not be set in development")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/cais/middleware/... -v -run TestSecurityHeaders`

Expected: FAIL — `SecurityHeaders` not defined.

- [ ] **Step 3: Write minimal implementation**

Create `pkg/cais/middleware/security.go`:

```go
package middleware

import (
	"fmt"
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
)

func SecurityHeaders(cfg cais.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", fmt.Sprintf(
				"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'%s",
				"",
			))
			if cfg.Env == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

Wire in `internal/app/app.go` (after `Recover`, before routes):

```go
r.Use(middleware.SecurityHeaders(cfg))
```

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/cais/middleware/... ./internal/app/... -v -run 'SecurityHeaders|TestApp'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/cais/middleware/security.go pkg/cais/middleware/security_test.go internal/app/app.go
git commit -m "feat(middleware): add security headers"
```

---

### Task 4: Rate limiting for auth and public forms

**Files:**

- Create: `pkg/cais/middleware/ratelimit.go`
- Create: `pkg/cais/middleware/ratelimit_test.go`
- Modify: `internal/app/routes.go` (or route registration file)

- [ ] **Step 1: Write the failing test**

Create `pkg/cais/middleware/ratelimit_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimit_blocksAfterBurst(t *testing.T) {
	lim := NewRateLimiter(2) // 2 requests per window
	h := lim.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "203.0.113.1:1234"
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", rr.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/cais/middleware/... -v -run TestRateLimit`

Expected: FAIL — `NewRateLimiter` not defined.

- [ ] **Step 3: Write minimal implementation**

Create `pkg/cais/middleware/ratelimit.go`:

```go
package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string][]time.Time
}

func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  time.Minute,
		buckets: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r) + ":" + r.URL.Path
		if !rl.allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	times := rl.buckets[key]
	filtered := times[:0]
	for _, ts := range times {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	if len(filtered) >= rl.limit {
		rl.buckets[key] = filtered
		return false
	}
	rl.buckets[key] = append(filtered, now)
	return true
}
```

Apply in route registration:

```go
loginLimit := middleware.NewRateLimiter(10)
contactLimit := middleware.NewRateLimiter(20)

r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
r.Post("/contact", contactLimit.Middleware(http.HandlerFunc(contact.Post)).ServeHTTP)
```

Adjust to match existing router API (may need a small wrapper if `Post` only accepts `HandlerFunc`).

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/cais/middleware/... -v -run TestRateLimit`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/cais/middleware/ratelimit.go pkg/cais/middleware/ratelimit_test.go internal/app/
git commit -m "feat(middleware): add per-IP rate limiting"
```

---

### Task 5: Proxy-aware client IP + production config validation

**Files:**

- Modify: `pkg/cais/middleware/logger.go`
- Modify: `pkg/cais/middleware/logger_test.go`
- Modify: `pkg/cais/config.go`
- Modify: `pkg/cais/config_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `pkg/cais/middleware/logger_test.go`:

```go
func TestClientIP_usesXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")
	req.RemoteAddr = "127.0.0.1:9999"
	if got := clientIP(req); got != "203.0.113.50" {
		t.Errorf("clientIP = %q, want 203.0.113.50", got)
	}
}
```

Add to `pkg/cais/config_test.go`:

```go
func TestConfig_Validate_requiresAppURLInProduction(t *testing.T) {
	cfg := Config{Env: "production", AdminToken: "secret", AppURL: ""}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error without APP_URL in production")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/cais/middleware/... ./pkg/cais/... -v -run 'TestClientIP|requiresAppURL'`

Expected: FAIL

- [ ] **Step 3: Implement**

In `pkg/cais/middleware/logger.go`, replace `clientIP`:

```go
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	if host := r.RemoteAddr; host != "" {
		if i := strings.LastIndex(host, ":"); i >= 0 {
			return host[:i]
		}
		return host
	}
	return "127.0.0.1"
}
```

In `pkg/cais/config.go`, extend `Validate`:

```go
func (c Config) Validate() error {
	if c.Env == "production" {
		if c.AdminToken == "" {
			return fmt.Errorf("ADMIN_TOKEN is required when ENV=production")
		}
		if c.AppURL == "" {
			return fmt.Errorf("APP_URL is required when ENV=production")
		}
	}
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `make test`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/cais/middleware/logger.go pkg/cais/middleware/logger_test.go pkg/cais/config.go pkg/cais/config_test.go
git commit -m "feat: proxy-aware client IP and APP_URL validation"
```

---

## Phase 2 — DX Helpers

### Task 6: `httpx.RenderPageOrPartial`

**Files:**

- Modify: `pkg/cais/httpx/httpx.go`
- Modify: `pkg/cais/httpx/httpx_test.go`
- Modify: `internal/handlers/contact.go`

- [ ] **Step 1: Write the failing test**

Add to `pkg/cais/httpx/httpx_test.go`:

```go
func TestRenderPageOrPartial_htmxUsesPartial(t *testing.T) {
	renderer := testRenderer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/contact", nil)
	req.Header.Set("HX-Request", "true")

	RenderPageOrPartial(rr, req, renderer, RenderOptions{
		Layout:  "base",
		Page:    "home",
		Partial: "greeting",
		Data:    map[string]string{"Name": "Ada"},
	})

	if !strings.Contains(rr.Body.String(), "Ada") {
		t.Errorf("body = %q, want partial content", rr.Body.String())
	}
}

func TestRenderPageOrPartial_fullPageWhenNotHTMX(t *testing.T) {
	renderer := testRenderer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contact", nil)

	RenderPageOrPartial(rr, req, renderer, RenderOptions{
		Layout:  "base",
		Page:    "home",
		Partial: "greeting",
		Data:    map[string]string{"Name": "Ada"},
	})

	if !strings.Contains(rr.Body.String(), "Ada") {
		t.Errorf("body = %q, want page content", rr.Body.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/cais/httpx/... -v -run TestRenderPageOrPartial`

Expected: FAIL — `RenderOptions` / `RenderPageOrPartial` not defined.

- [ ] **Step 3: Write minimal implementation**

Add to `pkg/cais/httpx/httpx.go`:

```go
type RenderOptions struct {
	Layout  string
	Page    string
	Partial string
	Data    any
	Status  int
}

func RenderPageOrPartial(w http.ResponseWriter, r *http.Request, renderer *cais.Renderer, opts RenderOptions) {
	if opts.Status != 0 {
		w.WriteHeader(opts.Status)
	}
	if cais.IsHTMX(r) {
		if err := RenderPartial(w, renderer, opts.Partial, opts.Data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	RenderOrError(w, renderer, opts.Layout, opts.Page, opts.Data)
}
```

Refactor `internal/handlers/contact.go`:

```go
func (h *ContactHandler) renderContactResponse(w http.ResponseWriter, r *http.Request, status int, partial string, data any) {
	httpx.RenderPageOrPartial(w, r, h.renderer, httpx.RenderOptions{
		Layout:  "base",
		Page:    "contact",
		Partial: partial,
		Data:    data,
		Status:  status,
	})
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/cais/httpx/... ./internal/handlers/... -v -run 'RenderPageOrPartial|Contact'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/cais/httpx/httpx.go pkg/cais/httpx/httpx_test.go internal/handlers/contact.go
git commit -m "feat(httpx): add RenderPageOrPartial for HTMX handlers"
```

---

### Task 7: Field-level validation errors

**Files:**

- Create: `pkg/cais/validate/errors.go`
- Create: `pkg/cais/validate/errors_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/cais/validate/errors_test.go`:

```go
package validate

import "testing"

func TestFieldErrors_AddAndFirst(t *testing.T) {
	var errs FieldErrors
	errs.Add("email", "email is invalid")
	errs.Add("name", "name is required")

	if errs.First() != "email is invalid" {
		t.Errorf("First() = %q", errs.First())
	}
	if errs.Has("email") != true {
		t.Error("Has(email) = false, want true")
	}
	if len(errs) != 2 {
		t.Fatalf("len = %d, want 2", len(errs))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/cais/validate/... -v -run TestFieldErrors`

Expected: FAIL

- [ ] **Step 3: Implement**

Create `pkg/cais/validate/errors.go`:

```go
package validate

type FieldErrors map[string]string

func (e *FieldErrors) Add(field, msg string) {
	if *e == nil {
		*e = make(FieldErrors)
	}
	(*e)[field] = msg
}

func (e FieldErrors) Has(field string) bool {
	_, ok := e[field]
	return ok
}

func (e FieldErrors) First() string {
	for _, msg := range e {
		return msg
	}
	return ""
}

func (e FieldErrors) Any() bool {
	return len(e) > 0
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/cais/validate/... -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/cais/validate/errors.go pkg/cais/validate/errors_test.go
git commit -m "feat(validate): add FieldErrors map"
```

---

### Task 8: Flash messages (redirect feedback)

**Files:**

- Create: `pkg/cais/flash/flash.go`
- Create: `pkg/cais/flash/flash_test.go`
- Create: `pkg/cais/middleware/flash.go`
- Create: `pkg/cais/middleware/flash_test.go`
- Modify: `internal/app/app.go`
- Modify: `web/templates/layouts/base.html` (flash banner partial)

- [ ] **Step 1: Write the failing test**

Create `pkg/cais/flash/flash_test.go`:

```go
package flash

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetAndConsume(t *testing.T) {
	rr := httptest.NewRecorder()
	Set(rr, "notice", "Saved successfully")

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()
	cookie := res.Cookies()[0]

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	msg, ok := Consume(req)
	if !ok || msg != "Saved successfully" {
		t.Fatalf("Consume = (%q, %v), want (Saved successfully, true)", msg, ok)
	}

	// Second read should be empty (consume-once)
	msg2, ok2 := Consume(req)
	if ok2 || msg2 != "" {
		t.Fatalf("second Consume = (%q, %v), want (_, false)", msg2, ok2)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/cais/flash/... -v -run TestSetAndConsume`

Expected: FAIL

- [ ] **Step 3: Implement flash package**

Create `pkg/cais/flash/flash.go`:

```go
package flash

import (
	"encoding/json"
	"net/http"
)

const cookieName = "cais_flash"

func Set(w http.ResponseWriter, kind, message string) {
	payload, _ := json.Marshal(map[string]string{"kind": kind, "message": message})
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    string(payload),
		Path:     "/",
		MaxAge:   60,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

func Consume(r *http.Request) (string, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	var data map[string]string
	if err := json.Unmarshal([]byte(c.Value), &data); err != nil {
		return "", false
	}
	return data["message"], true
}

func Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1})
}
```

Create `pkg/cais/middleware/flash.go`:

```go
package middleware

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais/flash"
)

type flashKey struct{}

func Flash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if msg, ok := flash.Consume(r); ok {
			flash.Clear(w)
			r = r.WithContext(contextWithFlash(r, msg))
		}
		next.ServeHTTP(w, r)
	})
}

func FlashMessage(r *http.Request) (string, bool) {
	msg, ok := r.Context().Value(flashKey{}).(string)
	return msg, ok
}
```

Add private `contextWithFlash` using `context.WithValue`.

Wire `middleware.Flash` in `internal/app/app.go` after session load.

Add to layout a conditional banner when `.Flash` is set — pass from a small helper in page data or template func.

- [ ] **Step 4: Use flash in auth logout/login redirect**

In `internal/handlers/auth.go` after successful login:

```go
flash.Set(w, "notice", "Bem-vindo!")
httpx.SeeOther(w, r, "/dashboard")
```

- [ ] **Step 5: Run tests and commit**

Run: `make test`

```bash
git add pkg/cais/flash/ pkg/cais/middleware/flash.go pkg/cais/middleware/flash_test.go internal/app/ internal/handlers/auth.go web/templates/
git commit -m "feat(flash): one-shot flash messages on redirect"
```

---

## Phase 3 — CLI DB Tooling

### Task 9: `cais db rollback` (last migration)

**Files:**

- Modify: `pkg/cais/migrate/migrate.go`
- Modify: `pkg/cais/migrate/migrate_test.go`
- Modify: `internal/cli/db.go`
- Modify: `internal/cli/db_test.go`
- Modify: `internal/cli/cli.go` (help text)

- [ ] **Step 1: Write the failing test**

Add to `pkg/cais/migrate/migrate_test.go`:

```go
func TestRollbackLast_removesLastAppliedMigration(t *testing.T) {
	db := testDB(t)
	fs := fstest.MapFS{
		"001_a.sql": &fstest.MapFile{Data: []byte("CREATE TABLE t1 (id INTEGER);")},
		"002_b.sql": &fstest.MapFile{Data: []byte("CREATE TABLE t2 (id INTEGER);")},
	}
	if err := Apply(db, fs, "."); err != nil {
		t.Fatal(err)
	}
	version, err := RollbackLast(db, fs, ".")
	if err != nil {
		t.Fatal(err)
	}
	if version != "002_b" {
		t.Fatalf("version = %q, want 002_b", version)
	}
	entries, err := Status(db, fs, ".")
	if err != nil {
		t.Fatal(err)
	}
	if entries[1].Applied {
		t.Fatal("002_b should be pending after rollback")
	}
}
```

**Note:** Rollback only removes the `schema_migrations` row — SQL down migrations are out of scope (YAGNI). Document this limitation in CLI help.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/cais/migrate/... -v -run TestRollbackLast`

Expected: FAIL

- [ ] **Step 3: Implement `RollbackLast`**

```go
func RollbackLast(db *sql.DB, migrations fs.FS, dir string) (string, error) {
	entries, err := Status(db, migrations, dir)
	if err != nil {
		return "", err
	}
	var lastApplied string
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Applied {
			lastApplied = entries[i].Version
			break
		}
	}
	if lastApplied == "" {
		return "", fmt.Errorf("no applied migrations to rollback")
	}
	if _, err := db.Exec("DELETE FROM schema_migrations WHERE version = ?", lastApplied); err != nil {
		return "", fmt.Errorf("rollback record %s: %w", lastApplied, err)
	}
	return lastApplied, nil
}
```

Add CLI command in `internal/cli/db.go`:

```go
case "rollback":
	return c.cmdDBRollback()
```

- [ ] **Step 4: Run tests and commit**

Run: `go test ./pkg/cais/migrate/... ./internal/cli/... -v -run 'Rollback|cmdDB'`

```bash
git add pkg/cais/migrate/ internal/cli/db.go internal/cli/db_test.go internal/cli/cli.go
git commit -m "feat(cli): add cais db rollback"
```

---

### Task 10: `cais db prune-sessions`

**Files:**

- Modify: `internal/cli/db.go`
- Modify: `internal/cli/db_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/cli/db_test.go`:

```go
func TestCLI_DBPruneSessions(t *testing.T) {
	// Use temp dir with :memory: or temp db, run prune-sessions, assert output contains "pruned"
}
```

- [ ] **Step 2–5: Implement, test, commit**

Wire `session.NewSQLiteStore(db).PruneExpired()` in new `cmdDBPruneSessions`.

```bash
git commit -m "feat(cli): add cais db prune-sessions"
```

---

## Phase 4 — Generator & Docs Sync

### Task 11: Update scaffolds to use new helpers

**Files:**

- Modify: `internal/cli/templates.go`
- Modify: `internal/cli/resource_gen.go`
- Modify: `internal/cli/scaffold_auth.go`
- Modify: `AGENTS.md`
- Modify: `README.md`

- [ ] **Step 1: Update generated handler template to use `httpx.RenderPageOrPartial`**
- [ ] **Step 2: Update auth scaffold for `CookieOptionsFromConfig` and flash**
- [ ] **Step 3: Update `cais g auth` tests in `internal/cli/scaffold_auth_test.go`**
- [ ] **Step 4: Document new env vars and commands in AGENTS.md**
- [ ] **Step 5: Run `make ci` and commit**

```bash
git commit -m "docs: sync scaffolds and AGENTS.md with framework improvements"
```

---

## Verification Checklist (before merge)

- [ ] `make test` — all packages green with `-race`
- [ ] `make lint` — no new golangci-lint issues
- [ ] `make ci` — full pipeline passes
- [ ] Manual smoke: `make dev` → login → contact form HTMX → logout flash
- [ ] Production smoke: `ENV=production ADMIN_TOKEN=x APP_URL=https://example.com make build && ./bin/server` boots without error

---

## Self-Review

| Requirement                   | Task    |
| ----------------------------- | ------- |
| Session expiry                | Task 1  |
| Secure cookies                | Task 2  |
| Security headers              | Task 3  |
| Rate limiting                 | Task 4  |
| Proxy IP + APP_URL validation | Task 5  |
| HTMX render helper            | Task 6  |
| Field validation map          | Task 7  |
| Flash messages                | Task 8  |
| DB rollback                   | Task 9  |
| Session prune CLI             | Task 10 |
| Scaffold sync                 | Task 11 |

**Gaps deferred to follow-up plans:**

- Resource pagination/search (`2026-07-01-cais-resource-pagination.md`)
- Structured JSON production logging
- `cais routes` command
- Password reset / user registration

---

## Suggested PR Stack

1. `feat/session-expiry` — Task 1–2
2. `feat/security-middleware` — Task 3–5
3. `feat/httpx-and-validate` — Task 6–7
4. `feat/flash-messages` — Task 8
5. `feat/cli-db-commands` — Task 9–10
6. `chore/scaffold-sync` — Task 11

Each PR should pass `make ci` independently where possible.
