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

	"github.com/puppe1990/cais/pkg/cais/jobs"
)

func (c *CLI) cmdJobs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cais jobs <work|status>")
	}
	switch args[0] {
	case "work":
		return c.cmdJobsWork(args[1:])
	case "status":
		return c.cmdJobsStatus()
	default:
		return fmt.Errorf("unknown jobs command %q (use work or status)", args[0])
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
