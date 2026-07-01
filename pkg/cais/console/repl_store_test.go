package console

import (
	"bytes"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

type fakeStore struct{}

func (fakeStore) Ping() string { return "pong" }

func TestRepl_CallsStoreMethod(t *testing.T) {
	var buf bytes.Buffer
	r := New(Options{
		AppName: "TestApp",
		Config:  cais.Config{Env: "development"},
		Bindings: map[string]any{
			"store": &fakeStore{},
		},
		Out: &buf,
	})

	if err := r.EvalLine(`store.Ping()`); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "pong") {
		t.Fatalf("output = %q", buf.String())
	}
}
