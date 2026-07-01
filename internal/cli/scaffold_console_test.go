package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldConsole_DryRunWritesNothing(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "noconsole")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "noconsole",
		ModulePath: "github.com/puppe1990/noconsole",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	consolePath := filepath.Join(appDir, "cmd/console/main.go")
	if err := os.Remove(consolePath); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldConsole(appDir, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(consolePath); !os.IsNotExist(err) {
		t.Error("dry-run should not create cmd/console/main.go")
	}
}
