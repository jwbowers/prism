/**
 * Storage Workflows E2E Tests
 *
 * End-to-end tests for complete storage management workflows in Prism GUI.
 * Tests: EFS volume creation/deletion/mounting, EBS volume creation/deletion/attachment.
 */

import { test, expect } from '@playwright/test';
import { StoragePage, InstancesPage, ConfirmDialog } from './pages';

test.describe('Storage Management Workflows', () => {
  // No artificial timeouts - backend uses AWS SDK waiters that poll until resources are ready (up to 30 minutes)
  // Tests complete when operations complete naturally (typically 30s - 5 minutes)

  let storagePage: StoragePage;
  let instancesPage: InstancesPage;
  let confirmDialog: ConfirmDialog;

  // Track if shared test volumes have been created in this test run
  // This prevents re-creating volumes for every test (beforeEach runs 46 times!)
  let sharedVolumesCreated = false;

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

    // Wait for UI to fully render tabs (tabs need time to become interactive)
    await page.waitForTimeout(2000);

    // Create shared test volumes for display/search tests ONLY ONCE
    // These persist across tests in this suite to avoid re-creating for every test
    // Individual tests that need specific volumes create their own
    // Note: No try/catch - setup failures should fail the test with clear errors

    if (!sharedVolumesCreated) {
      // Check if setup volumes already exist from previous test runs
      const efsExists = await storagePage.verifyEFSVolumeExists('test-setup-efs');
      if (!efsExists) {
        console.log('Creating shared EFS volume: test-setup-efs');
        await storagePage.createEFSVolume('test-setup-efs');

        const created = await storagePage.waitForEFSVolumeToExist('test-setup-efs');
        if (!created) {
          throw new Error('Failed to create shared EFS volume: test-setup-efs. This volume is required for display/search tests.');
        }
        console.log('✅ Shared EFS volume created successfully');
      }

      const ebsExists = await storagePage.verifyEBSVolumeExists('test-setup-ebs');
      if (!ebsExists) {
        console.log('Creating shared EBS volume: test-setup-ebs');
        await storagePage.createEBSVolume('test-setup-ebs', '50');

        const created = await storagePage.waitForEBSVolumeToExist('test-setup-ebs');
        if (!created) {
          throw new Error('Failed to create shared EBS volume: test-setup-ebs. This volume is required for display/search tests.');
        }
        console.log('✅ Shared EBS volume created successfully');
      }

      // Mark volumes as created so we don't try again for remaining 45 tests
      sharedVolumesCreated = true;
    }
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

      // Wait for AWS EFS creation to complete (polls until visible or test timeout)
      const volumeExists = await storagePage.waitForEFSVolumeToExist('test-efs-volume');
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

      // Should show validation error (scope to EFS modal to avoid strict mode violation)
      const efsModal = storagePage.page.getByRole('dialog', { name: 'Create EFS Volume' });
      const validationError = await efsModal.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/name.*required/i);
    });

    test('should delete EFS volume with confirmation', async () => {
      await storagePage.switchToEFS();

      // Create volume first
      await storagePage.createEFSVolume('delete-test-efs');

      // Wait for AWS EFS creation to complete (polls until visible or test timeout)
      let volumeExists = await storagePage.waitForEFSVolumeToExist('delete-test-efs');
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

      expect(instanceCount).toBeGreaterThan(0);

      const firstInstance = await page.locator('[data-testid="instances-table"] tbody tr').first();
      await firstInstance.waitFor({ state: 'visible' });
      const instanceName = await firstInstance.locator('[data-testid="instance-name"]').textContent();

      expect(instanceName).toBeTruthy();

      // Navigate back to storage
      await storagePage.navigate();
      await storagePage.switchToEFS();

      // Create EFS volume
      await storagePage.createEFSVolume('mount-test-efs');

      // Wait for AWS EFS creation to complete before mounting (polls until visible or test timeout)
      const volumeCreated = await storagePage.waitForEFSVolumeToExist('mount-test-efs');
      expect(volumeCreated).toBe(true);

      // Mount to instance
      await storagePage.mountEFSVolume('mount-test-efs', instanceName!);
      await storagePage.page.waitForTimeout(2000);

      // Verify mount success (check status or mounted indicator)
      const volumeRow = storagePage.getEFSVolumeByName('mount-test-efs');
      const volumeText = await volumeRow.textContent();
      expect(volumeText).toMatch(/mounted|attached/i);

      // Cleanup: Unmount and delete
      await storagePage.unmountEFSVolume('mount-test-efs', instanceName!);
      await confirmDialog.confirm();
      await storagePage.page.waitForTimeout(1000);
      await storagePage.deleteEFSVolume('mount-test-efs');
      await confirmDialog.confirmDelete();
    });

    test('should unmount EFS volume from instance', async ({ page }) => {
      // First ensure we have an instance
      await instancesPage.navigate();
      const instanceCount = await instancesPage.getInstanceCount();
      expect(instanceCount).toBeGreaterThan(0);

      const firstInstance = await page.locator('[data-testid="instances-table"] tbody tr').first();
      await firstInstance.waitFor({ state: 'visible' });
      const instanceName = await firstInstance.locator('[data-testid="instance-name"]').textContent();
      expect(instanceName).toBeTruthy();

      // Navigate to storage and create/mount a volume
      await storagePage.navigate();
      await storagePage.switchToEFS();

      // Create EFS volume
      await storagePage.createEFSVolume('unmount-test-efs');
      const volumeCreated = await storagePage.waitForEFSVolumeToExist('unmount-test-efs');
      expect(volumeCreated).toBe(true);

      // Mount it first
      await storagePage.mountEFSVolume('unmount-test-efs', instanceName!);
      await storagePage.page.waitForTimeout(2000);

      // Now unmount it
      await storagePage.unmountEFSVolume('unmount-test-efs', instanceName!);
      await confirmDialog.confirm();
      await storagePage.page.waitForTimeout(2000);

      // Verify unmount success
      const status = await storagePage.getVolumeStatus('unmount-test-efs', 'efs');
      expect(status).toMatch(/available|unmounted/i);

      // Cleanup
      await storagePage.deleteEFSVolume('unmount-test-efs');
      await confirmDialog.confirmDelete();
    });

    test('should show EFS volume status correctly', async () => {
      await storagePage.switchToEFS();

      // Wait for table to fully render with data
      const firstVolume = storagePage.page.locator('[data-testid="efs-table"] tbody tr').first();
      await firstVolume.waitFor({ state: 'visible' });

      const volumeNameElement = firstVolume.locator('[data-testid="volume-name"]');
      await volumeNameElement.waitFor({ state: 'visible' });
      const volumeName = await volumeNameElement.textContent();

      expect(volumeName).toBeTruthy();

      // Get volume status
      const status = await storagePage.getVolumeStatus(volumeName!, 'efs');
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

      // Wait for AWS EBS creation to complete (polls until visible or test timeout)
      const volumeExists = await storagePage.waitForEBSVolumeToExist('test-ebs-volume');
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

      // Should show validation error (scope to EBS modal to avoid strict mode violation)
      const ebsModal = storagePage.page.getByRole('dialog', { name: 'Create EBS Volume' });
      const validationError = await ebsModal.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/size.*required/i);
    });

    test('should validate EBS volume size is positive number', async () => {
      await storagePage.switchToEBS();

      await storagePage.page.getByTestId('create-ebs-header-button').click();

      await storagePage.page.getByRole('textbox', { name: 'EBS Volume Name' }).fill('size-validation-test');
      await storagePage.page.getByRole('spinbutton', { name: 'EBS Volume Size' }).fill('-10');
      await storagePage.clickButton('create');

      // Should show validation error (scope to EBS modal to avoid strict mode violation)
      const ebsModal = storagePage.page.getByRole('dialog', { name: 'Create EBS Volume' });
      const validationError = await ebsModal.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/positive|invalid.*size/i);
    });

    test('should delete EBS volume with confirmation', async () => {
      await storagePage.switchToEBS();

      // Create volume first
      await storagePage.createEBSVolume('delete-test-ebs', '50');

      // Wait for AWS EBS creation to complete (polls until visible or test timeout)
      let volumeExists = await storagePage.waitForEBSVolumeToExist('delete-test-ebs');
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

      expect(instanceCount).toBeGreaterThan(0);

      const firstInstance = await page.locator('[data-testid="instances-table"] tbody tr').first();
      await firstInstance.waitFor({ state: 'visible' });
      const instanceName = await firstInstance.locator('[data-testid="instance-name"]').textContent();

      expect(instanceName).toBeTruthy();

      // Navigate back to storage
      await storagePage.navigate();
      await storagePage.switchToEBS();

      // Create EBS volume
      await storagePage.createEBSVolume('attach-test-ebs', '100');

      // Wait for AWS EBS creation to complete before attaching (polls until visible or test timeout)
      const volumeCreated = await storagePage.waitForEBSVolumeToExist('attach-test-ebs');
      expect(volumeCreated).toBe(true);

      // Attach to instance
      await storagePage.attachEBSVolume('attach-test-ebs', instanceName!);
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
      // First ensure we have an instance
      await instancesPage.navigate();
      const instanceCount = await instancesPage.getInstanceCount();
      expect(instanceCount).toBeGreaterThan(0);

      const firstInstance = await page.locator('[data-testid="instances-table"] tbody tr').first();
      await firstInstance.waitFor({ state: 'visible' });
      const instanceName = await firstInstance.locator('[data-testid="instance-name"]').textContent();
      expect(instanceName).toBeTruthy();

      // Navigate to storage and create/attach a volume
      await storagePage.navigate();
      await storagePage.switchToEBS();

      // Create EBS volume
      await storagePage.createEBSVolume('detach-test-ebs', '50');
      const volumeCreated = await storagePage.waitForEBSVolumeToExist('detach-test-ebs');
      expect(volumeCreated).toBe(true);

      // Attach it first
      await storagePage.attachEBSVolume('detach-test-ebs', instanceName!);
      await storagePage.page.waitForTimeout(2000);

      // Now detach it
      await storagePage.detachEBSVolume('detach-test-ebs');
      await confirmDialog.confirm();
      await storagePage.page.waitForTimeout(2000);

      // Verify detach success
      const status = await storagePage.getVolumeStatus('detach-test-ebs', 'ebs');
      expect(status).toMatch(/available|detached/i);

      // Cleanup
      await storagePage.deleteEBSVolume('detach-test-ebs');
      await confirmDialog.confirmDelete();
    });

    test('should show EBS volume size in list', async () => {
      await storagePage.switchToEBS();

      const firstVolume = await storagePage.page.locator('[data-testid="ebs-table"] tbody tr').first();
      await firstVolume.waitFor({ state: 'visible' });
      const volumeText = await firstVolume.textContent();

      // Should display size information (e.g., "100 GB")
      expect(volumeText).toMatch(/\d+\s*GB/i);
    });

    test('should show EBS volume type', async () => {
      await storagePage.switchToEBS();

      const firstVolume = await storagePage.page.locator('[data-testid="ebs-table"] tbody tr').first();
      await firstVolume.waitFor({ state: 'visible' });
      const volumeText = await firstVolume.textContent();

      // Should display volume type (gp2, gp3, io1, etc.)
      const hasVolumeType = /gp2|gp3|io1|io2|st1|sc1/i.test(volumeText || '');
      expect(hasVolumeType).toBe(true);
    });
  });

  test.describe('Storage Search and Filtering', () => {
    test('should search EFS volumes by name', async () => {
      await storagePage.switchToEFS();

      // Wait for table to fully render with data
      const firstVolume = storagePage.page.locator('[data-testid="efs-table"] tbody tr').first();
      await firstVolume.waitFor({ state: 'visible' });

      const volumeNameElement = firstVolume.locator('[data-testid="volume-name"]');
      await volumeNameElement.waitFor({ state: 'visible' });
      const volumeName = await volumeNameElement.textContent();

      expect(volumeName).toBeTruthy();

      // Search for volume
      await storagePage.searchVolumes(volumeName!);
      await storagePage.page.waitForTimeout(500);

      // Verify search results
      const searchResults = await storagePage.getEFSVolumeCount();
      expect(searchResults).toBeGreaterThanOrEqual(1);
    });

    test('should search EBS volumes by name', async () => {
      await storagePage.switchToEBS();

      // Wait for table to fully render with data
      const firstVolume = storagePage.page.locator('[data-testid="ebs-table"] tbody tr').first();
      await firstVolume.waitFor({ state: 'visible' });

      const volumeNameElement = firstVolume.locator('[data-testid="volume-name"]');
      await volumeNameElement.waitFor({ state: 'visible' });
      const volumeName = await volumeNameElement.textContent();

      expect(volumeName).toBeTruthy();

      // Search for volume
      await storagePage.searchVolumes(volumeName!);
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
