package console

import (
	"bytes"
	"strings"
	"testing"

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
	for _, want := range []string{"store", "help", "sql"} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("help missing %q, got:\n%s", want, buf.String())
		}
	}
}
