package fixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
)

// FixtureRegistry tracks test resources for automatic cleanup
// Uses t.Cleanup() to ensure resources are always cleaned up, even on test failure
type FixtureRegistry struct {
	t             *testing.T
	client        client.PrismAPI
	instances     []string
	backups       []string
	volumes       []string
	ebsStorages   []string
	profiles      []string
	cleanupCalled bool
}

// NewFixtureRegistry creates a new fixture registry with automatic cleanup
// The cleanup function is registered with t.Cleanup() to run after the test
func NewFixtureRegistry(t *testing.T, apiClient client.PrismAPI) *FixtureRegistry {
	t.Helper()

	registry := &FixtureRegistry{
		t:      t,
		client: apiClient,
	}

	// Register cleanup function to run after test completes
	t.Cleanup(func() {
		if !registry.cleanupCalled {
			registry.cleanup()
		}
	})

	return registry
}

// Register adds a resource ID to the registry for cleanup
func (r *FixtureRegistry) Register(resourceType, id string) {
	r.t.Helper()

	switch resourceType {
	case "instance":
		r.instances = append(r.instances, id)
	case "backup":
		r.backups = append(r.backups, id)
	case "volume":
		r.volumes = append(r.volumes, id)
	case "ebs_storage":
		r.ebsStorages = append(r.ebsStorages, id)
	case "profile":
		r.profiles = append(r.profiles, id)
	default:
		r.t.Logf("Warning: Unknown resource type %q", resourceType)
	}
}

// cleanup removes all registered test resources in the correct order
// Order: backups → instances → EBS volumes → EFS volumes → profiles
func (r *FixtureRegistry) cleanup() {
	r.t.Helper()
	r.cleanupCalled = true

	ctx := context.Background()
	r.t.Log("Cleaning up test fixtures...")

	// Clean up backups first (fastest, no dependencies)
	for _, backupID := range r.backups {
		result, err := r.client.DeleteBackup(ctx, backupID)
		if err != nil {
			r.t.Logf("Warning: Failed to cleanup backup %s: %v", backupID, err)
		} else {
			r.t.Logf("Cleaned up backup: %s (savings: $%.2f/month)", backupID, result.StorageSavingsMonthly)
		}
	}

	// Clean up instances (must be deleted before volumes can be deleted)
	for _, instanceName := range r.instances {
		if err := r.client.DeleteInstance(ctx, instanceName); err != nil {
			r.t.Logf("Warning: Failed to cleanup instance %s: %v", instanceName, err)
		} else {
			r.t.Logf("Cleaned up instance: %s", instanceName)
			// Give AWS time to begin termination
			time.Sleep(2 * time.Second)
		}
	}

	// Clean up EBS volumes
	for _, storageName := range r.ebsStorages {
		if err := r.client.DeleteStorage(ctx, storageName); err != nil {
			r.t.Logf("Warning: Failed to cleanup EBS volume %s: %v", storageName, err)
		} else {
			r.t.Logf("Cleaned up EBS volume: %s", storageName)
		}
	}

	// Clean up EFS volumes
	for _, volumeName := range r.volumes {
		if err := r.client.DeleteVolume(ctx, volumeName); err != nil {
			r.t.Logf("Warning: Failed to cleanup EFS volume %s: %v", volumeName, err)
		} else {
			r.t.Logf("Cleaned up EFS volume: %s", volumeName)
		}
	}

	// Note: Profiles are managed locally via pkg/profile package, not through daemon API
	// They would need to be cleaned up differently if needed
	if len(r.profiles) > 0 {
		r.t.Logf("Note: %d profiles registered but profile cleanup not implemented (profiles managed locally)", len(r.profiles))
	}

	r.t.Log("Fixture cleanup complete")
}

// Cleanup manually triggers cleanup (useful for defer statements)
// Note: Cleanup is automatically called via t.Cleanup(), so this is optional
func (r *FixtureRegistry) Cleanup() {
	r.cleanup()
}

// logError is a helper to log errors consistently
func (r *FixtureRegistry) logError(resource, id string, err error) {
	r.t.Helper()
	r.t.Logf("Error with %s %s: %v", resource, id, err)
}

// formatDuration formats a duration for logging
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
