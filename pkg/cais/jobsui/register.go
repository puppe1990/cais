package jobsui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/csrf"
	"github.com/puppe1990/cais/pkg/cais/jobs"
)

type handler struct {
	store *jobs.Store
}

// Register mounts GET /jobs and retry/discard POSTs. Loopback only — job
// payloads can hold PII; operators use SSH tunnel in production (same as /logs).
func Register(r *cais.Router, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("jobsui: db is required")
	}
	if err := jobs.EnsureSchema(db); err != nil {
		return err
	}
	h := &handler{store: jobs.NewStore(db)}
	r.Get("/jobs", localOnly(h.serveDashboard))
	r.Get("/jobs/{id}", localOnly(cais.IntParam("id", h.serveJob)))
	r.Post("/jobs/requeue-stuck", localOnly(h.requeueStuck))
	r.Post("/jobs/prune-finished", localOnly(h.pruneFinished))
	r.Post("/jobs/{id}/retry", localOnly(cais.IntParam("id", h.retry)))
	r.Post("/jobs/{id}/discard", localOnly(cais.IntParam("id", h.discard)))
	return nil
}

func localOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isLoopback(r) {
			http.Error(w, "jobs dashboard only available on localhost", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func isLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (h *handler) serveDashboard(w http.ResponseWriter, r *http.Request) {
	view, err := h.loadDashboard(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("jobs dashboard: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTmpl.Execute(w, view); err != nil {
		http.Error(w, fmt.Sprintf("jobs dashboard render: %v", err), http.StatusInternalServerError)
	}
}

func (h *handler) retry(w http.ResponseWriter, r *http.Request, id int64) {
	redirectOrJobErr(w, r, h.store.RetryFailed(r.Context(), id))
}

func (h *handler) discard(w http.ResponseWriter, r *http.Request, id int64) {
	redirectOrJobErr(w, r, h.store.Discard(r.Context(), id))
}

func (h *handler) requeueStuck(w http.ResponseWriter, r *http.Request) {
	if _, err := h.store.RequeueOrphaned(r.Context(), jobs.DefaultWorkerStale); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/jobs", http.StatusSeeOther)
}

func (h *handler) pruneFinished(w http.ResponseWriter, r *http.Request) {
	if _, err := h.store.PruneFinished(r.Context(), 0); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/jobs", http.StatusSeeOther)
}

func (h *handler) serveJob(w http.ResponseWriter, r *http.Request, id int64) {
	rec, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, jobs.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	view := jobDetail{Job: *rec, CSRFToken: csrf.TokenFromRequest(r)}
	if err := detailTmpl.Execute(w, view); err != nil {
		http.Error(w, fmt.Sprintf("jobs detail render: %v", err), http.StatusInternalServerError)
	}
}

func redirectOrJobErr(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		http.Redirect(w, r, "/jobs", http.StatusSeeOther)
		return
	}
	if errors.Is(err, jobs.ErrNotFound) || errors.Is(err, jobs.ErrNotFailed) {
		http.NotFound(w, r)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (h *handler) loadDashboard(r *http.Request) (dashboard, error) {
	view, err := h.loadCounts(r.Context())
	if err != nil {
		return dashboard{}, err
	}
	view.Kind = strings.TrimSpace(r.URL.Query().Get("kind"))
	if err := h.loadTables(r.Context(), &view); err != nil {
		return dashboard{}, err
	}
	view.Workers, err = h.store.ListLiveWorkers(r.Context(), jobs.DefaultWorkerStale)
	if err != nil {
		return dashboard{}, err
	}
	view.CSRFToken = csrf.TokenFromRequest(r)
	view.WorkerHint = view.Ready > 0 && len(view.Workers) == 0
	view.MultiWorker = len(view.Workers) > 1
	return view, nil
}

func (h *handler) loadCounts(ctx context.Context) (dashboard, error) {
	counts, err := h.store.CountByStatus(ctx)
	if err != nil {
		return dashboard{}, err
	}
	view := dashboard{
		Ready:    counts[jobs.StatusReady],
		Running:  counts[jobs.StatusRunning],
		Finished: counts[jobs.StatusFinished],
		Failed:   counts[jobs.StatusFailed],
	}
	view.Scheduled, err = h.store.CountScheduled(ctx)
	if err != nil {
		return dashboard{}, err
	}
	view.Queues, err = h.store.CountByQueue(ctx)
	return view, err
}

func (h *handler) loadTables(ctx context.Context, view *dashboard) error {
	var err error
	kind := view.Kind
	if view.FailedJobs, err = h.store.List(ctx, jobs.ListFilter{Status: jobs.StatusFailed, Kind: kind, Limit: 50}); err != nil {
		return err
	}
	if view.RunningJobs, err = h.store.List(ctx, jobs.ListFilter{Status: jobs.StatusRunning, Kind: kind, Limit: 50}); err != nil {
		return err
	}
	if view.ReadyJobs, err = h.store.List(ctx, jobs.ListFilter{Status: jobs.StatusReady, Kind: kind, Limit: 50}); err != nil {
		return err
	}
	if view.FinishedJobs, err = h.store.List(ctx, jobs.ListFilter{Status: jobs.StatusFinished, Kind: kind, Limit: 20}); err != nil {
		return err
	}
	if view.DelayedJobs, err = h.store.ListScheduled(ctx, 50); err != nil {
		return err
	}
	view.Recurring, err = h.store.ListRecurring(ctx)
	return err
}
