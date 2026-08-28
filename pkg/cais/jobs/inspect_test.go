package jobs

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestList_filtersByStatusNewestFirst(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()

	readyID, err := Enqueue(ctx, store, Options{Kind: "ReadyOne"})
	if err != nil {
		t.Fatal(err)
	}
	doneID, err := Enqueue(ctx, store, Options{Kind: "DoneOne"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFinished(ctx, doneID); err != nil {
		t.Fatal(err)
	}

	got, err := store.List(ctx, ListFilter{Status: StatusReady, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != readyID || got[0].Kind != "ReadyOne" {
		t.Fatalf("ready list = %+v, want id=%d ReadyOne", got, readyID)
	}

	done, err := store.List(ctx, ListFilter{Status: StatusFinished})
	if err != nil {
		t.Fatal(err)
	}
	if len(done) != 1 || done[0].ID != doneID {
		t.Fatalf("finished list = %+v, want id=%d", done, doneID)
	}
}

func TestRetryFailed_requeuesAndResetsAttempts(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()

	id, err := Enqueue(ctx, store, Options{Kind: "Flaky", MaxAttempts: 1})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.Claim(ctx, DefaultQueue)
	if err != nil || job == nil {
		t.Fatal(err)
	}
	if err := store.MarkFailed(ctx, job.ID, errors.New("boom"), job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}

	if err := store.RetryFailed(ctx, id); err != nil {
		t.Fatal(err)
	}

	rows, err := store.List(ctx, ListFilter{Status: StatusReady})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != id || rows[0].Attempts != 0 {
		t.Fatalf("retried = %+v, want ready id=%d attempts=0", rows, id)
	}
}

func TestRetryFailed_rejectsNonFailed(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	id, err := Enqueue(ctx, store, Options{Kind: "Ready"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RetryFailed(ctx, id); !errors.Is(err, ErrNotFailed) {
		t.Fatalf("err = %v, want ErrNotFailed", err)
	}
}

func TestDiscard_deletesFailedOnly(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()

	failID, err := Enqueue(ctx, store, Options{Kind: "Dead", MaxAttempts: 1})
	if err != nil {
		t.Fatal(err)
	}
	keepID, err := Enqueue(ctx, store, Options{Kind: "Keep"})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.Claim(ctx, DefaultQueue)
	if err != nil || job == nil {
		t.Fatal(err)
	}
	if err := store.MarkFailed(ctx, job.ID, errors.New("dead"), job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}

	if err := store.Discard(ctx, failID); err != nil {
		t.Fatal(err)
	}
	if err := store.Discard(ctx, keepID); !errors.Is(err, ErrNotFailed) {
		t.Fatalf("discard ready: err = %v, want ErrNotFailed", err)
	}

	got, err := store.Get(ctx, failID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("get discarded = %+v err=%v, want ErrNotFound", got, err)
	}
}

func TestCountByQueue_groupsStatus(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if _, err := Enqueue(ctx, store, Options{Kind: "A", Queue: "mail"}); err != nil {
		t.Fatal(err)
	}
	id, err := Enqueue(ctx, store, Options{Kind: "B", Queue: "mail"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFinished(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := Enqueue(ctx, store, Options{Kind: "C"}); err != nil {
		t.Fatal(err)
	}

	got, err := store.CountByQueue(ctx)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]QueueCount{}
	for _, q := range got {
		byName[q.Queue] = q
	}
	if byName["mail"].Ready != 1 || byName["mail"].Finished != 1 {
		t.Fatalf("mail = %+v", byName["mail"])
	}
	if byName[DefaultQueue].Ready != 1 {
		t.Fatalf("default = %+v", byName[DefaultQueue])
	}
}

func TestListScheduled_returnsDueLaterRows(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	id, err := SetWait(ctx, store, 2*time.Hour, Options{Kind: "Later"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.ListScheduled(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != id || got[0].Kind != "Later" {
		t.Fatalf("scheduled = %+v, want id=%d Later", got, id)
	}
}

func TestPruneFinished_deletesFinishedRows(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	readyID, err := Enqueue(ctx, store, Options{Kind: "KeepReady"})
	if err != nil {
		t.Fatal(err)
	}
	doneID, err := Enqueue(ctx, store, Options{Kind: "Done"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFinished(ctx, doneID); err != nil {
		t.Fatal(err)
	}

	n, err := store.PruneFinished(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("pruned = %d, want 1", n)
	}
	if _, err := store.Get(ctx, doneID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("finished row still present: %v", err)
	}
	if _, err := store.Get(ctx, readyID); err != nil {
		t.Fatalf("ready row missing: %v", err)
	}
}
