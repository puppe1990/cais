package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound  = errors.New("job not found")
	ErrNotFailed = errors.New("job is not failed")
)

const defaultListLimit = 50
const maxListLimit = 200

// JobRecord is a jobs row for dashboards and CLI inspection.
type JobRecord struct {
	ID          int64
	Queue       string
	Kind        string
	Payload     string
	Priority    int
	Status      string
	Attempts    int
	MaxAttempts int
	LastError   string
	RunAt       string
	CreatedAt   string
	FinishedAt  string
	StartedAt   string
	WorkerID    string
}

// ScheduledRecord is a delayed row waiting for the dispatcher.
type ScheduledRecord struct {
	ID          int64
	Queue       string
	Kind        string
	Payload     string
	Priority    int
	MaxAttempts int
	RunAt       string
	CreatedAt   string
}

// ListFilter selects jobs for inspection. Empty Status/Queue means all.
type ListFilter struct {
	Status string
	Queue  string
	Kind   string
	Limit  int
}

// QueueCount is per-queue tallies by status.
type QueueCount struct {
	Queue    string
	Ready    int
	Running  int
	Finished int
	Failed   int
}

func (f ListFilter) limit() int {
	if f.Limit <= 0 {
		return defaultListLimit
	}
	if f.Limit > maxListLimit {
		return maxListLimit
	}
	return f.Limit
}

// List returns matching jobs, newest first.
func (s *Store) List(ctx context.Context, f ListFilter) ([]JobRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, queue, kind, payload, priority, status, attempts, max_attempts,
       COALESCE(last_error, ''), run_at, created_at, COALESCE(finished_at, ''),
       COALESCE(started_at, ''), COALESCE(worker_id, '')
FROM jobs
WHERE (? = '' OR status = ?) AND (? = '' OR queue = ?) AND (? = '' OR kind = ?)
ORDER BY id DESC
LIMIT ?`, f.Status, f.Status, f.Queue, f.Queue, f.Kind, f.Kind, f.limit())
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanJobRecords(rows)
}

// Get returns one job by id.
func (s *Store) Get(ctx context.Context, id int64) (*JobRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, queue, kind, payload, priority, status, attempts, max_attempts,
       COALESCE(last_error, ''), run_at, created_at, COALESCE(finished_at, ''),
       COALESCE(started_at, ''), COALESCE(worker_id, '')
FROM jobs WHERE id = ?`, id)
	rec, err := scanJobRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get job %d: %w", id, err)
	}
	return rec, nil
}

// RetryFailed moves a failed job back to ready with attempts reset.
func (s *Store) RetryFailed(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE jobs
SET status = ?, attempts = 0, run_at = datetime('now')
WHERE id = ? AND status = ?`, StatusReady, id, StatusFailed)
	if err != nil {
		return fmt.Errorf("retry job %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("retry job %d: %w", id, err)
	}
	if n == 0 {
		if _, getErr := s.Get(ctx, id); errors.Is(getErr, ErrNotFound) {
			return ErrNotFound
		}
		return ErrNotFailed
	}
	return nil
}

// Discard deletes a failed job.
func (s *Store) Discard(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM jobs WHERE id = ? AND status = ?`, id, StatusFailed)
	if err != nil {
		return fmt.Errorf("discard job %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("discard job %d: %w", id, err)
	}
	if n == 0 {
		if _, getErr := s.Get(ctx, id); errors.Is(getErr, ErrNotFound) {
			return ErrNotFound
		}
		return ErrNotFailed
	}
	return nil
}

// ListScheduled returns delayed jobs, soonest first.
func (s *Store) ListScheduled(ctx context.Context, limit int) ([]ScheduledRecord, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, queue, kind, payload, priority, max_attempts, run_at, created_at
FROM scheduled_jobs
ORDER BY run_at ASC, id ASC
LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list scheduled: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []ScheduledRecord
	for rows.Next() {
		var rec ScheduledRecord
		if err := rows.Scan(&rec.ID, &rec.Queue, &rec.Kind, &rec.Payload, &rec.Priority, &rec.MaxAttempts, &rec.RunAt, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// CountByQueue returns status counts grouped by queue name.
func (s *Store) CountByQueue(ctx context.Context) ([]QueueCount, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT queue, status, COUNT(*)
FROM jobs
GROUP BY queue, status
ORDER BY queue`)
	if err != nil {
		return nil, fmt.Errorf("count by queue: %w", err)
	}
	defer func() { _ = rows.Close() }()

	byQueue := map[string]*QueueCount{}
	var order []string
	for rows.Next() {
		var queue, status string
		var n int
		if err := rows.Scan(&queue, &status, &n); err != nil {
			return nil, err
		}
		qc, ok := byQueue[queue]
		if !ok {
			qc = &QueueCount{Queue: queue}
			byQueue[queue] = qc
			order = append(order, queue)
		}
		switch status {
		case StatusReady:
			qc.Ready = n
		case StatusRunning:
			qc.Running = n
		case StatusFinished:
			qc.Finished = n
		case StatusFailed:
			qc.Failed = n
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]QueueCount, 0, len(order))
	for _, name := range order {
		out = append(out, *byQueue[name])
	}
	return out, nil
}

// PruneFinished deletes finished jobs. olderThan <= 0 deletes all finished rows.
func (s *Store) PruneFinished(ctx context.Context, olderThan time.Duration) (int64, error) {
	var (
		res sql.Result
		err error
	)
	if olderThan <= 0 {
		res, err = s.db.ExecContext(ctx, `DELETE FROM jobs WHERE status = ?`, StatusFinished)
	} else {
		res, err = s.db.ExecContext(ctx, `
DELETE FROM jobs
WHERE status = ? AND finished_at <= datetime('now', ?)`,
			StatusFinished, fmt.Sprintf("-%d seconds", int(olderThan.Seconds())))
	}
	if err != nil {
		return 0, fmt.Errorf("prune finished: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune finished: %w", err)
	}
	return n, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanJobRecord(row scanner) (*JobRecord, error) {
	var rec JobRecord
	if err := row.Scan(
		&rec.ID, &rec.Queue, &rec.Kind, &rec.Payload, &rec.Priority, &rec.Status,
		&rec.Attempts, &rec.MaxAttempts, &rec.LastError, &rec.RunAt, &rec.CreatedAt, &rec.FinishedAt,
		&rec.StartedAt, &rec.WorkerID,
	); err != nil {
		return nil, err
	}
	return &rec, nil
}

func scanJobRecords(rows *sql.Rows) ([]JobRecord, error) {
	var out []JobRecord
	for rows.Next() {
		rec, err := scanJobRecord(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rec)
	}
	return out, rows.Err()
}
