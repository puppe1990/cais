package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/puppe1990/cais/internal/models"
	"github.com/puppe1990/cais/pkg/cais/session"
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

	err = s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&name)
	if err != nil {
		t.Fatalf("schema_migrations table not found: %v", err)
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
		t.Fatalf("count = %d, want 0", count)
	}

	if _, err := s.InsertContact(models.Contact{Name: "A", Email: "a@example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertContact(models.Contact{Name: "B", Email: "b@example.com"}); err != nil {
		t.Fatal(err)
	}

	count, err = s.CountContacts()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestStore_FindContact_notFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.FindContact(999)
	if err == nil {
		t.Fatal("expected error for missing contact")
	}
}

func TestStore_FindUserByEmail_developmentSeed(t *testing.T) {
	s, err := NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "demo@example.com" {
		t.Errorf("email = %q, want demo@example.com", user.Email)
	}
	if user.PasswordHash == "" {
		t.Error("password hash should be set for seeded user")
	}
}

func TestStore_FindUserByEmail_notFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.FindUserByEmail("missing@example.com")
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}

func TestStore_Sessions(t *testing.T) {
	s := newTestStore(t)
	if err := session.EnsureSQLiteSchema(s.DB()); err != nil {
		t.Fatal(err)
	}

	sess := s.Sessions()
	token, err := sess.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	userID, ok := sess.Get(token)
	if !ok || userID != 1 {
		t.Fatalf("userID = %d, ok = %v, want 1 true", userID, ok)
	}
}

func TestStore_PingAndDB(t *testing.T) {
	s := newTestStore(t)

	if err := s.Ping(); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if s.DB() == nil {
		t.Fatal("DB() returned nil")
	}
}

func TestStore_NewSQLiteStore_createsParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "app.db")

	s, err := NewSQLiteStore(path, "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("db file missing: %v", err)
	}
}

func TestSeedAuthData_skipsOutsideDevelopment(t *testing.T) {
	s, err := NewSQLiteStore(":memory:", "production")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.FindUserByEmail("demo@example.com")
	if err == nil {
		t.Fatal("expected no seeded user in production")
	}
}

func TestStore_Sessions_usesSQLiteBackend(t *testing.T) {
	s := newTestStore(t)

	sess := s.Sessions()
	if sess == nil {
		t.Fatal("Sessions() returned nil")
	}
}
