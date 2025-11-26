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

      // Get the dialog context
      const dialog = projectsPage.page.locator('[role="dialog"]').first();

      // Try to submit without username - use data-testid to find wrapper, then find input inside
      await projectsPage.page.getByTestId('user-email-input').locator('input').fill('test@example.com');
      await projectsPage.page.getByTestId('user-fullname-input').locator('input').fill('Test User');

      // Click the primary Create button within the dialog
      await dialog.locator('button[class*="variant-primary"]').click();

      // Should show validation error
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/username.*required/i);

      // Cancel
      await dialog.getByRole('button', { name: /cancel/i }).click();
    });

    test('should validate email format', async () => {
      // Email validation is now implemented
      const uniqueUsername = `testuser-${Date.now()}`;

      await projectsPage.page.getByTestId('create-user-button').click();

      // Get the dialog context
      const dialog = projectsPage.page.locator('[role="dialog"]').first();

      // Use data-testid to find wrapper, then find input inside
      await projectsPage.page.getByTestId('user-username-input').locator('input').fill(uniqueUsername);
      await projectsPage.page.getByTestId('user-email-input').locator('input').fill('invalid-email');
      await projectsPage.page.getByTestId('user-fullname-input').locator('input').fill('Test User');

      // Click the primary Create button within the dialog
      await dialog.locator('button[class*="variant-primary"]').click();

      // Should show validation error
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/email.*invalid/i);

      // Cancel
      await dialog.getByRole('button', { name: /cancel/i }).click();
    });

    test('should prevent duplicate usernames', async () => {
      // Backend validation test - verify UI displays error properly
      const uniqueUsername = `testuser-${Date.now()}`;

      // Create first user
      await projectsPage.createUser(uniqueUsername, 'test1@example.com', 'Test User 1');

      // Try to create second user with same username
      await projectsPage.page.getByRole('button', { name: /create user/i }).click();
      await projectsPage.fillInput('username', uniqueUsername);
      await projectsPage.fillInput('email', 'test2@example.com');
      await projectsPage.fillInput('full name', 'Test User 2');
      await projectsPage.clickButton('create');

      // Should show duplicate error
      const dialog = projectsPage.page.locator('[role="dialog"]').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/already exists|duplicate/i);

      // Cleanup
      await projectsPage.clickButton('cancel');
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
      await projectsPage.page.waitForTimeout(500);

      // Verify SSH key count updated in table
      await projectsPage.page.reload(); // Refresh to see updated count
      await projectsPage.page.waitForTimeout(1000);
      const updatedRow = projectsPage.getUserByUsername(uniqueUsername);
      const rowText = await updatedRow.textContent();
      expect(rowText).toMatch(/1/); // Should show 1 SSH key

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });

    test.skip('should display existing SSH keys', async () => {
      // TODO: Requires SSH key listing UI
      const uniqueUsername = `sshtest-${Date.now()}`;

      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'SSH Test User');

      // Generate key (via API or UI)
      // ...

      // View user details
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByRole('button', { name: /actions/i });
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      await projectsPage.page.getByRole('menuitem', { name: /view details/i }).click();

      // Verify SSH keys section shows keys
      const sshSection = projectsPage.page.locator('text=/ssh keys/i');
      expect(await sshSection.isVisible()).toBe(true);

      // Cleanup
      await projectsPage.navigateToUsers();
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });
  });

  test.describe('Workspace Provisioning', () => {
    test.skip('should provision user on workspace', async () => {
      // TODO: Requires workspace provisioning UI and active instance
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

      // Select workspace
      const workspaceSelect = projectsPage.page.getByLabel(/workspace/i);
      await workspaceSelect.selectOption({ index: 0 }); // Select first available

      // Confirm provisioning
      await projectsPage.clickButton('provision');
      await projectsPage.page.waitForTimeout(1000);

      // Verify workspace count updated
      const updatedRow = projectsPage.getUserByUsername(uniqueUsername);
      const rowText = await updatedRow.textContent();
      expect(rowText).toMatch(/1.*workspace/i);

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });

    test.skip('should show provisioned workspaces for user', async () => {
      // TODO: Requires user details view
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

      // Verify workspaces section
      const workspacesSection = projectsPage.page.locator('text=/provisioned workspaces/i');
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

    test.skip('should warn when deleting user with active workspaces', async () => {
      // TODO: Requires user with provisioned workspaces
      const uniqueUsername = `active-delete-test-${Date.now()}`;

      // Create user and provision on workspace
      // ...

      // Try to delete
      await projectsPage.deleteUser(uniqueUsername);

      // Should show warning
      const warningText = await projectsPage.page.locator('text=/active.*workspace|provisioned/i').textContent();
      expect(warningText).toBeTruthy();

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
      // This prevents race condition where initial getUsers() overwrites optimistic updates
      await projectsPage.page.waitForLoadState('networkidle');
      await projectsPage.page.waitForTimeout(1000);

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

    test.skip('should show user statistics', async () => {
      // TODO: Verify stats cards at top of page
      await projectsPage.navigateToUsers();

      // Check for overview stats
      const statsSection = projectsPage.page.locator('text=/total users|active users|ssh keys|provisioned workspaces/i').first();
      expect(await statsSection.isVisible()).toBe(true);
    });

    test.skip('should display UID for each user', async () => {
      // TODO: Verify UID column in table
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

    test.skip('should filter users by status', async () => {
      // TODO: Requires filter functionality
      await projectsPage.navigateToUsers();

      // Apply filter
      const filterSelect = projectsPage.page.getByLabel(/status/i);
      await filterSelect.selectOption('active');

      // Verify only active users shown
      const rows = projectsPage.getUserRows();
      const count = await rows.count();

      for (let i = 0; i < count; i++) {
        const row = rows.nth(i);
        const text = await row.textContent();
        expect(text).toContain('Active');
      }
    });
  });

  test.describe('User Status Management', () => {
    test.skip('should view user status details', async () => {
      // TODO: Requires user status view
      const uniqueUsername = `status-test-${Date.now()}`;

      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Status Test');

      // View status
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const actionsButton = userRow.getByRole('button', { name: /actions/i });
      await actionsButton.click();
      await projectsPage.page.waitForTimeout(300);

      await projectsPage.page.getByRole('menuitem', { name: /user status/i }).click();

      // Verify status dialog
      const statusDialog = projectsPage.page.locator('[role="dialog"]').first();
      expect(await statusDialog.isVisible()).toBe(true);

      // Close
      await projectsPage.clickButton('close');

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });

    test.skip('should update user status', async () => {
      // TODO: Requires user status editing
      const uniqueUsername = `status-update-test-${Date.now()}`;

      await projectsPage.createUser(uniqueUsername, `${uniqueUsername}@example.com`, 'Test User');

      // Change status (e.g., suspend user)
      // ...

      // Verify status updated in table
      const userRow = projectsPage.getUserByUsername(uniqueUsername);
      const userText = await userRow.textContent();
      expect(userText).toContain('Suspended');

      // Cleanup
      await projectsPage.deleteUser(uniqueUsername);
      await projectsPage.clickButton('delete');
    });
  });
});
