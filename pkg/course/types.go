package course

import (
	"errors"
	"fmt"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// Error constants
var (
	// ErrDuplicateCourse is returned when a course with the same code and semester already exists
	ErrDuplicateCourse = errors.New("course with this code and semester already exists")

	// ErrCourseNotFound is returned when the requested course does not exist
	ErrCourseNotFound = errors.New("course not found")

	// ErrMemberNotFound is returned when the requested member is not enrolled
	ErrMemberNotFound = errors.New("member not found")

	// ErrAlreadyEnrolled is returned when a user is already enrolled in the course
	ErrAlreadyEnrolled = errors.New("user is already enrolled in this course")

	// ErrNotAuthorized is returned when the caller lacks required permissions
	ErrNotAuthorized = errors.New("not authorized for this operation")

	// ErrCourseClosed is returned when an operation is attempted on a closed course
	ErrCourseClosed = errors.New("course is closed")
)

// CreateCourseRequest represents a request to create a new course
type CreateCourseRequest struct {
	// Code is the institutional course code (required, e.g. "CS229")
	Code string `json:"code"`

	// Title is the full course title (required)
	Title string `json:"title"`

	// Department is the offering department (optional)
	Department string `json:"department,omitempty"`

	// Semester identifies the offering period (required, e.g. "Fall 2024")
	Semester string `json:"semester"`

	// SemesterStart is when the course begins (required)
	SemesterStart time.Time `json:"semester_start"`

	// SemesterEnd is when the course ends (required)
	SemesterEnd time.Time `json:"semester_end"`

	// GracePeriodDays is days after end before auto-cleanup (default 7)
	GracePeriodDays int `json:"grace_period_days,omitempty"`

	// Owner is the primary instructor's user ID (required)
	Owner string `json:"owner"`

	// ApprovedTemplates is the initial template whitelist (empty = all allowed)
	ApprovedTemplates []string `json:"approved_templates,omitempty"`

	// PerStudentBudget is the default per-student budget in USD (0 = unlimited)
	PerStudentBudget float64 `json:"per_student_budget,omitempty"`

	// TotalBudget is the aggregate course cap in USD (0 = unlimited)
	TotalBudget float64 `json:"total_budget,omitempty"`

	// Tags are optional metadata
	Tags map[string]string `json:"tags,omitempty"`
}

// Validate validates the create course request
func (r *CreateCourseRequest) Validate() error {
	if r.Code == "" {
		return fmt.Errorf("course code is required")
	}
	if len(r.Code) > 20 {
		return fmt.Errorf("course code cannot exceed 20 characters")
	}
	if r.Title == "" {
		return fmt.Errorf("course title is required")
	}
	if len(r.Title) > 200 {
		return fmt.Errorf("course title cannot exceed 200 characters")
	}
	if r.Semester == "" {
		return fmt.Errorf("semester is required")
	}
	if r.Owner == "" {
		return fmt.Errorf("owner (instructor) is required")
	}
	if r.SemesterStart.IsZero() {
		return fmt.Errorf("semester start date is required")
	}
	if r.SemesterEnd.IsZero() {
		return fmt.Errorf("semester end date is required")
	}
	if !r.SemesterEnd.After(r.SemesterStart) {
		return fmt.Errorf("semester end must be after semester start")
	}
	if r.PerStudentBudget < 0 {
		return fmt.Errorf("per-student budget cannot be negative")
	}
	if r.TotalBudget < 0 {
		return fmt.Errorf("total budget cannot be negative")
	}
	return nil
}

// UpdateCourseRequest represents a request to update an existing course.
// All fields are optional pointers; nil means "do not change".
type UpdateCourseRequest struct {
	// Title updates the course title
	Title *string `json:"title,omitempty"`

	// Department updates the department
	Department *string `json:"department,omitempty"`

	// SemesterEnd extends or shortens the semester end date
	SemesterEnd *time.Time `json:"semester_end,omitempty"`

	// GracePeriodDays updates the cleanup grace period
	GracePeriodDays *int `json:"grace_period_days,omitempty"`

	// ApprovedTemplates replaces the template whitelist (nil = no change, empty slice = clear all)
	ApprovedTemplates []string `json:"approved_templates,omitempty"`

	// PerStudentBudget updates the default per-student budget
	PerStudentBudget *float64 `json:"per_student_budget,omitempty"`

	// TotalBudget updates the aggregate course budget cap
	TotalBudget *float64 `json:"total_budget,omitempty"`

	// Tags replaces the tag set
	Tags map[string]string `json:"tags,omitempty"`

	// Status manually sets the course status
	Status *types.CourseStatus `json:"status,omitempty"`
}

// EnrollRequest represents a request to enroll a single member
type EnrollRequest struct {
	// UserID is the member's user identifier (required unless Email is provided)
	UserID string `json:"user_id,omitempty"`

	// Email is the member's email address (used as identifier and for invitations)
	Email string `json:"email,omitempty"`

	// DisplayName is the member's human-readable name
	DisplayName string `json:"display_name,omitempty"`

	// Role is the member's role in the course (default: student)
	Role types.ClassRole `json:"role,omitempty"`

	// BudgetLimit overrides the course default for this member (0 = use course default)
	BudgetLimit float64 `json:"budget_limit,omitempty"`
}

// Validate validates the enroll request
func (r *EnrollRequest) Validate() error {
	if r.UserID == "" && r.Email == "" {
		return fmt.Errorf("either user_id or email is required")
	}
	if r.Role != "" {
		switch r.Role {
		case types.ClassRoleInstructor, types.ClassRoleTA, types.ClassRoleStudent, types.ClassRoleAuditor:
		default:
			return fmt.Errorf("invalid role %q: must be instructor, ta, student, or auditor", r.Role)
		}
	}
	return nil
}

// BulkEnrollRequest represents a request to enroll multiple members at once
type BulkEnrollRequest struct {
	// Members is the list of members to enroll
	Members []EnrollRequest `json:"members"`

	// DefaultRole is applied to any member with no explicit role (default: student)
	DefaultRole types.ClassRole `json:"default_role,omitempty"`

	// SendInvites sends invitation emails to each enrolled member when true
	SendInvites bool `json:"send_invites,omitempty"`
}

// RosterRow maps a single row from a CSV roster file
type RosterRow struct {
	// Email is the student's email address (required)
	Email string

	// DisplayName is the student's name
	DisplayName string

	// Budget is the per-student budget override in USD (0 = use course default)
	Budget float64

	// Role is the member role (optional; defaults to "student")
	Role string
}

// CourseFilter defines criteria for filtering course lists
type CourseFilter struct {
	// Owner filters to courses owned by this user ID
	Owner string `json:"owner,omitempty"`

	// Status filters by course lifecycle state
	Status *types.CourseStatus `json:"status,omitempty"`

	// Semester filters by semester identifier (exact match)
	Semester string `json:"semester,omitempty"`

	// Department filters by department
	Department string `json:"department,omitempty"`

	// Tags filters to courses matching all specified tags (AND logic)
	Tags map[string]string `json:"tags,omitempty"`
}

// Matches returns true if the course satisfies all filter criteria
func (f *CourseFilter) Matches(c *types.Course) bool {
	if f.Owner != "" && c.Owner != f.Owner {
		return false
	}
	if f.Status != nil && c.Status != *f.Status {
		return false
	}
	if f.Semester != "" && c.Semester != f.Semester {
		return false
	}
	if f.Department != "" && c.Department != f.Department {
		return false
	}
	for k, v := range f.Tags {
		if c.Tags[k] != v {
			return false
		}
	}
	return true
}

// CourseBudgetSummary summarizes budget usage across a course
type CourseBudgetSummary struct {
	// TotalBudget is the aggregate course cap in USD
	TotalBudget float64 `json:"total_budget"`

	// TotalSpent is the total spend across all students
	TotalSpent float64 `json:"total_spent"`

	// PerStudentDefault is the default allocation per student
	PerStudentDefault float64 `json:"per_student_default"`

	// Students lists per-student budget details
	Students []StudentBudgetInfo `json:"students"`
}

// StudentBudgetInfo holds per-student budget details
type StudentBudgetInfo struct {
	// UserID is the student's identifier
	UserID string `json:"user_id"`

	// DisplayName is the student's name
	DisplayName string `json:"display_name,omitempty"`

	// Email is the student's email
	Email string `json:"email,omitempty"`

	// BudgetLimit is the student's spending cap
	BudgetLimit float64 `json:"budget_limit"`

	// BudgetSpent is how much has been consumed
	BudgetSpent float64 `json:"budget_spent"`

	// Remaining is BudgetLimit - BudgetSpent (negative if over budget)
	Remaining float64 `json:"remaining"`
}

// TAResetRequest represents a request to reset a student's instance
type TAResetRequest struct {
	// StudentID is the user ID (or email) of the student whose instance to reset
	StudentID string `json:"student_id"`

	// Reason is a required explanation for the audit log
	Reason string `json:"reason"`

	// BackupRetentionDays is how long to keep the pre-reset snapshot (default 7)
	BackupRetentionDays int `json:"backup_retention_days,omitempty"`
}

// TADebugInfo is the read-only view a TA gets of a student's environment
type TADebugInfo struct {
	// CourseID is the course this debug info belongs to
	CourseID string `json:"course_id"`

	// StudentID is the user ID of the student
	StudentID string `json:"student_id"`

	// StudentEmail is the student's email
	StudentEmail string `json:"student_email,omitempty"`

	// Instances lists the student's current instances
	Instances []types.Instance `json:"instances"`

	// RecentEvents lists the last N instance state changes and errors
	RecentEvents []string `json:"recent_events,omitempty"`

	// BudgetSpent is how much of the student's budget has been consumed
	BudgetSpent float64 `json:"budget_spent"`

	// BudgetLimit is the student's spending cap (0 = unlimited)
	BudgetLimit float64 `json:"budget_limit"`
}
