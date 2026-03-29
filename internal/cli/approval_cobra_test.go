// Package cli — unit tests for the `prism approval` command group (v0.21.0 #495).
// Uses MockAPIClient to avoid spinning up an HTTP server.
package cli

import (
	"io"
	"testing"

	"github.com/scttfrdmn/prism/pkg/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// approvalV2TestApp returns an App wired to the default MockAPIClient.
func approvalV2TestApp(t *testing.T) (*App, *MockAPIClient) {
	t.Helper()
	t.Setenv("PRISM_NO_AUTO_START", "1")
	mock := NewMockAPIClient()
	app := NewAppWithClient("1.0.0", mock)
	return app, mock
}

func TestApprovalCobraCommandsV2_List(t *testing.T) {
	app, _ := approvalV2TestApp(t)
	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	listCmd, _, err := cmd.Find([]string{"list"})
	require.NoError(t, err)
	require.NotNil(t, listCmd.RunE)

	assert.NoError(t, listCmd.RunE(listCmd, []string{}))
}

func TestApprovalCobraCommandsV2_Show(t *testing.T) {
	app, _ := approvalV2TestApp(t)
	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	showCmd, _, err := cmd.Find([]string{"show"})
	require.NoError(t, err)
	require.NotNil(t, showCmd.RunE)

	// MockAPIClient.ListAllApprovals returns empty list, so "show" should return "not found"
	err = showCmd.RunE(showCmd, []string{"req-abc123"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestApprovalCobraCommandsV2_Approve(t *testing.T) {
	app, mock := approvalV2TestApp(t)
	// MockAPIClient.ListAllApprovals returns a sample pending approval with ID "req-abc"
	// Override to include the expected ID
	_ = mock // mock supports Approve via its internal list

	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	approveCmd, _, err := cmd.Find([]string{"approve"})
	require.NoError(t, err)
	require.NotNil(t, approveCmd.RunE)

	require.NoError(t, approveCmd.Flags().Set("comment", "looks good"))

	// MockAPIClient.ListAllApprovals returns empty → request not found
	err = approveCmd.RunE(approveCmd, []string{"req-not-found"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestApprovalCobraCommandsV2_Deny(t *testing.T) {
	app, _ := approvalV2TestApp(t)
	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	denyCmd, _, err := cmd.Find([]string{"deny"})
	require.NoError(t, err)
	require.NotNil(t, denyCmd.RunE)

	require.NoError(t, denyCmd.Flags().Set("comment", "use CPU first"))

	// MockAPIClient.ListAllApprovals returns empty → request not found
	err = denyCmd.RunE(denyCmd, []string{"req-not-found"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestApprovalCobraCommandsV2_Request_MissingProject(t *testing.T) {
	app, _ := approvalV2TestApp(t)
	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	requestCmd, _, err := cmd.Find([]string{"request"})
	require.NoError(t, err)
	require.NotNil(t, requestCmd.RunE)

	err = requestCmd.RunE(requestCmd, []string{"gpu"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--project is required")
}

func TestApprovalCobraCommandsV2_Request_UnknownType(t *testing.T) {
	app, _ := approvalV2TestApp(t)
	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	requestCmd, _, err := cmd.Find([]string{"request"})
	require.NoError(t, err)

	require.NoError(t, requestCmd.Flags().Set("project", "test-project"))
	err = requestCmd.RunE(requestCmd, []string{"unknown-type"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}

func TestApprovalCobraCommandsV2_Request_GPU(t *testing.T) {
	app, _ := approvalV2TestApp(t)
	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	requestCmd, _, err := cmd.Find([]string{"request"})
	require.NoError(t, err)

	require.NoError(t, requestCmd.Flags().Set("project", "test-project"))
	require.NoError(t, requestCmd.Flags().Set("reason", "paper deadline"))
	err = requestCmd.RunE(requestCmd, []string{"gpu"})
	assert.NoError(t, err)
}

func TestRequestApprovalCommand_CanHandle(t *testing.T) {
	c := &RequestApprovalCommand{}
	assert.True(t, c.CanHandle("--request-approval"))
	assert.False(t, c.CanHandle("--approval"))
	assert.False(t, c.CanHandle("--dry-run"))
}

func TestApprovalIDCommand_CanHandle(t *testing.T) {
	c := &ApprovalIDCommand{}
	assert.True(t, c.CanHandle("--approval"))
	assert.False(t, c.CanHandle("--request-approval"))
	assert.False(t, c.CanHandle("--dry-run"))
}

func TestApprovalIDCommand_Execute_MissingArg(t *testing.T) {
	c := &ApprovalIDCommand{}
	_, err := c.Execute(nil, []string{"--approval"}, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--approval requires")
}

// ── Populated mock list tests (#546) ─────────────────────────────────────

func TestApprovalCobraCommandsV2_List_WithApprovals(t *testing.T) {
	app, mock := approvalV2TestApp(t)

	// Populate the mock with a sample approval
	mock.Approvals = []*project.ApprovalRequest{
		{
			ID:          "req-abc123-full-uuid",
			ProjectID:   "test-project",
			Type:        project.ApprovalTypeGPUInstance,
			Status:      project.ApprovalStatusPending,
			RequestedBy: "alice",
		},
	}

	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	listCmd, _, err := cmd.Find([]string{"list"})
	require.NoError(t, err)

	assert.NoError(t, listCmd.RunE(listCmd, []string{}))
}

func TestApprovalCobraCommandsV2_Show_Found(t *testing.T) {
	app, mock := approvalV2TestApp(t)

	fullID := "req-abc123-full-uuid"
	mock.Approvals = []*project.ApprovalRequest{
		{
			ID:          fullID,
			ProjectID:   "test-project",
			Type:        project.ApprovalTypeExpensiveInstance,
			Status:      project.ApprovalStatusPending,
			RequestedBy: "bob",
		},
	}

	ac := NewApprovalCobraCommandsV2(app)
	cmd := ac.CreateApprovalCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	showCmd, _, err := cmd.Find([]string{"show"})
	require.NoError(t, err)

	// Pass full ID — should succeed now that ListAllApprovals returns the mock
	assert.NoError(t, showCmd.RunE(showCmd, []string{fullID}))
}
