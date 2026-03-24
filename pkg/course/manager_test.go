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
