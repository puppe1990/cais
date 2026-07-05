# Changelog

All notable changes to the Cais framework are documented here.

Format based on [Keep a Changelog](https://keepachangelog.com/). Versioning follows [Semantic Versioning](https://semver.org/).

## Unreleased

### Added

- `pkg/cais/chat`: `Truncate`, `SafeMessageBubble`, `TrimForDisplay`, `MaxMessageChars` — server-side safety and perf for large/polluted agent histories (addresses loading 1-2s, 500s on huge turns)
- `pkg/cais/chat`: `SelectWindowWithLastUser` — trims history window while pinning the most recent user message (removes need for tail+pinned hacks)
- `cais g stream chat` scaffold now demonstrates TrimForDisplay + SafeMessageBubble + SelectWindowWithLastUser in Show/ListMessages/Stream

### Added

- `pkg/cais/sqlite` package docs — WAL / busy_timeout guidance for SSE chat apps
- README — `testutil` chat assertion examples

### Added

- `testutil.AssertHTMLContains` / `testutil.AssertChatMarkers` — chat handler HTML assertions
- `cais dev` auto-bumps PWA `CACHE_VERSION` when `sw.js` is present
- `cais doctor --mobile` — chat enter-submit JS (`bindChatEnterSubmit`) and chat form CSS checks
- `cais g stream chat` handler tests — Show 404, PostMessage user bubble, `AssertChatMarkers`

### Added

- `cais.js` `bindChatEnterSubmit` — delegated Enter-to-send on `form[data-cais-chat-form]` (Shift+Enter newline)
- `cais.js` `dedupOptimisticUserBubble` — drops optimistic user bubble when server partial already includes it
- Chat form submit CSS — `inline-flex` button + scoped `htmx-indicator` / `htmx-request-hide` swap

### Changed

- `hxChatForm` no longer uses inline `hx-on:keydown` or `this.reset()` — input clear and Enter handled in `cais.js`
- `cais g stream chat` submit button uses `htmx-request-hide` + `htmx-indicator` pattern
- `cais g stream chat` uses mobile `cais-chat-shell` layout (sticky footer, viewport height)
- `#chat-messages` includes `overflow-x-hidden` — `cais doctor --mobile` warns when missing
- `finalizeChatStream` runs after `#chat-history` swaps; `pruneEmptyChatNodes` removes empty SSE slots

### Added

- `chat.DetailBubble` — collapsible tool/log output for agent streams

### Added

- `hxChatForm` calls `window.caisFinalizeChatStream` before submit to merge SSE stream slots
- `cais doctor --mobile` — chat agent JS finalize check and `#chat-messages` scroll container check
- `cais g stream chat` demo uses `pkg/cais/chat` — `event: stream` typing preview + timestamped `event: message`
- Scaffold `input.css` — generic chat styles (`.cais-chat-scroll-down`, `.cais-msg-time`, `.cais-thinking-dots`)
- `cais.js` agent chat module — `finalizeChatStream`, device-local timestamps, stick-to-bottom scroll, poll guard (opt-in `data-cais-chat`)
- `chat_sse_agent.html` partial — multi-slot agent chat (`#chat-history` + `#chat-stream` + `#chat-live`, `data-cais-chat`)
- `pkg/cais/chat` — generic SSE chat HTML helpers (`LiveBubble`, `MessageBubble`, `ThinkingHTML`, `WriteStream`, `WriteMessage`)
- `pkg/cais/stream` — `Flush`, `RelaySSE`, and `RelayAndCopy` for HTMX SSE through middleware-wrapped `ResponseWriter`s
- Logger skips misleading `Completed` log line for `/stream` and `/event` paths
- `cais.js` reconnects SSE after `hx-boost` swaps (`data-cais-sse-persist`, `htmx:sseClose` handler)
- `hxChatForm` template helper — Enter-to-send chat forms with thinking indicator
- `chat_sse.html` partial — `#chat-thinking` indicator and optional `data-cais-poll-url` fallback
- `cais g stream chat` — conversations + messages migration, SSE handler, HTMX chat UI
- `netutil.HealthPayload` — `/health` exposes `lan_urls` for mobile testing
- `cais doctor --mobile` — chat SSE pattern, SSE reconnect JS, health `lan_urls` checks
- `{{ flashMessage .Flash }}` template helper in `pkg/cais/forms`
- `cais pwa [--bump]` — refresh PWA assets; `--bump` increments `CACHE_VERSION` in `sw.js`
- `cais doctor --mobile` — flash template, Google Fonts CSP, and PWA cache version checks
- `boot.Print` LAN URL line via `pkg/cais/netutil`
- `cais.PortBusy` and dev-server warning when the configured port is already in use
- `cais.StringParams` — ergonomic two-param routes without nested `StringParam` callbacks
- `cais.IntStringParams` — int + string path params in one wrapper
- `cais link [path] [--unlink]` — go.mod `replace` for local framework development
- Scaffold partial `chat_sse.html` — append-only SSE chat pattern (`#chat-history` + `#chat-sse`)
- `cais doctor` warns when `sse-ext.min.js` is installed but `WriteTimeout > 0`

### Changed

- Scaffold and reference app default `WriteTimeout: 0` so long-lived SSE connections are not killed at 30s
- Scaffold uses system font stack (no Google Fonts `@import`) to avoid CSP console errors
- Scaffold layouts use `{{ flashMessage .Flash }}` instead of struct stringification

## [0.6.0] - 2026-07-04

### Added

- `pkg/cais/barcode` — Open Food Facts lookup client
- `pkg/cais/money` — `FormatBRL` for cent-based prices
- `middleware.LoadUserStats` / `UserStatsFrom` — gamification chrome in layouts
- `meta.Site.LoggedIn` — session flag for layout auth chrome
- `Config` security knobs: `PERMISSIONS_POLICY`, `CSP_MEDIA_SRC`, `CSP_CONNECT_SRC` (camera + barcode scan in PWA)
- `Router.StaticForEnv` — `no-store` for static assets in development
- `NewRendererForEnv` — disk template reload in development
- Scaffold partials: `icons.html`, `nav_links.html` on `cais new`

### Removed

- `cais g app supermarket` and `internal/cli/app_templates/supermarket/` — app UI belongs in apps, not the framework
- `pkg/cais/ui` — nav/icon HTML helpers (use app templates instead)

## [0.5.0] - 2026-07-03

### Added

#### HTMX UX (app shell)

- `pkg/cais/ui` — `navTab`, `makeNavTab`, `icon` helpers; `Site.ActiveNav` for tab highlighting
- `pkg/cais/htmxattrs` — `hxForm`, `hxDelete`, `hxBoostLink`, `hxPaginate`, `hxMorphOuter`
- Idiomorph extension bundled for `hx-swap="morph"` (`hx-ext="morph"` on layout body)
- Supermarket-style scaffold layout with `#cais-nav`, `#cais-main`, and hx-boost navigation
- Resource generator HTMX admin CRUD (inline delete, form partials, `RenderPageOrPartial` on 422)
- HTMX pagination with morph swap for admin and public lists (`--paginate`)
- `float` / `float?` field type in `cais g resource`
- `cais.SetToast`, `SetFocus`, `SetRetarget`, `SetTrigger` response header helpers

#### Security & sessions

- Session expiry in SQLite (`expires_at`, 7-day TTL, reject expired `Get`)
- `session.PruneExpired()` and `cais db prune-sessions`
- Secure cookies in production (`Config.CookieSecure()`, `session.CookieOptionsFromConfig`)
- Security headers middleware (`middleware.SecurityHeaders`) — CSP, HSTS, X-Frame-Options, Referrer-Policy
- Per-IP rate limiting (`middleware.NewRateLimiter`) on login and contact POST routes
- Trusted proxy support (`TRUSTED_PROXIES`, `middleware.ClientIP`)
- Production error sanitization (`Config.SanitizeErrors()`, `httpx.RenderOrError`)
- CSRF protection (`middleware.CSRF`, `meta.WithCSRF`, double-submit cookie)
- Session auth scaffold (`cais g auth`, login/logout, protected dashboard)
- Flash messages (`pkg/cais/flash`, `middleware.Flash`, one-shot redirect feedback)

#### HTMX & UI

- `httpx.RenderPageOrPartial` for HTMX-aware form responses
- `cais.js` — CSRF header injection, focus restore, optimistic toggles
- HTMX swap/loading CSS utilities (`.htmx-swapping`, `.htmx-settling`, `.htmx-request-hide`)
- `cais.SetTrigger` and `cais.SetRetarget` response helpers
- Optimistic bool toggles in generated resource admin (`data-cais-optimistic="toggle"`)
- Contact form HTMX polish (loading indicator, swap transitions)

#### Forms & validation

- `validate.FieldErrors` map with `Add`, `Has`, `First`, `Any`
- `validate.MinLength` and `validate.MaxLength`
- `pkg/cais/forms` — `csrfField`, `fieldError`, `makeField`, `fieldInput` template helpers
- `forms.FieldData` for labeled inputs, textareas, and checkboxes with inline errors
- `fieldSelect`, `makeSelectField`, `makeSelectFieldPtr` for foreign-key dropdowns
- Resource generator `category_id:references` and `category:belongs_to` field types (FK migration, `ListCategoryOptions`, admin select)

#### Router & render

- `Router.Put` and `Router.Patch` methods
- Partials parsed into the page template tree (`{{ template "name" . }}` works in pages and layouts)
- `pkg/cais/pagination` — offset/limit helpers for list pages
- `pkg/cais/cache` — in-memory TTL cache (`New`, `Get`, `Set`, `Delete`)

#### i18n

- `pkg/cais/i18n` — key-based locale catalogs (`LOCALE=en` default, `LOCALE=pt` supported)
- Template funcs `t`, `htmlLang`, `ogLocale` registered on the renderer
- i18n wired through scaffolds, handlers, and `meta` OG locale defaults

#### Background jobs

- `pkg/cais/jobs` — SQLite-backed queue (enqueue, delay, worker, dispatcher)
- Recurring cron scheduler (`recurring_tasks`, `jobs.RunScheduler`)
- `cais jobs work [--queues ...] [--concurrency N]` and `cais jobs status`
- `cais g job <name> [--cron "0 3 * * *"]` — scaffolds handler, registry, and `cmd/worker`
- Built-in `PruneSessions` job handler
- Jobs migration (`003_jobs.sql`, `004_recurring_tasks.sql`)

#### CLI — generators

- `cais g resource` defaults to session auth (`--admin-auth session|bearer`)
- `cais g resource --paginate` — admin index pagination (25/page, HTMX controls)
- `cais g resource --force` — overwrite existing generated files
- Resource admin show page (handler, template, route, tests)
- `cais g model` — model struct + migration + store methods (no handlers/UI)
- `cais g --dry-run` on all generators (resource, model, handler, page, migration, auth, console, ci, job)
- `cais destroy [--dry-run]` — resource, handler, model, auth, migration (unpatch routes, store, seeds, nav)
- `cais g ci` — add GitHub Actions, pre-commit, golangci-lint, Prettier to existing apps
- `cais new --module <path>` — override Go module path
- Nav marker `<!-- cais:nav -->` for reliable public link patching
- AST-based route patching (`internal/cli/patch`) via `insertBeforeFunctionEnd`
- `nextMigrationFile` — sequential migration numbering across generators
- Migration `-- up` / `-- down` sections in generated SQL files
- Welcome screen and i18n catalogs in `cais new` scaffolds

#### CLI — database & tooling

- Versioned migrations (`pkg/cais/migrate`, `cais db migrate`, `cais db status`)
- `cais db rollback` — roll back last migration (runs `-- down` SQL when present)
- `cais db seed` and `cais db seed --list`
- `cais routes` and `cais routes --verbose` (handler names + middleware)
- `cais version` — print framework version
- `cais doctor` — production readiness checks (`ADMIN_TOKEN`, `APP_URL`, CI tooling hints)
- Smoke scaffold test (`scripts/smoke-scaffold.sh`, `generate_smoke_test.go`)
- App-level integration tests: login → dashboard (flash) → logout; contact validation with CSRF (`internal/app/app_test.go`)

#### Config & health

- `validate.Email`, `validate.URL`, `validate.Required`
- `APP_URL` required in production (`cfg.Validate()`)
- DB-aware `/health` endpoint (503 `degraded` when SQLite is down)
- SQLite production defaults (`PRAGMA foreign_keys=ON`, WAL)

### Changed

- Admin auth requires `ADMIN_TOKEN` in production; Bearer header only (no query params)
- Generated resource admin routes use `RequireAuth` by default instead of `AdminAuth`
- Reference app aligned with `cais new` output (routes.go, auth, dashboard, contact validation)
- Blank app scaffold includes `Recover`, `SecurityHeaders`, server timeouts, and `/health`
- Contact scaffold uses `validate.FieldErrors` with name validation
- Auth migration template includes `expires_at` column (7-day default)
- `pwa.FS()` returns `(fs.FS, error)` instead of panicking on failure
- README, AGENTS.md, and `.env.example` synced with all new commands and env vars
- Jobs documentation in README and AGENTS (`cais g job`, `cais jobs work/status`, deploy notes)
- CSP `'unsafe-inline'` tradeoff documented (required for HTMX and inline service-worker script)

### Security

- Removed admin token via query string
- Constant-time admin token comparison
- Rate limiter bucket cleanup for stale entries
- Flash cookies use `HttpOnly` and `Secure` in production

### Fixed

- `cais g auth` migration includes session expiry column on fresh installs
- Local `CAIS_REPLACE` resolves from cwd for remote app directories
- Generated `dashboard_test` scaffold includes `cais` import
- AST route patch preserves `cais.IntParam(...)` lines through gofmt
- CAIS startup banner renders block Unicode logo clearly (no longer reads as "COTS")

### Deprecated

- `middleware.TokenAuth` — use `AdminAuth(cfg)` instead; scheduled for removal in a future release

## [0.4.7] - 2026-07-01

### Added

- Default OG preview and fullscreen PWA (`pkg/cais/meta`, `pkg/cais/pwa`)
- Go on Cais hero image and harbor PNG icon in scaffold assets

### Fixed

- Startup banner redrawn to spell CAIS clearly with block Unicode art

## [0.4.0] - 2026-07-01

### Added

- Interactive console REPL (`cais console`, `pkg/cais/console`) with SQL, history, reload, and typed bindings
- `/logs` development log viewer (`pkg/cais/devlog`, HTMX auto-refresh, localhost only)
- Cais-branded air startup banner (`pkg/cais/boot`)
