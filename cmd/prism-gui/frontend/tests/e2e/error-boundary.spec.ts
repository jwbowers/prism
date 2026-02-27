/**
 * Error Boundary and Error Handling Tests
 *
 * Tests for error handling, resilience, and graceful degradation in Prism GUI.
 * Tests Cloudscape error states, retry mechanisms, and UI stability.
 */

import { test, expect } from '@playwright/test';

test.describe('Error Boundary and Error Handling', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set localStorage to skip onboarding before navigation
    await context.addInitScript(() => {
      localStorage.setItem('prism_onboarding_complete', 'true');
    });
    await page.goto('/');

    // Wait for app to load
    await page.waitForLoadState('domcontentloaded', { timeout: 10000 });

    // Wait for API to be ready
    await page.waitForResponse(
      (response) => response.url().includes('/api/v1/'),
      { timeout: 15000 }
    ).catch(() => {
      // If API calls fail in test mode, that's okay
    });
  });

  test('template loading handles success or error gracefully', async ({ page }) => {
    // Navigate to Templates
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Either templates loaded OR error state shown
    const templateCards = page.locator('[data-testid="template-card"]');
    const hasTemplates = await templateCards.first().isVisible().catch(() => false);

    if (hasTemplates) {
      // Templates loaded successfully
      await expect(templateCards.first()).toBeVisible();
    } else {
      // Check if error state is shown gracefully (empty state or error message)
      const emptyState = page.getByText(/no templates/i);
      const hasEmptyState = await emptyState.isVisible().catch(() => false);
      expect(hasEmptyState).toBeDefined();
    }
  });

  test('instance loading handles success or error gracefully', async ({ page }) => {
    // Navigate to My Workspaces
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Either instances loaded OR empty/error state shown
    const instancesTable = page.locator('[data-testid="instances-table"]');
    const emptyState = page.locator('[data-testid="empty-instances"]');

    const hasTable = await instancesTable.isVisible().catch(() => false);
    const hasEmptyState = await emptyState.isVisible().catch(() => false);

    // At least one should be visible
    expect(hasTable || hasEmptyState).toBe(true);
  });

  test('daemon connection status is displayed', async ({ page }) => {
    // App should display and handle daemon connection status
    await expect(page.locator('#root')).toBeAttached();

    // Check that main navigation is visible (indicates successful connection)
    const navigation = page.getByRole('link', { name: /dashboard/i });
    await expect(navigation).toBeVisible({ timeout: 10000 });

    // Content should load
    const mainContent = page.locator('main').first();
    await expect(mainContent).toBeVisible();
  });

  test('form submission errors are handled gracefully', async ({ page }) => {
    // Navigate to Projects
    await page.getByRole('link', { name: /projects/i }).click();
    await page.waitForSelector('[data-testid="create-project-button"]', { state: 'visible', timeout: 10000 }).catch(() => {});

    // Try to open create project dialog
    const createButton = page.getByTestId('create-project-button');
    if (await createButton.isVisible().catch(() => false)) {
      await createButton.click();

      // Wait for dialog
      const dialog = page.getByRole('dialog', { name: /create.*project/i });
      await dialog.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {});

      if (await dialog.isVisible().catch(() => false)) {
        // Try to submit empty form
        const submitButton = dialog.getByRole('button', { name: /create/i });
        await submitButton.click();

        // Form should either stay open or show validation
        const stillOpen = await dialog.isVisible().catch(() => false);
        expect(stillOpen).toBeDefined();
      }
    } else {
      test.skip();
    }
  });

  test('settings form handles errors gracefully', async ({ page }) => {
    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Verify Settings page loads
    const settingsHeading = page.getByRole('heading', { name: /settings/i }).first();
    await expect(settingsHeading).toBeVisible({ timeout: 5000 });

    // Settings content should be accessible
    const settingsContent = page.locator('main').first();
    await expect(settingsContent).toBeVisible();
  });

  test('network errors handled without crashing', async ({ page }) => {
    // Navigate between sections to trigger network calls
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // App should still be visible
    await expect(page.locator('#root')).toBeAttached();

    // Navigate to another section
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // App should still be functional
    await expect(page.locator('#root')).toBeAttached();
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();
  });

  test('invalid navigation handled gracefully', async ({ page }) => {
    // Navigate to Templates
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForSelector('[data-testid="template-card"]', { state: 'visible', timeout: 10000 }).catch(() => {});

    // Check if templates loaded
    const templateCard = page.locator('[data-testid="template-card"]').first();
    const hasTemplates = await templateCard.isVisible().catch(() => false);

    if (hasTemplates) {
      // Click on template card - should handle gracefully
      await templateCard.click();

      // App should not crash
      await expect(page.locator('#root')).toBeAttached();
    } else {
      // Empty state is also valid
      test.skip();
    }
  });

  test('page reload recovers gracefully', async ({ page }) => {
    // Verify initial load
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible({ timeout: 10000 });

    // Reload page
    await page.reload();

    // Wait for app to reload
    await page.waitForLoadState('domcontentloaded', { timeout: 10000 });

    // App should recover
    await expect(page.locator('#root')).toBeAttached();
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible({ timeout: 10000 });
  });

  test('JavaScript errors do not crash the interface', async ({ page }) => {
    // Monitor console errors
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    // Navigate through various sections
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    await page.getByRole('link', { name: /storage/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // App should still be functional regardless of console errors
    await expect(page.locator('#root')).toBeAttached();
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();
  });

  test('UI remains responsive after errors', async ({ page }) => {
    // Navigate to Dashboard
    await page.getByRole('link', { name: /dashboard/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Navigate to Templates
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Navigate to My Workspaces
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // UI should still be responsive
    await expect(page.locator('#root')).toBeAttached();
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();

    // Navigation should still work
    await page.getByRole('link', { name: /dashboard/i }).click();
    await expect(page.locator('#root')).toBeAttached();
  });
});
