package update

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DetectInstallMethod determines how Prism was installed
func DetectInstallMethod() InstallMethod {
	// Get the path of the currently running binary
	execPath, err := os.Executable()
	if err != nil {
		return InstallMethodUnknown
	}

	// Resolve symlinks to get the actual binary path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	// Normalize path for comparison
	realPath = filepath.Clean(realPath)

	// Check for Homebrew installation (macOS/Linux)
	if isHomebrewInstall(realPath) {
		return InstallMethodHomebrew
	}

	// Check for Scoop installation (Windows)
	if runtime.GOOS == "windows" && isScoopInstall(realPath) {
		return InstallMethodScoop
	}

	// Check for source/development installation
	if isSourceInstall(realPath) {
		return InstallMethodSource
	}

	// Default to binary installation (direct download)
	return InstallMethodBinary
}

// isHomebrewInstall checks if binary is in Homebrew paths
func isHomebrewInstall(path string) bool {
	homebrewPaths := []string{
		"/opt/homebrew",       // Apple Silicon Macs
		"/usr/local/Homebrew", // Intel Macs
		"/home/linuxbrew",     // Linux Homebrew
		"/.linuxbrew",         // Alternative Linux path
	}

	for _, homebrewPath := range homebrewPaths {
		if strings.Contains(path, homebrewPath) {
			return true
		}
	}

	return false
}

// isScoopInstall checks if binary is in Scoop paths
func isScoopInstall(path string) bool {
	// Scoop typically installs to ~/scoop/apps or C:\ProgramData\scoop
	scoopPaths := []string{
		"\\scoop\\apps\\",
		"/scoop/apps/",
	}

	for _, scoopPath := range scoopPaths {
		if strings.Contains(path, scoopPath) {
			return true
		}
	}

	return false
}

// isSourceInstall checks if running from source/development directory
func isSourceInstall(path string) bool {
	// Check for common development paths
	devIndicators := []string{
		"/src/prism/",
		"/prism/bin/",
		"/go/src/",
		"/workspace/",
		"/dev/",
	}

	for _, indicator := range devIndicators {
		if strings.Contains(path, indicator) {
			// Additional check: look for .git directory in parent paths
			dir := filepath.Dir(path)
			for i := 0; i < 5; i++ { // Check up to 5 levels up
				gitPath := filepath.Join(dir, ".git")
				if _, err := os.Stat(gitPath); err == nil {
					return true
				}
				parentDir := filepath.Dir(dir)
				if parentDir == dir {
					break // Reached root
				}
				dir = parentDir
			}
		}
	}

	return false
}

// GetUpdateCommand returns the appropriate update command for the installation method
func GetUpdateCommand(method InstallMethod) string {
	switch method {
	case InstallMethodHomebrew:
		return "brew upgrade prism"
	case InstallMethodScoop:
		return "scoop update prism"
	case InstallMethodSource:
		return "git pull && make build"
	case InstallMethodBinary:
		return "Download latest release from https://github.com/scttfrdmn/prism/releases"
	default:
		return "Visit https://github.com/scttfrdmn/prism/releases for update instructions"
	}
}

// GetUpdateInstructions returns detailed update instructions
func GetUpdateInstructions(method InstallMethod, latestVersion string) string {
	switch method {
	case InstallMethodHomebrew:
		return "Update Prism via Homebrew:\n  $ brew upgrade prism"
	case InstallMethodScoop:
		return "Update Prism via Scoop:\n  $ scoop update prism"
	case InstallMethodSource:
		return "Update Prism from source:\n  $ git pull origin main\n  $ make build"
	case InstallMethodBinary:
		return "Download the latest binary:\n  1. Visit https://github.com/scttfrdmn/prism/releases/tag/" + latestVersion + "\n  2. Download the appropriate binary for your platform\n  3. Replace your current prism binary"
	default:
		return "Visit https://github.com/scttfrdmn/prism/releases for update instructions"
	}
}
