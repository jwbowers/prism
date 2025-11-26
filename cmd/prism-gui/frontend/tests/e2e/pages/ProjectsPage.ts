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
    await this.page.waitForTimeout(500);

    // Fill form fields
    await this.fillInput('project name', name);
    await this.fillInput('description', description);

    if (budget !== undefined) {
      await this.fillInput('budget', budget.toString());
    }

    // Submit
    await this.clickButton('create');
    await this.waitForDialogClose();
  }

  /**
   * Delete a project
   */
  async deleteProject(projectName: string) {
    await this.navigate();

    const projectRow = this.getProjectByName(projectName);

    // Click actions dropdown
    const actionsButton = projectRow.getByRole('button', { name: /actions/i });
    await actionsButton.click();
    await this.page.waitForTimeout(300);

    // Click delete option
    await this.page.getByRole('menuitem', { name: /delete/i }).click();
    await this.page.waitForTimeout(300);
  }

  /**
   * View project details
   */
  async viewProjectDetails(projectName: string) {
    await this.navigate();

    const projectRow = this.getProjectByName(projectName);
    const actionsButton = projectRow.getByRole('button', { name: /actions/i });
    await actionsButton.click();
    await this.page.waitForTimeout(300);

    await this.page.getByRole('menuitem', { name: /view details/i }).click();
    await this.waitForLoadingComplete();
  }

  /**
   * Verify project exists
   */
  async verifyProjectExists(name: string): Promise<boolean> {
    await this.navigate();
    const project = this.getProjectByName(name);
    return await this.elementExists(project);
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
  async waitForProjectToBeRemoved(projectName: string, timeout: number = 15000) {
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
      await this.page.waitForTimeout(500);
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
    await this.page.waitForTimeout(500);

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
    await this.page.waitForTimeout(300);

    await this.page.getByRole('menuitem', { name: /delete/i }).click();
    await this.page.waitForTimeout(300);
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
      await this.page.waitForTimeout(200);
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
      await this.page.waitForTimeout(500);
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
      await this.page.waitForTimeout(300);
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
      await this.page.waitForTimeout(300);
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
      await this.page.waitForTimeout(300);
    }
  }

  /**
   * Get all invitation rows
   */
  getInvitationRows(): Locator {
    return this.page.locator('[data-testid="invitations-table"] tbody tr, table tbody tr').filter({ hasText: /.+/ });
  }

  /**
   * Add invitation by token
   */
  async addInvitationToken(token: string) {
    await this.switchToIndividualInvitations();

    await this.fillInput('invitation token', token);
    await this.clickButton('add invitation');
    await this.page.waitForTimeout(500);
  }

  /**
   * Accept invitation
   */
  async acceptInvitation(projectName: string) {
    await this.switchToIndividualInvitations();

    const invitationRow = this.page.locator(`tr:has-text("${projectName}")`).first();
    const acceptButton = invitationRow.getByRole('button', { name: /accept/i });
    await acceptButton.click();
    await this.page.waitForTimeout(500);

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
    await this.page.waitForTimeout(500);

    if (reason) {
      await this.fillInput('reason', reason);
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

    // Fill email addresses
    const emailTextarea = this.page.getByLabel(/email.*addresses/i);
    await emailTextarea.fill(emails.join('\n'));

    // Select role
    const roleSelect = this.page.getByLabel(/role/i);
    await roleSelect.selectOption(role);

    // Select project using the test ID
    const projectSelect = this.page.getByTestId('bulk-invite-project-select');
    await projectSelect.click();
    await this.page.waitForTimeout(300);

    // Select the option with the matching project ID
    const projectOption = this.page.locator(`[data-value="${projectId}"]`);
    if (await this.elementExists(projectOption)) {
      await projectOption.click();
    }

    // Optional message
    if (message) {
      const messageTextarea = this.page.getByLabel(/message/i);
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
    await this.page.waitForTimeout(500);

    await this.fillInput('token name', name);
    await this.fillInput('redemption limit', redemptionLimit.toString());
    await this.selectOption('expires in', expiresIn);
    await this.selectOption('role', role);

    if (message) {
      await this.fillInput('welcome message', message);
    }

    await this.clickButton('create token');
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
    await this.page.waitForTimeout(500);

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
    await this.page.waitForTimeout(500);
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
   */
  async waitForDialogClose(timeout: number = 5000) {
    try {
      const dialog = this.page.locator('[role="dialog"]').first();
      const isVisible = await dialog.isVisible({ timeout: 500 });

      if (isVisible) {
        await this.page.waitForSelector('[role="dialog"]', {
          state: 'hidden',
          timeout
        });
        await this.page.waitForTimeout(500);
      }
    } catch {
      // Dialog already closed
    }
  }

  /**
   * Force close any open dialogs
   */
  async forceCloseDialogs() {
    try {
      const closeSelectors = [
        'button[aria-label="Close dialog"]',
        'button[aria-label="Close modal"]',
        '[role="dialog"] button:has-text("Cancel")',
        '[role="dialog"] button:has-text("Close")'
      ];

      for (const selector of closeSelectors) {
        const button = this.page.locator(selector).first();
        if (await button.isVisible({ timeout: 500 })) {
          await button.click();
          await this.page.waitForTimeout(500);
          break;
        }
      }

      await this.waitForDialogClose(2000);
    } catch {
      // No dialogs to close
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
            await this.page.waitForTimeout(500);
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
              await this.page.waitForTimeout(500);
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
        owner: 'test-owner',
        budget_limit: 1000,
        budget_period: 'monthly',
        status: 'active'
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
    email: string,
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
    await this.page.waitForTimeout(1000);

    const membersSection = this.page.locator('[data-testid="project-members"]');
    if (await membersSection.isVisible({ timeout: 5000 })) {
      const membersText = await membersSection.textContent();
      return membersText?.includes(username) || false;
    }

    return false;
  }
}
