package project

import (
	"fmt"
	"time"
)

// AWSDiscount represents a discovered AWS discount (EDP or PPA).
type AWSDiscount struct {
	DiscountID string     `json:"discount_id"`
	Type       string     `json:"type"` // "EDP" | "PPA"
	Percentage float64    `json:"percentage,omitempty"`
	Services   []string   `json:"services,omitempty"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Active     bool       `json:"active"`
}

// AWSCredit represents a time-limited AWS credit.
type AWSCredit struct {
	CreditID           string    `json:"credit_id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	InitialAmount      float64   `json:"initial_amount"`
	RemainingAmount    float64   `json:"remaining_amount"`
	StartDate          time.Time `json:"start_date"`
	ExpirationDate     time.Time `json:"expiration_date"`
	ApplicableServices []string  `json:"applicable_services,omitempty"`
	Active             bool      `json:"active"`
	DaysUntilExpiry    int       `json:"days_until_expiry"`
	ExpiresInWarning   bool      `json:"expires_in_warning"` // true if < 30 days
}

// CostAdjustments summarizes all active discounts and credits for a project.
type CostAdjustments struct {
	ProjectID           string        `json:"project_id"`
	Discounts           []AWSDiscount `json:"discounts"`
	Credits             []AWSCredit   `json:"credits"`
	TotalMonthlySavings float64       `json:"total_monthly_savings"`
	TotalCreditBalance  float64       `json:"total_credit_balance"`
	EffectiveCostRate   float64       `json:"effective_cost_rate"` // 0.90 = 10% discount
	UpdatedAt           time.Time     `json:"updated_at"`
}

// DiscoveryResult is what the discovery process returns.
type DiscoveryResult struct {
	Discounts []AWSDiscount
	Credits   []AWSCredit
	Error     error
}

// MockDiscovery returns simulated discounts and credits for environments
// without Cost Explorer access. Used as a fallback.
func MockDiscovery(projectID string) *CostAdjustments {
	return &CostAdjustments{
		ProjectID:           projectID,
		Discounts:           []AWSDiscount{},
		Credits:             []AWSCredit{},
		TotalMonthlySavings: 0,
		TotalCreditBalance:  0,
		EffectiveCostRate:   1.0,
		UpdatedAt:           time.Now(),
	}
}

// ApplyAdjustments applies discovered discounts to a cost estimate.
func ApplyAdjustments(baseCost float64, adj *CostAdjustments) float64 {
	return baseCost * adj.EffectiveCostRate
}

// SumMonthlySavings computes total monthly savings from all active discounts
// given the base monthly spend.
func SumMonthlySavings(discounts []AWSDiscount, monthlySpend float64) (totalSavings float64, effectiveRate float64) {
	effectiveRate = 1.0
	for _, d := range discounts {
		if !d.Active {
			continue
		}
		if d.Type == "EDP" {
			effectiveRate *= (1.0 - d.Percentage/100.0)
		}
	}
	totalSavings = monthlySpend * (1.0 - effectiveRate)
	return
}

// SumCreditBalance returns total remaining credit across all active credits.
func SumCreditBalance(credits []AWSCredit) float64 {
	total := 0.0
	for _, c := range credits {
		if c.Active && time.Now().Before(c.ExpirationDate) {
			total += c.RemainingAmount
		}
	}
	return total
}

// CreditExpiryWarnings returns credits expiring within warnDays days.
func CreditExpiryWarnings(credits []AWSCredit, warnDays int) []AWSCredit {
	var expiring []AWSCredit
	deadline := time.Now().AddDate(0, 0, warnDays)
	for _, c := range credits {
		if c.Active && c.ExpirationDate.Before(deadline) && time.Now().Before(c.ExpirationDate) {
			expiring = append(expiring, c)
		}
	}
	return expiring
}

// FormatAdjustmentSummary returns a human-readable summary string.
func FormatAdjustmentSummary(adj *CostAdjustments) string {
	if len(adj.Discounts) == 0 && len(adj.Credits) == 0 {
		return "No active discounts or credits discovered."
	}
	msg := fmt.Sprintf("Discounts: %d active, Credits: $%.2f remaining",
		len(adj.Discounts), adj.TotalCreditBalance)
	if adj.EffectiveCostRate < 1.0 {
		msg += fmt.Sprintf(" (%.0f%% discount applied)", (1-adj.EffectiveCostRate)*100)
	}
	return msg
}
