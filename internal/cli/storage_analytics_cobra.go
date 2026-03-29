package cli

import (
	"github.com/spf13/cobra"
)

// StorageAnalyticsCobraCommands handles `prism storage analytics` commands.
type StorageAnalyticsCobraCommands struct {
	app *App
}

// NewStorageAnalyticsCobraCommands creates new storage analytics cobra commands.
func NewStorageAnalyticsCobraCommands(app *App) *StorageAnalyticsCobraCommands {
	return &StorageAnalyticsCobraCommands{app: app}
}

// CreateAnalyticsCommand creates the `prism storage analytics` subcommand.
func (sc *StorageAnalyticsCobraCommands) CreateAnalyticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics [name]",
		Short: "Show storage cost and usage analytics",
		Long: `Show cost and usage analytics for storage volumes.

With no arguments, shows a summary of all storage volumes.
With a name argument, shows detailed analytics for that volume.

Examples:
  prism storage analytics
  prism storage analytics my-volume
  prism storage analytics --period weekly
  prism storage analytics my-volume --period monthly`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			period, _ := cmd.Flags().GetString("period")
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return sc.app.storageAnalytics(name, period)
		},
	}
	cmd.Flags().String("period", "daily", "Analytics period: daily, weekly, or monthly")

	return cmd
}
