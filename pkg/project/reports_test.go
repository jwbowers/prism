package project

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers for building test cost data points

func makeDataPoint(ts time.Time, dailyCost float64, instances []types.InstanceCost, storages []types.StorageCost) CostDataPoint {
	return CostDataPoint{
		Timestamp:     ts,
		TotalCost:     dailyCost,
		DailyCost:     dailyCost,
		InstanceCosts: instances,
		StorageCosts:  storages,
	}
}

func inMonth(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
}

// TestGenerateMonthlyReport_Empty verifies that a report with no history has zero spend.
func TestGenerateMonthlyReport_Empty(t *testing.T) {
	budget := &types.ProjectBudget{TotalBudget: 1000}
	report, err := GenerateMonthlyReport("proj-1", "2026-02", nil, budget)
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Equal(t, "proj-1", report.ProjectID)
	assert.Equal(t, "2026-02", report.Month)
	assert.Equal(t, 0.0, report.TotalSpend)
	assert.Empty(t, report.ByInstance)
	assert.Empty(t, report.ByStorage)
}

// TestGenerateMonthlyReport_WrongMonth ensures data points outside the target month are excluded.
func TestGenerateMonthlyReport_WrongMonth(t *testing.T) {
	history := []CostDataPoint{
		// January — before the target month
		makeDataPoint(inMonth(2026, time.January, 15), 50.0, nil, nil),
		// March — after the target month
		makeDataPoint(inMonth(2026, time.March, 1), 75.0, nil, nil),
	}

	budget := &types.ProjectBudget{TotalBudget: 1000}
	report, err := GenerateMonthlyReport("proj-1", "2026-02", history, budget)
	require.NoError(t, err)

	assert.Equal(t, 0.0, report.TotalSpend)
}

// TestGenerateMonthlyReport_WithData verifies that DailyCost values within the month are summed.
func TestGenerateMonthlyReport_WithData(t *testing.T) {
	ic := types.InstanceCost{
		InstanceName: "my-instance",
		InstanceType: "m5.large",
		RunningHours: 8.0,
		ComputeCost:  1.50,
		StorageCost:  0.10,
		TotalCost:    1.60,
	}
	sc := types.StorageCost{
		VolumeName: "vol-data",
		VolumeType: "EBS",
		SizeGB:     100,
		Cost:       0.25,
	}

	history := []CostDataPoint{
		makeDataPoint(inMonth(2026, time.February, 5), 10.0, []types.InstanceCost{ic}, []types.StorageCost{sc}),
		makeDataPoint(inMonth(2026, time.February, 10), 20.0, []types.InstanceCost{ic}, []types.StorageCost{sc}),
		// Point outside target month — must be excluded
		makeDataPoint(inMonth(2026, time.March, 1), 999.0, nil, nil),
	}

	budget := &types.ProjectBudget{TotalBudget: 500}
	report, err := GenerateMonthlyReport("proj-1", "2026-02", history, budget)
	require.NoError(t, err)

	// TotalSpend should be sum of DailyCost for February points only
	assert.InDelta(t, 30.0, report.TotalSpend, 0.001)

	// Instance and storage aggregation
	require.Len(t, report.ByInstance, 1)
	assert.Equal(t, "my-instance", report.ByInstance[0].InstanceName)
	assert.InDelta(t, 16.0, report.ByInstance[0].RunningHours, 0.001) // 8 * 2 points
	assert.InDelta(t, 3.0, report.ByInstance[0].ComputeCost, 0.001)   // 1.50 * 2
	assert.InDelta(t, 3.20, report.ByInstance[0].TotalCost, 0.001)    // 1.60 * 2

	require.Len(t, report.ByStorage, 1)
	assert.Equal(t, "vol-data", report.ByStorage[0].VolumeName)
	assert.InDelta(t, 0.50, report.ByStorage[0].Cost, 0.001) // 0.25 * 2
}

// TestGenerateMonthlyReport_Utilization checks that Utilization = TotalSpend / BudgetLimit.
func TestGenerateMonthlyReport_Utilization(t *testing.T) {
	history := []CostDataPoint{
		makeDataPoint(inMonth(2026, time.February, 14), 250.0, nil, nil),
	}

	// MonthlyAmount > 0 takes precedence over TotalBudget
	budget := &types.ProjectBudget{
		TotalBudget:   1000,
		MonthlyAmount: 500,
	}
	report, err := GenerateMonthlyReport("proj-1", "2026-02", history, budget)
	require.NoError(t, err)

	assert.InDelta(t, 500.0, report.BudgetLimit, 0.001)
	assert.InDelta(t, 0.5, report.Utilization, 0.001) // 250 / 500
}

// TestGenerateMonthlyReport_NoBudget verifies that nil budget results in zero BudgetLimit and no panic.
func TestGenerateMonthlyReport_NoBudget(t *testing.T) {
	history := []CostDataPoint{
		makeDataPoint(inMonth(2026, time.February, 1), 100.0, nil, nil),
	}

	report, err := GenerateMonthlyReport("proj-1", "2026-02", history, nil)
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Equal(t, 0.0, report.BudgetLimit)
	assert.Equal(t, 0.0, report.Utilization)
	assert.InDelta(t, 100.0, report.TotalSpend, 0.001)
}

// TestGenerateMonthlyReport_InvalidMonth verifies that a malformed month string returns an error.
func TestGenerateMonthlyReport_InvalidMonth(t *testing.T) {
	report, err := GenerateMonthlyReport("proj-1", "2026-13", nil, nil)
	assert.Error(t, err)
	assert.Nil(t, report)
}

// TestMonthlyReport_RenderText verifies that the text output contains key identifying fields.
func TestMonthlyReport_RenderText(t *testing.T) {
	history := []CostDataPoint{
		makeDataPoint(inMonth(2026, time.February, 20), 42.50, nil, nil),
	}
	budget := &types.ProjectBudget{TotalBudget: 200}
	report, err := GenerateMonthlyReport("my-project", "2026-02", history, budget)
	require.NoError(t, err)

	text := report.RenderText()

	assert.Contains(t, text, "my-project")
	assert.Contains(t, text, "2026-02")
	assert.Contains(t, text, "42.50")
}

// TestMonthlyReport_RenderCSV checks that CSV output has a header row and is parseable.
func TestMonthlyReport_RenderCSV(t *testing.T) {
	ic := types.InstanceCost{
		InstanceName: "csv-instance",
		InstanceType: "t3.medium",
		RunningHours: 24.0,
		ComputeCost:  1.0,
		StorageCost:  0.05,
		TotalCost:    1.05,
	}
	history := []CostDataPoint{
		makeDataPoint(inMonth(2026, time.February, 3), 1.05, []types.InstanceCost{ic}, nil),
	}
	report, err := GenerateMonthlyReport("csv-proj", "2026-02", history, nil)
	require.NoError(t, err)

	csv := report.RenderCSV()

	// Must contain the instance section header
	assert.Contains(t, csv, "instance_name")
	assert.Contains(t, csv, "instance_type")
	assert.Contains(t, csv, "running_hours")

	// Must contain instance data
	assert.Contains(t, csv, "csv-instance")
	assert.Contains(t, csv, "t3.medium")

	// Must be non-empty
	assert.NotEmpty(t, strings.TrimSpace(csv))
}

// TestMonthlyReport_RenderJSON verifies that RenderJSON produces valid JSON with the project_id field.
func TestMonthlyReport_RenderJSON(t *testing.T) {
	history := []CostDataPoint{
		makeDataPoint(inMonth(2026, time.February, 7), 15.0, nil, nil),
	}
	budget := &types.ProjectBudget{TotalBudget: 100}
	report, err := GenerateMonthlyReport("json-project", "2026-02", history, budget)
	require.NoError(t, err)

	jsonBytes, err := report.RenderJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)

	// Must be valid JSON
	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	// Must contain project_id field matching what was passed in
	projectID, ok := decoded["project_id"].(string)
	require.True(t, ok, "project_id field missing or not a string")
	assert.Equal(t, "json-project", projectID)

	// Month field must be present
	month, ok := decoded["month"].(string)
	require.True(t, ok, "month field missing or not a string")
	assert.Equal(t, "2026-02", month)
}

// TestMonthlyReport_RenderText_EmptyReport verifies that RenderText does not panic on zero-value report.
func TestMonthlyReport_RenderText_EmptyReport(t *testing.T) {
	report, err := GenerateMonthlyReport("empty-proj", "2026-02", nil, nil)
	require.NoError(t, err)

	// Must not panic
	text := report.RenderText()
	assert.NotEmpty(t, text)
	assert.Contains(t, text, "empty-proj")
	assert.Contains(t, text, "0.00")
}
