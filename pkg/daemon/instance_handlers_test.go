package daemon

import (
	"testing"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyExpectedTransitionalState(t *testing.T) {
	tests := []struct {
		name          string
		initialState  string
		expectedState string
		wantState     string
		wantHistory   bool
	}{
		{
			name:          "start: stopped → pending",
			initialState:  "stopped",
			expectedState: "pending",
			wantState:     "pending",
			wantHistory:   true,
		},
		{
			name:          "stop: running → stopping",
			initialState:  "running",
			expectedState: "stopping",
			wantState:     "stopping",
			wantHistory:   true,
		},
		{
			name:          "hibernate: running → stopping",
			initialState:  "running",
			expectedState: "stopping",
			wantState:     "stopping",
			wantHistory:   true,
		},
		{
			name:          "resume: hibernated → pending",
			initialState:  "hibernated",
			expectedState: "pending",
			wantState:     "pending",
			wantHistory:   true,
		},
		{
			name:          "no-op: already in expected state",
			initialState:  "stopping",
			expectedState: "stopping",
			wantState:     "stopping",
			wantHistory:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(t)

			// Seed instance in state
			instance := types.Instance{
				Name:  "test-instance",
				ID:    "i-test123",
				State: tt.initialState,
			}
			err := server.stateManager.SaveInstance(instance)
			require.NoError(t, err)

			// Apply optimistic state
			server.applyExpectedTransitionalState("test-instance", tt.expectedState)

			// Verify
			state, err := server.stateManager.LoadState()
			require.NoError(t, err)

			inst, exists := state.Instances["test-instance"]
			require.True(t, exists)
			assert.Equal(t, tt.wantState, inst.State)

			if tt.wantHistory {
				require.NotEmpty(t, inst.StateHistory)
				last := inst.StateHistory[len(inst.StateHistory)-1]
				assert.Equal(t, tt.initialState, last.FromState)
				assert.Equal(t, tt.expectedState, last.ToState)
				assert.Equal(t, "lifecycle_optimistic", last.Reason)
			} else {
				assert.Empty(t, inst.StateHistory, "no-op should not add history")
			}
		})
	}

	t.Run("nonexistent instance is no-op", func(t *testing.T) {
		server := createTestServer(t)
		// Should not panic
		server.applyExpectedTransitionalState("does-not-exist", "pending")
	})
}
