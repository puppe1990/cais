package stream

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"
)

type flushCounter struct {
	httptest.ResponseRecorder
	n int
}

func (f *flushCounter) Flush() {
	f.n++
}

// unwrapOnly mimics middleware that hides http.Flusher but exposes Unwrap (like statusRecorder).
type unwrapOnly struct {
	http.ResponseWriter
}

func (u *unwrapOnly) Unwrap() http.ResponseWriter {
	return u.ResponseWriter
}

func TestFlush_usesResponseControllerNotFlusherAssertion(t *testing.T) {
	counter := &flushCounter{}
	wrapped := &unwrapOnly{ResponseWriter: counter}

	if _, ok := any(wrapped).(http.Flusher); ok {
		t.Fatal("test setup: wrapper must not implement http.Flusher")
	}

	if err := Flush(wrapped); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}
	if counter.n != 1 {
		t.Errorf("underlying Flush calls = %d, want 1", counter.n)
	}
}

func TestFlush_throughLoggerMiddleware(t *testing.T) {
	counter := &flushCounter{}
	var logBuf bytes.Buffer
	handler := middleware.LoggerWithWriter(cais.Config{}, &logBuf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := Flush(w); err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/chat/stream", nil)
	handler.ServeHTTP(counter, req)

	if counter.n != 1 {
		t.Errorf("Flush calls = %d, want 1", counter.n)
	}
}

func TestRelaySSE_setsHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	RelaySSE(rr)

	if ct := rr.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
	if cc := rr.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", cc)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRelayAndCopy_forwardsEvents(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: message\ndata: hello\n\n"))
		_ = Flush(w)
		_, _ = w.Write([]byte("event: message\ndata: world\n\n"))
		_ = Flush(w)
	}))
	t.Cleanup(upstream.Close)

	resp, err := http.Get(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	rr := httptest.NewRecorder()
	RelaySSE(rr)
	n, err := RelayAndCopy(rr, resp.Body)
	if err != nil {
		t.Fatalf("RelayAndCopy: %v", err)
	}
	if n == 0 {
		t.Fatal("expected bytes copied")
	}
	body := rr.Body.String()
	if !strings.Contains(body, "data: hello") || !strings.Contains(body, "data: world") {
		t.Errorf("body = %q, want both SSE events", body)
	}
}
