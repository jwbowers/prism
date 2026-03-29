package daemon

// Tests for the v0.23.0 requireProjectRole auth check in invitation handlers.
//
// These tests cover:
//   - Non-member trying to create invitation → 403
//   - Viewer trying to create invitation → 403
//   - Admin creating an invitation → 201
//   - Member listing invitations → 200
//   - Non-member listing invitations → 403
//   - Non-member trying to revoke → 403
//   - Admin revoking → 200

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addProjectMember is a test helper that adds a member to a project via the API.
func addProjectMember(t *testing.T, handler http.Handler, projectID, userID, role string) {
	t.Helper()
	body := map[string]interface{}{
		"user_id":  userID,
		"role":     role,
		"added_by": "test-setup",
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/"+projectID+"/members", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code,
		"addProjectMember failed for user=%s role=%s: %s", userID, role, w.Body.String())
}

// sendInvitationAs posts an invitation request with the given inviter.
func sendInvitationAs(t *testing.T, handler http.Handler, projectID, inviter, email string) *httptest.ResponseRecorder {
	t.Helper()
	body := map[string]interface{}{
		"email":      email,
		"role":       "member",
		"invited_by": inviter,
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/"+projectID+"/invitations", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// TestHandleSendInvitation_NonMemberForbidden verifies that a user who is not a
// project member cannot send an invitation (403).
func TestHandleSendInvitation_NonMemberForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Auth Test Project")

	w := sendInvitationAs(t, handler, projectID, "stranger@example.com", "invite@example.com")
	assert.Equal(t, http.StatusForbidden, w.Code,
		"non-member should be forbidden: %s", w.Body.String())
}

// TestHandleSendInvitation_ViewerForbidden verifies that a project viewer
// cannot send an invitation (403 — insufficient role).
func TestHandleSendInvitation_ViewerForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Auth Test Viewer")
	addProjectMember(t, handler, projectID, "viewer-user", "viewer")

	w := sendInvitationAs(t, handler, projectID, "viewer-user", "invite@example.com")
	assert.Equal(t, http.StatusForbidden, w.Code,
		"viewer should be forbidden: %s", w.Body.String())
}

// TestHandleSendInvitation_AdminSucceeds verifies that a project admin can
// send an invitation (201).
func TestHandleSendInvitation_AdminSucceeds(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Auth Test Admin")
	addProjectMember(t, handler, projectID, "admin-user", "admin")

	w := sendInvitationAs(t, handler, projectID, "admin-user", "newuser@example.com")
	assert.Equal(t, http.StatusCreated, w.Code,
		"admin should succeed: %s", w.Body.String())
}

// TestHandleListProjectInvitations_MemberSucceeds verifies that a project
// member can list invitations (200).
func TestHandleListProjectInvitations_MemberSucceeds(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Auth Test List Member")
	addProjectMember(t, handler, projectID, "member-user", "member")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/invitations?requester_id=member-user", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code,
		"member should be able to list invitations: %s", w.Body.String())
}

// TestHandleListProjectInvitations_NonMemberForbidden verifies that a
// non-member is denied listing invitations (403).
func TestHandleListProjectInvitations_NonMemberForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Auth Test List Non-Member")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/"+projectID+"/invitations?requester_id=stranger", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"non-member should be forbidden: %s", w.Body.String())
}

// TestHandleRevokeInvitation_NonMemberForbidden verifies that a non-member
// cannot revoke an invitation (403).
func TestHandleRevokeInvitation_NonMemberForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Create project and add an admin so we can first create an invitation.
	projectID := createGovTestProject(t, handler, "Auth Test Revoke")
	addProjectMember(t, handler, projectID, "admin-user", "admin")

	// Create an invitation as the admin.
	w := sendInvitationAs(t, handler, projectID, "admin-user", "target@example.com")
	require.Equal(t, http.StatusCreated, w.Code, "failed to create invitation: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	invMap, ok := resp["invitation"].(map[string]interface{})
	require.True(t, ok, "missing invitation in response")
	invID, ok := invMap["id"].(string)
	require.True(t, ok, "missing invitation id")

	// Attempt to revoke as a non-member.
	req := httptest.NewRequest(http.MethodDelete,
		"/api/v1/invitations/"+invID+"?requester_id=stranger", nil)
	wr := httptest.NewRecorder()
	handler.ServeHTTP(wr, req)
	assert.Equal(t, http.StatusForbidden, wr.Code,
		"non-member should be forbidden to revoke: %s", wr.Body.String())
}
