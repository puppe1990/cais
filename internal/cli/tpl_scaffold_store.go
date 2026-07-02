package cli

const tplContactModel = `package models

import "time"

type Contact struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
}
`

const tplStore = `package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/session"
	caissqlite "github.com/puppe1990/cais/pkg/cais/sqlite"
	"github.com/puppe1990/cais/pkg/cais/sqllog"
	"{{.ModulePath}}/internal/models"
)

var ErrEmailTaken = errors.New("email already registered")

type Store interface {
	InsertContact(contact models.Contact) (int64, error)
	FindContact(id int64) (models.Contact, error)
	CountContacts() (int64, error)
	FindUserByEmail(email string) (models.User, error)
	CreateUser(email, passwordHash string) (int64, error)
	CreatePasswordResetToken(userID int64) (string, error)
	FindPasswordResetUserID(token string) (int64, bool)
	ResetPasswordWithToken(token, passwordHash string) error
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

	cfg := sqllog.ConfigForEnv(env)
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

func (s *SQLiteStore) CreateUser(email, passwordHash string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email, passwordHash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, ErrEmailTaken
		}
		return 0, fmt.Errorf("create user: %w", err)
	}
	return result.LastInsertId()
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
`

const tplStoreTest = `package store

import (
	"testing"

	"{{.ModulePath}}/internal/models"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStore_Migrations(t *testing.T) {
	s := newTestStore(t)

	var name string
	err := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='contacts'").Scan(&name)
	if err != nil {
		t.Fatalf("contacts table not found: %v", err)
	}
}

func TestStore_InsertContact(t *testing.T) {
	s := newTestStore(t)

	contact := models.Contact{Name: "Alice", Email: "alice@example.com"}
	id, err := s.InsertContact(contact)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Error("id = 0, want non-zero")
	}
}

func TestStore_FindContact(t *testing.T) {
	s := newTestStore(t)

	contact := models.Contact{Name: "Bob", Email: "bob@example.com"}
	id, err := s.InsertContact(contact)
	if err != nil {
		t.Fatal(err)
	}

	found, err := s.FindContact(id)
	if err != nil {
		t.Fatal(err)
	}
	if found.Name != "Bob" {
		t.Errorf("Name = %q, want %q", found.Name, "Bob")
	}
	if found.Email != "bob@example.com" {
		t.Errorf("Email = %q, want %q", found.Email, "bob@example.com")
	}
}

func TestStore_CountContacts(t *testing.T) {
	s := newTestStore(t)

	count, err := s.CountContacts()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	_, err = s.InsertContact(models.Contact{Name: "Alice", Email: "alice@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.InsertContact(models.Contact{Name: "Bob", Email: "bob@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	count, err = s.CountContacts()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}
`

const tplMigrations = `package store

import (
	"database/sql"
	"embed"

	"github.com/puppe1990/cais/pkg/cais/migrate"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func applyMigrations(db *sql.DB) error {
	return migrate.Apply(db, migrationFS, "migrations")
}
`

const tplMigration001 = `CREATE TABLE IF NOT EXISTS contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const tplStoreMinimal = `package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/sqllog"
	_ "modernc.org/sqlite"
)

type Store interface {
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

	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	cfg := sqllog.ConfigForEnv(env)
	if cfg.Enabled {
		cfg.Writer = devlog.MirrorDefault(os.Stdout)
	}
	return &SQLiteStore{db: sqllog.Wrap(db, cfg)}, nil
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
`

const tplStoreTestMinimal = `package store

import "testing"

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStore_Migrations(t *testing.T) {
	_ = newTestStore(t)
}
`
