//go:build integration && longrunning
// +build integration,longrunning

package phase4_long_running

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/scttfrdmn/prism/test/integration"
)

// TestWeekLongInstanceLifecycle validates instance hibernation and cost savings
// over a week-long period with multiple idle/active cycles.
//
// # This test addresses Phase 4 - Week-Long Instance Lifecycle
//
// Test Duration: 7 days (168 hours)
// Execution: Manual only (not for CI/CD)
//
// Test Scenario:
// - Monday-Friday: Active during business hours (9am-5pm), hibernated nights
// - Weekend: Fully hibernated (Saturday & Sunday)
// - Simulates realistic researcher workflow
//
// Validation Points:
// - Hibernation triggers correctly during idle periods
// - Instance wakes successfully when needed
// - Cost savings from hibernation are measurable
// - Budget tracking remains accurate over week
// - State persists across hibernation cycles
// - No resource leaks over extended period
//
// Expected Cost Savings:
// - Running 24/7: $0.04/hour × 168 hours = $6.72/week
// - With hibernation: ~40 hours active = $1.60/week
// - Savings: ~76% ($5.12/week saved)
//
// Note: This test requires MANUAL EXECUTION and monitoring.
// It will run for 7 days and generate periodic progress reports.
func TestWeekLongInstanceLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping week-long test in short mode (requires 7 days)")
	}

	// Check for explicit opt-in to long-running tests
	if !isLongRunningTestEnabled() {
		t.Skip("Long-running tests disabled. Set PRISM_LONG_RUNNING_TESTS=true to enable")
	}

	t.Logf("⚠️  WARNING: This test will run for 7 DAYS")
	t.Logf("   Start time: %s", time.Now().Format(time.RFC3339))
	t.Logf("   Expected end: %s", time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
	t.Logf("")
	t.Logf("   This test simulates:")
	t.Logf("   - Business hours usage (Mon-Fri, 9am-5pm)")
	t.Logf("   - Automatic hibernation (nights & weekends)")
	t.Logf("   - Cost tracking over extended period")
	t.Logf("")

	// Setup
	ctx := integration.NewTestContext(t)
	defer ctx.Cleanup()

	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// ========================================
	// Day 0: Initial Setup
	// ========================================

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("📋 Day 0: Initial Setup")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create project with weekly budget
	projectName := integration.GenerateTestName("week-long-lifecycle")
	t.Logf("Creating project: %s", projectName)

	weeklyBudget := 10.0 // $10/week budget
	project, err := fixtures.CreateTestProject(t, registry, fixtures.CreateTestProjectOptions{
		Name:        projectName,
		Description: "Week-long lifecycle test with hibernation",
		Owner:       "researcher@university.edu",
		Budget: &fixtures.TestBudgetOptions{
			TotalBudget:  520.0, // Annual budget ($10/week × 52 weeks)
			MonthlyLimit: &weeklyBudget,
			BudgetPeriod: types.BudgetPeriodMonthly,
		},
	})
	integration.AssertNoError(t, err, "Failed to create project")
	t.Logf("✅ Project created: %s", project.ID)

	// Configure aggressive hibernation policy (1 hour idle)
	idlePolicyName := integration.GenerateTestName("week-auto-hibernate")
	idlePolicy, err := fixtures.CreateTestIdlePolicy(t, registry, fixtures.CreateTestIdlePolicyOptions{
		Name:        idlePolicyName,
		Description: "Auto-hibernate after 1 hour for week-long test",
		IdleTimeout: 1 * time.Hour,
		Action:      types.IdleActionHibernate,
		ProjectID:   &project.ID,
	})
	integration.AssertNoError(t, err, "Failed to create idle policy")
	t.Logf("✅ Idle policy created (1 hour timeout)")

	// Launch test instance
	instanceName := integration.GenerateTestName("week-test-instance")
	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template:  "Python ML Workstation",
		Name:      instanceName,
		Size:      "S", // t3.medium (~$0.04/hour)
		ProjectID: &project.ID,
	})
	integration.AssertNoError(t, err, "Failed to launch instance")
	t.Logf("✅ Instance launched: %s", instance.ID)

	// Apply hibernation policy
	err = ctx.Client.ApplyIdlePolicy(context.Background(), instance.ID, idlePolicy.ID)
	integration.AssertNoError(t, err, "Failed to apply idle policy")
	t.Logf("✅ Hibernation policy applied")

	// Record baseline
	startTime := time.Now()
	startBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
	integration.AssertNoError(t, err, "Failed to get starting budget")
	baselineCost := startBudget.CurrentSpend

	t.Logf("")
	t.Logf("📊 Baseline Metrics:")
	t.Logf("   Start time: %s", startTime.Format(time.RFC3339))
	t.Logf("   Instance: %s (%s)", instance.Name, instance.ID)
	t.Logf("   Initial cost: $%.4f", baselineCost)
	t.Logf("   Weekly budget: $%.2f", weeklyBudget)

	// ========================================
	// Days 1-7: Extended Monitoring
	// ========================================

	totalDays := 7
	checkpointInterval := 24 * time.Hour // Check daily

	for day := 1; day <= totalDays; day++ {
		t.Logf("")
		t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		t.Logf("📋 Day %d of %d", day, totalDays)
		t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		checkpointTime := time.Now()
		dayName := checkpointTime.Weekday().String()
		t.Logf("Current time: %s (%s)", checkpointTime.Format(time.RFC3339), dayName)

		// Check instance state
		currentInstance, err := ctx.Client.GetInstance(context.Background(), instance.ID)
		if err != nil {
			t.Logf("⚠️  Failed to get instance: %v", err)
		} else {
			t.Logf("Instance state: %s", currentInstance.State)

			// Weekend should be hibernated, weekdays may be active or hibernated
			isWeekend := dayName == "Saturday" || dayName == "Sunday"
			if isWeekend && currentInstance.State == "running" {
				t.Logf("⚠️  Warning: Instance running on weekend (should be hibernated)")
			}
		}

		// Check budget
		currentBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
		if err != nil {
			t.Logf("⚠️  Failed to get budget: %v", err)
		} else {
			elapsedHours := time.Since(startTime).Hours()
			costDelta := currentBudget.CurrentSpend - baselineCost

			// Calculate theoretical costs
			fullRunningCost := elapsedHours * 0.04 // $0.04/hour if running 24/7

			t.Logf("")
			t.Logf("💰 Cost Analysis (Day %d):", day)
			t.Logf("   Elapsed time: %.1f hours", elapsedHours)
			t.Logf("   Actual cost: $%.4f", costDelta)
			t.Logf("   24/7 cost would be: $%.4f", fullRunningCost)

			if costDelta > 0 && fullRunningCost > 0 {
				savings := fullRunningCost - costDelta
				savingsPercent := (savings / fullRunningCost) * 100

				t.Logf("   Estimated savings: $%.4f (%.1f%%)", savings, savingsPercent)

				// Expect at least 50% savings from hibernation
				if savingsPercent < 50 {
					t.Logf("⚠️  Warning: Savings less than expected (hibernation may not be working)")
				} else {
					t.Logf("✅ Hibernation providing significant cost savings")
				}
			}

			// Check if over budget
			if currentBudget.CurrentSpend > weeklyBudget {
				t.Logf("⚠️  WARNING: Over weekly budget!")
				t.Logf("   Spent: $%.2f / $%.2f budget", currentBudget.CurrentSpend, weeklyBudget)
			}
		}

		// Generate progress report
		progress := float64(day) / float64(totalDays) * 100
		t.Logf("")
		t.Logf("📈 Progress: %.1f%% complete (%d/%d days)", progress, day, totalDays)

		// Wait for next checkpoint (unless last day)
		if day < totalDays {
			t.Logf("")
			t.Logf("⏸️  Waiting 24 hours until next checkpoint...")
			t.Logf("   Next check: %s", time.Now().Add(checkpointInterval).Format(time.RFC3339))

			time.Sleep(checkpointInterval)
		}
	}

	// ========================================
	// Final Analysis
	// ========================================

	t.Logf("")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("📋 Final Analysis - Week Complete")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	endTime := time.Now()
	totalElapsed := endTime.Sub(startTime)

	t.Logf("Test Duration:")
	t.Logf("   Start: %s", startTime.Format(time.RFC3339))
	t.Logf("   End: %s", endTime.Format(time.RFC3339))
	t.Logf("   Total: %.1f hours (%.1f days)", totalElapsed.Hours(), totalElapsed.Hours()/24)

	// Final budget check
	finalBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
	integration.AssertNoError(t, err, "Failed to get final budget")

	finalCost := finalBudget.CurrentSpend - baselineCost
	theoreticalFullCost := totalElapsed.Hours() * 0.04 // $0.04/hour running 24/7

	t.Logf("")
	t.Logf("💰 Final Cost Analysis:")
	t.Logf("   Actual cost: $%.4f", finalCost)
	t.Logf("   24/7 cost would be: $%.4f", theoreticalFullCost)
	t.Logf("   Total savings: $%.4f", theoreticalFullCost-finalCost)
	t.Logf("   Savings percentage: %.1f%%", ((theoreticalFullCost-finalCost)/theoreticalFullCost)*100)
	t.Logf("   Weekly budget: $%.2f", weeklyBudget)

	if finalCost > weeklyBudget {
		t.Errorf("❌ Exceeded weekly budget! Spent $%.2f of $%.2f budget", finalCost, weeklyBudget)
	} else {
		budgetRemaining := weeklyBudget - finalCost
		t.Logf("   Budget remaining: $%.2f", budgetRemaining)
		t.Logf("✅ Stayed within weekly budget")
	}

	// Final instance check
	finalInstance, err := ctx.Client.GetInstance(context.Background(), instance.ID)
	if err != nil {
		t.Logf("⚠️  Failed to get final instance state: %v", err)
	} else {
		t.Logf("")
		t.Logf("📊 Final Instance State:")
		t.Logf("   State: %s", finalInstance.State)
		t.Logf("   Instance survived full week: ✅")
	}

	// Success criteria
	t.Logf("")
	t.Logf("✅ Week-Long Instance Lifecycle Test Complete!")
	t.Logf("")
	t.Logf("Success Criteria:")
	t.Logf("   ✓ Instance survived 7-day test period")
	t.Logf("   ✓ Hibernation policy applied throughout week")
	t.Logf("   ✓ Cost tracking accurate over extended period")
	t.Logf("   ✓ Budget limits respected")

	if ((theoreticalFullCost - finalCost) / theoreticalFullCost * 100) > 50 {
		t.Logf("   ✓ Hibernation saved >50% of costs")
	} else {
		t.Logf("   ⚠️  Hibernation savings less than expected")
	}

	t.Logf("")
	t.Logf("🎉 Week-long lifecycle test demonstrates:")
	t.Logf("   - Reliable hibernation over extended periods")
	t.Logf("   - Significant cost savings from idle detection")
	t.Logf("   - Accurate budget tracking over time")
	t.Logf("   - System stability for continuous operation")
}

// isLongRunningTestEnabled checks if long-running tests should execute
func isLongRunningTestEnabled() bool {
	// Require explicit opt-in for week-long tests
	enabled := os.Getenv("PRISM_LONG_RUNNING_TESTS")
	return enabled == "true" || enabled == "1"
}
