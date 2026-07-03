package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais/session"
)

type stubStatsLoader struct {
	level  int
	points int
	rank   int
}

func (s stubStatsLoader) LoadStats(_ int64) (int, int, int, error) {
	return s.level, s.points, s.rank, nil
}

func TestLoadUserStats_setsContext(t *testing.T) {
	loader := stubStatsLoader{level: 3, points: 420, rank: 12}
	var got UserStats
	h := LoadUserStats(loader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats, ok := UserStatsFrom(r)
		if !ok {
			t.Fatal("expected stats in context")
		}
		got = stats
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = session.WithUserID(req, 7)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got.Level != 3 || got.Points != 420 || got.Rank != 12 {
		t.Fatalf("stats = %+v", got)
	}
}

func TestLoadUserStats_anonymous_skips(t *testing.T) {
	loader := stubStatsLoader{level: 1, points: 0, rank: 99}
	h := LoadUserStats(loader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserStatsFrom(r); ok {
			t.Fatal("anonymous request should not have stats")
		}
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
}
