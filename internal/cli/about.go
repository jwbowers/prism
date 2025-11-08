package cli

import (
	"fmt"
	"runtime"

	"github.com/scttfrdmn/prism/pkg/version"
	"github.com/spf13/cobra"
)

// NewAboutCommand creates a new about command
func NewAboutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "about",
		Short: "Show information about Prism",
		Long: `Display detailed information about Prism including version,
build information, platform details, and project links.`,
		Run: func(cmd *cobra.Command, args []string) {
			runAbout()
		},
	}

	return cmd
}

// runAbout displays comprehensive information about Prism
func runAbout() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║             Prism - Academic Research              ║")
	fmt.Println("║             Computing Platform                                 ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Version Information
	fmt.Println("📦 Version Information")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("   Version:        %s\n", version.GetVersion())
	if version.GitCommit != "" {
		commitShort := version.GitCommit
		if len(commitShort) > 8 {
			commitShort = commitShort[:8]
		}
		fmt.Printf("   Git Commit:     %s\n", commitShort)
	}
	if version.BuildDate != "" {
		fmt.Printf("   Build Date:     %s\n", version.BuildDate)
	}
	if version.GoVersion != "" {
		fmt.Printf("   Go Version:     %s\n", version.GoVersion)
	} else {
		fmt.Printf("   Go Version:     %s\n", runtime.Version())
	}
	fmt.Println()

	// Platform Information
	fmt.Println("💻 Platform Information")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("   OS:             %s\n", runtime.GOOS)
	fmt.Printf("   Architecture:   %s\n", runtime.GOARCH)
	fmt.Printf("   CPUs:           %d\n", runtime.NumCPU())
	fmt.Println()

	// Project Information
	fmt.Println("🔗 Project Links")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   Website:        https://cloudworkstation.io")
	fmt.Println("   Documentation:  https://docs.prism.io")
	fmt.Println("   GitHub:         https://github.com/scttfrdmn/prism")
	fmt.Println("   Issues:         https://github.com/scttfrdmn/prism/issues")
	fmt.Println()

	// Description
	fmt.Println("📝 About")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   Prism provides researchers with pre-configured cloud")
	fmt.Println("   computing environments for data analysis, machine learning, and")
	fmt.Println("   research computing. Launch production-ready environments in seconds")
	fmt.Println("   rather than spending hours on setup and configuration.")
	fmt.Println()

	// Features
	fmt.Println("✨ Key Features")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   • Pre-configured research templates (ML, R, Jupyter, etc.)")
	fmt.Println("   • Multi-modal access (CLI, TUI, GUI)")
	fmt.Println("   • Project-based budget management")
	fmt.Println("   • Cost optimization with hibernation")
	fmt.Println("   • Multi-user research collaboration")
	fmt.Println("   • Template marketplace and sharing")
	fmt.Println()

	// License and Copyright
	fmt.Println("📄 License")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   Copyright © 2024-2025 Scott Friedman and Prism Contributors")
	fmt.Println("   Licensed under the Apache License, Version 2.0")
	fmt.Println()

	// Quick Help
	fmt.Println("💡 Quick Start")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   cws --help             Show all available commands")
	fmt.Println("   prism templates          List available templates")
	fmt.Println("   prism launch <template> <name>  Launch a new workstation")
	fmt.Println("   prism tui                Launch terminal interface")
	fmt.Println("   prism gui                Launch graphical interface")
	fmt.Println()
}
