# Framework Roadmap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Implement full audit remediation per `docs/superpowers/specs/2026-07-01-framework-roadmap-design.md`.

**Architecture:** Six PR phases, TDD mandatory, work in `.worktrees/framework-roadmap` on branch `feat/framework-roadmap`.

**Tech Stack:** Go 1.22+, stdlib, existing Cais packages.

**Spec:** `docs/superpowers/specs/2026-07-01-framework-roadmap-design.md`

---

## Phase 1 — Admin Auth & Scaffold Sync

### Task 1: Resource generator session auth (default)

**Files:**

- Modify: `internal/cli/resource_gen.go`
- Modify: `internal/cli/resource.go` (`patchRoutesForResource`)
- Modify: `internal/cli/cli.go` (help text)
- Modify: `internal/cli/cli_test.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/cli/cli_test.go`:

```go
func TestScaffoldResource_DefaultAdminAuthUsesRequireAuth(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "items")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "items", ModulePath: "github.com/puppe1990/items",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "item", resourceOpts{Fields: "name:string"}); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	s := string(body)
	if !strings.Contains(s, `middleware.RequireAuth("/login")`) {
		t.Errorf("routes should use RequireAuth for session admin: %s", s)
	}
	if strings.Contains(s, "middleware.AdminAuth(cfg)") {
		t.Error("default should not use AdminAuth")
	}
}

func TestScaffoldResource_AdminAuthBearerFlag(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "apiitems")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "apiitems", ModulePath: "github.com/puppe1990/apiitems",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "item", resourceOpts{
		Fields: "name:string", AdminAuth: "bearer",
	}); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if !strings.Contains(string(body), "middleware.AdminAuth(cfg)") {
		t.Error("bearer flag should use AdminAuth")
	}
}
```

Update existing `TestScaffoldResource` assertion from `AdminAuth` to `RequireAuth`.

- [ ] **Step 2: Run tests — expect FAIL**

`go test ./internal/cli/... -v -run 'ScaffoldResource.*AdminAuth|TestScaffoldResource$'`

- [ ] **Step 3: Implement**

In `resource_gen.go`:

```go
type resourceOpts struct {
	Fields    string
	Public    bool
	Seed      bool
	AdminAuth string // "session" (default) or "bearer"
}
```

In `parseResourceOpts`, default `AdminAuth: "session"`, add `--admin-auth` flag validating `session` or `bearer`.

In `patchRoutesForResource`, branch on `data.AdminAuth` (pass via scaffoldData or resourceOpts embedded):

Session (default):

```go
fmt.Fprintf(&insert, "\tr.Group(middleware.RequireAuth(\"/login\"), func(g *cais.Router) {\n")
```

Bearer:

```go
fmt.Fprintf(&insert, "\tr.Group(middleware.AdminAuth(cfg), func(g *cais.Router) {\n")
```

Add `AdminAuth string` to `scaffoldData` or pass opts into patch function.

Update `cli.go` help: `[--admin-auth session|bearer]` default session.

- [ ] **Step 4: Run tests — expect PASS**

`go test ./internal/cli/... -v -run ScaffoldResource`

- [ ] **Step 5: Commit**

`git commit -m "feat(cli): default resource admin to session auth"`

---

### Task 2: Sync contact scaffold template

**Files:** `internal/cli/templates.go`, `internal/cli/scaffold_test.go` or `cli_test.go`

- [ ] Add name `FieldErrors` validation to `tplContactHandler` (match `internal/handlers/contact.go`)
- [ ] Add test asserting generated contact handler contains `errs.Add("name"`
- [ ] TDD: test fail → fix template → pass → commit `feat(cli): sync contact scaffold validation`

---

### Task 3: Sync blank app scaffold

**Files:** `internal/cli/templates.go` (`tplAppBlank`)

- [ ] Test: blank app template contains `middleware.Recover`, `SecurityHeaders`, `LoadSession`, `Flash`, `ReadTimeout`
- [ ] Update `tplAppBlank` to mirror `internal/app/app.go` middleware stack
- [ ] Commit `feat(cli): harden blank app scaffold middleware`

---

### Task 4: Auth migration expires_at

**Files:** `internal/cli/scaffold_auth.go` (`tplMigration002Auth`)

- [ ] Test: `cais g auth` migration includes `expires_at`
- [ ] Update migration template
- [ ] Commit `feat(cli): add expires_at to auth migration template`

---

### Task 5: Deprecate TokenAuth + README auth matrix

**Files:** `pkg/cais/middleware/auth.go`, `README.md`

- [ ] Add deprecation godoc on `TokenAuth`
- [ ] README section: Bearer vs session admin table
- [ ] Commit `docs: admin auth matrix and TokenAuth deprecation`

---

## Phase 2 — Security Hardening

### Task 6: TrustedProxies config + ClientIP

**Files:** `pkg/cais/config.go`, `pkg/cais/middleware/clientip.go` (new), `logger.go`, `ratelimit.go`, tests

- [ ] TDD `TRUSTED_PROXIES` env parsing
- [ ] `ClientIP(r, cfg)` — trust XFF only from trusted RemoteAddr
- [ ] Commit `feat: trusted proxy client IP`

### Task 7: Production error sanitization

**Files:** `pkg/cais/config.go`, `pkg/cais/httpx/httpx.go`, handlers

- [ ] `SanitizeErrors()` on Config
- [ ] `RenderOrError` accepts cfg, generic 500 in production
- [ ] Commit `feat(httpx): sanitize errors in production`

### Task 8: Cookie clear Secure + rate limit cleanup

**Files:** `session/cookie.go`, `middleware/ratelimit.go`, `boot/banner.go`

- [ ] TDD each fix → commit `feat: cookie and ratelimit hygiene`

---

## Phase 3 — Generator Robustness

### Task 9: Route patch anchor

### Task 10: Nav marker `<!-- cais:nav -->`

### Task 11: Migration down sections

### Task 12: `cais new --module`

### Task 13: `cais g --dry-run`

---

## Phase 4 — Rails Data Parity

### Task 14: Migration down SQL execution

### Task 15: `cais db seed`

### Task 16: `cais g model`

### Task 17: Router Put/Patch

---

## Phase 5 — Rails UI Parity

### Task 18: Resource pagination `--paginate`

### Task 19: `pkg/cais/forms` helpers

### Task 20: `cais routes` command

### Task 21: `pkg/cais/cache` minimal API

---

## Phase 6 — Docs & Coverage

### Task 22: README + .env.example + AGENTS.md

### Task 23: Test coverage gaps (console, devlog, boot, pwa)

### Task 24: `cais doctor` production checks

---

## Verification (each phase)

`go test ./... -race -count=1 && make lint && make ci`
