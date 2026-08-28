package jobsui

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/csrf"
	"github.com/puppe1990/cais/pkg/cais/jobs"
	"github.com/puppe1990/cais/pkg/cais/middleware"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	if err := jobs.EnsureSchema(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func newJobsRouter(t *testing.T, db *sql.DB) *cais.Router {
	t.Helper()
	r := cais.NewRouter()
	r.Use(middleware.CSRF(cais.Config{Env: "development"}))
	if err := Register(r, db); err != nil {
		t.Fatal(err)
	}
	return r
}

func getLocal(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestRegister_LocalhostShowsQueueCounts(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	if _, err := jobs.Enqueue(context.Background(), store, jobs.Options{Kind: "Ping"}); err != nil {
		t.Fatal(err)
	}
	r := newJobsRouter(t, db)

	rr := getLocal(t, r, "/jobs")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{"Cais Jobs", "Ping", "ready", "failed"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q:\n%s", want, body)
		}
	}
}

func TestRegister_BlocksNonLocalhost(t *testing.T) {
	r := newJobsRouter(t, testDB(t))
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	req.RemoteAddr = "203.0.113.1:80"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}

func TestRegister_RetryFailedPost(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	ctx := context.Background()
	id, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Flaky", MaxAttempts: 1})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.Claim(ctx, jobs.DefaultQueue)
	if err != nil || job == nil {
		t.Fatal(err)
	}
	if err := store.MarkFailed(ctx, job.ID, errors.New("boom"), job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}

	r := newJobsRouter(t, db)
	get := getLocal(t, r, "/jobs")
	token := csrfCookie(t, get)

	form := url.Values{csrf.FormField: {token}}
	req := httptest.NewRequest(http.MethodPost, "/jobs/"+strconv.FormatInt(id, 10)+"/retry", strings.NewReader(form.Encode()))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}

	rec, err := store.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != jobs.StatusReady {
		t.Fatalf("status = %q, want ready", rec.Status)
	}
}

func TestRegister_DiscardFailedPost(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	ctx := context.Background()
	id, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Dead", MaxAttempts: 1})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.Claim(ctx, jobs.DefaultQueue)
	if err != nil || job == nil {
		t.Fatal(err)
	}
	if err := store.MarkFailed(ctx, job.ID, errors.New("dead"), job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}

	r := newJobsRouter(t, db)
	get := getLocal(t, r, "/jobs")
	token := csrfCookie(t, get)

	form := url.Values{csrf.FormField: {token}}
	req := httptest.NewRequest(http.MethodPost, "/jobs/"+strconv.FormatInt(id, 10)+"/discard", strings.NewReader(form.Encode()))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if _, err := store.Get(ctx, id); !errors.Is(err, jobs.ErrNotFound) {
		t.Fatalf("expected discarded job gone, err=%v", err)
	}
}

func TestRegister_WorkerHintWhenReadyAndIdle(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	if _, err := jobs.Enqueue(context.Background(), store, jobs.Options{Kind: "Waiting"}); err != nil {
		t.Fatal(err)
	}
	rr := getLocal(t, newJobsRouter(t, db), "/jobs")
	if !strings.Contains(rr.Body.String(), "cais jobs work") {
		t.Fatalf("expected worker hint, body:\n%s", rr.Body.String())
	}
}

func TestRegister_RequeueStuckPost(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	ctx := context.Background()
	id, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Orphan"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Claim(ctx, jobs.DefaultQueue); err != nil {
		t.Fatal(err)
	}

	r := newJobsRouter(t, db)
	get := getLocal(t, r, "/jobs")
	if !strings.Contains(get.Body.String(), "Requeue stuck") {
		t.Fatalf("running jobs should offer requeue, body:\n%s", get.Body.String())
	}
	if !strings.Contains(get.Body.String(), "heartbeat") {
		t.Fatalf("requeue should mention heartbeat, body:\n%s", get.Body.String())
	}
	token := csrfCookie(t, get)
	form := url.Values{csrf.FormField: {token}}
	req := httptest.NewRequest(http.MethodPost, "/jobs/requeue-stuck", strings.NewReader(form.Encode()))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	rec, err := store.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != jobs.StatusReady {
		t.Fatalf("status = %q, want ready", rec.Status)
	}
}

func TestRegister_ShowsLiveWorkerAndHidesHint(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	ctx := context.Background()
	if _, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Waiting"}); err != nil {
		t.Fatal(err)
	}
	if err := store.TouchWorker(ctx, jobs.WorkerPulse{ID: "w1", Hostname: "box"}); err != nil {
		t.Fatal(err)
	}
	rr := getLocal(t, newJobsRouter(t, db), "/jobs")
	body := rr.Body.String()
	if strings.Contains(body, "cais jobs work") {
		t.Fatal("should not hint to start a worker when one is live")
	}
	if !strings.Contains(body, "w1") || !strings.Contains(body, "box") {
		t.Fatalf("missing worker pulse:\n%s", body)
	}
}

func TestRegister_WarnsOnTwoLiveWorkers(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	ctx := context.Background()
	if err := store.TouchWorker(ctx, jobs.WorkerPulse{ID: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := store.TouchWorker(ctx, jobs.WorkerPulse{ID: "b"}); err != nil {
		t.Fatal(err)
	}
	rr := getLocal(t, newJobsRouter(t, db), "/jobs")
	if !strings.Contains(rr.Body.String(), "one SQLite file") {
		t.Fatalf("expected multi-worker warning, body:\n%s", rr.Body.String())
	}
}

func TestRegister_FiltersByKind(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	ctx := context.Background()
	if _, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Mail"}); err != nil {
		t.Fatal(err)
	}
	if _, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Ping"}); err != nil {
		t.Fatal(err)
	}
	rr := getLocal(t, newJobsRouter(t, db), "/jobs?kind=Mail")
	body := rr.Body.String()
	if !strings.Contains(body, "Mail") {
		t.Fatalf("missing Mail:\n%s", body)
	}
	if strings.Contains(body, "Ping") {
		t.Fatalf("Ping should be filtered out:\n%s", body)
	}
}

func TestRegister_JobDetailPage(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	id, err := jobs.Enqueue(context.Background(), store, jobs.Options{Kind: "Mail", Payload: map[string]any{"to": "a@b.c"}})
	if err != nil {
		t.Fatal(err)
	}
	rr := getLocal(t, newJobsRouter(t, db), "/jobs/"+strconv.FormatInt(id, 10))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Mail") || !strings.Contains(body, "a@b.c") {
		t.Fatalf("detail missing payload:\n%s", body)
	}
}

func TestRegister_PruneFinishedPost(t *testing.T) {
	db := testDB(t)
	store := jobs.NewStore(db)
	ctx := context.Background()
	id, err := jobs.Enqueue(ctx, store, jobs.Options{Kind: "Done"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFinished(ctx, id); err != nil {
		t.Fatal(err)
	}
	r := newJobsRouter(t, db)
	get := getLocal(t, r, "/jobs")
	if !strings.Contains(get.Body.String(), "Clear finished") {
		t.Fatalf("missing prune button:\n%s", get.Body.String())
	}
	token := csrfCookie(t, get)
	form := url.Values{csrf.FormField: {token}}
	req := httptest.NewRequest(http.MethodPost, "/jobs/prune-finished", strings.NewReader(form.Encode()))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rr.Code)
	}
	if _, err := store.Get(ctx, id); err != jobs.ErrNotFound {
		t.Fatalf("expected pruned, err=%v", err)
	}
}

func TestRegister_NilDB(t *testing.T) {
	if err := Register(cais.NewRouter(), nil); err == nil {
		t.Fatal("expected error for nil db")
	}
}

func csrfCookie(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	for _, c := range rr.Result().Cookies() {
		if c.Name == csrf.CookieName {
			return c.Value
		}
	}
	t.Fatal("missing csrf cookie")
	return ""
}
