package sqllog

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais/logentry"
	"github.com/puppe1990/cais/pkg/cais/logtime"
)

type Config struct {
	Enabled bool
	Writer  io.Writer
	JSON    bool // structured JSON lines for /logs and agent tooling
}

type DB struct {
	inner  *sql.DB
	config Config
}

type Tx struct {
	inner  *sql.Tx
	config Config
}

func Wrap(db *sql.DB, cfg Config) *DB {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}
	return &DB{inner: db, config: cfg}
}

func (d *DB) Raw() *sql.DB {
	return d.inner
}

func (d *DB) Close() error {
	return d.inner.Close()
}

func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := d.inner.Exec(query, args...)
	d.log(query, args, start, time.Since(start), err)
	return result, err
}

func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.inner.Query(query, args...)
	d.log(query, args, start, time.Since(start), err)
	return rows, err
}

func (d *DB) QueryRow(query string, args ...any) *sql.Row {
	start := time.Now()
	row := d.inner.QueryRow(query, args...)
	d.log(query, args, start, time.Since(start), nil)
	return row
}

func (d *DB) Begin() (*Tx, error) {
	tx, err := d.inner.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{inner: tx, config: d.config}, nil
}

func (t *Tx) Exec(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := t.inner.Exec(query, args...)
	t.log(query, args, start, time.Since(start), err)
	return result, err
}

func (t *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.inner.Query(query, args...)
	t.log(query, args, start, time.Since(start), err)
	return rows, err
}

func (t *Tx) QueryRow(query string, args ...any) *sql.Row {
	start := time.Now()
	row := t.inner.QueryRow(query, args...)
	t.log(query, args, start, time.Since(start), nil)
	return row
}

func (t *Tx) Commit() error {
	return t.inner.Commit()
}

func (t *Tx) Rollback() error {
	return t.inner.Rollback()
}

func (d *DB) log(query string, args []any, at time.Time, elapsed time.Duration, err error) {
	if !d.config.Enabled {
		return
	}
	writeLog(d.config, query, args, at, elapsed, err)
}

func (t *Tx) log(query string, args []any, at time.Time, elapsed time.Duration, err error) {
	if !t.config.Enabled {
		return
	}
	writeLog(t.config, query, args, at, elapsed, err)
}

func writeLog(cfg Config, query string, args []any, at time.Time, elapsed time.Duration, err error) {
	query = strings.Join(strings.Fields(query), " ")
	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}
	if cfg.JSON {
		entry := logentry.Entry{
			Kind:       "sql",
			At:         at.UTC(),
			Operation:  operationLabel(query),
			Query:      query,
			Args:       args,
			DurationMS: float64(elapsed.Microseconds()) / 1000,
		}
		if err != nil {
			entry.Error = err.Error()
		}
		_ = logentry.Write(w, entry)
		return
	}
	label := operationLabel(query)
	line := fmt.Sprintf("  %s (%s)  %s  %s  at %s", label, formatDuration(elapsed), query, formatArgs(args), logtime.Format(at))
	if err != nil {
		line += fmt.Sprintf("  ERROR: %v", err)
	}
	_, _ = fmt.Fprintln(w, line)
}
