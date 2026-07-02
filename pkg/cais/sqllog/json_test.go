package sqllog

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestDB_LogsJSONWhenEnabled(t *testing.T) {
	raw, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })

	if _, err := raw.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT)`); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	db := Wrap(raw, Config{Enabled: true, JSON: true, Writer: &buf})
	if _, err := db.Exec(`INSERT INTO users (email) VALUES (?)`, "demo@example.com"); err != nil {
		t.Fatal(err)
	}

	line := strings.TrimSpace(buf.String())
	var got map[string]any
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, line)
	}
	if got["kind"] != "sql" {
		t.Errorf("kind = %v", got["kind"])
	}
	if got["operation"] != "User Create" {
		t.Errorf("operation = %v", got["operation"])
	}
	if !strings.Contains(got["query"].(string), "INSERT INTO users") {
		t.Errorf("query = %v", got["query"])
	}
}
