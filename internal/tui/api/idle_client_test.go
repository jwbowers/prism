package api

// Tests for TUI idle detection methods — verifies that the 4 methods
// (EnableIdleDetection, DisableIdleDetection, GetInstanceIdleStatus, UpdateIdlePolicy)
// correctly delegate to the underlying pkg/api/client HTTP client.

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apimock "github.com/scttfrdmn/prism/pkg/api/mock"
	"github.com/scttfrdmn/prism/pkg/idle"
)

// idleTrackingClient embeds the full MockClient and overrides the idle methods
// so tests can record calls and control return values.
type idleTrackingClient struct {
	apimock.MockClient
	applyCalls  []struct{ instance, policy string }
	removeCalls []struct{ instance, policy string }
	policies    []*idle.PolicyTemplate
	applyErr    error
	removeErr   error
	listErr     error
}

func (m *idleTrackingClient) ApplyIdlePolicy(_ context.Context, instance, policy string) error {
	m.applyCalls = append(m.applyCalls, struct{ instance, policy string }{instance, policy})
	return m.applyErr
}

func (m *idleTrackingClient) RemoveIdlePolicy(_ context.Context, instance, policy string) error {
	m.removeCalls = append(m.removeCalls, struct{ instance, policy string }{instance, policy})
	return m.removeErr
}

func (m *idleTrackingClient) GetInstanceIdlePolicies(_ context.Context, _ string) ([]*idle.PolicyTemplate, error) {
	return m.policies, m.listErr
}

func newIdleTrackingTUIClient(mock *idleTrackingClient) *TUIClient {
	return &TUIClient{client: mock}
}

// TestEnableIdleDetection_DelegatesToApplyIdlePolicy verifies that EnableIdleDetection
// calls the underlying ApplyIdlePolicy with the correct arguments.
func TestEnableIdleDetection_DelegatesToApplyIdlePolicy(t *testing.T) {
	mock := &idleTrackingClient{}
	c := newIdleTrackingTUIClient(mock)

	err := c.EnableIdleDetection(context.Background(), "my-instance", "balanced")
	require.NoError(t, err)
	require.Len(t, mock.applyCalls, 1)
	assert.Equal(t, "my-instance", mock.applyCalls[0].instance)
	assert.Equal(t, "balanced", mock.applyCalls[0].policy)
}

// TestEnableIdleDetection_PropagatesError ensures errors from ApplyIdlePolicy bubble up.
func TestEnableIdleDetection_PropagatesError(t *testing.T) {
	mock := &idleTrackingClient{applyErr: errors.New("apply failed")}
	c := newIdleTrackingTUIClient(mock)

	err := c.EnableIdleDetection(context.Background(), "inst", "policy")
	assert.EqualError(t, err, "apply failed")
}

// TestDisableIdleDetection_RemovesAllPolicies verifies that DisableIdleDetection
// calls RemoveIdlePolicy once per currently-applied policy.
func TestDisableIdleDetection_RemovesAllPolicies(t *testing.T) {
	mock := &idleTrackingClient{
		policies: []*idle.PolicyTemplate{
			{ID: "policy-a"},
			{ID: "policy-b"},
		},
	}
	c := newIdleTrackingTUIClient(mock)

	err := c.DisableIdleDetection(context.Background(), "my-instance")
	require.NoError(t, err)
	require.Len(t, mock.removeCalls, 2)
	assert.Equal(t, "policy-a", mock.removeCalls[0].policy)
	assert.Equal(t, "policy-b", mock.removeCalls[1].policy)
}

// TestDisableIdleDetection_NoPolicies verifies no error when instance has no policies.
func TestDisableIdleDetection_NoPolicies(t *testing.T) {
	mock := &idleTrackingClient{policies: nil}
	c := newIdleTrackingTUIClient(mock)

	err := c.DisableIdleDetection(context.Background(), "my-instance")
	assert.NoError(t, err)
	assert.Empty(t, mock.removeCalls)
}

// TestGetInstanceIdleStatus_EnabledWhenPoliciesExist verifies that Enabled is true
// and Policy/Threshold are populated when the instance has applied policies.
func TestGetInstanceIdleStatus_EnabledWhenPoliciesExist(t *testing.T) {
	mock := &idleTrackingClient{
		policies: []*idle.PolicyTemplate{
			{
				ID: "balanced",
				Schedules: []idle.Schedule{
					{IdleMinutes: 30},
				},
			},
		},
	}
	c := newIdleTrackingTUIClient(mock)

	status, err := c.GetInstanceIdleStatus(context.Background(), "my-instance")
	require.NoError(t, err)
	assert.True(t, status.Enabled)
	assert.Equal(t, "balanced", status.Policy)
	assert.Equal(t, 30, status.Threshold)
}

// TestGetInstanceIdleStatus_DisabledWhenNoPolicies verifies Enabled is false
// when no policies are applied.
func TestGetInstanceIdleStatus_DisabledWhenNoPolicies(t *testing.T) {
	mock := &idleTrackingClient{policies: nil}
	c := newIdleTrackingTUIClient(mock)

	status, err := c.GetInstanceIdleStatus(context.Background(), "my-instance")
	require.NoError(t, err)
	assert.False(t, status.Enabled)
	assert.Empty(t, status.Policy)
}

// TestGetInstanceIdleStatus_PropagatesError ensures list errors bubble up.
func TestGetInstanceIdleStatus_PropagatesError(t *testing.T) {
	mock := &idleTrackingClient{listErr: errors.New("daemon unavailable")}
	c := newIdleTrackingTUIClient(mock)

	_, err := c.GetInstanceIdleStatus(context.Background(), "my-instance")
	assert.EqualError(t, err, "daemon unavailable")
}

// TestUpdateIdlePolicy_IsNoOp verifies the intentional no-op behaviour.
func TestUpdateIdlePolicy_IsNoOp(t *testing.T) {
	mock := &idleTrackingClient{}
	c := newIdleTrackingTUIClient(mock)

	err := c.UpdateIdlePolicy(context.Background(), IdlePolicyUpdateRequest{
		Name:      "my-policy",
		Threshold: 45,
		Action:    "stop",
	})
	assert.NoError(t, err)
	assert.Empty(t, mock.applyCalls, "UpdateIdlePolicy should not call ApplyIdlePolicy")
}
