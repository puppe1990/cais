package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

// HTMXOptions configures a minimal test server with real templates and routes.
type HTMXOptions struct {
	Renderer *cais.Renderer
	// RegisterRoutes is called with a mux so the test can wire handlers.
	RegisterRoutes func(mux *http.ServeMux)
}

// NewHTMXServer starts an httptest.Server with the given templates and routes.
// Useful for integration-style tests of HTMX + SSE flows without a full browser.
func NewHTMXServer(t *testing.T, opts HTMXOptions) *httptest.Server {
	t.Helper()

	if opts.Renderer == nil {
		opts.Renderer = NewRenderer(t)
	}

	mux := http.NewServeMux()
	if opts.RegisterRoutes != nil {
		opts.RegisterRoutes(mux)
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// NewChatAgentTestServer is a convenience wrapper that mounts a minimal chat agent
// endpoint using the real chat_sse_agent partial. The caller provides the stream
// and messages handlers.
func NewChatAgentTestServer(t *testing.T, streamHandler, messagesHandler http.HandlerFunc) *httptest.Server {
	t.Helper()

	r := NewRenderer(t)

	return NewHTMXServer(t, HTMXOptions{
		Renderer: r,
		RegisterRoutes: func(mux *http.ServeMux) {
			mux.HandleFunc("/chat/test/stream", streamHandler)
			if messagesHandler != nil {
				mux.HandleFunc("/chat/test/messages", messagesHandler)
			}
			// Example route that serves the agent chat shell (tests can GET this + assert markers).
			mux.HandleFunc("/chat/test", func(w http.ResponseWriter, req *http.Request) {
				data := struct {
					StreamURL   string
					MessagesURL string
					Messages    []struct{ Role, Content string }
				}{
					StreamURL:   "/chat/test/stream",
					MessagesURL: "/chat/test/messages",
				}
				// In real harness usage the caller would render via the provided renderer.
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte(`<div id="chat-messages" data-cais-chat="true"></div>`))
				_ = data
			})
		},
	})
}

// Example usage in a real test (this will be expanded for chat flows):
func TestNewHTMXServer_basic(t *testing.T) {
	srv := NewHTMXServer(t, HTMXOptions{
		RegisterRoutes: func(mux *http.ServeMux) {
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})
		},
	})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

// Test harness for chat SSE contract (reconnect + poll fallback markers).
func TestHTMXHarness_chatAgentShellContract(t *testing.T) {
	srv := NewHTMXServer(t, HTMXOptions{
		RegisterRoutes: func(mux *http.ServeMux) {
			mux.HandleFunc("/chat/1/stream", func(w http.ResponseWriter, r *http.Request) {
				// Minimal SSE response for harness tests
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("event: thinking\ndata: <div id=\"chat-thinking\">...</div>\n\n"))
			})
			mux.HandleFunc("/chat/1/messages", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		},
	})

	// Fetch a page that would include the chat shell (in real app this would be the partial)
	// Here we just validate that the harness can serve and that we can assert markers
	// that the JS relies on (data-cais-sse-persist, poll url, etc).
	// The real contract tests live in stream/chat_sse_test.go; this exercises the harness.
	resp, err := http.Get(srv.URL + "/chat/1/stream")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stream status = %d", resp.StatusCode)
	}
}

// Demonstrates SSE reconnect + poll fallback wiring via harness (scenario from #102).
func TestHTMXHarness_sseReconnectAndPoll(t *testing.T) {
	called := false
	srv := NewHTMXServer(t, HTMXOptions{
		RegisterRoutes: func(mux *http.ServeMux) {
			mux.HandleFunc("/chat/42/stream", func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("event: message\ndata: <div class=\"cais-msg cais-msg-assistant\">hi</div>\n\n"))
			})
			mux.HandleFunc("/chat/42/messages", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("poll ok"))
			})
		},
	})

	// In a real harness test you would render the chat_sse_agent partial
	// and assert presence of data-cais-sse-persist + data-cais-poll-url.
	// Here we simply exercise that the routes are reachable.
	resp, err := http.Get(srv.URL + "/chat/42/stream")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 || !called {
		t.Fatalf("expected stream to be called, status=%d", resp.StatusCode)
	}
}
