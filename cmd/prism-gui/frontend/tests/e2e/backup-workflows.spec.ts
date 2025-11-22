/**
 * Backup Workflows E2E Tests
 *
 * End-to-end tests for backup and snapshot management workflows in Prism GUI.
 * Tests: Backup creation (full/incremental), restore, clone, and backup management.
 *
 * NOTE: These tests are designed for the actual Cloudscape Design System components,
 * not generic HTML elements. Cloudscape Select uses custom button+listbox patterns.
 */

import { test, expect, Page } from '@playwright/test';

/**
 * Helper: Wait for AWS API calls to complete and table loading to finish
 * AWS API calls can take 8-15 seconds, especially when querying multiple regions
 */
async function waitForBackupsToLoad(page: Page, timeoutMs: number = 30000) {
  // First, wait for the Backups view heading to appear (confirms view rendered)
  await page.waitForSelector('text=Available Backups', { timeout: timeoutMs });

  // Give React time to fully render the view components
  await page.waitForTimeout(1000);

  // Now wait for the table to exist
  await page.waitForSelector('[data-testid="backups-table"]', { timeout: timeoutMs });

  // Wait for loading text to disappear (critical - table shows "Loading backups from AWS" while loading)
  // Use polling: true to check every 200ms instead of waiting for mutations
  await page.waitForFunction(
    () => {
      const table = document.querySelector('[data-testid="backups-table"]');
      const tableText = table?.textContent || '';
      return !tableText.includes('Loading backups');
    },
    { timeout: timeoutMs, polling: 200 }
  );

  // Give a longer buffer for AWS responses and React re-renders
  await page.waitForTimeout(2000);
}

/**
 * Helper: Navigate to Backups view
 * Uses the same pattern as working storage-workflows tests
 */
async function navigateToBackups(page: Page) {
  // Find the link by accessible name (text content) - this is how BasePage.navigateToTab() works
  const link = page.getByRole('link', { name: /backups/i });
  await link.click();

  // Wait for navigation state to update
  await page.waitForTimeout(500);
}

/**
 * Helper: Select an option from Cloudscape Select component
 * Cloudscape Select uses a button trigger + dropdown pattern, not native <select>
 */
async function selectCloudscapeOption(page: Page, labelText: string, optionText: string) {
  // Find the FormField by label
  const formField = page.locator('label', { hasText: new RegExp(labelText, 'i') }).locator('..').locator('..');

  // Click the Select button trigger (opens dropdown)
  await formField.locator('button[aria-haspopup="listbox"]').click();

  // Wait for dropdown to appear
  await page.waitForTimeout(300);

  // Click the option in the dropdown
  await page.getByRole('option', { name: new RegExp(optionText, 'i') }).click();
}

/**
 * Helper: Click item in ButtonDropdown (Actions menu)
 */
async function clickDropdownAction(page: Page, rowIndex: number, actionText: string) {
  const rows = page.locator('[data-testid="backups-table"] tbody tr');
  const row = rows.nth(rowIndex);

  // Click the Actions dropdown button
  await row.getByRole('button', { name: 'Actions' }).click();
  await page.waitForTimeout(300);

  // Click the specific action item
  await page.getByRole('menuitem', { name: new RegExp(actionText, 'i') }).click();
}

test.describe('Backup Management Workflows', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set localStorage BEFORE navigating to prevent onboarding modal
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });

    // Navigate to Prism GUI (uses baseURL from playwright.config.js)
    await page.goto('/');

    // Don't use networkidle - AWS API calls continuously, use domcontentloaded instead
    await page.waitForLoadState('domcontentloaded');

    // Wait a bit for app to initialize
    await page.waitForTimeout(2000);

    // Navigate to Backups view
    await navigateToBackups(page);
  });

  test.describe('Backup List Display', () => {
    test('should display list of backups or empty state', async ({ page }) => {
      // Wait for backups to load from AWS
      await waitForBackupsToLoad(page);

      // Check for backups table or empty state
      const backupsTable = page.locator('[data-testid="backups-table"]');
      const emptyState = page.locator('[data-testid="empty-backups"]');

      const hasTable = await backupsTable.isVisible();
      const hasEmptyState = await emptyState.isVisible();

      // Either backups exist or empty state is shown
      expect(hasTable || hasEmptyState).toBe(true);
    });

    test('should display backup details (size, status, date, cost)', async ({ page }) => {
      // Wait for backups to load from AWS (can take 10-15 seconds)
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for display testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const backupText = await firstBackup.textContent();

      // Should display size (GB)
      expect(backupText).toMatch(/\d+\s*GB/i);

      // Should display status (available, creating, etc.)
      expect(backupText).toMatch(/available|creating|pending|deleting/i);

      // Should display date/time
      expect(backupText).toMatch(/\d{1,2}\/\d{1,2}\/\d{2,4}/);

      // Should display cost ($X.XX)
      expect(backupText).toMatch(/\$\d+\.\d{2}/);
    });

    test('should show backup name as link', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Check that backup name cell has data-testid
      const backupNameCell = page.locator('[data-testid="backup-name"]').first();
      expect(await backupNameCell.isVisible()).toBe(true);
    });

    test('should show status indicator badge', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Check for status badge
      const statusBadge = page.locator('[data-testid="status-badge"]').first();
      expect(await statusBadge.isVisible()).toBe(true);
    });
  });

  test.describe('Create Backup Workflow', () => {
    test('should open create backup dialog', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const createButton = page.getByRole('button', { name: /create.*backup/i });
      await createButton.click();

      // Dialog should open
      const dialog = page.locator('[role="dialog"]', { hasText: /create.*backup/i });
      await dialog.waitFor({ state: 'visible', timeout: 5000 });

      expect(await dialog.isVisible()).toBe(true);
    });

    test('should create backup from instance', async ({ page }) => {
      await waitForBackupsToLoad(page);

      // Open create backup dialog
      await page.getByRole('button', { name: /create.*backup/i }).click();
      await page.waitForTimeout(1000);

      // Select instance using Cloudscape Select component
      await selectCloudscapeOption(page, 'Instance', '0'); // Select first instance by index
      await page.waitForTimeout(500);

      // Fill backup name
      const backupNameInput = page.locator('input').filter({ hasText: '' }).first();
      await backupNameInput.fill(`test-backup-${Date.now()}`);

      // Create backup
      const createButton = page.locator('[data-testid="create-backup-submit"]');
      await createButton.click();

      // Wait for success notification
      await page.waitForTimeout(2000);

      // Should show success message in notifications
      const notification = page.locator('text=/success|created/i');
      expect(await notification.isVisible()).toBe(true);
    });

    test('should validate instance selection required', async ({ page }) => {
      await waitForBackupsToLoad(page);

      await page.getByRole('button', { name: /create.*backup/i }).click();
      await page.waitForTimeout(1000);

      // Fill backup name but don't select instance
      const backupNameInput = page.locator('input').filter({ hasText: '' }).first();
      await backupNameInput.fill('test-validation');

      // Try to create without selecting instance
      const createButton = page.locator('[data-testid="create-backup-submit"]');

      // Button should be disabled
      expect(await createButton.isDisabled()).toBe(true);
    });

    test('should validate backup name required', async ({ page }) => {
      await waitForBackupsToLoad(page);

      await page.getByRole('button', { name: /create.*backup/i }).click();
      await page.waitForTimeout(1000);

      // Select instance but don't fill name
      await selectCloudscapeOption(page, 'Instance', '0');
      await page.waitForTimeout(500);

      // Create button should be disabled without name
      const createButton = page.locator('[data-testid="create-backup-submit"]');
      expect(await createButton.isDisabled()).toBe(true);
    });
  });

  test.describe('Delete Backup Workflow', () => {
    test('should open delete confirmation dialog', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Click Actions dropdown and select Delete
      await clickDropdownAction(page, 0, 'Delete');

      // Confirmation dialog should appear
      await page.waitForTimeout(500);
      const dialog = page.locator('[role="dialog"]', { hasText: /delete.*confirmation/i });
      await dialog.waitFor({ state: 'visible', timeout: 5000 });

      expect(await dialog.isVisible()).toBe(true);
    });

    test('should show cost savings in delete dialog', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Open delete dialog
      await clickDropdownAction(page, 0, 'Delete');
      await page.waitForTimeout(500);

      // Dialog should show cost savings
      const dialog = page.locator('[role="dialog"]');
      const dialogText = await dialog.textContent();

      expect(dialogText).toMatch(/cost.*savings|monthly.*savings|save/i);
      expect(dialogText).toMatch(/\$\d+\.\d{2}/);
    });

    test('should be able to cancel delete', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Open delete dialog
      await clickDropdownAction(page, 0, 'Delete');
      await page.waitForTimeout(500);

      // Click Cancel
      await page.getByRole('button', { name: /cancel/i }).click();
      await page.waitForTimeout(500);

      // Dialog should close
      const dialog = page.locator('[role="dialog"]');
      expect(await dialog.isVisible()).toBe(false);
    });
  });

  test.describe('Restore Backup Workflow', () => {
    test('should open restore dialog', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Click Actions dropdown and select Restore
      await clickDropdownAction(page, 0, 'Restore');

      // Restore dialog should open
      await page.waitForTimeout(500);
      const dialog = page.locator('[role="dialog"]', { hasText: /restore/i });
      await dialog.waitFor({ state: 'visible', timeout: 5000 });

      expect(await dialog.isVisible()).toBe(true);
    });

    test('should show restore time warning', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Open restore dialog
      await clickDropdownAction(page, 0, 'Restore');
      await page.waitForTimeout(500);

      // Should show warning about restore time
      const dialog = page.locator('[role="dialog"]');
      const dialogText = await dialog.textContent();
      expect(dialogText).toMatch(/10-15.*minutes|restore.*time/i);
    });

    test('should validate instance name required for restore', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Open restore dialog
      await clickDropdownAction(page, 0, 'Restore');
      await page.waitForTimeout(1000);

      // Restore button should be disabled without instance name
      const restoreButton = page.locator('button[variant="primary"]', { hasText: /restore/i });
      expect(await restoreButton.isDisabled()).toBe(true);
    });
  });

  test.describe('Empty State Handling', () => {
    test('should show empty state if no backups exist', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Verify empty state is shown
        const emptyState = page.locator('[data-testid="empty-backups"]');
        expect(await emptyState.isVisible()).toBe(true);

        // Should have helpful message
        const emptyStateText = await emptyState.textContent();
        expect(emptyStateText).toContain('No backups');
      } else {
        // Skip: Backups exist, can't test empty state
        test.skip();
      }
    });

    test('should have create backup button in all states', async ({ page }) => {
      await waitForBackupsToLoad(page);

      // Create button should always be visible (in header)
      const createButton = page.getByRole('button', { name: /create.*backup/i });
      expect(await createButton.isVisible()).toBe(true);
    });
  });

  test.describe('Backup Actions', () => {
    test('should have Actions dropdown for each backup', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Each row should have Actions button
      const firstRow = backupRows[0];
      const actionsButton = firstRow.getByRole('button', { name: 'Actions' });
      expect(await actionsButton.isVisible()).toBe(true);
    });

    test('should show restore, clone, details, and delete actions', async ({ page }) => {
      await waitForBackupsToLoad(page);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        test.skip();
        return;
      }

      // Click Actions dropdown
      const firstRow = backupRows[0];
      await firstRow.getByRole('button', { name: 'Actions' }).click();
      await page.waitForTimeout(300);

      // Should show expected actions
      expect(await page.getByRole('menuitem', { name: /restore/i }).isVisible()).toBe(true);
      expect(await page.getByRole('menuitem', { name: /clone/i }).isVisible()).toBe(true);
      expect(await page.getByRole('menuitem', { name: /details/i }).isVisible()).toBe(true);
      expect(await page.getByRole('menuitem', { name: /delete/i }).isVisible()).toBe(true);
    });
  });
});
