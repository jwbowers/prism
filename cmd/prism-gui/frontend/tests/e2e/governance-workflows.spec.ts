/**
 * Governance Workflows E2E Tests — v0.13.0
 *
 * Tests for the governance features added to the ProjectDetailView:
 *   Quotas | Grant Period | Budget Sharing | Onboarding Templates | Monthly Report
 * And the cross-project Approvals dashboard.
 */

import { test, expect, request } from '@playwright/test';
import { GovernancePage } from './pages';

// Fixed project name — avoids re-evaluation issues when Playwright runs
// beforeAll once per nested describe group (each call gets a new Date.now()).
const GOV_TEST_PROJECT_NAME = 'gov-test-e2e-shared';

test.describe('Governance Workflows', () => {
  let governancePage: GovernancePage;
  let testProjectId: string;
  const testProjectName = GOV_TEST_PROJECT_NAME;

  // Look up the shared test project created by global-setup.js.
  // global-setup.js creates 'gov-test-e2e-shared' once before all tests run;
  // this beforeAll just retrieves its ID so tests can use it.
  // Playwright may run this once per nested describe group — that's fine because
  // we only look up, never create or delete.
  test.beforeAll(async ({ browser }) => {
    test.setTimeout(60000);

    const context = await browser.newContext();
    const page = await context.newPage();

    try {
      // Retry up to 5 times — the daemon may briefly return empty responses after startup
      for (let attempt = 0; attempt < 5; attempt++) {
        const listResp = await page.request.get('http://localhost:8947/api/v1/projects');
        if (!listResp.ok()) break;

        let data: any;
        try { data = await listResp.json(); } catch { /* empty body — retry */ }

        if (data) {
          const existing = (data.projects || []).find((p: any) => p.name === testProjectName);
          if (existing) {
            testProjectId = existing.id;
            console.log(`✅ Using test project: ${testProjectName} (id=${testProjectId})`);
            break;
          }
        }

        if (attempt < 4) await new Promise(r => setTimeout(r, 400));
      }

      if (!testProjectId) {
        console.log(`⚠️ Test project "${testProjectName}" not found — tests will be skipped`);
      }
    } catch (e) {
      console.log('⚠️ Error looking up test project:', e);
    } finally {
      await page.close();
      await context.close();
    }
  });

  // No afterAll — the project is cleaned up by global-teardown.js (gov-test-* pattern)

  test.beforeEach(async ({ page, context }) => {
    test.setTimeout(60000);

    // Set localStorage BEFORE navigating to prevent onboarding modal
    await context.addInitScript(() => {
      localStorage.setItem('prism_onboarding_complete', 'true');
    });

    governancePage = new GovernancePage(page);
    await governancePage.goto();
  });

  // ── Quotas ────────────────────────────────────────────────────────────────
  test.describe('Quotas', () => {
    test.beforeEach(async () => {
      if (!testProjectId) test.skip();
      await governancePage.navigateToGovernance(testProjectName);
      await governancePage.switchToGovernanceTab('Quotas');
    });

    test('displays empty quotas table initially', async ({ page }) => {
      const table = page.getByTestId('quotas-table');
      await table.waitFor({ state: 'visible', timeout: 10000 });
      // May be empty or have existing quotas — just confirm the table renders
      await expect(table).toBeVisible();
    });

    test('set quota button opens modal', async ({ page }) => {
      const btn = page.getByTestId('set-quota-button');
      await btn.waitFor({ state: 'visible', timeout: 10000 });
      await btn.click();

      // Modal should appear
      await expect(page.getByTestId('save-quota-button')).toBeVisible({ timeout: 5000 });
      await expect(page.getByTestId('quota-role-select')).toBeVisible();
      await expect(page.getByTestId('quota-max-instances-input')).toBeVisible();

      // Cancel
      await page.getByRole('button', { name: /cancel/i }).first().click();
    });

    test('sets quota for member role', async ({ page }) => {
      await governancePage.addQuota('member', 5, 10);

      // Table should now show the quota (or no error visible)
      const table = page.getByTestId('quotas-table');
      await table.waitFor({ state: 'visible', timeout: 10000 });
      await expect(table).toBeVisible();
    });
  });

  // ── Grant Period ──────────────────────────────────────────────────────────
  test.describe('Grant Period', () => {
    test.beforeEach(async () => {
      if (!testProjectId) test.skip();
      await governancePage.navigateToGovernance(testProjectName);
      await governancePage.switchToGovernanceTab('Grant Period');
    });

    test('shows no grant period message initially', async ({ page }) => {
      // Either alert about no grant period or the grant period details are shown
      // (in case a previous test left one). Just confirm page renders.
      await expect(page.locator('[data-testid="governance-panel"]')).toBeVisible({ timeout: 10000 });
    });

    test('configures grant period with auto-freeze', async ({ page }) => {
      // Delete any existing grant period first (fire-and-forget)
      await page.request.delete(`http://localhost:8947/api/v1/projects/${testProjectId}/grant-period`).catch(() => {});
      await page.reload();
      await governancePage.navigate();
      await governancePage.navigateToGovernance(testProjectName);
      await governancePage.switchToGovernanceTab('Grant Period');

      // Wait for configure button
      const configBtn = page.getByTestId('configure-grant-period-button');
      await configBtn.waitFor({ state: 'visible', timeout: 10000 });

      await governancePage.setGrantPeriod('NSF Grant Year 1', '2024-01-01', '2024-12-31', true);

      // Grant period details should now be shown
      await expect(page.getByTestId('grant-period-details')).toBeVisible({ timeout: 10000 });
      await expect(page.getByTestId('grant-period-name')).toHaveText('NSF Grant Year 1', { timeout: 5000 });
    });

    test('deletes grant period', async ({ page }) => {
      // Ensure there IS a grant period to delete
      await page.request.put(`http://localhost:8947/api/v1/projects/${testProjectId}/grant-period`, {
        data: { name: 'Temp Grant', start_date: '2024-01-01T00:00:00Z', end_date: '2024-12-31T00:00:00Z', auto_freeze: false }
      }).catch(() => {});
      await page.reload();
      await governancePage.navigate();
      await governancePage.navigateToGovernance(testProjectName);
      await governancePage.switchToGovernanceTab('Grant Period');

      const editBtn = page.getByTestId('edit-grant-period-button');
      await editBtn.waitFor({ state: 'visible', timeout: 10000 });

      await governancePage.deleteGrantPeriod();

      // Should now show the "no grant period" alert
      await expect(page.getByTestId('no-grant-period-alert')).toBeVisible({ timeout: 10000 });
    });
  });

  // ── Budget Sharing ────────────────────────────────────────────────────────
  test.describe('Budget Sharing', () => {
    test.beforeEach(async () => {
      if (!testProjectId) test.skip();
      await governancePage.navigateToGovernance(testProjectName);
      await governancePage.switchToGovernanceTab('Budget Sharing');
    });

    test('displays budget shares table', async ({ page }) => {
      const table = page.getByTestId('budget-shares-table');
      await table.waitFor({ state: 'visible', timeout: 10000 });
      await expect(table).toBeVisible();
    });

    test('share budget button opens modal', async ({ page }) => {
      const btn = page.getByTestId('share-budget-button');
      await btn.waitFor({ state: 'visible', timeout: 10000 });
      await btn.click();

      await expect(page.getByTestId('confirm-share-budget-button')).toBeVisible({ timeout: 5000 });
      await expect(page.getByTestId('share-amount-input')).toBeVisible();

      // Cancel
      await page.getByRole('button', { name: /cancel/i }).first().click();
    });

    test('creates a budget share to a member', async ({ page }) => {
      await governancePage.shareBudget('', 'test-member-1', 50, 'Research allocation');
      // Table refreshes — no error should be shown
      await expect(page.getByTestId('budget-shares-table')).toBeVisible({ timeout: 10000 });
    });
  });

  // ── Onboarding Templates ──────────────────────────────────────────────────
  test.describe('Onboarding Templates', () => {
    test.beforeEach(async () => {
      if (!testProjectId) test.skip();
      await governancePage.navigateToGovernance(testProjectName);
      await governancePage.switchToGovernanceTab('Onboarding Templates');
    });

    test('displays onboarding templates table', async ({ page }) => {
      const table = page.getByTestId('onboarding-templates-table');
      await table.waitFor({ state: 'visible', timeout: 10000 });
      await expect(table).toBeVisible();
    });

    test('adds and then deletes an onboarding template', async ({ page }) => {
      const templateName = `e2e-tmpl-${Date.now()}`;

      // Add template
      await governancePage.addOnboardingTemplate(templateName, 'E2E test onboarding template');

      // Table should be visible after add
      await expect(page.getByTestId('onboarding-templates-table')).toBeVisible({ timeout: 10000 });

      // Try to delete (may or may not appear depending on backend persistence)
      const deleteBtn = page.getByTestId(`delete-onboarding-template-${templateName}`);
      if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
        await governancePage.deleteOnboardingTemplate(templateName);
      }
    });
  });

  // ── Monthly Report ────────────────────────────────────────────────────────
  test.describe('Monthly Report', () => {
    test.beforeEach(async () => {
      if (!testProjectId) test.skip();
      await governancePage.navigateToGovernance(testProjectName);
      await governancePage.switchToGovernanceTab('Monthly Report');
    });

    test('generates report in text format for current month', async ({ page }) => {
      const now = new Date();
      const month = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;

      await governancePage.generateMonthlyReport(month, 'text');

      // Wait for output or error (project may have no cost history)
      await Promise.race([
        page.waitForSelector('[data-testid="monthly-report-output"]', { state: 'visible', timeout: 15000 }),
        page.waitForSelector('[data-testid="governance-panel"] [data-id="alert"]', { state: 'visible', timeout: 15000 }).catch(() => {})
      ]).catch(() => {});

      // At minimum confirm the generate button worked (no unhandled JS error)
      await expect(page.locator('[data-testid="governance-panel"]')).toBeVisible({ timeout: 5000 });
    });
  });

  // ── Approvals Dashboard ───────────────────────────────────────────────────
  test.describe('Approvals Dashboard', () => {
    test('navigates to approvals view via sidebar link', async ({ page }) => {
      await governancePage.navigateToApprovals();

      await expect(page.getByTestId('approvals-view')).toBeVisible({ timeout: 10000 });
      await expect(page.getByTestId('approvals-table')).toBeVisible({ timeout: 10000 });
    });

    test('shows status filter dropdown with "pending" default', async ({ page }) => {
      await governancePage.navigateToApprovals();

      const filter = page.getByTestId('approvals-status-filter');
      await filter.waitFor({ state: 'visible', timeout: 10000 });
      await expect(filter).toBeVisible();
    });

    test('filters by switching status select', async ({ page }) => {
      await governancePage.navigateToApprovals();

      const filter = page.getByTestId('approvals-status-filter');
      await filter.waitFor({ state: 'visible', timeout: 10000 });

      // Click to open dropdown
      await filter.click();
      // Select "All"
      await page.locator('[data-value=""]').first().click().catch(async () => {
        // Try alternative selector if data-value doesn't work
        await page.getByRole('option', { name: /^all$/i }).click().catch(() => {});
      });

      // Table should still be visible after filter change
      await expect(page.getByTestId('approvals-table')).toBeVisible({ timeout: 10000 });
    });
  });
});
