// Package cli tests for v0.12.0 budget governance subcommands.
package cli

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// budgetTestApp returns an App wired to the default MockAPIClient.
// The default mock has a project with ID="test-project".
// Budget commands call GetProject(ctx, name) where name is matched against proj.ID,
// so tests must pass "test-project" as the project argument.
func budgetTestApp(t *testing.T) (*App, *MockAPIClient) {
	t.Helper()
	t.Setenv("PRISM_NO_AUTO_START", "1")
	mock := NewMockAPIClient()
	app := NewAppWithClient("1.0.0", mock)
	return app, mock
}

// ── rollover ────────────────────────────────────────────────────────────────

func TestBudgetCommands_Rollover_MissingArg(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createRolloverCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// No args → cobra ExactArgs(1) validation error
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestBudgetCommands_Rollover_Enable(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createRolloverCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"test-project", "--enable"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestBudgetCommands_Rollover_WithCap(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createRolloverCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"test-project", "--enable", "--cap", "500"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestBudgetCommands_Rollover_Disable(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createRolloverCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"test-project", "--disable"})
	err := cmd.Execute()
	require.NoError(t, err)
}

// ── share ────────────────────────────────────────────────────────────────────

func TestBudgetCommands_Share_MissingFlags(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createShareCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// No --from/--to/--amount → RunE returns validation error
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestBudgetCommands_Share_Success(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createShareCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// "test-project" exists in the mock (matched by ID)
	cmd.SetArgs([]string{"--from", "test-project", "--to", "proj-b", "--amount", "100"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestBudgetCommands_Share_MissingFrom(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createShareCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--to", "proj-b", "--amount", "100"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestBudgetCommands_Share_APIError(t *testing.T) {
	app, mock := budgetTestApp(t)
	mock.ShouldReturnError = true
	mock.ErrorMessage = "budget share failed"
	bc := NewBudgetCommands(app)
	cmd := bc.createShareCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// ShouldReturnError causes GetProject to fail first
	cmd.SetArgs([]string{"--from", "test-project", "--to", "proj-b", "--amount", "100"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// ── reallocate ───────────────────────────────────────────────────────────────

func TestBudgetCommands_Reallocate_MissingFlags(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createReallocateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// No flags → validation error from RunE
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestBudgetCommands_Reallocate_Success(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createReallocateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// --project "test-project" exists in mock
	cmd.SetArgs([]string{
		"--from", "user1",
		"--to", "user2",
		"--amount", "100",
		"--project", "test-project",
	})
	err := cmd.Execute()
	require.NoError(t, err)
}

// ── borrow ───────────────────────────────────────────────────────────────────

func TestBudgetCommands_Borrow_MissingExpires(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createBorrowCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// No --expires → validation error from RunE
	cmd.SetArgs([]string{"--from", "test-project", "--amount", "100"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestBudgetCommands_Borrow_Success(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createBorrowCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--from", "test-project", "--amount", "100", "--expires", "14d"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestBudgetCommands_Borrow_InvalidExpires(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createBorrowCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--from", "test-project", "--amount", "100", "--expires", "badvalue"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// ── emergency ────────────────────────────────────────────────────────────────

func TestBudgetCommands_Emergency_MissingAmount(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createEmergencyCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// No --amount or --reason → validation error from RunE
	cmd.SetArgs([]string{"test-project"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestBudgetCommands_Emergency_MissingReason(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createEmergencyCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// --amount provided but no --reason
	cmd.SetArgs([]string{"test-project", "--amount", "500"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestBudgetCommands_Emergency_Success(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createEmergencyCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"test-project", "--amount", "500", "--reason", "critical deadline"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestBudgetCommands_Emergency_MissingArg(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createEmergencyCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// No positional arg → cobra ExactArgs(1) error
	cmd.SetArgs([]string{"--amount", "500", "--reason", "deadline"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// ── report ───────────────────────────────────────────────────────────────────

func TestBudgetCommands_Report_DefaultMonth(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createReportCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// No --month → defaults to current month
	cmd.SetArgs([]string{"test-project"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestBudgetCommands_Report_FormatJSON(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createReportCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"test-project", "--format", "json"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestBudgetCommands_Report_ExplicitMonth(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createReportCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"test-project", "--month", "2026-02"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestBudgetCommands_Report_APIError(t *testing.T) {
	app, mock := budgetTestApp(t)
	// Only make GetMonthlyReport fail, not GetProject.
	// We do this by wiring ShouldReturnError after the project lookup would succeed.
	// Since MockAPIClient.ShouldReturnError affects all methods, we test the error
	// propagation path by setting the error flag (it will fail at GetProject, which
	// is still an error propagation test for the report command).
	mock.ShouldReturnError = true
	mock.ErrorMessage = "report generation failed"
	bc := NewBudgetCommands(app)
	cmd := bc.createReportCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"test-project"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestBudgetCommands_Report_MissingArg(t *testing.T) {
	app, _ := budgetTestApp(t)
	bc := NewBudgetCommands(app)
	cmd := bc.createReportCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// No positional arg → cobra ExactArgs(1) error
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
}
