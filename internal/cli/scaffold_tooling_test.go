package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldTooling_CITriggersMainAndMaster(t *testing.T) {
	// #147
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "ciapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "ciapp",
		ModulePath: "github.com/puppe1990/ciapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	ci, err := os.ReadFile(filepath.Join(appDir, ".github/workflows/ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(ci)
	if !strings.Contains(body, "main") || !strings.Contains(body, "master") {
		t.Errorf("ci.yml should trigger on main and master, got:\n%s", body)
	}
	if !strings.Contains(body, "branches: [main, master]") {
		t.Errorf("ci.yml missing branches: [main, master], got:\n%s", body)
	}
}

func TestScaffoldTooling_PrettierAllowsEmptyOrIgnoresWebSrc(t *testing.T) {
	// #146 — committing only web/src must not fail prettier hook
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "fmtapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "fmtapp",
		ModulePath: "github.com/puppe1990/fmtapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	pre, err := os.ReadFile(filepath.Join(appDir, ".pre-commit-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(pre)
	// Either no-error-on-unmatched-pattern, or pre-commit exclude of web/src so hook is skipped
	hasEmptyOK := strings.Contains(body, "no-error-on-unmatched-pattern")
	hasExcludeSrc := strings.Contains(body, "web/src")
	if !hasEmptyOK && !hasExcludeSrc {
		t.Error("pre-commit prettier must tolerate web/src-only commits (args or exclude)")
	}
}

func TestScaffoldTooling_PreCommitUsesGoimportsNotOnlyGofmt(t *testing.T) {
	// #148
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "impapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "impapp",
		ModulePath: "github.com/puppe1990/impapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	pre, err := os.ReadFile(filepath.Join(appDir, ".pre-commit-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(pre)
	if strings.Contains(body, "entry: go fmt ./...") {
		t.Error("pre-commit must not use bare go fmt (misses goimports local-prefixes)")
	}
	if !strings.Contains(body, "goimports") {
		t.Error("pre-commit must run goimports to match CI golangci formatters")
	}
}
