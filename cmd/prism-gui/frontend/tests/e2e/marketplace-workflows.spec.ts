/**
 * Template Marketplace Workflows E2E Tests — v0.28.1
 *
 * Tests for the Template Marketplace view (Settings > Advanced > Template Marketplace).
 * This view allows discovering and installing community-contributed research templates.
 *
 * Coverage:
 *   - Navigation to Template Marketplace via Settings Advanced sub-nav
 *   - Marketplace header visible
 *   - Search input present
 *   - Category filter present
 *
 * Requires: prismd running with PRISM_TEST_MODE=true
 */

import { test, expect } from '@playwright/test';
import { MarketplacePage } from './pages/MarketplacePage';

test.describe('Template Marketplace', () => {
  let marketplacePage: MarketplacePage;

  test.beforeEach(async ({ page, context }) => {
    // Dismiss onboarding modal before navigation
    await context.addInitScript(() => {
      localStorage.setItem('prism_onboarding_complete', 'true');
    });
    marketplacePage = new MarketplacePage(page);
    await marketplacePage.goto();
  });

  test('navigates to Template Marketplace via Settings Advanced sub-nav', async ({ page }) => {
    await marketplacePage.navigate();

    // Template Marketplace header should be visible
    await expect(page.getByRole('heading', { name: /template marketplace/i })).toBeVisible({ timeout: 10000 });
  });

  test('shows search input for templates', async ({ page }) => {
    await marketplacePage.navigate();

    // Search input should be present
    const searchInput = page.getByPlaceholder(/search templates/i);
    await expect(searchInput).toBeVisible({ timeout: 5000 });
  });

  test('shows category filter', async ({ page }) => {
    await marketplacePage.navigate();

    // Category filter label should be present
    await expect(page.locator('text=Category').first()).toBeVisible({ timeout: 5000 });
  });
});
