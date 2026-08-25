package jobs

import (
	"context"
	"fmt"
	"sync"
)

// Registry maps job kinds to handlers. The map is guarded because workers read
// it concurrently with late Register calls (#174).
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(kind string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[kind] = h
}

func (r *Registry) Perform(ctx context.Context, kind string, payload []byte) error {
	r.mu.RLock()
	h, ok := r.handlers[kind]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown job kind %q", kind)
	}
	return h(ctx, payload)
}
