package jobs

import (
	"database/sql"
	"fmt"
	"strings"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  queue TEXT NOT NULL DEFAULT 'default',
  kind TEXT NOT NULL,
  payload TEXT NOT NULL DEFAULT '{}',
  priority INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'ready',
  attempts INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  last_error TEXT,
  run_at TEXT NOT NULL DEFAULT (datetime('now')),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  finished_at TEXT,
  started_at TEXT,
  worker_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_jobs_poll ON jobs (queue, status, run_at, priority, id);

CREATE TABLE IF NOT EXISTS scheduled_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  queue TEXT NOT NULL DEFAULT 'default',
  kind TEXT NOT NULL,
  payload TEXT NOT NULL DEFAULT '{}',
  priority INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  run_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_run_at ON scheduled_jobs (run_at);

CREATE TABLE IF NOT EXISTS recurring_tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  queue TEXT NOT NULL DEFAULT 'default',
  kind TEXT NOT NULL UNIQUE,
  payload TEXT NOT NULL DEFAULT '{}',
  cron TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  last_run TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS job_workers (
  id TEXT PRIMARY KEY,
  hostname TEXT,
  pid INTEGER NOT NULL DEFAULT 0,
  queues TEXT NOT NULL DEFAULT 'default',
  concurrency INTEGER NOT NULL DEFAULT 1,
  heartbeat_at TEXT NOT NULL,
  started_at TEXT NOT NULL
);
`

// EnsureSchema creates jobs tables when missing and adds columns on existing DBs.
func EnsureSchema(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("jobs schema: %w", err)
	}
	// CREATE TABLE IF NOT EXISTS does not add columns; ignore duplicate-column on old files.
	for _, stmt := range []string{
		`ALTER TABLE jobs ADD COLUMN started_at TEXT`,
		`ALTER TABLE jobs ADD COLUMN worker_id TEXT`,
	} {
		if _, err := db.Exec(stmt); err != nil && !isDupColumn(err) {
			return fmt.Errorf("jobs migrate: %w", err)
		}
	}
	return nil
}

func isDupColumn(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate column")
}
