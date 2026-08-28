package jobs

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorker_runsRegisteredJob(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx, cancel := context.WithCancel(context.Background())

	var ran atomic.Bool
	reg := NewRegistry()
	reg.Register("Ping", func(ctx context.Context, payload []byte) error {
		ran.Store(true)
		cancel()
		return nil
	})

	if _, err := Enqueue(ctx, store, Options{Kind: "Ping"}); err != nil {
		t.Fatal(err)
	}

	w := NewWorker(WorkerConfig{
		Store:            store,
		Registry:         reg,
		Concurrency:      1,
		PollInterval:     20 * time.Millisecond,
		DispatchInterval: time.Hour,
	})
	go func() { _ = w.Run(ctx) }()

	deadline := time.After(2 * time.Second)
	for !ran.Load() {
		select {
		case <-deadline:
			t.Fatal("job did not run")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestWorker_writesHeartbeatAndPruneRecurring(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := NewWorker(WorkerConfig{
		Store:             store,
		Registry:          NewRegistry(),
		Concurrency:       1,
		PollInterval:      20 * time.Millisecond,
		DispatchInterval:  time.Hour,
		SchedulerInterval: time.Hour,
	})
	done := make(chan struct{})
	go func() {
		_ = w.Run(ctx)
		close(done)
	}()

	deadline := time.After(2 * time.Second)
	for {
		live, err := store.ListLiveWorkers(ctx, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if len(live) == 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("worker heartbeat never appeared")
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}

	tasks, err := store.ListRecurring(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, task := range tasks {
		if task.Kind == KindPruneFinished {
			found = true
		}
	}
	if !found {
		t.Fatalf("recurring = %+v, want %s", tasks, KindPruneFinished)
	}

	cancel()
	<-done
	live, err := store.ListLiveWorkers(context.Background(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 0 {
		t.Fatalf("heartbeat survived shutdown: %+v", live)
	}
}

func TestPruneFinishedHandler_deletesOldFinished(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()
	id, err := Enqueue(ctx, store, Options{Kind: "Done"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFinished(ctx, id); err != nil {
		t.Fatal(err)
	}
	h := PruneFinishedHandler(db)
	if err := h(ctx, []byte(`{"older_than_hours":0}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, id); err != ErrNotFound {
		t.Fatalf("expected pruned, err=%v", err)
	}
}

func TestPruneSessionsHandler(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	h := PruneSessionsHandler(db)
	if err := h(ctx, nil); err != nil {
		t.Fatal(err)
	}
}
