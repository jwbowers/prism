import { test, expect } from '@playwright/test';

test.describe('Budget Workflows', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set localStorage BEFORE navigating to prevent onboarding modal
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');
  });

  test('should display budget overview', async ({ page }) => {
    // Navigate to budgets
    await page.click('text=Budgets');

    // Verify budget overview page elements
    await expect(page.locator('text=Budget Overview')).toBeVisible();
    await expect(page.locator('text=Total Budgets')).toBeVisible();
    await expect(page.locator('text=Total Allocated')).toBeVisible();
    await expect(page.locator('text=Total Spent')).toBeVisible();
    await expect(page.locator('text=Active Alerts')).toBeVisible();

    // Verify table is present
    await expect(page.locator('[data-testid="budgets-table"]')).toBeVisible();
  });

  test('should create budget pool', async ({ page }) => {
    // Navigate to budgets
    await page.click('text=Budgets');

    // Wait for page to load
    await page.waitForSelector('[data-testid="budgets-table"]');

    // Open create budget modal (click the first button, which is in the header)
    await page.locator('button:has-text("Create Budget Pool")').first().click();

    // Wait for modal to appear - use visible dialog selector
    const dialog = page.locator('[role="dialog"]:visible').first();
    await dialog.waitFor({ state: 'visible' });
    await page.waitForTimeout(500); // Allow modal animation to complete

    // Fill in budget details using accessible labels - scope to dialog
    const budgetName = `Test Budget Pool ${Date.now()}`;
    await dialog.getByLabel('Budget Name').fill(budgetName);
    await dialog.getByLabel('Description').fill('Test budget for E2E');
    await dialog.getByLabel('Total Amount (USD)').fill('50000');
    await dialog.getByLabel('Alert Threshold (%)').fill('80');

    // Submit form - scope to dialog
    await dialog.getByRole('button', { name: /create budget/i }).click();

    // Verify success notification
    await expect(page.locator('text=Budget Created')).toBeVisible({ timeout: 10000 });
    await expect(page.locator(`text=Budget pool "${budgetName}" created successfully`)).toBeVisible();
  });

  test('should filter budgets by status', async ({ page }) => {
    // Navigate to budgets
    await page.click('text=Budgets');

    // Verify filter dropdown is visible
    const filterSelect = page.locator('[data-testid="budget-filter-select"]');
    await expect(filterSelect).toBeVisible();

    // Test filtering to Warning
    await filterSelect.click();
    await page.click('text=Warning (80-95%)');
    await expect(page.locator('[data-testid="budgets-table"]')).toBeVisible();

    // Test filtering to Critical
    await filterSelect.click();
    await page.click('text=Critical (≥95%)');
    await expect(page.locator('[data-testid="budgets-table"]')).toBeVisible();

    // Test filtering back to All
    await filterSelect.click();
    await page.click('text=All Budgets');
    await expect(page.locator('[data-testid="budgets-table"]')).toBeVisible();
  });

  test('should display budget actions dropdown', async ({ page }) => {
    // Navigate to budgets
    await page.click('text=Budgets');

    // Wait for table to load
    await page.waitForSelector('[data-testid="budgets-table"]');

    // Look for any budget actions dropdown
    const actionsDropdown = page.locator('[data-testid^="budget-actions-"]').first();

    // Only test if budgets exist
    const count = await actionsDropdown.count();
    if (count > 0) {
      await actionsDropdown.click();

      // Verify action menu items
      await expect(page.locator('text=View Summary')).toBeVisible();
      await expect(page.locator('text=Manage Allocations')).toBeVisible();
      await expect(page.locator('text=Spending Report')).toBeVisible();
      await expect(page.locator('text=Edit Budget')).toBeVisible();
      await expect(page.locator('text=Delete').last()).toBeVisible();
    }
  });

  test('should show validation error for empty budget name', async ({ page }) => {
    // Navigate to budgets
    await page.click('text=Budgets');
    await page.waitForSelector('[data-testid="budgets-table"]');

    // Open create budget modal (click the first button, which is in the header)
    await page.locator('button:has-text("Create Budget Pool")').first().click();
    const dialog = page.locator('[role="dialog"]:visible').first();
    await dialog.waitFor({ state: 'visible' });
    await page.waitForTimeout(500); // Allow modal animation to complete

    // Try to submit without filling name - scope to dialog
    await dialog.getByLabel('Total Amount (USD)').fill('50000');
    await dialog.getByRole('button', { name: /create budget/i }).click();

    // Verify validation error
    await expect(page.locator('text=Budget name is required')).toBeVisible();
  });

  test('should show validation error for invalid amount', async ({ page }) => {
    // Navigate to budgets
    await page.click('text=Budgets');
    await page.waitForSelector('[data-testid="budgets-table"]');

    // Open create budget modal (click the first button, which is in the header)
    await page.locator('button:has-text("Create Budget Pool")').first().click();
    const dialog = page.locator('[role="dialog"]:visible').first();
    await dialog.waitFor({ state: 'visible' });
    await page.waitForTimeout(500); // Allow modal animation to complete

    // Fill name but invalid amount - scope to dialog
    await dialog.getByLabel('Budget Name').fill('Test Budget');
    await dialog.getByLabel('Total Amount (USD)').fill('0');
    await dialog.getByRole('button', { name: /create budget/i }).click();

    // Verify validation error
    await expect(page.locator('text=Total amount must be greater than 0')).toBeVisible();
  });

  test('should cancel budget creation', async ({ page }) => {
    // Navigate to budgets
    await page.click('text=Budgets');
    await page.waitForSelector('[data-testid="budgets-table"]');

    // Open create budget modal (click the first button, which is in the header)
    await page.locator('button:has-text("Create Budget Pool")').first().click();
    const dialog = page.locator('[role="dialog"]:visible').first();
    await dialog.waitFor({ state: 'visible' });
    await page.waitForTimeout(500); // Allow modal animation to complete

    // Fill some data - scope to dialog
    await dialog.getByLabel('Budget Name').fill('Test Budget to Cancel');

    // Click cancel
    await dialog.getByRole('button', { name: /cancel/i }).click();

    // Verify modal is closed
    await page.waitForTimeout(500);
    const visibleDialogs = await page.locator('[role="dialog"]:visible').count();
    expect(visibleDialogs).toBe(0);
  });

  test('should refresh budgets data', async ({ page }) => {
    // Navigate to budgets
    await page.click('text=Budgets');

    // Click refresh button
    await page.click('button:has-text("Refresh")');

    // Verify table is still visible after refresh
    await expect(page.locator('[data-testid="budgets-table"]')).toBeVisible();
  });
});
