package cais

import (
	"embed"
	"io/fs"
	"strings"
	"testing"
)

//go:embed testdata/templates/*
var testTemplates embed.FS

func testRenderer(t *testing.T) *Renderer {
	t.Helper()
	tmplFS, err := fs.Sub(testTemplates, "testdata/templates")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRenderer(tmplFS)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRender_Page(t *testing.T) {
	r := testRenderer(t)

	var buf strings.Builder
	err := r.Render(&buf, "base", "home", map[string]string{"Name": "World"})
	if err != nil {
		t.Fatal(err)
	}

	body := buf.String()
	if !strings.Contains(body, "Hello, World!") {
		t.Errorf("body missing greeting, got: %s", body)
	}
	if !strings.Contains(body, "<title>Home</title>") {
		t.Errorf("body missing title, got: %s", body)
	}
}

func TestRender_PageIncludesPartials(t *testing.T) {
	r := testRenderer(t)

	var buf strings.Builder
	err := r.Render(&buf, "base", "with_partial", map[string]string{"Name": "Bob"})
	if err != nil {
		t.Fatal(err)
	}

	body := buf.String()
	if !strings.Contains(body, "<p>Greetings, Bob</p>") {
		t.Errorf("body missing partial content, got: %s", body)
	}
}

func TestRender_Partial(t *testing.T) {
	r := testRenderer(t)

	var buf strings.Builder
	err := r.RenderPartial(&buf, "greeting", map[string]string{"Name": "Alice"})
	if err != nil {
		t.Fatal(err)
	}

	body := buf.String()
	if body != "<p>Greetings, Alice</p>" {
		t.Errorf("body = %q, want %q", body, "<p>Greetings, Alice</p>")
	}
}

func TestRender_EmbedFS(t *testing.T) {
	tmplFS, err := fs.Sub(testTemplates, "testdata/templates")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewRenderer(tmplFS)
	if err != nil {
		t.Fatalf("NewRenderer with embed.FS failed: %v", err)
	}
}
