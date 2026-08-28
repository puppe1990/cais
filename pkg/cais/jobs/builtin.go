package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/puppe1990/cais/pkg/cais/session"
)

const (
	KindPruneSessions = "PruneSessions"
	KindPruneFinished = "PruneFinished"
	pruneFinishedCron = "0 4 * * *"
)

// PruneSessionsHandler deletes expired session rows.
func PruneSessionsHandler(db *sql.DB) Handler {
	return func(ctx context.Context, _ []byte) error {
		if err := session.EnsureSQLiteSchema(db); err != nil {
			return err
		}
		_, err := session.NewSQLiteStore(db).PruneExpired()
		return err
	}
}

// PruneFinishedHandler deletes finished jobs. Payload: {"older_than_hours":24}; 0 = all.
func PruneFinishedHandler(db *sql.DB) Handler {
	return func(ctx context.Context, payload []byte) error {
		hours := 24
		var p struct {
			OlderThanHours *int `json:"older_than_hours"`
		}
		if len(payload) > 0 && json.Unmarshal(payload, &p) == nil && p.OlderThanHours != nil {
			hours = *p.OlderThanHours
		}
		_, err := NewStore(db).PruneFinished(ctx, time.Duration(hours)*time.Hour)
		return err
	}
}

// DefaultRegistry returns built-in framework jobs.
func DefaultRegistry(db *sql.DB) *Registry {
	r := NewRegistry()
	r.Register(KindPruneSessions, PruneSessionsHandler(db))
	r.Register(KindPruneFinished, PruneFinishedHandler(db))
	return r
}
