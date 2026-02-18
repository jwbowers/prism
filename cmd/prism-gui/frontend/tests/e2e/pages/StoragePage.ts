/**
 * Storage Page Object
 *
 * Page object for the Storage tab in Prism GUI.
 * Handles EFS and EBS storage management.
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class StoragePage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to Storage tab
   */
  async navigate() {
    await this.navigateToTab('storage');
    await this.waitForLoadingComplete();
  }

  /**
   * Switch to EFS tab
   */
  async switchToEFS() {
    const efsTab = this.page.getByRole('tab', { name: /efs/i });
    await efsTab.click();
    await this.waitForLoadingComplete();
  }

  /**
   * Switch to EBS tab
   */
  async switchToEBS() {
    const ebsTab = this.page.getByRole('tab', { name: /ebs/i });
    await ebsTab.click();
    await this.waitForLoadingComplete();
  }

  /**
   * Get all EFS volume rows
   */
  getEFSVolumeRows(): Locator {
    return this.page.locator('[data-testid="efs-table"] tbody tr');
  }

  /**
   * Get all EBS volume rows
   */
  getEBSVolumeRows(): Locator {
    return this.page.locator('[data-testid="ebs-table"] tbody tr');
  }

  /**
   * Get EFS volume by name
   */
  getEFSVolumeByName(name: string): Locator {
    return this.page.locator(`[data-testid="efs-table"] tr:has-text("${name}")`);
  }

  /**
   * Get EBS volume by name
   */
  getEBSVolumeByName(name: string): Locator {
    return this.page.locator(`[data-testid="ebs-table"] tr:has-text("${name}")`);
  }

  /**
   * Create EFS volume
   */
  async createEFSVolume(name: string) {
    await this.switchToEFS();
    const createButton = this.page.getByTestId('create-efs-header-button');
    await createButton.click();

    await this.page.getByRole('textbox', { name: 'EFS Volume Name' }).fill(name);
    await this.clickButton('create');
  }

  /**
   * Create EBS volume
   */
  async createEBSVolume(name: string, size: string) {
    await this.switchToEBS();
    const createButton = this.page.getByTestId('create-ebs-header-button');
    await createButton.click();

    await this.page.getByRole('textbox', { name: 'EBS Volume Name' }).fill(name);
    await this.page.getByRole('spinbutton', { name: 'EBS Volume Size' }).fill(size);
    await this.clickButton('create');
  }

  /**
   * Delete EFS volume
   * Waits for volume to be in "available" state before deletion (AWS requires this)
   */
  async deleteEFSVolume(name: string) {
    // Wait for volume to be available before attempting deletion
    // AWS doesn't allow deleting volumes in "creating" state
    const isAvailable = await this.waitForVolumeState(name, 'efs', 'available');
    if (!isAvailable) {
      throw new Error(`EFS volume "${name}" did not reach available state within timeout`);
    }

    await this.switchToEFS();
    const volume = this.getEFSVolumeByName(name);

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click();

    // Wait for menu to appear and click Delete option
    const deleteOption = this.page.getByRole('menuitem', { name: 'Delete', exact: true });
    await deleteOption.click();
  }

  /**
   * Delete EBS volume
   * Waits for volume to be in "available" state before deletion (AWS requires this)
   */
  async deleteEBSVolume(name: string) {
    // Wait for volume to be available before attempting deletion
    // AWS doesn't allow deleting volumes in "creating" state
    const isAvailable = await this.waitForVolumeState(name, 'ebs', 'available');
    if (!isAvailable) {
      throw new Error(`EBS volume "${name}" did not reach available state within timeout`);
    }

    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(name);

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click();

    // Wait for menu to appear and click Delete option
    const deleteOption = this.page.getByRole('menuitem', { name: 'Delete', exact: true });
    await deleteOption.click();
  }

  /**
   * Mount EFS volume to instance
   * Waits for volume to be in "available" state before mounting (AWS requires this)
   */
  async mountEFSVolume(volumeName: string, instanceName: string) {
    // Wait for volume to be available before attempting mount
    const isAvailable = await this.waitForVolumeState(volumeName, 'efs', 'available');
    if (!isAvailable) {
      throw new Error(`EFS volume "${volumeName}" did not reach available state within timeout`);
    }

    await this.switchToEFS();
    const volume = this.getEFSVolumeByName(volumeName);

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click();

    // Wait for menu to appear and click Mount option
    const mountOption = this.page.getByRole('menuitem', { name: 'Mount', exact: true });
    await mountOption.click();

    // Select instance in dialog
    const instanceSelect = this.page.getByLabel(/instance/i);
    await instanceSelect.selectOption(instanceName);
    await this.clickButton('mount');
  }

  /**
   * Unmount EFS volume from instance
   */
  async unmountEFSVolume(volumeName: string, instanceName: string) {
    await this.switchToEFS();
    const volume = this.getEFSVolumeByName(volumeName);

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click();

    // Wait for menu to appear and click Unmount option
    const unmountOption = this.page.getByRole('menuitem', { name: 'Unmount', exact: true });
    await unmountOption.click();

    // Confirm unmount
    await this.clickButton('unmount');
  }

  /**
   * Attach EBS volume to instance
   * Waits for volume to be in "available" state before attaching (AWS requires this)
   */
  async attachEBSVolume(volumeName: string, instanceName: string) {
    // Wait for volume to be available before attempting attach
    const isAvailable = await this.waitForVolumeState(volumeName, 'ebs', 'available');
    if (!isAvailable) {
      throw new Error(`EBS volume "${volumeName}" did not reach available state within timeout`);
    }

    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(volumeName);

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click();

    // Wait for menu to appear and click Attach option
    const attachOption = this.page.getByRole('menuitem', { name: 'Attach', exact: true });
    await attachOption.click();

    // Select instance in dialog
    const instanceSelect = this.page.getByLabel(/instance/i);
    await instanceSelect.selectOption(instanceName);
    await this.clickButton('attach');
  }

  /**
   * Detach EBS volume from instance
   */
  async detachEBSVolume(volumeName: string) {
    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(volumeName);

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click();

    // Wait for menu to appear and click Detach option
    const detachOption = this.page.getByRole('menuitem', { name: 'Detach', exact: true });
    await detachOption.click();

    // Confirm detach
    await this.clickButton('detach');
  }

  /**
   * Search volumes
   */
  async searchVolumes(query: string) {
    // Use data-testid for storage-specific search to avoid strict mode violations
    const searchInput = this.page.getByTestId('storage-search-input').or(
      this.page.locator('input[placeholder*="Search"]').first()
    );
    await searchInput.fill(query);
    // Wait for the table to update after search filter
    await this.page.waitForLoadState('domcontentloaded');
  }

  /**
   * Get EFS volume count
   */
  async getEFSVolumeCount(): Promise<number> {
    await this.switchToEFS();
    return await this.getEFSVolumeRows().count();
  }

  /**
   * Get EBS volume count
   */
  async getEBSVolumeCount(): Promise<number> {
    await this.switchToEBS();
    return await this.getEBSVolumeRows().count();
  }

  /**
   * Verify EFS volume exists
   */
  async verifyEFSVolumeExists(name: string): Promise<boolean> {
    await this.switchToEFS();
    const volume = this.getEFSVolumeByName(name);
    return await this.elementExists(volume);
  }

  /**
   * Verify EBS volume exists
   */
  async verifyEBSVolumeExists(name: string): Promise<boolean> {
    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(name);
    return await this.elementExists(volume);
  }

  /**
   * Wait for EFS volume to appear (deterministic DOM polling)
   * AWS EFS creation timing varies by region and load (typically 30-180+ seconds)
   * Uses Playwright's built-in waitFor() with test-level timeout (no hard-coded limits)
   * Relies on storage state monitor to update volume state from AWS
   */
  async waitForEFSVolumeToExist(name: string): Promise<boolean> {
    await this.switchToEFS();
    const volume = this.getEFSVolumeByName(name);
    try {
      // No timeout specified - uses test's configured timeout (e.g., 180s)
      await volume.waitFor({ state: 'visible' });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Wait for EBS volume to appear (deterministic DOM polling)
   * AWS EBS creation timing varies by region and load (typically 60-180+ seconds)
   * Uses Playwright's built-in waitFor() with test-level timeout (no hard-coded limits)
   * Relies on storage state monitor to update volume state from AWS
   */
  async waitForEBSVolumeToExist(name: string): Promise<boolean> {
    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(name);
    try {
      // No timeout specified - uses test's configured timeout (e.g., 180s)
      await volume.waitFor({ state: 'visible' });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get volume status
   */
  async getVolumeStatus(name: string, type: 'efs' | 'ebs'): Promise<string | null> {
    if (type === 'efs') {
      await this.switchToEFS();
      const volume = this.getEFSVolumeByName(name);
      const statusBadge = volume.locator('[data-testid="status-badge"]');
      return await this.getTextContent(statusBadge);
    } else {
      await this.switchToEBS();
      const volume = this.getEBSVolumeByName(name);
      const statusBadge = volume.locator('[data-testid="status-badge"]');
      return await this.getTextContent(statusBadge);
    }
  }

  /**
   * Wait for volume to reach specific state (deterministic DOM polling)
   * AWS EFS/EBS transitions: creating → available → in-use → deleting → deleted
   * Uses Playwright's waitForFunction with test-level timeout (no hard-coded limits)
   * Relies on storage state monitor to update volume state from AWS
   */
  async waitForVolumeState(
    name: string,
    type: 'efs' | 'ebs',
    targetState: string
  ): Promise<boolean> {
    if (type === 'efs') {
      await this.switchToEFS();
    } else {
      await this.switchToEBS();
    }

    const volume = type === 'efs' ? this.getEFSVolumeByName(name) : this.getEBSVolumeByName(name);
    const statusBadge = volume.locator('[data-testid="status-badge"]');

    try {
      // Wait for status badge to contain the target state
      // No timeout specified - uses test's configured timeout
      await this.page.waitForFunction(
        (args) => {
          const { volumeName, targetState: target, volumeType } = args;
          const table = volumeType === 'efs'
            ? document.querySelector('[data-testid="efs-table"]')
            : document.querySelector('[data-testid="ebs-table"]');

          if (!table) return false;

          const rows = Array.from(table.querySelectorAll('tbody tr'));
          const volumeRow = rows.find(row => row.textContent?.includes(volumeName));

          if (!volumeRow) return false;

          const statusBadge = volumeRow.querySelector('[data-testid="status-badge"]');
          if (!statusBadge) return false;

          const statusText = statusBadge.textContent?.toLowerCase().trim() || '';
          return statusText === target.toLowerCase();
        },
        { volumeName: name, targetState, volumeType: type }
      );
      return true;
    } catch {
      return false;
    }
  }
}
