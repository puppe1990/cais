package console

import "strings"

type History struct {
	lines []string
}

func NewHistory() *History {
	return &History{}
}

func (h *History) Add(line string) {
	line = strings.TrimSpace(line)
	if line == "" || line == "exit" || line == "quit" {
		return
	}
	h.lines = append(h.lines, line)
}

func (h *History) List() []string {
	out := make([]string, len(h.lines))
	copy(out, h.lines)
	return out
}

func (h *History) At(n int) string {
	if n < 1 || n > len(h.lines) {
		return ""
	}
	return h.lines[n-1]
}

func (h *History) Last() string {
	if len(h.lines) == 0 {
		return ""
	}
	return h.lines[len(h.lines)-1]
}

func (h *History) Lines() []string {
	return h.lines
}
