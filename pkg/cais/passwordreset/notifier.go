package passwordreset

import (
	"log"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/mail"
)

type smtpNotifier struct {
	sender  mail.Sender
	appURL  string
	appName string
}

func (n smtpNotifier) NotifyReset(email, token string) error {
	subject, body := mail.BuildResetMessage(n.appName, n.appURL, token)
	return n.sender.Send(email, subject, body)
}

func notifierFromMail(cfg mail.Config, appURL, appName string, sender mail.Sender) Notifier {
	if sender == nil {
		sender = mail.NewSender(cfg)
	}
	return smtpNotifier{sender: sender, appURL: appURL, appName: appName}
}

// NotifierFromConfig returns SMTP email in production when SMTP_HOST is set,
// otherwise LogNotifier (development-friendly).
func NotifierFromConfig(cfg cais.Config, appName string) Notifier {
	mcfg := mail.ConfigFrom(cfg)
	if cfg.Env == "production" && mcfg.Enabled() {
		return notifierFromMail(mcfg, cfg.AppURL, appName, nil)
	}
	if mcfg.Enabled() {
		return notifierFromMail(mcfg, cfg.AppURL, appName, nil)
	}
	return LogNotifier{Out: log.Writer(), AppURL: cfg.AppURL}
}

// DefaultNotifier returns the best notifier for cfg (LogNotifier fallback).
func DefaultNotifier(cfg cais.Config, appName string) Notifier {
	return NotifierFromConfig(cfg, appName)
}
