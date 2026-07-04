package stream

import (
	"io"
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

// RelayAndCopy streams bytes from src to w, flushing after each read so SSE clients
// receive events promptly through middleware-wrapped ResponseWriters.
func RelayAndCopy(w http.ResponseWriter, src io.Reader) (int64, error) {
	buf := make([]byte, 4096)
	var total int64
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			written, writeErr := w.Write(buf[:n])
			total += int64(written)
			if writeErr != nil {
				return total, writeErr
			}
			_ = Flush(w)
		}
		if readErr != nil {
			if readErr == io.EOF {
				return total, nil
			}
			return total, readErr
		}
	}
}
