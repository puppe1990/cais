package console

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestOptions_outAndIn_defaults(t *testing.T) {
	var o Options
	if o.out() != os.Stdout {
		t.Fatal("out() should default to os.Stdout")
	}
	if o.in() != os.Stdin {
		t.Fatal("in() should default to os.Stdin")
	}
}

func TestOptions_outAndIn_custom(t *testing.T) {
	var inBuf, outBuf bytes.Buffer
	o := Options{In: &inBuf, Out: &outBuf}
	if o.out() != &outBuf {
		t.Fatal("out() should use custom writer")
	}
	if o.in() != &inBuf {
		t.Fatal("in() should use custom reader")
	}
}

func TestRepl_printHistory_listsEntries(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{Out: &buf})
	r.recordHistory(`answer`)
	r.recordHistory(`cfg.DBPath`)

	if err := r.HandleLine("history"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "1  answer") || !strings.Contains(out, "2  cfg.DBPath") {
		t.Fatalf("history output:\n%s", out)
	}
}

func TestRepl_printHistory_empty(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{Out: &buf})
	r.printHistory()
	if !strings.Contains(buf.String(), "(empty)") {
		t.Fatalf("output = %q", buf.String())
	}
}

func TestRepl_reload_refreshesBindings(t *testing.T) {
	var buf bytes.Buffer
	calls := 0
	r := New(Options{
		Out: &buf,
		Bindings: map[string]any{
			"count": 1,
		},
		Reload: func() (map[string]any, error) {
			calls++
			return map[string]any{"count": 2}, nil
		},
	})
	if err := r.HandleLine("reload"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("reload calls = %d", calls)
	}
	if r.opts.Bindings["count"] != 2 {
		t.Fatalf("bindings = %#v", r.opts.Bindings)
	}
	if !strings.Contains(buf.String(), "Reloaded!") {
		t.Fatalf("output = %q", buf.String())
	}
}

func TestRepl_reload_requiresCallback(t *testing.T) {
	r := New(Options{Out: io.Discard})
	if err := r.HandleLine("reload"); err == nil {
		t.Fatal("expected error when Reload is nil")
	}
}

func TestRepl_EvalLine_syntaxError(t *testing.T) {
	r := New(Options{Out: io.Discard})
	if err := r.EvalLine("{"); err == nil {
		t.Fatal("expected syntax error")
	}
}
