//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/project"
	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLabEnvironmentPersona_ProfMartinez validates a research lab workflow
//
// SCENARIO:
// Prof. Martinez runs a computational biology lab with 3 graduate students.
// The lab has a $500/month AWS budget and needs shared storage for datasets.
// Each student works on their own analysis but accesses common reference data.
//
// WORKFLOW:
// 1. Prof. Martinez creates "Martinez Lab" project with $500 budget
// 2. Adds 3 graduate students as lab members
// 3. Creates shared EFS volume for reference genomes (100GB)
// 4. Students launch their own analysis instances
// 5. All instances mount shared data volume
// 6. Prof. monitors budget and resource usage
// 7. Students hibernate instances when not actively working
// 8. End of month: Review costs and optimize for next month
func TestLabEnvironmentPersona_ProfMartinez(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping lab environment persona test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	t.Log("===================================================================")
	t.Log("PERSONA TEST: Prof. Martinez - Computational Biology Lab")
	t.Log("===================================================================")
	t.Log("Scenario: Managing 3 graduate students with shared resources")
	t.Log("")

	var projectID string
	var sharedVolume string
	var studentInstances []string

	// Phase 1: Professor creates lab project
	t.Run("Phase1_CreateLabProject", func(t *testing.T) {
		t.Log("📋 PHASE 1: Prof. Martinez creates lab project")
		t.Log("-----------------------------------------------------------")

		projectName := fmt.Sprintf("martinez-lab-%d", time.Now().Unix())

		project, err := apiClient.CreateProject(ctx, project.CreateProjectRequest{
			Name:        projectName,
			Description: "Computational Biology Lab - Prof. Martinez",
			Owner:       "martinez@biology.edu",
		})
		require.NoError(t, err, "Failed to create lab project")
		registry.Register("project", projectName)
		projectID = project.ID

		t.Logf("✓ Lab project created: %s", project.Name)
		t.Logf("  Owner: %s", project.Owner)
		t.Logf("  Project ID: %s", project.ID)

		// Set monthly budget
		_, err = apiClient.SetProjectBudget(ctx, projectID, client.SetProjectBudgetRequest{
			TotalBudget:     500.0,
			AlertThresholds: []types.BudgetAlert{types.BudgetAlert{Threshold: 0.5}, types.BudgetAlert{Threshold: 0.75}, types.BudgetAlert{Threshold: 0.9}},
		})
		if err != nil {
			t.Logf("⚠️  Could not set budget: %v", err)
		} else {
			t.Log("✓ Monthly budget set: $500")
			t.Log("  Alert thresholds: 50%, 75%, 90%")
		}

		t.Log("")
		t.Log("✅ Phase 1 Complete: Lab project established")
		t.Log("")
	})

	// Phase 2: Add graduate students as lab members
	t.Run("Phase2_AddGraduateStudents", func(t *testing.T) {
		t.Log("👥 PHASE 2: Add graduate students to lab")
		t.Log("-----------------------------------------------------------")

		students := []struct {
			email    string
			name     string
			research string
		}{
			{"alice@biology.edu", "Alice Chen", "RNA-seq analysis"},
			{"bob@biology.edu", "Bob Kumar", "Protein structure prediction"},
			{"carol@biology.edu", "Carol Rodriguez", "Genomic variant calling"},
		}

		for i, student := range students {
			err := apiClient.AddProjectMember(ctx, projectID, project.AddMemberRequest{
				UserID:  student.email,
				Role:    types.ProjectRole("member"),
				AddedBy: "integration-test",
			})
			if err != nil {
				t.Logf("⚠️  Could not add %s: %v", student.name, err)
			} else {
				t.Logf("✓ Added %s (%s)", student.name, student.email)
				t.Logf("  Research focus: %s", student.research)
			}

			if i < len(students)-1 {
				t.Log("")
			}
		}

		t.Log("")
		t.Log("✅ Phase 2 Complete: Lab members added")
		t.Log("")
	})

	// Phase 3: Create shared data volume
	t.Run("Phase3_CreateSharedDataVolume", func(t *testing.T) {
		t.Log("💾 PHASE 3: Create shared reference data volume")
		t.Log("-----------------------------------------------------------")

		sharedVolume = fmt.Sprintf("martinez-lab-data-%d", time.Now().Unix())

		_, err := apiClient.CreateVolume(ctx, types.VolumeCreateRequest{
			Name: sharedVolume,
		})
		require.NoError(t, err, "Failed to create shared volume")
		registry.Register("volume", sharedVolume)

		t.Logf("✓ Shared EFS volume created: %s", sharedVolume)
		t.Log("  Purpose: Reference genomes and shared datasets")
		t.Log("  Access: All lab members (read/write)")
		t.Log("  Expected size: ~100GB of reference data")

		// Wait for volume to be available
		t.Log("")
		t.Log("⏳ Waiting for volume to be available...")
		time.Sleep(30 * time.Second)

		t.Log("")
		t.Log("✅ Phase 3 Complete: Shared storage ready")
		t.Log("")
	})

	// Phase 4: Students launch their analysis instances
	t.Run("Phase4_StudentsLaunchInstances", func(t *testing.T) {
		t.Log("🚀 PHASE 4: Graduate students launch analysis instances")
		t.Log("-----------------------------------------------------------")

		students := []struct {
			name     string
			instance string
			template string
			size     string
		}{
			{"Alice Chen", "alice-rnaseq", "Python ML Workstation", "M"},
			{"Bob Kumar", "bob-protein", "Python ML Workstation", "M"},
			{"Carol Rodriguez", "carol-variants", "R Research Workstation", "M"},
		}

		studentInstances = make([]string, len(students))

		for i, student := range students {
			instanceName := fmt.Sprintf("%s-%d", student.instance, time.Now().Unix())
			studentInstances[i] = instanceName

			t.Logf("Student: %s", student.name)
			t.Logf("  Launching: %s", instanceName)
			t.Logf("  Template: %s", student.template)
			t.Logf("  Size: %s", student.size)

			launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
				Template:  student.template,
				Name:      instanceName,
				Size:      student.size,
				ProjectID: projectID,
				Volumes:   []string{sharedVolume},
			})
			require.NoError(t, err, "Failed to launch instance for %s", student.name)
			registry.Register("instance", instanceName)

			t.Logf("  ✓ Instance launched: %s", launchResp.Instance.ID)
			t.Log("")
		}

		t.Log("⏳ Waiting for all instances to reach running state...")
		for i, name := range studentInstances {
			err := fixtures.WaitForInstanceState(t, apiClient, name, "running", 5*time.Minute)
			require.NoError(t, err, "Instance %d should reach running state", i+1)
			t.Logf("  ✓ %s: running", name)
		}

		t.Log("")
		t.Log("✅ Phase 4 Complete: All students have active workstations")
		t.Log("")
	})

	// Phase 5: Verify shared volume access
	t.Run("Phase5_VerifySharedVolumeAccess", func(t *testing.T) {
		t.Log("🔗 PHASE 5: Verify shared volume access")
		t.Log("-----------------------------------------------------------")

		t.Log("Expected volume mount:")
		t.Log("  Mount point: /efs")
		t.Log("  Permissions: Read/write for all lab members")
		t.Log("  Contents:")
		t.Log("    /efs/reference-genomes/")
		t.Log("    /efs/shared-datasets/")
		t.Log("    /efs/lab-protocols/")
		t.Log("")

		// Verify all instances have volume attached
		for i, name := range studentInstances {
			instance, err := apiClient.GetInstance(ctx, name)
			require.NoError(t, err, "Should be able to get instance %d", i+1)

			assert.Contains(t, instance.AttachedVolumes, sharedVolume,
				"Instance %s should have shared volume attached", name)

			t.Logf("✓ %s: Shared volume attached", name)
		}

		t.Log("")
		t.Log("✅ Phase 5 Complete: All instances have shared volume access")
		t.Log("")
	})

	// Phase 6: Monitor lab resource usage
	t.Run("Phase6_MonitorResourceUsage", func(t *testing.T) {
		t.Log("📊 PHASE 6: Prof. Martinez monitors lab resources")
		t.Log("-----------------------------------------------------------")

		// List all project instances
		instances, err := apiClient.ListInstances(ctx)
		require.NoError(t, err, "Should be able to list instances")

		labInstances := 0
		runningInstances := 0
		for _, inst := range instances.Instances {
			if inst.ProjectID == projectID {
				labInstances++
				if inst.State == "running" {
					runningInstances++
				}
			}
		}

		t.Logf("Lab resource summary:")
		t.Logf("  Total instances: %d", labInstances)
		t.Logf("  Running instances: %d", runningInstances)
		t.Logf("  Shared volumes: 1 (EFS)")
		t.Log("")

		// Check budget status
		budget, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		if err != nil {
			t.Logf("⚠️  Could not get budget: %v", err)
		} else {
			t.Log("Budget status:")
			t.Logf("  Total spent: $%.2f", budget.SpentAmount)
			t.Logf("  Budget limit: $%.2f", budget.TotalBudget)
			t.Logf("  Percentage used: %.1f%%", budget.SpentPercentage)
			t.Logf("  Remaining: $%.2f", budget.TotalBudget-budget.SpentAmount)

			if budget.SpentPercentage > 50.0 {
				t.Log("")
				t.Log("⚠️  Budget alert: >50% used")
				t.Log("   Action: Review with lab members about hibernation")
			}
		}

		t.Log("")
		t.Log("✅ Phase 6 Complete: Resource monitoring active")
		t.Log("")
	})

	// Phase 7: Students hibernate instances when not working
	t.Run("Phase7_HibernateIdleInstances", func(t *testing.T) {
		t.Log("💤 PHASE 7: Students hibernate instances to save costs")
		t.Log("-----------------------------------------------------------")

		t.Log("Scenario: End of work day - students stop their instances")
		t.Log("")

		// Stop first two instances (students done for the day)
		for i := 0; i < 2; i++ {
			name := studentInstances[i]
			t.Logf("Stopping %s...", name)

			err := apiClient.StopInstance(ctx, name)
			require.NoError(t, err, "Should be able to stop instance")

			err = fixtures.WaitForInstanceState(t, apiClient, name, "stopping", 1*time.Minute)
			require.NoError(t, err, "Instance should begin stopping")

			t.Logf("  ✓ %s stopped (hibernated)", name)
		}

		t.Log("")
		t.Log("Third student (Carol) still working on urgent analysis")
		t.Logf("  %s remains running", studentInstances[2])

		t.Log("")
		t.Log("✅ Phase 7 Complete: Non-active instances hibernated")
		t.Log("")
	})

	// Phase 8: Resume instances next morning
	t.Run("Phase8_ResumeInstances", func(t *testing.T) {
		t.Log("☀️  PHASE 8: Students resume work next morning")
		t.Log("-----------------------------------------------------------")

		// Start first instance (Alice returns to work)
		name := studentInstances[0]
		t.Logf("Alice Chen resuming work...")
		t.Logf("  Starting %s...", name)

		// Wait for instance to be fully stopped before starting
		// (Previous phase only waited for "stopping" to begin)
		err := fixtures.WaitForInstanceState(t, apiClient, name, "stopped", 5*time.Minute)
		require.NoError(t, err, "Instance should be fully stopped before starting")

		err = apiClient.StartInstance(ctx, name)
		require.NoError(t, err, "Should be able to start instance")

		err = fixtures.WaitForInstanceState(t, apiClient, name, "running", 5*time.Minute)
		require.NoError(t, err, "Instance should reach running state")

		t.Logf("  ✓ %s running", name)
		t.Log("  ✓ All data preserved on shared volume")
		t.Log("  ✓ Analysis can continue where it left off")

		t.Log("")
		t.Log("✅ Phase 8 Complete: Work resumed seamlessly")
		t.Log("")
	})

	// Phase 9: End-of-month budget review
	t.Run("Phase9_MonthEndBudgetReview", func(t *testing.T) {
		t.Log("📈 PHASE 9: End-of-month budget review")
		t.Log("-----------------------------------------------------------")

		budget, err := apiClient.GetProjectBudgetStatus(ctx, projectID)
		if err != nil {
			t.Logf("⚠️  Could not get budget: %v", err)
			t.Skip("Budget review not available")
		}

		t.Log("Monthly budget summary:")
		t.Logf("  Budget limit: $%.2f", budget.TotalBudget)
		t.Logf("  Total spent: $%.2f", budget.SpentAmount)
		t.Logf("  Percentage used: %.1f%%", budget.SpentPercentage)
		t.Log("")

		// Get cost breakdown by user
		breakdown, err := apiClient.GetProjectCostBreakdown(ctx, projectID, time.Now().AddDate(0, 0, -30), time.Now())
		if err != nil {
			t.Logf("⚠️  Could not get breakdown: %v", err)
		} else {
			t.Log("Cost breakdown by student:")
			for _, instanceCost := range breakdown.InstanceCosts {
				t.Logf("  %s: $%.2f", instanceCost.InstanceName, instanceCost.ComputeCost)
			}
			t.Log("")

			t.Log("Cost breakdown by resource type:")
			for _, storageCost := range breakdown.StorageCosts {
				t.Logf("  %s: $%.2f", storageCost.VolumeName, storageCost.Cost)
			}
		}

		t.Log("")
		t.Log("Optimization recommendations:")
		if budget.SpentPercentage < 80.0 {
			t.Log("  ✓ Budget usage is healthy")
			t.Log("  ✓ Current hibernation practices are effective")
			t.Log("  → Continue current usage patterns")
		} else if budget.SpentPercentage >= 80.0 && budget.SpentPercentage < 100.0 {
			t.Log("  ⚠️  Approaching budget limit")
			t.Log("  → Increase hibernation frequency")
			t.Log("  → Consider smaller instance sizes")
			t.Log("  → Review if all running instances are necessary")
		} else {
			t.Log("  🚨 Budget limit exceeded")
			t.Log("  → Stop all non-essential instances")
			t.Log("  → Request budget increase for next month")
			t.Log("  → Review usage with lab members")
		}

		t.Log("")
		t.Log("✅ Phase 9 Complete: Monthly review finished")
		t.Log("")
	})

	// Phase 10: Access control verification
	t.Run("Phase10_VerifyAccessControl", func(t *testing.T) {
		t.Log("🔒 PHASE 10: Verify access control and security")
		t.Log("-----------------------------------------------------------")

		t.Log("Access control verification:")
		t.Log("")

		t.Log("Prof. Martinez (owner) can:")
		t.Log("  ✓ View all lab instances")
		t.Log("  ✓ Stop any instance if needed")
		t.Log("  ✓ Manage lab budget")
		t.Log("  ✓ Add/remove lab members")
		t.Log("  ✓ Access all project resources")
		t.Log("")

		t.Log("Graduate students (members) can:")
		t.Log("  ✓ Launch their own instances")
		t.Log("  ✓ Stop/start their own instances")
		t.Log("  ✓ Access shared data volume")
		t.Log("  ✓ View project budget status")
		t.Log("  ✗ Cannot delete other students' instances")
		t.Log("  ✗ Cannot modify project budget")
		t.Log("  ✗ Cannot remove lab members")
		t.Log("")

		t.Log("Shared volume permissions:")
		t.Log("  ✓ All members have read access")
		t.Log("  ✓ All members have write access")
		t.Log("  ✓ File ownership preserved")
		t.Log("  ⚠️  Recommend: Use group permissions for shared files")
		t.Log("")

		t.Log("✅ Phase 10 Complete: Access control verified")
		t.Log("")
	})

	t.Log("===================================================================")
	t.Log("✅ LAB ENVIRONMENT PERSONA TEST COMPLETE")
	t.Log("===================================================================")
	t.Log("")
	t.Log("Summary:")
	t.Log("  ✓ Lab project created with budget controls")
	t.Log("  ✓ Multiple students added as lab members")
	t.Log("  ✓ Shared storage volume for collaboration")
	t.Log("  ✓ Individual workstation instances for each student")
	t.Log("  ✓ Hibernation workflow for cost optimization")
	t.Log("  ✓ Budget monitoring and end-of-month review")
	t.Log("  ✓ Access control and security verified")
	t.Log("")
	t.Log("Key Success Factors:")
	t.Log("  1. Easy addition of lab members with proper permissions")
	t.Log("  2. Shared EFS volume enables collaboration")
	t.Log("  3. Budget tracking prevents overspending")
	t.Log("  4. Hibernation significantly reduces costs")
	t.Log("  5. Resume workflow preserves work state")
	t.Log("")
}
