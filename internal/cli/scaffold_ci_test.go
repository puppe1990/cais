package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldCI_AddsMissingTooling(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "legacy")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "legacy",
		ModulePath: "github.com/puppe1990/legacy",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		".github/workflows/ci.yml",
		".pre-commit-config.yaml",
		".golangci.yml",
		".prettierrc.json",
		".prettierignore",
	} {
		if err := os.Remove(filepath.Join(appDir, path)); err != nil {
			t.Fatal(err)
		}
	}

	oldMakefile := `.PHONY: dev build test css css-watch

test:
	go test ./... -race -count=1
`
	if err := os.WriteFile(filepath.Join(appDir, "Makefile"), []byte(oldMakefile), 0o644); err != nil {
		t.Fatal(err)
	}

	oldPackageJSON := `{
  "private": true,
  "devDependencies": {
    "prettier": "^3.5.3",
    "tailwindcss": "^3.4.17"
  },
  "scripts": {
    "format": "prettier --write .",
    "format:check": "prettier --check ."
  }
}
`
	if err := os.WriteFile(filepath.Join(appDir, "package.json"), []byte(oldPackageJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldCI(appDir, scaffoldData{ModulePath: "github.com/puppe1990/legacy"}, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		".github/workflows/ci.yml",
		".pre-commit-config.yaml",
		".golangci.yml",
		".prettierrc.json",
		".prettierignore",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	makefile, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(makefile), "ci:") {
		t.Error("Makefile missing ci target after scaffoldCI")
	}

	pkg, err := os.ReadFile(filepath.Join(appDir, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(pkg), `"test"`) {
		t.Error("package.json missing test script after scaffoldCI")
	}

	golangci, err := os.ReadFile(filepath.Join(appDir, ".golangci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(golangci), "github.com/puppe1990/legacy") {
		t.Error(".golangci.yml missing module local-prefix")
	}
}

func TestScaffoldCI_DryRunWritesNothing(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "cidry")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "cidry",
		ModulePath: "github.com/puppe1990/cidry",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		".github/workflows/ci.yml",
		".pre-commit-config.yaml",
		".golangci.yml",
		".prettierrc.json",
		".prettierignore",
	} {
		if err := os.Remove(filepath.Join(appDir, path)); err != nil {
			t.Fatal(err)
		}
	}

	makefileBefore, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	pkgBefore, err := os.ReadFile(filepath.Join(appDir, "package.json"))
	if err != nil {
		t.Fatal(err)
	}

	if err := scaffoldCI(appDir, scaffoldData{ModulePath: "github.com/puppe1990/cidry"}, true); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		".github/workflows/ci.yml",
		".pre-commit-config.yaml",
		".golangci.yml",
		".prettierrc.json",
		".prettierignore",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); !os.IsNotExist(err) {
			t.Errorf("dry-run should not create %s", path)
		}
	}

	makefileAfter, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	if string(makefileAfter) != string(makefileBefore) {
		t.Error("dry-run should not modify Makefile")
	}

	pkgAfter, err := os.ReadFile(filepath.Join(appDir, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(pkgAfter) != string(pkgBefore) {
		t.Error("dry-run should not modify package.json")
	}
}

func TestScaffoldCI_Idempotent(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "ready")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "ready",
		ModulePath: "github.com/puppe1990/ready",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldCI(appDir, scaffoldData{ModulePath: "github.com/puppe1990/ready"}, false); err != nil {
		t.Fatal(err)
	}
	makefileBefore, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}

	if err := scaffoldCI(appDir, scaffoldData{ModulePath: "github.com/puppe1990/ready"}, false); err != nil {
		t.Fatal(err)
	}
	makefileAfter, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	if string(makefileBefore) != string(makefileAfter) {
		t.Error("second scaffoldCI run should not change Makefile")
	}
}
