package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CapacityBlockCobraCommands groups capacity-block commands.
type CapacityBlockCobraCommands struct {
	app *App
}

// NewCapacityBlockCobraCommands creates the capacity-block command group.
func NewCapacityBlockCobraCommands(app *App) *CapacityBlockCobraCommands {
	return &CapacityBlockCobraCommands{app: app}
}

// CreateCapacityBlockCommand builds the 'capacity-block' top-level command.
func (cb *CapacityBlockCobraCommands) CreateCapacityBlockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capacity-block",
		Short: "Manage EC2 Capacity Blocks for ML GPU workloads",
		Long: `Reserve and manage EC2 Capacity Blocks to guarantee GPU availability
for scheduled ML training runs.

Examples:
  prism capacity-block reserve --type p3.8xlarge --count 2 \
      --start "2026-04-01T09:00:00Z" --hours 8
  prism capacity-block list
  prism capacity-block show <id>
  prism capacity-block cancel <id>`,
	}

	cmd.AddCommand(
		cb.createReserveCommand(),
		cb.createListCommand(),
		cb.createShowCommand(),
		cb.createCancelCommand(),
	)
	return cmd
}

func (cb *CapacityBlockCobraCommands) createReserveCommand() *cobra.Command {
	var instanceType, az, start string
	var count, hours int

	cmd := &cobra.Command{
		Use:   "reserve",
		Short: "Reserve a new capacity block",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cb.app.capacityBlockReserve(instanceType, az, start, count, hours)
		},
	}
	cmd.Flags().StringVar(&instanceType, "type", "", "EC2 instance type (e.g. p3.8xlarge) [required]")
	cmd.Flags().StringVar(&az, "az", "", "Availability zone (optional)")
	cmd.Flags().StringVar(&start, "start", "", "Start time RFC3339 (e.g. 2026-04-01T09:00:00Z) [required]")
	cmd.Flags().IntVar(&count, "count", 1, "Number of instances")
	cmd.Flags().IntVar(&hours, "hours", 1, "Duration in hours (1, 2, 4, 8, 12, 24)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("start")
	return cmd
}

func (cb *CapacityBlockCobraCommands) createListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List capacity blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cb.app.capacityBlockList()
		},
	}
}

func (cb *CapacityBlockCobraCommands) createShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show details of a capacity block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cb.app.capacityBlockShow(args[0])
		},
	}
}

func (cb *CapacityBlockCobraCommands) createCancelCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "Cancel a capacity block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cb.app.capacityBlockCancel(args[0])
		},
	}
}

// capacityBlockClient returns the capBlockHTTP interface or an error.
func (a *App) capacityBlockClient() (capBlockHTTP, error) {
	hc, ok := a.apiClient.(capBlockHTTP)
	if !ok {
		return nil, fmt.Errorf("capacity-block commands require daemon connection")
	}
	return hc, nil
}
