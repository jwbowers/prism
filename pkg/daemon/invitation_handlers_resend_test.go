package daemon

// Tests for v0.25.0 Gap L: handleResendInvitation email dispatch.
//
// These tests verify:
//   - A project admin can resend an invitation (200)
//   - A non-member cannot resend an invitation (403)
//   - Email is dispatched (fire-and-forget via LogEmailSender — no panic, no error returned)

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resendInvitationAs posts a resend request for the given invitation ID with an optional requester.
func resendInvitationAs(t *testing.T, handler http.Handler, invitationID, requesterID string) *httptest.ResponseRecorder {
	t.Helper()
	url := "/api/v1/invitations/" + invitationID + "/resend"
	if requesterID != "" {
		url += "?requester_id=" + requesterID
	}
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// TestHandleResendInvitation_AdminSucceeds verifies that a project admin can
// resend an invitation and receives 200.
func TestHandleResendInvitation_AdminSucceeds(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Resend Admin Project")
	addProjectMember(t, handler, projectID, "admin-user", "admin")

	// Create an invitation as admin.
	w := sendInvitationAs(t, handler, projectID, "admin-user", "target@example.com")
	require.Equal(t, http.StatusCreated, w.Code, "setup: %s", w.Body.String())

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	invMap := createResp["invitation"].(map[string]interface{})
	invID := invMap["id"].(string)

	// Resend as admin — should succeed (200) and not panic on email dispatch.
	wr := resendInvitationAs(t, handler, invID, "admin-user")
	assert.Equal(t, http.StatusOK, wr.Code,
		"admin should be able to resend: %s", wr.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(wr.Body.Bytes(), &resp))
	assert.Equal(t, "Invitation resent successfully", resp["message"])
}

// TestHandleResendInvitation_NonMemberForbidden verifies that a non-member
// cannot resend an invitation (403).
func TestHandleResendInvitation_NonMemberForbidden(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Resend Non-Member Project")
	addProjectMember(t, handler, projectID, "admin-user", "admin")

	// Create an invitation as admin.
	w := sendInvitationAs(t, handler, projectID, "admin-user", "target2@example.com")
	require.Equal(t, http.StatusCreated, w.Code, "setup: %s", w.Body.String())

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	invMap := createResp["invitation"].(map[string]interface{})
	invID := invMap["id"].(string)

	// Attempt resend as a non-member stranger.
	wr := resendInvitationAs(t, handler, invID, "stranger@example.com")
	assert.Equal(t, http.StatusForbidden, wr.Code,
		"non-member should be forbidden from resending: %s", wr.Body.String())
}

// TestHandleResendInvitation_NoEmailPanic verifies that email dispatch does
// not panic or return an error to the caller when the project lookup succeeds.
// (The LogEmailSender used in tests writes to stdout and always returns nil.)
func TestHandleResendInvitation_NoEmailPanic(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Resend Email No-Panic Project")
	addProjectMember(t, handler, projectID, "owner-user", "owner")

	w := sendInvitationAs(t, handler, projectID, "owner-user", "nopanic@example.com")
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	invMap := createResp["invitation"].(map[string]interface{})
	invID := invMap["id"].(string)

	// Resend as owner — email dispatch must not panic.
	assert.NotPanics(t, func() {
		wr := resendInvitationAs(t, handler, invID, "owner-user")
		assert.Equal(t, http.StatusOK, wr.Code)
	})
}
