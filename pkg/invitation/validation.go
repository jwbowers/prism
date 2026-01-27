// Package invitation provides credential lifecycle validation for invitations
package invitation

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/scttfrdmn/prism/pkg/types"
)

// ProfileManager defines the interface for profile operations
type ProfileManager interface {
	GetProfile(ctx context.Context, profileID string) (*Profile, error)
	GetProfileCredentials(ctx context.Context, profileID string) (aws.Credentials, error)
}

// Profile represents a Prism profile with AWS credentials
type Profile struct {
	ID         string
	Name       string
	AWSProfile string
	Region     string
	ExpiresAt  *time.Time
}

// ValidationResult contains the result of invitation validation
type ValidationResult struct {
	Valid              bool      `json:"valid"`
	ProfileID          string    `json:"profile_id"`
	ProfileExists      bool      `json:"profile_exists"`
	CredentialsValid   bool      `json:"credentials_valid"`
	IsExpired          bool      `json:"is_expired"`
	ExpiresAt          time.Time `json:"expires_at"`
	ValidatedAt        time.Time `json:"validated_at"`
	ErrorMessage       string    `json:"error_message,omitempty"`
	CredentialIdentity string    `json:"credential_identity,omitempty"` // AWS Account ID from STS
}

// CredentialStatus represents the current status of invitation credentials
type CredentialStatus struct {
	InvitationID       string     `json:"invitation_id"`
	Status             string     `json:"status"` // "valid", "expired", "invalid"
	ExpiresAt          time.Time  `json:"expires_at"`
	RemainingTime      *string    `json:"remaining_time,omitempty"`
	LastValidated      *time.Time `json:"last_validated,omitempty"`
	CredentialIdentity string     `json:"credential_identity,omitempty"`
}

// ValidateInvitation validates invitation profile association and credentials
func (m *Manager) ValidateInvitation(ctx context.Context, invitationID string, profileMgr ProfileManager) (*ValidationResult, error) {
	m.mutex.RLock()
	invitation, exists := m.invitations[invitationID]
	m.mutex.RUnlock()

	if !exists {
		return nil, types.ErrInvitationNotFound
	}

	result := &ValidationResult{
		Valid:       false,
		ProfileID:   invitation.ProfileID,
		ValidatedAt: time.Now(),
	}

	// Check if invitation has expired
	if invitation.IsExpired() {
		result.IsExpired = true
		result.ErrorMessage = "invitation has expired"
		result.ExpiresAt = invitation.ExpiresAt
		return result, nil
	}

	result.ExpiresAt = invitation.ExpiresAt

	// Check profile association
	if invitation.ProfileID == "" {
		result.ErrorMessage = "invitation missing profile association"
		return result, nil
	}

	// Validate profile exists
	profile, err := profileMgr.GetProfile(ctx, invitation.ProfileID)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("invitation references non-existent profile: %v", err)
		return result, nil
	}

	result.ProfileExists = true

	// Test credentials
	credentialsValid, identity, err := m.TestCredentials(ctx, profileMgr, invitation.ProfileID)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("credential validation failed: %v", err)
		return result, nil
	}

	result.CredentialsValid = credentialsValid
	result.CredentialIdentity = identity

	// Check profile expiration
	if profile.ExpiresAt != nil && time.Now().After(*profile.ExpiresAt) {
		result.IsExpired = true
		result.ErrorMessage = "profile credentials have expired"
		return result, nil
	}

	// All checks passed
	result.Valid = true
	return result, nil
}

// TestCredentials tests if credentials work by making AWS STS GetCallerIdentity call
func (m *Manager) TestCredentials(ctx context.Context, profileMgr ProfileManager, profileID string) (bool, string, error) {
	// Get profile credentials
	creds, err := profileMgr.GetProfileCredentials(ctx, profileID)
	if err != nil {
		return false, "", fmt.Errorf("failed to get profile credentials: %w", err)
	}

	// Create AWS config with credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: creds,
		}),
	)
	if err != nil {
		return false, "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Test credentials with STS GetCallerIdentity
	client := sts.NewFromConfig(cfg)
	output, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return false, "", err
	}

	// Credentials work, return account identity
	identity := ""
	if output.Account != nil {
		identity = *output.Account
	}

	return true, identity, nil
}

// GetCredentialStatus returns the current credential status for an invitation
func (m *Manager) GetCredentialStatus(ctx context.Context, invitationID string, profileMgr ProfileManager) (*CredentialStatus, error) {
	m.mutex.RLock()
	invitation, exists := m.invitations[invitationID]
	m.mutex.RUnlock()

	if !exists {
		return nil, types.ErrInvitationNotFound
	}

	status := &CredentialStatus{
		InvitationID: invitationID,
		ExpiresAt:    invitation.ExpiresAt,
	}

	// Check expiration
	if invitation.IsExpired() {
		status.Status = "expired"
		return status, nil
	}

	// Calculate remaining time
	remaining := time.Until(invitation.ExpiresAt)
	if remaining > 0 {
		remainingStr := formatDuration(remaining)
		status.RemainingTime = &remainingStr
	}

	// If profile ID exists, test credentials
	if invitation.ProfileID != "" {
		credentialsValid, identity, err := m.TestCredentials(ctx, profileMgr, invitation.ProfileID)
		if err != nil || !credentialsValid {
			status.Status = "invalid"
			return status, nil
		}

		status.Status = "valid"
		status.CredentialIdentity = identity
		now := time.Now()
		status.LastValidated = &now
	} else {
		status.Status = "invalid"
	}

	return status, nil
}

// EnforceExpiration checks if invitation has expired and updates status
func (m *Manager) EnforceExpiration(ctx context.Context, invitationID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	invitation, exists := m.invitations[invitationID]
	if !exists {
		return types.ErrInvitationNotFound
	}

	// Check if expired
	if invitation.IsExpired() && invitation.Status != types.InvitationExpired {
		invitation.Status = types.InvitationExpired
		if err := m.save(); err != nil {
			return fmt.Errorf("failed to save invitation status: %w", err)
		}
	}

	return nil
}

// ValidateCredentialsExpired tests that expired credentials are rejected
// This is used for security testing to ensure time-boxed access is enforced
func (m *Manager) ValidateCredentialsExpired(ctx context.Context, invitationID string, profileMgr ProfileManager) error {
	m.mutex.RLock()
	invitation, exists := m.invitations[invitationID]
	m.mutex.RUnlock()

	if !exists {
		return types.ErrInvitationNotFound
	}

	// Check invitation is actually expired
	if !invitation.IsExpired() {
		return fmt.Errorf("invitation still valid, cannot test expiration (expires at %v)", invitation.ExpiresAt)
	}

	// If no profile associated, can't test
	if invitation.ProfileID == "" {
		return fmt.Errorf("no profile associated with invitation")
	}

	// Attempt to use credentials (should fail for time-boxed credentials)
	credentialsValid, _, err := m.TestCredentials(ctx, profileMgr, invitation.ProfileID)
	if err == nil && credentialsValid {
		// This is a SECURITY ISSUE: expired credentials still work
		return fmt.Errorf("SECURITY: expired invitation credentials still work after expiration")
	}

	// Expected: credentials should be rejected
	return nil
}

// formatDuration formats a duration into human-readable string
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours >= 48 {
		days := hours / 24
		return fmt.Sprintf("%d days", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}
