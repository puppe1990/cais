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
}
