package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais/jobs"
)

func TestCLI_JobsStatusEmpty(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "jobsapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "jobsapp", ModulePath: "github.com/puppe1990/jobsapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.cmdJobsStatus(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ready:") {
		t.Fatalf("output = %q", buf.String())
	}
}

func TestCLI_JobsEnqueueAndStatus(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "jobsq")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "jobsq", ModulePath: "github.com/puppe1990/jobsq",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	db, _, cleanup, err := openAppDB(appDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if err := jobs.EnsureSchema(db); err != nil {
		t.Fatal(err)
	}
	store := jobs.NewStore(db)
	if _, err := jobs.Enqueue(context.Background(), store, jobs.Options{Kind: jobs.KindPruneSessions}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.cmdJobsStatus(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ready:") || !strings.Contains(buf.String(), "1") {
		t.Fatalf("output = %q", buf.String())
	}
}

func TestCLI_JobsStatus_showsQueueAndRecurring(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "jobsdash")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "jobsdash", ModulePath: "github.com/puppe1990/jobsdash",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	db, _, cleanup, err := openAppDB(appDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if err := jobs.EnsureSchema(db); err != nil {
		t.Fatal(err)
	}
	store := jobs.NewStore(db)
	if _, err := jobs.Enqueue(context.Background(), store, jobs.Options{Kind: "Mail", Queue: "mail"}); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertRecurring(context.Background(), jobs.RecurringOptions{
		Kind: jobs.KindPruneSessions, Cron: "0 3 * * *",
	}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.cmdJobsStatus(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"mail", "PruneSessions", "0 3 * * *"} {
		if !strings.Contains(out, want) {
			t.Errorf("status missing %q: %q", want, out)
		}
	}
}

func TestCLI_JobsRetryAndDiscard(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "jobsact")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "jobsact", ModulePath: "github.com/puppe1990/jobsact",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	db, _, cleanup, err := openAppDB(appDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if err := jobs.EnsureSchema(db); err != nil {
		t.Fatal(err)
	}
	store := jobs.NewStore(db)
	ctx := context.Background()
	id, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Dead", MaxAttempts: 1})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.Claim(ctx, jobs.DefaultQueue)
	if err != nil || job == nil {
		t.Fatal(err)
	}
	if err := store.MarkFailed(ctx, job.ID, errTestJob, job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}

	c := &CLI{Out: &bytes.Buffer{}}
	if err := c.cmdJobs([]string{"retry", itoa64(id)}); err != nil {
		t.Fatal(err)
	}
	rec, err := store.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != jobs.StatusReady {
		t.Fatalf("after retry status=%q", rec.Status)
	}

	job, err = store.Claim(ctx, jobs.DefaultQueue)
	if err != nil || job == nil {
		t.Fatal(err)
	}
	if err := store.MarkFailed(ctx, job.ID, errTestJob, job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}
	if err := c.cmdJobs([]string{"discard", itoa64(id)}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, id); err != jobs.ErrNotFound {
		t.Fatalf("expected discarded, err=%v", err)
	}
}

func TestCLI_JobsPrune(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "jobsprune")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "jobsprune", ModulePath: "github.com/puppe1990/jobsprune",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	db, _, cleanup, err := openAppDB(appDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if err := jobs.EnsureSchema(db); err != nil {
		t.Fatal(err)
	}
	store := jobs.NewStore(db)
	ctx := context.Background()
	id, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Done"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFinished(ctx, id); err != nil {
		t.Fatal(err)
	}

	c := &CLI{Out: &bytes.Buffer{}}
	if err := c.cmdJobs([]string{"prune"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, id); err != jobs.ErrNotFound {
		t.Fatalf("expected pruned, err=%v", err)
	}
}

var errTestJob = errJobFailed{}

type errJobFailed struct{}

func (errJobFailed) Error() string { return "boom" }

func itoa64(id int64) string {
	return strconv.FormatInt(id, 10)
}
