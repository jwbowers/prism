package phase1_workflows

import (
	"os"
	"testing"
)

// TestMain is the entry point for the test suite
func TestMain(m *testing.M) {
	// Run tests
	exitCode := m.Run()

	// Note: Cleanup is now handled by individual tests using t.Cleanup()
	// and the fixture registry pattern

	// Exit with test result code
	os.Exit(exitCode)
}
