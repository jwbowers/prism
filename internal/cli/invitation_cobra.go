package cli

import (
	"github.com/spf13/cobra"
)

// InvitationCobraCommands handles invitation management commands
type InvitationCobraCommands struct {
	app *App
}

// NewInvitationCobraCommands creates new invitation cobra commands
func NewInvitationCobraCommands(app *App) *InvitationCobraCommands {
	return &InvitationCobraCommands{app: app}
}

// CreateInvitationCommand creates the invitation command with subcommands
func (ic *InvitationCobraCommands) CreateInvitationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invitation",
		Short: "Manage project invitations",
		Long: `Accept, decline, or view invitations to collaborate on projects.
Use 'prism project invite' to send new invitations.`,
	}

	// Add subcommands
	cmd.AddCommand(
		ic.createAddCommand(),
		ic.createListMyCommand(),
		ic.createAcceptCommand(),
		ic.createDeclineCommand(),
		ic.createInfoCommand(),
		ic.createRemoveCommand(),
	)

	return cmd
}

// createListMyCommand lists invitations received by the user
func (ic *InvitationCobraCommands) createListMyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your pending invitations",
		Long:  "List all invitations you've received to collaborate on projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			status, _ := cmd.Flags().GetString("status")

			invArgs := []string{"list"}
			if email != "" {
				invArgs = append(invArgs, "--email", email)
			}
			if status != "" {
				invArgs = append(invArgs, "--status", status)
			}

			return ic.app.Invitation(invArgs)
		},
	}

	cmd.Flags().String("email", "", "Filter by email address")
	cmd.Flags().String("status", "pending", "Filter by status (pending, accepted, declined, expired)")

	return cmd
}

// createAcceptCommand accepts an invitation
func (ic *InvitationCobraCommands) createAcceptCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "accept <token>",
		Short: "Accept a project invitation",
		Long: `Accept an invitation to collaborate on a project using the invitation token.
The token is provided in the invitation email.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ic.app.Invitation([]string{"accept", args[0]})
		},
	}
}

// createDeclineCommand declines an invitation
func (ic *InvitationCobraCommands) createDeclineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decline <token>",
		Short: "Decline a project invitation",
		Long: `Decline an invitation to collaborate on a project.
You can optionally provide a reason for declining.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reason, _ := cmd.Flags().GetString("reason")

			declineArgs := []string{"decline", args[0]}
			if reason != "" {
				declineArgs = append(declineArgs, "--reason", reason)
			}

			return ic.app.Invitation(declineArgs)
		},
	}

	cmd.Flags().String("reason", "", "Reason for declining (optional)")

	return cmd
}

// createInfoCommand shows invitation details
func (ic *InvitationCobraCommands) createInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info <token>",
		Short: "Show invitation details",
		Long:  "Display detailed information about an invitation including project details and expiration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ic.app.Invitation([]string{"info", args[0]})
		},
	}
}

// createAddCommand adds an invitation token to the local cache
func (ic *InvitationCobraCommands) createAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <token>",
		Short: "Add an invitation token to your local cache",
		Long: `Add an invitation token to your local cache for easy management.
This allows you to accept invitations by project name instead of remembering the full token.

Example:
  prism invitation add eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  prism invitation list
  prism invitation accept ml-research`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ic.app.Invitation([]string{"add", args[0]})
		},
	}
}

// createRemoveCommand removes an invitation from the local cache
func (ic *InvitationCobraCommands) createRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <project-name>",
		Short: "Remove an invitation from your local cache",
		Long: `Remove an invitation from your local cache by project name.
This only removes the cached invitation - it does not decline the invitation.

Example:
  prism invitation remove ml-research`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ic.app.Invitation([]string{"remove", args[0]})
		},
	}
}

// AddInvitationCommandsToProject adds invitation-related subcommands to the project command
func (pc *ProjectCobraCommands) AddInvitationCommands() []*cobra.Command {
	return []*cobra.Command{
		pc.createInviteCommand(),
		pc.createInvitationsCommand(),
	}
}

// createInviteCommand sends an invitation
func (pc *ProjectCobraCommands) createInviteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invite <project> <email>",
		Short: "Invite a user to collaborate on a project",
		Long: `Send an invitation to a user to collaborate on a project with a specific role.
The user will receive an email with an invitation token they can use to accept.

Invitation tokens expire after the specified duration (default: 7 days) or on a specific date.
Use --expires-in for relative duration or --expires-on for absolute date (mutually exclusive).

Examples:
  prism project invite my-project user@example.com --role admin
  prism project invite my-project user@example.com --expires-in 14d
  prism project invite my-project user@example.com --expires-on 2025-12-31`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			role, _ := cmd.Flags().GetString("role")
			message, _ := cmd.Flags().GetString("message")
			expiresIn, _ := cmd.Flags().GetString("expires-in")
			expiresOn, _ := cmd.Flags().GetString("expires-on")

			inviteArgs := []string{"invite", args[0], args[1]}
			if role != "" {
				inviteArgs = append(inviteArgs, "--role", role)
			}
			if message != "" {
				inviteArgs = append(inviteArgs, "--message", message)
			}
			if expiresIn != "" && expiresIn != "7d" {
				inviteArgs = append(inviteArgs, "--expires-in", expiresIn)
			}
			if expiresOn != "" {
				inviteArgs = append(inviteArgs, "--expires-on", expiresOn)
			}

			return pc.app.Project(inviteArgs)
		},
	}

	cmd.Flags().String("role", "member", "Role to assign (owner, admin, member, viewer)")
	cmd.Flags().String("message", "", "Optional personal message to include in invitation")
	cmd.Flags().String("expires-in", "7d", "Invitation expires in duration (e.g., 7d, 14d, 48h)")
	cmd.Flags().String("expires-on", "", "Invitation expires on date (e.g., 2025-12-31, \"2025-12-31 15:00\")")

	return cmd
}

// createInvitationsCommand manages project invitations
func (pc *ProjectCobraCommands) createInvitationsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invitations",
		Short: "Manage project invitations",
		Long:  "List, resend, or revoke project invitations",
	}

	cmd.AddCommand(
		pc.createInvitationsListCommand(),
		pc.createInvitationsResendCommand(),
		pc.createInvitationsRevokeCommand(),
		pc.createInvitationsBulkCommand(),
		createInvitationsSharedCommand(pc.app),
	)

	return cmd
}

// createInvitationsListCommand lists project invitations
func (pc *ProjectCobraCommands) createInvitationsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List invitations for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			status, _ := cmd.Flags().GetString("status")

			listArgs := []string{"invitations", "list", args[0]}
			if status != "" {
				listArgs = append(listArgs, "--status", status)
			}

			return pc.app.Project(listArgs)
		},
	}

	cmd.Flags().String("status", "", "Filter by status (pending, accepted, declined, expired)")

	return cmd
}

// createInvitationsResendCommand resends an invitation
func (pc *ProjectCobraCommands) createInvitationsResendCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "resend <invitation-id>",
		Short: "Resend a pending invitation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pc.app.Project([]string{"invitations", "resend", args[0]})
		},
	}
}

// createInvitationsRevokeCommand revokes an invitation
func (pc *ProjectCobraCommands) createInvitationsRevokeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <invitation-id>",
		Short: "Revoke a pending invitation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pc.app.Project([]string{"invitations", "revoke", args[0]})
		},
	}
}

// createInvitationsBulkCommand sends bulk invitations from file or inline emails
func (pc *ProjectCobraCommands) createInvitationsBulkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk <project>",
		Short: "Send bulk invitations from a file or inline list",
		Long: `Send invitations to multiple users at once from a CSV file, plain text file, or inline email list.

File Formats:
  CSV Format:     email,role,message
                  alice@example.com,admin,Welcome!
                  bob@example.com,member,

  Plain Text:     One email per line
                  alice@example.com
                  bob@example.com

  Inline:         Comma or space-separated emails
                  alice@example.com, bob@example.com

Examples:
  # From CSV file with custom roles and messages
  prism project invitations bulk ml-lab --file students.csv

  # From plain text file with default role
  prism project invitations bulk ml-lab --file emails.txt --role member

  # Inline email list
  prism project invitations bulk ml-lab --emails "alice@edu.com, bob@edu.com"

  # Custom default message
  prism project invitations bulk ml-lab --file emails.txt --message "Welcome to the team!"

  # Custom expiration
  prism project invitations bulk ml-lab --file emails.txt --expires-in 14d`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			emails, _ := cmd.Flags().GetString("emails")
			role, _ := cmd.Flags().GetString("role")
			message, _ := cmd.Flags().GetString("message")
			expiresIn, _ := cmd.Flags().GetString("expires-in")
			expiresOn, _ := cmd.Flags().GetString("expires-on")

			bulkArgs := []string{"invitations", "bulk", args[0]}

			if file != "" {
				bulkArgs = append(bulkArgs, "--file", file)
			}
			if emails != "" {
				bulkArgs = append(bulkArgs, "--emails", emails)
			}
			if role != "" {
				bulkArgs = append(bulkArgs, "--role", role)
			}
			if message != "" {
				bulkArgs = append(bulkArgs, "--message", message)
			}
			if expiresIn != "" && expiresIn != "7d" {
				bulkArgs = append(bulkArgs, "--expires-in", expiresIn)
			}
			if expiresOn != "" {
				bulkArgs = append(bulkArgs, "--expires-on", expiresOn)
			}

			return pc.app.Project(bulkArgs)
		},
	}

	cmd.Flags().String("file", "", "File path (CSV or plain text)")
	cmd.Flags().String("emails", "", "Comma-separated email list (alternative to --file)")
	cmd.Flags().String("role", "member", "Default role for all invitations (owner, admin, member, viewer)")
	cmd.Flags().String("message", "", "Default message for all invitations")
	cmd.Flags().String("expires-in", "7d", "Invitation expires in duration (e.g., 7d, 14d, 48h)")
	cmd.Flags().String("expires-on", "", "Invitation expires on date (e.g., 2025-12-31)")

	// Mark file and emails as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("file", "emails")

	return cmd
}
