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

	// Inherit environment (includes PRISM_USE_LOCALSTACK, AWS_PROFILE=localstack, etc.)
	// Do not override AWS_PROFILE here: appending would cause the last value to win
	// (shell/execve behavior), overwriting the localstack profile set by the test runner.
	cmd.Env = append(os.Environ(),
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
		output := runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "-y")
		assert.Contains(t, output, "launched successfully", "Launch should succeed")
		assert.Contains(t, output, instanceName, "Output should contain instance name")

		// Register for cleanup
		registry.Register("instance", instanceName)
	})

	// Step 2: List instances
	t.Run("List", func(t *testing.T) {
		output := runCLI(t, "workspace", "list")
		assert.Contains(t, output, instanceName, "List should show the instance")
		assert.Contains(t, output, "running", "Instance should be running")
	})

	// Step 3: Get connection info
	t.Run("Connect", func(t *testing.T) {
		output := runCLI(t, "workspace", "connect", instanceName)
		assert.Contains(t, output, "ssh", "Connect should show SSH command")
		assert.Contains(t, output, instanceName, "Connect should reference instance")
	})

	// Step 4: Stop instance
	t.Run("Stop", func(t *testing.T) {
		output := runCLI(t, "workspace", "stop", instanceName)
		assert.Contains(t, output, "Stopping workspace", "Stop should print progress message")

		// Wait for instance to stop
		time.Sleep(5 * time.Second)

		// Verify stopped state
		listOutput := runCLI(t, "workspace", "list")
		assert.Contains(t, listOutput, "stopped", "Instance should be stopped")
	})

	// Step 5: Start instance
	t.Run("Start", func(t *testing.T) {
		output := runCLI(t, "workspace", "start", instanceName)
		assert.Contains(t, output, "Starting workspace", "Start should print progress message")

		// Wait for instance to start
		time.Sleep(5 * time.Second)

		// Verify running state
		listOutput := runCLI(t, "workspace", "list")
		assert.Contains(t, listOutput, "running", "Instance should be running again")
	})

	// Step 6: Delete instance
	t.Run("Delete", func(t *testing.T) {
		output := runCLI(t, "workspace", "delete", instanceName)
		assert.Contains(t, output, "Deleting workspace", "Delete should print progress message")
	})

	// Cleanup happens automatically via fixtures
}

// TestCLI_List_MultipleInstances tests list command with multiple instances
func TestCLI_List_MultipleInstances(t *testing.T) {
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

	// Launch two instances
	runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instance1Name, "--size", "S", "-y")
	registry.Register("instance", instance1Name)

	runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instance2Name, "--size", "S", "-y")
	registry.Register("instance", instance2Name)

	time.Sleep(3 * time.Second) // Wait for instances to appear

	// Test: List all
	t.Run("ListAll", func(t *testing.T) {
		output := runCLI(t, "workspace", "list")
		assert.Contains(t, output, instance1Name)
		assert.Contains(t, output, instance2Name)
	})
}

// TestCLI_WorkspaceStatus tests workspace status visibility after state changes
func TestCLI_WorkspaceStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("cli-status-test-%d", time.Now().Unix())

	// Launch instance
	runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "-y")
	registry.Register("instance", instanceName)

	// Test: List shows running instance
	t.Run("ListRunning", func(t *testing.T) {
		output := runCLI(t, "workspace", "list")
		assert.Contains(t, output, instanceName)
		assert.Contains(t, output, "running")
	})

	// Stop instance
	runCLI(t, "workspace", "stop", instanceName)
	time.Sleep(5 * time.Second)

	// Test: List shows stopped instance
	t.Run("ListStopped", func(t *testing.T) {
		output := runCLI(t, "workspace", "list")
		assert.Contains(t, output, instanceName)
		assert.Contains(t, output, "stopped")
	})
}

// TestCLI_Delete_Workspace tests delete workflow
func TestCLI_Delete_Workspace(t *testing.T) {
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
	runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "-y")
	registry.Register("instance", instanceName)

	// Test: Delete workspace
	t.Run("Delete", func(t *testing.T) {
		output := runCLI(t, "workspace", "delete", instanceName)
		assert.Contains(t, output, "Deleting workspace", "Delete should print progress message")
	})
}

// TestCLI_Launch_DryRun tests dry-run validation
func TestCLI_Launch_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	instanceName := fmt.Sprintf("cli-dryrun-test-%d", time.Now().Unix())

	// Test: Dry run should validate without creating instance
	output := runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "--dry-run")

	assert.Contains(t, output, "Dry run", "Should indicate dry run mode")
	assert.Contains(t, output, instanceName, "Should show instance name")

	// Verify instance was NOT created
	listOutput := runCLI(t, "workspace", "list")
	assert.NotContains(t, listOutput, instanceName, "Dry run should not create instance")
}

// TestCLI_List_EmptyState tests list with no instances
func TestCLI_List_EmptyState(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	// Note: This test assumes there are some instances
	// It's hard to test truly empty state without cleaning everything
	output := runCLI(t, "workspace", "list")

	// Should not error even if no instances
	assert.NotContains(t, output, "Error")
	assert.NotContains(t, output, "panic")
}

// TestCLI_Stop_NonExistent tests stop on non-existent instance
func TestCLI_Stop_NonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test requires AWS")
	}

	instanceName := "nonexistent-instance-" + fmt.Sprint(time.Now().Unix())

	// Test: Stop non-existent instance should error gracefully
	output := runCLIExpectError(t, "workspace", "stop", instanceName)

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
	runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "-y")
	registry.Register("instance", instanceName)

	// Test: Start already running instance (should handle gracefully)
	output := runCLI(t, "workspace", "start", instanceName)

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
	runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "-y")
	registry.Register("instance", instanceName)
	runCLI(t, "workspace", "delete", instanceName)

	// Test: Delete again should error gracefully
	output := runCLIExpectError(t, "workspace", "delete", instanceName)

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
	output := runCLIExpectError(t, "workspace", "launch", "NonExistentTemplate", instanceName)

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
	output := runCLIExpectError(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "INVALID")

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
	runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "-y")
	registry.Register("instance", instanceName)
	runCLI(t, "workspace", "stop", instanceName)
	time.Sleep(5 * time.Second)

	// Test: Connect to stopped instance should give helpful message
	output := runCLI(t, "workspace", "connect", instanceName)

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
	output := runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "--project", projectName, "-y")
	registry.Register("instance", instanceName)

	assert.Contains(t, output, "launched successfully")
	assert.Contains(t, output, instanceName)

	// Verify instance is associated with project
	instance, err := apiClient.GetInstance(ctx, instanceName)
	require.NoError(t, err)
	assert.Equal(t, projectName, instance.ProjectID)
}

// TestCLI_OutputFormat tests list command output format
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
	runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceName, "--size", "S", "-y")
	registry.Register("instance", instanceName)

	// Test: Default output format (table)
	t.Run("DefaultFormat", func(t *testing.T) {
		output := runCLI(t, "workspace", "list")
		// Should have table-like format with columns
		assert.True(t, strings.Contains(output, "NAME") || strings.Contains(output, "Name"), "Should have column headers")
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
		runCLI(t, "workspace", "launch", "ubuntu-22-04-x86", instanceNames[i], "--size", "S", "-y")
		registry.Register("instance", instanceNames[i])
		time.Sleep(2 * time.Second) // Stagger launches
	}

	// Test: List should show all instances
	t.Run("ListMultiple", func(t *testing.T) {
		output := runCLI(t, "workspace", "list")
		for _, name := range instanceNames {
			assert.Contains(t, output, name, "Should list all instances")
		}
	})

	// Test: Stop all instances
	t.Run("StopMultiple", func(t *testing.T) {
		for _, name := range instanceNames {
			output := runCLI(t, "workspace", "stop", name)
			assert.Contains(t, output, "Stopping workspace", "Stop should print progress message")
		}
	})

	// Test: Delete all instances
	t.Run("DeleteMultiple", func(t *testing.T) {
		for _, name := range instanceNames {
			output := runCLI(t, "workspace", "delete", name)
			assert.Contains(t, output, "Deleting workspace", "Delete should print progress message")
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
		assert.Contains(t, output, "workspace", "Help should list workspace command")
	})

	// Test: Workspace help
	t.Run("WorkspaceHelp", func(t *testing.T) {
		output := runCLI(t, "workspace", "--help")
		assert.Contains(t, output, "launch", "Workspace help should list launch subcommand")
		assert.Contains(t, output, "list", "Workspace help should list list subcommand")
		assert.Contains(t, output, "delete", "Workspace help should list delete subcommand")
	})

	// Test: Launch help
	t.Run("LaunchHelp", func(t *testing.T) {
		output := runCLI(t, "workspace", "launch", "--help")
		assert.Contains(t, output, "template", "Launch help should mention template")
		assert.Contains(t, output, "size", "Launch help should mention size")
	})

	// Test: List help
	t.Run("ListHelp", func(t *testing.T) {
		output := runCLI(t, "workspace", "list", "--help")
		assert.Contains(t, output, "list", "List help should describe list command")
	})
}
