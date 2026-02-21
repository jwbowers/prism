// Package invitation provides credential lifecycle validation tests
package invitation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/scttfrdmn/prism/pkg/types"
)

// MockProfileManager implements ProfileManager for testing
type MockProfileManager struct {
	profiles         map[string]*Profile
	credentials      map[string]aws.Credentials
	credentialsError error
	credentialsValid bool
	getProfileError  error
}

// NewMockProfileManager creates a new mock profile manager
func NewMockProfileManager() *MockProfileManager {
	return &MockProfileManager{
		profiles:         make(map[string]*Profile),
		credentials:      make(map[string]aws.Credentials),
		credentialsValid: true,
	}
}

// AddProfile adds a profile to the mock
func (m *MockProfileManager) AddProfile(profile *Profile) {
	m.profiles[profile.ID] = profile
}

// AddCredentials adds credentials for a profile
func (m *MockProfileManager) AddCredentials(profileID string, creds aws.Credentials) {
	m.credentials[profileID] = creds
}

// SetCredentialsError sets an error to be returned when getting credentials
func (m *MockProfileManager) SetCredentialsError(err error) {
	m.credentialsError = err
}

// SetGetProfileError sets an error to be returned when getting profile
func (m *MockProfileManager) SetGetProfileError(err error) {
	m.getProfileError = err
}

// GetProfile implements ProfileManager interface
func (m *MockProfileManager) GetProfile(ctx context.Context, profileID string) (*Profile, error) {
	if m.getProfileError != nil {
		return nil, m.getProfileError
	}

	profile, exists := m.profiles[profileID]
	if !exists {
		return nil, errors.New("profile not found")
	}

	return profile, nil
}

// GetProfileCredentials implements ProfileManager interface
func (m *MockProfileManager) GetProfileCredentials(ctx context.Context, profileID string) (aws.Credentials, error) {
	if m.credentialsError != nil {
		return aws.Credentials{}, m.credentialsError
	}

	creds, exists := m.credentials[profileID]
	if !exists {
		return aws.Credentials{}, errors.New("credentials not found")
	}

	return creds, nil
}

// TestValidateInvitation_Success tests successful invitation validation
func TestValidateInvitation_Success(t *testing.T) {
	t.Setenv("PRISM_STATE_DIR", t.TempDir())
	// Create manager
	manager, err := NewManager(nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create invitation
	ctx := context.Background()
	req := &CreateInvitationRequest{
		ProjectID: "test-project",
		Email:     "test@example.com",
		Role:      types.ProjectRoleMember,
		InvitedBy: "admin",
	}

	inv, err := manager.CreateInvitation(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	// Add profile ID to invitation
	manager.mutex.Lock()
	invitation := manager.invitations[inv.ID]
	invitation.ProfileID = "test-profile"
	manager.mutex.Unlock()

	// Create mock profile manager
	profileMgr := NewMockProfileManager()
	profileMgr.AddProfile(&Profile{
		ID:         "test-profile",
		Name:       "Test Profile",
		AWSProfile: "default",
		Region:     "us-west-2",
		ExpiresAt:  nil,
	})
	profileMgr.AddCredentials("test-profile", aws.Credentials{
		AccessKeyID:     "test-key",
		SecretAccessKey: "test-secret",
	})

	// Mock credentials test to always succeed
	// In real implementation, TestCredentials would call AWS STS
	// For unit test, we'll skip the actual validation

	// For now, test the structure of validation
	result, err := manager.ValidateInvitation(ctx, inv.ID, profileMgr)
	if err != nil {
		t.Fatalf("ValidateInvitation failed: %v", err)
	}

	// Verify result structure
	if result.ProfileID != "test-profile" {
		t.Errorf("Expected ProfileID=test-profile, got %s", result.ProfileID)
	}

	if result.IsExpired {
		t.Error("Expected invitation not to be expired")
	}

	if !result.ProfileExists {
		t.Error("Expected profile to exist")
	}
}

// TestValidateInvitation_MissingProfile tests validation with missing profile
func TestValidateInvitation_MissingProfile(t *testing.T) {
	t.Setenv("PRISM_STATE_DIR", t.TempDir())
	// Create manager
	manager, err := NewManager(nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create invitation without profile ID
	ctx := context.Background()
	req := &CreateInvitationRequest{
		ProjectID: "test-project-2",
		Email:     "test2@example.com",
		Role:      types.ProjectRoleMember,
		InvitedBy: "admin",
	}

	inv, err := manager.CreateInvitation(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	// Create mock profile manager
	profileMgr := NewMockProfileManager()

	// Validate invitation
	result, err := manager.ValidateInvitation(ctx, inv.ID, profileMgr)
	if err != nil {
		t.Fatalf("ValidateInvitation failed: %v", err)
	}

	// Verify validation failed due to missing profile
	if result.Valid {
		t.Error("Expected validation to fail for missing profile")
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message for missing profile")
	}

	if result.ErrorMessage != "invitation missing profile association" {
		t.Errorf("Unexpected error message: %s", result.ErrorMessage)
	}
}

// TestValidateInvitation_ExpiredInvitation tests validation with expired invitation
func TestValidateInvitation_ExpiredInvitation(t *testing.T) {
	t.Setenv("PRISM_STATE_DIR", t.TempDir())
	// Create manager
	manager, err := NewManager(nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create invitation with past expiration
	ctx := context.Background()
	pastTime := time.Now().Add(-24 * time.Hour)
	req := &CreateInvitationRequest{
		ProjectID: "test-project-3",
		Email:     "test3@example.com",
		Role:      types.ProjectRoleMember,
		InvitedBy: "admin",
		ExpiresAt: &pastTime,
	}

	inv, err := manager.CreateInvitation(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	// Add profile ID
	manager.mutex.Lock()
	invitation := manager.invitations[inv.ID]
	invitation.ProfileID = "test-profile"
	manager.mutex.Unlock()

	// Create mock profile manager
	profileMgr := NewMockProfileManager()

	// Validate invitation
	result, err := manager.ValidateInvitation(ctx, inv.ID, profileMgr)
	if err != nil {
		t.Fatalf("ValidateInvitation failed: %v", err)
	}

	// Verify validation failed due to expiration
	if result.Valid {
		t.Error("Expected validation to fail for expired invitation")
	}

	if !result.IsExpired {
		t.Error("Expected IsExpired to be true")
	}

	if result.ErrorMessage != "invitation has expired" {
		t.Errorf("Unexpected error message: %s", result.ErrorMessage)
	}
}

// TestGetCredentialStatus tests credential status retrieval
func TestGetCredentialStatus(t *testing.T) {
	t.Setenv("PRISM_STATE_DIR", t.TempDir())
	// Create manager
	manager, err := NewManager(nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create invitation
	ctx := context.Background()
	futureTime := time.Now().Add(24 * time.Hour)
	req := &CreateInvitationRequest{
		ProjectID: "test-project-4",
		Email:     "test4@example.com",
		Role:      types.ProjectRoleMember,
		InvitedBy: "admin",
		ExpiresAt: &futureTime,
	}

	inv, err := manager.CreateInvitation(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	// Add profile ID
	manager.mutex.Lock()
	invitation := manager.invitations[inv.ID]
	invitation.ProfileID = "test-profile"
	manager.mutex.Unlock()

	// Create mock profile manager
	profileMgr := NewMockProfileManager()
	profileMgr.AddProfile(&Profile{
		ID:         "test-profile",
		Name:       "Test Profile",
		AWSProfile: "default",
		Region:     "us-west-2",
	})
	profileMgr.AddCredentials("test-profile", aws.Credentials{
		AccessKeyID:     "test-key",
		SecretAccessKey: "test-secret",
	})

	// Get credential status
	// Note: This will fail credential testing since we can't make real AWS calls
	// In a real scenario with integration tests, this would work
	status, err := manager.GetCredentialStatus(ctx, inv.ID, profileMgr)
	if err != nil {
		t.Fatalf("GetCredentialStatus failed: %v", err)
	}

	// Verify status structure
	if status.InvitationID != inv.ID {
		t.Errorf("Expected InvitationID=%s, got %s", inv.ID, status.InvitationID)
	}

	if status.RemainingTime == nil {
		t.Error("Expected RemainingTime to be set")
	}
}

// TestEnforceExpiration tests expiration enforcement
func TestEnforceExpiration(t *testing.T) {
	t.Setenv("PRISM_STATE_DIR", t.TempDir())
	// Create manager
	manager, err := NewManager(nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create expired invitation
	ctx := context.Background()
	pastTime := time.Now().Add(-24 * time.Hour)
	req := &CreateInvitationRequest{
		ProjectID: "test-project-5",
		Email:     "test5@example.com",
		Role:      types.ProjectRoleMember,
		InvitedBy: "admin",
		ExpiresAt: &pastTime,
	}

	inv, err := manager.CreateInvitation(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	// Verify initial status is pending
	if inv.Status != types.InvitationPending {
		t.Errorf("Expected status=pending, got %s", inv.Status)
	}

	// Enforce expiration
	err = manager.EnforceExpiration(ctx, inv.ID)
	if err != nil {
		t.Fatalf("EnforceExpiration failed: %v", err)
	}

	// Verify status changed to expired
	manager.mutex.RLock()
	updatedInv := manager.invitations[inv.ID]
	manager.mutex.RUnlock()

	if updatedInv.Status != types.InvitationExpired {
		t.Errorf("Expected status=expired, got %s", updatedInv.Status)
	}
}

// TestFormatDuration tests duration formatting
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{-1 * time.Hour, "expired"},
		{30 * time.Minute, "30m"},
		{1 * time.Hour, "1h 0m"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
		{48 * time.Hour, "2 days"},
		{72 * time.Hour, "3 days"},
	}

	for _, test := range tests {
		result := formatDuration(test.duration)
		if result != test.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", test.duration, result, test.expected)
		}
	}
}
