package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/scttfrdmn/prism/pkg/update"
	"github.com/scttfrdmn/prism/pkg/version"
	"github.com/spf13/cobra"
)

// NewVersionCommand creates a new version command
func NewVersionCommand() *cobra.Command {
	var checkUpdate bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display Prism version information including build details.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show version info
			fmt.Println(version.GetCLIVersionInfo())

			// Check for updates if requested
			if checkUpdate {
				fmt.Println()
				checkForUpdates()
			}
		},
	}

	cmd.Flags().BoolVar(&checkUpdate, "check-update", false, "Check for available updates")

	return cmd
}

// checkForUpdates checks for and displays update information
func checkForUpdates() {
	fmt.Println("🔍 Checking for updates...")

	checker, err := update.NewChecker()
	if err != nil {
		fmt.Printf("❌ Failed to initialize update checker: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	updateInfo, err := checker.CheckForUpdates(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to check for updates: %v\n", err)
		fmt.Println("💡 You can manually check for updates at https://github.com/scttfrdmn/prism/releases")
		return
	}

	if updateInfo.IsUpdateAvailable {
		fmt.Println(update.FormatUpdateMessage(updateInfo))
	} else {
		fmt.Printf("✅ You're running the latest version (%s)\n", updateInfo.CurrentVersion)
	}
}
