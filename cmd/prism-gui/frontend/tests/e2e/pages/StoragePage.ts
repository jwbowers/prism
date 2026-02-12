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
   */
  async deleteEFSVolume(name: string) {
    await this.switchToEFS();
    const volume = this.getEFSVolumeByName(name);
    const deleteButton = volume.getByRole('button', { name: /delete/i });
    await deleteButton.click();
  }

  /**
   * Delete EBS volume
   */
  async deleteEBSVolume(name: string) {
    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(name);
    const deleteButton = volume.getByRole('button', { name: /delete/i });
    await deleteButton.click();
  }

  /**
   * Mount EFS volume to instance
   */
  async mountEFSVolume(volumeName: string, instanceName: string) {
    await this.switchToEFS();
    const volume = this.getEFSVolumeByName(volumeName);
    const mountButton = volume.getByRole('button', { name: /mount/i });
    await mountButton.click();

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
    const unmountButton = volume.getByRole('button', { name: /unmount/i });
    await unmountButton.click();

    // Confirm unmount
    await this.clickButton('unmount');
  }

  /**
   * Attach EBS volume to instance
   */
  async attachEBSVolume(volumeName: string, instanceName: string) {
    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(volumeName);
    const attachButton = volume.getByRole('button', { name: /attach/i });
    await attachButton.click();

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
    const detachButton = volume.getByRole('button', { name: /detach/i });
    await detachButton.click();

    // Confirm detach
    await this.clickButton('detach');
  }

  /**
   * Search volumes
   */
  async searchVolumes(query: string) {
    const searchInput = this.page.getByPlaceholder(/search/i);
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
}
