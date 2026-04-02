/**
 * AMI Management Workflows E2E Tests — v0.28.1
 *
 * Tests for the AMI Management view (Settings > Advanced > AMI Management).
 * This view allows managing Amazon Machine Images for fast workspace launching.
 *
 * Coverage:
 *   - Navigation to AMI Management via Settings Advanced sub-nav
 *   - AMI Management header and stats counters visible
 *   - Build AMI button present
 *   - Tab navigation (AMIs, Builds, Regions)
 *
 * Requires: prismd running with PRISM_TEST_MODE=true
 */

import { test, expect } from '@playwright/test';
import { AMIPage } from './pages/AMIPage';

test.describe('AMI Management', () => {
  let amiPage: AMIPage;

  test.beforeEach(async ({ page, context }) => {
    // Dismiss onboarding modal before navigation
    await context.addInitScript(() => {
      localStorage.setItem('prism_onboarding_complete', 'true');
    });
    amiPage = new AMIPage(page);
    await amiPage.goto();
  });

  test('navigates to AMI Management view via Settings Advanced sub-nav', async ({ page }) => {
    await amiPage.navigate();

    // AMI Management header should be visible
    await expect(page.getByRole('heading', { name: /ami management/i })).toBeVisible({ timeout: 10000 });
  });

  test('displays AMI Management stats counters', async ({ page }) => {
    await amiPage.navigate();

    // Stats containers should be visible (Total AMIs, Total Size, Monthly Cost, Regions)
    await expect(page.getByRole('heading', { name: 'Total AMIs' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'Total Size' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'Monthly Cost' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'Regions' }).first()).toBeVisible({ timeout: 5000 });
  });

  test('shows Build AMI button', async ({ page }) => {
    await amiPage.navigate();

    // Build AMI primary action button should be present
    // Use .first() because Cloudscape keeps modal content in DOM even when hidden
    const buildButton = page.getByRole('button', { name: /build ami/i }).first();
    await expect(buildButton).toBeVisible({ timeout: 5000 });
  });

  test('can switch between AMI tabs (AMIs, Builds, Regions)', async ({ page }) => {
    await amiPage.navigate();

    // Verify tabs are present
    await expect(page.getByRole('tab', { name: /^amis$/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('tab', { name: /build status/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('tab', { name: /regional coverage/i })).toBeVisible({ timeout: 5000 });

    // Click Builds tab
    await amiPage.clickBuildsTab();
    // Verify we're on builds tab (tab should be selected)
    const buildsTab = page.getByRole('tab', { name: /build status/i });
    await expect(buildsTab).toBeVisible();

    // Click Regions tab
    await amiPage.clickRegionsTab();
    const regionsTab = page.getByRole('tab', { name: /regional coverage/i });
    await expect(regionsTab).toBeVisible();

    // Navigate back to AMIs tab
    await amiPage.clickAMIsTab();
    const amisTab = page.getByRole('tab', { name: /^amis$/i });
    await expect(amisTab).toBeVisible();
  });
});
