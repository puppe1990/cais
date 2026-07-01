package jobs

import (
	"context"
	"testing"
	"time"
)

func TestRunScheduler_enqueuesOnCronMatch(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()

	if err := store.UpsertRecurring(ctx, RecurringOptions{
		Kind: "Nightly",
		Cron: "0 3 * * *",
	}); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 1, 3, 0, 0, 0, time.UTC)
	n, err := RunScheduler(ctx, store, now)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("enqueued = %d, want 1", n)
	}

	counts, err := store.CountByStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if counts[StatusReady] != 1 {
		t.Fatalf("ready = %d", counts[StatusReady])
	}

	n, err = RunScheduler(ctx, store, now)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("second run enqueued = %d, want 0", n)
	}
}
