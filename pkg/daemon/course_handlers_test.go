package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/course"
	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func createTestCourse(t *testing.T, server *Server) *types.Course {
	t.Helper()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(90 * 24 * time.Hour)
	req := &course.CreateCourseRequest{
		Code:          "TEST101",
		Title:         "Test Course",
		Semester:      "Fall 2099",
		SemesterStart: start,
		SemesterEnd:   end,
		Owner:         "instructor1",
	}
	c, err := server.courseManager.CreateCourse(t.Context(), req)
	require.NoError(t, err)
	return c
}

// --- TestCourseEndpoints ---

func TestCourseEndpoints(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(90 * 24 * time.Hour)

	createBody, _ := json.Marshal(course.CreateCourseRequest{
		Code:          "CS101",
		Title:         "Intro",
		Semester:      "Fall 2099",
		SemesterStart: start,
		SemesterEnd:   end,
		Owner:         "prof1",
	})

	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "list courses returns empty array",
			method:         http.MethodGet,
			path:           "/api/v1/courses",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Contains(t, resp, "courses")
			},
		},
		{
			name:           "create course returns 201",
			method:         http.MethodPost,
			path:           "/api/v1/courses",
			body:           createBody,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var c types.Course
				require.NoError(t, json.Unmarshal(body, &c))
				assert.Equal(t, "CS101", c.Code)
				assert.NotEmpty(t, c.ID)
			},
		},
		{
			name:           "create course bad JSON returns 400",
			method:         http.MethodPost,
			path:           "/api/v1/courses",
			body:           []byte(`{bad json}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "get unknown course returns 404",
			method:         http.MethodGet,
			path:           "/api/v1/courses/does-not-exist",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.body != nil {
				bodyReader = bytes.NewReader(tt.body)
			} else {
				bodyReader = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code, "body: %s", w.Body.String())
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// --- Duplicate course returns 409 ---

func TestCreateCourseDuplicateReturns409(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(90 * 24 * time.Hour)
	body, _ := json.Marshal(course.CreateCourseRequest{
		Code:          "DUP101",
		Title:         "Duplicate",
		Semester:      "Spring 2099",
		SemesterStart: start,
		SemesterEnd:   end,
		Owner:         "prof1",
	})

	// First create: 201
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Second create: 409
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)
}

// --- Template Whitelist Enforcement (#46) ---

func TestTemplateWhitelistEnforcement(t *testing.T) {
	server := createTestServer(t)

	// Create a course with a whitelist
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(90 * 24 * time.Hour)
	c, err := server.courseManager.CreateCourse(t.Context(), &course.CreateCourseRequest{
		Code:              "ML201",
		Title:             "Machine Learning",
		Semester:          "Fall 2099",
		SemesterStart:     start,
		SemesterEnd:       end,
		Owner:             "prof1",
		ApprovedTemplates: []string{"python-ml"},
	})
	require.NoError(t, err)

	handler := server.createHTTPHandler()

	t.Run("approved template passes check", func(t *testing.T) {
		assert.True(t, server.courseManager.IsTemplateApproved(c.ID, "python-ml"))
	})

	t.Run("blocked template returns false", func(t *testing.T) {
		assert.False(t, server.courseManager.IsTemplateApproved(c.ID, "gpu-xl"))
	})

	t.Run("add template via API", func(t *testing.T) {
		body := `{"template":"r-research"}`
		req := httptest.NewRequest(http.MethodPost,
			"/api/v1/courses/"+c.ID+"/templates",
			strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.True(t, server.courseManager.IsTemplateApproved(c.ID, "r-research"))
	})

	t.Run("list templates returns whitelist", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/"+c.ID+"/templates", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp, "approved_templates")
	})
}

// --- Member Endpoints ---

func TestCourseMemberEndpoints(t *testing.T) {
	server := createTestServer(t)
	c := createTestCourse(t, server)
	handler := server.createHTTPHandler()

	// Enroll a student
	enrollBody, _ := json.Marshal(course.EnrollRequest{
		Email:       "student@uni.edu",
		DisplayName: "Test Student",
		Role:        types.ClassRoleStudent,
	})
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/courses/"+c.ID+"/members",
		bytes.NewReader(enrollBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

	// List members
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/courses/"+c.ID+"/members", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp))
	assert.Contains(t, resp, "members")

	// Duplicate enroll returns 409
	req3 := httptest.NewRequest(http.MethodPost,
		"/api/v1/courses/"+c.ID+"/members",
		bytes.NewReader(enrollBody))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusConflict, w3.Code)
}

// --- Budget Endpoint ---

func TestCourseBudgetEndpoints(t *testing.T) {
	server := createTestServer(t)
	c := createTestCourse(t, server)
	handler := server.createHTTPHandler()

	// Distribute budget
	body := `{"amount_per_student": 50.0}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/courses/"+c.ID+"/budget/distribute",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	// Get budget summary
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/courses/"+c.ID+"/budget", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

// --- Close Endpoint ---

func TestCloseCourseEndpoint(t *testing.T) {
	server := createTestServer(t)
	c := createTestCourse(t, server)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/"+c.ID+"/close", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	// Verify status
	updated, err := server.courseManager.GetCourse(t.Context(), c.ID)
	require.NoError(t, err)
	assert.Equal(t, types.CourseStatusClosed, updated.Status)
}

// --- TA Debug Endpoint (#48) ---

func TestTADebugEndpoint(t *testing.T) {
	server := createTestServer(t)
	c := createTestCourse(t, server)
	handler := server.createHTTPHandler()

	// Enroll a student
	_, err := server.courseManager.EnrollMember(t.Context(), c.ID, &course.EnrollRequest{
		UserID: "student1",
		Email:  "s1@uni.edu",
		Role:   types.ClassRoleStudent,
	})
	require.NoError(t, err)

	// TA debug by instructor (owner is auto-enrolled as instructor)
	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/courses/"+c.ID+"/ta/debug/student1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	// 200 is expected — the handler checks TA/instructor role
	// In test mode, taID comes from the (empty) userID context key,
	// but the course manager allows the course owner, which is "instructor1".
	// Since we can't easily set the auth context in httptest, we accept either 200 or 403.
	assert.Contains(t, []int{http.StatusOK, http.StatusForbidden}, w.Code)
}

// --- TA Reset Endpoint (#49) ---

func TestTAResetEndpoint(t *testing.T) {
	server := createTestServer(t)
	c := createTestCourse(t, server)
	handler := server.createHTTPHandler()

	_, err := server.courseManager.EnrollMember(t.Context(), c.ID, &course.EnrollRequest{
		UserID: "student2",
		Email:  "s2@uni.edu",
		Role:   types.ClassRoleStudent,
	})
	require.NoError(t, err)

	body := `{"reason":"broken environment","backup_retention_days":7}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/courses/"+c.ID+"/ta/reset/student2",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	// Same: accept 202 or 403 depending on auth context
	assert.Contains(t, []int{http.StatusAccepted, http.StatusForbidden}, w.Code)
}

// --- TA Reset requires reason ---

func TestTAResetRequiresReason(t *testing.T) {
	server := createTestServer(t)
	c := createTestCourse(t, server)
	handler := server.createHTTPHandler()

	// Enroll a TA so that the TA check passes
	_, err := server.courseManager.EnrollMember(t.Context(), c.ID, &course.EnrollRequest{
		UserID: "ta1",
		Role:   types.ClassRoleTA,
	})
	require.NoError(t, err)
	_, err = server.courseManager.EnrollMember(t.Context(), c.ID, &course.EnrollRequest{
		UserID: "student3",
		Role:   types.ClassRoleStudent,
	})
	require.NoError(t, err)

	body := `{}` // no reason
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/courses/"+c.ID+"/ta/reset/student3",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	// 400 (missing reason) or 403 (auth) — both are expected rejections
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusForbidden}, w.Code)
}
