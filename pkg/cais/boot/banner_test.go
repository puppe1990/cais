package boot

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintDevBanner_ShowsCaisVersion(t *testing.T) {
	var buf bytes.Buffer
	PrintDevBanner(&buf, "0.4.2")

	out := buf.String()
	for _, want := range []string{"Cais", "v0.4.2", "hot reload"} {
		if !strings.Contains(out, want) {
			t.Fatalf("banner missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "air") || strings.Contains(out, "AIR") {
		t.Fatalf("banner should not mention air:\n%s", out)
	}
	lines := strings.Split(strings.TrimSuffix(devBannerArt, "\n"), "\n")
	if len(lines) < 5 {
		t.Fatal("banner art should have 5 lines")
	}
	// A has a pointed top, not a round cap like O
	if !strings.Contains(lines[0], `/\`) {
		t.Fatalf("A should have pointed top, got:\n%s", lines[0])
	}
	// I is a thin column — no full-width top bar (that reads as T)
	if strings.Contains(lines[0], "_____") && strings.Count(lines[0], "_____") > 1 {
		t.Fatalf("I should not look like T on first line:\n%s", lines[0])
	}
}
