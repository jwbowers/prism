package cli

import (
	"context"
	"fmt"

	"github.com/scttfrdmn/prism/pkg/aws"
	"github.com/spf13/cobra"
)

// HealthCobraCommands handles AWS Health monitoring commands
type HealthCobraCommands struct {
	app *App
}

// NewHealthCobraCommands creates a new HealthCobraCommands instance
func NewHealthCobraCommands(app *App) *HealthCobraCommands {
	return &HealthCobraCommands{app: app}
}

// CreateHealthCommand creates the aws-health command
func (c *HealthCobraCommands) CreateHealthCommand() *cobra.Command {
	var allRegions bool

	cmd := &cobra.Command{
		Use:   "aws-health",
		Short: "Monitor AWS service health",
		Long: `Check AWS service health status for EC2 in your region or across all regions.

Monitors:
- Active service issues
- Scheduled maintenance
- Service impact levels
- Alternative healthy regions

Note: Requires AWS Business or Enterprise Support plan for full functionality.
      Without support, assumes services are healthy.`,
		Example: `  prism admin aws-health                    # Check current region
  prism admin aws-health --all-regions      # Check all regions`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return c.checkHealth(allRegions)
		},
	}

	cmd.Flags().BoolVar(&allRegions, "all-regions", false, "Check health across all AWS regions")

	return cmd
}

// checkHealth monitors AWS service health
func (c *HealthCobraCommands) checkHealth(allRegions bool) error {
	ctx := context.Background()

	// Get AWS manager
	manager, err := c.app.getAWSManager()
	if err != nil {
		return fmt.Errorf("failed to initialize AWS manager: %w", err)
	}

	healthMonitor := manager.GetHealthMonitor()

	if allRegions {
		return c.checkAllRegions(ctx, healthMonitor, manager.GetDefaultRegion())
	}

	return c.checkCurrentRegion(ctx, healthMonitor, manager.GetDefaultRegion())
}

// checkCurrentRegion checks health for the current region
func (c *HealthCobraCommands) checkCurrentRegion(ctx context.Context, healthMonitor *aws.HealthMonitor, region string) error {
	status, err := healthMonitor.CheckServiceHealth(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to check service health: %w", err)
	}

	// Display results
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🏥 AWS Service Health - %s\n", region)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	fmt.Printf("Service:        %s\n", status.Service)
	fmt.Printf("Region:         %s\n", status.Region)
	fmt.Printf("Active Events:  %d\n", status.ActiveEvents)
	fmt.Printf("Impact Level:   %s\n", status.ImpactLevel)
	fmt.Println()

	// Status indicator
	if status.IsHealthy {
		fmt.Println("✅ Status: HEALTHY")
	} else {
		fmt.Println("❌ Status: UNHEALTHY")
	}

	fmt.Println()
	fmt.Printf("💡 Recommended Action:\n   %s\n", status.RecommendedAction)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Display event details if any
	if len(status.Events) > 0 {
		fmt.Println()
		fmt.Println("Active Events:")
		for i, event := range status.Events {
			fmt.Printf("\n%d. %s\n", i+1, event.EventTypeCode)
			fmt.Printf("   Status:   %s\n", event.StatusCode)
			fmt.Printf("   Category: %s\n", event.EventTypeCategory)
			fmt.Printf("   Started:  %s\n", event.StartTime.Format("2006-01-02 15:04:05 MST"))
			if event.EndTime != nil {
				fmt.Printf("   Ended:    %s\n", event.EndTime.Format("2006-01-02 15:04:05 MST"))
			}
		}
		fmt.Println()
	}

	return nil
}

// checkAllRegions checks health across all AWS regions
func (c *HealthCobraCommands) checkAllRegions(ctx context.Context, healthMonitor *aws.HealthMonitor, currentRegion string) error {
	// Common AWS regions
	regions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
		"ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
	}

	summary, err := healthMonitor.GetRegionHealthSummary(ctx, regions)
	if err != nil {
		return fmt.Errorf("failed to get region health summary: %w", err)
	}

	// Format and display report
	report := healthMonitor.FormatHealthReport(summary)
	fmt.Println(report)

	// Suggest alternatives if current region is affected
	if len(summary.AffectedRegions) > 0 {
		for _, affected := range summary.AffectedRegions {
			if affected == currentRegion {
				fmt.Println()
				fmt.Printf("⚠️  Your current region (%s) is affected!\n", currentRegion)
				if len(summary.HealthyRegions) > 0 {
					fmt.Println()
					fmt.Println("💡 Alternative Healthy Regions:")
					for _, healthy := range summary.HealthyRegions {
						fmt.Printf("   • %s\n", healthy)
					}
					fmt.Println()
					fmt.Println("To switch regions, use: prism profile switch <profile-name>")
				}
				break
			}
		}
	}

	return nil
}
