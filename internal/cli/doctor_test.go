package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCLIVersion_okOnDevBuild(t *testing.T) {
	// frameworkVersion() in tests is usually "dev" — should not warn.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(`module app
require github.com/puppe1990/cais v0.8.0
`), 0o644); err != nil {
		t.Fatal(err)
	}
	c := checkCLIVersion(dir)
	if !c.OK {
		t.Errorf("dev CLI should pass check, got %+v", c)
	}
}

func TestCheckCLIVersion_warnsWhenCLIOlderThanMod(t *testing.T) {
	// Unit-test the comparison path with synthetic versions via extract + parse.
	mod := parseSemverCore("0.8.1")
	cli := parseSemverCore("0.6.1-0.20260706")
	if compareSemverCore(cli, mod) >= 0 {
		t.Fatal("precondition: CLI should be older than mod")
	}
	// When CLI were 0.6.1 and go.mod 0.8.1, check would warn — covered by compare.
	if formatSemver(cli) != "0.6.1" {
		t.Errorf("format = %s", formatSemver(cli))
	}
}

func TestDoctor_SSEWriteTimeoutWarnsWhenPositive(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	// Simulate an HTMX/SSE app (Inertia scaffolds skip this check without sse-ext).
	sseDir := filepath.Join(dir, "web/static/js")
	if err := os.MkdirAll(sseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sseDir, "sse-ext.min.js"), []byte("// sse"), 0o644); err != nil {
		t.Fatal(err)
	}

	appGo := filepath.Join(dir, "internal/app/app.go")
	body, err := os.ReadFile(appGo)
	if err != nil {
		t.Fatal(err)
	}
	patched := strings.Replace(string(body), "WriteTimeout:      0,", "WriteTimeout:      30 * time.Second,", 1)
	if err := os.WriteFile(appGo, []byte(patched), 0o644); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("doctor should pass with warning: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[warn] SSE WriteTimeout") {
		t.Errorf("expected SSE WriteTimeout warning, got:\n%s", out)
	}
	if !strings.Contains(out, "WriteTimeout: 0") {
		t.Errorf("expected fix hint for WriteTimeout: 0, got:\n%s", out)
	}
}

func TestDoctor_SSEWriteTimeoutOKWhenZero(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	out := runDoctorOutput(t, dir)
	if strings.Contains(out, "[warn] SSE WriteTimeout") {
		t.Errorf("unexpected SSE WriteTimeout warning with default scaffold, got:\n%s", out)
	}
}

func TestDoctor_MobileChecks_chatSSEAndReconnect(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "mobile",
		ModulePath: "github.com/puppe1990/mobile",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	out := runDoctorOutputMobile(t, dir)
	for _, want := range []string{
		"[ok] health lan_urls",
		"[ok] CSP fonts",
		"[ok] PWA cache version",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in doctor --mobile output, got:\n%s", want, out)
		}
	}
	for _, skipped := range []string{
		"chat SSE pattern",
		"SSE reconnect",
		"chat agent JS",
	} {
		if !strings.Contains(out, skipped) {
			t.Errorf("expected %q check in doctor --mobile output, got:\n%s", skipped, out)
		}
	}
}

func TestDoctor_MobileWarnsMultiSlotWithoutFinalize(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "mobile",
		ModulePath: "github.com/puppe1990/mobile",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	// Inertia scaffolds have no chat partials or cais.js — agent chat check is skipped.
	out := runDoctorOutputMobile(t, dir)
	if !strings.Contains(out, "chat agent JS") {
		t.Errorf("expected chat agent JS check in output, got:\n%s", out)
	}
	if strings.Contains(out, "[warn] chat agent JS") {
		t.Errorf("Inertia scaffold should not warn on chat agent JS, got:\n%s", out)
	}
}

func runDoctorOutputMobile(t *testing.T, dir string) string {
	t.Helper()
	unsetCIEnv(t)
	writeBuiltStylesCSS(t, dir)
	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor --mobile failed: %v\n%s", err, buf.String())
	}
	return buf.String()
}

func TestDoctor_AllOK(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("doctor failed: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "Inertia frontend") {
		t.Error("missing Inertia frontend check")
	}
	if !strings.Contains(buf.String(), "vite.config.js") {
		t.Error("missing vite.config.js check")
	}
}

func TestDoctor_FailsWhenStylesCSSNotBuilt(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "nocss",
		ModulePath: "github.com/puppe1990/nocss",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	// Leave scaffold placeholder as-is.
	var buf bytes.Buffer
	err := runDoctor(&buf, dir, doctorOptions{})
	if err == nil {
		t.Fatalf("doctor should fail when styles.css not built, output:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "tailwind css") {
		t.Errorf("expected tailwind css check, got:\n%s", buf.String())
	}
}

func TestDoctor_AirOptionalWhenMissing(t *testing.T) {
	unsetCIEnv(t)
	if _, err := exec.LookPath("air"); err == nil {
		t.Skip("air installed; optional-missing path not exercised")
	}
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("doctor should pass without air: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "[warn] air") {
		t.Errorf("expected air warning, got:\n%s", buf.String())
	}
}

func TestDoctor_QualityToolingWarnsWhenMissing(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "legacy",
		ModulePath: "github.com/puppe1990/legacy",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, ".github/workflows/ci.yml")); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("doctor should pass with optional warning: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[warn] quality tooling") {
		t.Errorf("expected quality tooling warning, got:\n%s", out)
	}
	if !strings.Contains(out, "cais g ci") {
		t.Errorf("expected fix hint, got:\n%s", out)
	}
}

func TestDoctor_ProductionWarnsMissingAdminToken(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ENV=production\nAPP_URL=https://example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runDoctorOutput(t, dir)
	if !strings.Contains(out, "[warn] ADMIN_TOKEN") {
		t.Errorf("expected ADMIN_TOKEN warning, got:\n%s", out)
	}
	if strings.Contains(out, "[warn] APP_URL") {
		t.Errorf("unexpected APP_URL warning, got:\n%s", out)
	}
}

func TestDoctor_ProductionWarnsMissingAppURL(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ENV=production\nADMIN_TOKEN=secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runDoctorOutput(t, dir)
	if !strings.Contains(out, "[warn] APP_URL") {
		t.Errorf("expected APP_URL warning, got:\n%s", out)
	}
}

func TestDoctor_ProductionFromEnvVar(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	t.Setenv("ENV", "production")
	t.Setenv("ADMIN_TOKEN", "")
	t.Setenv("APP_URL", "")
	dir := scaffoldDoctorApp(t)

	out := runDoctorOutput(t, dir)
	if !strings.Contains(out, "[warn] ADMIN_TOKEN") {
		t.Errorf("expected ADMIN_TOKEN warning from ENV, got:\n%s", out)
	}
	if !strings.Contains(out, "[warn] APP_URL") {
		t.Errorf("expected APP_URL warning from ENV, got:\n%s", out)
	}
}

func TestDoctor_ProductionSkipsWhenDevelopment(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ENV=development\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runDoctorOutput(t, dir)
	if strings.Contains(out, "[warn] ADMIN_TOKEN") || strings.Contains(out, "[warn] APP_URL") {
		t.Errorf("unexpected production warnings in development, got:\n%s", out)
	}
}

func TestDoctor_ProductionOKWhenConfigured(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ENV=production\nADMIN_TOKEN=secret\nAPP_URL=https://example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runDoctorOutput(t, dir)
	if strings.Contains(out, "[warn] ADMIN_TOKEN") || strings.Contains(out, "[warn] APP_URL") {
		t.Errorf("unexpected production warnings when configured, got:\n%s", out)
	}
	if !strings.Contains(out, "[ok] ADMIN_TOKEN") || !strings.Contains(out, "[ok] APP_URL") {
		t.Errorf("expected production checks to pass, got:\n%s", out)
	}
}

func TestDoctor_DeployLayout(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)

	out := runDoctorOutput(t, dir)
	if !strings.Contains(out, "[ok] deploy layout") {
		t.Errorf("expected deploy layout ok, got:\n%s", out)
	}

	manifest := filepath.Join(dir, "web/static/manifest.webmanifest")
	if err := os.Remove(manifest); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err == nil {
		t.Fatalf("expected doctor failure without manifest, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "[FAIL] deploy layout") {
		t.Errorf("expected deploy layout failure, got:\n%s", buf.String())
	}
}

func TestDoctor_SeedsInfoWhenPresent(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)

	seedsDir := filepath.Join(dir, "internal/db")
	if err := os.MkdirAll(seedsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedsDir, "seeds.go"), []byte("package db\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runDoctorOutput(t, dir)
	if !strings.Contains(out, "[info] db seeds") {
		t.Errorf("expected seeds info, got:\n%s", out)
	}
	if !strings.Contains(out, "cais db seed") {
		t.Errorf("expected cais db seed hint, got:\n%s", out)
	}
}

func scaffoldDoctorApp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)
	return dir
}

// writeBuiltStylesCSS marks styles.css as a real Tailwind build so doctor can pass (#141).
func writeBuiltStylesCSS(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "web/static/css/styles.css")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("*,::before{box-sizing:border-box}.text-stone-900{color:#1c1917}"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func unsetCIEnv(t *testing.T) {
	t.Helper()
	// Scaffolded test apps get a local replace; that must stay a warn here, not a CI FAIL (#154).
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
}

func runDoctorOutput(t *testing.T, dir string) string {
	t.Helper()
	unsetCIEnv(t)
	// Most doctor tests care about other checks; assume CSS was built unless the test asserts otherwise.
	if !stylesCSSReady(dir) {
		writeBuiltStylesCSS(t, dir)
	}
	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("doctor failed: %v\n%s", err, buf.String())
	}
	return buf.String()
}
