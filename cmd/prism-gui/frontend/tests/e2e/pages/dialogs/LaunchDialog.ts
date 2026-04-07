/**
 * Launch Dialog Page Object
 *
 * Page object for the Launch Instance dialog in Prism GUI.
 * Handles the complete instance launch workflow.
 */

import { Page, Locator } from '@playwright/test';

export class LaunchDialog {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Get dialog container
   * Matches any "Launch [Template Name]" dialog (e.g., "Launch Python Machine Learning")
   * Uses flexible pattern to distinguish from "Quick Start - Launch Workspace"
   */
  getDialog(): Locator {
    // Match "Launch" followed by template name, excluding "Quick Start"
    return this.page.locator('[role="dialog"]').filter({ hasText: /^Launch\s+(?!Quick Start)/ });
  }

  /**
   * Wait for dialog to open
   */
  async waitForDialog() {
    await this.getDialog().waitFor({ state: 'visible', timeout: 10000 });
  }

  /**
   * Fill instance/workspace name
   */
  async fillInstanceName(name: string) {
    const nameInput = this.page.getByLabel(/workspace name/i);
    await nameInput.fill(name);
  }

  /**
   * Select template
   */
  async selectTemplate(templateName: string) {
    const templateSelect = this.page.getByLabel(/template/i);
    await templateSelect.selectOption({ label: templateName });
  }

  /**
   * Select instance size
   * Uses Cloudscape Select component interaction pattern
   */
  async selectSize(size: string) {
    // The cloudscape-shim Select renders as a native <select> wrapped in a div.
    // The FormField label "Workspace size" is htmlFor-associated with the native select,
    // so use getByLabel + selectOption (value-based, no dropdown click needed).
    const sizeSelect = this.page.getByLabel(/workspace size/i);
    await sizeSelect.selectOption(size);
  }

  /**
   * Select instance type
   */
  async selectInstanceType(type: string) {
    const typeSelect = this.page.getByLabel(/instance type/i);
    await typeSelect.selectOption(type);
  }

  /**
   * Select region
   */
  async selectRegion(region: string) {
    const regionSelect = this.page.getByLabel(/region/i);
    await regionSelect.selectOption(region);
  }

  /**
   * Enable Spot instance
   */
  async enableSpot() {
    const spotCheckbox = this.page.getByLabel(/spot instance/i);
    await spotCheckbox.check();
  }

  /**
   * Enable hibernation
   */
  async enableHibernation() {
    const hibernationCheckbox = this.page.getByLabel(/hibernation/i);
    await hibernationCheckbox.check();
  }

  /**
   * Select storage volume
   */
  async selectStorageVolume(volumeName: string) {
    const volumeSelect = this.page.getByLabel(/storage.*volume/i);
    await volumeSelect.selectOption(volumeName);
  }

  /**
   * Add EBS volume
   */
  async addEBSVolume(name: string, size: string) {
    const addButton = this.page.getByRole('button', { name: /add.*ebs/i });
    await addButton.click();

    await this.page.getByLabel(/ebs.*name/i).fill(name);
    await this.page.getByLabel(/ebs.*size/i).fill(size);
  }

  /**
   * Enable dry run mode
   */
  async enableDryRun() {
    const dryRunCheckbox = this.page.getByLabel(/dry.*run/i);
    await dryRunCheckbox.check();
  }

  /**
   * Click Launch button
   */
  async clickLaunch() {
    const launchButton = this.page.getByRole('button', { name: /launch workspace/i });
    await launchButton.click();
  }

  /**
   * Click Cancel button
   */
  async clickCancel() {
    const cancelButton = this.page.getByRole('button', { name: /cancel/i });
    await cancelButton.click();
  }

  /**
   * Get validation error message
   * Checks Cloudscape FormField errorText rendering and data-testid elements
   */
  async getValidationError(): Promise<string | null> {
    // Cloudscape FormField renders errorText inside [class*="error-icon"] or similar
    // Try multiple selectors for compatibility
    const selectors = [
      '[data-testid="validation-error"]',
      '[class*="form-field__error"]',
      '[class*="error__message"]',
      '.awsui-form-field__error'
    ];
    for (const selector of selectors) {
      const el = this.page.locator(selector).first();
      if (await el.isVisible().catch(() => false)) {
        return await el.textContent();
      }
    }
    // Also check for disabled Launch button + visible errorText as form hint
    const errorHint = this.page.locator('text=/name.*required/i').first();
    if (await errorHint.isVisible().catch(() => false)) {
      return await errorHint.textContent();
    }
    return null;
  }

  /**
   * Verify cost estimate is shown
   */
  async verifyCostEstimate(): Promise<boolean> {
    const costText = this.page.locator('[data-testid="cost-estimate"]');
    return await costText.isVisible();
  }

  /**
   * Get cost estimate value
   * Returns null if cost estimate is not displayed in this dialog variant
   */
  async getCostEstimate(): Promise<string | null> {
    const costText = this.page.locator('[data-testid="cost-estimate"]');
    try {
      // Use short timeout - cost estimate may not be in all dialog variants
      return await costText.textContent({ timeout: 2000 });
    } catch {
      return null;
    }
  }

  /**
   * Complete basic launch workflow
   */
  async launchInstance(instanceName: string, templateName: string) {
    await this.waitForDialog();
    await this.fillInstanceName(instanceName);
    await this.selectTemplate(templateName);
    await this.clickLaunch();
  }

  /**
   * Complete advanced launch workflow
   */
  async launchInstanceAdvanced(options: {
    name: string;
    template: string;
    size?: string;
    instanceType?: string;
    region?: string;
    spot?: boolean;
    hibernation?: boolean;
    storageVolume?: string;
    dryRun?: boolean;
  }) {
    await this.waitForDialog();
    await this.fillInstanceName(options.name);
    await this.selectTemplate(options.template);

    if (options.size) {
      await this.selectSize(options.size);
    }

    if (options.instanceType) {
      await this.selectInstanceType(options.instanceType);
    }

    if (options.region) {
      await this.selectRegion(options.region);
    }

    if (options.spot) {
      await this.enableSpot();
    }

    if (options.hibernation) {
      await this.enableHibernation();
    }

    if (options.storageVolume) {
      await this.selectStorageVolume(options.storageVolume);
    }

    if (options.dryRun) {
      await this.enableDryRun();
    }

    await this.clickLaunch();
  }

  /**
   * Verify dialog is open
   */
  async isOpen(): Promise<boolean> {
    return await this.getDialog().isVisible();
  }

  /**
   * Verify dialog is closed
   */
  async isClosed(): Promise<boolean> {
    return !(await this.getDialog().isVisible());
  }
}
