package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/puppe1990/cais/internal/models"
	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/session"
	caissqlite "github.com/puppe1990/cais/pkg/cais/sqlite"
	"github.com/puppe1990/cais/pkg/cais/sqllog"
)

type Store interface {
	InsertContact(contact models.Contact) (int64, error)
	FindContact(id int64) (models.Contact, error)
	CountContacts() (int64, error)
	FindUserByEmail(email string) (models.User, error)
	Sessions() session.Store
	Ping() error
	Close() error
}

type SQLiteStore struct {
	db *sqllog.DB
}

func NewSQLiteStore(dsn string, env string) (*SQLiteStore, error) {
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
	if err := caissqlite.Configure(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	cfg := sqllog.Config{Enabled: sqllog.EnabledForEnv(env)}
	if cfg.Enabled {
		cfg.Writer = devlog.MirrorDefault(os.Stdout)
	}
	wrapped := sqllog.Wrap(db, cfg)
	if err := seedAuthData(wrapped.Raw(), env); err != nil {
		_ = wrapped.Close()
		return nil, err
	}
	return &SQLiteStore{db: wrapped}, nil
}

func seedAuthData(db *sql.DB, env string) error {
	if env != "development" {
		return nil
	}
	if err := session.EnsureSQLiteSchema(db); err != nil {
		return err
	}
	hash, err := session.HashPassword("password")
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT OR IGNORE INTO users (email, password_hash) VALUES (?, ?)", "demo@example.com", hash)
	return err
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

func (s *SQLiteStore) CountContacts() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count contacts: %w", err)
	}
	return count, nil
}

func (s *SQLiteStore) FindUserByEmail(email string) (models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (s *SQLiteStore) Sessions() session.Store {
	return session.NewSQLiteStore(s.db.Raw())
}

func (s *SQLiteStore) Ping() error {
	return s.db.Raw().Ping()
}

func (s *SQLiteStore) DB() *sql.DB {
	return s.db.Raw()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
