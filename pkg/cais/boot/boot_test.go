package boot

import (
	"bytes"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestPrint_IncludesRailsStyleLines(t *testing.T) {
	var buf bytes.Buffer
	Print(&buf, Options{
		AppName: "PulseFit",
		Config: cais.Config{
			Port:   ":8080",
			DBPath: "./data/app.db",
			Env:    "development",
		},
		Version: "0.3.1",
	})

	out := buf.String()
	for _, want := range []string{
		"=> Booting PulseFit (Cais v0.3.1)",
		"=> Environment: development",
		"=> Database:    sqlite3 (./data/app.db)",
		"=> Listening on http://127.0.0.1:8080",
		"=> Ctrl-C to stop",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n got:\n%s", want, out)
		}
	}
}

func TestPrint_ShowsPortShiftLine(t *testing.T) {
	var buf bytes.Buffer
	Print(&buf, Options{
		AppName: "PulseFit",
		Config: cais.Config{
			Port:   ":8081",
			DBPath: "./data/app.db",
			Env:    "development",
		},
		Version:         "0.3.2",
		PortShiftedFrom: ":8080",
	})

	if !strings.Contains(buf.String(), "=> Port :8080 in use, using :8081") {
		t.Fatalf("got:\n%s", buf.String())
	}
}

func TestListenURL(t *testing.T) {
	if got := ListenURL(":3000"); got != "http://127.0.0.1:3000" {
		t.Fatalf("ListenURL(:3000) = %q", got)
	}
	if got := ListenURL("0.0.0.0:8080"); got != "http://0.0.0.0:8080" {
		t.Fatalf("ListenURL(0.0.0.0:8080) = %q", got)
	}
}
