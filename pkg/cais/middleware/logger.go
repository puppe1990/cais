package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais/logtime"
)

func Logger(next http.Handler) http.Handler {
	return LoggerWithWriter(log.Writer(), next)
}

func LoggerTo(w io.Writer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return LoggerWithWriter(w, next)
	}
}

func LoggerWithWriter(w io.Writer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if skipRequestLog(r.URL.Path) {
			next.ServeHTTP(rw, r)
			return
		}

		start := time.Now()
		remote := clientIP(r)
		_, _ = fmt.Fprintf(w, "Started %s %q for %s at %s\n", r.Method, r.URL.Path, remote, logtime.Format(start))

		rec := &statusRecorder{ResponseWriter: rw, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		_, _ = fmt.Fprintf(
			w,
			"Completed %s in %s at %s\n",
			statusLabel(rec.status),
			formatDuration(time.Since(start)),
			logtime.Now(),
		)
	})
}

func skipRequestLog(path string) bool {
	return path == "/health" || path == "/logs" || strings.HasPrefix(path, "/static/")
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	if host := r.RemoteAddr; host != "" {
		if i := strings.LastIndex(host, ":"); i >= 0 {
			return host[:i]
		}
		return host
	}
	return "127.0.0.1"
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
