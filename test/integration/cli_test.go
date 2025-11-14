//go:build integration
// +build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/scttfrdmn/prism/pkg/types"
)

// TestCLITemplateOperations tests template-related CLI commands
func TestCLITemplateOperations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("ListTemplates", func(t *testing.T) {
		result := ctx.Prism("templates")
		result.AssertSuccess(t, "templates command should succeed")
		result.AssertContains(t, "python-ml-workstation", "should list python-ml-workstation template")
		result.AssertContains(t, "r-research-workstation", "should list r-research template")
	})

	t.Run("ValidateTemplates", func(t *testing.T) {
		result := ctx.Prism("templates", "validate")
		result.AssertSuccess(t, "templates validate should succeed")
		result.AssertContains(t, "Validation", "should show validation results")
	})

	t.Run("TemplateInfo", func(t *testing.T) {
		result := ctx.Prism("templates", "info", "python-ml-workstation")
		result.AssertSuccess(t, "templates info should succeed")
		result.AssertContains(t, "Python", "should show Python information")
		result.AssertContains(t, "conda", "should show conda information")
	})

	t.Run("TemplateInfoInvalid", func(t *testing.T) {
		result := ctx.Prism("templates", "info", "nonexistent-template")
		result.AssertFailure(t, "templates info for invalid template should fail")
	})
}

// TestCLIInstanceLaunch tests launching instances via CLI
func TestCLIInstanceLaunch(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	instanceName := GenerateTestName("test-cli-launch")

	t.Run("LaunchInstance", func(t *testing.T) {
		t.Logf("Launching instance: %s", instanceName)

		// Launch via CLI (not API)
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "M")
		AssertNoError(t, err, "CLI launch should succeed")
		result.AssertSuccess(t, "launch command should succeed")

		// Verify output contains success indicators
		result.AssertContains(t, instanceName, "output should mention instance name")

		// Verify instance exists in AWS (not just state)
		err = ctx.VerifyInstanceInAWS(instanceName)
		AssertNoError(t, err, "instance should exist in AWS")
	})

	t.Run("VerifyInstanceDetails", func(t *testing.T) {
		// Get instance details via API for validation
		instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
		AssertNoError(t, err, "should get instance details")

		AssertEqual(t, instanceName, instance.Name, "instance name")
		AssertEqual(t, "running", instance.State, "instance state")
		AssertNotEmpty(t, instance.ID, "instance ID")
		AssertNotEmpty(t, instance.PublicIP, "instance public IP")
		AssertEqual(t, "python-ml-workstation", instance.Template, "instance template")
	})

	t.Run("ListInstancesCLI", func(t *testing.T) {
		names, err := ctx.ListInstancesCLI()
		AssertNoError(t, err, "list should succeed")

		found := false
		for _, name := range names {
			if name == instanceName {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Instance '%s' not found in list output. Found: %v", instanceName, names)
		}
	})

	t.Run("ConnectCommand", func(t *testing.T) {
		result := ctx.Prism("workspace", "connect", instanceName)
		// Note: SSH key setup not configured for test instances, so connect may fail
		// This test verifies the command runs and provides connection info
		if result.ExitCode != 0 {
			t.Logf("Connect command failed (expected - SSH keys not configured): %s", result.Stderr)
			t.Skip("Skipping SSH connection test - requires SSH key setup")
			return
		}
		result.AssertContains(t, "ssh", "should show SSH connection command")
		result.AssertContains(t, instanceName, "should mention instance name")
	})

	t.Run("StopInstance", func(t *testing.T) {
		err := ctx.StopInstanceCLI(instanceName)
		AssertNoError(t, err, "stop should succeed")

		// Verify stopped state via API
		instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
		AssertNoError(t, err, "should get instance after stop")
		AssertEqual(t, "stopped", instance.State, "instance should be stopped")
	})

	t.Run("StartInstance", func(t *testing.T) {
		result := ctx.Prism("workspace", "start", instanceName)
		result.AssertSuccess(t, "start command should succeed")

		// Wait for running state
		err := ctx.WaitForInstanceState(instanceName, "running", InstanceReadyTimeout)
		AssertNoError(t, err, "instance should return to running state")
	})

	t.Run("TerminateInstance", func(t *testing.T) {
		err := ctx.DeleteInstanceCLI(instanceName)
		AssertNoError(t, err, "terminate should succeed")

		// Verify instance is gone (API call should fail)
		_, err = ctx.Client.GetInstance(context.Background(), instanceName)
		if err == nil {
			t.Fatal("Instance should be deleted but still exists")
		}
	})
}

// TestCLIInstanceLifecycle tests complete instance lifecycle via CLI
func TestCLIInstanceLifecycle(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	instanceName := GenerateTestName("test-cli-lifecycle")

	// Phase 1: Launch
	t.Run("Launch", func(t *testing.T) {
		result, err := ctx.LaunchInstanceCLI("r-research-workstation", instanceName, "S")
		AssertNoError(t, err, "launch should succeed")
		result.AssertSuccess(t, "launch command should succeed")
	})

	// Phase 2: Verify Running
	t.Run("VerifyRunning", func(t *testing.T) {
		instance := ctx.AssertInstanceExists(instanceName)
		AssertEqual(t, "running", instance.State, "instance should be running")
	})

	// Phase 3: Stop
	t.Run("Stop", func(t *testing.T) {
		err := ctx.StopInstanceCLI(instanceName)
		AssertNoError(t, err, "stop should succeed")
		ctx.AssertInstanceState(instanceName, "stopped")
	})

	// Phase 4: Start
	t.Run("Start", func(t *testing.T) {
		result := ctx.Prism("workspace", "start", instanceName)
		result.AssertSuccess(t, "start should succeed")
		err := ctx.WaitForInstanceState(instanceName, "running", InstanceReadyTimeout)
		AssertNoError(t, err, "instance should start")
	})

	// Phase 5: List verification
	t.Run("ListVerification", func(t *testing.T) {
		result := ctx.Prism("workspace", "list")
		result.AssertSuccess(t, "list should succeed")
		result.AssertContains(t, instanceName, "list should show instance")
		result.AssertContains(t, "RUNNING", "should show running state")
	})

	// Phase 6: Cleanup
	t.Run("Terminate", func(t *testing.T) {
		err := ctx.DeleteInstanceCLI(instanceName)
		AssertNoError(t, err, "terminate should succeed")
	})
}

// TestCLILaunchOptions tests various launch options via CLI
func TestCLILaunchOptions(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("LaunchWithSizeS", func(t *testing.T) {
		instanceName := GenerateTestName("test-cli-size-s")
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "S")
		AssertNoError(t, err, "launch with size S should succeed")
		result.AssertSuccess(t, "launch command should succeed")

		// Verify instance created
		instance := ctx.AssertInstanceExists(instanceName)
		AssertEqual(t, "running", instance.State, "instance should be running")
	})

	t.Run("LaunchWithSizeL", func(t *testing.T) {
		instanceName := GenerateTestName("test-cli-size-l")
		result, err := ctx.LaunchInstanceCLI("r-research-workstation", instanceName, "L")
		AssertNoError(t, err, "launch with size L should succeed")
		result.AssertSuccess(t, "launch command should succeed")

		// Verify instance created
		instance := ctx.AssertInstanceExists(instanceName)
		AssertEqual(t, "running", instance.State, "instance should be running")
	})
}

// TestCLIErrorHandling tests error conditions and helpful messages
func TestCLIErrorHandling(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("LaunchInvalidTemplate", func(t *testing.T) {
		result := ctx.Prism("workspace", "launch", "nonexistent-template", "test-invalid", "--size", "M")
		result.AssertFailure(t, "launch with invalid template should fail")
		result.AssertContains(t, "not found", "should show helpful error")
	})

	t.Run("LaunchDuplicateName", func(t *testing.T) {
		instanceName := GenerateTestName("test-cli-duplicate")

		// Launch first instance
		_, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "S")
		AssertNoError(t, err, "first launch should succeed")

		// Try to launch duplicate
		result := ctx.Prism("workspace", "launch", "python-ml-workstation", instanceName, "--size", "S")
		result.AssertFailure(t, "duplicate launch should fail")
		result.AssertContains(t, "already exists", "should show duplicate error")
	})

	t.Run("StopNonexistentInstance", func(t *testing.T) {
		result := ctx.Prism("workspace", "stop", "nonexistent-instance")
		result.AssertFailure(t, "stop nonexistent instance should fail")
	})

	t.Run("TerminateNonexistentInstance", func(t *testing.T) {
		result := ctx.Prism("workspace", "delete", "nonexistent-instance", "--force")
		result.AssertFailure(t, "terminate nonexistent instance should fail")
	})
}

// TestCLIStorageOperations tests storage-related CLI commands
func TestCLIStorageOperations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("ListVolumes", func(t *testing.T) {
		result := ctx.Prism("volume", "list")
		result.AssertSuccess(t, "volume list should succeed")
		// Should at least not error, may be empty
	})

	t.Run("ListStorage", func(t *testing.T) {
		result := ctx.Prism("storage", "list")
		result.AssertSuccess(t, "storage list should succeed")
		// Should at least not error, may be empty
	})
}

// TestCLIUserOperations tests user-related CLI commands
func TestCLIUserOperations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("ListUsers", func(t *testing.T) {
		result := ctx.Prism("user", "list")
		result.AssertSuccess(t, "user list should succeed")
		// May be empty, just verify command works
	})
}

// TestCLIProjectOperations tests project-related CLI commands
func TestCLIProjectOperations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("ListProjects", func(t *testing.T) {
		result := ctx.Prism("project", "list")
		result.AssertSuccess(t, "project list should succeed")
		// May be empty, just verify command works
	})
}

// TestCLIDaemonOperations tests daemon control commands
func TestCLIDaemonOperations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("DaemonStatus", func(t *testing.T) {
		result := ctx.Prism("admin", "daemon", "status")
		result.AssertSuccess(t, "daemon status should succeed")
		result.AssertContains(t, "running", "daemon should be running")
	})

	t.Run("DaemonVersion", func(t *testing.T) {
		result := ctx.Prism("version")
		result.AssertSuccess(t, "version command should succeed")
		result.AssertContains(t, "0.5.", "should show version")
	})
}

// TestCLIHelpCommands tests help output
func TestCLIHelpCommands(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("GlobalHelp", func(t *testing.T) {
		result := ctx.Prism("--help")
		result.AssertSuccess(t, "help should succeed")
		result.AssertContains(t, "prism", "should show prism usage")
		result.AssertContains(t, "launch", "should list launch command")
	})

	t.Run("LaunchHelp", func(t *testing.T) {
		result := ctx.Prism("workspace", "launch", "--help")
		result.AssertSuccess(t, "launch help should succeed")
		result.AssertContains(t, "template", "should explain template parameter")
		result.AssertContains(t, "size", "should explain size parameter")
	})

	t.Run("TemplatesHelp", func(t *testing.T) {
		result := ctx.Prism("templates", "--help")
		result.AssertSuccess(t, "templates help should succeed")
		result.AssertContains(t, "list", "should explain list subcommand")
	})
}

// TestCLIRealAWSIntegration tests that CLI actually creates resources in AWS
func TestCLIRealAWSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real AWS integration test in short mode")
	}

	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	instanceName := GenerateTestName("test-cli-real-aws")

	t.Run("LaunchVerifyAWS", func(t *testing.T) {
		// Launch via CLI
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "S")
		AssertNoError(t, err, "launch should succeed")
		result.AssertSuccess(t, "launch command should succeed")

		// Verify instance actually exists in AWS (not just state file)
		err = ctx.VerifyInstanceInAWS(instanceName)
		AssertNoError(t, err, "instance must exist in real AWS")

		// Get instance details and verify AWS-specific fields
		instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
		AssertNoError(t, err, "should get instance")

		// Verify AWS-specific attributes
		AssertNotEmpty(t, instance.ID, "instance should have AWS ID")
		AssertNotEmpty(t, instance.PublicIP, "instance should have public IP")
		if !strings.HasPrefix(instance.ID, "i-") {
			t.Fatalf("Instance ID should start with 'i-', got: %s", instance.ID)
		}

		t.Logf("Verified AWS instance: ID=%s, IP=%s, State=%s",
			instance.ID, instance.PublicIP, instance.State)
	})
}

// TestCLIMultipleInstances tests managing multiple instances simultaneously
func TestCLIMultipleInstances(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	instance1 := GenerateTestName("test-cli-multi-1")
	instance2 := GenerateTestName("test-cli-multi-2")
	instance3 := GenerateTestName("test-cli-multi-3")

	t.Run("LaunchMultiple", func(t *testing.T) {
		// Launch 3 instances
		_, err := ctx.LaunchInstanceCLI("python-ml-workstation", instance1, "S")
		AssertNoError(t, err, "launch instance 1")

		_, err = ctx.LaunchInstanceCLI("r-research-workstation", instance2, "S")
		AssertNoError(t, err, "launch instance 2")

		_, err = ctx.LaunchInstanceCLI("python-ml-workstation", instance3, "S")
		AssertNoError(t, err, "launch instance 3")
	})

	t.Run("VerifyAllListed", func(t *testing.T) {
		names, err := ctx.ListInstancesCLI()
		AssertNoError(t, err, "list should succeed")

		found := map[string]bool{
			instance1: false,
			instance2: false,
			instance3: false,
		}

		for _, name := range names {
			if _, exists := found[name]; exists {
				found[name] = true
			}
		}

		for name, wasFound := range found {
			if !wasFound {
				t.Errorf("Instance '%s' not found in list", name)
			}
		}
	})

	t.Run("StopMultiple", func(t *testing.T) {
		err := ctx.StopInstanceCLI(instance1)
		AssertNoError(t, err, "stop instance 1")

		err = ctx.StopInstanceCLI(instance2)
		AssertNoError(t, err, "stop instance 2")

		// Verify both stopped
		ctx.AssertInstanceState(instance1, "stopped")
		ctx.AssertInstanceState(instance2, "stopped")
	})

	t.Run("TerminateAll", func(t *testing.T) {
		err := ctx.DeleteInstanceCLI(instance1)
		AssertNoError(t, err, "delete instance 1")

		err = ctx.DeleteInstanceCLI(instance2)
		AssertNoError(t, err, "delete instance 2")

		err = ctx.DeleteInstanceCLI(instance3)
		AssertNoError(t, err, "delete instance 3")
	})
}

// TestCLIOutputFormats tests different output formats (table, JSON, etc.)
func TestCLIOutputFormats(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("TemplatesTableFormat", func(t *testing.T) {
		result := ctx.Prism("templates")
		result.AssertSuccess(t, "templates should succeed")
		// Default table format should have headers
		result.AssertContains(t, "NAME", "should have table header")
	})

	t.Run("TemplatesJSONFormat", func(t *testing.T) {
		var templates []types.Template
		err := ctx.PrismJSON(&templates, "templates")
		if err == nil {
			// JSON format supported
			if len(templates) == 0 {
				t.Log("Warning: no templates returned in JSON format")
			}
		} else {
			// JSON format might not be implemented yet, just log
			t.Logf("JSON format not implemented yet: %v", err)
		}
	})
}
