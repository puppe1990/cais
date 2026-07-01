# Cais

Full-stack Go framework for mini apps on Lightsail: server-side HTML, HTMX, Tailwind, and SQLite.

## Stack

- Go 1.22+ (`net/http` stdlib)
- `html/template` + HTMX 2.x
- PWA by default (manifest, service worker, offline page, icons)
- Tailwind CSS 3.x
- SQLite (`modernc.org/sqlite`, no CGO)

## CI and pre-commit

GitHub Actions runs tests, `golangci-lint`, and Prettier on every push/PR to `main`.

```bash
make pre-commit-install   # once: installs git hooks
make ci                   # test + lint + format-check locally
```

Pre-commit hooks run: trailing whitespace, Prettier, `go fmt`, `go test`, and `golangci-lint`.

## CLI (Rails-style)

Install the `cais` command:

```bash
make install-cli
export PATH="$HOME/go/bin:$PATH"
```

| Command                                                                         | Description                                   |
| ------------------------------------------------------------------------------- | --------------------------------------------- |
| `cais new <app> [dir]`                                                          | Scaffold a new app (home, contact, dashboard) |
| `cais new <app> [dir] --minimal`                                                | Slim app (home only)                          |
| `cais g handler <name>`                                                         | Handler + test + page + route                 |
| `cais g resource <name> [--fields title:string,url:url] [--public] [--no-seed]` | Full CRUD + optional public page              |
| `cais g page <name>`                                                            | Page template only                            |
| `cais g migration <name>`                                                       | SQL migration file                            |
| `cais doctor`                                                                   | Check htmx, air, go.mod, CSS                  |
| `cais server`                                                                   | Run `go run ./cmd/server`                     |
| `cais test`                                                                     | Run `go test ./...`                           |

```bash
cais new dashboard ../dashboard
cd ../dashboard && npm install && make dev
```

## Quick start

Requires Go on your PATH and `~/go/bin` for hot reload (`air`):

```bash
export PATH="$HOME/go/bin:$PATH"
make pwa      # regenerate manifest, icons, service worker
make dev      # http://localhost:8080
make test     # full test suite
make build    # builds bin/cais
make docker   # optimized image
```

## Structure

```
pkg/cais/          → framework (router, render, config, htmx)
internal/app/      → bootstrap and routes
internal/handlers/ → HTTP handlers
internal/store/    → SQLite + migrations
web/templates/     → HTML
web/static/        → CSS + JS
cmd/server/        → entry point
```

## Framework APIs

**Router** — path params and route groups:

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Group(middleware.Protect, func(g *cais.Router) {
  g.Get("/admin/items", admin.Index)
  g.Get("/admin/items/{id}/edit", cais.IntParam("id", admin.Edit))
})
```

**httpx** — less render boilerplate:

```go
httpx.RenderOrError(w, renderer, "base", "home", data)
httpx.RenderPartial(w, renderer, "errors", data)
httpx.SeeOther(w, r, "/admin")
```

**testutil** — shared test helpers for scaffolded apps:

```go
renderer := testutil.NewRenderer(t)
req := testutil.NewRequest(http.MethodGet, "/items/1", testutil.PathValue("id", "1"))
```

**Admin auth** — opt-in via `ADMIN_TOKEN` env (no-op when unset):

```go
r.Get("/admin/products", middleware.Protect(admin.Index))
```

## Environment variables

| Variable      | Default         | Description                         |
| ------------- | --------------- | ----------------------------------- |
| `PORT`        | `:8080`         | Server port                         |
| `DB_PATH`     | `./data/app.db` | SQLite file path                    |
| `ENV`         | `development`   | Environment                         |
| `ADMIN_TOKEN` | _(empty)_       | Bearer/query token for admin routes |

## Deploy (Lightsail)

```bash
make docker
docker run -p 8080:8080 -v cais-data:/app/data cais:latest
```

Health check: `GET /health` → `{"status":"ok"}`

## AI-assisted development

See [AGENTS.md](AGENTS.md) — mandatory TDD, handler conventions, HTMX, and store patterns.
