package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
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
