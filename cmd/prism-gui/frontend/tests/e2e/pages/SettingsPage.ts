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

    await this.page.getByTestId('profile-name-input').fill(name);
    await this.page.getByTestId('aws-profile-input').fill(awsProfile);
    await this.page.getByTestId('region-input').fill(region);
    await this.clickButton('create');
  }

  /**
   * Update profile
   */
  async updateProfile(name: string, newRegion: string) {
    await this.switchToProfiles();
    const profile = this.getProfileByName(name);
    const editButton = profile.getByTestId(`edit-profile-${name}`);
    await editButton.click();

    await this.page.getByTestId('region-input').fill(newRegion);
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
}
