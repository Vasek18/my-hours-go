package mailer

import (
	"context"
	"log"
)

// Mailer sends transactional messages. The interface keeps handlers decoupled
// from delivery, so swapping in a real SMTP implementation later is trivial.
type Mailer interface {
	SendPasswordReset(ctx context.Context, toEmail, resetURL string) error
	SendEmailChangeConfirmation(ctx context.Context, toEmail, confirmURL string) error
}

// ConsoleMailer "sends" mail by logging it. Used in development so no real
// emails go out and reset links are easy to grab from the logs.
type ConsoleMailer struct{}

func (ConsoleMailer) SendPasswordReset(_ context.Context, toEmail, resetURL string) error {
	log.Printf("[mailer] password reset for %s\n  reset link: %s", toEmail, resetURL)
	return nil
}

func (ConsoleMailer) SendEmailChangeConfirmation(_ context.Context, toEmail, confirmURL string) error {
	log.Printf("[mailer] email change confirmation for %s\n  confirm link: %s", toEmail, confirmURL)
	return nil
}
