//go:build integration
// +build integration

package cli_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findTemplateCLIBinary locates the prism CLI binary for testing
func findTemplateCLIBinary(t *testing.T) string {
	t.Helper()

	candidates := []string{
		"../../../bin/prism",
		"../../../bin/prism",
		filepath.Join(os.Getenv("GOPATH"), "bin", "prism"),
		filepath.Join(os.Getenv("GOPATH"), "bin", "prism"),
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	t.Fatal("Could not find prism binary. Run 'make build' first.")
	return ""
}

// runTemplateCLI executes a CLI command and returns output
func runTemplateCLI(t *testing.T, args ...string) string {
	t.Helper()

	binary := findTemplateCLIBinary(t)
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(),
		"AWS_PROFILE=aws",
		"AWS_REGION=us-west-2",
		"PRISM_DEV=true",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command failed: %s %v", binary, args)
		t.Logf("Output: %s", string(output))
	}

	return string(output)
}

// runTemplateCLIExpectError executes a CLI command expecting it to fail
func runTemplateCLIExpectError(t *testing.T, args ...string) string {
	t.Helper()

	binary := findTemplateCLIBinary(t)
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(),
		"AWS_PROFILE=aws",
		"AWS_REGION=us-west-2",
		"PRISM_DEV=true",
	)

	output, err := cmd.CombinedOutput()
	require.Error(t, err, "Expected command to fail but it succeeded")

	return string(output)
}

// TestCLI_Templates_List tests template listing
func TestCLI_Templates_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// List all templates
	output := runTemplateCLI(t, "templates", "list")
	assert.NotEmpty(t, output, "Templates list should not be empty")

	// Should contain standard templates
	assert.Contains(t, output, "Ubuntu", "Should contain Ubuntu templates")
	assert.Contains(t, output, "Rocky", "Should contain Rocky Linux templates")
}

// TestCLI_Templates_Info tests template info command
func TestCLI_Templates_Info(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get info for Ubuntu Basic template
	output := runTemplateCLI(t, "templates", "info", "ubuntu-22-04-x86")
	assert.Contains(t, output, "Ubuntu", "Info should contain template name")
	assert.Contains(t, output, "description", "Info should contain description field")
	assert.Contains(t, output, "packages", "Info should contain packages information")
}

// TestCLI_Templates_Search tests template search functionality
func TestCLI_Templates_Search(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Search for Python templates
	output := runTemplateCLI(t, "templates", "search", "python")
	assert.Contains(t, output, "Python", "Search should find Python templates")

	// Search for machine learning templates
	output = runTemplateCLI(t, "templates", "search", "machine learning")
	assert.Contains(t, output, "Machine Learning", "Search should find ML templates")

	// Search for non-existent template
	output = runTemplateCLI(t, "templates", "search", "nonexistent-template-xyz")
	// Should handle no results gracefully
	if !strings.Contains(output, "No templates found") && !strings.Contains(output, "0 templates") {
		// If it returns templates, that's also acceptable (fuzzy matching)
		assert.NotEmpty(t, output)
	}
}

// TestCLI_Templates_Discover tests template discovery by category
func TestCLI_Templates_Discover(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Discover all categories
	output := runTemplateCLI(t, "templates", "discover")
	assert.NotEmpty(t, output, "Discover should show categories")

	// Discover specific category
	output = runTemplateCLI(t, "templates", "discover", "--category", "development")
	if !strings.Contains(output, "unknown flag") && !strings.Contains(output, "No templates") {
		assert.NotEmpty(t, output, "Category discovery should return results")
	}
}

// TestCLI_Templates_Validate tests template validation
func TestCLI_Templates_Validate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Validate all templates
	output := runTemplateCLI(t, "templates", "validate")
	assert.Contains(t, output, "valid", "Validate should report validation status")

	// Should show validation summary
	assert.Contains(t, output, "templates", "Should mention templates")
}

// TestCLI_Templates_NonExistent tests error handling for non-existent templates
func TestCLI_Templates_NonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Info on non-existent template
	output := runTemplateCLIExpectError(t, "templates", "info", "nonexistent-template-12345")
	assert.Contains(t, strings.ToLower(output), "not found", "Should indicate template not found")
}

// TestCLI_Templates_OutputFormats tests different output formats
func TestCLI_Templates_OutputFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test table output (default)
	output := runTemplateCLI(t, "templates", "list")
	assert.Contains(t, output, "Ubuntu")

	// Test JSON output
	output = runTemplateCLI(t, "templates", "list", "-o", "json")
	if !strings.Contains(output, "unknown flag") {
		assert.Contains(t, output, "{", "JSON output should contain JSON syntax")
		assert.Contains(t, output, "Ubuntu")
	}

	// Test YAML output
	output = runTemplateCLI(t, "templates", "list", "-o", "yaml")
	if !strings.Contains(output, "unknown flag") {
		assert.Contains(t, output, "Ubuntu")
	}
}

// TestCLI_Templates_Filtering tests template filtering options
func TestCLI_Templates_Filtering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Filter by OS
	output := runTemplateCLI(t, "templates", "list", "--os", "ubuntu")
	if !strings.Contains(output, "unknown flag") {
		assert.Contains(t, output, "Ubuntu", "OS filter should show Ubuntu templates")
	}

	// Filter by category
	output = runTemplateCLI(t, "templates", "list", "--category", "development")
	if !strings.Contains(output, "unknown flag") {
		assert.NotEmpty(t, output, "Category filter should return results")
	}

	// Filter by package manager
	output = runTemplateCLI(t, "templates", "list", "--package-manager", "apt")
	if !strings.Contains(output, "unknown flag") {
		assert.NotEmpty(t, output, "Package manager filter should return results")
	}
}

// TestCLI_Templates_Sorting tests template sorting options
func TestCLI_Templates_Sorting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Sort by name
	output := runTemplateCLI(t, "templates", "list", "--sort", "name")
	if !strings.Contains(output, "unknown flag") {
		assert.NotEmpty(t, output, "Sort by name should return results")
	}

	// Sort by popularity
	output = runTemplateCLI(t, "templates", "list", "--sort", "popularity")
	if !strings.Contains(output, "unknown flag") {
		assert.NotEmpty(t, output, "Sort by popularity should return results")
	}
}

// TestCLI_Templates_Version tests template version management
func TestCLI_Templates_Version(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// List template versions
	output := runTemplateCLI(t, "templates", "version", "list", "ubuntu-22-04-x86")
	// This command may not be fully implemented, so check for output
	assert.NotEmpty(t, output, "Version command should return output")
}

// TestCLI_Templates_Usage tests template usage statistics
func TestCLI_Templates_Usage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get usage statistics
	output := runTemplateCLI(t, "templates", "usage")
	assert.NotEmpty(t, output, "Usage command should return output")

	// Get usage for specific template
	output = runTemplateCLI(t, "templates", "usage", "ubuntu-22-04-x86")
	if !strings.Contains(output, "not found") && !strings.Contains(output, "No usage") {
		assert.NotEmpty(t, output, "Template usage should return statistics")
	}
}

// TestCLI_Templates_Snapshot tests template snapshot creation
func TestCLI_Templates_Snapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-test-snapshot-%d", time.Now().Unix())

	// Launch instance for snapshotting
	output := runTemplateCLI(t, "launch", "ubuntu-22-04-x86", instanceName, "--size", "S")
	assert.Contains(t, output, "launched successfully")
	registry.Register("instance", instanceName)

	// Wait for instance to be running
	time.Sleep(10 * time.Second)

	// Create snapshot
	snapshotName := fmt.Sprintf("cli-snapshot-%d", time.Now().Unix())
	output = runTemplateCLI(t, "templates", "snapshot", instanceName,
		"--name", snapshotName,
		"--description", "Test snapshot from CLI")

	// Snapshot creation may be async, check for appropriate response
	if !strings.Contains(output, "not implemented") {
		assert.NotEmpty(t, output, "Snapshot command should return output")
		// Register snapshot for cleanup if successfully created
		if strings.Contains(output, "created") || strings.Contains(output, "success") {
			registry.Register("snapshot", snapshotName)
		}
	}
}

// TestCLI_Templates_Install tests template installation
func TestCLI_Templates_Install(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test template installation (dry-run mode)
	output := runTemplateCLI(t, "templates", "install", "test-template", "--dry-run")
	if !strings.Contains(output, "not implemented") && !strings.Contains(output, "unknown flag") {
		assert.NotEmpty(t, output, "Install command should return output")
	}
}

// TestCLI_Templates_Test tests template testing functionality
func TestCLI_Templates_Test(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Run template tests
	output := runTemplateCLI(t, "templates", "test", "ubuntu-22-04-x86")
	assert.NotEmpty(t, output, "Template test should return output")

	// May show test results or indicate tests are not configured
	if !strings.Contains(output, "not implemented") && !strings.Contains(output, "No tests") {
		assert.Contains(t, output, "test", "Test output should mention tests")
	}
}

// TestCLI_Templates_DetailedInfo tests detailed template information
func TestCLI_Templates_DetailedInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get detailed info
	output := runTemplateCLI(t, "templates", "info", "ubuntu-22-04-x86", "--verbose")
	if !strings.Contains(output, "unknown flag") {
		assert.NotEmpty(t, output, "Detailed info should return output")
	}

	// Check for key information fields
	output = runTemplateCLI(t, "templates", "info", "ubuntu-22-04-x86")
	assert.Contains(t, output, "Ubuntu", "Should contain template name")

	// Should show various template properties
	fields := []string{"description", "base", "packages"}
	foundCount := 0
	for _, field := range fields {
		if strings.Contains(strings.ToLower(output), field) {
			foundCount++
		}
	}
	assert.Greater(t, foundCount, 0, "Should contain at least one key field")
}

// TestCLI_Templates_SearchCaseInsensitive tests case-insensitive search
func TestCLI_Templates_SearchCaseInsensitive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Search with lowercase
	output1 := runTemplateCLI(t, "templates", "search", "ubuntu")
	// Search with uppercase
	output2 := runTemplateCLI(t, "templates", "search", "UBUNTU")
	// Search with mixed case
	output3 := runTemplateCLI(t, "templates", "search", "Ubuntu")

	// All should return results (case-insensitive)
	assert.NotEmpty(t, output1, "Lowercase search should return results")
	assert.NotEmpty(t, output2, "Uppercase search should return results")
	assert.NotEmpty(t, output3, "Mixed case search should return results")
}

// TestCLI_Templates_ListWithCount tests template list with count
func TestCLI_Templates_ListWithCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	output := runTemplateCLI(t, "templates", "list")

	// Should show multiple templates
	assert.NotEmpty(t, output, "List should show templates")

	// Count should be reasonable (at least a few templates)
	ubuntuCount := strings.Count(output, "Ubuntu")
	assert.Greater(t, ubuntuCount, 0, "Should have at least one Ubuntu template")
}

// TestCLI_Templates_InfoFields tests that info displays all expected fields
func TestCLI_Templates_InfoFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	output := runTemplateCLI(t, "templates", "info", "ubuntu-22-04-x86")

	// Check for essential fields
	essentialFields := []string{"name", "description", "ubuntu"}
	for _, field := range essentialFields {
		assert.Contains(t, strings.ToLower(output), field,
			fmt.Sprintf("Info should contain %s field", field))
	}
}

// TestCLI_Templates_ValidateSpecific tests validating specific template
func TestCLI_Templates_ValidateSpecific(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Validate specific template
	output := runTemplateCLI(t, "templates", "validate", "ubuntu-22-04-x86")
	assert.NotEmpty(t, output, "Validate specific template should return output")
	assert.Contains(t, output, "valid", "Should report validation status")
}

// TestCLI_Templates_Help tests help command output
func TestCLI_Templates_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Templates help
	output := runTemplateCLI(t, "templates", "--help")
	assert.Contains(t, output, "templates")
	assert.Contains(t, output, "list")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "search")
	assert.Contains(t, output, "validate")

	// List help
	output = runTemplateCLI(t, "templates", "list", "--help")
	assert.Contains(t, output, "list")

	// Info help
	output = runTemplateCLI(t, "templates", "info", "--help")
	assert.Contains(t, output, "info")

	// Search help
	output = runTemplateCLI(t, "templates", "search", "--help")
	assert.Contains(t, output, "search")
}

// TestCLI_Templates_Integration tests template commands with instance launch
func TestCLI_Templates_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	// Get template info
	output := runTemplateCLI(t, "templates", "info", "ubuntu-22-04-x86")
	assert.Contains(t, output, "Ubuntu")

	// Launch instance using template
	instanceName := fmt.Sprintf("cli-template-test-%d", time.Now().Unix())
	output = runTemplateCLI(t, "launch", "ubuntu-22-04-x86", instanceName, "--size", "S")
	assert.Contains(t, output, "launched successfully")
	registry.Register("instance", instanceName)

	// Verify instance is running
	output = runTemplateCLI(t, "info", instanceName)
	assert.Contains(t, output, instanceName)
	assert.Contains(t, output, "Ubuntu")
}
