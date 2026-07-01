package cli

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/migrate"
	"github.com/puppe1990/cais/pkg/cais/session"

	_ "modernc.org/sqlite"
)

func (c *CLI) cmdDB(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cais db <migrate|status|rollback|prune-sessions|seed>")
	}
	switch args[0] {
	case "migrate":
		return c.cmdDBMigrate()
	case "status":
		return c.cmdDBStatus()
	case "rollback":
		return c.cmdDBRollback()
	case "prune-sessions":
		return c.cmdDBPruneSessions()
	case "seed":
		return c.cmdDBSeed(args[1:])
	default:
		return fmt.Errorf("unknown db command %q (use migrate, status, rollback, prune-sessions, or seed)", args[0])
	}
}

func (c *CLI) cmdDBSeed(args []string) error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	seedsPath := filepath.Join(dir, "internal/db/seeds.go")
	if _, err := os.Stat(seedsPath); err != nil {
		return fmt.Errorf("internal/db/seeds.go not found")
	}

	for _, arg := range args {
		if arg == "--list" {
			items, err := listSeedsInDir(dir)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				_, _ = fmt.Fprintln(c.Out, "=> RunSeeds (no named seed helpers found)")
				return nil
			}
			_, _ = fmt.Fprintln(c.Out, "=> Seeds:")
			for _, item := range items {
				_, _ = fmt.Fprintf(c.Out, "  - %s\n", item)
			}
			return nil
		}
	}
	runnerPath := filepath.Join(dir, "internal/db/runseed_main.go")
	if _, err := os.Stat(runnerPath); err != nil {
		module := moduleFromDir(dir)
		content := fmt.Sprintf(tplRunSeed, module, module)
		if err := os.MkdirAll(filepath.Dir(runnerPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(runnerPath, []byte(content), 0o644); err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintln(c.Out, "=> Running seeds")
	if err := runCmd(dir, "go", "run", "./internal/db/runseed_main.go"); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(c.Out, "=> Seeds complete")
	return nil
}

const tplRunSeed = `//go:build ignore

package main

import (
	"log"

	"%s/internal/db"
	"%s/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
)

func main() {
	cfg := cais.Load()
	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if err := db.RunSeeds(s); err != nil {
		log.Fatal(err)
	}
}
`

func (c *CLI) cmdDBMigrate() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	db, migrationsDir, cleanup, err := openAppDB(dir)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := migrate.ApplyDir(db, migrationsDir); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(c.Out, "=> Migrations up to date")
	return nil
}

func (c *CLI) cmdDBRollback() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	db, migrationsDir, cleanup, err := openAppDB(dir)
	if err != nil {
		return err
	}
	defer cleanup()

	result, err := migrate.RollbackLastDir(db, migrationsDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(c.Out, "=> Rolled back %s\n", result.Version)
	if !result.RanDownSQL {
		_, _ = fmt.Fprintln(c.Out, "   Warning: no -- down section; only schema_migrations record was removed")
	}
	return nil
}

func (c *CLI) cmdDBPruneSessions() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	db, _, cleanup, err := openAppDB(dir)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := session.EnsureSQLiteSchema(db); err != nil {
		return err
	}
	n, err := session.NewSQLiteStore(db).PruneExpired()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(c.Out, "=> Pruned %d expired session(s)\n", n)
	return nil
}

func (c *CLI) cmdDBStatus() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	db, migrationsDir, cleanup, err := openAppDB(dir)
	if err != nil {
		return err
	}
	defer cleanup()

	entries, err := migrate.StatusDir(db, migrationsDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		state := "pending"
		if e.Applied {
			state = "applied"
		}
		_, _ = fmt.Fprintf(c.Out, "  %s  %s\n", state, e.Version)
	}
	return nil
}

func openAppDB(appDir string) (*sql.DB, string, func(), error) {
	cfg := cais.Load()
	migrationsDir := filepath.Join(appDir, "internal", "store", "migrations")
	if _, err := os.Stat(migrationsDir); err != nil {
		return nil, "", nil, fmt.Errorf("migrations dir not found: %s", migrationsDir)
	}

	if cfg.DBPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
			return nil, "", nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, "", nil, fmt.Errorf("open db: %w", err)
	}

	cleanup := func() { _ = db.Close() }
	return db, migrationsDir, cleanup, nil
}
