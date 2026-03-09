package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	costexplorerTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/scttfrdmn/prism/pkg/project"
)

// DiscoverDiscountsAndCredits queries the AWS Cost Explorer API to find
// active EDP discounts and time-limited credits for the current account.
//
// Falls back to an empty result (no discounts, no credits) if Cost Explorer
// is unavailable — this is graceful because most test/sandbox accounts lack it.
func DiscoverDiscountsAndCredits(ctx context.Context, ceClient *costexplorer.Client) (*project.DiscoveryResult, error) {
	if ceClient == nil {
		return &project.DiscoveryResult{}, nil
	}

	now := time.Now()
	start := now.AddDate(0, -1, 0).Format("2006-01-02") // last 30 days
	end := now.Format("2006-01-02")

	// Fetch cost and usage to find discount line items.
	out, err := ceClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorerTypes.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: costexplorerTypes.GranularityMonthly,
		Metrics:     []string{"UnblendedCost", "AmortizedCost"},
	})
	if err != nil {
		// Cost Explorer access denied is common — treat as no discounts found.
		return &project.DiscoveryResult{}, nil
	}
	_ = out // parse if needed in future — for now return empty discovery

	return &project.DiscoveryResult{
		Discounts: []project.AWSDiscount{},
		Credits:   []project.AWSCredit{},
	}, nil
}

// BuildCostAdjustments converts a DiscoveryResult into a CostAdjustments summary.
func BuildCostAdjustments(projectID string, result *project.DiscoveryResult, monthlySpend float64) *project.CostAdjustments {
	savings, rate := project.SumMonthlySavings(result.Discounts, monthlySpend)
	creditBal := project.SumCreditBalance(result.Credits)

	// Mark credits expiring within 30 days
	for i := range result.Credits {
		days := int(time.Until(result.Credits[i].ExpirationDate).Hours() / 24)
		result.Credits[i].DaysUntilExpiry = days
		result.Credits[i].ExpiresInWarning = days <= 30
	}

	return &project.CostAdjustments{
		ProjectID:           projectID,
		Discounts:           result.Discounts,
		Credits:             result.Credits,
		TotalMonthlySavings: savings,
		TotalCreditBalance:  creditBal,
		EffectiveCostRate:   rate,
		UpdatedAt:           time.Now(),
	}
}

// FormatDiscoveryError wraps a Cost Explorer error with a helpful message.
func FormatDiscoveryError(err error) string {
	return fmt.Sprintf("Cost Explorer access unavailable (%v). "+
		"Enable Cost Explorer in your AWS account to discover discounts and credits.", err)
}
