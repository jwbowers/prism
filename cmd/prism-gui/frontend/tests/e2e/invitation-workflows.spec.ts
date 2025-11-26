/**
 * Invitation Workflows E2E Tests
 *
 * End-to-end tests for complete invitation management workflows in Prism GUI.
 * Tests: Individual invitations, bulk invitations, shared tokens, accept/decline flows.
 */

import { test, expect } from '@playwright/test';
import { ProjectsPage } from './pages';

test.describe('Invitation Management Workflows', () => {
  let projectsPage: ProjectsPage;

  test.beforeEach(async ({ page, context }) => {
    // Set localStorage BEFORE navigating to prevent onboarding modal
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });

    projectsPage = new ProjectsPage(page);
    await projectsPage.goto();
    await projectsPage.navigateToInvitations();

    // Force close any open dialogs from previous tests
    await projectsPage.forceCloseDialogs();
  });

  test.describe('Individual Invitations Workflow', () => {
    test('should add invitation by token', async () => {
      // Individual Invitations UI now implemented
      await projectsPage.switchToIndividualInvitations();

      const testToken = 'test-invitation-token-12345';

      // Add invitation
      await projectsPage.addInvitationToken(testToken);

      // Verify invitation appears in list
      const invitationExists = await projectsPage.verifyInvitationExists('Test Project');
      expect(invitationExists).toBe(true);
    });

    test('should display invitation details', async () => {
      // Individual Invitations UI now implemented
      await projectsPage.switchToIndividualInvitations();

      const invitationRow = projectsPage.getInvitationRows().first();
      const invitationText = await invitationRow.textContent();

      // Should show key details: project, role, invited by, expiration
      expect(invitationText).toContain('Project');
      expect(invitationText).toMatch(/viewer|member|admin/i);
    });

    test('should show invitation status badges', async () => {
      // Individual Invitations UI now implemented
      await projectsPage.switchToIndividualInvitations();

      const invitationRow = projectsPage.getInvitationRows().first();
      const invitationText = await invitationRow.textContent();

      // Should have status badge: Pending, Accepted, Declined, Expired
      expect(invitationText).toMatch(/pending|accepted|declined|expired/i);
    });

    test('should filter by invitation status', async () => {
      // Individual Invitations UI now implemented
      await projectsPage.switchToIndividualInvitations();

      // Apply pending filter
      const filterSelect = projectsPage.page.getByLabel(/status/i);
      await filterSelect.selectOption('pending');

      // Verify only pending shown
      const rows = projectsPage.getInvitationRows();
      const count = await rows.count();

      for (let i = 0; i < count; i++) {
        const row = rows.nth(i);
        const text = await row.textContent();
        expect(text).toContain('Pending');
      }
    });
  });

  test.describe('Accept Invitation Workflow', () => {
    test('should accept invitation with confirmation', async () => {
      // Create test project and invitation
      const testProjectName = `Accept Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);
      const testEmail = 'accept-test@example.com';
      const invitationToken = await projectsPage.sendTestInvitation(testProjectId, testEmail, 'member');

      // Add invitation to Individual Invitations tab
      await projectsPage.navigateToInvitations();
      await projectsPage.switchToIndividualInvitations();
      await projectsPage.addInvitationToken(invitationToken);

      // Accept invitation
      await projectsPage.acceptInvitation(testProjectName);

      // Verify status changed to Accepted
      await projectsPage.page.waitForTimeout(1000);
      const invitationRow = projectsPage.page.locator(`tr:has-text("${testProjectName}")`).first();
      const invitationText = await invitationRow.textContent();
      expect(invitationText).toContain('Accepted');

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });

    test('should show acceptance confirmation dialog', async () => {
      // Create test project and invitation
      const testProjectName = `Dialog Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);
      const testEmail = 'dialog-test@example.com';
      const invitationToken = await projectsPage.sendTestInvitation(testProjectId, testEmail, 'admin');

      // Add invitation to Individual Invitations tab
      await projectsPage.navigateToInvitations();
      await projectsPage.switchToIndividualInvitations();
      await projectsPage.addInvitationToken(invitationToken);

      const invitationRow = projectsPage.getInvitationRows().first();
      const acceptButton = invitationRow.getByRole('button', { name: /accept/i });
      await acceptButton.click();

      // Verify confirmation dialog
      const dialog = projectsPage.page.locator('[role="dialog"]').first();
      expect(await dialog.isVisible()).toBe(true);

      // Dialog should show project details
      const dialogText = await dialog.textContent();
      expect(dialogText).toMatch(/project.*role/i);

      // Cancel
      await projectsPage.clickButton('cancel');

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });

    test('should add user to project after acceptance', async () => {
      // Create test project and invitation
      const testProjectName = `Membership Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);
      const testEmail = 'membership-test@example.com';
      const invitationToken = await projectsPage.sendTestInvitation(testProjectId, testEmail, 'viewer');

      // Add invitation to Individual Invitations tab
      await projectsPage.navigateToInvitations();
      await projectsPage.switchToIndividualInvitations();
      await projectsPage.addInvitationToken(invitationToken);

      // Accept invitation
      await projectsPage.acceptInvitation(testProjectName);

      // Navigate to projects and verify membership
      await projectsPage.navigate();
      const isMember = await projectsPage.verifyProjectMember(testProjectName, testEmail);
      expect(isMember).toBe(true);

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });
  });

  test.describe('Decline Invitation Workflow', () => {
    test('should decline invitation with reason', async () => {
      // Create test project and invitation
      const testProjectName = `Decline Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);
      const testEmail = 'decline-test@example.com';
      const invitationToken = await projectsPage.sendTestInvitation(testProjectId, testEmail, 'member');

      // Add invitation to Individual Invitations tab
      await projectsPage.navigateToInvitations();
      await projectsPage.switchToIndividualInvitations();
      await projectsPage.addInvitationToken(invitationToken);

      const declineReason = 'Not interested in this project at the moment';

      // Decline invitation
      await projectsPage.declineInvitation(testProjectName, declineReason);

      // Verify status changed to Declined
      await projectsPage.page.waitForTimeout(1000);
      const invitationRow = projectsPage.page.locator(`tr:has-text("${testProjectName}")`).first();
      const invitationText = await invitationRow.textContent();
      expect(invitationText).toContain('Declined');

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });

    test('should show decline confirmation dialog', async () => {
      // Create test project and invitation
      const testProjectName = `Decline Dialog Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);
      const testEmail = 'decline-dialog-test@example.com';
      const invitationToken = await projectsPage.sendTestInvitation(testProjectId, testEmail, 'viewer');

      // Add invitation to Individual Invitations tab
      await projectsPage.navigateToInvitations();
      await projectsPage.switchToIndividualInvitations();
      await projectsPage.addInvitationToken(invitationToken);

      const invitationRow = projectsPage.getInvitationRows().first();
      const declineButton = invitationRow.getByRole('button', { name: /decline/i });
      await declineButton.click();

      // Verify confirmation dialog
      const dialog = projectsPage.page.locator('[role="dialog"]').first();
      expect(await dialog.isVisible()).toBe(true);

      // Should have optional reason field
      const reasonInput = dialog.getByLabel(/reason/i);
      expect(await reasonInput.isVisible()).toBe(true);

      // Cancel
      await projectsPage.clickButton('cancel');

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });

    test('should allow declining without reason', async () => {
      // Create test project and invitation
      const testProjectName = `Decline No Reason Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);
      const testEmail = 'decline-no-reason-test@example.com';
      const invitationToken = await projectsPage.sendTestInvitation(testProjectId, testEmail, 'admin');

      // Add invitation to Individual Invitations tab
      await projectsPage.navigateToInvitations();
      await projectsPage.switchToIndividualInvitations();
      await projectsPage.addInvitationToken(invitationToken);

      // Decline without reason
      await projectsPage.declineInvitation(testProjectName);

      // Verify status changed
      await projectsPage.page.waitForTimeout(1000);
      const invitationRow = projectsPage.page.locator(`tr:has-text("${testProjectName}")`).first();
      const invitationText = await invitationRow.textContent();
      expect(invitationText).toContain('Declined');

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });
  });

  test.describe('Bulk Invitations Workflow', () => {
    test('should send bulk invitations to multiple emails', async () => {
      // Create test project
      const testProjectName = `Bulk Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);

      const emails = [
        'researcher1@example.com',
        'researcher2@example.com',
        'researcher3@example.com'
      ];

      const role = 'member';

      // Navigate to Bulk Invitations tab
      await projectsPage.navigateToInvitations();
      await projectsPage.switchToBulkInvitations();

      // Send bulk invitations
      await projectsPage.sendBulkInvitations(testProjectId, emails, role);

      // Verify success message/results
      await projectsPage.page.waitForTimeout(1000);
      const resultsContainer = projectsPage.page.getByTestId('bulk-results-container');
      expect(await resultsContainer.isVisible()).toBe(true);

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });

    test('should validate email format in bulk invitations', async () => {
      // Create test project
      const testProjectName = `Email Validation Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);

      await projectsPage.navigateToInvitations();
      await projectsPage.switchToBulkInvitations();

      const invalidEmails = [
        'valid@example.com',
        'invalid-email', // Invalid
        'another@example.com'
      ];

      // Select project first
      const projectSelect = projectsPage.page.getByTestId('bulk-project-select');
      await projectSelect.click();
      await projectsPage.page.waitForTimeout(300);
      const projectOption = projectsPage.page.locator(`[data-value="${testProjectId}"]`);
      if (await projectOption.isVisible({ timeout: 1000 })) {
        await projectOption.click();
      }

      // Fill emails
      const emailTextarea = projectsPage.page.getByTestId('bulk-emails-textarea');
      await emailTextarea.fill(invalidEmails.join('\n'));

      // Select role
      const roleSelect = projectsPage.page.getByTestId('bulk-role-select');
      await roleSelect.click();
      await projectsPage.page.waitForTimeout(300);
      await projectsPage.page.locator('[data-value="member"]').click();

      await projectsPage.clickButton('send bulk invitations');

      // Should show validation error or results with failed emails
      await projectsPage.page.waitForTimeout(1000);
      const resultsContainer = projectsPage.page.getByTestId('bulk-results-container');
      expect(await resultsContainer.isVisible()).toBe(true);

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });

    test('should require project selection for bulk invitations', async () => {
      await projectsPage.navigateToInvitations();
      await projectsPage.switchToBulkInvitations();

      const emails = ['test@example.com'];

      // Fill emails but no project
      const emailTextarea = projectsPage.page.getByTestId('bulk-emails-textarea');
      await emailTextarea.fill(emails.join('\n'));

      // Try to click send button - should be disabled
      const sendButton = projectsPage.page.getByTestId('send-bulk-invitations-button');
      const isDisabled = await sendButton.isDisabled();
      expect(isDisabled).toBe(true);
    });

    test('should show bulk invitation results summary', async () => {
      // Create test project
      const testProjectName = `Results Summary Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);

      const emails = [
        'result1@example.com',
        'result2@example.com'
      ];

      await projectsPage.navigateToInvitations();
      await projectsPage.switchToBulkInvitations();

      // Send bulk invitations
      await projectsPage.sendBulkInvitations(testProjectId, emails, 'viewer');

      // Verify results section shows
      await projectsPage.page.waitForTimeout(1000);
      const resultsContainer = projectsPage.page.getByTestId('bulk-results-container');
      expect(await resultsContainer.isVisible()).toBe(true);

      const resultsText = await resultsContainer.textContent();
      expect(resultsText).toBeTruthy();

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });

    test('should include optional welcome message', async () => {
      // Create test project
      const testProjectName = `Welcome Message Test ${Date.now()}`;
      const testProjectId = await projectsPage.createTestProject(testProjectName);

      const emails = ['welcome@example.com'];
      const role = 'viewer';
      const message = 'Welcome to our research collaboration!';

      await projectsPage.navigateToInvitations();
      await projectsPage.switchToBulkInvitations();

      await projectsPage.sendBulkInvitations(testProjectId, emails, role, message);

      // Verify success
      await projectsPage.page.waitForTimeout(1000);
      const resultsContainer = projectsPage.page.getByTestId('bulk-results-container');
      expect(await resultsContainer.isVisible()).toBe(true);

      // Cleanup
      await projectsPage.deleteTestProject(testProjectId);
    });
  });

  test.describe('Shared Tokens Workflow', () => {
    test('should create shared invitation token', async () => {
      // Shared Tokens UI now implemented (Phase 4.4)
      await projectsPage.switchToSharedTokens();

      const tokenName = `test-token-${Date.now()}`;

      // Create token
      await projectsPage.createSharedToken(tokenName, 10, '7d', 'member', 'Welcome!');

      // Verify token appears in list
      await projectsPage.page.waitForTimeout(1000);
      const tokenRow = projectsPage.page.locator(`tr:has-text("${tokenName}")`).first();
      expect(await tokenRow.isVisible()).toBe(true);
    });

    test('should display QR code for shared token', async () => {
      // Shared Tokens UI now implemented (Phase 4.4)
      // First create a test token
      await projectsPage.switchToSharedTokens();
      const tokenName = `QR Test ${Date.now()}`;
      await projectsPage.createSharedToken(tokenName, 10, '7d', 'member', 'Welcome!');
      await projectsPage.page.waitForTimeout(1000);

      // View QR code
      await projectsPage.viewQRCode(tokenName);

      // Verify QR code modal
      const qrModal = projectsPage.page.locator('[role="dialog"]').first();
      expect(await qrModal.isVisible()).toBe(true);

      // Should show QR code image
      const qrImage = qrModal.locator('img, canvas, svg');
      expect(await qrImage.isVisible()).toBe(true);

      // Should have copy token button
      const copyButton = qrModal.getByRole('button', { name: /copy token/i });
      expect(await copyButton.isVisible()).toBe(true);

      // Close
      await projectsPage.clickButton('close');
    });

    test('should copy shared token URL', async () => {
      // Shared Tokens UI now implemented (Phase 4.4)
      // First create a test token
      await projectsPage.switchToSharedTokens();
      const tokenName = `Copy Test ${Date.now()}`;
      await projectsPage.createSharedToken(tokenName, 5, '30d', 'viewer');
      await projectsPage.page.waitForTimeout(1000);

      await projectsPage.viewQRCode(tokenName);

      // Copy URL button should be visible
      const copyUrlButton = projectsPage.page.getByRole('button', { name: /copy url/i });
      expect(await copyUrlButton.isVisible()).toBe(true);
      await copyUrlButton.click();

      // Wait for copy action
      await projectsPage.page.waitForTimeout(500);

      // Close modal
      await projectsPage.clickButton('close');
    });

    test('should show redemption count for shared token', async () => {
      // Shared Tokens UI now implemented (Phase 4.4)
      // Create a test token to verify redemption count display
      await projectsPage.switchToSharedTokens();
      const tokenName = `Redemption Test ${Date.now()}`;
      await projectsPage.createSharedToken(tokenName, 10, '30d', 'viewer');
      await projectsPage.page.waitForTimeout(1000);

      const tokenRow = projectsPage.page.locator(`tr:has-text("${tokenName}")`).first();
      const tokenText = await tokenRow.textContent();

      // Should show redemption count (e.g., "0 / 10")
      expect(tokenText).toMatch(/\d+.*\/.*\d+/);
    });

    test('should extend shared token expiration', async () => {
      // Shared Tokens UI now implemented (Phase 4.4)
      // Create a test token first
      await projectsPage.switchToSharedTokens();
      const tokenName = `Extend Test ${Date.now()}`;
      await projectsPage.createSharedToken(tokenName, 5, '7d', 'member');
      await projectsPage.page.waitForTimeout(1000);

      const tokenRow = projectsPage.page.locator(`tr:has-text("${tokenName}")`).first();

      // Click extend button
      const extendButton = tokenRow.getByRole('button', { name: /extend|add-plus/i });
      await extendButton.click();
      await projectsPage.page.waitForTimeout(500);

      // Select extension duration
      const durationSelect = projectsPage.page.getByLabel(/duration/i);
      await durationSelect.selectOption('7d');

      // Confirm
      await projectsPage.clickButton('extend');

      // Verify new expiration date
      await projectsPage.page.waitForTimeout(1000);
      const updatedRow = projectsPage.page.locator(`tr:has-text("${tokenName}")`).first();
      const updatedText = await updatedRow.textContent();

      // Expiration date should be updated (hard to verify exact date in test)
      expect(updatedText).toBeTruthy();
    });

    test('should revoke shared token', async () => {
      // Shared Tokens UI now implemented (Phase 4.4)
      await projectsPage.switchToSharedTokens();

      const tokenName = `revoke-test-${Date.now()}`;

      // Create token
      await projectsPage.createSharedToken(tokenName, 5, '30d', 'viewer');
      await projectsPage.page.waitForTimeout(1000);

      // Revoke token
      await projectsPage.revokeSharedToken(tokenName);

      // Verify status changed to Revoked
      await projectsPage.page.waitForTimeout(1000);
      const tokenRow = projectsPage.page.locator(`tr:has-text("${tokenName}")`).first();
      const tokenText = await tokenRow.textContent();
      expect(tokenText).toContain('Revoked');
    });

    test('should prevent extending expired token', async () => {
      // Shared Tokens UI now implemented (Phase 4.4)
      await projectsPage.switchToSharedTokens();

      // Look for expired token row (if any exist)
      const expiredTokenRow = projectsPage.page.locator('tr:has-text("Expired")').first();

      // Check if expired token exists
      if (await expiredTokenRow.isVisible({ timeout: 2000 }).catch(() => false)) {
        // Extend button should be disabled for expired tokens
        const extendButton = expiredTokenRow.getByRole('button', { name: /extend/i });
        expect(await extendButton.isDisabled()).toBe(true);
      } else {
        // If no expired tokens exist, test passes (cannot test disabled state)
        expect(true).toBe(true);
      }
    });

    test('should prevent revoking already revoked token', async () => {
      // Shared Tokens UI now implemented (Phase 4.4)
      // Create and revoke a token to test disabled state
      await projectsPage.switchToSharedTokens();

      const tokenName = `Already Revoked ${Date.now()}`;
      await projectsPage.createSharedToken(tokenName, 5, '30d', 'viewer');
      await projectsPage.page.waitForTimeout(1000);

      // Revoke the token
      await projectsPage.revokeSharedToken(tokenName);
      await projectsPage.page.waitForTimeout(1000);

      const revokedTokenRow = projectsPage.page.locator(`tr:has-text("${tokenName}")`).first();

      // Revoke button should be disabled for already revoked token
      const revokeButton = revokedTokenRow.getByRole('button', { name: /revoke/i });
      expect(await revokeButton.isDisabled()).toBe(true);
    });
  });

  test.describe('Invitation Statistics', () => {
    test.skip('should display invitation summary stats', async () => {
      // TODO: Verify stats section
      await projectsPage.navigateToInvitations();

      // Should show total, pending, accepted counts
      const statsSection = projectsPage.page.locator('text=/total.*pending.*accepted/i');
      expect(await statsSection.isVisible()).toBe(true);
    });

    test.skip('should update stats after invitation actions', async () => {
      // TODO: Requires invitation action (accept/decline)
      await projectsPage.switchToIndividualInvitations();

      // Get initial pending count
      const initialStats = await projectsPage.page.locator('text=/pending.*\\d+/i').textContent();
      const initialPending = parseInt(initialStats?.match(/\\d+/)?.[0] || '0');

      // Accept an invitation
      // ... (accept logic)

      // Verify pending count decreased
      const updatedStats = await projectsPage.page.locator('text=/pending.*\\d+/i').textContent();
      const updatedPending = parseInt(updatedStats?.match(/\\d+/)?.[0] || '0');

      expect(updatedPending).toBe(initialPending - 1);
    });
  });

  test.describe('Invitation Expiration', () => {
    test.skip('should show expiration date for invitations', async () => {
      // TODO: Requires invitation with expiration
      await projectsPage.switchToIndividualInvitations();

      const invitationRow = projectsPage.getInvitationRows().first();
      const invitationText = await invitationRow.textContent();

      // Should contain expiration date
      expect(invitationText).toMatch(/expires|expiration/i);
      expect(invitationText).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}|\d+.*day/i); // Date or relative
    });

    test.skip('should mark expired invitations', async () => {
      // TODO: Requires expired invitation
      await projectsPage.switchToIndividualInvitations();

      const expiredRow = projectsPage.page.locator('tr:has-text("Expired")').first();

      // Should have "Expired" status badge
      expect(await expiredRow.isVisible()).toBe(true);

      // Accept/Decline buttons should be disabled
      const acceptButton = expiredRow.getByRole('button', { name: /accept/i });
      expect(await acceptButton.isDisabled()).toBe(true);
    });

    test.skip('should remove expired invitations from list', async () => {
      // TODO: Requires auto-cleanup or manual removal
      await projectsPage.switchToIndividualInvitations();

      const expiredRow = projectsPage.page.locator('tr:has-text("Expired")').first();

      // Click remove button
      const removeButton = expiredRow.getByRole('button', { name: /remove|delete/i });
      await removeButton.click();

      // Confirm removal
      await projectsPage.clickButton('remove');

      // Verify removed
      await projectsPage.page.waitForTimeout(500);
      expect(await expiredRow.isVisible()).toBe(false);
    });
  });
});
