package project

import (
	"fmt"
	"math"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// ForecastMonth is a per-month spending projection.
type ForecastMonth struct {
	Month           string   `json:"month"` // "2026-04"
	ProjectedSpend  float64  `json:"projected_spend"`
	CumulativeSpend float64  `json:"cumulative_spend"`
	RemainingBudget float64  `json:"remaining_budget"`
	IsProjected     bool     `json:"is_projected"` // false for historical months
	ActualSpend     *float64 `json:"actual_spend,omitempty"`
}

// GrantRenewalAlert fires when budget exhaustion is predicted before a grant renewal date.
type GrantRenewalAlert struct {
	RenewalDate       time.Time `json:"renewal_date"`
	DaysBeforeRenewal int       `json:"days_before_renewal"` // exhaustion N days before renewal
	ShortfallAmount   float64   `json:"shortfall_amount"`
	Message           string    `json:"message"`
}

// ShortfallPrediction contains all forward-looking budget analysis.
type ShortfallPrediction struct {
	ProjectID             string             `json:"project_id"`
	GeneratedAt           time.Time          `json:"generated_at"`
	CurrentDailyRate      float64            `json:"current_daily_rate"`
	TrendSlope            float64            `json:"trend_slope"` // $/day change per day (acceleration)
	PredictedExhaustionAt *time.Time         `json:"predicted_exhaustion_at,omitempty"`
	DaysUntilExhaustion   *int               `json:"days_until_exhaustion,omitempty"`
	WillExceedBudget      bool               `json:"will_exceed_budget"`
	ProbabilityExceed     float64            `json:"probability_exceed"` // 0.0–1.0
	ConfidenceLevel       string             `json:"confidence_level"`   // low / medium / high
	MonthlyForecasts      []ForecastMonth    `json:"monthly_forecasts"`
	GrantRenewalAlert     *GrantRenewalAlert `json:"grant_renewal_alert,omitempty"`
}

// BudgetPredictor uses cost history to forecast future spending.
type BudgetPredictor struct {
	burnCalc    *BurnRateCalculator
	surplusCalc *SurplusCalculator
}

// NewBudgetPredictor creates a new predictor with default calculators.
func NewBudgetPredictor() *BudgetPredictor {
	return &BudgetPredictor{
		burnCalc:    &BurnRateCalculator{},
		surplusCalc: &SurplusCalculator{},
	}
}

// Predict performs forward projection using linear regression on recent cost history.
//
// months is the number of future months to forecast.  grantRenewalDate may be nil.
func (p *BudgetPredictor) Predict(
	projectID string,
	history []CostDataPoint,
	budget *types.ProjectBudget,
	months int,
	grantRenewalDate *time.Time,
) (*ShortfallPrediction, error) {
	now := time.Now()

	dailyRate := p.burnCalc.DailyBurnRate(history, 7)
	_, slope := linearRegressionSlope(history)

	remaining := budget.TotalBudget - budget.SpentAmount
	if remaining < 0 {
		remaining = 0
	}

	// Estimate exhaustion date.
	var exhaustionAt *time.Time
	var daysUntil *int
	willExceed := false
	prob := 0.0

	if dailyRate > 0 && remaining > 0 {
		days := remaining / dailyRate
		t := now.Add(time.Duration(days*24) * time.Hour)
		exhaustionAt = &t
		d := int(math.Round(days))
		daysUntil = &d
		willExceed = true
		prob = 1.0
	} else if dailyRate <= 0 {
		willExceed = false
		prob = 0.0
	}

	// Confidence based on data density: high = 14+ days, medium = 7+, low otherwise.
	confidence := "low"
	if len(history) >= 2 {
		span := history[len(history)-1].Timestamp.Sub(history[0].Timestamp).Hours() / 24
		if span >= 14 {
			confidence = "high"
		} else if span >= 7 {
			confidence = "medium"
		}
	}

	// Build monthly forecasts.
	forecasts := buildMonthlyForecasts(history, budget.SpentAmount, budget.TotalBudget, dailyRate, slope, months)

	// Grant renewal alert.
	var renewalAlert *GrantRenewalAlert
	if grantRenewalDate != nil && exhaustionAt != nil && exhaustionAt.Before(*grantRenewalDate) {
		daysGap := int(math.Round(grantRenewalDate.Sub(*exhaustionAt).Hours() / 24))
		shortfall := dailyRate * float64(daysGap)
		renewalAlert = &GrantRenewalAlert{
			RenewalDate:       *grantRenewalDate,
			DaysBeforeRenewal: daysGap,
			ShortfallAmount:   shortfall,
			Message: fmt.Sprintf(
				"Budget exhaustion predicted %d days before grant renewal on %s — $%.2f shortfall",
				daysGap, grantRenewalDate.Format("2006-01-02"), shortfall,
			),
		}
	}

	return &ShortfallPrediction{
		ProjectID:             projectID,
		GeneratedAt:           now,
		CurrentDailyRate:      dailyRate,
		TrendSlope:            slope,
		PredictedExhaustionAt: exhaustionAt,
		DaysUntilExhaustion:   daysUntil,
		WillExceedBudget:      willExceed,
		ProbabilityExceed:     prob,
		ConfidenceLevel:       confidence,
		MonthlyForecasts:      forecasts,
		GrantRenewalAlert:     renewalAlert,
	}, nil
}

// linearRegressionSlope computes the least-squares intercept and slope ($/day) of
// cumulative spending vs. time.  Returns (intercept, slope).
func linearRegressionSlope(history []CostDataPoint) (float64, float64) {
	n := len(history)
	if n < 2 {
		return 0, 0
	}
	t0 := history[0].Timestamp
	var sumX, sumY, sumXY, sumX2 float64
	fn := float64(n)
	for _, p := range history {
		x := p.Timestamp.Sub(t0).Hours() / 24 // days since first point
		y := p.TotalCost
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := fn*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, 0
	}
	slope := (fn*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / fn
	return intercept, slope
}

// buildMonthlyForecasts produces a slice of ForecastMonth combining historical data
// with forward projections.
func buildMonthlyForecasts(
	history []CostDataPoint,
	spentToDate, totalBudget float64,
	dailyRate, slope float64,
	futureMonths int,
) []ForecastMonth {
	now := time.Now()
	var forecasts []ForecastMonth

	// One historical month (current month-to-date as actual).
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var mtdSpend float64
	for _, p := range history {
		if !p.Timestamp.Before(monthStart) {
			// Use daily cost field; fall back to TotalCost delta.
			if p.DailyCost > 0 {
				mtdSpend += p.DailyCost
			}
		}
	}
	cumulative := spentToDate
	remaining := totalBudget - cumulative
	actual := mtdSpend
	forecasts = append(forecasts, ForecastMonth{
		Month:           now.Format("2006-01"),
		ProjectedSpend:  dailyRate * 30,
		CumulativeSpend: cumulative,
		RemainingBudget: remaining,
		IsProjected:     false,
		ActualSpend:     &actual,
	})

	// Future months.
	rate := dailyRate
	for i := 1; i <= futureMonths; i++ {
		rate += slope * 30 // adjust for trend
		if rate < 0 {
			rate = 0
		}
		projected := rate * 30
		cumulative += projected
		remaining = totalBudget - cumulative
		if remaining < 0 {
			remaining = 0
		}
		target := monthStart.AddDate(0, i, 0)
		forecasts = append(forecasts, ForecastMonth{
			Month:           target.Format("2006-01"),
			ProjectedSpend:  projected,
			CumulativeSpend: cumulative,
			RemainingBudget: remaining,
			IsProjected:     true,
		})
	}
	return forecasts
}
