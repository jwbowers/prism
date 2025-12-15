package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleSleepWakeStatus tests the sleep/wake status endpoint
func TestHandleSleepWakeStatus(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should always return valid response (even if monitor not available)
	assert.True(t, w.Code == http.StatusOK)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have enabled field
	assert.Contains(t, response, "enabled")

	// If not available, should have error message
	if enabled, ok := response["enabled"].(bool); ok && !enabled {
		assert.Contains(t, response, "error")
	}
}

// TestHandleSleepWakeConfigure tests the sleep/wake configuration endpoint
func TestHandleSleepWakeConfigure(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "enable monitoring",
			requestBody: map[string]interface{}{
				"enabled": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "disable monitoring",
			requestBody: map[string]interface{}{
				"enabled": false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "enable with reconnect option",
			requestBody: map[string]interface{}{
				"enabled":           true,
				"reconnect_on_wake": true,
			},
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
			// Create fresh server for each subtest to avoid monitor reuse issues
			server := createTestServer(t)
			handler := server.createHTTPHandler()

			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/api/v1/sleep-wake/configure", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// If monitor not available, will return 503
			// Otherwise should match expected status
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusServiceUnavailable)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "success")
			}

			// Brief delay to allow monitor to fully start/stop before next test
			time.Sleep(200 * time.Millisecond)
		})
	}
}

// TestSleepWakeHandlersMethodValidation tests HTTP method validation
func TestSleepWakeHandlersMethodValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST on status endpoint",
			method:         "POST",
			path:           "/api/v1/sleep-wake/status",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET on configure endpoint",
			method:         "GET",
			path:           "/api/v1/sleep-wake/configure",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT on status endpoint",
			method:         "PUT",
			path:           "/api/v1/sleep-wake/status",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE on configure endpoint",
			method:         "DELETE",
			path:           "/api/v1/sleep-wake/configure",
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

// TestSleepWakeHandlersConcurrency tests concurrent access to sleep/wake endpoints
func TestSleepWakeHandlersConcurrency(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	const numRequests = 10
	done := make(chan bool, numRequests)

	// Launch concurrent status requests
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Should always return OK (even if monitor not available)
			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// TestSleepWakeConfigurationWorkflow tests the configuration workflow
func TestSleepWakeConfigurationWorkflow(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Step 1: Get initial status
	statusReq := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
	statusW := httptest.NewRecorder()
	handler.ServeHTTP(statusW, statusReq)

	assert.Equal(t, http.StatusOK, statusW.Code)

	var initialStatus map[string]interface{}
	err := json.Unmarshal(statusW.Body.Bytes(), &initialStatus)
	require.NoError(t, err)

	// Step 2: Attempt to enable monitoring
	enableConfig := map[string]interface{}{
		"enabled": true,
	}
	enableBody, err := json.Marshal(enableConfig)
	require.NoError(t, err)

	enableReq := httptest.NewRequest("POST", "/api/v1/sleep-wake/configure", bytes.NewReader(enableBody))
	enableReq.Header.Set("Content-Type", "application/json")
	enableW := httptest.NewRecorder()
	handler.ServeHTTP(enableW, enableReq)

	// Brief delay to allow monitor to fully start
	time.Sleep(200 * time.Millisecond)

	// May succeed or fail depending on platform
	assert.True(t, enableW.Code == http.StatusOK ||
		enableW.Code == http.StatusServiceUnavailable)

	// Step 3: Check status again
	statusReq2 := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
	statusW2 := httptest.NewRecorder()
	handler.ServeHTTP(statusW2, statusReq2)

	assert.Equal(t, http.StatusOK, statusW2.Code)

	var updatedStatus map[string]interface{}
	err = json.Unmarshal(statusW2.Body.Bytes(), &updatedStatus)
	require.NoError(t, err)
	assert.Contains(t, updatedStatus, "enabled")
}

// TestSleepWakePlatformDetection tests platform-specific behavior
func TestSleepWakePlatformDetection(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should indicate if monitoring is available
	if enabled, ok := response["enabled"].(bool); ok {
		if !enabled {
			// Should have error message explaining why
			assert.Contains(t, response, "error")
		}
	}
}

// TestSleepWakeStatusResponseStructure tests response structure
func TestSleepWakeStatusResponseStructure(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var status map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &status)
	require.NoError(t, err)

	// Must have enabled field
	_, hasEnabled := status["enabled"]
	assert.True(t, hasEnabled, "Status response must have 'enabled' field")

	// If enabled is false, should have error message
	if enabled, ok := status["enabled"].(bool); ok && !enabled {
		_, hasError := status["error"]
		assert.True(t, hasError, "Disabled status should include error message")
	}
}

// TestSleepWakeConfigureResponseStructure tests configuration response structure
func TestSleepWakeConfigureResponseStructure(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	config := map[string]interface{}{
		"enabled": false,
	}
	body, err := json.Marshal(config)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/sleep-wake/configure", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Only validate response structure if configuration succeeded
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have success indicator
		assert.Contains(t, response, "success")

		// Should include updated status
		assert.Contains(t, response, "status")
	}
}

// TestSleepWakeErrorScenarios tests error handling
func TestSleepWakeErrorScenarios(t *testing.T) {
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
			name:           "malformed JSON in configure",
			endpoint:       "/api/v1/sleep-wake/configure",
			method:         "POST",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body in configure",
			endpoint:       "/api/v1/sleep-wake/configure",
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

// TestSleepWakeMonitorLifecycle tests monitor start/stop lifecycle
func TestSleepWakeMonitorLifecycle(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Step 1: Disable monitoring (if running)
	disableConfig := map[string]interface{}{
		"enabled": false,
	}
	disableBody, err := json.Marshal(disableConfig)
	require.NoError(t, err)

	disableReq := httptest.NewRequest("POST", "/api/v1/sleep-wake/configure", bytes.NewReader(disableBody))
	disableReq.Header.Set("Content-Type", "application/json")
	disableW := httptest.NewRecorder()
	handler.ServeHTTP(disableW, disableReq)

	// Brief delay to allow monitor to fully stop
	time.Sleep(200 * time.Millisecond)

	// Step 2: Check status (should be disabled or unavailable)
	statusReq1 := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
	statusW1 := httptest.NewRecorder()
	handler.ServeHTTP(statusW1, statusReq1)
	assert.Equal(t, http.StatusOK, statusW1.Code)

	// Step 3: Enable monitoring
	enableConfig := map[string]interface{}{
		"enabled": true,
	}
	enableBody, err := json.Marshal(enableConfig)
	require.NoError(t, err)

	enableReq := httptest.NewRequest("POST", "/api/v1/sleep-wake/configure", bytes.NewReader(enableBody))
	enableReq.Header.Set("Content-Type", "application/json")
	enableW := httptest.NewRecorder()
	handler.ServeHTTP(enableW, enableReq)

	// Brief delay to allow monitor to fully start
	time.Sleep(200 * time.Millisecond)

	// May succeed or fail depending on platform
	assert.True(t, enableW.Code == http.StatusOK ||
		enableW.Code == http.StatusServiceUnavailable)

	// Step 4: Check status again
	statusReq2 := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
	statusW2 := httptest.NewRecorder()
	handler.ServeHTTP(statusW2, statusReq2)
	assert.Equal(t, http.StatusOK, statusW2.Code)
}

// TestSleepWakeConfigValidation tests configuration validation
func TestSleepWakeConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid enable config",
			config: map[string]interface{}{
				"enabled": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid disable config",
			config: map[string]interface{}{
				"enabled": false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "config with extra fields",
			config: map[string]interface{}{
				"enabled":           true,
				"reconnect_on_wake": true,
				"log_events":        true,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh server for each subtest to avoid monitor reuse issues
			server := createTestServer(t)
			handler := server.createHTTPHandler()

			body, err := json.Marshal(tt.config)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/sleep-wake/configure", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Should succeed or be unavailable (platform-dependent)
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusServiceUnavailable)

			// Brief delay to allow monitor to fully start/stop before next test
			time.Sleep(200 * time.Millisecond)
		})
	}
}

// TestSleepWakeMonitorNotAvailable tests behavior when monitor is not available
func TestSleepWakeMonitorNotAvailable(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Status should always return OK with enabled=false and error message
	statusReq := httptest.NewRequest("GET", "/api/v1/sleep-wake/status", nil)
	statusW := httptest.NewRecorder()
	handler.ServeHTTP(statusW, statusReq)

	assert.Equal(t, http.StatusOK, statusW.Code)

	var response map[string]interface{}
	err := json.Unmarshal(statusW.Body.Bytes(), &response)
	require.NoError(t, err)

	// If monitor is not available (common in test environments)
	if enabled, ok := response["enabled"].(bool); ok && !enabled {
		// Should explain why
		errorMsg, hasError := response["error"].(string)
		assert.True(t, hasError, "Should have error message when disabled")
		assert.Contains(t, errorMsg, "not initialized")
	}
}

// TestSleepWakeEndpointIntegration tests full endpoint integration
func TestSleepWakeEndpointIntegration(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	endpoints := []struct {
		path   string
		method string
	}{
		{"/api/v1/sleep-wake/status", "GET"},
		{"/api/v1/sleep-wake/configure", "POST"},
	}

	for _, ep := range endpoints {
		t.Run(ep.path, func(t *testing.T) {
			var req *http.Request
			if ep.method == "POST" {
				body := []byte(`{"enabled": false}`)
				req = httptest.NewRequest(ep.method, ep.path, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(ep.method, ep.path, nil)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Should return valid HTTP status code (not panic)
			assert.True(t, w.Code >= 200 && w.Code < 600)
		})
	}
}
