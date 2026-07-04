package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
)

func Logger(cfg cais.Config) func(http.Handler) http.Handler {
	return LoggerTo(cfg, log.Writer())
}

func LoggerTo(cfg cais.Config, w io.Writer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return LoggerWithWriter(cfg, w, next)
	}
}

// LoggerWithWriter logs requests. JSON when cfg.LogJSON() (default in dev/production); LOG_FORMAT=text for Rails-style.
func LoggerWithWriter(cfg cais.Config, w io.Writer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if skipRequestLog(r.URL.Path) {
			next.ServeHTTP(rw, r)
			return
		}

		start := time.Now()
		remote := ClientIP(r, cfg)
		logRequestStarted(w, cfg, r.Method, r.URL.Path, remote, start)

		rec := &statusRecorder{ResponseWriter: rw, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Long-lived SSE paths log Started only — Completed duration is misleading.
		if !skipCompletedLog(r.URL.Path) {
			logRequestCompleted(w, cfg, r.Method, r.URL.Path, remote, rec.status, time.Since(start))
		}
	})
}

func skipRequestLog(path string) bool {
	return path == "/health" || path == "/logs" || strings.HasPrefix(path, "/static/")
}

func skipCompletedLog(path string) bool {
	return strings.HasSuffix(path, "/stream") || path == "/event" || strings.HasSuffix(path, "/event")
}

func statusLabel(code int) string {
	text := http.StatusText(code)
	if text == "" {
		text = "Unknown"
	}
	return fmt.Sprintf("%d %s", code, text)
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	case d < time.Second:
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Flush delegates to the underlying writer so SSE and other streaming responses work.
func (r *statusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
