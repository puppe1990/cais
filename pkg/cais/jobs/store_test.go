package jobs

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

var errTestFail = errors.New("job failed")

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := EnsureSchema(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestEnqueue_andClaim(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()

	id, err := Enqueue(ctx, store, Options{Kind: "Demo", Payload: map[string]any{"n": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("id = 0")
	}

	job, err := store.Claim(ctx, DefaultQueue)
	if err != nil {
		t.Fatal(err)
	}
	if job == nil {
		t.Fatal("expected claimed job")
	}
	if job.Kind != "Demo" || job.Attempts != 1 {
		t.Fatalf("job = %+v", job)
	}
}

func TestSetWait_dispatchedLater(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()

	_, err := SetWait(ctx, store, 2*time.Hour, Options{Kind: "Later"})
	if err != nil {
		t.Fatal(err)
	}

	job, err := store.Claim(ctx, DefaultQueue)
	if err != nil {
		t.Fatal(err)
	}
	if job != nil {
		t.Fatal("scheduled job should not be claimable yet")
	}

	n, err := store.DispatchDue(ctx, time.Now().UTC().Add(3*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("dispatched = %d, want 1", n)
	}

	job, err = store.Claim(ctx, DefaultQueue)
	if err != nil || job == nil || job.Kind != "Later" {
		t.Fatalf("job after dispatch = %+v, err = %v", job, err)
	}
}

func TestMarkFailed_retriesThenFails(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()

	id, err := Enqueue(ctx, store, Options{Kind: "Flaky", MaxAttempts: 2})
	if err != nil {
		t.Fatal(err)
	}

	job, err := store.Claim(ctx, DefaultQueue)
	if err != nil || job == nil {
		t.Fatal(err)
	}
	if err := store.MarkFailed(ctx, job.ID, errTestFail, job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM jobs WHERE id = ?`, id).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != StatusReady {
		t.Fatalf("status = %q, want ready (retry)", status)
	}

	time.Sleep(3 * time.Second)

	job, err = store.Claim(ctx, DefaultQueue)
	if err != nil || job == nil {
		t.Fatalf("second claim: job=%v err=%v", job, err)
	}
	if err := store.MarkFailed(ctx, job.ID, errTestFail, job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT status FROM jobs WHERE id = ?`, id).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != StatusFailed {
		t.Fatalf("status = %q, want failed", status)
	}
}

func TestCountByStatus(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()

	if _, err := Enqueue(ctx, store, Options{Kind: "A"}); err != nil {
		t.Fatal(err)
	}
	id, err := Enqueue(ctx, store, Options{Kind: "B"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFinished(ctx, id); err != nil {
		t.Fatal(err)
	}

	counts, err := store.CountByStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if counts[StatusReady] != 1 || counts[StatusFinished] != 1 {
		t.Fatalf("counts = %#v", counts)
	}
}
