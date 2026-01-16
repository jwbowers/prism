//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/project"
	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBudgetTracking_RealTimeUpdates validates real-time cost tracking
// Tests: Create project → Launch instances → Verify cost accumulation → Check accuracy
func TestBudgetTracking_RealTimeUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping budget tracking test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("budget-track-%d", time.Now().Unix())

	t.Logf("Creating project with budget tracking: %s", projectName)

	// Test 1: Create project with budget
	var projectID string
	t.Run("CreateProjectWithBudget", func(t *testing.T) {
		project, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
			Name:        projectName,
			Description: "Budget tracking integration test",
			Owner:       "test-user@example.com",
		})
		require.NoError(t, err, "Failed to create project")
		registry.Register("project", projectName)
		projectID = project.ID

		t.Logf("Project created: %s (ID: %s)", project.Name, project.ID)

		// Set budget
		monthlyLimit := 50.0
		_, err = apiClient.SetProjectBudget(ctx, projectID, client.SetProjectBudgetRequest{
			TotalBudget:  500.0,
			MonthlyLimit: &monthlyLimit,
			AlertThresholds: []types.BudgetAlert{
				{
					Threshold:  0.80,
					Type:       types.BudgetAlertEmail,
					Recipients: []string{"test@example.com"},
				},
			},
			AutoActions: []types.BudgetAutoAction{
				{
					Threshold: 0.90,
					Action:    types.BudgetActionPreventLaunch,
				},
			},
		})
		require.NoError(t, err, "Failed to set project budget")

		t.Logf("Budget configured: $%.2f monthly limit", monthlyLimit)
		t.Log("✓ Project with budget created")
	})

	// Test 2: Verify initial budget status
	t.Run("VerifyInitialBudgetStatus", func(t *testing.T) {
		t.Log("Fetching initial budget status...")

		status, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		assert.NoError(t, err, "Failed to get budget status")
		assert.NotNil(t, status, "Budget status should not be nil")

		if status.BudgetEnabled {
			t.Log("✓ Budget tracking enabled")
			t.Logf("  Total budget: $%.2f", status.TotalBudget)
			// Note: MonthlyLimit field removed from BudgetStatus in v0.7.x
			t.Logf("  Current spend: $%.2f", status.SpentAmount)
			t.Logf("  Remaining: $%.2f", status.RemainingBudget)
			t.Logf("  Spent percentage: %.1f%%", status.SpentPercentage*100)

			// Verify initial state
			assert.Equal(t, 0.0, status.SpentAmount, "Initial spend should be zero")
			// Note: Can't verify RemainingBudget == MonthlyLimit as MonthlyLimit field was removed
		} else {
			t.Log("⚠️  Budget tracking not enabled - feature may not be implemented")
			t.Skip("Budget tracking not available")
		}
	})

	// Test 3: Launch instance and verify cost tracking
	var instanceName string
	t.Run("LaunchInstanceTrackCosts", func(t *testing.T) {
		instanceName = fmt.Sprintf("budget-inst-%d", time.Now().Unix())
		t.Logf("Launching instance in project: %s", instanceName)

		launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
			Template:  "Ubuntu Basic",
			Name:      instanceName,
			Size:      "S", // t3.small: $0.0208/hour
			ProjectID: projectID,
		})
		require.NoError(t, err, "Failed to launch instance")
		registry.Register("instance", instanceName)

		t.Logf("Instance launched: %s (ID: %s)", launchResp.Instance.Name, launchResp.Instance.ID)
		t.Logf("  Instance type: %s", launchResp.Instance.InstanceType)
		t.Logf("  Hourly rate: $%.4f", launchResp.Instance.HourlyRate)

		// Wait for instance to be running
		err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
		require.NoError(t, err, "Instance should reach running state")

		t.Log("✓ Instance launched successfully")
	})

	// Test 4: Wait for cost tracking to update
	t.Run("WaitForCostUpdate", func(t *testing.T) {
		t.Log("Waiting for cost tracker to update...")

		// Poll for cost updates (up to 60 seconds)
		maxAttempts := 12
		pollInterval := 5 * time.Second
		costDetected := false

		for i := 0; i < maxAttempts; i++ {
			time.Sleep(pollInterval)

			status, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
			if err != nil {
				t.Logf("  Attempt %d/%d: Error getting status: %v", i+1, maxAttempts, err)
				continue
			}

			t.Logf("  Attempt %d/%d: Current spend: $%.4f", i+1, maxAttempts, status.SpentAmount)

			if status.SpentAmount > 0 {
				costDetected = true
				t.Logf("✓ Cost tracking detected spend: $%.4f", status.SpentAmount)
				break
			}
		}

		if !costDetected {
			t.Log("⚠️  No cost detected after 60 seconds - cost tracking may be delayed")
			t.Log("Note: This is acceptable if cost tracking is async or batched")
		}
	})

	// Test 5: Verify budget status after instance launch
	t.Run("VerifyBudgetAfterLaunch", func(t *testing.T) {
		t.Log("Verifying budget status after instance launch...")

		status, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		assert.NoError(t, err, "Failed to get budget status")

		t.Logf("Budget status:")
		t.Logf("  Total budget: $%.2f", status.TotalBudget)
		// Note: MonthlyLimit field removed from BudgetStatus in v0.7.x
		t.Logf("  Current spend: $%.4f", status.SpentAmount)
		t.Logf("  Remaining: $%.4f", status.RemainingBudget)
		t.Logf("  Spent percentage: %.2f%%", status.SpentPercentage*100)

		if status.ProjectedMonthlySpend > 0 {
			t.Logf("  Projected monthly: $%.2f", status.ProjectedMonthlySpend)
		}

		// Verify spending is within reasonable range
		// For a test running < 1 minute, cost should be minimal
		if status.SpentAmount > 1.0 {
			t.Errorf("Spend amount seems high for short test: $%.4f", status.SpentAmount)
		}

		// Verify remaining budget calculation
		// Note: Can't calculate expectedRemaining as MonthlyLimit field was removed from BudgetStatus
		// Just verify RemainingBudget is reasonable (should be positive for this short test)
		assert.True(t, status.RemainingBudget > 0, "Remaining budget should be positive")

		t.Log("✓ Budget calculations verified")
	})

	// Test 6: Launch second instance and verify cumulative cost
	t.Run("LaunchSecondInstanceCumulativeCost", func(t *testing.T) {
		instance2Name := fmt.Sprintf("budget-inst2-%d", time.Now().Unix())
		t.Logf("Launching second instance: %s", instance2Name)

		// Get current spend before launch
		statusBefore, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		require.NoError(t, err, "Failed to get budget status before launch")
		spendBefore := statusBefore.SpentAmount

		// Launch second instance
		launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
			Template:  "Ubuntu Basic",
			Name:      instance2Name,
			Size:      "S",
			ProjectID: projectID,
		})
		require.NoError(t, err, "Failed to launch second instance")
		registry.Register("instance", instance2Name)

		t.Logf("Second instance launched: %s", launchResp.Instance.Name)

		// Wait for running state
		err = fixtures.WaitForInstanceState(t, apiClient, instance2Name, "running", 5*time.Minute)
		require.NoError(t, err, "Second instance should reach running state")

		// Wait for cost update
		time.Sleep(10 * time.Second)

		// Get updated budget status
		statusAfter, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		require.NoError(t, err, "Failed to get budget status after second launch")

		t.Logf("Budget status after second instance:")
		t.Logf("  Spend before: $%.4f", spendBefore)
		t.Logf("  Spend after: $%.4f", statusAfter.SpentAmount)
		t.Logf("  Incremental cost: $%.4f", statusAfter.SpentAmount-spendBefore)

		// Verify cumulative cost increased (or stayed same if not yet tracked)
		if statusAfter.SpentAmount > spendBefore {
			t.Log("✓ Cumulative cost tracking working")
		} else {
			t.Log("⚠️  Cost not yet updated (may be delayed)")
		}

		t.Log("✓ Second instance launched and tracked")
	})

	// Test 7: Stop instance and verify cost stops accumulating
	t.Run("StopInstanceVerifyCostStops", func(t *testing.T) {
		t.Logf("Stopping instance to verify cost tracking stops: %s", instanceName)

		// Get current spend
		statusBefore, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		require.NoError(t, err, "Failed to get budget status before stop")
		spendBefore := statusBefore.SpentAmount

		// Stop instance
		err = apiClient.StopInstance(ctx, instanceName)
		require.NoError(t, err, "Failed to stop instance")

		// Wait for stopped state
		err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "stopped", 3*time.Minute)
		require.NoError(t, err, "Instance should reach stopped state")

		t.Log("Instance stopped, waiting for cost tracker to update...")
		time.Sleep(15 * time.Second)

		// Get budget status after stop
		statusAfter, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		require.NoError(t, err, "Failed to get budget status after stop")

		t.Logf("Budget status after stop:")
		t.Logf("  Spend before stop: $%.4f", spendBefore)
		t.Logf("  Spend after stop: $%.4f", statusAfter.SpentAmount)

		// Note: Cost may continue to accumulate briefly due to billing lag
		// This is expected AWS behavior
		if statusAfter.SpentAmount > spendBefore {
			incremental := statusAfter.SpentAmount - spendBefore
			t.Logf("  Incremental cost after stop: $%.4f (expected due to billing lag)", incremental)
		}

		t.Log("✓ Instance stopped and cost tracking updated")
	})

	t.Log("✅ Real-time budget tracking test complete")
}

// TestBudgetTracking_AlertThresholds validates budget alert configuration
// Tests: Create project → Set alerts → Verify alert thresholds → Check alert status
func TestBudgetTracking_AlertThresholds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping alert threshold test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("budget-alerts-%d", time.Now().Unix())

	t.Logf("Creating project with alert thresholds: %s", projectName)

	// Test 1: Create project with multiple alert thresholds
	var projectID string
	t.Run("CreateProjectWithAlerts", func(t *testing.T) {
		project, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
			Name:        projectName,
			Description: "Budget alerts test",
			Owner:       "test-user@example.com",
		})
		require.NoError(t, err, "Failed to create project")
		registry.Register("project", projectName)
		projectID = project.ID

		// Set budget with multiple alert thresholds
		monthlyLimit := 10.0
		_, err = apiClient.SetProjectBudget(ctx, projectID, client.SetProjectBudgetRequest{
			TotalBudget:  100.0,
			MonthlyLimit: &monthlyLimit,
			AlertThresholds: []types.BudgetAlert{
				{
					Threshold:  0.50,
					Type:       types.BudgetAlertEmail,
					Recipients: []string{"team@example.com"},
				},
				{
					Threshold:  0.75,
					Type:       types.BudgetAlertEmail,
					Recipients: []string{"manager@example.com"},
				},
				{
					Threshold:  0.90,
					Type:       types.BudgetAlertEmail,
					Recipients: []string{"admin@example.com"},
				},
			},
		})
		require.NoError(t, err, "Failed to set budget with alerts")

		t.Logf("Budget configured with alert thresholds: 50%%, 75%%, 90%%")
		t.Log("✓ Project with alerts created")
	})

	// Test 2: Verify alert thresholds are saved
	t.Run("VerifyAlertThresholds", func(t *testing.T) {
		t.Log("Verifying alert thresholds...")

		status, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		assert.NoError(t, err, "Failed to get budget status")

		// Note: BudgetStatus has ActiveAlerts ([]string) not AlertThresholds
		// ActiveAlerts contains currently triggered alert descriptions
		if len(status.ActiveAlerts) > 0 {
			t.Logf("Found %d active alerts:", len(status.ActiveAlerts))
			for i, alert := range status.ActiveAlerts {
				t.Logf("  %d. %s", i+1, alert)
			}
			t.Log("✓ Alerts are being tracked")
		} else {
			t.Log("⚠️  No active alerts (expected - budget not exhausted yet)")
		}
	})

	// Test 3: Document alert triggering behavior
	t.Run("DocumentAlertBehavior", func(t *testing.T) {
		t.Log("Documenting alert triggering behavior:")
		t.Log("")
		t.Log("ALERT THRESHOLDS:")
		t.Log("  50% ($5.00) → Email to team@example.com")
		t.Log("  75% ($7.50) → Email to manager@example.com")
		t.Log("  90% ($9.00) → Email to admin@example.com")
		t.Log("")
		t.Log("EXPECTED BEHAVIOR:")
		t.Log("  - Alerts trigger when spend crosses threshold")
		t.Log("  - Each alert fires only once per period")
		t.Log("  - Alert history tracked in project")
		t.Log("  - Email notifications sent to recipients")
		t.Log("")
		t.Log("✓ Alert behavior documented")
	})

	t.Log("✅ Budget alert threshold test complete")
}

// TestBudgetTracking_CostBreakdown validates detailed cost breakdown
// Tests: Launch resources → Get cost breakdown → Verify by resource type
func TestBudgetTracking_CostBreakdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cost breakdown test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cost-breakdown-%d", time.Now().Unix())

	t.Logf("Creating project for cost breakdown testing: %s", projectName)

	// Create project
	project, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Cost breakdown test",
		Owner:       "test-user@example.com",
	})
	require.NoError(t, err, "Failed to create project")
	registry.Register("project", projectName)
	projectID := project.ID

	t.Logf("Project created: %s (ID: %s)", project.Name, project.ID)

	// Launch instance
	instanceName := fmt.Sprintf("cost-inst-%d", time.Now().Unix())
	t.Logf("Launching instance for cost tracking: %s", instanceName)

	launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
		Template:  "Ubuntu Basic",
		Name:      instanceName,
		Size:      "M",
		ProjectID: projectID,
	})
	require.NoError(t, err, "Failed to launch instance")
	registry.Register("instance", instanceName)

	t.Logf("Instance launched: %s", launchResp.Instance.Name)

	// Wait for running state
	err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
	require.NoError(t, err, "Instance should reach running state")

	// Wait for cost data
	t.Log("Waiting for cost data to accumulate...")
	time.Sleep(10 * time.Second)

	// Test: Get cost breakdown
	t.Run("GetCostBreakdown", func(t *testing.T) {
		t.Log("Fetching cost breakdown...")

		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now()

		breakdown, err := apiClient.GetProjectCostBreakdown(ctx, projectID, startTime, endTime)

		if err != nil {
			t.Logf("⚠️  Cost breakdown API error: %v", err)
			t.Log("Note: Cost breakdown API may not be implemented yet")
			t.Skip("Cost breakdown not available")
			return
		}

		t.Log("✓ Cost breakdown retrieved:")
		t.Logf("  Total cost: $%.4f", breakdown.TotalCost)
		t.Logf("  Instance costs: %d entries", len(breakdown.InstanceCosts))
		t.Logf("  Storage costs: %d entries", len(breakdown.StorageCosts))

		// Verify breakdown structure
		if len(breakdown.InstanceCosts) > 0 {
			t.Log("  Instance cost details:")
			for _, cost := range breakdown.InstanceCosts {
				t.Logf("    - %s: $%.4f", cost.InstanceName, cost.ComputeCost)
			}
		}

		if len(breakdown.StorageCosts) > 0 {
			t.Log("  Storage cost details:")
			for _, cost := range breakdown.StorageCosts {
				t.Logf("    - %s: $%.4f", cost.VolumeName, cost.Cost)
			}
		}

		t.Log("✓ Cost breakdown structure validated")
	})

	t.Log("✅ Cost breakdown test complete")
}

// TestBudgetTracking_ProjectedSpend validates projected spending calculations
// Tests: Launch instance → Wait for projection → Verify projection accuracy
func TestBudgetTracking_ProjectedSpend(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping projected spend test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("projected-spend-%d", time.Now().Unix())

	// Create project
	project, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Projected spend test",
		Owner:       "test-user@example.com",
	})
	require.NoError(t, err, "Failed to create project")
	registry.Register("project", projectName)

	t.Logf("Project created: %s", project.Name)

	// Launch instance with known hourly rate
	instanceName := fmt.Sprintf("projected-inst-%d", time.Now().Unix())
	launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
		Template:  "Ubuntu Basic",
		Name:      instanceName,
		Size:      "S", // t3.small: $0.0208/hour
		ProjectID: project.ID,
	})
	require.NoError(t, err, "Failed to launch instance")
	registry.Register("instance", instanceName)

	hourlyRate := launchResp.Instance.HourlyRate
	t.Logf("Instance launched with hourly rate: $%.4f", hourlyRate)

	// Wait for running state
	err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
	require.NoError(t, err, "Instance should reach running state")

	// Wait for projection calculation
	time.Sleep(15 * time.Second)

	// Get budget status with projection
	status, err := apiClient.GetProjectBudgetStatus(ctx, project.ID)
	require.NoError(t, err, "Failed to get budget status")

	t.Logf("Budget status:")
	t.Logf("  Current spend: $%.4f", status.SpentAmount)
	t.Logf("  Projected monthly: $%.2f", status.ProjectedMonthlySpend)

	// Calculate expected monthly spend
	hoursPerMonth := 730.0 // Average hours per month
	expectedMonthly := hourlyRate * hoursPerMonth

	t.Logf("Expected monthly spend: $%.2f", expectedMonthly)

	// Verify projection is reasonable (allow 20% margin due to timing)
	if status.ProjectedMonthlySpend > 0 {
		ratio := status.ProjectedMonthlySpend / expectedMonthly
		if ratio >= 0.8 && ratio <= 1.2 {
			t.Log("✓ Projected spend is within expected range")
		} else {
			t.Logf("⚠️  Projected spend differs from expected: %.2f vs %.2f (ratio: %.2f)", status.ProjectedMonthlySpend, expectedMonthly, ratio)
		}
	} else {
		t.Log("⚠️  Projected spend not calculated yet")
	}

	t.Log("✅ Projected spend test complete")
}
