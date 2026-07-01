package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestRateLimit_blocksAfterBurst(t *testing.T) {
	lim := NewRateLimiter(2, cais.Config{}) // 2 requests per window
	h := lim.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "203.0.113.1:1234"
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", rr.Code)
	}
}

func TestRateLimiter_cleansStaleBucketsWhenOverCapacity(t *testing.T) {
	lim := NewRateLimiter(10, cais.Config{})
	stale := time.Now().Add(-2 * time.Minute)
	for i := 0; i < 1001; i++ {
		lim.buckets[fmt.Sprintf("stale-%d", i)] = []time.Time{stale}
	}
	if len(lim.buckets) != 1001 {
		t.Fatalf("buckets = %d, want 1001", len(lim.buckets))
	}

	if !lim.allow("fresh-key") {
		t.Fatal("allow() should succeed for fresh key")
	}
	if len(lim.buckets) > 1000 {
		t.Errorf("buckets = %d, want <= 1000 after cleanup", len(lim.buckets))
	}
	for i := 0; i < 1001; i++ {
		key := fmt.Sprintf("stale-%d", i)
		if _, ok := lim.buckets[key]; ok {
			t.Errorf("stale bucket %q should be removed", key)
		}
	}
}

func TestRateLimiter_retainsActiveBucketsDuringCleanup(t *testing.T) {
	lim := NewRateLimiter(10, cais.Config{})
	recent := time.Now().Add(-10 * time.Second)
	for i := 0; i < 1001; i++ {
		lim.buckets[fmt.Sprintf("active-%d", i)] = []time.Time{recent}
	}
	if len(lim.buckets) != 1001 {
		t.Fatalf("buckets = %d, want 1001", len(lim.buckets))
	}

	if !lim.allow("another-key") {
		t.Fatal("allow() should succeed")
	}
	if len(lim.buckets) != 1002 {
		t.Errorf("buckets = %d, want 1002 (active buckets retained + new key)", len(lim.buckets))
	}
}
