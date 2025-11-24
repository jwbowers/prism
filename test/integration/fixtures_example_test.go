//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleBackupWorkflow demonstrates using the fixtures package for integration tests
// This example creates a real instance, creates a backup, and automatic cleanup happens via t.Cleanup()
func ExampleBackupWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Initialize API client pointing to local daemon
	client := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		Region:     "us-west-2",
	})

	// Create fixture registry - cleanup is automatic via t.Cleanup()
	registry := fixtures.NewFixtureRegistry(t, client)

	// Step 1: Create test instance
	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "Ubuntu Basic",
		Name:     "backup-example",
		Size:     "S", // Small size for cost efficiency
	})
	require.NoError(t, err, "Failed to create test instance")
	require.NotNil(t, instance)
	assert.Equal(t, "running", instance.State)

	// Step 2: Create backup from instance
	backup, err := fixtures.CreateTestBackup(t, registry, fixtures.CreateTestBackupOptions{
		InstanceID:  instance.Name,
		Name:        "example-backup",
		Description: "Example integration test backup",
	})
	require.NoError(t, err, "Failed to create backup")
	require.NotNil(t, backup)
	assert.Equal(t, "available", backup.Status)

	// Step 3: Verify backup via API
	ctx := context.Background()
	retrievedBackup, err := client.GetBackup(ctx, backup.Name)
	require.NoError(t, err)
	assert.Equal(t, backup.Name, retrievedBackup.Name)

	// Cleanup happens automatically via t.Cleanup() - no explicit cleanup needed!
	// The registry will delete: backups first, then instances
}

// NOTE: Profile fixtures are not included because profiles are managed
// locally via pkg/profile package, not through the daemon API

// ExampleStorageWorkflow demonstrates storage fixture usage
func ExampleStorageWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		Region:     "us-west-2",
	})

	registry := fixtures.NewFixtureRegistry(t, client)

	// Create EFS volume
	volume, err := fixtures.CreateTestVolume(t, registry, fixtures.CreateTestVolumeOptions{
		Name:            "example-volume",
		PerformanceMode: "generalPurpose",
	})
	require.NoError(t, err)
	require.NotNil(t, volume)
	assert.Equal(t, "available", volume.State)

	// Create EBS storage
	storage, err := fixtures.CreateTestEBSStorage(t, registry, fixtures.CreateTestEBSStorageOptions{
		Name:       "example-ebs",
		SizeGB:     10,
		VolumeType: "gp3",
	})
	require.NoError(t, err)
	require.NotNil(t, storage)
	assert.Equal(t, "available", storage.State)

	// Automatic cleanup via t.Cleanup()
}

// ExampleCompleteTestEnvironment demonstrates creating a complete test environment
// with multiple resources that depend on each other
func ExampleCompleteTestEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running integration test")
	}

	client := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		Region:     "us-west-2",
	})

	registry := fixtures.NewFixtureRegistry(t, client)
	ctx := context.Background()

	// Step 1: Create storage volumes
	efsVolume, err := fixtures.CreateTestVolume(t, registry, fixtures.CreateTestVolumeOptions{
		Name:            "test-env-efs",
		PerformanceMode: "generalPurpose",
	})
	require.NoError(t, err)
	t.Logf("Created EFS volume: %s", efsVolume.FilesystemID)

	ebsStorage, err := fixtures.CreateTestEBSStorage(t, registry, fixtures.CreateTestEBSStorageOptions{
		Name:       "test-env-ebs",
		SizeGB:     10,
		VolumeType: "gp3",
	})
	require.NoError(t, err)
	t.Logf("Created EBS storage: %s (Volume ID: %s)", ebsStorage.Name, ebsStorage.VolumeID)

	// Step 2: Create instance
	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "Ubuntu Basic",
		Name:     "test-env-instance",
		Size:     "S",
	})
	require.NoError(t, err)
	t.Logf("Created instance: %s (ID: %s)", instance.Name, instance.ID)

	// Step 3: Attach volumes to instance
	err = client.AttachVolume(ctx, efsVolume.Name, instance.Name)
	require.NoError(t, err)
	t.Logf("Attached EFS volume to instance")

	err = client.AttachStorage(ctx, ebsStorage.Name, instance.Name)
	require.NoError(t, err)
	t.Logf("Attached EBS storage to instance")

	// Step 4: Create backup
	backup, err := fixtures.CreateTestBackup(t, registry, fixtures.CreateTestBackupOptions{
		InstanceID:  instance.Name,
		Name:        "test-env-backup",
		Description: "Complete environment backup",
	})
	require.NoError(t, err)
	t.Logf("Created backup: %s", backup.Name)

	// Verify all resources exist
	_, err = client.GetVolume(ctx, efsVolume.Name)
	require.NoError(t, err)

	_, err = client.GetStorage(ctx, ebsStorage.Name)
	require.NoError(t, err)

	_, err = client.GetInstance(ctx, instance.Name)
	require.NoError(t, err)

	_, err = client.GetBackup(ctx, backup.Name)
	require.NoError(t, err)

	t.Log("✓ Complete test environment created and verified")

	// Automatic cleanup via t.Cleanup() - resources are cleaned up in correct order:
	// 1. Backups (fastest, no dependencies)
	// 2. Instances (must terminate before volumes can be deleted)
	// 3. EBS storage
	// 4. EFS volumes
}

// ExampleManualCleanup demonstrates explicit cleanup (optional)
func ExampleManualCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		Region:     "us-west-2",
	})

	registry := fixtures.NewFixtureRegistry(t, client)

	// Create instance
	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "Ubuntu Basic",
		Name:     "manual-cleanup-example",
		Size:     "S",
	})
	require.NoError(t, err)

	// Do some testing...
	assert.Equal(t, "running", instance.State)

	// Manually trigger cleanup if needed (optional - normally automatic)
	// This is useful if you want to cleanup mid-test
	registry.Cleanup()

	// Verify instance was terminated
	time.Sleep(5 * time.Second)
	ctx := context.Background()
	terminatedInstance, err := client.GetInstance(ctx, instance.Name)
	if err == nil {
		// Instance might still exist but should be terminating/terminated
		assert.Contains(t, []string{"terminating", "terminated"}, terminatedInstance.State)
	}

	// Note: t.Cleanup() will still run, but FixtureRegistry tracks that cleanup
	// was already called and won't double-cleanup
}
