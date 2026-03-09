// Package project provides monthly budget report generation for Prism.
//
// Reports are delivered as text, CSV, or JSON. PDF output can be generated
// externally by piping text output through pandoc/wkhtmltopdf.
//
// Usage:
//
//	prism budget report my-project --month 2026-02
//	prism budget report my-project --month 2026-02 --format csv > report.csv
package project

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// InstanceSpendLine is a single instance row in a monthly report
type InstanceSpendLine struct {
	InstanceName string  `json:"instance_name"`
	InstanceType string  `json:"instance_type"`
	RunningHours float64 `json:"running_hours"`
	ComputeCost  float64 `json:"compute_cost"`
	StorageCost  float64 `json:"storage_cost"`
	TotalCost    float64 `json:"total_cost"`
}

// StorageSpendLine is a single storage row in a monthly report
type StorageSpendLine struct {
	VolumeName string  `json:"volume_name"`
	VolumeType string  `json:"volume_type"`
	SizeGB     float64 `json:"size_gb"`
	Cost       float64 `json:"cost"`
}

// MonthlyReport contains a full monthly budget report for a project
type MonthlyReport struct {
	ProjectID   string              `json:"project_id"`
	Month       string              `json:"month"` // "2026-02"
	TotalSpend  float64             `json:"total_spend"`
	ByInstance  []InstanceSpendLine `json:"by_instance"`
	ByStorage   []StorageSpendLine  `json:"by_storage"`
	BudgetLimit float64             `json:"budget_limit"`
	Utilization float64             `json:"utilization"` // 0.0-1.0
	GeneratedAt time.Time           `json:"generated_at"`
}

// GenerateMonthlyReport builds a MonthlyReport from cost history for the given month.
// month format: "2026-02" (YYYY-MM)
func GenerateMonthlyReport(projectID, month string, history []CostDataPoint, budget *types.ProjectBudget) (*MonthlyReport, error) {
	// Parse the target month
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, fmt.Errorf("invalid month format %q: use YYYY-MM", month)
	}

	periodStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	// Aggregate cost data points that fall within the month
	instanceMap := make(map[string]*InstanceSpendLine)
	storageMap := make(map[string]*StorageSpendLine)
	var totalSpend float64

	for _, pt := range history {
		if pt.Timestamp.Before(periodStart) || !pt.Timestamp.Before(periodEnd) {
			continue
		}

		totalSpend += pt.DailyCost

		for _, ic := range pt.InstanceCosts {
			entry, ok := instanceMap[ic.InstanceName]
			if !ok {
				entry = &InstanceSpendLine{
					InstanceName: ic.InstanceName,
					InstanceType: ic.InstanceType,
				}
				instanceMap[ic.InstanceName] = entry
			}
			entry.RunningHours += ic.RunningHours
			entry.ComputeCost += ic.ComputeCost
			entry.StorageCost += ic.StorageCost
			entry.TotalCost += ic.TotalCost
		}

		for _, sc := range pt.StorageCosts {
			entry, ok := storageMap[sc.VolumeName]
			if !ok {
				entry = &StorageSpendLine{
					VolumeName: sc.VolumeName,
					VolumeType: sc.VolumeType,
					SizeGB:     sc.SizeGB,
				}
				storageMap[sc.VolumeName] = entry
			}
			entry.Cost += sc.Cost
		}
	}

	// Convert maps to slices
	instances := make([]InstanceSpendLine, 0, len(instanceMap))
	for _, v := range instanceMap {
		instances = append(instances, *v)
	}

	storages := make([]StorageSpendLine, 0, len(storageMap))
	for _, v := range storageMap {
		storages = append(storages, *v)
	}

	// Calculate utilization
	var budgetLimit float64
	if budget != nil {
		budgetLimit = budget.TotalBudget
		if budget.MonthlyAmount > 0 {
			budgetLimit = budget.MonthlyAmount
		}
	}

	var utilization float64
	if budgetLimit > 0 {
		utilization = totalSpend / budgetLimit
	}

	return &MonthlyReport{
		ProjectID:   projectID,
		Month:       month,
		TotalSpend:  totalSpend,
		ByInstance:  instances,
		ByStorage:   storages,
		BudgetLimit: budgetLimit,
		Utilization: utilization,
		GeneratedAt: time.Now(),
	}, nil
}

// RenderText returns a human-readable text report
func (r *MonthlyReport) RenderText() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Monthly Budget Report — %s\n", r.Month))
	b.WriteString(strings.Repeat("=", 60) + "\n")
	b.WriteString(fmt.Sprintf("Project:     %s\n", r.ProjectID))
	b.WriteString(fmt.Sprintf("Generated:   %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05 UTC")))
	b.WriteString(fmt.Sprintf("Total Spend: $%.2f\n", r.TotalSpend))
	if r.BudgetLimit > 0 {
		b.WriteString(fmt.Sprintf("Budget:      $%.2f (%.1f%% utilized)\n", r.BudgetLimit, r.Utilization*100))
	}

	if len(r.ByInstance) > 0 {
		b.WriteString("\nInstances:\n")
		b.WriteString(fmt.Sprintf("  %-28s %-14s %8s %10s %10s\n",
			"Name", "Type", "Hours", "Compute", "Total"))
		b.WriteString("  " + strings.Repeat("-", 74) + "\n")
		for _, inst := range r.ByInstance {
			b.WriteString(fmt.Sprintf("  %-28s %-14s %8.1f %10s %10s\n",
				truncate(inst.InstanceName, 28),
				inst.InstanceType,
				inst.RunningHours,
				fmt.Sprintf("$%.2f", inst.ComputeCost),
				fmt.Sprintf("$%.2f", inst.TotalCost),
			))
		}
	}

	if len(r.ByStorage) > 0 {
		b.WriteString("\nStorage:\n")
		b.WriteString(fmt.Sprintf("  %-28s %-10s %8s %10s\n", "Name", "Type", "GB", "Cost"))
		b.WriteString("  " + strings.Repeat("-", 60) + "\n")
		for _, vol := range r.ByStorage {
			b.WriteString(fmt.Sprintf("  %-28s %-10s %8.0f %10s\n",
				truncate(vol.VolumeName, 28),
				vol.VolumeType,
				vol.SizeGB,
				fmt.Sprintf("$%.2f", vol.Cost),
			))
		}
	}

	return b.String()
}

// RenderCSV returns a CSV-formatted report (two sections: instances, storage)
func (r *MonthlyReport) RenderCSV() string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header metadata
	_ = w.Write([]string{"# Monthly Budget Report", r.Month, r.ProjectID})
	_ = w.Write([]string{"# Total Spend", fmt.Sprintf("$%.2f", r.TotalSpend)})
	_ = w.Write([]string{})

	// Instance section
	_ = w.Write([]string{"instance_name", "instance_type", "running_hours", "compute_cost", "storage_cost", "total_cost"})
	for _, inst := range r.ByInstance {
		_ = w.Write([]string{
			inst.InstanceName,
			inst.InstanceType,
			fmt.Sprintf("%.2f", inst.RunningHours),
			fmt.Sprintf("%.4f", inst.ComputeCost),
			fmt.Sprintf("%.4f", inst.StorageCost),
			fmt.Sprintf("%.4f", inst.TotalCost),
		})
	}

	_ = w.Write([]string{})

	// Storage section
	_ = w.Write([]string{"volume_name", "volume_type", "size_gb", "cost"})
	for _, vol := range r.ByStorage {
		_ = w.Write([]string{
			vol.VolumeName,
			vol.VolumeType,
			fmt.Sprintf("%.0f", vol.SizeGB),
			fmt.Sprintf("%.4f", vol.Cost),
		})
	}

	w.Flush()
	return buf.String()
}

// RenderJSON returns the report as indented JSON bytes
func (r *MonthlyReport) RenderJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// truncate shortens a string to maxLen characters, appending "…" if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
