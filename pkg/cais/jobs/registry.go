package jobs

import (
	"context"
	"fmt"
)

// Registry maps job kinds to handlers.
type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(kind string, h Handler) {
	r.handlers[kind] = h
}

func (r *Registry) Perform(ctx context.Context, kind string, payload []byte) error {
	h, ok := r.handlers[kind]
	if !ok {
		return fmt.Errorf("unknown job kind %q", kind)
	}
	return h(ctx, payload)
}
