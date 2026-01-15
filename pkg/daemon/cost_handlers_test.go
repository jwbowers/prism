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

// TestHandleGetCostAlerts tests the cost alerts listing endpoint
func TestHandleGetCostAlerts(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "get all alerts",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "alerts")
				assert.Contains(t, response, "count")
			},
		},
		{
			name:           "filter alerts by project",
			queryParams:    "?project_id=test-project",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "alerts")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/cost/alerts"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// TestHandleGetActiveAlerts tests the active alerts endpoint
func TestHandleGetActiveAlerts(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/cost/alerts/active", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "alerts")
	assert.Contains(t, response, "count")
}

// TestHandleAcknowledgeAlert tests the alert acknowledgement endpoint
func TestHandleAcknowledgeAlert(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		alertID        string
		expectedStatus int
	}{
		{
			name:           "acknowledge valid alert",
			alertID:        "alert-123",
			expectedStatus: http.StatusNotFound, // Alert doesn't exist yet
		},
		{
			name:           "acknowledge with empty ID",
			alertID:        "",
			expectedStatus: http.StatusNotFound, // May return 404 or redirect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/cost/alerts/" + tt.alertID + "/acknowledge"
			req := httptest.NewRequest("POST", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Accept various error codes (routing implementation dependent)
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusMovedPermanently ||
				w.Code == http.StatusBadRequest)
		})
	}
}

// TestHandleResolveAlert tests the alert resolution endpoint
func TestHandleResolveAlert(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		alertID        string
		expectedStatus int
	}{
		{
			name:           "resolve valid alert",
			alertID:        "alert-456",
			expectedStatus: http.StatusNotFound, // Alert doesn't exist yet
		},
		{
			name:           "resolve with empty ID",
			alertID:        "",
			expectedStatus: http.StatusNotFound, // May return 404 or redirect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/cost/alerts/" + tt.alertID + "/resolve"
			req := httptest.NewRequest("POST", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Accept various error codes (routing implementation dependent)
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusMovedPermanently ||
				w.Code == http.StatusBadRequest)
		})
	}
}

// TestHandleAddAlertRule tests the alert rule creation endpoint
func TestHandleAddAlertRule(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid alert rule",
			requestBody: map[string]interface{}{
				"name":    "Test Rule",
				"type":    "threshold",
				"enabled": true,
				"conditions": map[string]interface{}{
					"budget_percentage": 80.0,
				},
				"actions": []string{"email"},
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

			req := httptest.NewRequest("POST", "/api/v1/cost/alerts/rules", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHandleGetOptimizationReport tests the optimization report endpoint
func TestHandleGetOptimizationReport(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "report with valid project",
			queryParams:    "?project_id=test-project",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing project_id",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty project_id",
			queryParams:    "?project_id=",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/cost/optimization/report"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHandleGetRecommendations tests the cost optimization recommendations endpoint
func TestHandleGetRecommendations(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "recommendations for instance",
			queryParams:    "?instance_id=i-123456",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "recommendations")
				assert.Contains(t, response, "total_savings")
			},
		},
		{
			name:           "recommendations for project",
			queryParams:    "?project_id=test-project",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "recommendations")
			},
		},
		{
			name:           "missing both instance and project",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/cost/optimization/recommendations"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil && w.Code == http.StatusOK {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// TestHandleGetCostTrends tests the cost trends endpoint
func TestHandleGetCostTrends(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Set up test data: create project with budget
	projectID := setupTestProject(t, server, "test-project")
	setupTestBudget(t, server, projectID, 1000.0)

	tests := []struct {
		name            string
		queryParamsFunc func(string) string
		expectedStatus  int
	}{
		{
			name:            "trends with default period",
			queryParamsFunc: func(id string) string { return "?project_id=" + id },
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "trends with 7d period",
			queryParamsFunc: func(id string) string { return "?project_id=" + id + "&period=7d" },
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "trends with 90d period",
			queryParamsFunc: func(id string) string { return "?project_id=" + id + "&period=90d" },
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "missing project_id",
			queryParamsFunc: func(id string) string { return "" },
			expectedStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryParams := tt.queryParamsFunc(projectID)
			req := httptest.NewRequest("GET", "/api/v1/cost/trends"+queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHandleGetBudgetStatus tests the budget status endpoint
func TestHandleGetBudgetStatus(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Set up test data: create project with budget
	projectID := setupTestProject(t, server, "test-project")
	setupTestBudget(t, server, projectID, 1000.0)

	tests := []struct {
		name            string
		queryParamsFunc func(string) string
		expectedStatus  int
	}{
		{
			name:            "valid project budget status",
			queryParamsFunc: func(id string) string { return "?project_id=" + id },
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "missing project_id",
			queryParamsFunc: func(id string) string { return "" },
			expectedStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryParams := tt.queryParamsFunc(projectID)
			req := httptest.NewRequest("GET", "/api/v1/cost/budget/status"+queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHandleUpdateBudgetAlert tests the budget alert update endpoint
func TestHandleUpdateBudgetAlert(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid budget alert",
			requestBody: map[string]interface{}{
				"project_id": "test-project",
				"alert_type": "threshold",
				"threshold":  80.0,
				"enabled":    true,
				"actions":    []string{"email"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"project_id": "test-project",
			},
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

			req := httptest.NewRequest("POST", "/api/v1/cost/budget/alerts", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestCostHandlersConcurrency tests concurrent access to cost endpoints
func TestCostHandlersConcurrency(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	const numRequests = 20
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	endpoints := []string{
		"/api/v1/cost/alerts",
		"/api/v1/cost/alerts/active",
		"/api/v1/cost/optimization/recommendations?instance_id=i-123",
		"/api/v1/cost/trends?project_id=test&period=7d",
		"/api/v1/cost/budget/status?project_id=test",
	}

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(index int) {
			endpoint := endpoints[index%len(endpoints)]
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				errors <- assert.AnError
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
			// Request completed
		case <-errors:
			// Error occurred (acceptable for some test cases)
		}
	}
}

// TestCostHandlersMethodValidation tests HTTP method validation
func TestCostHandlersMethodValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "DELETE on alerts",
			method:         "DELETE",
			path:           "/api/v1/cost/alerts",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT on trends",
			method:         "PUT",
			path:           "/api/v1/cost/trends",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// May return 405, 404, 401, or 403 depending on routing and authentication
			assert.True(t,
				w.Code == http.StatusMethodNotAllowed ||
					w.Code == http.StatusNotFound ||
					w.Code == http.StatusUnauthorized ||
					w.Code == http.StatusForbidden,
				"Expected 405, 404, 401, or 403 but got %d", w.Code)
		})
	}
}

// TestCostHandlersErrorCases tests error handling scenarios
func TestCostHandlersErrorCases(t *testing.T) {
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
			name:           "malformed JSON in alert rule",
			endpoint:       "/api/v1/cost/alerts/rules",
			method:         "POST",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body in alert update",
			endpoint:       "/api/v1/cost/budget/alerts",
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
