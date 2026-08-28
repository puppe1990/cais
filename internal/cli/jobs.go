package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/puppe1990/cais/pkg/cais/jobs"
)

func (c *CLI) cmdJobs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cais jobs <work|status|retry|discard|prune>")
	}
	switch args[0] {
	case "work":
		return c.cmdJobsWork(args[1:])
	case "status":
		return c.cmdJobsStatus()
	case "retry":
		return c.cmdJobsRetry(args[1:])
	case "discard":
		return c.cmdJobsDiscard(args[1:])
	case "prune":
		return c.cmdJobsPrune(args[1:])
	default:
		return fmt.Errorf("unknown jobs command %q (use work, status, retry, discard, or prune)", args[0])
	}
}

func (c *CLI) cmdJobsStatus() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	db, _, cleanup, err := openAppDB(dir)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := jobs.EnsureSchema(db); err != nil {
		return err
	}
	store := jobs.NewStore(db)
	ctx := context.Background()

	counts, err := store.CountByStatus(ctx)
	if err != nil {
		return err
	}
	scheduled, err := store.CountScheduled(ctx)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(c.Out, "=> Job queue")
	for _, status := range []string{jobs.StatusReady, jobs.StatusRunning, jobs.StatusFinished, jobs.StatusFailed} {
		_, _ = fmt.Fprintf(c.Out, "  %-9s %d\n", status+":", counts[status])
	}
	_, _ = fmt.Fprintf(c.Out, "  scheduled: %d\n", scheduled)

	if err := c.printJobQueues(ctx, store); err != nil {
		return err
	}
	if err := c.printLiveWorkers(ctx, store); err != nil {
		return err
	}
	return c.printRecurringTasks(ctx, store)
}

func (c *CLI) printLiveWorkers(ctx context.Context, store *jobs.Store) error {
	live, err := store.ListLiveWorkers(ctx, jobs.DefaultWorkerStale)
	if err != nil {
		return err
	}
	if len(live) == 0 {
		_, _ = fmt.Fprintln(c.Out, "=> Workers    none (start with cais jobs work)")
		return nil
	}
	_, _ = fmt.Fprintln(c.Out, "=> Workers")
	for _, w := range live {
		_, _ = fmt.Fprintf(c.Out, "  %s  queues=%s concurrency=%d heartbeat=%s\n",
			w.ID, w.Queues, w.Concurrency, w.HeartbeatAt)
	}
	if len(live) > 1 {
		_, _ = fmt.Fprintln(c.Out, "  warning: multiple workers on one SQLite file serialize writes")
	}
	return nil
}

func (c *CLI) cmdJobsRetry(args []string) error {
	id, err := parseJobID(args)
	if err != nil {
		return err
	}
	return c.withJobStore(func(store *jobs.Store) error {
		if err := store.RetryFailed(context.Background(), id); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.Out, "=> retried job %d\n", id)
		return nil
	})
}

func (c *CLI) cmdJobsDiscard(args []string) error {
	id, err := parseJobID(args)
	if err != nil {
		return err
	}
	return c.withJobStore(func(store *jobs.Store) error {
		if err := store.Discard(context.Background(), id); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.Out, "=> discarded job %d\n", id)
		return nil
	})
}

func (c *CLI) cmdJobsPrune(args []string) error {
	older := time.Duration(0)
	for i := 0; i < len(args); i++ {
		if args[i] == "--older" {
			if i+1 >= len(args) {
				return fmt.Errorf("--older requires a duration (e.g. 24h)")
			}
			d, err := time.ParseDuration(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid --older %q: %w", args[i+1], err)
			}
			older = d
			i++
			continue
		}
		return fmt.Errorf("unknown prune flag %q", args[i])
	}
	return c.withJobStore(func(store *jobs.Store) error {
		n, err := store.PruneFinished(context.Background(), older)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.Out, "=> pruned %d finished job(s)\n", n)
		return nil
	})
}

func (c *CLI) withJobStore(fn func(*jobs.Store) error) error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	db, _, cleanup, err := openAppDB(dir)
	if err != nil {
		return err
	}
	defer cleanup()
	if err := jobs.EnsureSchema(db); err != nil {
		return err
	}
	return fn(jobs.NewStore(db))
}

func parseJobID(args []string) (int64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("usage: cais jobs retry|discard <id>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid job id %q", args[0])
	}
	return id, nil
}

func (c *CLI) printJobQueues(ctx context.Context, store *jobs.Store) error {
	queues, err := store.CountByQueue(ctx)
	if err != nil {
		return err
	}
	if len(queues) == 0 {
		return nil
	}
	_, _ = fmt.Fprintln(c.Out, "=> Queues")
	for _, q := range queues {
		_, _ = fmt.Fprintf(c.Out, "  %-12s ready=%d running=%d finished=%d failed=%d\n",
			q.Queue, q.Ready, q.Running, q.Finished, q.Failed)
	}
	return nil
}

func (c *CLI) printRecurringTasks(ctx context.Context, store *jobs.Store) error {
	tasks, err := store.ListRecurring(ctx)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}
	_, _ = fmt.Fprintln(c.Out, "=> Recurring")
	for _, task := range tasks {
		last := "never"
		if task.LastRun != nil {
			last = task.LastRun.UTC().Format("2006-01-02 15:04")
		}
		_, _ = fmt.Fprintf(c.Out, "  %-20s %s  last=%s\n", task.Kind, task.Cron, last)
	}
	return nil
}

func (c *CLI) cmdJobsWork(args []string) error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	workerMain := filepath.Join(dir, "cmd/worker/main.go")
	if _, err := os.Stat(workerMain); err == nil {
		_, _ = fmt.Fprintln(c.Out, "=> Running cmd/worker")
		return runCmd(dir, "go", append([]string{"run", "./cmd/worker"}, args...)...)
	}
	db, _, cleanup, err := openAppDB(dir)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := jobs.EnsureSchema(db); err != nil {
		return err
	}

	queues := []string{jobs.DefaultQueue}
	concurrency := 2
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--queues":
			if i+1 >= len(args) {
				return fmt.Errorf("--queues requires a value")
			}
			i++
			queues = splitCSV(args[i])
		case "--concurrency":
			if i+1 >= len(args) {
				return fmt.Errorf("--concurrency requires a value")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				return fmt.Errorf("invalid --concurrency %q", args[i])
			}
			concurrency = n
		default:
			return fmt.Errorf("unknown flag %q", args[i])
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker := jobs.NewWorker(jobs.WorkerConfig{
		Store:       jobs.NewStore(db),
		Registry:    jobs.DefaultRegistry(db),
		Queues:      queues,
		Concurrency: concurrency,
	})
	_, _ = fmt.Fprintf(c.Out, "=> Jobs worker (queues=%s, concurrency=%d)\n", strings.Join(queues, ","), concurrency)
	return worker.Run(ctx)
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
