//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIProfileOperations tests basic profile management operations
func TestCLIProfileOperations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	profileName := GenerateTestName("test-profile")
	renamedProfile := GenerateTestName("renamed-profile")

	t.Run("ListProfiles", func(t *testing.T) {
		result := ctx.Prism("profiles", "list")
		result.AssertSuccess(t, "profiles list should succeed")

		// Should show default profiles or empty list
		if result.ExitCode == 0 {
			t.Logf("Profiles list output: %s", result.Stdout)
		}
	})

	t.Run("ShowCurrent", func(t *testing.T) {
		result := ctx.Prism("profiles", "current")
		result.AssertSuccess(t, "profiles current should succeed")
		t.Logf("Current profile: %s", result.Stdout)
	})

	t.Run("AddProfile", func(t *testing.T) {
		result := ctx.Prism("profiles", "add", "personal", profileName,
			"--aws-profile", TestAWSProfile,
			"--region", TestAWSRegion,
		)
		result.AssertSuccess(t, "profiles add should succeed")

		// Verify profile was added
		listResult := ctx.Prism("profiles", "list")
		listResult.AssertSuccess(t, "profiles list should succeed after add")
		listResult.AssertContains(t, profileName, "should list newly added profile")
	})

	t.Run("SwitchProfile", func(t *testing.T) {
		result := ctx.Prism("profiles", "switch", profileName)
		result.AssertSuccess(t, "profiles switch should succeed")

		// Verify switch by checking current profile
		currentResult := ctx.Prism("profiles", "current")
		currentResult.AssertSuccess(t, "profiles current should succeed after switch")
		currentResult.AssertContains(t, profileName, "current profile should be the switched profile")
	})

	t.Run("ValidateProfile", func(t *testing.T) {
		result := ctx.Prism("profiles", "validate", profileName)
		// Validation may fail if AWS credentials are not configured
		// So we just check it doesn't crash
		if result.ExitCode != 0 {
			t.Logf("Profile validation failed (expected if AWS credentials not configured): %s", result.Stderr)
		} else {
			t.Logf("Profile validation succeeded: %s", result.Stdout)
		}
	})

	t.Run("UpdateProfile", func(t *testing.T) {
		// Try updating region
		result := ctx.Prism("profiles", "update", profileName,
			"--region", "us-east-1",
		)
		result.AssertSuccess(t, "profiles update should succeed")
	})

	t.Run("RenameProfile", func(t *testing.T) {
		result := ctx.Prism("profiles", "rename", profileName, renamedProfile)
		result.AssertSuccess(t, "profiles rename should succeed")

		// Verify renamed profile exists
		listResult := ctx.Prism("profiles", "list")
		listResult.AssertSuccess(t, "profiles list should succeed after rename")
		listResult.AssertContains(t, renamedProfile, "should list renamed profile")
		// Note: Profile ID stays the same after rename, only the NAME changes
	})

	t.Run("RemoveProfile", func(t *testing.T) {
		// Switch back to default profile before removing
		ctx.Prism("profiles", "switch", "default")

		// Note: profiles remove requires the PROFILE ID, which stays the same after rename
		result := ctx.Prism("profiles", "remove", profileName)
		result.AssertSuccess(t, "profiles remove should succeed")

		// Verify profile was removed
		listResult := ctx.Prism("profiles", "list")
		listResult.AssertSuccess(t, "profiles list should succeed after remove")
		listResult.AssertNotContains(t, renamedProfile, "should not list removed profile")
	})
}

// TestCLIProfileExportImport tests profile export/import functionality
func TestCLIProfileExportImport(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	profileName1 := GenerateTestName("export-profile-1")
	profileName2 := GenerateTestName("export-profile-2")
	exportFile := filepath.Join(os.TempDir(), fmt.Sprintf("prism-profiles-%d.json", os.Getpid()))
	defer os.Remove(exportFile)

	t.Run("CreateProfiles", func(t *testing.T) {
		// Create first profile
		result := ctx.Prism("profiles", "add", "personal", profileName1,
			"--aws-profile", TestAWSProfile,
			"--region", "us-west-2",
		)
		result.AssertSuccess(t, "first profile add should succeed")

		// Create second profile
		result = ctx.Prism("profiles", "add", "personal", profileName2,
			"--aws-profile", TestAWSProfile,
			"--region", "us-east-1",
		)
		result.AssertSuccess(t, "second profile add should succeed")
	})

	t.Run("ExportProfiles", func(t *testing.T) {
		result := ctx.Prism("profiles", "export", exportFile, "--format", "json")
		result.AssertSuccess(t, "profiles export should succeed")

		// Verify export file exists
		if _, err := os.Stat(exportFile); os.IsNotExist(err) {
			t.Fatalf("Export file should exist at %s", exportFile)
		}

		t.Logf("Exported profiles to: %s", exportFile)
	})

	t.Run("RemoveProfiles", func(t *testing.T) {
		// Switch to default before removing
		ctx.Prism("profiles", "switch", "default")

		// Remove both profiles
		result := ctx.Prism("profiles", "remove", profileName1)
		result.AssertSuccess(t, "remove first profile should succeed")

		result = ctx.Prism("profiles", "remove", profileName2)
		result.AssertSuccess(t, "remove second profile should succeed")
	})

	t.Run("ImportProfiles", func(t *testing.T) {
		result := ctx.Prism("profiles", "import", exportFile)
		result.AssertSuccess(t, "profiles import should succeed")

		// Verify profiles were imported
		listResult := ctx.Prism("profiles", "list")
		listResult.AssertSuccess(t, "profiles list should succeed after import")
		listResult.AssertContains(t, profileName1, "should list first imported profile")
		listResult.AssertContains(t, profileName2, "should list second imported profile")
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Clean up imported profiles
		ctx.Prism("profiles", "switch", "default")
		ctx.Prism("profiles", "remove", profileName1)
		ctx.Prism("profiles", "remove", profileName2)
	})
}

// TestCLIProfileErrorHandling tests error scenarios
func TestCLIProfileErrorHandling(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	profileName := GenerateTestName("error-test-profile")

	t.Run("AddDuplicateProfile", func(t *testing.T) {
		// Add profile
		result := ctx.Prism("profiles", "add", "personal", profileName,
			"--aws-profile", TestAWSProfile,
			"--region", TestAWSRegion,
		)
		result.AssertSuccess(t, "first profile add should succeed")

		// Try to add same profile again
		result = ctx.Prism("profiles", "add", "personal", profileName,
			"--aws-profile", TestAWSProfile,
			"--region", TestAWSRegion,
		)
		result.AssertFailure(t, "duplicate profile add should fail")

		// Cleanup
		ctx.Prism("profiles", "switch", "default")
		ctx.Prism("profiles", "remove", profileName)
	})

	t.Run("RemoveNonExistentProfile", func(t *testing.T) {
		nonExistentProfile := GenerateTestName("nonexistent")
		result := ctx.Prism("profiles", "remove", nonExistentProfile)
		result.AssertFailure(t, "removing nonexistent profile should fail")
	})

	t.Run("SwitchToNonExistentProfile", func(t *testing.T) {
		nonExistentProfile := GenerateTestName("nonexistent")
		result := ctx.Prism("profiles", "switch", nonExistentProfile)
		result.AssertFailure(t, "switching to nonexistent profile should fail")
	})

	t.Run("RenameNonExistentProfile", func(t *testing.T) {
		nonExistentProfile := GenerateTestName("nonexistent")
		newName := GenerateTestName("new-name")
		result := ctx.Prism("profiles", "rename", nonExistentProfile, newName)
		result.AssertFailure(t, "renaming nonexistent profile should fail")
	})

	t.Run("UpdateNonExistentProfile", func(t *testing.T) {
		nonExistentProfile := GenerateTestName("nonexistent")
		result := ctx.Prism("profiles", "update", nonExistentProfile, "--region", "us-west-1")
		result.AssertFailure(t, "updating nonexistent profile should fail")
	})

	t.Run("ValidateNonExistentProfile", func(t *testing.T) {
		nonExistentProfile := GenerateTestName("nonexistent")
		result := ctx.Prism("profiles", "validate", nonExistentProfile)
		result.AssertFailure(t, "validating nonexistent profile should fail")
	})

	t.Run("ImportNonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(os.TempDir(), fmt.Sprintf("nonexistent-%d.json", os.Getpid()))
		result := ctx.Prism("profiles", "import", nonExistentFile)
		result.AssertFailure(t, "importing from nonexistent file should fail")
	})
}

// TestCLIProfileSwitchingAffectsOperations tests that switching profiles affects subsequent operations
func TestCLIProfileSwitchingAffectsOperations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	profileName := GenerateTestName("ops-test-profile")

	t.Run("CreateProfile", func(t *testing.T) {
		result := ctx.Prism("profiles", "add", "personal", profileName,
			"--aws-profile", TestAWSProfile,
			"--region", "us-east-1", // Different region
		)
		result.AssertSuccess(t, "profile add should succeed")
	})

	t.Run("SwitchProfile", func(t *testing.T) {
		result := ctx.Prism("profiles", "switch", profileName)
		result.AssertSuccess(t, "profile switch should succeed")
	})

	t.Run("VerifyProfileUsed", func(t *testing.T) {
		// Verify current profile is the switched one
		result := ctx.Prism("profiles", "current")
		result.AssertSuccess(t, "profiles current should succeed")
		result.AssertContains(t, profileName, "should show switched profile")

		// Any subsequent operations should use this profile
		// This could be verified by checking region in instance launch, etc.
		t.Logf("Current profile after switch: %s", result.Stdout)
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Switch back to default
		result := ctx.Prism("profiles", "switch", "default")
		result.AssertSuccess(t, "switch back to default should succeed")

		// Remove test profile
		result = ctx.Prism("profiles", "remove", profileName)
		result.AssertSuccess(t, "profile remove should succeed")
	})
}

// TestCLIProfileSetupWizard tests the interactive profile setup wizard (if testable)
func TestCLIProfileSetupWizard(t *testing.T) {
	t.Skip("Interactive setup wizard requires input automation - implement with expect or similar")

	// Future: Implement with expect-like tool or mock stdin
	// For now, manual testing only
}

// TestCLIProfileRegionVariations tests profiles with different regions
func TestCLIProfileRegionVariations(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	regions := []string{"us-west-2", "us-east-1", "eu-west-1"}
	profiles := make([]string, len(regions))

	t.Run("CreateRegionalProfiles", func(t *testing.T) {
		for i, region := range regions {
			profileName := GenerateTestName(fmt.Sprintf("profile-%s", strings.ReplaceAll(region, "-", "")))
			profiles[i] = profileName

			result := ctx.Prism("profiles", "add", "personal", profileName,
				"--aws-profile", TestAWSProfile,
				"--region", region,
			)
			result.AssertSuccess(t, fmt.Sprintf("profile add for %s should succeed", region))
			t.Logf("Created profile %s for region %s", profileName, region)
		}
	})

	t.Run("ListAllProfiles", func(t *testing.T) {
		result := ctx.Prism("profiles", "list")
		result.AssertSuccess(t, "profiles list should succeed")

		// Verify all profiles are listed
		for i, profileName := range profiles {
			result.AssertContains(t, profileName, fmt.Sprintf("should list profile %d", i+1))
		}
	})

	t.Run("SwitchBetweenRegions", func(t *testing.T) {
		for _, profileName := range profiles {
			result := ctx.Prism("profiles", "switch", profileName)
			result.AssertSuccess(t, fmt.Sprintf("switch to %s should succeed", profileName))

			// Verify current profile
			currentResult := ctx.Prism("profiles", "current")
			currentResult.AssertSuccess(t, "profiles current should succeed")
			currentResult.AssertContains(t, profileName, "should show switched profile")
		}
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Switch back to default
		ctx.Prism("profiles", "switch", "default")

		// Remove all test profiles
		for _, profileName := range profiles {
			result := ctx.Prism("profiles", "remove", profileName)
			result.AssertSuccess(t, fmt.Sprintf("remove %s should succeed", profileName))
		}
	})
}

// TestCLIProfileValidation tests profile AWS credential validation
func TestCLIProfileValidation(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	validProfileName := GenerateTestName("valid-profile")
	invalidProfileName := GenerateTestName("invalid-profile")

	t.Run("ValidateValidProfile", func(t *testing.T) {
		// Create profile with test AWS credentials
		result := ctx.Prism("profiles", "add", "personal", validProfileName,
			"--aws-profile", TestAWSProfile,
			"--region", TestAWSRegion,
		)
		result.AssertSuccess(t, "profile add should succeed")

		// Validate profile
		result = ctx.Prism("profiles", "validate", validProfileName)
		// Validation may succeed or fail depending on AWS credentials
		// Just verify command doesn't crash
		t.Logf("Validation result for valid profile: exit=%d, stdout=%s, stderr=%s",
			result.ExitCode, result.Stdout, result.Stderr)

		// Cleanup
		ctx.Prism("profiles", "switch", "default")
		ctx.Prism("profiles", "remove", validProfileName)
	})

	t.Run("ValidateInvalidProfile", func(t *testing.T) {
		// Create profile with invalid AWS credentials
		result := ctx.Prism("profiles", "add", "personal", invalidProfileName,
			"--aws-profile", "nonexistent-aws-profile",
			"--region", TestAWSRegion,
		)
		result.AssertSuccess(t, "profile add should succeed even with invalid AWS profile")

		// Validate should fail
		result = ctx.Prism("profiles", "validate", invalidProfileName)
		result.AssertFailure(t, "validation should fail for invalid AWS credentials")
		t.Logf("Validation error for invalid profile: %s", result.Stderr)

		// Cleanup
		ctx.Prism("profiles", "switch", "default")
		ctx.Prism("profiles", "remove", invalidProfileName)
	})
}
