package cli

import (
	"fmt"

	"github.com/scttfrdmn/prism/pkg/version"
	"github.com/spf13/cobra"
)

// NewVersionCommand creates a new version command
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display Prism version information including build details.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetCLIVersionInfo())
		},
	}

	return cmd
}
