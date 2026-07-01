package session

import (
	"database/sql"
	"fmt"
)

const sqliteSchema = `CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY NOT NULL,
  user_id INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);`

// EnsureSQLiteSchema creates the sessions table when missing.
func EnsureSQLiteSchema(db *sql.DB) error {
	if _, err := db.Exec(sqliteSchema); err != nil {
		return fmt.Errorf("sessions schema: %w", err)
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
	if _, err := s.db.Exec("INSERT INTO sessions (token, user_id) VALUES (?, ?)", token, userID); err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return token, nil
}

func (s *SQLiteStore) Get(token string) (int64, bool) {
	var id int64
	err := s.db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", token).Scan(&id)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *SQLiteStore) Delete(token string) {
	_, _ = s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
}
