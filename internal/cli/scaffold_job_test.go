package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldJob_createsWorkerAndHandler(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "jobapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "jobapp", ModulePath: "github.com/puppe1990/jobapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldJob(appDir, "send_welcome", jobOpts{}); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/jobs/send_welcome.go",
		"internal/jobs/send_welcome_test.go",
		"internal/jobs/registry.go",
		"cmd/worker/main.go",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	reg, err := os.ReadFile(filepath.Join(appDir, "internal/jobs/registry.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(reg), `reg.Register("SendWelcome", PerformSendWelcome(s))`) {
		t.Errorf("registry missing SendWelcome registration:\n%s", reg)
	}
}

func TestScaffoldJob_withCronPatchesSeeds(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "cronjob")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName: "cronjob", ModulePath: "github.com/puppe1990/cronjob",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldJob(appDir, "prune_sessions", jobOpts{Cron: "0 3 * * *"}); err != nil {
		t.Fatal(err)
	}

	seeds, err := os.ReadFile(filepath.Join(appDir, "internal/db/seeds.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(seeds)
	if !strings.Contains(body, `Kind: "PruneSessions"`) || !strings.Contains(body, `Cron: "0 3 * * *"`) {
		t.Fatalf("seeds missing recurring job:\n%s", body)
	}
}
