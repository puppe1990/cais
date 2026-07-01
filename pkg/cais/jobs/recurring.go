package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RecurringTask is a cron-driven job definition.
type RecurringTask struct {
	ID      int64
	Queue   string
	Kind    string
	Payload []byte
	Cron    string
	LastRun *time.Time
}

// RecurringOptions registers a recurring task.
type RecurringOptions struct {
	Queue   string
	Kind    string
	Payload any
	Cron    string
}

func (s *Store) UpsertRecurring(ctx context.Context, opts RecurringOptions) error {
	o, raw, err := Options{Queue: opts.Queue, Kind: opts.Kind, Payload: opts.Payload}.normalized()
	if err != nil {
		return err
	}
	if opts.Cron == "" {
		return fmt.Errorf("cron expression is required")
	}
	if err := ValidateCron(opts.Cron); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO recurring_tasks (queue, kind, payload, cron, enabled)
VALUES (?, ?, ?, ?, 1)
ON CONFLICT(kind) DO UPDATE SET
  queue = excluded.queue,
  payload = excluded.payload,
  cron = excluded.cron,
  enabled = 1`,
		o.Queue, o.Kind, string(raw), opts.Cron,
	)
	return err
}

func (s *Store) ListRecurring(ctx context.Context) ([]RecurringTask, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, queue, kind, payload, cron, last_run
FROM recurring_tasks
WHERE enabled = 1
ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []RecurringTask
	for rows.Next() {
		var t RecurringTask
		var payload string
		var lastRun sql.NullString
		if err := rows.Scan(&t.ID, &t.Queue, &t.Kind, &payload, &t.Cron, &lastRun); err != nil {
			return nil, err
		}
		t.Payload = []byte(payload)
		if lastRun.Valid {
			parsed, err := time.ParseInLocation("2006-01-02 15:04:05", lastRun.String, time.UTC)
			if err == nil {
				t.LastRun = &parsed
			}
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) UpdateRecurringLastRun(ctx context.Context, id int64, at time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE recurring_tasks SET last_run = ? WHERE id = ?`,
		formatTime(at), id,
	)
	return err
}

// RunScheduler enqueues jobs for due recurring tasks.
func RunScheduler(ctx context.Context, store *Store, now time.Time) (int, error) {
	tasks, err := store.ListRecurring(ctx)
	if err != nil {
		return 0, err
	}
	var n int
	for _, task := range tasks {
		run, err := ShouldRunRecurring(task.Cron, task.LastRun, now)
		if err != nil {
			return n, err
		}
		if !run {
			continue
		}
		if _, err := store.insertReady(ctx, Options{
			Queue: task.Queue, Kind: task.Kind, Payload: task.Payload,
		}, task.Payload, now); err != nil {
			return n, err
		}
		if err := store.UpdateRecurringLastRun(ctx, task.ID, now); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
