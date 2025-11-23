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

    // Force close any open dialogs from previous tests
    await settingsPage.forceCloseDialogs();

    await settingsPage.switchToProfiles();

    // Clean up any test profiles from previous runs
    try {
      await settingsPage.cleanupTestProfiles(/^test-|^duplicate-|^switch-test-|^preserve-test|^update-test|^validation-test|^export-test|^delete-test|^cancel-delete-test|^active-delete-test|^list-test-/);
    } catch (error) {
      // Cleanup failed, but we can continue
      console.log('Profile cleanup failed:', error);
    }
  });

  test.describe('Create Profile Workflow', () => {
    test('should create a new profile with valid configuration', async () => {
      // Use unique name to avoid conflicts with previous runs
      const uniqueName = `test-profile-${Date.now()}`;

      // Step 1: Click create profile button
      await settingsPage.page.getByTestId('create-profile-button').click();

      // Step 2: Fill profile form (Cloudscape Input wraps input in div, need to find input inside)
      await settingsPage.page.getByTestId('profile-name-input').locator('input').fill(uniqueName);
      await settingsPage.page.getByTestId('aws-profile-input').locator('input').fill('default');
      await settingsPage.page.getByTestId('region-input').locator('input').fill('us-west-2');

      // Step 3: Submit form
      await settingsPage.clickButton('create');

      // Step 4: Wait for dialog to close (profile created successfully)
      await settingsPage.waitForDialogClose();

      // Step 5: Verify profile appears in list
      const profileExists = await settingsPage.verifyProfileExists(uniqueName);
      expect(profileExists).toBe(true);

      // Cleanup: Delete test profile
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });

    test('should validate profile name is required', async () => {
      await settingsPage.page.getByTestId('create-profile-button').click();

      // Leave name empty
      await settingsPage.page.getByTestId('aws-profile-input').locator('input').fill('default');
      await settingsPage.page.getByTestId('region-input').locator('input').fill('us-east-1');
      await settingsPage.clickButton('create');

      // Should show validation error (get first visible dialog)
      const dialog = settingsPage.page.locator('[role="dialog"]').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/name.*required/i);
    });

    test('should validate region format', async () => {
      await settingsPage.page.getByTestId('create-profile-button').click();

      await settingsPage.page.getByTestId('profile-name-input').locator('input').fill('test-region-validation');
      await settingsPage.page.getByTestId('aws-profile-input').locator('input').fill('default');
      await settingsPage.page.getByTestId('region-input').locator('input').fill('invalid-region');
      await settingsPage.clickButton('create');

      // Should show validation error
      const dialog = settingsPage.page.locator('[role="dialog"]').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/region/i);

      // Cleanup - cancel the dialog
      await settingsPage.clickButton('cancel');
    });

    test('should prevent duplicate profile names', async () => {
      const uniqueName = `duplicate-test-${Date.now()}`;

      // Create first profile
      await settingsPage.page.getByTestId('create-profile-button').click();
      await settingsPage.page.getByTestId('profile-name-input').locator('input').fill(uniqueName);
      await settingsPage.page.getByTestId('aws-profile-input').locator('input').fill('default');
      await settingsPage.page.getByTestId('region-input').locator('input').fill('us-west-2');
      await settingsPage.clickButton('create');
      await settingsPage.waitForDialogClose();

      // Try to create second profile with same name
      await settingsPage.page.getByTestId('create-profile-button').click();
      await settingsPage.page.getByTestId('profile-name-input').locator('input').fill(uniqueName);
      await settingsPage.page.getByTestId('aws-profile-input').locator('input').fill('default');
      await settingsPage.page.getByTestId('region-input').locator('input').fill('us-east-1');
      await settingsPage.clickButton('create');

      // Should show duplicate error
      const dialog = settingsPage.page.locator('[role="dialog"]').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/already exists|duplicate/i);

      // Cleanup
      await settingsPage.clickButton('cancel');
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });
  });

  test.describe('Switch Profile Workflow', () => {
    test('should switch between profiles successfully', async () => {
      const name1 = `switch-test-1-${Date.now()}`;
      const name2 = `switch-test-2-${Date.now()}`;

      // Create test profiles
      await settingsPage.createProfile(name1, 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);
      await settingsPage.createProfile(name2, 'default', 'us-east-1');
      await settingsPage.page.waitForTimeout(1000);

      // Switch to first profile
      await settingsPage.switchProfile(name1);
      await settingsPage.page.waitForTimeout(1000);

      // Verify current profile is updated
      const currentProfile = await settingsPage.getCurrentProfile();
      expect(currentProfile).toContain(name1);

      // Switch to second profile
      await settingsPage.switchProfile(name2);
      await settingsPage.page.waitForTimeout(1000);

      // Verify current profile changed
      const newCurrentProfile = await settingsPage.getCurrentProfile();
      expect(newCurrentProfile).toContain(name2);

      // Cleanup
      await settingsPage.deleteProfile(name1);
      await settingsPage.clickButton('delete');
      await settingsPage.page.waitForTimeout(500);
      await settingsPage.deleteProfile(name2);
      await settingsPage.clickButton('delete');
    });

    test('should preserve profile settings after switch', async () => {
      const uniqueName = `preserve-test-${Date.now()}`;

      // Create profile with specific settings
      await settingsPage.createProfile(uniqueName, 'test-profile', 'ap-northeast-1');
      await settingsPage.page.waitForTimeout(1000);

      // Switch to it
      await settingsPage.switchProfile(uniqueName);
      await settingsPage.page.waitForTimeout(1000);

      // Verify settings are preserved
      const profileRow = settingsPage.getProfileByName(uniqueName);
      const profileText = await profileRow.textContent();
      expect(profileText).toContain('ap-northeast-1');

      // Cleanup
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });
  });

  test.describe('Update Profile Workflow', () => {
    test('should update profile region successfully', async () => {
      const uniqueName = `update-test-${Date.now()}`;

      // Create profile
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Update profile
      await settingsPage.updateProfile(uniqueName, 'eu-west-1');
      await settingsPage.page.waitForTimeout(1000);

      // Verify update
      const profileRow = settingsPage.getProfileByName(uniqueName);
      const profileText = await profileRow.textContent();
      expect(profileText).toContain('eu-west-1');

      // Cleanup
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });

    test('should not allow updating to invalid region', async () => {
      const uniqueName = `validation-test-${Date.now()}`;

      // Create profile
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Try to update with invalid region
      const profile = settingsPage.getProfileByName(uniqueName);
      await profile.getByRole('button', { name: /edit/i }).click();
      await settingsPage.page.waitForTimeout(500);

      // Fill the region input directly
      await settingsPage.page.getByTestId('region-input').locator('input').fill('invalid-region-name');
      await settingsPage.clickButton('save');

      // Should show validation error
      const dialog = settingsPage.page.locator('[role="dialog"]').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/region/i);

      // Cleanup
      await settingsPage.clickButton('cancel');
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });
  });

  test.describe('Export Profile Workflow', () => {
    test.skip('should export profile configuration', async () => {
      const uniqueName = `export-test-${Date.now()}`;

      // Create profile to export
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Start download listener
      const downloadPromise = settingsPage.page.waitForEvent('download');

      // Export profile
      await settingsPage.exportProfile(uniqueName);

      // Verify download started
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toMatch(new RegExp(`${uniqueName}.*\\.json`, 'i'));

      // Cleanup
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });
  });

  test.describe('Import Profile Workflow', () => {
    test.skip('should import profile from valid JSON file', async ({ page }) => {
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

    test.skip('should reject invalid profile JSON', async ({ page }) => {
      // Verify import validation would work
      // This would require actual file upload simulation
      const importButton = page.getByRole('button', { name: /import/i });
      expect(await importButton.isVisible()).toBe(true);
    });
  });

  test.describe('Delete Profile Workflow', () => {
    test('should delete profile with confirmation', async () => {
      const uniqueName = `delete-test-${Date.now()}`;

      // Create profile to delete
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Verify profile exists
      let profileExists = await settingsPage.verifyProfileExists(uniqueName);
      expect(profileExists).toBe(true);

      // Delete profile
      await settingsPage.deleteProfile(uniqueName);

      // Confirm deletion
      await settingsPage.clickButton('delete');
      await settingsPage.page.waitForTimeout(1000);

      // Verify profile is removed
      profileExists = await settingsPage.verifyProfileExists(uniqueName);
      expect(profileExists).toBe(false);
    });

    test('should cancel profile deletion', async () => {
      const uniqueName = `cancel-delete-test-${Date.now()}`;

      // Create profile
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);

      // Start deletion
      await settingsPage.deleteProfile(uniqueName);

      // Cancel deletion
      await settingsPage.clickButton('cancel');
      await settingsPage.page.waitForTimeout(500);

      // Verify profile still exists
      const profileExists = await settingsPage.verifyProfileExists(uniqueName);
      expect(profileExists).toBe(true);

      // Cleanup
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });

    test('should prevent deleting currently active profile', async () => {
      const uniqueName = `active-delete-test-${Date.now()}`;

      // Create and switch to profile
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(1000);
      await settingsPage.switchProfile(uniqueName);
      await settingsPage.page.waitForTimeout(1000);

      // Try to delete active profile
      await settingsPage.deleteProfile(uniqueName);

      // Should show warning about active profile
      const warningText = await settingsPage.page.locator('text=/active.*profile|cannot.*delete.*active/i').textContent();
      expect(warningText).toBeTruthy();

      // Cancel and switch to different profile before cleanup
      await settingsPage.clickButton('cancel');
    });
  });

  test.describe('Profile Listing and Display', () => {
    test('should display all profiles in list', async () => {
      const name1 = `list-test-1-${Date.now()}`;
      const name2 = `list-test-2-${Date.now()}`;

      // Get initial count
      const initialCount = await settingsPage.getProfileCount();

      // Create multiple profiles
      await settingsPage.createProfile(name1, 'default', 'us-west-2');
      await settingsPage.page.waitForTimeout(500);
      await settingsPage.createProfile(name2, 'default', 'us-east-1');
      await settingsPage.page.waitForTimeout(500);

      // Verify count increased
      const newCount = await settingsPage.getProfileCount();
      expect(newCount).toBe(initialCount + 2);

      // Cleanup
      await settingsPage.deleteProfile(name1);
      await settingsPage.clickButton('delete');
      await settingsPage.page.waitForTimeout(500);
      await settingsPage.deleteProfile(name2);
      await settingsPage.clickButton('delete');
    });

    test('should show current profile indicator', async () => {
      const currentProfile = await settingsPage.getCurrentProfile();
      expect(currentProfile).toBeTruthy();
    });
  });
});
