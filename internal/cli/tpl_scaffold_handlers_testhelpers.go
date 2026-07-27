// Inertia handler scaffold templates (ported from Cais demo).
package cli

const tplHelpersTest = `package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

func testSite() meta.Site {
	return meta.Site{AppName: "{{.AppName}}", AppURL: "https://cais.example.com"}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}

func setupTestRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	root := projectRoot(t)
	layout := filepath.Join(root, "web", "templates", "layouts", "base.html")
	if _, err := os.Stat(layout); err != nil {
		return cais.NewRendererStub(i18n.DefaultCatalog())
	}
	r, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), i18n.DefaultCatalog())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}`

const tplInertiaTest = `package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	inertia "github.com/romsar/gonertia/v3"
)

const testInertiaRoot = ` + "`" + `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8" />{{"{{ .inertiaHead }}"}}</head>
<body>{{"{{ .inertia }}"}}</body>
</html>` + "`" + `

func setupTestInertia(t *testing.T) *inertia.Inertia {
	t.Helper()
	i, err := inertia.New(testInertiaRoot)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func inertiaRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("X-Inertia", "true")
	return req
}

func parseInertiaJSON(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("not json: %v body=%s", err, rr.Body.String())
	}
	return payload
}

func assertInertiaComponent(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()
	payload := parseInertiaJSON(t, rr)
	if payload["component"] != want {
		t.Errorf("component = %v, want %s", payload["component"], want)
	}
}

func assertInertiaErrors(t *testing.T, rr *httptest.ResponseRecorder, keys ...string) {
	t.Helper()
	payload := parseInertiaJSON(t, rr)
	props, ok := payload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", payload)
	}
	errors, ok := props["errors"].(map[string]any)
	if !ok || len(errors) == 0 {
		t.Fatalf("missing errors in props: %v", props)
	}
	for _, k := range keys {
		if _, ok := errors[k]; !ok {
			t.Errorf("errors missing key %q: %v", k, errors)
		}
	}
}

func assertInertiaProp(t *testing.T, rr *httptest.ResponseRecorder, key string) any {
	t.Helper()
	payload := parseInertiaJSON(t, rr)
	props, ok := payload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", payload)
	}
	v, ok := props[key]
	if !ok {
		t.Fatalf("props missing %q: %v", key, props)
	}
	return v
}`
