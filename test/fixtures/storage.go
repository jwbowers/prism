package fixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/types"
)

// CreateTestVolumeOptions contains options for creating a test EFS volume
type CreateTestVolumeOptions struct {
	Name            string
	PerformanceMode string
}

// CreateTestVolume creates a test EFS volume for integration tests
// The volume is automatically registered for cleanup via the registry
func CreateTestVolume(t *testing.T, registry *FixtureRegistry, opts CreateTestVolumeOptions) (*types.StorageVolume, error) {
	t.Helper()

	// Set defaults
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("test-volume-%d", time.Now().Unix())
	}
	if opts.PerformanceMode == "" {
		opts.PerformanceMode = "generalPurpose"
	}

	ctx := context.Background()

	// Create volume
	volumeReq := types.VolumeCreateRequest{
		Name:            opts.Name,
		PerformanceMode: opts.PerformanceMode,
	}

	t.Logf("Creating test EFS volume: %s (performance: %s)", opts.Name, opts.PerformanceMode)
	volume, err := registry.client.CreateVolume(ctx, volumeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	// Register for cleanup
	registry.Register("volume", opts.Name)
	t.Logf("EFS volume created: %s (FS ID: %s)", opts.Name, volume.FileSystemID)

	// Wait for volume to be available
	t.Logf("Waiting for volume %s to reach available state...", opts.Name)
	availableVolume, err := waitForVolumeState(ctx, registry.client, opts.Name, "available", 2*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("volume did not reach available state: %w", err)
	}

	t.Logf("EFS volume %s is now available", opts.Name)
	return availableVolume, nil
}

// CreateTestEBSStorageOptions contains options for creating a test EBS storage volume
type CreateTestEBSStorageOptions struct {
	Name       string
	SizeGB     int
	VolumeType string
}

// CreateTestEBSStorage creates a test EBS storage volume for integration tests
// The storage is automatically registered for cleanup via the registry
func CreateTestEBSStorage(t *testing.T, registry *FixtureRegistry, opts CreateTestEBSStorageOptions) (*types.StorageVolume, error) {
	t.Helper()

	// Set defaults
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("test-ebs-%d", time.Now().Unix())
	}
	if opts.SizeGB == 0 {
		opts.SizeGB = 10 // Minimum size for cost efficiency
	}
	if opts.VolumeType == "" {
		opts.VolumeType = "gp3" // Latest generation GP volume
	}

	ctx := context.Background()

	// Create EBS storage
	// Convert size to string format (API accepts "S", "M", "L" or specific GB like "10")
	sizeStr := fmt.Sprintf("%d", opts.SizeGB)
	storageReq := types.StorageCreateRequest{
		Name:       opts.Name,
		Size:       sizeStr,
		VolumeType: opts.VolumeType,
	}

	t.Logf("Creating test EBS storage: %s (size: %dGB, type: %s)", opts.Name, opts.SizeGB, opts.VolumeType)
	storage, err := registry.client.CreateStorage(ctx, storageReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create EBS storage: %w", err)
	}

	// Register for cleanup
	registry.Register("ebs_storage", opts.Name)
	t.Logf("EBS storage created: %s (Volume ID: %s)", opts.Name, storage.VolumeID)

	// Wait for storage to be available
	t.Logf("Waiting for EBS storage %s to reach available state...", opts.Name)
	availableStorage, err := waitForStorageState(ctx, registry.client, opts.Name, "available", 1*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("EBS storage did not reach available state: %w", err)
	}

	t.Logf("EBS storage %s is now available", opts.Name)
	return availableStorage, nil
}

// waitForVolumeState polls the EFS volume until it reaches the target state or times out
func waitForVolumeState(ctx context.Context, client client.PrismAPI, volumeName, targetState string, timeout time.Duration) (*types.StorageVolume, error) {
	startTime := time.Now()
	pollInterval := 10 * time.Second

	for time.Since(startTime) < timeout {
		volume, err := client.GetVolume(ctx, volumeName)
		if err != nil {
			// Ignore errors, keep polling
			time.Sleep(pollInterval)
			continue
		}

		if volume.State == targetState {
			return volume, nil
		}

		// Check if volume is in an error state
		if volume.State == "error" || volume.State == "failed" {
			return nil, fmt.Errorf("volume entered error state: %s", volume.State)
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("volume %s did not reach state %s within %v", volumeName, targetState, timeout)
}

// waitForStorageState polls the EBS storage until it reaches the target state or times out
func waitForStorageState(ctx context.Context, client client.PrismAPI, storageName, targetState string, timeout time.Duration) (*types.StorageVolume, error) {
	startTime := time.Now()
	pollInterval := 5 * time.Second // EBS is faster, poll more frequently

	for time.Since(startTime) < timeout {
		storage, err := client.GetStorage(ctx, storageName)
		if err != nil {
			// Ignore errors, keep polling
			time.Sleep(pollInterval)
			continue
		}

		if storage.State == targetState {
			return storage, nil
		}

		// Check if storage is in an error state
		if storage.State == "error" || storage.State == "failed" {
			return nil, fmt.Errorf("EBS storage entered error state: %s", storage.State)
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("EBS storage %s did not reach state %s within %v", storageName, targetState, timeout)
}
