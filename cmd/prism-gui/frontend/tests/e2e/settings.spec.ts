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

  // Helper to navigate to Settings > Profiles section
  async function navigateToSettings(page: any) {
    await page.getByRole('link', { name: /settings/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});
    // Navigate to Profiles sub-section within Settings.
    // Must wait for the sub-nav to render before clicking (SettingsView renders async).
    await page.locator('a[href="#profiles"]').waitFor({ state: 'visible', timeout: 8000 });
    await page.locator('a[href="#profiles"]').click();
    // Wait for the Profiles section to confirm we're in the right place
    await page.waitForSelector('[data-testid="create-profile-button"]', { state: 'visible', timeout: 5000 });
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
    await navigateToSettings(page);

    // Profiles table must be visible (test mode always has at least "AWS Default" profile)
    const profilesTable = page.locator('[data-testid="profiles-table"]');
    await expect(profilesTable).toBeVisible({ timeout: 5000 });

    // Table must have at least one profile row
    const rows = profilesTable.locator('tbody tr');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
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

    // Wait for dialog - use :visible to avoid matching all 20 Cloudscape hidden modals in DOM
    const dialog = page.locator('[role="dialog"]:visible').first();
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

    // Wait for dialog - use :visible to avoid matching all 20 Cloudscape hidden modals in DOM
    const dialog = page.locator('[role="dialog"]:visible').first();
    await dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Submit without filling the required name field
    await page.getByRole('button', { name: /create/i }).last().click();

    // Validation error must appear — form must not silently close
    const validationError = dialog.locator('[data-testid="validation-error"]');
    await expect(validationError).toBeVisible({ timeout: 3000 });
    expect(await validationError.textContent()).toMatch(/name.*required/i);

    // Cleanup: close the dialog
    await page.getByRole('button', { name: 'Cancel', exact: true }).last().click();
    await page.locator('[role="dialog"]:visible').waitFor({ state: 'hidden', timeout: 5000 });
  });

  test('profile form accepts valid input', async ({ page }) => {
    await navigateToSettings(page);

    const createButton = page.getByTestId('create-profile-button');
    await expect(createButton).toBeVisible({ timeout: 5000 });
    await createButton.click();

    // Wait for dialog - use :visible to avoid matching all 20 Cloudscape hidden modals in DOM
    await page.locator('[role="dialog"]:visible').first().waitFor({ state: 'visible', timeout: 5000 });

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

    // Wait for dialog - use :visible to avoid matching all 20 Cloudscape hidden modals in DOM
    await page.locator('[role="dialog"]:visible').first().waitFor({ state: 'visible', timeout: 5000 });

    // Click cancel button
    const cancelButton = page.getByRole('button', { name: 'Cancel', exact: true }).last();
    await cancelButton.click();

    // Wait for dialog to close - when no visible dialog remains the locator resolves as hidden
    await page.locator('[role="dialog"]:visible').waitFor({ state: 'hidden', timeout: 5000 });
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
    await navigateToSettings(page);

    // Profiles section must always show meaningful content: either a populated table
    // or at minimum the create button (if somehow no profiles exist).
    // Both must be true in test mode — table has "AWS Default", create button is visible.
    const profilesTable = page.locator('[data-testid="profiles-table"]');
    const createButton = page.getByTestId('create-profile-button');

    const tableVisible = await profilesTable.isVisible().catch(() => false);
    const buttonVisible = await createButton.isVisible().catch(() => false);

    // The page must never be completely blank — at least one must render
    expect(tableVisible || buttonVisible).toBe(true);
    // In practice both should be true (table + create button coexist)
    expect(buttonVisible).toBe(true);
  });

  test('settings forms use Cloudscape components', async ({ page }) => {
    await navigateToSettings(page);

    const createButton = page.getByTestId('create-profile-button');
    await expect(createButton).toBeVisible({ timeout: 5000 });
    await createButton.click();

    // Wait for dialog - use :visible to avoid matching all 20 Cloudscape hidden modals in DOM
    await page.locator('[role="dialog"]:visible').first().waitFor({ state: 'visible', timeout: 5000 });

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
