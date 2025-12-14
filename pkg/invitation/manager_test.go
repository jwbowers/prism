package invitation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmailSender is a test implementation of EmailSender
type MockEmailSender struct {
	sentInvitations   []*types.Invitation
	sentConfirmations []*types.Invitation
	sendError         error
	confirmError      error
}

func (m *MockEmailSender) SendInvitation(ctx context.Context, invitation *types.Invitation, project *types.Project, inviter string) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.sentInvitations = append(m.sentInvitations, invitation)
	return nil
}

func (m *MockEmailSender) SendAcceptanceConfirmation(ctx context.Context, invitation *types.Invitation, project *types.Project) error {
	if m.confirmError != nil {
		return m.confirmError
	}
	m.sentConfirmations = append(m.sentConfirmations, invitation)
	return nil
}

// createTestManager creates a manager with a temporary invitations file
func createTestManager(t *testing.T) (*Manager, *MockEmailSender, string) {
	tmpDir := t.TempDir()
	invitationsPath := filepath.Join(tmpDir, "invitations.json")

	mockSender := &MockEmailSender{}

	manager := &Manager{
		invitationsPath: invitationsPath,
		invitations:     make(map[string]*types.Invitation),
		tokenIndex:      make(map[string]*types.Invitation),
		projectIndex:    make(map[string][]*types.Invitation),
		emailIndex:      make(map[string][]*types.Invitation),
		emailSender:     mockSender,
	}

	return manager, mockSender, tmpDir
}

// createTestRequest creates a standard test invitation request
func createTestRequest(projectID, email string) *CreateInvitationRequest {
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	return &CreateInvitationRequest{
		ProjectID: projectID,
		Email:     email,
		Role:      "viewer",
		InvitedBy: "owner@example.com",
		ExpiresAt: &expiresAt,
	}
}

func TestNewManager(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	mockSender := &MockEmailSender{}
	manager, err := NewManager(mockSender)

	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.invitations)
	assert.NotNil(t, manager.tokenIndex)
	assert.NotNil(t, manager.projectIndex)
	assert.NotNil(t, manager.emailIndex)
	assert.Equal(t, mockSender, manager.emailSender)
}

func TestCreateInvitation(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	req := &CreateInvitationRequest{
		ProjectID: "project-1",
		Email:     "user@example.com",
		Role:      "viewer",
		InvitedBy: "owner@example.com",
		ExpiresAt: &expiresAt,
		Message:   "Join our project",
	}

	invitation, err := manager.CreateInvitation(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, invitation)
	assert.NotEmpty(t, invitation.ID)
	assert.NotEmpty(t, invitation.Token)
	assert.Equal(t, req.ProjectID, invitation.ProjectID)
	assert.Equal(t, req.Email, invitation.Email)
	assert.Equal(t, req.Role, invitation.Role)
	assert.Equal(t, types.InvitationPending, invitation.Status)
	assert.Equal(t, req.InvitedBy, invitation.InvitedBy)
	assert.Equal(t, req.Message, invitation.Message)
	assert.False(t, invitation.InvitedAt.IsZero())
	assert.False(t, invitation.ExpiresAt.IsZero())

	// Verify invitation was stored
	stored, err := manager.GetInvitation(ctx, invitation.ID)
	require.NoError(t, err)
	assert.Equal(t, invitation.ID, stored.ID)
}

func TestCreateInvitation_DuplicatePending(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	req := createTestRequest("project-1", "user@example.com")

	// Create first invitation
	_, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Try to create duplicate
	_, err = manager.CreateInvitation(ctx, req)
	assert.ErrorIs(t, err, types.ErrDuplicateInvitation)
}

func TestGetInvitationByToken(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	req := createTestRequest("project-1", "user@example.com")

	created, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Get by token
	retrieved, err := manager.GetInvitationByToken(ctx, created.Token)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Token, retrieved.Token)

	// Test invalid token
	_, err = manager.GetInvitationByToken(ctx, "invalid-token")
	assert.ErrorIs(t, err, types.ErrInvitationNotFound)
}

func TestListInvitations(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create multiple invitations
	projects := []string{"project-1", "project-2"}
	emails := []string{"user1@example.com", "user2@example.com"}

	for _, projectID := range projects {
		for _, email := range emails {
			req := createTestRequest(projectID, email)
			_, err := manager.CreateInvitation(ctx, req)
			require.NoError(t, err)
		}
	}

	// List all
	all, err := manager.ListInvitations(ctx, &types.InvitationFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 4)

	// Filter by project
	project1Invites, err := manager.ListInvitations(ctx, &types.InvitationFilter{
		ProjectID: "project-1",
	})
	require.NoError(t, err)
	assert.Len(t, project1Invites, 2)

	// Filter by email
	user1Invites, err := manager.ListInvitations(ctx, &types.InvitationFilter{
		Email: "user1@example.com",
	})
	require.NoError(t, err)
	assert.Len(t, user1Invites, 2)

	// Filter by status
	pendingInvites, err := manager.ListInvitations(ctx, &types.InvitationFilter{
		Status: types.InvitationPending,
	})
	require.NoError(t, err)
	assert.Len(t, pendingInvites, 4)
}

func TestAcceptInvitation(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	req := &CreateInvitationRequest{
		ProjectID: "project-1",
		Email:     "user@example.com",
		Role:      "viewer",
		InvitedBy: "owner@example.com",
	}

	created, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Accept invitation
	accepted, err := manager.AcceptInvitation(ctx, created.Token)
	require.NoError(t, err)
	assert.Equal(t, types.InvitationAccepted, accepted.Status)
	assert.False(t, accepted.AcceptedAt.IsZero())

	// Try to accept again - should fail
	_, err = manager.AcceptInvitation(ctx, created.Token)
	assert.Error(t, err)
}

func TestDeclineInvitation(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	req := &CreateInvitationRequest{
		ProjectID: "project-1",
		Email:     "user@example.com",
		Role:      "viewer",
		InvitedBy: "owner@example.com",
	}

	created, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Decline invitation
	reason := "Not interested"
	declined, err := manager.DeclineInvitation(ctx, created.Token, reason)
	require.NoError(t, err)
	assert.Equal(t, types.InvitationDeclined, declined.Status)
	assert.Equal(t, reason, declined.DeclineReason)
	assert.False(t, declined.DeclinedAt.IsZero())
}

func TestRevokeInvitation(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	req := &CreateInvitationRequest{
		ProjectID: "project-1",
		Email:     "user@example.com",
		Role:      "viewer",
		InvitedBy: "owner@example.com",
	}

	created, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Revoke invitation
	err = manager.RevokeInvitation(ctx, created.ID)
	require.NoError(t, err)

	// Verify status changed
	revoked, err := manager.GetInvitation(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, types.InvitationRevoked, revoked.Status)
}

func TestGetProjectSummary(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	projectID := "project-1"

	// Create invitations with different statuses
	statuses := []struct {
		email  string
		status types.InvitationStatus
	}{
		{"user1@example.com", types.InvitationPending},
		{"user2@example.com", types.InvitationPending},
		{"user3@example.com", types.InvitationAccepted},
		{"user4@example.com", types.InvitationDeclined},
	}

	for _, s := range statuses {
		req := &CreateInvitationRequest{
			ProjectID: projectID,
			Email:     s.email,
			Role:      "viewer",
			InvitedBy: "owner@example.com",
		}
		inv, err := manager.CreateInvitation(ctx, req)
		require.NoError(t, err)

		// Update status if not pending
		if s.status != types.InvitationPending {
			inv.Status = s.status
			if s.status == types.InvitationAccepted {
				now := time.Now()
				inv.AcceptedAt = &now
			} else if s.status == types.InvitationDeclined {
				now := time.Now()
				inv.DeclinedAt = &now
			}
			manager.invitations[inv.ID] = inv
		}
	}

	// Get summary
	summary, err := manager.GetProjectSummary(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, projectID, summary.ProjectID)
	assert.Equal(t, 4, summary.TotalInvitations)
	assert.Equal(t, 2, summary.PendingCount)
	assert.Equal(t, 1, summary.AcceptedCount)
	assert.Equal(t, 1, summary.DeclinedCount)
}

func TestCleanupExpired(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create expired invitation
	expiredTime := time.Now().Add(-24 * time.Hour) // Already expired
	expiredReq := &CreateInvitationRequest{
		ProjectID: "project-1",
		Email:     "expired@example.com",
		Role:      "viewer",
		InvitedBy: "owner@example.com",
		ExpiresAt: &expiredTime,
	}
	expiredInv, err := manager.CreateInvitation(ctx, expiredReq)
	require.NoError(t, err)

	// Create valid invitation
	validReq := createTestRequest("project-1", "valid@example.com")
	_, err = manager.CreateInvitation(ctx, validReq)
	require.NoError(t, err)

	// Cleanup expired
	cleaned, err := manager.CleanupExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, cleaned)

	// Verify expired invitation status changed
	expired, err := manager.GetInvitation(ctx, expiredInv.ID)
	require.NoError(t, err)
	assert.Equal(t, types.InvitationExpired, expired.Status)

	// Verify valid invitation still pending
	validInvites, err := manager.ListInvitations(ctx, &types.InvitationFilter{
		Status: types.InvitationPending,
	})
	require.NoError(t, err)
	assert.Len(t, validInvites, 1)
}

func TestSaveAndLoad(t *testing.T) {
	manager, _, tmpDir := createTestManager(t)
	ctx := context.Background()

	// Create invitation
	req := &CreateInvitationRequest{
		ProjectID: "project-1",
		Email:     "user@example.com",
		Role:      "viewer",
		InvitedBy: "owner@example.com",
	}
	created, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Save
	err = manager.save()
	require.NoError(t, err)

	// Create new manager to load
	invitationsPath := filepath.Join(tmpDir, "invitations.json")
	mockSender := &MockEmailSender{}

	newManager := &Manager{
		invitationsPath: invitationsPath,
		invitations:     make(map[string]*types.Invitation),
		tokenIndex:      make(map[string]*types.Invitation),
		projectIndex:    make(map[string][]*types.Invitation),
		emailIndex:      make(map[string][]*types.Invitation),
		emailSender:     mockSender,
	}

	// Load
	err = newManager.load()
	require.NoError(t, err)

	// Verify invitation was loaded
	loaded, err := newManager.GetInvitation(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, loaded.ID)
	assert.Equal(t, created.Token, loaded.Token)
	assert.Equal(t, created.Email, loaded.Email)
}

func TestGenerateSecureToken(t *testing.T) {
	token1, err := generateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := generateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)
}

func TestResendInvitation(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	req := createTestRequest("project-1", "user@example.com")

	created, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Resend invitation
	err = manager.ResendInvitation(ctx, created.ID)
	require.NoError(t, err)

	// Verify resend count incremented
	updated, err := manager.GetInvitation(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, updated.ResendCount)
	assert.NotNil(t, updated.LastResent)

	// Try to resend again
	err = manager.ResendInvitation(ctx, created.ID)
	require.NoError(t, err)

	updated, err = manager.GetInvitation(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, updated.ResendCount)

	// Try to resend non-existent invitation
	err = manager.ResendInvitation(ctx, "invalid-id")
	assert.ErrorIs(t, err, types.ErrInvitationNotFound)

	// Try to resend accepted invitation (should fail)
	_, err = manager.AcceptInvitation(ctx, created.Token)
	require.NoError(t, err)

	err = manager.ResendInvitation(ctx, created.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, types.ErrInvitationNotPending)
}

func TestCreateBulkInvitations(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	bulkReq := &types.BulkInvitationRequest{
		DefaultRole:    "viewer",
		DefaultMessage: "Welcome to the project",
		ExpiresAt:      &expiresAt,
		Invitations: []types.BulkInvitationEntry{
			{Email: "user1@example.com"},
			{Email: "user2@example.com", Role: "editor"},
			{Email: "user3@example.com", Message: "Custom message"},
		},
	}

	response, err := manager.CreateBulkInvitations(ctx, "project-1", "owner@example.com", bulkReq)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 3, response.Summary.Total)
	assert.Equal(t, 3, response.Summary.Sent)
	assert.Equal(t, 0, response.Summary.Failed)
	assert.Equal(t, 0, response.Summary.Skipped)
	assert.Len(t, response.Results, 3)

	// Verify invitations were created
	invitations, err := manager.ListInvitations(ctx, &types.InvitationFilter{
		ProjectID: "project-1",
	})
	require.NoError(t, err)
	assert.Len(t, invitations, 3)

	// Try to create duplicate (should be skipped)
	response, err = manager.CreateBulkInvitations(ctx, "project-1", "owner@example.com", bulkReq)
	require.NoError(t, err)
	assert.Equal(t, 3, response.Summary.Total)
	assert.Equal(t, 0, response.Summary.Sent)
	assert.Equal(t, 3, response.Summary.Skipped)
}

func TestCreateBulkInvitations_WithInvalidEmails(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	bulkReq := &types.BulkInvitationRequest{
		DefaultRole: "viewer",
		Invitations: []types.BulkInvitationEntry{
			{Email: "valid@example.com"},
			{Email: ""}, // Empty email should fail validation
			{Email: "another@example.com"},
		},
	}

	response, err := manager.CreateBulkInvitations(ctx, "project-1", "owner@example.com", bulkReq)
	require.NoError(t, err)
	assert.Equal(t, 3, response.Summary.Total)
	assert.Equal(t, 2, response.Summary.Sent)
	assert.Equal(t, 1, response.Summary.Failed)
}
