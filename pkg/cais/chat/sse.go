package chat

import (
	"net/http"
	"time"

	"github.com/puppe1990/cais/pkg/cais/stream"
)

// WriteStream emits event: stream for live token/tool updates into #chat-live.
func WriteStream(w http.ResponseWriter, html string) error {
	return stream.WriteEvent(w, "stream", html)
}

// WriteMessage emits event: message for finalized bubbles appended to #chat-stream.
func WriteMessage(w http.ResponseWriter, html string) error {
	return stream.WriteEvent(w, "message", html)
}

// WriteThinking emits event: thinking for the thinking indicator swap.
func WriteThinking(w http.ResponseWriter, html string) error {
	return stream.WriteEvent(w, "thinking", html)
}

// WriteUnsafeLive writes a pre-rendered HTML fragment to the live slot (see chat.UnsafeLiveHTML).
func WriteUnsafeLive(w http.ResponseWriter, htmlContent string) error {
	return WriteStream(w, UnsafeLiveHTML(htmlContent))
}

// WriteUnsafeMessage writes a pre-rendered final message (see chat.UnsafeMessageHTML).
func WriteUnsafeMessage(w http.ResponseWriter, role Role, htmlContent string, at time.Time) error {
	return WriteMessage(w, UnsafeMessageHTML(role, htmlContent, at))
}
