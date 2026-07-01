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

| Command                   | Description                   |
| ------------------------- | ----------------------------- |
| `cais new <app> [dir]`    | Scaffold a new app            |
| `cais g handler <name>`   | Handler + test + page + route |
| `cais g page <name>`      | Page template only            |
| `cais g migration <name>` | SQL migration file            |
| `cais server`             | Run `go run ./cmd/server`     |
| `cais test`               | Run `go test ./...`           |

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

## Environment variables

| Variable  | Default         | Description      |
| --------- | --------------- | ---------------- |
| `PORT`    | `:8080`         | Server port      |
| `DB_PATH` | `./data/app.db` | SQLite file path |
| `ENV`     | `development`   | Environment      |

## Deploy (Lightsail)

```bash
make docker
docker run -p 8080:8080 -v cais-data:/app/data cais:latest
```

Health check: `GET /health` → `{"status":"ok"}`

## AI-assisted development

See [AGENTS.md](AGENTS.md) — mandatory TDD, handler conventions, HTMX, and store patterns.
