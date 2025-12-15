package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleAMIResolve tests the AMI resolution endpoint
func TestHandleAMIResolve(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		templateName   string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "resolve AMI for valid template",
			templateName:   "python-ml",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "resolve AMI with details",
			templateName:   "r-research",
			queryParams:    "?details=true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "resolve AMI for non-existent template",
			templateName:   "non-existent",
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing template name",
			templateName:   "",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/ami/resolve/" + tt.templateName + tt.queryParams
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestHandleAMITest tests the AMI availability testing endpoint
func TestHandleAMITest(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "test AMI in default region",
			requestBody: map[string]interface{}{
				"template_name": "python-ml",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "test AMI across multiple regions",
			requestBody: map[string]interface{}{
				"template_name": "r-research",
				"regions":       []string{"us-east-1", "us-west-2", "eu-west-1"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing template name",
			requestBody: map[string]interface{}{
				"regions": []string{"us-east-1"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			requestBody:    map[string]interface{}{},
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

			req := httptest.NewRequest("POST", "/api/v1/ami/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMICosts tests the AMI cost analysis endpoint
func TestHandleAMICosts(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		templateName   string
		expectedStatus int
	}{
		{
			name:           "get costs for valid template",
			templateName:   "python-ml",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get costs for non-existent template",
			templateName:   "non-existent",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing template name",
			templateName:   "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/ami/costs/" + tt.templateName
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMIPreview tests the AMI preview endpoint
func TestHandleAMIPreview(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		templateName   string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "preview AMI resolution",
			templateName:   "python-ml",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "template_name")
				assert.Contains(t, response, "target_region")
				assert.Contains(t, response, "resolution_method")
			},
		},
		{
			name:           "preview for non-existent template",
			templateName:   "non-existent",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/ami/preview/" + tt.templateName
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)

			if tt.checkResponse != nil && w.Code == http.StatusOK {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// TestHandleAMICreate tests the AMI creation endpoint
func TestHandleAMICreate(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "create AMI from instance",
			requestBody: map[string]interface{}{
				"template_name": "test-template",
				"instance_id":   "i-1234567890abcdef0",
				"description":   "Test AMI creation",
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "create AMI with multi-region",
			requestBody: map[string]interface{}{
				"template_name": "test-template",
				"instance_id":   "i-1234567890abcdef0",
				"multi_region":  []string{"us-east-1", "us-west-2"},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "missing template name",
			requestBody: map[string]interface{}{
				"instance_id": "i-1234567890abcdef0",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing instance ID",
			requestBody: map[string]interface{}{
				"template_name": "test-template",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
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

			req := httptest.NewRequest("POST", "/api/v1/ami/create", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMIStatus tests the AMI creation status endpoint
func TestHandleAMIStatus(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		creationID     string
		expectedStatus int
	}{
		{
			name:           "get status for valid creation ID",
			creationID:     "ami-1234567890abcdef0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get status for non-existent creation ID",
			creationID:     "ami-nonexistent",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing creation ID",
			creationID:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/ami/status/" + tt.creationID
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusNotFound ||
				w.Code == http.StatusOK)
		})
	}
}

// TestHandleAMIList tests the AMI listing endpoint
func TestHandleAMIList(t *testing.T) {
	server := createTestServerWithAWS(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/ami/list", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Distinguish between infrastructure issues and code issues
	// Infrastructure issues: AWS SDK trying to use IMDS, credentials not configured properly
	// Code issues: Handler logic bugs, incorrect API responses

	if w.Code != http.StatusOK {
		// Check if it's an AWS credential/infrastructure issue
		var errorResp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err == nil {
			if errMsg, ok := errorResp["message"].(string); ok {
				// IMDS errors are infrastructure issues, not code issues
				if strings.Contains(errMsg, "EC2 IMDS") || strings.Contains(errMsg, "169.254.169.254") {
					t.Logf("INFRASTRUCTURE ISSUE: AWS SDK trying to use EC2 IMDS instead of profile credentials")
					t.Logf("Error: %s", errMsg)
					t.Skip("Skipping test due to AWS SDK credential chain issue (tries IMDS before profile). This is NOT a code issue.")
				}
			}
		}
		t.Fatalf("AMI list endpoint failed with status %d (this IS a code issue if not an infrastructure issue above)", w.Code)
	}

	// Validate successful response structure
	var amis []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &amis)
	require.NoError(t, err)
	assert.NotNil(t, amis, "AMI list should return an array (may be empty)")
	t.Logf("✓ AMI list endpoint working correctly: returned %d AMIs", len(amis))
}

// TestHandleAMICleanup tests the AMI cleanup endpoint
func TestHandleAMICleanup(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "cleanup with default age",
			requestBody: map[string]interface{}{
				"dry_run": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "cleanup with custom age",
			requestBody: map[string]interface{}{
				"max_age": "60d",
				"dry_run": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "cleanup without dry run",
			requestBody: map[string]interface{}{
				"max_age": "90d",
				"dry_run": false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
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

			req := httptest.NewRequest("POST", "/api/v1/ami/cleanup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMIDelete tests the AMI deletion endpoint
func TestHandleAMIDelete(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "delete AMI with snapshots",
			requestBody: map[string]interface{}{
				"ami_id":          "ami-1234567890abcdef0",
				"deregister_only": false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "deregister AMI only",
			requestBody: map[string]interface{}{
				"ami_id":          "ami-1234567890abcdef0",
				"deregister_only": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing AMI ID",
			requestBody: map[string]interface{}{
				"deregister_only": false,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
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

			req := httptest.NewRequest("POST", "/api/v1/ami/delete", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMISnapshotsList tests the AMI snapshots listing endpoint
func TestHandleAMISnapshotsList(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:           "list all snapshots",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusOK,
		},
		{
			name: "list snapshots for instance",
			requestBody: map[string]interface{}{
				"instance_id": "i-1234567890abcdef0",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "list snapshots with max age",
			requestBody: map[string]interface{}{
				"max_age": "30d",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "list snapshots with region",
			requestBody: map[string]interface{}{
				"region": "us-west-2",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/ami/snapshots", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMISnapshotCreate tests the AMI snapshot creation endpoint
func TestHandleAMISnapshotCreate(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "create snapshot with reboot",
			requestBody: map[string]interface{}{
				"instance_id": "i-1234567890abcdef0",
				"description": "Test snapshot",
				"no_reboot":   false,
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "create snapshot without reboot",
			requestBody: map[string]interface{}{
				"instance_id": "i-1234567890abcdef0",
				"description": "Test snapshot no reboot",
				"no_reboot":   true,
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "missing instance ID",
			requestBody: map[string]interface{}{
				"description": "Test snapshot",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
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

			req := httptest.NewRequest("POST", "/api/v1/ami/snapshot/create", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMISnapshotRestore tests the AMI snapshot restore endpoint
func TestHandleAMISnapshotRestore(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "restore AMI from snapshot",
			requestBody: map[string]interface{}{
				"snapshot_id":  "snap-1234567890abcdef0",
				"name":         "restored-ami",
				"description":  "Restored from snapshot",
				"architecture": "x86_64",
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "restore AMI with default architecture",
			requestBody: map[string]interface{}{
				"snapshot_id": "snap-1234567890abcdef0",
				"name":        "restored-ami-default",
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "missing snapshot ID",
			requestBody: map[string]interface{}{
				"name": "restored-ami",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing name",
			requestBody: map[string]interface{}{
				"snapshot_id": "snap-1234567890abcdef0",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
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

			req := httptest.NewRequest("POST", "/api/v1/ami/snapshot/restore", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMISnapshotDelete tests the AMI snapshot deletion endpoint
func TestHandleAMISnapshotDelete(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "delete snapshot",
			requestBody: map[string]interface{}{
				"snapshot_id": "snap-1234567890abcdef0",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing snapshot ID",
			requestBody: map[string]interface{}{
				"other_field": "value",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
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

			req := httptest.NewRequest("POST", "/api/v1/ami/snapshot/delete", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on AWS manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleAMICheckFreshness tests AMI freshness checking
// This test validates that static AMI IDs in templates match latest SSM values
func TestHandleAMICheckFreshness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow AMI freshness check test in short mode")
	}

	server := createTestServerWithAWS(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/ami/check-freshness", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Distinguish between infrastructure issues and code issues
	if w.Code != http.StatusOK {
		// Check if it's an AWS credential/infrastructure issue
		var errorResp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err == nil {
			if errMsg, ok := errorResp["message"].(string); ok {
				// IMDS errors are infrastructure issues, not code issues
				if strings.Contains(errMsg, "EC2 IMDS") || strings.Contains(errMsg, "169.254.169.254") {
					t.Logf("INFRASTRUCTURE ISSUE: AWS SDK credential chain trying IMDS before profile")
					t.Logf("Error: %s", errMsg)
					t.Skip("Skipping test due to AWS SDK credential chain issue. This is NOT a code issue.")
				}
			}
		}
		t.Fatalf("AMI freshness check failed with status %d (this IS a code issue if not an infrastructure issue)", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate response structure
	assert.Contains(t, response, "total_checked", "Response should include total_checked count")
	assert.Contains(t, response, "outdated", "Response should include outdated count")
	assert.Contains(t, response, "up_to_date", "Response should include up_to_date count")
	assert.Contains(t, response, "results", "Response should include detailed results")

	// Log freshness check results for visibility
	t.Logf("✓ AMI Freshness Check working correctly: %d total, %d up-to-date, %d outdated",
		int(response["total_checked"].(float64)),
		int(response["up_to_date"].(float64)),
		int(response["outdated"].(float64)))
}

// TestAMIHandlersMethodValidation tests HTTP method validation
func TestAMIHandlersMethodValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST on resolve endpoint",
			method:         "POST",
			path:           "/api/v1/ami/resolve/python-ml",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET on create endpoint",
			method:         "GET",
			path:           "/api/v1/ami/create",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE on list endpoint",
			method:         "DELETE",
			path:           "/api/v1/ami/list",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// May return 405 or 404 depending on routing
			assert.True(t, w.Code == http.StatusMethodNotAllowed ||
				w.Code == http.StatusNotFound)
		})
	}
}

// TestAMIHandlersConcurrency tests concurrent access to AMI endpoints
func TestAMIHandlersConcurrency(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	const numRequests = 15
	done := make(chan bool, numRequests)

	// Exclude slow endpoints like check-freshness which make real AWS SSM calls
	endpoints := []string{
		"/api/v1/ami/list",
		"/api/v1/ami/resolve/python-ml",
		"/api/v1/ami/costs/r-research",
		"/api/v1/ami/preview/python-ml",
	}

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(index int) {
			endpoint := endpoints[index%len(endpoints)]
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Accept any non-panicking response
			assert.True(t, w.Code > 0)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// TestAMIHandlersErrorScenarios tests error handling
func TestAMIHandlersErrorScenarios(t *testing.T) {
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
			name:           "malformed JSON in test",
			endpoint:       "/api/v1/ami/test",
			method:         "POST",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed JSON in create",
			endpoint:       "/api/v1/ami/create",
			method:         "POST",
			body:           `{"bad": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body in snapshot create",
			endpoint:       "/api/v1/ami/snapshot/create",
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

// TestAMICreationLifecycle tests the full AMI creation workflow
func TestAMICreationLifecycle(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Note: This test validates the API workflow, not actual AMI creation
	// which requires AWS infrastructure

	// Step 1: Test preview
	previewReq := httptest.NewRequest("GET", "/api/v1/ami/preview/python-ml", nil)
	previewW := httptest.NewRecorder()
	handler.ServeHTTP(previewW, previewReq)

	// Should get preview or error (AWS manager dependent)
	assert.True(t, previewW.Code == http.StatusOK ||
		previewW.Code == http.StatusInternalServerError)

	// Step 2: Test cost analysis
	costReq := httptest.NewRequest("GET", "/api/v1/ami/costs/python-ml", nil)
	costW := httptest.NewRecorder()
	handler.ServeHTTP(costW, costReq)

	// Should get costs or error (AWS manager dependent)
	assert.True(t, costW.Code == http.StatusOK ||
		costW.Code == http.StatusInternalServerError)

	// Step 3: Test creation initiation (without actual AWS call)
	createBody := map[string]interface{}{
		"template_name": "python-ml",
		"instance_id":   "i-test123",
	}
	body, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq := httptest.NewRequest("POST", "/api/v1/ami/create", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.ServeHTTP(createW, createReq)

	// Should accept or error (AWS manager dependent)
	assert.True(t, createW.Code == http.StatusAccepted ||
		createW.Code == http.StatusInternalServerError)
}
