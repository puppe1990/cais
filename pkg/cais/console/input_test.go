package console

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRepl_readLine_scannerMode(t *testing.T) {
	var out bytes.Buffer
	r := New(Options{
		In:  strings.NewReader("hello\n"),
		Out: &out,
	})
	r.initInput()

	line, err := r.readLine()
	if err != nil {
		t.Fatalf("readLine() error = %v", err)
	}
	if line != "hello" {
		t.Fatalf("readLine() = %q, want %q", line, "hello")
	}
	if !strings.Contains(out.String(), ">> ") {
		t.Fatalf("output missing prompt, got %q", out.String())
	}
}

func TestRepl_readLine_scannerEOF(t *testing.T) {
	r := New(Options{
		In:  strings.NewReader(""),
		Out: io.Discard,
	})
	r.initInput()

	_, err := r.readLine()
	if err != io.EOF {
		t.Fatalf("readLine() error = %v, want io.EOF", err)
	}
}
