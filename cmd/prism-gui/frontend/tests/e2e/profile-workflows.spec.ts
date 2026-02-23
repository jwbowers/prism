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
      const dialog = settingsPage.page.locator('[role="dialog"]:visible').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/name.*required/i);
    });

    test('should validate region format', async () => {
      const uniqueName = `region-test-${Date.now()}`;

      await settingsPage.page.getByTestId('create-profile-button').click();

      await settingsPage.page.getByTestId('profile-name-input').locator('input').fill(uniqueName);
      await settingsPage.page.getByTestId('aws-profile-input').locator('input').fill('default');
      await settingsPage.page.getByTestId('region-input').locator('input').fill('invalid-region');
      await settingsPage.clickButton('create');

      // Should show validation error
      const dialog = settingsPage.page.locator('[role="dialog"]:visible').first();
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
      const dialog = settingsPage.page.locator('[role="dialog"]:visible').first();
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

      // Create test profiles with polling to ensure they exist
      await settingsPage.createProfile(name1, 'default', 'us-west-2');
      await settingsPage.waitForProfileToExist(name1);
      await settingsPage.createProfile(name2, 'default', 'us-east-1');
      await settingsPage.waitForProfileToExist(name2);

      // Switch to first profile (polls until it becomes current)
      await settingsPage.switchProfile(name1);

      // Verify current profile is updated
      const currentProfile = await settingsPage.getCurrentProfile();
      expect(currentProfile).toContain(name1);

      // Switch to second profile (polls until it becomes current)
      await settingsPage.switchProfile(name2);

      // Verify current profile changed
      const newCurrentProfile = await settingsPage.getCurrentProfile();
      expect(newCurrentProfile).toContain(name2);

      // Cleanup - switch to AWS Default before deleting test profiles
      // (cannot delete the currently active profile)
      await settingsPage.switchProfile('AWS Default');

      // Delete test profiles - poll after each delete to ensure completion
      await settingsPage.deleteProfile(name1);
      await settingsPage.clickButton('delete');
      await settingsPage.waitForProfileToBeRemoved(name1);
      await settingsPage.deleteProfile(name2);
      await settingsPage.clickButton('delete');
      await settingsPage.waitForProfileToBeRemoved(name2);
    });

    test('should preserve profile settings after switch', async () => {
      const uniqueName = `preserve-test-${Date.now()}`;

      // Create profile with specific settings and poll until it exists
      await settingsPage.createProfile(uniqueName, 'test-profile', 'ap-northeast-1');
      await settingsPage.waitForProfileToExist(uniqueName);

      // Switch to it (polls until it becomes current)
      await settingsPage.switchProfile(uniqueName);

      // Verify settings are preserved
      const profileRow = settingsPage.getProfileByName(uniqueName);
      const profileText = await profileRow.textContent();
      expect(profileText).toContain('ap-northeast-1');

      // Cleanup - switch to AWS Default before deleting
      await settingsPage.switchProfile('AWS Default');
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
      await settingsPage.waitForProfileToBeRemoved(uniqueName);
    });
  });

  test.describe('Update Profile Workflow', () => {
    test('should update profile region successfully', async () => {
      const uniqueName = `update-test-${Date.now()}`;

      // Create profile
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');

      // Update profile
      await settingsPage.updateProfile(uniqueName, 'eu-west-1');

      // Poll until the region update is reflected in the UI
      await settingsPage.waitForProfileRegion(uniqueName, 'eu-west-1');

      // Verify update
      const profileRow = settingsPage.getProfileByName(uniqueName);
      const profileText = await profileRow.textContent();
      expect(profileText).toContain('eu-west-1');

      // Cleanup
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
      await settingsPage.waitForProfileToBeRemoved(uniqueName);
    });

    test('should not allow updating to invalid region', async () => {
      const uniqueName = `validation-test-${Date.now()}`;

      // Create profile and wait for it to exist
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.waitForProfileToExist(uniqueName);

      // Try to update with invalid region
      const profile = settingsPage.getProfileByName(uniqueName);
      await profile.getByRole('button', { name: /edit/i }).click();

      // Wait for dialog to open
      await settingsPage.page.locator('[role="dialog"]:visible').last().waitFor({ state: 'visible' });

      // Fill the region input directly (edit dialog uses edit-region-input, not region-input)
      await settingsPage.page.getByTestId('edit-region-input').locator('input').fill('invalid-region-name');
      await settingsPage.clickButton('save');

      // Should show validation error
      const dialog = settingsPage.page.locator('[role="dialog"]:visible').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/region/i);

      // Cleanup
      await settingsPage.clickButton('cancel');
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });
  });

  test.describe('Delete Profile Workflow', () => {
    test('should delete profile with confirmation', async () => {
      const uniqueName = `delete-test-${Date.now()}`;

      // Create profile to delete
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');

      // Verify profile exists
      let profileExists = await settingsPage.verifyProfileExists(uniqueName);
      expect(profileExists).toBe(true);

      // Delete profile
      await settingsPage.deleteProfile(uniqueName);

      // Confirm deletion
      await settingsPage.clickButton('delete');

      // Poll until profile is removed from the list
      await settingsPage.waitForProfileToBeRemoved(uniqueName);

      // Verify profile is removed
      profileExists = await settingsPage.verifyProfileExists(uniqueName);
      expect(profileExists).toBe(false);
    });

    test('should cancel profile deletion', async () => {
      const uniqueName = `cancel-delete-test-${Date.now()}`;

      // Create profile and wait for it to exist
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.waitForProfileToExist(uniqueName);

      // Start deletion
      await settingsPage.deleteProfile(uniqueName);

      // Cancel deletion
      await settingsPage.clickButton('cancel');
      await settingsPage.waitForDialogClose();

      // Verify profile still exists
      const profileExists = await settingsPage.verifyProfileExists(uniqueName);
      expect(profileExists).toBe(true);

      // Cleanup
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
    });

    test('should prevent deleting currently active profile', async () => {
      const uniqueName = `active-delete-test-${Date.now()}`;

      // Create and switch to profile (switchProfile polls until current)
      await settingsPage.createProfile(uniqueName, 'default', 'us-west-2');
      await settingsPage.waitForProfileToExist(uniqueName);
      await settingsPage.switchProfile(uniqueName);

      // Verify delete button is NOT present for active profile (good UX)
      const deleteButton = settingsPage.page.getByTestId(`delete-profile-${uniqueName}`);
      await expect(deleteButton).not.toBeVisible();

      // Cleanup - switch to AWS Default and delete test profile
      await settingsPage.switchProfile('AWS Default');
      await settingsPage.deleteProfile(uniqueName);
      await settingsPage.clickButton('delete');
      await settingsPage.waitForProfileToBeRemoved(uniqueName);
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
      await settingsPage.createProfile(name2, 'default', 'us-east-1');

      // Verify count increased
      const newCount = await settingsPage.getProfileCount();
      expect(newCount).toBe(initialCount + 2);

      // Cleanup - poll after each delete to ensure it completes
      await settingsPage.deleteProfile(name1);
      await settingsPage.clickButton('delete');
      await settingsPage.waitForProfileToBeRemoved(name1);

      await settingsPage.deleteProfile(name2);
      await settingsPage.clickButton('delete');
      await settingsPage.waitForProfileToBeRemoved(name2);
    });

    test('should show current profile indicator', async () => {
      const currentProfile = await settingsPage.getCurrentProfile();
      expect(currentProfile).toBeTruthy();
    });
  });
});
