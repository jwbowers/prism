package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/idle"
	"github.com/scttfrdmn/prism/pkg/project"
	"github.com/scttfrdmn/prism/pkg/types"
)

// setupTestProject creates a test project in the server's project manager
// Returns the created project for use in tests
func setupTestProject(t *testing.T, server *Server, name string) *types.Project {
	t.Helper()

	if server.projectManager == nil {
		t.Skip("Project manager not initialized - skipping test")
		return nil
	}

	ctx := context.Background()

	// Create project with test defaults
	req := project.CreateProjectRequest{
		Name:        name,
		Description: "Test project for handler tests",
		Owner:       "test-owner@example.com",
	}

	proj, err := server.projectManager.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	t.Logf("Created test project: %s (ID: %s)", proj.Name, proj.ID)
	return proj
}

// setupTestBudget creates a test budget for a project
// Returns the budget configuration for use in tests
func setupTestBudget(t *testing.T, server *Server, projectID string, totalBudget float64) *types.ProjectBudget {
	t.Helper()

	if server.budgetManager == nil {
		t.Skip("Budget manager not initialized - skipping test")
		return nil
	}

	ctx := context.Background()

	// Create budget with test defaults
	budgetReq := project.SetBudgetRequest{
		TotalBudget:  totalBudget,
		MonthlyLimit: &totalBudget, // Same as total for simplicity
		BudgetPeriod: types.BudgetPeriodMonthly,
		AlertThresholds: []types.BudgetAlert{
			{
				Threshold:  0.8, // 80% warning
				Type:       types.BudgetAlertTypeWarning,
				Recipients: []string{"test@example.com"},
			},
			{
				Threshold:  0.95, // 95% critical
				Type:       types.BudgetAlertTypeCritical,
				Recipients: []string{"test@example.com"},
			},
		},
	}

	budget, err := server.budgetManager.SetBudget(ctx, projectID, budgetReq)
	if err != nil {
		t.Fatalf("Failed to setup test budget: %v", err)
	}

	t.Logf("Created test budget for project %s: $%.2f", projectID, totalBudget)
	return budget
}

// setupTestIdlePolicy creates a test idle policy
// Returns the policy ID for use in tests
func setupTestIdlePolicy(t *testing.T, server *Server, policyID string) *idle.Policy {
	t.Helper()

	if server.policyService == nil {
		t.Skip("Policy service not initialized - skipping test")
		return nil
	}

	ctx := context.Background()

	// Create idle policy with test defaults
	policy := &idle.Policy{
		ID:          policyID,
		Name:        "Test " + policyID + " Policy",
		Description: "Test policy for handler tests",
		IdleTimeout: 30 * time.Minute,
		Actions: []idle.PolicyAction{
			{
				Type:      idle.ActionTypeStop,
				Threshold: 30 * time.Minute,
			},
		},
		Schedule: &idle.PolicySchedule{
			Enabled:   true,
			StartTime: "08:00",
			EndTime:   "18:00",
			Timezone:  "America/Los_Angeles",
			Days:      []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := server.policyService.CreatePolicy(ctx, policy)
	if err != nil {
		t.Fatalf("Failed to setup test idle policy: %v", err)
	}

	t.Logf("Created test idle policy: %s", policyID)
	return policy
}

// setupTestSnapshot creates a test snapshot in the server's state
// Returns the snapshot for use in tests
func setupTestSnapshot(t *testing.T, server *Server, instanceName string) *types.Snapshot {
	t.Helper()

	if server.stateManager == nil {
		t.Skip("State manager not initialized - skipping test")
		return nil
	}

	// Create snapshot with test defaults
	snapshot := &types.Snapshot{
		ID:           "snap-test-12345",
		InstanceName: instanceName,
		Name:         instanceName + "-snapshot-1",
		Description:  "Test snapshot for handler tests",
		State:        types.SnapshotStateAvailable,
		VolumeID:     "vol-test-12345",
		VolumeSize:   100,
		StartTime:    time.Now().Add(-1 * time.Hour),
		Progress:     "100%",
		Encrypted:    true,
		Tags: map[string]string{
			"Name":        instanceName + "-snapshot-1",
			"Environment": "test",
			"CreatedBy":   "test-suite",
		},
	}

	// Store snapshot in state
	state, err := server.stateManager.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if state.Snapshots == nil {
		state.Snapshots = make(map[string]*types.Snapshot)
	}
	state.Snapshots[snapshot.ID] = snapshot

	err = server.stateManager.Save(state)
	if err != nil {
		t.Fatalf("Failed to save state with snapshot: %v", err)
	}

	t.Logf("Created test snapshot: %s for instance %s", snapshot.ID, instanceName)
	return snapshot
}

// setupTestInstance creates a test instance in the server's state
// Returns the instance for use in tests
func setupTestInstance(t *testing.T, server *Server, instanceName string) *types.Instance {
	t.Helper()

	if server.stateManager == nil {
		t.Skip("State manager not initialized - skipping test")
		return nil
	}

	// Create instance with test defaults
	instance := &types.Instance{
		ID:           "i-test-12345",
		Name:         instanceName,
		Template:     "test-template",
		InstanceType: "t3.micro",
		State:        "running",
		PublicIP:     "54.123.45.67",
		PrivateIP:    "10.0.1.10",
		LaunchTime:   time.Now().Add(-2 * time.Hour),
		Region:       "us-west-2",
		Tags: map[string]string{
			"Name":        instanceName,
			"Environment": "test",
		},
		EstimatedCost: 0.0104, // t3.micro hourly rate
	}

	// Store instance in state
	state, err := server.stateManager.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if state.Instances == nil {
		state.Instances = make(map[string]*types.Instance)
	}
	state.Instances[instanceName] = instance

	err = server.stateManager.Save(state)
	if err != nil {
		t.Fatalf("Failed to save state with instance: %v", err)
	}

	t.Logf("Created test instance: %s (ID: %s)", instanceName, instance.ID)
	return instance
}

// setupTestSecurityConfig creates test security configuration
// Returns the config for use in tests
func setupTestSecurityConfig(t *testing.T, server *Server) *types.SecurityConfig {
	t.Helper()

	if server.securityManager == nil {
		t.Skip("Security manager not initialized - skipping test")
		return nil
	}

	ctx := context.Background()

	// Create security config with test defaults
	config := &types.SecurityConfig{
		EnabledChecks: []string{
			"ssh_keys",
			"security_groups",
			"iam_roles",
			"encryption",
		},
		AutoRemediation: false,
		ScanInterval:    24 * time.Hour,
		Severity:        "medium",
	}

	err := server.securityManager.SetConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to setup test security config: %v", err)
	}

	t.Logf("Created test security config with %d checks", len(config.EnabledChecks))
	return config
}

// setupTestMarketplaceTemplate creates a test marketplace template tracking entry
// Returns the template metadata for use in tests
func setupTestMarketplaceTemplate(t *testing.T, server *Server, templateSlug string) *types.MarketplaceTemplate {
	t.Helper()

	if server.stateManager == nil {
		t.Skip("State manager not initialized - skipping test")
		return nil
	}

	// Create marketplace template tracking entry
	template := &types.MarketplaceTemplate{
		Slug:        templateSlug,
		Name:        "Test " + templateSlug,
		Description: "Test marketplace template",
		Author:      "test-author",
		Version:     "1.0.0",
		Downloads:   100,
		Rating:      4.5,
		Tags:        []string{"test", "development"},
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now(),
		Featured:    false,
	}

	// Store in state
	state, err := server.stateManager.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if state.MarketplaceTemplates == nil {
		state.MarketplaceTemplates = make(map[string]*types.MarketplaceTemplate)
	}
	state.MarketplaceTemplates[templateSlug] = template

	err = server.stateManager.Save(state)
	if err != nil {
		t.Fatalf("Failed to save state with marketplace template: %v", err)
	}

	t.Logf("Created test marketplace template: %s", templateSlug)
	return template
}

// setupTestProjectMember adds a test member to a project
// Returns the member for use in tests
func setupTestProjectMember(t *testing.T, server *Server, projectID, userID string, role types.ProjectRole) *types.ProjectMember {
	t.Helper()

	if server.projectManager == nil {
		t.Skip("Project manager not initialized - skipping test")
		return nil
	}

	ctx := context.Background()

	// Add member to project
	req := project.AddMemberRequest{
		UserID:  userID,
		Role:    role,
		AddedBy: "test-admin",
	}

	err := server.projectManager.AddMember(ctx, projectID, req)
	if err != nil {
		t.Fatalf("Failed to setup test project member: %v", err)
	}

	// Retrieve the added member
	members, err := server.projectManager.ListMembers(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to retrieve project members: %v", err)
	}

	// Find the member we just added
	for _, member := range members {
		if member.UserID == userID {
			t.Logf("Added test member %s to project %s with role %s", userID, projectID, role)
			return member
		}
	}

	t.Fatalf("Failed to find added member %s in project %s", userID, projectID)
	return nil
}
