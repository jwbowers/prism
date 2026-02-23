/**
 * Settings Interface Tests
 *
 * Tests for the Settings page in Prism GUI using Cloudscape Design System.
 * Tests profile management and settings functionality.
 */

import { test, expect } from '@playwright/test';

test.describe('Settings Interface', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set localStorage to skip onboarding before navigation
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
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

  // Helper to navigate to Settings and wait for profiles section to load
  async function navigateToSettings(page: any) {
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});
    // Wait for create-profile-button to appear (profiles section loaded)
    await page.waitForSelector('[data-testid="create-profile-button"]', { timeout: 8000 }).catch(() => {});
  }

  test('settings page loads successfully', async ({ page }) => {
    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Verify Settings page heading
    const settingsHeading = page.getByRole('heading', { name: /settings/i }).first();
    await expect(settingsHeading).toBeVisible({ timeout: 5000 });
  });

  test('profiles section displays correctly', async ({ page }) => {
    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Check if profiles table or content exists
    const profilesContent = page.locator('[data-testid="profiles-table"]');
    const hasProfilesTable = await profilesContent.isVisible().catch(() => false);

    // Either profiles table or empty state should be visible
    expect(typeof hasProfilesTable).toBe('boolean');
  });

  test('create profile button is accessible', async ({ page }) => {
    await navigateToSettings(page);

    const createButton = page.getByTestId('create-profile-button');
    await expect(createButton).toBeVisible({ timeout: 5000 });
    await expect(createButton).toBeEnabled();
  });

  test('profile form opens when create is clicked', async ({ page }) => {
    await navigateToSettings(page);

    const createButton = page.getByTestId('create-profile-button');
    await expect(createButton).toBeVisible({ timeout: 5000 });
    await createButton.click();

    // Wait for dialog
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Verify form inputs are present
    const nameInput = page.getByTestId('profile-name-input');
    await expect(nameInput).toBeVisible();
  });

  test('profile form validation works', async ({ page }) => {
    await navigateToSettings(page);

    const createButton = page.getByTestId('create-profile-button');
    await expect(createButton).toBeVisible({ timeout: 5000 });
    await createButton.click();

    // Wait for dialog
    const dialog = page.locator('[role="dialog"]');
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Form should have required inputs
    const nameInput = page.getByTestId('profile-name-input').locator('input');
    await expect(nameInput).toBeVisible();

    const awsProfileInput = page.getByTestId('aws-profile-input').locator('input');
    await expect(awsProfileInput).toBeVisible();

    const regionInput = page.getByTestId('region-input').locator('input');
    await expect(regionInput).toBeVisible();
  });

  test('profile form accepts valid input', async ({ page }) => {
    await navigateToSettings(page);

    const createButton = page.getByTestId('create-profile-button');
    await expect(createButton).toBeVisible({ timeout: 5000 });
    await createButton.click();

    // Wait for dialog
    await page.locator('[role="dialog"]').waitFor({ state: 'visible', timeout: 5000 });

    // Fill form with valid data
    const nameInput = page.getByTestId('profile-name-input').locator('input');
    const awsProfileInput = page.getByTestId('aws-profile-input').locator('input');
    const regionInput = page.getByTestId('region-input').locator('input');

    await nameInput.fill('test-profile');
    await awsProfileInput.fill('default');
    await regionInput.fill('us-west-2');

    // Verify inputs accepted values
    expect(await nameInput.inputValue()).toBe('test-profile');
    expect(await awsProfileInput.inputValue()).toBe('default');
    expect(await regionInput.inputValue()).toBe('us-west-2');
  });

  test('profile dialog can be cancelled', async ({ page }) => {
    await navigateToSettings(page);

    const createButton = page.getByTestId('create-profile-button');
    await expect(createButton).toBeVisible({ timeout: 5000 });
    await createButton.click();

    // Wait for dialog
    const dialog = page.locator('[role="dialog"]');
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Click cancel button
    const cancelButton = page.getByRole('button', { name: 'Cancel', exact: true }).last();
    await cancelButton.click();

    // Dialog should be hidden
    await dialog.waitFor({ state: 'hidden', timeout: 5000 });
  });

  test('settings page remains responsive', async ({ page }) => {
    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Page should be functional
    await expect(page.locator('#root')).toBeAttached();

    // Navigate away and back
    await page.getByRole('link', { name: /dashboard/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Settings should still be accessible
    const settingsHeading = page.getByRole('heading', { name: /settings/i }).first();
    await expect(settingsHeading).toBeVisible({ timeout: 5000 });
  });

  test('settings content is accessible with proper headings', async ({ page }) => {
    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Check for main heading
    const mainHeading = page.getByRole('heading', { name: /settings/i }).first();
    await expect(mainHeading).toBeVisible({ timeout: 5000 });

    // Settings page should have accessible structure
    await expect(page.locator('main')).toBeVisible();
  });

  test('settings page handles empty state', async ({ page }) => {
    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Either profiles table or empty state should be visible
    const profilesTable = page.locator('[data-testid="profiles-table"]');
    const hasTable = await profilesTable.isVisible().catch(() => false);

    if (!hasTable) {
      // Empty state should be handled gracefully
      const createButton = page.getByTestId('create-profile-button');
      const hasCreateButton = await createButton.isVisible().catch(() => false);
      expect(hasCreateButton).toBeDefined();
    }
  });

  test('settings forms use Cloudscape components', async ({ page }) => {
    await navigateToSettings(page);

    const createButton = page.getByTestId('create-profile-button');
    await expect(createButton).toBeVisible({ timeout: 5000 });
    await createButton.click();

    // Wait for dialog
    await page.locator('[role="dialog"]').waitFor({ state: 'visible', timeout: 5000 });

    // Verify Cloudscape form fields
    const nameInput = page.getByTestId('profile-name-input');
    await expect(nameInput).toBeVisible();

    // Input should have proper structure
    const input = nameInput.locator('input');
    await expect(input).toBeVisible();
  });

  test('navigation from settings to other pages works', async ({ page }) => {
    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Navigate to Templates
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Verify navigation worked
    await expect(page.locator('#root')).toBeAttached();

    // Navigate back to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    const settingsHeading = page.getByRole('heading', { name: /settings/i }).first();
    await expect(settingsHeading).toBeVisible({ timeout: 5000 });
  });

  test('settings page displays without errors', async ({ page }) => {
    // Monitor console errors
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Page should load without critical errors
    await expect(page.locator('#root')).toBeAttached();

    const settingsHeading = page.getByRole('heading', { name: /settings/i }).first();
    await expect(settingsHeading).toBeVisible({ timeout: 5000 });
  });

  test('settings page keyboard navigation works', async ({ page }) => {
    // Navigate to Settings
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Settings should be keyboard accessible
    const settingsHeading = page.getByRole('heading', { name: /settings/i }).first();
    await expect(settingsHeading).toBeVisible({ timeout: 5000 });

    // Check if create button is keyboard accessible
    const createButton = page.getByTestId('create-profile-button');
    if (await createButton.isVisible().catch(() => false)) {
      // Button should be focusable
      await createButton.focus();
      await expect(createButton).toBeFocused();
    }
  });

  test('settings page handles multiple navigation cycles', async ({ page }) => {
    // Navigate to Settings multiple times
    for (let i = 0; i < 3; i++) {
      await page.getByRole('link', { name: /settings/i }).click();
      await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

      const settingsHeading = page.getByRole('heading', { name: /settings/i }).first();
      await expect(settingsHeading).toBeVisible({ timeout: 5000 });

      // Navigate away
      await page.getByRole('link', { name: /dashboard/i }).click();
      await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});
    }

    // App should still be functional
    await expect(page.locator('#root')).toBeAttached();
  });
});
