package daemon

// Tests for v0.24.0 Gaps I, J, K in handleBulkInvitation:
//   - Gap I: Non-member attempting bulk invite → 403
//   - Gap I: Admin can bulk invite → 201
//   - Gap K: Email send is triggered for each "sent" result (fire-and-forget, no panic)

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sendBulkInvitationAs posts a bulk invitation request with the given inviter.
func sendBulkInvitationAs(t *testing.T, handler http.Handler, projectID, inviter string, emails ...string) *httptest.ResponseRecorder {
	t.Helper()
	entries := make([]map[string]interface{}, 0, len(emails))
	for _, e := range emails {
		entries = append(entries, map[string]interface{}{"email": e})
	}
	body := map[string]interface{}{
		"invited_by":   inviter,
		"default_role": "member",
		"invitations":  entries,
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/"+projectID+"/invitations/bulk", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// TestHandleBulkInvitation_NonMemberForbidden verifies that a user who is not a
// project member receives 403 when attempting a bulk invitation (Gap I).
func TestHandleBulkInvitation_NonMemberForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Bulk Auth Project Non-Member")

	w := sendBulkInvitationAs(t, handler, projectID, "stranger@example.com",
		"invite1@example.com", "invite2@example.com")

	assert.Equal(t, http.StatusForbidden, w.Code,
		"non-member should be forbidden from bulk invitations: %s", w.Body.String())
}

// TestHandleBulkInvitation_ViewerForbidden verifies that a project viewer
// cannot send bulk invitations (403 — insufficient role).
func TestHandleBulkInvitation_ViewerForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Bulk Auth Project Viewer")
	addProjectMember(t, handler, projectID, "viewer-user", "viewer")

	w := sendBulkInvitationAs(t, handler, projectID, "viewer-user",
		"invite1@example.com")

	assert.Equal(t, http.StatusForbidden, w.Code,
		"viewer should be forbidden from bulk invitations: %s", w.Body.String())
}

// TestHandleBulkInvitation_AdminSucceeds verifies that a project admin can
// send bulk invitations and receives 201 (Gap I — admin path).
func TestHandleBulkInvitation_AdminSucceeds(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Bulk Auth Project Admin")
	addProjectMember(t, handler, projectID, "admin-user", "admin")

	w := sendBulkInvitationAs(t, handler, projectID, "admin-user",
		"bulk1@example.com", "bulk2@example.com")

	require.Equal(t, http.StatusCreated, w.Code,
		"admin should be able to send bulk invitations: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	summary, ok := resp["summary"].(map[string]interface{})
	require.True(t, ok, "response should contain summary")
	assert.Equal(t, float64(2), summary["total"], "total should match number of emails")
}

// TestHandleBulkInvitation_SystemBackwardCompat verifies that omitting
// invited_by (backward-compatible "system" sentinel) still creates invitations
// without a 403 — system callers are not subject to the role check (Gap J).
func TestHandleBulkInvitation_SystemBackwardCompat(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Bulk System Compat Project")

	body := map[string]interface{}{
		"default_role": "member",
		"invitations": []map[string]interface{}{
			{"email": "system-invite@example.com"},
		},
		// No invited_by field — server should default to "system"
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/"+projectID+"/invitations/bulk", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code,
		"system caller (no invited_by) should not require role check: %s", w.Body.String())
}

// TestHandleBulkInvitation_EmailsSentForSentResults verifies that email sending
// is attempted for each "sent" result (Gap K). The test server uses the log
// email sender, which does not error, so no panics should occur and the
// response should still be 201.
func TestHandleBulkInvitation_EmailsSentForSentResults(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Bulk Email Project")
	addProjectMember(t, handler, projectID, "owner-user", "owner")

	w := sendBulkInvitationAs(t, handler, projectID, "owner-user",
		"email-a@example.com", "email-b@example.com", "email-c@example.com")

	require.Equal(t, http.StatusCreated, w.Code,
		"bulk invite with email sending should not fail: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	summary, ok := resp["summary"].(map[string]interface{})
	require.True(t, ok)
	// All 3 invitations should have been sent (no duplicates, all new emails).
	assert.Equal(t, float64(3), summary["sent"],
		"all 3 invitations should be sent and email dispatch should not block")
}
