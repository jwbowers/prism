/**
 * Idle Detection Page Object
 *
 * Page object for the Idle Detection section in Prism GUI
 * (Settings > Advanced > Idle Detection).
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class IdlePage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to Idle Detection via Settings > Advanced > Idle Detection
   */
  async navigate() {
    await this.navigateToSettingsAdvanced('Idle Detection');
    await this.waitForIdleView();
  }

  /**
   * Wait for the Idle Detection view to be visible
   */
  async waitForIdleView() {
    await this.page.waitForSelector('text=Idle Detection', { state: 'visible', timeout: 10000 });
  }

  /**
   * Get the idle policies table
   */
  getPoliciesTable(): Locator {
    return this.page.locator('[data-testid="idle-policies-table"]');
  }

  /**
   * Click the "Idle Policies" tab
   */
  async clickPoliciesTab() {
    await this.page.getByRole('tab', { name: /idle policies/i }).click();
  }

  /**
   * Click the "Schedules" tab
   */
  async clickSchedulesTab() {
    await this.page.getByRole('tab', { name: /schedules/i }).click();
  }

  /**
   * Wait for idle policies API call
   */
  async waitForIdleData() {
    await this.waitForApiCall('/api/v1/idle/policies').catch(() => {
      // Non-critical
    });
  }

  /**
   * Get stats containers (Active Policies, Total Policies, etc.)
   */
  getStatContainer(label: string): Locator {
    return this.page.locator(`text=${label}`).first();
  }
}
