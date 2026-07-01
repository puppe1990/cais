package sqlite

import "database/sql"

// Configure applies production-friendly defaults for SQLite.
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
