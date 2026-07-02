package logentry

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestWrite_sqlEntry(t *testing.T) {
	var buf bytes.Buffer
	at := time.Date(2026, 7, 2, 15, 4, 5, 0, time.UTC)
	err := Write(&buf, Entry{
		Kind:       "sql",
		At:         at,
		Operation:  "User Create",
		Query:      "INSERT INTO users (email) VALUES (?)",
		Args:       []any{"demo@example.com"},
		DurationMS: 1.2,
	})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if got["kind"] != "sql" {
		t.Errorf("kind = %v", got["kind"])
	}
	if got["operation"] != "User Create" {
		t.Errorf("operation = %v", got["operation"])
	}
	if got["query"] != "INSERT INTO users (email) VALUES (?)" {
		t.Errorf("query = %v", got["query"])
	}
	args, ok := got["args"].([]any)
	if !ok || len(args) != 1 || args[0] != "demo@example.com" {
		t.Errorf("args = %v", got["args"])
	}
	if got["duration_ms"] != 1.2 {
		t.Errorf("duration_ms = %v", got["duration_ms"])
	}
}

func TestWrite_requestEntry(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, Entry{
		Kind:       "request",
		Phase:      "completed",
		Method:     "GET",
		Path:       "/login",
		Status:     200,
		Remote:     "127.0.0.1",
		DurationMS: 3.5,
		At:         time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}

	var got Entry
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Kind != "request" || got.Phase != "completed" || got.Status != 200 {
		t.Errorf("got %+v", got)
	}
}
