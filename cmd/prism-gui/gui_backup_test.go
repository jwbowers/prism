package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Backup represents a backup/snapshot of an instance
type Backup struct {
	ID         string  `json:"id"`
	InstanceID string  `json:"instance_id"`
	Name       string  `json:"name"`
	CreatedAt  string  `json:"created_at"`
	SizeGB     int     `json:"size_gb"`
	Status     string  `json:"status"` // "creating", "available", "deleting"
	Type       string  `json:"type"`   // "full", "incremental"
	CostGB     float64 `json:"cost_gb,omitempty"`
}

// BackupRequest represents a backup creation request
type BackupRequest struct {
	InstanceID  string `json:"instance_id"`
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"` // "full" or "incremental"
	Description string `json:"description,omitempty"`
}

// RestoreRequest represents a backup restore request
type RestoreRequest struct {
	BackupID     string `json:"backup_id"`
	InstanceName string `json:"instance_name,omitempty"`
}

// TestListBackups tests listing backups
func TestListBackups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/backups" {
			t.Errorf("Expected /api/v1/backups, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		backups := []Backup{
			{
				ID:         "backup-123",
				InstanceID: "i-1234567890abcdef0",
				Name:       "my-ml-research-backup",
				CreatedAt:  "2025-11-15T10:00:00Z",
				SizeGB:     50,
				Status:     "available",
				Type:       "full",
				CostGB:     0.05,
			},
			{
				ID:         "backup-456",
				InstanceID: "i-1234567890abcdef0",
				Name:       "my-ml-research-incremental",
				CreatedAt:  "2025-11-16T10:00:00Z",
				SizeGB:     10,
				Status:     "available",
				Type:       "incremental",
				CostGB:     0.05,
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(backups)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/backups")
	if err != nil {
		t.Fatalf("ListBackups request failed: %v", err)
	}
	defer resp.Body.Close()

	var backups []Backup
	if err := json.NewDecoder(resp.Body).Decode(&backups); err != nil {
		t.Fatalf("Failed to decode backups: %v", err)
	}

	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}

	// Verify full backup
	if backups[0].Type != "full" || backups[0].SizeGB != 50 {
		t.Errorf("Expected full backup, got %+v", backups[0])
	}

	// Verify incremental backup
	if backups[1].Type != "incremental" || backups[1].SizeGB != 10 {
		t.Errorf("Expected incremental backup, got %+v", backups[1])
	}
}

// TestGetBackup tests getting a specific backup
func TestGetBackup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/api/v1/backups/") {
			t.Errorf("Expected /api/v1/backups/*, got %s", r.URL.Path)
		}

		backup := Backup{
			ID:         "backup-test",
			InstanceID: "i-test123",
			Name:       "test-backup",
			CreatedAt:  "2025-11-16T12:00:00Z",
			SizeGB:     75,
			Status:     "available",
			Type:       "full",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(backup)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/backups/backup-test")
	if err != nil {
		t.Fatalf("GetBackup request failed: %v", err)
	}
	defer resp.Body.Close()

	var backup Backup
	if err := json.NewDecoder(resp.Body).Decode(&backup); err != nil {
		t.Fatalf("Failed to decode backup: %v", err)
	}

	if backup.SizeGB != 75 {
		t.Errorf("Expected size 75GB, got %d", backup.SizeGB)
	}
	if backup.Type != "full" {
		t.Errorf("Expected type 'full', got '%s'", backup.Type)
	}
}

// TestCreateBackup tests backup creation
func TestCreateBackup(t *testing.T) {
	var receivedRequest BackupRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		backup := Backup{
			ID:         "backup-new123",
			InstanceID: receivedRequest.InstanceID,
			Name:       receivedRequest.Name,
			CreatedAt:  time.Now().Format(time.RFC3339),
			Status:     "creating",
			Type:       receivedRequest.Type,
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(backup)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	backupReq := BackupRequest{
		InstanceID:  "i-test123",
		Name:        "test-backup",
		Type:        "full",
		Description: "Test backup for instance",
	}

	reqBody, _ := json.Marshal(backupReq)
	resp, err := service.client.Post(
		server.URL+"/api/v1/backups",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("CreateBackup request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// Verify request
	if receivedRequest.InstanceID != "i-test123" {
		t.Errorf("Expected instance i-test123, got %s", receivedRequest.InstanceID)
	}

	var backup Backup
	if err := json.NewDecoder(resp.Body).Decode(&backup); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if backup.Status != "creating" {
		t.Errorf("Expected status 'creating', got '%s'", backup.Status)
	}
}

// TestRestoreBackup tests backup restore
func TestRestoreBackup(t *testing.T) {
	restoredBackup := ""
	newInstanceName := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/restore") {
			t.Errorf("Expected path with /restore, got %s", r.URL.Path)
		}

		// Extract backup ID from path
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "backups" && i+1 < len(parts) {
				restoredBackup = parts[i+1]
				break
			}
		}

		var req RestoreRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
			newInstanceName = req.InstanceName
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":      "restoring",
			"instance_id": "i-restored123",
		})
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	restoreReq := RestoreRequest{
		BackupID:     "backup-123",
		InstanceName: "restored-instance",
	}

	reqBody, _ := json.Marshal(restoreReq)
	resp, err := service.client.Post(
		server.URL+"/api/v1/backups/backup-123/restore",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("RestoreBackup request failed: %v", err)
	}
	defer resp.Body.Close()

	if restoredBackup != "backup-123" {
		t.Errorf("Expected to restore backup-123, got %s", restoredBackup)
	}
	if newInstanceName != "restored-instance" {
		t.Errorf("Expected instance name 'restored-instance', got '%s'", newInstanceName)
	}
}

// TestDeleteBackup tests backup deletion
func TestDeleteBackup(t *testing.T) {
	deletedBackup := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		parts := strings.Split(r.URL.Path, "/")
		deletedBackup = parts[len(parts)-1]

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	req, _ := http.NewRequest(
		http.MethodDelete,
		server.URL+"/api/v1/backups/backup-old",
		nil,
	)

	resp, err := service.client.Do(req)
	if err != nil {
		t.Fatalf("DeleteBackup request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	if deletedBackup != "backup-old" {
		t.Errorf("Expected to delete 'backup-old', got '%s'", deletedBackup)
	}
}

// TestBackupValidation tests backup validation
func TestBackupValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req BackupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
			return
		}

		// Validate instance ID
		if req.InstanceID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "instance ID required"})
			return
		}

		// Validate backup type
		if req.Type != "" && req.Type != "full" && req.Type != "incremental" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid backup type"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Backup{Name: req.Name})
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Test missing instance ID
	t.Run("MissingInstanceID", func(t *testing.T) {
		reqBody, _ := json.Marshal(BackupRequest{Name: "test"})
		resp, err := service.client.Post(
			server.URL+"/api/v1/backups",
			"application/json",
			strings.NewReader(string(reqBody)),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	// Test invalid type
	t.Run("InvalidType", func(t *testing.T) {
		reqBody, _ := json.Marshal(BackupRequest{
			InstanceID: "i-test",
			Name:       "test",
			Type:       "invalid",
		})
		resp, err := service.client.Post(
			server.URL+"/api/v1/backups",
			"application/json",
			strings.NewReader(string(reqBody)),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

// TestIncrementalBackup tests incremental backup creation
func TestIncrementalBackup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req BackupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// For incremental backup, check if there's a previous backup
		if req.Type == "incremental" {
			// Simulate checking for previous backup
			backup := Backup{
				ID:         "backup-incremental",
				InstanceID: req.InstanceID,
				Name:       req.Name,
				CreatedAt:  time.Now().Format(time.RFC3339),
				Status:     "creating",
				Type:       "incremental",
				SizeGB:     5, // Smaller than full backup
			}

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(backup)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "no previous backup found for incremental",
			})
		}
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	backupReq := BackupRequest{
		InstanceID: "i-test123",
		Name:       "incremental-backup",
		Type:       "incremental",
	}

	reqBody, _ := json.Marshal(backupReq)
	resp, err := service.client.Post(
		server.URL+"/api/v1/backups",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("Incremental backup request failed: %v", err)
	}
	defer resp.Body.Close()

	var backup Backup
	if err := json.NewDecoder(resp.Body).Decode(&backup); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if backup.Type != "incremental" {
		t.Errorf("Expected incremental type, got '%s'", backup.Type)
	}
	if backup.SizeGB >= 50 {
		t.Errorf("Expected smaller size for incremental, got %d", backup.SizeGB)
	}
}

// TestBackupErrorHandling tests error handling for backup operations
func TestBackupErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/backups/backup-notfound":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "backup not found"})

		case "/api/v1/backups/backup-creating/restore":
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "backup still creating"})

		case "/api/v1/backups":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "instance not found"})
			}

		default:
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		}
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Test not found
	t.Run("NotFound", func(t *testing.T) {
		resp, err := service.client.Get(server.URL + "/api/v1/backups/backup-notfound")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	// Test restore conflict (backup still creating)
	t.Run("RestoreConflict", func(t *testing.T) {
		resp, err := service.client.Post(
			server.URL+"/api/v1/backups/backup-creating/restore",
			"application/json",
			nil,
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp.StatusCode)
		}
	})

	// Test instance not found for backup
	t.Run("InstanceNotFound", func(t *testing.T) {
		reqBody, _ := json.Marshal(BackupRequest{
			InstanceID: "i-notfound",
			Name:       "test",
		})
		resp, err := service.client.Post(
			server.URL+"/api/v1/backups",
			"application/json",
			strings.NewReader(string(reqBody)),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

// TestCloneFromBackup tests creating a new instance from a backup (clone)
func TestCloneFromBackup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/clone") {
			t.Errorf("Expected path with /clone, got %s", r.URL.Path)
		}

		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":        "cloning",
			"instance_id":   "i-cloned123",
			"instance_name": req["instance_name"],
		})
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	reqBody, _ := json.Marshal(map[string]string{
		"instance_name": "cloned-instance",
	})

	resp, err := service.client.Post(
		server.URL+"/api/v1/backups/backup-123/clone",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("Clone request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "cloning" {
		t.Errorf("Expected status 'cloning', got '%s'", result["status"])
	}
	if result["instance_name"] != "cloned-instance" {
		t.Errorf("Expected instance_name 'cloned-instance', got '%s'", result["instance_name"])
	}
}
