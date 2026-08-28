package cli

// tplAgentsMD is the LLM/agent conventions file shipped with every cais new app.
// Keep it app-scoped (not framework internals). Agents load this first.
const tplAgentsMD = "# {{.AppName}} — AI Conventions\n\n" +
	"Primary reader is often an LLM agent. Prefer small greps, small modules, and headless tests.\n\n" +
	"## Rule #1: TDD is mandatory\n\n" +
	"Before writing production code:\n\n" +
	"1. Write the test (`*_test.go`, or `web/src/pages/*.test.js` for Svelte)\n" +
	"2. Run focused: `go test ./... -v -run TestName` / `npm run test:fe`\n" +
	"3. Confirm it **fails** for the right reason\n" +
	"4. Write the **minimal** code to pass\n" +
	"5. Run `cais test` (and frontend tests when JS changed)\n\n" +
	"## Clean code for agents\n\n" +
	"| Priority | Rule |\n" +
	"| -------- | ---- |\n" +
	"| 1 | **Small units** — functions ~4–20 lines; files target 200–300 lines, hard cap ~500 |\n" +
	"| 2 | **SRP** — one reason to change per file/package |\n" +
	"| 3 | **Greppable names** — unique domain nouns; avoid `data`, `handler`, `Manager`, `util` as primary names |\n" +
	"| 4 | **Comments = WHY** — security, SQLite, CSRF/cookie, Inertia JSON vs form. No narrating WHAT |\n" +
	"| 5 | **Inject deps** — handlers take `Store`, Inertia, `cais.Config` via constructor |\n" +
	"| 6 | **Early returns** — max ~2 nesting levels |\n" +
	"| 7 | **Errors with values** — `fmt.Errorf(\"...: %w\", err)` |\n" +
	"| 8 | **Headless tests** — SQLite `:memory:`; no manual seed for unit tests |\n\n" +
	"## Layout\n\n" +
	"| Path | Responsibility |\n" +
	"| ---- | -------------- |\n" +
	"| `cmd/server/` | Entry point |\n" +
	"| `internal/app/` | Bootstrap, `registerRoutes` |\n" +
	"| `internal/handlers/` | HTTP handlers (Inertia + Svelte) |\n" +
	"| `internal/store/` | SQLite + migrations |\n" +
	"| `internal/models/` | Domain structs |\n" +
	"| `web/templates/app.html` | Inertia root shell |\n" +
	"| `web/src/pages/` | Svelte pages |\n" +
	"| `web/src/components/` | Shared Svelte components |\n" +
	"| `web/static/` | CSS, Vite build, PWA |\n\n" +
	"Patch markers (do not remove): `registerRoutes`, `Close() error`, `<!-- cais:nav -->`.\n\n" +
	"## Inertia + Svelte\n\n" +
	"Handlers render **Inertia + Svelte** only:\n\n" +
	"```go\n" +
	"_ = h.inertia.Render(w, r, \"Contact\", inertia.Props{\"site\": meta.ForRequest(h.site, r)})\n\n" +
	"// Validation — re-render same component\n" +
	"ve := make(inertia.ValidationErrors)\n" +
	"ve[\"email\"] = \"Invalid email\"\n" +
	"ctx := inertia.SetValidationErrors(r.Context(), ve)\n" +
	"_ = h.inertia.Render(w, r.WithContext(ctx), \"Contact\", inertia.Props{})\n\n" +
	"// Flash on redirect — cais cookie API only (not inertia.SetFlash)\n" +
	"flash.Set(w, \"notice\", \"Saved!\", cfg.CookieSecure())\n" +
	"h.inertia.Redirect(w, r, \"/dashboard\", http.StatusSeeOther)\n" +
	"// Next request: middleware.Flash → flash.MessageFromRequest → props[\"flash\"]\n" +
	"```\n\n" +
	"Svelte pages (`@inertiajs/svelte` + Svelte 5):\n\n" +
	"- `useForm` returns a **reactive object**, not a store — no `$form`\n" +
	"- Do **not** reactive-write props into the form (`$: form.x = prop`) — can blank the page\n" +
	"- Prefer local state for derived UI; assign into form on submit\n" +
	"- Password fields: `PasswordInput.svelte`\n" +
	"- Mutations: `form.post` / `router.post` (`use:inertia` is for GET only)\n\n" +
	"Parse bodies with `httpx.ParseFormOrJSON` (Inertia posts JSON).\n\n" +
	"## Auth, CSRF, flash\n\n" +
	"- Session middleware: `LoadSession` + `Flash` + `CSRF(cfg)`\n" +
	"- Protect routes: `middleware.RequireAuth(\"/login\")` / `RequireAuthFunc`\n" +
	"- CSRF: double-submit cookie `cais_csrf` + form field or `X-CSRF-Token`\n" +
	"- Flash: **only** `flash.Set` + read via `flash.MessageFromRequest` into Inertia props\n" +
	"- Dev demo user (when seeded): `demo@example.com` / `password`\n\n" +
	"## New page / resource\n\n" +
	"```bash\n" +
	"cais g handler settings     # handler + test + web/src/pages/Settings.svelte + route\n" +
	"cais g page about           # Svelte page only\n" +
	"cais g resource bookmark --fields title:string,url:url,notes:text?\n" +
	"cais g model tag --fields name:string\n" +
	"cais g migration add_notes\n" +
	"cais g auth                 # if app was --blank/--minimal\n" +
	"cais db migrate\n" +
	"```\n\n" +
	"Or by hand:\n\n" +
	"1. Go test in `internal/handlers/`\n" +
	"2. Optional Vitest in `web/src/pages/*.test.js`\n" +
	"3. Svelte page + handler + route in `internal/app/routes.go`\n\n" +
	"## Commands\n\n" +
	"```bash\n" +
	"cais install          # npm + go mod tidy (+ Tailwind build)\n" +
	"cais dev              # air + tailwind + vite watch\n" +
	"cais test             # go test ./...\n" +
	"npm run test:fe       # Vitest (Svelte)\n" +
	"make ci               # test + lint + format-check\n" +
	"cais doctor [--mobile]\n" +
	"cais routes\n" +
	"cais db migrate | status | rollback | seed\n" +
	"cais jobs work | status\n" +
	"```\n\n" +
	"`GET /jobs` — localhost queue dashboard (heartbeats, retry/discard, prune, `?kind=`). Production: SSH tunnel.\n\n" +
	"## Do not\n\n" +
	"- Parse templates per request\n" +
	"- Use inline CSS (Tailwind classes)\n" +
	"- Mock the database (use SQLite `:memory:`)\n" +
	"- Grow files past ~500 lines without splitting\n" +
	"- Ship features without a headless test\n" +
	"- Use `$form` store syntax or Svelte 4 `new App()`\n" +
	"- Call `inertia.SetFlash` (no-op without FlashDataProvider — use `flash.Set`)\n" +
	"- Reactive-assign Inertia props into `useForm` fields\n"
