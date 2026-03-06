package cli

import (
	"github.com/spf13/cobra"
)

// SchoolsCobraCommands handles institution school registry commands
type SchoolsCobraCommands struct {
	app *App
}

// NewSchoolsCobraCommands creates new schools cobra commands
func NewSchoolsCobraCommands(app *App) *SchoolsCobraCommands {
	return &SchoolsCobraCommands{app: app}
}

// CreateSchoolsCommand creates the schools command with subcommands
func (sc *SchoolsCobraCommands) CreateSchoolsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schools",
		Short:   "Find your institution's AWS portal",
		GroupID: "system",
		Long: `Browse the community registry of institutional AWS portals.

Many universities and research institutions provide AWS accounts to their
researchers at no cost, or through sponsored credit programs. Use this
command to find the portal URL for your institution.

Registry source: https://github.com/scttfrdmn/prism-school-registry`,
	}

	cmd.AddCommand(
		sc.createListCommand(),
		sc.createSearchCommand(),
		sc.createInfoCommand(),
	)

	return cmd
}

func (sc *SchoolsCobraCommands) createListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered institutions",
		Long:  `List all institutions registered in the Prism school registry.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return sc.app.Schools([]string{"list"})
		},
	}
}

func (sc *SchoolsCobraCommands) createSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search by name, domain, or location",
		Long: `Search the school registry by institution name, email domain, state, or country.

Examples:
  prism schools search stanford
  prism schools search "university of washington"
  prism schools search mit.edu`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return sc.app.Schools(append([]string{"search"}, args...))
		},
	}
}

func (sc *SchoolsCobraCommands) createInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info <id-or-name>",
		Short: "Show portal URL and setup instructions",
		Long: `Show detailed information about an institution's AWS portal and setup steps.

The <id-or-name> argument can be the institution's short ID (e.g. "mit", "stanford")
or a partial name match.

Examples:
  prism schools info mit
  prism schools info stanford
  prism schools info uw`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return sc.app.Schools(append([]string{"info"}, args...))
		},
	}
}
