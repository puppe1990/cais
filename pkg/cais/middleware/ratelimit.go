package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string][]time.Time
}

func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  time.Minute,
		buckets: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r) + ":" + r.URL.Path
		if !rl.allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	times := rl.buckets[key]
	filtered := times[:0]
	for _, ts := range times {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	if len(filtered) >= rl.limit {
		rl.setBucket(key, filtered)
		return false
	}
	rl.setBucket(key, append(filtered, now))
	return true
}

func (rl *RateLimiter) setBucket(key string, times []time.Time) {
	if len(times) == 0 {
		delete(rl.buckets, key)
		return
	}
	rl.buckets[key] = times
}
