package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Job is a claimed row ready for execution.
type Job struct {
	ID          int64
	Queue       string
	Kind        string
	Payload     []byte
	Priority    int
	Attempts    int
	MaxAttempts int
}

// Store persists jobs in SQLite.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) insertReady(ctx context.Context, opts Options, payload []byte, runAt time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
INSERT INTO jobs (queue, kind, payload, priority, max_attempts, run_at, status)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		opts.Queue, opts.Kind, string(payload), opts.Priority, opts.MaxAttempts,
		formatTime(runAt), StatusReady,
	)
	if err != nil {
		return 0, fmt.Errorf("insert job: %w", err)
	}
	return res.LastInsertId()
}

func (s *Store) insertScheduled(ctx context.Context, opts Options, payload []byte, runAt time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
INSERT INTO scheduled_jobs (queue, kind, payload, priority, max_attempts, run_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		opts.Queue, opts.Kind, string(payload), opts.Priority, opts.MaxAttempts, formatTime(runAt),
	)
	if err != nil {
		return 0, fmt.Errorf("insert scheduled job: %w", err)
	}
	return res.LastInsertId()
}

// DispatchDue moves scheduled jobs whose run_at has passed into the ready queue.
func (s *Store) DispatchDue(ctx context.Context, now time.Time) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
SELECT queue, kind, payload, priority, max_attempts, run_at
FROM scheduled_jobs
WHERE run_at <= ?
ORDER BY run_at ASC, id ASC`, formatTime(now))
	if err != nil {
		return 0, fmt.Errorf("select scheduled: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type row struct {
		queue, kind, payload string
		priority, max        int
		runAt                string
	}
	var pending []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.queue, &r.kind, &r.payload, &r.priority, &r.max, &r.runAt); err != nil {
			return 0, err
		}
		pending = append(pending, r)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, r := range pending {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO jobs (queue, kind, payload, priority, max_attempts, run_at, status)
VALUES (?, ?, ?, ?, ?, datetime('now'), ?)`,
			r.queue, r.kind, r.payload, r.priority, r.max, StatusReady,
		); err != nil {
			return 0, fmt.Errorf("promote scheduled: %w", err)
		}
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM scheduled_jobs WHERE run_at <= ?`, formatTime(now))
	if err != nil {
		return 0, fmt.Errorf("delete scheduled: %w", err)
	}
	n, _ := res.RowsAffected()
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(n), nil
}

// Claim reserves the next ready job in queue.
// Uses a single UPDATE ... RETURNING (SQLite 3.35+) instead of FOR UPDATE SKIP LOCKED.
func (s *Store) Claim(ctx context.Context, queue string) (*Job, error) {
	var j Job
	var payload string
	err := s.db.QueryRowContext(ctx, `
UPDATE jobs
SET status = ?, attempts = attempts + 1, last_error = NULL
WHERE id = (
  SELECT id FROM jobs
  WHERE queue = ? AND status = ? AND run_at <= datetime('now')
  ORDER BY priority ASC, id ASC
  LIMIT 1
)
RETURNING id, kind, payload, priority, attempts, max_attempts`,
		StatusRunning, queue, StatusReady,
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

// MarkFinished sets status to finished.
func (s *Store) MarkFinished(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE jobs SET status = ?, finished_at = datetime('now') WHERE id = ?`,
		StatusFinished, id,
	)
	return err
}

// MarkFailed records failure; requeues with backoff or marks failed permanently.
func (s *Store) MarkFailed(ctx context.Context, id int64, jobErr error, attempts, maxAttempts int) error {
	msg := ""
	if jobErr != nil {
		msg = jobErr.Error()
	}
	if attempts < maxAttempts {
		backoff := retryDelay(attempts)
		_, err := s.db.ExecContext(ctx, `
UPDATE jobs SET status = ?, last_error = ?, run_at = datetime('now', ?) WHERE id = ?`,
			StatusReady, msg, fmt.Sprintf("+%d seconds", int(backoff.Seconds())), id,
		)
		return err
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE jobs SET status = ?, last_error = ?, finished_at = datetime('now') WHERE id = ?`,
		StatusFailed, msg, id,
	)
	return err
}

// CountByStatus returns job counts grouped by status.
func (s *Store) CountByStatus(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM jobs GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]int{}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		out[status] = n
	}
	return out, rows.Err()
}

// CountScheduled returns rows waiting in scheduled_jobs.
func (s *Store) CountScheduled(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM scheduled_jobs`).Scan(&n)
	return n, err
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	sec := 1 << min(attempt, 6) // cap at 64s
	return time.Duration(sec) * time.Second
}

func formatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}
