package course

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testManager creates a Manager backed by a temporary directory
func testManager(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	orig := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", orig) })
	_ = os.Setenv("HOME", dir)

	m, err := NewManager()
	require.NoError(t, err)
	return m
}

// testCreateReq returns a valid CreateCourseRequest
func testCreateReq() *CreateCourseRequest {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(90 * 24 * time.Hour)
	return &CreateCourseRequest{
		Code:          "CS101",
		Title:         "Intro to Computing",
		Department:    "Computer Science",
		Semester:      "Spring 2025",
		SemesterStart: start,
		SemesterEnd:   end,
		Owner:         "instructor1",
	}
}

// --- CreateCourse ---

func TestCreateCourse(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	req := testCreateReq()

	c, err := m.CreateCourse(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "CS101", c.Code)
	assert.Equal(t, "instructor1", c.Owner)
	assert.Equal(t, types.CourseStatusActive, c.Status) // started yesterday
	assert.Len(t, c.Members, 1)                         // owner auto-enrolled
	assert.Equal(t, types.ClassRoleInstructor, c.Members[0].Role)
}

func TestCreateCourse_DefaultGracePeriod(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	req := testCreateReq()
	req.GracePeriodDays = 0

	c, err := m.CreateCourse(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 7, c.GracePeriodDays)
}

func TestCreateCourse_PendingStatus(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	req := testCreateReq()
	req.SemesterStart = time.Now().Add(24 * time.Hour) // starts tomorrow

	c, err := m.CreateCourse(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, types.CourseStatusPending, c.Status)
}

func TestCreateCourse_DuplicateRejected(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	_, err := m.CreateCourse(ctx, testCreateReq())
	require.NoError(t, err)

	_, err = m.CreateCourse(ctx, testCreateReq())
	assert.ErrorIs(t, err, ErrDuplicateCourse)
}

func TestCreateCourse_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	req := testCreateReq()

	tests := []struct {
		name   string
		mutate func(*CreateCourseRequest)
	}{
		{"missing code", func(r *CreateCourseRequest) { r.Code = "" }},
		{"missing title", func(r *CreateCourseRequest) { r.Title = "" }},
		{"missing semester", func(r *CreateCourseRequest) { r.Semester = "" }},
		{"missing owner", func(r *CreateCourseRequest) { r.Owner = "" }},
		{"end before start", func(r *CreateCourseRequest) { r.SemesterEnd = r.SemesterStart.Add(-time.Hour) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := *req
			tt.mutate(&r)
			_, err := m.CreateCourse(ctx, &r)
			assert.Error(t, err)
		})
	}
}

// --- EnrollMember ---

func TestEnrollMember(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, err := m.CreateCourse(ctx, testCreateReq())
	require.NoError(t, err)

	mb, err := m.EnrollMember(ctx, c.ID, &EnrollRequest{
		UserID: "student1",
		Email:  "s1@uni.edu",
		Role:   types.ClassRoleStudent,
	})
	require.NoError(t, err)
	assert.Equal(t, "student1", mb.UserID)
	assert.Equal(t, types.ClassRoleStudent, mb.Role)
}

func TestEnrollMember_DefaultRole(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	mb, err := m.EnrollMember(ctx, c.ID, &EnrollRequest{Email: "s2@uni.edu"})
	require.NoError(t, err)
	assert.Equal(t, types.ClassRoleStudent, mb.Role)
}

func TestEnrollMember_DuplicateRejected(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, err := m.EnrollMember(ctx, c.ID, &EnrollRequest{Email: "s3@uni.edu"})
	require.NoError(t, err)

	_, err = m.EnrollMember(ctx, c.ID, &EnrollRequest{Email: "s3@uni.edu"})
	assert.ErrorIs(t, err, ErrAlreadyEnrolled)
}

// --- BulkEnroll ---

func TestBulkEnroll(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	req := &BulkEnrollRequest{
		Members: []EnrollRequest{
			{Email: "a@uni.edu", DisplayName: "Alice"},
			{Email: "b@uni.edu", DisplayName: "Bob"},
			{Email: ""}, // invalid — no email or userID
		},
		DefaultRole: types.ClassRoleStudent,
	}

	enrolled, rowErrors, err := m.BulkEnroll(ctx, c.ID, req)
	require.NoError(t, err)
	assert.Equal(t, 2, enrolled)
	assert.Len(t, rowErrors, 1) // the empty-email row
}

func TestBulkEnroll_SkipsDuplicates(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	req := &BulkEnrollRequest{
		Members: []EnrollRequest{
			{Email: "x@uni.edu"},
			{Email: "x@uni.edu"}, // duplicate
		},
	}

	enrolled, rowErrors, err := m.BulkEnroll(ctx, c.ID, req)
	require.NoError(t, err)
	assert.Equal(t, 1, enrolled)
	assert.Len(t, rowErrors, 1)
}

// --- Template Whitelist (#46) ---

func TestIsTemplateApproved_EmptyListAllowsAll(t *testing.T) {
	m := testManager(t)
	c, _ := m.CreateCourse(context.Background(), testCreateReq())

	assert.True(t, m.IsTemplateApproved(c.ID, "any-template"))
	assert.True(t, m.IsTemplateApproved(c.ID, "gpu-xl"))
}

func TestIsTemplateApproved_EnforcesList(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	req := testCreateReq()
	req.ApprovedTemplates = []string{"python-ml", "r-research"}
	c, _ := m.CreateCourse(ctx, req)

	assert.True(t, m.IsTemplateApproved(c.ID, "python-ml"))
	assert.True(t, m.IsTemplateApproved(c.ID, "r-research"))
	assert.False(t, m.IsTemplateApproved(c.ID, "gpu-xl"))
}

func TestAddRemoveApprovedTemplate(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	require.NoError(t, m.AddApprovedTemplate(ctx, c.ID, "python-ml"))
	assert.False(t, m.IsTemplateApproved(c.ID, "r-research")) // list now enforced
	assert.True(t, m.IsTemplateApproved(c.ID, "python-ml"))

	require.NoError(t, m.RemoveApprovedTemplate(ctx, c.ID, "python-ml"))
	assert.True(t, m.IsTemplateApproved(c.ID, "any-template")) // list empty again
}

// --- DistributeBudget (#47) ---

func TestDistributeBudget(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	// Enroll two students
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{Email: "s1@uni.edu"})
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{Email: "s2@uni.edu"})

	require.NoError(t, m.DistributeBudget(ctx, c.ID, 50.0))

	// Verify per-student limits
	summary, err := m.GetBudgetSummary(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, 50.0, summary.PerStudentDefault)
	assert.Len(t, summary.Students, 2)
	for _, s := range summary.Students {
		assert.Equal(t, 50.0, s.BudgetLimit)
	}
}

// --- ParseRosterCSV (#47) ---

func TestParseRosterCSV(t *testing.T) {
	csvData := []byte("email,display_name,budget,role\nalice@uni.edu,Alice,50,student\nbob@uni.edu,Bob,,\n")
	rows, err := ParseRosterCSV(csvData)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "alice@uni.edu", rows[0].Email)
	assert.Equal(t, "Alice", rows[0].DisplayName)
	assert.Equal(t, 50.0, rows[0].Budget)
	assert.Equal(t, "bob@uni.edu", rows[1].Email)
}

func TestParseRosterCSV_MissingHeader(t *testing.T) {
	_, err := ParseRosterCSV([]byte("name,budget\nalice,50\n"))
	assert.Error(t, err) // no 'email' column
}

func TestParseRosterCSV_EmptySkipped(t *testing.T) {
	csvData := []byte("email,display_name\n\nalice@uni.edu,Alice\n")
	rows, err := ParseRosterCSV(csvData)
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

// --- CloseCourse ---

func TestCloseCourse(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())
	assert.Equal(t, types.CourseStatusActive, c.Status)

	require.NoError(t, m.CloseCourse(ctx, c.ID))

	updated, err := m.GetCourse(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, types.CourseStatusClosed, updated.Status)
}

// --- CheckAndCloseExpiredCourses ---

func TestCheckAndCloseExpiredCourses(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	req.SemesterStart = time.Now().Add(-60 * 24 * time.Hour) // started 60 days ago
	req.SemesterEnd = time.Now().Add(-30 * 24 * time.Hour)   // ended 30 days ago
	req.GracePeriodDays = 7                                  // grace expired 23 days ago

	c, _ := m.CreateCourse(ctx, req)
	// Force to active so the check can close it
	_, _ = m.UpdateCourse(ctx, c.ID, &UpdateCourseRequest{
		Status: func() *types.CourseStatus { s := types.CourseStatusActive; return &s }(),
	})

	closed, err := m.CheckAndCloseExpiredCourses(ctx)
	require.NoError(t, err)
	assert.Contains(t, closed, c.ID)

	updated, _ := m.GetCourse(ctx, c.ID)
	assert.Equal(t, types.CourseStatusClosed, updated.Status)
}

// --- Persistence ---

func TestPersistence(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	defer func() { _ = os.Setenv("HOME", origHome) }()
	_ = os.Setenv("HOME", dir)

	m1, _ := NewManager()
	c, _ := m1.CreateCourse(ctx, testCreateReq())

	// Create a second manager pointing at the same directory
	m2, err := NewManager()
	require.NoError(t, err)

	loaded, err := m2.GetCourse(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, c.Code, loaded.Code)
}

// --- v0.16.0 tests ---

// TestCheckStudentBudget verifies per-student budget enforcement.
func TestCheckStudentBudget(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	req.PerStudentBudget = 10.0
	c, err := m.CreateCourse(ctx, req)
	require.NoError(t, err)

	_, err = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "alice", Email: "alice@uni.edu", Role: types.ClassRoleStudent})
	require.NoError(t, err)

	// Under budget — should allow
	err = m.CheckStudentBudget(c.ID, "alice", 5.0)
	assert.NoError(t, err)

	// Record some spend
	_ = m.RecordSpend(ctx, c.ID, "alice", 8.0)

	// Over budget — should block
	err = m.CheckStudentBudget(c.ID, "alice", 5.0)
	require.Error(t, err)
	var budgetErr *BudgetExceededError
	require.ErrorAs(t, err, &budgetErr)
	assert.Equal(t, 8.0, budgetErr.Spent)
	assert.Equal(t, 10.0, budgetErr.Limit)

	// Unknown student — should not block
	err = m.CheckStudentBudget(c.ID, "unknown", 100.0)
	assert.NoError(t, err)

	// Unknown course — should not block
	err = m.CheckStudentBudget("bad-course-id", "alice", 5.0)
	assert.NoError(t, err)
}

// TestCheckStudentBudget_Unlimited verifies unlimited budgets are never blocked.
func TestCheckStudentBudget_Unlimited(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	req.PerStudentBudget = 0 // unlimited
	c, _ := m.CreateCourse(ctx, req)
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "alice", Role: types.ClassRoleStudent})
	_ = m.RecordSpend(ctx, c.ID, "alice", 9999.0)

	err := m.CheckStudentBudget(c.ID, "alice", 9999.0)
	assert.NoError(t, err)
}

// ── v0.22.5 gap tests (#545) ──────────────────────────────────────────────

// --- GetCourse ---

func TestGetCourse(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	created, _ := m.CreateCourse(ctx, testCreateReq())

	got, err := m.GetCourse(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "CS101", got.Code)
}

func TestGetCourse_NotFound(t *testing.T) {
	m := testManager(t)
	_, err := m.GetCourse(context.Background(), "no-such-id")
	assert.ErrorIs(t, err, ErrCourseNotFound)
}

// --- ListCourses ---

func TestListCourses_Empty(t *testing.T) {
	m := testManager(t)
	courses, err := m.ListCourses(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, courses)
}

func TestListCourses_WithOwnerFilter(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req1 := testCreateReq()
	req1.Owner = "owner-a"
	_, _ = m.CreateCourse(ctx, req1)

	req2 := testCreateReq()
	req2.Semester = "Fall 2025" // different semester to avoid duplicate
	req2.Owner = "owner-b"
	_, _ = m.CreateCourse(ctx, req2)

	courses, err := m.ListCourses(ctx, &CourseFilter{Owner: "owner-a"})
	require.NoError(t, err)
	require.Len(t, courses, 1)
	assert.Equal(t, "owner-a", courses[0].Owner)
}

// --- UpdateCourse ---

func TestUpdateCourse(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	newTitle := "Advanced Computing"
	newDept := "Engineering"
	updated, err := m.UpdateCourse(ctx, c.ID, &UpdateCourseRequest{
		Title:      &newTitle,
		Department: &newDept,
	})
	require.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)
	assert.Equal(t, newDept, updated.Department)

	// Verify persisted
	reloaded, _ := m.GetCourse(ctx, c.ID)
	assert.Equal(t, newTitle, reloaded.Title)
}

// --- DeleteCourse ---

func TestDeleteCourse(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	require.NoError(t, m.DeleteCourse(ctx, c.ID))

	_, err := m.GetCourse(ctx, c.ID)
	assert.ErrorIs(t, err, ErrCourseNotFound)
}

// --- GetMember ---

func TestGetMember(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, err := m.EnrollMember(ctx, c.ID, &EnrollRequest{
		UserID: "student-x",
		Email:  "sx@uni.edu",
		Role:   types.ClassRoleStudent,
	})
	require.NoError(t, err)

	mb, err := m.GetMember(ctx, c.ID, "student-x")
	require.NoError(t, err)
	assert.Equal(t, "student-x", mb.UserID)
}

func TestGetMember_NotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, err := m.GetMember(ctx, c.ID, "ghost")
	assert.ErrorIs(t, err, ErrMemberNotFound)
}

// --- UnenrollMember ---

func TestUnenrollMember(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, err := m.EnrollMember(ctx, c.ID, &EnrollRequest{Email: "leave@uni.edu"})
	require.NoError(t, err)

	require.NoError(t, m.UnenrollMember(ctx, c.ID, "leave@uni.edu"))

	_, err = m.GetMember(ctx, c.ID, "leave@uni.edu")
	assert.ErrorIs(t, err, ErrMemberNotFound)
}

// --- UpdateMember ---

func TestUpdateMember_Role(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, err := m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "upgrade-me", Email: "u@uni.edu"})
	require.NoError(t, err)

	require.NoError(t, m.UpdateMember(ctx, c.ID, "upgrade-me", types.ClassRoleTA, 0))

	mb, _ := m.GetMember(ctx, c.ID, "upgrade-me")
	assert.Equal(t, types.ClassRoleTA, mb.Role)
}

// --- RecordSpend ---

func TestRecordSpend(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, err := m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "spender", Email: "sp@uni.edu"})
	require.NoError(t, err)

	require.NoError(t, m.RecordSpend(ctx, c.ID, "spender", 12.50))

	mb, _ := m.GetMember(ctx, c.ID, "spender")
	assert.Equal(t, 12.50, mb.BudgetSpent)
}

// --- GetBudgetSummary ---

func TestGetBudgetSummary(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	req.PerStudentBudget = 100.0
	c, _ := m.CreateCourse(ctx, req)

	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "s1", Email: "s1@uni.edu", Role: types.ClassRoleStudent})
	_ = m.RecordSpend(ctx, c.ID, "s1", 40.0)

	summary, err := m.GetBudgetSummary(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, summary.PerStudentDefault)
	assert.Equal(t, 40.0, summary.TotalSpent)
	require.Len(t, summary.Students, 1)
	assert.Equal(t, 40.0, summary.Students[0].BudgetSpent)
	assert.Equal(t, 60.0, summary.Students[0].Remaining)
}

// --- ImportRosterCSV ---

func TestImportRosterCSV(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	rows := []RosterRow{
		{Email: "r1@uni.edu", DisplayName: "Row One"},
		{Email: "r2@uni.edu", DisplayName: "Row Two"},
		{Email: "r3@uni.edu", DisplayName: "Row Three"},
	}

	enrolled, rowErrors, err := m.ImportRosterCSV(ctx, c.ID, rows, false)
	require.NoError(t, err)
	assert.Equal(t, 3, enrolled)
	assert.Empty(t, rowErrors)
}

// --- ResetStudentInstance ---

func TestResetStudentInstance(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	// Enroll TA and student
	_, err := m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "ta1", Email: "ta1@uni.edu", Role: types.ClassRoleTA})
	require.NoError(t, err)
	_, err = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "stu1", Email: "stu1@uni.edu", Role: types.ClassRoleStudent})
	require.NoError(t, err)

	err = m.ResetStudentInstance(ctx, c.ID, "ta1", &TAResetRequest{
		StudentID: "stu1",
		Reason:    "test reset",
	})
	assert.NoError(t, err)
}

func TestResetStudentInstance_Unauthorized(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "regular", Email: "r@uni.edu", Role: types.ClassRoleStudent})

	err := m.ResetStudentInstance(ctx, c.ID, "regular", &TAResetRequest{StudentID: "other", Reason: "x"})
	assert.ErrorIs(t, err, ErrNotAuthorized)
}

// TestParseCanvasCSV verifies Canvas LMS roster parsing.
func TestParseCanvasCSV(t *testing.T) {
	csv := `Student,ID,SIS Login ID,Section,SIS User ID,Email Address,Points Possible
Alice Liddell,1,alice,CS101-A,101,alice@uni.edu,N/A
Bob Baker,2,bob,CS101-A,102,bob@uni.edu,N/A
Points Possible,,,,,,100`

	rows, err := ParseCanvasCSV([]byte(csv))
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "alice@uni.edu", rows[0].Email)
	assert.Equal(t, "Alice Liddell", rows[0].DisplayName)
	assert.Equal(t, "bob@uni.edu", rows[1].Email)
}

// TestParseBlackboardCSV verifies Blackboard LMS roster parsing.
func TestParseBlackboardCSV(t *testing.T) {
	csv := `Last Name,First Name,Username,Student ID,Last Access,Email
Liddell,Alice,alice101,S001,2026-01-01,alice@uni.edu
Baker,Bob,bob202,S002,2026-01-02,bob@uni.edu`

	rows, err := ParseBlackboardCSV([]byte(csv))
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "alice@uni.edu", rows[0].Email)
	assert.Equal(t, "Liddell, Alice", rows[0].DisplayName) // Blackboard "Last, First" format
	assert.Equal(t, "alice101", rows[0].Role)              // username stored in Role field
}

// TestGetProvisioningContext verifies provisioning context resolution.
func TestGetProvisioningContext(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	req.DefaultTemplate = "python-ml"
	c, _ := m.CreateCourse(ctx, req)
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "alice", Email: "alice@uni.edu", Role: types.ClassRoleStudent, BudgetLimit: 20.0})

	// No override — should use DefaultTemplate
	pctx, err := m.GetProvisioningContext(ctx, c.ID, "alice", "")
	require.NoError(t, err)
	assert.Equal(t, "python-ml", pctx.Template)
	assert.Equal(t, "alice", pctx.StudentID)
	assert.Equal(t, 20.0, pctx.BudgetLimit)

	// With override
	pctx, err = m.GetProvisioningContext(ctx, c.ID, "alice", "r-research")
	require.NoError(t, err)
	assert.Equal(t, "r-research", pctx.Template)
}

// TestGetCourseOverview verifies the TA dashboard overview.
func TestGetCourseOverview(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	c, _ := m.CreateCourse(ctx, req)
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "alice", Role: types.ClassRoleStudent})
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "bob", Role: types.ClassRoleStudent})

	overview, err := m.GetCourseOverview(ctx, c.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, c.ID, overview.CourseID)
	assert.Equal(t, 2, overview.TotalStudents)
}

// TestGetUsageReport verifies the semester usage report.
func TestGetUsageReport(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	req := testCreateReq()
	req.TotalBudget = 100.0
	c, _ := m.CreateCourse(ctx, req)
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{UserID: "alice", Email: "alice@uni.edu", Role: types.ClassRoleStudent, BudgetLimit: 50.0})
	_ = m.RecordSpend(ctx, c.ID, "alice", 12.5)

	report, err := m.GetUsageReport(ctx, c.ID,
		map[string]float64{"alice": 4.0},
		map[string]int{"alice": 1},
	)
	require.NoError(t, err)
	assert.Equal(t, c.ID, report.CourseID)
	assert.InDelta(t, 12.5, report.TotalSpent, 0.001)
	require.Len(t, report.Students, 1)
	assert.Equal(t, 4.0, report.Students[0].InstanceHours)
	assert.Equal(t, 1, report.Students[0].InstanceCount)
}

// TestListArchiveEligibleCourses verifies courses eligible for archiving.
func TestListArchiveEligibleCourses(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	// Course that ended long ago (past grace period)
	past := &CreateCourseRequest{
		Code:            "OLD101",
		Title:           "Old Course",
		Semester:        "Fall 2020",
		SemesterStart:   time.Now().Add(-200 * 24 * time.Hour),
		SemesterEnd:     time.Now().Add(-100 * 24 * time.Hour),
		Owner:           "prof1",
		GracePeriodDays: 7,
	}
	old, _ := m.CreateCourse(ctx, past)
	// Manually set to closed so it's eligible
	m.mutex.Lock()
	m.courses[old.ID].Status = types.CourseStatusClosed
	m.mutex.Unlock()

	// Active course — not eligible
	active, _ := m.CreateCourse(ctx, testCreateReq())
	_ = active

	eligible, err := m.ListArchiveEligibleCourses(ctx)
	require.NoError(t, err)
	require.Len(t, eligible, 1)
	assert.Equal(t, old.ID, eligible[0].ID)
}

// TestMarkCourseArchived verifies the archive transition.
func TestMarkCourseArchived(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	c, _ := m.CreateCourse(ctx, testCreateReq())

	// Must be closed first
	_ = m.CloseCourse(ctx, c.ID)

	err := m.MarkCourseArchived(ctx, c.ID)
	require.NoError(t, err)

	updated, _ := m.GetCourse(ctx, c.ID)
	assert.Equal(t, types.CourseStatusArchived, updated.Status)
}

// TestMarkCourseArchived_MustBeClosed verifies active courses cannot be archived directly.
func TestMarkCourseArchived_MustBeClosed(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	c, _ := m.CreateCourse(ctx, testCreateReq())
	// Still active — archive should fail
	err := m.MarkCourseArchived(ctx, c.ID)
	require.Error(t, err)
}

// --- v0.19.0: TA Access Management (#48, #160) ---

func TestGrantTAAccess_Success(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	ta, err := m.GrantTAAccess(ctx, c.ID, "ta@uni.edu", "Alice TA")
	require.NoError(t, err)
	require.NotNil(t, ta)
	assert.Equal(t, types.ClassRoleTA, ta.Role)
	assert.Equal(t, "ta@uni.edu", ta.Email)
}

func TestGrantTAAccess_Idempotent(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, err := m.GrantTAAccess(ctx, c.ID, "ta@uni.edu", "Alice TA")
	require.NoError(t, err)

	// Granting again returns the existing TA, no error
	ta2, err := m.GrantTAAccess(ctx, c.ID, "ta@uni.edu", "Alice TA")
	require.NoError(t, err)
	require.NotNil(t, ta2)
	assert.Equal(t, types.ClassRoleTA, ta2.Role)
}

func TestGrantTAAccess_CourseNotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	_, err := m.GrantTAAccess(ctx, "nonexistent", "ta@uni.edu", "")
	require.Error(t, err)
}

func TestListTAAccess(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, _ = m.GrantTAAccess(ctx, c.ID, "ta1@uni.edu", "TA One")
	_, _ = m.GrantTAAccess(ctx, c.ID, "ta2@uni.edu", "TA Two")

	// Also enroll a regular student — should not appear in TA list
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{Email: "student@uni.edu", Role: types.ClassRoleStudent})

	tas, err := m.ListTAAccess(ctx, c.ID)
	require.NoError(t, err)
	assert.Len(t, tas, 2)
	for _, ta := range tas {
		assert.Equal(t, types.ClassRoleTA, ta.Role)
	}
}

func TestRevokeTAAccess_Success(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, _ = m.GrantTAAccess(ctx, c.ID, "ta@uni.edu", "Alice TA")

	err := m.RevokeTAAccess(ctx, c.ID, "ta@uni.edu")
	require.NoError(t, err)

	tas, _ := m.ListTAAccess(ctx, c.ID)
	assert.Len(t, tas, 0)
}

func TestRevokeTAAccess_NotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	err := m.RevokeTAAccess(ctx, c.ID, "nobody@uni.edu")
	require.Error(t, err)
}

func TestLogTASSHConnect_RequiresTARole(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	// Enroll as student — not TA → must be denied
	_, _ = m.EnrollMember(ctx, c.ID, &EnrollRequest{Email: "stu@uni.edu", Role: types.ClassRoleStudent})

	entry := TAAccessEntry{
		TAID:      "stu@uni.edu",
		StudentID: "other@uni.edu",
		Reason:    "debug",
	}
	err := m.LogTASSHConnect(ctx, c.ID, entry)
	require.Error(t, err) // ErrNotAuthorized
}

func TestLogTASSHConnect_Success(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, _ = m.GrantTAAccess(ctx, c.ID, "ta@uni.edu", "TA")

	entry := TAAccessEntry{
		TAID:      "ta@uni.edu",
		StudentID: "stu@uni.edu",
		Reason:    "office hours",
	}
	require.NoError(t, m.LogTASSHConnect(ctx, c.ID, entry))
}

// --- v0.19.0: Template Enforcement (#47) ---

func TestCheckTemplateAllowed_EmptyList(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	c, _ := m.CreateCourse(ctx, testCreateReq())

	// No approved templates set → all allowed
	err := m.CheckTemplateAllowed(c.ID, "python-ml")
	require.NoError(t, err)
}

func TestCheckTemplateAllowed_Approved(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_ = m.AddApprovedTemplate(ctx, c.ID, "python-ml")
	_ = m.AddApprovedTemplate(ctx, c.ID, "r-research")

	require.NoError(t, m.CheckTemplateAllowed(c.ID, "python-ml"))
}

func TestCheckTemplateAllowed_Rejected(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_ = m.AddApprovedTemplate(ctx, c.ID, "python-ml")

	err := m.CheckTemplateAllowed(c.ID, "gpu-cluster")
	require.Error(t, err)
}

func TestCheckTemplateAllowed_UnknownCourse(t *testing.T) {
	m := testManager(t)
	// Unknown course → IsTemplateApproved returns true (permissive) → no error
	err := m.CheckTemplateAllowed("no-such-course", "anything")
	require.NoError(t, err)
}

// --- v0.19.0: Shared Course Materials (#167) ---

func TestSetAndGetCourseMaterials(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	err := m.SetCourseMaterials(ctx, c.ID, "fs-abc123", "/mnt/materials", 50)
	require.NoError(t, err)

	vol, err := m.GetCourseMaterials(ctx, c.ID)
	require.NoError(t, err)
	require.NotNil(t, vol)
	assert.Equal(t, "fs-abc123", vol.EFSID)
	assert.Equal(t, "/mnt/materials", vol.MountPath)
	assert.Equal(t, 50, vol.SizeGB)
	assert.Equal(t, "available", vol.State)
}

func TestGetCourseMaterials_None(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	vol, err := m.GetCourseMaterials(ctx, c.ID)
	require.NoError(t, err)
	assert.Nil(t, vol)
}

func TestGetCourseMaterials_CourseNotFound(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)

	_, err := m.GetCourseMaterials(ctx, "nonexistent")
	require.Error(t, err)
}

func TestSetCourseMaterials_Persisted(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_ = m.SetCourseMaterials(ctx, c.ID, "fs-persist", "/mnt/shared", 20)

	// Re-read from same manager
	updated, _ := m.GetCourse(ctx, c.ID)
	assert.Equal(t, "fs-persist", updated.SharedMaterialsEFSID)
	assert.Equal(t, "/mnt/shared", updated.SharedMaterialsMountPath)
	assert.Equal(t, 20, updated.SharedMaterialsSizeGB)
}

func TestSetCourseMaterials_AuditLogged(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_ = m.SetCourseMaterials(ctx, c.ID, "fs-xyz", "/mnt/data", 30)

	entries, err := m.QueryAuditLog(c.ID, "", time.Time{}, 100)
	require.NoError(t, err)
	found := false
	for _, e := range entries {
		if e.Action == AuditActionMaterialsCreate {
			found = true
			break
		}
	}
	assert.True(t, found, "materials.create audit entry expected")
}

// --- v0.19.0: TA Access Audit (#48) ---

func TestGrantTAAccess_AuditLogged(t *testing.T) {
	ctx := context.Background()
	m := testManager(t)
	c, _ := m.CreateCourse(ctx, testCreateReq())

	_, _ = m.GrantTAAccess(ctx, c.ID, "ta@uni.edu", "Alice")

	entries, err := m.QueryAuditLog(c.ID, "", time.Time{}, 100)
	require.NoError(t, err)
	found := false
	for _, e := range entries {
		if e.Action == AuditActionTAAccessGrant {
			found = true
			assert.Equal(t, "ta@uni.edu", e.Target)
			break
		}
	}
	assert.True(t, found, "ta.access.grant audit entry expected")
}
