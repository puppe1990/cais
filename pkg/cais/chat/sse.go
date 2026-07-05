package chat

import (
	"fmt"
	"net/http"

	"github.com/puppe1990/cais/pkg/cais/stream"
)

// WriteStream emits event: stream for live token/tool updates into #chat-live.
func WriteStream(w http.ResponseWriter, html string) error {
	_, err := fmt.Fprintf(w, "event: stream\ndata: %s\n\n", html)
	if err != nil {
		return err
	}
	return stream.Flush(w)
}

// WriteMessage emits event: message for finalized bubbles appended to #chat-stream.
func WriteMessage(w http.ResponseWriter, html string) error {
	_, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", html)
	if err != nil {
		return err
	}
	return stream.Flush(w)
}

// WriteThinking emits event: thinking for the thinking indicator swap.
func WriteThinking(w http.ResponseWriter, html string) error {
	_, err := fmt.Fprintf(w, "event: thinking\ndata: %s\n\n", html)
	if err != nil {
		return err
	}
	return stream.Flush(w)
}
