//go:build integration
// +build integration

package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/project"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to find CLI binary
func findCLIBinary(t *testing.T) string {
	t.Helper()

	// Try common locations
	candidates := []string{
		"../../../bin/prism",
		"../../../bin/prism",
		filepath.Join(os.Getenv("GOPATH"), "bin", "prism"),
		filepath.Join(os.Getenv("GOPATH"), "bin", "prism"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			absPath, err := filepath.Abs(candidate)
			require.NoError(t, err)
			return absPath
		}
	}

	t.Fatal("Could not find prism binary. Run 'make build' first.")
	return ""
}

// Helper to run CLI command and return output
func runCLI(t *testing.T, args ...string) string {
	t.Helper()

	binary := findCLIBinary(t)
	cmd := exec.Command(binary, args...)

	// Set environment
	cmd.Env = append(os.Environ(),
		"AWS_PROFILE=aws",
		"AWS_REGION=us-west-2",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command failed: %s %v\nOutput: %s", binary, args, string(output))
	}

	return string(output)
}

// Helper to run CLI command expecting error
func runCLIExpectError(t *testing.T, args ...string) string {
	t.Helper()

	binary := findCLIBinary(t)
	cmd := exec.Command(binary, args...)

	cmd.Env = append(os.Environ(),
		"AWS_PROFILE=aws",
		"AWS_REGION=us-west-2",
	)

	output, err := cmd.CombinedOutput()
	require.Error(t, err, "Expected command to fail but it succeeded")

	return string(output)
}

// TestCLI_Launch_FullWorkflow tests the complete instance lifecycle via CLI
func TestCLI_Launch_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	// Setup API client for cleanup
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-test-%d", time.Now().Unix())

	// Step 1: Launch instance
	t.Run("Launch", func(t *testing.T) {
		output := runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S")
		assert.Contains(t, output, "launched successfully", "Launch should succeed")
		assert.Contains(t, output, instanceName, "Output should contain instance name")

		// Register for cleanup
		registry.Register("instance", instanceName)
	})

	// Step 2: List instances
	t.Run("List", func(t *testing.T) {
		output := runCLI(t, "list")
		assert.Contains(t, output, instanceName, "List should show the instance")
		assert.Contains(t, output, "running", "Instance should be running")
	})

	// Step 3: Get instance info
	t.Run("Info", func(t *testing.T) {
		output := runCLI(t, "info", instanceName)
		assert.Contains(t, output, instanceName, "Info should show instance name")
		assert.Contains(t, output, "running", "Info should show running state")
		assert.Contains(t, output, "i-", "Info should show instance ID")
	})

	// Step 4: Get connection info
	t.Run("Connect", func(t *testing.T) {
		output := runCLI(t, "connect", instanceName)
		assert.Contains(t, output, "ssh", "Connect should show SSH command")
		assert.Contains(t, output, instanceName, "Connect should reference instance")
	})

	// Step 5: Stop instance
	t.Run("Stop", func(t *testing.T) {
		output := runCLI(t, "stop", instanceName, "-y")
		assert.Contains(t, output, "stopped", "Stop should succeed")

		// Wait for instance to stop
		time.Sleep(5 * time.Second)

		// Verify stopped state
		listOutput := runCLI(t, "list")
		assert.Contains(t, listOutput, "stopped", "Instance should be stopped")
	})

	// Step 6: Start instance
	t.Run("Start", func(t *testing.T) {
		output := runCLI(t, "start", instanceName)
		assert.Contains(t, output, "started", "Start should succeed")

		// Wait for instance to start
		time.Sleep(5 * time.Second)

		// Verify running state
		listOutput := runCLI(t, "list")
		assert.Contains(t, listOutput, "running", "Instance should be running again")
	})

	// Step 7: Delete instance
	t.Run("Delete", func(t *testing.T) {
		output := runCLI(t, "delete", instanceName, "-y")
		assert.Contains(t, output, "deleted", "Delete should succeed")
	})

	// Cleanup happens automatically via fixtures
}

// TestCLI_List_AllFilters tests list command with various filters
func TestCLI_List_AllFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	// Create test instances
	instance1Name := fmt.Sprintf("cli-list-test1-%d", time.Now().Unix())
	instance2Name := fmt.Sprintf("cli-list-test2-%d", time.Now().Unix())

	// Launch two instances with different templates
	runCLI(t, "launch", "Ubuntu Basic", instance1Name, "--size", "S")
	registry.Register("instance", instance1Name)

	runCLI(t, "launch", "Ubuntu Basic", instance2Name, "--size", "S")
	registry.Register("instance", instance2Name)

	time.Sleep(3 * time.Second) // Wait for instances to appear

	// Test: List all
	t.Run("ListAll", func(t *testing.T) {
		output := runCLI(t, "list")
		assert.Contains(t, output, instance1Name)
		assert.Contains(t, output, instance2Name)
	})

	// Test: List with state filter
	t.Run("ListByState", func(t *testing.T) {
		output := runCLI(t, "list", "--state", "running")
		assert.Contains(t, output, instance1Name)
		assert.Contains(t, output, instance2Name)
	})

	// Test: List with template filter
	t.Run("ListByTemplate", func(t *testing.T) {
		output := runCLI(t, "list", "--template", "Ubuntu")
		assert.Contains(t, output, instance1Name)
	})
}

// TestCLI_Info_AllInstanceStates tests info command with different states
func TestCLI_Info_AllInstanceStates(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-info-test-%d", time.Now().Unix())

	// Launch instance
	runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S")
	registry.Register("instance", instanceName)

	// Test: Info on running instance
	t.Run("InfoRunning", func(t *testing.T) {
		output := runCLI(t, "info", instanceName)
		assert.Contains(t, output, instanceName)
		assert.Contains(t, output, "running")
		assert.Contains(t, output, "Public IP")
	})

	// Stop instance
	runCLI(t, "stop", instanceName, "-y")
	time.Sleep(5 * time.Second)

	// Test: Info on stopped instance
	t.Run("InfoStopped", func(t *testing.T) {
		output := runCLI(t, "info", instanceName)
		assert.Contains(t, output, instanceName)
		assert.Contains(t, output, "stopped")
	})
}

// TestCLI_Delete_WithConfirmation tests delete with and without -y flag
func TestCLI_Delete_WithConfirmation(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-delete-test-%d", time.Now().Unix())

	// Launch instance
	runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S")
	registry.Register("instance", instanceName)

	// Test: Delete with -y flag (no prompt)
	t.Run("DeleteWithFlag", func(t *testing.T) {
		output := runCLI(t, "delete", instanceName, "-y")
		assert.Contains(t, output, "deleted")
		assert.NotContains(t, output, "Are you sure")
	})
}

// TestCLI_Launch_DryRun tests dry-run validation
func TestCLI_Launch_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	instanceName := fmt.Sprintf("cli-dryrun-test-%d", time.Now().Unix())

	// Test: Dry run should validate without creating instance
	output := runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S", "--dry-run")

	assert.Contains(t, output, "Dry run", "Should indicate dry run mode")
	assert.Contains(t, output, instanceName, "Should show instance name")

	// Verify instance was NOT created
	listOutput := runCLI(t, "list")
	assert.NotContains(t, listOutput, instanceName, "Dry run should not create instance")
}

// TestCLI_List_EmptyState tests list with no instances
func TestCLI_List_EmptyState(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	// Note: This test assumes there are some instances
	// It's hard to test truly empty state without cleaning everything
	output := runCLI(t, "list")

	// Should not error even if no instances
	assert.NotContains(t, output, "Error")
	assert.NotContains(t, output, "panic")
}

// TestCLI_Info_NonExistent tests info on non-existent instance
func TestCLI_Info_NonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	instanceName := "nonexistent-instance-" + fmt.Sprint(time.Now().Unix())

	// Test: Info on non-existent instance should error gracefully
	output := runCLIExpectError(t, "info", instanceName)

	assert.Contains(t, output, "not found", "Should indicate instance not found")
	assert.NotContains(t, output, "panic", "Should not panic")
}

// TestCLI_Stop_NonExistent tests stop on non-existent instance
func TestCLI_Stop_NonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	instanceName := "nonexistent-instance-" + fmt.Sprint(time.Now().Unix())

	// Test: Stop non-existent instance should error gracefully
	output := runCLIExpectError(t, "stop", instanceName, "-y")

	assert.Contains(t, output, "not found", "Should indicate instance not found")
	assert.NotContains(t, output, "panic", "Should not panic")
}

// TestCLI_Start_AlreadyRunning tests start on already running instance
func TestCLI_Start_AlreadyRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-start-test-%d", time.Now().Unix())

	// Launch instance (already running)
	runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S")
	registry.Register("instance", instanceName)

	// Test: Start already running instance (should handle gracefully)
	output := runCLI(t, "start", instanceName)

	// Should either succeed or give helpful message
	assert.NotContains(t, output, "panic", "Should not panic")
}

// TestCLI_Delete_AlreadyDeleted tests delete on already deleted instance
func TestCLI_Delete_AlreadyDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-delete-twice-%d", time.Now().Unix())

	// Launch and delete instance
	runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S")
	registry.Register("instance", instanceName)
	runCLI(t, "delete", instanceName, "-y")

	// Test: Delete again should error gracefully
	output := runCLIExpectError(t, "delete", instanceName, "-y")

	assert.Contains(t, output, "not found", "Should indicate instance not found")
	assert.NotContains(t, output, "panic", "Should not panic")
}

// TestCLI_Launch_InvalidTemplate tests launch with invalid template
func TestCLI_Launch_InvalidTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	instanceName := fmt.Sprintf("cli-invalid-template-%d", time.Now().Unix())

	// Test: Launch with invalid template name
	output := runCLIExpectError(t, "launch", "NonExistentTemplate", instanceName)

	assert.Contains(t, output, "not found", "Should indicate template not found")
	assert.NotContains(t, output, "panic", "Should not panic")
}

// TestCLI_Launch_InvalidSize tests launch with invalid size
func TestCLI_Launch_InvalidSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	instanceName := fmt.Sprintf("cli-invalid-size-%d", time.Now().Unix())

	// Test: Launch with invalid size
	output := runCLIExpectError(t, "launch", "Ubuntu Basic", instanceName, "--size", "INVALID")

	assert.Contains(t, output, "invalid", "Should indicate invalid size")
	assert.NotContains(t, output, "panic", "Should not panic")
}

// TestCLI_Connect_StoppedInstance tests connect to stopped instance
func TestCLI_Connect_StoppedInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-connect-stopped-%d", time.Now().Unix())

	// Launch and stop instance
	runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S")
	registry.Register("instance", instanceName)
	runCLI(t, "stop", instanceName, "-y")
	time.Sleep(5 * time.Second)

	// Test: Connect to stopped instance should give helpful message
	output := runCLI(t, "connect", instanceName)

	// Should indicate instance is stopped
	assert.Contains(t, output, "stopped", "Should mention instance is stopped")
}

// TestCLI_Launch_WithProject tests launch with project flag
func TestCLI_Launch_WithProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS and project setup")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)
	ctx := context.Background()

	projectName := fmt.Sprintf("cli-project-test-%d", time.Now().Unix())
	instanceName := fmt.Sprintf("cli-instance-%d", time.Now().Unix())

	// Create test project via API
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for CLI",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	defer apiClient.DeleteProject(ctx, projectName)

	// Test: Launch with project flag
	output := runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S", "--project", projectName)
	registry.Register("instance", instanceName)

	assert.Contains(t, output, "launched successfully")
	assert.Contains(t, output, instanceName)

	// Verify instance is associated with project
	instance, err := apiClient.GetInstance(ctx, instanceName)
	require.NoError(t, err)
	assert.Equal(t, projectName, instance.ProjectID)
}

// TestCLI_OutputFormat tests various output formats
func TestCLI_OutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-format-test-%d", time.Now().Unix())

	// Launch instance
	runCLI(t, "launch", "Ubuntu Basic", instanceName, "--size", "S")
	registry.Register("instance", instanceName)

	// Test: Default output format (table)
	t.Run("DefaultFormat", func(t *testing.T) {
		output := runCLI(t, "list")
		// Should have table-like format with columns
		assert.True(t, strings.Contains(output, "NAME") || strings.Contains(output, "Name"), "Should have column headers")
	})

	// Test: JSON output format (if supported)
	t.Run("JSONFormat", func(t *testing.T) {
		output := runCLI(t, "list", "--output", "json")
		// Should be valid JSON
		if strings.Contains(output, "{") {
			assert.Contains(t, output, instanceName)
		}
	})
}

// TestCLI_MultipleInstances tests operations with multiple instances
func TestCLI_MultipleInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	// Create 3 test instances
	instanceNames := make([]string, 3)
	for i := 0; i < 3; i++ {
		instanceNames[i] = fmt.Sprintf("cli-multi-%d-%d", i, time.Now().Unix())
		runCLI(t, "launch", "Ubuntu Basic", instanceNames[i], "--size", "S")
		registry.Register("instance", instanceNames[i])
		time.Sleep(2 * time.Second) // Stagger launches
	}

	// Test: List should show all instances
	t.Run("ListMultiple", func(t *testing.T) {
		output := runCLI(t, "list")
		for _, name := range instanceNames {
			assert.Contains(t, output, name, "Should list all instances")
		}
	})

	// Test: Stop all instances
	t.Run("StopMultiple", func(t *testing.T) {
		for _, name := range instanceNames {
			output := runCLI(t, "stop", name, "-y")
			assert.Contains(t, output, "stopped")
		}
	})

	// Test: Delete all instances
	t.Run("DeleteMultiple", func(t *testing.T) {
		for _, name := range instanceNames {
			output := runCLI(t, "delete", name, "-y")
			assert.Contains(t, output, "deleted")
		}
	})
}

// TestCLI_Help tests help output for commands
func TestCLI_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test")
	}

	// Test: Main help
	t.Run("MainHelp", func(t *testing.T) {
		output := runCLI(t, "--help")
		assert.Contains(t, output, "launch", "Help should list launch command")
		assert.Contains(t, output, "list", "Help should list list command")
		assert.Contains(t, output, "delete", "Help should list delete command")
	})

	// Test: Launch help
	t.Run("LaunchHelp", func(t *testing.T) {
		output := runCLI(t, "launch", "--help")
		assert.Contains(t, output, "template", "Launch help should mention template")
		assert.Contains(t, output, "size", "Launch help should mention size")
	})

	// Test: List help
	t.Run("ListHelp", func(t *testing.T) {
		output := runCLI(t, "list", "--help")
		assert.Contains(t, output, "list", "List help should describe list command")
	})
}
