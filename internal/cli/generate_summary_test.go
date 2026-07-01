package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCLI_GenerateResource_printsNextSteps(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := t.TempDir()
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "genapp",
		ModulePath: "github.com/puppe1990/genapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(appDir); err != nil {
		t.Fatal(err)
	}

	if err := c.Run([]string{"g", "resource", "item", "--fields", "name:string", "--no-seed"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "cais db migrate") {
		t.Errorf("missing migrate hint: %q", out)
	}
	if !strings.Contains(out, "cais test") {
		t.Errorf("missing test hint: %q", out)
	}
}
