# Cais

Framework full stack em Go para mini apps no Lightsail: server-side HTML, HTMX, Tailwind e SQLite.

## Stack

- Go 1.22+ (`net/http` stdlib)
- `html/template` + HTMX 2.x
- Tailwind CSS 3.x
- SQLite (`modernc.org/sqlite`, sem CGO)

## Quick start

Requer Go no PATH e `~/go/bin` para o hot-reload (`air`):

```bash
export PATH="$HOME/go/bin:$PATH"
make dev      # http://localhost:8080
make test     # suite completa
make build    # gera bin/cais
make docker   # imagem otimizada
```

## Estrutura

```
pkg/cais/          → framework (router, render, config, htmx)
internal/app/      → bootstrap e rotas
internal/handlers/ → handlers HTTP
internal/store/    → SQLite + migrations
web/templates/     → HTML
web/static/        → CSS + JS
cmd/server/        → entry point
```

## Variáveis de ambiente

| Variável | Default | Descrição |
|----------|---------|-----------|
| `PORT` | `:8080` | Porta do servidor |
| `DB_PATH` | `./data/app.db` | Caminho do SQLite |
| `ENV` | `development` | Ambiente |

## Deploy (Lightsail)

```bash
make docker
docker run -p 8080:8080 -v cais-data:/app/data cais:latest
```

Health check: `GET /health` → `{"status":"ok"}`

## Desenvolvimento com IA

Leia [AGENTS.md](AGENTS.md) — TDD obrigatório, convenções de handlers, HTMX e store.