package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestLogger_RailsStyleRequestLog(t *testing.T) {
	var buf bytes.Buffer
	handler := LoggerWithWriter(cais.Config{}, &buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	out := buf.String()
	if !strings.Contains(out, "Started GET \"/login\" for 127.0.0.1 at ") {
		t.Errorf("missing Started line with timestamp, got:\n%s", out)
	}
	if !strings.Contains(out, "Completed 200 OK in") || !strings.Contains(out, " at ") {
		t.Errorf("missing Completed line with timestamp, got:\n%s", out)
	}
}

func TestLogger_SkipsStaticAssets(t *testing.T) {
	var buf bytes.Buffer
	handler := LoggerWithWriter(cais.Config{}, &buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/static/css/styles.css", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log for static asset, got:\n%s", buf.String())
	}
}

func TestLogger_JSONInDevelopment(t *testing.T) {
	var buf bytes.Buffer
	handler := LoggerWithWriter(cais.Config{Env: "development"}, &buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 JSON lines, got:\n%s", buf.String())
	}
	for i, phase := range []string{"started", "completed"} {
		var got map[string]any
		if err := json.Unmarshal([]byte(lines[i]), &got); err != nil {
			t.Fatalf("line %d: invalid JSON: %v", i, err)
		}
		if got["kind"] != "request" || got["phase"] != phase {
			t.Errorf("line %d = %v", i, got)
		}
	}
}

func TestLogger_SlowRequestMarksDuration(t *testing.T) {
	var buf bytes.Buffer
	handler := LoggerWithWriter(cais.Config{}, &buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "10.0.0.1:80"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !strings.Contains(buf.String(), "Completed 201 Created in") {
		t.Fatalf("got:\n%s", buf.String())
	}
}
