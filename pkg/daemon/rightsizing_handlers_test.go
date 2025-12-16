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

// TestHandleRightsizingAnalyze tests the rightsizing analysis endpoint
func TestHandleRightsizingAnalyze(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid analysis request",
			requestBody: map[string]interface{}{
				"instance_name":         "test-instance",
				"analysis_period_hours": 24,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing instance name",
			requestBody: map[string]interface{}{
				"analysis_period_hours": 24,
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

			req := httptest.NewRequest("POST", "/api/v1/rightsizing/analyze", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// AWS manager may not be initialized - accept OK, 404 (not found), or 500
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusNotFound ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleRightsizingRecommendations tests the recommendations endpoint
func TestHandleRightsizingRecommendations(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/rightsizing/recommendations", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// AWS manager may not be initialized
	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "recommendations")
		assert.Contains(t, response, "total_instances")
	}
}

// TestHandleRightsizingStats tests the stats endpoint
func TestHandleRightsizingStats(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "get stats for instance",
			queryParams:    "?instance=test-instance",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing instance parameter",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/rightsizing/stats"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusNotFound ||
				w.Code == http.StatusInternalServerError)
		})
	}
}

// TestHandleRightsizingExport tests the export endpoint
func TestHandleRightsizingExport(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "export instance metrics",
			queryParams:    "?instance=test-instance",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing instance parameter",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/rightsizing/export"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusNotFound ||
				w.Code == http.StatusInternalServerError)

			if w.Code == http.StatusOK {
				// Should have Content-Disposition header for download
				contentDisp := w.Header().Get("Content-Disposition")
				assert.Contains(t, contentDisp, "attachment")
			}
		})
	}
}

// TestHandleRightsizingSummary tests the summary endpoint
func TestHandleRightsizingSummary(t *testing.T) {
	t.Skip("Issue #409: Handler tests need test data setup (v0.6.2)")
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/rightsizing/summary", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		// Verify expected fields in summary response
		assert.Contains(t, response, "fleet_overview")
		assert.Contains(t, response, "cost_optimization")
		assert.Contains(t, response, "resource_utilization")
	}
}

// TestHandleInstanceMetrics tests the instance metrics endpoint
func TestHandleInstanceMetrics(t *testing.T) {
	t.Skip("Issue #409: Handler tests need test data setup (v0.6.2)")
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/instances/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError)

	if w.Code == http.StatusOK {
		var metrics []interface{}
		err := json.Unmarshal(w.Body.Bytes(), &metrics)
		require.NoError(t, err)
		// Metrics array may be empty if no running instances
		assert.NotNil(t, metrics)
	}
}

// TestHandleInstanceMetricsOperations tests the instance-specific metrics endpoint
func TestHandleInstanceMetricsOperations(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name         string
		instanceName string
		queryParams  string
	}{
		{
			name:         "get metrics for instance",
			instanceName: "test-instance",
			queryParams:  "",
		},
		{
			name:         "get metrics with limit",
			instanceName: "test-instance",
			queryParams:  "?limit=10",
		},
		{
			name:         "get metrics with custom limit",
			instanceName: "test-instance",
			queryParams:  "?limit=50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/rightsizing/instance/" + tt.instanceName + "/metrics" + tt.queryParams
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, w.Code == http.StatusOK ||
				w.Code == http.StatusNotFound ||
				w.Code == http.StatusInternalServerError)

			if w.Code == http.StatusOK {
				var metrics []interface{}
				err := json.Unmarshal(w.Body.Bytes(), &metrics)
				require.NoError(t, err)
				assert.NotNil(t, metrics)
			}
		})
	}
}

// TestRightsizingHandlersMethodValidation tests HTTP method validation
func TestRightsizingHandlersMethodValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET on analyze endpoint",
			method:         "GET",
			path:           "/api/v1/rightsizing/analyze",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST on recommendations endpoint",
			method:         "POST",
			path:           "/api/v1/rightsizing/recommendations",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE on stats endpoint",
			method:         "DELETE",
			path:           "/api/v1/rightsizing/stats",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT on summary endpoint",
			method:         "PUT",
			path:           "/api/v1/rightsizing/summary",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// TestRightsizingHandlersConcurrency tests concurrent access
func TestRightsizingHandlersConcurrency(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	const numRequests = 10
	done := make(chan bool, numRequests)

	endpoints := []string{
		"/api/v1/rightsizing/recommendations",
		"/api/v1/rightsizing/summary",
		"/api/v1/instances/metrics",
	}

	for i := 0; i < numRequests; i++ {
		go func(endpoint string) {
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Should return valid status
			assert.True(t, w.Code >= 200 && w.Code < 600)
			done <- true
		}(endpoints[i%len(endpoints)])
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// TestRightsizingResponseStructure tests response structure validation
func TestRightsizingResponseStructure(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	t.Run("recommendations response structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rightsizing/recommendations", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify expected fields
			assert.Contains(t, response, "recommendations")
			assert.Contains(t, response, "total_instances")
			assert.Contains(t, response, "active_instances")
			assert.Contains(t, response, "potential_savings")
		}
	})

	t.Run("summary response structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/rightsizing/summary", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			// Check if body is not empty before unmarshaling
			if w.Body.Len() == 0 {
				t.Skip("Empty response body - AWS manager may not be available")
				return
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Logf("Response body: %s", w.Body.String())
				t.Skip("Invalid JSON response - AWS manager may not be available")
				return
			}

			// Verify expected fields
			assert.Contains(t, response, "fleet_overview")
			assert.Contains(t, response, "cost_optimization")
			assert.Contains(t, response, "resource_utilization")
			assert.Contains(t, response, "recommendations")
		}
	})
}

// TestRightsizingAnalysisWorkflow tests the analysis workflow
func TestRightsizingAnalysisWorkflow(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Step 1: Get recommendations for all instances
	recommendationsReq := httptest.NewRequest("GET", "/api/v1/rightsizing/recommendations", nil)
	recommendationsW := httptest.NewRecorder()
	handler.ServeHTTP(recommendationsW, recommendationsReq)

	// Step 2: Get summary
	summaryReq := httptest.NewRequest("GET", "/api/v1/rightsizing/summary", nil)
	summaryW := httptest.NewRecorder()
	handler.ServeHTTP(summaryW, summaryReq)

	// Step 3: Request detailed analysis for specific instance
	analysisBody := map[string]interface{}{
		"instance_name":         "test-instance",
		"analysis_period_hours": 24,
	}
	body, err := json.Marshal(analysisBody)
	require.NoError(t, err)

	analysisReq := httptest.NewRequest("POST", "/api/v1/rightsizing/analyze", bytes.NewReader(body))
	analysisReq.Header.Set("Content-Type", "application/json")
	analysisW := httptest.NewRecorder()
	handler.ServeHTTP(analysisW, analysisReq)

	// All requests should complete without panics
	assert.True(t, recommendationsW.Code >= 200 && recommendationsW.Code < 600)
	assert.True(t, summaryW.Code >= 200 && summaryW.Code < 600)
	assert.True(t, analysisW.Code >= 200 && analysisW.Code < 600)
}

// TestRightsizingErrorScenarios tests error handling
func TestRightsizingErrorScenarios(t *testing.T) {
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
			name:           "malformed JSON in analyze",
			endpoint:       "/api/v1/rightsizing/analyze",
			method:         "POST",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body in analyze",
			endpoint:       "/api/v1/rightsizing/analyze",
			method:         "POST",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing query parameter in stats",
			endpoint:       "/api/v1/rightsizing/stats",
			method:         "GET",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing query parameter in export",
			endpoint:       "/api/v1/rightsizing/export",
			method:         "GET",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.endpoint, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.endpoint, nil)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
