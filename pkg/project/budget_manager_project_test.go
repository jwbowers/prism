package project

// Tests for BudgetManager project enrichment via SetProjectManager.
//
// Verifies that GetBudgetSummary.ProjectNames and
// GetProjectFundingSummary.ProjectName / DefaultAllocationID are populated
// when a project manager is injected.

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupEnrichedManagers creates a BudgetManager and a Manager sharing the same
// temp state dir, then injects the Manager into the BudgetManager.
func setupEnrichedManagers(t *testing.T) (*BudgetManager, *Manager) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "prism-enriched-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	orig := os.Getenv("PRISM_STATE_DIR")
	t.Cleanup(func() {
		if orig == "" {
			_ = os.Unsetenv("PRISM_STATE_DIR")
		} else {
			_ = os.Setenv("PRISM_STATE_DIR", orig)
		}
	})
	_ = os.Setenv("PRISM_STATE_DIR", tempDir)

	bm, err := NewBudgetManager()
	require.NoError(t, err)

	pm, err := NewManager()
	require.NoError(t, err)

	bm.SetProjectManager(pm)
	return bm, pm
}

// TestGetBudgetSummary_ProjectNamesPopulated verifies that project names appear
// in the summary when SetProjectManager has been called.
func TestGetBudgetSummary_ProjectNamesPopulated(t *testing.T) {
	bm, pm := setupEnrichedManagers(t)
	ctx := context.Background()

	// Create a project.
	proj, err := pm.CreateProject(ctx, &CreateProjectRequest{
		Name:        "Enrichment Test Project",
		Description: "for budget summary test",
		Owner:       "test-owner",
	})
	require.NoError(t, err)

	// Create a budget.
	budget, err := bm.CreateBudget(ctx, &CreateBudgetRequest{
		Name:        "Test Budget",
		TotalAmount: 5000.0,

		Period:    "monthly",
		CreatedBy: "test-owner",
	})
	require.NoError(t, err)

	// Allocate budget to the project.
	_, err = bm.CreateAllocation(ctx, &CreateAllocationRequest{
		BudgetID:        budget.ID,
		ProjectID:       proj.ID,
		AllocatedAmount: 1000.0,
		AllocatedBy:     "test-owner",
	})
	require.NoError(t, err)

	// GetBudgetSummary should now include the project name.
	summary, err := bm.GetBudgetSummary(ctx, budget.ID)
	require.NoError(t, err)

	assert.Equal(t, "Enrichment Test Project", summary.ProjectNames[proj.ID],
		"ProjectNames should map project ID to its name")
}

// TestGetProjectFundingSummary_NameAndDefaultPopulated verifies that ProjectName
// and DefaultAllocationID are populated when a project manager is injected.
func TestGetProjectFundingSummary_NameAndDefaultPopulated(t *testing.T) {
	bm, pm := setupEnrichedManagers(t)
	ctx := context.Background()

	proj, err := pm.CreateProject(ctx, &CreateProjectRequest{
		Name:        "Funding Summary Project",
		Description: "for funding summary test",
		Owner:       "test-owner",
	})
	require.NoError(t, err)

	budget, err := bm.CreateBudget(ctx, &CreateBudgetRequest{
		Name:        "Funding Budget",
		TotalAmount: 8000.0,

		Period:    "monthly",
		CreatedBy: "test-owner",
	})
	require.NoError(t, err)

	alloc, err := bm.CreateAllocation(ctx, &CreateAllocationRequest{
		BudgetID:        budget.ID,
		ProjectID:       proj.ID,
		AllocatedAmount: 2000.0,
		AllocatedBy:     "test-owner",
	})
	require.NoError(t, err)

	// Set default allocation on the project.
	require.NoError(t, pm.SetDefaultAllocation(ctx, proj.ID, alloc.ID))

	// GetProjectFundingSummary should include name and default allocation.
	summary, err := bm.GetProjectFundingSummary(ctx, proj.ID)
	require.NoError(t, err)

	assert.Equal(t, "Funding Summary Project", summary.ProjectName,
		"ProjectName should be populated from project manager")
	require.NotNil(t, summary.DefaultAllocationID,
		"DefaultAllocationID should be populated from project")
	assert.Equal(t, alloc.ID, *summary.DefaultAllocationID)
}

// TestGetBudgetSummary_WithoutProjectManager verifies backward-compat:
// empty ProjectNames when no project manager is set.
func TestGetBudgetSummary_WithoutProjectManager(t *testing.T) {
	bm := setupTestBudgetManager(t)
	ctx := context.Background()

	budget, err := bm.CreateBudget(ctx, &CreateBudgetRequest{
		Name:        "No-PM Budget",
		TotalAmount: 1000.0,

		Period:    "monthly",
		CreatedBy: "owner",
	})
	require.NoError(t, err)

	summary, err := bm.GetBudgetSummary(ctx, budget.ID)
	require.NoError(t, err)

	assert.Empty(t, summary.ProjectNames, "ProjectNames should be empty without project manager")
}
