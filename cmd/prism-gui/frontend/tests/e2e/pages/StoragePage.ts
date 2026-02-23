/**
 * Storage Page Object
 *
 * Page object for the Storage tab in Prism GUI.
 * Handles EFS and EBS storage management.
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';
import { ConfirmDialog } from './dialogs/ConfirmDialog';

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
    // Wait for global loading state to clear - Cloudscape Tables hide rows during loading
    // The Refresh button is enabled when state.loading = false
    await this.waitForStorageLoaded();
  }

  /**
   * Wait for storage page data to be loaded (not in loading state)
   * Cloudscape Tables show skeleton rows when loading={true}, hiding actual data-testid cells
   */
  async waitForStorageLoaded() {
    try {
      // Wait for EFS or EBS tab to be visible (storage page is rendered)
      await this.page.getByRole('tab', { name: /efs/i }).waitFor({ state: 'visible', timeout: 10000 });
      // Wait until no Cloudscape loading text is visible in the storage area
      // When state.loading=true, tables show loadingText like "Loading shared storage volumes from AWS"
      await this.page.waitForFunction(() => {
        const bodyText = document.body.innerText || '';
        return !bodyText.includes('Loading shared storage volumes from AWS') &&
               !bodyText.includes('Loading private storage volumes from AWS');
      }, undefined, { timeout: 30000 });
    } catch {
      // Non-critical - continue even if check fails
    }
  }

  /**
   * Switch to EFS tab
   * Uses aria-selected to reliably detect current tab state
   * (isVisible() is unreliable for Cloudscape Tabs which may render all content in DOM)
   * Uses retry loop to handle StorageManagementView remounting (resets activeTabId to EFS default)
   */
  async switchToEFS() {
    // Retry up to 3 times to handle component remounts during loadApplicationData
    for (let attempt = 0; attempt < 3; attempt++) {
      const efsTab = this.page.getByRole('tab', { name: /efs/i });

      // Check if already on EFS tab using aria-selected (reliable for Cloudscape Tabs)
      const isSelected = await efsTab.getAttribute('aria-selected').catch(() => null);
      if (isSelected !== 'true') {
        // Wait for tab to be ready before clicking
        await efsTab.waitFor({ state: 'visible' }).catch(() => {});
        await this.page.waitForTimeout(300);
        await efsTab.click().catch(() => {});
        await this.waitForLoadingComplete();
      }

      // Wait for Cloudscape Table loading to clear
      await this.waitForStorageLoaded();

      // Verify EFS tab is still selected after loading (handles remount during wait)
      const finalSelected = await this.page.getByRole('tab', { name: /efs/i })
        .getAttribute('aria-selected').catch(() => null);
      if (finalSelected === 'true') break;
      // Remount reset the tab - retry
      await this.page.waitForTimeout(500);
    }
  }

  /**
   * Switch to EBS tab
   * Uses aria-selected to reliably detect current tab state
   * (isVisible() is unreliable for Cloudscape Tabs which may render all content in DOM)
   * Uses retry loop to handle StorageManagementView remounting (resets activeTabId to EFS default)
   */
  async switchToEBS() {
    // Retry up to 3 times to handle component remounts during loadApplicationData
    for (let attempt = 0; attempt < 3; attempt++) {
      const ebsTab = this.page.getByRole('tab', { name: /ebs/i });

      // Check if already on EBS tab using aria-selected (reliable for Cloudscape Tabs)
      const isSelected = await ebsTab.getAttribute('aria-selected').catch(() => null);
      if (isSelected !== 'true') {
        // Wait for tab to be ready before clicking
        await ebsTab.waitFor({ state: 'visible' }).catch(() => {});
        await this.page.waitForTimeout(300);
        await ebsTab.click().catch(() => {});
        await this.waitForLoadingComplete();
      }

      // Wait for Cloudscape Table loading to clear
      await this.waitForStorageLoaded();

      // Verify EBS tab is still selected after loading (handles remount during wait)
      const finalSelected = await this.page.getByRole('tab', { name: /ebs/i })
        .getAttribute('aria-selected').catch(() => null);
      if (finalSelected === 'true') break;
      // Remount reset the tab - retry
      await this.page.waitForTimeout(500);
    }
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
   * Uses .filter() with volume-name testid for reliable row matching
   * (avoids :has-text() strict mode violations when multiple rows contain similar text)
   */
  getEFSVolumeByName(name: string): Locator {
    return this.page.locator('[data-testid="efs-table"] tbody tr').filter({
      has: this.page.locator('[data-testid="volume-name"]').filter({ hasText: name })
    });
  }

  /**
   * Get EBS volume by name
   * Uses .filter() with volume-name testid for reliable row matching
   */
  getEBSVolumeByName(name: string): Locator {
    return this.page.locator('[data-testid="ebs-table"] tbody tr').filter({
      has: this.page.locator('[data-testid="volume-name"]').filter({ hasText: name })
    });
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

    // Wait for the specific volume row to be visible before clicking
    // loadApplicationData (triggered by fire-and-forget callbacks or the 30s interval) puts the
    // table into loading state, hiding rows. We need to wait for the loading cycle to complete.
    await volume.waitFor({ state: 'visible', timeout: 30000 });

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click({ timeout: 30000 });

    // Wait for menu to appear and click Delete option
    const deleteOption = this.page.getByRole('menuitem', { name: 'Delete', exact: true });
    await deleteOption.click();
  }

  /**
   * Delete EBS volume
   */
  async deleteEBSVolume(name: string) {
    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(name);

    // Wait for the specific volume row to be visible before clicking
    // loadApplicationData (triggered by fire-and-forget callbacks or the 30s interval) puts the
    // table into loading state, hiding rows. We need to wait for the loading cycle to complete.
    await volume.waitFor({ state: 'visible', timeout: 30000 });

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click({ timeout: 30000 });

    // Wait for menu to appear and click Delete option
    const deleteOption = this.page.getByRole('menuitem', { name: 'Delete', exact: true });
    await deleteOption.click();
  }

  /**
   * Delete EBS volume if it exists in the table (resilient cleanup)
   * Handles cases where the volume row may not be visible (daemon dead, table empty, etc.)
   * Also handles the confirmation dialog automatically.
   */
  async deleteEBSVolumeIfExists(name: string, confirmDialog: ConfirmDialog): Promise<void> {
    try {
      await this.switchToEBS();
      const volume = this.getEBSVolumeByName(name);
      const count = await volume.count();
      if (count === 0) return; // Not in table
      const actionsButton = volume.getByRole('button', { name: 'Actions' });
      await actionsButton.click({ timeout: 5000 });
      const deleteOption = this.page.getByRole('menuitem', { name: 'Delete', exact: true });
      await deleteOption.click({ timeout: 5000 });
      await confirmDialog.confirmDelete();
    } catch {
      // Volume not found or daemon dead - nothing to clean up
    }
  }

  /**
   * Delete EFS volume if it exists in the table (resilient cleanup)
   * Handles cases where the volume row may not be visible (daemon dead, table empty, etc.)
   * Also handles the confirmation dialog automatically.
   */
  async deleteEFSVolumeIfExists(name: string, confirmDialog: ConfirmDialog): Promise<void> {
    try {
      await this.switchToEFS();
      const volume = this.getEFSVolumeByName(name);
      const count = await volume.count();
      if (count === 0) return; // Not in table
      const actionsButton = volume.getByRole('button', { name: 'Actions' });
      await actionsButton.click({ timeout: 5000 });
      const deleteOption = this.page.getByRole('menuitem', { name: 'Delete', exact: true });
      await deleteOption.click({ timeout: 5000 });
      await confirmDialog.confirmDelete();
    } catch {
      // Volume not found or daemon dead - nothing to clean up
    }
  }

  /**
   * Mount EFS volume to instance
   */
  async mountEFSVolume(volumeName: string, instanceName: string) {
    await this.switchToEFS();
    const volume = this.getEFSVolumeByName(volumeName);

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click();

    // Wait for menu to appear and click Mount option
    const mountOption = this.page.getByRole('menuitem', { name: 'Mount', exact: true });
    await mountOption.click();

    // Select instance in Cloudscape Select dialog:
    // 1. Click the trigger button to open the dropdown
    const instanceSelectTrigger = this.page.getByTestId('mount-instance-select');
    await instanceSelectTrigger.click();
    // 2. Click the option matching the instance name
    // Use .first() because Cloudscape Select renders both the trigger display and the dropdown option
    // with role="option", so strict mode would fail without disambiguating with .first()
    await this.page.getByRole('option', { name: instanceName }).first().click();

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
   */
  async attachEBSVolume(volumeName: string, instanceName: string) {
    await this.switchToEBS();
    const volume = this.getEBSVolumeByName(volumeName);

    // Click the Actions dropdown button
    const actionsButton = volume.getByRole('button', { name: 'Actions' });
    await actionsButton.click();

    // Wait for menu to appear and click Attach option
    const attachOption = this.page.getByRole('menuitem', { name: 'Attach', exact: true });
    await attachOption.click();

    // Select instance in Cloudscape Select dialog:
    // 1. Click the trigger button to open the dropdown (using unique attach-instance-select testid)
    const instanceSelectTrigger = this.page.getByTestId('attach-instance-select');
    await instanceSelectTrigger.click();
    // 2. Click the option matching the instance name
    // Use .first() because Cloudscape Select renders both the trigger display and the dropdown option
    // with role="option", so strict mode would fail without disambiguating with .first()
    await this.page.getByRole('option', { name: instanceName }).first().click();

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
    // Determine which search input to use based on which tab is active
    // Use aria-selected on the tab element (reliable for Cloudscape Tabs)
    const efsTab = this.page.getByRole('tab', { name: /efs/i });
    const isEfsSelected = await efsTab.getAttribute('aria-selected');
    const testId = isEfsSelected === 'true' ? 'efs-search-input' : 'ebs-search-input';

    // Use React evaluation trick to reliably update the controlled TextFilter input.
    // Neither fill() nor pressSequentially reliably trigger Cloudscape's React onChange:
    // - fill() may not trigger the event at all for Cloudscape's controlled input
    // - pressSequentially loses focus after each keystroke when the table re-renders
    // Solution: Use the native value setter + dispatch 'input' event directly,
    // which bypasses React's value wrapper and triggers the synthetic event system.
    await this.page.evaluate(({ testId: tid, query: q }) => {
      const el = document.querySelector(`[data-testid="${tid}"] input`) as HTMLInputElement;
      if (!el) return;
      // Use native value setter to bypass React's controlled input wrapper
      const nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')?.set;
      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(el, q);
      }
      // Dispatch input event to trigger React's synthetic onChange handler
      el.dispatchEvent(new Event('input', { bubbles: true, cancelable: true }));
    }, { testId, query });

    // Wait for React to re-render with filtered results
    await this.waitForStorageLoaded();
  }

  /**
   * Get EFS volume count
   */
  async getEFSVolumeCount(): Promise<number> {
    await this.switchToEFS();
    // Use getVisibleVolumeCount to filter out loading/empty state rows
    return await this.getVisibleVolumeCount('efs');
  }

  /**
   * Get EBS volume count
   */
  async getEBSVolumeCount(): Promise<number> {
    await this.switchToEBS();
    // Use getVisibleVolumeCount to filter out loading/empty state rows
    return await this.getVisibleVolumeCount('ebs');
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
   * Wait for volume data to be visible in table
   */
  async waitForVolumeDataVisible(volumeType: 'ebs' | 'efs'): Promise<void> {
    const tableTestId = volumeType === 'ebs' ? 'ebs-table' : 'efs-table';

    // Wait for table rows with actual data (not just loading state)
    await this.page.waitForFunction(
      (testId) => {
        const table = document.querySelector(`[data-testid="${testId}"]`);
        if (!table) return false;
        const rows = table.querySelectorAll('tbody tr');
        return rows.length > 0 && rows[0].textContent && rows[0].textContent.trim().length > 0;
      },
      tableTestId,
      { timeout: 15000 }
    );
  }

  /**
   * Get visible volume count after filtering
   * Counts only rows that have a volume-name cell (filters out loading/empty/no-match rows)
   */
  async getVisibleVolumeCount(volumeType: 'ebs' | 'efs'): Promise<number> {
    const tableTestId = volumeType === 'ebs' ? 'ebs-table' : 'efs-table';
    // Only count rows that have a volume-name testid - this reliably excludes
    // Cloudscape skeleton loading rows, empty state rows, and no-match rows
    return await this.page.locator(`[data-testid="${tableTestId}"] tbody tr:has([data-testid="volume-name"])`).count();
  }

  /**
   * Wait for EFS volume row to appear in table
   * Note: Volume may still be in "creating" state - use waitForVolumeState() if you need "available"
   * Uses 5-minute timeout to accommodate AWS EFS creation (60-180s typically)
   */
  async waitForEFSVolumeToExist(name: string): Promise<boolean> {
    await this.switchToEFS();

    // Wait for volume row to appear in table
    // Explicit 5-minute timeout because AWS EFS creation takes 60-180s
    // actionTimeout (10s) is too short for AWS operations
    const volume = this.getEFSVolumeByName(name);
    try {
      await volume.waitFor({ state: 'visible', timeout: 300000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Wait for EFS volume row to disappear from table
   * Used after delete to confirm removal from UI (backend deletes + loadApplicationData refreshes)
   * Uses waitForFunction to poll until table is NOT in loading state AND volume is gone.
   * (Cannot use waitFor({ state: 'hidden' }) - Cloudscape loading state also hides rows)
   * Uses 3-minute timeout to accommodate AWS EFS deletion (30-120s typically)
   */
  async waitForEFSVolumeToDisappear(name: string): Promise<boolean> {
    await this.switchToEFS();
    try {
      await this.page.waitForFunction(
        (args) => {
          const { volumeName } = args;
          // Check that table is not in loading state
          const bodyText = document.body.innerText || '';
          if (bodyText.includes('Loading shared storage volumes from AWS')) return false;
          // Check that the volume row is gone
          const table = document.querySelector('[data-testid="efs-table"]');
          if (!table) return false;
          const rows = Array.from(table.querySelectorAll('tbody tr'));
          return !rows.some(row => row.textContent?.includes(volumeName));
        },
        { volumeName: name },
        { timeout: 180000 }
      );
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Wait for EBS volume row to appear in table
   * Note: Volume may still be in "creating" state - use waitForVolumeState() if you need "available"
   * Uses polling loop because StorageManagementView is defined inside App component body,
   * causing it to remount on every loadApplicationData call (resetting activeTabId to 'shared'/EFS).
   * Re-switches to EBS on each poll iteration to handle these remounts.
   * Uses 5-minute total timeout to accommodate AWS EBS creation (30-120s typically)
   */
  async waitForEBSVolumeToExist(name: string): Promise<boolean> {
    const startTime = Date.now();
    const timeout = 300000; // 5 minutes

    while (Date.now() - startTime < timeout) {
      // Re-switch to EBS tab (may have reset to EFS due to StorageManagementView remount)
      await this.switchToEBS();
      const volume = this.getEBSVolumeByName(name);
      if (await volume.isVisible()) return true;
      // Wait before next poll (give time for volume to appear)
      await this.page.waitForTimeout(3000);
    }
    return false;
  }

  /**
   * Wait for EBS volume row to disappear from table
   * Uses polling loop with atomic DOM check to handle StorageManagementView remounting.
   *
   * Root cause: StorageManagementView is defined inline inside App component body.
   * This creates a new function reference on every App re-render, causing React to
   * unmount/remount the component and reset activeTabId to 'shared' (EFS default).
   * After a remount, the EBS panel is hidden - so isVisible() returns false even
   * though the volume is still in the DOM. This is a false positive.
   *
   * Fix: Use page.evaluate() to atomically check BOTH:
   * 1. The EBS tab has aria-selected="true" (truly active, not just clicked)
   * 2. The volume is absent from the EBS table
   * If the EBS tab is not selected, skip this iteration (tab reset happened).
   *
   * Uses 3-minute total timeout to accommodate AWS EBS deletion (30-120s typically)
   */
  async waitForEBSVolumeToDisappear(name: string): Promise<boolean> {
    const startTime = Date.now();
    const timeout = 180000; // 3 minutes

    while (Date.now() - startTime < timeout) {
      // Re-switch to EBS tab (may have reset to EFS due to StorageManagementView remount)
      await this.switchToEBS();
      // Allow React to process the tab click and re-render
      await this.page.waitForTimeout(300);

      // Atomic DOM check: verify EBS tab IS active AND volume is absent
      // (Not two separate calls - that would have a race condition between them)
      const result = await this.page.evaluate((volumeName) => {
        // Verify EBS tab is truly selected in the DOM
        const tabs = Array.from(document.querySelectorAll('[role="tab"]'));
        const ebsTab = tabs.find(t => t.textContent?.includes('Private (EBS)'));
        if (!ebsTab || ebsTab.getAttribute('aria-selected') !== 'true') {
          return { tabActive: false, volumeGone: false, loading: false };
        }

        // Check loading state
        const bodyText = document.body.innerText || '';
        if (bodyText.includes('Loading private storage volumes from AWS')) {
          return { tabActive: true, volumeGone: false, loading: true };
        }

        // Check if volume is in EBS table
        const ebsTable = document.querySelector('[data-testid="ebs-table"]');
        if (!ebsTable) return { tabActive: true, volumeGone: true, loading: false };

        const rows = Array.from(ebsTable.querySelectorAll('tbody tr'));
        const volumePresent = rows.some(row => row.textContent?.includes(volumeName));
        return { tabActive: true, volumeGone: !volumePresent, loading: false };
      }, name);

      if (result.tabActive && result.volumeGone && !result.loading) {
        return true;
      }

      await this.page.waitForTimeout(3000);
    }
    return false;
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
    // Use polling loop to handle StorageManagementView remounting (resets activeTabId to EFS default)
    // For EBS volumes especially, we must re-switch after each state monitor cycle
    const startTime = Date.now();
    const timeout = 360000; // 6 minutes: accommodates AWS state transitions + state monitor polling

    while (Date.now() - startTime < timeout) {
      // Re-switch to correct tab (handles StorageManagementView remount resetting to EFS)
      if (type === 'efs') {
        await this.switchToEFS();
      } else {
        await this.switchToEBS();
      }

      // Check if the volume has reached the target state
      const table = type === 'efs' ? '[data-testid="efs-table"]' : '[data-testid="ebs-table"]';
      const found = await this.page.evaluate(
        (args) => {
          const { tableSelector, volumeName, target } = args;
          const tableEl = document.querySelector(tableSelector);
          if (!tableEl) return false;
          const rows = Array.from(tableEl.querySelectorAll('tbody tr'));
          const volumeRow = rows.find(row => row.textContent?.includes(volumeName));
          if (!volumeRow) return false;
          const badge = volumeRow.querySelector('[data-testid="status-badge"]');
          if (!badge) return false;
          return badge.textContent?.toLowerCase().trim() === target.toLowerCase();
        },
        { tableSelector: table, volumeName: name, target: targetState }
      );

      if (found) return true;
      await this.page.waitForTimeout(3000);
    }
    return false;
  }
}
