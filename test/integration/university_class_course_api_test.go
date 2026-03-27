//go:build integration
// +build integration

// Package integration — University Class Persona using the /api/v1/courses/* API (v0.17.0 port).
//
// This is the v0.17.0 port of TestUniversityClassPersona_ProfThompson.
// The original test uses the project/budget API; this version uses the
// dedicated course API endpoints introduced in v0.14.0.
package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	apiclient "github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUniversityClassPersona_CourseAPI replicates the Prof. Thompson classroom
// scenario but drives every operation through the /api/v1/courses/* endpoints.
//
// SCENARIO:
// Prof. Thompson teaches "Data Science 101" with 5 students (reduced from 20)
// and 1 TA. Each student needs an identical Python/Jupyter environment.
//
// WORKFLOW:
// 1. Create course via POST /api/v1/courses
// 2. Enroll TA and students via POST /api/v1/courses/{id}/members
// 3. Whitelist a template
// 4. Distribute per-student budget
// 5. TA views course overview (GET /api/v1/courses/{id}/overview)
// 6. TA views audit log (GET /api/v1/courses/{id}/audit)
// 7. Close and archive the course
func TestUniversityClassPersona_CourseAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test — requires running daemon")
	}

	ctx := context.Background()
	apiClient := apiclient.NewClientWithOptions("http://localhost:8947", apiclient.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)
	hc, ok := apiClient.(*apiclient.HTTPClient)
	require.True(t, ok)

	const numStudents = 5
	courseCode := fmt.Sprintf("DATASCI101-%d", time.Now().UnixMilli()%100000)

	t.Log("=================================================================")
	t.Log("PERSONA TEST (Course API): Prof. Thompson — Data Science 101")
	t.Log("=================================================================")

	// ── Phase 1: Create course ────────────────────────────────────────────

	t.Run("Phase1_CreateCourse", func(t *testing.T) {
		t.Logf("Creating course: %s", courseCode)

		created, err := hc.CreateCourse(ctx, map[string]interface{}{
			"code":               courseCode,
			"title":              "Data Science 101",
			"department":         "Computer Science",
			"semester":           "Spring-2099",
			"semester_start":     "2099-01-15T00:00:00Z",
			"semester_end":       "2099-05-15T00:00:00Z",
			"owner":              "thompson@cs.university.edu",
			"per_student_budget": 24.0,
			"total_budget":       1200.0,
		})
		require.NoError(t, err, "CreateCourse")

		courseID, _ := created["code"].(string)
		// Use the ID field for subsequent operations
		idField, _ := created["id"].(string)
		require.NotEmpty(t, idField, "expected course ID")
		registry.RegisterCourse(idField)

		t.Logf("✓ Course created: %s (ID: %s, code: %s)", "Data Science 101", idField, courseID)
	})

	// We need the course ID for subsequent phases — fetch it from the list
	var courseID string
	t.Run("Phase1b_GetCourseID", func(t *testing.T) {
		listed, err := hc.ListCourses(ctx, "")
		require.NoError(t, err)
		courses, _ := listed["courses"].([]interface{})
		for _, raw := range courses {
			c, _ := raw.(map[string]interface{})
			if c["code"] == courseCode {
				courseID, _ = c["id"].(string)
				break
			}
		}
		require.NotEmpty(t, courseID, "could not find course in list")
		t.Logf("Found course ID: %s", courseID)
	})

	if courseID == "" {
		t.Fatal("course ID not resolved — cannot continue")
	}

	// ── Phase 2: Enroll TA and students ──────────────────────────────────

	t.Run("Phase2_EnrollMembers", func(t *testing.T) {
		// Enroll head TA
		_, err := hc.EnrollCourseMember(ctx, courseID, map[string]interface{}{
			"user_id":      "alex-ta",
			"email":        "alex@cs.university.edu",
			"display_name": "Alex (TA)",
			"role":         "ta",
		})
		require.NoError(t, err, "enroll TA")
		t.Log("✓ Enrolled TA: Alex")

		// Enroll students
		for i := 1; i <= numStudents; i++ {
			uid := fmt.Sprintf("student%d", i)
			_, err := hc.EnrollCourseMember(ctx, courseID, map[string]interface{}{
				"user_id":      uid,
				"email":        fmt.Sprintf("%s@uni.edu", uid),
				"display_name": fmt.Sprintf("Student %d", i),
				"role":         "student",
			})
			require.NoError(t, err, "enroll student %d", i)
		}
		t.Logf("✓ Enrolled %d students", numStudents)

		// Verify member count
		members, err := hc.ListCourseMembers(ctx, courseID, "")
		require.NoError(t, err)
		memberList, _ := members["members"].([]interface{})
		assert.GreaterOrEqual(t, len(memberList), numStudents+1, "unexpected member count")
		t.Logf("✓ Total members: %d", len(memberList))
	})

	// ── Phase 3: Whitelist template ───────────────────────────────────────

	t.Run("Phase3_WhitelistTemplate", func(t *testing.T) {
		err := hc.AddCourseTemplate(ctx, courseID, "python-ml")
		require.NoError(t, err, "AddCourseTemplate")
		t.Log("✓ Template 'python-ml' whitelisted")

		templates, err := hc.ListCourseTemplates(ctx, courseID)
		require.NoError(t, err)
		t.Logf("✓ Approved templates: %v", templates)
	})

	// ── Phase 4: Distribute per-student budget ────────────────────────────

	t.Run("Phase4_DistributeBudget", func(t *testing.T) {
		result, err := hc.DistributeCourseBudget(ctx, courseID, 24.0)
		require.NoError(t, err, "DistributeCourseBudget")
		t.Logf("✓ Budget distributed: %v", result)

		summary, err := hc.GetCourseBudget(ctx, courseID)
		require.NoError(t, err, "GetCourseBudget")
		t.Logf("✓ Budget summary: %v", summary)
	})

	// ── Phase 5: TA views course overview ────────────────────────────────

	t.Run("Phase5_TAOverview", func(t *testing.T) {
		overview, err := hc.GetCourseOverview(ctx, courseID)
		require.NoError(t, err, "GetCourseOverview")
		t.Logf("✓ Course overview: %v", overview)
	})

	// ── Phase 6: TA views audit log ──────────────────────────────────────

	t.Run("Phase6_AuditLog", func(t *testing.T) {
		audit, err := hc.GetCourseAuditLog(ctx, courseID, "", "", 0)
		require.NoError(t, err, "GetCourseAuditLog")
		entries, _ := audit["entries"].([]interface{})
		assert.Greater(t, len(entries), 0, "expected audit entries")
		t.Logf("✓ Audit entries: %d", len(entries))
	})

	// ── Phase 7: Usage report ────────────────────────────────────────────

	t.Run("Phase7_UsageReport", func(t *testing.T) {
		report, err := hc.GetCourseReport(ctx, courseID, "json")
		require.NoError(t, err, "GetCourseReport")
		assert.NotNil(t, report)
		t.Logf("✓ Usage report: %v", keyList(report))
	})

	// ── Phase 8: Close and archive ───────────────────────────────────────

	t.Run("Phase8_CloseAndArchive", func(t *testing.T) {
		err := hc.CloseCourse(ctx, courseID)
		require.NoError(t, err, "CloseCourse")
		t.Log("✓ Course closed")

		archiveResult, err := hc.ArchiveCourse(ctx, courseID)
		require.NoError(t, err, "ArchiveCourse")
		t.Logf("✓ Course archived: %v", archiveResult)
	})
}
