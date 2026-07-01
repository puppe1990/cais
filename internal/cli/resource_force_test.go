package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldResource_forceOverwrites(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "forceapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "forceapp",
		ModulePath: "github.com/puppe1990/forceapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	opts := resourceOpts{Fields: "name:string", Seed: false}
	if err := scaffoldResource(appDir, "item", opts); err != nil {
		t.Fatal(err)
	}
	opts.Force = true
	opts.Fields = "title:string"
	if err := scaffoldResource(appDir, "item", opts); err != nil {
		t.Fatal(err)
	}
	model, err := os.ReadFile(filepath.Join(appDir, "internal/models/item.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(model), "Title") {
		t.Error("expected model to be overwritten with Title field")
	}
}
