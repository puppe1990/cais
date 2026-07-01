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

func TestDevBannerArt_usesBlockCAISLogo(t *testing.T) {
	lines := strings.Split(strings.TrimRight(devBannerArt, "\n"), "\n")
	if len(lines) != 6 {
		t.Fatalf("want 6 lines, got %d", len(lines))
	}
	if !strings.Contains(devBannerArt, "██████╗") {
		t.Fatal("banner should use block CAIS logo")
	}
}

func TestWriteDevSeedWarning_development(t *testing.T) {
	var buf bytes.Buffer
	WriteDevSeedWarning(&buf, "development")

	out := buf.String()
	if !strings.Contains(out, "demo@example.com") {
		t.Fatalf("warning missing demo user:\n%s", out)
	}
	if !strings.Contains(out, "development only") {
		t.Fatalf("warning missing development qualifier:\n%s", out)
	}
}

func TestWriteDevSeedWarning_production(t *testing.T) {
	var buf bytes.Buffer
	WriteDevSeedWarning(&buf, "production")
	if buf.Len() != 0 {
		t.Fatalf("production should not print seed warning, got:\n%s", buf.String())
	}
}
