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

| Directory            | Responsibility                                      |
| -------------------- | --------------------------------------------------- |
| `pkg/cais/`          | Framework: config, router, render, htmx, middleware |
| `internal/app/`      | Bootstrap: route and dependency wiring              |
| `internal/handlers/` | HTTP handlers                                       |
| `internal/store/`    | SQLite persistence                                  |
| `web/templates/`     | HTML templates (layouts, pages, partials)           |
| `web/static/`        | Tailwind CSS, HTMX, PWA (manifest, sw.js, icons)    |
| `pkg/cais/pwa/`      | Default PWA assets generator                        |
| `cmd/server/`        | Entry point                                         |

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

## Commands

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
