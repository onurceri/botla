package services

import (
	"context"
	"errors"

	"github.com/onurceri/botla-app/pkg/logger"
)

// ErrEmailNotConfigured is returned by an EmailSender that has no delivery
// transport. Callers treat it as "not sent" rather than as a failure.
var ErrEmailNotConfigured = errors.New("email delivery is not configured")

// EmailMessage is a plain-text notification email.
type EmailMessage struct {
	To      []string
	Subject string
	Body    string
}

// EmailSender delivers notification emails (e.g. handoff alerts to operators).
//
// Botla does not ship a mail provider integration yet: the only implementation
// is NoopEmailSender. A real transport (SMTP, SES, ...) should implement this
// interface, be built from pkg/config in cmd/server, and be injected via
// HandoffService.EmailSender.
type EmailSender interface {
	Send(ctx context.Context, msg EmailMessage) error
}

// NoopEmailSender logs the message metadata and reports ErrEmailNotConfigured.
// It never logs the body, which can contain visitor conversation content.
type NoopEmailSender struct {
	Log *logger.Logger
}

var _ EmailSender = NoopEmailSender{}

// Send implements EmailSender.
func (s NoopEmailSender) Send(_ context.Context, msg EmailMessage) error {
	if s.Log != nil {
		s.Log.Info("email_not_sent_no_provider", map[string]any{
			"to":          msg.To,
			"subject":     msg.Subject,
			"body_length": len(msg.Body),
		})
	}
	return ErrEmailNotConfigured
}
