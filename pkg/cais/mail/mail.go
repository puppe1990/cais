package mail

import (
	"fmt"
	"io"
	"net/smtp"
	"net/url"
	"os"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
)

// Config holds SMTP connection settings from environment variables.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// ConfigFrom reads SMTP_* env vars. Port defaults to 587 when host is set.
func ConfigFrom(_ cais.Config) Config {
	cfg := Config{
		Host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:     strings.TrimSpace(os.Getenv("SMTP_PORT")),
		User:     strings.TrimSpace(os.Getenv("SMTP_USER")),
		Password: strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		From:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
	}
	if cfg.Host != "" && cfg.Port == "" {
		cfg.Port = "587"
	}
	return cfg
}

// Enabled reports whether outbound email is configured.
func (c Config) Enabled() bool {
	return c.Host != "" && c.From != ""
}

// Sender delivers plain-text email.
type Sender interface {
	Send(to, subject, body string) error
}

// LogSender writes messages to Out (stdout when nil).
type LogSender struct {
	Out io.Writer
}

func (s LogSender) Send(to, subject, body string) error {
	out := s.Out
	if out == nil {
		out = os.Stdout
	}
	_, err := fmt.Fprintf(out, "mail to=%s subject=%q\n%s\n", to, subject, body)
	return err
}

// SMTPSender sends mail via net/smtp. Uses STARTTLS on port 587 by default.
type SMTPSender struct {
	Config Config
}

func (s SMTPSender) Send(to, subject, body string) error {
	addr := s.Config.Host + ":" + s.Config.Port
	msg := buildMessage(s.Config.From, to, subject, body)
	auth := smtp.PlainAuth("", s.Config.User, s.Config.Password, s.Config.Host)
	return smtp.SendMail(addr, auth, s.Config.From, []string{to}, []byte(msg))
}

func buildMessage(from, to, subject, body string) string {
	var b strings.Builder
	b.WriteString("From: ")
	b.WriteString(from)
	b.WriteString("\r\nTo: ")
	b.WriteString(to)
	b.WriteString("\r\nSubject: ")
	b.WriteString(subject)
	b.WriteString("\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(body)
	return b.String()
}

// BuildResetMessage returns subject and body for a password reset email.
func BuildResetMessage(appName, appURL, token string) (subject, body string) {
	link := strings.TrimRight(appURL, "/") + "/reset-password?token=" + url.QueryEscape(token)
	if appName == "" {
		appName = "Cais"
	}
	subject = fmt.Sprintf("Reset your %s password", appName)
	body = fmt.Sprintf("Hello,\n\nUse the link below to reset your %s password:\n\n%s\n\nIf you did not request this, you can ignore this email.\n", appName, link)
	return subject, body
}

// NewSender returns SMTPSender when configured, otherwise LogSender.
func NewSender(cfg Config) Sender {
	if cfg.Enabled() {
		return SMTPSender{Config: cfg}
	}
	return LogSender{}
}
