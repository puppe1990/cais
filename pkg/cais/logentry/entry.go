package logentry

import (
	"encoding/json"
	"io"
	"time"
)

// Entry is one JSON log line for development tooling (/logs, agents, grep).
type Entry struct {
	Kind       string    `json:"kind"`
	Phase      string    `json:"phase,omitempty"`
	At         time.Time `json:"at"`
	Method     string    `json:"method,omitempty"`
	Path       string    `json:"path,omitempty"`
	Status     int       `json:"status,omitempty"`
	Remote     string    `json:"remote,omitempty"`
	Operation  string    `json:"operation,omitempty"`
	Query      string    `json:"query,omitempty"`
	Args       []any     `json:"args,omitempty"`
	DurationMS float64   `json:"duration_ms,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// Write marshals entry as a single JSON line.
func Write(w io.Writer, e Entry) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}
