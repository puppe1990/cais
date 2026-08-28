package jobs

import (
	"context"
	"testing"
	"time"
)

func TestTouchWorker_listsAsLive(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()

	if err := store.TouchWorker(ctx, WorkerPulse{
		ID: "w1", Queues: "default", Concurrency: 2,
	}); err != nil {
		t.Fatal(err)
	}

	live, err := store.ListLiveWorkers(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 1 || live[0].ID != "w1" || live[0].Concurrency != 2 {
		t.Fatalf("live = %+v", live)
	}
}

func TestListLiveWorkers_excludesStale(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()
	if err := store.TouchWorker(ctx, WorkerPulse{ID: "old"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE job_workers SET heartbeat_at = datetime('now', '-2 minutes')`); err != nil {
		t.Fatal(err)
	}
	live, err := store.ListLiveWorkers(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 0 {
		t.Fatalf("stale worker still live: %+v", live)
	}
}

func TestRequeueOrphaned_skipsJobsOfLiveWorker(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if err := store.TouchWorker(ctx, WorkerPulse{ID: "alive"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Enqueue(ctx, store, Options{Kind: "Owned"}); err != nil {
		t.Fatal(err)
	}
	job, err := store.ClaimFor(ctx, DefaultQueue, "alive")
	if err != nil || job == nil {
		t.Fatalf("claim: %v %+v", err, job)
	}

	n, err := store.RequeueOrphaned(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("requeued live worker job: %d", n)
	}
	got, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusRunning {
		t.Fatalf("status = %q, want running", got.Status)
	}
}

func TestRequeueOrphaned_recoversDeadWorkerJobs(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if _, err := Enqueue(ctx, store, Options{Kind: "Orphan"}); err != nil {
		t.Fatal(err)
	}
	job, err := store.ClaimFor(ctx, DefaultQueue, "dead")
	if err != nil || job == nil {
		t.Fatal(err)
	}

	n, err := store.RequeueOrphaned(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("requeued = %d, want 1", n)
	}
	got, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusReady {
		t.Fatalf("status = %q, want ready", got.Status)
	}
}

func TestList_filtersByKind(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if _, err := Enqueue(ctx, store, Options{Kind: "Mail"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Enqueue(ctx, store, Options{Kind: "Ping"}); err != nil {
		t.Fatal(err)
	}
	got, err := store.List(ctx, ListFilter{Kind: "Mail"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "Mail" {
		t.Fatalf("list by kind = %+v", got)
	}
}
