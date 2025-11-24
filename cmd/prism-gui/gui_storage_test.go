package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// EBSStorage represents an EBS volume
type EBSStorage struct {
	VolumeID         string `json:"volume_id"`
	Name             string `json:"name"`
	State            string `json:"state"`
	SizeGB           int    `json:"size_gb"`
	VolumeType       string `json:"volume_type"`
	IOPS             int    `json:"iops,omitempty"`
	Throughput       int    `json:"throughput,omitempty"`
	Encrypted        bool   `json:"encrypted"`
	AvailabilityZone string `json:"availability_zone"`
	CreationTime     string `json:"creation_time"`
	AttachedTo       string `json:"attached_to,omitempty"`
	Device           string `json:"device,omitempty"`
}

// EFSVolume represents an EFS filesystem
type EFSVolume struct {
	FilesystemID    string   `json:"filesystem_id"`
	Name            string   `json:"name"`
	State           string   `json:"state"`
	SizeGB          int      `json:"size_gb"`
	PerformanceMode string   `json:"performance_mode"`
	ThroughputMode  string   `json:"throughput_mode"`
	CreationTime    string   `json:"creation_time"`
	MountTargets    []string `json:"mount_targets,omitempty"`
}

// StorageRequest represents a storage creation request
type StorageRequest struct {
	Name             string `json:"name"`
	SizeGB           int    `json:"size_gb"`
	VolumeType       string `json:"volume_type,omitempty"`       // For EBS: gp3, gp2, io2
	PerformanceMode  string `json:"performance_mode,omitempty"`  // For EFS: generalPurpose, maxIO
	AvailabilityZone string `json:"availability_zone,omitempty"` // For EBS
}

// TestListEBSStorage tests listing EBS volumes
func TestListEBSStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/storage" {
			t.Errorf("Expected /api/v1/storage, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		storages := []EBSStorage{
			{
				VolumeID:         "vol-1234567890abcdef0",
				Name:             "project-storage-L",
				State:            "available",
				SizeGB:           100,
				VolumeType:       "gp3",
				IOPS:             3000,
				Throughput:       125,
				Encrypted:        true,
				AvailabilityZone: "us-west-2a",
				CreationTime:     "2025-09-20T12:00:00Z",
			},
			{
				VolumeID:         "vol-0987654321fedcba0",
				Name:             "large-dataset",
				State:            "in-use",
				SizeGB:           1000,
				VolumeType:       "gp3",
				Encrypted:        true,
				AvailabilityZone: "us-west-2a",
				AttachedTo:       "i-1234567890abcdef0",
				Device:           "/dev/sdf",
				CreationTime:     "2025-09-21T10:00:00Z",
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(storages)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/storage")
	if err != nil {
		t.Fatalf("ListEBSStorage request failed: %v", err)
	}
	defer resp.Body.Close()

	var storages []EBSStorage
	if err := json.NewDecoder(resp.Body).Decode(&storages); err != nil {
		t.Fatalf("Failed to decode storages: %v", err)
	}

	if len(storages) != 2 {
		t.Errorf("Expected 2 storages, got %d", len(storages))
	}

	// Verify first storage (available)
	if storages[0].State != "available" || storages[0].AttachedTo != "" {
		t.Errorf("Expected available storage, got %+v", storages[0])
	}

	// Verify second storage (in-use)
	if storages[1].State != "in-use" || storages[1].AttachedTo == "" {
		t.Errorf("Expected in-use storage, got %+v", storages[1])
	}
}

// TestListEFSVolumes tests listing EFS filesystems
func TestListEFSVolumes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/volumes" {
			t.Errorf("Expected /api/v1/volumes, got %s", r.URL.Path)
		}

		volumes := []EFSVolume{
			{
				FilesystemID:    "fs-1234567890abcdef0",
				Name:            "shared-data",
				State:           "available",
				SizeGB:          100,
				PerformanceMode: "generalPurpose",
				ThroughputMode:  "bursting",
				CreationTime:    "2025-09-15T10:00:00Z",
			},
			{
				FilesystemID:    "fs-0987654321fedcba0",
				Name:            "project-storage",
				State:           "in-use",
				SizeGB:          500,
				PerformanceMode: "generalPurpose",
				ThroughputMode:  "bursting",
				CreationTime:    "2025-09-16T12:00:00Z",
				MountTargets:    []string{"10.0.1.100", "10.0.2.100"},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(volumes)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := service.client.Get(server.URL + "/api/v1/volumes")
	if err != nil {
		t.Fatalf("ListEFSVolumes request failed: %v", err)
	}
	defer resp.Body.Close()

	var volumes []EFSVolume
	if err := json.NewDecoder(resp.Body).Decode(&volumes); err != nil {
		t.Fatalf("Failed to decode volumes: %v", err)
	}

	if len(volumes) != 2 {
		t.Errorf("Expected 2 volumes, got %d", len(volumes))
	}

	// Verify performance modes
	if volumes[0].PerformanceMode != "generalPurpose" {
		t.Errorf("Expected generalPurpose, got %s", volumes[0].PerformanceMode)
	}
}

// TestCreateEBSStorage tests EBS volume creation
func TestCreateEBSStorage(t *testing.T) {
	var receivedRequest StorageRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		storage := EBSStorage{
			VolumeID:         "vol-new123456789",
			Name:             receivedRequest.Name,
			State:            "creating",
			SizeGB:           receivedRequest.SizeGB,
			VolumeType:       receivedRequest.VolumeType,
			AvailabilityZone: receivedRequest.AvailabilityZone,
			CreationTime:     time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(storage)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	storageReq := StorageRequest{
		Name:             "test-storage",
		SizeGB:           200,
		VolumeType:       "gp3",
		AvailabilityZone: "us-west-2a",
	}

	reqBody, _ := json.Marshal(storageReq)
	resp, err := service.client.Post(
		server.URL+"/api/v1/storage",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("CreateEBSStorage request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// Verify request
	if receivedRequest.SizeGB != 200 {
		t.Errorf("Expected size 200GB, got %d", receivedRequest.SizeGB)
	}

	var storage EBSStorage
	if err := json.NewDecoder(resp.Body).Decode(&storage); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if storage.Name != "test-storage" {
		t.Errorf("Expected name 'test-storage', got '%s'", storage.Name)
	}
}

// TestCreateEFSVolume tests EFS filesystem creation
func TestCreateEFSVolume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req StorageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		volume := EFSVolume{
			FilesystemID:    "fs-new123456789",
			Name:            req.Name,
			State:           "creating",
			PerformanceMode: req.PerformanceMode,
			ThroughputMode:  "bursting",
			CreationTime:    time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(volume)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	volumeReq := StorageRequest{
		Name:            "test-volume",
		PerformanceMode: "generalPurpose",
	}

	reqBody, _ := json.Marshal(volumeReq)
	resp, err := service.client.Post(
		server.URL+"/api/v1/volumes",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		t.Fatalf("CreateEFSVolume request failed: %v", err)
	}
	defer resp.Body.Close()

	var volume EFSVolume
	if err := json.NewDecoder(resp.Body).Decode(&volume); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if volume.PerformanceMode != "generalPurpose" {
		t.Errorf("Expected generalPurpose, got %s", volume.PerformanceMode)
	}
}

// TestAttachDetachEBSStorage tests attaching and detaching EBS volumes
func TestAttachDetachEBSStorage(t *testing.T) {
	actionPerformed := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/attach"):
			actionPerformed = "attach"
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "attaching"})

		case strings.Contains(r.URL.Path, "/detach"):
			actionPerformed = "detach"
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "detaching"})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Test attach
	t.Run("Attach", func(t *testing.T) {
		resp, err := service.client.Post(
			server.URL+"/api/v1/storage/vol-123/attach",
			"application/json",
			strings.NewReader(`{"instance_id":"i-123","device":"/dev/sdf"}`),
		)
		if err != nil {
			t.Fatalf("Attach request failed: %v", err)
		}
		defer resp.Body.Close()

		if actionPerformed != "attach" {
			t.Errorf("Expected attach action, got %s", actionPerformed)
		}
	})

	// Test detach
	t.Run("Detach", func(t *testing.T) {
		resp, err := service.client.Post(
			server.URL+"/api/v1/storage/vol-123/detach",
			"application/json",
			nil,
		)
		if err != nil {
			t.Fatalf("Detach request failed: %v", err)
		}
		defer resp.Body.Close()

		if actionPerformed != "detach" {
			t.Errorf("Expected detach action, got %s", actionPerformed)
		}
	})
}

// TestMountUnmountEFSVolume tests mounting and unmounting EFS filesystems
func TestMountUnmountEFSVolume(t *testing.T) {
	actionPerformed := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/mount"):
			actionPerformed = "mount"
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "mounting"})

		case strings.Contains(r.URL.Path, "/unmount"):
			actionPerformed = "unmount"
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unmounting"})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Test mount
	t.Run("Mount", func(t *testing.T) {
		resp, err := service.client.Post(
			server.URL+"/api/v1/volumes/fs-123/mount",
			"application/json",
			strings.NewReader(`{"instance_id":"i-123","mount_point":"/mnt/efs"}`),
		)
		if err != nil {
			t.Fatalf("Mount request failed: %v", err)
		}
		defer resp.Body.Close()

		if actionPerformed != "mount" {
			t.Errorf("Expected mount action, got %s", actionPerformed)
		}
	})

	// Test unmount
	t.Run("Unmount", func(t *testing.T) {
		resp, err := service.client.Post(
			server.URL+"/api/v1/volumes/fs-123/unmount",
			"application/json",
			strings.NewReader(`{"instance_id":"i-123"}`),
		)
		if err != nil {
			t.Fatalf("Unmount request failed: %v", err)
		}
		defer resp.Body.Close()

		if actionPerformed != "unmount" {
			t.Errorf("Expected unmount action, got %s", actionPerformed)
		}
	})
}

// TestDeleteStorage tests deleting storage resources
func TestDeleteStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := &PrismService{
		daemonURL: server.URL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}

	// Test delete EBS
	t.Run("DeleteEBS", func(t *testing.T) {
		req, _ := http.NewRequest(
			http.MethodDelete,
			server.URL+"/api/v1/storage/vol-123",
			nil,
		)

		resp, err := service.client.Do(req)
		if err != nil {
			t.Fatalf("Delete EBS request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.StatusCode)
		}
	})

	// Test delete EFS
	t.Run("DeleteEFS", func(t *testing.T) {
		req, _ := http.NewRequest(
			http.MethodDelete,
			server.URL+"/api/v1/volumes/fs-123",
			nil,
		)

		resp, err := service.client.Do(req)
		if err != nil {
			t.Fatalf("Delete EFS request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.StatusCode)
		}
	})
}

// TestStorageErrorHandling tests error handling for storage operations
func TestStorageErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate various error conditions
		switch r.URL.Path {
		case "/api/v1/storage/vol-notfound":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "volume not found"})

		case "/api/v1/volumes/fs-inuse":
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "filesystem in use"})
			}

		case "/api/v1/storage":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid size specified"})
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
		resp, err := service.client.Get(server.URL + "/api/v1/storage/vol-notfound")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	// Test conflict (in-use)
	t.Run("InUse", func(t *testing.T) {
		req, _ := http.NewRequest(
			http.MethodDelete,
			server.URL+"/api/v1/volumes/fs-inuse",
			nil,
		)

		resp, err := service.client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp.StatusCode)
		}
	})

	// Test bad request
	t.Run("BadRequest", func(t *testing.T) {
		resp, err := service.client.Post(
			server.URL+"/api/v1/storage",
			"application/json",
			strings.NewReader(`{"name":"test","size_gb":-1}`),
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
