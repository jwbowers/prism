package sleepwake

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		name     string
		event    EventType
		expected string
	}{
		{"Sleep event", EventSleep, "sleep"},
		{"Wake event", EventWake, "wake"},
		{"Shutdown event", EventShutdown, "shutdown"},
		{"Unknown event", EventType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.String())
		})
	}
}

func TestEvent_JSON(t *testing.T) {
	now := time.Now()
	event := Event{
		Type:      EventSleep,
		Timestamp: now,
		Source:    "iokit",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)

	// Unmarshal back
	var decoded Event
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.Type, decoded.Type)
	assert.Equal(t, event.Source, decoded.Source)
	assert.True(t, decoded.Timestamp.Unix() == now.Unix())
}

func TestHibernationMode_Values(t *testing.T) {
	assert.Equal(t, HibernationMode("idle_only"), HibernationModeIdleOnly)
	assert.Equal(t, HibernationMode("all"), HibernationModeAll)
	assert.Equal(t, HibernationMode("manual_only"), HibernationModeManualOnly)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled)
	assert.True(t, config.HibernateOnSleep)
	assert.Equal(t, HibernationModeIdleOnly, config.HibernationMode)
	assert.Equal(t, 10*time.Second, config.IdleCheckTimeout)
	assert.False(t, config.ResumeOnWake)
	assert.Equal(t, 30*time.Second, config.GracePeriod)
	assert.Empty(t, config.ExcludedInstances)
	assert.NotEmpty(t, config.StateFilePath)
}

func TestConfig_JSON(t *testing.T) {
	config := Config{
		Enabled:           true,
		HibernateOnSleep:  true,
		HibernationMode:   HibernationModeIdleOnly,
		IdleCheckTimeout:  10 * time.Second,
		ResumeOnWake:      false,
		GracePeriod:       30 * time.Second,
		ExcludedInstances: []string{"instance-1", "instance-2"},
		StateFilePath:     "/tmp/state.json",
	}

	// Marshal to JSON
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Unmarshal back
	var decoded Config
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, config.Enabled, decoded.Enabled)
	assert.Equal(t, config.HibernateOnSleep, decoded.HibernateOnSleep)
	assert.Equal(t, config.HibernationMode, decoded.HibernationMode)
	assert.Equal(t, config.IdleCheckTimeout, decoded.IdleCheckTimeout)
	assert.Equal(t, config.ResumeOnWake, decoded.ResumeOnWake)
	assert.Equal(t, config.GracePeriod, decoded.GracePeriod)
	assert.Equal(t, config.ExcludedInstances, decoded.ExcludedInstances)
}

func TestNewState(t *testing.T) {
	state := NewState("/tmp/test_state.json")

	assert.NotNil(t, state)
	assert.NotNil(t, state.HibernatedInstances)
	assert.Empty(t, state.HibernatedInstances)
	assert.Equal(t, "/tmp/test_state.json", state.StateFilePath)
	assert.Zero(t, state.TotalSleepEvents)
	assert.Zero(t, state.TotalWakeEvents)
	assert.Zero(t, state.TotalHibernated)
	assert.Zero(t, state.TotalResumed)
}

func TestState_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	// Create state with data
	state := NewState(stateFile)
	state.LastSleepTime = time.Now().Add(-1 * time.Hour)
	state.LastWakeTime = time.Now()
	state.HibernatedInstances["instance-1"] = time.Now().Add(-30 * time.Minute)
	state.HibernatedInstances["instance-2"] = time.Now().Add(-20 * time.Minute)
	state.TotalSleepEvents = 5
	state.TotalWakeEvents = 5
	state.TotalHibernated = 10
	state.TotalResumed = 8

	// Save
	err := state.Save()
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(stateFile)
	require.NoError(t, err)

	// Load into new state
	newState := NewState(stateFile)
	err = newState.Load()
	require.NoError(t, err)

	// Verify loaded data
	assert.Equal(t, state.TotalSleepEvents, newState.TotalSleepEvents)
	assert.Equal(t, state.TotalWakeEvents, newState.TotalWakeEvents)
	assert.Equal(t, state.TotalHibernated, newState.TotalHibernated)
	assert.Equal(t, state.TotalResumed, newState.TotalResumed)
	assert.Len(t, newState.HibernatedInstances, 2)
	assert.Contains(t, newState.HibernatedInstances, "instance-1")
	assert.Contains(t, newState.HibernatedInstances, "instance-2")
}

func TestState_Load_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nonexistent.json")

	state := NewState(stateFile)
	err := state.Load()

	// Should not error for nonexistent file, just use defaults
	assert.NoError(t, err)
	assert.Empty(t, state.HibernatedInstances)
}

func TestState_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nested", "dir", "state.json")

	state := NewState(stateFile)
	state.TotalSleepEvents = 1

	err := state.Save()
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(filepath.Dir(stateFile))
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(stateFile)
	require.NoError(t, err)
}

func TestState_RecordSleepEvent(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	state := NewState(stateFile)
	initialEvents := state.TotalSleepEvents

	beforeCall := time.Now()
	state.RecordSleepEvent()

	assert.GreaterOrEqual(t, state.LastSleepTime.Unix(), beforeCall.Unix())
	assert.Equal(t, initialEvents+1, state.TotalSleepEvents)
}

func TestState_RecordWakeEvent(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	state := NewState(stateFile)
	initialEvents := state.TotalWakeEvents

	beforeCall := time.Now()
	state.RecordWakeEvent()

	assert.GreaterOrEqual(t, state.LastWakeTime.Unix(), beforeCall.Unix())
	assert.Equal(t, initialEvents+1, state.TotalWakeEvents)
}

func TestState_AddHibernatedInstance(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	state := NewState(stateFile)
	initialCount := state.TotalHibernated

	state.AddHibernatedInstance("instance-1")

	assert.Equal(t, initialCount+1, state.TotalHibernated)
	assert.Contains(t, state.HibernatedInstances, "instance-1")
	assert.False(t, state.HibernatedInstances["instance-1"].IsZero())
}

func TestState_RecordResumed(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	state := NewState(stateFile)
	initialCount := state.TotalResumed

	state.RecordResumed()

	assert.Equal(t, initialCount+1, state.TotalResumed)
}

func TestState_GetHibernatedInstances(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	state := NewState(stateFile)
	state.HibernatedInstances["instance-1"] = time.Now().Add(-1 * time.Hour)
	state.HibernatedInstances["instance-2"] = time.Now().Add(-30 * time.Minute)

	instances := state.GetHibernatedInstances()

	assert.Len(t, instances, 2)
	assert.Contains(t, instances, "instance-1")
	assert.Contains(t, instances, "instance-2")
}

func TestState_RemoveHibernatedInstance(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	state := NewState(stateFile)
	state.HibernatedInstances["instance-1"] = time.Now()
	state.HibernatedInstances["instance-2"] = time.Now()

	state.RemoveHibernatedInstance("instance-1")

	assert.NotContains(t, state.HibernatedInstances, "instance-1")
	assert.Contains(t, state.HibernatedInstances, "instance-2")
}

func TestState_ClearHibernatedInstances(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	state := NewState(stateFile)
	state.HibernatedInstances["instance-1"] = time.Now()
	state.HibernatedInstances["instance-2"] = time.Now()

	state.ClearHibernatedInstances()

	assert.Empty(t, state.HibernatedInstances)
}

func TestState_GetStats(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	state := NewState(stateFile)
	state.RecordSleepEvent()
	state.RecordWakeEvent()
	state.AddHibernatedInstance("instance-1")
	state.RecordResumed()

	stats := state.GetStats()

	assert.Equal(t, int64(1), stats.TotalSleepEvents)
	assert.Equal(t, int64(1), stats.TotalWakeEvents)
	assert.Equal(t, int64(1), stats.TotalHibernated)
	assert.Equal(t, int64(1), stats.TotalResumed)
	assert.Equal(t, 1, stats.ActivelyTracked)
	assert.False(t, stats.LastSleepTime.IsZero())
	assert.False(t, stats.LastWakeTime.IsZero())
}
