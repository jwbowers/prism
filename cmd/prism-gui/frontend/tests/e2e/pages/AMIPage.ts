/**
 * AMI Management Page Object
 *
 * Page object for the AMI Management section in Prism GUI (Settings > Advanced > AMI Management).
 * Provides navigation and interaction helpers for AMI management tests.
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class AMIPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to the AMI Management section via Settings > Advanced > AMI Management
   */
  async navigate() {
    await this.navigateToSettingsAdvanced('AMI Management');
    await this.waitForAMIView();
  }

  /**
   * Wait for the AMI Management view to be fully loaded
   */
  async waitForAMIView() {
    await this.page.waitForSelector('text=AMI Management', { state: 'visible', timeout: 10000 });
  }

  /**
   * Get the AMI count from the header counter
   */
  getHeaderCounter(): Locator {
    return this.page.locator('text=/\\(\\d+ AMIs\\)/');
  }

  /**
   * Click the "Build AMI" button
   */
  async clickBuildAMI() {
    await this.page.getByRole('button', { name: /build ami/i }).click();
  }

  /**
   * Click the "AMIs" tab
   */
  async clickAMIsTab() {
    await this.page.getByRole('tab', { name: /^amis$/i }).click();
  }

  /**
   * Click the "Builds" tab
   */
  async clickBuildsTab() {
    await this.page.getByRole('tab', { name: /build status/i }).click();
  }

  /**
   * Click the "Regional Coverage" tab
   */
  async clickRegionsTab() {
    await this.page.getByRole('tab', { name: /regional coverage/i }).click();
  }

  /**
   * Wait for the AMI images API call to complete
   */
  async waitForAMIData() {
    await this.waitForApiCall('/api/v1/ami/images').catch(() => {
      // Non-critical: test environment may not always have this endpoint
    });
  }
}
