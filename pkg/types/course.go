// Package types provides course and education management types for Prism.
//
// This file defines the core types for course-based resource organization and
// education management, enabling instructors to organize student instances,
// enforce template policies, and track per-student budgets.
package types

import "time"

// ClassRole defines the role of a member within a course
type ClassRole string

const (
	// ClassRoleInstructor is the course owner with full control
	ClassRoleInstructor ClassRole = "instructor"

	// ClassRoleTA is a teaching assistant with limited admin capabilities
	ClassRoleTA ClassRole = "ta"

	// ClassRoleStudent is a standard student with resource launch privileges
	ClassRoleStudent ClassRole = "student"

	// ClassRoleAuditor has read-only access with no budget allocation
	ClassRoleAuditor ClassRole = "auditor"
)

// CourseStatus indicates the lifecycle state of a course
type CourseStatus string

const (
	// CourseStatusPending means the course has been created but the semester has not started
	CourseStatusPending CourseStatus = "pending"

	// CourseStatusActive means the course is currently running
	CourseStatusActive CourseStatus = "active"

	// CourseStatusClosed means the semester has ended; resources in grace period
	CourseStatusClosed CourseStatus = "closed"

	// CourseStatusArchived means the course and all resources have been cleaned up
	CourseStatusArchived CourseStatus = "archived"
)

// Course represents an academic course with associated student workspaces and budgets
type Course struct {
	// ID is the unique course identifier
	ID string `json:"id"`

	// Code is the institutional course code (e.g. "CS229")
	Code string `json:"code"`

	// Title is the full course title (e.g. "Machine Learning")
	Title string `json:"title"`

	// Department is the offering department (optional)
	Department string `json:"department,omitempty"`

	// Semester identifies the offering period (e.g. "Fall 2024")
	Semester string `json:"semester"`

	// SemesterStart is when the course begins
	SemesterStart time.Time `json:"semester_start"`

	// SemesterEnd is when the course ends
	SemesterEnd time.Time `json:"semester_end"`

	// GracePeriodDays is how many days after SemesterEnd before auto-cleanup (default 7)
	GracePeriodDays int `json:"grace_period_days"`

	// Owner is the primary instructor's user ID
	Owner string `json:"owner"`

	// Members lists all enrolled class members (students, TAs, auditors)
	Members []ClassMember `json:"members,omitempty"`

	// ApprovedTemplates is the whitelist of allowed workspace templates.
	// When empty, all templates are allowed (no restriction).
	ApprovedTemplates []string `json:"approved_templates,omitempty"`

	// PerStudentBudget is the default budget allocation per student in USD (0 = unlimited)
	PerStudentBudget float64 `json:"per_student_budget,omitempty"`

	// TotalBudget is the aggregate course budget cap in USD (0 = unlimited)
	TotalBudget float64 `json:"total_budget,omitempty"`

	// Tags are optional metadata for organization and reporting
	Tags map[string]string `json:"tags,omitempty"`

	// CreatedAt is when the course was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the course was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// Status indicates the current course lifecycle state
	Status CourseStatus `json:"status"`

	// LaunchPrevented prevents new instance launches when true (set by budget enforcement)
	LaunchPrevented bool `json:"launch_prevented,omitempty"`
}

// ClassMember represents an enrolled course participant
type ClassMember struct {
	// UserID is the member's user identifier
	UserID string `json:"user_id"`

	// Email is the member's email address (used for invitations and notifications)
	Email string `json:"email"`

	// DisplayName is the member's human-readable name
	DisplayName string `json:"display_name,omitempty"`

	// Role defines the member's permissions within the course
	Role ClassRole `json:"role"`

	// AddedAt is when the member was enrolled
	AddedAt time.Time `json:"added_at"`

	// AddedBy is the user ID of whoever enrolled this member
	AddedBy string `json:"added_by"`

	// ExpiresAt is when membership expires (defaults to SemesterEnd + GracePeriodDays)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// BudgetSpent tracks how much of the member's allocation has been consumed
	BudgetSpent float64 `json:"budget_spent,omitempty"`

	// BudgetLimit is the per-member override (0 = use course PerStudentBudget)
	BudgetLimit float64 `json:"budget_limit,omitempty"`
}
