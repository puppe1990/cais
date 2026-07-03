# Framework Roadmap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Steps use checkbox (`- [x]` done, `- [ ]` pending) syntax.

**Goal:** Implement full audit remediation per `docs/superpowers/specs/2026-07-01-framework-roadmap-design.md`.

**Architecture:** Six PR phases, TDD mandatory.

**Tech Stack:** Go 1.22+, stdlib, existing Cais packages.

**Spec:** `docs/superpowers/specs/2026-07-01-framework-roadmap-design.md`

**Last updated:** 2026-07-03

---

## Implementation status

| Phase | Theme                      | Status                                                          |
| ----- | -------------------------- | --------------------------------------------------------------- |
| 1     | Admin auth & scaffold sync | **Shipped**                                                     |
| 2     | Security hardening         | **Shipped**                                                     |
| 3     | Generator robustness       | **Shipped**                                                     |
| 4     | Rails data parity          | **Shipped**                                                     |
| 5     | Rails UI & DX              | **Shipped** (+ `FieldData`, `destroy`, `MinLength`/`MaxLength`) |
| 6     | Docs & coverage            | **Shipped**; cache render integration **deferred**              |

**Roadmap complete** except deferred items listed under [Remaining](#remaining).

**Also shipped (post-roadmap):** `cais destroy`, `cais g --dry-run`, migration numbering via `nextMigrationFile`, `cais version`, `cais db seed --list`, `cais routes --verbose`, `g resource --force`, jobs queue (`cais jobs work`), mail + password reset/signup in `cais g auth`, structured JSON dev + production logs (`LOG_FORMAT`), CLI scaffold splits (PRs #39–#44), provenance comments (#46), blank app session middleware (#47), coverage push for console/devlog/store (#48), `boot/version.go` tests (#50).

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
- [x] Blank app: `LoadSession` + `Flash` + minimal `Sessions()` on store (#47)

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
- [x] Periodic rate-limiter bucket cleanup (`cleanupBuckets` on each request)

---

## Phase 3 — Generator Robustness

### Task 9: Route patch anchor

- [x] `insertBeforeFunctionEnd` delegates to `internal/cli/patch.InsertBeforeFuncEnd` (go/ast; source not reformatted so `cais.IntParam` lines stay intact)
- [x] Integration tests: nested `IntParam` in admin groups (`patch_integration_test.go`, `patch/ast_test.go`)

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

- [x] `cache.New`, `Get`, `Set`, `Delete` + tests
- [ ] Render-layer cache integration (deferred; no template fragment caching yet)

---

## Phase 6 — Docs & Coverage

### Task 22: README + .env.example + AGENTS.md

- [x] Admin auth matrix, new CLI commands, form/validation examples
- [x] Destroy, dry-run, seed, routes, version documented
- [x] CSP `'unsafe-inline'` tradeoff note in README
- [x] `LOG_FORMAT` / JSON logging documented (#49)

### Task 23: Test coverage gaps

- [x] `pkg/cais/console` → 84.5% (target 70%+)
- [x] `pkg/cais/devlog` → 94.8%
- [x] `pkg/cais/boot/version.go` → `versionFrom` 100% (#50)
- [x] `pkg/cais/pwa` — `FS()` returns `(fs.FS, error)` (no panic)
- [x] `internal/store` pagination/seed coverage → 76.9%

### Task 24: `cais doctor` production checks

- [x] Warns missing `ADMIN_TOKEN`, `APP_URL` in production
- [x] Quality tooling warning + `cais g ci` hint

---

## Remaining

Deferred or not part of the original six phases:

- [ ] Cache render integration (fragment caching in renderer)
- [ ] External docs site
- [ ] Nonce-based CSP (HTMX inline scripts conflict)
- [ ] Accept-Language i18n v2 / `cais g locale` (basic `LOCALE` catalog shipped)
- [ ] esbuild / JS bundling
- [ ] REST PUT/DELETE in generated admin (still POST)
- [ ] Resource show page in generator
- [ ] Richer FK/associations in generator (basic `references` / `belongs_to` field shipped)

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
