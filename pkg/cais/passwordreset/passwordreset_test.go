package passwordreset

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildResetURL(t *testing.T) {
	got := BuildResetURL("https://app.example.com", "abc+def")
	want := "https://app.example.com/reset-password?token=abc%2Bdef"
	if got != want {
		t.Errorf("BuildResetURL = %q, want %q", got, want)
	}
}

func TestBuildResetURL_trimsTrailingSlash(t *testing.T) {
	got := BuildResetURL("https://app.example.com/", "tok")
	if got != "https://app.example.com/reset-password?token=tok" {
		t.Errorf("BuildResetURL = %q", got)
	}
}

func TestLogNotifier_writesResetLink(t *testing.T) {
	var buf bytes.Buffer
	n := LogNotifier{Out: &buf, AppURL: "http://localhost:8080"}
	if err := n.NotifyReset("user@example.com", "secret-token"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("log missing email: %q", out)
	}
	if !strings.Contains(out, "http://localhost:8080/reset-password?token=secret-token") {
		t.Errorf("log missing reset URL: %q", out)
	}
}

func TestNewToken_unique(t *testing.T) {
	a, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || len(a) < 32 {
		t.Fatalf("tokens = %q, %q", a, b)
	}
}
