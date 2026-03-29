package invitation

import (
	"context"
	"fmt"
	"os"

	"github.com/scttfrdmn/prism/pkg/types"
)

// NewEmailSender returns an EmailSender based on environment configuration.
//
//   - If PRISM_SMTP_HOST is set, returns an SMTPEmailSender.
//   - Otherwise returns a LogEmailSender that prints to stdout.
func NewEmailSender() EmailSender {
	host := os.Getenv("PRISM_SMTP_HOST")
	if host == "" {
		return &LogEmailSender{}
	}

	port := os.Getenv("PRISM_SMTP_PORT")
	if port == "" {
		port = "587"
	}

	from := os.Getenv("PRISM_EMAIL_FROM")
	if from == "" {
		fmt.Println("[email] PRISM_SMTP_HOST set but PRISM_EMAIL_FROM is empty; falling back to log sender")
		return &LogEmailSender{}
	}

	return &SMTPEmailSender{
		host: host,
		port: port,
		user: os.Getenv("PRISM_SMTP_USER"),
		pass: os.Getenv("PRISM_SMTP_PASS"),
		from: from,
	}
}

// SendInvitationEmail dispatches an invitation email via the configured sender.
// Errors are logged but do not fail the caller (fire-and-forget).
func (m *Manager) SendInvitationEmail(ctx context.Context, inv *types.Invitation, project *types.Project, inviter string) {
	if m.emailSender == nil {
		return
	}
	if err := m.emailSender.SendInvitation(ctx, inv, project, inviter); err != nil {
		fmt.Printf("[invitation] email send failed for %s: %v\n", inv.Email, err)
	}
}

// SendAcceptanceEmail dispatches an acceptance confirmation via the configured sender.
// Errors are logged but do not fail the caller.
func (m *Manager) SendAcceptanceEmail(ctx context.Context, inv *types.Invitation, project *types.Project) {
	if m.emailSender == nil {
		return
	}
	if err := m.emailSender.SendAcceptanceConfirmation(ctx, inv, project); err != nil {
		fmt.Printf("[invitation] acceptance email send failed for %s: %v\n", inv.Email, err)
	}
}
