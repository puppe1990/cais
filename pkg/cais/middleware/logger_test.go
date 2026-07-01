package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLogger_RailsStyleRequestLog(t *testing.T) {
	var buf bytes.Buffer
	handler := LoggerWithWriter(&buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	out := buf.String()
	if !strings.Contains(out, "Started GET \"/login\" for 127.0.0.1") {
		t.Errorf("missing Started line, got:\n%s", out)
	}
	if !strings.Contains(out, "Completed 200 OK in") {
		t.Errorf("missing Completed line, got:\n%s", out)
	}
}

func TestLogger_SkipsStaticAssets(t *testing.T) {
	var buf bytes.Buffer
	handler := LoggerWithWriter(&buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/static/css/styles.css", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log for static asset, got:\n%s", buf.String())
	}
}

func TestLogger_SlowRequestMarksDuration(t *testing.T) {
	var buf bytes.Buffer
	handler := LoggerWithWriter(&buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
