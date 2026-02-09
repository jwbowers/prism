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
   * Clean up old test projects (projects with test-related names)
   * This helps prevent test pollution and keeps pagination manageable
   */
  async cleanupOldTestProjects() {
    await this.navigate();

    // Set filter to "All Projects" to see everything
    const filterSelect = this.page.getByTestId('project-filter-select');
    if (await filterSelect.isVisible({ timeout: 2000 }).catch(() => false)) {
      await filterSelect.click({ timeout: 2000 }).catch(() => {});
      await this.page.locator('[data-value="all"]').click({ timeout: 2000 }).catch(() => {});
      await this.page.waitForTimeout(1000);
    }

    // Look for test project patterns and delete them
    const testPatterns = [
      /^list-test-\d+-\d+$/,
      /^active-project-\d+$/,
      /^suspended-project-\d+$/,
      /^cancel-delete-test-\d+$/,
      /^delete-test-\d+$/
    ];

    // Get all project name cells
    const cells = this.page.getByRole('cell').filter({ hasText: /-test-|-project-/ });
    const count = await cells.count();

    // Iterate through pages to find and delete test projects
    let deletedCount = 0;
    for (let i = 0; i < count && deletedCount < 50; i++) {
      try {
        const cellText = await cells.nth(i).textContent({ timeout: 1000 });
        if (!cellText) continue;

        // Check if this matches any test pattern
        const isTestProject = testPatterns.some(pattern => pattern.test(cellText));
        if (isTestProject) {
          try {
            await this.deleteProject(cellText);
            await this.page.getByTestId('confirm-delete-button').click({ timeout: 2000 });
            await this.waitForProjectToBeRemoved(cellText);
            deletedCount++;
            // Navigate back to page 1 after each deletion
            await this.navigate();
          } catch (e) {
            // Skip if deletion fails (project might not exist anymore)
            continue;
          }
        }
      } catch (e) {
        // Skip if we can't read the cell
        continue;
      }
    }
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
   * NOTE: Most tests should use deleteProjectViaAPI() for cleanup - it's faster and more reliable
   * This method is primarily for tests that specifically test the deletion UI workflow
   * @param projectName Name of the project to delete
   */
  async deleteProject(projectName: string) {
    await this.navigate();

    // Wait for project to be visible on page (assumes it's on current page, typically page 1 for new projects)
    await this.page.getByRole('cell', { name: projectName, exact: true }).waitFor({ state: 'visible', timeout: 10000 });

    const projectRow = this.getProjectByName(projectName);

    // Click actions dropdown with retry logic for element detachment
    const actionsButton = projectRow.getByRole('button', { name: /actions/i });
    await actionsButton.waitFor({ state: 'visible', timeout: 5000 });

    // Retry click if element gets detached (common during table updates)
    let clickSuccess = false;
    for (let attempt = 0; attempt < 3 && !clickSuccess; attempt++) {
      try {
        await actionsButton.click({ timeout: 5000 });
        clickSuccess = true;
      } catch (e) {
        if (attempt === 2) throw e; // Throw on final attempt
        await this.page.waitForTimeout(1000); // Wait for table to stabilize
      }
    }

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

    // Poll for project to appear in table (handles async UI updates)
    const projectExists = await this.verifyProjectExists(projectName);
    if (!projectExists) {
      throw new Error(`Project "${projectName}" not found in projects table`);
    }

    const projectRow = this.getProjectByName(projectName);
    const actionsButton = projectRow.getByRole('button', { name: /actions/i });
    await actionsButton.waitFor({ state: 'visible', timeout: 5000 });
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
    // Don't navigate - just check if the project exists in the current visible page
    // Use same selector as createProject() for consistency (cell by exact name)
    // With pagination, only checks current page - project must be visible
    const maxWait = 10000; // 10 seconds
    const startTime = Date.now();

    while (Date.now() - startTime < maxWait) {
      // Use cell selector (same as createProject) instead of row has-text
      const cell = this.page.getByRole('cell', { name, exact: true });
      if (await this.elementExists(cell)) {
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

  /**
   * Delete project via API (for test cleanup - much faster and more reliable than UI)
   */
  async deleteProjectViaAPI(projectName: string) {
    try {
      // Get all projects
      const response = await this.page.request.get('http://localhost:8947/api/v1/projects');
      if (!response.ok()) {
        console.warn(`Failed to get projects: ${response.status()}`);
        return;
      }

      const data = await response.json();
      const projects = data.projects || [];

      // Find project by name
      const project = projects.find((p: any) => p.name === projectName);
      if (!project) {
        console.warn(`Project "${projectName}" not found via API`);
        return;
      }

      // Delete project
      const deleteResponse = await this.page.request.delete(`http://localhost:8947/api/v1/projects/${project.id}`);
      if (!deleteResponse.ok()) {
        console.warn(`Failed to delete project "${projectName}": ${deleteResponse.status()}`);
        return;
      }

      // Wait for UI to update
      await this.page.waitForTimeout(500);
    } catch (error) {
      console.warn(`Error deleting project "${projectName}" via API:`, error);
    }
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
    message?: string,
    projectId?: string
  ) {
    await this.switchToSharedTokens();

    await this.page.getByRole('button', { name: /create shared token/i }).click();

    // Wait for dialog to appear using data-testid
    const dialog = this.page.locator('[data-testid="create-shared-token-modal"]');
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Select project if provided, otherwise select first available project
    const projectSelect = dialog.locator('[data-testid="token-project-select"]');
    await projectSelect.click();
    await this.page.waitForTimeout(300);
    // Click first option in dropdown
    await this.page.getByRole('option').first().click();

    await this.fillInput('token name', name);
    await this.fillInput('redemption limit', redemptionLimit.toString());

    // Select expires in
    const expiresSelect = dialog.locator('[data-testid="expires-in-select"]');
    await expiresSelect.click();
    await this.page.waitForTimeout(300);
    await this.page.getByRole('option', { name: expiresIn }).click();

    // Select role
    const roleSelect = dialog.locator('[data-testid="shared-token-role-select"]');
    await roleSelect.click();
    await this.page.waitForTimeout(300);
    await this.page.getByRole('option', { name: role }).click();

    if (message) {
      // Cloudscape Textarea wraps the actual textarea in a span with data-testid
      // So we need to find the textarea element within it
      const messageTextarea = dialog.locator('[data-testid="token-message-textarea"]').locator('textarea');
      await messageTextarea.fill(message);
    }

    await this.page.locator('[data-testid="create-token-button"]').click();
    await this.waitForDialogClose();
  }

  /**
   * Revoke shared token
   */
  async revokeSharedToken(tokenName: string) {
    await this.switchToSharedTokens();

    // Wait for token row to appear
    const tokenRow = this.page.locator(`tr:has-text("${tokenName}")`).first();
    await tokenRow.waitFor({ state: 'visible', timeout: 10000 });

    // Click revoke button using text match
    const revokeButton = tokenRow.getByRole('button', { name: 'Revoke' });
    await revokeButton.click();

    // Wait for confirmation modal using data-testid
    const dialog = this.page.locator('[data-testid="revoke-token-modal"]');
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Confirm by clicking the confirm button
    const confirmButton = dialog.locator('[data-testid="confirm-revoke-button"]');
    await confirmButton.click();

    await this.waitForDialogClose();
  }

  /**
   * Extend shared token expiration
   */
  async extendSharedToken(tokenName: string, days: '7' | '30' | '90' = '7') {
    await this.switchToSharedTokens();

    // Wait for token row to appear
    const tokenRow = this.page.locator(`tr:has-text("${tokenName}")`).first();
    await tokenRow.waitFor({ state: 'visible', timeout: 10000 });

    // Click extend button using text match
    const extendButton = tokenRow.getByRole('button', { name: 'Extend' });
    await extendButton.click();

    // Wait for extend modal using data-testid
    const dialog = this.page.locator('[data-testid="extend-token-modal"]');
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Select duration using Cloudscape Select component
    const durationSelect = dialog.locator('[data-testid="extend-duration-select"]');
    await durationSelect.click();
    await this.page.waitForTimeout(300);
    await this.page.getByRole('option', { name: `${days} days` }).click();

    // Confirm extension
    const confirmButton = dialog.locator('[data-testid="confirm-extend-button"]');
    await confirmButton.click();

    await this.waitForDialogClose();
  }

  /**
   * View QR code for shared token
   */
  async viewQRCode(tokenName: string) {
    await this.switchToSharedTokens();

    // Wait for token row to appear in table
    const tokenRow = this.page.locator(`tr:has-text("${tokenName}")`).first();
    await tokenRow.waitFor({ state: 'visible', timeout: 10000 });

    const viewButton = tokenRow.getByRole('button', { name: /view/i });
    await viewButton.click();

    // Wait for QR code dialog to appear - header is "Shared Invitation Token"
    const dialog = this.page.locator('[data-testid="qr-code-modal"]');
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
   * Filter invitations by status (Pending, Accepted, Declined, etc.)
   */
  async filterInvitationsByStatus(status: 'All' | 'Pending' | 'Accepted' | 'Declined' | 'Expired' | 'Revoked'): Promise<void> {
    // Click the filter button
    const filterButton = this.page.getByRole('button', { name: /filter by status/i });
    await filterButton.click();

    // Wait for dropdown to appear and click the status option
    await this.page.waitForTimeout(300);
    const statusOption = this.page.getByRole('option', { name: new RegExp(status, 'i') });
    await statusOption.click();

    // Wait for table to update
    await this.waitForNetworkIdle();
  }

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
    const result = await this.page.evaluate(async (args) => {
      const api = (window as any).__apiClient;
      if (!api) {
        throw new Error('API client not found - page may not be loaded');
      }
      try {
        console.log(`Creating invitation for project ${args.projectId}, email ${args.email}, role ${args.role}`);
        const invitation = await api.sendInvitation(
          args.projectId,
          args.email,
          args.role,
          'Test invitation message'
        );
        console.log(`Invitation created successfully: token=${invitation.token}, status=${invitation.status}`);
        return { token: invitation.token, status: invitation.status, error: null };
      } catch (err: any) {
        console.error('Failed to create invitation:', err);
        return { token: null, status: null, error: err.message || String(err) };
      }
    }, { projectId, email, role });

    if (result.error) {
      throw new Error(`Failed to send test invitation: ${result.error}`);
    }

    if (!result.token) {
      throw new Error('Invitation created but no token returned');
    }

    console.log(`Test invitation created: token=${result.token}, status=${result.status}`);
    return result.token;
  }

  /**
   * Get project members via API
   */
  async getProjectMembers(projectId: string): Promise<any[]> {
    return await this.page.evaluate(async (id) => {
      const api = (window as any).__apiClient;
      if (!api) {
        throw new Error('API client not found - page may not be loaded');
      }
      try {
        const project = await api.getProject(id);
        return project.members || [];
      } catch (err: any) {
        console.error(`Failed to get project members: ${err.message || String(err)}`);
        throw err;
      }
    }, projectId);
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
   * Clean up all test invitations via API
   * Fetches all invitations and deletes them to prevent test data pollution
   * Only revokes pending invitations (accepted/declined cannot be revoked)
   */
  async cleanupTestInvitations(): Promise<void> {
    await this.page.evaluate(async () => {
      const api = (window as any).__apiClient;
      try {
        // Get all invitations for the test user
        const invitations = await api.getMyInvitations('test-user@example.com');

        // Only revoke pending invitations (backend rejects revocation of accepted/declined)
        const pendingInvitations = invitations.filter((inv: any) => inv.status === 'pending');

        console.log(`Cleanup: Found ${invitations.length} invitations, ${pendingInvitations.length} pending`);

        // Delete each pending invitation
        for (const invitation of pendingInvitations) {
          try {
            await api.revokeInvitation(invitation.id);
          } catch (err) {
            console.error(`Failed to delete invitation ${invitation.id}:`, err);
          }
        }
      } catch (err) {
        console.error('Failed to cleanup test invitations:', err);
      }
    });
  }

  /**
   * Clean up test projects that match invitation workflow patterns
   * Deletes projects via API to prevent database pollution
   */
  async cleanupInvitationTestProjects(): Promise<void> {
    await this.page.evaluate(async () => {
      const api = (window as any).__apiClient;
      try {
        // Get all projects
        const projects = await api.getProjects();

        // Test project name patterns used in invitation tests
        const testPatterns = [
          /^Test-Project-\d+$/,
          /^Accept Test \d+$/,
          /^Decline Test \d+$/,
          /^Bulk Test \d+$/,
          /^Token Test \d+$/,
          /^Add Token Test \d+$/,
          /^test-project-\d+$/i
        ];

        // Delete matching projects
        for (const project of projects) {
          const matches = testPatterns.some(pattern => pattern.test(project.name));
          if (matches) {
            try {
              await api.deleteProject(project.id);
            } catch (err) {
              console.error(`Failed to delete project ${project.name}:`, err);
            }
          }
        }
      } catch (err) {
        console.error('Failed to cleanup invitation test projects:', err);
      }
    });
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
      'spend-test',
      // Invitation workflow test projects
      'Membership Test',
      'Dialog Test',
      'Stats Test',
      'Expiration Test',
      'Email Validation Test',
      'Welcome Message Test',
      'Decline Test',
      'Decline Dialog Test',
      'Results Summary Test',
      'Decline Reason Test',
      'test-collab-project',
      'test-invitation-project'
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

  // ==================== TIMING UTILITIES (Issue #366) ====================

  /**
   * Wait for API response to complete before proceeding
   * Helps prevent timing issues where UI checks happen before API calls finish
   */
  async waitForAPIResponse(urlPattern: string | RegExp, timeout: number = 10000): Promise<void> {
    await this.page.waitForResponse(
      (response) => {
        const url = response.url();
        if (typeof urlPattern === 'string') {
          return url.includes(urlPattern);
        }
        return urlPattern.test(url);
      },
      { timeout }
    );
    // Add small buffer for React re-render
    await this.page.waitForTimeout(500);
  }

  /**
   * Wait for network to become idle (all requests completed)
   * Useful after operations that trigger multiple API calls
   */
  async waitForNetworkIdle(timeout: number = 5000): Promise<void> {
    await this.page.waitForLoadState('networkidle', { timeout });
    await this.page.waitForTimeout(500);
  }

  /**
   * Poll until a condition is true or timeout
   * Generic helper for waiting for dynamic conditions
   */
  async pollUntil(
    condition: () => Promise<boolean>,
    options: { timeout?: number; interval?: number; errorMessage?: string } = {}
  ): Promise<void> {
    const timeout = options.timeout || 15000;
    const interval = options.interval || 500;
    const errorMessage = options.errorMessage || 'Condition not met within timeout';

    const startTime = Date.now();

    while (Date.now() - startTime < timeout) {
      if (await condition()) {
        return;
      }
      await this.page.waitForTimeout(interval);
    }

    throw new Error(`${errorMessage} (timeout: ${timeout}ms)`);
  }

  /**
   * Wait for table to update with new data
   * Polls until row count changes or specific row appears
   */
  async waitForTableUpdate(
    expectedMinRows: number = 1,
    timeout: number = 10000
  ): Promise<void> {
    await this.pollUntil(
      async () => {
        const rows = this.page.locator('table tbody tr');
        const count = await rows.count();
        return count >= expectedMinRows;
      },
      {
        timeout,
        errorMessage: `Table did not update with at least ${expectedMinRows} rows`
      }
    );
  }

  /**
   * Wait for specific row to appear in table
   * More robust than simple waitFor - retries if row doesn't exist
   */
  async waitForTableRow(
    text: string,
    timeout: number = 10000
  ): Promise<void> {
    await this.pollUntil(
      async () => {
        const row = this.page.locator(`table tbody tr:has-text("${text}")`);
        const count = await row.count();
        return count > 0;
      },
      {
        timeout,
        errorMessage: `Row containing "${text}" did not appear`
      }
    );
  }

  /**
   * Wait for dialog to be ready (visible and fully rendered)
   * More robust than simple waitFor - handles React rendering delays
   */
  async waitForDialogReady(
    dialogNamePattern: string | RegExp,
    timeout: number = 10000
  ): Promise<void> {
    // Wait for dialog to appear in DOM
    const dialog = typeof dialogNamePattern === 'string'
      ? this.page.getByRole('dialog', { name: new RegExp(dialogNamePattern, 'i') })
      : this.page.getByRole('dialog', { name: dialogNamePattern });

    await dialog.waitFor({ state: 'visible', timeout });

    // Wait for dialog to be fully rendered (buttons, inputs, etc.)
    await this.page.waitForTimeout(500);

    // Verify dialog is still visible (not a flicker)
    await dialog.waitFor({ state: 'visible', timeout: 2000 });
  }

  /**
   * Wait for button to become enabled
   * Polls until button is no longer disabled
   */
  async waitForButtonEnabled(
    button: Locator,
    timeout: number = 10000
  ): Promise<void> {
    await this.pollUntil(
      async () => {
        const isDisabled = await button.getAttribute('disabled');
        return isDisabled === null;
      },
      {
        timeout,
        errorMessage: 'Button did not become enabled'
      }
    );
  }
}
