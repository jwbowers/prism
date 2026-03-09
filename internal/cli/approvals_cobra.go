// Package cli provides approval workflow CLI commands for Prism v0.12.0.
//
// These commands allow researchers to request GPU/expensive resources and PIs to
// manage the approval queue:
//
//	prism request gpu-workstation my-project
//	prism approvals list my-project
//	prism approvals dashboard
//	prism approve <request-id>
//	prism deny <request-id> --note "use CPU instead"
package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/scttfrdmn/prism/pkg/project"
	"github.com/spf13/cobra"
)

// ApprovalCobraCommands provides the approval workflow CLI commands
type ApprovalCobraCommands struct {
	app *App
}

// NewApprovalCobraCommands creates new approval CLI commands
func NewApprovalCobraCommands(app *App) *ApprovalCobraCommands {
	return &ApprovalCobraCommands{app: app}
}

// CreateRequestCommand creates `prism request` for submitting approval requests
func (ac *ApprovalCobraCommands) CreateRequestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request <type> <project>",
		Short: "Request approval for a resource or action",
		Long: `Submit an approval request to the project owner or admin.

Types:
  gpu-workstation     Request access to a GPU instance
  expensive-instance  Request access to an instance costing >$2/hr
  budget-overage      Request approval to exceed the budget limit
  emergency           Request an emergency budget increase
  sub-budget          Request a sub-budget delegation

Examples:
  prism request gpu-workstation my-project
  prism request budget-overage my-project --reason "critical deadline"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeName := args[0]
			projectName := args[1]
			reason, _ := cmd.Flags().GetString("reason")

			var approvalType project.ApprovalType
			switch typeName {
			case "gpu-workstation", "gpu":
				approvalType = project.ApprovalTypeGPUInstance
			case "expensive-instance", "expensive":
				approvalType = project.ApprovalTypeExpensiveInstance
			case "budget-overage", "overage":
				approvalType = project.ApprovalTypeBudgetOverage
			case "emergency":
				approvalType = project.ApprovalTypeEmergency
			case "sub-budget", "sub":
				approvalType = project.ApprovalTypeSubBudget
			default:
				return fmt.Errorf("unknown request type %q (use gpu-workstation, expensive-instance, budget-overage, emergency, sub-budget)", typeName)
			}

			proj, err := ac.app.apiClient.GetProject(ac.app.ctx, projectName)
			if err != nil {
				return fmt.Errorf("project %q not found: %w", projectName, err)
			}

			req, err := ac.app.apiClient.SubmitApproval(ac.app.ctx, proj.ID, approvalType,
				map[string]interface{}{"type": typeName}, reason)
			if err != nil {
				return fmt.Errorf("failed to submit request: %w", err)
			}

			fmt.Printf("✅ Request submitted (ID: %s)\n", req.ID)
			fmt.Printf("   Type:    %s\n", req.Type)
			fmt.Printf("   Project: %s\n", projectName)
			fmt.Printf("   Status:  %s (pending approval)\n", req.Status)
			fmt.Printf("   Expires: %s\n", req.ExpiresAt.Format("2006-01-02"))
			fmt.Printf("\nYour PI can approve with: prism approve %s\n", req.ID)
			return nil
		},
	}
	cmd.Flags().String("reason", "", "Justification for the request")
	return cmd
}

// CreateApproveCommand creates `prism approve <request-id>`
func (ac *ApprovalCobraCommands) CreateApproveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <request-id>",
		Short: "Approve a pending request",
		Long: `Approve a pending approval request.

Examples:
  prism approve abc-123
  prism approve abc-123 --note "approved for paper deadline"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]
			note, _ := cmd.Flags().GetString("note")

			// We need to get the project ID from the request, but the API
			// currently requires a project ID for the approval endpoint.
			// Use the admin dashboard approach: list all pending approvals.
			reqs, err := ac.app.apiClient.ListAllApprovals(ac.app.ctx, project.ApprovalStatusPending)
			if err != nil {
				return fmt.Errorf("failed to look up request: %w", err)
			}

			var targetProjectID string
			for _, r := range reqs {
				if r.ID == requestID {
					targetProjectID = r.ProjectID
					break
				}
			}

			if targetProjectID == "" {
				return fmt.Errorf("approval request %q not found", requestID)
			}

			approved, err := ac.app.apiClient.ApproveRequest(ac.app.ctx, targetProjectID, requestID, note)
			if err != nil {
				return fmt.Errorf("failed to approve request: %w", err)
			}

			fmt.Printf("✅ Request %s approved\n", approved.ID)
			if note != "" {
				fmt.Printf("   Note: %s\n", note)
			}
			return nil
		},
	}
	cmd.Flags().String("note", "", "Optional note to the requester")
	return cmd
}

// CreateDenyCommand creates `prism deny <request-id>`
func (ac *ApprovalCobraCommands) CreateDenyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deny <request-id>",
		Short: "Deny a pending request",
		Long: `Deny a pending approval request.

Examples:
  prism deny abc-123 --note "use CPU instance instead"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]
			note, _ := cmd.Flags().GetString("note")

			reqs, err := ac.app.apiClient.ListAllApprovals(ac.app.ctx, project.ApprovalStatusPending)
			if err != nil {
				return fmt.Errorf("failed to look up request: %w", err)
			}

			var targetProjectID string
			for _, r := range reqs {
				if r.ID == requestID {
					targetProjectID = r.ProjectID
					break
				}
			}

			if targetProjectID == "" {
				return fmt.Errorf("approval request %q not found", requestID)
			}

			denied, err := ac.app.apiClient.DenyRequest(ac.app.ctx, targetProjectID, requestID, note)
			if err != nil {
				return fmt.Errorf("failed to deny request: %w", err)
			}

			fmt.Printf("✅ Request %s denied\n", denied.ID)
			if note != "" {
				fmt.Printf("   Note: %s\n", note)
			}
			return nil
		},
	}
	cmd.Flags().String("note", "", "Reason for denial (shown to requester)")
	return cmd
}

// CreateApprovalsCommand creates `prism approvals` with list/dashboard subcommands
func (ac *ApprovalCobraCommands) CreateApprovalsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approvals",
		Short: "Manage approval requests",
		Long: `View and manage governance approval requests.

Examples:
  prism approvals list my-project
  prism approvals dashboard`,
	}

	// List subcommand
	listCmd := &cobra.Command{
		Use:   "list [project]",
		Short: "List approval requests for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, _ := cmd.Flags().GetString("status")

			var reqs []*project.ApprovalRequest
			var err error

			if len(args) > 0 {
				proj, err := ac.app.apiClient.GetProject(ac.app.ctx, args[0])
				if err != nil {
					return fmt.Errorf("project %q not found: %w", args[0], err)
				}
				reqs, err = ac.app.apiClient.ListApprovals(ac.app.ctx, proj.ID, project.ApprovalStatus(status))
				if err != nil {
					return fmt.Errorf("failed to list approvals: %w", err)
				}
			} else {
				reqs, err = ac.app.apiClient.ListAllApprovals(ac.app.ctx, project.ApprovalStatus(status))
				if err != nil {
					return fmt.Errorf("failed to list approvals: %w", err)
				}
			}

			if len(reqs) == 0 {
				fmt.Println("No approval requests found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tTYPE\tSTATUS\tREQUESTED BY\tCREATED\n")
			for _, r := range reqs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					r.ID[:8]+"…",
					r.Type,
					r.Status,
					r.RequestedBy,
					r.CreatedAt.Format("2006-01-02"),
				)
			}
			_ = w.Flush()
			return nil
		},
	}
	listCmd.Flags().String("status", "", "Filter by status (pending, approved, denied, expired)")
	cmd.AddCommand(listCmd)

	// Dashboard subcommand (#153)
	dashCmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Show all pending approvals across owned projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			reqs, err := ac.app.apiClient.ListAllApprovals(ac.app.ctx, project.ApprovalStatusPending)
			if err != nil {
				return fmt.Errorf("failed to load dashboard: %w", err)
			}

			if len(reqs) == 0 {
				fmt.Println("✅ No pending approval requests.")
				return nil
			}

			fmt.Printf("⏳ Pending Approvals (%d)\n\n", len(reqs))
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tTYPE\tPROJECT\tREQUESTED BY\tREASON\n")
			for _, r := range reqs {
				reason := r.Reason
				if len(reason) > 40 {
					reason = reason[:37] + "…"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					r.ID[:8]+"…",
					r.Type,
					r.ProjectID,
					r.RequestedBy,
					reason,
				)
			}
			_ = w.Flush()

			fmt.Printf("\nTo approve: prism approve <request-id>\n")
			fmt.Printf("To deny:    prism deny <request-id> --note \"reason\"\n")
			return nil
		},
	}
	cmd.AddCommand(dashCmd)

	return cmd
}
