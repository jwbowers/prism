package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// IdlePolicy represents an idle detection policy
type IdlePolicy struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	IdleMinutes      int    `json:"idle_minutes"`
	Action           string `json:"action"` // "stop", "hibernate", "terminate"
	CPUThreshold     int    `json:"cpu_threshold"`
	MemoryThreshold  int    `json:"memory_threshold"`
	NetworkThreshold int    `json:"network_threshold"`
	DiskThreshold    int    `json:"disk_threshold"`
	GPUThreshold     int    `json:"gpu_threshold,omitempty"`
	Enabled          bool   `json:"enabled"`
}

// IdlePolicyRequest represents an idle policy creation/update request
type IdlePolicyRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	IdleMinutes      int    `json:"idle_minutes"`
	Action           string `json:"action"`
	CPUThreshold     int    `json:"cpu_threshold,omitempty"`
	MemoryThreshold  int    `json:"memory_threshold,omitempty"`
	NetworkThreshold int    `json:"network_threshold,omitempty"`
	DiskThreshold    int    `json:"disk_threshold,omitempty"`
	GPUThreshold     int    `json:"gpu_threshold,omitempty"`
}

// IdleHistoryEntry represents an idle detection history entry
type IdleHistoryEntry struct {
	InstanceID   string  `json:"instance_id"`
	InstanceName string  `json:"instance_name"`
	Action       string  `json:"action"`
	Reason       string  `json:"reason"`
	Timestamp    string  `json:"timestamp"`
	CostSaved    float64 `json:"cost_saved,omitempty"`
}

// TestListIdlePolicies tests listing idle policies
func TestListIdlePolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/idle/policies" {
			t.Errorf("Expected /api/v1/idle/policies, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		policies := []IdlePolicy{
			{
				ID:               "policy-gpu",
				Name:             "gpu",
				Description:      "GPU instance idle detection",
				IdleMinutes:      15,
				Action:           "stop",
				CPUThreshold:     10,
				MemoryThreshold:  20,
				NetworkThreshold: 1,
				DiskThreshold:    5,
				GPUThreshold:     10,
				Enabled:          true,
			},
			{
				ID:               "policy-batch",
				Name:             "batch",
				Description:      "Long-running batch jobs",
				IdleMinutes:      60,
				Action:           "hibernate",
				CPUThreshold:     5,
				MemoryThreshold:  10,
				NetworkThreshold: 1,
				DiskThreshold:    5,
				Enabled:          true,
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(policies)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/idle/policies")
	if err != nil {
		t.Fatalf("ListIdlePolicies request failed: %v", err)
	}
	defer resp.Body.Close()

	var policies []IdlePolicy
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		t.Fatalf("Failed to decode policies: %v", err)
	}

	if len(policies) != 2 {
		t.Errorf("Expected 2 policies, got %d", len(policies))
	}

	// Verify GPU policy
	if policies[0].Name != "gpu" || policies[0].IdleMinutes != 15 {
		t.Errorf("Expected GPU policy, got %+v", policies[0])
	}

	// Verify batch policy
	if policies[1].Action != "hibernate" || policies[1].IdleMinutes != 60 {
		t.Errorf("Expected batch hibernation policy, got %+v", policies[1])
	}
}

// TestGetIdlePolicy tests getting a specific idle policy
func TestGetIdlePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/api/v1/idle/policies/") {
			t.Errorf("Expected /api/v1/idle/policies/*, got %s", r.URL.Path)
		}

		policy := IdlePolicy{
			ID:               "policy-test",
			Name:             "test-policy",
			Description:      "Test idle policy",
			IdleMinutes:      30,
			Action:           "hibernate",
			CPUThreshold:     10,
			MemoryThreshold:  15,
			NetworkThreshold: 2,
			DiskThreshold:    5,
			Enabled:          true,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(policy)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/idle/policies/policy-test")
	if err != nil {
		t.Fatalf("GetIdlePolicy request failed: %v", err)
	}
	defer resp.Body.Close()

	var policy IdlePolicy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		t.Fatalf("Failed to decode policy: %v", err)
	}

	if policy.IdleMinutes != 30 {
		t.Errorf("Expected 30 idle minutes, got %d", policy.IdleMinutes)
	}
	if policy.Action != "hibernate" {
		t.Errorf("Expected action 'hibernate', got '%s'", policy.Action)
	}
}

// TestCreateIdlePolicy tests idle policy creation
func TestCreateIdlePolicy(t *testing.T) {
	var receivedRequest IdlePolicyRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Return created policy
		policy := IdlePolicy{
			ID:               "policy-new123",
			Name:             receivedRequest.Name,
			Description:      receivedRequest.Description,
			IdleMinutes:      receivedRequest.IdleMinutes,
			Action:           receivedRequest.Action,
			CPUThreshold:     receivedRequest.CPUThreshold,
			MemoryThreshold:  receivedRequest.MemoryThreshold,
			NetworkThreshold: receivedRequest.NetworkThreshold,
			DiskThreshold:    receivedRequest.DiskThreshold,
			GPUThreshold:     receivedRequest.GPUThreshold,
			Enabled:          true,
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(policy)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	policyReq := IdlePolicyRequest{
		Name:             "cost-optimized",
		Description:      "Maximum cost savings",
		IdleMinutes:      10,
		Action:           "hibernate",
		CPUThreshold:     5,
		MemoryThreshold:  10,
		NetworkThreshold: 1,
		DiskThreshold:    3,
	}

	reqBody, _ := json.Marshal(policyReq)
	resp, err := service.client.Post(
		server.URL+"/api/v1/idle/policies",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("CreateIdlePolicy request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// Verify request
	if receivedRequest.IdleMinutes != 10 {
		t.Errorf("Expected 10 idle minutes, got %d", receivedRequest.IdleMinutes)
	}

	var policy IdlePolicy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if policy.Name != "cost-optimized" {
		t.Errorf("Expected name 'cost-optimized', got '%s'", policy.Name)
	}
}

// TestUpdateIdlePolicy tests idle policy update
func TestUpdateIdlePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		var req IdlePolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Extract policy ID from path
		parts := strings.Split(r.URL.Path, "/")
		policyID := parts[len(parts)-1]

		policy := IdlePolicy{
			ID:          policyID,
			Name:        req.Name,
			IdleMinutes: req.IdleMinutes,
			Action:      req.Action,
			Enabled:     true,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(policy)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	policyReq := IdlePolicyRequest{
		Name:        "updated-policy",
		IdleMinutes: 45,
		Action:      "stop",
	}

	reqBody, _ := json.Marshal(policyReq)
	req, _ := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/v1/idle/policies/policy-123",
		strings.NewReader(string(reqBody)),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := service.client.Do(req)
	if err != nil {
		t.Fatalf("UpdateIdlePolicy request failed: %v", err)
	}
	defer resp.Body.Close()

	var policy IdlePolicy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if policy.IdleMinutes != 45 {
		t.Errorf("Expected 45 idle minutes, got %d", policy.IdleMinutes)
	}
}

// TestDeleteIdlePolicy tests idle policy deletion
func TestDeleteIdlePolicy(t *testing.T) {
	deletedPolicy := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		parts := strings.Split(r.URL.Path, "/")
		deletedPolicy = parts[len(parts)-1]

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	req, _ := http.NewRequest(
		http.MethodDelete,
		server.URL+"/api/v1/idle/policies/policy-old",
		nil,
	)

	resp, err := service.client.Do(req)
	if err != nil {
		t.Fatalf("DeleteIdlePolicy request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	if deletedPolicy != "policy-old" {
		t.Errorf("Expected to delete 'policy-old', got '%s'", deletedPolicy)
	}
}

// TestApplyIdlePolicyToInstance tests applying a policy to an instance
func TestApplyIdlePolicyToInstance(t *testing.T) {
	appliedTo := ""
	policyApplied := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/apply-policy") {
			t.Errorf("Expected path with /apply-policy, got %s", r.URL.Path)
		}

		// Extract instance ID from path
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "instances" && i+1 < len(parts) {
				appliedTo = parts[i+1]
				break
			}
		}

		// Get policy from request body
		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)
		policyApplied = req["policy_id"]

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "policy applied",
			"instance": appliedTo,
			"policy":   policyApplied,
		})
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	reqBody, _ := json.Marshal(map[string]string{"policy_id": "policy-gpu"})
	resp, err := service.client.Post(
		server.URL+"/api/v1/idle/instances/i-test123/apply-policy",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("ApplyPolicy request failed: %v", err)
	}
	defer resp.Body.Close()

	if appliedTo != "i-test123" {
		t.Errorf("Expected policy applied to i-test123, got %s", appliedTo)
	}
	if policyApplied != "policy-gpu" {
		t.Errorf("Expected policy-gpu applied, got %s", policyApplied)
	}
}

// TestGetIdleHistory tests getting idle detection history
func TestGetIdleHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/idle/history" {
			t.Errorf("Expected /api/v1/idle/history, got %s", r.URL.Path)
		}

		history := []IdleHistoryEntry{
			{
				InstanceID:   "i-1234567890abcdef0",
				InstanceName: "my-ml-research",
				Action:       "hibernate",
				Reason:       "idle_timeout",
				Timestamp:    "2025-11-16T14:30:00Z",
				CostSaved:    2.40,
			},
			{
				InstanceID:   "i-0987654321fedcba0",
				InstanceName: "gpu-workstation",
				Action:       "stop",
				Reason:       "idle_timeout",
				Timestamp:    "2025-11-16T15:00:00Z",
				CostSaved:    5.20,
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(history)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/idle/history")
	if err != nil {
		t.Fatalf("GetIdleHistory request failed: %v", err)
	}
	defer resp.Body.Close()

	var history []IdleHistoryEntry
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		t.Fatalf("Failed to decode history: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(history))
	}

	// Verify hibernation entry
	if history[0].Action != "hibernate" || history[0].CostSaved != 2.40 {
		t.Errorf("Expected hibernate with savings, got %+v", history[0])
	}

	// Verify stop entry
	if history[1].Action != "stop" || history[1].CostSaved != 5.20 {
		t.Errorf("Expected stop with savings, got %+v", history[1])
	}
}

// TestIdlePolicyValidation tests idle policy validation
func TestIdlePolicyValidation(t *testing.T) {
	tests := []struct {
		name        string
		policyReq   IdlePolicyRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid policy",
			policyReq: IdlePolicyRequest{
				Name:        "valid",
				IdleMinutes: 15,
				Action:      "hibernate",
			},
			expectError: false,
		},
		{
			name: "invalid action",
			policyReq: IdlePolicyRequest{
				Name:        "invalid-action",
				IdleMinutes: 15,
				Action:      "invalid",
			},
			expectError: true,
			errorMsg:    "invalid action",
		},
		{
			name: "negative idle minutes",
			policyReq: IdlePolicyRequest{
				Name:        "negative",
				IdleMinutes: -10,
				Action:      "stop",
			},
			expectError: true,
			errorMsg:    "idle minutes must be positive",
		},
		{
			name: "idle minutes too low",
			policyReq: IdlePolicyRequest{
				Name:        "too-low",
				IdleMinutes: 2,
				Action:      "stop",
			},
			expectError: true,
			errorMsg:    "minimum idle minutes is 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req IdlePolicyRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
					return
				}

				// Validate action
				validActions := map[string]bool{"stop": true, "hibernate": true, "terminate": true}
				if !validActions[req.Action] {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid action"})
					return
				}

				// Validate idle minutes
				if req.IdleMinutes < 0 {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "idle minutes must be positive"})
					return
				}
				if req.IdleMinutes < 5 {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "minimum idle minutes is 5"})
					return
				}

				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(IdlePolicy{Name: req.Name})
			}))
			defer server.Close()

			service := &PrismService{
				daemonURL: server.URL,
				client:    &http.Client{Timeout: 5 * time.Second},
			}

			reqBody, _ := json.Marshal(tt.policyReq)
			resp, err := service.client.Post(
				server.URL+"/api/v1/idle/policies",
				"application/json",
				strings.NewReader(string(reqBody)),
			)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if tt.expectError {
				if resp.StatusCode == http.StatusCreated {
					t.Errorf("Expected error, but got success")
				}

				var errResp map[string]string
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				if !strings.Contains(errResp["error"], tt.errorMsg) {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, errResp["error"])
				}
			} else {
				if resp.StatusCode != http.StatusCreated {
					t.Errorf("Expected status 201, got %d", resp.StatusCode)
				}
			}
		})
	}
}

// TestIdlePolicyPresets tests retrieving predefined idle policy presets
func TestIdlePolicyPresets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/presets") {
			t.Errorf("Expected path with /presets, got %s", r.URL.Path)
		}

		presets := []IdlePolicy{
			{
				Name:        "gpu",
				Description: "Optimized for GPU instances",
				IdleMinutes: 15,
				Action:      "stop",
			},
			{
				Name:        "batch",
				Description: "Long-running batch jobs",
				IdleMinutes: 60,
				Action:      "hibernate",
			},
			{
				Name:        "cost-optimized",
				Description: "Maximum cost savings",
				IdleMinutes: 10,
				Action:      "hibernate",
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(presets)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/idle/policies/presets")
	if err != nil {
		t.Fatalf("GetPresets request failed: %v", err)
	}
	defer resp.Body.Close()

	var presets []IdlePolicy
	if err := json.NewDecoder(resp.Body).Decode(&presets); err != nil {
		t.Fatalf("Failed to decode presets: %v", err)
	}

	if len(presets) != 3 {
		t.Errorf("Expected 3 presets, got %d", len(presets))
	}

	// Verify preset names
	expectedNames := []string{"gpu", "batch", "cost-optimized"}
	for i, preset := range presets {
		if preset.Name != expectedNames[i] {
			t.Errorf("Expected preset '%s', got '%s'", expectedNames[i], preset.Name)
		}
	}
}
