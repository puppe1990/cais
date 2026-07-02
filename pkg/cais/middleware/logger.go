package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/logentry"
	"github.com/puppe1990/cais/pkg/cais/logtime"
)

func Logger(cfg cais.Config) func(http.Handler) http.Handler {
	return LoggerTo(cfg, log.Writer())
}

func LoggerTo(cfg cais.Config, w io.Writer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return LoggerWithWriter(cfg, w, next)
	}
}

// LoggerWithWriter logs requests. Development emits logentry JSON; production keeps Rails-style text.
func LoggerWithWriter(cfg cais.Config, w io.Writer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if skipRequestLog(r.URL.Path) {
			next.ServeHTTP(rw, r)
			return
		}

		start := time.Now()
		remote := ClientIP(r, cfg)
		if cfg.Env == "development" {
			_ = logentry.Write(w, logentry.Entry{
				Kind:   "request",
				Phase:  "started",
				At:     start.UTC(),
				Method: r.Method,
				Path:   r.URL.Path,
				Remote: remote,
			})
		} else {
			_, _ = fmt.Fprintf(w, "Started %s %q for %s at %s\n", r.Method, r.URL.Path, remote, logtime.Format(start))
		}

		rec := &statusRecorder{ResponseWriter: rw, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		elapsed := time.Since(start)
		if cfg.Env == "development" {
			_ = logentry.Write(w, logentry.Entry{
				Kind:       "request",
				Phase:      "completed",
				At:         time.Now().UTC(),
				Method:     r.Method,
				Path:       r.URL.Path,
				Status:     rec.status,
				Remote:     remote,
				DurationMS: float64(elapsed.Microseconds()) / 1000,
			})
		} else {
			_, _ = fmt.Fprintf(
				w,
				"Completed %s in %s at %s\n",
				statusLabel(rec.status),
				formatDuration(elapsed),
				logtime.Now(),
			)
		}
	})
}

func skipRequestLog(path string) bool {
	return path == "/health" || path == "/logs" || strings.HasPrefix(path, "/static/")
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
