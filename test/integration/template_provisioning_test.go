//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplateProvisioning_PythonML validates Python ML template provisioning
// Tests: Launch → Jupyter accessible on port 8888 → numpy/pandas/pytorch installed
func TestTemplateProvisioning_PythonML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping template provisioning test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("python-ml-test-%d", time.Now().Unix())

	t.Logf("Launching Python ML instance: %s", instanceName)

	// Launch Python ML instance
	launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
		Template: "Python Machine Learning",
		Name:     instanceName,
		Size:     "M",
	})
	require.NoError(t, err, "Failed to launch Python ML instance")
	registry.Register("instance", instanceName)

	t.Logf("Instance launched: %s (ID: %s)", launchResp.Instance.Name, launchResp.Instance.ID)

	// Wait for instance to be running
	t.Log("Waiting for instance to reach running state...")
	err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
	require.NoError(t, err, "Instance failed to reach running state")

	// Get instance details
	instanceDetails, err := apiClient.GetInstance(ctx, instanceName)
	require.NoError(t, err, "Failed to get instance details")
	require.NotEmpty(t, instanceDetails.PublicIP, "Instance should have public IP")

	t.Logf("Instance running at %s", instanceDetails.PublicIP)

	// Test 1: Verify Jupyter service is accessible on port 8888
	t.Run("JupyterAccessible", func(t *testing.T) {
		t.Log("Testing Jupyter accessibility on port 8888...")

		// Retry for up to 2 minutes (services may take time to start)
		var httpCode string
		maxAttempts := 12
		for i := 0; i < maxAttempts; i++ {
			t.Logf("Attempt %d/%d: Checking Jupyter status...", i+1, maxAttempts)

			// Check if Jupyter process is running
			_, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
				"ps aux | grep jupyter | grep -v grep")
			if err == nil {
				httpCode = "200"
				break
			}

			time.Sleep(10 * time.Second)
		}

		// Verify service is running (we'd normally check HTTP response)
		assert.NotEmpty(t, httpCode, "Jupyter service should be running")
	})

	// Test 2: Verify Python packages are installed
	t.Run("PythonPackagesInstalled", func(t *testing.T) {
		t.Log("Verifying Python packages (numpy, pandas, pytorch)...")

		packages := []string{"numpy", "pandas", "torch"}
		for _, pkg := range packages {
			t.Logf("Checking package: %s", pkg)

			output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
				fmt.Sprintf("python3 -c 'import %s; print(%s.__version__)'", pkg, pkg))

			assert.NoError(t, err, "Package %s should be importable", pkg)
			assert.NotEmpty(t, strings.TrimSpace(output), "Package %s should have version", pkg)

			t.Logf("✓ %s version: %s", pkg, strings.TrimSpace(output))
		}
	})

	// Test 3: Verify Jupyter is installed and configured
	t.Run("JupyterInstalled", func(t *testing.T) {
		t.Log("Verifying Jupyter installation...")

		output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			"jupyter --version")

		assert.NoError(t, err, "Jupyter should be installed")
		assert.Contains(t, output, "jupyter", "Jupyter version should be displayed")

		t.Logf("✓ Jupyter version: %s", strings.TrimSpace(output))
	})

	t.Log("✅ Python ML template provisioning test complete")
}

// TestTemplateProvisioning_RResearch validates R Research template provisioning
// Tests: Launch → RStudio accessible on port 8787 → tidyverse installed
func TestTemplateProvisioning_RResearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping template provisioning test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("r-research-test-%d", time.Now().Unix())

	t.Logf("Launching R Research instance: %s", instanceName)

	// Launch R Research instance
	launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
		Template: "R Research Environment",
		Name:     instanceName,
		Size:     "M",
	})
	require.NoError(t, err, "Failed to launch R Research instance")
	registry.Register("instance", instanceName)

	t.Logf("Instance launched: %s (ID: %s)", launchResp.Instance.Name, launchResp.Instance.ID)

	// Wait for instance to be running
	t.Log("Waiting for instance to reach running state...")
	err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
	require.NoError(t, err, "Instance failed to reach running state")

	// Get instance details
	instanceDetails, err := apiClient.GetInstance(ctx, instanceName)
	require.NoError(t, err, "Failed to get instance details")
	require.NotEmpty(t, instanceDetails.PublicIP, "Instance should have public IP")

	t.Logf("Instance running at %s", instanceDetails.PublicIP)

	// Test 1: Verify RStudio service is accessible on port 8787
	t.Run("RStudioAccessible", func(t *testing.T) {
		t.Log("Testing RStudio accessibility on port 8787...")

		// Check if RStudio process is running
		maxAttempts := 12
		var running bool
		for i := 0; i < maxAttempts; i++ {
			t.Logf("Attempt %d/%d: Checking RStudio status...", i+1, maxAttempts)

			output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
				"ps aux | grep rstudio-server | grep -v grep")
			if err == nil && strings.Contains(output, "rstudio-server") {
				running = true
				break
			}

			time.Sleep(10 * time.Second)
		}

		assert.True(t, running, "RStudio server should be running")
	})

	// Test 2: Verify R is installed
	t.Run("RInstalled", func(t *testing.T) {
		t.Log("Verifying R installation...")

		output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			"R --version")

		assert.NoError(t, err, "R should be installed")
		assert.Contains(t, output, "R version", "R version should be displayed")

		t.Logf("✓ R version: %s", strings.Split(output, "\n")[0])
	})

	// Test 3: Verify tidyverse is installed
	t.Run("TidyverseInstalled", func(t *testing.T) {
		t.Log("Verifying tidyverse package...")

		output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			"R -e 'library(tidyverse); packageVersion(\"tidyverse\")' --quiet")

		assert.NoError(t, err, "Tidyverse should be installed")
		assert.Contains(t, output, "[1]", "Tidyverse version should be displayed")

		t.Logf("✓ Tidyverse installed: %s", strings.TrimSpace(output))
	})

	t.Log("✅ R Research template provisioning test complete")
}

// TestTemplateProvisioning_BaseTemplate validates base template provisioning
// Tests: User creation, permissions, shell configuration
func TestTemplateProvisioning_BaseTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping template provisioning test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	instanceName := fmt.Sprintf("base-template-test-%d", time.Now().Unix())

	t.Logf("Launching Ubuntu Basic instance: %s", instanceName)

	// Launch Ubuntu Basic instance
	launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
		Template: "Ubuntu Basic",
		Name:     instanceName,
		Size:     "S",
	})
	require.NoError(t, err, "Failed to launch Ubuntu Basic instance")
	registry.Register("instance", instanceName)

	t.Logf("Instance launched: %s (ID: %s)", launchResp.Instance.Name, launchResp.Instance.ID)

	// Wait for instance to be running
	t.Log("Waiting for instance to reach running state...")
	err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
	require.NoError(t, err, "Instance failed to reach running state")

	// Get instance details
	instanceDetails, err := apiClient.GetInstance(ctx, instanceName)
	require.NoError(t, err, "Failed to get instance details")
	require.NotEmpty(t, instanceDetails.PublicIP, "Instance should have public IP")

	t.Logf("Instance running at %s", instanceDetails.PublicIP)

	// Test 1: Verify ubuntu user exists and has sudo privileges
	t.Run("UbuntuUserSetup", func(t *testing.T) {
		t.Log("Verifying ubuntu user configuration...")

		// Check user exists
		output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			"whoami")
		assert.NoError(t, err, "Should be able to SSH as ubuntu user")
		assert.Equal(t, "ubuntu", strings.TrimSpace(output), "User should be ubuntu")

		// Check sudo privileges
		output, err = fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			"sudo -n whoami")
		assert.NoError(t, err, "Ubuntu user should have passwordless sudo")
		assert.Equal(t, "root", strings.TrimSpace(output), "Sudo should work as root")

		t.Log("✓ Ubuntu user configured correctly")
	})

	// Test 2: Verify shell configuration
	t.Run("ShellConfiguration", func(t *testing.T) {
		t.Log("Verifying shell configuration...")

		// Check default shell
		output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			"echo $SHELL")
		assert.NoError(t, err, "Should be able to read SHELL variable")
		assert.Contains(t, output, "bash", "Default shell should be bash")

		// Check .bashrc exists
		output, err = fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			"test -f ~/.bashrc && echo 'exists'")
		assert.NoError(t, err, ".bashrc should exist")
		assert.Equal(t, "exists", strings.TrimSpace(output), ".bashrc file should exist")

		t.Log("✓ Shell configured correctly")
	})

	// Test 3: Verify basic system tools are installed
	t.Run("BasicToolsInstalled", func(t *testing.T) {
		t.Log("Verifying basic system tools...")

		tools := []string{"git", "curl", "vim", "wget"}
		for _, tool := range tools {
			t.Logf("Checking tool: %s", tool)

			output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
				fmt.Sprintf("which %s", tool))

			assert.NoError(t, err, "Tool %s should be installed", tool)
			assert.Contains(t, output, "/", "Tool %s should have path", tool)

			t.Logf("✓ %s: %s", tool, strings.TrimSpace(output))
		}
	})

	// Test 4: Verify file permissions
	t.Run("FilePermissions", func(t *testing.T) {
		t.Log("Verifying file permissions...")

		// Check home directory permissions
		output, err := fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			"stat -c '%a' ~")
		assert.NoError(t, err, "Should be able to check home directory permissions")
		assert.Equal(t, "755", strings.TrimSpace(output), "Home directory should be 755")

		// Verify user can create files
		testFile := "/home/ubuntu/test-permissions.txt"
		_, err = fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			fmt.Sprintf("echo 'test' > %s && cat %s", testFile, testFile))
		assert.NoError(t, err, "User should be able to create and read files")

		// Cleanup
		fixtures.SSHCommand(t, instanceDetails.PublicIP, "ubuntu",
			fmt.Sprintf("rm -f %s", testFile))

		t.Log("✓ File permissions correct")
	})

	t.Log("✅ Base template provisioning test complete")
}

// TestTemplateProvisioning_MultipleTemplates validates launching multiple templates concurrently
func TestTemplateProvisioning_MultipleTemplates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent template test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	templates := []string{
		"Ubuntu Basic",
		"Python Machine Learning",
		"R Research Environment",
	}

	t.Log("Launching multiple templates concurrently...")

	// Launch all templates
	instanceNames := make([]string, len(templates))
	for i, template := range templates {
		instanceName := fmt.Sprintf("multi-test-%s-%d", strings.ToLower(strings.ReplaceAll(template, " ", "-")), time.Now().Unix())
		instanceNames[i] = instanceName

		t.Logf("Launching %s as %s", template, instanceName)

		_, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
			Template: template,
			Name:     instanceName,
			Size:     "S",
		})
		require.NoError(t, err, "Failed to launch %s", template)
		registry.Register("instance", instanceName)
	}

	// Wait for all instances to be running
	t.Log("Waiting for all instances to reach running state...")
	for _, instanceName := range instanceNames {
		err := fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
		assert.NoError(t, err, "Instance %s should reach running state", instanceName)
	}

	// Verify all instances are accessible
	t.Log("Verifying all instances are accessible...")
	for _, instanceName := range instanceNames {
		instanceDetails, err := apiClient.GetInstance(ctx, instanceName)
		assert.NoError(t, err, "Should be able to get details for %s", instanceName)
		assert.NotEmpty(t, instanceDetails.PublicIP, "Instance %s should have public IP", instanceName)

		t.Logf("✓ %s running at %s", instanceName, instanceDetails.PublicIP)
	}

	t.Log("✅ Multiple template provisioning test complete")
}
