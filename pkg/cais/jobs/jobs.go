package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	StatusReady    = "ready"
	StatusRunning  = "running"
	StatusFinished = "finished"
	StatusFailed   = "failed"

	DefaultQueue       = "default"
	DefaultMaxAttempts = 3
)

// Handler runs a job payload.
type Handler func(ctx context.Context, payload []byte) error

// Options describes a job to enqueue.
type Options struct {
	Queue       string
	Kind        string
	Payload     any
	Priority    int
	MaxAttempts int
}

func (o Options) normalized() (Options, []byte, error) {
	if o.Queue == "" {
		o.Queue = DefaultQueue
	}
	if o.Kind == "" {
		return o, nil, fmt.Errorf("job kind is required")
	}
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = DefaultMaxAttempts
	}
	var raw []byte
	switch p := o.Payload.(type) {
	case nil:
		raw = []byte("{}")
	case []byte:
		raw = p
	default:
		b, err := json.Marshal(p)
		if err != nil {
			return o, nil, fmt.Errorf("marshal payload: %w", err)
		}
		raw = b
	}
	return o, raw, nil
}

// Enqueue inserts a ready job.
func Enqueue(ctx context.Context, store *Store, opts Options) (int64, error) {
	o, raw, err := opts.normalized()
	if err != nil {
		return 0, err
	}
	return store.insertReady(ctx, o, raw, time.Now().UTC())
}

// SetWait schedules a job for a future run (dispatcher promotes to jobs).
func SetWait(ctx context.Context, store *Store, wait time.Duration, opts Options) (int64, error) {
	o, raw, err := opts.normalized()
	if err != nil {
		return 0, err
	}
	return store.insertScheduled(ctx, o, raw, time.Now().UTC().Add(wait))
}
