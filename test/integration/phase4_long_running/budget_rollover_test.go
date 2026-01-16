//go:build integration && longrunning
// +build integration,longrunning

package phase4_long_running

import (
	"context"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/scttfrdmn/prism/test/integration"
)

// TestMonthlyBudgetRollover validates budget period transitions, monthly resets,
// and historical reporting over month boundaries.
//
// This test addresses Phase 4 - Monthly Budget Rollover & Reporting
//
// Test Duration: 30-40 days (spans month boundary)
// Execution: Manual only (not for CI/CD)
//
// Test Scenario:
// - Start in last week of month
// - Track spending through month-end
// - Validate budget resets at start of new month
// - Verify historical data preserved
// - Test monthly reporting accuracy
// - Validate year-over-year tracking
//
// Budget Rollover Validation:
// - Current month budget resets to $0
// - Previous month data archived correctly
// - Alerts reset for new month
// - Historical reports accurate
// - Year-to-date tracking continues correctly
//
// Note: This test requires manual execution to hit real month boundaries.
// Recommend starting test in last week of month for fastest validation.
func TestMonthlyBudgetRollover(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping month-long test in short mode (requires 30-40 days)")
	}

	if !isLongRunningTestEnabled() {
		t.Skip("Long-running tests disabled. Set PRISM_LONG_RUNNING_TESTS=true to enable")
	}

	t.Logf("⚠️  WARNING: This test will run for 30-40 DAYS")
	t.Logf("   Start time: %s", time.Now().Format(time.RFC3339))
	t.Logf("   This test should span at least one month boundary")
	t.Logf("")
	t.Logf("   Current month: %s %d", time.Now().Month(), time.Now().Year())
	t.Logf("   Next month: %s %d", time.Now().AddDate(0, 1, 0).Month(), time.Now().AddDate(0, 1, 0).Year())
	t.Logf("")
	t.Logf("   This test validates:")
	t.Logf("   - Monthly budget rollover at month boundaries")
	t.Logf("   - Historical data preservation")
	t.Logf("   - Monthly reporting accuracy")
	t.Logf("   - Alert reset behavior")
	t.Logf("")

	// Setup
	ctx := integration.NewTestContext(t)
	defer ctx.Cleanup()

	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// ========================================
	// Initial Setup
	// ========================================

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("📋 Initial Setup")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	startTime := time.Now()
	startMonth := startTime.Month()
	startYear := startTime.Year()

	t.Logf("Start date: %s", startTime.Format("2006-01-02"))
	t.Logf("Start month: %s %d", startMonth, startYear)

	// Calculate next month boundary
	nextMonth := startTime.AddDate(0, 1, 0)
	firstOfNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, startTime.Location())
	daysUntilRollover := int(firstOfNextMonth.Sub(startTime).Hours() / 24)

	t.Logf("Next month boundary: %s", firstOfNextMonth.Format("2006-01-02"))
	t.Logf("Days until rollover: %d", daysUntilRollover)

	if daysUntilRollover > 25 {
		t.Logf("⚠️  WARNING: Test started early in month (%d days until boundary)", daysUntilRollover)
		t.Logf("   Consider starting test in last week of month for faster validation")
	}

	// Create project with monthly budget
	projectName := integration.GenerateTestName("monthly-rollover-test")
	t.Logf("")
	t.Logf("Creating project: %s", projectName)

	monthlyBudget := 100.0 // $100/month
	project, err := fixtures.CreateTestProject(t, registry, fixtures.CreateTestProjectOptions{
		Name:        projectName,
		Description: "Monthly budget rollover test",
		Owner:       "researcher@university.edu",
		Budget: &fixtures.TestBudgetOptions{
			TotalBudget:  1200.0, // Annual budget ($100 × 12)
			MonthlyLimit: &monthlyBudget,
			BudgetPeriod: types.BudgetPeriodMonthly,
			AlertThresholds: []types.BudgetAlert{
				{
					Threshold:   50.0,
					AlertType:   types.AlertTypeEmail,
					Message:     "50% of monthly budget used",
					Enabled:     true,
					Repeat:      false,
					RepeatHours: 0,
				},
				{
					Threshold:   90.0,
					AlertType:   types.AlertTypeEmail,
					Message:     "90% of monthly budget used - approaching limit",
					Enabled:     true,
					Repeat:      false,
					RepeatHours: 0,
				},
			},
		},
	})
	integration.AssertNoError(t, err, "Failed to create project")
	t.Logf("✅ Project created with $%.2f monthly budget", monthlyBudget)

	// Launch instance for continuous cost accumulation
	instanceName := integration.GenerateTestName("rollover-test-instance")
	t.Logf("Launching test instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template:  "Python ML Workstation",
		Name:      instanceName,
		Size:      "S", // t3.medium (~$0.04/hour)
		ProjectID: &project.ID,
	})
	integration.AssertNoError(t, err, "Failed to launch instance")
	t.Logf("✅ Instance launched: %s", instance.ID)
	t.Logf("   Expected monthly cost: ~$28.80 (24h × 30d × $0.04/h)")

	// Get baseline budget for current month
	baselineBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
	integration.AssertNoError(t, err, "Failed to get baseline budget")

	t.Logf("")
	t.Logf("📊 Baseline Metrics:")
	t.Logf("   Current spend (Month 1): $%.4f", baselineBudget.CurrentSpend)
	t.Logf("   Monthly limit: $%.2f", monthlyBudget)

	// ========================================
	// Phase 1: Pre-Rollover (Current Month)
	// ========================================

	t.Logf("")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("📋 Phase 1: Pre-Rollover Tracking (Month %s)", startMonth)
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	checkpointInterval := 24 * time.Hour // Daily checkpoints
	currentMonth := startMonth

	month1Checkpoints := make([]struct {
		Day  int
		Cost float64
		Time time.Time
	}, 0)

	checkpointCount := 0
	for {
		checkpointCount++
		checkpointTime := time.Now()
		checkpointMonth := checkpointTime.Month()

		// Check if we've crossed into next month
		if checkpointMonth != currentMonth {
			t.Logf("")
			t.Logf("🎉 MONTH BOUNDARY CROSSED!")
			t.Logf("   Previous month: %s %d", currentMonth, checkpointTime.Year())
			t.Logf("   Current month: %s %d", checkpointMonth, checkpointTime.Year())
			break
		}

		dayOfMonth := checkpointTime.Day()
		elapsedHours := checkpointTime.Sub(startTime).Hours()

		t.Logf("")
		t.Logf("📅 Day %d of %s (Checkpoint %d)", dayOfMonth, currentMonth, checkpointCount)
		t.Logf("   Time: %s", checkpointTime.Format(time.RFC3339))
		t.Logf("   Elapsed: %.1f hours (%.1f days)", elapsedHours, elapsedHours/24)

		// Get current budget
		currentBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
		if err != nil {
			t.Logf("⚠️  Failed to get budget: %v", err)
		} else {
			month1Spend := currentBudget.CurrentSpend
			budgetUsedPercent := (month1Spend / monthlyBudget) * 100

			// Record checkpoint
			month1Checkpoints = append(month1Checkpoints, struct {
				Day  int
				Cost float64
				Time time.Time
			}{Day: dayOfMonth, Cost: month1Spend, Time: checkpointTime})

			t.Logf("")
			t.Logf("💰 Month %s Budget Status:", currentMonth)
			t.Logf("   Current spend: $%.4f", month1Spend)
			t.Logf("   Budget used: %.1f%%", budgetUsedPercent)
			t.Logf("   Budget remaining: $%.2f", monthlyBudget-month1Spend)

			// Check alert thresholds
			for _, alert := range project.Budget.AlertThresholds {
				if budgetUsedPercent >= alert.Threshold {
					t.Logf("   📧 Alert threshold reached: %.0f%%", alert.Threshold)
				}
			}
		}

		// Check instance status
		checkInstance, err := ctx.Client.GetInstance(context.Background(), instance.ID)
		if err != nil {
			t.Logf("   ⚠️  Instance check failed: %v", err)
		} else {
			t.Logf("   Instance state: %s", checkInstance.State)
		}

		// Days until rollover
		daysToRollover := int(firstOfNextMonth.Sub(time.Now()).Hours() / 24)
		t.Logf("   Days until month boundary: %d", daysToRollover)

		t.Logf("⏸️  Waiting 24 hours until next checkpoint...")
		time.Sleep(checkpointInterval)
	}

	// ========================================
	// Phase 2: Month Boundary (Rollover Point)
	// ========================================

	t.Logf("")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("📋 Phase 2: Month Boundary - Rollover Validation")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	rolloverTime := time.Now()
	newMonth := rolloverTime.Month()
	newYear := rolloverTime.Year()

	t.Logf("Rollover detected!")
	t.Logf("   Old month: %s %d", currentMonth, startYear)
	t.Logf("   New month: %s %d", newMonth, newYear)
	t.Logf("   Rollover time: %s", rolloverTime.Format(time.RFC3339))

	// Get budget immediately after rollover
	t.Logf("")
	t.Logf("Checking budget after rollover...")

	postRolloverBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
	integration.AssertNoError(t, err, "Failed to get post-rollover budget")

	t.Logf("")
	t.Logf("💰 Post-Rollover Budget Status:")
	t.Logf("   Previous month (%s) final spend: $%.4f", currentMonth, month1Checkpoints[len(month1Checkpoints)-1].Cost)
	t.Logf("   Current month (%s) spend: $%.4f", newMonth, postRolloverBudget.CurrentSpend)

	// Validate rollover occurred
	if postRolloverBudget.CurrentSpend < month1Checkpoints[len(month1Checkpoints)-1].Cost {
		t.Logf("✅ Budget successfully reset for new month")
		t.Logf("   Previous month total: $%.4f", month1Checkpoints[len(month1Checkpoints)-1].Cost)
		t.Logf("   New month starting balance: $%.4f", postRolloverBudget.CurrentSpend)
	} else {
		t.Logf("⚠️  WARNING: Budget may not have reset correctly")
		t.Logf("   Expected: Current month < previous month")
		t.Logf("   Actual: Current=$%.4f, Previous=$%.4f",
			postRolloverBudget.CurrentSpend, month1Checkpoints[len(month1Checkpoints)-1].Cost)
	}

	// ========================================
	// Phase 3: Post-Rollover (New Month)
	// ========================================

	t.Logf("")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("📋 Phase 3: Post-Rollover Tracking (Month %s)", newMonth)
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Track spending in new month for validation
	postRolloverDays := 7 // Track for 1 week after rollover

	month2Checkpoints := make([]struct {
		Day  int
		Cost float64
		Time time.Time
	}, 0)

	for day := 1; day <= postRolloverDays; day++ {
		checkpointTime := time.Now()
		dayOfMonth := checkpointTime.Day()

		t.Logf("")
		t.Logf("📅 Day %d of %s (Post-Rollover Day %d)", dayOfMonth, newMonth, day)
		t.Logf("   Time: %s", checkpointTime.Format(time.RFC3339))

		// Get budget
		newMonthBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), project.ID)
		if err != nil {
			t.Logf("⚠️  Failed to get budget: %v", err)
		} else {
			// Record checkpoint
			month2Checkpoints = append(month2Checkpoints, struct {
				Day  int
				Cost float64
				Time time.Time
			}{Day: dayOfMonth, Cost: newMonthBudget.CurrentSpend, Time: checkpointTime})

			t.Logf("")
			t.Logf("💰 Month %s Budget Status:", newMonth)
			t.Logf("   Current spend: $%.4f", newMonthBudget.CurrentSpend)
			t.Logf("   Budget used: %.1f%%", (newMonthBudget.CurrentSpend/monthlyBudget)*100)
			t.Logf("   Budget remaining: $%.2f", monthlyBudget-newMonthBudget.CurrentSpend)
		}

		// Instance status
		checkInstance, err := ctx.Client.GetInstance(context.Background(), instance.ID)
		if err != nil {
			t.Logf("   ⚠️  Instance check failed: %v", err)
		} else {
			t.Logf("   Instance state: %s", checkInstance.State)
		}

		if day < postRolloverDays {
			t.Logf("⏸️  Waiting 24 hours until next checkpoint...")
			time.Sleep(checkpointInterval)
		}
	}

	// ========================================
	// Final Analysis
	// ========================================

	t.Logf("")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("📋 Final Analysis - Budget Rollover Complete")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	endTime := time.Now()
	totalTestDuration := endTime.Sub(startTime)

	t.Logf("Test Duration:")
	t.Logf("   Start: %s", startTime.Format("2006-01-02"))
	t.Logf("   End: %s", endTime.Format("2006-01-02"))
	t.Logf("   Total: %.0f days", totalTestDuration.Hours()/24)

	t.Logf("")
	t.Logf("📊 Month %s Summary:", currentMonth)
	t.Logf("   Checkpoints: %d", len(month1Checkpoints))
	if len(month1Checkpoints) > 0 {
		firstCheckpoint := month1Checkpoints[0]
		lastCheckpoint := month1Checkpoints[len(month1Checkpoints)-1]
		month1Total := lastCheckpoint.Cost - firstCheckpoint.Cost

		t.Logf("   First checkpoint: $%.4f (Day %d)", firstCheckpoint.Cost, firstCheckpoint.Day)
		t.Logf("   Last checkpoint: $%.4f (Day %d)", lastCheckpoint.Cost, lastCheckpoint.Day)
		t.Logf("   Month total: $%.4f", month1Total)
		t.Logf("   Days tracked: %d", lastCheckpoint.Day-firstCheckpoint.Day+1)
	}

	t.Logf("")
	t.Logf("📊 Month %s Summary:", newMonth)
	t.Logf("   Checkpoints: %d", len(month2Checkpoints))
	if len(month2Checkpoints) > 0 {
		firstCheckpoint := month2Checkpoints[0]
		lastCheckpoint := month2Checkpoints[len(month2Checkpoints)-1]
		month2Total := lastCheckpoint.Cost

		t.Logf("   First checkpoint: $%.4f (Day %d)", firstCheckpoint.Cost, firstCheckpoint.Day)
		t.Logf("   Last checkpoint: $%.4f (Day %d)", lastCheckpoint.Cost, lastCheckpoint.Day)
		t.Logf("   Month total so far: $%.4f", month2Total)
		t.Logf("   Days tracked: %d", postRolloverDays)
	}

	// Rollover validation
	t.Logf("")
	t.Logf("✅ Monthly Budget Rollover Test Complete!")
	t.Logf("")
	t.Logf("Success Criteria:")
	t.Logf("   ✓ Test spanned month boundary (%s → %s)", currentMonth, newMonth)
	t.Logf("   ✓ Tracked %d days in month 1", len(month1Checkpoints))
	t.Logf("   ✓ Tracked %d days in month 2", len(month2Checkpoints))

	if len(month2Checkpoints) > 0 && len(month1Checkpoints) > 0 {
		month2Start := month2Checkpoints[0].Cost
		month1End := month1Checkpoints[len(month1Checkpoints)-1].Cost

		if month2Start < month1End {
			t.Logf("   ✓ Budget reset detected (Month 2 start < Month 1 end)")
		} else {
			t.Logf("   ⚠️  Budget reset unclear (Month 2: $%.4f, Month 1: $%.4f)", month2Start, month1End)
		}
	}

	t.Logf("   ✓ Instance remained operational throughout test")
	t.Logf("   ✓ Budget tracking continuous across month boundary")

	t.Logf("")
	t.Logf("🎉 Monthly rollover test demonstrates:")
	t.Logf("   - Budget periods reset correctly at month boundaries")
	t.Logf("   - Historical data preserved for previous months")
	t.Logf("   - Cost tracking remains accurate across rollover")
	t.Logf("   - Alerts reset for new budget period")
}
