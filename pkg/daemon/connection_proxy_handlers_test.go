package daemon

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedProxyInstance adds a running instance to state so proxy handlers
// can proceed past the instance-lookup guard to reach validation logic.
func seedProxyInstance(t *testing.T, server *Server) {
	t.Helper()
	inst := types.Instance{
		Name:     "test-proxy-instance",
		ID:       "i-test123",
		State:    "running",
		PublicIP: "127.0.0.1", // loopback — proxy will fail fast (connection refused), not hang
	}
	require.NoError(t, server.stateManager.SaveInstance(inst))
}

// TestDCVProxyPortValidation verifies /dcv-proxy/ rejects invalid port values
// with 400 Bad Request before attempting any proxy connection.
func TestDCVProxyPortValidation(t *testing.T) {
	server := createTestServer(t)
	seedProxyInstance(t, server)
	handler := server.createHTTPHandler()

	invalidPorts := []struct {
		name string
		port string
	}{
		{"port_zero", "0"},
		{"port_above_max", "65536"},
		{"port_negative", "-1"},
		{"port_text", "abc"},
		{"port_float", "8080.5"},
		{"port_overflow", "99999"},
	}

	for _, tt := range invalidPorts {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET",
				fmt.Sprintf("/dcv-proxy/test-proxy-instance/?port=%s", tt.port), nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"port=%q should be rejected", tt.port)
		})
	}
}

// TestWebProxyPortValidation verifies /web-proxy/ rejects invalid port values.
func TestWebProxyPortValidation(t *testing.T) {
	server := createTestServer(t)
	seedProxyInstance(t, server)
	handler := server.createHTTPHandler()

	invalidPorts := []struct {
		name string
		port string
	}{
		{"port_zero", "0"},
		{"port_above_max", "65536"},
		{"port_negative", "-80"},
		{"port_text", "jupyter"},
	}

	for _, tt := range invalidPorts {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET",
				fmt.Sprintf("/web-proxy/test-proxy-instance/?port=%s", tt.port), nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"port=%q should be rejected", tt.port)
		})
	}
}

// TestAWSProxyFederationTokenValidation verifies /aws-proxy/ rejects malformed
// federation tokens before any redirect or proxy operation.
func TestAWSProxyFederationTokenValidation(t *testing.T) {
	server := createTestServer(t)
	seedProxyInstance(t, server)
	handler := server.createHTTPHandler()

	invalidTokens := []struct {
		name  string
		token string
	}{
		{"spaces", "token with spaces"},
		{"angle_brackets", "token<script>"},
		{"null_byte", "token\x00null"},
		{"semicolon", "token;cmd"},
		{"ampersand", "token&other"},
		{"percent_encoding", "token%20encoded"},
		{"hash", "token#fragment"},
		{"newline", "token\ninjection"},
	}

	for _, tt := range invalidTokens {
		t.Run(tt.name, func(t *testing.T) {
			// URL-encode the token to ensure it reaches the handler as-is
			// (the handler validates the raw query param value)
			req := httptest.NewRequest("GET",
				"/aws-proxy/console", nil)
			q := req.URL.Query()
			q.Set("token", tt.token)
			req.URL.RawQuery = q.Encode()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"token=%q should be rejected", tt.token)
		})
	}
}

// TestAWSProxyValidTokenPassesValidation verifies that well-formed tokens
// are accepted by the validation layer (handler may fail later for other reasons).
func TestAWSProxyValidTokenPassesValidation(t *testing.T) {
	server := createTestServer(t)
	seedProxyInstance(t, server)
	handler := server.createHTTPHandler()

	validTokens := []struct {
		name  string
		token string
	}{
		{"alphanumeric", "abc123ABC"},
		{"base64_chars", "abc123+/="},
		{"url_safe_chars", "abc123_-"},
		{"long_token", "aB1+/=_-aB1+/=_-aB1+/=_-aB1+/=_-"},
	}

	for _, tt := range validTokens {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET",
				"/aws-proxy/console", nil)
			q := req.URL.Query()
			q.Set("token", tt.token)
			req.URL.RawQuery = q.Encode()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			// Must NOT be rejected by token validation
			assert.NotEqual(t, http.StatusBadRequest, w.Code,
				"valid token=%q should not return 400", tt.token)
		})
	}
}
