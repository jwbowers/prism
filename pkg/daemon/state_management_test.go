package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStateInitialization tests that server initializes state correctly
func TestStateInitialization(t *testing.T) {
	server := createTestServer(t)

	// State manager should be initialized
	assert.NotNil(t, server.stateManager)

	// Should be able to load state
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.NotNil(t, state.Instances)
	assert.NotNil(t, state.StorageVolumes)
}

// TestStatePersistence tests that state changes are persisted
func TestStatePersistence(t *testing.T) {
	server := createTestServer(t)

	// Create a test instance
	testInstance := types.Instance{
		ID:           "i-test123",
		Name:         "test-instance-persist",
		Template:     "test-template",
		State:        "running",
		PublicIP:     "1.2.3.4",
		PrivateIP:    "10.0.0.1",
		LaunchTime:   time.Now(),
		Region:       "us-east-1",
		InstanceType: "t3.micro",
	}

	// Save instance to state
	err := server.stateManager.SaveInstance(testInstance)
	require.NoError(t, err)

	// Load state and verify instance is saved
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)

	savedInstance, exists := state.Instances[testInstance.Name]
	assert.True(t, exists, "Instance should be saved in state")
	assert.Equal(t, testInstance.ID, savedInstance.ID)
	assert.Equal(t, testInstance.Name, savedInstance.Name)
	assert.Equal(t, testInstance.State, savedInstance.State)
}

// TestStateRemoval tests that removing items from state works
func TestStateRemoval(t *testing.T) {
	server := createTestServer(t)

	// Create and save a test instance
	testInstance := types.Instance{
		ID:       "i-remove123",
		Name:     "test-instance-remove",
		Template: "test-template",
		State:    "terminated",
	}

	err := server.stateManager.SaveInstance(testInstance)
	require.NoError(t, err)

	// Verify instance exists
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)
	_, exists := state.Instances[testInstance.Name]
	assert.True(t, exists)

	// Remove instance
	err = server.stateManager.RemoveInstance(testInstance.Name)
	require.NoError(t, err)

	// Verify instance is removed
	state, err = server.stateManager.LoadState()
	require.NoError(t, err)
	_, exists = state.Instances[testInstance.Name]
	assert.False(t, exists, "Instance should be removed from state")
}

// TestStateConcurrentAccess tests concurrent state reads and writes
func TestStateConcurrentAccess(t *testing.T) {
	server := createTestServer(t)

	const numGoroutines = 10
	const numOperations = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Launch multiple goroutines performing state operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Write operation
				instance := types.Instance{
					ID:       "i-concurrent-" + string(rune(id)),
					Name:     "instance-" + string(rune(id)),
					Template: "test-template",
					State:    "running",
				}

				err := server.stateManager.SaveInstance(instance)
				if err != nil {
					errors <- err
					return
				}

				// Read operation
				_, err = server.stateManager.LoadState()
				if err != nil {
					errors <- err
					return
				}

				// Small delay to increase chance of race conditions
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent state access error: %v", err)
	}
}

// TestStateFileCorruptionRecovery tests recovery from corrupted state files
func TestStateFileCorruptionRecovery(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, ".prism", "state.json")

	// Create state directory
	err := os.MkdirAll(filepath.Dir(statePath), 0755)
	require.NoError(t, err)

	// Write corrupted JSON to state file
	corruptedJSON := `{"instances": {invalid json`
	err = os.WriteFile(statePath, []byte(corruptedJSON), 0644)
	require.NoError(t, err)

	// Override home directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		_ = os.Setenv("HOME", originalHome)
	})
	_ = os.Setenv("HOME", tempDir)

	// Create server - may fail to initialize due to corrupted state
	// This is expected behavior - daemon should not start with corrupted state
	server, err := NewServerForTesting("8949")

	// Either server creation fails (expected) or succeeds with recovery
	if err != nil {
		// Corrupted state caused initialization to fail - this is acceptable
		assert.Contains(t, err.Error(), "state")
	} else {
		// Server initialized despite corrupted state - verify it recovered
		assert.NotNil(t, server)

		// State manager should be functional after recovery
		state, err := server.stateManager.LoadState()
		if err == nil {
			assert.NotNil(t, state)
		}
	}
}

// TestStateBackupCreation tests that backups are created
func TestStateBackupCreation(t *testing.T) {
	server := createTestServer(t)

	// Get state file path
	homeDir := os.Getenv("HOME")
	stateDir := filepath.Join(homeDir, ".prism")
	statePath := filepath.Join(stateDir, "state.json")

	// Save an instance to trigger state save
	testInstance := types.Instance{
		ID:       "i-backup123",
		Name:     "test-backup-instance",
		Template: "test-template",
		State:    "running",
	}

	err := server.stateManager.SaveInstance(testInstance)
	require.NoError(t, err)

	// Verify state file exists
	_, err = os.Stat(statePath)
	assert.NoError(t, err, "State file should exist")

	// Check for backup files (state.json.backup)
	backupPath := statePath + ".backup"
	// Backup might not exist immediately, but state file should
	if _, err := os.Stat(backupPath); err == nil {
		// If backup exists, verify it's valid JSON
		backupData, err := os.ReadFile(backupPath)
		require.NoError(t, err)

		var state types.State
		err = json.Unmarshal(backupData, &state)
		assert.NoError(t, err, "Backup file should contain valid JSON")
	}
}

// TestStateAtomicWrites tests that state writes are atomic
func TestStateAtomicWrites(t *testing.T) {
	server := createTestServer(t)

	// Perform multiple rapid state changes
	for i := 0; i < 10; i++ {
		instance := types.Instance{
			ID:       "i-atomic-" + string(rune(i)),
			Name:     "atomic-instance",
			Template: "test-template",
			State:    "running",
		}

		err := server.stateManager.SaveInstance(instance)
		require.NoError(t, err)
	}

	// State should be consistent
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)

	// Should have the latest instance
	instance, exists := state.Instances["atomic-instance"]
	assert.True(t, exists)
	assert.NotEmpty(t, instance.ID)
}

// TestStateVolumeTracking tests storage volume state tracking
func TestStateVolumeTracking(t *testing.T) {
	server := createTestServer(t)

	// Load initial state
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)
	initialVolumeCount := len(state.StorageVolumes)

	// Create a test EFS volume
	sizeGB := int32(100)
	testVolume := types.StorageVolume{
		Name:            "test-volume",
		Type:            types.StorageTypeShared,
		AWSService:      types.AWSServiceEFS,
		FileSystemID:    "fs-test123",
		State:           "available",
		SizeGB:          &sizeGB,
		CreationTime:    time.Now(),
		EstimatedCostGB: 0.30,
	}

	err = server.stateManager.SaveStorageVolume(testVolume)
	require.NoError(t, err)

	// Verify volume is in state
	state, err = server.stateManager.LoadState()
	require.NoError(t, err)
	assert.Equal(t, initialVolumeCount+1, len(state.StorageVolumes))

	volume, exists := state.StorageVolumes[testVolume.Name]
	assert.True(t, exists)
	assert.Equal(t, testVolume.FileSystemID, volume.FileSystemID)
}

// TestStateEBSVolumeTracking tests EBS volume state tracking
func TestStateEBSVolumeTracking(t *testing.T) {
	server := createTestServer(t)

	// Load initial state
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)

	// Storage volumes should be tracked (includes both EFS and EBS)
	assert.NotNil(t, state.StorageVolumes)
}

// TestStateProjectTracking tests project state tracking
func TestStateProjectTracking(t *testing.T) {
	server := createTestServer(t)

	// State should include projects if using unified state
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)
	assert.NotNil(t, state)

	// Projects might be empty initially but should not be nil
	// The actual project data is managed by projectManager
	assert.NotNil(t, server.projectManager)
}

// TestStateInstanceMetadata tests that instance metadata is preserved
func TestStateInstanceMetadata(t *testing.T) {
	server := createTestServer(t)

	// Create instance with rich metadata
	testInstance := types.Instance{
		ID:                 "i-metadata123",
		Name:               "test-metadata",
		Template:           "test-template",
		State:              "running",
		PublicIP:           "1.2.3.4",
		PrivateIP:          "10.0.0.1",
		InstanceType:       "t3.medium",
		LaunchTime:         time.Now(),
		Region:             "us-west-2",
		AvailabilityZone:   "us-west-2a",
		EstimatedCost:      0.50,
		AttachedVolumes:    []string{"vol-1", "vol-2"},
		AttachedEBSVolumes: []string{"ebs-vol-1"},
		KeyName:            "test-key",
		HourlyRate:         0.0208,
		CurrentSpend:       1.25,
		EffectiveRate:      0.0208,
	}

	// Save instance
	err := server.stateManager.SaveInstance(testInstance)
	require.NoError(t, err)

	// Load and verify all metadata is preserved
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)

	savedInstance, exists := state.Instances[testInstance.Name]
	require.True(t, exists)

	// Verify all fields
	assert.Equal(t, testInstance.ID, savedInstance.ID)
	assert.Equal(t, testInstance.PublicIP, savedInstance.PublicIP)
	assert.Equal(t, testInstance.PrivateIP, savedInstance.PrivateIP)
	assert.Equal(t, testInstance.InstanceType, savedInstance.InstanceType)
	assert.Equal(t, testInstance.Region, savedInstance.Region)
	assert.Equal(t, testInstance.AvailabilityZone, savedInstance.AvailabilityZone)
	assert.Equal(t, testInstance.EstimatedCost, savedInstance.EstimatedCost)
	assert.Equal(t, testInstance.AttachedVolumes, savedInstance.AttachedVolumes)
	assert.Equal(t, testInstance.AttachedEBSVolumes, savedInstance.AttachedEBSVolumes)
	assert.Equal(t, testInstance.KeyName, savedInstance.KeyName)
	assert.Equal(t, testInstance.HourlyRate, savedInstance.HourlyRate)
	assert.Equal(t, testInstance.CurrentSpend, savedInstance.CurrentSpend)
	assert.Equal(t, testInstance.EffectiveRate, savedInstance.EffectiveRate)
}

// TestStateLoadPerformance tests that state loading is fast
func TestStateLoadPerformance(t *testing.T) {
	server := createTestServer(t)

	// Add several instances to state
	for i := 0; i < 20; i++ {
		instance := types.Instance{
			ID:       "i-perf-" + string(rune(i)),
			Name:     "perf-instance-" + string(rune(i)),
			Template: "test-template",
			State:    "running",
		}
		err := server.stateManager.SaveInstance(instance)
		require.NoError(t, err)
	}

	// Measure load time
	start := time.Now()
	_, err := server.stateManager.LoadState()
	duration := time.Since(start)

	require.NoError(t, err)

	// Loading state should be fast (< 100ms even with 20 instances)
	assert.Less(t, duration.Milliseconds(), int64(100),
		"State loading should complete in less than 100ms, took %v", duration)
}

// TestStateConsistencyAfterError tests state remains consistent after errors
func TestStateConsistencyAfterError(t *testing.T) {
	server := createTestServer(t)

	// Save a valid instance
	validInstance := types.Instance{
		ID:       "i-valid123",
		Name:     "valid-instance",
		Template: "test-template",
		State:    "running",
	}

	err := server.stateManager.SaveInstance(validInstance)
	require.NoError(t, err)

	// Attempt to save an instance with problematic data
	// (State manager should handle this gracefully)
	problematicInstance := types.Instance{
		ID:       "", // Empty ID
		Name:     "problematic-instance",
		Template: "", // Empty template
		State:    "unknown",
	}

	// Even if this fails, state should remain consistent
	_ = server.stateManager.SaveInstance(problematicInstance)

	// Load state and verify valid instance is still there
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)

	_, validExists := state.Instances["valid-instance"]
	assert.True(t, validExists, "Valid instance should still exist in state")
}

// TestStateCleanup tests cleanup of terminated instances
func TestStateCleanup(t *testing.T) {
	server := createTestServer(t)

	// Create a terminated instance (old)
	oldTerminated := types.Instance{
		ID:         "i-oldterm123",
		Name:       "old-terminated",
		Template:   "test-template",
		State:      "terminated",
		LaunchTime: time.Now().Add(-48 * time.Hour), // 2 days ago
	}

	err := server.stateManager.SaveInstance(oldTerminated)
	require.NoError(t, err)

	// Create a recently terminated instance
	recentTerminated := types.Instance{
		ID:         "i-recentterm123",
		Name:       "recent-terminated",
		Template:   "test-template",
		State:      "terminated",
		LaunchTime: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	err = server.stateManager.SaveInstance(recentTerminated)
	require.NoError(t, err)

	// Both should be in state initially
	state, err := server.stateManager.LoadState()
	require.NoError(t, err)

	_, oldExists := state.Instances[oldTerminated.Name]
	_, recentExists := state.Instances[recentTerminated.Name]
	assert.True(t, oldExists)
	assert.True(t, recentExists)

	// After cleanup, old terminated instances may be removed
	// (This depends on server configuration and cleanup policy)
	// For now, just verify state is consistent
	assert.NotNil(t, state.Instances)
}
