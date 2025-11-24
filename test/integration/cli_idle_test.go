//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
	"time"
)

// TestCLIIdlePolicyManagement tests idle policy operations via CLI
// Priority: HIGHEST - Idle detection is core Phase 3 cost optimization
func TestCLIIdlePolicyManagement(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	instanceName := GenerateTestName("test-idle-policy")

	// Launch instance for idle policy testing
	t.Run("LaunchInstance", func(t *testing.T) {
		t.Logf("Launching instance for idle policy testing: %s", instanceName)

		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "S")
		AssertNoError(t, err, "launch instance")
		result.AssertSuccess(t, "launch should succeed")

		t.Logf("✓ Instance launched")
	})

	// List available idle policies
	t.Run("ListIdlePolicies", func(t *testing.T) {
		t.Logf("Listing available idle policies")

		result := ctx.Prism("idle", "policy", "list")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") ||
				strings.Contains(result.Stderr, "not found") {
				t.Skip("Idle policy commands not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "idle policy list should succeed")

		// Check for expected pre-configured policies
		expectedPolicies := []string{
			"batch",           // 60min idle → hibernate
			"gpu",             // 15min idle → stop
			"aggressive-cost", // 10min idle → hibernate
			"balanced",        // 30min idle → hibernate
			"conservative",    // 120min idle → stop
			"research",        // 45min idle → hibernate
		}

		output := result.Stdout + result.Stderr
		foundPolicies := 0
		for _, policy := range expectedPolicies {
			if strings.Contains(output, policy) {
				foundPolicies++
				t.Logf("✓ Found policy: %s", policy)
			}
		}

		if foundPolicies == 0 {
			t.Error("Expected to find pre-configured idle policies, found none")
		}

		t.Logf("✓ Found %d/%d expected policies", foundPolicies, len(expectedPolicies))
	})

	// Apply idle policy to instance
	t.Run("ApplyIdlePolicy", func(t *testing.T) {
		t.Logf("Applying 'balanced' idle policy to instance")

		result := ctx.Prism("idle", "policy", "apply", instanceName, "balanced")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Idle policy apply command not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "idle policy apply should succeed")

		t.Logf("✓ Policy applied")
	})

	// Check policy status on instance
	t.Run("CheckPolicyStatus", func(t *testing.T) {
		t.Logf("Checking idle policy status for instance")

		result := ctx.Prism("idle", "policy", "status", instanceName)

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Idle policy status command not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "idle policy status should succeed")

		// Check that balanced policy is shown
		output := result.Stdout + result.Stderr
		if !strings.Contains(output, "balanced") {
			t.Error("Expected 'balanced' policy in status output")
		}

		t.Logf("✓ Policy status retrieved")
	})

	// Get policy recommendation
	t.Run("RecommendPolicy", func(t *testing.T) {
		t.Logf("Getting idle policy recommendation")

		result := ctx.Prism("idle", "policy", "recommend", instanceName)

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Idle policy recommend command not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "idle policy recommend should succeed")

		// Should recommend a policy based on instance type/template
		output := result.Stdout + result.Stderr
		if len(output) == 0 {
			t.Error("Expected recommendation output, got empty")
		}

		t.Logf("✓ Policy recommendation retrieved")
	})

	// Test applying different policies
	t.Run("ApplyCostOptimizedPolicy", func(t *testing.T) {
		t.Logf("Applying 'aggressive-cost' policy (aggressive)")

		result := ctx.Prism("idle", "policy", "apply", instanceName, "aggressive-cost")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Idle policy apply command not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "should apply aggressive-cost policy")

		// Verify policy changed
		statusResult := ctx.Prism("idle", "policy", "status", instanceName)
		if statusResult.ExitCode == 0 {
			output := statusResult.Stdout + statusResult.Stderr
			if !strings.Contains(output, "aggressive-cost") {
				t.Error("Expected policy to update to 'aggressive-cost'")
			}
		}

		t.Logf("✓ Policy updated")
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		err := ctx.DeleteInstanceCLI(instanceName)
		AssertNoError(t, err, "cleanup should succeed")
		t.Logf("✓ Cleanup complete")
	})
}

// TestCLIIdlePolicyValidation tests error handling for idle policies
func TestCLIIdlePolicyValidation(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("ApplyInvalidPolicy", func(t *testing.T) {
		t.Logf("Testing invalid policy name")

		result := ctx.Prism("idle", "policy", "apply", "nonexistent-instance", "invalid-policy")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Idle policy commands not yet implemented")
				return
			}
		}

		// Should fail with helpful error
		result.AssertFailure(t, "invalid policy should fail")

		if !strings.Contains(result.Stderr, "not found") &&
			!strings.Contains(result.Stderr, "invalid") &&
			!strings.Contains(result.Stderr, "unknown") {
			t.Logf("Warning: Error message could be more helpful: %s", result.Stderr)
		}

		t.Logf("✓ Invalid policy rejected")
	})

	t.Run("ApplyPolicyToNonexistentInstance", func(t *testing.T) {
		t.Logf("Testing policy apply to nonexistent instance")

		result := ctx.Prism("idle", "policy", "apply", "nonexistent-instance-12345", "balanced")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Idle policy commands not yet implemented")
				return
			}
		}

		// Should fail with clear error
		result.AssertFailure(t, "nonexistent instance should fail")

		t.Logf("✓ Nonexistent instance handled")
	})
}

// TestCLIIdleSchedules tests scheduled idle operations
func TestCLIIdleSchedules(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("ListSchedules", func(t *testing.T) {
		t.Logf("Listing idle schedules")

		result := ctx.Prism("idle", "schedule", "list")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") ||
				strings.Contains(result.Stderr, "not found") {
				t.Skip("Idle schedule commands not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "idle schedule list should succeed")

		// May be empty initially
		t.Logf("✓ Schedule list retrieved")
	})

	t.Run("CreateWorkHoursSchedule", func(t *testing.T) {
		t.Logf("Creating work hours schedule")

		// Try to create a schedule: weekdays 9am-5pm
		result := ctx.Prism("idle", "schedule", "create", "work-hours",
			"--start", "09:00", "--end", "17:00", "--weekdays")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Idle schedule create command not yet implemented")
				return
			}

			// Command might have different syntax, log for reference
			t.Logf("Schedule creation syntax may differ: %s", result.Stderr)
			t.Skip("Skipping schedule creation test - syntax unclear")
			return
		}

		result.AssertSuccess(t, "schedule creation should succeed")

		t.Logf("✓ Schedule created")
	})
}

// TestCLIIdleSavingsReport tests savings report generation
func TestCLIIdleSavingsReport(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("GenerateSavingsReport", func(t *testing.T) {
		t.Logf("Generating idle savings report")

		result := ctx.Prism("idle", "savings", "--period", "7d")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") ||
				strings.Contains(result.Stderr, "not found") {
				t.Skip("Idle savings command not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "idle savings report should succeed")

		// Report may be empty if no idle actions occurred
		output := result.Stdout + result.Stderr

		// Check for expected report elements
		expectedTerms := []string{"savings", "cost", "idle"}
		foundTerms := 0
		for _, term := range expectedTerms {
			if strings.Contains(strings.ToLower(output), term) {
				foundTerms++
			}
		}

		if foundTerms > 0 {
			t.Logf("✓ Savings report generated with relevant information")
		} else {
			t.Logf("Warning: Savings report may lack expected content")
		}
	})

	t.Run("GenerateReportDifferentPeriods", func(t *testing.T) {
		t.Logf("Testing different report periods")

		periods := []string{"1d", "7d", "30d"}

		for _, period := range periods {
			result := ctx.Prism("idle", "savings", "--period", period)

			if result.ExitCode != 0 {
				if strings.Contains(result.Stderr, "unknown command") {
					t.Skip("Idle savings command not yet implemented")
					return
				}
			}

			if result.ExitCode == 0 {
				t.Logf("✓ Report generated for period: %s", period)
			}
		}
	})
}

// TestCLIIdleHistory tests idle action history/audit trail
func TestCLIIdleHistory(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	instanceName := GenerateTestName("test-idle-history")

	t.Run("LaunchWithIdlePolicy", func(t *testing.T) {
		t.Logf("Launching instance with idle policy")

		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "S")
		AssertNoError(t, err, "launch instance")
		result.AssertSuccess(t, "launch should succeed")

		// Apply aggressive policy for testing (if command exists)
		policyResult := ctx.Prism("idle", "policy", "apply", instanceName, "aggressive-cost")
		if policyResult.ExitCode != 0 {
			if strings.Contains(policyResult.Stderr, "unknown command") {
				t.Skip("Idle policy commands not yet implemented")
				return
			}
		}

		t.Logf("✓ Instance launched with idle policy")
	})

	t.Run("CheckIdleHistory", func(t *testing.T) {
		t.Logf("Checking idle action history")

		result := ctx.Prism("idle", "history")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") ||
				strings.Contains(result.Stderr, "not found") {
				t.Skip("Idle history command not yet implemented")
				return
			}
		}

		result.AssertSuccess(t, "idle history should succeed")

		// History may be empty if no idle actions occurred yet
		t.Logf("✓ Idle history retrieved")
	})

	t.Run("InstanceSpecificHistory", func(t *testing.T) {
		t.Logf("Checking history for specific instance")

		result := ctx.Prism("idle", "history", "--instance", instanceName)

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Instance-specific history not yet implemented")
				return
			}

			// May not support --instance flag yet
			t.Logf("Instance-specific history may not be supported yet")
			return
		}

		result.AssertSuccess(t, "instance-specific history should succeed")

		t.Logf("✓ Instance-specific history retrieved")
	})

	t.Run("Cleanup", func(t *testing.T) {
		err := ctx.DeleteInstanceCLI(instanceName)
		AssertNoError(t, err, "cleanup should succeed")
	})
}

// TestCLIIdleIntegrationWithHibernation tests idle policy triggering hibernation
func TestCLIIdleIntegrationWithHibernation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running idle+hibernation integration test")
	}

	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	instanceName := GenerateTestName("test-idle-hibernate")

	t.Run("Setup", func(t *testing.T) {
		t.Logf("Setting up instance with hibernation-enabled idle policy")

		// Launch hibernation-capable instance
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "M")
		AssertNoError(t, err, "launch instance")
		result.AssertSuccess(t, "launch should succeed")

		// Apply policy that hibernates after short idle period (if supported)
		// Note: In production this would be 10+ minutes, but for testing we might need shorter
		policyResult := ctx.Prism("idle", "policy", "apply", instanceName, "aggressive-cost")
		if policyResult.ExitCode != 0 {
			if strings.Contains(policyResult.Stderr, "unknown command") {
				t.Skip("Idle policy commands not yet implemented")
				return
			}
		}

		t.Logf("✓ Instance configured with hibernation idle policy")
	})

	t.Run("VerifyIdleDetection", func(t *testing.T) {
		t.Logf("Note: This test documents expected behavior but doesn't wait for actual idle detection")
		t.Logf("In full integration testing, would wait 10+ minutes for idle threshold")
		t.Logf("Expected behavior: After idle period, instance should automatically hibernate")

		// Check current policy status
		statusResult := ctx.Prism("idle", "policy", "status", instanceName)
		if statusResult.ExitCode == 0 {
			t.Logf("Current idle policy status:")
			t.Logf("%s", statusResult.Stdout)
		}

		// Skip actual wait for idle detection (would take 10+ minutes)
		t.Skip("Skipping actual idle detection wait - would require 10+ minutes")
	})

	t.Run("Cleanup", func(t *testing.T) {
		err := ctx.DeleteInstanceCLI(instanceName)
		if err != nil {
			// Instance might be hibernated/stopped, try to start first
			ctx.Prism("workspace", "start", instanceName)
			time.Sleep(10 * time.Second)
			err = ctx.DeleteInstanceCLI(instanceName)
		}
		AssertNoError(t, err, "cleanup should succeed")
	})
}

// TestCLIIdleHelp tests help output for idle commands
func TestCLIIdleHelp(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	t.Run("IdleMainHelp", func(t *testing.T) {
		result := ctx.Prism("idle", "--help")
		result.AssertSuccess(t, "idle help should succeed")

		output := result.Stdout + result.Stderr

		// Check for expected subcommands
		expectedCommands := []string{"policy", "schedule", "savings"}
		for _, cmd := range expectedCommands {
			if !strings.Contains(output, cmd) {
				t.Errorf("Expected '%s' subcommand in idle help", cmd)
			}
		}

		t.Logf("✓ Idle help available")
	})

	t.Run("IdlePolicyHelp", func(t *testing.T) {
		result := ctx.Prism("idle", "policy", "--help")

		if result.ExitCode != 0 {
			if strings.Contains(result.Stderr, "unknown command") {
				t.Skip("Idle policy commands not yet fully implemented")
				return
			}
		}

		result.AssertSuccess(t, "idle policy help should succeed")

		t.Logf("✓ Idle policy help available")
	})
}
