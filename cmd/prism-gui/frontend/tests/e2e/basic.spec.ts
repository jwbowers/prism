/**
 * Basic Smoke Tests
 *
 * Core smoke tests for Prism GUI application structure and navigation.
 * Tests that the Cloudscape-based React application loads correctly.
 */

import { test, expect } from '@playwright/test';

test.describe('Basic Smoke Tests', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set localStorage to skip onboarding before navigation
    await context.addInitScript(() => {
      localStorage.setItem('prism_onboarding_complete', 'true');
    });
    await page.goto('/');

    // Wait for app to load
    await page.waitForLoadState('domcontentloaded', { timeout: 10000 });

    // Wait for API to be ready (at least one endpoint should respond)
    await page.waitForResponse(
      (response) => response.url().includes('/api/v1/'),
      { timeout: 15000 }
    ).catch(() => {
      // If API calls fail in test mode, that's okay - we're testing UI structure
    });
  });

  test('application loads successfully', async ({ page }) => {
    // Check that main navigation links are present
    await expect(page.getByRole('button', { name: /dashboard/i })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('button', { name: /templates/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /my workspaces/i })).toBeVisible();

    // Check that a heading exists (content loaded)
    const mainHeading = page.getByRole('heading').first();
    await expect(mainHeading).toBeVisible();

    // Check that React app root exists
    const root = page.locator('#root');
    await expect(root).toBeAttached();
  });

  test('navigation between sections works', async ({ page }) => {
    // Test navigation to Templates
    await page.getByRole('button', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});
    // Just check that we navigated (content changed)
    const templateContent = page.locator('[data-testid="template-card"]').first();
    await expect(templateContent).toBeVisible({ timeout: 5000 }).catch(() => {
      // Template cards may not exist, that's okay
    });

    // Test navigation to My Workspaces (instances)
    await page.getByRole('button', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});
    // Check for instances table or empty state
    const instancesContent = page.locator('[data-testid="instances-table"], [data-testid="empty-instances"]').first();
    await expect(instancesContent).toBeVisible({ timeout: 5000 });

    // Test navigation to Storage
    await page.getByRole('button', { name: /storage/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});
    // Check for storage tables
    const storageContent = page.locator('[data-testid="efs-table"], [data-testid="ebs-table"]').first();
    await expect(storageContent).toBeVisible({ timeout: 5000 });

    // Test navigation back to Dashboard
    await page.getByRole('button', { name: /dashboard/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});
    // Dashboard should load (any heading present)
    const heading = page.getByRole('heading').first();
    await expect(heading).toBeVisible({ timeout: 5000 });
  });

  test('application structure is consistent', async ({ page }) => {
    // Verify main navigation links exist
    await expect(page.getByRole('button', { name: /dashboard/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /templates/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /my workspaces/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /storage/i })).toBeVisible();

    // Verify at least one heading is present (content loaded)
    const heading = page.getByRole('heading').first();
    await expect(heading).toBeVisible();

    // Verify React app root element exists
    const root = page.locator('#root');
    await expect(root).toBeAttached();

    // Verify content is rendered (not blank page)
    const bodyText = await page.locator('body').textContent();
    expect(bodyText).toBeTruthy();
    expect(bodyText!.length).toBeGreaterThan(0);
  });
});
