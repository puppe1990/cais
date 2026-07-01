package store

import (
	"testing"

	"github.com/matheuspuppe/cais/internal/models"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
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