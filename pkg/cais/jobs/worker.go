package jobs

import (
	"context"
	"log"
	"os"
	"strings"
	"time"
)

// WorkerConfig configures the job worker and dispatcher loops.
type WorkerConfig struct {
	Store             *Store
	Registry          *Registry
	Queues            []string
	Concurrency       int
	PollInterval      time.Duration
	DispatchInterval  time.Duration
	SchedulerInterval time.Duration
	Logger            *log.Logger
}

// Worker processes jobs from SQLite.
type Worker struct {
	cfg WorkerConfig
	id  string
}

func NewWorker(cfg WorkerConfig) *Worker {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = time.Second
	}
	if cfg.DispatchInterval <= 0 {
		cfg.DispatchInterval = time.Second
	}
	if cfg.SchedulerInterval <= 0 {
		cfg.SchedulerInterval = time.Minute
	}
	if len(cfg.Queues) == 0 {
		cfg.Queues = []string{DefaultQueue}
	}
	if cfg.Concurrency < 1 {
		cfg.Concurrency = 1
	}
	if cfg.Registry == nil {
		cfg.Registry = NewRegistry()
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Worker{cfg: cfg}
}

// Run starts dispatcher and worker goroutines until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	w.id = NewWorkerID()
	pulse := w.pulse()
	if err := w.cfg.Store.TouchWorker(ctx, pulse); err != nil {
		w.cfg.Logger.Printf("jobs heartbeat: %v", err)
	}
	defer func() { _ = w.cfg.Store.RemoveWorker(context.Background(), w.id) }()

	// Recover jobs whose worker heartbeat is gone (#172) without stealing live work.
	if n, err := w.cfg.Store.RequeueOrphaned(ctx, DefaultWorkerStale); err != nil {
		w.cfg.Logger.Printf("jobs requeue-orphaned: %v", err)
	} else if n > 0 {
		w.cfg.Logger.Printf("jobs requeued %d orphaned job(s) from previous run", n)
	}

	if err := w.ensureFinishedPrune(ctx); err != nil {
		w.cfg.Logger.Printf("jobs prune-finished recurring: %v", err)
	}

	errCh := make(chan error, w.cfg.Concurrency+2)

	go func() {
		ticker := time.NewTicker(w.cfg.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.cfg.Store.TouchWorker(ctx, pulse); err != nil {
					w.cfg.Logger.Printf("jobs heartbeat: %v", err)
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(w.cfg.DispatchInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := w.cfg.Store.DispatchDue(ctx, time.Now().UTC()); err != nil {
					w.cfg.Logger.Printf("jobs dispatcher: %v", err)
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(w.cfg.SchedulerInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := RunScheduler(ctx, w.cfg.Store, time.Now().UTC()); err != nil {
					w.cfg.Logger.Printf("jobs scheduler: %v", err)
				} else if n > 0 {
					w.cfg.Logger.Printf("jobs scheduler: enqueued %d recurring task(s)", n)
				}
			}
		}
	}()

	for i := 0; i < w.cfg.Concurrency; i++ {
		go func() {
			ticker := time.NewTicker(w.cfg.PollInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := w.pollOnce(ctx); err != nil {
						errCh <- err
						return
					}
				}
			}
		}()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (w *Worker) pollOnce(ctx context.Context) error {
	for _, queue := range w.cfg.Queues {
		job, err := w.cfg.Store.ClaimFor(ctx, queue, w.id)
		if err != nil {
			return err
		}
		if job == nil {
			continue
		}
		w.runJob(ctx, job)
	}
	return nil
}

func (w *Worker) runJob(ctx context.Context, job *Job) {
	err := w.cfg.Registry.Perform(ctx, job.Kind, job.Payload)
	if err == nil {
		if markErr := w.cfg.Store.MarkFinished(ctx, job.ID); markErr != nil {
			w.cfg.Logger.Printf("jobs finish id=%d: %v", job.ID, markErr)
		}
		w.cfg.Logger.Printf("jobs finished id=%d kind=%s", job.ID, job.Kind)
		return
	}
	if markErr := w.cfg.Store.MarkFailed(ctx, job.ID, err, job.Attempts, job.MaxAttempts); markErr != nil {
		w.cfg.Logger.Printf("jobs fail id=%d: %v (mark: %v)", job.ID, err, markErr)
		return
	}
	w.cfg.Logger.Printf("jobs failed id=%d kind=%s: %v", job.ID, job.Kind, err)
}

func (w *Worker) pulse() WorkerPulse {
	host, _ := os.Hostname()
	return WorkerPulse{
		ID:          w.id,
		Hostname:    host,
		PID:         os.Getpid(),
		Queues:      strings.Join(w.cfg.Queues, ","),
		Concurrency: w.cfg.Concurrency,
	}
}

func (w *Worker) ensureFinishedPrune(ctx context.Context) error {
	return w.cfg.Store.UpsertRecurring(ctx, RecurringOptions{
		Kind:    KindPruneFinished,
		Cron:    pruneFinishedCron,
		Payload: map[string]any{"older_than_hours": 24},
	})
}
