package middleware

import (
	"fmt"
	"io"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/logentry"
	"github.com/puppe1990/cais/pkg/cais/logtime"
)

func logRequestStarted(w io.Writer, cfg cais.Config, method, path, remote string, at time.Time) {
	if cfg.LogJSON() {
		_ = logentry.Write(w, logentry.Entry{
			Kind:   "request",
			Phase:  "started",
			At:     at.UTC(),
			Method: method,
			Path:   path,
			Remote: remote,
		})
		return
	}
	_, _ = fmt.Fprintf(w, "Started %s %q for %s at %s\n", method, path, remote, logtime.Format(at))
}

func logRequestCompleted(w io.Writer, cfg cais.Config, method, path, remote string, status int, elapsed time.Duration) {
	if cfg.LogJSON() {
		_ = logentry.Write(w, logentry.Entry{
			Kind:       "request",
			Phase:      "completed",
			At:         time.Now().UTC(),
			Method:     method,
			Path:       path,
			Status:     status,
			Remote:     remote,
			DurationMS: float64(elapsed.Microseconds()) / 1000,
		})
		return
	}
	_, _ = fmt.Fprintf(
		w,
		"Completed %s in %s at %s\n",
		statusLabel(status),
		formatDuration(elapsed),
		logtime.Now(),
	)
}
