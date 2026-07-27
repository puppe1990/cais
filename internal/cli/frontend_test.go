package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMinimalCaisApp(t *testing.T, dir string, withVite bool) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testapp\nrequire github.com/puppe1990/cais v0.3.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "input.css"), []byte("@tailwind base;"), 0o644); err != nil {
		t.Fatal(err)
	}
	if withVite {
		if err := os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte("// vite"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"build":"node -e \"require('fs').mkdirSync('web/static/build',{recursive:true});require('fs').writeFileSync('web/static/build/.built','ok')\""}}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestHasViteApp_falseWithoutConfig(t *testing.T) {
	dir := t.TempDir()
	writeMinimalCaisApp(t, dir, false)
	if hasViteApp(dir) {
		t.Fatal("expected false without vite.config.js")
	}
}

func TestHasViteApp_trueWithConfigAndBuildScript(t *testing.T) {
	dir := t.TempDir()
	writeMinimalCaisApp(t, dir, true)
	if !hasViteApp(dir) {
		t.Fatal("expected true with vite.config.js and build script")
	}
}

func TestRunViteBuild_skipsWithoutVite(t *testing.T) {
	dir := t.TempDir()
	writeMinimalCaisApp(t, dir, false)
	if err := runViteBuild(dir); err != nil {
		t.Fatalf("expected nil skip, got %v", err)
	}
}

func TestRunViteBuild_runsNpmBuildWhenVitePresent(t *testing.T) {
	dir := t.TempDir()
	writeMinimalCaisApp(t, dir, true)
	if err := runViteBuild(dir); err != nil {
		t.Fatalf("runViteBuild: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/static/build/.built")); err != nil {
		t.Fatalf("expected build output: %v", err)
	}
}

func TestViteWatchArgs_usesNpmBuildWatch(t *testing.T) {
	args := viteWatchArgs()
	want := []string{"run", "build", "--", "--watch"}
	if len(args) != len(want) {
		t.Fatalf("viteWatchArgs = %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("viteWatchArgs = %v, want %v", args, want)
		}
	}
}

func TestStartViteWatch_skipsWithoutVite(t *testing.T) {
	dir := t.TempDir()
	writeMinimalCaisApp(t, dir, false)
	cmd, err := startViteWatch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Fatal("expected nil cmd without vite app")
	}
}

func TestStartViteWatch_errorsWithoutNodeModules(t *testing.T) {
	dir := t.TempDir()
	writeMinimalCaisApp(t, dir, true)
	_, err := startViteWatch(dir)
	if err == nil {
		t.Fatal("expected error when node_modules missing")
	}
}
