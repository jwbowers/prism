package fixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/project"
	"github.com/scttfrdmn/prism/pkg/types"
)

// CreateTestProjectOptions contains options for creating a test project
type CreateTestProjectOptions struct {
	Name        string
	Description string
	Owner       string
	Budget      *TestBudgetOptions
}

// TestBudgetOptions contains options for project budget configuration
type TestBudgetOptions struct {
	TotalBudget     float64
	MonthlyLimit    *float64
	AlertThresholds []types.BudgetAlert
	AutoActions     []types.BudgetAutoAction
	BudgetPeriod    types.BudgetPeriod
}

// CreateTestProject creates a test project for integration tests
// The project is automatically registered for cleanup via the registry
func CreateTestProject(t *testing.T, registry *FixtureRegistry, opts CreateTestProjectOptions) (*types.Project, error) {
	t.Helper()

	// Set defaults
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("test-project-%d", time.Now().Unix())
	}
	if opts.Description == "" {
		opts.Description = "Integration test project"
	}
	if opts.Owner == "" {
		opts.Owner = "test-user@example.com"
	}

	ctx := context.Background()

	// Create project request
	createReq := project.CreateProjectRequest{
		Name:        opts.Name,
		Description: opts.Description,
		Owner:       opts.Owner,
	}

	t.Logf("Creating test project: %s", opts.Name)
	proj, err := registry.client.CreateProject(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Register for cleanup
	registry.Register("project", proj.ID)
	t.Logf("Project created: %s (ID: %s)", opts.Name, proj.ID)

	// Set budget if specified
	if opts.Budget != nil {
		t.Logf("Setting project budget: $%.2f", opts.Budget.TotalBudget)

		budgetReq := client.SetProjectBudgetRequest{
			TotalBudget:     opts.Budget.TotalBudget,
			MonthlyLimit:    opts.Budget.MonthlyLimit,
			AlertThresholds: opts.Budget.AlertThresholds,
			AutoActions:     opts.Budget.AutoActions,
			BudgetPeriod:    opts.Budget.BudgetPeriod,
		}

		_, err = registry.client.SetProjectBudget(ctx, proj.ID, budgetReq)
		if err != nil {
			return nil, fmt.Errorf("failed to set project budget: %w", err)
		}

		t.Logf("Budget configured successfully")
	}

	return proj, nil
}

// CreateTestProjectMember adds a member to a test project
func CreateTestProjectMember(t *testing.T, registry *FixtureRegistry, projectID, userID string, role types.ProjectRole) error {
	t.Helper()

	ctx := context.Background()

	addMemberReq := project.AddMemberRequest{
		UserID:  userID,
		Role:    role,
		AddedBy: "test-admin",
	}

	t.Logf("Adding project member: %s (role: %s)", userID, role)
	err := registry.client.AddProjectMember(ctx, projectID, addMemberReq)
	if err != nil {
		return fmt.Errorf("failed to add project member: %w", err)
	}

	t.Logf("Member added successfully")
	return nil
}
