package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/matheuspuppe/cais/internal/models"
)

type Store interface {
	InsertContact(contact models.Contact) (int64, error)
	FindContact(id int64) (models.Contact, error)
	Close() error
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	if dsn != ":memory:" {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) InsertContact(contact models.Contact) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO contacts (name, email) VALUES (?, ?)",
		contact.Name, contact.Email,
	)
	if err != nil {
		return 0, fmt.Errorf("insert contact: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) FindContact(id int64) (models.Contact, error) {
	var c models.Contact
	err := s.db.QueryRow(
		"SELECT id, name, email, created_at FROM contacts WHERE id = ?",
		id,
	).Scan(&c.ID, &c.Name, &c.Email, &c.CreatedAt)
	if err != nil {
		return models.Contact{}, fmt.Errorf("find contact: %w", err)
	}
	return c, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
