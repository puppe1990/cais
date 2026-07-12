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

| Directory                | Responsibility                                                                        |
| ------------------------ | ------------------------------------------------------------------------------------- |
| `pkg/cais/`              | Framework: config, router, render, htmx, middleware                                   |
| `pkg/cais/httpx/`        | Render and redirect helpers for handlers                                              |
| `pkg/cais/meta/`         | Open Graph / Twitter preview (`Site`, `PreviewHTML`)                                  |
| `pkg/cais/session/`      | Cookie sessions (`SignIn`, `SignOut`, `Store`)                                        |
| `pkg/cais/boot/`         | Rails-style startup banner                                                            |
| `pkg/cais/devlog/`       | Development log buffer + `/logs` viewer                                               |
| `pkg/cais/sqllog/`       | SQL query logging wrapper (`Wrap`, `EnabledForEnv`)                                   |
| `pkg/cais/console/`      | Interactive REPL (yaegi + SQL)                                                        |
| `pkg/cais/csrf/`         | CSRF tokens (double-submit cookie)                                                    |
| `pkg/cais/validate/`     | Form field validation helpers                                                         |
| `pkg/cais/forms/`        | Template helpers (`csrfField`, `fieldError`, `makeField`, `fieldInput`)               |
| `pkg/cais/i18n/`         | Locale catalogs (`LOCALE` env, `t` template func)                                     |
| `pkg/cais/testutil/`     | Test helpers (`NewRenderer`, `NewRequest`, `AssertHTMLContains`, `AssertChatMarkers`) |
| `pkg/cais/pwa/`          | Default PWA assets generator (manifest, icons, og.png)                                |
| `pkg/cais/cache/`        | In-memory TTL cache + stable `Key`/`Hash` for ETags                                   |
| `pkg/cais/pagination/`   | Offset/limit helpers for list pages                                                   |
| `pkg/cais/stream/`       | SSE relay (`RelaySSE`, `WriteEvent`, `Flush`)                                           |
| `pkg/cais/chat/`         | Chat bubbles, tool UI, SSE writers (`WriteStream`, `WriteMessage`)                    |
| `pkg/cais/middleware/`   | CSRF, sessions, auth, rate limits, security headers                                   |
| `pkg/cais/passwordreset/`| Password-reset tokens + notifier interface                                            |
| `pkg/cais/sqlite/`       | WAL, busy timeout, foreign keys (`Configure`)                                         |
| `pkg/cais/netutil/`      | Health payload + LAN URLs for mobile testing                                          |
| `pkg/cais/migrate/`      | `schema_migrations` runner (idempotent on boot)                                       |
| `pkg/cais/flash/`        | One-shot flash cookies                                                                |
| `pkg/cais/jobs/`         | SQLite background job queue                                                           |
| `internal/cli/`          | Generators (`cais new`, `cais g`, `cais destroy`) — **Inertia + Svelte scaffolds**    |
| `internal/app/`          | This repo's dogfood app bootstrap (`deps.Inertia`)                                    |
| `internal/handlers/`     | HTTP handlers (gonertia + optional httpx fallback)                                    |
| `internal/store/`        | SQLite persistence                                                                    |
| `web/templates/app.html` | Inertia root shell (loads Vite bundle)                                                |
| `web/src/pages/`         | Svelte 5 pages (`@inertiajs/svelte`)                                                  |
| `web/static/build/`      | Vite production output (`npm run build`)                                              |
| `web/static/`            | Tailwind CSS, PWA assets, legacy HTMX JS                                              |
| `web/embed.go`           | `embed.FS` for templates in production                                                |
| `cmd/server/`            | Entry point                                                                           |

### Generated apps (`cais new`)

Same layout as this repo's demo: `app.html` + `web/src/pages/*.svelte` + gonertia + Vite → `web/static/build/`. Default scaffold uses `pwa.InstallForInertia` (no HTMX JS bundles).

`cais g handler` / `cais g page` generate **Svelte** pages in `web/src/pages/`.

`cais g resource` / `cais g stream chat` still generate **HTMX/html** admin and chat templates until ported to Svelte.

## Router path params and groups

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Post("/chat/{id}/permissions/{permID}/approve", cais.StringParams("id", "permID", chat.ApprovePermission))
r.Get("/items/{id}/{slug}", cais.IntStringParams("id", "slug", items.Show))
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
auth := handlers.NewAuthHandler(renderer, store, site, store.Sessions(), cfg, catalog, inertia)
session.SignIn(w, sessions, r, userID, session.CookieOptionsFromConfig(cfg))
flash.Set(w, "notice", "Bem-vindo!", cfg.CookieSecure())
```

Dev seed user: `demo@example.com` / `password`. Sessions persist in SQLite via `session.NewSQLiteStore`.

**Session expiry** — cookies and DB rows expire after 7 days (`sessionTTL` / `defaultMaxAge`). SQLite stores `expires_at`; expired rows are ignored on lookup. Prune stale rows with `cais db prune-sessions` (or call `session.Store.PruneExpired()`).

**Production cookies** — `session.CookieOptionsFromConfig(cfg)` sets `Secure` when `cfg.CookieSecure()` is true (`ENV=production`).

## Inertia + Svelte (default frontend)

Handlers accept `*inertia.Inertia` and branch on `h.inertia != nil` (Inertia path) vs `httpx.RenderOrError` (legacy HTMX templates still in this repo for chat/admin).

```go
// GET
_ = h.inertia.Render(w, r, "Contact", inertia.Props{"site": meta.ForRequest(h.site, r)})

// Validation errors (422) — re-render same component; gonertia fills props.errors
ve := make(inertia.ValidationErrors)
ve["email"] = "Invalid email"
ctx := inertia.SetValidationErrors(r.Context(), ve)
_ = h.inertia.Render(w, r.WithContext(ctx), "Contact", inertia.Props{})

// Flash on redirect
ctx := inertia.SetFlash(r.Context(), inertia.Flash{"notice": "Saved!"})
h.inertia.Redirect(w, r.WithContext(ctx), "/dashboard", http.StatusSeeOther)
```

Svelte pages use `@inertiajs/svelte`:

```svelte
<script>
  import { useForm } from '@inertiajs/svelte'
  export let errors = {}
  let form = useForm({ name: '', email: '' })
  function submit() { $form.post('/contact') }
</script>
```

Vite builds to `web/static/build/` (`npm run build`). `app.html` loads `/static/build/assets/main.js`.

**Frontend TDD** — Vitest + Testing Library in `web/src/pages/*.test.js`:

```bash
npm run test:fe   # or make js-test in framework repo (also runs pkg/cais/js/*.test.mjs)
```

## New page (scaffolded app)

1. Go test in `internal/handlers/foo_test.go` — `setupTestInertia` + `inertiaRequest` + `assertInertiaComponent` (see `inertia_test.go`)
2. Optional Svelte test in `web/src/pages/Foo.test.js`
3. Svelte page in `web/src/pages/Foo.svelte`
4. Handler — `h.inertia.Render(w, r, "Foo", inertia.Props{"site": meta.ForRequest(h.site, r)})`
5. Register the route in `internal/app/routes.go`

Or use `cais g handler foo` (creates handler + test + `web/src/pages/Foo.svelte` + route patch).

Pass `meta.SiteFrom(appName, cfg.AppURL)` from bootstrap for OG/Twitter props in Inertia pages.

**Inertia integration tests** (`internal/app/app_test.go`) — set `X-Inertia: true` on requests; successful POSTs return `303` redirects; validation returns `200` with `props.errors`.

## CSRF

- `middleware.CSRF(cfg)` on the router (validates POST/PUT/DELETE/PATCH)
- Double-submit cookie (`cais_csrf`) — token in cookie + form field or `X-CSRF-Token` header; no server-side token store

**HTMX templates** — pass `meta.ForRequest(site, r)` in page data; layout renders `<meta name="csrf-token">` + `cais.js` sends `X-CSRF-Token` on HTMX requests. HTML forms: `{{ csrfField .CSRFToken }}` or hidden `csrf_token` input.

**Inertia + Svelte** — `@inertiajs/svelte` `useForm` posts include the CSRF cookie automatically; middleware still validates cookie + body/header.

**Integration tests** — GET page first (read `csrf` cookie), then POST with matching `csrf_token` field + cookie (and `X-Inertia: true` for Inertia routes).

## Flash messages

- `middleware.Flash` on the router (after `LoadSession`)
- Set on redirect: `flash.Set(w, "notice", "Saved!", cfg.CookieSecure())` — read in templates via `meta.ForRequest(site, r)` → `.Flash`
- Layouts: `{{ flashMessage .Flash }}` (`pkg/cais/forms`) — never `{{ .Flash }}` (stringifies the struct)
- One-shot: consumed on the next request

## Mobile PWA

- `boot.Print` shows **LAN** URLs for phone testing on Wi‑Fi
- `GET /health` returns `lan_urls` via `netutil.HealthPayload` — use this array, never concatenate `APP_URL` + port manually
- `cais pwa --bump` increments `CACHE_VERSION` in `sw.js` after template/HTML changes
- `cais dev` auto-bumps `CACHE_VERSION` when `sw.js` exists (fresh assets on phone without a manual bump)
- `cais doctor --mobile` checks flash markup, Google Fonts CSP, SW cache, chat SSE partial, SSE reconnect, chat agent JS (`finalizeChatStream`), chat enter-submit JS (`bindChatEnterSubmit`), chat form CSS (`.cais-chat-shell`), `#chat-messages` scroll container, and health `lan_urls`
- Scaffold `input.css` uses system fonts (no `fonts.googleapis.com` — blocked by default CSP)

**Mobile SSE checklist:** `cais doctor --mobile` → `cais pwa --bump` → open boot **LAN** URL on phone → stay on chat page while agent responds (avoid hx-boost away mid-stream)

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

## SSE / streaming

Cais ships `sse-ext.min.js` (HTMX SSE extension). Long-lived streams need server and handler setup:

| Setting        | Normal handlers | SSE routes                                             |
| -------------- | --------------- | ------------------------------------------------------ |
| `WriteTimeout` | `30s` ok        | **`0`** (disabled) — scaffold default                  |
| Flush          | N/A             | `stream.Flush(w)` — never assert `http.Flusher` on `w` |

```go
import (
    "github.com/puppe1990/cais/pkg/cais/chat"
    "github.com/puppe1990/cais/pkg/cais/stream"
)

func streamHandler(w http.ResponseWriter, r *http.Request) {
    stream.RelaySSE(w)
    _ = chat.WriteStream(w, chat.LiveBubble("token…"))   // event: stream
    _ = chat.WriteMessage(w, chat.MessageBubble(chat.RoleAssistant, "done", time.Now().UTC()))
    _ = stream.WriteEvent(w, "thinking", chat.ThinkingHTML("…")) // prefer over manual fmt.Fprintf
    _ = stream.Flush(w) // never w.(http.Flusher) — middleware may hide Flusher
}

// Proxy upstream SSE (e.g. OpenCode /event) to the browser:
func relayHandler(w http.ResponseWriter, upstream *http.Response) {
    stream.RelaySSE(w)
    _, _ = stream.RelayAndCopy(w, upstream.Body)
}
```

**Chat template patterns** — two partials ship with `cais new`:

| Partial               | Use case                        | DOM                                                                    |
| --------------------- | ------------------------------- | ---------------------------------------------------------------------- |
| `chat_sse.html`       | Echo / simple bots              | Single `#chat-history`; SSE appends with `beforeend`                   |
| `chat_sse_agent.html` | Agent streaming (tokens, tools) | `#chat-history` + `#chat-stream` + `#chat-live` under `#chat-messages` |

Simple mode — `chat_sse.html`: `#chat-sse` uses `sse-swap="message"` + `hx-swap="beforeend"` + `hx-target="#chat-history"`. Set `data-cais-sse-persist="true"` so `cais.js` reconnects SSE after `hx-boost` navigation.

Agent mode — `chat_sse_agent.html`: opt-in with `data-cais-chat="true"`. SSE events:

| Event      | Target           | Swap        | Purpose                                   |
| ---------- | ---------------- | ----------- | ----------------------------------------- |
| `stream`   | `#chat-live`     | `innerHTML` | Live token/tool bubble (`data-cais-live`) |
| `message`  | `#chat-stream`   | `beforeend` | Finalized assistant chunks                |
| `thinking` | `#chat-thinking` | `outerHTML` | Thinking indicator                        |

**Ordering pitfall** — `hxChatForm` posts user bubbles to `#chat-history` with `beforeend`. If agent SSE writes to sibling containers (`#chat-stream`, `#chat-live`) without merging first, the user input appears above in-flight assistant output. Agent mode requires `cais.js` `finalizeChatStream()` (runs automatically when `data-cais-chat` is present; also exposed as `window.caisFinalizeChatStream`).

**Server helpers** — `pkg/cais/chat`: `LiveBubble`, `MessageBubble`, `SafeMessageBubble`, `ToolCallBubble`, `ToolResultBubble`, `DetailBubble`, `Truncate`, `SelectWindowWithLastUser`, `UnsafeLiveHTML` / `UnsafeMessageHTML` (caller sanitizes HTML), `WriteStream`, `WriteMessage` (wrap `stream.WriteEvent`).

**Chat form** — `{{ hxChatForm "/chat/{id}/messages" "#chat-thinking" }}` on the `<form>`: `cais.js` `bindChatEnterSubmit` sends on Enter, Shift+Enter newline. Input clears on submit; use `data-cais-chat-optimistic="true"` only when the POST partial does **not** return a user bubble (otherwise `dedupOptimisticUserBubble` drops the duplicate). Optional `data-cais-poll-url` on `#chat-sse` enables history refresh fallback when SSE fails (poll is skipped while stream slots are active).

`cais doctor` warns when `sse-ext.min.js` is present and `WriteTimeout > 0` in `internal/app/app.go`.

**Chat handler tests** — `cais g stream chat` generates handler tests using `testutil.AssertChatMarkers` (Show) and `testutil.AssertHTMLContains` (PostMessage bubble). Missing records return `http.NotFound` (not 500):

```go
conv, err := h.store.FindConversationByID(id)
if err != nil {
    http.NotFound(w, r)
    return
}
```

## HTMX interactions

Used by `cais g resource` / `cais g stream chat` admin and chat UIs until ported to Svelte.

- Partial in `web/templates/partials/` — each file has `{{ define "name" }}` matching the filename
- Partials are parsed into full pages too, so `{{ template "name" . }}` works in pages and layouts
- For HTMX swaps, handler returns the partial via `RenderPartial` (not a full layout)
- Template attributes: `hx-post`, `hx-target`, `hx-swap`
- Handler: prefer `httpx.WritePage` (thin wrapper over `RenderPageOrPartial`) for forms that return partial on HTMX and full page otherwise
- Test with `req.Header.Set("HX-Request", "true")`

```go
httpx.WritePage(w, r, renderer, httpx.PageConfig{
  Layout: "base", Page: "contact", Partial: "contact_errors", Data: data, Status: 422,
}, cfg)
```

**List caching** — combine `cache.Key(...)` + `cache.Hash(version)` with `httpx.NotModified` / `httpx.SetETag` for 304 responses on stable list pages.

## Form validation

Use `validate.Email`, `validate.URL`, `validate.Required`, `validate.MinLength`, `validate.MaxLength` for single-field checks. For multiple fields, collect errors in `validate.FieldErrors`:

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

**Form helpers** (`pkg/cais/forms`, registered on the renderer):

```html
{{ fieldInput (makeField "name" "Name" .Name "text" true .Errors) }}
```

`makeField` returns `forms.FieldData`; `fieldInput` renders input/textarea/checkbox + error. Resource generator admin forms use these by default.

Foreign-key selects:

```html
{{ fieldSelect (makeSelectField "category_id" "Category" .Item.CategoryID .CategoryOptions true
.Errors) }}
```

## Foreign keys in generators

`cais g resource post --fields title:string,category_id:references` (or `category:belongs_to`):

- Migration column: `INTEGER [NOT NULL] REFERENCES categories(id)`
- Store: `ListCategoryOptions()` — `SELECT id, COALESCE(name, title, id) FROM categories`
- Admin form: `fieldSelect` / `makeSelectField` populated from options

Generate the parent resource first (`cais g resource category --fields name:string`). Referenced table needs a `name` or `title` column for labels.

## Integration tests (auth + contact)

Multi-step flows belong in `internal/app/app_test.go` (full router + CSRF + session + Inertia):

1. `GET /login` or `/contact` — read `csrf` cookie
2. `POST` with matching `csrf_token` field + cookie; set `X-Inertia: true` for Inertia routes
3. Follow session/flash cookies on subsequent requests

See `TestApp_LoginPost_withCSRF_redirects`, `TestApp_AuthFlow_loginDashboardLogout`, `TestApp_ContactPost_validationWithCSRF_returns422`, `TestApp_Smoke_contactInertia_loginDashboardLogout`.

## HTMX UX (app-like feel)

HTMX layouts load `cais-core.js` (or full `cais.js`) after `htmx.min.js` — CSRF header, focus restore, optimistic toggles, nav tab sync. Chat pages add `cais-chat.js` (testable logic in `cais-chat-logic.mjs`).

- **App shell** — `#cais-main` + `#cais-nav`; `navTab` links use `hx-boost` (swap main only, no full reload)
- **Small targets** — swap `#form-errors` or `this`, not whole lists
- **Transitions** — `hx-swap="innerHTML swap:150ms"` on forms; `outerHTML swap:150ms` on toggles
- **Forms** — `{{ hxForm "/path" "#errors" "#spinner" }}` (`pkg/cais/htmxattrs`); `.htmx-request-hide` on submit label
- **Bool toggles** — `data-cais-optimistic="toggle"` for instant class flip (see resource generator)
- **Count / remove** — `data-cais-optimistic="count"` or `"remove"` for feed-style actions (rollback on error)
- **Morph swaps** — default for nav (`hx-swap="morph:innerHTML"`) and paginate for better mobile/SPA state. The old `data-cais-view-transition` attr was removed (was dead code). Use `hxMorphOuter` helper for toggles.
- **CSS** — `input.css`: `.htmx-swapping`, `.htmx-settling`, `.cais-skeleton`, `.cais-toast-enter`
- **Response headers** — `cais.SetToast`, `cais.SetFocus(w, "#field")`, `cais.SetRetarget`, `cais.SetTrigger`
- **Admin CRUD** — `cais g resource` generates `hxForm` admin forms, inline delete (`hx-swap="delete"`), `RenderPageOrPartial` on 422

## New table

1. Store test with `":memory:"` before the migration
2. SQL in `internal/store/migrations/NNN_name.sql`
3. Methods on the `store.Store` interface
4. Wrap DB with `sqllog.Wrap` in `NewSQLiteStore` for development query logs
5. Migrations tracked in `schema_migrations` via `pkg/cais/migrate` (idempotent on boot)

**SQLite concurrency (SSE / chat)** — scaffold `NewSQLiteStore` calls `sqlite.Configure`: `journal_mode=WAL`, `busy_timeout=5000`, `foreign_keys=ON`, `MaxOpenConns(1)`. WAL allows concurrent readers while a writer holds the lock briefly; `busy_timeout` retries instead of immediate `SQLITE_BUSY`. SSE handlers poll the DB in a loop — keep writes short and avoid long transactions during streams. For heavy write concurrency, consider a dedicated writer queue.

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

- `middleware.LoggerTo(devlog.MirrorDefault(...))` — JSON request logs when `cfg.LogJSON()` (`kind: request`); `LOG_FORMAT=text` opts out
- `sqllog.ConfigForEnv(env)` — SQL JSON logs in development (`kind: sql`); plain text when `JSON: false`
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
cais g [--dry-run] stream chat              # SSE chat: conversations, messages, stream relay scaffold
cais g [--dry-run] resource bookmark --fields title:string,url:url,notes:text? --public --paginate --force
cais destroy [--dry-run] resource bookmark   # undo generator output
cais destroy [--dry-run] model bookmark      # remove model + migration + store methods
cais destroy [--dry-run] handler settings
cais destroy [--dry-run] auth                # remove login/auth scaffolding
cais destroy [--dry-run] migration add_tags  # remove *_add_tags.sql
cais g [--dry-run] model bookmark --fields title:string,url:url
cais g [--dry-run] handler settings
cais g [--dry-run] page about
cais g [--dry-run] migration add_tags
cais g [--dry-run] auth       # login/logout + protected dashboard
cais g [--dry-run] console    # scaffold cmd/console/main.go
cais g [--dry-run] ci         # add CI/pre-commit to existing apps
cais g [--dry-run] job send_welcome --cron "0 3 * * *"
cais doctor [--mobile]        # verify setup (Inertia/Vite, htmx, air, go.mod, PWA/mobile)
cais pwa [--bump]             # write/refresh PWA assets; --bump invalidates SW cache
cais link [path] [--unlink]   # go.mod replace for local Cais dev
cais routes                   # list routes from internal/app/routes.go
```

**Aliases:** `cais g` → generate, `cais i` → install, `cais b` → build, `cais s` → server, `cais c` → console.

Field types: `string`, `text`, `url`, `bool`, `int`, `date`, `references` (or `name:belongs_to`). Suffix `?` for optional.

**Resource options:** `--public` (public list page), `--paginate` (admin index pagination, 25/page), `--no-seed` (skip demo data), `--admin-auth session|bearer` (default: session).

**Model generator** — `cais g model` creates model struct, migration, and store methods only (no handlers, templates, or routes). Use for data layer without admin CRUD.

**Dry-run** — `cais g --dry-run ...` and `cais destroy --dry-run ...` print planned changes without writing files.

**Destroy** — `cais destroy resource|handler|model <name>` removes generated files and unpatches `routes.go`, `store.go`, `seeds.go`, and layout nav where applicable. `destroy auth` also reverts `app.go` session middleware. `destroy migration` removes matching `*_<name>.sql` only (does not roll back `schema_migrations`).

**Demo seed** — `cais g resource` (unless `--no-seed`) generates `SeedDemo*` store methods and wires them into `cmd/server/main.go` at boot.

## App commands (run from a Cais app)

```bash
cais install  # npm install + go mod tidy
cais css      # build Tailwind
cais dev      # hot reload + tailwind watch
cais build    # bin/server
cais server   # go run ./cmd/server
cais test     # go test ./...
npm run test:fe           # vitest (Svelte pages; also in scaffold package.json)
cais doctor [--mobile]  # verify Inertia/Vite, htmx, air, go.mod, PWA/mobile
cais pwa [--bump]       # write/refresh PWA assets
cais console  # Rails-style REPL (store, cfg, db + sql)
cais routes   # list HTTP routes from internal/app/routes.go
cais link [../Cais] [--unlink]  # go.mod replace for local framework dev
cais db migrate        # run pending migrations
cais db status         # list applied/pending migrations
cais db rollback       # roll back last migration (runs -- down SQL when present)
cais db prune-sessions # delete expired login sessions from SQLite
cais db seed           # run internal/db/seeds.go (idempotent; safe in production for catalog data)
cais db seed --list    # list seed helpers referenced in seeds.go
cais jobs work [--queues default,mail] [--concurrency 2]
cais jobs status
cais routes --verbose  # routes with handler names and middleware
cais version           # print framework version
```

## Background jobs

SQLite queue in `pkg/cais/jobs` (same DB file as the app). See [jobs design](docs/superpowers/specs/2026-07-01-jobs-design.md).

```bash
cais g job prune_sessions --cron "0 3 * * *"  # internal/jobs/*.go + registry + cmd/worker
cais db migrate                                # jobs + recurring_tasks tables
cais jobs work --concurrency 2                   # worker + delayed-job dispatcher
cais jobs status
```

Enqueue from handlers:

```go
jobs.Enqueue(ctx, jobStore, jobs.Options{Kind: "SendWelcome", Payload: data})
```

Register handlers in `internal/jobs/registry.go`. Built-in: `PruneSessions`. Production: run `cais jobs work` as a separate process next to `bin/server`.

**Generator troubleshooting** — if `could not patch routes.go` or `could not patch store`, check that `registerRoutes` and `Close() error` markers exist. Public nav links need `<!-- cais:nav -->` in the layout (or `</nav>`). Run `cais db migrate` after `g resource` / `g model` / `g auth`.

Console bindings: `store`, `cfg`, `db`, plus any custom keys in `Bindings`. Commands: `help`, `sql`, `reload`, `history`, `!N`/`!!`, `exit`. Arrow keys when stdin is a TTY.

`/logs` — development-only log viewer (localhost). Shows request + SQL logs with HTMX auto-refresh.

## CLI generator layout

The `cais` CLI lives in `internal/cli/`. Scaffold templates are split by responsibility so agents can grep a single file instead of loading a 2400-line monolith.

| Path                                      | Responsibility                                                 |
| ----------------------------------------- | -------------------------------------------------------------- |
| `internal/cli/cli.go`                     | Command routing (`new`, `g`, `destroy`, `db`, …)               |
| `internal/cli/scaffold.go`                | `cais new` orchestration and `writeTemplate`                   |
| `internal/cli/resource.go`                | `cais g resource` orchestration (writes files, calls patches)  |
| `internal/cli/resource_patch.go`          | Patches store, routes, layout nav, seeds, main for resources   |
| `internal/cli/resource_gen_*.go`          | Resource code generation (store, admin, public, HTML, fields)  |
| `internal/cli/tpl_scaffold_*.go`          | Embedded `const tpl*` for `cais new` scaffolding               |
| `internal/cli/tpl_scaffold_main.go`       | `cmd/server/main.go` (full + blank)                            |
| `internal/cli/tpl_scaffold_app_core.go`   | `internal/app/app.go` (full + blank)                           |
| `internal/cli/tpl_scaffold_routes.go`     | `internal/app/routes.go` (full, minimal, blank)                |
| `internal/cli/tpl_scaffold_console.go`    | `cmd/console/main.go`                                          |
| `internal/cli/tpl_scaffold_auth.go`       | Auth Go templates (handler, store, model, migration, tests)    |
| `internal/cli/tpl_scaffold_auth_pages.go` | Auth HTML page templates (`login`, `signup`, reset)            |
| `internal/cli/tpl_scaffold_inertia.go`    | Vite, Svelte, `app.html`, `web/src/main.js`                      |
| `internal/cli/tpl_scaffold_handlers_inertia.go` | Inertia handler scaffolds (home, contact, auth, dashboard) |
| `internal/cli/tpl_scaffold_web.go`        | Legacy HTMX layout fragments (resource/chat generators)        |
| `internal/cli/tpl_stream_chat.go`         | `cais g stream chat` templates and handlers                    |
| `internal/cli/stream.go`                  | `cais g stream chat` orchestration                             |
| `internal/cli/scaffold_auth.go`           | `cais g auth` orchestration and store/app/route patches        |
| `internal/cli/patch.go`                   | AST-safe patches into generated apps (`routes.go`, `store.go`) |
| `internal/cli/patch/`                     | `go/ast` helpers — regex patches break nested `cais.IntParam`  |
| `internal/cli/doctor.go`                  | `cais doctor` checks (Inertia, Vite, HTMX, PWA, mobile)        |
| `internal/cli/pwa_cmd.go`                 | `cais pwa` asset writer                                        |
| `internal/cli/commands.go`                | `cais install`, `cais css`, `cais dev`                         |
| `internal/cli/destroy.go`                 | `cais destroy` — reverses generators                           |

**Generator tests** (split by domain — run focused suites while editing generators):

| Test file                    | Scope                                  |
| ---------------------------- | -------------------------------------- |
| `cli_help_test.go`           | `cais help` output                     |
| `cli_new_test.go`            | `cais new` (full, minimal, blank)      |
| `resource_scaffold_test.go`  | `cais g resource`                      |
| `scaffold_handler_test.go`   | `cais g handler` route patching        |
| `scaffold_model_test.go`     | `cais g model`                         |
| `scaffold_migration_test.go` | `cais g migration` numbering           |
| `generate_dryrun_test.go`    | `--dry-run` generators                 |
| `patch_gomod_test.go`        | `replace` directive in scaffolded apps |

```bash
go test ./internal/cli/... -run TestScaffoldResource -count=1
go test ./internal/cli/... -run TestCLI_New -count=1
go test ./internal/cli/... -count=1
```

**Patch markers** — generated apps must keep `registerRoutes`, `Close() error`, and `<!-- cais:nav -->` (or `</nav>`) for destroy/generator patches to work.

## Framework commands (Cais repo)

```bash
make test-v         # TDD: watch RED/GREEN
make test           # Go validation with -race
make js-test        # Vitest Svelte + cais-chat-logic .mjs tests
make lint           # golangci-lint
make format         # prettier --write
make ci             # test + js-test + lint + format-check
make dev            # hot reload + tailwind watch
make build          # bin/cais
make install-cli    # go install ./cmd/cais
make pwa            # regenerate PWA assets (manifest fullscreen, icons, og.png)
make docker         # ~15-20MB image
```

CI builds Vite assets (`npm ci && npm run build`) before `go test`.

## Production deploy (Lightsail / systemd)

Cross-compile and ship static assets beside the binary:

```bash
npm run build   # Vite → web/static/build/
cais build --os linux --arch amd64 -o bin/server-linux
tar czf release.tar.gz bin/server-linux web/static
```

- Guide: `docs/deploy/lightsail-systemd.md`
- Template: `deploy/systemd/cais-app.service.example`
- `cais doctor` checks `web/static` + `manifest.webmanifest`
- Set `STATIC_DIR` / `TEMPLATES_DIR` when `WorkingDirectory` is not the app root
- Dev-only seeds (demo user) do not run when `ENV=production`; use `cais db seed` for catalog data

Pre-commit (tests, lint, prettier): `make pre-commit-install` once, then hooks run on every commit.

## Agent code style

- Comment **why** on non-obvious constraints (security tradeoffs, SQLite limits, boot-time vs per-request).
- Do not narrate what the code does — names and tests already cover that.
- Prefer file- or func-level provenance over inline noise; agents load whole files, not single lines.

## Do not

- Parse templates per request (use `cais.NewRenderer`)
- Use inline CSS (use Tailwind classes in templates)
- Mock the database (use SQLite `:memory:`)
- Import `internal/` from `pkg/cais/` (avoids import cycles)
