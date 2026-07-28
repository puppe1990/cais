# Go on Cais

![Go on Cais](pkg/cais/pwa/assets/go-on-cais.jpg)

Full-stack Go framework for mini apps (Lightsail-friendly): **Inertia.js + Svelte 5**, Tailwind, and SQLite — with a Rails-style CLI.

This repository is the **framework + CLI** only. Generate apps with `cais new`.

## Stack

| Layer     | Choice                                                                       |
| --------- | ---------------------------------------------------------------------------- |
| Language  | Go 1.26 (`net/http` stdlib; see `go.mod`)                                    |
| Frontend  | **Inertia.js + Svelte 5** (`@inertiajs/svelte` + Vite → `web/static/build/`) |
| CSS       | Tailwind CSS 3.x                                                             |
| DB        | SQLite (`modernc.org/sqlite`, no CGO)                                        |
| PWA       | Manifest, service worker, offline page, icons, fullscreen                    |
| Meta      | Open Graph / Twitter via `pkg/cais/meta`                                     |
| Legacy UI | HTMX helpers in `pkg/cais/` for `cais g stream chat` / non-Inertia resource  |

## Quick start

```bash
export PATH="$HOME/go/bin:$PATH"
make install-cli          # installs cais CLI
cais new myapp
cd myapp && cais install && cais dev   # http://localhost:8080
```

**Developing the framework itself:**

```bash
export PATH="$HOME/go/bin:$PATH"
make install-cli
make test                 # go test ./... -race
make ci                   # test + js-test + lint + format-check
# Local CLI against this checkout:
cais link .               # from an app dir, or set CAIS_REPLACE
```

Demo login in a fresh scaffold (dev seed): `demo@example.com` / `password`.

## CLI (Rails-style)

```bash
make install-cli
export PATH="$HOME/go/bin:$PATH"
```

| Command                                                                                        | Description                                                                              |
| ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `cais new <app> [dir] [--minimal\|--blank] [--module path]`                                    | Scaffold app (Inertia + Svelte by default)                                               |
| `cais g [--dry-run] handler\|page\|resource\|model\|migration\|auth\|console\|ci\|job\|stream` | Generators                                                                               |
| `cais destroy [--dry-run] resource\|handler\|model\|auth\|migration`                           | Undo generators                                                                          |
| `cais install`                                                                                 | `npm install` + `go mod tidy`                                                            |
| `cais dev`                                                                                     | **air + Tailwind watch + `vite build --watch`** (Svelte rebuilds on `web/src/**` change) |
| `cais css` / `cais build` / `cais server` / `cais test`                                        | CSS, binary, run, tests                                                                  |
| `cais console`                                                                                 | REPL (store, cfg, db + SQL)                                                              |
| `cais routes [--verbose]`                                                                      | List routes from `internal/app/routes.go`                                                |
| `cais db migrate\|status\|rollback\|prune-sessions\|seed`                                      | Migrations & seeds                                                                       |
| `cais jobs work\|status`                                                                       | SQLite background jobs                                                                   |
| `cais doctor [--mobile]`                                                                       | Verify Inertia/Vite, PWA, mobile SSE, SW cache strategy                                  |
| `cais pwa [--bump]`                                                                            | Write/refresh PWA assets; **migrates** `sw.js` to network-first SPA; `--bump` cache      |
| `cais link [path] [--unlink]`                                                                  | Local `go.mod replace` for framework dev                                                 |
| `cais version`                                                                                 | Framework version                                                                        |

Field types for generators: `string`, `text`, `url`, `bool`, `int`, `date`, `references` (or `name:belongs_to`). Suffix `?` for optional.

```bash
cais g resource bookmark --fields title:string,url:url,notes:text? --public --paginate
cais g handler settings   # Go handler + test + web/src/pages/Settings.svelte
```

## Development experience (in generated apps)

- **Port auto-pick** if `:8080` is busy
- **Boot banner** with LAN URLs for phone testing on Wi‑Fi
- **Logs** — JSON request (`kind: request`) + SQL (`kind: sql`); `LOG_FORMAT=text` for plain text
- **`/logs`** — localhost-only HTMX log viewer in development
- **Frontend** — `cais dev` rebuilds Vite assets when Svelte sources change
- **PWA** — SW is **network-first** for `/static/build/` and `/static/css/`; `cais pwa --bump` after HTML changes on phones

## Structure

```
pkg/cais/              framework packages (router, httpx, session, jobs, pwa, …)
internal/cli/          cais CLI + scaffold templates (split by domain)
cmd/cais/              CLI entry point
cmd/pwagen/            helper to write PWA assets into a directory
scripts/               smoke-scaffold + smoke-production (via cais new)
```

Scaffolded apps get `cmd/server`, `internal/app`, `internal/handlers`, `web/src`, etc. — not this repo.

## Inertia + Svelte (generated apps)

Handlers render components via gonertia:

```go
_ = h.inertia.Render(w, r, "Login", inertia.Props{
  "site": meta.ForRequest(h.site, r),
})
// Validation: inertia.SetValidationErrors → re-render same component
// Flash redirect: inertia.SetFlash + h.inertia.Redirect(..., 303)
```

Svelte pages use `useForm` as a **reactive object** (not a store — no `$form`):

```svelte
<script>
  import { useForm, router } from '@inertiajs/svelte'
  let form = useForm({ email: '', password: '' })
  function submit() { form.post('/login') }
  function logout() { router.post('/logout') }  // use:inertia is for GET only
</script>
```

**Svelte 5 footgun:** do not push Inertia props into the form via reactive statements (`$: form.item_id = items[0].id`). Prefer local state for derived UI and set form fields on submit.

**JSON bodies** — Inertia posts `application/json`. Handlers should use:

```go
if err := httpx.ParseFormOrJSON(r); err != nil { /* ... */ }
email := r.FormValue("email")
```

## Framework APIs (highlights)

**Router**

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
  g.Get("/dashboard", dashboard.ServeHTTP)
})
```

**httpx** — `RenderOrError`, `WritePage`, `SeeOther`, `ParseFormOrJSON`, `FormTruthy`, ETag helpers.

**Sessions** — cookie auth (7-day TTL), `session.SignIn` / `SignOut`, `cais db prune-sessions`.

**CSRF** — double-submit cookie `cais_csrf` + form field / `X-CSRF-Token`.

**Jobs** — SQLite queue, no Redis:

```bash
cais g job send_welcome --cron "0 3 * * *"
cais jobs work --concurrency 2
```

## Framework commands

```bash
make test           # go test ./... -race
make test-v         # verbose
make js-test        # pkg/cais/js unit tests
make lint           # golangci-lint
make format         # prettier --write
make ci             # test + js-test + lint + format-check
make build          # bin/cais
make install-cli    # go install ./cmd/cais
```

CI runs Go tests, JS unit tests, lint, Prettier, and smoke (`cais new` + production boot of a scaffolded app).

## Production deploy (generated apps)

```bash
npm run build   # Vite → web/static/build/
cais build --os linux --arch amd64 -o bin/server-linux
tar czf release.tar.gz bin/server-linux web/static
```

- Guide: `docs/deploy/lightsail-systemd.md`
- Template: `deploy/systemd/cais-app.service.example`

## License

See [LICENSE](LICENSE).
