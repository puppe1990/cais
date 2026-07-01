package devlog

import (
	"strings"
	"testing"
)

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
