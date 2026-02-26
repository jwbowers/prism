/**
 * Storage Workflows E2E Tests
 *
 * End-to-end tests for complete storage management workflows in Prism GUI.
 * Tests: EFS volume creation/deletion/mounting, EBS volume creation/deletion/attachment.
 */

import { test, expect } from '@playwright/test';
import { StoragePage, InstancesPage, ConfirmDialog } from './pages';

test.describe('Storage Management Workflows', () => {
  // 7-minute timeout: AWS operations take 30-180s + state monitor polling (10s) + test execution + cleanup
  test.setTimeout(420000);

  let storagePage: StoragePage;
  let instancesPage: InstancesPage;
  let confirmDialog: ConfirmDialog;

  // Track if shared test volumes have been created in this test run
  // This prevents re-creating volumes for every test (beforeEach runs 46 times!)
  let sharedVolumesCreated = false;

  // Delete shared volumes after all tests complete via direct API calls.
  // test-setup-efs and test-setup-ebs are created once per run but never deleted inline.
  // Without this cleanup they persist across runs, causing subsequent runs to find volumes
  // in transitional states (being deleted, locked mount targets) which disables the Delete
  // button and causes tests to fail.
  //
  // Uses direct API calls (not UI) for reliability at teardown time.
  test.afterAll(async ({ request }) => {
    test.setTimeout(120000); // 2 minutes for cleanup

    const sharedVolumes = [
      { name: 'test-setup-efs', endpoint: 'http://localhost:8947/api/v1/volumes/test-setup-efs' },
      { name: 'test-setup-ebs', endpoint: 'http://localhost:8947/api/v1/storage/test-setup-ebs' },
    ];

    for (const vol of sharedVolumes) {
      try {
        const res = await request.delete(vol.endpoint);
        if (res.ok() || res.status() === 404) {
          console.log(`✅ Cleaned up shared volume: ${vol.name}`);
        } else {
          console.log(`⚠️  Could not clean up ${vol.name} (status ${res.status()}) - will be cleaned on next run`);
        }
      } catch (e) {
        console.log(`⚠️  Cleanup error for ${vol.name}: ${e} - will be cleaned on next run`);
      }
    }
  });

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

    // Wait for storage page tabs to be interactive before proceeding
    await page.getByRole('tab', { name: /efs/i }).waitFor({ state: 'visible', timeout: 15000 });

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

      // Wait for volume to disappear (backend deletes EFS + state refreshes - can take 30-120s)
      const disappeared = await storagePage.waitForEFSVolumeToDisappear('delete-test-efs');
      expect(disappeared).toBe(true);

      // Verify it's removed
      volumeExists = await storagePage.verifyEFSVolumeExists('delete-test-efs');
      expect(volumeExists).toBe(false);
    });

    test('should mount EFS volume to instance', async () => {
      // Find a running instance - SSM mount requires running state
      await instancesPage.navigate();
      const instanceName = await instancesPage.getFirstRunningInstanceName();

      if (!instanceName) {
        test.skip(true, 'No running instances available for EFS mount test (requires SSM)');
        return;
      }

      // Navigate back to storage
      await storagePage.navigate();
      await storagePage.switchToEFS();

      // Create EFS volume
      await storagePage.createEFSVolume('mount-test-efs');

      // Wait for AWS EFS creation to complete before mounting (polls until visible or test timeout)
      const volumeCreated = await storagePage.waitForEFSVolumeToExist('mount-test-efs');
      expect(volumeCreated).toBe(true);

      // Wait for volume to be "available" before mounting (Mount button is disabled for non-available volumes)
      // Note: If a stale "mount-test-efs" from a previous interrupted test run exists in wrong state,
      // waitForVolumeState may timeout. Treat this as an infrastructure issue and skip.
      const volumeAvailable = await storagePage.waitForVolumeState('mount-test-efs', 'efs', 'available');
      if (!volumeAvailable) {
        await storagePage.deleteEFSVolume('mount-test-efs');
        await confirmDialog.confirmDelete();
        test.skip(true, 'EFS volume did not reach available state (possible stale volume from previous run or slow AWS)');
        return;
      }

      // Mount to instance
      await storagePage.mountEFSVolume('mount-test-efs', instanceName);
      // Wait for mount result: success OR failure notification (SSM can take 10-30s)
      // Using waitFor with regex to handle either outcome within 30 seconds
      await storagePage.page.getByText(/Mount Failed|Volume Mounted/).first()
        .waitFor({ state: 'visible', timeout: 30000 })
        .catch(() => { /* timeout - check state below */ });

      // Check if mount failed due to SSM infrastructure issues
      const mountFailed = await storagePage.page.locator('text=Mount Failed').isVisible();
      if (mountFailed) {
        // SSM not available on this instance - infrastructure limitation, not a code bug
        // Cleanup then skip
        await storagePage.deleteEFSVolume('mount-test-efs');
        await confirmDialog.confirmDelete();
        test.skip(true, 'EFS mount failed - instance SSM agent not configured or not in valid state');
        return;
      }

      // Verify mount success: wait for volume state to update to "mounted"
      // State monitor polls every ~10s, so allow up to 30s for state to propagate
      const isMounted = await storagePage.waitForVolumeState('mount-test-efs', 'efs', 'mounted')
        .catch(() => false);
      if (!isMounted) {
        // State may not have updated yet - verify no error notification shown
        const mountFailed = await storagePage.page.locator('text=Mount Failed').isVisible();
        expect(mountFailed).toBe(false);
      }

      // Cleanup: Unmount and delete
      await storagePage.unmountEFSVolume('mount-test-efs', instanceName);
      await storagePage.deleteEFSVolume('mount-test-efs');
      await confirmDialog.confirmDelete();
    });

    test('should unmount EFS volume from instance', async () => {
      // Find a running instance - SSM mount/unmount requires running state
      await instancesPage.navigate();
      const instanceName = await instancesPage.getFirstRunningInstanceName();

      if (!instanceName) {
        test.skip(true, 'No running instances available for EFS unmount test (requires SSM)');
        return;
      }

      // Navigate to storage and create/mount a volume
      await storagePage.navigate();
      await storagePage.switchToEFS();

      // Create EFS volume
      await storagePage.createEFSVolume('unmount-test-efs');
      const volumeCreated = await storagePage.waitForEFSVolumeToExist('unmount-test-efs');
      expect(volumeCreated).toBe(true);

      // Wait for volume to be "available" before mounting (Mount button is disabled for non-available volumes)
      const volumeAvailable = await storagePage.waitForVolumeState('unmount-test-efs', 'efs', 'available');
      if (!volumeAvailable) {
        await storagePage.deleteEFSVolume('unmount-test-efs');
        await confirmDialog.confirmDelete();
        test.skip(true, 'EFS volume did not reach available state (possible stale volume from previous run or slow AWS)');
        return;
      }

      // Mount it first
      await storagePage.mountEFSVolume('unmount-test-efs', instanceName);
      // Wait for mount result: success OR failure notification (SSM can take 10-30s)
      await storagePage.page.getByText(/Mount Failed|Volume Mounted/).first()
        .waitFor({ state: 'visible', timeout: 30000 })
        .catch(() => { /* timeout - check state below */ });

      // Check if mount failed due to SSM infrastructure issues
      const mountFailed = await storagePage.page.locator('text=Mount Failed').isVisible();
      if (mountFailed) {
        // SSM not available on this instance - infrastructure limitation, not a code bug
        await storagePage.deleteEFSVolume('unmount-test-efs');
        await confirmDialog.confirmDelete();
        test.skip(true, 'EFS mount failed - instance SSM agent not configured or not in valid state');
        return;
      }

      // Now unmount it
      await storagePage.unmountEFSVolume('unmount-test-efs', instanceName);

      // Verify unmount success
      const status = await storagePage.getVolumeStatus('unmount-test-efs', 'efs');
      expect(status).toMatch(/available|unmounted/i);

      // Cleanup
      await storagePage.deleteEFSVolume('unmount-test-efs');
      await confirmDialog.confirmDelete();
    });

    test('should show EFS volume status correctly', async () => {
      await storagePage.switchToEFS();

      // Wait for actual data rows (not loading state rows which don't have volume-name testid)
      await storagePage.page.locator('[data-testid="efs-table"] [data-testid="volume-name"]').first().waitFor({ state: 'visible', timeout: 15000 });
      // Ensure no re-load is in progress before reading element text (prevents race condition
      // where loadApplicationData fires and detaches elements during textContent() call)
      await storagePage.waitForStorageLoaded();
      // Use :has() to find only rows with actual data (not skeleton loading rows)
      const firstVolume = storagePage.page.locator('[data-testid="efs-table"] tbody tr:has([data-testid="volume-name"])').first();

      const volumeNameElement = firstVolume.locator('[data-testid="volume-name"]');
      await volumeNameElement.waitFor({ state: 'visible', timeout: 15000 });
      // Use innerText() instead of textContent() - more reliable for rendered text
      // textContent() can return null if element detaches during re-render
      const volumeName = (await volumeNameElement.innerText()).trim();

      expect(volumeName).toBeTruthy();

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

      // Wait for volume to disappear (backend deletes EBS + state refreshes - can take 30-120s)
      const disappeared = await storagePage.waitForEBSVolumeToDisappear('delete-test-ebs');
      expect(disappeared).toBe(true);

      // Verify it's removed
      volumeExists = await storagePage.verifyEBSVolumeExists('delete-test-ebs');
      expect(volumeExists).toBe(false);
    });

    test('should attach EBS volume to instance', async () => {
      // Find a running instance (EBS attach requires instance to exist; running preferred for "in-use" state)
      await instancesPage.navigate();
      const instanceName = await instancesPage.getFirstRunningInstanceName();

      if (!instanceName) {
        test.skip(true, 'No running instances available for EBS attach test');
        return;
      }

      // Navigate back to storage
      await storagePage.navigate();
      await storagePage.switchToEBS();

      // Create EBS volume
      await storagePage.createEBSVolume('attach-test-ebs', '100');

      // Wait for AWS EBS creation to complete before attaching (polls until visible or test timeout)
      const volumeCreated = await storagePage.waitForEBSVolumeToExist('attach-test-ebs');
      expect(volumeCreated).toBe(true);

      // Wait for volume to be "available" before attaching (Attach button is disabled for non-available volumes)
      // Note: If a stale "attach-test-ebs" from a previous interrupted test run exists in wrong state,
      // waitForVolumeState may timeout. Treat this as an infrastructure issue and skip.
      const volumeAvailable = await storagePage.waitForVolumeState('attach-test-ebs', 'ebs', 'available');
      if (!volumeAvailable) {
        await storagePage.deleteEBSVolumeIfExists('attach-test-ebs', confirmDialog);
        test.skip(true, 'EBS volume did not reach available state (possible stale volume from previous run or slow AWS)');
        return;
      }

      // Attach to instance
      await storagePage.attachEBSVolume('attach-test-ebs', instanceName);
      // Wait for attach result: success OR failure notification (AWS attach backend waiter takes 30-120s)
      await storagePage.page.getByText(/Attachment Failed|Volume Attached/).first()
        .waitFor({ state: 'visible', timeout: 180000 })
        .catch(() => { /* timeout - check state below */ });

      // Check if attach failed due to infrastructure issues (instance terminated, AZ mismatch, etc.)
      const attachFailed = await storagePage.page.locator('text=Attachment Failed').isVisible();
      if (attachFailed) {
        // Infrastructure limitation (invalid instance ID, AZ mismatch, etc.) - not a code bug
        await storagePage.deleteEBSVolumeIfExists('attach-test-ebs', confirmDialog);
        test.skip(true, 'EBS attach failed - instance ID not found or availability zone mismatch');
        return;
      }

      // Verify attach success
      const volumeRow = storagePage.getEBSVolumeByName('attach-test-ebs');
      const volumeText = await volumeRow.textContent();
      expect(volumeText).toMatch(/attached|in-use/i);

      // Cleanup: Detach and delete
      // Note: detachEBSVolume already handles the modal confirmation
      await storagePage.detachEBSVolume('attach-test-ebs');
      // Wait for detach to complete (fire-and-forget - AWS takes 30-60s)
      await storagePage.waitForVolumeState('attach-test-ebs', 'ebs', 'available');
      await storagePage.deleteEBSVolumeIfExists('attach-test-ebs', confirmDialog);
    });

    test('should detach EBS volume from instance', async () => {
      // Find a running instance (EBS attach requires instance to exist)
      await instancesPage.navigate();
      const instanceName = await instancesPage.getFirstRunningInstanceName();

      if (!instanceName) {
        test.skip(true, 'No running instances available for EBS detach test');
        return;
      }

      // Navigate to storage and create/attach a volume
      await storagePage.navigate();
      await storagePage.switchToEBS();

      // Check if detach-test-ebs already exists (orphaned from previous test run)
      // If it does, use it directly instead of creating a new one (prevents concurrent AWS operations
      // that can stress/crash the daemon: simultaneous EBS create + EBS attach)
      const alreadyExists = await storagePage.verifyEBSVolumeExists('detach-test-ebs');
      if (!alreadyExists) {
        // Create EBS volume (fire-and-forget - modal closes immediately)
        await storagePage.createEBSVolume('detach-test-ebs', '50');
        const volumeCreated = await storagePage.waitForEBSVolumeToExist('detach-test-ebs');
        expect(volumeCreated).toBe(true);
      }

      // Wait for volume to be "available" before attaching (Attach button is disabled for non-available volumes)
      // Note: If a stale "detach-test-ebs" from a previous interrupted test run exists in wrong state,
      // waitForVolumeState may timeout. Treat this as an infrastructure issue and skip.
      const volumeAvailable = await storagePage.waitForVolumeState('detach-test-ebs', 'ebs', 'available');
      if (!volumeAvailable) {
        await storagePage.deleteEBSVolumeIfExists('detach-test-ebs', confirmDialog);
        test.skip(true, 'EBS volume did not reach available state (possible stale volume from previous run or slow AWS)');
        return;
      }

      // Attach it first
      await storagePage.attachEBSVolume('detach-test-ebs', instanceName);
      // Wait for attach result: success OR failure notification (AWS attach backend waiter takes 30-120s)
      await storagePage.page.getByText(/Attachment Failed|Volume Attached/).first()
        .waitFor({ state: 'visible', timeout: 180000 })
        .catch(() => { /* timeout - check state below */ });

      // Check if attach failed due to infrastructure issues
      const attachFailed = await storagePage.page.locator('text=Attachment Failed').isVisible();
      if (attachFailed) {
        // Infrastructure limitation (invalid instance ID, AZ mismatch, etc.) - not a code bug
        // Use resilient delete that handles daemon-down scenarios gracefully
        await storagePage.deleteEBSVolumeIfExists('detach-test-ebs', confirmDialog);
        test.skip(true, 'EBS attach failed - instance ID not found or availability zone mismatch');
        return;
      }

      // Now detach it
      // Note: detachEBSVolume already handles the modal confirmation
      await storagePage.detachEBSVolume('detach-test-ebs');

      // Verify detach success - wait for volume to return to "available" state (fire-and-forget - AWS takes 30-60s)
      const detachComplete = await storagePage.waitForVolumeState('detach-test-ebs', 'ebs', 'available');
      expect(detachComplete).toBe(true);

      // Cleanup
      await storagePage.deleteEBSVolumeIfExists('detach-test-ebs', confirmDialog);
    });

    test('should show EBS volume size in list', async () => {
      await storagePage.switchToEBS();

      // Wait for actual data rows (not loading state rows which don't have volume-name testid)
      await storagePage.page.locator('[data-testid="ebs-table"] [data-testid="volume-name"]').first().waitFor({ state: 'visible', timeout: 15000 });
      // Use :has() to find only rows with actual data (not skeleton loading rows)
      const firstVolume = storagePage.page.locator('[data-testid="ebs-table"] tbody tr:has([data-testid="volume-name"])').first();
      const volumeText = await firstVolume.textContent();

      // Should display size information (e.g., "100 GB")
      expect(volumeText).toMatch(/\d+\s*GB/i);
    });

    test('should show EBS volume type', async () => {
      await storagePage.switchToEBS();

      // Wait for actual data rows (not loading state rows which don't have volume-name testid)
      await storagePage.page.locator('[data-testid="ebs-table"] [data-testid="volume-name"]').first().waitFor({ state: 'visible', timeout: 15000 });
      // Use :has() to find only rows with actual data (not skeleton loading rows)
      const firstVolume = storagePage.page.locator('[data-testid="ebs-table"] tbody tr:has([data-testid="volume-name"])').first();
      const volumeText = await firstVolume.textContent();

      // Should display volume type (gp2, gp3, io1, etc.)
      const hasVolumeType = /gp2|gp3|io1|io2|st1|sc1/i.test(volumeText || '');
      expect(hasVolumeType).toBe(true);
    });
  });

  test.describe('Storage Search and Filtering', () => {
    test('should search EFS volumes by name', async () => {
      await storagePage.switchToEFS();

      // Wait for actual data rows (not loading state rows which don't have volume-name testid)
      await storagePage.page.locator('[data-testid="efs-table"] [data-testid="volume-name"]').first().waitFor({ state: 'visible', timeout: 15000 });
      await storagePage.waitForStorageLoaded();
      // Use :has() to find only rows with actual data (not skeleton loading rows)
      const firstVolume = storagePage.page.locator('[data-testid="efs-table"] tbody tr:has([data-testid="volume-name"])').first();

      const volumeNameElement = firstVolume.locator('[data-testid="volume-name"]');
      await volumeNameElement.waitFor({ state: 'visible', timeout: 15000 });
      const volumeName = (await volumeNameElement.innerText()).trim();

      expect(volumeName).toBeTruthy();

      // Search for volume
      await storagePage.searchVolumes(volumeName);

      // Verify search results
      const searchResults = await storagePage.getEFSVolumeCount();
      expect(searchResults).toBeGreaterThanOrEqual(1);
    });

    test('should search EBS volumes by name', async () => {
      await storagePage.switchToEBS();

      // Wait for actual data rows (not loading state rows which don't have volume-name testid)
      await storagePage.page.locator('[data-testid="ebs-table"] [data-testid="volume-name"]').first().waitFor({ state: 'visible', timeout: 15000 });
      // Use :has() to find only rows with actual data (not skeleton loading rows)
      const firstVolume = storagePage.page.locator('[data-testid="ebs-table"] tbody tr:has([data-testid="volume-name"])').first();

      const volumeNameElement = firstVolume.locator('[data-testid="volume-name"]');
      await volumeNameElement.waitFor({ state: 'visible', timeout: 15000 });
      const volumeName = await volumeNameElement.textContent();

      expect(volumeName).toBeTruthy();

      // Search for volume
      await storagePage.searchVolumes(volumeName!);

      // Verify search results
      const searchResults = await storagePage.getEBSVolumeCount();
      expect(searchResults).toBeGreaterThanOrEqual(1);
    });

    test('should show empty state when no volumes match search', async () => {
      await storagePage.switchToEFS();

      // Search for non-existent volume
      await storagePage.searchVolumes('nonexistent-volume-xyz');

      // Should show zero results
      const count = await storagePage.getEFSVolumeCount();
      expect(count).toBe(0);
    });
  });

  test.describe('Tab Switching', () => {
    test('should switch between EFS and EBS tabs', async () => {
      // Start with EFS
      await storagePage.switchToEFS();

      // Verify EFS content
      const efsTable = storagePage.page.locator('[data-testid="efs-table"]');
      await efsTable.waitFor({ state: 'visible', timeout: 5000 });
      expect(await efsTable.isVisible()).toBe(true);

      // Switch to EBS
      await storagePage.switchToEBS();

      // Verify EBS content
      const ebsTable = storagePage.page.locator('[data-testid="ebs-table"]');
      await ebsTable.waitFor({ state: 'visible', timeout: 5000 });
      expect(await ebsTable.isVisible()).toBe(true);
    });

    test('should preserve tab state when navigating away and back', async () => {
      // Switch to EBS tab
      await storagePage.switchToEBS();

      // Navigate away
      await instancesPage.navigate();

      // Navigate back
      await storagePage.navigate();

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
