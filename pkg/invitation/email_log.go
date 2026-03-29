package invitation

import (
	"context"
	"fmt"

	"github.com/scttfrdmn/prism/pkg/types"
)

// LogEmailSender is the default EmailSender implementation.
// It logs invitation details to stdout instead of sending real emails.
// Used when no SMTP configuration is provided.
type LogEmailSender struct{}

// SendInvitation logs an invitation dispatch.
func (l *LogEmailSender) SendInvitation(ctx context.Context, inv *types.Invitation, project *types.Project, inviter string) error {
	projectName := inv.ProjectID
	if project != nil {
		projectName = project.Name
	}
	fmt.Printf("[email] invitation to %s for project %q (token: %s, role: %s, invited_by: %s)\n",
		inv.Email, projectName, inv.Token, inv.Role, inviter)
	return nil
}

// SendAcceptanceConfirmation logs an acceptance confirmation.
func (l *LogEmailSender) SendAcceptanceConfirmation(ctx context.Context, inv *types.Invitation, project *types.Project) error {
	projectName := inv.ProjectID
	if project != nil {
		projectName = project.Name
	}
	fmt.Printf("[email] acceptance confirmation to %s for project %q\n", inv.Email, projectName)
	return nil
}
