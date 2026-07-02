package passwordreset

import (
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/mail"
)

type stubSender struct {
	to, subject, body string
}

func (s *stubSender) Send(to, subject, body string) error {
	s.to = to
	s.subject = subject
	s.body = body
	return nil
}

func TestSMTPNotifier_sendsResetEmail(t *testing.T) {
	stub := &stubSender{}
	n := smtpNotifier{
		sender:  stub,
		appURL:  "https://app.example.com",
		appName: "MyApp",
	}
	if err := n.NotifyReset("user@example.com", "tok123"); err != nil {
		t.Fatal(err)
	}
	if stub.to != "user@example.com" {
		t.Errorf("to = %q", stub.to)
	}
	if !strings.Contains(stub.body, "reset-password?token=tok123") {
		t.Errorf("body = %q", stub.body)
	}
}

func TestNotifierFromConfig_logsWithoutSMTP(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	cfg := cais.Config{AppURL: "http://localhost:8080", Env: "development"}
	n := NotifierFromConfig(cfg, "MyApp")
	if _, ok := n.(LogNotifier); !ok {
		t.Fatalf("notifier type = %T, want LogNotifier", n)
	}
}

func TestNotifierFromConfig_usesSMTPWhenConfigured(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_FROM", "noreply@example.com")
	t.Setenv("SMTP_PORT", "587")

	cfg := cais.Config{AppURL: "https://app.example.com", Env: "production"}
	n := NotifierFromConfig(cfg, "MyApp")
	if _, ok := n.(smtpNotifier); !ok {
		t.Fatalf("notifier type = %T, want smtpNotifier", n)
	}
}

func TestNotifierFromConfig_usesSMTPInDevelopmentWhenConfigured(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_FROM", "noreply@example.com")

	stub := &stubSender{}
	cfg := mail.Config{Host: "smtp.example.com", Port: "587", From: "noreply@example.com"}
	n := notifierFromMail(cfg, "https://app.example.com", "MyApp", stub)
	if err := n.NotifyReset("a@b.com", "token"); err != nil {
		t.Fatal(err)
	}
	if stub.to != "a@b.com" {
		t.Errorf("to = %q", stub.to)
	}
}
