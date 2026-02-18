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

    // Wait for the Launch New Workspace button to be visible (deterministic wait)
    // This ensures the InstancesView has fully rendered
    await this.page.getByRole('button', { name: /launch.*workspace/i }).waitFor({
      state: 'visible',
      timeout: 10000
    });

    // Wait for instances table to be populated
    await this.waitForInstancesTable();
  }

  /**
   * Wait for instances table to be ready
   * This ensures the table has loaded (either with data or empty state)
   */
  async waitForInstancesTable() {
    // Wait for the instances API call to complete
    try {
      await this.page.waitForResponse(
        response => response.url().includes('/api/v1/instances') && response.status() === 200,
        { timeout: 10000 }
      );
    } catch {
      // API might have already been called, continue
    }

    // Wait a moment for React to render the response
    await this.page.waitForTimeout(500);

    // Ensure either table with rows or empty state is visible
    await this.page.waitForFunction(() => {
      const table = document.querySelector('[data-testid="instances-table"] tbody');
      const emptyState = document.querySelector('[data-testid="empty-instances"]');
      return (table && table.children.length > 0) || emptyState !== null;
    }, { timeout: 5000 }).catch(() => {
      // If neither condition is met, that's ok - tests will handle appropriately
    });
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
    // Wait for the table to update after search filter
    await this.page.waitForLoadState('domcontentloaded');
  }

  /**
   * Refresh instance list
   */
  async refreshInstances() {
    const refreshButton = this.page.getByTestId('refresh-instances-button');
    await refreshButton.click();
    await this.waitForLoadingComplete();
  }

  /**
   * Open launch dialog
   */
  async openLaunchDialog() {
    const launchButton = this.page.getByRole('button', { name: /launch.*workspace/i });
    await launchButton.click();

    // Give Cloudscape a moment to start the dialog animation
    // The dialog exists immediately but needs time to remove awsui_hidden class
    await this.page.waitForTimeout(500);
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
      // Wait for next API refresh response
      await this.page.waitForResponse(
        response => response.url().includes('/api/v1/instances'),
        { timeout: 3000 }
      ).catch(() => {
        // Timeout is acceptable, continue polling
      });
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
