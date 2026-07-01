package console

import (
	"bytes"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestRepl_ReloadRefreshesBindings(t *testing.T) {
	var buf bytes.Buffer
	n := 0
	r := New(Options{
		AppName: "TestApp",
		Config:  cais.Config{Env: "development"},
		Bindings: map[string]any{
			"n": 0,
		},
		Reload: func() (map[string]any, error) {
			n++
			return map[string]any{"n": n}, nil
		},
		Out: &buf,
	})

	if err := r.EvalLine(`n`); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "0") {
		t.Fatalf("before reload: %q", buf.String())
	}

	buf.Reset()
	if err := r.HandleLine("reload"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Reloaded") {
		t.Fatalf("reload message: %q", buf.String())
	}

	buf.Reset()
	if err := r.EvalLine(`n`); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "1") {
		t.Fatalf("after reload: %q", buf.String())
	}
}

func TestRepl_HistoryCommand(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{AppName: "TestApp", Out: &buf})
	r.recordHistory("store.Ping()")
	r.recordHistory("sql SELECT 1")

	if err := r.HandleLine("history"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"1", "store.Ping()", "2", "sql SELECT 1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("history missing %q, got:\n%s", want, out)
		}
	}
}

func TestRepl_RecallHistoryCommand(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{
		AppName: "TestApp",
		Config:  cais.Config{Env: "development"},
		Bindings: map[string]any{
			"answer": 42,
		},
		Out: &buf,
	})
	r.recordHistory(`answer`)
	r.recordHistory("help")

	buf.Reset()
	if err := r.HandleLine("!1"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "42") {
		t.Fatalf("!1 output = %q", buf.String())
	}
}

func TestRepl_ExtraBindingCallsMethod(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{
		AppName: "TestApp",
		Config:  cais.Config{Env: "development"},
		Bindings: map[string]any{
			"svc": &fakeStore{},
		},
		Out: &buf,
	})

	if err := r.EvalLine(`svc.Ping()`); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "pong") {
		t.Fatalf("output = %q", buf.String())
	}
}
