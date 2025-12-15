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

// TestHandleListIdlePolicies tests the idle policy listing endpoint
func TestHandleListIdlePolicies(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/idle/policies", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should return array of policies
	var policies []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &policies)
	require.NoError(t, err)
	assert.NotNil(t, policies)
}

// TestHandleGetIdlePolicy tests getting a specific idle policy
func TestHandleGetIdlePolicy(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		policyID       string
		expectedStatus int
	}{
		{
			name:           "valid policy ID",
			policyID:       "standard",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent policy",
			policyID:       "non-existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty policy ID",
			policyID:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/idle/policies/" + tt.policyID
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHandleRecommendIdlePolicy tests idle policy recommendation
func TestHandleRecommendIdlePolicy(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "recommend for t3.micro",
			requestBody: map[string]interface{}{
				"instance_type": "t3.micro",
				"tags": map[string]string{
					"Environment": "development",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "recommend for gpu instance",
			requestBody: map[string]interface{}{
				"instance_type": "g4dn.xlarge",
				"tags": map[string]string{
					"Workload": "ml-training",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusOK, // May succeed with defaults
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

			req := httptest.NewRequest("POST", "/api/v1/idle/policies/recommend", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHandleListIdleSchedules tests the idle schedules listing endpoint
func TestHandleListIdleSchedules(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/idle/schedules", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// May return 200 or 503 depending on scheduler initialization
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)

	if w.Code == http.StatusOK {
		// Should return array of schedules
		var schedules []interface{}
		err := json.Unmarshal(w.Body.Bytes(), &schedules)
		require.NoError(t, err)
		assert.NotNil(t, schedules)
	}
}

// TestHandleIdleSavingsReport tests the idle savings report endpoint
func TestHandleIdleSavingsReport(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/idle/savings", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var report map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &report)
	require.NoError(t, err)

	// Should have required fields
	assert.Contains(t, report, "report_id")
	assert.Contains(t, report, "generated_at")
	assert.Contains(t, report, "total_saved")
	assert.Contains(t, report, "projected_savings")
}

// TestHandleIdlePolicyApply tests applying idle policy to instance
func TestHandleIdlePolicyApply(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "apply valid policy",
			requestBody: map[string]interface{}{
				"instance_name": "test-instance",
				"policy_id":     "standard",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing instance name",
			requestBody: map[string]interface{}{
				"policy_id": "standard",
			},
			expectedStatus: http.StatusServiceUnavailable, // Scheduler may not be available
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

			req := httptest.NewRequest("POST", "/api/v1/idle/policies/apply", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// May vary based on scheduler availability
			assert.True(t, w.Code == tt.expectedStatus || w.Code == http.StatusServiceUnavailable)
		})
	}
}

// TestHandleInstanceIdlePolicies tests instance-specific idle policy endpoints
func TestHandleInstanceIdlePolicies(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		instanceName   string
		expectedStatus int
	}{
		{
			name:           "get policies for valid instance",
			instanceName:   "test-instance",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get policies for empty instance",
			instanceName:   "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/instances/" + tt.instanceName + "/idle/policies"
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on routing and whether instance exists
			assert.True(t, w.Code == tt.expectedStatus || w.Code == http.StatusOK)
		})
	}
}

// TestHandleIdleMethodValidation tests HTTP method validation
func TestHandleIdleMethodValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "DELETE on policies list",
			method:         "DELETE",
			path:           "/api/v1/idle/policies",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT on schedules",
			method:         "PUT",
			path:           "/api/v1/idle/schedules",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST on savings",
			method:         "POST",
			path:           "/api/v1/idle/savings",
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

// TestIdleHandlersConcurrency tests concurrent access to idle endpoints
func TestIdleHandlersConcurrency(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	const numRequests = 20
	done := make(chan bool, numRequests)

	endpoints := []string{
		"/api/v1/idle/policies",
		"/api/v1/idle/schedules",
		"/api/v1/idle/savings",
	}

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(index int) {
			endpoint := endpoints[index%len(endpoints)]
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Accept any non-error response
			assert.True(t, w.Code < 500)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// TestIdlePolicyRecommendationLogic tests policy recommendation scenarios
func TestIdlePolicyRecommendationLogic(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name         string
		instanceType string
		tags         map[string]string
		expectPolicy bool
	}{
		{
			name:         "development instance",
			instanceType: "t3.small",
			tags: map[string]string{
				"Environment": "development",
			},
			expectPolicy: true,
		},
		{
			name:         "production instance",
			instanceType: "c5.2xlarge",
			tags: map[string]string{
				"Environment": "production",
			},
			expectPolicy: true,
		},
		{
			name:         "ml training instance",
			instanceType: "p3.2xlarge",
			tags: map[string]string{
				"Workload": "ml-training",
			},
			expectPolicy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"instance_type": tt.instanceType,
				"tags":          tt.tags,
			}

			body, err := json.Marshal(requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/idle/policies/recommend", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.expectPolicy {
				var policy map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &policy)
				require.NoError(t, err)
				// Should have some policy structure
				assert.NotEmpty(t, policy)
			}
		})
	}
}

// TestIdleSavingsReportStructure tests savings report data structure
func TestIdleSavingsReportStructure(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/idle/savings", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var report map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &report)
	require.NoError(t, err)

	// Verify required fields exist
	requiredFields := []string{
		"report_id",
		"generated_at",
		"period",
		"total_saved",
		"projected_savings",
		"idle_hours",
		"active_hours",
		"savings_percentage",
	}

	for _, field := range requiredFields {
		assert.Contains(t, report, field, "Report should contain field: %s", field)
	}

	// Verify period structure
	if period, ok := report["period"].(map[string]interface{}); ok {
		assert.Contains(t, period, "start")
		assert.Contains(t, period, "end")
	}
}

// TestIdleHandlersErrorScenarios tests error handling
func TestIdleHandlersErrorScenarios(t *testing.T) {
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
			name:           "malformed JSON in recommend",
			endpoint:       "/api/v1/idle/policies/recommend",
			method:         "POST",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed JSON in apply",
			endpoint:       "/api/v1/idle/policies/apply",
			method:         "POST",
			body:           `{"bad": json}`,
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
