/**
 * Instances Page Object
 *
 * Page object for the Instances tab in Prism GUI.
 * Handles instance listing, management actions, and status monitoring.
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class InstancesPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to Instances tab (labeled "My Workspaces" in UI)
   */
  async navigate() {
    await this.navigateToTab('workspaces');
    await this.waitForLoadingComplete();
  }

  /**
   * Get all instance rows (Cloudscape Table component)
   */
  getInstanceRows(): Locator {
    return this.page.locator('[data-testid="instances-table"] tbody tr, .awsui-table tbody tr');
  }

  /**
   * Get instance row by name
   */
  getInstanceByName(name: string): Locator {
    return this.page.locator(`tr:has-text("${name}")`);
  }

  /**
   * Get instance count
   */
  async getInstanceCount(): Promise<number> {
    const rows = this.getInstanceRows();
    return await rows.count();
  }

  /**
   * Check if empty state is shown
   */
  async hasEmptyState(): Promise<boolean> {
    const emptyState = this.page.locator('[data-testid="empty-instances"], .awsui-table-empty');
    return await this.elementExists(emptyState);
  }

  /**
   * Start an instance
   */
  async startInstance(name: string) {
    const instance = this.getInstanceByName(name);
    const startButton = instance.getByRole('button', { name: /start/i });
    await startButton.click();
  }

  /**
   * Stop an instance
   */
  async stopInstance(name: string) {
    const instance = this.getInstanceByName(name);
    const stopButton = instance.getByRole('button', { name: /stop/i });
    await stopButton.click();
  }

  /**
   * Terminate an instance (with confirmation)
   */
  async terminateInstance(name: string) {
    const instance = this.getInstanceByName(name);
    const terminateButton = instance.getByRole('button', { name: /terminate/i });
    await terminateButton.click();
  }

  /**
   * Hibernate an instance
   */
  async hibernateInstance(name: string) {
    const instance = this.getInstanceByName(name);
    const hibernateButton = instance.getByRole('button', { name: /hibernate/i });
    await hibernateButton.click();
  }

  /**
   * Resume a hibernated instance
   */
  async resumeInstance(name: string) {
    const instance = this.getInstanceByName(name);
    const resumeButton = instance.getByRole('button', { name: /resume/i });
    await resumeButton.click();
  }

  /**
   * Connect to an instance (view connection info)
   */
  async connectToInstance(name: string) {
    const instance = this.getInstanceByName(name);
    const connectButton = instance.getByRole('button', { name: /connect/i });
    await connectButton.click();
  }

  /**
   * Get instance status
   */
  async getInstanceStatus(name: string): Promise<string | null> {
    const instance = this.getInstanceByName(name);
    const statusBadge = instance.locator('[data-testid="status-badge"], .awsui-badge');
    return await this.getTextContent(statusBadge);
  }

  /**
   * Filter instances by status
   */
  async filterByStatus(status: string) {
    const filterSelect = this.page.getByLabel(/filter.*status/i);
    await filterSelect.selectOption(status);
    await this.waitForLoadingComplete();
  }

  /**
   * Search instances by name
   */
  async searchInstances(query: string) {
    const searchInput = this.page.getByPlaceholder(/search.*instances/i);
    await searchInput.fill(query);
    await this.page.waitForTimeout(500); // Allow search to filter
  }

  /**
   * Refresh instance list
   */
  async refreshInstances() {
    const refreshButton = this.page.getByRole('button', { name: /refresh/i });
    await refreshButton.click();
    await this.waitForLoadingComplete();
  }

  /**
   * Open launch dialog
   */
  async openLaunchDialog() {
    const launchButton = this.page.getByRole('button', { name: /launch.*instance/i });
    await launchButton.click();
  }

  /**
   * Verify instance appears in list
   */
  async verifyInstanceExists(name: string): Promise<boolean> {
    const instance = this.getInstanceByName(name);
    return await this.elementExists(instance);
  }

  /**
   * Wait for instance status to change
   */
  async waitForInstanceStatus(name: string, expectedStatus: string, timeout: number = 30000) {
    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      const status = await this.getInstanceStatus(name);
      if (status && status.toLowerCase().includes(expectedStatus.toLowerCase())) {
        return true;
      }
      await this.page.waitForTimeout(2000); // Check every 2 seconds
    }
    return false;
  }

  /**
   * Get instance cost estimate
   */
  async getInstanceCost(name: string): Promise<string | null> {
    const instance = this.getInstanceByName(name);
    const costText = instance.locator('[data-testid="cost-estimate"]');
    return await this.getTextContent(costText);
  }
}
