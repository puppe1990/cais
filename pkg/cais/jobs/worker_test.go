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

func TestPruneSessionsHandler(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	h := PruneSessionsHandler(db)
	if err := h(ctx, nil); err != nil {
		t.Fatal(err)
	}
}
