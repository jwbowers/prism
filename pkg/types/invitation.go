// Package types provides invitation management types for Prism v0.5.11+
//
// This file defines the core types for the user invitation system, enabling
// project owners to invite collaborators with role-based permissions.
package types

import (
	"errors"
	"time"
)

// Invitation represents a project invitation sent to a user
type Invitation struct {
	// ID is the unique invitation identifier
	ID string `json:"id"`

	// ProjectID is the project being invited to
	ProjectID string `json:"project_id"`

	// ProjectName is the name of the project (enriched for display)
	ProjectName string `json:"project_name,omitempty"`

	// Email is the invitee's email address
	Email string `json:"email"`

	// Role is the role to be assigned when invitation is accepted
	Role ProjectRole `json:"role"`

	// Token is the secure random token used for acceptance/decline
	Token string `json:"token"`

	// InvitedBy is the user ID of who sent the invitation
	InvitedBy string `json:"invited_by"`

	// ProfileID is the profile associated with this invitation (for credential validation)
	ProfileID string `json:"profile_id,omitempty"`

	// InvitedAt is when the invitation was sent
	InvitedAt time.Time `json:"invited_at"`

	// ExpiresAt is when the invitation expires (default: 7 days)
	ExpiresAt time.Time `json:"expires_at"`

	// Status is the current invitation status
	Status InvitationStatus `json:"status"`

	// AcceptedAt is when the invitation was accepted (if applicable)
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`

	// DeclinedAt is when the invitation was declined (if applicable)
	DeclinedAt *time.Time `json:"declined_at,omitempty"`

	// DeclineReason is an optional reason for declining
	DeclineReason string `json:"decline_reason,omitempty"`

	// ResendCount tracks how many times the invitation was resent
	ResendCount int `json:"resend_count"`

	// LastResent is when the invitation was last resent
	LastResent *time.Time `json:"last_resent,omitempty"`

	// Message is an optional personal message from the inviter
	Message string `json:"message,omitempty"`
}

// InvitationStatus represents the current status of an invitation
type InvitationStatus string

const (
	// InvitationPending indicates the invitation is awaiting response
	InvitationPending InvitationStatus = "pending"

	// InvitationAccepted indicates the invitation was accepted
	InvitationAccepted InvitationStatus = "accepted"

	// InvitationDeclined indicates the invitation was declined
	InvitationDeclined InvitationStatus = "declined"

	// InvitationExpired indicates the invitation expired
	InvitationExpired InvitationStatus = "expired"

	// InvitationRevoked indicates the invitation was revoked by inviter
	InvitationRevoked InvitationStatus = "revoked"
)

// InvitationSummary provides an overview of invitations for a project
type InvitationSummary struct {
	// ProjectID is the project identifier
	ProjectID string `json:"project_id"`

	// TotalInvitations is the total number of invitations sent
	TotalInvitations int `json:"total_invitations"`

	// PendingCount is the number of pending invitations
	PendingCount int `json:"pending_count"`

	// AcceptedCount is the number of accepted invitations
	AcceptedCount int `json:"accepted_count"`

	// DeclinedCount is the number of declined invitations
	DeclinedCount int `json:"declined_count"`

	// ExpiredCount is the number of expired invitations
	ExpiredCount int `json:"expired_count"`

	// RecentInvitations lists the 10 most recent invitations
	RecentInvitations []*Invitation `json:"recent_invitations"`
}

// InvitationFilter provides filtering options for listing invitations
type InvitationFilter struct {
	// ProjectID filters by project
	ProjectID string `json:"project_id,omitempty"`

	// Email filters by invitee email
	Email string `json:"email,omitempty"`

	// Status filters by invitation status
	Status InvitationStatus `json:"status,omitempty"`

	// InvitedBy filters by inviter user ID
	InvitedBy string `json:"invited_by,omitempty"`

	// IncludeExpired includes expired invitations (default: false)
	IncludeExpired bool `json:"include_expired"`

	// Limit limits the number of results (default: 50)
	Limit int `json:"limit,omitempty"`

	// Offset skips N results for pagination
	Offset int `json:"offset,omitempty"`
}

// IsExpired checks if the invitation has expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt) && i.Status == InvitationPending
}

// CanAccept checks if the invitation can be accepted
func (i *Invitation) CanAccept() bool {
	return i.Status == InvitationPending && !i.IsExpired()
}

// CanDecline checks if the invitation can be declined
func (i *Invitation) CanDecline() bool {
	return i.Status == InvitationPending && !i.IsExpired()
}

// CanResend checks if the invitation can be resent
func (i *Invitation) CanResend() bool {
	return i.Status == InvitationPending && !i.IsExpired()
}

// CanRevoke checks if the invitation can be revoked
func (i *Invitation) CanRevoke() bool {
	return i.Status == InvitationPending
}

// Validate validates the invitation data
func (i *Invitation) Validate() error {
	if i.ProjectID == "" {
		return ErrInvalidProjectID
	}
	if i.Email == "" {
		return ErrInvalidEmail
	}
	if i.Role == "" {
		return ErrInvalidRole
	}
	if i.Token == "" {
		return ErrInvalidToken
	}
	if i.InvitedBy == "" {
		return ErrInvalidInviter
	}
	return nil
}

// BulkInvitationRequest represents a bulk invitation request (v0.5.12+)
type BulkInvitationRequest struct {
	// Invitations is the list of invitations to create
	Invitations []BulkInvitationEntry `json:"invitations"`

	// DefaultRole is applied when invitation entry has no role specified
	DefaultRole ProjectRole `json:"default_role,omitempty"`

	// DefaultMessage is applied when invitation entry has no message
	DefaultMessage string `json:"default_message,omitempty"`

	// ExpiresIn is the default expiration duration (e.g., "7d", "30d")
	ExpiresIn string `json:"expires_in,omitempty"`

	// ExpiresAt is an absolute expiration time (mutually exclusive with ExpiresIn)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// BulkInvitationEntry represents a single invitation in a bulk request
type BulkInvitationEntry struct {
	// Email is the invitee's email address (required)
	Email string `json:"email"`

	// Role overrides DefaultRole if specified
	Role ProjectRole `json:"role,omitempty"`

	// Message overrides DefaultMessage if specified
	Message string `json:"message,omitempty"`
}

// BulkInvitationResponse represents the response from a bulk invitation operation
type BulkInvitationResponse struct {
	// Summary provides aggregate statistics
	Summary BulkInvitationSummary `json:"summary"`

	// Results contains the outcome for each invitation
	Results []BulkInvitationResult `json:"results"`

	// Message is a human-readable summary
	Message string `json:"message"`
}

// BulkInvitationSummary provides aggregate statistics for a bulk operation
type BulkInvitationSummary struct {
	// Total is the total number of invitations processed
	Total int `json:"total"`

	// Sent is the number of successfully sent invitations
	Sent int `json:"sent"`

	// Skipped is the number of skipped invitations (duplicates, already members)
	Skipped int `json:"skipped"`

	// Failed is the number of failed invitations
	Failed int `json:"failed"`
}

// BulkInvitationResult represents the result for a single invitation in a bulk operation
type BulkInvitationResult struct {
	// Email is the invitee's email address
	Email string `json:"email"`

	// Status is the result status: "sent", "skipped", or "failed"
	Status string `json:"status"`

	// InvitationID is the ID of the created invitation (only for "sent" status)
	InvitationID string `json:"invitation_id,omitempty"`

	// Reason explains why the invitation was skipped
	Reason string `json:"reason,omitempty"`

	// Error contains the error message for failed invitations
	Error string `json:"error,omitempty"`
}

// Invitation-related errors (using simple error values for comparability)
var (
	ErrInvitationNotFound      = errors.New("invitation not found")
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrInvitationAlreadyUsed   = errors.New("invitation has already been accepted or declined")
	ErrInvitationRevoked       = errors.New("invitation has been revoked")
	ErrInvalidToken            = errors.New("invalid invitation token")
	ErrInvalidEmail            = errors.New("invalid email address")
	ErrInvalidRole             = errors.New("invalid role")
	ErrInvalidInviter          = errors.New("invalid inviter")
	ErrDuplicateInvitation     = errors.New("an invitation for this email already exists for this project")
	ErrInvitationNotPending    = errors.New("invitation is not in pending status")
	ErrInsufficientPermissions = errors.New("insufficient permissions to perform this action")
	ErrProjectNotFound         = errors.New("project not found")
	ErrInvalidProjectID        = errors.New("invalid project ID")
)

// SharedInvitationToken represents a shared token that multiple users can redeem (v0.5.13+)
//
// Shared tokens are ideal for workshops, conferences, and guest lectures where:
// - A single token code works for multiple users (e.g., first 60 redeemers)
// - No pre-registration or email list required
// - QR code support for easy mobile access
// - First-come-first-served access control
type SharedInvitationToken struct {
	// ID is the unique token identifier
	ID string `json:"id"`

	// ProjectID is the project this token grants access to
	ProjectID string `json:"project_id"`

	// Token is the human-readable token code (e.g., WORKSHOP-NEURIPS-2025)
	Token string `json:"token"`

	// Name is the display name for this token
	Name string `json:"name"`

	// Role is the role assigned to users who redeem this token
	Role ProjectRole `json:"role"`

	// Message is an optional welcome message for redeemers
	Message string `json:"message,omitempty"`

	// CreatedBy is the user ID of who created the token
	CreatedBy string `json:"created_by"`

	// CreatedAt is when the token was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the token expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// RedemptionLimit is the maximum number of redemptions allowed
	RedemptionLimit int `json:"redemption_limit"`

	// RedemptionCount is the current number of redemptions
	RedemptionCount int `json:"redemption_count"`

	// Status is the current token status
	Status SharedTokenStatus `json:"status"`

	// Redemptions is the audit trail of redemptions
	Redemptions []SharedTokenRedemption `json:"redemptions"`
}

// SharedTokenStatus represents the current status of a shared token
type SharedTokenStatus string

const (
	// SharedTokenActive indicates the token is active and can be redeemed
	SharedTokenActive SharedTokenStatus = "active"

	// SharedTokenExpired indicates the token has expired
	SharedTokenExpired SharedTokenStatus = "expired"

	// SharedTokenRevoked indicates the token was revoked by creator
	SharedTokenRevoked SharedTokenStatus = "revoked"

	// SharedTokenExhausted indicates all redemptions have been used
	SharedTokenExhausted SharedTokenStatus = "exhausted"
)

// SharedTokenRedemption represents a single redemption of a shared token
type SharedTokenRedemption struct {
	// RedeemedBy is the user email/ID who redeemed the token
	RedeemedBy string `json:"redeemed_by"`

	// RedeemedAt is when the redemption occurred
	RedeemedAt time.Time `json:"redeemed_at"`

	// Position is the redemption order (1, 2, 3, ...)
	Position int `json:"position"`
}

// CreateSharedTokenRequest represents a request to create a shared token
type CreateSharedTokenRequest struct {
	// ProjectID is the project this token grants access to (inferred from URL path if not provided)
	ProjectID string `json:"project_id,omitempty"`

	// Name is the display name for this token
	Name string `json:"name"`

	// Role is the role assigned to redeemers
	Role ProjectRole `json:"role"`

	// Message is an optional welcome message
	Message string `json:"message,omitempty"`

	// RedemptionLimit is the maximum number of redemptions
	RedemptionLimit int `json:"redemption_limit"`

	// ExpiresIn is the expiration duration (e.g., "7d")
	ExpiresIn string `json:"expires_in,omitempty"`

	// ExpiresAt is an absolute expiration time
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// CreatedBy is the user ID creating the token (inferred from auth context if not provided)
	CreatedBy string `json:"created_by,omitempty"`
}

// RedeemSharedTokenRequest represents a request to redeem a shared token
type RedeemSharedTokenRequest struct {
	// Token is the token code to redeem
	Token string `json:"token"`

	// RedeemedBy is the user email/ID redeeming the token
	RedeemedBy string `json:"redeemed_by"`
}

// RedeemSharedTokenResponse represents the response from redeeming a shared token
type RedeemSharedTokenResponse struct {
	// Success indicates if redemption was successful
	Success bool `json:"success"`

	// ProjectID is the project the user was added to
	ProjectID string `json:"project_id"`

	// Role is the role assigned to the user
	Role ProjectRole `json:"role"`

	// RedemptionPosition is the user's redemption order
	RedemptionPosition int `json:"redemption_position"`

	// RemainingRedemptions is how many redemptions are left
	RemainingRedemptions int `json:"remaining_redemptions"`

	// Message is a human-readable message
	Message string `json:"message"`
}

// IsExpired checks if the shared token has expired
func (t *SharedInvitationToken) IsExpired() bool {
	return t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt)
}

// IsExhausted checks if all redemptions have been used
func (t *SharedInvitationToken) IsExhausted() bool {
	return t.RedemptionCount >= t.RedemptionLimit
}

// CanRedeem checks if the token can still be redeemed
func (t *SharedInvitationToken) CanRedeem() bool {
	return t.Status == SharedTokenActive && !t.IsExpired() && !t.IsExhausted()
}

// RemainingRedemptions returns the number of remaining redemptions
func (t *SharedInvitationToken) RemainingRedemptions() int {
	remaining := t.RedemptionLimit - t.RedemptionCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// UpdateStatus updates the token status based on current state
func (t *SharedInvitationToken) UpdateStatus() {
	if t.Status == SharedTokenRevoked {
		return // Revoked tokens stay revoked
	}

	if t.IsExpired() {
		t.Status = SharedTokenExpired
	} else if t.IsExhausted() {
		t.Status = SharedTokenExhausted
	} else {
		t.Status = SharedTokenActive
	}
}

// Validate validates the shared token data
func (t *SharedInvitationToken) Validate() error {
	if t.ProjectID == "" {
		return ErrInvalidProjectID
	}
	if t.Token == "" {
		return ErrInvalidToken
	}
	if t.Name == "" {
		return errors.New("token name is required")
	}
	if t.Role == "" {
		return ErrInvalidRole
	}
	if t.RedemptionLimit <= 0 {
		return errors.New("redemption limit must be positive")
	}
	if t.CreatedBy == "" {
		return ErrInvalidInviter
	}
	return nil
}

// Shared token errors
var (
	ErrSharedTokenNotFound  = errors.New("shared token not found")
	ErrSharedTokenExpired   = errors.New("shared token has expired")
	ErrSharedTokenRevoked   = errors.New("shared token has been revoked")
	ErrSharedTokenExhausted = errors.New("shared token has reached redemption limit")
	ErrAlreadyRedeemed      = errors.New("you have already redeemed this token")
	ErrInvalidRedemption    = errors.New("invalid redemption request")
)
