package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// FileOpsCobraCommands groups SSM file-transfer commands.
type FileOpsCobraCommands struct {
	app *App
}

// NewFileOpsCobraCommands creates the file ops command group.
func NewFileOpsCobraCommands(app *App) *FileOpsCobraCommands {
	return &FileOpsCobraCommands{app: app}
}

// CreateFilesCommand builds the 'workspace files' sub-command tree.
func (fc *FileOpsCobraCommands) CreateFilesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "Transfer files to/from a running instance via SSM",
		Long: `Transfer files between your local machine and a running instance.

Files are relayed through a temporary S3 bucket — the instance must be running
and have the SSM agent + an IAM role that allows s3:GetObject/PutObject on the
Prism temp bucket (prism-temp-{acct}-{region}).

Examples:
  prism workspace files list  my-instance --path /home/ec2-user
  prism workspace files push  my-instance ./data.csv /home/ec2-user/data.csv
  prism workspace files pull  my-instance /home/ec2-user/results.csv ./results.csv`,
	}

	var listPath string
	listCmd := &cobra.Command{
		Use:   "list <instance>",
		Short: "List files on a running instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fc.app.fileOpsList(args[0], listPath)
		},
	}
	listCmd.Flags().StringVar(&listPath, "path", "/home", "Remote path to list")

	pushCmd := &cobra.Command{
		Use:   "push <instance> <local-path> <remote-path>",
		Short: "Upload a local file to a running instance",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fc.app.fileOpsPush(args[0], args[1], args[2])
		},
	}

	pullCmd := &cobra.Command{
		Use:   "pull <instance> <remote-path> <local-path>",
		Short: "Download a file from a running instance",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fc.app.fileOpsPull(args[0], args[1], args[2])
		},
	}

	cmd.AddCommand(listCmd, pushCmd, pullCmd)
	return cmd
}

// fileOpsClient returns the HTTPClient or an error.
func (a *App) fileOpsClient() (fileOpsHTTP, error) {
	hc, ok := a.apiClient.(fileOpsHTTP)
	if !ok {
		return nil, fmt.Errorf("file-ops commands require daemon connection")
	}
	return hc, nil
}
