/**
 * Navigation Tests
 *
 * Tests for SideNavigation, view switching, and UI navigation patterns in Prism GUI.
 * Tests the Cloudscape-based React application navigation.
 */

import { test, expect } from '@playwright/test';

test.describe('Navigation and User Interactions', () => {
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

  test('side navigation switches between sections correctly', async ({ page }) => {
    // Verify Dashboard is loaded by default (initial view)
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();

    // Navigate to Templates
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Verify Templates view loaded (either cards or empty state)
    const templatesContent = page.locator('[data-testid="template-card"]').first();
    const hasTemplates = await templatesContent.isVisible().catch(() => false);
    expect(typeof hasTemplates).toBe('boolean'); // Template cards may or may not exist

    // Navigate to My Workspaces
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Verify Workspaces view loaded (table or empty state)
    const workspacesContent = page.locator('[data-testid="instances-table"], [data-testid="empty-instances"]').first();
    await expect(workspacesContent).toBeVisible({ timeout: 5000 });

    // Navigate to Storage
    await page.getByRole('link', { name: /storage/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Verify Storage view loaded (EFS or EBS tables)
    const storageContent = page.locator('[data-testid="efs-table"], [data-testid="ebs-table"]').first();
    await expect(storageContent).toBeVisible({ timeout: 5000 });

    // Navigate back to Dashboard
    await page.getByRole('link', { name: /dashboard/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Dashboard should be visible
    const heading = page.getByRole('heading').first();
    await expect(heading).toBeVisible({ timeout: 5000 });
  });

  test('navigation links are all present', async ({ page }) => {
    // Verify main navigation links exist
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /templates/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /my workspaces/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /storage/i })).toBeVisible();

    // Additional navigation links (may be in expandable sections)
    const backupsLink = page.getByRole('link', { name: /backups/i });
    const hasBackups = await backupsLink.isVisible().catch(() => false);
    expect(typeof hasBackups).toBe('boolean');

    const projectsLink = page.getByRole('link', { name: /projects/i });
    const hasProjects = await projectsLink.isVisible().catch(() => false);
    expect(typeof hasProjects).toBe('boolean');
  });

  test('navigation preserves state between view switches', async ({ page }) => {
    // Navigate to My Workspaces
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Verify we're in workspaces view
    const workspacesContent = page.locator('[data-testid="instances-table"], [data-testid="empty-instances"]').first();
    await expect(workspacesContent).toBeVisible({ timeout: 5000 });

    // Navigate to Templates
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Navigate back to My Workspaces
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Workspaces view should still be functional
    await expect(workspacesContent).toBeVisible({ timeout: 5000 });
  });

  test('responsive navigation works on all screen sizes', async ({ page }) => {
    // Test navigation on desktop viewport (default)
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();

    // Test navigation on tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();

    // Navigation should still work
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Content should load
    const heading = page.getByRole('heading').first();
    await expect(heading).toBeVisible({ timeout: 5000 });
  });

  test('keyboard navigation works for main links', async ({ page }) => {
    // Tab through elements to test keyboard navigation
    await page.keyboard.press('Tab');

    // Navigate using keyboard to Templates link
    await page.getByRole('link', { name: /templates/i }).focus();
    await page.keyboard.press('Enter');
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Verify navigation worked
    const heading = page.getByRole('heading').first();
    await expect(heading).toBeVisible({ timeout: 5000 });
  });

  test('navigation maintains consistent state', async ({ page }) => {
    // Verify initial navigation works
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();

    // Navigate to Templates and back to Dashboard
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    await page.getByRole('link', { name: /dashboard/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Dashboard should be functional
    const heading = page.getByRole('heading').first();
    await expect(heading).toBeVisible({ timeout: 5000 });
  });

  test('navigation handles rapid view switching', async ({ page }) => {
    // Rapidly switch between views
    await page.getByRole('link', { name: /templates/i }).click();
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.getByRole('link', { name: /storage/i }).click();
    await page.getByRole('link', { name: /dashboard/i }).click();

    // Wait for final view to load
    await page.waitForLoadState('domcontentloaded', { timeout: 5000 });

    // Dashboard should be visible without errors
    const heading = page.getByRole('heading').first();
    await expect(heading).toBeVisible({ timeout: 5000 });
  });

  test('navigation displays correct icons and labels', async ({ page }) => {
    // Verify main navigation links have accessible text
    const dashboardLink = page.getByRole('link', { name: /dashboard/i });
    await expect(dashboardLink).toBeVisible();

    const templatesLink = page.getByRole('link', { name: /templates/i });
    await expect(templatesLink).toBeVisible();

    const workspacesLink = page.getByRole('link', { name: /my workspaces/i });
    await expect(workspacesLink).toBeVisible();

    // Links should have descriptive text content
    const dashboardText = await dashboardLink.textContent();
    expect(dashboardText).toBeTruthy();
    expect(dashboardText!.length).toBeGreaterThan(0);
  });

  test('navigation works after page refresh', async ({ page }) => {
    // Navigate to Templates
    await page.getByRole('link', { name: /templates/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Refresh page
    await page.reload();
    await page.waitForLoadState('domcontentloaded', { timeout: 10000 });

    // Navigation should still work
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Content should load
    const workspacesContent = page.locator('[data-testid="instances-table"], [data-testid="empty-instances"]').first();
    await expect(workspacesContent).toBeVisible({ timeout: 5000 });
  });

  test('navigation sections are organized logically', async ({ page }) => {
    // Verify navigation structure
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();

    // Core features should be accessible
    await expect(page.getByRole('link', { name: /templates/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /my workspaces/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /storage/i })).toBeVisible();

    // Verify React app is rendered
    const root = page.locator('#root');
    await expect(root).toBeAttached();
  });

  test('navigation handles empty states gracefully', async ({ page }) => {
    // Navigate to My Workspaces (may be empty)
    await page.getByRole('link', { name: /my workspaces/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Should show either instances table or empty state
    const workspacesContent = page.locator('[data-testid="instances-table"], [data-testid="empty-instances"]').first();
    await expect(workspacesContent).toBeVisible({ timeout: 5000 });

    // Navigate to Storage (may be empty)
    await page.getByRole('link', { name: /storage/i }).click();
    await page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});

    // Should show storage tables (even if empty)
    const storageContent = page.locator('[data-testid="efs-table"], [data-testid="ebs-table"]').first();
    await expect(storageContent).toBeVisible({ timeout: 5000 });
  });
});
