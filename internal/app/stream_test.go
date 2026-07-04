package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/cais/pkg/cais/stream"
)

// TestApp_SSEFlushThroughMiddleware verifies stream.Flush works through the same
// middleware stack used by New (logger wraps ResponseWriter).
func TestApp_SSEFlushThroughMiddleware(t *testing.T) {
	t.Parallel()

	sseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream.RelaySSE(w)
		_, _ = w.Write([]byte("event: message\ndata: ping\n\n"))
		if err := stream.Flush(w); err != nil {
			t.Errorf("stream.Flush: %v", err)
		}
	})

	cfg := cais.Config{Env: "test"}
	h := middleware.SecurityHeaders(cfg)(
		middleware.Recover(
			middleware.Logger(cfg)(sseHandler),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/chat/1/stream", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
	if !strings.Contains(rr.Body.String(), "data: ping") {
		t.Errorf("body = %q, want SSE event", rr.Body.String())
	}
}
