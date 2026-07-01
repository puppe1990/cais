# Cais — AI Conventions

## Rule #1: TDD is mandatory

Before writing production code:

1. Write the test in `*_test.go`
2. Run: `go test ./... -v -run TestName`
3. Confirm it **fails** for the right reason (missing feature, not a typo)
4. Write the **minimal** code to make it pass
5. Run: `make test`
6. Only then refactor

## Structure

| Directory            | Responsibility                                          |
| -------------------- | ------------------------------------------------------- |
| `pkg/cais/`          | Framework: config, router, render, htmx, middleware     |
| `pkg/cais/httpx/`    | Render and redirect helpers for handlers                |
| `pkg/cais/meta/`     | Open Graph / Twitter preview (`Site`, `PreviewHTML`)    |
| `pkg/cais/session/`  | Cookie sessions (`SignIn`, `SignOut`, `Store`)          |
| `pkg/cais/boot/`     | Rails-style startup banner                              |
| `pkg/cais/devlog/`   | Development log buffer + `/logs` viewer                 |
| `pkg/cais/sqllog/`   | SQL query logging wrapper (`Wrap`, `EnabledForEnv`)     |
| `pkg/cais/console/`  | Interactive REPL (yaegi + SQL)                          |
| `pkg/cais/csrf/`     | CSRF tokens (double-submit cookie)                      |
| `pkg/cais/validate/` | Form field validation helpers                           |
| `pkg/cais/forms/`    | Template helpers (`csrfField`, `fieldError`)            |
| `pkg/cais/i18n/`     | Locale catalogs (`LOCALE` env, `t` template func)       |
| `pkg/cais/testutil/` | Test helpers (`NewRenderer`, `NewRequest`, path values) |
| `pkg/cais/pwa/`      | Default PWA assets generator (manifest, icons, og.png)  |
| `internal/app/`      | Bootstrap: route and dependency wiring                  |
| `internal/handlers/` | HTTP handlers                                           |
| `internal/store/`    | SQLite persistence                                      |
| `web/templates/`     | HTML templates (layouts, pages, partials)               |
| `web/static/`        | Tailwind CSS, HTMX, PWA (manifest, sw.js, icons)        |
| `cmd/server/`        | Entry point                                             |

## Router path params and groups

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Group(middleware.Protect, func(g *cais.Router) {
  g.Post("/admin/items/{id}", cais.IntParam("id", admin.Update))
})
```

## Admin protection

| Mode                    | Middleware                         | Generator flag            |
| ----------------------- | ---------------------------------- | ------------------------- |
| Browser admin (default) | `middleware.RequireAuth("/login")` | `cais g resource` default |
| Bearer token API        | `middleware.AdminAuth(cfg)`        | `--admin-auth bearer`     |

Set `ADMIN_TOKEN` in production (`cfg.Validate()` fails on boot if missing). `AdminAuth` accepts Bearer header only, no query params. No-op in development when unset.

`cais g resource` defaults to session auth (`--admin-auth session`). Use `--admin-auth bearer` for token-only admin APIs without login pages.

## Session auth

`cais new` includes login/logout and protects `/dashboard`. Add to existing apps with `cais g auth`.

```go
r.Use(middleware.LoadSession(deps.Store.Sessions()))
r.Use(middleware.Flash)
r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
auth := handlers.NewAuthHandler(renderer, store, site, store.Sessions(), cfg)
session.SignIn(w, sessions, r, userID, session.CookieOptionsFromConfig(cfg))
flash.Set(w, "notice", "Bem-vindo!", cfg.CookieSecure())
```

Dev seed user: `demo@example.com` / `password`. Sessions persist in SQLite via `session.NewSQLiteStore`.

**Session expiry** — cookies and DB rows expire after 7 days (`sessionTTL` / `defaultMaxAge`). SQLite stores `expires_at`; expired rows are ignored on lookup. Prune stale rows with `cais db prune-sessions` (or call `session.Store.PruneExpired()`).

**Production cookies** — `session.CookieOptionsFromConfig(cfg)` sets `Secure` when `cfg.CookieSecure()` is true (`ENV=production`).

## New page

1. Test in `internal/handlers/foo_test.go`
2. Template in `web/templates/pages/foo.html`
3. Handler in `internal/handlers/foo.go` — embed `meta.Site` in page data
4. Register the route in `internal/app/app.go`

Pass `meta.SiteFrom(appName, cfg.AppURL)` from bootstrap so layouts render correct OG/Twitter tags (`absURL` template func).

## CSRF

- `middleware.CSRF(cfg)` on the router (validates POST/PUT/DELETE/PATCH)
- Pass `meta.ForRequest(site, r)` in page data (CSRF + flash) — layout renders `<meta name="csrf-token">` + HTMX header script
- HTML forms: `{{ csrfField .CSRFToken }}` (`pkg/cais/forms`, registered on the renderer) or `<input type="hidden" name="csrf_token" value="{{ .CSRFToken }}" />`
- Field errors in templates: `{{ fieldError .Errors "email" }}`
- Integration tests: GET page first (cookie), then POST with matching token

## Flash messages

- `middleware.Flash` on the router (after `LoadSession`)
- Set on redirect: `flash.Set(w, "notice", "Saved!", cfg.CookieSecure())` — read in templates via `meta.ForRequest(site, r)` → `.Flash`
- One-shot: consumed on the next request

## Security headers

- `middleware.SecurityHeaders(cfg)` on the router (after `Recover`)
- Sets `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`
- Adds `Strict-Transport-Security` in production (`ENV=production`)
- CSRF and flash cookies use `Secure` when `cfg.CookieSecure()` is true
- Session rotates on login (invalidates previous token)

## Rate limiting

Wrap sensitive POST routes with per-IP token buckets:

```go
loginLimit := middleware.NewRateLimiter(10)   // 10 req/min
contactLimit := middleware.NewRateLimiter(20) // 20 req/min
r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
r.Post("/contact", contactLimit.Middleware(http.HandlerFunc(contact.Post)).ServeHTTP)
```

Rate limiters use `middleware.ClientIP(r, cfg)` — set `TRUSTED_PROXIES` when behind a reverse proxy.

## HTMX interactions

- Partial in `web/templates/partials/` — each file has `{{ define "name" }}` matching the filename
- Partials are parsed into full pages too, so `{{ template "name" . }}` works in pages and layouts
- For HTMX swaps, handler returns the partial via `RenderPartial` (not a full layout)
- Template attributes: `hx-post`, `hx-target`, `hx-swap`
- Handler: use `httpx.RenderPageOrPartial` for forms that return partial on HTMX and full page otherwise
- Test with `req.Header.Set("HX-Request", "true")`

```go
httpx.RenderPageOrPartial(w, r, renderer, httpx.RenderOptions{
  Layout: "base", Page: "contact", Partial: "contact_errors", Data: data, Status: 422,
})
```

## Form validation

Use `validate.Email`, `validate.URL`, `validate.Required` for single-field checks. For multiple fields, collect errors in `validate.FieldErrors`:

```go
var errs validate.FieldErrors
if item.Name == "" {
  errs.Add("name", "Name is required")
}
if errs.Any() {
  // re-render form with errs map — templates use {{ fieldError .Errors "name" }}
}
```

Pass `errs` as `.Errors` in page data when re-rendering forms.

## HTMX UX (app-like feel)

Layout loads `cais.js` after `htmx.min.js` — CSRF header, focus restore, optimistic toggles.

- **Small targets** — swap `#form-errors` or `this`, not whole lists
- **Transitions** — `hx-swap="innerHTML swap:150ms"` on forms; `outerHTML swap:150ms` on toggles
- **Forms** — `hx-indicator` + `hx-disabled-elt="button[type='submit']"`; hide label with `.htmx-request-hide`
- **Bool toggles** — `data-cais-optimistic="toggle"` for instant class flip (see resource generator)
- **Optional** — `data-cais-view-transition` enables View Transitions API when supported
- **CSS** — `input.css` includes `.htmx-swapping` / `.htmx-settling` fade utilities
- **Response headers** — `cais.SetTrigger(w, "event")`, `cais.SetRetarget(w, "#id")` when needed

## New table

1. Store test with `":memory:"` before the migration
2. SQL in `internal/store/migrations/NNN_name.sql`
3. Methods on the `store.Store` interface
4. Wrap DB with `sqllog.Wrap` in `NewSQLiteStore` for development query logs
5. Migrations tracked in `schema_migrations` via `pkg/cais/migrate` (idempotent on boot)

**Migration down sections** — use `-- up` / `-- down` markers in `.sql` files (generator default for resources):

```sql
-- up
CREATE TABLE bookmarks (...);

-- down
DROP TABLE IF EXISTS bookmarks;
```

`cais db rollback` executes the `-- down` SQL when present, then removes the `schema_migrations` row. Without a down section, only the record is removed.

## Development logging

In `ENV=development`:

- `middleware.LoggerTo(devlog.MirrorDefault(...))` — timestamped request logs
- `sqllog.Wrap(db, sqllog.Config{Enabled: true})` — SQL query + duration logs
- `devlog.Register(r, cfg.Env, buf)` — mounts `/logs` (localhost only, HTMX refresh)

Boot banner via `boot.Print` in `cmd/server/main.go`. Port auto-pick via `cais.ResolvePort` when preferred port is busy.

Set `APP_URL` for absolute OG image URLs. **`APP_URL` is required when `ENV=production`** — `cfg.Validate()` fails on boot if missing.

Set `TRUSTED_PROXIES` (comma-separated IPs) when behind a reverse proxy so `middleware.ClientIP` trusts `X-Forwarded-For` for rate limiting and logging.

Set `LOCALE=en` (default) or `LOCALE=pt` for UI strings via `pkg/cais/i18n`. See [i18n design](docs/superpowers/specs/2026-07-01-i18n-design.md).

## CLI generators

```bash
cais new myapp              # includes GitHub Actions CI, pre-commit, golangci-lint, Prettier
cais new myapp --minimal
cais new myapp --blank
cais new myapp --module github.com/acme/myapp
cais g [--dry-run] resource bookmark --fields title:string,url:url,notes:text? --public --paginate
cais g [--dry-run] model bookmark --fields title:string,url:url
cais g [--dry-run] handler settings
cais g [--dry-run] page about
cais g [--dry-run] migration add_tags
cais g [--dry-run] auth       # login/logout + protected dashboard
cais g console                # scaffold cmd/console/main.go
cais g ci                     # add CI/pre-commit to existing apps
cais doctor                   # verify htmx, air, go.mod
cais routes                   # list routes from internal/app/routes.go
```

Field types: `string`, `text`, `url`, `bool`, `int`, `date`. Suffix `?` for optional.

**Resource options:** `--public` (public list page), `--paginate` (admin index pagination, 25/page), `--no-seed` (skip demo data), `--admin-auth session|bearer` (default: session).

**Model generator** — `cais g model` creates model struct, migration, and store methods only (no handlers, templates, or routes). Use for data layer without admin CRUD.

**Dry-run** — `cais g --dry-run ...` prints files that would be created without writing them.

**Demo seed** — `cais g resource` (unless `--no-seed`) generates `SeedDemo*` store methods and wires them into `cmd/server/main.go` at boot.

## App commands (run from a Cais app)

```bash
cais install  # npm install + go mod tidy
cais css      # build Tailwind
cais dev      # hot reload + tailwind watch
cais build    # bin/server
cais server   # go run ./cmd/server
cais test     # go test ./...
cais doctor   # verify htmx, air, go.mod
cais console  # Rails-style REPL (store, cfg, db + sql)
cais routes   # list HTTP routes from internal/app/routes.go
cais db migrate        # run pending migrations
cais db status         # list applied/pending migrations
cais db rollback       # roll back last migration (runs -- down SQL when present)
cais db prune-sessions # delete expired login sessions from SQLite
```

Console bindings: `store`, `cfg`, `db`, plus any custom keys in `Bindings`. Commands: `help`, `sql`, `reload`, `history`, `!N`/`!!`, `exit`. Arrow keys when stdin is a TTY.

`/logs` — development-only log viewer (localhost). Shows request + SQL logs with HTMX auto-refresh.

## Framework commands (Cais repo)

```bash
make test-v   # TDD: watch RED/GREEN
make test     # validation with -race
make lint     # golangci-lint
make format   # prettier --write
make ci       # test + lint + format-check
make dev      # hot reload + tailwind watch
make build    # bin/cais
make pwa      # regenerate PWA assets (manifest, icons, og.png)
make docker   # ~15-20MB image
```

Pre-commit (tests, lint, prettier): `make pre-commit-install` once, then hooks run on every commit.

## Do not

- Parse templates per request (use `cais.NewRenderer`)
- Use inline CSS (use Tailwind classes in templates)
- Mock the database (use SQLite `:memory:`)
- Import `internal/` from `pkg/cais/` (avoids import cycles)
