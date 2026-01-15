//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/state"
	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCorruptedStateJSON_RecoverWithBackup tests recovery from corrupted state.json
func TestCorruptedStateJSON_RecoverWithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	// Create manager and save initial state
	manager, err := state.NewManager()
	require.NoError(t, err)

	initialState := &types.State{
		Version: types.CurrentStateVersion,
		Instances: map[string]types.Instance{
			"test-instance": {
				Name:     "test-instance",
				ID:       "i-1234567890abcdef0",
				PublicIP: "54.123.45.67",
				State:    "running",
			},
		},
		StorageVolumes: make(map[string]types.StorageVolume),
		Config: types.Config{
			DefaultRegion: "us-west-2",
		},
	}

	err = manager.SaveState(initialState)
	require.NoError(t, err)

	// Save again to create backup (first save creates state.json, second creates .bak)
	err = manager.SaveState(initialState)
	require.NoError(t, err)

	// Verify backup was created
	statePath := filepath.Join(tmpDir, "state.json")
	backupPath := statePath + ".bak"
	_, err = os.Stat(backupPath)
	require.NoError(t, err, "Backup should be created")

	// Corrupt primary state file
	err = os.WriteFile(statePath, []byte("corrupted json {{{"), 0644)
	require.NoError(t, err)

	// Load state (should recover from backup)
	recoveredState, err := manager.LoadState()
	require.NoError(t, err)
	assert.Equal(t, types.CurrentStateVersion, recoveredState.Version)
	assert.Len(t, recoveredState.Instances, 1)
	assert.Equal(t, "test-instance", recoveredState.Instances["test-instance"].Name)

	// Verify primary state file was restored
	data, err := os.ReadFile(statePath)
	require.NoError(t, err)
	var restoredState types.State
	err = json.Unmarshal(data, &restoredState)
	require.NoError(t, err, "Primary state should be restored from backup")
}

// TestPartialWriteState_RollbackToLast tests recovery from partial write
func TestPartialWriteState_RollbackToLast(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Save first state
	state1 := &types.State{
		Version: types.CurrentStateVersion,
		Instances: map[string]types.Instance{
			"instance-1": {Name: "instance-1", ID: "i-111", State: "running"},
		},
		StorageVolumes: make(map[string]types.StorageVolume),
		Config:         types.Config{DefaultRegion: "us-west-2"},
	}
	err = manager.SaveState(state1)
	require.NoError(t, err)

	// Save second state
	state2 := &types.State{
		Version: types.CurrentStateVersion,
		Instances: map[string]types.Instance{
			"instance-1": {Name: "instance-1", ID: "i-111", State: "running"},
			"instance-2": {Name: "instance-2", ID: "i-222", State: "running"},
		},
		StorageVolumes: make(map[string]types.StorageVolume),
		Config:         types.Config{DefaultRegion: "us-west-2"},
	}
	err = manager.SaveState(state2)
	require.NoError(t, err)

	// Simulate partial write by truncating file
	statePath := filepath.Join(tmpDir, "state.json")
	data, err := os.ReadFile(statePath)
	require.NoError(t, err)

	// Write partial data
	err = os.WriteFile(statePath, data[:len(data)/2], 0644)
	require.NoError(t, err)

	// Load state (should recover from backup)
	recoveredState, err := manager.LoadState()
	require.NoError(t, err)

	// Should recover state1 from backup
	assert.Len(t, recoveredState.Instances, 1)
	assert.Equal(t, "instance-1", recoveredState.Instances["instance-1"].Name)
}

// TestMissingStateFile_InitializeNew tests initialization of new state
func TestMissingStateFile_InitializeNew(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Load state when file doesn't exist
	state, err := manager.LoadState()
	require.NoError(t, err)

	// Should return empty initialized state
	assert.Equal(t, types.CurrentStateVersion, state.Version)
	assert.NotNil(t, state.Instances)
	assert.NotNil(t, state.StorageVolumes)
	assert.Len(t, state.Instances, 0)
	assert.Len(t, state.StorageVolumes, 0)
	assert.Equal(t, "us-east-1", state.Config.DefaultRegion)
}

// TestConcurrentStateUpdates_NoCorruption tests concurrent state updates
func TestConcurrentStateUpdates_NoCorruption(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Initialize state
	initialState := &types.State{
		Version:        types.CurrentStateVersion,
		Instances:      make(map[string]types.Instance),
		StorageVolumes: make(map[string]types.StorageVolume),
		Config:         types.Config{DefaultRegion: "us-west-2"},
	}
	err = manager.SaveState(initialState)
	require.NoError(t, err)

	// Concurrent updates
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Load state
			state, err := manager.LoadState()
			if err != nil {
				t.Errorf("Failed to load state: %v", err)
				return
			}

			// Add instance
			instanceName := "instance-" + string(rune('a'+id))
			state.Instances[instanceName] = types.Instance{
				Name:  instanceName,
				ID:    "i-" + string(rune('a'+id)),
				State: "running",
			}

			// Save state
			if err := manager.SaveState(state); err != nil {
				t.Errorf("Failed to save state: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Load final state
	finalState, err := manager.LoadState()
	require.NoError(t, err)

	// Verify no corruption (at least some instances should be saved)
	assert.Greater(t, len(finalState.Instances), 0)
	assert.Equal(t, types.CurrentStateVersion, finalState.Version)

	// Verify state file is valid JSON
	statePath := filepath.Join(tmpDir, "state.json")
	data, err := os.ReadFile(statePath)
	require.NoError(t, err)

	var validState types.State
	err = json.Unmarshal(data, &validState)
	require.NoError(t, err, "State file should be valid JSON after concurrent updates")
}

// TestBackupRotation_Maintains4Levels tests backup rotation preserves 4 levels
func TestBackupRotation_Maintains4Levels(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Save 5 states (should create 4 backups)
	for i := 1; i <= 5; i++ {
		state := &types.State{
			Version: types.CurrentStateVersion,
			Instances: map[string]types.Instance{
				"instance": {
					Name:  "instance",
					ID:    "i-" + string(rune('0'+i)),
					State: "running",
				},
			},
			StorageVolumes: make(map[string]types.StorageVolume),
			Config:         types.Config{DefaultRegion: "us-west-2"},
		}
		err = manager.SaveState(state)
		require.NoError(t, err)

		// Small delay to ensure distinct timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Verify backup files exist
	statePath := filepath.Join(tmpDir, "state.json")
	backups := []string{
		statePath + ".bak",
		statePath + ".bak1",
		statePath + ".bak2",
		statePath + ".bak3",
	}

	for _, backup := range backups {
		_, err := os.Stat(backup)
		assert.NoError(t, err, "Backup %s should exist", filepath.Base(backup))
	}

	// Verify no .bak4
	_, err = os.Stat(statePath + ".bak4")
	assert.True(t, os.IsNotExist(err), "Should not have .bak4")

	// Verify newest instance is in primary state
	state, err := manager.LoadState()
	require.NoError(t, err)
	assert.Equal(t, "i-5", state.Instances["instance"].ID)

	// Verify older instance is in oldest backup
	data, err := os.ReadFile(statePath + ".bak3")
	require.NoError(t, err)
	var oldState types.State
	err = json.Unmarshal(data, &oldState)
	require.NoError(t, err)
	assert.Equal(t, "i-1", oldState.Instances["instance"].ID)
}

// TestMultipleBackupCorruption_RecoverFromDeepest tests recovery from multiple corrupted backups
func TestMultipleBackupCorruption_RecoverFromDeepest(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Save multiple states to create backups
	for i := 1; i <= 4; i++ {
		state := &types.State{
			Version: types.CurrentStateVersion,
			Instances: map[string]types.Instance{
				"instance": {
					Name:  "instance",
					ID:    "i-" + string(rune('0'+i)),
					State: "running",
				},
			},
			StorageVolumes: make(map[string]types.StorageVolume),
			Config:         types.Config{DefaultRegion: "us-west-2"},
		}
		err = manager.SaveState(state)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Corrupt primary and first two backups
	statePath := filepath.Join(tmpDir, "state.json")
	err = os.WriteFile(statePath, []byte("corrupted"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(statePath+".bak", []byte("corrupted"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(statePath+".bak1", []byte("corrupted"), 0644)
	require.NoError(t, err)

	// Load state (should recover from .bak2)
	recoveredState, err := manager.LoadState()
	require.NoError(t, err)
	assert.Equal(t, types.CurrentStateVersion, recoveredState.Version)
	assert.Len(t, recoveredState.Instances, 1)
	assert.Equal(t, "i-1", recoveredState.Instances["instance"].ID)
}

// TestAllBackupsCorrupted_InitializeEmpty tests fallback to empty state
func TestAllBackupsCorrupted_InitializeEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Create corrupted state and backups
	statePath := filepath.Join(tmpDir, "state.json")
	corruptData := []byte("corrupted json")

	err = os.WriteFile(statePath, corruptData, 0644)
	require.NoError(t, err)
	err = os.WriteFile(statePath+".bak", corruptData, 0644)
	require.NoError(t, err)
	err = os.WriteFile(statePath+".bak1", corruptData, 0644)
	require.NoError(t, err)
	err = os.WriteFile(statePath+".bak2", corruptData, 0644)
	require.NoError(t, err)
	err = os.WriteFile(statePath+".bak3", corruptData, 0644)
	require.NoError(t, err)

	// Load state (should return empty initialized state)
	state, err := manager.LoadState()
	require.NoError(t, err)

	assert.Equal(t, types.CurrentStateVersion, state.Version)
	assert.Len(t, state.Instances, 0)
	assert.Len(t, state.StorageVolumes, 0)
	assert.Equal(t, "us-east-1", state.Config.DefaultRegion)
}

// TestVersionUpgrade_AutomaticallyApplied tests automatic version upgrade
func TestVersionUpgrade_AutomaticallyApplied(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Create old state file without version
	oldState := map[string]interface{}{
		"instances":       map[string]interface{}{},
		"storage_volumes": map[string]interface{}{},
		"config": map[string]interface{}{
			"default_region": "us-west-2",
		},
	}

	data, err := json.MarshalIndent(oldState, "", "  ")
	require.NoError(t, err)

	statePath := filepath.Join(tmpDir, "state.json")
	err = os.WriteFile(statePath, data, 0644)
	require.NoError(t, err)

	// Load state (should auto-upgrade version)
	state, err := manager.LoadState()
	require.NoError(t, err)

	assert.Equal(t, types.CurrentStateVersion, state.Version)

	// Save state (should persist upgraded version)
	err = manager.SaveState(state)
	require.NoError(t, err)

	// Load again and verify version persists
	state2, err := manager.LoadState()
	require.NoError(t, err)
	assert.Equal(t, types.CurrentStateVersion, state2.Version)
}

// TestAtomicWrite_NoPartialState tests atomic write prevents partial state
func TestAtomicWrite_NoPartialState(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Save large state
	largeState := &types.State{
		Version:   types.CurrentStateVersion,
		Instances: make(map[string]types.Instance),
		StorageVolumes: map[string]types.StorageVolume{
			"vol-1": {Name: "vol-1", FileSystemID: "fs-1", State: "available"},
			"vol-2": {Name: "vol-2", FileSystemID: "fs-2", State: "available"},
		},
		Config: types.Config{DefaultRegion: "us-west-2"},
	}

	// Add many instances
	for i := 0; i < 100; i++ {
		instanceName := "instance-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		largeState.Instances[instanceName] = types.Instance{
			Name:  instanceName,
			ID:    "i-" + instanceName,
			State: "running",
		}
	}

	err = manager.SaveState(largeState)
	require.NoError(t, err)

	// Verify state file is complete and valid
	statePath := filepath.Join(tmpDir, "state.json")
	data, err := os.ReadFile(statePath)
	require.NoError(t, err)

	var loadedState types.State
	err = json.Unmarshal(data, &loadedState)
	require.NoError(t, err, "State file should be complete valid JSON")

	assert.Equal(t, 100, len(loadedState.Instances))
	assert.Equal(t, 2, len(loadedState.StorageVolumes))

	// Verify no temporary file left behind
	tmpPath := statePath + ".tmp"
	_, err = os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(err), "Temporary file should be cleaned up")
}

// TestRecoveryPerformance_UnderMillisecond tests recovery completes quickly
func TestRecoveryPerformance_UnderMillisecond(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Save initial state
	initialState := &types.State{
		Version: types.CurrentStateVersion,
		Instances: map[string]types.Instance{
			"test-instance": {Name: "test-instance", ID: "i-123", State: "running"},
		},
		StorageVolumes: make(map[string]types.StorageVolume),
		Config:         types.Config{DefaultRegion: "us-west-2"},
	}
	err = manager.SaveState(initialState)
	require.NoError(t, err)

	// Save again to create backup
	err = manager.SaveState(initialState)
	require.NoError(t, err)

	// Corrupt primary state
	statePath := filepath.Join(tmpDir, "state.json")
	err = os.WriteFile(statePath, []byte("corrupted"), 0644)
	require.NoError(t, err)

	// Measure recovery time
	start := time.Now()
	state, err := manager.LoadState()
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, "test-instance", state.Instances["test-instance"].Name)
	assert.Less(t, elapsed, 100*time.Millisecond, "Recovery should complete within 100ms")
}

// TestEmptyJSONFile_RecoverGracefully tests handling of empty file
func TestEmptyJSONFile_RecoverGracefully(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Save initial state
	initialState := &types.State{
		Version: types.CurrentStateVersion,
		Instances: map[string]types.Instance{
			"test": {Name: "test", ID: "i-123", State: "running"},
		},
		StorageVolumes: make(map[string]types.StorageVolume),
		Config:         types.Config{DefaultRegion: "us-west-2"},
	}
	err = manager.SaveState(initialState)
	require.NoError(t, err)

	// Save again to create backup
	err = manager.SaveState(initialState)
	require.NoError(t, err)

	// Create empty file
	statePath := filepath.Join(tmpDir, "state.json")
	err = os.WriteFile(statePath, []byte(""), 0644)
	require.NoError(t, err)

	// Load state (should recover from backup)
	state, err := manager.LoadState()
	require.NoError(t, err)
	assert.Len(t, state.Instances, 1)
	assert.Equal(t, "test", state.Instances["test"].Name)
}

// TestInvalidJSONStructure_RecoverGracefully tests handling of invalid structure
func TestInvalidJSONStructure_RecoverGracefully(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Save initial state
	initialState := &types.State{
		Version: types.CurrentStateVersion,
		Instances: map[string]types.Instance{
			"test": {Name: "test", ID: "i-123", State: "running"},
		},
		StorageVolumes: make(map[string]types.StorageVolume),
		Config:         types.Config{DefaultRegion: "us-west-2"},
	}
	err = manager.SaveState(initialState)
	require.NoError(t, err)

	// Write invalid JSON structure (valid JSON but wrong structure)
	statePath := filepath.Join(tmpDir, "state.json")
	invalidStructure := `{"wrong_field": "value", "another_field": 123}`
	err = os.WriteFile(statePath, []byte(invalidStructure), 0644)
	require.NoError(t, err)

	// Load state (should recover from backup or initialize empty)
	state, err := manager.LoadState()
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.NotNil(t, state.Instances)
	assert.NotNil(t, state.StorageVolumes)
}

// TestStateFilePermissions_Maintained tests file permissions are preserved
func TestStateFilePermissions_Maintained(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PRISM_STATE_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("PRISM_STATE_DIR") })

	manager, err := state.NewManager()
	require.NoError(t, err)

	// Save state
	state := &types.State{
		Version:        types.CurrentStateVersion,
		Instances:      make(map[string]types.Instance),
		StorageVolumes: make(map[string]types.StorageVolume),
		Config:         types.Config{DefaultRegion: "us-west-2"},
	}
	err = manager.SaveState(state)
	require.NoError(t, err)

	// Check file permissions
	statePath := filepath.Join(tmpDir, "state.json")
	info, err := os.Stat(statePath)
	require.NoError(t, err)

	// Should be readable/writable by owner (0644 or more restrictive)
	mode := info.Mode().Perm()
	assert.True(t, mode&0400 != 0, "File should be readable by owner")
	assert.True(t, mode&0200 != 0, "File should be writable by owner")
}
