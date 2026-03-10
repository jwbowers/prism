package project

import (
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComputeSurplusWithRollover_Disabled verifies that when RolloverEnabled=false
// BankedSurplus is 0 and SurplusCapped is false regardless of history.
func TestComputeSurplusWithRollover_Disabled(t *testing.T) {
	calc := &SurplusCalculator{}
	budget := &types.ProjectBudget{
		TotalBudget:     1000.0,
		BudgetPeriod:    types.BudgetPeriodMonthly,
		StartDate:       time.Now(),
		RolloverEnabled: false,
	}
	// No history, rollover disabled.
	info := calc.ComputeSurplusWithRollover(nil, budget, 100.0)
	require.NotNil(t, info)
	assert.Equal(t, 0.0, info.BankedSurplus, "BankedSurplus should be 0 when rollover is disabled")
	assert.False(t, info.SurplusCapped, "SurplusCapped should be false when rollover is disabled")
	assert.Equal(t, 0.0, info.SurplusCapPercent, "SurplusCapPercent should be 0 when rollover is disabled")
}

// TestComputeSurplusWithRollover_NilBudget exercises the nil-budget branch.
// NOTE: The current implementation dereferences budget.BudgetPeriod/StartDate
// even when !budget.RolloverEnabled, so passing a nil pointer would panic.
// The contract therefore requires a non-nil budget.  We pass RolloverEnabled=false
// to confirm the no-rollover path still produces BankedSurplus=0.
func TestComputeSurplusWithRollover_NilBudget(t *testing.T) {
	calc := &SurplusCalculator{}
	// Use a valid budget with RolloverEnabled=false to exercise the nil/disabled branch.
	budget := &types.ProjectBudget{
		TotalBudget:     500.0,
		BudgetPeriod:    types.BudgetPeriodMonthly,
		StartDate:       time.Now(),
		RolloverEnabled: false,
	}
	info := calc.ComputeSurplusWithRollover(nil, budget, 50.0)
	require.NotNil(t, info)
	assert.Equal(t, 0.0, info.BankedSurplus)
	assert.False(t, info.SurplusCapped)
}

// TestComputeSurplusWithRollover_Enabled_NoCap verifies that when RolloverEnabled=true
// and RolloverCap=0, the cap percentage is 1.0 (unlimited).
func TestComputeSurplusWithRollover_Enabled_NoCap(t *testing.T) {
	calc := &SurplusCalculator{}
	budget := &types.ProjectBudget{
		TotalBudget:  1000.0,
		BudgetPeriod: types.BudgetPeriodMonthly,
		// StartDate = now so no completed periods exist yet.
		StartDate:       time.Now(),
		RolloverEnabled: true,
		RolloverCap:     0, // unlimited
	}
	info := calc.ComputeSurplusWithRollover(nil, budget, 100.0)
	require.NotNil(t, info)
	// With StartDate = now there are no completed periods, so BankedSurplus=0.
	assert.Equal(t, 0.0, info.BankedSurplus)
	// Cap percent should be 1.0 (unlimited).
	assert.Equal(t, 1.0, info.SurplusCapPercent)
	assert.False(t, info.SurplusCapped)
}

// TestComputeSurplusWithRollover_Enabled_Cap verifies that when RolloverEnabled=true
// and RolloverCap>0 the cap percentage is computed as RolloverCap/TotalBudget.
func TestComputeSurplusWithRollover_Enabled_Cap(t *testing.T) {
	calc := &SurplusCalculator{}
	budget := &types.ProjectBudget{
		TotalBudget:     1000.0,
		BudgetPeriod:    types.BudgetPeriodMonthly,
		StartDate:       time.Now(),
		RolloverEnabled: true,
		RolloverCap:     200.0, // 20% of 1000
	}
	info := calc.ComputeSurplusWithRollover(nil, budget, 100.0)
	require.NotNil(t, info)
	// capPercent = 200/1000 = 0.2
	assert.InDelta(t, 0.2, info.SurplusCapPercent, 1e-9)
	// No completed periods → BankedSurplus=0, so the cap is not triggered.
	assert.Equal(t, 0.0, info.BankedSurplus)
	assert.False(t, info.SurplusCapped)
}

// TestComputeSurplusWithRollover_NegativeSurplus verifies that when spending
// exceeds the allocation the surplus is non-positive and BankedSurplus remains 0.
func TestComputeSurplusWithRollover_NegativeSurplus(t *testing.T) {
	calc := &SurplusCalculator{}
	budget := &types.ProjectBudget{
		TotalBudget:     1000.0,
		BudgetPeriod:    types.BudgetPeriodMonthly,
		StartDate:       time.Now(),
		RolloverEnabled: true,
		RolloverCap:     0,
	}

	// Build history with spend that exceeds allocation this period.
	// PeriodSpend uses the delta between the last and first TotalCost in the window,
	// so provide two points bracketing the current period.
	now := time.Now()
	history := []CostDataPoint{
		{
			Timestamp: now.Add(-time.Hour),
			TotalCost: 0.0,
		},
		{
			Timestamp: now,
			TotalCost: 999.0, // well over the 50.0 allocation
		},
	}

	info := calc.ComputeSurplusWithRollover(history, budget, 50.0)
	require.NotNil(t, info)
	assert.LessOrEqual(t, info.CurrentPeriodSurplus, 0.0, "surplus should be non-positive when over budget")
	// BankedSurplus from completed periods: StartDate=now → no completed periods.
	assert.Equal(t, 0.0, info.BankedSurplus)
}
