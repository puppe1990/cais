package jobs

import (
	"context"
	"sync"
	"testing"
)

// #174: Register wrote the handlers map while workers read it concurrently.
func TestRegistry_concurrentRegisterPerform(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Register("Kind", func(ctx context.Context, payload []byte) error { return nil })
			_ = r.Perform(context.Background(), "Unknown", nil)
		}()
	}
	wg.Wait()
}
