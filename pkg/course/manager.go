// Package course provides course and education management functionality for Prism.
//
// This package implements course-based resource organization, template policy
// enforcement, and per-student budget tracking that enable instructors to
// manage classroom computing environments at scale.
package course

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/scttfrdmn/prism/pkg/types"
)

// Manager handles course lifecycle, member enrollment, template policies, and budgets
type Manager struct {
	coursesPath string
	mutex       sync.RWMutex
	courses     map[string]*types.Course
}

// NewManager creates a new course manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	stateDir := filepath.Join(homeDir, ".prism")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	manager := &Manager{
		coursesPath: filepath.Join(stateDir, "courses.json"),
		courses:     make(map[string]*types.Course),
	}

	if err := manager.loadCourses(); err != nil {
		return nil, fmt.Errorf("failed to load courses: %w", err)
	}

	return manager, nil
}

// --- Lifecycle ---

// CreateCourse creates a new course
func (m *Manager) CreateCourse(ctx context.Context, req *CreateCourseRequest) (*types.Course, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid course request: %w", err)
	}

	// Check for duplicate (same code + semester)
	for _, c := range m.courses {
		if strings.EqualFold(c.Code, req.Code) && strings.EqualFold(c.Semester, req.Semester) {
			return nil, ErrDuplicateCourse
		}
	}

	graceDays := req.GracePeriodDays
	if graceDays <= 0 {
		graceDays = 7
	}

	// Determine initial status based on current date
	now := time.Now()
	status := types.CourseStatusPending
	if now.After(req.SemesterStart) && now.Before(req.SemesterEnd) {
		status = types.CourseStatusActive
	}

	course := &types.Course{
		ID:                uuid.New().String(),
		Code:              req.Code,
		Title:             req.Title,
		Department:        req.Department,
		Semester:          req.Semester,
		SemesterStart:     req.SemesterStart,
		SemesterEnd:       req.SemesterEnd,
		GracePeriodDays:   graceDays,
		Owner:             req.Owner,
		Members:           []types.ClassMember{},
		ApprovedTemplates: req.ApprovedTemplates,
		PerStudentBudget:  req.PerStudentBudget,
		TotalBudget:       req.TotalBudget,
		Tags:              req.Tags,
		CreatedAt:         now,
		UpdatedAt:         now,
		Status:            status,
	}

	// Enroll the owner as instructor
	if req.Owner != "" {
		expiresAt := req.SemesterEnd.AddDate(0, 0, graceDays)
		course.Members = append(course.Members, types.ClassMember{
			UserID:    req.Owner,
			Role:      types.ClassRoleInstructor,
			AddedAt:   now,
			AddedBy:   req.Owner,
			ExpiresAt: &expiresAt,
		})
	}

	m.courses[course.ID] = course
	if err := m.saveCourses(); err != nil {
		delete(m.courses, course.ID)
		return nil, fmt.Errorf("failed to save course: %w", err)
	}

	copy := *course
	return &copy, nil
}

// GetCourse retrieves a course by ID
func (m *Manager) GetCourse(ctx context.Context, courseID string) (*types.Course, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, ok := m.courses[courseID]
	if !ok {
		return nil, ErrCourseNotFound
	}
	copy := *c
	return &copy, nil
}

// GetCourseByCode retrieves a course by its code and semester
func (m *Manager) GetCourseByCode(ctx context.Context, code, semester string) (*types.Course, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, c := range m.courses {
		if strings.EqualFold(c.Code, code) && strings.EqualFold(c.Semester, semester) {
			copy := *c
			return &copy, nil
		}
	}
	return nil, ErrCourseNotFound
}

// ListCourses retrieves all courses matching the optional filter
func (m *Manager) ListCourses(ctx context.Context, filter *CourseFilter) ([]*types.Course, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*types.Course
	for _, c := range m.courses {
		if filter != nil && !filter.Matches(c) {
			continue
		}
		copy := *c
		results = append(results, &copy)
	}
	return results, nil
}

// UpdateCourse applies a partial update to an existing course
func (m *Manager) UpdateCourse(ctx context.Context, courseID string, req *UpdateCourseRequest) (*types.Course, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return nil, ErrCourseNotFound
	}

	if req.Title != nil {
		c.Title = *req.Title
	}
	if req.Department != nil {
		c.Department = *req.Department
	}
	if req.SemesterEnd != nil {
		c.SemesterEnd = *req.SemesterEnd
	}
	if req.GracePeriodDays != nil {
		c.GracePeriodDays = *req.GracePeriodDays
	}
	if req.ApprovedTemplates != nil {
		c.ApprovedTemplates = req.ApprovedTemplates
	}
	if req.PerStudentBudget != nil {
		c.PerStudentBudget = *req.PerStudentBudget
	}
	if req.TotalBudget != nil {
		c.TotalBudget = *req.TotalBudget
	}
	if req.Tags != nil {
		c.Tags = req.Tags
	}
	if req.Status != nil {
		c.Status = *req.Status
	}
	c.UpdatedAt = time.Now()

	if err := m.saveCourses(); err != nil {
		return nil, fmt.Errorf("failed to save course update: %w", err)
	}

	copy := *c
	return &copy, nil
}

// CloseCourse transitions a course to the "closed" state
func (m *Manager) CloseCourse(ctx context.Context, courseID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return ErrCourseNotFound
	}

	c.Status = types.CourseStatusClosed
	c.UpdatedAt = time.Now()

	return m.saveCourses()
}

// DeleteCourse permanently removes a course record
func (m *Manager) DeleteCourse(ctx context.Context, courseID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.courses[courseID]; !ok {
		return ErrCourseNotFound
	}

	delete(m.courses, courseID)
	return m.saveCourses()
}

// --- Member Management ---

// EnrollMember adds a single member to the course
func (m *Manager) EnrollMember(ctx context.Context, courseID string, req *EnrollRequest) (*types.ClassMember, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid enroll request: %w", err)
	}

	c, ok := m.courses[courseID]
	if !ok {
		return nil, ErrCourseNotFound
	}
	if c.Status == types.CourseStatusArchived {
		return nil, ErrCourseClosed
	}

	role := req.Role
	if role == "" {
		role = types.ClassRoleStudent
	}

	// Determine effective user ID
	userID := req.UserID
	if userID == "" {
		userID = req.Email
	}

	// Check for duplicate (only compare email when both sides are non-empty)
	for _, mb := range c.Members {
		if mb.UserID == userID {
			return nil, ErrAlreadyEnrolled
		}
		if mb.Email != "" && req.Email != "" && mb.Email == req.Email {
			return nil, ErrAlreadyEnrolled
		}
	}

	addedBy, _ := ctx.Value("caller_id").(string)
	if addedBy == "" {
		addedBy = c.Owner
	}

	expiresAt := c.SemesterEnd.AddDate(0, 0, c.GracePeriodDays)
	member := types.ClassMember{
		UserID:      userID,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Role:        role,
		AddedAt:     time.Now(),
		AddedBy:     addedBy,
		ExpiresAt:   &expiresAt,
		BudgetLimit: req.BudgetLimit,
	}

	c.Members = append(c.Members, member)
	c.UpdatedAt = time.Now()

	if err := m.saveCourses(); err != nil {
		// roll back in-memory change
		c.Members = c.Members[:len(c.Members)-1]
		return nil, fmt.Errorf("failed to save enrollment: %w", err)
	}

	memberCopy := member
	return &memberCopy, nil
}

// enrollMemberInternal is the lock-free version used by BulkEnroll (caller holds write lock)
func (m *Manager) enrollMemberInternal(c *types.Course, req *EnrollRequest, defaultRole types.ClassRole, addedBy string) error {
	role := req.Role
	if role == "" {
		role = defaultRole
	}
	if role == "" {
		role = types.ClassRoleStudent
	}

	userID := req.UserID
	if userID == "" {
		userID = req.Email
	}

	for _, mb := range c.Members {
		if mb.UserID == userID || (req.Email != "" && mb.Email == req.Email) {
			return ErrAlreadyEnrolled
		}
	}

	expiresAt := c.SemesterEnd.AddDate(0, 0, c.GracePeriodDays)
	c.Members = append(c.Members, types.ClassMember{
		UserID:      userID,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Role:        role,
		AddedAt:     time.Now(),
		AddedBy:     addedBy,
		ExpiresAt:   &expiresAt,
		BudgetLimit: req.BudgetLimit,
	})
	return nil
}

// BulkEnroll adds multiple members in one operation.
// Partial failures skip the bad row and continue; row errors are returned alongside the count.
func (m *Manager) BulkEnroll(ctx context.Context, courseID string, req *BulkEnrollRequest) (int, []error, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return 0, nil, ErrCourseNotFound
	}
	if c.Status == types.CourseStatusArchived {
		return 0, nil, ErrCourseClosed
	}

	callerID, _ := ctx.Value("caller_id").(string)
	if callerID == "" {
		callerID = c.Owner
	}

	defaultRole := req.DefaultRole
	if defaultRole == "" {
		defaultRole = types.ClassRoleStudent
	}

	var rowErrors []error
	enrolled := 0

	for i, member := range req.Members {
		if err := member.Validate(); err != nil {
			rowErrors = append(rowErrors, fmt.Errorf("row %d: %w", i, err))
			continue
		}
		if err := m.enrollMemberInternal(c, &member, defaultRole, callerID); err != nil {
			rowErrors = append(rowErrors, fmt.Errorf("row %d (%s): %w", i, member.Email, err))
			continue
		}
		enrolled++
	}

	if enrolled > 0 {
		c.UpdatedAt = time.Now()
		if err := m.saveCourses(); err != nil {
			return 0, rowErrors, fmt.Errorf("failed to save bulk enrollment: %w", err)
		}
	}

	return enrolled, rowErrors, nil
}

// UpdateMember changes a member's role or budget limit
func (m *Manager) UpdateMember(ctx context.Context, courseID, userID string, role types.ClassRole, budgetLimit float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return ErrCourseNotFound
	}

	for i, mb := range c.Members {
		if mb.UserID == userID || mb.Email == userID {
			if role != "" {
				c.Members[i].Role = role
			}
			if budgetLimit >= 0 {
				c.Members[i].BudgetLimit = budgetLimit
			}
			c.UpdatedAt = time.Now()
			return m.saveCourses()
		}
	}
	return ErrMemberNotFound
}

// UnenrollMember removes a member from the course
func (m *Manager) UnenrollMember(ctx context.Context, courseID, userID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return ErrCourseNotFound
	}

	for i, mb := range c.Members {
		if mb.UserID == userID || mb.Email == userID {
			c.Members = append(c.Members[:i], c.Members[i+1:]...)
			c.UpdatedAt = time.Now()
			return m.saveCourses()
		}
	}
	return ErrMemberNotFound
}

// ListMembers returns enrolled members, optionally filtered by role
func (m *Manager) ListMembers(ctx context.Context, courseID string, role *types.ClassRole) ([]types.ClassMember, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, ok := m.courses[courseID]
	if !ok {
		return nil, ErrCourseNotFound
	}

	if role == nil {
		out := make([]types.ClassMember, len(c.Members))
		copy(out, c.Members)
		return out, nil
	}

	var result []types.ClassMember
	for _, mb := range c.Members {
		if mb.Role == *role {
			result = append(result, mb)
		}
	}
	return result, nil
}

// GetMember returns a single member by user ID or email
func (m *Manager) GetMember(ctx context.Context, courseID, userID string) (*types.ClassMember, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, ok := m.courses[courseID]
	if !ok {
		return nil, ErrCourseNotFound
	}

	for _, mb := range c.Members {
		if mb.UserID == userID || mb.Email == userID {
			copy := mb
			return &copy, nil
		}
	}
	return nil, ErrMemberNotFound
}

// --- Template Whitelist (#46) ---

// AddApprovedTemplate adds a template slug to the course whitelist
func (m *Manager) AddApprovedTemplate(ctx context.Context, courseID, templateSlug string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return ErrCourseNotFound
	}

	for _, t := range c.ApprovedTemplates {
		if t == templateSlug {
			return nil // already present
		}
	}

	c.ApprovedTemplates = append(c.ApprovedTemplates, templateSlug)
	c.UpdatedAt = time.Now()
	return m.saveCourses()
}

// RemoveApprovedTemplate removes a template slug from the course whitelist
func (m *Manager) RemoveApprovedTemplate(ctx context.Context, courseID, templateSlug string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return ErrCourseNotFound
	}

	for i, t := range c.ApprovedTemplates {
		if t == templateSlug {
			c.ApprovedTemplates = append(c.ApprovedTemplates[:i], c.ApprovedTemplates[i+1:]...)
			c.UpdatedAt = time.Now()
			return m.saveCourses()
		}
	}
	return nil // not present; no-op
}

// IsTemplateApproved returns true if the template is allowed for this course.
// Returns true when the course has no whitelist (empty ApprovedTemplates = unrestricted).
func (m *Manager) IsTemplateApproved(courseID, templateSlug string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, ok := m.courses[courseID]
	if !ok {
		return true // unknown course: don't block
	}

	if len(c.ApprovedTemplates) == 0 {
		return true // no restriction
	}

	for _, t := range c.ApprovedTemplates {
		if t == templateSlug {
			return true
		}
	}
	return false
}

// --- Budget (#47) ---

// DistributeBudget sets per-student budget allocations for all currently enrolled students
func (m *Manager) DistributeBudget(ctx context.Context, courseID string, amountPerStudent float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return ErrCourseNotFound
	}

	c.PerStudentBudget = amountPerStudent
	for i, mb := range c.Members {
		if mb.Role == types.ClassRoleStudent && mb.BudgetLimit == 0 {
			c.Members[i].BudgetLimit = amountPerStudent
		}
	}
	c.UpdatedAt = time.Now()
	return m.saveCourses()
}

// RecordSpend increments the spend counter for a member
func (m *Manager) RecordSpend(ctx context.Context, courseID, userID string, amount float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	c, ok := m.courses[courseID]
	if !ok {
		return ErrCourseNotFound
	}

	for i, mb := range c.Members {
		if mb.UserID == userID || mb.Email == userID {
			c.Members[i].BudgetSpent += amount
			c.UpdatedAt = time.Now()
			return m.saveCourses()
		}
	}
	return ErrMemberNotFound
}

// GetBudgetSummary returns budget usage across the entire course
func (m *Manager) GetBudgetSummary(ctx context.Context, courseID string) (*CourseBudgetSummary, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, ok := m.courses[courseID]
	if !ok {
		return nil, ErrCourseNotFound
	}

	summary := &CourseBudgetSummary{
		TotalBudget:       c.TotalBudget,
		PerStudentDefault: c.PerStudentBudget,
	}

	for _, mb := range c.Members {
		if mb.Role != types.ClassRoleStudent {
			continue
		}
		limit := mb.BudgetLimit
		if limit == 0 {
			limit = c.PerStudentBudget
		}
		summary.TotalSpent += mb.BudgetSpent
		summary.Students = append(summary.Students, StudentBudgetInfo{
			UserID:      mb.UserID,
			DisplayName: mb.DisplayName,
			Email:       mb.Email,
			BudgetLimit: limit,
			BudgetSpent: mb.BudgetSpent,
			Remaining:   limit - mb.BudgetSpent,
		})
	}

	return summary, nil
}

// --- Roster CSV Import (#47) ---

// ImportRosterCSV parses a CSV reader and bulk-enrolls students.
// CSV columns: email, display_name, budget (optional), role (optional)
// Header row is required. Malformed rows are skipped; errors returned per row.
func (m *Manager) ImportRosterCSV(ctx context.Context, courseID string, rows []RosterRow, sendInvites bool) (int, []error, error) {
	reqs := make([]EnrollRequest, 0, len(rows))
	for _, r := range rows {
		role := types.ClassRole(r.Role)
		if role == "" {
			role = types.ClassRoleStudent
		}
		reqs = append(reqs, EnrollRequest{
			Email:       r.Email,
			DisplayName: r.DisplayName,
			Role:        role,
			BudgetLimit: r.Budget,
		})
	}

	return m.BulkEnroll(ctx, courseID, &BulkEnrollRequest{
		Members:     reqs,
		DefaultRole: types.ClassRoleStudent,
		SendInvites: sendInvites,
	})
}

// ParseRosterCSV parses CSV bytes into RosterRow slice.
// Expected columns (with header): email, display_name, budget, role
func ParseRosterCSV(data []byte) ([]RosterRow, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have a header row and at least one data row")
	}

	// Build column index from header
	header := make(map[string]int)
	for i, col := range records[0] {
		header[strings.ToLower(strings.TrimSpace(col))] = i
	}
	emailIdx, ok := header["email"]
	if !ok {
		return nil, fmt.Errorf("CSV must have an 'email' column")
	}

	var rows []RosterRow
	for _, rec := range records[1:] {
		if len(rec) == 0 {
			continue
		}
		email := ""
		if emailIdx < len(rec) {
			email = strings.TrimSpace(rec[emailIdx])
		}
		if email == "" {
			continue
		}

		row := RosterRow{Email: email}

		if idx, ok := header["display_name"]; ok && idx < len(rec) {
			row.DisplayName = strings.TrimSpace(rec[idx])
		}
		if idx, ok := header["name"]; ok && row.DisplayName == "" && idx < len(rec) {
			row.DisplayName = strings.TrimSpace(rec[idx])
		}
		if idx, ok := header["budget"]; ok && idx < len(rec) {
			if v, err := strconv.ParseFloat(strings.TrimSpace(rec[idx]), 64); err == nil {
				row.Budget = v
			}
		}
		if idx, ok := header["role"]; ok && idx < len(rec) {
			row.Role = strings.TrimSpace(rec[idx])
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// --- TA Operations (#48, #49) ---

// GetStudentDebugInfo returns read-only debug information for a student in a course.
// The caller (taID) must be a TA or instructor in the course.
func (m *Manager) GetStudentDebugInfo(ctx context.Context, courseID, taID, studentID string) (*TADebugInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, ok := m.courses[courseID]
	if !ok {
		return nil, ErrCourseNotFound
	}

	if !m.isTAOrInstructorLocked(c, taID) {
		return nil, ErrNotAuthorized
	}

	var studentMember *types.ClassMember
	for i, mb := range c.Members {
		if mb.UserID == studentID || mb.Email == studentID {
			studentMember = &c.Members[i]
			break
		}
	}
	if studentMember == nil {
		return nil, ErrMemberNotFound
	}

	limit := studentMember.BudgetLimit
	if limit == 0 {
		limit = c.PerStudentBudget
	}

	return &TADebugInfo{
		CourseID:     courseID,
		StudentID:    studentMember.UserID,
		StudentEmail: studentMember.Email,
		Instances:    nil, // populated by the handler using the AWS manager
		BudgetSpent:  studentMember.BudgetSpent,
		BudgetLimit:  limit,
	}, nil
}

// ResetStudentInstance records a TA reset request in the course audit trail.
// The actual AWS snapshot + re-provision is handled by the daemon handler (requires awsManager).
func (m *Manager) ResetStudentInstance(ctx context.Context, courseID, taID string, req *TAResetRequest) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, ok := m.courses[courseID]
	if !ok {
		return ErrCourseNotFound
	}

	if !m.isTAOrInstructorLocked(c, taID) {
		return ErrNotAuthorized
	}

	// Verify student is enrolled
	for _, mb := range c.Members {
		if mb.UserID == req.StudentID || mb.Email == req.StudentID {
			return nil // authorized; actual reset done by daemon handler
		}
	}
	return ErrMemberNotFound
}

// --- Maintenance ---

// CheckAndCloseExpiredCourses transitions courses past their end date + grace period to "closed"
func (m *Manager) CheckAndCloseExpiredCourses(ctx context.Context) ([]string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	var closed []string

	for id, c := range m.courses {
		if c.Status != types.CourseStatusActive {
			continue
		}
		deadline := c.SemesterEnd.AddDate(0, 0, c.GracePeriodDays)
		if now.After(deadline) {
			c.Status = types.CourseStatusClosed
			c.UpdatedAt = now
			closed = append(closed, id)
		}
	}

	if len(closed) > 0 {
		if err := m.saveCourses(); err != nil {
			return nil, fmt.Errorf("failed to save course closures: %w", err)
		}
	}

	return closed, nil
}

// --- Internal Helpers ---

// isTAOrInstructorLocked checks if userID is a TA or instructor. Caller must hold at least RLock.
func (m *Manager) isTAOrInstructorLocked(c *types.Course, userID string) bool {
	if c.Owner == userID {
		return true
	}
	for _, mb := range c.Members {
		if (mb.UserID == userID || mb.Email == userID) &&
			(mb.Role == types.ClassRoleTA || mb.Role == types.ClassRoleInstructor) {
			return true
		}
	}
	return false
}

// loadCourses loads the course map from disk
func (m *Manager) loadCourses() error {
	if _, err := os.Stat(m.coursesPath); os.IsNotExist(err) {
		return nil // no file yet; start empty
	}

	data, err := os.ReadFile(m.coursesPath)
	if err != nil {
		return fmt.Errorf("failed to read courses file: %w", err)
	}

	var courses map[string]*types.Course
	if err := json.Unmarshal(data, &courses); err != nil {
		return fmt.Errorf("failed to parse courses file: %w", err)
	}

	m.courses = courses
	return nil
}

// saveCourses persists the course map to disk atomically
func (m *Manager) saveCourses() error {
	data, err := json.MarshalIndent(m.courses, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal courses: %w", err)
	}

	tempPath := m.coursesPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary courses file: %w", err)
	}

	if err := os.Rename(tempPath, m.coursesPath); err != nil {
		return fmt.Errorf("failed to rename courses file: %w", err)
	}

	return nil
}
