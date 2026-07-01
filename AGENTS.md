# Cais — Convenções para IA

## Regra #1: TDD obrigatório

Antes de escrever código de produção:

1. Escreva o teste em `*_test.go`
2. Rode: `go test ./... -v -run TestNome`
3. Confirme que **falha** pelo motivo certo (feature ausente, não typo)
4. Escreva o código **mínimo** para passar
5. Rode: `make test`
6. Só então refatore

## Estrutura

| Pasta | Responsabilidade |
|-------|------------------|
| `pkg/cais/` | Framework: config, router, render, htmx, middleware |
| `internal/app/` | Bootstrap: wiring de rotas e dependências |
| `internal/handlers/` | Handlers HTTP |
| `internal/store/` | Persistência SQLite |
| `web/templates/` | Templates HTML (layouts, pages, partials) |
| `web/static/` | CSS Tailwind compilado + HTMX vendor |
| `cmd/server/` | Entry point |

## Nova página

1. Teste em `internal/handlers/foo_test.go`
2. Template em `web/templates/pages/foo.html`
3. Handler em `internal/handlers/foo.go`
4. Registrar rota em `internal/app/app.go`

## Interação HTMX

- Partial em `web/templates/partials/`
- Atributos no template: `hx-post`, `hx-target`, `hx-swap`
- Handler: se `cais.IsHTMX(r)` → `RenderPartial`, senão → `Render` com layout
- Teste com `req.Header.Set("HX-Request", "true")`

## Nova tabela

1. Teste store com `":memory:"` antes da migration
2. SQL em `internal/store/migrations/NNN_nome.sql`
3. Métodos na interface `store.Store`

## Comandos

```bash
make test-v   # TDD: ver RED/GREEN
make test     # validação com -race
make dev      # hot reload + tailwind watch
make build    # bin/cais
make docker   # imagem ~15-20MB
```

## Não fazer

- Parsear templates por request (usar `cais.NewRenderer`)
- CSS inline (usar classes Tailwind nos templates)
- Mocks de banco (usar SQLite `:memory:`)
- Importar `internal/` de `pkg/cais/` (evita ciclos)