package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestDoctor_AllOK(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir); err != nil {
		t.Fatalf("doctor failed: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "htmx.min.js") {
		t.Error("missing htmx check")
	}
}
