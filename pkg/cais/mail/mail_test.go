package mail

import (
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestConfigFrom_readsEnv(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_USER", "user")
	t.Setenv("SMTP_PASSWORD", "secret")
	t.Setenv("SMTP_FROM", "noreply@example.com")

	cfg := ConfigFrom(cais.Config{})
	if cfg.Host != "smtp.example.com" {
		t.Errorf("Host = %q", cfg.Host)
	}
	if cfg.Port != "2525" {
		t.Errorf("Port = %q", cfg.Port)
	}
	if cfg.From != "noreply@example.com" {
		t.Errorf("From = %q", cfg.From)
	}
	if !cfg.Enabled() {
		t.Error("Enabled() = false, want true")
	}
}

func TestConfig_Enabled_falseWhenHostEmpty(t *testing.T) {
	if (Config{}).Enabled() {
		t.Error("Enabled() = true for empty config")
	}
}

// #171: to/subject reach SMTP headers verbatim; user input with CRLF must be
// rejected before it can inject Bcc:/extra headers.
func TestSMTPSender_rejectsHeaderInjection(t *testing.T) {
	s := SMTPSender{Config: Config{Host: "smtp.example.com", Port: "587", From: "noreply@example.com"}}

	err := s.Send("victim@example.com\r\nBcc: attacker@example.com", "Hi", "body")
	if err == nil || !strings.Contains(err.Error(), "invalid recipient") {
		t.Errorf("recipient CRLF injection not rejected by validation, got: %v", err)
	}
	err = s.Send("victim@example.com", "Hi\r\nBcc: attacker@example.com", "body")
	if err == nil || !strings.Contains(err.Error(), "subject contains newline") {
		t.Errorf("subject CRLF injection not rejected by validation, got: %v", err)
	}
	if err := s.Send("not-an-email", "Hi", "body"); err == nil || !strings.Contains(err.Error(), "invalid recipient") {
		t.Errorf("invalid recipient not rejected by validation, got: %v", err)
	}
}

func TestValidateHeaders_acceptsCleanInput(t *testing.T) {
	if err := validateHeaders("noreply@example.com", "user@example.com", "Hello"); err != nil {
		t.Errorf("clean input rejected: %v", err)
	}
}

func TestBuildResetMessage_containsLink(t *testing.T) {
	subject, body := BuildResetMessage("MyApp", "https://app.example.com", "abc123")
	if subject == "" {
		t.Fatal("empty subject")
	}
	if !strings.Contains(body, "https://app.example.com/reset-password?token=abc123") {
		t.Errorf("body missing reset link: %q", body)
	}
	if !strings.Contains(body, "MyApp") {
		t.Errorf("body missing app name: %q", body)
	}
}

func TestLogSender_writesMessage(t *testing.T) {
	var buf strings.Builder
	s := LogSender{Out: &buf}
	if err := s.Send("user@example.com", "Subject", "Body line"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"user@example.com", "Subject", "Body line"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q: %s", want, got)
		}
	}
}
