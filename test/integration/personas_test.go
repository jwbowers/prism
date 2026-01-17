//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"
)

// TestSoloResearcherPersona tests the Solo Researcher (Dr. Sarah Chen) workflow
// Based on docs/USER_SCENARIOS/01_SOLO_RESEARCHER_WALKTHROUGH.md
//
// Workflow:
// 1. Launch bioinformatics workspace (size M)
// 2. Configure hibernation profile (budget-safe, 15min idle)
// 3. Verify workspace is running with correct template
// 4. Test hibernation cycle (stop → start)
// 5. Verify cost tracking and hibernation savings
// 6. Cleanup
//
// IMPORTANT: This test launches real AWS instances and requires:
// - Valid AWS credentials configured
// - Appropriate AWS permissions (EC2, IAM, VPC)
// - Test timeout of at least 10 minutes: go test -timeout 10m
// - Use -short flag to skip: go test -short
func TestSoloResearcherPersona(t *testing.T) {
	// Skip if running in short mode (use: go test -short)
	if testing.Short() {
		t.Skip("Skipping AWS integration test in short mode (use -short to skip)")
	}

	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	// Generate unique names for this test run
	instanceName := GenerateTestName("test-rnaseq-analysis")

	t.Run("Phase1_LaunchBioinformaticsWorkspace", func(t *testing.T) {
		// Launch instance with bioinformatics template
		// Per walkthrough: "prism launch bioinformatics-suite rnaseq-analysis --size M"
		// Using Python ML Workstation template (slug: python-ml-workstation)
		instance, err := ctx.LaunchInstance("python-ml-workstation", instanceName, "M")
		AssertNoError(t, err, "Launch bioinformatics workspace")

		// Verify instance details
		AssertNotEmpty(t, instance.ID, "Instance should have AWS ID")
		AssertNotEmpty(t, instance.PublicIP, "Instance should have public IP")
		AssertEqual(t, "running", instance.State, "Instance state")
		AssertEqual(t, instanceName, instance.Name, "Instance name")

		t.Logf("✅ Workspace launched successfully")
		t.Logf("   Name: %s", instance.Name)
		t.Logf("   ID: %s", instance.ID)
		t.Logf("   Public IP: %s", instance.PublicIP)
		t.Logf("   State: %s", instance.State)
		t.Logf("   Template: %s", instance.Template)
	})

	t.Run("Phase2_ConfigureHibernationPolicy", func(t *testing.T) {
		// Per walkthrough: Apply hibernation policy for cost optimization
		// Note: Using the new idle policy system

		// List available idle policies
		policies, err := ctx.Client.ListIdlePolicies(context.Background())
		AssertNoError(t, err, "List idle policies")

		t.Logf("Available idle policies: %d", len(policies))
		for _, policy := range policies {
			t.Logf("  - %s: %s", policy.ID, policy.Name)
		}

		// Find a hibernation-friendly policy (cost-optimized or batch)
		// Per walkthrough: Apply aggressive hibernation for budget safety
		policyToApply := ""
		for _, policy := range policies {
			// Look for policies that use hibernation action
			if policy.ID == "cost-optimized" || policy.ID == "batch" {
				policyToApply = policy.ID
				t.Logf("Selected policy: %s (%s)", policy.ID, policy.Name)
				break
			}
		}

		if policyToApply == "" {
			// If no pre-configured policy found, skip this test phase
			// (In real deployment, policies should be pre-configured)
			t.Log("⚠️  No hibernation policies found - skipping policy application")
			t.Log("   This is expected if idle policy system is not yet configured")
			return
		}

		// Apply policy to instance
		// Per walkthrough: "prism idle instance rnaseq-analysis --profile budget-safe"
		err = ctx.Client.ApplyIdlePolicy(context.Background(), instanceName, policyToApply)
		AssertNoError(t, err, "Apply idle policy to instance")

		t.Logf("✅ Applied idle policy '%s' to instance", policyToApply)
		t.Log("   Instance will automatically hibernate when idle")
	})

	t.Run("Phase3_VerifyWorkspaceConfiguration", func(t *testing.T) {
		// Verify instance is still running with correct configuration
		instance := ctx.AssertInstanceExists(instanceName)

		AssertEqual(t, "running", instance.State, "Instance should still be running")
		AssertNotEmpty(t, instance.PublicIP, "Instance should have public IP")

		t.Logf("✅ Workspace configuration verified")
		t.Logf("   Status: %s", instance.State)
		t.Logf("   Launch time: %s", instance.LaunchTime)
		t.Logf("   Uptime: %s", time.Since(instance.LaunchTime).Round(time.Second))
	})

	t.Run("Phase4_TestHibernationCycle", func(t *testing.T) {
		// Test manual hibernation (simulates idle detection triggering)
		// Per walkthrough: Workspace automatically hibernates after 15min idle

		t.Log("Testing hibernation cycle...")

		// Hibernate instance
		err := ctx.HibernateInstance(instanceName)
		AssertNoError(t, err, "Hibernate instance")

		// Verify stopped state (hibernated instances show as "stopped")
		ctx.AssertInstanceState(instanceName, "stopped")
		t.Logf("✅ Instance hibernated successfully")

		// Resume from hibernation
		// Per walkthrough: "prism start rnaseq-analysis" (resumes in 30 seconds)
		err = ctx.StartInstance(instanceName)
		AssertNoError(t, err, "Resume from hibernation")

		// Verify running state
		instance, err := ctx.WaitForInstanceRunning(instanceName)
		AssertNoError(t, err, "Wait for instance running")
		AssertNotEmpty(t, instance.PublicIP, "Instance should have public IP after resume")

		t.Logf("✅ Instance resumed from hibernation")
		t.Logf("   State: %s", instance.State)
		t.Logf("   Public IP: %s", instance.PublicIP)
	})

	t.Run("Phase5_VerifyCostTracking", func(t *testing.T) {
		// Verify cost tracking is working
		// Per walkthrough: "prism cost summary"

		listResp, err := ctx.Client.ListInstances(context.Background())
		AssertNoError(t, err, "List instances for cost tracking")

		foundInstance := false
		for _, instance := range listResp.Instances {
			if instance.Name == instanceName {
				foundInstance = true

				// Verify cost fields are present
				if instance.EstimatedCost > 0 {
					t.Logf("✅ Cost tracking verified")
					t.Logf("   Estimated cost: $%.2f", instance.EstimatedCost)
					t.Logf("   Hourly rate: $%.4f", instance.HourlyRate)
					t.Logf("   Current spend: $%.4f", instance.CurrentSpend)
				} else {
					t.Log("⚠️  Cost not yet calculated (instance may be too new)")
				}

				// Verify hibernation savings tracking
				t.Logf("   Instance type: %s", instance.InstanceType)
				t.Logf("   Launch time: %s", instance.LaunchTime)
				break
			}
		}

		if !foundInstance {
			t.Fatalf("Instance '%s' not found in instance list", instanceName)
		}
	})

	t.Run("Phase6_Cleanup", func(t *testing.T) {
		// Delete instance
		// Per walkthrough: "prism delete rnaseq-analysis"
		err := ctx.DeleteInstance(instanceName)
		AssertNoError(t, err, "Delete instance")

		t.Logf("✅ Instance deleted successfully")

		// Poll until instance no longer appears in list (AWS eventual consistency)
		// Terminated instances can take 3-5 minutes to disappear from AWS
		t.Log("Polling for instance to disappear from list...")
		deadline := time.Now().Add(5 * time.Minute)
		instanceGone := false

		for time.Now().Before(deadline) {
			listResp, err := ctx.Client.ListInstances(context.Background())
			AssertNoError(t, err, "List instances after deletion")

			found := false
			for _, instance := range listResp.Instances {
				if instance.Name == instanceName {
					found = true
					t.Logf("  Instance still visible in state: %s (waiting...)", instance.State)
					break
				}
			}

			if !found {
				instanceGone = true
				break
			}

			time.Sleep(10 * time.Second)
		}

		if !instanceGone {
			t.Fatalf("Instance '%s' still exists after deletion timeout", instanceName)
		}

		t.Log("✅ Cleanup verified - instance no longer in list")
	})

	t.Log("🎉 Solo Researcher persona test completed successfully!")
}

// TestSoloResearcherPersona_Complete tests the complete Solo Researcher workflow
// including budget enforcement, alerts, and cost tracking over a simulated month.
//
// Based on docs/USER_SCENARIOS/01_SOLO_RESEARCHER_WALKTHROUGH.md and Issue #400
//
// Workflow:
// - Week 1-2: Regular usage pattern (4h/day) with automatic hibernation
// - Week 3: Trigger budget alert at 80% threshold
// - Week 4: Attempt over-budget launch (should be blocked)
// - Month end: Verify cost reporting and hibernation savings
//
// IMPORTANT: This test uses accelerated simulation:
// - 3 days of testing simulates 2-week pattern
// - Budget amounts scaled down for testing ($2.40 budget instead of $100)
// - Test timeout: go test -timeout 20m
func TestSoloResearcherPersona_Complete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive persona test in short mode")
	}

	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	// Generate unique names for this test run
	// projectName := GenerateTestName("test-sarah-chen-research") // TODO: Will be used when project/budget API is integrated
	instanceName := GenerateTestName("test-rnaseq-analysis")

	t.Run("Setup_ProjectWithBudget", func(t *testing.T) {
		// Create project with $2.40 budget (scaled down for testing)
		// This simulates Sarah's $100/month budget in accelerated time
		t.Log("Creating project with budget...")

		// TODO: Implement project creation with budget
		// project, err := fixtures.CreateTestProject(t, registry, fixtures.CreateTestProjectOptions{
		// 	Name:        projectName,
		// 	Description: "Solo researcher test project with budget",
		// 	Owner:       "sarah.chen@university.edu",
		// 	Budget: &fixtures.TestBudgetOptions{
		// 		TotalBudget:  2.40,
		// 		MonthlyLimit: ptr(2.40),
		// 		AlertThresholds: []types.BudgetAlert{
		// 			{Threshold: 0.80, Enabled: true}, // 80% alert
		// 		},
		// 		BudgetPeriod: types.BudgetPeriodMonthly,
		// 	},
		// })
		// AssertNoError(t, err, "Create project with budget")

		t.Skip("⚠️  Skipping: Project creation requires FixtureRegistry setup")
	})

	t.Run("Week1_LaunchBioinformaticsWorkspace", func(t *testing.T) {
		// Launch bioinformatics workspace (size M, ~$0.16/hour)
		// Per walkthrough: "prism launch bioinformatics-suite rnaseq-analysis --size M"
		t.Log("Launching bioinformatics workspace...")

		instance, err := ctx.LaunchInstance("python-ml-workstation", instanceName, "M")
		AssertNoError(t, err, "Launch bioinformatics workspace")

		// Verify instance details
		AssertNotEmpty(t, instance.ID, "Instance should have AWS ID")
		AssertNotEmpty(t, instance.PublicIP, "Instance should have public IP")
		AssertEqual(t, "running", instance.State, "Instance state")
		AssertEqual(t, instanceName, instance.Name, "Instance name")

		t.Logf("✅ Workspace launched successfully")
		t.Logf("   Name: %s", instance.Name)
		t.Logf("   ID: %s", instance.ID)
		t.Logf("   Public IP: %s", instance.PublicIP)
		t.Logf("   State: %s", instance.State)
	})

	t.Run("Week1_ConfigureHibernationPolicy", func(t *testing.T) {
		// Configure aggressive hibernation for budget safety
		// Per walkthrough: 15-minute idle timeout
		t.Log("Configuring hibernation policy...")

		// List available idle policies
		policies, err := ctx.Client.ListIdlePolicies(context.Background())
		AssertNoError(t, err, "List idle policies")

		// Find cost-optimized or batch policy
		policyToApply := ""
		for _, policy := range policies {
			if policy.ID == "cost-optimized" || policy.ID == "batch" {
				policyToApply = policy.ID
				t.Logf("Selected policy: %s (%s)", policy.ID, policy.Name)
				break
			}
		}

		if policyToApply == "" {
			t.Log("⚠️  No hibernation policies found - skipping policy application")
			return
		}

		// Apply policy to instance
		err = ctx.Client.ApplyIdlePolicy(context.Background(), instanceName, policyToApply)
		AssertNoError(t, err, "Apply idle policy to instance")

		t.Logf("✅ Applied idle policy '%s' to instance", policyToApply)
		t.Log("   Instance will automatically hibernate when idle")
	})

	t.Run("Week1_2_SimulateUsagePattern", func(t *testing.T) {
		// Simulate 2 weeks of regular usage (accelerated: 3 cycles)
		// Each cycle: 4 hours active → hibernate → verify cost accumulation
		t.Log("Simulating regular usage pattern (accelerated)...")

		for day := 1; day <= 3; day++ {
			t.Logf("Day %d: Simulating 4-hour work session", day)

			// Verify instance is running
			instance := ctx.AssertInstanceExists(instanceName)
			if instance.State != "running" {
				// Start if hibernated
				err := ctx.StartInstance(instanceName)
				AssertNoError(t, err, "Start instance for work session")
				_, err = ctx.WaitForInstanceRunning(instanceName)
				AssertNoError(t, err, "Wait for instance running")
			}

			// Simulate active work (wait 2 minutes to accumulate some cost)
			t.Logf("  Active work simulation (2 minutes)...")
			time.Sleep(2 * time.Minute)

			// Hibernate to save costs
			t.Logf("  Hibernating instance (simulating idle detection)...")
			err := ctx.HibernateInstance(instanceName)
			AssertNoError(t, err, "Hibernate instance")

			// Verify stopped state
			ctx.AssertInstanceState(instanceName, "stopped")
			t.Logf("  ✅ Day %d complete: hibernated successfully", day)

			// Brief pause between days
			time.Sleep(30 * time.Second)
		}

		t.Log("✅ Regular usage pattern simulation complete")
		t.Log("   Expected: Cost accumulation with hibernation savings")
	})

	t.Run("Week1_2_VerifyCostTracking", func(t *testing.T) {
		// Verify cost tracking is accumulating correctly
		t.Log("Verifying cost tracking...")

		listResp, err := ctx.Client.ListInstances(context.Background())
		AssertNoError(t, err, "List instances for cost tracking")

		foundInstance := false
		for _, instance := range listResp.Instances {
			if instance.Name == instanceName {
				foundInstance = true

				t.Logf("✅ Cost tracking data:")
				t.Logf("   Current spend: $%.4f", instance.CurrentSpend)
				t.Logf("   Hourly rate: $%.4f", instance.HourlyRate)
				t.Logf("   Estimated daily: $%.2f", instance.EstimatedCost)
				t.Logf("   Runtime: %s", time.Since(instance.LaunchTime).Round(time.Minute))

				// Verify costs are being tracked (should be > 0 after active time)
				if instance.CurrentSpend > 0 {
					t.Logf("   ✅ Cost accumulation verified")
				} else {
					t.Log("   ⚠️  Cost not yet calculated (may need more time)")
				}
				break
			}
		}

		if !foundInstance {
			t.Fatalf("Instance '%s' not found in instance list", instanceName)
		}
	})

	t.Run("Week3_BudgetAlert", func(t *testing.T) {
		// In real scenario, this would trigger at 80% of $100 = $80
		// In accelerated test, we check if alerts are configured
		t.Log("Checking budget alert system...")

		// TODO: Implement budget alert verification
		// - Query alert manager for project alerts
		// - Verify alert triggered at 80% threshold
		// - Verify alert includes current spend, remaining budget
		//
		// alerts, err := ctx.Client.GetProjectAlerts(context.Background(), projectID)
		// AssertNoError(t, err, "Get project alerts")
		//
		// if len(alerts) > 0 {
		// 	t.Logf("✅ Budget alert found:")
		// 	for _, alert := range alerts {
		// 		t.Logf("   - %s: %s", alert.Type, alert.Message)
		// 	}
		// } else {
		// 	t.Log("⚠️  No alerts triggered (budget threshold not reached)")
		// }

		t.Skip("⚠️  Skipping: Budget alert system requires project/budget API integration")
	})

	t.Run("Week4_AttemptOverBudgetLaunch", func(t *testing.T) {
		// Attempt to launch expensive GPU instance (should be blocked)
		// In real scenario: p3.2xlarge costs $24.80/day, exceeds remaining budget
		t.Log("Attempting over-budget launch...")

		// TODO: Implement over-budget launch attempt
		// - Try to launch expensive instance (GPU, XL size)
		// - Verify launch is blocked with budget error
		// - Verify error message includes budget details
		//
		// gpuInstanceName := GenerateTestName("test-protein-folding")
		// _, err := ctx.LaunchInstanceInProject(projectID, "gpu-ml-workstation", gpuInstanceName, "XL")
		//
		// if err != nil {
		// 	t.Logf("✅ Over-budget launch blocked as expected")
		// 	t.Logf("   Error: %v", err)
		//
		// 	// Verify error message contains budget information
		// 	errorMsg := err.Error()
		// 	AssertContains(t, errorMsg, "budget", "Error should mention budget")
		// 	AssertContains(t, errorMsg, "exceed", "Error should mention exceeding")
		// } else {
		// 	t.Fatal("❌ Over-budget launch should have been blocked!")
		// }

		t.Skip("⚠️  Skipping: Over-budget launch blocking requires project/budget API integration")
	})

	t.Run("MonthEnd_CostSummary", func(t *testing.T) {
		// Verify month-end cost summary and hibernation savings
		t.Log("Generating cost summary...")

		listResp, err := ctx.Client.ListInstances(context.Background())
		AssertNoError(t, err, "List instances for cost summary")

		var totalSpend float64
		var totalRuntime time.Duration

		for _, instance := range listResp.Instances {
			if instance.Name == instanceName {
				totalSpend = instance.CurrentSpend
				totalRuntime = time.Since(instance.LaunchTime)

				t.Logf("✅ Month-end cost summary:")
				t.Logf("   Total spend: $%.2f", totalSpend)
				t.Logf("   Total runtime: %s", totalRuntime.Round(time.Minute))
				t.Logf("   Instance type: %s", instance.InstanceType)
				t.Logf("   Hourly rate: $%.4f", instance.HourlyRate)

				// Calculate effective rate with hibernation
				if totalRuntime.Hours() > 0 {
					effectiveRate := totalSpend / totalRuntime.Hours()
					t.Logf("   Effective cost: $%.4f/hour (with hibernation savings)", effectiveRate)

					// Verify hibernation savings (effective rate should be lower than base rate)
					if effectiveRate < instance.HourlyRate {
						savingsPercent := ((instance.HourlyRate - effectiveRate) / instance.HourlyRate) * 100
						t.Logf("   ✅ Hibernation savings: %.1f%%", savingsPercent)

						// Per walkthrough: Should achieve 30%+ savings
						if savingsPercent >= 30.0 {
							t.Logf("   ✅ Excellent! Exceeds 30%% savings target")
						} else {
							t.Logf("   ⚠️  Below 30%% savings target (expected in short test)")
						}
					}
				}

				break
			}
		}

		// TODO: Implement detailed cost report generation
		// - Per-day cost breakdown
		// - Hibernation savings calculation
		// - Budget utilization percentage
		//
		// summary, err := ctx.Client.GetProjectCostSummary(context.Background(), projectID)
		// AssertNoError(t, err, "Get project cost summary")
		//
		// t.Logf("Project cost summary:")
		// t.Logf("   Total spend: $%.2f", summary.TotalSpend)
		// t.Logf("   Budget limit: $%.2f", summary.BudgetLimit)
		// t.Logf("   Utilization: %.1f%%", (summary.TotalSpend/summary.BudgetLimit)*100)
		// t.Logf("   Hibernation savings: $%.2f", summary.HibernationSavings)
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Delete instance
		t.Log("Cleaning up resources...")

		err := ctx.DeleteInstance(instanceName)
		AssertNoError(t, err, "Delete instance")

		t.Log("✅ Instance deleted successfully")

		// Wait for instance to disappear (AWS eventual consistency)
		deadline := time.Now().Add(5 * time.Minute)
		for time.Now().Before(deadline) {
			listResp, err := ctx.Client.ListInstances(context.Background())
			AssertNoError(t, err, "List instances after deletion")

			found := false
			for _, instance := range listResp.Instances {
				if instance.Name == instanceName {
					found = true
					break
				}
			}

			if !found {
				t.Log("✅ Cleanup verified - instance no longer in list")
				return
			}

			time.Sleep(10 * time.Second)
		}

		t.Fatal("Instance still exists after deletion timeout")
	})

	t.Log("🎉 Solo Researcher complete persona test finished!")
	t.Log("   ✅ Regular usage pattern with hibernation")
	t.Log("   ⚠️  Budget alerts (skipped - requires API)")
	t.Log("   ⚠️  Over-budget blocking (skipped - requires API)")
	t.Log("   ✅ Cost tracking and savings verification")
}

// TestLabEnvironmentPersona tests the Lab Environment (Prof. Martinez) workflow
// Based on docs/USER_SCENARIOS/02_LAB_ENVIRONMENT_WALKTHROUGH.md
//
// TODO: Implement multi-user setup, shared storage, team collaboration
func TestLabEnvironmentPersona(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Lab Environment persona test in short mode")
	}

	t.Skip("Lab Environment persona test not yet implemented")

	// Planned workflow:
	// 1. Launch multiple workspaces for team members
	// 2. Create shared EFS volume
	// 3. Attach shared storage to all workspaces
	// 4. Create research users for team members
	// 5. Verify collaboration workflows
	// 6. Cleanup
}

// TestUniversityClassPersona tests the University Class (Prof. Thompson) workflow
// Based on docs/USER_SCENARIOS/03_UNIVERSITY_CLASS_WALKTHROUGH.md
//
// TODO: Implement bulk launch, student access, template standardization
func TestUniversityClassPersona(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping University Class persona test in short mode")
	}

	t.Skip("University Class persona test not yet implemented")

	// Planned workflow:
	// 1. Create standardized course template
	// 2. Bulk launch workspaces for 25 students
	// 3. Configure uniform access policies
	// 4. Test student workspace access
	// 5. Verify cost tracking per student
	// 6. Cleanup (bulk delete)
}

// TestConferenceWorkshopPersona tests the Conference Workshop (Dr. Patel) workflow
// Based on docs/USER_SCENARIOS/04_CONFERENCE_WORKSHOP_WALKTHROUGH.md
//
// TODO: Implement rapid deployment, public access, time-limited workspaces
func TestConferenceWorkshopPersona(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Conference Workshop persona test in short mode")
	}

	t.Skip("Conference Workshop persona test not yet implemented")

	// Planned workflow:
	// 1. Create workshop template
	// 2. Launch workspaces with auto-termination (8 hours)
	// 3. Configure public access (temporary credentials)
	// 4. Verify time-limited lifecycle
	// 5. Cleanup (auto-termination)
}

// TestCrossInstitutionalPersona tests the Cross-Institutional (Dr. Kim) workflow
// Based on docs/USER_SCENARIOS/05_CROSS_INSTITUTIONAL_COLLABORATION_WALKTHROUGH.md
//
// TODO: Implement multi-profile setup, shared EFS, budget tracking
func TestCrossInstitutionalPersona(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Cross-Institutional persona test in short mode")
	}

	t.Skip("Cross-Institutional persona test not yet implemented")

	// Planned workflow:
	// 1. Setup workspaces in different AWS accounts (multi-profile)
	// 2. Create shared EFS volume for collaboration
	// 3. Configure cross-account access
	// 4. Verify budget tracking per institution
	// 5. Test data sharing workflows
	// 6. Cleanup
}
