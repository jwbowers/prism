/**
 * Idle Detection Workflows E2E Tests — v0.28.1
 *
 * Tests for the Idle Detection view (Settings > Advanced > Idle Detection).
 * This view provides automatic cost optimization through idle detection and hibernation.
 *
 * Coverage:
 *   - Navigation to Idle Detection via Settings Advanced sub-nav
 *   - Idle Detection header and stats counters visible
 *   - Tab navigation (Idle Policies, Schedules)
 *   - Idle policies table present
 *
 * Requires: prismd running with PRISM_TEST_MODE=true
 */

import { test, expect } from '@playwright/test';
import { IdlePage } from './pages/IdlePage';

test.describe('Idle Detection', () => {
  let idlePage: IdlePage;

  test.beforeEach(async ({ page, context }) => {
    // Dismiss onboarding modal before navigation
    await context.addInitScript(() => {
      localStorage.setItem('prism_onboarding_complete', 'true');
    });
    idlePage = new IdlePage(page);
    await idlePage.goto();
  });

  test('navigates to Idle Detection view via Settings Advanced sub-nav', async ({ page }) => {
    await idlePage.navigate();

    // Idle Detection & Hibernation header should be visible
    // Use .first() because "About Idle Detection" h2 also matches the pattern
    await expect(page.getByRole('heading', { name: /idle detection/i }).first()).toBeVisible({ timeout: 10000 });
  });

  test('displays idle detection stats counters', async ({ page }) => {
    await idlePage.navigate();

    // Stats containers should be visible
    await expect(page.locator('text=Active Policies')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=Total Policies')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=Monitored Workspaces')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=Cost Savings')).toBeVisible({ timeout: 5000 });
  });

  test('shows idle policies table on Policies tab', async ({ page }) => {
    await idlePage.navigate();

    // Policies tab is active by default — idle policies table should be present
    const policiesTable = idlePage.getPoliciesTable();
    await expect(policiesTable).toBeVisible({ timeout: 5000 });
  });

  test('can switch to Schedules tab', async ({ page }) => {
    await idlePage.navigate();

    // Schedules tab should be present
    const schedulesTab = page.getByRole('tab', { name: /schedules/i });
    await expect(schedulesTab).toBeVisible({ timeout: 5000 });

    // Click Schedules tab
    await idlePage.clickSchedulesTab();

    // Tab should still be visible after click
    await expect(schedulesTab).toBeVisible();
  });
});
