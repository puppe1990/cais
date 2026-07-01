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
