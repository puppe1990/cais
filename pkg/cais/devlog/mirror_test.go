package devlog

import (
	"bytes"
	"strings"
	"testing"
)

func TestMirror_WritesToBoth(t *testing.T) {
	buf := NewBuffer(10)
	var out bytes.Buffer
	w := Mirror(&out, buf)

	_, _ = w.Write([]byte("hello\n"))

	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("dst = %q", out.String())
	}
	if !strings.Contains(buf.Text(), "hello") {
		t.Fatalf("buffer = %q", buf.Text())
	}
}
