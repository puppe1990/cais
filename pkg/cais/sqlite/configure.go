// Package sqlite applies opinionated PRAGMA defaults for Cais apps.
//
// SSE and chat handlers poll SQLite in long-lived requests while other
// handlers write messages. WAL mode allows concurrent readers; busy_timeout
// retries lock contention instead of failing immediately. MaxOpenConns(1)
// matches SQLite's single-writer model — keep write transactions short during streams.
package sqlite

import "database/sql"

// Configure applies production-friendly defaults for SQLite (WAL, busy_timeout, foreign_keys).
// Called from scaffold NewSQLiteStore on boot.
func Configure(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return err
		}
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return nil
}
