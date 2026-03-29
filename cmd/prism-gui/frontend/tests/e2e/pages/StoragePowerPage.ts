/**
 * StoragePowerPage — Page object for v0.20.0 storage power features.
 *
 * Covers:
 *   - S3 Mounts tab (#22c)
 *   - Instance Files tab (#30b)
 *   - Storage Analytics tab (#23c)
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class StoragePowerPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /** Navigate to the Storage view. */
  async navigate(): Promise<void> {
    await this.navigateToTab('storage');
    await this.waitForLoadingComplete();
    // Wait for the Tabs component to render
    await this.page.getByRole('tab', { name: /efs/i }).waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});
  }

  /** Switch to the S3 Mounts tab. */
  async switchToS3Mounts(): Promise<void> {
    const tab = this.page.getByRole('tab', { name: /s3 mounts/i });
    await tab.waitFor({ state: 'visible', timeout: 10000 });
    const selected = await tab.getAttribute('aria-selected').catch(() => null);
    if (selected !== 'true') {
      await tab.click();
      await this.page.waitForTimeout(300);
    }
  }

  /** Switch to the Instance Files tab. */
  async switchToFiles(): Promise<void> {
    const tab = this.page.getByRole('tab', { name: /instance files/i });
    await tab.waitFor({ state: 'visible', timeout: 10000 });
    const selected = await tab.getAttribute('aria-selected').catch(() => null);
    if (selected !== 'true') {
      await tab.click();
      await this.page.waitForTimeout(300);
    }
  }

  /** Switch to the Analytics tab. */
  async switchToAnalytics(): Promise<void> {
    const tab = this.page.getByRole('tab', { name: /analytics/i });
    await tab.waitFor({ state: 'visible', timeout: 10000 });
    const selected = await tab.getAttribute('aria-selected').catch(() => null);
    if (selected !== 'true') {
      await tab.click();
      await this.page.waitForTimeout(300);
    }
  }

  /** Select an instance in the S3 Mounts instance selector. */
  async selectS3Instance(name: string): Promise<void> {
    const select = this.page.getByTestId('s3-instance-select');
    await select.click();
    await this.page.getByRole('option', { name }).click();
  }

  /** Select an instance in the Instance Files instance selector. */
  async selectFilesInstance(name: string): Promise<void> {
    const select = this.page.getByTestId('files-instance-select');
    await select.click();
    await this.page.getByRole('option', { name }).click();
  }

  /** Click the Load Mounts button to fetch S3 mounts for the selected instance. */
  async loadS3Mounts(): Promise<void> {
    await this.page.getByTestId('load-s3-mounts-button').click();
    await this.page.waitForTimeout(500);
  }

  /** Returns the rows locator for the S3 mounts table. */
  getS3MountRows(): Locator {
    return this.page.getByTestId('s3-mounts-table').locator('tbody tr');
  }

  /** Returns the rows locator for the analytics table. */
  getAnalyticsRows(): Locator {
    return this.page.getByTestId('analytics-table').locator('tbody tr');
  }

  /** Returns the rows locator for the instance files table. */
  getFileRows(): Locator {
    return this.page.getByTestId('files-table').locator('tbody tr');
  }

  /** Select a period in the Analytics period selector. */
  async selectAnalyticsPeriod(period: string): Promise<void> {
    const select = this.page.getByTestId('analytics-period-select');
    await select.click();
    await this.page.getByRole('option', { name: new RegExp(period, 'i') }).click();
  }

  /** Click the Refresh button on the Analytics tab. */
  async refreshAnalytics(): Promise<void> {
    await this.page.getByTestId('refresh-analytics-button').click();
    await this.page.waitForTimeout(500);
  }
}
