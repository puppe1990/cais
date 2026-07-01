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
		return fmt.Errorf("usage: cais db <migrate|status|rollback|prune-sessions>")
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
	default:
		return fmt.Errorf("unknown db command %q (use migrate, status, rollback, or prune-sessions)", args[0])
	}
}

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

	version, err := migrate.RollbackLastDir(db, migrationsDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(c.Out, "=> Rolled back %s (does not run SQL down migrations)\n", version)
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
