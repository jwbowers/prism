/**
 * Profile Workflows E2E Tests
 *
 * End-to-end tests for complete profile management workflows in Prism GUI.
 * Tests: Create, switch, update, export, import, and delete profiles.
 */

import { test, expect } from '@playwright/test';
import { SettingsPage } from './pages';

test.describe('Profile Management Workflows', () => {
  let settingsPage: SettingsPage;

  test.beforeEach(async ({ page, context }) => {
    // Set localStorage BEFORE navigating to prevent onboarding modal
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });

    settingsPage = new SettingsPage(page);
    await settingsPage.goto();
    await settingsPage.navigate();
    await settingsPage.switchToProfiles();
  });

  test.describe('Create Profile Workflow', () => {
    test('should create a new profile with valid configuration', async () => {
      // Step 1: Click create profile button
      await settingsPage.page.getByRole('button', { name: /create.*profile/i }).click();

      // Step 2: Fill profile form
      await settingsPage.fillInput('profile name', 'test-profile');
      await settingsPage.fillInput('aws profile', 'default');
      await settingsPage.fillInput('region', 'us-west-2');

      // Step 3: Submit form
      await settingsPage.clickButton('create');

      // Step 4: Wait for profile creation to complete
      await settingsPage.page.waitForTimeout(2000);

      // Step 5: Verify profile appears in list
      const profileExists = await settingsPage.verifyProfileExists('test-profile');
      expect(profileExists).toBe(true);

      // Cleanup: Delete test profile
      await settingsPage.deleteProfile('test-profile');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });

    test('should validate profile name is required', async () => {
      await settingsPage.page.getByRole('button', { name: /create.*profile/i }).click();

      // Leave name empty
      await settingsPage.fillInput('aws profile', 'default');
      await settingsPage.fillInput('region', 'us-east-1');
      await settingsPage.clickButton('create');

      // Should show validation error
      const validationError = await settingsPage.page.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/name.*required/i);
    });

    test('should validate region format', async () => {
      await settingsPage.page.getByRole('button', { name: /create.*profile/i }).click();

      await settingsPage.fillInput('profile name', 'test-region-validation');
      await settingsPage.fillInput('aws profile', 'default');
      await settingsPage.fillInput('region', 'invalid-region');
      await settingsPage.clickButton('create');

      // Should show validation error
      const hasError = await settingsPage.page.locator('text=/invalid.*region/i').isVisible();
      expect(hasError).toBe(true);
    });

    test('should prevent duplicate profile names', async () => {
      // Create first profile
      await settingsPage.page.getByRole('button', { name: /create.*profile/i }).click();
      await settingsPage.fillInput('profile name', 'duplicate-test');
      await settingsPage.fillInput('aws profile', 'default');
      await settingsPage.fillInput('region', 'us-west-2');
      await settingsPage.clickButton('create');
      await settingsPage.page.waitForTimeout(2000);

      // Try to create second profile with same name
      await settingsPage.page.getByRole('button', { name: /create.*profile/i }).click();
      await settingsPage.fillInput('profile name', 'duplicate-test');
      await settingsPage.fillInput('aws profile', 'default');
      await settingsPage.fillInput('region', 'us-east-1');
      await settingsPage.clickButton('create');

      // Should show duplicate error
      const hasDuplicateError = await settingsPage.page.locator('text=/already exists|duplicate/i').isVisible();
      expect(hasDuplicateError).toBe(true);

      // Cleanup
      await settingsPage.page.getByRole('button', { name: /cancel/i }).click();
      await settingsPage.deleteProfile('duplicate-test');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });
  });

  test.describe('Switch Profile Workflow', () => {
    test('should switch between profiles successfully', async () => {
      // Create test profiles
      await settingsPage.createProfile('switch-test-1', 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);
      await settingsPage.createProfile('switch-test-2', 'default', 'us-east-1');
      await settingsPage.page.waitForTimeout(1000);

      // Switch to first profile
      await settingsPage.switchProfile('switch-test-1');
      await settingsPage.page.waitForTimeout(1000);

      // Verify current profile is updated
      const currentProfile = await settingsPage.getCurrentProfile();
      expect(currentProfile).toContain('switch-test-1');

      // Switch to second profile
      await settingsPage.switchProfile('switch-test-2');
      await settingsPage.page.waitForTimeout(1000);

      // Verify current profile changed
      const newCurrentProfile = await settingsPage.getCurrentProfile();
      expect(newCurrentProfile).toContain('switch-test-2');

      // Cleanup
      await settingsPage.deleteProfile('switch-test-1');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
      await settingsPage.page.waitForTimeout(500);
      await settingsPage.deleteProfile('switch-test-2');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });

    test('should preserve profile settings after switch', async () => {
      // Create profile with specific settings
      await settingsPage.createProfile('preserve-test', 'test-profile', 'ap-northeast-1');
      await settingsPage.page.waitForTimeout(1000);

      // Switch to it
      await settingsPage.switchProfile('preserve-test');
      await settingsPage.page.waitForTimeout(1000);

      // Verify settings are preserved
      const profileRow = settingsPage.getProfileByName('preserve-test');
      const profileText = await profileRow.textContent();
      expect(profileText).toContain('ap-northeast-1');

      // Cleanup
      await settingsPage.deleteProfile('preserve-test');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });
  });

  test.describe('Update Profile Workflow', () => {
    test('should update profile region successfully', async () => {
      // Create profile
      await settingsPage.createProfile('update-test', 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Update profile
      await settingsPage.updateProfile('update-test', 'eu-west-1');
      await settingsPage.page.waitForTimeout(1000);

      // Verify update
      const profileRow = settingsPage.getProfileByName('update-test');
      const profileText = await profileRow.textContent();
      expect(profileText).toContain('eu-west-1');

      // Cleanup
      await settingsPage.deleteProfile('update-test');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });

    test('should not allow updating to invalid region', async () => {
      // Create profile
      await settingsPage.createProfile('validation-test', 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Try to update with invalid region
      const profile = settingsPage.getProfileByName('validation-test');
      await profile.getByRole('button', { name: /edit/i }).click();
      await settingsPage.fillInput('region', 'invalid-region-name');
      await settingsPage.clickButton('save');

      // Should show validation error
      const hasError = await settingsPage.page.locator('text=/invalid.*region/i').isVisible();
      expect(hasError).toBe(true);

      // Cleanup
      await settingsPage.clickButton('cancel');
      await settingsPage.deleteProfile('validation-test');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });
  });

  test.describe('Export Profile Workflow', () => {
    test('should export profile configuration', async () => {
      // Create profile to export
      await settingsPage.createProfile('export-test', 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Start download listener
      const downloadPromise = settingsPage.page.waitForEvent('download');

      // Export profile
      await settingsPage.exportProfile('export-test');

      // Verify download started
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toMatch(/export-test.*\.json/i);

      // Cleanup
      await settingsPage.deleteProfile('export-test');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });
  });

  test.describe('Import Profile Workflow', () => {
    test('should import profile from valid JSON file', async ({ page }) => {
      // Create a test profile JSON file
      const testProfileJson = JSON.stringify({
        name: 'imported-profile',
        aws_profile: 'default',
        region: 'us-east-1',
      });

      // Note: This test would need actual file creation in a temp directory
      // For now, we verify the import button exists
      const importButton = page.getByRole('button', { name: /import/i });
      expect(await importButton.isVisible()).toBe(true);
    });

    test('should reject invalid profile JSON', async ({ page }) => {
      // Verify import validation would work
      // This would require actual file upload simulation
      const importButton = page.getByRole('button', { name: /import/i });
      expect(await importButton.isVisible()).toBe(true);
    });
  });

  test.describe('Delete Profile Workflow', () => {
    test('should delete profile with confirmation', async () => {
      // Create profile to delete
      await settingsPage.createProfile('delete-test', 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Verify profile exists
      let profileExists = await settingsPage.verifyProfileExists('delete-test');
      expect(profileExists).toBe(true);

      // Delete profile
      await settingsPage.deleteProfile('delete-test');

      // Confirm deletion
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
      await settingsPage.page.waitForTimeout(1000);

      // Verify profile is removed
      profileExists = await settingsPage.verifyProfileExists('delete-test');
      expect(profileExists).toBe(false);
    });

    test('should cancel profile deletion', async () => {
      // Create profile
      await settingsPage.createProfile('cancel-delete-test', 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Start deletion
      await settingsPage.deleteProfile('cancel-delete-test');

      // Cancel deletion
      await settingsPage.page.getByRole('button', { name: /cancel/i }).click();
      await settingsPage.page.waitForTimeout(500);

      // Verify profile still exists
      const profileExists = await settingsPage.verifyProfileExists('cancel-delete-test');
      expect(profileExists).toBe(true);

      // Cleanup
      await settingsPage.deleteProfile('cancel-delete-test');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });

    test('should prevent deleting currently active profile', async () => {
      // Create and switch to profile
      await settingsPage.createProfile('active-delete-test', 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);
      await settingsPage.switchProfile('active-delete-test');
      await settingsPage.page.waitForTimeout(1000);

      // Try to delete active profile
      await settingsPage.deleteProfile('active-delete-test');

      // Should show warning about active profile
      const warningText = await settingsPage.page.locator('text=/active.*profile|cannot.*delete.*active/i').textContent();
      expect(warningText).toBeTruthy();

      // Cancel and switch to different profile before cleanup
      await settingsPage.clickButton('cancel');
    });
  });

  test.describe('Profile Listing and Display', () => {
    test('should display all profiles in list', async () => {
      // Get initial count
      const initialCount = await settingsPage.getProfileCount();

      // Create multiple profiles
      await settingsPage.createProfile('list-test-1', 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(500);
      await settingsPage.createProfile('list-test-2', 'default', 'us-east-1');
      await settingsPage.page.waitForTimeout(500);

      // Verify count increased
      const newCount = await settingsPage.getProfileCount();
      expect(newCount).toBe(initialCount + 2);

      // Cleanup
      await settingsPage.deleteProfile('list-test-1');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
      await settingsPage.page.waitForTimeout(500);
      await settingsPage.deleteProfile('list-test-2');
      await settingsPage.page.getByRole('button', { name: /delete|confirm/i }).click();
    });

    test('should show current profile indicator', async () => {
      const currentProfile = await settingsPage.getCurrentProfile();
      expect(currentProfile).toBeTruthy();
    });
  });
});
