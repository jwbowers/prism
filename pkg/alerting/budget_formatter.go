package alerting

import (
	"fmt"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// FormatBudgetThresholdAlert creates an Alert when a budget threshold is crossed.
func FormatBudgetThresholdAlert(projectID, projectName string, threshold, spentAmount, totalBudget float64) Alert {
	pct := 0.0
	if totalBudget > 0 {
		pct = spentAmount / totalBudget * 100
	}
	severity := AlertSeverityWarning
	if pct >= 90 {
		severity = AlertSeverityCritical
	}
	return Alert{
		ID:          fmt.Sprintf("budget-threshold-%s-%d", projectID, time.Now().Unix()),
		Severity:    severity,
		Title:       fmt.Sprintf("Budget Alert: %s reached %.0f%%", projectName, pct),
		Body:        fmt.Sprintf("Project %q has spent $%.2f of $%.2f (%.1f%%). Threshold: %.0f%%.", projectName, spentAmount, totalBudget, pct, threshold*100),
		ProjectID:   projectID,
		ProjectName: projectName,
		Tags:        map[string]string{"type": "threshold", "threshold": fmt.Sprintf("%.0f%%", threshold*100)},
		CreatedAt:   time.Now(),
	}
}

// FormatShortfallAlert creates an Alert from a predictive shortfall warning.
func FormatShortfallAlert(projectID, projectName string, dailyRate float64, daysUntil int, exhaustionDate time.Time) Alert {
	return Alert{
		ID:          fmt.Sprintf("budget-shortfall-%s-%d", projectID, time.Now().Unix()),
		Severity:    AlertSeverityWarning,
		Title:       fmt.Sprintf("Budget Forecast: %s may exhaust in %d days", projectName, daysUntil),
		Body:        fmt.Sprintf("At $%.2f/day, project %q budget will be exhausted around %s.", dailyRate, projectName, exhaustionDate.Format("2006-01-02")),
		ProjectID:   projectID,
		ProjectName: projectName,
		Tags:        map[string]string{"type": "shortfall", "days_until": fmt.Sprintf("%d", daysUntil)},
		CreatedAt:   time.Now(),
	}
}

// FormatCushionAlert creates an Alert when the budget cushion threshold is crossed.
func FormatCushionAlert(projectID, projectName string, headroomPct, remaining, total float64, mode string) Alert {
	return Alert{
		ID:       fmt.Sprintf("budget-cushion-%s-%d", projectID, time.Now().Unix()),
		Severity: AlertSeverityCritical,
		Title:    fmt.Sprintf("Budget Cushion Triggered: %s", projectName),
		Body: fmt.Sprintf(
			"Project %q has $%.2f remaining (%.1f%% of $%.2f). Cushion action: %s.",
			projectName, remaining, remaining/total*100, total, mode,
		),
		ProjectID:   projectID,
		ProjectName: projectName,
		Tags:        map[string]string{"type": "cushion", "mode": mode},
		CreatedAt:   time.Now(),
	}
}

// FormatAutoActionAlert creates an Alert when an automatic budget action executes.
func FormatAutoActionAlert(projectID, projectName string, action types.BudgetActionType, spentAmount, totalBudget float64) Alert {
	return Alert{
		ID:       fmt.Sprintf("budget-action-%s-%d", projectID, time.Now().Unix()),
		Severity: AlertSeverityCritical,
		Title:    fmt.Sprintf("Budget Auto-Action: %s — %s", projectName, action),
		Body: fmt.Sprintf(
			"Automatic action %q triggered for project %q. Spent: $%.2f / $%.2f.",
			action, projectName, spentAmount, totalBudget,
		),
		ProjectID:   projectID,
		ProjectName: projectName,
		Tags:        map[string]string{"type": "auto_action", "action": string(action)},
		CreatedAt:   time.Now(),
	}
}
