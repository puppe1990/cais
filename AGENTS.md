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
| `pkg/cais/testutil/` | Test helpers (`NewRenderer`, `NewRequest`, path values) |
| `internal/app/`      | Bootstrap: route and dependency wiring                  |
| `internal/handlers/` | HTTP handlers                                           |
| `internal/store/`    | SQLite persistence                                      |
| `web/templates/`     | HTML templates (layouts, pages, partials)               |
| `web/static/`        | Tailwind CSS, HTMX, PWA (manifest, sw.js, icons)        |
| `pkg/cais/pwa/`      | Default PWA assets generator                            |
| `cmd/server/`        | Entry point                                             |

## Router path params and groups

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Group(middleware.Protect, func(g *cais.Router) {
  g.Post("/admin/items/{id}", cais.IntParam("id", admin.Update))
})
```

## Admin protection

Set `ADMIN_TOKEN` in production. Use `middleware.Protect` on admin routes — no-op when env is empty.

## New page

1. Test in `internal/handlers/foo_test.go`
2. Template in `web/templates/pages/foo.html`
3. Handler in `internal/handlers/foo.go`
4. Register the route in `internal/app/app.go`

## HTMX interactions

- Partial in `web/templates/partials/`
- Template attributes: `hx-post`, `hx-target`, `hx-swap`
- Handler: if `cais.IsHTMX(r)` → `RenderPartial`, else → `Render` with layout
- Test with `req.Header.Set("HX-Request", "true")`

## New table

1. Store test with `":memory:"` before the migration
2. SQL in `internal/store/migrations/NNN_name.sql`
3. Methods on the `store.Store` interface

## CLI generators

```bash
cais new myapp --minimal
cais g resource bookmark --fields title:string,url:url,notes:text? --public
cais doctor                    # verify htmx, air, go.mod
cais g handler settings
cais g console              # scaffold cmd/console/main.go
```

Field types: `string`, `text`, `url`, `bool`, `int`. Suffix `?` for optional.

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
make docker   # ~15-20MB image
```

Pre-commit (tests, lint, prettier): `make pre-commit-install` once, then hooks run on every commit.

## Do not

- Parse templates per request (use `cais.NewRenderer`)
- Use inline CSS (use Tailwind classes in templates)
- Mock the database (use SQLite `:memory:`)
- Import `internal/` from `pkg/cais/` (avoids import cycles)
