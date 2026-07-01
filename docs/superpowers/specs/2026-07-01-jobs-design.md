# Jobs Design — SQLite queue (Solid Queue–style)

**Date:** 2026-07-01
**Status:** v1 implemented in `pkg/cais/jobs`
**Stack:** stdlib + SQLite (`modernc.org/sqlite`), no Redis

## Goal

Background job processing for Cais mini apps using the **same database** as the app. Inspired by [Solid Queue](https://github.com/rails/solid_queue): workers, dispatchers, and (later) schedulers — without external queue services.

## Architecture

```
HTTP handler ──Enqueue──► jobs (ready)
                SetWait──► scheduled_jobs ──dispatcher──► jobs (ready)
                                                      │
recurring_tasks ──scheduler (v2)──────────────────────┘
                                                      │
                                              worker(s) claim FOR UPDATE SKIP LOCKED
                                                      │
                                              Registry.Perform(kind, payload)
```

| Role       | Cais v1                | Process                             |
| ---------- | ---------------------- | ----------------------------------- |
| Worker     | `cais jobs work`       | Claims and runs jobs                |
| Dispatcher | goroutine in worker    | Moves due `scheduled_jobs` → `jobs` |
| Scheduler  | v2 (`recurring_tasks`) | Cron → enqueue                      |

On Lightsail, one `cais jobs work` process with `--concurrency N` is enough (N goroutines, same DB file).

## Schema

### `jobs` (ready / running / done / failed)

| Column       | Type       | Notes                                    |
| ------------ | ---------- | ---------------------------------------- |
| id           | INTEGER PK |                                          |
| queue        | TEXT       | default `default`                        |
| kind         | TEXT       | handler name, e.g. `PruneSessions`       |
| payload      | TEXT       | JSON                                     |
| priority     | INTEGER    | lower = sooner                           |
| status       | TEXT       | `ready`, `running`, `finished`, `failed` |
| attempts     | INTEGER    |                                          |
| max_attempts | INTEGER    | default 3                                |
| last_error   | TEXT       |                                          |
| run_at       | DATETIME   | not before this time                     |
| created_at   | DATETIME   |                                          |
| finished_at  | DATETIME   |                                          |

### `scheduled_jobs` (delayed)

Same fields except no `status` / `attempts` — promoted by dispatcher when `run_at <= now`.

### `recurring_tasks` (v2)

Cron expression + kind + payload; scheduler enqueues on interval.

## Claiming (concurrency)

SQLite in Cais uses atomic `UPDATE ... RETURNING` (no `SKIP LOCKED` in modernc driver):

```sql
UPDATE jobs SET status = 'running', attempts = attempts + 1
WHERE id = (
  SELECT id FROM jobs WHERE queue = ? AND status = 'ready' ...
  ORDER BY priority ASC, id ASC LIMIT 1
)
RETURNING id, kind, payload, attempts, max_attempts;
```

**One DB file = one host**; multiple goroutines/`--concurrency` on the same machine are supported.

## API (`pkg/cais/jobs`)

```go
jobs.EnsureSchema(db)

jobs.Enqueue(ctx, store, jobs.Options{Kind: "SendWelcome", Payload: data})
jobs.SetWait(ctx, store, 1*time.Hour, opts)  // scheduled_jobs

registry := jobs.NewRegistry()
registry.Register("PruneSessions", jobs.PruneSessionsHandler(db))

worker := jobs.NewWorker(jobs.WorkerConfig{
    Store: store, Registry: registry,
    Queues: []string{"default"}, Concurrency: 2,
})
worker.Run(ctx)  // dispatcher + worker loops
```

## CLI

| Command                                                    | Description             |
| ---------------------------------------------------------- | ----------------------- |
| `cais jobs work [--queues default,mail] [--concurrency 2]` | Run worker + dispatcher |
| `cais jobs status`                                         | Count jobs by status    |

Built-in job: **PruneSessions** (replaces manual-only `cais db prune-sessions` for scheduled use).

## Retry

On handler error: if `attempts < max_attempts`, set `status=ready`, `run_at=now+backoff`. Else `status=failed`, store `last_error`.

## v2 (shipped)

- `recurring_tasks` table + `RunScheduler` (cron: `*`, `N`, `*/N`)
- `cais g job <name> [--cron "0 3 * * *"]` — handler, registry, `cmd/worker/main.go`
- `cais jobs work` delegates to `go run ./cmd/worker` when present

## Out of scope

- Admin UI for failed jobs
- Redis / Postgres adapters
- Per-kind concurrency limits
- Distributed workers across multiple DB replicas

## Verification

```bash
go test ./pkg/cais/jobs/... -race
cais jobs status
cais jobs work --concurrency 1
```
