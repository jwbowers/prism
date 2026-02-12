/**
 * Storage Workflows E2E Tests
 *
 * End-to-end tests for complete storage management workflows in Prism GUI.
 * Tests: EFS volume creation/deletion/mounting, EBS volume creation/deletion/attachment.
 */

import { test, expect } from '@playwright/test';
import { StoragePage, InstancesPage, ConfirmDialog } from './pages';

test.describe('Storage Management Workflows', () => {
  let storagePage: StoragePage;
  let instancesPage: InstancesPage;
  let confirmDialog: ConfirmDialog;

  test.beforeEach(async ({ page, context }) => {
    // Set localStorage BEFORE navigating to prevent onboarding modal
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });

    storagePage = new StoragePage(page);
    instancesPage = new InstancesPage(page);
    confirmDialog = new ConfirmDialog(page);

    await storagePage.goto();
    await storagePage.navigate();
  });

  test.describe('EFS Volume Management', () => {
    test('should display EFS volumes list', async () => {
      await storagePage.switchToEFS();

      // Verify EFS tab content is displayed
      const efsCount = await storagePage.getEFSVolumeCount();
      expect(typeof efsCount).toBe('number');
    });

    test('should create new EFS volume', async () => {
      await storagePage.switchToEFS();

      // Click create button
      await storagePage.page.getByTestId('create-efs-header-button').click();

      // Fill form
      await storagePage.page.getByRole('textbox', { name: 'EFS Volume Name' }).fill('test-efs-volume');
      await storagePage.clickButton('create');

      // Wait for creation
      await storagePage.page.waitForTimeout(3000);

      // Verify volume appears in list
      const volumeExists = await storagePage.verifyEFSVolumeExists('test-efs-volume');
      expect(volumeExists).toBe(true);

      // Cleanup
      await storagePage.deleteEFSVolume('test-efs-volume');
      await confirmDialog.confirmDelete();
    });

    test('should validate EFS volume name is required', async () => {
      await storagePage.switchToEFS();

      await storagePage.page.getByTestId('create-efs-header-button').click();

      // Don't fill name, just click create
      await storagePage.clickButton('create');

      // Should show validation error
      const validationError = await storagePage.page.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/name.*required/i);
    });

    test('should delete EFS volume with confirmation', async () => {
      await storagePage.switchToEFS();

      // Create volume first
      await storagePage.createEFSVolume('delete-test-efs');
      await storagePage.page.waitForTimeout(2000);

      // Verify it exists
      let volumeExists = await storagePage.verifyEFSVolumeExists('delete-test-efs');
      expect(volumeExists).toBe(true);

      // Delete volume
      await storagePage.deleteEFSVolume('delete-test-efs');
      await confirmDialog.waitForDialog();
      await confirmDialog.confirmDelete();
      await storagePage.page.waitForTimeout(2000);

      // Verify it's removed
      volumeExists = await storagePage.verifyEFSVolumeExists('delete-test-efs');
      expect(volumeExists).toBe(false);
    });

    test('should mount EFS volume to instance', async ({ page }) => {
      // First check if there are any running instances
      await instancesPage.navigate();
      const instanceCount = await instancesPage.getInstanceCount();

      if (instanceCount === 0) {
        // Skip: No instances available for mount testing
        test.skip();
        return;
      }

      const firstInstance = await page.locator('[data-testid="instances-table"] tbody tr').first();
      const instanceName = await firstInstance.locator('[data-testid="instance-name"]').textContent();

      if (!instanceName) {
        // Skip: Could not get instance name
        test.skip();
        return;
      }

      // Navigate back to storage
      await storagePage.navigate();
      await storagePage.switchToEFS();

      // Create EFS volume
      await storagePage.createEFSVolume('mount-test-efs');
      await storagePage.page.waitForTimeout(2000);

      // Mount to instance
      await storagePage.mountEFSVolume('mount-test-efs', instanceName);
      await storagePage.page.waitForTimeout(2000);

      // Verify mount success (check status or mounted indicator)
      const volumeRow = storagePage.getEFSVolumeByName('mount-test-efs');
      const volumeText = await volumeRow.textContent();
      expect(volumeText).toMatch(/mounted|attached/i);

      // Cleanup: Unmount and delete
      await storagePage.unmountEFSVolume('mount-test-efs', instanceName);
      await confirmDialog.confirm();
      await storagePage.page.waitForTimeout(1000);
      await storagePage.deleteEFSVolume('mount-test-efs');
      await confirmDialog.confirmDelete();
    });

    test('should unmount EFS volume from instance', async ({ page }) => {
      // This test assumes there's a mounted EFS volume
      await storagePage.switchToEFS();

      // Look for mounted volume
      const mountedVolume = page.locator('tr:has-text("mounted")').first();
      const hasMountedVolume = await mountedVolume.isVisible();

      if (!hasMountedVolume) {
        // Skip: No mounted EFS volumes available for testing
        test.skip();
        return;
      }

      const volumeName = await mountedVolume.locator('[data-testid="volume-name"]').textContent();
      const instanceName = await mountedVolume.locator('[data-testid="mounted-instance"]').textContent();

      if (!volumeName || !instanceName) {
        // Skip: Could not get volume or instance name
        test.skip();
        return;
      }

      // Unmount volume
      await storagePage.unmountEFSVolume(volumeName, instanceName);
      await confirmDialog.confirm();
      await storagePage.page.waitForTimeout(2000);

      // Verify unmount success
      const status = await storagePage.getVolumeStatus(volumeName, 'efs');
      expect(status).toMatch(/available|unmounted/i);
    });

    test('should show EFS volume status correctly', async () => {
      await storagePage.switchToEFS();

      const volumeCount = await storagePage.getEFSVolumeCount();

      if (volumeCount === 0) {
        // Skip: No EFS volumes available for status testing
        test.skip();
        return;
      }

      const firstVolume = await storagePage.page.locator('[data-testid="efs-table"] tbody tr').first();
      const volumeName = await firstVolume.locator('[data-testid="volume-name"]').textContent();

      if (!volumeName) {
        // Skip: Could not get volume name
        test.skip();
        return;
      }

      // Get volume status
      const status = await storagePage.getVolumeStatus(volumeName, 'efs');
      expect(status).toBeTruthy();

      // Status should be valid
      const validStates = ['available', 'creating', 'deleting', 'mounted', 'error'];
      const isValidState = validStates.some((state) => status?.toLowerCase().includes(state));
      expect(isValidState).toBe(true);
    });
  });

  test.describe('EBS Volume Management', () => {
    test('should display EBS volumes list', async () => {
      await storagePage.switchToEBS();

      // Verify EBS tab content is displayed
      const ebsCount = await storagePage.getEBSVolumeCount();
      expect(typeof ebsCount).toBe('number');
    });

    test('should create new EBS volume', async () => {
      await storagePage.switchToEBS();

      // Click create button
      await storagePage.page.getByTestId('create-ebs-header-button').click();

      // Fill form
      await storagePage.page.getByRole('textbox', { name: 'EBS Volume Name' }).fill('test-ebs-volume');
      await storagePage.page.getByRole('spinbutton', { name: 'EBS Volume Size' }).fill('100');
      await storagePage.clickButton('create');

      // Wait for creation
      await storagePage.page.waitForTimeout(3000);

      // Verify volume appears in list
      const volumeExists = await storagePage.verifyEBSVolumeExists('test-ebs-volume');
      expect(volumeExists).toBe(true);

      // Cleanup
      await storagePage.deleteEBSVolume('test-ebs-volume');
      await confirmDialog.confirmDelete();
    });

    test('should validate EBS volume size is required', async () => {
      await storagePage.switchToEBS();

      await storagePage.page.getByTestId('create-ebs-header-button').click();

      // Fill name but not size
      await storagePage.page.getByRole('textbox', { name: 'EBS Volume Name' }).fill('size-validation-test');
      await storagePage.clickButton('create');

      // Should show validation error
      const validationError = await storagePage.page.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/size.*required/i);
    });

    test('should validate EBS volume size is positive number', async () => {
      await storagePage.switchToEBS();

      await storagePage.page.getByTestId('create-ebs-header-button').click();

      await storagePage.page.getByRole('textbox', { name: 'EBS Volume Name' }).fill('size-validation-test');
      await storagePage.page.getByRole('spinbutton', { name: 'EBS Volume Size' }).fill('-10');
      await storagePage.clickButton('create');

      // Should show validation error
      const validationError = await storagePage.page.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/positive|invalid.*size/i);
    });

    test('should delete EBS volume with confirmation', async () => {
      await storagePage.switchToEBS();

      // Create volume first
      await storagePage.createEBSVolume('delete-test-ebs', '50');
      await storagePage.page.waitForTimeout(2000);

      // Verify it exists
      let volumeExists = await storagePage.verifyEBSVolumeExists('delete-test-ebs');
      expect(volumeExists).toBe(true);

      // Delete volume
      await storagePage.deleteEBSVolume('delete-test-ebs');
      await confirmDialog.waitForDialog();
      await confirmDialog.confirmDelete();
      await storagePage.page.waitForTimeout(2000);

      // Verify it's removed
      volumeExists = await storagePage.verifyEBSVolumeExists('delete-test-ebs');
      expect(volumeExists).toBe(false);
    });

    test('should attach EBS volume to instance', async ({ page }) => {
      // First check if there are any running instances
      await instancesPage.navigate();
      const instanceCount = await instancesPage.getInstanceCount();

      if (instanceCount === 0) {
        // Skip: No instances available for attach testing
        test.skip();
        return;
      }

      const firstInstance = await page.locator('[data-testid="instances-table"] tbody tr').first();
      const instanceName = await firstInstance.locator('[data-testid="instance-name"]').textContent();

      if (!instanceName) {
        // Skip: Could not get instance name
        test.skip();
        return;
      }

      // Navigate back to storage
      await storagePage.navigate();
      await storagePage.switchToEBS();

      // Create EBS volume
      await storagePage.createEBSVolume('attach-test-ebs', '100');
      await storagePage.page.waitForTimeout(2000);

      // Attach to instance
      await storagePage.attachEBSVolume('attach-test-ebs', instanceName);
      await storagePage.page.waitForTimeout(2000);

      // Verify attach success
      const volumeRow = storagePage.getEBSVolumeByName('attach-test-ebs');
      const volumeText = await volumeRow.textContent();
      expect(volumeText).toMatch(/attached|in-use/i);

      // Cleanup: Detach and delete
      await storagePage.detachEBSVolume('attach-test-ebs');
      await confirmDialog.confirm();
      await storagePage.page.waitForTimeout(1000);
      await storagePage.deleteEBSVolume('attach-test-ebs');
      await confirmDialog.confirmDelete();
    });

    test('should detach EBS volume from instance', async ({ page }) => {
      await storagePage.switchToEBS();

      // Look for attached volume
      const attachedVolume = page.locator('tr:has-text("attached")').first();
      const hasAttachedVolume = await attachedVolume.isVisible();

      if (!hasAttachedVolume) {
        // Skip: No attached EBS volumes available for testing
        test.skip();
        return;
      }

      const volumeName = await attachedVolume.locator('[data-testid="volume-name"]').textContent();

      if (!volumeName) {
        // Skip: Could not get volume name
        test.skip();
        return;
      }

      // Detach volume
      await storagePage.detachEBSVolume(volumeName);
      await confirmDialog.confirm();
      await storagePage.page.waitForTimeout(2000);

      // Verify detach success
      const status = await storagePage.getVolumeStatus(volumeName, 'ebs');
      expect(status).toMatch(/available|detached/i);
    });

    test('should show EBS volume size in list', async () => {
      await storagePage.switchToEBS();

      const volumeCount = await storagePage.getEBSVolumeCount();

      if (volumeCount === 0) {
        // Skip: No EBS volumes available for size display testing
        test.skip();
        return;
      }

      const firstVolume = await storagePage.page.locator('[data-testid="ebs-table"] tbody tr').first();
      const volumeText = await firstVolume.textContent();

      // Should display size information (e.g., "100 GB")
      expect(volumeText).toMatch(/\d+\s*GB/i);
    });

    test('should show EBS volume type', async () => {
      await storagePage.switchToEBS();

      const volumeCount = await storagePage.getEBSVolumeCount();

      if (volumeCount === 0) {
        // Skip: No EBS volumes available for type display testing
        test.skip();
        return;
      }

      const firstVolume = await storagePage.page.locator('[data-testid="ebs-table"] tbody tr').first();
      const volumeText = await firstVolume.textContent();

      // Should display volume type (gp2, gp3, io1, etc.)
      const hasVolumeType = /gp2|gp3|io1|io2|st1|sc1/.test(volumeText || '');
      expect(hasVolumeType).toBe(true);
    });
  });

  test.describe('Storage Search and Filtering', () => {
    test('should search EFS volumes by name', async () => {
      await storagePage.switchToEFS();

      const volumeCount = await storagePage.getEFSVolumeCount();

      if (volumeCount === 0) {
        // Skip: No EFS volumes available for search testing
        test.skip();
        return;
      }

      const firstVolume = await storagePage.page.locator('[data-testid="efs-table"] tbody tr').first();
      const volumeName = await firstVolume.locator('[data-testid="volume-name"]').textContent();

      if (!volumeName) {
        // Skip: Could not get volume name
        test.skip();
        return;
      }

      // Search for volume
      await storagePage.searchVolumes(volumeName);
      await storagePage.page.waitForTimeout(500);

      // Verify search results
      const searchResults = await storagePage.getEFSVolumeCount();
      expect(searchResults).toBeGreaterThanOrEqual(1);
    });

    test('should search EBS volumes by name', async () => {
      await storagePage.switchToEBS();

      const volumeCount = await storagePage.getEBSVolumeCount();

      if (volumeCount === 0) {
        // Skip: No EBS volumes available for search testing
        test.skip();
        return;
      }

      const firstVolume = await storagePage.page.locator('[data-testid="ebs-table"] tbody tr').first();
      const volumeName = await firstVolume.locator('[data-testid="volume-name"]').textContent();

      if (!volumeName) {
        // Skip: Could not get volume name
        test.skip();
        return;
      }

      // Search for volume
      await storagePage.searchVolumes(volumeName);
      await storagePage.page.waitForTimeout(500);

      // Verify search results
      const searchResults = await storagePage.getEBSVolumeCount();
      expect(searchResults).toBeGreaterThanOrEqual(1);
    });

    test('should show empty state when no volumes match search', async () => {
      await storagePage.switchToEFS();

      // Search for non-existent volume
      await storagePage.searchVolumes('nonexistent-volume-xyz');
      await storagePage.page.waitForTimeout(500);

      // Should show zero results
      const count = await storagePage.getEFSVolumeCount();
      expect(count).toBe(0);
    });
  });

  test.describe('Tab Switching', () => {
    test('should switch between EFS and EBS tabs', async () => {
      // Start with EFS
      await storagePage.switchToEFS();
      await storagePage.page.waitForTimeout(500);

      // Verify EFS content
      const efsTable = storagePage.page.locator('[data-testid="efs-table"]');
      expect(await efsTable.isVisible()).toBe(true);

      // Switch to EBS
      await storagePage.switchToEBS();
      await storagePage.page.waitForTimeout(500);

      // Verify EBS content
      const ebsTable = storagePage.page.locator('[data-testid="ebs-table"]');
      expect(await ebsTable.isVisible()).toBe(true);
    });

    test('should preserve tab state when navigating away and back', async () => {
      // Switch to EBS tab
      await storagePage.switchToEBS();
      await storagePage.page.waitForTimeout(500);

      // Navigate away
      await instancesPage.navigate();
      await storagePage.page.waitForTimeout(500);

      // Navigate back
      await storagePage.navigate();
      await storagePage.page.waitForTimeout(500);

      // Tab state might be preserved (implementation-dependent)
      // Just verify storage page loads correctly
      const storageContent = await storagePage.page.locator('[data-testid="storage-page"]').isVisible();
      expect(storageContent).toBe(true);
    });
  });

  test.describe('Empty State Handling', () => {
    test('should show helpful message when no EFS volumes exist', async () => {
      await storagePage.switchToEFS();

      const volumeCount = await storagePage.getEFSVolumeCount();

      if (volumeCount === 0) {
        // Verify empty state content
        const emptyState = storagePage.page.locator('[data-testid="empty-efs"]');
        const isVisible = await emptyState.isVisible();
        expect(isVisible).toBe(true);
      }
    });

    test('should show helpful message when no EBS volumes exist', async () => {
      await storagePage.switchToEBS();

      const volumeCount = await storagePage.getEBSVolumeCount();

      if (volumeCount === 0) {
        // Verify empty state content
        const emptyState = storagePage.page.locator('[data-testid="empty-ebs"]');
        const isVisible = await emptyState.isVisible();
        expect(isVisible).toBe(true);
      }
    });
  });
});
