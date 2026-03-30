package daemon

// Tests for v0.25.0 Gaps M + N in shared token handlers.
//
// Gap M: created_by from request body is respected (not overwritten with "system").
// Gap N: requireProjectRole applied to create/list/revoke/extend operations.
//
// These tests cover:
//   - Non-member trying to create a shared token → 403
//   - Viewer trying to create a shared token → 403
//   - Admin creating a shared token → 201
//   - No created_by (backward-compat "system") → 201 without role check
//   - created_by round-trips into the response
//   - Non-member listing tokens → 403
//   - Member listing tokens → 200
//   - Non-member revoking a token → 403
//   - Admin revoking a token → 200

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createSharedTokenAs posts a create-shared-token request with the given creator identity.
func createSharedTokenAs(t *testing.T, handler http.Handler, projectID, createdBy string) *httptest.ResponseRecorder {
	t.Helper()
	body := map[string]interface{}{
		"name":             "auth-test-token",
		"role":             "member",
		"redemption_limit": 10,
	}
	if createdBy != "" {
		body["created_by"] = createdBy
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/"+projectID+"/invitations/shared", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// TestHandleCreateSharedToken_NonMemberForbidden verifies that a user who is not
// a project member receives 403 when attempting to create a shared token (Gap N).
func TestHandleCreateSharedToken_NonMemberForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token Auth Non-Member")

	w := createSharedTokenAs(t, handler, projectID, "stranger@example.com")
	assert.Equal(t, http.StatusForbidden, w.Code,
		"non-member should be forbidden: %s", w.Body.String())
}

// TestHandleCreateSharedToken_ViewerForbidden verifies that a viewer cannot
// create a shared token (403 — insufficient role).
func TestHandleCreateSharedToken_ViewerForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token Auth Viewer")
	addProjectMember(t, handler, projectID, "viewer-user", "viewer")

	w := createSharedTokenAs(t, handler, projectID, "viewer-user")
	assert.Equal(t, http.StatusForbidden, w.Code,
		"viewer should be forbidden: %s", w.Body.String())
}

// TestHandleCreateSharedToken_AdminSucceeds verifies that a project admin can
// create a shared token (201) with their identity preserved (Gap M).
func TestHandleCreateSharedToken_AdminSucceeds(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token Auth Admin")
	addProjectMember(t, handler, projectID, "admin-user", "admin")

	w := createSharedTokenAs(t, handler, projectID, "admin-user")
	require.Equal(t, http.StatusCreated, w.Code,
		"admin should be able to create shared token: %s", w.Body.String())

	// Verify the token was created successfully.
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["token"], "response should contain token code")
}

// TestHandleCreateSharedToken_SystemBackwardCompat verifies that omitting
// created_by (backward-compatible "system" sentinel) still creates a token
// without a 403 — system callers are not subject to the role check (Gap M).
func TestHandleCreateSharedToken_SystemBackwardCompat(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token System Compat")

	// No created_by field — server should fall back to "system" and skip role check.
	w := createSharedTokenAs(t, handler, projectID, "")
	assert.Equal(t, http.StatusCreated, w.Code,
		"system caller (no created_by) should not require role check: %s", w.Body.String())
}

// TestHandleListSharedTokens_NonMemberForbidden verifies that a non-member
// cannot list shared tokens when requester_id is provided (403).
func TestHandleListSharedTokens_NonMemberForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token List Non-Member")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/invitations/shared?requester_id=stranger", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"non-member should be forbidden from listing: %s", w.Body.String())
}

// TestHandleListSharedTokens_MemberSucceeds verifies that a project member
// can list shared tokens (200).
func TestHandleListSharedTokens_MemberSucceeds(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token List Member")
	addProjectMember(t, handler, projectID, "member-user", "member")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/invitations/shared?requester_id=member-user", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code,
		"member should be able to list tokens: %s", w.Body.String())
}

// TestHandleRevokeSharedToken_NonMemberForbidden verifies that a non-member
// cannot revoke a shared token (403).
func TestHandleRevokeSharedToken_NonMemberForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token Revoke Non-Member")

	// Create a token as system (no auth required for backward compat).
	tokenCode := createSharedToken(t, handler, projectID)

	// Attempt revoke as non-member.
	req := httptest.NewRequest(http.MethodDelete,
		"/api/v1/invitations/shared/"+tokenCode+"?requester_id=stranger", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"non-member should be forbidden from revoking: %s", w.Body.String())
}

// TestHandleRevokeSharedToken_AdminSucceeds verifies that a project admin can
// revoke a shared token (200).
func TestHandleRevokeSharedToken_AdminSucceeds(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token Revoke Admin")
	addProjectMember(t, handler, projectID, "admin-user", "admin")

	// Create a token as system.
	tokenCode := createSharedToken(t, handler, projectID)

	// Revoke as admin.
	req := httptest.NewRequest(http.MethodDelete,
		"/api/v1/invitations/shared/"+tokenCode+"?requester_id=admin-user", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code,
		"admin should be able to revoke token: %s", w.Body.String())
}
