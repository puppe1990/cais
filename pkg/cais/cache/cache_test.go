package cache

import (
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	c := New[string](time.Minute)

	c.Set("greeting", "hello")

	got, ok := c.Get("greeting")
	if !ok {
		t.Fatal("Get returned false for existing key")
	}
	if got != "hello" {
		t.Errorf("Get = %q, want %q", got, "hello")
	}

	_, ok = c.Get("missing")
	if ok {
		t.Error("Get returned true for missing key")
	}
}

func TestExpiryAfterTTL(t *testing.T) {
	c := New[string](50 * time.Millisecond)

	c.Set("temp", "value")

	got, ok := c.Get("temp")
	if !ok || got != "value" {
		t.Fatalf("Get before expiry = (%q, %v), want (value, true)", got, ok)
	}

	time.Sleep(60 * time.Millisecond)

	_, ok = c.Get("temp")
	if ok {
		t.Error("Get returned true for expired key")
	}
}

func TestDelete(t *testing.T) {
	c := New[string](time.Minute)

	c.Set("temp", "value")
	c.Delete("temp")

	_, ok := c.Get("temp")
	if ok {
		t.Error("Get returned true after Delete")
	}

	c.Delete("missing") // should not panic
}

func TestConcurrentAccess(t *testing.T) {
	c := New[int](time.Minute)

	const goroutines = 32
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				c.Set("counter", n*iterations+j)
			}
		}(i)

		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = c.Get("counter")
			}
		}()
	}

	wg.Wait()

	_, ok := c.Get("counter")
	if !ok {
		t.Fatal("Get returned false after concurrent access")
	}
}

func TestKey(t *testing.T) {
	k1 := Key("list", 42, "en", true)
	k2 := Key("list", 42, "en", true)
	if k1 != k2 {
		t.Errorf("Key not stable: %s != %s", k1, k2)
	}
	if k1 == "" {
		t.Error("Key returned empty")
	}

	k3 := Key("list", 99)
	if k3 == k1 {
		t.Error("different parts produced same key")
	}
}

func TestHash_stableAndShort(t *testing.T) {
	type listVer struct {
		Count   int
		LastMod string
		Page    int
	}

	v1 := listVer{Count: 80, LastMod: "2026-07-05T10:00", Page: 1}
	h1 := Hash(v1)
	h2 := Hash(v1)
	if h1 != h2 {
		t.Error("Hash not stable for same value")
	}
	if len(h1) == 0 || len(h1) > 32 {
		t.Errorf("Hash length suspicious: %d", len(h1))
	}

	v2 := listVer{Count: 81, LastMod: "2026-07-05T10:00", Page: 1}
	h3 := Hash(v2)
	if h3 == h1 {
		t.Error("different values produced same hash")
	}
}
