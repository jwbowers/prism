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

// TestParseDuration tests the parseDuration helper function
func TestParseDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Duration
		expectError bool
	}{
		{
			name:     "valid days",
			input:    "7d",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "valid hours",
			input:    "24h",
			expected: 24 * time.Hour,
		},
		{
			name:     "valid minutes",
			input:    "30m",
			expected: 30 * time.Minute,
		},
		{
			name:     "single digit days",
			input:    "1d",
			expected: 24 * time.Hour,
		},
		{
			name:        "too short",
			input:       "d",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid unit",
			input:       "7w",
			expectError: true,
		},
		{
			name:        "invalid number",
			input:       "abcd",
			expectError: true,
		},
		{
			name:        "no unit",
			input:       "7",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestTokenExpiration tests token expiration edge cases
func TestTokenExpiration(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create invitation that expires in 1 millisecond
	expiresAt := time.Now().Add(1 * time.Millisecond)
	req := &CreateInvitationRequest{
		ProjectID: "project-1",
		Email:     "user@example.com",
		Role:      "viewer",
		InvitedBy: "owner@example.com",
		ExpiresAt: &expiresAt,
	}

	invitation, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Try to accept expired invitation
	_, err = manager.AcceptInvitation(ctx, invitation.Token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestConcurrentInvitationAcceptance tests concurrent acceptance attempts
func TestConcurrentInvitationAcceptance(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create invitation
	req := createTestRequest("project-1", "user@example.com")
	invitation, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Launch multiple concurrent acceptance attempts
	const numAttempts = 10
	results := make(chan error, numAttempts)

	for i := 0; i < numAttempts; i++ {
		go func() {
			_, err := manager.AcceptInvitation(ctx, invitation.Token)
			results <- err
		}()
	}

	// Collect results
	var successCount, errorCount int
	for i := 0; i < numAttempts; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	// Only one should succeed, rest should fail
	assert.Equal(t, 1, successCount, "Only one acceptance should succeed")
	assert.Equal(t, numAttempts-1, errorCount, "All other attempts should fail")
}

// TestBulkInvitationsPartialFailure tests bulk invitations with some failures
func TestBulkInvitationsPartialFailure(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create a scenario where some invitations will fail
	// First, create a pending invitation for one email
	req1 := createTestRequest("project-1", "existing@example.com")
	_, err := manager.CreateInvitation(ctx, req1)
	require.NoError(t, err)

	// Now try bulk invitation including the existing email
	bulkReq := &types.BulkInvitationRequest{
		DefaultRole: "viewer",
		Invitations: []types.BulkInvitationEntry{
			{Email: "new1@example.com"},
			{Email: "existing@example.com"}, // Should fail - already pending
			{Email: "new2@example.com"},
			{Email: ""}, // Should fail - invalid email
		},
	}

	response, err := manager.CreateBulkInvitations(ctx, "project-1", "owner@example.com", bulkReq)
	require.NoError(t, err)

	// Verify partial success
	assert.Equal(t, 4, response.Summary.Total)
	assert.Equal(t, 2, response.Summary.Sent)    // new1 and new2
	assert.Equal(t, 1, response.Summary.Failed)  // empty email
	assert.Equal(t, 1, response.Summary.Skipped) // existing duplicate
	assert.Len(t, response.Results, 4)

	// Verify invitations in system (1 original + 2 new successful ones = 3 total)
	invs, err := manager.ListInvitations(ctx, &types.InvitationFilter{})
	require.NoError(t, err)
	assert.Len(t, invs, 3)
}

// TestResendInvitationTracking tests resend counter and timestamp tracking
func TestResendInvitationTracking(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create invitation
	req := createTestRequest("project-1", "user@example.com")
	invitation, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Initial state
	assert.Equal(t, 0, invitation.ResendCount)
	assert.Nil(t, invitation.LastResent)

	// First resend should succeed and increment counter
	err = manager.ResendInvitation(ctx, invitation.ID)
	require.NoError(t, err)

	// Check updated state
	inv, err := manager.GetInvitation(ctx, invitation.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, inv.ResendCount)
	assert.NotNil(t, inv.LastResent)

	// Second resend should also succeed and increment counter
	err = manager.ResendInvitation(ctx, invitation.ID)
	require.NoError(t, err)

	inv, err = manager.GetInvitation(ctx, invitation.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, inv.ResendCount)

	// Accept invitation - resend should no longer work
	_, err = manager.AcceptInvitation(ctx, invitation.Token)
	require.NoError(t, err)

	// Resend should fail for accepted invitation
	err = manager.ResendInvitation(ctx, invitation.ID)
	assert.Error(t, err)
}

// TestAcceptInvitationProjectMemberSync tests project member sync on acceptance
func TestAcceptInvitationProjectMemberSync(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create invitation
	req := createTestRequest("project-1", "user@example.com")
	req.Role = "contributor"
	invitation, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Accept invitation
	accepted, err := manager.AcceptInvitation(ctx, invitation.Token)
	require.NoError(t, err)
	require.NotNil(t, accepted)

	// Verify invitation was updated
	inv, err := manager.GetInvitation(ctx, invitation.ID)
	require.NoError(t, err)
	assert.Equal(t, types.InvitationAccepted, inv.Status)
	assert.NotNil(t, inv.AcceptedAt)
}

// TestCleanupExpiredInvitations tests cleanup of expired invitations
func TestCleanupExpiredInvitations(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create mix of active and expired invitations
	activeExpiry := time.Now().Add(7 * 24 * time.Hour)
	expiredExpiry := time.Now().Add(-1 * time.Hour)

	// Active invitation
	req1 := createTestRequest("project-1", "active@example.com")
	req1.ExpiresAt = &activeExpiry
	active, err := manager.CreateInvitation(ctx, req1)
	require.NoError(t, err)

	// Expired invitation
	req2 := createTestRequest("project-1", "expired@example.com")
	req2.ExpiresAt = &expiredExpiry
	expired, err := manager.CreateInvitation(ctx, req2)
	require.NoError(t, err)

	// Run cleanup
	count, err := manager.CleanupExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Should clean up 1 expired invitation")

	// Verify active invitation still exists and is active
	activeInv, err := manager.GetInvitation(ctx, active.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.InvitationPending, activeInv.Status)

	// Verify expired invitation status was changed to expired
	expiredInv, err := manager.GetInvitation(ctx, expired.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.InvitationExpired, expiredInv.Status)
}

// TestGetInvitationByTokenEdgeCases tests edge cases for token lookup
func TestGetInvitationByTokenEdgeCases(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		token       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty token",
			token:       "",
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name:        "invalid token",
			token:       "invalid-token-12345",
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name:        "very long token",
			token:       string(make([]byte, 1000)),
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.GetInvitationByToken(ctx, tt.token)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDeclineInvitationEdgeCases tests edge cases for declining invitations
func TestDeclineInvitationEdgeCases(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create invitation
	req := createTestRequest("project-1", "user@example.com")
	invitation, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Decline invitation
	declined, err := manager.DeclineInvitation(ctx, invitation.Token, "Not interested")
	require.NoError(t, err)
	require.NotNil(t, declined)

	// Verify invitation was declined
	inv, err := manager.GetInvitation(ctx, invitation.ID)
	require.NoError(t, err)
	assert.Equal(t, types.InvitationDeclined, inv.Status)

	// Try to decline again - should fail
	_, err = manager.DeclineInvitation(ctx, invitation.Token, "Already declined")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already")

	// Try to accept declined invitation - should fail
	_, err = manager.AcceptInvitation(ctx, invitation.Token)
	assert.Error(t, err)
}

// TestRevokeInvitationEdgeCases tests edge cases for revoking invitations
func TestRevokeInvitationEdgeCases(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create and accept invitation
	req := createTestRequest("project-1", "user@example.com")
	invitation, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	_, err = manager.AcceptInvitation(ctx, invitation.Token)
	require.NoError(t, err)

	// Try to revoke accepted invitation - should fail
	err = manager.RevokeInvitation(ctx, invitation.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in pending status")
}

// TestValidationEdgeCases tests invitation validation edge cases
func TestValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		req         *CreateInvitationRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing project ID",
			req: &CreateInvitationRequest{
				Email:     "user@example.com",
				Role:      "viewer",
				InvitedBy: "owner@example.com",
			},
			expectError: true,
			errorMsg:    "invalid project ID",
		},
		{
			name: "missing email",
			req: &CreateInvitationRequest{
				ProjectID: "project-1",
				Role:      "viewer",
				InvitedBy: "owner@example.com",
			},
			expectError: true,
			errorMsg:    "invalid email address",
		},
		{
			name: "valid email with any format",
			req: &CreateInvitationRequest{
				ProjectID: "project-1",
				Email:     "not-an-email",
				Role:      "viewer",
				InvitedBy: "owner@example.com",
			},
			expectError: false,
		},
		{
			name: "missing role",
			req: &CreateInvitationRequest{
				ProjectID: "project-1",
				Email:     "user@example.com",
				InvitedBy: "owner@example.com",
			},
			expectError: true,
			errorMsg:    "invalid role",
		},
		{
			name: "missing invited by",
			req: &CreateInvitationRequest{
				ProjectID: "project-1",
				Email:     "user@example.com",
				Role:      "viewer",
			},
			expectError: true,
			errorMsg:    "invalid inviter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRebuildIndexes tests the rebuild indexes functionality
func TestRebuildIndexes(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create multiple invitations for same project
	for i := 0; i < 3; i++ {
		req := createTestRequest("project-1", "user"+string(rune('0'+i))+"@example.com")
		_, err := manager.CreateInvitation(ctx, req)
		require.NoError(t, err)
	}

	// Clear indexes to simulate corruption
	manager.tokenIndex = make(map[string]*types.Invitation)
	manager.projectIndex = make(map[string][]*types.Invitation)
	manager.emailIndex = make(map[string][]*types.Invitation)

	// Rebuild indexes
	manager.rebuildIndexes()

	// Verify indexes were rebuilt
	assert.Len(t, manager.tokenIndex, 3)
	assert.Len(t, manager.projectIndex, 1)
	assert.Len(t, manager.projectIndex["project-1"], 3)
	assert.Len(t, manager.emailIndex, 3)
}

// TestConcurrentInvitationOperations tests concurrent read/write operations
func TestConcurrentInvitationOperations(t *testing.T) {
	manager, _, _ := createTestManager(t)
	ctx := context.Background()

	// Create initial invitation
	req := createTestRequest("project-1", "user@example.com")
	invitation, err := manager.CreateInvitation(ctx, req)
	require.NoError(t, err)

	// Launch concurrent operations
	const numOps = 20
	done := make(chan bool, numOps)

	// Mix of reads and writes
	for i := 0; i < numOps; i++ {
		if i%4 == 0 {
			// Create new invitation
			go func(idx int) {
				req := createTestRequest("project-1", "user"+string(rune('0'+idx))+"@example.com")
				_, _ = manager.CreateInvitation(ctx, req)
				done <- true
			}(i)
		} else if i%4 == 1 {
			// List invitations
			go func() {
				_, _ = manager.ListInvitations(ctx, &types.InvitationFilter{})
				done <- true
			}()
		} else if i%4 == 2 {
			// Get invitation
			go func() {
				_, _ = manager.GetInvitation(ctx, invitation.ID)
				done <- true
			}()
		} else {
			// Get by token
			go func() {
				_, _ = manager.GetInvitationByToken(ctx, invitation.Token)
				done <- true
			}()
		}
	}

	// Wait for all operations to complete
	for i := 0; i < numOps; i++ {
		<-done
	}

	// Verify system is still functional
	invs, err := manager.ListInvitations(ctx, &types.InvitationFilter{})
	require.NoError(t, err)
	assert.Greater(t, len(invs), 0)
}
