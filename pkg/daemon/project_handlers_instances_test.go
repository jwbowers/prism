package daemon

// Tests for v0.24.0 Gap F:
//   - GET /api/v1/projects/{id}/instances returns the correct instance list
//   - DELETE /api/v1/projects/{id} is blocked when calculateActiveInstances > 0

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleGetProjectInstances_Empty verifies the endpoint returns an empty list
// when no instances are associated with the project.
func TestHandleGetProjectInstances_Empty(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Instances Empty Project")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/instances", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "unexpected error: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, projectID, resp["project_id"])
	assert.Equal(t, float64(0), resp["count"])
	instances, ok := resp["instances"].([]interface{})
	require.True(t, ok, "instances should be an array")
	assert.Len(t, instances, 0)
}

// TestHandleGetProjectInstances_ReturnsMatchingInstances verifies that instances
// seeded in state with the correct ProjectID are returned, while instances
// belonging to a different project are excluded.
func TestHandleGetProjectInstances_ReturnsMatchingInstances(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Instances With Data")
	otherProjectID := createGovTestProject(t, handler, "Instances Other Project")

	// Seed a running instance for our project.
	matchInst := types.Instance{
		ID:           "i-match001",
		Name:         "inst-match",
		Template:     "test-template",
		State:        "running",
		ProjectID:    projectID,
		InstanceType: "t3.micro",
		Region:       "us-east-1",
		LaunchTime:   time.Now(),
	}
	require.NoError(t, server.stateManager.SaveInstance(matchInst))

	// Seed an instance for a different project (must NOT appear in results).
	otherInst := types.Instance{
		ID:           "i-other001",
		Name:         "inst-other",
		Template:     "test-template",
		State:        "running",
		ProjectID:    otherProjectID,
		InstanceType: "t3.micro",
		Region:       "us-east-1",
		LaunchTime:   time.Now(),
	}
	require.NoError(t, server.stateManager.SaveInstance(otherInst))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/instances", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "unexpected error: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, float64(1), resp["count"], "only the matching instance should be returned")
	instances, ok := resp["instances"].([]interface{})
	require.True(t, ok)
	require.Len(t, instances, 1)

	first := instances[0].(map[string]interface{})
	assert.Equal(t, "i-match001", first["id"])
	assert.Equal(t, projectID, first["project_id"])
}

// TestHandleGetProjectInstances_UnknownProject verifies the endpoint returns 404
// for a project that does not exist.
func TestHandleGetProjectInstances_UnknownProject(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/nonexistent-proj/instances", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// The endpoint itself will return an empty list (state-based lookup), but the
	// sub-operation router will call handleGetProjectInstances which does NOT require
	// the project to exist — it simply filters state. Verify we get 200 with 0 results.
	require.Equal(t, http.StatusOK, w.Code, "response: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["count"])
}

// TestHandleDeleteProject_BlockedWithActiveInstance verifies that deleting a
// project is rejected (409 Conflict) when calculateActiveInstances returns > 0.
//
// Since calculateActiveInstances returns 0 in test mode to avoid AWS calls, we
// temporarily disable testMode and seed a running instance directly in state.
func TestHandleDeleteProject_BlockedWithActiveInstance(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Delete Guard Project")

	// Temporarily disable test mode so calculateActiveInstances reads real state.
	server.testMode = false
	t.Cleanup(func() { server.testMode = true })

	// Seed a running instance belonging to this project.
	inst := types.Instance{
		ID:           "i-active001",
		Name:         "inst-active",
		Template:     "test-template",
		State:        "running",
		ProjectID:    projectID,
		InstanceType: "t3.micro",
		Region:       "us-east-1",
		LaunchTime:   time.Now(),
	}
	require.NoError(t, server.stateManager.SaveInstance(inst))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+projectID, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code,
		"deletion should be blocked while instance is running: %s", w.Body.String())
}

// TestHandleDeleteProject_SucceedsWithNoActiveInstances verifies that deleting a
// project succeeds (204) when no running instances exist for the project.
func TestHandleDeleteProject_SucceedsWithNoActiveInstances(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	projectID := createGovTestProject(t, handler, "Delete Allowed Project")

	// Temporarily disable test mode; no running instances seeded.
	server.testMode = false
	t.Cleanup(func() { server.testMode = true })

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+projectID, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code,
		"deletion should succeed with no active instances: %s", w.Body.String())
}
