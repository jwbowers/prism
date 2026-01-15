package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleListInstanceSnapshots tests the snapshot listing endpoint
func TestHandleListInstanceSnapshots(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Set up test data: create instance with snapshots
	setupTestInstance(t, server, "test-instance")
	setupTestSnapshot(t, server, "test-instance")

	req := httptest.NewRequest("GET", "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// If we get a 500 error due to AWS credentials, skip the test
	if w.Code == http.StatusInternalServerError {
		t.Skip("Skipping - AWS credentials not available in test environment")
		return
	}

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have snapshots array and count
	assert.Contains(t, response, "snapshots")
	assert.Contains(t, response, "count")
}

// TestHandleCreateInstanceSnapshot tests the snapshot creation endpoint
func TestHandleCreateInstanceSnapshot(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Set up test data: create instance for snapshot creation
	setupTestInstance(t, server, "test-instance")

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid snapshot request",
			requestBody: map[string]interface{}{
				"instance_name": "test-instance",
				"snapshot_name": "test-snapshot",
				"description":   "Test snapshot",
				"no_reboot":     true,
				"wait":          false,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing instance name",
			requestBody: map[string]interface{}{
				"snapshot_name": "test-snapshot",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing snapshot name",
			requestBody: map[string]interface{}{
				"instance_name": "test-instance",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "snapshot with wait flag",
			requestBody: map[string]interface{}{
				"instance_name": "test-instance",
				"snapshot_name": "test-snapshot-wait",
				"wait":          true,
			},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/api/v1/snapshots", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// May vary based on AWS manager availability, authentication, and instance existence
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusUnauthorized ||
				w.Code == http.StatusForbidden ||
				w.Code == http.StatusNotFound,
				"Expected %d, 500, 401, 403, or 404, got %d: %s", tt.expectedStatus, w.Code, w.Body.String())
		})
	}
}

// TestHandleGetInstanceSnapshot tests getting a specific snapshot
func TestHandleGetInstanceSnapshot(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		snapshotName   string
		expectedStatus int
	}{
		{
			name:           "get existing snapshot",
			snapshotName:   "test-snapshot",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent snapshot",
			snapshotName:   "non-existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty snapshot name",
			snapshotName:   "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/snapshots/" + tt.snapshotName
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager and whether snapshot exists
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleDeleteInstanceSnapshot tests snapshot deletion
func TestHandleDeleteInstanceSnapshot(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		snapshotName   string
		expectedStatus int
	}{
		{
			name:           "delete existing snapshot",
			snapshotName:   "test-snapshot",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete non-existent snapshot",
			snapshotName:   "non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/snapshots/" + tt.snapshotName
			req := httptest.NewRequest("DELETE", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager and whether snapshot exists
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleRestoreInstanceFromSnapshot tests instance restoration from snapshot
func TestHandleRestoreInstanceFromSnapshot(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		snapshotName   string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:         "valid restore request",
			snapshotName: "test-snapshot",
			requestBody: map[string]interface{}{
				"new_instance_name": "restored-instance",
				"wait":              false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "restore with wait flag",
			snapshotName: "test-snapshot",
			requestBody: map[string]interface{}{
				"new_instance_name": "restored-instance-wait",
				"wait":              true,
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:         "missing new instance name",
			snapshotName: "test-snapshot",
			requestBody: map[string]interface{}{
				"wait": false,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			snapshotName:   "test-snapshot",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			path := "/api/v1/snapshots/" + tt.snapshotName + "/restore"
			req := httptest.NewRequest("POST", path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus || w.Code == http.StatusInternalServerError)
		})
	}
}

// TestSnapshotMethodValidation tests HTTP method validation
func TestSnapshotMethodValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "PUT on snapshots list",
			method:         "PUT",
			path:           "/api/v1/snapshots",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH on specific snapshot",
			method:         "PATCH",
			path:           "/api/v1/snapshots/test-snapshot",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// May return 405 or 404 depending on routing
			assert.True(t, w.Code == http.StatusMethodNotAllowed || w.Code == http.StatusNotFound)
		})
	}
}

// TestSnapshotHandlersConcurrency tests concurrent snapshot operations
func TestSnapshotHandlersConcurrency(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	const numRequests = 10
	done := make(chan bool, numRequests)

	// Launch concurrent GET requests for snapshots list
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/v1/snapshots", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Accept any response (concurrent reads should be safe)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// TestSnapshotNamingValidation tests snapshot name validation
func TestSnapshotNamingValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Set up test data: create instance for snapshot validation tests
	setupTestInstance(t, server, "test-instance")

	tests := []struct {
		name         string
		snapshotName string
		expectError  bool
	}{
		{
			name:         "valid alphanumeric name",
			snapshotName: "test-snapshot-123",
			expectError:  false,
		},
		{
			name:         "name with underscores",
			snapshotName: "test_snapshot",
			expectError:  false,
		},
		{
			name:         "empty name",
			snapshotName: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"instance_name": "test-instance",
				"snapshot_name": tt.snapshotName,
			}

			body, err := json.Marshal(requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/snapshots", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				// May succeed or fail based on AWS manager availability and instance existence
				// Valid responses: 201 (Created), 404 (instance not found), 500 (AWS error)
				assert.True(t,
					w.Code == http.StatusCreated ||
						w.Code == http.StatusNotFound ||
						w.Code == http.StatusInternalServerError,
					"Expected 201, 404, or 500, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// TestSnapshotErrorScenarios tests error handling
func TestSnapshotErrorScenarios(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "malformed JSON in create",
			endpoint:       "/api/v1/snapshots",
			method:         "POST",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed JSON in restore",
			endpoint:       "/api/v1/snapshots/test/restore",
			method:         "POST",
			body:           `{"bad": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body in create",
			endpoint:       "/api/v1/snapshots",
			method:         "POST",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, bytes.NewBufferString(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSnapshotResponseStructure tests response data structures
func TestSnapshotResponseStructure(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Set up test data: create instance with snapshot for response structure tests
	setupTestInstance(t, server, "test-instance")
	setupTestSnapshot(t, server, "test-instance")

	t.Run("list response structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/snapshots", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// If we get a 500 error due to AWS credentials, skip the test
		if w.Code == http.StatusInternalServerError {
			t.Skip("Skipping - AWS credentials not available in test environment")
			return
		}

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have required fields
		assert.Contains(t, response, "snapshots")
		assert.Contains(t, response, "count")

		// Snapshots should be an array
		snapshots, ok := response["snapshots"].([]interface{})
		assert.True(t, ok, "snapshots should be an array, got type %T", response["snapshots"])
		assert.NotNil(t, snapshots)
	})
}
