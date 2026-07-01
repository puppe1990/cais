# Cais

Full-stack Go framework for mini apps on Lightsail: server-side HTML, HTMX, Tailwind, and SQLite.

## Stack

- Go 1.22+ (`net/http` stdlib)
- `html/template` + HTMX 2.x
- Tailwind CSS 3.x
- SQLite (`modernc.org/sqlite`, no CGO)

## Quick start

Requires Go on your PATH and `~/go/bin` for hot reload (`air`):

```bash
export PATH="$HOME/go/bin:$PATH"
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

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `:8080` | Server port |
| `DB_PATH` | `./data/app.db` | SQLite file path |
| `ENV` | `development` | Environment |

## Deploy (Lightsail)

```bash
make docker
docker run -p 8080:8080 -v cais-data:/app/data cais:latest
```

Health check: `GET /health` → `{"status":"ok"}`

## AI-assisted development

See [AGENTS.md](AGENTS.md) — mandatory TDD, handler conventions, HTMX, and store patterns.