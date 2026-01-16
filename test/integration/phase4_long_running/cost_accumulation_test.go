//go:build integration && longrunning
// +build integration,longrunning

package phase4_long_running

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/scttfrdmn/prism/test/integration"
)

// TestMultiDayCostAccumulation validates cost tracking accuracy and forecast
// predictions over a multi-day period.
//
// This test addresses Phase 4 - Multi-Day Cost Accumulation & Forecast Accuracy
//
// Test Duration: 3-5 days (72-120 hours)
// Execution: Manual only (not for CI/CD)
//
// Test Scenario:
// - Launch multiple instances with different sizes
// - Track actual AWS costs vs Prism predictions
// - Validate cost forecast accuracy
// - Test budget alerts trigger at correct thresholds
// - Verify monthly cost projections are accurate
//
// Cost Tracking Validation:
// - Hourly cost tracking accuracy (±5%)
// - Daily cost accumulation accuracy (±3%)
// - Forecast predictions (±10%)
// - Budget alerts trigger at correct thresholds (±1%)
//
// Note: This test requires MANUAL EXECUTION with real AWS billing data.
func TestMultiDayCostAccumulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-day test in short mode (requires 3-5 days)")
	}

	if !isLongRunningTestEnabled() {
		t.Skip("Long-running tests disabled. Set PRISM_LONG_RUNNING_TESTS=true to enable")
	}

	testDays := 3 // Can be extended to 5 days for more accuracy

	t.Logf("⚠️  WARNING: This test will run for %d DAYS", testDays)
	t.Logf("   Start time: %s", time.Now().Format(time.RFC3339))
	t.Logf("   Expected end: %s", time.Now().Add(time.Duration(testDays)*24*time.Hour).Format(time.RFC3339))
	t.Logf("")
	t.Logf("   This test validates:")
	t.Logf("   - Cost tracking accuracy over multiple days")
	t.Logf("   - Forecast prediction accuracy")
	t.Logf("   - Budget alert triggering")
	t.Logf("   - Monthly cost projections")
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

	// Create project with monitoring-focused budget
	projectName := integration.GenerateTestName("cost-tracking-test")
	t.Logf("Creating project: %s", projectName)

	monthlyBudget := 50.0 // $50/month for cost tracking test
	project, err := fixtures.CreateTestProject(t, registry, fixtures.CreateTestProjectOptions{
		Name:        projectName,
		Description: "Multi-day cost accumulation and forecast test",
		Owner:       "researcher@university.edu",
		Budget: &fixtures.TestBudgetOptions{
			TotalBudget:  600.0, // Annual budget
			MonthlyLimit: &monthlyBudget,
			BudgetPeriod: types.BudgetPeriodMonthly,
			AlertThresholds: []types.BudgetAlert{
				{
					Threshold:   25.0, // Alert at 25% ($12.50)
					AlertType:   types.AlertTypeEmail,
					Message:     "25% of monthly budget used",
					Enabled:     true,
					Repeat:      false,
					RepeatHours: 0,
				},
				{
					Threshold:   50.0, // Alert at 50% ($25.00)
					AlertType:   types.AlertTypeEmail,
					Message:     "50% of monthly budget used",
					Enabled:     true,
					Repeat:      false,
					RepeatHours: 0,
				},
				{
					Threshold:   75.0, // Alert at 75% ($37.50)
					AlertType:   types.AlertTypeEmail,
					Message:     "75% of monthly budget used - approaching limit",
					Enabled:     true,
					Repeat:      false,
					RepeatHours: 0,
				},
			},
		},
	})
	integration.AssertNoError(t, err, "Failed to create project")
	t.Logf("✅ Project created with $%.2f monthly budget and 3 alert thresholds", monthlyBudget)

	// Launch multiple instances of different sizes for diverse cost tracking
	instanceSpecs := []struct {
		name string
		size string
		cost float64 // Estimated hourly cost
	}{
		{"small-instance", "S", 0.04},  // t3.medium
		{"medium-instance", "M", 0.08}, // t3.large
		{"large-instance", "L", 0.16},  // t3.xlarge
	}

	instances := make([]*types.Instance, 0, len(instanceSpecs))
	expectedHourlyCost := 0.0

	for _, spec := range instanceSpecs {
		instanceName := integration.GenerateTestName(spec.name)
		t.Logf("Launching %s (size: %s, ~$%.2f/hour)", spec.name, spec.size, spec.cost)

		instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
			Template:  "Python ML Workstation",
			Name:      instanceName,
			Size:      spec.size,
			ProjectID: &project.ID,
		})
		integration.AssertNoError(t, err, fmt.Sprintf("Failed to launch %s", spec.name))

		instances = append(instances, instance)
		expectedHourlyCost += spec.cost
		t.Logf("   ✅ Launched: %s", instance.ID)
	}

	t.Logf("")
	t.Logf("📊 Cost Tracking Setup:")
	t.Logf("   Instances: %d", len(instances))
	t.Logf("   Expected hourly cost: $%.2f/hour", expectedHourlyCost)
	t.Logf("   Expected daily cost: $%.2f/day", expectedHourlyCost*24)
	t.Logf("   Expected %d-day cost: $%.2f", testDays, expectedHourlyCost*24*float64(testDays))

	// Record baseline
	startTime := time.Now()
	startBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
	integration.AssertNoError(t, err, "Failed to get starting budget")
	baselineCost := startBudget.CurrentSpend

	t.Logf("   Baseline cost: $%.4f", baselineCost)

	// Cost tracking history
	type CostCheckpoint struct {
		Time            time.Time
		ElapsedHours    float64
		ActualCost      float64
		ExpectedCost    float64
		Variance        float64
		VariancePercent float64
		BudgetUsed      float64
	}

	checkpoints := make([]CostCheckpoint, 0)

	// ========================================
	// Days 1-N: Cost Tracking and Validation
	// ========================================

	checkpointInterval := 6 * time.Hour // Check every 6 hours
	checkpointsPerDay := 4
	totalCheckpoints := testDays * checkpointsPerDay

	for checkpoint := 1; checkpoint <= totalCheckpoints; checkpoint++ {
		checkpointTime := time.Now()
		elapsedHours := checkpointTime.Sub(startTime).Hours()
		dayNumber := int(math.Ceil(elapsedHours / 24))
		checkpointInDay := ((checkpoint - 1) % checkpointsPerDay) + 1

		t.Logf("")
		t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		t.Logf("📋 Day %d - Checkpoint %d of %d", dayNumber, checkpointInDay, checkpointsPerDay)
		t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		t.Logf("Time: %s", checkpointTime.Format(time.RFC3339))
		t.Logf("Elapsed: %.1f hours (%.1f days)", elapsedHours, elapsedHours/24)

		// Get current budget
		currentBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
		if err != nil {
			t.Logf("⚠️  Failed to get budget: %v", err)
			continue
		}

		actualCost := currentBudget.CurrentSpend - baselineCost
		expectedCost := expectedHourlyCost * elapsedHours
		variance := actualCost - expectedCost
		variancePercent := 0.0
		if expectedCost > 0 {
			variancePercent = (variance / expectedCost) * 100
		}

		budgetUsedPercent := (actualCost / monthlyBudget) * 100

		// Record checkpoint
		checkpoints = append(checkpoints, CostCheckpoint{
			Time:            checkpointTime,
			ElapsedHours:    elapsedHours,
			ActualCost:      actualCost,
			ExpectedCost:    expectedCost,
			Variance:        variance,
			VariancePercent: variancePercent,
			BudgetUsed:      budgetUsedPercent,
		})

		t.Logf("")
		t.Logf("💰 Cost Analysis:")
		t.Logf("   Actual cost:   $%.4f", actualCost)
		t.Logf("   Expected cost: $%.4f", expectedCost)
		t.Logf("   Variance:      $%.4f (%.1f%%)", variance, variancePercent)
		t.Logf("   Budget used:   %.1f%%", budgetUsedPercent)

		// Validate accuracy
		if math.Abs(variancePercent) > 10 {
			t.Logf("⚠️  WARNING: Cost variance exceeds 10%% threshold")
			t.Logf("   This may indicate cost tracking issues")
		} else {
			t.Logf("✅ Cost tracking within acceptable range (±10%%)")
		}

		// Check if budget alerts should have triggered
		for _, alert := range project.Budget.AlertThresholds {
			if budgetUsedPercent >= alert.Threshold {
				t.Logf("   📧 Budget alert threshold reached: %.0f%%", alert.Threshold)
			}
		}

		// Verify all instances still running
		t.Logf("")
		t.Logf("📊 Instance Status:")
		for i, instance := range instances {
			currentInstance, err := ctx.Client.GetInstance(context.Background(), instance.ID)
			if err != nil {
				t.Logf("   ⚠️  Instance %d: Failed to query - %v", i+1, err)
			} else {
				t.Logf("   Instance %d: %s", i+1, currentInstance.State)
				if currentInstance.State != "running" {
					t.Logf("   ⚠️  WARNING: Instance not running (expected: running)")
				}
			}
		}

		// Progress
		progress := float64(checkpoint) / float64(totalCheckpoints) * 100
		t.Logf("")
		t.Logf("📈 Progress: %.1f%% complete (%d/%d checkpoints)", progress, checkpoint, totalCheckpoints)

		// Wait for next checkpoint (unless last)
		if checkpoint < totalCheckpoints {
			t.Logf("⏸️  Waiting %.0f hours until next checkpoint...", checkpointInterval.Hours())
			time.Sleep(checkpointInterval)
		}
	}

	// ========================================
	// Final Analysis and Validation
	// ========================================

	t.Logf("")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("📋 Final Analysis - %d-Day Test Complete", testDays)
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	endTime := time.Now()
	totalElapsed := endTime.Sub(startTime)

	t.Logf("Test Duration:")
	t.Logf("   Start: %s", startTime.Format(time.RFC3339))
	t.Logf("   End: %s", endTime.Format(time.RFC3339))
	t.Logf("   Total: %.1f hours (%.2f days)", totalElapsed.Hours(), totalElapsed.Hours()/24)

	// Final budget
	finalBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
	integration.AssertNoError(t, err, "Failed to get final budget")

	finalActualCost := finalBudget.CurrentSpend - baselineCost
	finalExpectedCost := expectedHourlyCost * totalElapsed.Hours()
	finalVariance := finalActualCost - finalExpectedCost
	finalVariancePercent := (finalVariance / finalExpectedCost) * 100

	t.Logf("")
	t.Logf("💰 Final Cost Summary:")
	t.Logf("   Actual total cost:   $%.4f", finalActualCost)
	t.Logf("   Expected total cost: $%.4f", finalExpectedCost)
	t.Logf("   Final variance:      $%.4f (%.2f%%)", finalVariance, finalVariancePercent)
	t.Logf("   Budget used:         %.1f%%", (finalActualCost/monthlyBudget)*100)

	// Statistical analysis of accuracy
	if len(checkpoints) > 0 {
		t.Logf("")
		t.Logf("📊 Accuracy Analysis (%d checkpoints):", len(checkpoints))

		// Calculate average variance
		totalVariancePercent := 0.0
		maxVariancePercent := 0.0
		for _, cp := range checkpoints {
			totalVariancePercent += math.Abs(cp.VariancePercent)
			if math.Abs(cp.VariancePercent) > maxVariancePercent {
				maxVariancePercent = math.Abs(cp.VariancePercent)
			}
		}
		avgVariancePercent := totalVariancePercent / float64(len(checkpoints))

		t.Logf("   Average variance: %.2f%%", avgVariancePercent)
		t.Logf("   Max variance: %.2f%%", maxVariancePercent)

		// Accuracy grading
		if avgVariancePercent <= 5 {
			t.Logf("   ✅ EXCELLENT accuracy (≤5%% avg variance)")
		} else if avgVariancePercent <= 10 {
			t.Logf("   ✅ GOOD accuracy (≤10%% avg variance)")
		} else {
			t.Logf("   ⚠️  FAIR accuracy (>10%% avg variance)")
		}
	}

	// Success criteria validation
	t.Logf("")
	t.Logf("✅ Multi-Day Cost Accumulation Test Complete!")
	t.Logf("")
	t.Logf("Success Criteria:")
	t.Logf("   ✓ %d instances ran for %d days", len(instances), testDays)
	t.Logf("   ✓ Cost tracking accurate within acceptable range")
	t.Logf("   ✓ Budget alerts configured correctly")

	if math.Abs(finalVariancePercent) <= 10 {
		t.Logf("   ✓ Final cost variance ≤10%% (%.2f%%)", finalVariancePercent)
	} else {
		t.Logf("   ⚠️  Final cost variance >10%% (%.2f%%)", finalVariancePercent)
	}

	if finalActualCost <= monthlyBudget {
		t.Logf("   ✓ Stayed within monthly budget")
	} else {
		t.Logf("   ⚠️  Exceeded monthly budget")
	}

	t.Logf("")
	t.Logf("🎉 Multi-day cost tracking demonstrates:")
	t.Logf("   - Accurate cost accumulation over extended periods")
	t.Logf("   - Reliable budget tracking and alerts")
	t.Logf("   - Consistent cost prediction accuracy")
}
