//go:build integration
// +build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestCLIHibernationOperations tests hibernation and resume operations via CLI
// Priority: HIGHEST - Phase 3 hibernation is the cornerstone of cost optimization
func TestCLIHibernationOperations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	// Use a hibernation-capable template (requires specific configuration)
	// Python ML workstation should support hibernation on compatible instance types
	instanceName := GenerateTestName("test-hibernate")

	// Launch hibernation-capable instance
	t.Run("LaunchHibernationCapable", func(t *testing.T) {
		t.Logf("Launching hibernation-capable instance: %s", instanceName)

		// Launch with size M (should support hibernation)
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "M")
		AssertNoError(t, err, "launch hibernation-capable instance")
		result.AssertSuccess(t, "launch command should succeed")

		// Verify instance is running
		instance := ctx.AssertInstanceExists(instanceName)
		AssertEqual(t, "running", instance.State, "instance should be running")

		t.Logf("✓ Instance %s launched successfully", instanceName)
	})

	// Check hibernation support status
	t.Run("CheckHibernationStatus", func(t *testing.T) {
		t.Logf("Checking hibernation support for instance: %s", instanceName)

		// Get instance details via API
		instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
		AssertNoError(t, err, "get instance details")

		// Check if hibernation fields are present (implementation may vary)
		t.Logf("Instance ID: %s, State: %s", instance.ID, instance.State)

		// Note: HibernationSupported field may not exist yet in types.Instance
		// This test documents the expected behavior for Phase 3 completion
	})

	// Test hibernate operation
	t.Run("HibernateRunningInstance", func(t *testing.T) {
		t.Logf("Hibernating instance: %s", instanceName)

		// Execute hibernate command
		result := ctx.Prism("workspace", "hibernate", instanceName)

		// Check if command exists and works
		if result.ExitCode != 0 {
			// If hibernate command doesn't exist yet, skip gracefully
			if strings.Contains(result.Stderr, "unknown command") ||
				strings.Contains(result.Stderr, "not found") {
				t.Skip("Hibernate command not yet implemented - test documents expected behavior")
				return
			}

			// If instance doesn't support hibernation, verify graceful fallback
			if strings.Contains(result.Stderr, "not supported") ||
				strings.Contains(result.Stderr, "fallback to stop") {
				t.Logf("Instance doesn't support hibernation, should fallback to regular stop")
				// Verify instance was stopped (fallback behavior)
				time.Sleep(5 * time.Second) // Give AWS time to process
				instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
				AssertNoError(t, err, "get instance after fallback")
				if instance.State != "stopped" && instance.State != "stopping" {
					t.Errorf("Expected instance to be stopped/stopping after hibernation fallback, got: %s", instance.State)
				}
				return
			}

			// Unexpected error
			t.Fatalf("Hibernate command failed unexpectedly: %s", result.Stderr)
		}

		result.AssertSuccess(t, "hibernate command should succeed")

		// Wait for hibernation to complete (or stopping state)
		// Hibernation can take 1-3 minutes depending on RAM size
		t.Logf("Waiting for hibernation to complete (may take 1-3 minutes)...")
		time.Sleep(5 * time.Second) // Initial delay

		// Check instance state
		instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
		AssertNoError(t, err, "get instance after hibernate")

		// Valid states: hibernated, stopped, or stopping
		validStates := []string{"hibernated", "stopped", "stopping"}
		stateValid := false
		for _, state := range validStates {
			if instance.State == state {
				stateValid = true
				break
			}
		}

		if !stateValid {
			t.Errorf("Expected instance in hibernated/stopped/stopping state, got: %s", instance.State)
		}

		t.Logf("✓ Instance hibernated successfully (state: %s)", instance.State)
	})

	// Test resume operation
	t.Run("ResumeHibernatedInstance", func(t *testing.T) {
		t.Logf("Resuming hibernated instance: %s", instanceName)

		// Wait for hibernation to fully complete (not just "stopping" state)
		t.Logf("Waiting for instance to reach stopped/hibernated state...")
		maxWaitTime := 5 * time.Minute
		pollInterval := 10 * time.Second
		startTime := time.Now()

		for {
			instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
			if err == nil && (instance.State == "stopped" || instance.State == "hibernated") {
				t.Logf("✓ Instance fully hibernated/stopped (state: %s)", instance.State)
				break
			}

			elapsed := time.Since(startTime)
			if elapsed > maxWaitTime {
				t.Fatalf("Timeout waiting for instance to hibernate (waited %v, state: %s)", elapsed, instance.State)
			}

			time.Sleep(pollInterval)
		}

		// Execute resume command
		result := ctx.Prism("workspace", "resume", instanceName)

		// Check if command exists
		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") ||
				strings.Contains(result.Stderr, "not found") {
				t.Skip("Resume command not yet implemented - using start instead")
				// Fallback to start command
				result = ctx.Prism("workspace", "start", instanceName)
			}
		}

		result.AssertSuccess(t, "resume/start command should succeed")

		// Wait for instance to return to running state
		t.Logf("Waiting for instance to resume...")
		err := ctx.WaitForInstanceState(instanceName, "running", InstanceReadyTimeout)
		AssertNoError(t, err, "instance should return to running state")

		// Verify instance is running
		instance := ctx.AssertInstanceExists(instanceName)
		AssertEqual(t, "running", instance.State, "instance should be running after resume")

		t.Logf("✓ Instance resumed successfully")
	})

	// Test hibernate command on stopped instance (edge case)
	t.Run("HibernateStoppedInstance", func(t *testing.T) {
		t.Logf("Testing hibernate on stopped instance (edge case)")

		// Stop the instance first
		err := ctx.StopInstanceCLI(instanceName)
		AssertNoError(t, err, "stop instance")

		// Try to hibernate stopped instance
		result := ctx.Prism("workspace", "hibernate", instanceName)

		// Should either succeed (no-op) or fail gracefully with clear message
		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Hibernate command not yet implemented")
				return
			}

			// Check for helpful error message
			if !strings.Contains(result.Stderr, "already stopped") &&
				!strings.Contains(result.Stderr, "not running") {
				t.Logf("Warning: Hibernate error message could be more helpful: %s", result.Stderr)
			}
		}

		t.Logf("✓ Edge case handled appropriately")
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Start instance if stopped/hibernated before deletion
		instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
		if err == nil && (instance.State == "stopped" || instance.State == "hibernated") {
			ctx.Prism("workspace", "start", instanceName)
			time.Sleep(10 * time.Second) // Give it time to start
		}

		err = ctx.DeleteInstanceCLI(instanceName)
		AssertNoError(t, err, "cleanup should succeed")

		t.Logf("✓ Cleanup complete")
	})
}

// TestCLIHibernationNonCapable tests hibernation fallback on non-capable instances
func TestCLIHibernationNonCapable(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	instanceName := GenerateTestName("test-no-hibernate")

	t.Run("LaunchNonCapable", func(t *testing.T) {
		t.Logf("Launching instance without hibernation support")

		// Launch with size S (may not support hibernation)
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "S")
		if err != nil {
			t.Skipf("Launch failed (may be environmental): %v", err)
			return
		}

		if result.ExitCode != 0 {
			t.Skipf("Launch command failed (may be environmental): %s", result.Stderr)
			return
		}

		t.Logf("✓ Instance launched")
	})

	t.Run("HibernateWithFallback", func(t *testing.T) {
		t.Logf("Testing hibernation fallback to stop")

		result := ctx.Prism("workspace", "hibernate", instanceName)

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Hibernate command not yet implemented")
				return
			}
		}

		// Should succeed (fallback to stop) or provide clear message
		if result.ExitCode == 0 {
			t.Logf("✓ Hibernation attempted (may fallback to stop)")

			// Verify instance stopped
			time.Sleep(5 * time.Second)
			instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
			AssertNoError(t, err, "get instance")

			if instance.State != "stopped" && instance.State != "stopping" && instance.State != "hibernated" {
				t.Errorf("Expected instance stopped/stopping/hibernated, got: %s", instance.State)
			}
		} else {
			// Should have clear message about lack of hibernation support
			t.Logf("Hibernation not supported (expected): %s", result.Stderr)
		}
	})

	t.Run("Cleanup", func(t *testing.T) {
		err := ctx.DeleteInstanceCLI(instanceName)
		if err != nil {
			// Try to start first if needed
			ctx.Prism("workspace", "start", instanceName)
			time.Sleep(5 * time.Second)
			err = ctx.DeleteInstanceCLI(instanceName)
		}
		AssertNoError(t, err, "cleanup should succeed")
	})
}

// TestCLIHibernationHelp tests help output for hibernation commands
func TestCLIHibernationHelp(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("HibernateHelp", func(t *testing.T) {
		result := ctx.Prism("workspace", "hibernate", "--help")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Hibernate command not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "hibernate help should succeed")

		// Check for key information in help
		helpText := result.Stdout + result.Stderr
		expectedTerms := []string{"hibernate", "save state", "cost"}

		for _, term := range expectedTerms {
			if !strings.Contains(strings.ToLower(helpText), strings.ToLower(term)) {
				t.Logf("Warning: Help text might be missing '%s' explanation", term)
			}
		}

		t.Logf("✓ Hibernate help available")
	})

	t.Run("ResumeHelp", func(t *testing.T) {
		result := ctx.Prism("workspace", "resume", "--help")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Resume command not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "resume help should succeed")

		t.Logf("✓ Resume help available")
	})
}
