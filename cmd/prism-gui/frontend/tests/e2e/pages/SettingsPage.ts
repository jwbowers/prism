/**
 * Settings Page Object
 *
 * Page object for the Settings tab in Prism GUI.
 * Handles profile management, idle policies, and configuration.
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class SettingsPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to Settings tab
   */
  async navigate() {
    await this.navigateToTab('settings');
    await this.waitForLoadingComplete();
  }

  /**
   * Switch to Profiles section
   */
  async switchToProfiles() {
    // Settings uses SideNavigation with links, not tabs
    const profilesLink = this.page.getByRole('link', { name: /^profiles$/i });
    if (await this.elementExists(profilesLink)) {
      await profilesLink.click();
      await this.waitForLoadingComplete();
    }
  }

  /**
   * Switch to Idle Policies section
   */
  async switchToIdlePolicies() {
    // Settings uses SideNavigation with links, not tabs
    const idleLink = this.page.getByRole('link', { name: /idle.*detection/i });
    if (await this.elementExists(idleLink)) {
      await idleLink.click();
      await this.waitForLoadingComplete();
    }
  }

  /**
   * Get all profile rows
   */
  getProfileRows(): Locator {
    return this.page.locator('[data-testid="profiles-table"] tbody tr');
  }

  /**
   * Get profile by name
   */
  getProfileByName(name: string): Locator {
    return this.page.locator(`[data-testid="profiles-table"] tr:has-text("${name}")`);
  }

  /**
   * Create profile
   */
  async createProfile(name: string, awsProfile: string, region: string) {
    await this.switchToProfiles();
    await this.page.getByTestId('create-profile-button').click();
    await this.page.waitForTimeout(500); // Wait for dialog to open

    // Cloudscape Input wraps input in a div, so we need to find the input inside
    await this.page.getByTestId('profile-name-input').locator('input').fill(name);
    await this.page.getByTestId('aws-profile-input').locator('input').fill(awsProfile);
    await this.page.getByTestId('region-input').locator('input').fill(region);
    await this.clickButton('create');

    // Wait for dialog to close (profile created successfully)
    await this.waitForDialogClose();
  }

  /**
   * Update profile
   */
  async updateProfile(name: string, newRegion: string) {
    await this.switchToProfiles();
    const profile = this.getProfileByName(name);
    const editButton = profile.getByTestId(`edit-profile-${name}`);
    await editButton.click();
    await this.page.waitForTimeout(500); // Wait for dialog

    // Cloudscape Input wraps input in a div
    await this.page.getByTestId('region-input').locator('input').fill(newRegion);
    await this.clickButton('save');
  }

  /**
   * Delete profile
   */
  async deleteProfile(name: string) {
    await this.switchToProfiles();
    await this.page.getByTestId(`delete-profile-${name}`).click();
  }

  /**
   * Switch profile
   */
  async switchProfile(name: string) {
    await this.switchToProfiles();
    await this.page.getByTestId(`switch-profile-${name}`).click();
  }

  /**
   * Export profile
   */
  async exportProfile(name: string) {
    await this.switchToProfiles();
    const profile = this.getProfileByName(name);
    const exportButton = profile.getByTestId(`export-profile-${name}`);
    await exportButton.click();
  }

  /**
   * Import profile
   */
  async importProfile(filePath: string) {
    await this.switchToProfiles();
    const importButton = this.page.getByRole('button', { name: /import/i });
    await importButton.click();

    // Upload file
    const fileInput = this.page.locator('input[type="file"]');
    await fileInput.setInputFiles(filePath);
    await this.clickButton('import');
  }

  /**
   * Get current profile name
   */
  async getCurrentProfile(): Promise<string | null> {
    await this.switchToProfiles();

    // Wait for profiles table to load
    await this.page.waitForSelector('[data-testid="profiles-table"]', { timeout: 5000 });

    // Wait a moment for the API call to complete and profiles to render
    await this.page.waitForTimeout(1000);

    const currentBadge = this.page.getByTestId('current-profile-badge');
    return await this.getTextContent(currentBadge);
  }

  /**
   * Get all idle policy rows
   */
  getIdlePolicyRows(): Locator {
    return this.page.locator('[data-testid="idle-policies-table"] tbody tr');
  }

  /**
   * Get idle policy by name
   */
  getIdlePolicyByName(name: string): Locator {
    return this.page.locator(`[data-testid="idle-policies-table"] tr:has-text("${name}")`);
  }

  /**
   * Create idle policy
   */
  async createIdlePolicy(name: string, idleMinutes: string, action: string) {
    await this.switchToIdlePolicies();
    const createButton = this.page.getByRole('button', { name: /create.*policy/i });
    await createButton.click();

    await this.fillInput('policy name', name);
    await this.fillInput('idle minutes', idleMinutes);
    await this.selectOption('action', action);
    await this.clickButton('create');
  }

  /**
   * Update idle policy
   */
  async updateIdlePolicy(name: string, newIdleMinutes: string) {
    await this.switchToIdlePolicies();
    const policy = this.getIdlePolicyByName(name);
    const editButton = policy.getByRole('button', { name: /edit/i });
    await editButton.click();

    await this.fillInput('idle minutes', newIdleMinutes);
    await this.clickButton('save');
  }

  /**
   * Delete idle policy
   */
  async deleteIdlePolicy(name: string) {
    await this.switchToIdlePolicies();
    const policy = this.getIdlePolicyByName(name);
    const deleteButton = policy.getByRole('button', { name: /delete/i });
    await deleteButton.click();
  }

  /**
   * Apply idle policy to instance
   */
  async applyIdlePolicy(policyName: string, instanceName: string) {
    await this.switchToIdlePolicies();
    const policy = this.getIdlePolicyByName(policyName);
    const applyButton = policy.getByRole('button', { name: /apply/i });
    await applyButton.click();

    const instanceSelect = this.page.getByLabel(/instance/i);
    await instanceSelect.selectOption(instanceName);
    await this.clickButton('apply');
  }

  /**
   * View idle history
   */
  async viewIdleHistory() {
    await this.switchToIdlePolicies();
    const historyButton = this.page.getByRole('button', { name: /history/i });
    await historyButton.click();
  }

  /**
   * Verify profile exists
   */
  async verifyProfileExists(name: string): Promise<boolean> {
    await this.switchToProfiles();
    const profile = this.getProfileByName(name);
    return await this.elementExists(profile);
  }

  /**
   * Verify idle policy exists
   */
  async verifyIdlePolicyExists(name: string): Promise<boolean> {
    await this.switchToIdlePolicies();
    const policy = this.getIdlePolicyByName(name);
    return await this.elementExists(policy);
  }

  /**
   * Get profile count
   */
  async getProfileCount(): Promise<number> {
    await this.switchToProfiles();
    return await this.getProfileRows().count();
  }

  /**
   * Get idle policy count
   */
  async getIdlePolicyCount(): Promise<number> {
    await this.switchToIdlePolicies();
    return await this.getIdlePolicyRows().count();
  }

  /**
   * Clean up test profiles by deleting profiles matching a pattern
   */
  async cleanupTestProfiles(namePattern: RegExp) {
    await this.switchToProfiles();

    // Get all profile rows
    const rows = this.getProfileRows();
    const count = await rows.count();

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i);
      const text = await row.textContent();

      if (text && namePattern.test(text)) {
        // Extract profile name from the row
        const match = text.match(/^([a-zA-Z0-9\-_]+)/);
        if (match) {
          const profileName = match[1];
          try {
            // Try to delete it
            await this.deleteProfile(profileName);
            await this.page.waitForTimeout(200);

            // Confirm deletion
            const confirmButton = this.page.getByRole('button', { name: /delete|confirm/i });
            if (await confirmButton.isVisible({ timeout: 1000 })) {
              await confirmButton.click();
              await this.page.waitForTimeout(500);
            }
          } catch (error) {
            // Profile might not be deletable (could be current), skip it
            // Try to close any dialog
            try {
              const cancelButton = this.page.getByRole('button', { name: /cancel/i });
              if (await cancelButton.isVisible({ timeout: 500 })) {
                await cancelButton.click();
              }
            } catch {
              // Ignore
            }
          }
        }
      }
    }
  }

  /**
   * Wait for any dialog to close
   */
  async waitForDialogClose(timeout: number = 5000) {
    try {
      await this.page.waitForSelector('[role="dialog"]', {
        state: 'hidden',
        timeout
      });
    } catch {
      // Dialog might already be closed
    }
  }

  /**
   * Force close any open dialogs
   */
  async forceCloseDialogs() {
    try {
      // Try multiple methods to close dialogs
      const closeSelectors = [
        'button[aria-label="Close dialog"]',
        'button[aria-label="Close modal"]',
        '.awsui_dismiss-control button',
        '[role="dialog"] button:has-text("Cancel")',
        '[role="dialog"] button:has-text("Close")'
      ];

      for (const selector of closeSelectors) {
        const button = this.page.locator(selector).first();
        if (await button.isVisible({ timeout: 500 })) {
          await button.click();
          await this.page.waitForTimeout(500);
          break;
        }
      }

      // Wait for dialog to actually close
      await this.waitForDialogClose(2000);
    } catch {
      // No dialogs to close or already closed
    }
  }
}
