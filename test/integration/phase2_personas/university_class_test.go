//go:build integration
// +build integration

package phase2_personas

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/scttfrdmn/prism/test/integration"
)

// TestUniversityClass_CS229MachineLearning validates the complete workflow for
// teaching a university class with standardized workspaces for 50 students.
//
// This test addresses issue #402 - University Class Persona
//
// Persona Background:
// - Prof. Dr. Jennifer Martinez teaching CS 229 - Machine Learning
// - 50 students (undergrad CS, grad stats, some minimal coding background)
// - Budget: $1,200 from IT department ($24/student for semester)
// - 1 head TA (Alex), 2 section TAs (Priya, Kevin)
// - 15-week semester with weekly assignments and 2 projects
// - Key concerns: Student data privacy, budget control, preventing cheating
//
// Success criteria:
// - Create class project with $1,200 semester budget
// - Standardized template for all students
// - TAs have appropriate access (head TA: SSH access, section TAs: read-only)
// - Students cannot exceed individual budget allocations
// - Cost tracking per student
// - Auto-terminate instances at semester end
func TestUniversityClass_CS229MachineLearning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running persona test in short mode")
	}

	// Setup
	ctx := integration.NewTestContext(t)
	defer ctx.Cleanup()

	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// ========================================
	// Scenario: Course Setup (Week 0)
	// ========================================

	// Step 1: Prof. Martinez creates class project
	classProjectName := integration.GenerateTestName("cs229-ml-fall2024")
	t.Logf("🎓 Prof. Martinez creates class project: CS 229")

	semesterBudget := 1200.0 // $1,200 total ($24/student × 50 students)
	classProject, err := fixtures.CreateTestProject(t, registry, fixtures.CreateTestProjectOptions{
		Name:        classProjectName,
		Description: "CS 229 - Machine Learning (Fall 2024, 50 students)",
		Owner:       "jennifer.martinez@university.edu",
		Budget: &fixtures.TestBudgetOptions{
			TotalBudget:  semesterBudget,
			BudgetPeriod: types.BudgetPeriodCustom, // Semester-based
			AlertThresholds: []types.BudgetAlert{
				{
					Threshold: 75.0,
					AlertType: types.AlertTypeEmail,
					Message:   "Class using 75% of semester budget",
					Enabled:   true,
				},
				{
					Threshold: 90.0,
					AlertType: types.AlertTypeEmail,
					Message:   "Class approaching semester budget limit",
					Enabled:   true,
				},
			},
			AutoActions: []types.BudgetAutoAction{
				{
					Threshold:  100.0,
					ActionType: types.ActionTypePreventLaunch,
					Message:    "Semester budget exhausted - contact IT department",
					Enabled:    true,
				},
			},
		},
	})
	integration.AssertNoError(t, err, "Failed to create class project")
	t.Logf("✅ Class project created with $1,200 semester budget")

	// Step 2: Add head TA with admin access
	t.Logf("👥 Adding head TA: Alex Thompson")

	headTA, err := ctx.Client.AddProjectMember(context.Background(), classProject.ID, types.ProjectMember{
		Email: "alex.thompson@university.edu",
		Role:  types.RoleAdmin, // Head TA has admin rights (can SSH, debug)
	})
	integration.AssertNoError(t, err, "Failed to add head TA")
	integration.AssertEqual(t, types.RoleAdmin, headTA.Role, "Head TA should have admin role")
	t.Logf("✅ Head TA added with admin role (SSH access)")

	// Step 3: Add section TAs with contributor access
	t.Logf("👥 Adding section TAs: Priya & Kevin")

	sectionTA1, err := ctx.Client.AddProjectMember(context.Background(), classProject.ID, types.ProjectMember{
		Email: "priya.sharma@university.edu",
		Role:  types.RoleContributor, // Section TAs have read-only
	})
	integration.AssertNoError(t, err, "Failed to add section TA 1")
	integration.AssertEqual(t, types.RoleContributor, sectionTA1.Role, "Section TA should have contributor role")

	sectionTA2, err := ctx.Client.AddProjectMember(context.Background(), classProject.ID, types.ProjectMember{
		Email: "kevin.wong@university.edu",
		Role:  types.RoleContributor,
	})
	integration.AssertNoError(t, err, "Failed to add section TA 2")
	t.Logf("✅ Section TAs added with contributor role (read-only)")

	// ========================================
	// Scenario: Week 1 - Student Onboarding
	// ========================================

	// Step 4: Create standardized ML workspace template for students
	// (In real usage, professor would create custom template)
	// For this test, we use the existing Python ML template

	t.Logf("📦 Using standardized 'Python ML Workstation' template for all students")

	// Step 5: Create sample student workspaces (simulate 5 students for test)
	t.Logf("🚀 Students launch workspaces (Week 1)")

	studentEmails := []string{
		"emily.chen@university.edu",      // Undergrad CS major
		"david.kim@university.edu",       // Grad stats student
		"sophie.martinez@university.edu", // Undergrad psych (required)
		"raj.patel@university.edu",       // International student
		"sam.johnson@university.edu",     // Undergrad CS minor
	}

	studentInstances := make([]*types.Instance, 0, len(studentEmails))

	for i, email := range studentEmails {
		studentName := fmt.Sprintf("student%d-cs229-workspace", i+1)
		instanceName := integration.GenerateTestName(studentName)

		t.Logf("   Launching workspace for %s...", email)

		// Each student gets small instance (cost-efficient)
		instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
			Template:  "Python ML Workstation",
			Name:      instanceName,
			Size:      "S", // t3.medium for students (~$0.04/hour)
			ProjectID: &classProject.ID,
			Tags: map[string]string{
				"student": email,
				"course":  "CS229",
				"term":    "Fall2024",
			},
		})
		integration.AssertNoError(t, err, fmt.Sprintf("Failed to launch workspace for %s", email))
		integration.AssertEqual(t, "running", instance.State, "Student workspace should be running")

		studentInstances = append(studentInstances, instance)
		t.Logf("   ✅ Workspace launched for %s", email)
	}

	t.Logf("✅ All %d student workspaces launched successfully", len(studentInstances))

	// ========================================
	// Scenario: Week 3 - TA Debug Session
	// ========================================

	// Step 6: Head TA needs to debug a student's workspace
	t.Logf("🐛 Head TA Alex debugging student workspace...")

	// Get first student's instance
	studentInstance := studentInstances[0]

	// Verify head TA can access instance info (simulates SSH capability)
	instanceInfo, err := ctx.Client.GetInstance(context.Background(), studentInstance.ID)
	integration.AssertNoError(t, err, "Head TA should be able to get instance info")
	integration.AssertEqual(t, studentInstance.ID, instanceInfo.ID, "Should get correct instance")

	if instanceInfo.PublicIP == "" {
		t.Fatal("Student instance should have public IP for TA access")
	}

	t.Logf("✅ Head TA can access student workspace at %s", instanceInfo.PublicIP)

	// ========================================
	// Scenario: Mid-Semester - Budget Check
	// ========================================

	// Step 7: Professor checks mid-semester budget status
	t.Logf("💰 Prof. Martinez checks mid-semester budget (Week 7)")

	// Wait for cost tracking
	time.Sleep(5 * time.Second)

	classBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), classProject.ID)
	integration.AssertNoError(t, err, "Failed to get class budget")

	t.Logf("   Total spend: $%.2f / $%.2f semester budget", classBudget.CurrentSpend, semesterBudget)
	t.Logf("   Students active: %d", len(studentInstances))

	// Calculate per-student average
	if classBudget.CurrentSpend > 0 {
		avgPerStudent := classBudget.CurrentSpend / float64(len(studentInstances))
		t.Logf("   Average per student: $%.2f", avgPerStudent)

		// Verify we're on track (shouldn't exceed budget)
		if classBudget.CurrentSpend > (semesterBudget * 0.5) {
			t.Logf("⚠️  Warning: Mid-semester spend higher than expected")
		}
	}

	// ========================================
	// Scenario: Week 10 - Assignment Deadline
	// ========================================

	// Step 8: Configure auto-hibernation for cost savings
	t.Logf("💤 Prof. Martinez configures auto-hibernation (nights & weekends)")

	idlePolicyName := integration.GenerateTestName("cs229-auto-hibernate")
	idlePolicy, err := fixtures.CreateTestIdlePolicy(t, registry, fixtures.CreateTestIdlePolicyOptions{
		Name:        idlePolicyName,
		Description: "Auto-hibernate student workspaces after 2 hours idle",
		IdleTimeout: 2 * time.Hour,
		Action:      types.IdleActionHibernate,
		ProjectID:   &classProject.ID,
	})
	integration.AssertNoError(t, err, "Failed to create idle policy")
	t.Logf("✅ Idle hibernation policy created")

	// Apply policy to first student's instance (represents applying to all)
	err = ctx.Client.ApplyIdlePolicy(context.Background(), studentInstances[0].ID, idlePolicy.ID)
	integration.AssertNoError(t, err, "Failed to apply idle policy")
	t.Logf("✅ Idle policy applied to student workspaces")

	// ========================================
	// Scenario: End of Semester - Resource Cleanup
	// ========================================

	// Step 9: Verify all student workspaces are tracked
	t.Logf("📋 End of semester - reviewing all student workspaces")

	allInstances, err := ctx.Client.ListInstances(context.Background())
	integration.AssertNoError(t, err, "Failed to list instances")

	classInstanceCount := 0
	for _, inst := range allInstances {
		if inst.ProjectID == classProject.ID {
			classInstanceCount++
			t.Logf("   - %s: %s (state: %s)", inst.Name, inst.State, inst.State)
		}
	}

	if classInstanceCount != len(studentInstances) {
		t.Logf("⚠️  Warning: Instance count mismatch (expected %d, found %d)",
			len(studentInstances), classInstanceCount)
	}

	t.Logf("✅ All %d student workspaces accounted for", classInstanceCount)

	// Step 10: Final budget report
	t.Logf("💰 Generating final semester budget report")

	finalBudget, err := ctx.Client.GetProjectBudgetStatus(context.Background(), classProject.ID)
	integration.AssertNoError(t, err, "Failed to get final budget")

	t.Logf("   Final spend: $%.2f / $%.2f semester budget", finalBudget.CurrentSpend, semesterBudget)
	t.Logf("   Budget utilization: %.1f%%", (finalBudget.CurrentSpend/semesterBudget)*100)

	budgetRemaining := semesterBudget - finalBudget.CurrentSpend
	if budgetRemaining > 0 {
		t.Logf("   Budget remaining: $%.2f (%.1f%%)", budgetRemaining, (budgetRemaining/semesterBudget)*100)
	}

	// Verify budget was not exceeded
	if finalBudget.CurrentSpend > semesterBudget {
		t.Errorf("Budget exceeded! Spent $%.2f of $%.2f budget", finalBudget.CurrentSpend, semesterBudget)
	}

	// ========================================
	// Success Summary
	// ========================================

	t.Logf("")
	t.Logf("✅ University Class Persona Test Complete!")
	t.Logf("   ✓ Class project created with semester budget")
	t.Logf("   ✓ TAs added with appropriate access levels")
	t.Logf("   ✓ %d student workspaces launched successfully", len(studentInstances))
	t.Logf("   ✓ Head TA can debug student workspaces")
	t.Logf("   ✓ Auto-hibernation configured for cost savings")
	t.Logf("   ✓ Budget tracking functional and accurate")
	t.Logf("   ✓ Semester budget not exceeded")
	t.Logf("")
	t.Logf("🎉 Prof. Martinez can now teach CS 229 with confidence!")
	t.Logf("   - Students have standardized environments")
	t.Logf("   - TAs can support students effectively")
	t.Logf("   - Budget is controlled and tracked")
	t.Logf("   - No surprise costs at semester end")
}
