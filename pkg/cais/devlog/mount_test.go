package devlog

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrepare_development(t *testing.T) {
	t.Cleanup(func() { SetDefault(nil) })

	buf := Prepare("development")
	if buf == nil {
		t.Fatal("expected buffer in development")
	}
	if Default() != buf {
		t.Fatal("Default() should return Prepare buffer")
	}
}

func TestPrepare_productionClearsDefault(t *testing.T) {
	SetDefault(NewBuffer(10))
	t.Cleanup(func() { SetDefault(nil) })

	if got := Prepare("production"); got != nil {
		t.Fatalf("Prepare(production) = %v, want nil", got)
	}
	if Default() != nil {
		t.Fatal("Default() should be nil outside development")
	}
}

func TestMirrorDefault_writesToDefaultBuffer(t *testing.T) {
	SetDefault(NewBuffer(10))
	t.Cleanup(func() { SetDefault(nil) })

	var out bytes.Buffer
	w := MirrorDefault(&out)
	if _, err := w.Write([]byte("line\n")); err != nil {
		t.Fatal(err)
	}
	if out.String() != "line\n" {
		t.Fatalf("dst = %q", out.String())
	}
	if !strings.Contains(Default().Text(), "line") {
		t.Fatalf("buffer = %q", Default().Text())
	}
}

func TestMirrorDefault_nilDestination(t *testing.T) {
	SetDefault(NewBuffer(10))
	t.Cleanup(func() { SetDefault(nil) })

	w := MirrorDefault(nil)
	if n, err := w.Write([]byte("only-buffer\n")); err != nil || n != 12 {
		t.Fatalf("Write = (%d, %v)", n, err)
	}
	if !strings.Contains(Default().Text(), "only-buffer") {
		t.Fatal("expected buffer write")
	}
}
