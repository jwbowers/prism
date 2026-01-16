/**
 * Projects Page Object
 *
 * Page object for the Projects tab in Prism GUI.
 * Handles project management, user management, and invitation workflows.
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class ProjectsPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to Projects tab
   */
  async navigate() {
    await this.navigateToTab('projects');
    await this.waitForLoadingComplete();
  }

  /**
   * Navigate to Users tab
   */
  async navigateToUsers() {
    await this.navigateToTab('users');
    await this.waitForLoadingComplete();

    // Wait for the Create User button to be visible (deterministic wait)
    // This ensures the UserManagementView has fully rendered
    await this.page.waitForSelector('[data-testid="create-user-button"]', {
      state: 'visible',
      timeout: 10000
    });
  }

  /**
   * Navigate to Invitations tab
   */
  async navigateToInvitations() {
    await this.navigateToTab('invitations');
    await this.waitForLoadingComplete();
  }

  // ==================== PROJECT MANAGEMENT ====================

  /**
   * Get all project rows from the projects table
   */
  getProjectRows(): Locator {
    return this.page.locator('[data-testid="projects-table"] tbody tr, table tbody tr').filter({ hasText: /.+/ });
  }

  /**
   * Get project by name
   */
  getProjectByName(name: string): Locator {
    return this.page.locator(`tr:has-text("${name}")`).first();
  }

  /**
   * Create a new project
   */
  async createProject(name: string, description: string, budget?: number) {
    await this.navigate();

    // Click Create Project button using test ID
    await this.page.getByTestId('create-project-button').click();

    // Wait for Create Project dialog to appear by name (deterministic, avoids strict mode violation)
    await this.page.getByRole('dialog', { name: /create.*project/i }).waitFor({ state: 'visible', timeout: 5000 });

    // Fill form fields using data-testids
    await this.fillInput('project name', name);

    // Use data-testid for description (Cloudscape wraps textarea)
    await this.page.getByTestId('project-description-input').locator('textarea').fill(description);

    if (budget !== undefined) {
      // Use data-testid for budget input (Cloudscape wraps input)
      await this.page.getByTestId('project-budget-input').locator('input').fill(budget.toString());
    }

    // Submit - use data-testid for the Create button in the dialog footer
    const createButton = this.page.getByTestId('create-project-submit-button');

    // Ensure button is ready and clickable
    await createButton.waitFor({ state: 'visible' });

    // Click the button to submit the form
    await createButton.click();

    // Wait for dialog to close by checking that it's no longer visible
    await this.page.locator('[role="dialog"]:has-text("Create New Project")').waitFor({ state: 'hidden', timeout: 10000 });

    // Wait for the project to appear in the table by looking for the project name
    // This ensures the POST succeeded and the table refreshed
    await this.page.getByRole('cell', { name, exact: true }).waitFor({ state: 'visible', timeout: 10000 });

    // Extra wait to ensure Cloudscape table has fully rendered
    await this.page.waitForTimeout(500);
  }

  /**
   * Delete a project - opens confirmation modal, test must confirm deletion
   * @param projectName Name of the project to delete
   */
  async deleteProject(projectName: string) {
    await this.navigate();

    // Wait for the project to appear in the table before trying to interact with it
    await this.page.getByRole('cell', { name: projectName, exact: true }).waitFor({ state: 'visible', timeout: 10000 });

    const projectRow = this.getProjectByName(projectName);

    // Click actions dropdown
    const actionsButton = projectRow.getByRole('button', { name: /actions/i });
    await actionsButton.waitFor({ state: 'visible', timeout: 5000 });
    await actionsButton.click();

    // Wait for menu to appear (deterministic)
    const deleteOption = this.page.getByRole('menuitem', { name: /delete/i });
    await deleteOption.waitFor({ state: 'visible', timeout: 3000 });
    await deleteOption.click();

    // Wait for delete confirmation modal dialog to be visible
    await this.page.getByRole('dialog', { name: /delete/i }).waitFor({ state: 'visible', timeout: 5000 });
    await this.page.getByTestId('confirm-delete-button').waitFor({ state: 'visible', timeout: 5000 });
  }

  /**
   * Confirm project deletion by clicking delete button in modal
   * Note: Projects don't require name confirmation, just a simple click
   */
  async confirmDeletion() {
    // deleteProject() already waited for modal and button to be visible
    const deleteButton = this.page.getByTestId('confirm-delete-button');
    await deleteButton.click();
  }

  /**
   * View project details
   */
  async viewProjectDetails(projectName: string) {
    await this.navigate();

    const projectRow = this.getProjectByName(projectName);
    const actionsButton = projectRow.getByRole('button', { name: /actions/i });
    await actionsButton.click();

    // Wait for menu to appear
    const viewDetailsOption = this.page.getByRole('menuitem', { name: /view details/i });
    await viewDetailsOption.waitFor({ state: 'visible', timeout: 3000 });
    await viewDetailsOption.click();
    await this.waitForLoadingComplete();
  }

  /**
   * Verify project exists
   * Note: Does NOT navigate - assumes we're already on the Projects page
   * Waits/polls for the project to appear in the table (handles async UI updates)
   */
  async verifyProjectExists(name: string): Promise<boolean> {
    // Don't navigate - just check if the project exists in the current table
    // Poll for the project to appear (table may take time to refresh after creation)
    const maxWait = 10000; // 10 seconds
    const startTime = Date.now();

    while (Date.now() - startTime < maxWait) {
      const project = this.getProjectByName(name);
      if (await this.elementExists(project)) {
        return true; // Project found!
      }
      // Wait a bit before checking again
      await this.page.waitForTimeout(200);
    }

    return false; // Project not found after waiting
  }

  /**
   * Get project count
   */
  async getProjectCount(): Promise<number> {
    await this.navigate();
    return await this.getProjectRows().count();
  }

  /**
   * Wait for project to be removed from list
   */
  async waitForProjectToBeRemoved(projectName: string, timeout: number = 30000) {
    await this.waitForDialogClose();

    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      try {
        const exists = await this.verifyProjectExists(projectName);
        if (!exists) {
          return; // Success - project removed
        }
      } catch (error) {
        // Continue polling
      }
      // Wait for DOM to update or short timeout
      await this.page.waitForLoadState('domcontentloaded', { timeout: 500 }).catch(() => {});
      await this.page.waitForTimeout(200); // Small delay between polls
    }
    throw new Error(`Project "${projectName}" was not removed within ${timeout}ms`);
  }

  // ==================== USER MANAGEMENT ====================

  /**
   * Get all user rows from the users table
   * Excludes empty state rows that show "No users found"
   */
  getUserRows(): Locator {
    return this.page.locator('[data-testid="users-table"] tbody tr, table tbody tr')
      .filter({ hasText: /.+/ })
      .filter({ hasNotText: /No users found/ });
  }

  /**
   * Get user by username
   */
  getUserByUsername(username: string): Locator {
    return this.page.locator(`tr:has-text("${username}")`).first();
  }

  /**
   * Create a new user
   * Note: Caller must ensure they are on the Users page first
   */
  async createUser(username: string, email: string, fullName: string) {
    // Click Create User button using test ID
    await this.page.getByTestId('create-user-button').click();

    // Wait for the specific Create User modal to be visible
    // Cloudscape renders all modals in DOM but hides them with CSS
    // We need to target the visible one by its header text
    const dialog = this.page.getByRole('dialog', { name: /create new user/i });
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Use test IDs to avoid strict mode violations with multiple email fields
    // Cloudscape wraps inputs in divs, so we need to find the input inside
    await this.page.getByTestId('user-username-input').locator('input').fill(username);
    await this.page.getByTestId('user-email-input').locator('input').fill(email);
    await this.page.getByTestId('user-fullname-input').locator('input').fill(fullName);

    await this.clickButton('create');
    await this.waitForDialogClose();

    // Wait for the newly created user to appear in the list
    // This ensures state has fully updated before returning
    await this.page.waitForSelector(`tr:has-text("${username}")`, { timeout: 5000 });
  }

  /**
   * Delete a user
   * Note: Assumes we're already on the Users page
   */
  async deleteUser(username: string) {
    const userRow = this.getUserByUsername(username);
    const actionsButton = userRow.getByRole('button', { name: /actions/i });
    await actionsButton.click();

    // Wait for menu to appear
    const deleteOption = this.page.getByRole('menuitem', { name: /delete/i });
    await deleteOption.waitFor({ state: 'visible', timeout: 3000 });
    await deleteOption.click();

    // Wait for confirmation dialog - Cloudscape renders all modals in DOM but hides them with CSS
    // Target the specific delete user dialog by its header text
    const dialog = this.page.getByRole('dialog', { name: /delete user\?/i });
    await dialog.waitFor({ state: 'visible', timeout: 3000 });
  }

  /**
   * Verify user exists
   */
  async verifyUserExists(username: string): Promise<boolean> {
    await this.navigateToUsers();
    const user = this.getUserByUsername(username);
    return await this.elementExists(user);
  }

  /**
   * Get user count
   */
  async getUserCount(): Promise<number> {
    // Don't navigate - just count the rows currently displayed
    // This avoids triggering another data refresh which could return stale data
    return await this.getUserRows().count();
  }

  /**
   * Wait for user count to reach expected value
   * Polls until count stabilizes at the expected value
   */
  async waitForUserCount(expectedCount: number, timeout: number = 10000): Promise<void> {
    const startTime = Date.now();
    let lastCount = -1;
    let stableCount = 0;

    while (Date.now() - startTime < timeout) {
      const currentCount = await this.getUserRows().count();

      if (currentCount === expectedCount) {
        // Count matches - verify it stays stable for 2 checks
        stableCount++;
        if (stableCount >= 2) {
          return; // Success - count is stable at expected value
        }
      } else {
        stableCount = 0; // Reset stability counter
      }

      lastCount = currentCount;
      // Wait for DOM to update or short timeout
      await this.page.waitForLoadState('domcontentloaded', { timeout: 200 }).catch(() => {});
    }

    throw new Error(`User count did not reach ${expectedCount} within ${timeout}ms. Last count: ${lastCount}`);
  }

  /**
   * Wait for user to be removed from list
   */
  async waitForUserToBeRemoved(username: string, timeout: number = 15000) {
    await this.waitForDialogClose();

    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      try {
        const exists = await this.verifyUserExists(username);
        if (!exists) {
          return; // Success
        }
      } catch (error) {
        // Continue polling
      }
      // Wait for DOM to update or short timeout
      await this.page.waitForLoadState('domcontentloaded', { timeout: 500 }).catch(() => {});
      await this.page.waitForTimeout(200); // Small delay between polls
    }
    throw new Error(`User "${username}" was not removed within ${timeout}ms`);
  }

  // ==================== INVITATIONS MANAGEMENT ====================

  /**
   * Switch to Individual Invitations tab
   */
  async switchToIndividualInvitations() {
    await this.navigateToInvitations();

    // Click the "Individual" tab
    const individualTab = this.page.getByRole('tab', { name: /individual/i });
    if (await this.elementExists(individualTab)) {
      await individualTab.click();
      // Wait for tab panel to be visible
      await this.page.getByRole('tabpanel', { name: /individual/i }).waitFor({ state: 'visible', timeout: 3000 });
    }
  }

  /**
   * Switch to Bulk Invitations tab
   */
  async switchToBulkInvitations() {
    await this.navigateToInvitations();

    const bulkTab = this.page.getByRole('tab', { name: /bulk/i });
    if (await this.elementExists(bulkTab)) {
      await bulkTab.click();
      // Wait for tab panel to be visible
      await this.page.getByRole('tabpanel', { name: /bulk/i }).waitFor({ state: 'visible', timeout: 3000 });
    }
  }

  /**
   * Switch to Shared Tokens tab
   */
  async switchToSharedTokens() {
    await this.navigateToInvitations();

    const sharedTab = this.page.getByRole('tab', { name: /shared/i });
    if (await this.elementExists(sharedTab)) {
      await sharedTab.click();
      // Wait for tab panel to be visible
      await this.page.getByRole('tabpanel', { name: /shared/i }).waitFor({ state: 'visible', timeout: 3000 });
    }
  }

  /**
   * Get all invitation rows
   */
  getInvitationRows(): Locator {
    return this.page.locator('[data-testid="invitations-table"] tbody tr, table tbody tr').filter({ hasText: /.+/ });
  }

  /**
   * Get specific cell text from an invitation row
   * Column mapping: project (0), role (1), invited_by (2), expires (3), status (4), actions (5)
   */
  async getInvitationCellText(row: Locator, columnIndex: number): Promise<string> {
    const cell = row.locator('td').nth(columnIndex);
    const text = await cell.textContent();
    return text?.trim() || '';
  }

  /**
   * Get invitation data as an object from a row
   */
  async getInvitationData(row: Locator): Promise<{
    project: string;
    role: string;
    invitedBy: string;
    expires: string;
    status: string;
  }> {
    return {
      project: await this.getInvitationCellText(row, 0),
      role: await this.getInvitationCellText(row, 1),
      invitedBy: await this.getInvitationCellText(row, 2),
      expires: await this.getInvitationCellText(row, 3),
      status: await this.getInvitationCellText(row, 4),
    };
  }

  /**
   * Add invitation by token and optionally wait for it to appear
   * @param token - The invitation token to add
   * @param expectedProjectName - Optional project name to wait for. If provided, waits for the invitation row to appear.
   */
  async addInvitationToken(token: string, expectedProjectName?: string) {
    if (!token) {
      throw new Error('Token parameter is required for addInvitationToken()');
    }

    await this.switchToIndividualInvitations();

    // Scope to the Individual tab panel to avoid modal conflicts
    const individualPanel = this.page.getByRole('tabpanel', { name: 'Individual' });

    // Find the actual input element inside the Cloudscape Input component
    // data-testid is on the wrapper div, need to find the <input> inside
    const tokenInput = individualPanel.getByTestId('invitation-token-input').locator('input');
    await tokenInput.fill(token);

    // Use data-testid for the button scoped to the tab panel
    const addButton = individualPanel.getByTestId('add-invitation-button');

    // Click and wait for the invitation row to appear in the table
    await addButton.click();

    // If expectedProjectName is provided, wait for the specific invitation row to appear
    // This proves the API call completed AND React state updated AND table re-rendered
    if (expectedProjectName) {
      const invitationRow = this.page.locator(`tr:has-text("${expectedProjectName}")`).first();
      await invitationRow.waitFor({ state: 'visible', timeout: 10000 });
    }
  }

  /**
   * Accept invitation
   */
  async acceptInvitation(projectName: string) {
    await this.switchToIndividualInvitations();

    // Wait for the invitation row to be visible
    // This handles cases where the invitation was just added and may not be immediately visible
    const invitationRow = this.page.locator(`tr:has-text("${projectName}")`).first();
    await invitationRow.waitFor({ state: 'visible', timeout: 10000 });

    const acceptButton = invitationRow.getByRole('button', { name: /accept/i });
    await acceptButton.click();

    // Wait for confirmation modal - Cloudscape-specific targeting
    const dialog = this.page.getByRole('dialog', { name: /accept invitation/i });
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Confirm in modal
    await this.clickButton('accept');
    await this.waitForDialogClose();
  }

  /**
   * Decline invitation
   */
  async declineInvitation(projectName: string, reason?: string) {
    await this.switchToIndividualInvitations();

    const invitationRow = this.page.locator(`tr:has-text("${projectName}")`).first();
    const declineButton = invitationRow.getByRole('button', { name: /decline/i });
    await declineButton.click();

    // Wait for confirmation modal - Cloudscape-specific targeting
    const dialog = this.page.getByRole('dialog', { name: /decline invitation/i });
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    if (reason) {
      // Cloudscape wraps textarea in a span, need to target the actual textarea
      const reasonTextarea = this.page.getByTestId('decline-reason-textarea').locator('textarea');
      await reasonTextarea.fill(reason);
    }

    await this.clickButton('decline');
    await this.waitForDialogClose();
  }

  /**
   * Send bulk invitations
   */
  async sendBulkInvitations(
    projectId: string,
    emails: string[],
    role: 'viewer' | 'member' | 'admin',
    message?: string
  ) {
    await this.switchToBulkInvitations();

    // Fill email addresses - Cloudscape wraps textarea in a span, need to target the actual textarea
    const emailTextarea = this.page.getByTestId('bulk-emails-textarea').locator('textarea');
    await emailTextarea.fill(emails.join('\n'));

    // Select role - use specific test ID to avoid strict mode violation (multiple role selectors on page)
    const roleSelect = this.page.getByTestId('bulk-role-select');
    await roleSelect.click();
    await this.page.waitForTimeout(300);
    await this.page.locator(`[data-value="${role}"]`).click();

    // Select project using the test ID
    const projectSelect = this.page.getByTestId('bulk-invite-project-select');
    await projectSelect.click();

    // Wait for dropdown options to appear
    const projectOption = this.page.locator(`[data-value="${projectId}"]`);
    await projectOption.waitFor({ state: 'visible', timeout: 3000 });
    await projectOption.click();

    // Optional message - Cloudscape wraps textarea in a span, need to target the actual textarea
    if (message) {
      const messageTextarea = this.page.getByTestId('bulk-message-textarea').locator('textarea');
      await messageTextarea.fill(message);
    }

    // Submit
    await this.clickButton('send bulk invitations');
    await this.waitForLoadingComplete();
  }

  /**
   * Create shared token
   */
  async createSharedToken(
    name: string,
    redemptionLimit: number,
    expiresIn: '1d' | '7d' | '30d' | '90d',
    role: 'viewer' | 'member' | 'admin',
    message?: string
  ) {
    await this.switchToSharedTokens();

    await this.page.getByRole('button', { name: /create shared token/i }).click();

    // Wait for dialog to appear using data-testid
    const dialog = this.page.locator('[data-testid="create-shared-token-modal"]');
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    await this.fillInput('token name', name);
    await this.fillInput('redemption limit', redemptionLimit.toString());

    // Use data-testid for Cloudscape Select components
    await this.page.locator('[data-testid="expires-in-select"]').click();
    await this.page.locator(`[data-value="${expiresIn}"]`).click();

    // Use specific test ID to avoid strict mode violation with other role selectors
    await this.page.locator('[data-testid="shared-token-role-select"]').click();
    await this.page.locator(`[data-value="${role}"]`).click();

    if (message) {
      await this.fillInput('welcome message', message);
    }

    await this.page.locator('[data-testid="create-token-button"]').click();
    await this.waitForDialogClose();
  }

  /**
   * Revoke shared token
   */
  async revokeSharedToken(tokenName: string) {
    await this.switchToSharedTokens();

    const tokenRow = this.page.locator(`tr:has-text("${tokenName}")`).first();

    // Click revoke icon button (close icon)
    const revokeButton = tokenRow.getByRole('button', { name: /revoke|close/i });
    await revokeButton.click();

    // Wait for confirmation modal with name-based selector
    const dialog = this.page.getByRole('dialog', { name: /revoke|confirm/i });
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Confirm
    await this.clickButton('revoke');
    await this.waitForDialogClose();
  }

  /**
   * View QR code for shared token
   */
  async viewQRCode(tokenName: string) {
    await this.switchToSharedTokens();

    const tokenRow = this.page.locator(`tr:has-text("${tokenName}")`).first();
    const viewButton = tokenRow.getByRole('button', { name: /view/i });
    await viewButton.click();

    // Wait for QR code dialog to appear
    const dialog = this.page.getByRole('dialog', { name: /qr code/i });
    await dialog.waitFor({ state: 'visible', timeout: 5000 });
  }

  /**
   * Get specific cell text from a shared token row
   * Column mapping: name (0), project (1), role (2), redemptions (3), expires (4), status (5), actions (6)
   */
  async getSharedTokenCellText(row: Locator, columnIndex: number): Promise<string> {
    const cell = row.locator('td').nth(columnIndex);
    const text = await cell.textContent();
    return text?.trim() || '';
  }

  /**
   * Get shared token data as an object from a row
   */
  async getSharedTokenData(row: Locator): Promise<{
    name: string;
    project: string;
    role: string;
    redemptions: string;
    expires: string;
    status: string;
  }> {
    return {
      name: await this.getSharedTokenCellText(row, 0),
      project: await this.getSharedTokenCellText(row, 1),
      role: await this.getSharedTokenCellText(row, 2),
      redemptions: await this.getSharedTokenCellText(row, 3),
      expires: await this.getSharedTokenCellText(row, 4),
      status: await this.getSharedTokenCellText(row, 5),
    };
  }

  /**
   * Get invitation count
   */
  async getInvitationCount(): Promise<number> {
    await this.switchToIndividualInvitations();
    return await this.getInvitationRows().count();
  }

  /**
   * Verify invitation exists
   */
  async verifyInvitationExists(projectName: string): Promise<boolean> {
    await this.switchToIndividualInvitations();
    const invitation = this.page.locator(`tr:has-text("${projectName}")`).first();
    return await this.elementExists(invitation);
  }

  // ==================== COMMON UTILITIES ====================

  /**
   * Wait for dialog to close
   * Cloudscape-specific: Check for visible modals rather than waiting for all dialogs to be hidden
   * since Cloudscape renders all modals in DOM but hides them with CSS
   */
  async waitForDialogClose(timeout: number = 5000) {
    try {
      // Check if any dialog is currently visible (not just in DOM)
      const visibleDialogs = this.page.locator('[role="dialog"]:visible');
      const count = await visibleDialogs.count();

      if (count > 0) {
        // Wait for visible dialogs to become hidden
        await visibleDialogs.first().waitFor({ state: 'hidden', timeout });
        // Wait for animation to complete
        await this.page.waitForTimeout(300);
      }
    } catch {
      // Dialog already closed or timeout - safe to continue
    }
  }

  /**
   * Force close any open dialogs
   * Cloudscape-specific: Only target visible dialogs, not all dialogs in DOM
   */
  async forceCloseDialogs() {
    try {
      // Check for visible dialogs (not just dialogs in DOM)
      const visibleDialogs = this.page.locator('[role="dialog"]:visible');
      const count = await visibleDialogs.count();

      if (count === 0) {
        return; // No visible dialogs to close
      }

      // Try to find and click close button in the visible dialog
      const closeSelectors = [
        'button[aria-label="Close dialog"]',
        'button[aria-label="Close modal"]',
        '[role="dialog"]:visible button:has-text("Cancel")',
        '[role="dialog"]:visible button:has-text("Close")'
      ];

      for (const selector of closeSelectors) {
        const button = this.page.locator(selector).first();
        if (await button.isVisible({ timeout: 500 })) {
          await button.click();
          // Wait for animation
          await this.page.waitForTimeout(500);
          break;
        }
      }

      await this.waitForDialogClose(2000);
    } catch {
      // No dialogs to close or already closed
    }
  }

  /**
   * Clean up test projects by pattern
   */
  async cleanupTestProjects(namePattern: RegExp) {
    await this.navigate();

    const rows = this.getProjectRows();
    const count = await rows.count();

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i);
      const text = await row.textContent();

      if (text && namePattern.test(text)) {
        // Extract project name
        const match = text.match(/^([a-zA-Z0-9\-_]+)/);
        if (match) {
          const projectName = match[1];
          try {
            await this.deleteProject(projectName);
            await this.clickButton('delete');
            await this.waitForDialogClose();
          } catch {
            // Skip if deletion fails
            try {
              await this.clickButton('cancel');
            } catch {
              // Ignore
            }
          }
        }
      }
    }
  }

  /**
   * Clean up test users by pattern
   * Note: Assumes we're already on the Users page
   * Repeatedly finds and deletes the first matching user until none remain
   */
  async cleanupTestUsers(namePattern: RegExp) {
    let attempts = 0;
    const maxAttempts = 50; // Safety limit to prevent infinite loops

    while (attempts < maxAttempts) {
      const rows = this.getUserRows();
      const count = await rows.count();
      let found = false;

      // Find the first matching user
      for (let i = 0; i < count; i++) {
        const row = rows.nth(i);
        const text = await row.textContent();

        if (text && namePattern.test(text)) {
          const match = text.match(/^([a-zA-Z0-9\-_]+)/);
          if (match) {
            const username = match[1];
            try {
              await this.deleteUser(username);
              await this.clickButton('delete');
              await this.waitForDialogClose();
              found = true;
              break; // Exit loop to refresh row list after deletion
            } catch {
              try {
                await this.clickButton('cancel');
              } catch {
                // Ignore
              }
            }
          }
        }
      }

      if (!found) {
        // No more matching users found
        break;
      }
      attempts++;
    }
  }

  // ==================== INVITATION TEST HELPERS ====================

  /**
   * Create test project via API for invitation testing
   */
  async createTestProject(name: string): Promise<string> {
    const projectId = await this.page.evaluate(async (projectName) => {
      const api = (window as any).__apiClient;
      const project = await api.createProject({
        name: projectName,
        description: 'Test project for invitation workflows',
        owner: 'test-owner'
      });
      return project.id;
    }, name);

    return projectId;
  }

  /**
   * Send invitation to test user and return invitation token
   */
  async sendTestInvitation(
    projectId: string,
    email: string = 'test-user@example.com',
    role: 'viewer' | 'member' | 'admin'
  ): Promise<string> {
    const token = await this.page.evaluate(async (args) => {
      const api = (window as any).__apiClient;
      const invitation = await api.sendInvitation(
        args.projectId,
        args.email,
        args.role,
        'Test invitation message'
      );
      return invitation.token;
    }, { projectId, email, role });

    return token;
  }

  /**
   * Delete test project via API
   */
  async deleteTestProject(projectId: string): Promise<void> {
    await this.page.evaluate(async (id) => {
      const api = (window as any).__apiClient;
      try {
        await api.deleteProject(id);
      } catch (err) {
        console.error('Failed to delete test project:', err);
      }
    }, projectId);
  }

  /**
   * Verify user appears in project members list
   */
  async verifyProjectMember(projectName: string, username: string): Promise<boolean> {
    await this.viewProjectDetails(projectName);

    // Wait for members section to load
    const membersSection = this.page.locator('[data-testid="project-members"]');
    await membersSection.waitFor({ state: 'visible', timeout: 5000 });

    const membersText = await membersSection.textContent();
    return membersText?.includes(username) || false;
  }

  // ==================== CLEANUP UTILITIES ====================

  /**
   * Clean up all test projects (names starting with common test prefixes)
   * Similar to cleanupTestInvitations for invitations
   */
  async cleanupTestProjects() {
    await this.navigate();

    const testPrefixes = [
      'test-project',
      'budget-test',
      'delete-test',
      'duplicate-test',
      'view-test',
      'active-project',
      'suspended-project',
      'budget-view-test',
      'cancel-delete-test',
      'list-test',
      'alert-test',
      'exceeded-test',
      'spend-test'
    ];

    // Get all project rows
    const rows = this.getProjectRows();
    const count = await rows.count();

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i);
      const rowText = await row.textContent();

      // Check if row matches any test prefix
      const isTestProject = testPrefixes.some(prefix =>
        rowText?.toLowerCase().includes(prefix.toLowerCase())
      );

      if (isTestProject && rowText) {
        // Extract project name from the row (first cell/link usually)
        const nameLink = row.locator('a').first();
        const projectName = await nameLink.textContent();

        if (projectName) {
          try {
            await this.deleteProject(projectName.trim());
            await this.confirmDeletion();
            await this.waitForProjectToBeRemoved(projectName.trim());
          } catch (error) {
            // Project might already be deleted or not found, continue
            console.log(`Failed to delete project ${projectName}:`, error);
          }
        }
      }
    }
  }
}
