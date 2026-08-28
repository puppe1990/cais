package jobsui

import (
	"html/template"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais/jobs"
)

type dashboard struct {
	CSRFToken    string
	Kind         string
	Ready        int
	Running      int
	Finished     int
	Failed       int
	Scheduled    int
	Queues       []jobs.QueueCount
	FailedJobs   []jobs.JobRecord
	RunningJobs  []jobs.JobRecord
	ReadyJobs    []jobs.JobRecord
	FinishedJobs []jobs.JobRecord
	DelayedJobs  []jobs.ScheduledRecord
	Recurring    []jobs.RecurringTask
	Workers      []jobs.WorkerPulse
	WorkerHint   bool
	MultiWorker  bool
}

var pageTmpl = template.Must(template.New("jobs").Funcs(template.FuncMap{
	"trunc": trunc,
	"lastRun": func(t *time.Time) string {
		if t == nil {
			return "never"
		}
		return t.UTC().Format("2006-01-02 15:04")
	},
}).Parse(pageHTML))

func trunc(s string, n int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if n < 1 || len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

const pageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="refresh" content="5">
  <title>Cais Jobs</title>
  <style>
    :root { color-scheme: dark; }
    body { margin: 0; font-family: ui-sans-serif, system-ui, sans-serif; background: #0f172a; color: #e2e8f0; }
    header { padding: 1rem 1.25rem; border-bottom: 1px solid #1e293b; display: flex; justify-content: space-between; gap: 1rem; align-items: center; flex-wrap: wrap; }
    h1 { margin: 0; font-size: 0.95rem; font-weight: 600; }
    .badge { font-size: 0.7rem; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.08em; }
    main { padding: 1.25rem; display: grid; gap: 1.25rem; }
    .stats { display: flex; flex-wrap: wrap; gap: .5rem; }
    .stat { background: #1e293b; border-radius: 8px; padding: .55rem .85rem; min-width: 5.5rem; }
    .stat b { display: block; font-size: 1.15rem; }
    .stat span { font-size: .7rem; color: #94a3b8; text-transform: uppercase; letter-spacing: .06em; }
    .stat.failed b { color: #f87171; }
    .stat.running b { color: #38bdf8; }
    .hint { background: #422006; color: #fdba74; padding: .75rem 1rem; border-radius: 8px; font-size: .9rem; }
    code { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
    h2 { margin: 0 0 .6rem; font-size: .8rem; color: #94a3b8; text-transform: uppercase; letter-spacing: .08em; }
    table { width: 100%; border-collapse: collapse; font-size: .85rem; }
    th, td { text-align: left; padding: .45rem .5rem; border-bottom: 1px solid #1e293b; vertical-align: top; }
    th { color: #94a3b8; font-weight: 500; }
    .muted { color: #64748b; }
    .err { color: #fca5a5; max-width: 28rem; word-break: break-word; }
    form { display: inline; }
    button { background: #334155; color: #e2e8f0; border: 0; border-radius: 6px; padding: .25rem .55rem; cursor: pointer; font-size: .75rem; }
    button.danger { background: #7f1d1d; }
    .empty { color: #64748b; font-size: .85rem; }
    a { color: #7dd3fc; text-decoration: none; }
  </style>
</head>
<body>
  <header>
    <h1>Cais Jobs</h1>
    <span class="badge">localhost only · auto-refresh 5s</span>
  </header>
  <main>
    <div class="stats">
      <div class="stat"><b>{{.Ready}}</b><span>ready</span></div>
      <div class="stat running"><b>{{.Running}}</b><span>running</span></div>
      <div class="stat"><b>{{.Finished}}</b><span>finished</span></div>
      <div class="stat failed"><b>{{.Failed}}</b><span>failed</span></div>
      <div class="stat"><b>{{.Scheduled}}</b><span>scheduled</span></div>
    </div>
    <form method="get" action="/jobs">
      <input name="kind" value="{{.Kind}}" placeholder="filter kind" style="background:#1e293b;color:#e2e8f0;border:1px solid #334155;border-radius:6px;padding:.35rem .6rem;">
      <button type="submit">Filter</button>
      {{if .Kind}}<a href="/jobs" class="muted">clear</a>{{end}}
    </form>
    {{if .WorkerHint}}
    <div class="hint">Ready jobs are waiting. Start a worker with <code>cais jobs work</code>.</div>
    {{end}}
    {{if .MultiWorker}}
    <div class="hint">{{len .Workers}} workers are live on one SQLite file — writes serialize. Run a single <code>cais jobs work</code>.</div>
    {{end}}
    <section>
      <h2>Workers</h2>
      {{if .Workers}}
      <table>
        <tr><th>id</th><th>host</th><th>queues</th><th>concurrency</th><th>heartbeat</th></tr>
        {{range .Workers}}
        <tr><td>{{.ID}}</td><td>{{.Hostname}}</td><td>{{.Queues}}</td><td>{{.Concurrency}}</td><td class="muted">{{.HeartbeatAt}}</td></tr>
        {{end}}
      </table>
      {{else}}<p class="empty">No live worker heartbeat.</p>{{end}}
    </section>
    {{if .Queues}}
    <section>
      <h2>Queues</h2>
      <table>
        <tr><th>queue</th><th>ready</th><th>running</th><th>finished</th><th>failed</th></tr>
        {{range .Queues}}
        <tr><td>{{.Queue}}</td><td>{{.Ready}}</td><td>{{.Running}}</td><td>{{.Finished}}</td><td>{{.Failed}}</td></tr>
        {{end}}
      </table>
    </section>
    {{end}}
    <section>
      <h2>Failed</h2>
      {{if .FailedJobs}}
      <table>
        <tr><th>id</th><th>kind</th><th>queue</th><th>error</th><th></th></tr>
        {{range .FailedJobs}}
        <tr>
          <td><a href="/jobs/{{.ID}}">{{.ID}}</a></td>
          <td><a href="/jobs?kind={{.Kind}}">{{.Kind}}</a></td>
          <td>{{.Queue}}</td>
          <td class="err">{{trunc .LastError 160}}</td>
          <td>
            <form method="post" action="/jobs/{{.ID}}/retry">
              <input type="hidden" name="csrf_token" value="{{$.CSRFToken}}">
              <button type="submit">Retry</button>
            </form>
            <form method="post" action="/jobs/{{.ID}}/discard">
              <input type="hidden" name="csrf_token" value="{{$.CSRFToken}}">
              <button type="submit" class="danger">Discard</button>
            </form>
          </td>
        </tr>
        {{end}}
      </table>
      {{else}}<p class="empty">No failed jobs.</p>{{end}}
    </section>
    <section>
      <h2>Ready</h2>
      {{if .ReadyJobs}}
      <table>
        <tr><th>id</th><th>kind</th><th>queue</th><th>run at</th><th>payload</th></tr>
        {{range .ReadyJobs}}
        <tr>
          <td><a href="/jobs/{{.ID}}">{{.ID}}</a></td>
          <td><a href="/jobs?kind={{.Kind}}">{{.Kind}}</a></td>
          <td>{{.Queue}}</td>
          <td class="muted">{{.RunAt}}</td>
          <td class="muted">{{trunc .Payload 80}}</td>
        </tr>
        {{end}}
      </table>
      {{else}}<p class="empty">Queue is empty.</p>{{end}}
    </section>
    <section>
      <h2>Running</h2>
      {{if gt .Running 0}}
      <form method="post" action="/jobs/requeue-stuck">
        <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
        <button type="submit">Requeue stuck</button>
      </form>
      <p class="empty">Requeues jobs whose worker heartbeat is gone. Live workers keep their running jobs.</p>
      {{end}}
      {{if .RunningJobs}}
      <table>
        <tr><th>id</th><th>kind</th><th>queue</th><th>attempts</th></tr>
        {{range .RunningJobs}}
        <tr><td><a href="/jobs/{{.ID}}">{{.ID}}</a></td><td>{{.Kind}}</td><td>{{.Queue}}</td><td>{{.Attempts}}/{{.MaxAttempts}}</td></tr>
        {{end}}
      </table>
      {{else}}<p class="empty">Nothing running.</p>{{end}}
    </section>
    <section>
      <h2>Scheduled</h2>
      {{if .DelayedJobs}}
      <table>
        <tr><th>id</th><th>kind</th><th>queue</th><th>run at</th></tr>
        {{range .DelayedJobs}}
        <tr><td>{{.ID}}</td><td>{{.Kind}}</td><td>{{.Queue}}</td><td>{{.RunAt}}</td></tr>
        {{end}}
      </table>
      {{else}}<p class="empty">No delayed jobs.</p>{{end}}
    </section>
    <section>
      <h2>Recurring</h2>
      {{if .Recurring}}
      <table>
        <tr><th>kind</th><th>cron</th><th>queue</th><th>last run</th></tr>
        {{range .Recurring}}
        <tr><td>{{.Kind}}</td><td><code>{{.Cron}}</code></td><td>{{.Queue}}</td><td class="muted">{{lastRun .LastRun}}</td></tr>
        {{end}}
      </table>
      {{else}}<p class="empty">No recurring tasks. <code>cais g job</code> with <code>--cron</code> adds one.</p>{{end}}
    </section>
    <section>
      <h2>Finished (latest)</h2>
      {{if gt .Finished 0}}
      <form method="post" action="/jobs/prune-finished">
        <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
        <button type="submit">Clear finished</button>
      </form>
      {{end}}
      {{if .FinishedJobs}}
      <table>
        <tr><th>id</th><th>kind</th><th>finished</th></tr>
        {{range .FinishedJobs}}
        <tr><td><a href="/jobs/{{.ID}}">{{.ID}}</a></td><td>{{.Kind}}</td><td class="muted">{{.FinishedAt}}</td></tr>
        {{end}}
      </table>
      {{else}}<p class="empty">No finished jobs yet.</p>{{end}}
    </section>
  </main>
</body>
</html>
`
