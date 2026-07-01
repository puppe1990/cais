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

Set `ADMIN_TOKEN` in production (`cfg.Validate()` fails on boot if missing). Use `middleware.AdminAuth(cfg)` on admin route groups — Bearer header only, no query params. No-op in development when unset.

## Session auth

`cais new` includes login/logout and protects `/dashboard`. Add to existing apps with `cais g auth`.

```go
r.Use(middleware.LoadSession(deps.Store.Sessions()))
r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
session.SignIn(w, sessions, userID, session.CookieOptions{})
```

Dev seed user: `demo@example.com` / `password`. Sessions persist in SQLite via `session.NewSQLiteStore`.

## New page

1. Test in `internal/handlers/foo_test.go`
2. Template in `web/templates/pages/foo.html`
3. Handler in `internal/handlers/foo.go` — embed `meta.Site` in page data
4. Register the route in `internal/app/app.go`

Pass `meta.SiteFrom(appName, cfg.AppURL)` from bootstrap so layouts render correct OG/Twitter tags (`absURL` template func).

## CSRF

- `middleware.CSRF` on the router (validates POST/PUT/DELETE/PATCH)
- Pass `meta.WithCSRF(site, r)` in page data — layout renders `<meta name="csrf-token">` + HTMX header script
- HTML forms: `<input type="hidden" name="csrf_token" value="{{ .CSRFToken }}" />`
- Integration tests: GET page first (cookie), then POST with matching token

## HTMX interactions

- Partial in `web/templates/partials/`
- Template attributes: `hx-post`, `hx-target`, `hx-swap`
- Handler: if `cais.IsHTMX(r)` → `RenderPartial`, else → `Render` with layout
- Test with `req.Header.Set("HX-Request", "true")`

## New table

1. Store test with `":memory:"` before the migration
2. SQL in `internal/store/migrations/NNN_name.sql`
3. Methods on the `store.Store` interface
4. Wrap DB with `sqllog.Wrap` in `NewSQLiteStore` for development query logs
5. Migrations tracked in `schema_migrations` via `pkg/cais/migrate` (idempotent on boot)

## Development logging

In `ENV=development`:

- `middleware.LoggerTo(devlog.MirrorDefault(...))` — timestamped request logs
- `sqllog.Wrap(db, sqllog.Config{Enabled: true})` — SQL query + duration logs
- `devlog.Register(r, cfg.Env, buf)` — mounts `/logs` (localhost only, HTMX refresh)

Boot banner via `boot.Print` in `cmd/server/main.go`. Port auto-pick via `cais.ResolvePort` when preferred port is busy.

Set `APP_URL` for absolute OG image URLs in production.

## CLI generators

```bash
cais new myapp --minimal
cais new myapp --blank
cais g resource bookmark --fields title:string,url:url,notes:text? --public
cais doctor                    # verify htmx, air, go.mod
cais g handler settings
cais g console              # scaffold cmd/console/main.go
cais g auth                 # login/logout + protected dashboard
```

Field types: `string`, `text`, `url`, `bool`, `int`, `date`. Suffix `?` for optional.

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
cais db migrate  # run pending migrations
cais db status   # list applied/pending migrations
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
