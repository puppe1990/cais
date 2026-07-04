package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_PWA_Bump(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "pwa",
		ModulePath: "github.com/puppe1990/pwa",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	c := &CLI{Out: &bytes.Buffer{}}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := c.cmdPWA([]string{"--bump"}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "web/static/js/sw.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "CACHE_VERSION = 2") {
		t.Errorf("expected bumped CACHE_VERSION, got:\n%s", body)
	}
}
