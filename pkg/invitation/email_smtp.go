package invitation

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/scttfrdmn/prism/pkg/types"
)

// SMTPEmailSender sends invitation emails via SMTP.
// Configure with environment variables:
//
//	PRISM_SMTP_HOST  — SMTP server hostname (required to activate)
//	PRISM_SMTP_PORT  — SMTP server port (default: 587)
//	PRISM_SMTP_USER  — SMTP username (optional, for AUTH PLAIN)
//	PRISM_SMTP_PASS  — SMTP password (optional, for AUTH PLAIN)
//	PRISM_EMAIL_FROM — Sender address (required)
type SMTPEmailSender struct {
	host string
	port string
	user string
	pass string
	from string
}

var invitationEmailTmpl = template.Must(template.New("invite").Parse(`From: {{.From}}
To: {{.To}}
Subject: You've been invited to join {{.ProjectName}} on Prism
MIME-version: 1.0
Content-Type: text/plain; charset="UTF-8"

Hi,

{{.InviterName}} has invited you to join the project "{{.ProjectName}}" on Prism as a {{.Role}}.

{{if .Message}}
Personal message from {{.InviterName}}:
{{.Message}}

{{end}}
To accept your invitation, use the following token:
  Token: {{.Token}}

  prism project invitations accept {{.Token}}

Or visit: http://localhost:8947/api/v1/invitations/{{.Token}}/accept

This invitation expires in 7 days.

If you did not expect this invitation, you can safely ignore this email.

— The Prism Team
`))

var acceptanceTmpl = template.Must(template.New("accept").Parse(`From: {{.From}}
To: {{.To}}
Subject: Welcome to {{.ProjectName}} on Prism
MIME-version: 1.0
Content-Type: text/plain; charset="UTF-8"

Hi,

Your invitation to join "{{.ProjectName}}" on Prism has been accepted successfully.

You now have access to the project. You can launch workspaces using:
  prism workspace launch <template> --project {{.ProjectName}}

— The Prism Team
`))

// SendInvitation sends an invitation email.
func (s *SMTPEmailSender) SendInvitation(ctx context.Context, inv *types.Invitation, project *types.Project, inviter string) error {
	projectName := inv.ProjectID
	if project != nil {
		projectName = project.Name
	}

	data := map[string]string{
		"From":        s.from,
		"To":          inv.Email,
		"ProjectName": projectName,
		"InviterName": inviter,
		"Role":        string(inv.Role),
		"Token":       inv.Token,
		"Message":     inv.Message,
	}

	var buf bytes.Buffer
	if err := invitationEmailTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render invitation email: %w", err)
	}

	return s.send(inv.Email, buf.Bytes())
}

// SendAcceptanceConfirmation sends a welcome confirmation email.
func (s *SMTPEmailSender) SendAcceptanceConfirmation(ctx context.Context, inv *types.Invitation, project *types.Project) error {
	projectName := inv.ProjectID
	if project != nil {
		projectName = project.Name
	}

	data := map[string]string{
		"From":        s.from,
		"To":          inv.Email,
		"ProjectName": projectName,
	}

	var buf bytes.Buffer
	if err := acceptanceTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render acceptance email: %w", err)
	}

	return s.send(inv.Email, buf.Bytes())
}

func (s *SMTPEmailSender) send(to string, body []byte) error {
	addr := s.host + ":" + s.port

	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, body); err != nil {
		return fmt.Errorf("smtp send to %s: %w", to, err)
	}
	return nil
}
