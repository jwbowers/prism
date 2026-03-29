package daemon

// Tests for the v0.23.0 Gap B fix: redeeming a shared token adds the redeemer
// as a project member.
//
// These tests cover:
//   - Successful redemption adds redeemer to project.Members
//   - Redeeming twice by the same user (duplicate) is accepted gracefully

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createSharedToken creates a shared token for the given project and returns the
// token code string.
func createSharedToken(t *testing.T, handler http.Handler, projectID string) string {
	t.Helper()
	body := map[string]interface{}{
		"name":             "test-token",
		"role":             "member",
		"redemption_limit": 10,
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/"+projectID+"/invitations/shared", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code,
		"failed to create shared token: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	code, ok := resp["token"].(string)
	require.True(t, ok, "response missing token field")
	require.NotEmpty(t, code)
	return code
}

// redeemSharedToken calls POST /api/v1/invitations/shared/redeem and returns
// the recorder so callers can inspect status and body.
func redeemSharedToken(t *testing.T, handler http.Handler, tokenCode, redeemedBy string) *httptest.ResponseRecorder {
	t.Helper()
	body := map[string]interface{}{
		"token":       tokenCode,
		"redeemed_by": redeemedBy,
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/invitations/shared/redeem", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// getProjectMembers calls GET /api/v1/projects/{id}/members and returns the
// parsed member list.  The endpoint returns a JSON array directly.
func getProjectMembers(t *testing.T, handler http.Handler, projectID string) []map[string]interface{} {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/members", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code,
		"getProjectMembers failed: %s", w.Body.String())

	var rawSlice []interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rawSlice))

	members := make([]map[string]interface{}, 0, len(rawSlice))
	for _, m := range rawSlice {
		if mm, ok := m.(map[string]interface{}); ok {
			members = append(members, mm)
		}
	}
	return members
}

// hasProjectMember returns true if userID is present in members.
func hasProjectMember(members []map[string]interface{}, userID string) bool {
	for _, m := range members {
		if uid, _ := m["user_id"].(string); uid == userID {
			return true
		}
	}
	return false
}

// TestRedeemSharedToken_AddsProjectMember verifies that after a successful
// redemption the redeemer appears in the project member list.
func TestRedeemSharedToken_AddsProjectMember(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token Member Test")
	tokenCode := createSharedToken(t, handler, projectID)

	w := redeemSharedToken(t, handler, tokenCode, "alice@example.com")
	assert.Equal(t, http.StatusOK, w.Code,
		"redemption should succeed: %s", w.Body.String())

	// The redeemer should now be in the project member list.
	members := getProjectMembers(t, handler, projectID)
	assert.True(t, hasProjectMember(members, "alice@example.com"),
		"alice should be a project member after redemption; members: %v", members)
}

// TestRedeemSharedToken_DuplicateIsGraceful verifies that redeeming the same
// token twice by the same user does not return an error on the first call, even
// though the second add is a duplicate.
func TestRedeemSharedToken_DuplicateIsGraceful(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Shared Token Duplicate Test")
	tokenCode := createSharedToken(t, handler, projectID)

	// First redemption.
	w1 := redeemSharedToken(t, handler, tokenCode, "bob@example.com")
	assert.Equal(t, http.StatusOK, w1.Code,
		"first redemption should succeed: %s", w1.Body.String())

	// Second attempt by the same user — the token manager returns ErrAlreadyRedeemed,
	// so the handler returns 400, but the project membership side-effect is idempotent.
	w2 := redeemSharedToken(t, handler, tokenCode, "bob@example.com")
	// 200 or 400 are both acceptable here; the important thing is no panic.
	assert.NotEqual(t, http.StatusInternalServerError, w2.Code,
		"duplicate redemption must not cause 500: %s", w2.Body.String())
}
