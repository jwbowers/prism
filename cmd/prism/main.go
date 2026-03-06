// Prism CLI client - Launch research computing environments.
//
// The prism command-line tool provides a simple interface for managing cloud
// research workstations. It communicates with the Prism daemon (prismd)
// to launch pre-configured environments optimized for academic research.
//
// Core Commands:
//
//	prism launch template-name instance-name  # Launch new research environment
//	prism list                                # Show running instances and costs
//	prism connect instance-name               # Get connection information
//	prism stop/start instance-name            # Manage instance lifecycle
//
// Storage Commands:
//
//	prism volumes create/list/delete          # Manage EFS shared storage
//	prism storage create/list/delete          # Manage EBS high-performance storage
//
// Examples:
//
//	prism launch r-research my-analysis       # Launch R environment
//	prism launch python-ml gpu-training --size GPU-L  # Launch ML environment
//	prism list                                # Show all instances
//	prism connect my-analysis                 # Get SSH/web URLs
//
// The CLI implements Prism's "Default to Success" principle -
// every command works out of the box with smart defaults while providing
// advanced options for power users.
package main

import (
	"fmt"
	"os"

	"github.com/scttfrdmn/prism/internal/cli"
	"github.com/scttfrdmn/prism/pkg/version"
)

func main() {
	// Create app
	cliApp := cli.NewApp(version.GetVersion())

	// Use the Cobra-based system for all commands
	err := cliApp.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", cli.FormatErrorForCLI(err, "command execution"))
		os.Exit(1)
	}
}
