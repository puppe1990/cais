package jobs

import (
	"context"
	"database/sql"

	"github.com/puppe1990/cais/pkg/cais/session"
)

const KindPruneSessions = "PruneSessions"

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

// DefaultRegistry returns built-in framework jobs.
func DefaultRegistry(db *sql.DB) *Registry {
	r := NewRegistry()
	r.Register(KindPruneSessions, PruneSessionsHandler(db))
	return r
}
