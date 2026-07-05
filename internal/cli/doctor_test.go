package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_SSEWriteTimeoutWarnsWhenPositive(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
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
		"[ok] chat SSE pattern",
		"[ok] SSE reconnect",
		"[ok] health lan_urls",
		"[ok] chat agent JS",
		"[ok] chat scroll container",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in doctor --mobile output, got:\n%s", want, out)
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
	jsPath := filepath.Join(dir, "web/static/js/cais.js")
	body, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatal(err)
	}
	stripped := strings.ReplaceAll(string(body), "finalizeChatStream", "")
	if err := os.WriteFile(jsPath, []byte(stripped), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runDoctorOutputMobile(t, dir)
	if !strings.Contains(out, "[warn] chat agent JS") {
		t.Errorf("expected chat agent JS warning when finalizeChatStream removed, got:\n%s", out)
	}
}

func runDoctorOutputMobile(t *testing.T, dir string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor --mobile failed: %v\n%s", err, buf.String())
	}
	return buf.String()
}

func TestDoctor_AllOK(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("doctor failed: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "htmx.min.js") {
		t.Error("missing htmx check")
	}
	if !strings.Contains(buf.String(), "sse-ext.min.js") {
		t.Error("missing sse extension check")
	}
}

func TestDoctor_AirOptionalWhenMissing(t *testing.T) {
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

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("doctor should pass without air: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "[warn] air") {
		t.Errorf("expected air warning, got:\n%s", buf.String())
	}
}

func TestDoctor_QualityToolingWarnsWhenMissing(t *testing.T) {
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
	return dir
}

func runDoctorOutput(t *testing.T, dir string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("doctor failed: %v\n%s", err, buf.String())
	}
	return buf.String()
}
