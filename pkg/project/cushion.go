package project

import (
	"context"
	"fmt"

	"github.com/scttfrdmn/prism/pkg/alerting"
)

// CushionMode defines the action taken when the budget headroom threshold is crossed.
type CushionMode string

const (
	// CushionModeHibernate hibernates all running instances for the project.
	CushionModeHibernate CushionMode = "hibernate"
	// CushionModeStop stops all running instances (no resume).
	CushionModeStop CushionMode = "stop"
	// CushionModePreventLaunch blocks new instance launches without stopping running ones.
	CushionModePreventLaunch CushionMode = "prevent_launch"
	// CushionModeNotifyOnly sends an alert but takes no automated action.
	CushionModeNotifyOnly CushionMode = "notify_only"
)

// CushionConfig defines the automatic safety headroom for a project budget.
type CushionConfig struct {
	// Enabled activates the cushion system.
	Enabled bool `json:"enabled"`
	// HeadroomPercent is the fractional budget headroom (e.g. 0.10 = 10%).
	// Ignored when HeadroomFixed is set.
	HeadroomPercent float64 `json:"headroom_percent"`
	// HeadroomFixed is a fixed-dollar headroom amount (takes precedence over HeadroomPercent).
	HeadroomFixed *float64 `json:"headroom_fixed_usd,omitempty"`
	// Mode controls what action is taken when the cushion threshold is crossed.
	Mode CushionMode `json:"mode"`
	// NotifyBeforeAction sends a warning alert before taking the configured action.
	NotifyBeforeAction bool `json:"notify_before_action"`
	// WarnLeadHours is how many hours before the projected action to send the warning.
	WarnLeadHours int `json:"warn_lead_hours"`
}

// CushionStatus describes whether the cushion threshold has been reached.
type CushionStatus struct {
	// Triggered is true when the remaining budget is at or below the headroom.
	Triggered bool `json:"triggered"`
	// Headroom is the configured dollar headroom.
	Headroom float64 `json:"headroom"`
	// Remaining is the actual remaining budget.
	Remaining float64 `json:"remaining"`
	// Message is a human-readable status description.
	Message string `json:"message"`
}

// CushionEvaluator evaluates and executes budget cushion actions.
type CushionEvaluator struct {
	executor ActionExecutor
	alerter  alerting.AlertDispatcher
}

// NewCushionEvaluator creates a CushionEvaluator.
func NewCushionEvaluator(executor ActionExecutor, alerter alerting.AlertDispatcher) *CushionEvaluator {
	return &CushionEvaluator{executor: executor, alerter: alerter}
}

// Evaluate checks whether the cushion threshold has been crossed.
//
// Returns the cushion status. Call Execute if status.Triggered is true.
func (ce *CushionEvaluator) Evaluate(status *BudgetStatus, config CushionConfig) CushionStatus {
	if !config.Enabled || status == nil || !status.BudgetEnabled {
		return CushionStatus{Message: "cushion not enabled"}
	}

	headroom := ce.headroomDollars(config, status.TotalBudget)
	remaining := status.RemainingBudget

	triggered := remaining <= headroom
	msg := fmt.Sprintf("remaining $%.2f > headroom $%.2f (SAFE)", remaining, headroom)
	if triggered {
		msg = fmt.Sprintf("remaining $%.2f ≤ headroom $%.2f — TRIGGERED (mode: %s)", remaining, headroom, config.Mode)
	}
	return CushionStatus{
		Triggered: triggered,
		Headroom:  headroom,
		Remaining: remaining,
		Message:   msg,
	}
}

// Execute runs the configured cushion action for the given project.
func (ce *CushionEvaluator) Execute(ctx context.Context, projectID, projectName string, config CushionConfig, status BudgetStatus) error {
	if ce.alerter != nil {
		alert := alerting.FormatCushionAlert(
			projectID, projectName,
			config.HeadroomPercent, status.RemainingBudget, status.TotalBudget,
			string(config.Mode),
		)
		_ = ce.alerter.Send(ctx, alert)
	}

	if ce.executor == nil {
		return nil
	}

	switch config.Mode {
	case CushionModeHibernate:
		return ce.executor.ExecuteHibernateAll(projectID)
	case CushionModeStop:
		return ce.executor.ExecuteStopAll(projectID)
	case CushionModePreventLaunch:
		return ce.executor.ExecutePreventLaunch(projectID)
	case CushionModeNotifyOnly:
		// alert already sent above
		return nil
	default:
		return fmt.Errorf("unknown cushion mode: %s", config.Mode)
	}
}

// headroomDollars returns the effective dollar headroom from config.
func (ce *CushionEvaluator) headroomDollars(config CushionConfig, totalBudget float64) float64 {
	if config.HeadroomFixed != nil && *config.HeadroomFixed > 0 {
		return *config.HeadroomFixed
	}
	return totalBudget * config.HeadroomPercent
}
