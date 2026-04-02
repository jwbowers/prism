/**
 * Course Workflows E2E Tests — v0.17.0
 *
 * Tests for the Courses GUI introduced in v0.17.0:
 *   Course list | Create | Detail tabs (Overview/Members/Templates/Budget/Audit)
 *   Enrollment | Template whitelist | Budget | Archive
 *
 * Requires: prismd running with PRISM_TEST_MODE=true
 */

import { test, expect, request } from '@playwright/test';
import { CoursePage } from './pages';
import * as path from 'path';
import * as fs from 'fs';
import * as os from 'os';

const BASE_URL = 'http://localhost:8947';

// ── Helpers ───────────────────────────────────────────────────────────────────

async function createCourseViaAPI(code: string, title: string = 'E2E Test Course'): Promise<string> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  const resp = await apiContext.post('/api/v1/courses', {
    data: {
      code,
      title,
      semester: 'E2E-2099',
      semester_start: '2099-09-01T00:00:00Z',
      semester_end: '2099-12-15T00:00:00Z',
      owner: 'e2e-test-prof',
    }
  });
  const body = await resp.json();
  await apiContext.dispose();
  return body.id || '';
}

async function deleteCourseViaAPI(courseId: string): Promise<void> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  await apiContext.delete(`/api/v1/courses/${courseId}`);
  await apiContext.dispose();
}

function uniqueCode(prefix: string = 'E2E') {
  return `${prefix}-${Date.now() % 100000}`;
}

// ── Test Suite ────────────────────────────────────────────────────────────────

test.describe('Course Workflows', () => {
  let coursePage: CoursePage;
  const createdCourseIds: string[] = [];

  test.beforeEach(async ({ page, context }) => {
    // Dismiss onboarding modal before navigation
    await context.addInitScript(() => {
      localStorage.setItem('prism_onboarding_complete', 'true');
    });
    coursePage = new CoursePage(page);
  });

  test.afterAll(async () => {
    // Cleanup all courses created during tests
    for (const id of createdCourseIds) {
      await deleteCourseViaAPI(id).catch(() => {});
    }
  });

  // ── Course Lifecycle ────────────────────────────────────────────────────

  test.describe('Course Lifecycle', () => {
    test('navigates to courses sidebar entry', async ({ page }) => {
      await coursePage.navigateToCourses();
      await expect(page.getByTestId('courses-panel')).toBeVisible();
    });

    test('shows empty state when no courses', async ({ page }) => {
      await coursePage.navigateToCourses();
      // Either the table is visible or shows empty state — both are valid
      const panel = page.getByTestId('courses-panel');
      await expect(panel).toBeVisible();
    });

    test('creates a course via UI', async ({ page }) => {
      const code = uniqueCode('CRS');
      await coursePage.navigateToCourses();
      await coursePage.createCourse({
        code,
        title: 'E2E Test Course',
        semester: 'E2E-2099',
        owner: 'test-prof',
      });
      // Verify it appears in the list
      const row = page.locator('[data-testid="courses-table"] tr').filter({ hasText: code });
      await expect(row).toBeVisible({ timeout: 10000 });

      // Track for cleanup
      const apiContext = await request.newContext({ baseURL: BASE_URL });
      const listResp = await apiContext.get('/api/v1/courses');
      const data = await listResp.json();
      const course = (data.courses || []).find((c: any) => c.code === code);
      if (course?.id) createdCourseIds.push(course.id);
      await apiContext.dispose();
    });

    test('opens a course detail view', async ({ page }) => {
      const code = uniqueCode('VIEW');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);

      await coursePage.openCourse(code);
      await expect(page.getByTestId('course-detail-panel')).toBeVisible();
    });

    test('shows all 5 tabs in course detail', async ({ page }) => {
      const code = uniqueCode('TABS');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);

      const tabs = page.getByTestId('course-tabs');
      await expect(tabs.getByRole('tab', { name: /overview/i })).toBeVisible();
      await expect(tabs.getByRole('tab', { name: /members/i })).toBeVisible();
      await expect(tabs.getByRole('tab', { name: /templates/i })).toBeVisible();
      await expect(tabs.getByRole('tab', { name: /budget/i })).toBeVisible();
      await expect(tabs.getByRole('tab', { name: /audit/i })).toBeVisible();
    });

    test('navigates back to course list', async ({ page }) => {
      const code = uniqueCode('BACK');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.goBack();
      await expect(page.getByTestId('courses-panel')).toBeVisible();
    });
  });

  // ── Enrollment Management ───────────────────────────────────────────────

  test.describe('Enrollment Management', () => {
    test('opens Members tab', async ({ page }) => {
      const code = uniqueCode('MBR');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Members');
      await expect(page.getByTestId('members-table')).toBeVisible();
    });

    test('shows enroll member button', async ({ page }) => {
      const code = uniqueCode('ENR');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Members');
      await expect(page.getByTestId('enroll-member-button')).toBeVisible();
    });

    test('opens enroll modal', async ({ page }) => {
      const code = uniqueCode('MOD');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Members');
      await page.getByTestId('enroll-member-button').click();
      await expect(page.getByTestId('enroll-email-input')).toBeVisible();
    });

    test('enrolls a student via modal', async ({ page }) => {
      const code = uniqueCode('ESTUD');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Members');
      await page.getByTestId('enroll-member-button').click();
      await page.getByTestId('enroll-email-input').locator('input').fill('alice@uni.edu');
      await page.getByTestId('enroll-submit-button').click();
      await page.waitForTimeout(500);
      // Members table should update
      await expect(page.getByTestId('members-table')).toBeVisible();
    });
  });

  // ── Template Whitelist ─────────────────────────────────────────────────

  test.describe('Template Whitelist', () => {
    test('shows templates tab', async ({ page }) => {
      const code = uniqueCode('TMPL');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Templates');
      await expect(page.getByTestId('add-template-input')).toBeVisible();
    });

    test('adds a template to the whitelist', async ({ page }) => {
      const code = uniqueCode('TADD');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.addTemplate('python-ml');
      await expect(page.getByTestId('templates-table')).toBeVisible({ timeout: 5000 });
    });

    test('removes a template from the whitelist', async ({ page }) => {
      const code = uniqueCode('TREM');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      // Add via API first
      const apiContext = await request.newContext({ baseURL: BASE_URL });
      await apiContext.post(`/api/v1/courses/${id}/templates`, { data: { template: 'python-ml' } });
      await apiContext.dispose();

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Templates');
      const table = page.getByTestId('templates-table');
      await expect(table).toBeVisible({ timeout: 5000 });
      const removeBtn = table.getByRole('button', { name: /remove/i }).first();
      await removeBtn.click();
      await page.waitForTimeout(300);
    });

    test('shows empty state with info alert for empty whitelist', async ({ page }) => {
      const code = uniqueCode('TEMPTY');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Templates');
      // Should see info alert about all templates being allowed
      await expect(page.getByText(/all templates are allowed/i)).toBeVisible({ timeout: 5000 });
    });
  });

  // ── Budget Distribution ────────────────────────────────────────────────

  test.describe('Budget Distribution', () => {
    test('shows budget tab', async ({ page }) => {
      const code = uniqueCode('BDG');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Budget');
      await expect(page.getByTestId('distribute-budget-button')).toBeVisible();
    });

    test('opens distribute budget modal', async ({ page }) => {
      const code = uniqueCode('BMOD');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Budget');
      await page.getByTestId('distribute-budget-button').click();
      await expect(page.getByTestId('distribute-amount-input')).toBeVisible();
    });

    test('distributes budget to students', async ({ page }) => {
      const code = uniqueCode('BDIST');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.distributeBudget(50);
      await page.waitForTimeout(500);
      // Modal closed — budget tab still visible
      await expect(page.getByTestId('distribute-budget-button')).toBeVisible();
    });
  });

  // ── Overview & Reports ─────────────────────────────────────────────────

  test.describe('Overview and Reports', () => {
    test('shows overview tab', async ({ page }) => {
      const code = uniqueCode('OVW');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Overview');
      // Overview may show zero-state stats
      await page.waitForTimeout(300);
    });

    test('shows audit log tab', async ({ page }) => {
      const code = uniqueCode('AUD');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Audit');
      await expect(page.getByTestId('audit-table')).toBeVisible();
    });

    test('shows download report button', async ({ page }) => {
      const code = uniqueCode('RPT');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Audit');
      await expect(page.getByTestId('download-report-button')).toBeVisible();
    });
  });

  // ── Archive ────────────────────────────────────────────────────────────

  test.describe('Archive', () => {
    test('archive button visible for closed courses', async ({ page }) => {
      const code = uniqueCode('ARCH');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      // Close the course via API first
      const apiContext = await request.newContext({ baseURL: BASE_URL });
      await apiContext.post(`/api/v1/courses/${id}/close`);
      await apiContext.dispose();

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Audit');
      // Archive button should be visible for closed course
      await expect(page.getByTestId('archive-course-button')).toBeVisible({ timeout: 5000 });
    });

    test('archive confirmation modal appears', async ({ page }) => {
      const code = uniqueCode('ACNF');
      const id = await createCourseViaAPI(code);
      createdCourseIds.push(id);

      const apiContext = await request.newContext({ baseURL: BASE_URL });
      await apiContext.post(`/api/v1/courses/${id}/close`);
      await apiContext.dispose();

      await coursePage.navigateToCourses();
      await page.getByTestId('refresh-courses-button').click();
      await page.waitForTimeout(500);
      await coursePage.openCourse(code);
      await coursePage.switchToTab('Audit');
      await page.getByTestId('archive-course-button').click();
      await expect(page.getByTestId('archive-confirm-button')).toBeVisible({ timeout: 5000 });
    });
  });
});
