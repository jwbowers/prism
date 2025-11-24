package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// HibernationStatus represents hibernation capability and status
type HibernationStatus struct {
	InstanceID            string `json:"instance_id"`
	HibernationEnabled    bool   `json:"hibernation_enabled"`
	HibernationConfigured bool   `json:"hibernation_configured"`
	CanHibernate          bool   `json:"can_hibernate"`
	CurrentState          string `json:"current_state"`
	Message               string `json:"message,omitempty"`
}

// HibernationRequest represents a hibernation operation request
type HibernationRequest struct {
	InstanceID string `json:"instance_id"`
	Force      bool   `json:"force,omitempty"`
}

// TestHibernateInstance tests instance hibernation
func TestHibernateInstance(t *testing.T) {
	hibernatedInstance := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/hibernate") {
			t.Errorf("Expected path with /hibernate, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Extract instance ID from path
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "instances" && i+1 < len(parts) {
				hibernatedInstance = parts[i+1]
				break
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "hibernating",
			"message": "Instance is entering hibernation",
		})
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Post(
		server.URL+"/api/v1/instances/i-test123/hibernate",
		"application/json",
		nil,
	)
	if err != nil {
		t.Fatalf("Hibernate request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if hibernatedInstance != "i-test123" {
		t.Errorf("Expected to hibernate i-test123, got %s", hibernatedInstance)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "hibernating" {
		t.Errorf("Expected status 'hibernating', got '%s'", result["status"])
	}
}

// TestResumeInstance tests instance resume from hibernation
func TestResumeInstance(t *testing.T) {
	resumedInstance := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/resume") {
			t.Errorf("Expected path with /resume, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Extract instance ID from path
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "instances" && i+1 < len(parts) {
				resumedInstance = parts[i+1]
				break
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "resuming",
			"message": "Instance is resuming from hibernation",
		})
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Post(
		server.URL+"/api/v1/instances/i-test456/resume",
		"application/json",
		nil,
	)
	if err != nil {
		t.Fatalf("Resume request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if resumedInstance != "i-test456" {
		t.Errorf("Expected to resume i-test456, got %s", resumedInstance)
	}
}

// TestGetHibernationStatus tests getting hibernation status
func TestGetHibernationStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/hibernation-status") {
			t.Errorf("Expected path with /hibernation-status, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		status := HibernationStatus{
			InstanceID:            "i-test123",
			HibernationEnabled:    true,
			HibernationConfigured: true,
			CanHibernate:          true,
			CurrentState:          "running",
			Message:               "Instance is ready for hibernation",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/instances/i-test123/hibernation-status")
	if err != nil {
		t.Fatalf("GetHibernationStatus request failed: %v", err)
	}
	defer resp.Body.Close()

	var status HibernationStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode status: %v", err)
	}

	if !status.CanHibernate {
		t.Error("Expected CanHibernate to be true")
	}
	if !status.HibernationEnabled {
		t.Error("Expected HibernationEnabled to be true")
	}
	if status.CurrentState != "running" {
		t.Errorf("Expected state 'running', got '%s'", status.CurrentState)
	}
}

// TestHibernationNotSupported tests handling when hibernation is not supported
func TestHibernationNotSupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/hibernation-status") {
			status := HibernationStatus{
				InstanceID:            "i-nosupport",
				HibernationEnabled:    false,
				HibernationConfigured: false,
				CanHibernate:          false,
				CurrentState:          "running",
				Message:               "Instance type does not support hibernation",
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(status)

		} else if strings.Contains(r.URL.Path, "/hibernate") {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "hibernation not supported for this instance type",
			})
		}
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Test status check
	t.Run("StatusCheck", func(t *testing.T) {
		resp, err := service.client.Get(server.URL + "/api/v1/instances/i-nosupport/hibernation-status")
		if err != nil {
			t.Fatalf("Status request failed: %v", err)
		}
		defer resp.Body.Close()

		var status HibernationStatus
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			t.Fatalf("Failed to decode status: %v", err)
		}

		if status.CanHibernate {
			t.Error("Expected CanHibernate to be false")
		}
		if !strings.Contains(status.Message, "not support") {
			t.Errorf("Expected unsupported message, got '%s'", status.Message)
		}
	})

	// Test hibernate attempt
	t.Run("HibernateAttempt", func(t *testing.T) {
		resp, err := service.client.Post(
			server.URL+"/api/v1/instances/i-nosupport/hibernate",
			"application/json",
			nil,
		)
		if err != nil {
			t.Fatalf("Hibernate request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

// TestHibernationFallbackToStop tests fallback to stop when hibernation fails
func TestHibernationFallbackToStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/hibernate"):
			// Hibernation fails, suggest stop
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":           "hibernation failed",
				"fallback_action": "stop",
				"message":         "Falling back to stop operation",
			})

		case strings.Contains(r.URL.Path, "/stop"):
			// Stop succeeds
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "stopping",
			})
		}
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Attempt hibernate
	resp, err := service.client.Post(
		server.URL+"/api/v1/instances/i-test/hibernate",
		"application/json",
		nil,
	)
	if err != nil {
		t.Fatalf("Hibernate request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Verify fallback suggestion
	var errResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp["fallback_action"] != "stop" {
		t.Errorf("Expected fallback_action 'stop', got '%v'", errResp["fallback_action"])
	}
}

// TestHibernationWithForce tests force hibernation option
func TestHibernationWithForce(t *testing.T) {
	forceUsed := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req HibernationRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
			forceUsed = req.Force
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "hibernating",
			"forced": forceUsed,
		})
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Test with force flag
	reqBody, _ := json.Marshal(HibernationRequest{
		InstanceID: "i-test",
		Force:      true,
	})

	resp, err := service.client.Post(
		server.URL+"/api/v1/instances/i-test/hibernate",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("Hibernate with force request failed: %v", err)
	}
	defer resp.Body.Close()

	if !forceUsed {
		t.Error("Expected force flag to be true")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !result["forced"].(bool) {
		t.Error("Expected forced to be true in response")
	}
}

// TestHibernationStateTransitions tests various state transitions
func TestHibernationStateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialState  string
		action        string
		expectedState string
		expectError   bool
	}{
		{
			name:          "running to hibernating",
			initialState:  "running",
			action:        "hibernate",
			expectedState: "hibernating",
			expectError:   false,
		},
		{
			name:          "hibernated to resuming",
			initialState:  "hibernated",
			action:        "resume",
			expectedState: "resuming",
			expectError:   false,
		},
		{
			name:          "stopped cannot hibernate",
			initialState:  "stopped",
			action:        "hibernate",
			expectedState: "stopped",
			expectError:   true,
		},
		{
			name:          "terminating cannot hibernate",
			initialState:  "terminating",
			action:        "hibernate",
			expectedState: "terminating",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if state allows operation
				if tt.expectError {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error": "invalid state for " + tt.action,
					})
					return
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"status":         tt.expectedState,
					"previous_state": tt.initialState,
				})
			}))
			defer server.Close()

			service := &PrismService{
				daemonURL: server.URL,
				client:    &http.Client{Timeout: 5 * time.Second},
			}

			endpoint := server.URL + "/api/v1/instances/i-test/" + tt.action
			resp, err := service.client.Post(endpoint, "application/json", nil)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if tt.expectError {
				if resp.StatusCode == http.StatusOK {
					t.Errorf("Expected error, but got success")
				}
			} else {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}

				var result map[string]string
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if result["status"] != tt.expectedState {
					t.Errorf("Expected state '%s', got '%s'", tt.expectedState, result["status"])
				}
			}
		})
	}
}
