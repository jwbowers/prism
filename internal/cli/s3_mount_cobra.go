package cli

import (
	"github.com/spf13/cobra"
)

// S3MountCobraCommands handles `prism storage s3` commands (Cobra layer)
type S3MountCobraCommands struct {
	app *App
}

// NewS3MountCobraCommands creates new S3 mount cobra commands.
func NewS3MountCobraCommands(app *App) *S3MountCobraCommands {
	return &S3MountCobraCommands{app: app}
}

// CreateS3Command creates the `prism storage s3` subcommand group.
func (sc *S3MountCobraCommands) CreateS3Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Manage S3 bucket mounts on instances",
		Long: `Mount and unmount S3 buckets on running instances via mountpoint-s3 or s3fs.

Examples:
  prism storage s3 list my-instance
  prism storage s3 mount my-instance my-bucket /mnt/data
  prism storage s3 mount my-instance my-bucket /mnt/data --method s3fs --read-only
  prism storage s3 unmount my-instance /mnt/data`,
	}

	mountCmd := &cobra.Command{
		Use:   "mount <instance> <bucket> <mount-path>",
		Short: "Mount an S3 bucket on a running instance",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			method, _ := cmd.Flags().GetString("method")
			readOnly, _ := cmd.Flags().GetBool("read-only")
			return sc.app.s3MountMount(args[0], args[1], args[2], method, readOnly)
		},
	}
	mountCmd.Flags().String("method", "mountpoint", "Mount method: mountpoint or s3fs")
	mountCmd.Flags().Bool("read-only", false, "Mount as read-only")

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list <instance>",
			Short: "List active S3 mounts on an instance",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return sc.app.s3MountList(args[0])
			},
		},
		mountCmd,
		&cobra.Command{
			Use:   "unmount <instance> <mount-path>",
			Short: "Unmount an S3 bucket from an instance",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return sc.app.s3MountUnmount(args[0], args[1])
			},
		},
	)

	return cmd
}
