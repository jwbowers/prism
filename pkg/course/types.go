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

	// DefaultTemplate is the template slug auto-assigned to students on workspace provision.
	DefaultTemplate string `json:"default_template,omitempty"`

	// AutoProvisionOnEnroll creates a workspace for the student automatically upon
	// enrollment when true (requires DefaultTemplate to be set).
	AutoProvisionOnEnroll bool `json:"auto_provision_on_enroll,omitempty"`
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

	// DefaultTemplate updates the default workspace template
	DefaultTemplate *string `json:"default_template,omitempty"`

	// AutoProvisionOnEnroll updates the auto-provision setting
	AutoProvisionOnEnroll *bool `json:"auto_provision_on_enroll,omitempty"`
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

// --- v0.16.0 types ---

// BudgetExceededError is returned when a launch would exceed the student's allocation.
type BudgetExceededError struct {
	UserID    string
	Spent     float64
	Limit     float64
	Requested float64
}

func (e *BudgetExceededError) Error() string {
	return fmt.Sprintf("budget exceeded: $%.2f spent of $%.2f limit (requesting ~$%.2f more)",
		e.Spent, e.Limit, e.Requested)
}

// ErrBudgetExceeded is the sentinel error for student budget enforcement.
var ErrBudgetExceeded = errors.New("student budget exceeded")

// RosterFormat identifies the source LMS format of an uploaded CSV file.
type RosterFormat string

const (
	// RosterFormatPrism is the native Prism CSV format (email,display_name,role,budget).
	RosterFormatPrism RosterFormat = "prism"
	// RosterFormatCanvas is the Canvas LMS student roster export format.
	RosterFormatCanvas RosterFormat = "canvas"
	// RosterFormatBlackboard is the Blackboard Learn user list export format.
	RosterFormatBlackboard RosterFormat = "blackboard"
)

// ProvisioningContext carries what the daemon needs to launch a workspace on behalf of a student.
type ProvisioningContext struct {
	CourseID     string
	StudentID    string
	StudentEmail string
	Template     string
	BudgetLimit  float64
	BudgetSpent  float64
}

// ArchiveResult is returned by the course archive operation.
type ArchiveResult struct {
	CourseID         string   `json:"course_id"`
	InstancesStopped []string `json:"instances_stopped"`
	SnapshotsCreated []string `json:"snapshots_created"`
	Errors           []string `json:"errors,omitempty"`
}

// StudentStatus summarises a single student's current state for the TA dashboard.
type StudentStatus struct {
	UserID       string           `json:"user_id"`
	Email        string           `json:"email"`
	DisplayName  string           `json:"display_name,omitempty"`
	Instances    []types.Instance `json:"instances"`
	BudgetSpent  float64          `json:"budget_spent"`
	BudgetLimit  float64          `json:"budget_limit"`
	BudgetStatus string           `json:"budget_status"` // "ok" | "warning" (>80%) | "exceeded"
	LastActive   *time.Time       `json:"last_active,omitempty"`
}

// CourseOverview is the payload for GET /api/v1/courses/{id}/overview.
type CourseOverview struct {
	CourseID         string          `json:"course_id"`
	CourseCode       string          `json:"course_code"`
	TotalStudents    int             `json:"total_students"`
	ActiveInstances  int             `json:"active_instances"`
	TotalBudgetSpent float64         `json:"total_budget_spent"`
	Students         []StudentStatus `json:"students"`
	GeneratedAt      time.Time       `json:"generated_at"`
}

// StudentUsageRecord is one row in the semester usage report.
type StudentUsageRecord struct {
	UserID        string  `json:"user_id"`
	Email         string  `json:"email"`
	DisplayName   string  `json:"display_name,omitempty"`
	TotalSpent    float64 `json:"total_spent"`
	BudgetLimit   float64 `json:"budget_limit"`
	InstanceHours float64 `json:"instance_hours"`
	InstanceCount int     `json:"instance_count"`
}

// UsageReport is the payload for GET /api/v1/courses/{id}/report.
type UsageReport struct {
	CourseID    string               `json:"course_id"`
	CourseCode  string               `json:"course_code"`
	Semester    string               `json:"semester"`
	TotalSpent  float64              `json:"total_spent"`
	TotalBudget float64              `json:"total_budget"`
	Students    []StudentUsageRecord `json:"students"`
	GeneratedAt time.Time            `json:"generated_at"`
}

// TAAccessEntry records a single TA SSH access session for the audit trail.
type TAAccessEntry struct {
	// TAID is the user ID (or email) of the TA who connected
	TAID string `json:"ta_id"`
	// TAEmail is the TA's email for display purposes
	TAEmail string `json:"ta_email,omitempty"`
	// StudentID is the user ID (or email) of the student whose instance was accessed
	StudentID string `json:"student_id"`
	// StudentEmail is the student's email for display purposes
	StudentEmail string `json:"student_email,omitempty"`
	// Reason is a required explanation provided at access time
	Reason string `json:"reason"`
	// ConnectedAt is when the SSH session was initiated
	ConnectedAt time.Time `json:"connected_at"`
	// InstanceID is the AWS instance ID that was accessed
	InstanceID string `json:"instance_id,omitempty"`
	// PublicIP is the public IP address of the student's instance
	PublicIP string `json:"public_ip,omitempty"`
}

// TASSHConnectRequest is the payload for POST /courses/{id}/ta-access/connect
type TASSHConnectRequest struct {
	// StudentID is the user ID or email of the student to connect to
	StudentID string `json:"student_id"`
	// Reason is a mandatory explanation for the audit log
	Reason string `json:"reason"`
}

// SharedMaterialsVolume describes the EFS volume used for shared course materials.
type SharedMaterialsVolume struct {
	// CourseID links back to the owning course
	CourseID string `json:"course_id"`
	// EFSID is the AWS EFS filesystem ID
	EFSID string `json:"efs_id"`
	// SizeGB is the advisory provisioned size in GB (EFS is elastic)
	SizeGB int `json:"size_gb"`
	// MountPath is where the EFS is mounted inside student instances
	MountPath string `json:"mount_path"`
	// State is "creating" | "available" | "error"
	State string `json:"state"`
	// CreatedAt is when the volume was created
	CreatedAt time.Time `json:"created_at"`
	// MountedInstanceCount is how many student instances currently have it mounted
	MountedInstanceCount int `json:"mounted_instance_count"`
}

// CourseMaterialsCreateRequest is the payload for POST /courses/{id}/materials
type CourseMaterialsCreateRequest struct {
	// SizeGB is the advisory size (EFS is elastic; default 50)
	SizeGB int `json:"size_gb,omitempty"`
	// MountPath is where to mount in student instances (default "/mnt/course-materials")
	MountPath string `json:"mount_path,omitempty"`
}

// WorkspaceResetResult is returned by the reset endpoint.
type WorkspaceResetResult struct {
	// StudentID identifies whose workspace was reset
	StudentID string `json:"student_id"`
	// BackupSnapshotID is the snapshot ID created before reset (empty if --no-backup)
	BackupSnapshotID string `json:"backup_snapshot_id,omitempty"`
	// BackupDownloadURL is a pre-signed URL for the student to retrieve their backed-up work
	BackupDownloadURL string `json:"backup_download_url,omitempty"`
	// BackupExpiresAt is when the download URL expires (7 days by default)
	BackupExpiresAt *time.Time `json:"backup_expires_at,omitempty"`
	// Status is "reset_scheduled" | "reset_complete"
	Status string `json:"status"`
}

// TAAccessGrantRequest is the payload for POST /courses/{id}/ta-access
type TAAccessGrantRequest struct {
	// Email is the TA's email address to grant access
	Email string `json:"email"`
	// DisplayName is the TA's display name (optional)
	DisplayName string `json:"display_name,omitempty"`
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
