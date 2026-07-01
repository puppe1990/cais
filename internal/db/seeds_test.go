package db

import (
	"testing"

	"github.com/puppe1990/cais/internal/store"
)

func TestRunSeeds_insertsContact(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	if err := RunSeeds(s); err != nil {
		t.Fatal(err)
	}
	n, err := s.CountContacts()
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("contacts = %d, want at least 1", n)
	}
}
