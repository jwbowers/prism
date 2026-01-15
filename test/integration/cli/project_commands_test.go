//go:build integration
// +build integration

package cli_test

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

// findCLIBinary locates the prism CLI binary for testing
func findProjectCLIBinary(t *testing.T) string {
	t.Helper()

	candidates := []string{
		"../../../bin/prism",
		"../../../bin/cws",
		filepath.Join(os.Getenv("GOPATH"), "bin", "prism"),
		filepath.Join(os.Getenv("GOPATH"), "bin", "cws"),
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

	t.Fatal("Could not find prism/cws binary. Run 'make build' first.")
	return ""
}

// runProjectCLI executes a CLI command and returns output
func runProjectCLI(t *testing.T, args ...string) string {
	t.Helper()

	binary := findProjectCLIBinary(t)
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

// runProjectCLIExpectError executes a CLI command expecting it to fail
func runProjectCLIExpectError(t *testing.T, args ...string) string {
	t.Helper()

	binary := findProjectCLIBinary(t)
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

// TestCLI_Project_CreateListDelete tests basic project lifecycle
func TestCLI_Project_CreateListDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectName := fmt.Sprintf("cli-test-project-%d", time.Now().Unix())

	// Create project
	output := runProjectCLI(t, "project", "create", projectName,
		"--description", "Test project for CLI",
		"--owner", "test-owner")
	assert.Contains(t, output, "created successfully", "Project creation output should indicate success")

	// List projects
	output = runProjectCLI(t, "project", "list")
	assert.Contains(t, output, projectName, "Project list should contain created project")

	// Get project info
	output = runProjectCLI(t, "project", "info", projectName)
	assert.Contains(t, output, projectName, "Project info should contain project name")
	assert.Contains(t, output, "Test project for CLI", "Project info should contain description")

	// Delete project
	output = runProjectCLI(t, "project", "delete", projectName, "-y")
	assert.Contains(t, output, "deleted", "Project deletion output should indicate success")

	// Verify deletion
	output = runProjectCLI(t, "project", "list")
	assert.NotContains(t, output, projectName, "Project list should not contain deleted project")
}

// TestCLI_Project_Info tests project info command
func TestCLI_Project_Info(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-info-%d", time.Now().Unix())

	// Create project via API
	proj, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for info command",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// Get info
	output := runProjectCLI(t, "project", "info", projectName)
	assert.Contains(t, output, projectName)
	assert.Contains(t, output, "Test project for info command")
	assert.Contains(t, output, proj.ID)
}

// TestCLI_Project_Budget tests budget management commands
func TestCLI_Project_Budget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-budget-%d", time.Now().Unix())

	// Create project
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for budget commands",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// Set budget
	output := runProjectCLI(t, "project", "budget", "set", projectName,
		"--limit", "1000",
		"--period", "monthly")
	assert.Contains(t, output, "budget", "Budget set output should mention budget")

	// Show budget
	output = runProjectCLI(t, "project", "budget", "show", projectName)
	assert.Contains(t, output, "1000", "Budget show should display limit")
	assert.Contains(t, output, "monthly", "Budget show should display period")

	// Get budget alerts
	output = runProjectCLI(t, "project", "budget", "alerts", projectName)
	// Should show no alerts initially (or handle appropriately)
	assert.NotEmpty(t, output, "Budget alerts should return output")
}

// TestCLI_Project_Members tests member management commands
func TestCLI_Project_Members(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-members-%d", time.Now().Unix())

	// Create project
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for member commands",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// List members (initially just owner)
	output := runProjectCLI(t, "project", "members", "list", projectName)
	assert.Contains(t, output, "test-owner", "Members list should contain owner")

	// Add member
	output = runProjectCLI(t, "project", "members", "add", projectName,
		"--email", "newmember@example.com",
		"--role", "member")
	assert.Contains(t, output, "added", "Member add output should indicate success")

	// List members again
	output = runProjectCLI(t, "project", "members", "list", projectName)
	assert.Contains(t, output, "newmember@example.com", "Members list should contain new member")

	// Remove member
	output = runProjectCLI(t, "project", "members", "remove", projectName,
		"--email", "newmember@example.com", "-y")
	assert.Contains(t, output, "removed", "Member remove output should indicate success")

	// Verify removal
	output = runProjectCLI(t, "project", "members", "list", projectName)
	assert.NotContains(t, output, "newmember@example.com", "Members list should not contain removed member")
}

// TestCLI_Project_Invitations tests invitation management commands
func TestCLI_Project_Invitations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-invites-%d", time.Now().Unix())

	// Create project
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for invitation commands",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// Invite user
	output := runProjectCLI(t, "project", "invite", projectName,
		"--email", "invitee@example.com",
		"--role", "member",
		"--message", "Welcome to the project")
	assert.Contains(t, output, "invitation sent", "Invite output should indicate invitation sent")

	// List invitations
	output = runProjectCLI(t, "project", "invitations", "list", projectName)
	assert.Contains(t, output, "invitee@example.com", "Invitations list should contain invitee")

	// Cancel invitation
	output = runProjectCLI(t, "project", "invitations", "cancel", projectName,
		"--email", "invitee@example.com", "-y")
	assert.Contains(t, output, "cancelled", "Cancel invitation output should indicate success")

	// Verify cancellation
	output = runProjectCLI(t, "project", "invitations", "list", projectName)
	assert.NotContains(t, output, "invitee@example.com", "Invitations list should not contain cancelled invitation")
}

// TestCLI_Project_Instances tests instance management within projects
func TestCLI_Project_Instances(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-instances-%d", time.Now().Unix())
	instanceName := fmt.Sprintf("cli-test-inst-%d", time.Now().Unix())

	// Create project
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for instance commands",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// Launch instance in project
	output := runProjectCLI(t, "launch", "Ubuntu Basic", instanceName,
		"--project", projectName,
		"--size", "S")
	assert.Contains(t, output, "launched successfully")
	registry.Register("instance", instanceName)

	// List instances in project
	output = runProjectCLI(t, "project", "instances", projectName)
	assert.Contains(t, output, instanceName, "Project instances should contain launched instance")

	// Verify instance shows project association
	output = runProjectCLI(t, "info", instanceName)
	assert.Contains(t, output, projectName, "Instance info should show project association")
}

// TestCLI_Project_Templates tests template listing within projects
func TestCLI_Project_Templates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-templates-%d", time.Now().Unix())

	// Create project
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for template commands",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// List templates (should show all available templates for project)
	output := runProjectCLI(t, "project", "templates", projectName)
	assert.NotEmpty(t, output, "Templates list should not be empty")
	// Should contain standard templates
	assert.Contains(t, output, "Ubuntu", "Templates should include Ubuntu templates")
}

// TestCLI_Project_NonExistent tests error handling for non-existent projects
func TestCLI_Project_NonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	nonExistentProject := "nonexistent-project-12345"

	// Info on non-existent project
	output := runProjectCLIExpectError(t, "project", "info", nonExistentProject)
	assert.Contains(t, strings.ToLower(output), "not found", "Should indicate project not found")

	// Delete non-existent project
	output = runProjectCLIExpectError(t, "project", "delete", nonExistentProject, "-y")
	assert.Contains(t, strings.ToLower(output), "not found", "Should indicate project not found")

	// Show budget for non-existent project
	output = runProjectCLIExpectError(t, "project", "budget", "show", nonExistentProject)
	assert.Contains(t, strings.ToLower(output), "not found", "Should indicate project not found")
}

// TestCLI_Project_EmptyList tests project list when no projects exist
func TestCLI_Project_EmptyList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Note: This test assumes a clean state, which may not always be true
	// In a real CI environment, there might be other projects
	output := runProjectCLI(t, "project", "list")
	// Should handle empty list gracefully
	assert.NotEmpty(t, output, "Project list should return output even if empty")
}

// TestCLI_Project_DuplicateName tests creating project with duplicate name
func TestCLI_Project_DuplicateName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-dup-%d", time.Now().Unix())

	// Create first project
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "First project",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// Try to create duplicate
	output := runProjectCLIExpectError(t, "project", "create", projectName,
		"--description", "Duplicate project",
		"--owner", "test-owner")
	assert.Contains(t, strings.ToLower(output), "already exists", "Should indicate project already exists")
}

// TestCLI_Project_OutputFormats tests different output formats
func TestCLI_Project_OutputFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-format-%d", time.Now().Unix())

	// Create project
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for output formats",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// Test table output (default)
	output := runProjectCLI(t, "project", "list")
	assert.Contains(t, output, projectName)

	// Test JSON output
	output = runProjectCLI(t, "project", "list", "-o", "json")
	assert.Contains(t, output, projectName)
	assert.Contains(t, output, "{", "JSON output should contain JSON syntax")

	// Test YAML output (if supported)
	output = runProjectCLI(t, "project", "list", "-o", "yaml")
	if !strings.Contains(output, "unknown flag") {
		assert.Contains(t, output, projectName)
	}
}

// TestCLI_Project_BudgetAlerts tests budget alert commands
func TestCLI_Project_BudgetAlerts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	projectName := fmt.Sprintf("cli-test-alerts-%d", time.Now().Unix())

	// Create project
	_, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
		Name:        projectName,
		Description: "Test project for budget alerts",
		Owner:       "test-owner",
	})
	require.NoError(t, err)
	registry.Register("project", projectName)

	// Set budget
	runProjectCLI(t, "project", "budget", "set", projectName,
		"--limit", "500",
		"--period", "monthly")

	// Get alerts
	output := runProjectCLI(t, "project", "budget", "alerts", projectName)
	assert.NotEmpty(t, output, "Budget alerts should return output")

	// Test alert threshold
	output = runProjectCLI(t, "project", "budget", "set-alert", projectName,
		"--threshold", "80",
		"--email", "admin@example.com")
	if !strings.Contains(output, "unknown command") {
		assert.Contains(t, output, "alert", "Alert set output should mention alert")
	}
}

// TestCLI_Project_Help tests help command output
func TestCLI_Project_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Project help
	output := runProjectCLI(t, "project", "--help")
	assert.Contains(t, output, "Manage projects")
	assert.Contains(t, output, "create")
	assert.Contains(t, output, "list")
	assert.Contains(t, output, "delete")

	// Budget help
	output = runProjectCLI(t, "project", "budget", "--help")
	assert.Contains(t, output, "budget")

	// Members help
	output = runProjectCLI(t, "project", "members", "--help")
	assert.Contains(t, output, "members")
}
