# Cais Framework Roadmap — Full Audit Remediation

**Date:** 2026-07-01
**Status:** Draft — awaiting review
**Scope:** Option C — complete remediation of audit findings
**Prerequisite:** `docs/superpowers/plans/2026-07-01-cais-framework-improvements.md` (Phases 1–4 largely shipped)

---

## Goal

Close every gap identified in the 2026-07-01 framework audit so that:

1. Generated apps work in production without manual fixes
2. Security defaults are correct behind reverse proxies
3. CLI generators stay in sync with the reference app
4. Rails-like DX gaps (seed, routes, pagination, forms) are addressed
5. Documentation matches implementation

---

## Execution Approaches (3 options)

### Approach A — Sequential PR stack (recommended)

Six independent PRs, each passes `make ci` alone. Order by risk/impact:

| PR  | Theme                      | Ships                                                    |
| --- | -------------------------- | -------------------------------------------------------- |
| 1   | Admin auth + scaffold sync | Production-unblocking                                    |
| 2   | Security hardening         | Trusted proxy, error sanitization, cookie hygiene        |
| 3   | Generator robustness       | AST patches, migrations, `--module`, `--dry-run`         |
| 4   | Rails parity (data)        | `db seed`, migration down, `cais g model`                |
| 5   | Rails parity (UI)          | Pagination, form helpers, `cais routes`                  |
| 6   | Docs + test coverage       | README, AGENTS.md, `.env.example`, low-coverage packages |

**Pros:** Reviewable increments, rollback per PR, matches existing Cais PR-stack culture.
**Cons:** Full roadmap takes multiple sessions.

### Approach B — Two mega-phases

Phase 1 = PRs 1–3 (correctness + security). Phase 2 = PRs 4–6 (features + docs).

**Pros:** Faster perceived progress.
**Cons:** Larger diffs, harder review, more merge conflict risk.

### Approach C — Feature flags per generator

Ship framework packages first; gate generator output behind `--next` flag until stable.

**Pros:** Zero breakage for existing scaffold users.
**Cons:** Maintenance burden of dual templates; YAGNI for a pre-1.0 framework.

**Recommendation:** Approach A. Cais already documents PR stacks; each phase is independently testable.

---

## Architecture Overview

```
pkg/cais/
  config.go          + TrustedProxies, SanitizeErrors
  httpx/             + RenderOrError production-safe
  middleware/
    auth.go          + AdminSessionAuth (dual mode)
    clientip.go      + trusted proxy parsing (extract from logger)
    ratelimit.go     + bucket cleanup
  forms/             NEW — template helpers (csrfField, field error)
  pagination/        NEW — offset/limit helpers
  cache/             NEW — in-memory TTL cache (stdlib)
  migrate/           + Down SQL support (*.down.sql or -- down section)

internal/cli/
  resource.go        RequireAuth for admin HTML (default)
  templates.go       sync all scaffolds with reference app
  routes_cmd.go      NEW — cais routes
  seed.go            NEW — cais db seed
  doctor.go          production readiness checks
  patch/             NEW — go/ast route patching (optional refactor)

internal/handlers/   reference implementations
internal/app/        route registry for `cais routes`
```

No new runtime dependencies. Stdlib + existing stack only.

---

## Phase 1 — Admin Auth & Scaffold Sync (Critical)

### Problem

`cais g resource` wraps admin CRUD in `middleware.AdminAuth(cfg)` (Bearer-only). HTML forms cannot authenticate in production when `ADMIN_TOKEN` is set.

### Design

**Default:** Generated admin routes use session auth (same as dashboard):

```go
r.Group(func(g *cais.Router) {
  g.Use(middleware.RequireAuth("/login"))
  g.Get("/admin/items", admin.Index)
  // ...
})
```

**Flag for API-style admin:** `cais g resource bookmark --admin-auth bearer` keeps current `AdminAuth` behavior.

**Reference app:** Document in README that `AdminAuth` is for API/scripts; browser admin uses `RequireAuth`.

### Scaffold sync checklist

| Template                 | Fix                                                                                                    |
| ------------------------ | ------------------------------------------------------------------------------------------------------ |
| `tplContactHandler`      | Add `FieldErrors` name validation (match `internal/handlers/contact.go`)                               |
| `tplContactTest`         | Assert name-required rejection                                                                         |
| `tplAppBlank`            | Add `Recover`, `SecurityHeaders`, `LoadSession`, `Flash`, `Static`, server timeouts, DB health handler |
| `tplMigration002Auth`    | Add `expires_at` to sessions table                                                                     |
| `tplGenericHandler`      | Use `httpx.RenderOrError` + `meta.ForRequest` + i18n catalog                                           |
| Admin resource templates | Use `RequireAuth` group; pass `meta.ForRequest` consistently                                           |

### Auth API clarity

Add `docs/packages/auth.md` (or README section) matrix:

| Middleware              | Use case                                  |
| ----------------------- | ----------------------------------------- |
| `RequireAuth("/login")` | Browser pages (dashboard, admin CRUD)     |
| `AdminAuth(cfg)`        | API/scripts with Bearer token             |
| `Protect(cfg, h)`       | Single handler + Bearer (deprecated path) |

Remove `TokenAuth` in v0.2; mark deprecated in godoc now.

### Tests

- `internal/cli/cli_test.go`: resource scaffold emits `RequireAuth` by default
- `internal/cli/cli_test.go`: `--admin-auth bearer` emits `AdminAuth`
- `scaffold_test.go`: blank app includes security middleware
- `scaffold_test.go`: contact template includes name validation

---

## Phase 2 — Security Hardening

### 2.1 Trusted proxy configuration

**Config:**

```go
// TRUSTED_PROXIES=127.0.0.1,10.0.0.0/8  (comma-separated IPs/CIDRs)
// Empty = never trust X-Forwarded-For / X-Real-IP (use RemoteAddr only)
TrustedProxies []string
```

**New package function** `middleware.ClientIP(r, cfg)`:

- If `RemoteAddr` is in trusted set → parse `X-Forwarded-For` (first hop)
- Otherwise → `RemoteAddr` only

Extract from `logger.go`; update `ratelimit.go` to use shared helper.

### 2.2 Production error sanitization

**Config:** `SanitizeErrors() bool` → true when `ENV=production`.

**httpx changes:**

```go
func RenderOrError(w, renderer, layout, page, data, cfg) {
  if err := RenderPage(...); err != nil {
    log.Printf("render error: %v", err)
    if cfg.SanitizeErrors() {
      http.Error(w, "internal server error", 500)
    } else {
      http.Error(w, err.Error(), 500)
    }
  }
}
```

Same pattern for handlers that expose `err.Error()` on 500 (contact, auth, generated handlers).

### 2.3 Cookie hygiene

`session.ClearCookie` must pass `Secure` from `CookieOptions` when clearing.

### 2.4 Rate limiter memory

Add periodic cleanup in `RateLimiter`: drop stale bucket keys on each `allow()` when `len(buckets) > 1000` or entries older than 2× window.

### 2.5 Dev seed warning

When `ENV=development` and seed user exists, `boot.Print` shows:

```
⚠ Demo user: demo@example.com / password
```

`cais doctor` adds optional warn check for default seed in non-dev `ENV`.

### 2.6 CSP (document, not block)

Keep `'unsafe-inline'` for HTMX compatibility. Document tradeoff in README; nonce-based CSP deferred.

### Tests

- `clientip_test.go`: trusted vs untrusted proxy
- `httpx_test.go`: sanitized 500 in production
- `ratelimit_test.go`: bucket cleanup
- `session/cookie_test.go`: clear with Secure

---

## Phase 3 — Generator Robustness

### 3.1 Route patching

Replace `strings.LastIndex(content, "\n}\n")` with anchor-based patch:

```go
const routesRegisterAnchor = "func registerRoutes("
```

Find closing brace of `registerRoutes` via brace counting or `go/parser`. Fallback to current behavior with clear error.

### 3.2 Layout nav patching

Replace brittle string replace with HTML comment marker in generated `base.html`:

```html
<!-- cais:nav -->
```

Generator inserts nav links after marker.

### 3.3 Migration stubs

`cais g migration create_items` generates:

```sql
-- migration: create_items
-- up
CREATE TABLE items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- down
DROP TABLE IF EXISTS items;
```

### 3.4 `cais new --module`

```bash
cais new myapp --module github.com/acme/myapp
```

Overrides `github.com/puppe1990/{slug}` default.

### 3.5 `cais g --dry-run`

Print files that would be created/patched without writing.

### 3.6 `gofmtGoFiles` scope

Extend to `pkg/` if generated, or document limitation in CLI help.

### Tests

- Route patch on edited `routes.go` (with extra functions below)
- Migration down section parsed by `migrate` package
- `--module` flag in `cais new`
- `--dry-run` produces no filesystem changes

---

## Phase 4 — Rails Parity (Data Layer)

### 4.1 Migration down SQL

Extend `pkg/cais/migrate`:

- Parse `-- down` section from `.sql` files (or sibling `NNN_name.down.sql`)
- `cais db rollback` executes down SQL then removes `schema_migrations` row
- Transaction-wrapped; failure rolls back record deletion

If no down section: current behavior (record-only rollback) with CLI warning.

### 4.2 `cais db seed`

```bash
cais db seed           # run internal/db/seeds.go
cais db seed --list    # show available seed functions
```

Scaffold includes `internal/db/seeds.go` with `SeedDemoUser` (if auth) and per-resource `SeedDemo*` calls.

Generator `cais g resource` appends seed function unless `--no-seed`.

### 4.3 `cais g model`

```bash
cais g model bookmark --fields title:string,url:url
```

Generates: model struct, migration, store methods (no handlers/templates). Lighter than full resource.

### 4.4 Router `Put` / `Patch`

```go
func (r *Router) Put(pattern string, handler http.HandlerFunc)
func (r *Router) Patch(pattern string, handler http.HandlerFunc)
```

Mirror `Post`/`Delete` registration.

### Tests

- `migrate_test.go`: down SQL executed on rollback
- `cli/db_test.go`: seed command output
- `cli_test.go`: model generator creates expected files
- `router_test.go`: Put/Patch routing

---

## Phase 5 — Rails Parity (UI & DX)

### 5.1 Resource pagination

**Store:**

```go
ListBookmarks(page, perPage int) ([]models.Bookmark, int, error)
// returns items, total count, error
```

**Generator flag:** `cais g resource bookmark --paginate` (default perPage=25).

**Template:** HTMX-friendly page controls partial `admin_{resource}_pagination`.

### 5.2 Form helpers (`pkg/cais/forms`)

Template funcs registered via renderer:

```html
{{ csrfField .CSRFToken }} {{ fieldError .Errors "email" }}
```

`forms.FieldData` struct for templates. Generator uses in admin forms.

### 5.3 `cais routes`

```bash
cais routes
# GET  /
# GET  /contact
# POST /contact
# ...
```

Parse `internal/app/routes.go` with `go/parser` + regex for `cais.IntParam`/`StringParam` patterns. No runtime registry required for v1.

### 5.4 In-memory cache (`pkg/cais/cache`)

```go
cache.New[string, T](ttl time.Duration)
c.Get(key) (T, bool)
c.Set(key, val)
```

For future fragment caching; ship minimal API with tests. Not wired into render by default (YAGNI).

### Tests

- Pagination store method + handler test
- Form helper template execution test
- `cais routes` CLI test against fixture `routes.go`

---

## Phase 6 — Documentation & Test Coverage

### 6.1 README updates

- Admin auth matrix (Bearer vs session)
- Lightsail/reverse-proxy deploy guide (`TRUSTED_PROXIES`, `APP_URL`)
- Align Go version: **1.22+** (minimum) / `go.mod` may pin newer
- New commands: `db seed`, `routes`, `g model`
- `validate.FieldErrors` example in Framework APIs
- i18n user guide (link to i18n spec)

### 6.2 `.env.example` (scaffold)

```env
PORT=:8080
DB_PATH=./data/app.db
ENV=development
APP_URL=
ADMIN_TOKEN=
LOCALE=en
TRUSTED_PROXIES=
```

Document `CAIS_REPLACE`, `CAIS_SKIP_TIDY` in README env table.

### 6.3 AGENTS.md

- Admin auth guidance for generators
- `forms` helpers
- Pagination in resources
- Migration down sections

### 6.4 Test coverage targets

| Package                      | Current | Target                            |
| ---------------------------- | ------- | --------------------------------- |
| `pkg/cais/console`           | ~48%    | 70%+                              |
| `pkg/cais/devlog` (local.go) | partial | `IsLoopback` covered              |
| `pkg/cais/boot/version.go`   | 0%      | basic version paths               |
| `pkg/cais/pwa`               | ~69%    | no panic in `FS()` — return error |
| `internal/store`             | ~39%    | 60%+ (pagination, seed)           |

### 6.5 `pwa.FS()` fix

Change `panic(err)` → `return nil, fmt.Errorf(...)`. Update callers.

---

## Out of Scope (explicit)

| Item                               | Reason                                |
| ---------------------------------- | ------------------------------------- |
| Action Mailer / SMTP               | Separate spec; large surface          |
| Background jobs / queue            | Separate spec                         |
| Nonce-based CSP                    | HTMX inline script conflict           |
| Accept-Language i18n               | Covered in i18n spec out-of-scope     |
| esbuild / JS bundling              | Tailwind + vendored HTMX sufficient   |
| Password reset / registration      | Auth expansion spec                   |
| Structured JSON production logging | Separate observability spec           |
| External docs site                 | README + AGENTS.md sufficient for now |

---

## Error Handling Conventions

| Context               | Production            | Development   |
| --------------------- | --------------------- | ------------- |
| Render failure        | 500 generic message   | `err.Error()` |
| DB failure in handler | 500 generic           | `err.Error()` |
| Validation            | 422 with field errors | same          |
| Auth failure          | 401/303 redirect      | same          |

Log full error server-side always.

---

## Verification (per phase)

- `go test ./... -race -count=1`
- `make lint`
- `make ci`
- Generator smoke: `cais new testapp && cd testapp && cais g resource item --fields name:string && make test`
- Production smoke: `ENV=production ADMIN_TOKEN=x APP_URL=https://example.com TRUSTED_PROXIES=127.0.0.1 ./bin/server`

---

## PR Stack Summary

1. `feat/admin-auth-scaffold-sync` — Phase 1
2. `feat/security-hardening` — Phase 2
3. `feat/generator-robustness` — Phase 3
4. `feat/rails-data-parity` — Phase 4
5. `feat/rails-ui-parity` — Phase 5
6. `docs/coverage-pwa-fix` — Phase 6

Each PR must pass `make ci` independently.

---

## Self-Review

- [x] No TBD/TODO placeholders
- [x] Admin auth contradiction resolved (session default, bearer opt-in)
- [x] Scoped to single roadmap; mailer/jobs explicitly deferred
- [x] Migration down format specified (`-- down` section)
- [x] Trusted proxy behavior explicit (empty = no trust)
- [x] Aligns with existing `2026-07-01-cais-framework-improvements.md` without duplicating shipped work
