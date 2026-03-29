// Package cli provides the prism approval command group (v0.21.0 #495).
//
// These commands expose the hierarchical launch approval workflow documented in
// issue #495. They complement (and share backend with) the older `prism approvals`
// and `prism request/approve/deny` commands introduced in v0.12.0.
//
// Canonical CLI UX from #495:
//
//	prism approval list [--project <name>] [--status pending|approved|denied]
//	prism approval show <id>
//	prism approval approve <id> [--comment "..."]
//	prism approval deny <id> [--comment "..."]
//	prism approval request <type> --project <name> [--reason "..."]
package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/scttfrdmn/prism/pkg/project"
	"github.com/spf13/cobra"
)

// ApprovalCobraCommandsV2 provides the unified prism approval command group (#495)
type ApprovalCobraCommandsV2 struct {
	app *App
}

// NewApprovalCobraCommandsV2 creates the singular `approval` command group.
func NewApprovalCobraCommandsV2(app *App) *ApprovalCobraCommandsV2 {
	return &ApprovalCobraCommandsV2{app: app}
}

// CreateApprovalCommand returns the root `prism approval` command with subcommands.
func (ac *ApprovalCobraCommandsV2) CreateApprovalCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval",
		Short: "Manage launch approval requests",
		Long: `View and manage hierarchical launch approval requests.

Approval requests are created when a workspace launch requires PI or admin
consent — either explicitly via --request-approval or automatically when
the estimated cost exceeds the project's approval threshold.

Examples:
  prism approval list
  prism approval list --project my-lab --status pending
  prism approval show req-abc123
  prism approval approve req-abc123 --comment "Approved for paper deadline"
  prism approval deny req-abc123 --comment "Use CPU instance instead"
  prism approval request gpu --project my-lab --reason "model training"`,
		GroupID: "collab",
	}

	cmd.AddCommand(ac.createListCommand())
	cmd.AddCommand(ac.createShowCommand())
	cmd.AddCommand(ac.createApproveCommand())
	cmd.AddCommand(ac.createDenyCommand())
	cmd.AddCommand(ac.createRequestCommand())

	return cmd
}

// createListCommand creates `prism approval list`
func (ac *ApprovalCobraCommandsV2) createListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List approval requests",
		Long: `List approval requests, optionally filtered by project or status.

Without arguments, lists all pending requests across projects you own or admin.

Examples:
  prism approval list
  prism approval list --project my-lab
  prism approval list --status approved`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName, _ := cmd.Flags().GetString("project")
			status, _ := cmd.Flags().GetString("status")

			var reqs []*project.ApprovalRequest
			var err error

			if projectName != "" {
				proj, err := ac.app.apiClient.GetProject(ac.app.ctx, projectName)
				if err != nil {
					return fmt.Errorf("project %q not found: %w", projectName, err)
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
			fmt.Fprintf(w, "ID\tTYPE\tSTATUS\tPROJECT\tREQUESTED BY\tCREATED\n")
			for _, r := range reqs {
				idShort := r.ID
				if len(idShort) > 12 {
					idShort = idShort[:12] + "…"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					idShort,
					r.Type,
					r.Status,
					r.ProjectID,
					r.RequestedBy,
					r.CreatedAt.Format("2006-01-02"),
				)
			}
			_ = w.Flush()
			return nil
		},
	}
	cmd.Flags().String("project", "", "Filter by project name")
	cmd.Flags().String("status", "", "Filter by status: pending, approved, denied, expired")
	return cmd
}

// createShowCommand creates `prism approval show <id>`
func (ac *ApprovalCobraCommandsV2) createShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show details of an approval request",
		Long: `Show full details of an approval request by ID.

Examples:
  prism approval show req-abc123-full-uuid`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]

			// Look up the project for this approval by listing all pending requests
			// (workaround: API requires project ID for single-approval fetch)
			allReqs, err := ac.app.apiClient.ListAllApprovals(ac.app.ctx, "")
			if err != nil {
				return fmt.Errorf("failed to look up request: %w", err)
			}

			var projectID string
			for _, r := range allReqs {
				if r.ID == requestID || (len(requestID) > 8 && len(r.ID) >= len(requestID) && r.ID[:len(requestID)] == requestID) {
					projectID = r.ProjectID
					requestID = r.ID // resolve full ID if partial was given
					break
				}
			}

			if projectID == "" {
				return fmt.Errorf("approval request %q not found", requestID)
			}

			req, err := ac.app.apiClient.GetApproval(ac.app.ctx, projectID, requestID)
			if err != nil {
				return fmt.Errorf("failed to fetch approval: %w", err)
			}

			fmt.Printf("ID:           %s\n", req.ID)
			fmt.Printf("Type:         %s\n", req.Type)
			fmt.Printf("Status:       %s\n", req.Status)
			fmt.Printf("Project:      %s\n", req.ProjectID)
			fmt.Printf("Requested by: %s\n", req.RequestedBy)
			fmt.Printf("Created:      %s\n", req.CreatedAt.Format(time.RFC3339))
			fmt.Printf("Expires:      %s\n", req.ExpiresAt.Format(time.RFC3339))
			if req.Reason != "" {
				fmt.Printf("Reason:       %s\n", req.Reason)
			}
			if req.ReviewedBy != "" {
				fmt.Printf("Reviewed by:  %s\n", req.ReviewedBy)
				if req.ReviewedAt != nil {
					fmt.Printf("Reviewed at:  %s\n", req.ReviewedAt.Format(time.RFC3339))
				}
				if req.ReviewNote != "" {
					fmt.Printf("Review note:  %s\n", req.ReviewNote)
				}
			}
			if len(req.Details) > 0 {
				fmt.Printf("Details:\n")
				for k, v := range req.Details {
					fmt.Printf("  %s: %v\n", k, v)
				}
			}

			if req.Status == project.ApprovalStatusPending {
				fmt.Printf("\nTo approve: prism approval approve %s\n", req.ID)
				fmt.Printf("To deny:    prism approval deny %s --comment \"reason\"\n", req.ID)
			} else if req.Status == project.ApprovalStatusApproved {
				fmt.Printf("\nLaunch with this approval:\n")
				fmt.Printf("  prism workspace launch <template> <name> --approval %s\n", req.ID)
			}

			return nil
		},
	}
}

// createApproveCommand creates `prism approval approve <id>`
func (ac *ApprovalCobraCommandsV2) createApproveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a pending request",
		Long: `Approve a pending launch approval request.

Only project admins and owners can approve requests.

Examples:
  prism approval approve req-abc123
  prism approval approve req-abc123 --comment "Approved for paper deadline"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]
			comment, _ := cmd.Flags().GetString("comment")

			reqs, err := ac.app.apiClient.ListAllApprovals(ac.app.ctx, project.ApprovalStatusPending)
			if err != nil {
				return fmt.Errorf("failed to look up request: %w", err)
			}

			var projectID string
			for _, r := range reqs {
				if r.ID == requestID {
					projectID = r.ProjectID
					break
				}
			}

			if projectID == "" {
				return fmt.Errorf("pending approval request %q not found", requestID)
			}

			approved, err := ac.app.apiClient.ApproveRequest(ac.app.ctx, projectID, requestID, comment)
			if err != nil {
				return fmt.Errorf("failed to approve request: %w", err)
			}

			fmt.Printf("✅ Approved: %s\n", approved.ID)
			if comment != "" {
				fmt.Printf("   Comment: %s\n", comment)
			}
			fmt.Printf("\nRequester can now launch with:\n")
			fmt.Printf("  prism workspace launch <template> <name> --approval %s\n", approved.ID)
			return nil
		},
	}
	cmd.Flags().String("comment", "", "Optional comment shown to the requester")
	return cmd
}

// createDenyCommand creates `prism approval deny <id>`
func (ac *ApprovalCobraCommandsV2) createDenyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deny <id>",
		Short: "Deny a pending request",
		Long: `Deny a pending launch approval request.

Examples:
  prism approval deny req-abc123 --comment "Use CPU instance for initial testing"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]
			comment, _ := cmd.Flags().GetString("comment")

			reqs, err := ac.app.apiClient.ListAllApprovals(ac.app.ctx, project.ApprovalStatusPending)
			if err != nil {
				return fmt.Errorf("failed to look up request: %w", err)
			}

			var projectID string
			for _, r := range reqs {
				if r.ID == requestID {
					projectID = r.ProjectID
					break
				}
			}

			if projectID == "" {
				return fmt.Errorf("pending approval request %q not found", requestID)
			}

			denied, err := ac.app.apiClient.DenyRequest(ac.app.ctx, projectID, requestID, comment)
			if err != nil {
				return fmt.Errorf("failed to deny request: %w", err)
			}

			fmt.Printf("❌ Denied: %s\n", denied.ID)
			if comment != "" {
				fmt.Printf("   Comment: %s\n", comment)
			}
			return nil
		},
	}
	cmd.Flags().String("comment", "", "Reason for denial (shown to requester)")
	return cmd
}

// createRequestCommand creates `prism approval request <type> --project <name>`
func (ac *ApprovalCobraCommandsV2) createRequestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request <type>",
		Short: "Submit an approval request",
		Long: `Submit a launch approval request to your project PI or admin.

Types:
  gpu             Request GPU instance access
  expensive       Request access to an instance costing >$2/hr
  budget-overage  Request approval to exceed the budget limit
  emergency       Request an emergency budget increase
  sub-budget      Request a sub-budget delegation

Examples:
  prism approval request gpu --project my-lab
  prism approval request expensive --project my-lab --reason "model training deadline"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeName := args[0]
			projectName, _ := cmd.Flags().GetString("project")
			reason, _ := cmd.Flags().GetString("reason")

			if projectName == "" {
				return fmt.Errorf("--project is required")
			}

			var approvalType project.ApprovalType
			switch typeName {
			case "gpu", "gpu-workstation":
				approvalType = project.ApprovalTypeGPUInstance
			case "expensive", "expensive-instance":
				approvalType = project.ApprovalTypeExpensiveInstance
			case "budget-overage", "overage":
				approvalType = project.ApprovalTypeBudgetOverage
			case "emergency":
				approvalType = project.ApprovalTypeEmergency
			case "sub-budget", "sub":
				approvalType = project.ApprovalTypeSubBudget
			default:
				return fmt.Errorf("unknown type %q — use: gpu, expensive, budget-overage, emergency, sub-budget", typeName)
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

			fmt.Printf("✅ Approval request submitted (ID: %s)\n", req.ID)
			fmt.Printf("   Type:    %s\n", req.Type)
			fmt.Printf("   Project: %s\n", projectName)
			fmt.Printf("   Status:  pending\n")
			fmt.Printf("   Expires: %s\n", req.ExpiresAt.Format("2006-01-02"))
			fmt.Printf("\nYour PI can approve with:\n")
			fmt.Printf("  prism approval approve %s\n", req.ID)
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project name (required)")
	cmd.Flags().String("reason", "", "Justification for the request")
	return cmd
}
