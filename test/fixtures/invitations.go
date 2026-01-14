package fixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/types"
)

// CreateTestInvitationOptions contains options for creating a test invitation
type CreateTestInvitationOptions struct {
	ProjectID string
	Email     string
	Role      types.ProjectRole
	InvitedBy string
	Message   string
	ExpiresAt *time.Time
}

// CreateTestInvitation creates a test invitation for integration tests
// The invitation is automatically registered for cleanup via the registry
func CreateTestInvitation(t *testing.T, registry *FixtureRegistry, opts CreateTestInvitationOptions) (*types.Invitation, error) {
	t.Helper()

	// Set defaults
	if opts.Email == "" {
		opts.Email = fmt.Sprintf("test-user-%d@university.edu", time.Now().Unix())
	}
	if opts.Role == "" {
		opts.Role = types.ProjectRoleMember
	}
	if opts.InvitedBy == "" {
		opts.InvitedBy = "test-pi@university.edu"
	}
	if opts.Message == "" {
		opts.Message = "Welcome to the research project!"
	}

	ctx := context.Background()

	// Create invitation request
	// Note: InvitedBy is determined by the API from the authenticated user
	inviteReq := client.SendInvitationRequest{
		Email:     opts.Email,
		Role:      opts.Role,
		Message:   opts.Message,
		ExpiresAt: opts.ExpiresAt,
	}

	t.Logf("Creating test invitation: %s (project: %s, role: %s)", opts.Email, opts.ProjectID, opts.Role)

	inviteResp, err := registry.client.SendInvitation(ctx, opts.ProjectID, inviteReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send invitation: %w", err)
	}

	// Register for cleanup
	registry.Register("invitation", inviteResp.Invitation.ID)
	t.Logf("Invitation created: %s (token: %s...)", inviteResp.Invitation.ID, inviteResp.Invitation.Token[:16])

	return inviteResp.Invitation, nil
}

// CreateTestBulkInvitationsOptions contains options for creating bulk test invitations
type CreateTestBulkInvitationsOptions struct {
	ProjectID      string
	Count          int               // Number of invitations to create
	EmailPrefix    string            // Optional: prefix for email addresses (default: "student")
	DefaultRole    types.ProjectRole // Optional: default role (default: member)
	DefaultMessage string            // Optional: default message
	InvitedBy      string            // Required: who is sending the invitations
	ExpiresIn      string            // Optional: expiration duration (e.g., "30d")
}

// CreateTestBulkInvitations creates bulk test invitations
// All invitations are automatically registered for cleanup via the registry
func CreateTestBulkInvitations(t *testing.T, registry *FixtureRegistry, opts CreateTestBulkInvitationsOptions) (*types.BulkInvitationResponse, error) {
	t.Helper()

	// Set defaults
	if opts.Count <= 0 {
		opts.Count = 10
	}
	if opts.EmailPrefix == "" {
		opts.EmailPrefix = "student"
	}
	if opts.DefaultRole == "" {
		opts.DefaultRole = types.ProjectRoleMember
	}
	if opts.DefaultMessage == "" {
		opts.DefaultMessage = "Welcome to the research project!"
	}
	if opts.InvitedBy == "" {
		opts.InvitedBy = "test-pi@university.edu"
	}
	if opts.ExpiresIn == "" {
		opts.ExpiresIn = "30d"
	}

	ctx := context.Background()

	// Generate invitation entries
	invitations := make([]types.BulkInvitationEntry, opts.Count)
	for i := 0; i < opts.Count; i++ {
		invitations[i] = types.BulkInvitationEntry{
			Email: fmt.Sprintf("%s%d@university.edu", opts.EmailPrefix, i+1),
		}
	}

	// Create bulk invitation request
	bulkReq := &types.BulkInvitationRequest{
		Invitations:    invitations,
		DefaultRole:    opts.DefaultRole,
		DefaultMessage: opts.DefaultMessage,
		ExpiresIn:      opts.ExpiresIn,
	}

	t.Logf("Creating bulk invitations: %d invitations for project %s", opts.Count, opts.ProjectID)

	bulkResp, err := registry.client.SendBulkInvitation(ctx, opts.ProjectID, bulkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send bulk invitations: %w", err)
	}

	// Register all successfully created invitations for cleanup
	for _, result := range bulkResp.Results {
		if result.Status == "sent" && result.InvitationID != "" {
			registry.Register("invitation", result.InvitationID)
		}
	}

	t.Logf("Bulk invitations created: %d sent, %d skipped, %d failed",
		bulkResp.Summary.Sent, bulkResp.Summary.Skipped, bulkResp.Summary.Failed)

	return bulkResp, nil
}

// AcceptTestInvitation accepts a test invitation by token
func AcceptTestInvitation(t *testing.T, client client.PrismAPI, token string) (*types.Invitation, error) {
	t.Helper()

	ctx := context.Background()

	t.Logf("Accepting invitation with token: %s...", token[:16])

	acceptResp, err := client.AcceptInvitation(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	t.Logf("Invitation accepted: %s", acceptResp.Invitation.ID)
	return acceptResp.Invitation, nil
}

// DeclineTestInvitation declines a test invitation by token with a reason
func DeclineTestInvitation(t *testing.T, client client.PrismAPI, token string, reason string) (*types.Invitation, error) {
	t.Helper()

	ctx := context.Background()

	if reason == "" {
		reason = "Not interested at this time"
	}

	t.Logf("Declining invitation with token: %s... (reason: %s)", token[:16], reason)

	declineResp, err := client.DeclineInvitation(ctx, token, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to decline invitation: %w", err)
	}

	t.Logf("Invitation declined: %s", declineResp.Invitation.ID)
	return declineResp.Invitation, nil
}

// GetTestInvitationByToken retrieves an invitation by token
func GetTestInvitationByToken(t *testing.T, client client.PrismAPI, token string) (*types.Invitation, error) {
	t.Helper()

	ctx := context.Background()

	inviteResp, err := client.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation by token: %w", err)
	}

	return inviteResp.Invitation, nil
}
