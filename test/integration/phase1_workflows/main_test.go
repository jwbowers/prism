package phase1_workflows

import (
	"os"
	"testing"

	"github.com/scttfrdmn/prism/test/integration"
)

// TestMain is the entry point for the test suite
// It ensures aggressive cleanup of any leaked instances at the end
func TestMain(m *testing.M) {
	// Run tests
	exitCode := m.Run()

	// CRITICAL: Aggressive cleanup to ensure NO instances left running
	// This is the final safety net that catches any leaked instances
	integration.AggressiveCleanup(m)

	// Exit with test result code
	os.Exit(exitCode)
}
