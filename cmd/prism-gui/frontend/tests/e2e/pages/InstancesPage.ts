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

    // Allow React to render the response
    await this.page.waitForLoadState('domcontentloaded');

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
   * Get the name of the first running instance
   * Returns null if no running instances are found
   * Used by mount/attach tests that require running instances with SSM
   */
  async getFirstRunningInstanceName(): Promise<string | null> {
    const rows = await this.page.locator('[data-testid="instances-table"] tbody tr').all();
    for (const row of rows) {
      const statusEl = row.locator('[data-testid="instance-status"]');
      const status = await statusEl.textContent().catch(() => '');
      if (status?.toLowerCase().includes('running')) {
        return await row.locator('[data-testid="instance-name"]').textContent();
      }
    }
    return null;
  }

  /**
   * Check if empty state is shown
   */
  async hasEmptyState(): Promise<boolean> {
    const emptyState = this.page.locator('[data-testid="empty-instances"], .awsui-table-empty');
    return await this.elementExists(emptyState);
  }

  /**
   * Click the Actions dropdown for an instance and select a menu item
   * The UI uses a ButtonDropdown component labeled "Actions" for all instance operations
   */
  private async clickInstanceAction(name: string, action: string) {
    const instance = this.getInstanceByName(name).first();
    const actionsButton = instance.getByRole('button', { name: 'Actions' });
    await actionsButton.click();
    // Cloudscape ButtonDropdown with expandToViewport renders items in a portal at page level.
    // Wait for the item to become visible (portal is added to DOM when dropdown opens),
    // then click it. Only the currently-open dropdown's items are visible at any time.
    const menuItem = this.page.getByRole('menuitem', { name: action, exact: true });
    await menuItem.waitFor({ state: 'visible', timeout: 5000 });
    await menuItem.click();
  }

  /**
   * Start an instance
   */
  async startInstance(name: string) {
    await this.clickInstanceAction(name, 'Start');
  }

  /**
   * Stop an instance
   */
  async stopInstance(name: string) {
    await this.clickInstanceAction(name, 'Stop');
  }

  /**
   * Delete/Terminate an instance (opens confirmation modal)
   * Note: The UI calls this action "Delete" in the Actions dropdown
   */
  async terminateInstance(name: string) {
    await this.clickInstanceAction(name, 'Delete');
  }

  /**
   * Hibernate an instance
   */
  async hibernateInstance(name: string) {
    await this.clickInstanceAction(name, 'Hibernate');
  }

  /**
   * Resume a hibernated instance
   */
  async resumeInstance(name: string) {
    await this.clickInstanceAction(name, 'Resume');
  }

  /**
   * Connect to an instance (view connection info)
   * Uses the dedicated connect button (data-testid) in the table row for reliability.
   * The direct button bypasses the ButtonDropdown portal which can be unreliable in tests.
   */
  async connectToInstance(name: string) {
    const connectBtn = this.page.getByTestId(`connect-btn-${name}`);
    await connectBtn.click();
  }

  /**
   * Get instance status text from the instance row
   */
  async getInstanceStatus(name: string): Promise<string | null> {
    const instance = this.getInstanceByName(name);
    // Status is in data-testid="instance-status" which contains a StatusIndicator
    // Use .first() to handle cases where duplicate rows exist (same instance in state + AWS response)
    const statusEl = instance.locator('[data-testid="instance-status"]').first();
    return await this.getTextContent(statusEl);
  }

  /**
   * Filter instances by status using the PropertyFilter component
   * Uses keyboard type to properly trigger PropertyFilter's state machine
   */
  async filterByStatus(status: string) {
    const input = this.page.getByPlaceholder(/search instances/i);
    await input.click();
    await input.clear();
    await this.page.keyboard.type(status);
    await this.page.keyboard.press('Enter');
    await this.waitForLoadingComplete();
  }

  /**
   * Search instances by name using the PropertyFilter component
   * Uses keyboard type to properly trigger PropertyFilter's state machine
   */
  async searchInstances(query: string) {
    const input = this.page.getByPlaceholder(/search instances/i);
    await input.click();
    await input.clear();
    await this.page.keyboard.type(query);
    await this.page.keyboard.press('Enter');
    await this.waitForLoadingComplete();
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

    // Wait for the dialog to become visible (Cloudscape removes awsui_hidden class)
    await this.page.getByRole('dialog').waitFor({ state: 'visible', timeout: 5000 });
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
