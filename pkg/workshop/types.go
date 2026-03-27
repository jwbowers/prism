// Package workshop provides workshop and event management functionality for Prism.
//
// This package implements time-bounded, event-style workshop management — conferences,
// tutorials, and hackathons where an organizer needs bulk workspace pre-provisioning,
// real-time participant monitoring, and auto-termination at a fixed end time.
package workshop

import (
	"errors"
	"fmt"
	"time"
)

// Error sentinels
var (
	ErrWorkshopNotFound    = errors.New("workshop not found")
	ErrParticipantNotFound = errors.New("participant not found")
	ErrWorkshopFull        = errors.New("workshop has reached maximum participant capacity")
	ErrWorkshopEnded       = errors.New("workshop has already ended")
	ErrConfigNotFound      = errors.New("workshop config not found")
	ErrDuplicateConfig     = errors.New("workshop config with this name already exists")
)

// WorkshopStatus indicates the lifecycle state of a workshop
type WorkshopStatus string

const (
	WorkshopStatusDraft    WorkshopStatus = "draft"
	WorkshopStatusActive   WorkshopStatus = "active"
	WorkshopStatusEnded    WorkshopStatus = "ended"
	WorkshopStatusArchived WorkshopStatus = "archived"
)

// WorkshopEvent represents a time-bounded event workspace session.
// Unlike a Course (semester-length, email-enrollment), a WorkshopEvent is
// short-lived, joined via a shared token, and supports bulk provisioning.
type WorkshopEvent struct {
	// ID is the unique workshop identifier
	ID string `json:"id"`

	// Title is the human-readable workshop name
	Title string `json:"title"`

	// Description is an optional longer description
	Description string `json:"description,omitempty"`

	// Owner is the organizer's user ID
	Owner string `json:"owner"`

	// Template is the default workspace template slug for participants
	Template string `json:"template"`

	// ApprovedTemplates is an optional whitelist of allowed templates (#176).
	// When empty, only Template is allowed. When non-empty, participants may choose
	// any listed template (and Template must be in the list).
	ApprovedTemplates []string `json:"approved_templates,omitempty"`

	// MaxParticipants caps the number of participants who can join (0 = unlimited)
	MaxParticipants int `json:"max_participants"`

	// BudgetPerParticipant is the per-participant cost allocation in USD (0 = unlimited)
	BudgetPerParticipant float64 `json:"budget_per_participant,omitempty"`

	// StartTime is when the workshop begins
	StartTime time.Time `json:"start_time"`

	// EndTime is when the workshop ends (all instances auto-stopped at this time)
	EndTime time.Time `json:"end_time"`

	// EarlyAccessHours is how many hours before StartTime participants can join (#181).
	// 0 means no early access (join only after StartTime).
	EarlyAccessHours int `json:"early_access_hours,omitempty"`

	// Status is the current workshop lifecycle state
	Status WorkshopStatus `json:"status"`

	// JoinToken is the shared token code participants use to join.
	// Auto-generated on creation. Redeeming the token adds the participant.
	JoinToken string `json:"join_token,omitempty"`

	// Participants lists all joined participants
	Participants []WorkshopParticipant `json:"participants,omitempty"`

	// PolicyRestrictions defines launch policy constraints for this workshop (#177).
	// When set, all participant workspace launches must satisfy these policies.
	PolicyRestrictions []string `json:"policy_restrictions,omitempty"`

	// CreatedAt is when the workshop was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the workshop was last modified
	UpdatedAt time.Time `json:"updated_at"`
}

// WorkshopParticipant represents a single joined participant
type WorkshopParticipant struct {
	// UserID is the participant's user identifier
	UserID string `json:"user_id"`

	// Email is the participant's email address
	Email string `json:"email,omitempty"`

	// DisplayName is the participant's human-readable name
	DisplayName string `json:"display_name,omitempty"`

	// JoinedAt is when the participant redeemed the join token
	JoinedAt time.Time `json:"joined_at"`

	// InstanceID is the AWS instance ID of the provisioned workspace (empty until provisioned)
	InstanceID string `json:"instance_id,omitempty"`

	// InstanceName is the logical name of the provisioned workspace
	InstanceName string `json:"instance_name,omitempty"`

	// Status is the participant's workspace state: pending|provisioned|running|stopped
	Status string `json:"status"`

	// Progress is 0-100 progress value for tracking workshop exercises (#184)
	Progress int `json:"progress,omitempty"`
}

// WorkshopDashboard is the live status view for the workshop organizer (#179)
type WorkshopDashboard struct {
	WorkshopID        string                `json:"workshop_id"`
	Title             string                `json:"title"`
	TotalParticipants int                   `json:"total_participants"`
	ActiveInstances   int                   `json:"active_instances"`
	StoppedInstances  int                   `json:"stopped_instances"`
	PendingInstances  int                   `json:"pending_instances"`
	TotalSpent        float64               `json:"total_spent"`
	TimeRemaining     string                `json:"time_remaining"`
	Status            WorkshopStatus        `json:"status"`
	Participants      []WorkshopParticipant `json:"participants"`
}

// WorkshopConfig is a reusable workshop configuration template (#183)
type WorkshopConfig struct {
	// Name is the unique config identifier
	Name string `json:"name"`

	// Template is the workspace template slug
	Template string `json:"template"`

	// MaxParticipants is the default capacity
	MaxParticipants int `json:"max_participants"`

	// BudgetPerParticipant is the default per-participant budget
	BudgetPerParticipant float64 `json:"budget_per_participant,omitempty"`

	// DurationHours is the default workshop length in hours
	DurationHours int `json:"duration_hours"`

	// EarlyAccessHours is the default early-access window
	EarlyAccessHours int `json:"early_access_hours,omitempty"`

	// Description is an optional description of this config
	Description string `json:"description,omitempty"`

	// CreatedAt is when the config was saved
	CreatedAt time.Time `json:"created_at"`
}

// CreateWorkshopRequest is the payload for creating a new workshop
type CreateWorkshopRequest struct {
	Title                string    `json:"title"`
	Description          string    `json:"description,omitempty"`
	Owner                string    `json:"owner"`
	Template             string    `json:"template"`
	ApprovedTemplates    []string  `json:"approved_templates,omitempty"`
	MaxParticipants      int       `json:"max_participants"`
	BudgetPerParticipant float64   `json:"budget_per_participant,omitempty"`
	StartTime            time.Time `json:"start_time"`
	EndTime              time.Time `json:"end_time"`
	EarlyAccessHours     int       `json:"early_access_hours,omitempty"`
	PolicyRestrictions   []string  `json:"policy_restrictions,omitempty"`
}

// Validate validates the create workshop request
func (r *CreateWorkshopRequest) Validate() error {
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if r.Owner == "" {
		return fmt.Errorf("owner is required")
	}
	if r.Template == "" {
		return fmt.Errorf("template is required")
	}
	if r.StartTime.IsZero() {
		return fmt.Errorf("start_time is required")
	}
	if r.EndTime.IsZero() {
		return fmt.Errorf("end_time is required")
	}
	if !r.EndTime.After(r.StartTime) {
		return fmt.Errorf("end_time must be after start_time")
	}
	if r.MaxParticipants < 0 {
		return fmt.Errorf("max_participants cannot be negative")
	}
	if r.BudgetPerParticipant < 0 {
		return fmt.Errorf("budget_per_participant cannot be negative")
	}
	return nil
}

// UpdateWorkshopRequest is the payload for updating an existing workshop
type UpdateWorkshopRequest struct {
	Title                *string    `json:"title,omitempty"`
	Description          *string    `json:"description,omitempty"`
	Template             *string    `json:"template,omitempty"`
	ApprovedTemplates    []string   `json:"approved_templates,omitempty"`
	MaxParticipants      *int       `json:"max_participants,omitempty"`
	BudgetPerParticipant *float64   `json:"budget_per_participant,omitempty"`
	EndTime              *time.Time `json:"end_time,omitempty"`
	EarlyAccessHours     *int       `json:"early_access_hours,omitempty"`
	PolicyRestrictions   []string   `json:"policy_restrictions,omitempty"`
}

// WorkshopFilter defines criteria for filtering workshop lists
type WorkshopFilter struct {
	Owner  string          `json:"owner,omitempty"`
	Status *WorkshopStatus `json:"status,omitempty"`
}

// Matches returns true if the workshop satisfies all filter criteria
func (f *WorkshopFilter) Matches(w *WorkshopEvent) bool {
	if f.Owner != "" && w.Owner != f.Owner {
		return false
	}
	if f.Status != nil && w.Status != *f.Status {
		return false
	}
	return true
}

// ProvisionResult is returned by ProvisionWorkshop
type ProvisionResult struct {
	WorkshopID  string   `json:"workshop_id"`
	Provisioned int      `json:"provisioned"`
	Skipped     int      `json:"skipped"`
	Errors      []string `json:"errors,omitempty"`
}

// EndResult is returned by EndWorkshop
type EndResult struct {
	WorkshopID string   `json:"workshop_id"`
	Stopped    int      `json:"stopped"`
	Errors     []string `json:"errors,omitempty"`
}

// DownloadManifest is returned by GetDownloadManifest (#180)
type DownloadManifest struct {
	WorkshopID   string            `json:"workshop_id"`
	Participants []ParticipantWork `json:"participants"`
	GeneratedAt  time.Time         `json:"generated_at"`
}

// ParticipantWork represents one participant's downloadable work
type ParticipantWork struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	InstanceName string `json:"instance_name,omitempty"`
	DownloadNote string `json:"download_note"`
}

// persistedState is the on-disk format for workshop storage
type persistedState struct {
	Workshops []WorkshopEvent  `json:"workshops"`
	Configs   []WorkshopConfig `json:"configs"`
}
