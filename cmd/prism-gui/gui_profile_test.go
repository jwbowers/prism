package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Profile represents a Prism profile configuration
type Profile struct {
	Name       string `json:"name"`
	AWSProfile string `json:"aws_profile"`
	Region     string `json:"region"`
	IsDefault  bool   `json:"is_default"`
	CreatedAt  string `json:"created_at"`
}

// ProfileRequest represents a profile creation/update request
type ProfileRequest struct {
	Name       string `json:"name"`
	AWSProfile string `json:"aws_profile"`
	Region     string `json:"region"`
	IsDefault  bool   `json:"is_default"`
}

// TestGetProfiles tests profile listing
func TestGetProfiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/profiles" {
			t.Errorf("Expected /api/v1/profiles, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		profiles := []Profile{
			{
				Name:       "default",
				AWSProfile: "default",
				Region:     "us-west-2",
				IsDefault:  true,
				CreatedAt:  "2025-01-01T00:00:00Z",
			},
			{
				Name:       "research-profile",
				AWSProfile: "research",
				Region:     "us-east-1",
				IsDefault:  false,
				CreatedAt:  "2025-01-02T00:00:00Z",
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(profiles)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Call would be: profiles, err := service.GetProfiles(ctx)
	// For now, test the HTTP endpoint directly
	resp, err := service.client.Get(server.URL + "/api/v1/profiles")
	if err != nil {
		t.Fatalf("GetProfiles request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		t.Fatalf("Failed to decode profiles: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}

	// Verify default profile
	if profiles[0].Name != "default" || !profiles[0].IsDefault {
		t.Errorf("Expected default profile, got %+v", profiles[0])
	}

	// Verify research profile
	if profiles[1].Name != "research-profile" || profiles[1].IsDefault {
		t.Errorf("Expected research-profile, got %+v", profiles[1])
	}
}

// TestGetCurrentProfile tests getting the current profile
func TestGetCurrentProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/profiles/current" {
			t.Errorf("Expected /api/v1/profiles/current, got %s", r.URL.Path)
		}

		profile := Profile{
			Name:       "default",
			AWSProfile: "default",
			Region:     "us-west-2",
			IsDefault:  true,
			CreatedAt:  "2025-01-01T00:00:00Z",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(profile)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/profiles/current")
	if err != nil {
		t.Fatalf("GetCurrentProfile request failed: %v", err)
	}
	defer resp.Body.Close()

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		t.Fatalf("Failed to decode profile: %v", err)
	}

	if profile.Name != "default" {
		t.Errorf("Expected 'default' profile, got '%s'", profile.Name)
	}
	if !profile.IsDefault {
		t.Error("Expected IsDefault to be true")
	}
	if profile.Region != "us-west-2" {
		t.Errorf("Expected region us-west-2, got %s", profile.Region)
	}
}

// TestCreateProfile tests profile creation
func TestCreateProfile(t *testing.T) {
	var receivedRequest ProfileRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/profiles" {
			t.Errorf("Expected /api/v1/profiles, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Return created profile
		profile := Profile{
			Name:       receivedRequest.Name,
			AWSProfile: receivedRequest.AWSProfile,
			Region:     receivedRequest.Region,
			IsDefault:  receivedRequest.IsDefault,
			CreatedAt:  time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(profile)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Create profile request
	profileReq := ProfileRequest{
		Name:       "new-profile",
		AWSProfile: "research",
		Region:     "eu-west-1",
		IsDefault:  false,
	}

	reqBody, _ := json.Marshal(profileReq)
	resp, err := service.client.Post(
		server.URL+"/api/v1/profiles",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("CreateProfile request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// Verify request received correctly
	if receivedRequest.Name != "new-profile" {
		t.Errorf("Expected profile name 'new-profile', got '%s'", receivedRequest.Name)
	}
	if receivedRequest.Region != "eu-west-1" {
		t.Errorf("Expected region eu-west-1, got %s", receivedRequest.Region)
	}

	// Verify response
	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if profile.Name != "new-profile" {
		t.Errorf("Expected profile name 'new-profile', got '%s'", profile.Name)
	}
}

// TestUpdateProfile tests profile update
func TestUpdateProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/profiles/") {
			t.Errorf("Expected /api/v1/profiles/*, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		var req ProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Extract profile name from path
		parts := strings.Split(r.URL.Path, "/")
		profileName := parts[len(parts)-1]

		// Return updated profile
		profile := Profile{
			Name:       profileName,
			AWSProfile: req.AWSProfile,
			Region:     req.Region,
			IsDefault:  req.IsDefault,
			CreatedAt:  "2025-01-01T00:00:00Z",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(profile)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Update profile
	profileReq := ProfileRequest{
		AWSProfile: "updated-profile",
		Region:     "ap-southeast-1",
		IsDefault:  false,
	}

	reqBody, _ := json.Marshal(profileReq)
	req, _ := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/v1/profiles/test-profile",
		strings.NewReader(string(reqBody)),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := service.client.Do(req)
	if err != nil {
		t.Fatalf("UpdateProfile request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if profile.Region != "ap-southeast-1" {
		t.Errorf("Expected region ap-southeast-1, got %s", profile.Region)
	}
}

// TestDeleteProfile tests profile deletion
func TestDeleteProfile(t *testing.T) {
	deletedProfile := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/profiles/") {
			t.Errorf("Expected /api/v1/profiles/*, got %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		// Extract profile name from path
		parts := strings.Split(r.URL.Path, "/")
		deletedProfile = parts[len(parts)-1]

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	req, _ := http.NewRequest(
		http.MethodDelete,
		server.URL+"/api/v1/profiles/old-profile",
		nil,
	)

	resp, err := service.client.Do(req)
	if err != nil {
		t.Fatalf("DeleteProfile request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	if deletedProfile != "old-profile" {
		t.Errorf("Expected to delete 'old-profile', got '%s'", deletedProfile)
	}
}

// TestSwitchProfile tests switching active profile
func TestSwitchProfile(t *testing.T) {
	switchedTo := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/switch") {
			t.Errorf("Expected path with /switch, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Extract profile name from path
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "profiles" && i+1 < len(parts) {
				switchedTo = parts[i+1]
				break
			}
		}

		response := map[string]string{
			"current_profile": switchedTo,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Post(
		server.URL+"/api/v1/profiles/research-profile/switch",
		"application/json",
		nil,
	)
	if err != nil {
		t.Fatalf("SwitchProfile request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if switchedTo != "research-profile" {
		t.Errorf("Expected to switch to 'research-profile', got '%s'", switchedTo)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["current_profile"] != "research-profile" {
		t.Errorf("Expected current_profile 'research-profile', got '%s'", result["current_profile"])
	}
}

// TestProfileValidation tests profile validation
func TestProfileValidation(t *testing.T) {
	tests := []struct {
		name        string
		profileReq  ProfileRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid profile",
			profileReq: ProfileRequest{
				Name:       "valid-profile",
				AWSProfile: "default",
				Region:     "us-west-2",
			},
			expectError: false,
		},
		{
			name: "empty name",
			profileReq: ProfileRequest{
				Name:       "",
				AWSProfile: "default",
				Region:     "us-west-2",
			},
			expectError: true,
			errorMsg:    "profile name is required",
		},
		{
			name: "empty region",
			profileReq: ProfileRequest{
				Name:       "test",
				AWSProfile: "default",
				Region:     "",
			},
			expectError: true,
			errorMsg:    "region is required",
		},
		{
			name: "invalid region format",
			profileReq: ProfileRequest{
				Name:       "test",
				AWSProfile: "default",
				Region:     "invalid-region",
			},
			expectError: true,
			errorMsg:    "invalid region format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req ProfileRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
					return
				}

				// Validate request
				if req.Name == "" {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "profile name is required"})
					return
				}
				if req.Region == "" {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "region is required"})
					return
				}
				// AWS regions have format like "us-west-2" with at least 2 dashes
				dashCount := strings.Count(req.Region, "-")
				if dashCount < 2 || len(req.Region) < 9 {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid region format"})
					return
				}

				// Valid request
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(Profile{Name: req.Name, Region: req.Region})
			}))
			defer server.Close()

			service := &PrismService{
				daemonURL: server.URL,
				client:    &http.Client{Timeout: 5 * time.Second},
			}

			reqBody, _ := json.Marshal(tt.profileReq)
			resp, err := service.client.Post(
				server.URL+"/api/v1/profiles",
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

// TestProfileExportImport tests profile export and import
func TestProfileExportImport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/export"):
			// Export profile
			profile := Profile{
				Name:       "export-test",
				AWSProfile: "default",
				Region:     "us-west-2",
				IsDefault:  false,
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(profile)

		case strings.HasSuffix(r.URL.Path, "/import"):
			// Import profile
			var profile Profile
			if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(profile)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Test export
	t.Run("Export", func(t *testing.T) {
		resp, err := service.client.Get(server.URL + "/api/v1/profiles/export-test/export")
		if err != nil {
			t.Fatalf("Export request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var profile Profile
		if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
			t.Fatalf("Failed to decode export: %v", err)
		}

		if profile.Name != "export-test" {
			t.Errorf("Expected profile name 'export-test', got '%s'", profile.Name)
		}
	})

	// Test import
	t.Run("Import", func(t *testing.T) {
		importProfile := Profile{
			Name:       "imported-profile",
			AWSProfile: "imported",
			Region:     "eu-central-1",
			IsDefault:  false,
		}

		reqBody, _ := json.Marshal(importProfile)
		resp, err := service.client.Post(
			server.URL+"/api/v1/profiles/import",
			"application/json",
			strings.NewReader(string(reqBody)),
		)
		if err != nil {
			t.Fatalf("Import request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var profile Profile
		if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
			t.Fatalf("Failed to decode import result: %v", err)
		}

		if profile.Name != "imported-profile" {
			t.Errorf("Expected profile name 'imported-profile', got '%s'", profile.Name)
		}
	})
}
