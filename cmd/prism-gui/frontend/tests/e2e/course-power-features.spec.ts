/**
 * Course Power Features E2E Tests — v0.19.0
 *
 * Tests for:
 *   #46  Template enforcement badge
 *   #47  Template whitelisting (GUI enforcement indicator)
 *   #48  TA Access management — grant, revoke, connect
 *   #49  Workspace reset
 *   #160 TA SSH access log (audit trail)
 *   #164 Shared materials creation and mount
 *   #167 Shared materials volume display
 *
 * Requires: prismd running with PRISM_TEST_MODE=true
 */

import { test, expect, request } from '@playwright/test';
import { CoursePage } from './pages';

const BASE_URL = 'http://localhost:8947';

// ── Helpers ───────────────────────────────────────────────────────────────────

function uniquePowerCode(prefix: string): string {
  // Keep under 20-char backend limit: prefix (≤8) + hyphen(1) + 5 digits = ≤14 chars
  return `${prefix}-${Date.now() % 100000}`;
}

async function createCourseViaAPI(code: string): Promise<{ id: string; code: string }> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  const resp = await apiContext.post('/api/v1/courses', {
    data: {
      code,
      title: `E2E Power Test ${code}`,
      semester: 'E2E-2099',
      semester_start: '2099-09-01T00:00:00Z',
      semester_end: '2099-12-15T00:00:00Z',
      owner: 'e2e-instructor',
    }
  });
  const body = await resp.json();
  await apiContext.dispose();
  return { id: body.id || '', code };
}

async function deleteCourseViaAPI(courseId: string): Promise<void> {
  if (!courseId) return;
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  await apiContext.delete(`/api/v1/courses/${courseId}`);
  await apiContext.dispose();
}

async function grantTAViaAPI(courseId: string, email: string): Promise<void> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  await apiContext.post(`/api/v1/courses/${courseId}/ta-access`, {
    data: { email, display_name: 'E2E TA' }
  });
  await apiContext.dispose();
}

async function addTemplateViaAPI(courseId: string, templateSlug: string): Promise<void> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  await apiContext.post(`/api/v1/courses/${courseId}/templates`, {
    data: { template: templateSlug }
  });
  await apiContext.dispose();
}

// ── Setup ─────────────────────────────────────────────────────────────────────

test.beforeEach(async ({ context }) => {
  await context.addInitScript(() => {
    localStorage.setItem('prism_onboarding_complete', 'true');
  });
});

// ── Test Suite: TA Access (#48, #160) ────────────────────────────────────────

test.describe('TA Access Management', () => {
  let courseId = '';
  let courseCode = '';

  test.beforeEach(async () => {
    const result = await createCourseViaAPI(uniquePowerCode('PWR-TA'));
    courseId = result.id;
    courseCode = result.code;
  });

  test.afterEach(async () => {
    await deleteCourseViaAPI(courseId);
  });

  test('TA Access tab renders in course detail', async ({ page }) => {
    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);

    await coursePage.switchToTab('TA Access');
    await expect(page.getByTestId('ta-access-table')).toBeVisible({ timeout: 10000 });
  });

  test('shows "No TAs" empty state initially', async ({ page }) => {
    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);
    await coursePage.openTAAccessTab();

    await expect(page.getByText(/no tas/i)).toBeVisible({ timeout: 5000 });
  });

  test('shows grant TA button', async ({ page }) => {
    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);
    await coursePage.openTAAccessTab();

    await expect(page.getByTestId('ta-grant-button')).toBeVisible({ timeout: 5000 });
  });

  test('lists TA granted via API', async ({ page }) => {
    await grantTAViaAPI(courseId, 'ta-e2e@uni.edu');

    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);
    await coursePage.openTAAccessTab();

    await expect(page.getByText('ta-e2e@uni.edu')).toBeVisible({ timeout: 10000 });
  });
});

// ── Test Suite: Template Enforcement (#46, #47) ───────────────────────────────

test.describe('Template Enforcement', () => {
  let courseId = '';
  let courseCode = '';

  test.beforeEach(async () => {
    const result = await createCourseViaAPI(uniquePowerCode('PWR-TPL'));
    courseId = result.id;
    courseCode = result.code;
  });

  test.afterEach(async () => {
    await deleteCourseViaAPI(courseId);
  });

  test('Templates tab shows no enforcement badge when whitelist is empty', async ({ page }) => {
    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);
    await coursePage.switchToTab('Templates');

    await page.waitForTimeout(500);
    await expect(page.getByTestId('enforcement-active-badge')).not.toBeVisible();
  });

  test('Templates tab shows enforcement badge when templates are set', async ({ page }) => {
    await addTemplateViaAPI(courseId, 'python-ml');

    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);
    await coursePage.switchToTab('Templates');

    await expect(page.getByTestId('enforcement-active-badge')).toBeVisible({ timeout: 10000 });
  });
});

// ── Test Suite: Shared Materials (#164, #167) ─────────────────────────────────

test.describe('Shared Course Materials', () => {
  let courseId = '';
  let courseCode = '';

  test.beforeEach(async () => {
    const result = await createCourseViaAPI(uniquePowerCode('PWR-MAT'));
    courseId = result.id;
    courseCode = result.code;
  });

  test.afterEach(async () => {
    await deleteCourseViaAPI(courseId);
  });

  test('Materials tab renders in course detail', async ({ page }) => {
    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);
    await coursePage.openMaterialsTab();

    // Create button visible when no volume exists
    await expect(page.getByTestId('create-materials-button')).toBeVisible({ timeout: 10000 });
  });

  test('creating materials volume via API shows EFS ID in UI', async ({ page }) => {
    // Create via API endpoint directly
    const apiContext = await request.newContext({ baseURL: BASE_URL });
    await apiContext.post(`/api/v1/courses/${courseId}/materials`, {
      data: { size_gb: 30, mount_path: '/mnt/shared' }
    });
    await apiContext.dispose();

    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);
    await coursePage.openMaterialsTab();

    // Should show EFS ID (not the create button)
    await expect(page.getByTestId('materials-efs-id')).toBeVisible({ timeout: 10000 });
  });

  test('mount button visible when materials volume exists', async ({ page }) => {
    const apiContext = await request.newContext({ baseURL: BASE_URL });
    await apiContext.post(`/api/v1/courses/${courseId}/materials`, {
      data: { size_gb: 20, mount_path: '/mnt/data' }
    });
    await apiContext.dispose();

    const coursePage = new CoursePage(page);
    await coursePage.navigateToCourses();
    await coursePage.openCourse(courseCode);
    await coursePage.openMaterialsTab();

    await expect(page.getByTestId('mount-materials-button')).toBeVisible({ timeout: 10000 });
  });
});
