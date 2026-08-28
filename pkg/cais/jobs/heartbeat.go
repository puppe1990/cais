package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

const DefaultWorkerStale = 15 * time.Second

// WorkerPulse is a live worker heartbeat row.
type WorkerPulse struct {
	ID          string
	Hostname    string
	PID         int
	Queues      string
	Concurrency int
	HeartbeatAt string
	StartedAt   string
}

// NewWorkerID returns a unique id for this process.
func NewWorkerID() string {
	host, _ := os.Hostname()
	if host == "" {
		host = "worker"
	}
	return fmt.Sprintf("%s-%d-%d", host, os.Getpid(), time.Now().UnixNano())
}

// TouchWorker upserts a heartbeat so the dashboard can tell the worker is alive.
func (s *Store) TouchWorker(ctx context.Context, pulse WorkerPulse) error {
	if pulse.ID == "" {
		return fmt.Errorf("worker id is required")
	}
	if pulse.Queues == "" {
		pulse.Queues = DefaultQueue
	}
	if pulse.Concurrency < 1 {
		pulse.Concurrency = 1
	}
	now := formatTime(time.Now().UTC())
	_, err := s.db.ExecContext(ctx, `
INSERT INTO job_workers (id, hostname, pid, queues, concurrency, heartbeat_at, started_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  hostname = excluded.hostname,
  pid = excluded.pid,
  queues = excluded.queues,
  concurrency = excluded.concurrency,
  heartbeat_at = excluded.heartbeat_at`,
		pulse.ID, pulse.Hostname, pulse.PID, pulse.Queues, pulse.Concurrency, now, now,
	)
	if err != nil {
		return fmt.Errorf("touch worker: %w", err)
	}
	return nil
}

// ListLiveWorkers returns heartbeats newer than stale.
func (s *Store) ListLiveWorkers(ctx context.Context, stale time.Duration) ([]WorkerPulse, error) {
	if stale <= 0 {
		stale = DefaultWorkerStale
	}
	mod := fmt.Sprintf("-%d seconds", int(stale.Seconds()))
	if stale < time.Second {
		mod = "-1 seconds"
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, COALESCE(hostname, ''), pid, queues, concurrency, heartbeat_at, started_at
FROM job_workers
WHERE heartbeat_at >= datetime('now', ?)
ORDER BY heartbeat_at DESC`, mod)
	if err != nil {
		return nil, fmt.Errorf("list workers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []WorkerPulse
	for rows.Next() {
		var p WorkerPulse
		if err := rows.Scan(&p.ID, &p.Hostname, &p.PID, &p.Queues, &p.Concurrency, &p.HeartbeatAt, &p.StartedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// RemoveWorker deletes a heartbeat on clean shutdown.
func (s *Store) RemoveWorker(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM job_workers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("remove worker %s: %w", id, err)
	}
	return nil
}

// RequeueOrphaned returns running jobs whose worker is missing or stale.
// Live workers keep their in-flight jobs — unlike RequeueStuck, this is safe
// to call while a worker is running (dashboard button, overlapping processes).
func (s *Store) RequeueOrphaned(ctx context.Context, stale time.Duration) (int64, error) {
	if stale <= 0 {
		stale = DefaultWorkerStale
	}
	mod := fmt.Sprintf("-%d seconds", int(stale.Seconds()))
	res, err := s.db.ExecContext(ctx, `
UPDATE jobs
SET status = ?, worker_id = NULL, started_at = NULL
WHERE status = ?
  AND (
    worker_id IS NULL OR worker_id = ''
    OR worker_id NOT IN (
      SELECT id FROM job_workers WHERE heartbeat_at >= datetime('now', ?)
    )
  )`, StatusReady, StatusRunning, mod)
	if err != nil {
		return 0, fmt.Errorf("requeue orphaned: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("requeue orphaned: %w", err)
	}
	return n, nil
}

// ClaimFor is Claim tagged with the worker id so RequeueOrphaned can skip live work.
func (s *Store) ClaimFor(ctx context.Context, queue, workerID string) (*Job, error) {
	var j Job
	var payload string
	err := s.db.QueryRowContext(ctx, `
UPDATE jobs
SET status = ?, attempts = attempts + 1, last_error = NULL,
    started_at = datetime('now'), worker_id = ?
WHERE id = (
  SELECT id FROM jobs
  WHERE queue = ? AND status = ? AND run_at <= datetime('now')
  ORDER BY priority ASC, id ASC
  LIMIT 1
)
RETURNING id, kind, payload, priority, attempts, max_attempts`,
		StatusRunning, workerID, queue, StatusReady,
	).Scan(&j.ID, &j.Kind, &payload, &j.Priority, &j.Attempts, &j.MaxAttempts)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim job: %w", err)
	}
	j.Queue = queue
	j.Payload = []byte(payload)
	return &j, nil
}
