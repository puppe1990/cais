package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCLI_GenerateConsoleDryRun(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "cliconsoledry")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "cliconsoledry",
		ModulePath: "github.com/puppe1990/cliconsoledry",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(appDir, "cmd/console/main.go")); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	c := &CLI{Out: io.Discard}
	if err := c.Run([]string{"g", "--dry-run", "console"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(appDir, "cmd/console/main.go")); !os.IsNotExist(err) {
		t.Error("CLI --dry-run should not create cmd/console/main.go")
	}
}

func TestCLI_GenerateCIDryRun(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "clicidry")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "clicidry",
		ModulePath: "github.com/puppe1990/clicidry",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".github/workflows/ci.yml", ".golangci.yml"} {
		if err := os.Remove(filepath.Join(appDir, path)); err != nil {
			t.Fatal(err)
		}
	}
	makefileBefore, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	c := &CLI{Out: io.Discard}
	if err := c.Run([]string{"g", "--dry-run", "ci"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(appDir, ".github/workflows/ci.yml")); !os.IsNotExist(err) {
		t.Error("CLI --dry-run should not create ci.yml")
	}
	makefileAfter, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	if string(makefileAfter) != string(makefileBefore) {
		t.Error("CLI --dry-run should not modify Makefile")
	}
}

func TestCLI_GenerateResourceDryRun(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "clidryrun")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "clidryrun",
		ModulePath: "github.com/puppe1990/clidryrun",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	c := &CLI{Out: io.Discard}
	if err := c.Run([]string{"g", "--dry-run", "resource", "post", "--fields", "title:string"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(appDir, "internal/models/post.go")); !os.IsNotExist(err) {
		t.Error("CLI --dry-run should not create post.go")
	}
}
