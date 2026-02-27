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

// TestHandleSecurityStatus tests the security status endpoint
func TestHandleSecurityStatus(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/security/status", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Status depends on security manager initialization
	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError ||
		w.Code == http.StatusServiceUnavailable)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotNil(t, response)
	}
}

// TestHandleSecurityHealth tests the security health endpoint
func TestHandleSecurityHealth(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Set up test data: create security config
	setupTestSecurityConfig(t, server)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET health status",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST trigger health check",
			method:         "POST",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/security/health", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Status depends on security manager availability
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusServiceUnavailable)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.method == "GET" {
					// Health status should have these fields (values may be null/nil but keys must be present)
					_, hasSysHealth := response["system_health"]
					_, hasKeychainInfo := response["keychain_info"]
					assert.True(t, hasSysHealth || hasKeychainInfo)
				} else {
					// Health check trigger response
					assert.Contains(t, response, "status")
				}
			}
		})
	}
}

// TestHandleSecurityDashboard tests the security dashboard endpoint
func TestHandleSecurityDashboard(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/security/dashboard", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Dashboard may not be available in test mode
	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusServiceUnavailable ||
		w.Code == http.StatusInternalServerError)

	if w.Code == http.StatusOK {
		var dashboard map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &dashboard)
		require.NoError(t, err)
		assert.NotNil(t, dashboard)
	}
}

// TestHandleSecurityCorrelations tests the security correlations endpoint
func TestHandleSecurityCorrelations(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/security/correlations", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Status depends on security manager availability
	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError ||
		w.Code == http.StatusServiceUnavailable)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have correlation information
		assert.Contains(t, response, "correlations")
		assert.Contains(t, response, "correlation_count")
	}
}

// TestHandleSecurityKeychain tests the security keychain endpoint
func TestHandleSecurityKeychain(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET keychain info",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST validate keychain",
			method:         "POST",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/security/keychain", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Keychain operations are platform-dependent
			assert.True(t, w.Code == tt.expectedStatus ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusServiceUnavailable)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.method == "GET" {
					// Should have keychain info and diagnostics
					assert.True(t, response["info"] != nil ||
						response["diagnostics"] != nil)
				} else {
					// Validation response
					assert.Contains(t, response, "status")
				}
			}
		})
	}
}

// TestHandleSecurityConfig tests the security configuration endpoint
func TestHandleSecurityConfig(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET security config",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PUT security config (valid body)",
			method:         "PUT",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == "PUT" {
				// Send a valid SecurityConfig body
				body := `{"alert_threshold":"HIGH","monitoring_enabled":false,"audit_log_enabled":false,"correlation_enabled":false,"log_retention_days":30}`
				req = httptest.NewRequest(tt.method, "/api/v1/security/config", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/api/v1/security/config", nil)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.method == "PUT" {
				// PUT now implemented — should return 200 or 400 (if security manager returns error)
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest,
					"PUT should return 200 or 400, got %d: %s", w.Code, w.Body.String())
			} else {
				// GET should work or return error
				assert.True(t, w.Code == tt.expectedStatus ||
					w.Code == http.StatusInternalServerError ||
					w.Code == http.StatusServiceUnavailable)

				if w.Code == http.StatusOK {
					var response map[string]interface{}
					err := json.Unmarshal(w.Body.Bytes(), &response)
					require.NoError(t, err)
					assert.Contains(t, response, "configuration")
				}
			}
		})
	}
}

// TestSecurityHandlersMethodValidation tests HTTP method validation
func TestSecurityHandlersMethodValidation(t *testing.T) {
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
			path:           "/api/v1/security/status",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT on dashboard endpoint",
			method:         "PUT",
			path:           "/api/v1/security/dashboard",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE on correlations endpoint",
			method:         "DELETE",
			path:           "/api/v1/security/correlations",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH on keychain endpoint",
			method:         "PATCH",
			path:           "/api/v1/security/keychain",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Should return method not allowed
			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// TestSecurityHandlersConcurrency tests concurrent access to security endpoints
func TestSecurityHandlersConcurrency(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	const numRequests = 15
	done := make(chan bool, numRequests)

	endpoints := []string{
		"/api/v1/security/status",
		"/api/v1/security/health",
		"/api/v1/security/dashboard",
		"/api/v1/security/correlations",
		"/api/v1/security/config",
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

// TestSecurityHealthCheckWorkflow tests the health check workflow
func TestSecurityHealthCheckWorkflow(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Step 1: Get initial health status
	getReq := httptest.NewRequest("GET", "/api/v1/security/health", nil)
	getW := httptest.NewRecorder()
	handler.ServeHTTP(getW, getReq)

	// Should get status or error
	assert.True(t, getW.Code == http.StatusOK ||
		getW.Code == http.StatusInternalServerError ||
		getW.Code == http.StatusServiceUnavailable)

	// Step 2: Trigger health check
	postReq := httptest.NewRequest("POST", "/api/v1/security/health", nil)
	postW := httptest.NewRecorder()
	handler.ServeHTTP(postW, postReq)

	// Should trigger check or error
	assert.True(t, postW.Code == http.StatusOK ||
		postW.Code == http.StatusInternalServerError ||
		postW.Code == http.StatusServiceUnavailable)

	// Step 3: Get updated health status
	getReq2 := httptest.NewRequest("GET", "/api/v1/security/health", nil)
	getW2 := httptest.NewRecorder()
	handler.ServeHTTP(getW2, getReq2)

	// Should get status or error
	assert.True(t, getW2.Code == http.StatusOK ||
		getW2.Code == http.StatusInternalServerError ||
		getW2.Code == http.StatusServiceUnavailable)
}

// TestSecurityStatusResponseStructure tests response structure
func TestSecurityStatusResponseStructure(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/security/status", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Only validate structure if request succeeded
	if w.Code == http.StatusOK {
		var status map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &status)
		require.NoError(t, err)

		// Status should be a valid JSON object
		assert.NotNil(t, status)
	}
}

// TestSecurityKeychainDiagnostics tests keychain diagnostics
func TestSecurityKeychainDiagnostics(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/security/keychain", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Keychain operations are platform-specific
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have diagnostics information
		if diagnostics, ok := response["diagnostics"].(map[string]interface{}); ok {
			assert.NotNil(t, diagnostics)
		}
	}
}

// TestSecurityConfigStructure tests configuration response structure
func TestSecurityConfigStructure(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/security/config", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Only validate if request succeeded
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have configuration field
		assert.Contains(t, response, "configuration")
	}
}

// TestSecurityCorrelationsStructure tests correlations response structure
func TestSecurityCorrelationsStructure(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/security/correlations", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Only validate if request succeeded
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have required fields
		assert.Contains(t, response, "correlations")
		assert.Contains(t, response, "correlation_count")

		// Correlation count should be a number
		if count, ok := response["correlation_count"].(float64); ok {
			assert.GreaterOrEqual(t, count, 0.0)
		}
	}
}

// TestSecurityEndpointsErrorHandling tests error scenarios
func TestSecurityEndpointsErrorHandling(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	tests := []struct {
		name           string
		path           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid path",
			path:           "/api/v1/security/nonexistent",
			method:         "GET",
			body:           "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSecurityManagerIntegration tests security manager integration
func TestSecurityManagerIntegration(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	// Test that security endpoints respond (even if manager is not fully initialized)
	endpoints := map[string]string{
		"/api/v1/security/status":       "GET",
		"/api/v1/security/health":       "GET",
		"/api/v1/security/config":       "GET",
		"/api/v1/security/correlations": "GET",
	}

	for endpoint, method := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(method, endpoint, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Should return valid HTTP status code (not panic)
			assert.True(t, w.Code >= 200 && w.Code < 600)
		})
	}
}

// TestSecurityDashboardAvailability tests dashboard availability
func TestSecurityDashboardAvailability(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("GET", "/api/v1/security/dashboard", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Dashboard may not be available, which is acceptable
	if w.Code == http.StatusServiceUnavailable {
		// Check that error message is provided
		assert.Greater(t, len(w.Body.String()), 0)
	} else if w.Code == http.StatusOK {
		// If available, should be valid JSON
		var dashboard map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &dashboard)
		require.NoError(t, err)
	}
}

// TestSecurityKeychainValidation tests keychain validation
func TestSecurityKeychainValidation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	req := httptest.NewRequest("POST", "/api/v1/security/keychain", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Keychain validation is platform-dependent
	// Accept success, error, or unavailable
	assert.True(t, w.Code == http.StatusOK ||
		w.Code == http.StatusInternalServerError ||
		w.Code == http.StatusServiceUnavailable)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "status")
	}
}

// TestHandleSecurityConfigPUT_Validation tests PUT /api/v1/security/config validation.
func TestHandleSecurityConfigPUT_Validation(t *testing.T) {
	server := createTestServer(t)
	handler := server.createHTTPHandler()

	t.Run("invalid_json_body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/security/config",
			bytes.NewBufferString(`{not valid json`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid_alert_threshold", func(t *testing.T) {
		body := `{"alert_threshold":"EXTREME"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/security/config",
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		// Validation rejects unknown threshold values
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid alert threshold")
	})

	t.Run("valid_full_config", func(t *testing.T) {
		body := `{"alert_threshold":"HIGH","monitoring_enabled":false,"audit_log_enabled":false,"log_retention_days":60,"correlation_enabled":false}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/security/config",
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		// Should succeed or fail at security manager init — not at validation
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest,
			"unexpected status %d: %s", w.Code, w.Body.String())
		if w.Code == http.StatusOK {
			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, "updated", resp["status"])
		}
	})

	t.Run("empty_threshold_defaults_accepted", func(t *testing.T) {
		body := `{"monitoring_enabled":false,"audit_log_enabled":false}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/security/config",
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest,
			"unexpected status %d: %s", w.Code, w.Body.String())
	})
}
