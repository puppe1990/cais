# Framework Roadmap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Steps use checkbox (`- [x]` done, `- [ ]` pending) syntax.

**Goal:** Implement full audit remediation per `docs/superpowers/specs/2026-07-01-framework-roadmap-design.md`.

**Architecture:** Six PR phases, TDD mandatory.

**Tech Stack:** Go 1.22+, stdlib, existing Cais packages.

**Spec:** `docs/superpowers/specs/2026-07-01-framework-roadmap-design.md`

**Last updated:** 2026-07-01

---

## Implementation status

| Phase | Theme                      | Status                                                                               |
| ----- | -------------------------- | ------------------------------------------------------------------------------------ |
| 1     | Admin auth & scaffold sync | **Shipped** (blank app LoadSession/Flash still optional)                             |
| 2     | Security hardening         | **Shipped**                                                                          |
| 3     | Generator robustness       | **Mostly shipped** (AST patch exists, not wired to production)                       |
| 4     | Rails data parity          | **Shipped**                                                                          |
| 5     | Rails UI & DX              | **Shipped** (+ `FieldData`, `destroy`, `MinLength`/`MaxLength` beyond original spec) |
| 6     | Docs & coverage            | **Docs shipped**; coverage targets & `pwa.FS()` panic fix **pending**                |

**Also shipped (post-roadmap):** `cais destroy` (resource/handler/model/auth/migration), `cais g --dry-run` for console/ci, migration numbering via `nextMigrationFile`, `cais version`, `cais db seed --list`, `cais routes --verbose`, `g resource --force`.

---

## Phase 1 — Admin Auth & Scaffold Sync

### Task 1: Resource generator session auth (default)

- [x] Tests: `TestScaffoldResource_DefaultAdminAuthUsesRequireAuth`, `TestScaffoldResource_AdminAuthBearerFlag`
- [x] `--admin-auth session|bearer` flag (default: session)
- [x] `patchRoutesForResource` uses `RequireAuth` or `AdminAuth` by flag
- [x] CLI help text

### Task 2: Sync contact scaffold template

- [x] `tplContactHandler` uses `validate.FieldErrors` + name validation
- [x] Test asserts `errs.Add("name"` in generated contact handler

### Task 3: Sync blank app scaffold

- [x] Blank app: `Recover`, `SecurityHeaders`, server timeouts, `/health`
- [ ] Blank app: `LoadSession` + `Flash` (only added via `cais g auth`, not in `tplAppBlank` by default)

### Task 4: Auth migration expires_at

- [x] `cais g auth` migration includes `expires_at` with 7-day default
- [x] `TestScaffoldAuth_migrationIncludesExpiresAt`

### Task 5: Deprecate TokenAuth + README auth matrix

- [x] `TokenAuth` deprecation godoc
- [x] README / AGENTS admin auth matrix (session vs Bearer)

---

## Phase 2 — Security Hardening

### Task 6: TrustedProxies config + ClientIP

- [x] `TRUSTED_PROXIES` env parsing on `Config`
- [x] `middleware.ClientIP(r, cfg)` with CIDR + `X-Real-IP` fallback
- [x] Tests in `clientip_test.go`

### Task 7: Production error sanitization

- [x] `Config.SanitizeErrors()`
- [x] `httpx.RenderOrError` generic 500 in production

### Task 8: Cookie clear Secure + rate limit cleanup

- [x] `CookieSecure()` + `CookieOptionsFromConfig`
- [x] Rate limiter on login/contact routes
- [ ] Periodic rate-limiter bucket cleanup (low priority)

---

## Phase 3 — Generator Robustness

### Task 9: Route patch anchor

- [x] String-based `insertBeforeFunctionEnd` (production default)
- [ ] AST route patch in `internal/cli/patch/` wired safely (gofmt breaks `cais.IntParam` lines today)

### Task 10: Nav marker `<!-- cais:nav -->`

- [x] Marker in layout templates; `patchLayoutNav` prefers marker over `</nav>`
- [x] Test: public resource nav link after marker

### Task 11: Migration down sections

- [x] `-- up` / `-- down` in generated migrations
- [x] `migrate.RollbackLast` runs down SQL; CLI warns when missing

### Task 12: `cais new --module`

- [x] `--module <path>` flag + tests

### Task 13: `cais g --dry-run` / `cais destroy --dry-run`

- [x] All generators including `console`, `ci`, `auth`, `resource`, `model`, `migration`
- [x] `destroy` dry-run for all targets

### Task 13b: Migration numbering (follow-up)

- [x] `nextMigrationFile` (max+1) for resource, model, auth, `g migration`

### Task 13c: `cais destroy` (follow-up)

- [x] `destroy resource|handler|model|auth|migration`
- [x] Unpatch routes (admin group block), store interface, imports, seeds, layout nav

---

## Phase 4 — Rails Data Parity

### Task 14: Migration down SQL execution

- [x] `pkg/cais/migrate` rollback with `-- down` section
- [x] `cais db rollback` + warning when no down SQL

### Task 15: `cais db seed`

- [x] `internal/db/seeds.go` scaffold + `cais db seed`
- [x] `cais db seed --list`

### Task 16: `cais g model`

- [x] Model struct + migration + store methods (no handlers/UI)

### Task 17: Router Put/Patch

- [x] `Router.Put` / `Router.Patch` + tests

---

## Phase 5 — Rails UI Parity

### Task 18: Resource pagination `--paginate`

- [x] `pkg/cais/pagination` + `List{Resource}(page, perPage)` store methods
- [x] Admin index pagination partial in generator

### Task 19: `pkg/cais/forms` helpers

- [x] `csrfField`, `fieldError`
- [x] `forms.FieldData`, `makeField`, `fieldInput` (generator admin forms)
- [x] `validate.MinLength`, `validate.MaxLength`

### Task 20: `cais routes` command

- [x] `cais routes` + `--verbose` (handler + middleware)

### Task 21: `pkg/cais/cache` minimal API

- [x] `cache.New`, `Get`, `Set` + tests
- [ ] `Delete` + render integration (deferred)

---

## Phase 6 — Docs & Coverage

### Task 22: README + .env.example + AGENTS.md

- [x] Admin auth matrix, new CLI commands, form/validation examples
- [x] Destroy, dry-run, seed, routes, version documented
- [ ] CSP `'unsafe-inline'` tradeoff note in README

### Task 23: Test coverage gaps

- [ ] `pkg/cais/console` → 70%+
- [ ] `pkg/cais/devlog`, `pkg/cais/boot/version.go`
- [ ] `pkg/cais/pwa` — `FS()` return error instead of panic
- [ ] `internal/store` pagination/seed coverage

### Task 24: `cais doctor` production checks

- [x] Warns missing `ADMIN_TOKEN`, `APP_URL` in production
- [x] Quality tooling warning + `cais g ci` hint

---

## Out of scope (separate specs)

- Action Mailer / SMTP
- Background jobs / queue
- Password reset / registration
- Structured JSON production logging
- External docs site
- Nonce-based CSP (HTMX conflict)
- Accept-Language i18n v2 / `cais g locale`
- esbuild / JS bundling
- REST PUT/DELETE in generated admin (still POST)
- Resource show page, FK/associations in generator

---

## Verification (each phase)

```bash
go test ./... -race -count=1
make lint
make ci
```

Generator smoke:

```bash
cais new testapp && cd testapp && cais g resource item --fields name:string && make test
```
