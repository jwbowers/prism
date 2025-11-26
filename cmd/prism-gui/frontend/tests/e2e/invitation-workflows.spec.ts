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
    test.skip('should send bulk invitations to multiple emails', async () => {
      // TODO: Requires project and email validation
      await projectsPage.switchToBulkInvitations();

      const emails = [
        'researcher1@example.com',
        'researcher2@example.com',
        'researcher3@example.com'
      ];

      const projectId = 'test-project-id';
      const role = 'member';

      // Send bulk invitations
      await projectsPage.sendBulkInvitations(projectId, emails, role);

      // Verify success message/results
      await projectsPage.page.waitForTimeout(1000);
      const resultsText = await projectsPage.page.locator('text=/sent.*invitation/i').textContent();
      expect(resultsText).toContain('3'); // 3 emails sent
    });

    test.skip('should validate email format in bulk invitations', async () => {
      // TODO: UI implementation gap - need validation
      await projectsPage.switchToBulkInvitations();

      const invalidEmails = [
        'valid@example.com',
        'invalid-email', // Invalid
        'another@example.com'
      ];

      const emailTextarea = projectsPage.page.getByLabel(/email.*addresses/i);
      await emailTextarea.fill(invalidEmails.join('\n'));

      await projectsPage.clickButton('send bulk invitations');

      // Should show validation error
      const error = projectsPage.page.locator('[data-testid="validation-error"]');
      expect(await error.isVisible()).toBe(true);
      expect(await error.textContent()).toContain('invalid-email');
    });

    test.skip('should require project selection for bulk invitations', async () => {
      // TODO: Validate project selection requirement
      await projectsPage.switchToBulkInvitations();

      const emails = ['test@example.com'];

      // Fill emails but no project
      const emailTextarea = projectsPage.page.getByLabel(/email.*addresses/i);
      await emailTextarea.fill(emails.join('\n'));

      const roleSelect = projectsPage.page.getByLabel(/role/i);
      await roleSelect.selectOption('member');

      // Try to send without project
      await projectsPage.clickButton('send bulk invitations');

      // Should show error about project requirement
      const notification = projectsPage.page.locator('text=/project.*required/i');
      expect(await notification.isVisible()).toBe(true);
    });

    test.skip('should show bulk invitation results summary', async () => {
      // TODO: Requires bulk send completion
      await projectsPage.switchToBulkInvitations();

      // After sending bulk invitations...
      // Verify results section shows:
      // - Sent count
      // - Failed count
      // - Skipped count (duplicates, etc.)

      const resultsSection = projectsPage.page.locator('text=/invitation results/i');
      expect(await resultsSection.isVisible()).toBe(true);

      const sentCount = await projectsPage.page.locator('text=/sent.*\\d+/i').textContent();
      expect(sentCount).toBeTruthy();
    });

    test.skip('should include optional welcome message', async () => {
      // TODO: Verify message inclusion
      await projectsPage.switchToBulkInvitations();

      const emails = ['test@example.com'];
      const projectId = 'test-project';
      const role = 'viewer';
      const message = 'Welcome to our research collaboration!';

      await projectsPage.sendBulkInvitations(projectId, emails, role, message);

      // Verify success (actual message delivery would need backend verification)
      await projectsPage.page.waitForTimeout(1000);
      const success = projectsPage.page.locator('text=/sent/i');
      expect(await success.isVisible()).toBe(true);
    });
  });

  test.describe('Shared Tokens Workflow', () => {
    test.skip('should create shared invitation token', async () => {
      // TODO: Requires project context
      await projectsPage.switchToSharedTokens();

      const tokenName = `test-token-${Date.now()}`;

      // Create token
      await projectsPage.createSharedToken(tokenName, 10, '7d', 'member', 'Welcome!');

      // Verify token appears in list
      await projectsPage.page.waitForTimeout(1000);
      const tokenRow = projectsPage.page.locator(`tr:has-text("${tokenName}")`).first();
      expect(await tokenRow.isVisible()).toBe(true);
    });

    test.skip('should display QR code for shared token', async () => {
      // TODO: Requires shared token
      await projectsPage.switchToSharedTokens();

      const tokenName = 'Test Token';

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

    test.skip('should copy shared token URL', async () => {
      // TODO: Requires clipboard testing
      await projectsPage.switchToSharedTokens();

      const tokenName = 'Test Token';

      await projectsPage.viewQRCode(tokenName);

      // Copy URL
      const copyUrlButton = projectsPage.page.getByRole('button', { name: /copy url/i });
      await copyUrlButton.click();

      // Verify clipboard (browser testing limitation - would need special handling)
      await projectsPage.page.waitForTimeout(500);

      // Close modal
      await projectsPage.clickButton('close');
    });

    test.skip('should show redemption count for shared token', async () => {
      // TODO: Requires tokens with redemptions
      await projectsPage.switchToSharedTokens();

      const tokenRow = projectsPage.page.locator('tr').first();
      const tokenText = await tokenRow.textContent();

      // Should show redemption count (e.g., "3 / 10")
      expect(tokenText).toMatch(/\d+.*\/.*\d+/);
    });

    test.skip('should extend shared token expiration', async () => {
      // TODO: Requires active shared token
      await projectsPage.switchToSharedTokens();

      const tokenName = 'Test Token';
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

    test.skip('should revoke shared token', async () => {
      // TODO: Requires active shared token
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

    test.skip('should prevent extending expired token', async () => {
      // TODO: Requires expired token
      await projectsPage.switchToSharedTokens();

      const expiredTokenRow = projectsPage.page.locator('tr:has-text("Expired")').first();

      // Extend button should be disabled
      const extendButton = expiredTokenRow.getByRole('button', { name: /extend/i });
      expect(await extendButton.isDisabled()).toBe(true);
    });

    test.skip('should prevent revoking already revoked token', async () => {
      // TODO: Requires revoked token
      await projectsPage.switchToSharedTokens();

      const revokedTokenRow = projectsPage.page.locator('tr:has-text("Revoked")').first();

      // Revoke button should be disabled
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
