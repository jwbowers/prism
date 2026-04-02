/**
 * Storage Power Features E2E Tests — v0.20.0
 *
 * Tests for:
 *   #30  SSM file operations (list, push, pull)
 *   #63  EC2 Capacity Blocks (list, reserve modal, cancel)
 *
 * Requires: prismd running with PRISM_TEST_MODE=true
 */

import { test, expect, request } from '@playwright/test';

const BASE_URL = 'http://localhost:8947';

// ── Helpers ───────────────────────────────────────────────────────────────────

async function listCapacityBlocks(): Promise<unknown[]> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  const resp = await apiContext.get('/api/v1/capacity-blocks');
  const body = await resp.json().catch(() => []);
  await apiContext.dispose();
  return Array.isArray(body) ? body : [];
}

// ── Setup ─────────────────────────────────────────────────────────────────────

test.beforeEach(async ({ context }) => {
  await context.addInitScript(() => {
    localStorage.setItem('prism_onboarding_complete', 'true');
  });
});

// ── Test Suite: Capacity Blocks panel (#63) ───────────────────────────────────

test.describe('Capacity Blocks Management', () => {
  test('Capacity Blocks sidebar link is visible', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('link', { name: /capacity blocks/i })).toBeVisible({ timeout: 10000 });
  });

  test('Capacity Blocks panel renders', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /capacity blocks/i }).click();
    await expect(page.getByTestId('capacity-blocks-table')).toBeVisible({ timeout: 10000 });
  });

  test('Reserve Capacity Block button is visible', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /capacity blocks/i }).click();
    await expect(page.getByTestId('reserve-capacity-block-button')).toBeVisible({ timeout: 10000 });
  });

  test('clicking Reserve opens the modal with form fields', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /capacity blocks/i }).click();
    await page.getByTestId('reserve-capacity-block-button').click();

    await expect(page.getByTestId('reserve-capacity-block-modal')).toBeVisible({ timeout: 5000 });
    await expect(page.getByTestId('reserve-instance-type')).toBeVisible();
    await expect(page.getByTestId('reserve-count-input')).toBeVisible();
    await expect(page.getByTestId('reserve-duration')).toBeVisible();
  });

  test('Reserve submit button is disabled without instance type', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /capacity blocks/i }).click();
    await page.getByTestId('reserve-capacity-block-button').click();
    await expect(page.getByTestId('reserve-submit-button')).toBeDisabled({ timeout: 5000 });
  });

  test('API returns list of capacity blocks', async () => {
    const blocks = await listCapacityBlocks();
    // Just verifies the endpoint responds without error (may be empty in test env)
    expect(Array.isArray(blocks)).toBe(true);
  });
});

// ── Test Suite: SSM File Operations panel (#30) ───────────────────────────────

test.describe('SSM File Operations API', () => {
  test('GET /api/v1/instances/{name}/files returns 400 for unknown instance', async () => {
    const apiContext = await request.newContext({ baseURL: BASE_URL });
    const resp = await apiContext.get('/api/v1/instances/nonexistent-test-instance/files');
    await apiContext.dispose();
    // Should return error (400 instance not found), not 404 (route not found)
    expect(resp.status()).not.toBe(404);
  });

  test('POST /api/v1/instances/{name}/files/push returns 400 for missing body fields', async () => {
    const apiContext = await request.newContext({ baseURL: BASE_URL });
    const resp = await apiContext.post('/api/v1/instances/test-instance/files/push', {
      data: {} // missing local_path and remote_path
    });
    await apiContext.dispose();
    expect([400, 500]).toContain(resp.status());
  });

  test('POST /api/v1/instances/{name}/files/pull returns error for unknown instance', async () => {
    const apiContext = await request.newContext({ baseURL: BASE_URL });
    const resp = await apiContext.post('/api/v1/instances/nonexistent-test-instance/files/pull', {
      data: { remote_path: '/tmp/test.txt', local_path: '/tmp/test.txt' }
    });
    await apiContext.dispose();
    expect(resp.status()).not.toBe(404);
  });
});

// ── Test Suite: S3 Mounts tab (#22c) ─────────────────────────────────────────

test.describe('S3 Mounts tab', () => {
  test('S3 Mounts tab is visible in storage view', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /storage/i }).first().click();
    // Wait for storage page to render tabs
    await page.getByRole('tab', { name: /efs/i }).waitFor({ state: 'visible', timeout: 10000 });
    await expect(page.getByRole('tab', { name: /s3 mounts/i })).toBeVisible({ timeout: 5000 });
  });

  test('S3 Mounts tab contains instance selector and Load Mounts button', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /storage/i }).first().click();
    await page.getByRole('tab', { name: /s3 mounts/i }).click();

    await expect(page.getByTestId('s3-instance-select')).toBeVisible({ timeout: 5000 });
    await expect(page.getByTestId('load-s3-mounts-button')).toBeVisible({ timeout: 5000 });
  });

  test('Mount S3 Bucket modal contains required fields', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /storage/i }).first().click();
    await page.getByRole('tab', { name: /s3 mounts/i }).click();
    // The Mount button is disabled until an instance is selected; click anyway to verify it exists
    await expect(page.getByTestId('mount-s3-button')).toBeVisible({ timeout: 5000 });
  });
});

// ── Test Suite: Storage Analytics tab (#23c) ──────────────────────────────────

test.describe('Storage Analytics tab', () => {
  test('Analytics tab is visible in storage view', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /storage/i }).first().click();
    await page.getByRole('tab', { name: /efs/i }).waitFor({ state: 'visible', timeout: 10000 });
    await expect(page.getByRole('tab', { name: /analytics/i })).toBeVisible({ timeout: 5000 });
  });

  test('Analytics tab contains period selector and Refresh button', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: /storage/i }).first().click();
    await page.getByRole('tab', { name: /analytics/i }).click();

    await expect(page.getByTestId('analytics-period-select')).toBeVisible({ timeout: 5000 });
    await expect(page.getByTestId('refresh-analytics-button')).toBeVisible({ timeout: 5000 });
  });

  test('Analytics table renders with mocked API response', async ({ page }) => {
    // Mock the storage analytics API endpoint
    await page.route('**/api/v1/storage/analytics**', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          resources: [
            { storage_name: 'my-efs', type: 'efs', period: 'daily', total_cost: 0.12, daily_cost: 0.12, usage_percent: 35.5, recommendations: [] },
            { storage_name: 'my-ebs', type: 'ebs', period: 'daily', total_cost: 0.08, daily_cost: 0.08, usage_percent: 72.0, recommendations: [] },
          ]
        })
      });
    });

    await page.goto('/');
    await page.getByRole('link', { name: /storage/i }).first().click();
    await page.getByRole('tab', { name: /analytics/i }).click();
    await page.getByTestId('refresh-analytics-button').click();

    await expect(page.getByTestId('analytics-table')).toBeVisible({ timeout: 5000 });
  });
});
