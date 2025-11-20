/**
 * Backup Workflows E2E Tests
 *
 * End-to-end tests for backup and snapshot management workflows in Prism GUI.
 * Tests: Backup creation (full/incremental), restore, clone, and backup management.
 */

import { test, expect } from '@playwright/test';
import { BasePage, ConfirmDialog } from './pages';

test.describe('Backup Management Workflows', () => {
  let basePage: BasePage;
  let confirmDialog: ConfirmDialog;

  test.beforeEach(async ({ page }) => {
    basePage = new BasePage(page);
    confirmDialog = new ConfirmDialog(page);

    await basePage.goto();
    await basePage.navigateToTab('backups');
  });

  test.describe('Backup List Display', () => {
    test('should display list of backups', async ({ page }) => {
      // Wait for backups to load
      await page.waitForTimeout(2000);

      // Check for backups table or empty state
      const backupsTable = page.locator('[data-testid="backups-table"]');
      const emptyState = page.locator('[data-testid="empty-backups"]');

      const hasBackups = await backupsTable.isVisible();
      const hasEmptyState = await emptyState.isVisible();

      // Either backups exist or empty state is shown
      expect(hasBackups || hasEmptyState).toBe(true);
    });

    test('should display backup details (size, type, date)', async ({ page }) => {
      // Wait for loading to complete (AWS API can take 10+ seconds)
      await page.waitForTimeout(2000);

      // Wait for either the table to appear with data or empty state (max 15s)
      await page.waitForSelector('[data-testid="backups-table"], [data-testid="empty-backups"]', {
        timeout: 15000
      });

      // Additional wait for data to populate
      await page.waitForTimeout(1000);

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

      // Should display type (full/incremental)
      expect(backupText).toMatch(/full|incremental/i);

      // Should display date/time
      expect(backupText).toMatch(/\d{4}-\d{2}-\d{2}|\d{1,2}\/\d{1,2}\/\d{2,4}/);
    });

    test('should show monthly cost for each backup', async ({ page }) => {
      // Wait for loading to complete (AWS API can take 10+ seconds)
      await page.waitForTimeout(2000);

      // Wait for either the table to appear with data or empty state (max 15s)
      await page.waitForSelector('[data-testid="backups-table"], [data-testid="empty-backups"]', {
        timeout: 15000
      });

      // Additional wait for data to populate
      await page.waitForTimeout(1000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for cost testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const backupText = await firstBackup.textContent();

      // Should display cost ($X.XX/month)
      expect(backupText).toMatch(/\$\d+\.\d{2}/);
    });
  });

  test.describe('Create Backup Workflow', () => {
    test('should open create backup dialog', async ({ page }) => {
      const createButton = page.getByRole('button', { name: /create.*backup/i });
      await createButton.click();

      // Dialog should open
      const dialog = page.locator('[role="dialog"]', { hasText: /create.*backup/i });
      await dialog.waitFor({ state: 'visible', timeout: 5000 });

      expect(await dialog.isVisible()).toBe(true);
    });

    test('should create full backup', async ({ page }) => {
      await page.getByRole('button', { name: /create.*backup/i }).click();

      // Wait for dialog
      await page.waitForTimeout(1000);

      // Select instance
      const instanceSelect = page.getByLabel(/instance/i);
      await instanceSelect.selectOption({ index: 0 }); // Select first instance

      // Fill backup name
      await basePage.fillInput('backup name', 'test-full-backup');

      // Select full backup type
      await basePage.selectOption('backup type', 'full');

      // Create backup
      await basePage.clickButton('create');

      // Wait for creation
      await page.waitForTimeout(3000);

      // Should show success message
      const successMessage = await page.locator('text=/success|created/i').isVisible();
      expect(successMessage).toBe(true);

      // Cleanup
      const backupRow = page.locator('tr:has-text("test-full-backup")');
      const deleteButton = backupRow.getByRole('button', { name: /delete/i });
      if (await deleteButton.isVisible()) {
        await deleteButton.click();
        await confirmDialog.confirmDelete();
      }
    });

    test('should create incremental backup', async ({ page }) => {
      await page.getByRole('button', { name: /create.*backup/i }).click();
      await page.waitForTimeout(1000);

      const instanceSelect = page.getByLabel(/instance/i);
      await instanceSelect.selectOption({ index: 0 });

      await basePage.fillInput('backup name', 'test-incremental-backup');
      await basePage.selectOption('backup type', 'incremental');
      await basePage.clickButton('create');

      await page.waitForTimeout(3000);

      const successMessage = await page.locator('text=/success|created/i').isVisible();
      expect(successMessage).toBe(true);

      // Cleanup
      const backupRow = page.locator('tr:has-text("test-incremental-backup")');
      const deleteButton = backupRow.getByRole('button', { name: /delete/i });
      if (await deleteButton.isVisible()) {
        await deleteButton.click();
        await confirmDialog.confirmDelete();
      }
    });

    test('should show estimated backup size', async ({ page }) => {
      await page.getByRole('button', { name: /create.*backup/i }).click();
      await page.waitForTimeout(1000);

      const instanceSelect = page.getByLabel(/instance/i);
      await instanceSelect.selectOption({ index: 0 });

      // Wait for size estimate to calculate
      await page.waitForTimeout(1000);

      // Should show size estimate
      const estimateText = page.locator('[data-testid="size-estimate"]');
      const hasEstimate = await estimateText.isVisible();

      if (hasEstimate) {
        const estimateValue = await estimateText.textContent();
        expect(estimateValue).toMatch(/\d+\s*GB/i);
      }

      await basePage.clickButton('cancel');
    });

    test('should validate backup name is required', async ({ page }) => {
      await page.getByRole('button', { name: /create.*backup/i }).click();
      await page.waitForTimeout(1000);

      const instanceSelect = page.getByLabel(/instance/i);
      await instanceSelect.selectOption({ index: 0 });

      // Don't fill name, just click create
      await basePage.clickButton('create');

      // Should show validation error
      const validationError = await page.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/name.*required/i);
    });

    test('should validate instance is selected', async ({ page }) => {
      await page.getByRole('button', { name: /create.*backup/i }).click();
      await page.waitForTimeout(1000);

      // Fill name but don't select instance
      await basePage.fillInput('backup name', 'test-validation');
      await basePage.clickButton('create');

      // Should show validation error
      const validationError = await page.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/instance.*required/i);
    });
  });

  test.describe('Delete Backup Workflow', () => {
    test('should delete backup with confirmation', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for deletion testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const backupName = await firstBackup.locator('[data-testid="backup-name"]').textContent();

      // Click delete button
      const deleteButton = firstBackup.getByRole('button', { name: /delete/i });
      await deleteButton.click();

      // Confirmation dialog should appear
      await confirmDialog.waitForDialog();

      // Should show cost savings information
      const hasCostSavings = await confirmDialog.hasCostSavings();
      expect(hasCostSavings).toBe(true);

      // Cancel for safety
      await confirmDialog.clickCancel();
    });

    test('should show cost savings after deletion', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for deletion testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const deleteButton = firstBackup.getByRole('button', { name: /delete/i });
      await deleteButton.click();

      await confirmDialog.waitForDialog();

      // Dialog should show how much will be saved
      const message = await confirmDialog.getMessage();
      expect(message).toMatch(/save.*\$|free.*GB/i);

      await confirmDialog.clickCancel();
    });
  });

  test.describe('Restore Backup Workflow', () => {
    test('should open restore dialog', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for restore testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const restoreButton = firstBackup.getByRole('button', { name: /restore/i });
      await restoreButton.click();

      // Restore dialog should open
      const dialog = page.locator('[role="dialog"]', { hasText: /restore/i });
      await dialog.waitFor({ state: 'visible', timeout: 5000 });

      expect(await dialog.isVisible()).toBe(true);

      // Cancel
      await basePage.clickButton('cancel');
    });

    test('should restore backup to new instance', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for restore testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const restoreButton = firstBackup.getByRole('button', { name: /restore/i });
      await restoreButton.click();

      // Wait for dialog
      await page.waitForTimeout(1000);

      // Fill new instance name
      await basePage.fillInput('new instance name', 'restored-test-instance');

      // Start restore
      await basePage.clickButton('restore');

      // Wait for restore to initiate
      await page.waitForTimeout(3000);

      // Should show success message
      const successMessage = await page.locator('text=/success|restoring/i').isVisible();
      expect(successMessage).toBe(true);
    });

    test('should warn about restore time', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for restore testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const restoreButton = firstBackup.getByRole('button', { name: /restore/i });
      await restoreButton.click();

      await page.waitForTimeout(1000);

      // Should show warning about restore time
      const warningText = page.locator('text=/may take.*minutes|restore.*time/i');
      const hasWarning = await warningText.isVisible();
      expect(hasWarning).toBe(true);

      await basePage.clickButton('cancel');
    });
  });

  test.describe('Clone from Backup Workflow', () => {
    test('should clone instance from backup', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for clone testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const cloneButton = firstBackup.getByRole('button', { name: /clone/i });

      const hasCloneButton = await cloneButton.isVisible();
      if (!hasCloneButton) {
        // Skip: Clone button not available
        test.skip();
        return;
      }

      await cloneButton.click();

      // Wait for dialog
      await page.waitForTimeout(1000);

      // Fill clone name
      await basePage.fillInput('clone name', 'cloned-test-instance');

      // Create clone
      await basePage.clickButton('clone');

      // Wait for clone to initiate
      await page.waitForTimeout(3000);

      // Should show success message
      const successMessage = await page.locator('text=/success|cloning/i').isVisible();
      expect(successMessage).toBe(true);
    });
  });

  test.describe('Backup Filtering', () => {
    test('should filter by backup type', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for filtering
        test.skip();
        return;
      }

      // Filter by full backups
      const filterSelect = page.getByLabel(/filter.*type/i);
      const hasFilter = await filterSelect.isVisible();

      if (hasFilter) {
        await filterSelect.selectOption('full');
        await page.waitForTimeout(500);

        // Verify filtering worked (all visible backups should be full)
        const visibleRows = await page.locator('[data-testid="backups-table"] tbody tr').all();
        for (const row of visibleRows) {
          const rowText = await row.textContent();
          expect(rowText).toMatch(/full/i);
        }
      }
    });

    test('should search backups by name', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for search
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const backupName = await firstBackup.locator('[data-testid="backup-name"]').textContent();

      if (!backupName) {
        // Skip: Could not get backup name
        test.skip();
        return;
      }

      // Search for backup
      const searchInput = page.getByPlaceholder(/search.*backups/i);
      await searchInput.fill(backupName);
      await page.waitForTimeout(500);

      // Verify search results
      const searchResults = await page.locator('[data-testid="backups-table"] tbody tr').all();
      expect(searchResults.length).toBeGreaterThanOrEqual(1);
    });
  });

  test.describe('Empty State Handling', () => {
    test('should show helpful message when no backups exist', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Verify empty state
        const emptyState = page.locator('[data-testid="empty-backups"]');
        const isVisible = await emptyState.isVisible();
        expect(isVisible).toBe(true);

        // Should have helpful content
        const emptyStateText = await emptyState.textContent();
        expect(emptyStateText?.length).toBeGreaterThan(0);
      }
    });

    test('should provide create backup button in empty state', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Create button should be visible
        const createButton = page.getByRole('button', { name: /create.*backup/i });
        const isVisible = await createButton.isVisible();
        expect(isVisible).toBe(true);
      }
    });
  });

  test.describe('Backup Status Monitoring', () => {
    test('should display backup status', async ({ page }) => {
      await page.waitForTimeout(2000);

      const backupRows = await page.locator('[data-testid="backups-table"] tbody tr').all();

      if (backupRows.length === 0) {
        // Skip: No backups available for status testing
        test.skip();
        return;
      }

      const firstBackup = backupRows[0];
      const statusBadge = firstBackup.locator('[data-testid="status-badge"]');

      if (await statusBadge.isVisible()) {
        const statusText = await statusBadge.textContent();

        // Valid backup states
        const validStates = ['available', 'creating', 'deleting', 'error'];
        const isValidState = validStates.some((state) => statusText?.toLowerCase().includes(state));
        expect(isValidState).toBe(true);
      }
    });
  });
});
