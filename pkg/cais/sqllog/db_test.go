package sqllog

import (
	"bytes"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestDB_LogsQueryInDevelopment(t *testing.T) {
	raw, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })

	if _, err := raw.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT)`); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	db := Wrap(raw, Config{Enabled: true, Writer: &buf})
	if _, err := db.Exec(`INSERT INTO users (email) VALUES (?)`, "demo@pulsefit.local"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "User Create") {
		t.Errorf("missing operation label, got:\n%s", out)
	}
	if !strings.Contains(out, `INSERT INTO users`) {
		t.Errorf("missing SQL, got:\n%s", out)
	}
	if !strings.Contains(out, `["demo@pulsefit.local"]`) {
		t.Errorf("missing args, got:\n%s", out)
	}
	if !strings.Contains(out, "User Create (") {
		t.Errorf("missing duration wrapper, got:\n%s", out)
	}
	if !regexp.MustCompile(` at \d{4}-\d{2}-\d{2} `).MatchString(out) {
		t.Errorf("missing timestamp, got:\n%s", out)
	}
}

func TestDB_SkipsLoggingWhenDisabled(t *testing.T) {
	raw, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })

	var buf bytes.Buffer
	db := Wrap(raw, Config{Enabled: false, Writer: &buf})
	if _, err := db.Exec("SELECT 1"); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no logs, got:\n%s", buf.String())
	}
}
