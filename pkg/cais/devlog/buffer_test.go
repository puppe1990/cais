package devlog

import (
	"strings"
	"testing"
)

func TestNewBuffer_defaultsMaxWhenZero(t *testing.T) {
	buf := NewBuffer(0)
	for i := 0; i < 600; i++ {
		_, _ = buf.Write([]byte("x\n"))
	}
	if len(buf.Lines()) != 500 {
		t.Fatalf("len = %d, want default max 500", len(buf.Lines()))
	}
}

func TestBuffer_TextEmpty(t *testing.T) {
	if got := NewBuffer(10).Text(); got != "" {
		t.Fatalf("Text() = %q, want empty", got)
	}
}

func TestBuffer_KeepsLastNLines(t *testing.T) {
	buf := NewBuffer(3)
	for _, line := range []string{"one", "two", "three", "four"} {
		_, _ = buf.Write([]byte(line + "\n"))
	}

	got := strings.Join(buf.Lines(), "|")
	if got != "two|three|four" {
		t.Fatalf("lines = %q", got)
	}
}

func TestBuffer_TextJoinsLines(t *testing.T) {
	buf := NewBuffer(10)
	_, _ = buf.Write([]byte("alpha\nbeta\n"))

	if !strings.HasSuffix(buf.Text(), "beta\n") {
		t.Fatalf("text = %q", buf.Text())
	}
}
