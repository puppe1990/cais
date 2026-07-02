package devlog

import (
	"strings"
	"testing"
)

func TestFormatForDisplay_requestJSON(t *testing.T) {
	in := `{"kind":"request","phase":"started","at":"2026-07-02T15:04:05Z","method":"GET","path":"/login","remote":"127.0.0.1"}
{"kind":"request","phase":"completed","at":"2026-07-02T15:04:05Z","method":"GET","path":"/login","status":200,"remote":"127.0.0.1","duration_ms":3.5}`

	got := FormatForDisplay(in)
	for _, want := range []string{"GET /login", "127.0.0.1", "200", "3.5ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFormatForDisplay_sqlJSON(t *testing.T) {
	in := `{"kind":"sql","at":"2026-07-02T15:04:05Z","operation":"User Create","query":"INSERT INTO users (email) VALUES (?)","args":["a@b.com"],"duration_ms":1.2}`

	got := FormatForDisplay(in)
	for _, want := range []string{"User Create", "INSERT INTO users", "1.2ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFormatForDisplay_plainTextPassthrough(t *testing.T) {
	in := "Started GET \"/login\" for 127.0.0.1"
	if got := FormatForDisplay(in); got != in+"\n" {
		t.Fatalf("got %q", got)
	}
}
