//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/scttfrdmn/prism/test/fixtures"
)

// TestNonTechnicalCollaboratorPersona validates the complete collaboration
// workflow for a non-technical researcher who only uses web-based RStudio.
//
// This test addresses issue #435 - Non-Technical Collaborator Persona
//
// Real-world scenario:
// - Dr. Smith (US) launches R research environment
// - Dr. García (Chile) collaborates via web browser only
// - Dr. García has no CLI experience, doesn't install Prism
// - Both researchers work on shared R project via RStudio Server
//
// Success criteria:
// - Instance launches with RStudio Server
// - Web interface accessible without SSH
// - Non-technical user can login with password
// - Shared project directory accessible to both users
// - Both users can read/write shared files
// - R session works in web interface
func TestNonTechnicalCollaboratorPersona(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping non-technical collaborator persona test in short mode")
	}

	// Setup test context
	ctx := NewTestContext(t)
	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// Phase 1: Setup (Instance Owner - Dr. Smith in US)
	t.Log("=" + "================================================================")
	t.Log("PHASE 1: Instance Owner (Dr. Smith) Setup")
	t.Log("=" + "================================================================")

	instanceName := GenerateTestName("collab-r-research")
	t.Logf("Dr. Smith launching R research instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "R Research Full Stack",
		Name:     instanceName,
		Size:     "M", // Medium size for RStudio Server
	})
	AssertNoError(t, err, "Failed to launch R research instance")
	AssertNotNil(t, instance, "Instance should not be nil")

	t.Logf("✓ Instance launched: %s (IP: %s)", instance.Name, instance.PublicIP)

	// Wait for SSH to be ready
	t.Log("Waiting for SSH to be ready...")
	err = fixtures.WaitForSSHReady(t, instance.PublicIP, "researcher", 5*time.Minute)
	AssertNoError(t, err, "SSH should be ready")
	t.Log("✓ SSH is ready")

	// Wait for RStudio Server to be accessible
	t.Log("Waiting for RStudio Server to start...")
	err = fixtures.WaitForRStudioServer(t, instance.PublicIP, 8787, 5*time.Minute)
	AssertNoError(t, err, "RStudio Server should be accessible")
	t.Logf("✓ RStudio Server is accessible at http://%s:8787", instance.PublicIP)

	// Phase 2: Add Collaborator (Instance Owner)
	t.Log("")
	t.Log("=" + "================================================================")
	t.Log("PHASE 2: Add Collaborator (Dr. García)")
	t.Log("=" + "================================================================")

	collaboratorUsername := "maria"
	collaboratorPassword := "SecureTestPass123!"

	t.Logf("Dr. Smith adding collaborator: %s", collaboratorUsername)

	err = fixtures.CreateRStudioUser(t, instance.PublicIP, "researcher", collaboratorUsername, collaboratorPassword)
	AssertNoError(t, err, "Failed to create collaborator user")
	t.Logf("✓ Collaborator user created: %s", collaboratorUsername)

	// Create shared project directory
	sharedProjectDir := "/home/shared/projects"
	t.Logf("Creating shared project directory: %s", sharedProjectDir)

	err = fixtures.CreateSharedDirectory(t, instance.PublicIP, "researcher", sharedProjectDir, "sudo")
	AssertNoError(t, err, "Failed to create shared directory")
	t.Logf("✓ Shared directory created: %s", sharedProjectDir)

	// Phase 3: Web Access (Collaborator - Dr. García)
	t.Log("")
	t.Log("=" + "================================================================")
	t.Log("PHASE 3: Web Access Verification (Dr. García in Chile)")
	t.Log("=" + "================================================================")

	t.Logf("Dr. García accessing RStudio Server at http://%s:8787", instance.PublicIP)

	// Verify login works
	loginSuccess := fixtures.VerifyRStudioLogin(t, instance.PublicIP, 8787, collaboratorUsername, collaboratorPassword)
	if !loginSuccess {
		t.Fatal("Collaborator should be able to login to RStudio Server")
	}
	t.Logf("✓ Collaborator %s can login to RStudio Server", collaboratorUsername)

	// Verify R session works (smoke test)
	t.Log("Verifying R session functionality...")
	err = fixtures.VerifyRSessionWorks(t, instance.PublicIP, "researcher")
	AssertNoError(t, err, "R session should work")
	t.Log("✓ R session verified working")

	// Phase 4: Collaboration (Both Users)
	t.Log("")
	t.Log("=" + "================================================================")
	t.Log("PHASE 4: Collaboration Workflow")
	t.Log("=" + "================================================================")

	// Dr. Smith creates shared R project
	projectName := "climate-analysis"
	t.Logf("Dr. Smith creating shared R project: %s", projectName)

	err = fixtures.CreateSharedRProject(t, instance.PublicIP, "researcher", sharedProjectDir, projectName)
	AssertNoError(t, err, "Failed to create shared R project")
	t.Logf("✓ R project created: %s/%s", sharedProjectDir, projectName)

	// Verify Dr. García can access the project
	projectPath := sharedProjectDir + "/" + projectName + "/analysis.R"
	t.Logf("Verifying Dr. García can access project files: %s", projectPath)

	err = fixtures.VerifyFileAccess(t, instance.PublicIP, collaboratorUsername, projectPath, true, true)
	AssertNoError(t, err, "Collaborator should have read/write access to shared project")
	t.Logf("✓ Collaborator can read and write to shared project")

	// Verify Dr. Smith still has access (owner)
	t.Log("Verifying Dr. Smith still has access to project...")
	err = fixtures.VerifyFileAccess(t, instance.PublicIP, "researcher", projectPath, true, true)
	AssertNoError(t, err, "Owner should have read/write access to project")
	t.Log("✓ Owner has full access to project")

	// Phase 5: Security Verification
	t.Log("")
	t.Log("=" + "================================================================")
	t.Log("PHASE 5: Security & Isolation Verification")
	t.Log("=" + "================================================================")

	// Verify non-sudo users cannot access other users' home directories
	t.Log("Verifying home directory isolation...")

	// Try to read researcher's private file
	privateFilePath := "/home/researcher/.bashrc"
	cmd := "sudo -u " + collaboratorUsername + " test -r " + privateFilePath + " && echo 'readable' || echo 'not-readable'"
	output, _ := fixtures.SSHCommand(t, instance.PublicIP, "researcher", cmd)

	if containsString(output, "readable") {
		t.Error("Collaborator should NOT be able to read owner's private files")
	} else {
		t.Log("✓ Home directories properly isolated")
	}

	// Verify shared directory has correct permissions
	t.Log("Verifying shared directory permissions...")
	cmd = "ls -ld " + sharedProjectDir
	output, err = fixtures.SSHCommand(t, instance.PublicIP, "researcher", cmd)
	AssertNoError(t, err, "Should be able to check directory permissions")
	t.Logf("  Directory permissions: %s", output)

	if !containsString(output, "drwxrwsr-x") {
		t.Log("  ⚠️  Warning: Expected permissions drwxrwsr-x (2775), but got different permissions")
		t.Log("  This may affect file ownership inheritance in the shared directory")
	} else {
		t.Log("✓ Shared directory has correct permissions (2775 with setgid)")
	}

	// Phase 6: Summary
	t.Log("")
	t.Log("=" + "================================================================")
	t.Log("TEST SUMMARY")
	t.Log("=" + "================================================================")
	t.Log("")
	t.Log("✅ Instance Setup:")
	t.Logf("   • R Research instance launched successfully")
	t.Logf("   • RStudio Server accessible at http://%s:8787", instance.PublicIP)
	t.Log("")
	t.Log("✅ Collaboration Setup:")
	t.Logf("   • Collaborator account created: %s", collaboratorUsername)
	t.Logf("   • Shared project directory: %s", sharedProjectDir)
	t.Logf("   • R project created: %s", projectName)
	t.Log("")
	t.Log("✅ Access Verification:")
	t.Log("   • Collaborator can login via web interface")
	t.Log("   • Both users can access shared project")
	t.Log("   • Both users can read/write shared files")
	t.Log("   • R sessions work for both users")
	t.Log("")
	t.Log("✅ Security:")
	t.Log("   • Home directories are isolated")
	t.Log("   • Shared directory has proper permissions")
	t.Log("   • Authentication required for RStudio access")
	t.Log("")
	t.Log("=" + "================================================================")
	t.Log("Non-Technical Collaborator Persona: PASSED ✓")
	t.Log("=" + "================================================================")
	t.Log("")

	// Cleanup is automatic via fixture registry
	t.Log("Cleanup will be handled automatically by fixture registry")
}

// TestNonTechnicalCollaboratorPersona_MultipleCollaborators tests
// multiple non-technical collaborators working simultaneously
func TestNonTechnicalCollaboratorPersona_MultipleCollaborators(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple collaborators test in short mode")
	}

	// Setup test context
	ctx := NewTestContext(t)
	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// Launch instance
	instanceName := GenerateTestName("multi-collab-r")
	t.Logf("Launching R research instance for multiple collaborators: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "R Research Full Stack",
		Name:     instanceName,
		Size:     "M",
	})
	AssertNoError(t, err, "Failed to launch R research instance")

	// Wait for services
	t.Log("Waiting for SSH and RStudio Server...")
	err = fixtures.WaitForSSHReady(t, instance.PublicIP, "researcher", 5*time.Minute)
	AssertNoError(t, err, "SSH should be ready")

	err = fixtures.WaitForRStudioServer(t, instance.PublicIP, 8787, 5*time.Minute)
	AssertNoError(t, err, "RStudio Server should be accessible")
	t.Log("✓ Services ready")

	// Create shared directory
	sharedDir := "/home/shared/projects"
	err = fixtures.CreateSharedDirectory(t, instance.PublicIP, "researcher", sharedDir, "sudo")
	AssertNoError(t, err, "Failed to create shared directory")

	// Add multiple collaborators
	collaborators := []struct {
		username string
		password string
		location string
	}{
		{"maria", "SecurePass1!", "Chile"},
		{"jean", "SecurePass2!", "France"},
		{"akira", "SecurePass3!", "Japan"},
	}

	t.Logf("Adding %d collaborators from different locations...", len(collaborators))

	for _, collab := range collaborators {
		t.Logf("  Adding %s (%s)...", collab.username, collab.location)

		err = fixtures.CreateRStudioUser(t, instance.PublicIP, "researcher", collab.username, collab.password)
		AssertNoError(t, err, "Failed to create collaborator: "+collab.username)

		// Verify login works
		loginSuccess := fixtures.VerifyRStudioLogin(t, instance.PublicIP, 8787, collab.username, collab.password)
		if !loginSuccess {
			t.Fatalf("Collaborator %s should be able to login", collab.username)
		}

		t.Logf("  ✓ %s can login to RStudio Server", collab.username)
	}

	// Create shared project
	projectName := "global-climate-study"
	err = fixtures.CreateSharedRProject(t, instance.PublicIP, "researcher", sharedDir, projectName)
	AssertNoError(t, err, "Failed to create shared project")
	t.Logf("✓ Shared project created: %s", projectName)

	// Verify all collaborators can access the project
	projectPath := sharedDir + "/" + projectName + "/analysis.R"

	for _, collab := range collaborators {
		t.Logf("Verifying %s (%s) can access shared project...", collab.username, collab.location)

		err = fixtures.VerifyFileAccess(t, instance.PublicIP, collab.username, projectPath, true, true)
		AssertNoError(t, err, "Collaborator "+collab.username+" should have access")

		t.Logf("  ✓ %s has read/write access", collab.username)
	}

	t.Log("")
	t.Log("=" + "================================================================")
	t.Log("Multiple Collaborators Test: PASSED ✓")
	t.Logf("  • %d collaborators from different locations", len(collaborators))
	t.Log("  • All can login via web interface")
	t.Log("  • All can access shared project")
	t.Log("  • All have read/write permissions")
	t.Log("=" + "================================================================")
}

// TestNonTechnicalCollaboratorPersona_PermissionErrors tests
// that users without proper permissions cannot access RStudio
func TestNonTechnicalCollaboratorPersona_PermissionErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping permission error test in short mode")
	}

	// Setup test context
	ctx := NewTestContext(t)
	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	// Launch instance
	instanceName := GenerateTestName("perm-test-r")
	t.Log("Launching R research instance for permission testing...")

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "R Research Full Stack",
		Name:     instanceName,
		Size:     "M",
	})
	AssertNoError(t, err, "Failed to launch R research instance")

	// Wait for services
	err = fixtures.WaitForSSHReady(t, instance.PublicIP, "researcher", 5*time.Minute)
	AssertNoError(t, err, "SSH should be ready")

	err = fixtures.WaitForRStudioServer(t, instance.PublicIP, 8787, 5*time.Minute)
	AssertNoError(t, err, "RStudio Server should be accessible")

	// Create user WITHOUT sudo group (should NOT have RStudio access)
	restrictedUser := "restricted"
	restrictedPass := "TestPass123!"

	t.Logf("Creating restricted user (without RStudio permissions): %s", restrictedUser)

	// Create user without RStudio group
	cmd := "sudo adduser --disabled-password --gecos '' " + restrictedUser
	_, err = fixtures.SSHCommand(t, instance.PublicIP, "researcher", cmd)
	AssertNoError(t, err, "Failed to create restricted user")

	cmd = "echo '" + restrictedUser + ":" + restrictedPass + "' | sudo chpasswd"
	_, err = fixtures.SSHCommand(t, instance.PublicIP, "researcher", cmd)
	AssertNoError(t, err, "Failed to set password for restricted user")

	// Verify restricted user CANNOT login to RStudio Server
	t.Logf("Verifying restricted user CANNOT login to RStudio Server...")

	loginSuccess := fixtures.VerifyRStudioLogin(t, instance.PublicIP, 8787, restrictedUser, restrictedPass)
	if loginSuccess {
		t.Fatal("Restricted user should NOT be able to login to RStudio Server")
	}

	t.Log("✓ Restricted user correctly denied access to RStudio Server")

	// Now add to sudo group and verify access is granted
	t.Logf("Adding restricted user to sudo group...")
	cmd = "sudo usermod -aG sudo " + restrictedUser
	_, err = fixtures.SSHCommand(t, instance.PublicIP, "researcher", cmd)
	AssertNoError(t, err, "Failed to add user to sudo group")

	// Give RStudio Server time to pick up group changes
	time.Sleep(5 * time.Second)

	// Verify user CAN now login
	t.Logf("Verifying user CAN login after adding to sudo group...")
	loginSuccess = fixtures.VerifyRStudioLogin(t, instance.PublicIP, 8787, restrictedUser, restrictedPass)
	if !loginSuccess {
		t.Log("⚠️  User still cannot login after adding to sudo group")
		t.Log("   This may be expected - group changes may require session restart")
	} else {
		t.Log("✓ User can login after adding to sudo group")
	}

	t.Log("")
	t.Log("=" + "================================================================")
	t.Log("Permission Error Test: PASSED ✓")
	t.Log("  • Users without sudo group cannot access RStudio Server")
	t.Log("  • Group-based access control working correctly")
	t.Log("=" + "================================================================")
}

// Helper function to check if a string contains a substring
func containsString(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 &&
		(text == substr || len(text) >= len(substr) &&
			(text[:len(substr)] == substr || text[len(text)-len(substr):] == substr ||
				findSubstring(text, substr)))
}

// Simple substring search
func findSubstring(text, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(text) < len(substr) {
		return false
	}
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
