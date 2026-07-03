package store

import (
	"testing"
)

func TestMercado_createAndConfirmReport(t *testing.T) {
	s := openTestStore(t)

	userID, err := s.CreateUser("a@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	otherID, err := s.CreateUser("b@test.com", "hash2")
	if err != nil {
		t.Fatal(err)
	}
	pid, err := s.CreateProduct("Leite", "123", "Laticínios")
	if err != nil {
		t.Fatal(err)
	}
	markets, err := s.ListSupermarkets()
	if err != nil || len(markets) == 0 {
		t.Fatalf("need seeded supermarkets: %v len=%d", err, len(markets))
	}
	reportID, err := s.CreatePriceReport(userID, pid, markets[0].ID, 549)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ConfirmPriceReport(reportID, userID); err != ErrOwnReport {
		t.Fatalf("expected ErrOwnReport, got %v", err)
	}
	count, err := s.ConfirmPriceReport(reportID, otherID)
	if err != nil || count != 1 {
		t.Fatalf("confirm = %d, %v", count, err)
	}
	_, err = s.ConfirmPriceReport(reportID, otherID)
	if err != ErrAlreadyConfirmed {
		t.Fatalf("expected ErrAlreadyConfirmed, got %v", err)
	}
}

func TestMercado_loadStats(t *testing.T) {
	s := openTestStore(t)
	userID, err := s.CreateUser("stats@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := s.CreateProduct("Pão", "999", "Padaria")
	markets, _ := s.ListSupermarkets()
	_, _ = s.CreatePriceReport(userID, pid, markets[0].ID, 399)

	level, points, rank, err := s.LoadStats(userID)
	if err != nil {
		t.Fatal(err)
	}
	if points < pointsPerReport {
		t.Fatalf("points = %d", points)
	}
	if level != levelFromPoints(points) {
		t.Fatalf("level = %d want %d", level, levelFromPoints(points))
	}
	if rank < 1 {
		t.Fatalf("rank = %d", rank)
	}
}

func openTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
