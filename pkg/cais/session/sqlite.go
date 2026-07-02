package session

import (
	"database/sql"
	"fmt"
	"time"
)

const sessionTTL = 7 * 24 * time.Hour

const sqliteSchema = `CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY NOT NULL,
  user_id INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  expires_at TEXT NOT NULL
);`

// EnsureSQLiteSchema creates the sessions table when missing.
func EnsureSQLiteSchema(db *sql.DB) error {
	if _, err := db.Exec(sqliteSchema); err != nil {
		return fmt.Errorf("sessions schema: %w", err)
	}
	return migrateSQLiteSchema(db)
}

// migrateSQLiteSchema adds expires_at for databases created before session TTL shipped.
// ALTER is idempotent via pragma check — avoids breaking existing app.db files on upgrade.
func migrateSQLiteSchema(db *sql.DB) error {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name = 'expires_at'`,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("sessions schema migrate: %w", err)
	}
	if count == 0 {
		_, err = db.Exec(`ALTER TABLE sessions ADD COLUMN expires_at TEXT NOT NULL DEFAULT (datetime('now', '+7 days'))`)
		if err != nil {
			return fmt.Errorf("sessions schema migrate: %w", err)
		}
	}
	return nil
}

// SQLiteStore persists sessions in SQLite.
type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Create(userID int64) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().UTC().Add(sessionTTL).Format("2006-01-02 15:04:05")
	if _, err := s.db.Exec(
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt,
	); err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return token, nil
}

func (s *SQLiteStore) Get(token string) (int64, bool) {
	var id int64
	err := s.db.QueryRow(
		"SELECT user_id FROM sessions WHERE token = ? AND expires_at > datetime('now')",
		token,
	).Scan(&id)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *SQLiteStore) Delete(token string) {
	_, _ = s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
}

func (s *SQLiteStore) PruneExpired() (int64, error) {
	res, err := s.db.Exec("DELETE FROM sessions WHERE expires_at <= datetime('now')")
	if err != nil {
		return 0, fmt.Errorf("prune expired sessions: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune expired sessions: %w", err)
	}
	return n, nil
}
