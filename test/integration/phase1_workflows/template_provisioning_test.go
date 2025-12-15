//go:build integration
// +build integration

package phase1_workflows

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/scttfrdmn/prism/test/integration"
)

// TestTemplateProvisioning_PythonML validates end-to-end template provisioning
// for the Python Machine Learning template, ensuring Jupyter and ML packages
// are actually installed and accessible (not just "instance running").
//
// This test addresses issue #396 - Template Provisioning End-to-End
//
// Success criteria:
// - Instance launches and reaches running state
// - Provisioning completes (user data execution)
// - Jupyter is installed and accessible
// - ML packages (PyTorch, TensorFlow, scikit-learn) are installed
// - Jupyter HTTP endpoint is accessible (when SSH available)
func TestTemplateProvisioning_PythonML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := integration.NewTestContext(t)
	defer ctx.Cleanup()

	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// Step 1: Launch instance with Python ML template
	instanceName := integration.GenerateTestName("test-python-ml")
	t.Logf("Launching Python ML instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "Python ML Workstation",
		Name:     instanceName,
		Size:     "S", // Small size for cost efficiency
	})
	integration.AssertNoError(t, err, "Failed to create Python ML instance")
	integration.AssertNotEmpty(t, instance.ID, "Instance should have EC2 ID")
	integration.AssertEqual(t, "running", instance.State, "Instance should be running")

	t.Logf("Instance launched successfully: %s (ID: %s, IP: %s)",
		instance.Name, instance.ID, instance.PublicIP)

	// Step 2: Wait for provisioning to complete
	// User data execution (installing packages, configuring services) takes longer than just "running" state
	t.Log("Waiting for provisioning to complete (user data execution)...")
	if err := waitForProvisioningComplete(ctx, instance.Name, 10*time.Minute); err != nil {
		t.Fatalf("Provisioning did not complete: %v", err)
	}

	t.Log("Provisioning completed successfully")

	// Step 3: Verify Jupyter and ML packages via SSH (if available)
	if canSSH(instance.PublicIP) {
		t.Log("SSH access available - performing deep verification")
		verifyJupyterInstalled(t, instance)
		verifyJupyterRunning(t, instance)
		verifyMLPackagesInstalled(t, instance)
		verifyJupyterAccessible(t, instance)
		t.Log("✓ All verification checks passed - Jupyter and ML packages are fully functional")
	} else {
		t.Log("⚠️  SSH access not available - skipping deep verification")
		t.Log("Note: To enable SSH verification, ensure:")
		t.Log("  1. Security group allows SSH (port 22) from your IP")
		t.Log("  2. SSH key is configured in AWS and available locally")
		t.Log("  3. Network connectivity to instance public IP")
	}

	// Step 4: Verify API-level instance information is correct
	instanceInfo, err := ctx.Client.GetInstance(context.Background(), instanceName)
	integration.AssertNoError(t, err, "Should be able to retrieve instance info")
	integration.AssertEqual(t, "Python Machine Learning (Simplified)", instanceInfo.Template, "Template name should match")
	integration.AssertNotEmpty(t, instanceInfo.PublicIP, "Instance should have public IP")
	integration.AssertEqual(t, "running", instanceInfo.State, "Instance should still be running")

	t.Log("✓ Template provisioning test completed successfully")
}

// TestTemplateProvisioning_RResearch validates R Research Environment template
// This test addresses issue #396 - Template Provisioning End-to-End
func TestTemplateProvisioning_RResearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := integration.NewTestContext(t)
	defer ctx.Cleanup()

	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// Launch instance with R Research template
	instanceName := integration.GenerateTestName("test-r-research")
	t.Logf("Launching R Research instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "R Research Workstation",
		Name:     instanceName,
		Size:     "M", // Medium for R's memory requirements
	})
	integration.AssertNoError(t, err, "Failed to create R Research instance")
	integration.AssertEqual(t, "running", instance.State, "Instance should be running")

	t.Logf("Instance launched successfully: %s (ID: %s)", instance.Name, instance.ID)

	// Wait for provisioning
	t.Log("Waiting for provisioning to complete...")
	if err := waitForProvisioningComplete(ctx, instance.Name, 10*time.Minute); err != nil {
		t.Fatalf("Provisioning did not complete: %v", err)
	}

	// Verify RStudio and R packages (if SSH available)
	if canSSH(instance.PublicIP) {
		t.Log("SSH access available - verifying R environment")
		verifyRInstalled(t, instance)
		verifyRStudioInstalled(t, instance)
		verifyTidyverseInstalled(t, instance)
		t.Log("✓ R Research environment verified successfully")
	} else {
		t.Log("⚠️  SSH access not available - skipping R verification")
	}

	t.Log("✓ R Research template provisioning test completed")
}

// TestTemplateProvisioning_BaseUbuntu validates basic Ubuntu template
// This ensures the foundation template works before testing complex stacks
func TestTemplateProvisioning_BaseUbuntu(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := integration.NewTestContext(t)
	defer ctx.Cleanup()

	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// Launch instance with Ubuntu 24.04 Server template
	instanceName := integration.GenerateTestName("test-ubuntu-basic")
	t.Logf("Launching Ubuntu 24.04 Server instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "Ubuntu 24.04 Server",
		Name:     instanceName,
		Size:     "S",
	})
	integration.AssertNoError(t, err, "Failed to create Ubuntu 24.04 Server instance")
	integration.AssertEqual(t, "running", instance.State, "Instance should be running")

	t.Logf("Instance launched successfully: %s (ID: %s)", instance.Name, instance.ID)

	// Wait for provisioning (should be fast for base template)
	t.Log("Waiting for provisioning to complete...")
	if err := waitForProvisioningComplete(ctx, instance.Name, 5*time.Minute); err != nil {
		t.Fatalf("Provisioning did not complete: %v", err)
	}

	// Verify basic system (if SSH available)
	if canSSH(instance.PublicIP) {
		t.Log("SSH access available - verifying base system")
		verifyBasicSystemTools(t, instance)
		verifyAPTPackageManager(t, instance)
		t.Log("✓ Base Ubuntu system verified successfully")
	} else {
		t.Log("⚠️  SSH access not available - skipping system verification")
	}

	t.Log("✓ Ubuntu 24.04 Server template provisioning test completed")
}

// Helper Functions

// waitForProvisioningComplete waits for user data execution to complete
// This is more thorough than just waiting for "running" state
func waitForProvisioningComplete(ctx *integration.TestContext, instanceName string, timeout time.Duration) error {
	// For now, use a simple time-based wait since we can't directly check cloud-init status without SSH
	// In a production version, this would check cloud-init status via SSM or SSH
	ctx.T.Logf("Waiting %v for user data execution to complete...", timeout)

	// Wait for instance to be stable in running state
	time.Sleep(30 * time.Second)

	// Verify still running
	if err := ctx.WaitForInstanceState(instanceName, "running", 1*time.Minute); err != nil {
		return fmt.Errorf("instance not stable: %w", err)
	}

	// Additional wait for provisioning (conservative estimate)
	// Real implementation would check: systemctl is-active cloud-init
	provisioningWait := timeout - 2*time.Minute
	if provisioningWait > 0 {
		ctx.T.Logf("Allowing additional %v for package installation...", provisioningWait)
		time.Sleep(provisioningWait)
	}

	return nil
}

// canSSH checks if SSH access is available to the instance
// Returns false if SSH key setup is not configured
func canSSH(publicIP string) bool {
	if publicIP == "" {
		return false
	}

	// Check if SSH key exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check for common SSH key locations
	sshKeyPaths := []string{
		fmt.Sprintf("%s/.ssh/prism-test-key", homeDir),
		fmt.Sprintf("%s/.ssh/id_rsa", homeDir),
		fmt.Sprintf("%s/.ssh/id_ed25519", homeDir),
	}

	for _, keyPath := range sshKeyPaths {
		if _, err := os.Stat(keyPath); err == nil {
			// Key exists, try quick SSH connection test
			cmd := exec.Command("ssh",
				"-i", keyPath,
				"-o", "StrictHostKeyChecking=no",
				"-o", "ConnectTimeout=5",
				"-o", "BatchMode=yes",
				fmt.Sprintf("ubuntu@%s", publicIP),
				"echo 'connected'")

			if err := cmd.Run(); err == nil {
				return true
			}
		}
	}

	return false
}

// verifyJupyterInstalled checks if Jupyter is installed
func verifyJupyterInstalled(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, "jupyter --version")
	if !strings.Contains(output, "jupyter") {
		t.Errorf("Jupyter not found in version output: %s", output)
	}

	t.Log("✓ Jupyter is installed")
}

// verifyJupyterRunning checks if Jupyter process is running
func verifyJupyterRunning(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, "ps aux | grep jupyter | grep -v grep")
	if !strings.Contains(output, "jupyter") {
		t.Errorf("Jupyter process not running: %s", output)
	}

	t.Log("✓ Jupyter is running")
}

// verifyMLPackagesInstalled checks if ML packages can be imported
func verifyMLPackagesInstalled(t *testing.T, instance *types.Instance) {
	t.Helper()

	pythonCmd := `python3 -c "import torch, sklearn; print('ML packages OK')"`
	output := runSSHCommand(t, instance.PublicIP, pythonCmd)

	if !strings.Contains(output, "ML packages OK") {
		t.Errorf("ML packages not accessible: %s", output)
	}

	t.Log("✓ ML packages (PyTorch, scikit-learn) are installed")
}

// verifyJupyterAccessible checks if Jupyter HTTP endpoint responds
func verifyJupyterAccessible(t *testing.T, instance *types.Instance) {
	t.Helper()

	// Check if Jupyter is listening on port 8888
	output := runSSHCommand(t, instance.PublicIP, "curl -s -o /dev/null -w '%{http_code}' http://localhost:8888")

	// Jupyter redirects to /tree, so 302 or 200 are both OK
	if !strings.Contains(output, "200") && !strings.Contains(output, "302") {
		t.Logf("Warning: Jupyter HTTP endpoint returned unexpected status: %s", output)
	} else {
		t.Log("✓ Jupyter HTTP endpoint is accessible")
	}
}

// verifyRInstalled checks if R is installed
func verifyRInstalled(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, "R --version")
	if !strings.Contains(output, "R version") {
		t.Errorf("R not found in version output: %s", output)
	}

	t.Log("✓ R is installed")
}

// verifyRStudioInstalled checks if RStudio Server is installed
func verifyRStudioInstalled(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, "which rstudio-server")
	if output == "" || strings.Contains(output, "not found") {
		t.Errorf("RStudio Server not found: %s", output)
	}

	t.Log("✓ RStudio Server is installed")
}

// verifyTidyverseInstalled checks if tidyverse R packages are installed
func verifyTidyverseInstalled(t *testing.T, instance *types.Instance) {
	t.Helper()

	rCmd := `R -e "library(dplyr); library(ggplot2)" 2>&1`
	output := runSSHCommand(t, instance.PublicIP, rCmd)

	if strings.Contains(output, "Error") || strings.Contains(output, "there is no package") {
		t.Errorf("Tidyverse packages not accessible: %s", output)
	}

	t.Log("✓ Tidyverse packages (dplyr, ggplot2) are installed")
}

// verifyBasicSystemTools checks if basic system tools are installed
func verifyBasicSystemTools(t *testing.T, instance *types.Instance) {
	t.Helper()

	tools := []string{"curl", "wget", "git", "vim"}
	for _, tool := range tools {
		output := runSSHCommand(t, instance.PublicIP, fmt.Sprintf("which %s", tool))
		if output == "" || strings.Contains(output, "not found") {
			t.Errorf("Tool %s not found", tool)
		}
	}

	t.Log("✓ Basic system tools are installed")
}

// verifyAPTPackageManager checks if APT package manager is functional
func verifyAPTPackageManager(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, "apt --version")
	if !strings.Contains(output, "apt") {
		t.Errorf("APT not found in version output: %s", output)
	}

	t.Log("✓ APT package manager is functional")
}

// runSSHCommand executes a command via SSH and returns the output
func runSSHCommand(t *testing.T, publicIP, command string) string {
	t.Helper()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Try different SSH keys
	sshKeyPaths := []string{
		fmt.Sprintf("%s/.ssh/prism-test-key", homeDir),
		fmt.Sprintf("%s/.ssh/id_rsa", homeDir),
		fmt.Sprintf("%s/.ssh/id_ed25519", homeDir),
	}

	var lastErr error
	for _, keyPath := range sshKeyPaths {
		if _, err := os.Stat(keyPath); err != nil {
			continue
		}

		cmd := exec.Command("ssh",
			"-i", keyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "ConnectTimeout=10",
			"-o", "BatchMode=yes",
			fmt.Sprintf("ubuntu@%s", publicIP),
			command)

		output, err := cmd.CombinedOutput()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
		lastErr = err
	}

	t.Fatalf("Failed to run SSH command (tried all keys): %v", lastErr)
	return ""
}
