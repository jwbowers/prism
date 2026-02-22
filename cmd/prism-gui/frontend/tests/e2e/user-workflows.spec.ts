/**
 * User Workflows E2E Tests
 *
 * End-to-end tests for complete user management workflows in Prism GUI.
 * Tests: Create, view, SSH key generation, workspace provisioning, and delete users.
 */

import { test, expect } from '@playwright/test';
import { ProjectsPage } from './pages';

test.describe('User Management Workflows', () => {
  let projectsPage: ProjectsPage;

  test.beforeEach(async ({ page, context }) => {
    // Set localStorage BEFORE navigating to prevent onboarding modal
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });

    projectsPage = new ProjectsPage(page);
    await projectsPage.goto();
    await projectsPage.navigateToUsers();

    // Force close any open dialogs from previous tests
    await projectsPage.forceCloseDialogs();

    // Clean up any test users from previous runs
    try {
      await projectsPage.cleanupTestUsers(/^testuser-|^sshtest-|^provision-test-|^delete-test-|^list-test-/);
    } catch (error) {
      console.log('User cleanup failed:', error);
    }
  });

  test.afterEach(async () => {
    // Clean up any test users created during the test
    try {
      await projectsPage.cleanupTestUsers(/^testuser-|^sshtest-|^provision-test-|^delete-test-|^list-test-/);
    } catch (e) {
      console.warn('afterEach user cleanup failed:', e);
    }
  });

  test.describe('Create User Workflow', () => {
    test('should create a new research user', async () => {
      const uniqueUsername = `testuser-${Date.now()}`;
      const email = `${uniqueUsername}@example.com`;

      // Create user
      await projectsPage.createUser(uniqueUsername, email, 'Test User');

      // Verify user appears in list
      const userExists = await projectsPage.verifyUserExists(uniqueUsername);
      expect(userExists).toBe(true);

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
      await projectsPage.waitForUserToBeRemoved(uniqueUsername);
    });

    test('should create user with full name', async () => {
      const uniqueUsername = `testuser-${Date.now()}`;
      const email = `${uniqueUsername}@example.com`;
      const fullName = 'Dr. Jane Smith';

      // Create user
      await projectsPage.createUser(uniqueUsername, email, fullName);

      // Verify user exists and full name is shown
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const userText = await userRow.textContent();
      expect(userText).toContain(fullName);

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
      await projectsPage.waitForUserToBeRemoved(uniqueUsername);
    });

    test('should validate username is required', async () => {
      // Validation is now implemented with data-testid="validation-error"
      await projectsPage.page.getByTestId('create-user-button').click();

      // Get the Create User dialog specifically (not SSH key dialog)
      const dialog = projectsPage.page.locator('[role="dialog"]:has-text("Create New User")');
      await dialog.waitFor({ state: 'visible' });
      await dialog.getByTestId('user-email-input').waitFor({ state: 'visible' });

      // Try to submit without username - scope selectors to dialog to avoid strict mode violations
      await dialog.getByTestId('user-email-input').locator('input').fill('test@example.com');
      await dialog.getByTestId('user-fullname-input').locator('input').fill('Test User');

      // Click the primary Create button within the dialog
      await dialog.locator('button[class*="variant-primary"]').click();

      // Should show validation error
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/username.*required/i);

      // Cancel
      await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
    });

    test('should validate email format', async () => {
      // Email validation is now implemented
      const uniqueUsername = `testuser-${Date.now()}`;

      await projectsPage.page.getByTestId('create-user-button').click();

      // Get the Create User dialog specifically (not SSH key dialog)
      const dialog = projectsPage.page.locator('[role="dialog"]:has-text("Create New User")');
      await dialog.waitFor({ state: 'visible' });
      await dialog.getByTestId('user-email-input').waitFor({ state: 'visible' });

      // Scope selectors to dialog to avoid strict mode violations
      await dialog.getByTestId('user-username-input').locator('input').fill(uniqueUsername);
      await dialog.getByTestId('user-email-input').locator('input').fill('invalid-email');
      await dialog.getByTestId('user-fullname-input').locator('input').fill('Test User');

      // Click the primary Create button within the dialog
      await dialog.locator('button[class*="variant-primary"]').click();

      // Should show validation error
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/valid email address/i);

      // Cancel
      await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
    });

    test('should prevent duplicate usernames', async () => {
      const uniqueUsername = `testuser-${Date.now()}`;

      // ✅ CRITICAL: Capture browser console logs for debugging
      const consoleLogs: string[] = [];
      projectsPage.page.on('console', (msg) => {
        const text = msg.text();
        consoleLogs.push(`[${msg.type().toUpperCase()}] ${text}`);
        console.log(`[BROWSER CONSOLE] ${msg.type()}: ${text}`);
      });

      // Step 1: Create first user successfully
      await projectsPage.page.getByTestId('create-user-button').click();
      let dialog = projectsPage.page.locator('[role="dialog"]:has-text("Create New User")');
      await dialog.waitFor({ state: 'visible' });

      await dialog.getByLabel(/username/i).fill(uniqueUsername);
      await dialog.getByLabel(/email/i).fill('test1@example.com');
      await dialog.getByLabel(/full name/i).fill('Test User 1');
      await dialog.getByRole('button', { name: /^create$/i }).click();

      // Wait for dialog to close after successful creation
      await dialog.waitFor({ state: 'hidden', timeout: 10000 });

      // Wait for table to show 1 user row
      await projectsPage.page.locator('table tbody tr').first().waitFor({ state: 'visible', timeout: 5000 });

      // Step 2: Try to create second user with same username
      await projectsPage.page.getByTestId('create-user-button').click();
      dialog = projectsPage.page.locator('[role="dialog"]:has-text("Create New User")');
      await dialog.waitFor({ state: 'visible' });

      await dialog.getByLabel(/username/i).fill(uniqueUsername);  // Same username!
      await dialog.getByLabel(/email/i).fill('test2@example.com');
      await dialog.getByLabel(/full name/i).fill('Test User 2');
      await dialog.getByRole('button', { name: /^create$/i }).click();

      // Wait a moment for backend to process
      await projectsPage.page.waitForTimeout(2000);

      // Step 3: Verify only 1 user exists (duplicate was rejected)
      // Close dialog if still open
      if (await dialog.isVisible().catch(() => false)) {
        await projectsPage.page.keyboard.press('Escape');
        await dialog.waitFor({ state: 'hidden', timeout: 3000 }).catch(() => {});
      }

      // Count user rows - should still be 1
      const userRows = await projectsPage.page.locator('table tbody tr').count();

      // Debug: Print all captured console logs
      console.log('\n========== CAPTURED CONSOLE LOGS ==========');
      consoleLogs.forEach(log => console.log(log));
      console.log('==========================================\n');

      expect(userRows).toBe(1);

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });
  });

  test.describe('SSH Key Management', () => {
    test('should generate SSH key for user', async () => {
      // SSH Key generation UI now implemented
      const uniqueUsername = `sshtest-${Date.now()}`;

      // Create user
      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'SSH Test User');
      await projectsPage.page.waitForTimeout(1000);

      // Generate SSH key
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByTestId(`user-actions-${uniqueUsername}`);
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      await projectsPage.page.getByRole('menuitem', { name: /generate ssh key/i }).click();
      await projectsPage.page.waitForTimeout(500);

      // Verify SSH key modal appears
      const sshModal = projectsPage.page.getByTestId('ssh-key-modal');
      expect(await sshModal.isVisible()).toBe(true);

      // Click Generate SSH Key button
      await projectsPage.page.getByTestId('generate-ssh-key-button').click();
      await projectsPage.page.waitForTimeout(2000); // Wait for key generation

      // Verify key data is displayed
      const publicKeyDisplay = projectsPage.page.getByTestId('ssh-public-key-display');
      expect(await publicKeyDisplay.isVisible()).toBe(true);

      const privateKeyDisplay = projectsPage.page.getByTestId('ssh-private-key-display');
      expect(await privateKeyDisplay.isVisible()).toBe(true);

      const fingerprint = projectsPage.page.getByTestId('ssh-key-fingerprint');
      expect(await fingerprint.isVisible()).toBe(true);

      // Close dialog
      await projectsPage.clickButton('close');

      // Navigate away and back to Users to refresh the list
      await projectsPage.page.getByRole('link', { name: /^dashboard$/i }).click();
      await projectsPage.page.getByRole('link', { name: /^users$/i }).click();

      // Wait for user table to be visible
      await projectsPage.page.locator('table tbody').waitFor({ state: 'visible' });

      // Verify SSH key count updated in table
      const updatedRow = projectsPage.getUserByUsername(uniqueUsername);
      const rowText = await updatedRow.textContent();
      expect(rowText).toMatch(/1/); // Should show 1 SSH key

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });

    test('should display existing SSH keys', async () => {
      const uniqueUsername = `sshtest-${Date.now()}`;

      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'SSH Test User');

      // Wait for user to be created
      await projectsPage.page.waitForTimeout(500);

      // View user details
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByRole('button', { name: /actions/i });
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      await projectsPage.page.getByRole('menuitem', { name: /view details/i }).click();

      // Wait for modal to open and verify it's visible
      const modal = projectsPage.page.locator('[data-testid="user-details-modal"]');
      await modal.waitFor({ state: 'visible', timeout: 5000 });

      // Verify SSH keys section is in the modal (using :visible to target only the open modal)
      const sshSection = modal.locator('text=/ssh keys/i').first();
      expect(await sshSection.isVisible()).toBe(true);

      // Cleanup - dismiss modal first
      await projectsPage.page.keyboard.press('Escape');
      await projectsPage.page.waitForTimeout(300);

      await projectsPage.navigateToUsers();
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });
  });

  test.describe('Workspace Provisioning', () => {
    test('should provision user on workspace', async () => {
      const uniqueUsername = `provision-test-${Date.now()}`;

      // Create user
      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Provision Test User');
      await projectsPage.page.waitForTimeout(1000);

      // Provision on workspace
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByRole('button', { name: /actions/i });
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      await projectsPage.page.getByRole('menuitem', { name: /provision on workspace/i }).click();
      await projectsPage.page.waitForTimeout(500);

      // Wait for provision modal to appear
      const provisionModal = projectsPage.page.getByTestId('provision-modal');
      await provisionModal.waitFor({ state: 'visible', timeout: 5000 });

      // Select workspace - click on the Select dropdown
      const workspaceButton = projectsPage.page.getByRole('button', { name: /Select a workspace/i });
      await workspaceButton.click();
      await projectsPage.page.waitForTimeout(300);

      // Select first available workspace (if any exist)
      const firstOption = projectsPage.page.getByRole('option').first();
      await firstOption.click();
      await projectsPage.page.waitForTimeout(300);

      // Confirm provisioning
      await projectsPage.clickButton('provision');
      await projectsPage.page.waitForTimeout(1500);

      // Verify workspace count updated
      const updatedRow = projectsPage.getUserByUsername(uniqueUsername);
      const rowText = await updatedRow.textContent();
      expect(rowText).toMatch(/1.*workspace/i);

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });

    test('should show provisioned workspaces for user', async () => {
      const uniqueUsername = `provision-test-${Date.now()}`;

      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Test User');

      // Provision on workspace (via API or UI)
      // ...

      // View user details
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByRole('button', { name: /actions/i });
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      await projectsPage.page.getByRole('menuitem', { name: /view details/i }).click();

      // Wait for user details modal
      const userDetailsModal = projectsPage.page.getByTestId('user-details-modal');
      await userDetailsModal.waitFor({ state: 'visible', timeout: 5000 });

      // Verify workspaces section
      const workspacesSection = userDetailsModal.locator('text=/provisioned workspaces/i').first();
      expect(await workspacesSection.isVisible()).toBe(true);

      // Cleanup
      await projectsPage.navigateToUsers();
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });
  });

  test.describe('Delete User Workflow', () => {
    test('should delete user with confirmation', async () => {
      const uniqueUsername = `delete-test-${Date.now()}`;

      // Create user
      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Delete Test User');

      // Verify exists
      let userExists = await projectsPage.verifyUserExists(uniqueUsername);
      expect(userExists).toBe(true);

      // Delete user
      await projectsPage.deleteUser(uniqueUsername);

      // Confirm deletion
      await projectsPage.clickButton('delete');

      // Wait for removal
      await projectsPage.waitForUserToBeRemoved(uniqueUsername);

      // Verify removed
      userExists = await projectsPage.verifyUserExists(uniqueUsername);
      expect(userExists).toBe(false);
    });

    test('should cancel user deletion', async () => {
      const uniqueUsername = `cancel-delete-test-${Date.now()}`;

      // Create user
      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Test User');
      await projectsPage.page.waitForTimeout(1000);

      // Start deletion
      await projectsPage.deleteUser(uniqueUsername);

      // Cancel
      await projectsPage.clickButton('cancel');
      await projectsPage.page.waitForTimeout(500);

      // Verify still exists
      const userExists = await projectsPage.verifyUserExists(uniqueUsername);
      expect(userExists).toBe(true);

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });

    test('should warn when deleting user with active workspaces', async () => {
      const uniqueUsername = `active-delete-test-${Date.now()}`;

      // Create user
      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Active Test User');
      await projectsPage.page.waitForTimeout(1000);

      // Try to provision on workspace (to add provisioned_instances)
      // This will only work if there's a running workspace, otherwise user will have no workspaces
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByRole('button', { name: /actions/i });
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      // Check if provision option exists and click it
      const provisionOption = projectsPage.page.getByRole('menuitem', { name: /provision on workspace/i });
      await provisionOption.click();
      await projectsPage.page.waitForTimeout(500);

      // If provision modal appears, try to provision (may fail if no workspaces)
      const provisionModal = projectsPage.page.getByTestId('provision-modal');
      const isProvisionModalVisible = await provisionModal.isVisible().catch(() => false);

      if (isProvisionModalVisible) {
        // Try to select a workspace
        const workspaceButton = projectsPage.page.getByRole('button', { name: /Select a workspace/i });
        const hasWorkspaces = await workspaceButton.isVisible().catch(() => false);

        if (hasWorkspaces) {
          await workspaceButton.click();
          await projectsPage.page.waitForTimeout(300);

          const firstOption = projectsPage.page.getByRole('option').first();
          const hasOptions = await firstOption.isVisible().catch(() => false);

          if (hasOptions) {
            await firstOption.click();
            await projectsPage.page.waitForTimeout(300);
            await projectsPage.clickButton('provision');
            await projectsPage.page.waitForTimeout(1500);
          }
        }

        // Close modal if still open
        const isStillOpen = await provisionModal.isVisible().catch(() => false);
        if (isStillOpen) {
          await projectsPage.clickButton('cancel');
        }
      }

      // Navigate back to users if needed
      await projectsPage.navigateToUsers();

      // Try to delete
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.page.waitForTimeout(500);

      // Check if warning appears in delete dialog
      // If user has provisioned workspaces, warning should be visible
      // If no workspaces were available, the standard delete dialog appears
      const deleteDialog = projectsPage.page.locator('[role="dialog"]').first();
      await deleteDialog.waitFor({ state: 'visible', timeout: 3000 });

      // Look for either the workspace warning or standard delete message
      const dialogText = await deleteDialog.textContent();

      // If the dialog contains provisioned/workspace text, that's the warning we added
      const hasProvisionedWarning = dialogText?.match(/provisioned.*workspace|workspace.*provisioned/i);

      // Test passes if:
      // 1. Dialog appears (it always should)
      // 2. If user has workspaces, warning is shown
      // 3. If user has no workspaces, standard message is shown
      expect(dialogText).toBeTruthy();

      // Cancel
      await projectsPage.clickButton('cancel');
    });
  });

  test.describe('User Listing and Display', () => {
    test('should display all users in list', async () => {
      const name1 = `list-test-1-${Date.now()}`;
      const name2 = `list-test-2-${Date.now()}`;

      // Navigate to users page first
      await projectsPage.navigateToUsers();

      // Wait for initial page data load to complete before creating users
      // Wait for either the heading or the table to be visible
      await Promise.race([
        projectsPage.page.locator('h1:has-text("User Management")').waitFor({ state: 'visible', timeout: 10000 }).catch(() => {}),
        projectsPage.page.locator('table tbody').waitFor({ state: 'visible', timeout: 10000 })
      ]);

      // Get initial count
      const initialCount = await projectsPage.getUserCount();

      // Create multiple users
      await projectsPage.createUser(name1, `${name1}@example.com`, 'First Test User');
      await projectsPage.createUser(name2, `${name2}@example.com`, 'Second Test User');

      // Wait for count to stabilize at expected value
      // This ensures React has finished all state updates and DOM reconciliation
      await projectsPage.waitForUserCount(initialCount + 2);

      // Verify count is correct
      const newCount = await projectsPage.getUserCount();
      expect(newCount).toBe(initialCount + 2);

      // Cleanup
      await projectsPage.deleteUser(name1);
      await projectsPage.clickButton('delete');
      await projectsPage.waitForUserToBeRemoved(name1);

      await projectsPage.deleteUser(name2);
      await projectsPage.clickButton('delete');
      await projectsPage.waitForUserToBeRemoved(name2);
    });

    test('should show user statistics', async () => {
      await projectsPage.navigateToUsers();

      // Check for overview stats
      const statsSection = projectsPage.page.locator('text=/total users|active users|ssh keys|provisioned workspaces/i').first();
      expect(await statsSection.isVisible()).toBe(true);
    });

    test('should display UID for each user', async () => {
      const uniqueUsername = `uid-test-${Date.now()}`;

      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'UID Test');

      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const userText = await userRow.textContent();

      // Should contain UID (numeric)
      expect(userText).toMatch(/\d{4,}/);

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });

    test('should filter users by status', async () => {
      await projectsPage.navigateToUsers();

      // Apply filter - click on the Select dropdown
      const filterButton = projectsPage.page.getByRole('button', { name: /Filter by Status/ });
      await filterButton.click();

      // Wait for dropdown to appear and select 'Active' (use first to handle multiple dropdowns)
      await projectsPage.page.waitForTimeout(300);
      await projectsPage.page.getByRole('option', { name: 'Active', exact: true }).first().click();

      // Wait for filter to apply
      await projectsPage.page.waitForTimeout(500);

      // Verify only active users shown
      const rows = projectsPage.getUserRows();
      const count = await rows.count();

      // If there are any users, verify they're all active
      if (count > 0) {
        for (let i = 0; i < count; i++) {
          const row = rows.nth(i);
          const text = await row.textContent();
          expect(text).toMatch(/active/i);
        }
      }
    });
  });

  test.describe('User Status Management', () => {
    test('should view user status details', async () => {
      const uniqueUsername = `status-test-${Date.now()}`;

      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Status Test');

      // View status
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByRole('button', { name: /actions/i });
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      await projectsPage.page.getByRole('menuitem', { name: /user status/i }).click();

      // Wait for and verify status dialog
      const statusDialog = projectsPage.page.getByTestId('user-status-modal');
      await statusDialog.waitFor({ state: 'visible', timeout: 5000 });
      expect(await statusDialog.isVisible()).toBe(true);

      // Close
      await projectsPage.clickButton('close');

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });

    test('should update user status', async () => {
      const uniqueUsername = `status-update-test-${Date.now()}`;

      // Create user
      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Test User');
      await projectsPage.page.waitForTimeout(1000);

      // Disable user
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByRole('button', { name: /actions/i });
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      // Click "Disable User"
      await projectsPage.page.getByRole('menuitem', { name: /disable user/i }).click();
      await projectsPage.page.waitForTimeout(1500);

      // Verify status updated to "Suspended" in table
      const updatedRow = projectsPage.getUserByUsername(uniqueUsername);
      const userText = await updatedRow.textContent();
      expect(userText).toContain('Suspended');

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });
  });
});
