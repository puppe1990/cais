package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatchGoModReplace_CaisAppsLayout(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	caisDir := filepath.Join(root, "Cais")
	appsDir := filepath.Join(root, "Cais-apps", "demo")
	for _, d := range []string{caisDir, filepath.Dir(appsDir)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(caisDir, "go.mod"), []byte("module github.com/puppe1990/cais\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldNewApp(appsDir, scaffoldData{
		AppName:    "demo",
		ModulePath: "github.com/puppe1990/demo",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	mod, err := os.ReadFile(filepath.Join(appsDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "replace github.com/puppe1990/cais => ../../Cais") {
		t.Errorf("go.mod missing sibling Cais replace: %s", mod)
	}
}

func TestPatchGoModReplace_RemoteAppDirFromCwd(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	caisDir := filepath.Join(root, "Cais")
	appsDir := filepath.Join(root, "Cais-apps")
	appDir := filepath.Join(root, "remote", "testapp")
	for _, d := range []string{caisDir, appsDir, filepath.Dir(appDir)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(caisDir, "go.mod"), []byte("module github.com/puppe1990/cais\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(appsDir)
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "testapp",
		ModulePath: "github.com/puppe1990/testapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	mod, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	wantRel := filepath.Join("..", "..", "Cais")
	want := "replace github.com/puppe1990/cais => " + wantRel
	if !strings.Contains(string(mod), want) {
		t.Errorf("go.mod missing Cais replace from cwd layout:\nwant substring %q\ngot:\n%s", want, mod)
	}
}

func TestPatchGoModReplace(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	parent := t.TempDir()
	appDir := filepath.Join(parent, "demo")
	caisDir := filepath.Join(parent, "Cais")
	if err := os.MkdirAll(caisDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(caisDir, "go.mod"), []byte("module github.com/puppe1990/cais\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "demo",
		ModulePath: "github.com/puppe1990/demo",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "replace github.com/puppe1990/cais => ../Cais") {
		t.Errorf("go.mod missing replace: %s", body)
	}
}
