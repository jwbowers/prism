package fixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/integration"
)

// CreateTestInstanceOptions contains options for creating a test instance
type CreateTestInstanceOptions struct {
	Template            string
	Name                string
	Size                string
	ProjectID           *string         // Optional: Associate instance with a project for budget enforcement
	FundingAllocationID string          // Optional: Specific allocation to charge (v0.6.2+)
	Context             context.Context // Optional: Custom context for timeouts/cancellation (chaos tests)
}

// CreateTestInstance creates a test instance for integration tests
// The instance is automatically registered for cleanup via the registry
func CreateTestInstance(t *testing.T, registry *FixtureRegistry, opts CreateTestInstanceOptions) (*types.Instance, error) {
	t.Helper()

	// Set defaults
	if opts.Template == "" {
		opts.Template = "Ubuntu Basic"
	}
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("test-instance-%d", time.Now().Unix())
	}
	if opts.Size == "" {
		opts.Size = "S" // Smallest size for cost efficiency
	}

	// CRITICAL: Acquire instance slot (blocks until available, max 2 concurrent)
	suiteManager := integration.GetSuiteManager()
	suiteManager.SetClient(registry.client) // Ensure client available for cleanup
	suiteManager.AcquireInstanceSlot(t, opts.Name)

	// Ensure slot is released when instance is done (LIFO cleanup via t.Cleanup)
	t.Cleanup(func() {
		suiteManager.ReleaseInstanceSlot(t, opts.Name)
	})

	// Use custom context if provided (for chaos tests with timeouts)
	// Otherwise use background context
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Launch instance
	projectID := ""
	if opts.ProjectID != nil {
		projectID = *opts.ProjectID
	}

	launchReq := types.LaunchRequest{
		Template:            opts.Template,
		Name:                opts.Name,
		Size:                opts.Size,
		ProjectID:           projectID,                // Associate with project if provided
		FundingAllocationID: opts.FundingAllocationID, // Specify funding source (v0.6.2+)
	}

	if projectID != "" {
		t.Logf("Creating test instance: %s (template: %s, size: %s, project: %s)", opts.Name, opts.Template, opts.Size, projectID)
	} else {
		t.Logf("Creating test instance: %s (template: %s, size: %s)", opts.Name, opts.Template, opts.Size)
	}
	launchResp, err := registry.client.LaunchInstance(ctx, launchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to launch instance: %w", err)
	}

	// Register for cleanup
	registry.Register("instance", opts.Name)
	t.Logf("Instance created: %s", opts.Name)

	// Wait for instance to be running
	t.Logf("Waiting for instance %s to reach running state...", opts.Name)
	runningInstance, err := waitForInstanceState(ctx, registry.client, opts.Name, "running", 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("instance did not reach running state: %w", err)
	}

	t.Logf("Instance %s is now running (ID: %s)", opts.Name, launchResp.Instance.ID)
	return runningInstance, nil
}

// CreateTestBackupOptions contains options for creating a test backup
type CreateTestBackupOptions struct {
	InstanceID  string
	Name        string
	Description string
}

// CreateTestBackup creates a test backup from an instance
// The backup is automatically registered for cleanup via the registry
func CreateTestBackup(t *testing.T, registry *FixtureRegistry, opts CreateTestBackupOptions) (*types.BackupInfo, error) {
	t.Helper()

	if opts.InstanceID == "" {
		return nil, fmt.Errorf("instance ID is required to create a backup")
	}

	// Set defaults
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("test-backup-%d", time.Now().Unix())
	}
	if opts.Description == "" {
		opts.Description = "Integration test backup"
	}

	ctx := context.Background()

	// Create backup
	backupReq := types.BackupCreateRequest{
		InstanceName: opts.InstanceID,
		BackupName:   opts.Name,
		Description:  opts.Description,
	}

	t.Logf("Creating test backup: %s (from instance: %s)", opts.Name, opts.InstanceID)
	backupResult, err := registry.client.CreateBackup(ctx, backupReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Register for cleanup
	registry.Register("backup", opts.Name)
	t.Logf("Backup created: %s", opts.Name)

	// Wait for backup to be available
	t.Logf("Waiting for backup %s to reach available state (this may take several minutes)...", opts.Name)
	availableBackup, err := waitForBackupState(ctx, registry.client, opts.Name, "available", 10*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("backup did not reach available state: %w", err)
	}

	t.Logf("Backup %s is now available (ID: %s)", opts.Name, backupResult.BackupID)
	return availableBackup, nil
}

// waitForInstanceState polls the instance until it reaches the target state or times out
func waitForInstanceState(ctx context.Context, client client.PrismAPI, instanceName, targetState string, timeout time.Duration) (*types.Instance, error) {
	startTime := time.Now()
	pollInterval := 10 * time.Second

	for time.Since(startTime) < timeout {
		instance, err := client.GetInstance(ctx, instanceName)
		if err != nil {
			// Ignore errors, keep polling
			time.Sleep(pollInterval)
			continue
		}

		if instance.State == targetState {
			return instance, nil
		}

		// Check if instance is in a terminal error state
		if instance.State == "terminated" || instance.State == "terminating" {
			return nil, fmt.Errorf("instance entered terminal state: %s", instance.State)
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("instance %s did not reach state %s within %v", instanceName, targetState, timeout)
}

// waitForBackupState polls the backup until it reaches the target state or times out
func waitForBackupState(ctx context.Context, client client.PrismAPI, backupName, targetState string, timeout time.Duration) (*types.BackupInfo, error) {
	startTime := time.Now()
	pollInterval := 15 * time.Second // Backups are slower, poll less frequently

	for time.Since(startTime) < timeout {
		backup, err := client.GetBackup(ctx, backupName)
		if err != nil {
			// Ignore errors, keep polling
			time.Sleep(pollInterval)
			continue
		}

		if backup.State == targetState {
			return backup, nil
		}

		// Check if backup is in an error state
		if backup.State == "failed" || backup.State == "error" {
			return nil, fmt.Errorf("backup entered error state: %s", backup.State)
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("backup %s did not reach state %s within %v", backupName, targetState, timeout)
}
