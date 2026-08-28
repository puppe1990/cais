package jobsui

import (
	"html/template"

	"github.com/puppe1990/cais/pkg/cais/jobs"
)

type jobDetail struct {
	Job       jobs.JobRecord
	CSRFToken string
}

var detailTmpl = template.Must(template.New("job").Parse(detailHTML))

const detailHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Job {{.Job.ID}}</title>
  <style>
    :root { color-scheme: dark; }
    body { margin: 0; font-family: ui-sans-serif, system-ui, sans-serif; background: #0f172a; color: #e2e8f0; }
    header, main { padding: 1rem 1.25rem; }
    header { border-bottom: 1px solid #1e293b; display: flex; gap: 1rem; align-items: center; }
    a { color: #7dd3fc; }
    dl { display: grid; grid-template-columns: 8rem 1fr; gap: .4rem 1rem; }
    dt { color: #94a3b8; }
    pre { background: #1e293b; padding: 1rem; border-radius: 8px; overflow: auto; white-space: pre-wrap; }
    .err { color: #fca5a5; }
    form { display: inline; }
    button { background: #334155; color: #e2e8f0; border: 0; border-radius: 6px; padding: .3rem .65rem; cursor: pointer; }
    button.danger { background: #7f1d1d; }
  </style>
</head>
<body>
  <header>
    <a href="/jobs">← Jobs</a>
    <h1>Job {{.Job.ID}}</h1>
  </header>
  <main>
    <dl>
      <dt>kind</dt><dd>{{.Job.Kind}}</dd>
      <dt>status</dt><dd>{{.Job.Status}}</dd>
      <dt>queue</dt><dd>{{.Job.Queue}}</dd>
      <dt>attempts</dt><dd>{{.Job.Attempts}} / {{.Job.MaxAttempts}}</dd>
      <dt>worker</dt><dd>{{.Job.WorkerID}}</dd>
      <dt>run at</dt><dd>{{.Job.RunAt}}</dd>
      <dt>started</dt><dd>{{.Job.StartedAt}}</dd>
      <dt>finished</dt><dd>{{.Job.FinishedAt}}</dd>
    </dl>
    {{if .Job.LastError}}<p class="err">{{.Job.LastError}}</p>{{end}}
    <h2>Payload</h2>
    <pre>{{.Job.Payload}}</pre>
    {{if eq .Job.Status "failed"}}
    <form method="post" action="/jobs/{{.Job.ID}}/retry">
      <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
      <button type="submit">Retry</button>
    </form>
    <form method="post" action="/jobs/{{.Job.ID}}/discard">
      <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
      <button type="submit" class="danger">Discard</button>
    </form>
    {{end}}
  </main>
</body>
</html>
`
