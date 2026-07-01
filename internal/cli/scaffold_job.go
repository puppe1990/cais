package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type jobOpts struct {
	Cron   string
	dryRun bool
}

func parseJobOpts(args []string) (jobOpts, error) {
	opts := jobOpts{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--cron":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--cron requires a value")
			}
			i++
			opts.Cron = args[i]
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opts, nil
}

func scaffoldJob(dir, name string, opts jobOpts) error {
	data := dataForHandler(name)
	data.ModulePath = readModulePath(dir)

	files := map[string]string{
		filepath.Join("internal/jobs", data.Snake+".go"):      buildJobHandler(data),
		filepath.Join("internal/jobs", data.Snake+"_test.go"): buildJobTest(data),
	}

	workerPath := filepath.Join(dir, "cmd/worker/main.go")
	if _, err := os.Stat(workerPath); os.IsNotExist(err) {
		files["cmd/worker/main.go"] = tplWorker
	}
	registryPath := filepath.Join(dir, "internal/jobs/registry.go")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		files["internal/jobs/registry.go"] = tplJobRegistry
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if path != filepath.Join("internal/jobs", data.Snake+".go") &&
			path != filepath.Join("internal/jobs", data.Snake+"_test.go") {
			if err := writeScaffoldTemplate(full, content, data, path, opts.dryRun); err != nil {
				return err
			}
			continue
		}
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := writeScaffoldFile(full, []byte(content), 0o644, path, opts.dryRun); err != nil {
			return err
		}
	}

	if err := patchJobRegistry(dir, data, opts.dryRun); err != nil {
		return err
	}
	if opts.Cron != "" {
		if err := patchJobRecurringSeed(dir, data, opts, opts.dryRun); err != nil {
			return err
		}
	}
	if opts.dryRun {
		return nil
	}
	return gofmtGoFiles(dir)
}

func buildJobHandler(data scaffoldData) string {
	return fmt.Sprintf(`package jobs

import (
	"context"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"
	"%s/internal/store"
)

// Perform%s runs the %s job.
func Perform%s(s store.Store) caisjobs.Handler {
	return func(ctx context.Context, payload []byte) error {
		_ = ctx
		_ = payload
		_ = s
		return nil
	}
}
`, data.ModulePath, data.Pascal, data.Snake, data.Pascal)
}

func buildJobTest(data scaffoldData) string {
	return fmt.Sprintf(`package jobs

import (
	"context"
	"testing"

	"%s/internal/store"
)

func TestPerform%s(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := Perform%s(s)
	if err := h(context.Background(), []byte("{}")); err != nil {
		t.Fatal(err)
	}
}
`, data.ModulePath, data.Pascal, data.Pascal)
}

func patchJobRegistry(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/jobs/registry.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	line := fmt.Sprintf("\treg.Register(%q, Perform%s(s))\n", data.Pascal, data.Pascal)
	if strings.Contains(content, line) {
		return nil
	}
	marker := "\t// cais:jobs-register"
	if !strings.Contains(content, marker) {
		return fmt.Errorf("internal/jobs/registry.go: missing %s marker", marker)
	}
	content = strings.Replace(content, marker, line+marker, 1)
	return updateScaffoldFile(path, []byte(content), "internal/jobs/registry.go", dryRun)
}

func patchJobRecurringSeed(dir string, data scaffoldData, opts jobOpts, dryRun bool) error {
	path := filepath.Join(dir, "internal/db/seeds.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("internal/db/seeds.go not found — run cais new first")
	}
	content := string(body)
	needle := fmt.Sprintf(`Kind: %q`, data.Pascal)
	if strings.Contains(content, needle) {
		return nil
	}
	block := fmt.Sprintf(`
	if err := jobStore.UpsertRecurring(context.Background(), caisjobs.RecurringOptions{
		Kind: %q,
		Cron: %q,
	}); err != nil {
		return err
	}
`, data.Pascal, opts.Cron)
	marker := "\t// cais:recurring-seeds"
	if !strings.Contains(content, marker) {
		return fmt.Errorf("seeds.go: missing %s marker — add it inside RunSeeds", marker)
	}
	content = strings.Replace(content, marker, block+marker, 1)

	if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/jobs") {
		content = strings.Replace(content,
			`import (`,
			`import (
	"context"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"`,
			1,
		)
	} else if !strings.Contains(content, `"context"`) {
		content = strings.Replace(content,
			`import (`,
			`import (
	"context"`,
			1,
		)
	}
	if !strings.Contains(content, "jobStore :=") {
		insert := `
	jobStore := caisjobs.NewStore(s.DB())
	if err := caisjobs.EnsureSchema(s.DB()); err != nil {
		return err
	}
`
		content = strings.Replace(content, "func RunSeeds(s store.Store) error {",
			"func RunSeeds(s store.Store) error {"+insert, 1)
	}
	return updateScaffoldFile(path, []byte(content), "internal/db/seeds.go", dryRun)
}

const tplJobRegistry = `package jobs

import (
	"database/sql"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"
	"{{.ModulePath}}/internal/store"
)

// RegisterAll wires app and framework jobs into the worker registry.
func RegisterAll(reg *caisjobs.Registry, db *sql.DB, s store.Store) {
	reg.Register(caisjobs.KindPruneSessions, caisjobs.PruneSessionsHandler(db))
	// cais:jobs-register
}
`

const tplWorker = `package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/puppe1990/cais/pkg/cais"
	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"
	appjobs "{{.ModulePath}}/internal/jobs"
	"{{.ModulePath}}/internal/store"
)

func main() {
	queues := flag.String("queues", "default", "comma-separated queue names")
	concurrency := flag.Int("concurrency", 2, "worker goroutines")
	flag.Parse()

	cfg := cais.Load()
	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	if err := caisjobs.EnsureSchema(s.DB()); err != nil {
		log.Fatal(err)
	}

	reg := caisjobs.NewRegistry()
	appjobs.RegisterAll(reg, s.DB(), s)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker := caisjobs.NewWorker(caisjobs.WorkerConfig{
		Store:       caisjobs.NewStore(s.DB()),
		Registry:    reg,
		Queues:      splitQueues(*queues),
		Concurrency: *concurrency,
	})
	log.Printf("=> Worker started (queues=%s, concurrency=%d)", *queues, *concurrency)
	if err := worker.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}

func splitQueues(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return []string{caisjobs.DefaultQueue}
	}
	return out
}
`
