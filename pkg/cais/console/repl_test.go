package console

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestRepl_EvaluatesBindingExpression(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{
		AppName: "TestApp",
		Config:  cais.Config{Env: "development", DBPath: ":memory:"},
		Bindings: map[string]any{
			"answer": 42,
		},
		Out: &buf,
	})

	if err := r.EvalLine(`answer`); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "42") {
		t.Fatalf("output = %q, want 42", buf.String())
	}
}

func TestRepl_HelpCommand(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{AppName: "PulseFit", Out: &buf})
	if err := r.HandleLine("help"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"cfg", "help", "sql", "reload", "history"} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("help missing %q, got:\n%s", want, buf.String())
		}
	}
}

func TestBindingNames_ordersStoreAndDBFirst(t *testing.T) {
	names := bindingNames(map[string]any{
		"zebra": 1,
		"store": 2,
		"db":    3,
	})
	if len(names) != 3 {
		t.Fatalf("names = %#v", names)
	}
	if names[0] != "store" || names[1] != "db" {
		t.Fatalf("names = %#v, want store and db first", names)
	}
}

func TestRepl_printValue_types(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{Out: &buf})

	r.printValue(nil)
	if !strings.Contains(buf.String(), "nil") {
		t.Fatalf("nil output = %q", buf.String())
	}

	buf.Reset()
	r.printValue("hello")
	if !strings.Contains(buf.String(), `"hello"`) {
		t.Fatalf("string output = %q", buf.String())
	}

	buf.Reset()
	r.printValue(struct{ Name string }{Name: "Cais"})
	if !strings.Contains(buf.String(), "Cais") {
		t.Fatalf("struct output = %q", buf.String())
	}
}

func TestRepl_runSQL(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE nums (n INTEGER); INSERT INTO nums VALUES (7)`); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	r := New(Options{
		Bindings: map[string]any{"db": db},
		Out:      &buf,
	})
	if err := r.HandleLine("sql SELECT n FROM nums"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "7") {
		t.Fatalf("sql output = %q", buf.String())
	}
}

func TestRepl_runSQL_missingDB(t *testing.T) {
	r := New(Options{Out: bytes.NewBuffer(nil)})
	if err := r.HandleLine("sql SELECT 1"); err == nil {
		t.Fatal("expected error when db binding missing")
	}
}

func TestRepl_recallHistory_doubleBang(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{
		Config: cais.Config{Env: "development"},
		Bindings: map[string]any{
			"answer": 99,
		},
		Out: &buf,
	})
	r.recordHistory(`answer`)

	buf.Reset()
	if err := r.HandleLine("!!"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "99") {
		t.Fatalf("!! output = %q", buf.String())
	}
}

func TestRepl_recallHistory_errors(t *testing.T) {
	r := New(Options{Out: bytes.NewBuffer(nil)})

	if err := r.HandleLine("!!"); err == nil {
		t.Fatal("expected error for !! with empty history")
	}
	if err := r.HandleLine("!bad"); err == nil {
		t.Fatal("expected error for invalid recall syntax")
	}
}

func TestRepl_shouldRecord_skipsMetaCommands(t *testing.T) {
	r := New(Options{Out: bytes.NewBuffer(nil)})
	for _, line := range []string{"help", "history", "exit", "quit", "reload", "!1", ""} {
		if r.shouldRecord(line) {
			t.Fatalf("shouldRecord(%q) = true, want false", line)
		}
	}
	if !r.shouldRecord(`store.Ping()`) {
		t.Fatal("shouldRecord expression = false, want true")
	}
}

func TestRepl_Loop_exitCommand(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{
		AppName: "TestApp",
		Config:  cais.Config{Env: "development"},
		In:      strings.NewReader("exit\n"),
		Out:     &buf,
	})
	if err := r.Loop(); err != nil {
		t.Fatalf("Loop() error = %v", err)
	}
	if !strings.Contains(buf.String(), "Bye!") {
		t.Fatalf("output = %q, want goodbye", buf.String())
	}
}

func TestRun_exitsCleanly(t *testing.T) {
	var buf bytes.Buffer
	err := Run(Options{
		AppName: "TestApp",
		Config:  cais.Config{Env: "development"},
		In:      strings.NewReader("quit\n"),
		Out:     &buf,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRepl_exitCommand(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{Out: &buf})
	if err := r.HandleLine("exit"); err != errExit {
		t.Fatalf("HandleLine(exit) error = %v, want errExit", err)
	}
}

func TestRepl_helpNotRecordedInHistory(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{
		AppName: "TestApp",
		In:      strings.NewReader("help\nexit\n"),
		Out:     &buf,
	})
	if err := r.Loop(); err != nil {
		t.Fatal(err)
	}
	if len(r.history.List()) != 0 {
		t.Fatalf("history = %#v, want empty (help/exit not recorded)", r.history.List())
	}
}

func TestRepl_closeInput_noPanic(t *testing.T) {
	r := New(Options{
		In:  strings.NewReader(""),
		Out: bytes.NewBuffer(nil),
	})
	r.initInput()
	r.closeInput()
}
