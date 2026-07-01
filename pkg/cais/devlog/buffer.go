package devlog

import (
	"strings"
	"sync"
)

type Buffer struct {
	mu    sync.RWMutex
	lines []string
	max   int
}

func NewBuffer(max int) *Buffer {
	if max < 1 {
		max = 500
	}
	return &Buffer{max: max}
}

func (b *Buffer) Write(p []byte) (int, error) {
	text := string(p)
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if line == "" && len(p) == 0 {
			continue
		}
		b.lines = append(b.lines, line)
		if len(b.lines) > b.max {
			b.lines = b.lines[len(b.lines)-b.max:]
		}
	}
	return len(p), nil
}

func (b *Buffer) Lines() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}

func (b *Buffer) Text() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.lines) == 0 {
		return ""
	}
	return strings.Join(b.lines, "\n") + "\n"
}
