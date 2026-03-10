package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createGovTestProject is a helper that creates a test project and returns its ID.
// It fails the test immediately if project creation fails.
func createGovTestProject(t *testing.T, handler http.Handler, name string) string {
	t.Helper()
	body := map[string]interface{}{
		"name":        name,
		"description": "governance test project",
		"owner":       "test-owner",
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "expected project creation to succeed: %s", w.Body.String())

	var proj map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &proj))
	id, ok := proj["id"].(string)
	require.True(t, ok, "project response missing id field")
	require.NotEmpty(t, id)
	return id
}

// ---------------------------------------------------------------------------
// Quota handler tests
// ---------------------------------------------------------------------------

// TestHandleProjectQuotas_Get tests GET /api/v1/projects/{id}/quotas returns 200
// with a role_quotas key (may be empty/null for projects with no budget).
func TestHandleProjectQuotas_Get(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Quota Get Test")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/quotas", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "role_quotas")
}

// TestHandleProjectQuotas_Put tests PUT /api/v1/projects/{id}/quotas sets a quota
// and the response contains the updated role_quotas list.
func TestHandleProjectQuotas_Put(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Quota Put Test")

	quota := map[string]interface{}{
		"role":            "member",
		"max_instances":   5,
		"max_spend_daily": 10.0,
	}
	data, err := json.Marshal(quota)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+projectID+"/quotas", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "PUT quotas should succeed: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "role_quotas")

	roleQuotas, ok := resp["role_quotas"].([]interface{})
	require.True(t, ok, "role_quotas should be a list")
	assert.Len(t, roleQuotas, 1, "should have exactly one quota after PUT")
}

// TestHandleProjectQuotas_Put_BadJSON tests that an invalid JSON body returns 400.
func TestHandleProjectQuotas_Put_BadJSON(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Quota BadJSON Test")

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+projectID+"/quotas", bytes.NewReader([]byte(`{bad json}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Grant period handler tests
// ---------------------------------------------------------------------------

// TestHandleGrantPeriod_PutGet tests a PUT-then-GET round-trip for grant periods.
func TestHandleGrantPeriod_PutGet(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "GrantPeriod PutGet Test")

	gp := map[string]interface{}{
		"name":        "NSF Year 1",
		"start_date":  time.Now().Format(time.RFC3339),
		"end_date":    time.Now().AddDate(1, 0, 0).Format(time.RFC3339),
		"auto_freeze": true,
	}
	data, err := json.Marshal(gp)
	require.NoError(t, err)

	// PUT
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+projectID+"/grant-period", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "PUT grant-period should succeed: %s", w.Body.String())
	var putResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &putResp))
	assert.Contains(t, putResp, "grant_period")
	assert.Contains(t, putResp, "message")

	// GET
	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/grant-period", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "GET grant-period should succeed: %s", w.Body.String())
	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getResp))
	assert.Contains(t, getResp, "grant_period")

	gpData, ok := getResp["grant_period"].(map[string]interface{})
	require.True(t, ok, "grant_period field should be an object")
	assert.Equal(t, "NSF Year 1", gpData["name"])
}

// TestHandleGrantPeriod_Delete tests that DELETE removes the grant period and a
// subsequent GET returns grant_period: null (not an error).
func TestHandleGrantPeriod_Delete(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "GrantPeriod Delete Test")

	// Set a grant period first
	gp := map[string]interface{}{
		"name":       "Temp Grant",
		"start_date": time.Now().Format(time.RFC3339),
		"end_date":   time.Now().AddDate(0, 6, 0).Format(time.RFC3339),
	}
	data, _ := json.Marshal(gp)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+projectID+"/grant-period", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// DELETE
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+projectID+"/grant-period", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "DELETE grant-period should succeed: %s", w.Body.String())
	var delResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &delResp))
	assert.Contains(t, delResp, "message")

	// GET after DELETE should return 200 with grant_period: null
	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/grant-period", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getResp))
	assert.Contains(t, getResp, "grant_period")
	// grant_period should be null after deletion
	assert.Nil(t, getResp["grant_period"])
}

// ---------------------------------------------------------------------------
// Approval handler tests
// ---------------------------------------------------------------------------

// TestHandleProjectApprovals_Submit tests POST /api/v1/projects/{id}/approvals.
// If the approval manager is initialised it returns 202; if not, 503.
// Both are valid outcomes in a test environment.
func TestHandleProjectApprovals_Submit(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Approvals Submit Test")

	body := map[string]interface{}{
		"type":    "gpu_instance",
		"reason":  "need GPU for training",
		"details": map[string]interface{}{"instance_type": "p3.2xlarge"},
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/approvals", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 202 Accepted when manager is present; 503 if not initialised in test mode
	assert.True(t,
		w.Code == http.StatusAccepted || w.Code == http.StatusServiceUnavailable,
		"expected 202 or 503, got %d: %s", w.Code, w.Body.String(),
	)

	if w.Code == http.StatusAccepted {
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp, "id")
		assert.Equal(t, "gpu_instance", resp["type"])
		assert.Equal(t, "pending", resp["status"])
	}
}

// TestHandleProjectApprovals_List tests GET /api/v1/projects/{id}/approvals returns 200
// with an approvals array when the manager is available, or 503 when not.
func TestHandleProjectApprovals_List(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Approvals List Test")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/approvals", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t,
		w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable,
		"expected 200 or 503, got %d: %s", w.Code, w.Body.String(),
	)

	if w.Code == http.StatusOK {
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp, "approvals")
		assert.Contains(t, resp, "total")
	}
}

// TestHandleProjectApprovals_SubmitThenList tests the full submit→list round-trip
// when the approval manager is initialised.
func TestHandleProjectApprovals_SubmitThenList(t *testing.T) {
	server := createTestServer(t)
	if server.approvalManager == nil {
		t.Skip("approval manager not initialised in this test environment")
	}
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Approvals SubmitList Test")

	// Submit a request
	submitBody := map[string]interface{}{
		"type":    "sub_budget",
		"reason":  "delegate to grad student",
		"details": map[string]interface{}{},
	}
	data, _ := json.Marshal(submitBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/approvals", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code)

	var submitResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &submitResp))
	approvalID, ok := submitResp["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, approvalID)

	// List approvals for the project
	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/approvals", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var listResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	approvals, ok := listResp["approvals"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(approvals), 1)
	assert.Equal(t, float64(len(approvals)), listResp["total"])
}

// TestHandleProjectApprovals_ApproveDeny tests the approve and deny actions
// when the approval manager is initialised.
func TestHandleProjectApprovals_ApproveDeny(t *testing.T) {
	server := createTestServer(t)
	if server.approvalManager == nil {
		t.Skip("approval manager not initialised in this test environment")
	}
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Approvals ApproveDeny Test")

	// Helper: submit an approval request and return its ID
	submitApproval := func(typ string) string {
		body := map[string]interface{}{
			"type":    typ,
			"reason":  "test",
			"details": map[string]interface{}{},
		}
		data, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/approvals", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		require.Equal(t, http.StatusAccepted, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		return resp["id"].(string)
	}

	// Test approve action
	t.Run("approve", func(t *testing.T) {
		approvalID := submitApproval("expensive_instance")

		approveBody := map[string]interface{}{"note": "looks good"}
		data, _ := json.Marshal(approveBody)
		req := httptest.NewRequest(http.MethodPost,
			"/api/v1/projects/"+projectID+"/approvals/"+approvalID+"/approve",
			bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "approve should return 200: %s", w.Body.String())
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "approved", resp["status"])
		assert.Equal(t, approvalID, resp["id"])
	})

	// Test deny action
	t.Run("deny", func(t *testing.T) {
		approvalID := submitApproval("budget_overage")

		denyBody := map[string]interface{}{"note": "budget exceeded policy"}
		data, _ := json.Marshal(denyBody)
		req := httptest.NewRequest(http.MethodPost,
			"/api/v1/projects/"+projectID+"/approvals/"+approvalID+"/deny",
			bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "deny should return 200: %s", w.Body.String())
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "denied", resp["status"])
		assert.Equal(t, approvalID, resp["id"])
	})
}

// ---------------------------------------------------------------------------
// Admin approvals handler tests
// ---------------------------------------------------------------------------

// TestHandleAdminApprovals tests GET /api/v1/admin/approvals always returns 200
// with an approvals array and total count (empty list when no approvals exist).
func TestHandleAdminApprovals(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/approvals", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "admin approvals should always return 200: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "approvals")
	assert.Contains(t, resp, "total")
}

// TestHandleAdminApprovals_StatusFilter tests GET /api/v1/admin/approvals?status=pending.
func TestHandleAdminApprovals_StatusFilter(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/approvals?status=pending", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "approvals")
}

// TestHandleAdminApprovals_MethodNotAllowed tests that non-GET methods return 405.
func TestHandleAdminApprovals_MethodNotAllowed(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/approvals", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// ---------------------------------------------------------------------------
// Budget share handler tests
// ---------------------------------------------------------------------------

// TestHandleProjectBudgetShare_Post tests POST /api/v1/projects/{id}/budget/share.
// A project with no budget set still returns 200 because ShareBudget skips the
// availability check when Budget is nil.
func TestHandleProjectBudgetShare_Post(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "BudgetShare Post Test")

	shareBody := map[string]interface{}{
		"amount": 100.0,
		"reason": "research allocation",
	}
	data, err := json.Marshal(shareBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/budget/share", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 200 with record (no budget set → availability check skipped)
	assert.Equal(t, http.StatusOK, w.Code, "budget share should succeed: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "id")
	assert.Contains(t, resp, "approved_by")
}

// TestHandleProjectBudgetShare_InsufficientBudget tests that sharing more than
// the available budget returns 400.
func TestHandleProjectBudgetShare_InsufficientBudget(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "BudgetShare Insufficient Test")

	// Set a budget of $50
	budgetBody := map[string]interface{}{
		"total_budget": 50.0,
	}
	budgetData, _ := json.Marshal(budgetBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+projectID+"/budget", bytes.NewReader(budgetData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "setting budget should succeed")

	// Attempt to share $200 (exceeds budget)
	shareBody := map[string]interface{}{
		"amount": 200.0,
		"reason": "exceeds budget",
	}
	data, _ := json.Marshal(shareBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/budget/share", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "should reject share exceeding budget: %s", w.Body.String())
}

// TestHandleProjectBudgetShare_ZeroAmount tests that sharing $0 is rejected.
func TestHandleProjectBudgetShare_ZeroAmount(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "BudgetShare ZeroAmount Test")

	shareBody := map[string]interface{}{
		"amount": 0.0,
	}
	data, _ := json.Marshal(shareBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/budget/share", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Budget shares list handler tests
// ---------------------------------------------------------------------------

// TestHandleProjectBudgetShares_Get tests GET /api/v1/projects/{id}/budget/shares
// returns 200 with an empty shares array (in-session-only persistence).
func TestHandleProjectBudgetShares_Get(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "BudgetShares Get Test")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/budget/shares", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "GET budget/shares should return 200: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "shares")
}

// ---------------------------------------------------------------------------
// Monthly report handler tests
// ---------------------------------------------------------------------------

// TestHandleProjectMonthlyReport_JSON tests GET ?format=json returns 200 JSON.
func TestHandleProjectMonthlyReport_JSON(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "MonthlyReport JSON Test")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/reports/monthly?month=2026-01&format=json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "JSON report should succeed: %s", w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	// Response should be valid JSON
	var resp interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
}

// TestHandleProjectMonthlyReport_Text tests GET ?format=text returns 200 plain text.
func TestHandleProjectMonthlyReport_Text(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "MonthlyReport Text Test")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/reports/monthly?month=2026-01&format=text", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "text report should succeed: %s", w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}

// TestHandleProjectMonthlyReport_CSV tests GET ?format=csv returns 200 CSV content.
func TestHandleProjectMonthlyReport_CSV(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "MonthlyReport CSV Test")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/reports/monthly?month=2026-01&format=csv", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "CSV report should succeed: %s", w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
}

// TestHandleProjectMonthlyReport_BadMonth tests that an invalid month returns 400.
func TestHandleProjectMonthlyReport_BadMonth(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "MonthlyReport BadMonth Test")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/reports/monthly?month=2026-13&format=json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "invalid month should return 400: %s", w.Body.String())
}

// TestHandleProjectMonthlyReport_NonExistentProject tests that a non-existent
// project ID returns 404.
func TestHandleProjectMonthlyReport_NonExistentProject(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/does-not-exist/reports/monthly?month=2026-01", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Onboarding template handler tests
// ---------------------------------------------------------------------------

// TestHandleOnboardingTemplates_AddList tests POST then GET round-trip for
// onboarding templates.
func TestHandleOnboardingTemplates_AddList(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "OnboardingTemplates AddList Test")

	// POST a new onboarding template
	tmpl := map[string]interface{}{
		"name":         "Welcome Kit",
		"budget_limit": 500.0,
		"templates":    []string{"python-ml"},
	}
	data, err := json.Marshal(tmpl)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/"+projectID+"/onboarding-templates",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "POST onboarding template should succeed: %s", w.Body.String())
	var postResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &postResp))
	assert.Contains(t, postResp, "message")

	// GET to verify the template was stored
	req = httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/onboarding-templates", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "GET onboarding templates should succeed: %s", w.Body.String())
	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getResp))
	assert.Contains(t, getResp, "onboarding_templates")

	templates, ok := getResp["onboarding_templates"].([]interface{})
	require.True(t, ok, "onboarding_templates should be a list")
	assert.Len(t, templates, 1, "should have exactly one onboarding template")

	first, ok := templates[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Welcome Kit", first["name"])
}

// TestHandleOnboardingTemplates_Delete tests POST then DELETE then GET returns empty list.
func TestHandleOnboardingTemplates_Delete(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "OnboardingTemplates Delete Test")

	// POST a template (no spaces in name to keep URL simple)
	tmpl := map[string]interface{}{
		"name":         "TempKit",
		"budget_limit": 100.0,
		"templates":    []string{},
	}
	data, _ := json.Marshal(tmpl)
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/"+projectID+"/onboarding-templates",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// DELETE by name
	req = httptest.NewRequest(http.MethodDelete,
		"/api/v1/projects/"+projectID+"/onboarding-templates/TempKit", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "DELETE onboarding template should succeed: %s", w.Body.String())
	var delResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &delResp))
	assert.Contains(t, delResp, "message")

	// GET: list should now be empty
	req = httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/onboarding-templates", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getResp))

	// onboarding_templates may be null (JSON null) or an empty array after deletion
	if templates := getResp["onboarding_templates"]; templates != nil {
		list, ok := templates.([]interface{})
		if ok {
			assert.Empty(t, list, "onboarding_templates should be empty after deletion")
		}
	}
}

// TestHandleOnboardingTemplates_DeleteMissingID tests that DELETE without a
// template name/ID returns 400.
func TestHandleOnboardingTemplates_DeleteMissingID(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "OnboardingTemplates DeleteMissingID Test")

	// DELETE without name segment — path has only 2 parts so handler returns 400
	req := httptest.NewRequest(http.MethodDelete,
		"/api/v1/projects/"+projectID+"/onboarding-templates", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
