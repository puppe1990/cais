package stream

import (
	"net/http"
	"time"
)

// Flush writes buffered data through middleware-wrapped ResponseWriters.
// Prefer this over http.Flusher type assertions — wrapped writers may not implement Flusher.
func Flush(w http.ResponseWriter) error {
	return http.NewResponseController(w).Flush()
}

// RelaySSE sets standard SSE headers and disables the write deadline for long-lived streams.
func RelaySSE(w http.ResponseWriter) *http.ResponseController {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{}) // zero time clears deadline
	return rc
}
