package cli

import (
	"github.com/spf13/cobra"
)

// WorkshopCobraCommands handles workshop management commands.
type WorkshopCobraCommands struct {
	app *App
}

// NewWorkshopCobraCommands creates new workshop cobra commands.
func NewWorkshopCobraCommands(app *App) *WorkshopCobraCommands {
	return &WorkshopCobraCommands{app: app}
}

// CreateWorkshopCommand creates the 'workshop' top-level command.
func (wc *WorkshopCobraCommands) CreateWorkshopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workshop",
		Short: "Manage workshops and events",
		Long: `Manage time-bounded workshops: conferences, tutorials, and hackathons.

Examples:
  prism workshop create --title "NeurIPS DL Tutorial" --template pytorch-ml \
        --max-participants 60 --start 2026-12-08T09:00:00 --end 2026-12-08T15:00:00
  prism workshop list
  prism workshop show <id>
  prism workshop provision <id>
  prism workshop dashboard <id>
  prism workshop end <id>
  prism workshop download <id>
  prism workshop config save <id> <name>
  prism workshop config list
  prism workshop config use <name> --title "New Workshop" --start 2027-01-10T09:00:00`,
	}

	cmd.AddCommand(
		wc.createListCommand(),
		wc.createCreateCommand(),
		wc.createShowCommand(),
		wc.createDeleteCommand(),
		wc.createProvisionCommand(),
		wc.createDashboardCommand(),
		wc.createEndCommand(),
		wc.createDownloadCommand(),
		wc.createConfigCommand(),
	)
	return cmd
}

func (wc *WorkshopCobraCommands) createListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workshops",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, _ := cmd.Flags().GetString("owner")
			status, _ := cmd.Flags().GetString("status")
			cliArgs := []string{"list"}
			if owner != "" {
				cliArgs = append(cliArgs, "--owner", owner)
			}
			if status != "" {
				cliArgs = append(cliArgs, "--status", status)
			}
			return wc.app.Workshop(cliArgs)
		},
	}
	cmd.Flags().String("owner", "", "Filter by owner user ID")
	cmd.Flags().String("status", "", "Filter by status (draft, active, ended, archived)")
	return cmd
}

func (wc *WorkshopCobraCommands) createCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workshop",
		RunE: func(cmd *cobra.Command, args []string) error {
			title, _ := cmd.Flags().GetString("title")
			template, _ := cmd.Flags().GetString("template")
			owner, _ := cmd.Flags().GetString("owner")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			maxPax, _ := cmd.Flags().GetInt("max-participants")
			budget, _ := cmd.Flags().GetFloat64("budget-per-participant")
			earlyAccess, _ := cmd.Flags().GetInt("early-access")
			desc, _ := cmd.Flags().GetString("description")

			cliArgs := []string{"create"}
			if title != "" {
				cliArgs = append(cliArgs, "--title", title)
			}
			if template != "" {
				cliArgs = append(cliArgs, "--template", template)
			}
			if owner != "" {
				cliArgs = append(cliArgs, "--owner", owner)
			}
			if start != "" {
				cliArgs = append(cliArgs, "--start", start)
			}
			if end != "" {
				cliArgs = append(cliArgs, "--end", end)
			}
			if maxPax > 0 {
				cliArgs = append(cliArgs, "--max-participants", itoa(maxPax))
			}
			if budget > 0 {
				cliArgs = append(cliArgs, "--budget-per-participant", ftoa(budget))
			}
			if earlyAccess > 0 {
				cliArgs = append(cliArgs, "--early-access", itoa(earlyAccess))
			}
			if desc != "" {
				cliArgs = append(cliArgs, "--description", desc)
			}
			return wc.app.Workshop(cliArgs)
		},
	}
	cmd.Flags().String("title", "", "Workshop title (required)")
	cmd.Flags().String("template", "", "Workspace template slug (required)")
	cmd.Flags().String("owner", "", "Organizer user ID (required)")
	cmd.Flags().String("start", "", "Start time in RFC3339 or YYYY-MM-DDTHH:MM:SS (required)")
	cmd.Flags().String("end", "", "End time in RFC3339 or YYYY-MM-DDTHH:MM:SS (required)")
	cmd.Flags().Int("max-participants", 0, "Maximum participant count (0 = unlimited)")
	cmd.Flags().Float64("budget-per-participant", 0, "Per-participant budget in USD")
	cmd.Flags().Int("early-access", 0, "Hours of early access before start time")
	cmd.Flags().String("description", "", "Workshop description")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("template")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")
	return cmd
}

func (wc *WorkshopCobraCommands) createShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show workshop details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.app.Workshop([]string{"show", args[0]})
		},
	}
}

func (wc *WorkshopCobraCommands) createDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a workshop",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.app.Workshop([]string{"delete", args[0]})
		},
	}
}

func (wc *WorkshopCobraCommands) createProvisionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "provision <id>",
		Short: "Bulk-provision workspaces for all participants",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.app.Workshop([]string{"provision", args[0]})
		},
	}
}

func (wc *WorkshopCobraCommands) createDashboardCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard <id>",
		Short: "Show live participant status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.app.Workshop([]string{"dashboard", args[0]})
		},
	}
}

func (wc *WorkshopCobraCommands) createEndCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "end <id>",
		Short: "End a workshop and stop all participant instances",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.app.Workshop([]string{"end", args[0]})
		},
	}
}

func (wc *WorkshopCobraCommands) createDownloadCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "download <id>",
		Short: "Show download manifest for participant work",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.app.Workshop([]string{"download", args[0]})
		},
	}
}

func (wc *WorkshopCobraCommands) createConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage reusable workshop config templates",
	}

	saveCmd := &cobra.Command{
		Use:   "save <workshop-id> <config-name>",
		Short: "Save a workshop's settings as a reusable config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.app.Workshop([]string{"config", "save", args[0], args[1]})
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List saved workshop configs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.app.Workshop([]string{"config", "list"})
		},
	}

	useCmd := &cobra.Command{
		Use:   "use <config-name>",
		Short: "Create a new workshop from a saved config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title, _ := cmd.Flags().GetString("title")
			start, _ := cmd.Flags().GetString("start")
			owner, _ := cmd.Flags().GetString("owner")

			cliArgs := []string{"config", "use", args[0]}
			if title != "" {
				cliArgs = append(cliArgs, "--title", title)
			}
			if start != "" {
				cliArgs = append(cliArgs, "--start", start)
			}
			if owner != "" {
				cliArgs = append(cliArgs, "--owner", owner)
			}
			return wc.app.Workshop(cliArgs)
		},
	}
	useCmd.Flags().String("title", "", "Workshop title (required)")
	useCmd.Flags().String("start", "", "Start time (required)")
	useCmd.Flags().String("owner", "", "Organizer user ID")
	_ = useCmd.MarkFlagRequired("title")
	_ = useCmd.MarkFlagRequired("start")

	cmd.AddCommand(saveCmd, listCmd, useCmd)
	return cmd
}
