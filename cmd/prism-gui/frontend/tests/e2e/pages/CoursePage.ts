/**
 * CoursePage Page Object — v0.17.0
 *
 * Playwright page object for Courses management in Prism GUI.
 */

import { Page } from '@playwright/test';
import { BasePage } from './BasePage';

export class CoursePage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /** Navigate to the Courses view via the sidebar. */
  async navigateToCourses() {
    await this.goto();
    const coursesLink = this.page.getByRole('link', { name: /^Courses$/i });
    await coursesLink.waitFor({ state: 'visible', timeout: 10000 });
    await coursesLink.click();
    await this.waitForCourseList();
    // Explicitly refresh the list so newly-created courses appear, then wait for the API response.
    const refreshBtn = this.page.getByTestId('refresh-courses-button');
    await refreshBtn.waitFor({ state: 'visible', timeout: 5000 });
    const courseResponsePromise = this.page.waitForResponse(
      response => response.url().includes('/api/v1/courses') && response.status() === 200,
      { timeout: 10000 }
    ).catch(() => {});
    await refreshBtn.click();
    await courseResponsePromise;
    await this.page.waitForTimeout(300);
  }

  /** Wait for the courses table to render and finish loading. */
  async waitForCourseList() {
    await this.page.waitForSelector('[data-testid="courses-panel"]', {
      state: 'visible',
      timeout: 15000
    });
    // Explicitly refresh courses to ensure the table reflects the latest daemon state.
    // Without this, openCourse() can race against the initial loadCourses() API call
    // and see an empty list when a course was just created in beforeEach.
    const refreshBtn = this.page.getByTestId('refresh-courses-button');
    if (await refreshBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      // Set up response waiter BEFORE clicking to capture the refresh response
      const responsePromise = this.page.waitForResponse(
        resp => resp.url().includes('/api/v1/courses') && resp.status() === 200,
        { timeout: 8000 }
      ).catch(() => {});
      await refreshBtn.click();
      await responsePromise;
    }
  }

  /**
   * Create a course via the Create Course button and modal.
   */
  async createCourse(data: {
    code: string;
    title: string;
    department?: string;
    semester?: string;
    start?: string;
    end?: string;
    owner?: string;
    budget?: string;
  }) {
    const createBtn = this.page.getByTestId('create-course-button');
    await createBtn.waitFor({ state: 'visible', timeout: 10000 });
    await createBtn.click();

    await this.page.waitForSelector('[data-testid="create-course-modal"]', {
      state: 'visible',
      timeout: 10000
    });

    await this.page.getByTestId('course-code-input').locator('input').fill(data.code);
    await this.page.getByTestId('course-title-input').locator('input').fill(data.title);

    if (data.department) {
      await this.page.getByLabel('Department', { exact: true }).fill(data.department);
    }
    if (data.semester) {
      await this.page.getByLabel('Semester', { exact: true }).fill(data.semester);
    }
    // Semester Start and End are required by backend — use keyboard type to trigger React state
    const startInput = this.page.getByLabel('Semester Start', { exact: true });
    await startInput.click();
    await startInput.press('Control+a');
    await this.page.keyboard.type(data.start || '2099-09-01');
    await this.page.keyboard.press('Tab');
    const endInput = this.page.getByLabel('Semester End', { exact: true });
    await endInput.click();
    await endInput.press('Control+a');
    await this.page.keyboard.type(data.end || '2099-12-15');
    await this.page.keyboard.press('Tab');
    if (data.owner) {
      await this.page.getByLabel('Owner (User ID)', { exact: true }).fill(data.owner);
    }
    if (data.budget) {
      await this.page.getByLabel('Per-Student Budget (USD)', { exact: true }).fill(data.budget);
    }

    await this.page.getByTestId('create-course-submit').click();
    // Modal closes on success
    await this.page.waitForSelector('[data-testid="create-course-modal"]', {
      state: 'hidden',
      timeout: 10000
    });
    await this.waitForCourseList();
  }

  /** Click on a course row to open its detail view. */
  async openCourse(code: string) {
    const row = this.page.locator('[data-testid="courses-table"] tr').filter({ hasText: code });
    await row.waitFor({ state: 'visible', timeout: 10000 });
    await row.click();
    await this.page.waitForSelector('[data-testid="course-detail-panel"]', {
      state: 'visible',
      timeout: 10000
    });
  }

  /** Switch to a named tab in the course detail panel. */
  async switchToTab(label: string) {
    const tab = this.page.getByTestId('course-tabs').getByRole('tab', { name: new RegExp(label, 'i') });
    await tab.waitFor({ state: 'visible', timeout: 10000 });
    await tab.click();
    await this.page.waitForTimeout(300);
  }

  // ── Members tab helpers ────────────────────────────────────────────────

  async enrollMember(email: string, role: string = 'student') {
    await this.switchToTab('Members');
    await this.page.getByTestId('enroll-member-button').click();
    await this.page.waitForTimeout(300);
    await this.page.getByTestId('enroll-email-input').locator('input').fill(email);
    const roleSelect = this.page.getByLabel(/role/i).last();
    await roleSelect.selectOption(role);
    await this.page.getByTestId('enroll-submit-button').click();
    await this.page.waitForTimeout(500);
  }

  async importRoster(filePath: string, format: string = 'prism') {
    await this.switchToTab('Members');
    const input = this.page.locator('input[type="file"]');
    await input.setInputFiles(filePath);
    await this.page.waitForTimeout(500);
  }

  async unenrollMember(userId: string) {
    await this.switchToTab('Members');
    const row = this.page.locator('[data-testid="members-table"] tr').filter({ hasText: userId });
    await row.getByRole('button', { name: /unenroll/i }).click();
    await this.page.waitForTimeout(300);
  }

  // ── Templates tab helpers ─────────────────────────────────────────────

  async addTemplate(slug: string) {
    await this.switchToTab('Templates');
    // Wait for templates content to load before interacting with input
    const addBtn = this.page.getByTestId('add-template-button');
    await addBtn.waitFor({ state: 'visible', timeout: 5000 });
    // Use pressSequentially with delay to ensure each keypress triggers Cloudscape onChange
    const templateInput = this.page.getByTestId('add-template-input').locator('input');
    await templateInput.click();
    await templateInput.pressSequentially(slug, { delay: 50 });
    await this.page.waitForTimeout(300); // allow React to batch-process state updates
    await addBtn.click();
    await this.page.waitForTimeout(1500);
  }

  async removeTemplate(slug: string) {
    await this.switchToTab('Templates');
    const row = this.page.locator('[data-testid="templates-table"] tr').filter({ hasText: slug });
    await row.getByRole('button', { name: /remove/i }).click();
    await this.page.waitForTimeout(300);
  }

  // ── Budget tab helpers ────────────────────────────────────────────────

  async distributeBudget(amount: number) {
    await this.switchToTab('Budget');
    await this.page.getByTestId('distribute-budget-button').click();
    await this.page.waitForTimeout(200);
    await this.page.getByTestId('distribute-amount-input').locator('input').fill(String(amount));
    await this.page.getByTestId('distribute-submit-button').click();
    await this.page.waitForTimeout(300);
  }

  // ── Audit & Reports tab helpers ───────────────────────────────────────

  async viewAuditLog() {
    await this.switchToTab('Audit');
    await this.page.waitForSelector('[data-testid="audit-table"]', { state: 'visible', timeout: 10000 });
  }

  async downloadReport() {
    await this.switchToTab('Audit');
    const downloadPromise = this.page.waitForEvent('download');
    await this.page.getByTestId('download-report-button').click();
    return downloadPromise;
  }

  async archiveCourse(courseCode: string) {
    await this.switchToTab('Audit');
    await this.page.getByTestId('archive-course-button').click();
    await this.page.waitForSelector('[data-testid="archive-confirm-button"]', { state: 'visible', timeout: 5000 });
    await this.page.getByTestId('archive-confirm-button').click();
    await this.page.waitForTimeout(500);
  }

  // ── Assertions ────────────────────────────────────────────────────────

  async getCourseStatus(code: string): Promise<string> {
    const row = this.page.locator('[data-testid="courses-table"] tr').filter({ hasText: code });
    await row.waitFor({ state: 'visible', timeout: 10000 });
    const badge = row.locator('.status-badge, [class*="badge"]').first();
    return badge.textContent() || '';
  }

  async getStudentBudgetStatus(userId: string): Promise<string> {
    const row = this.page.locator('[data-testid="budget-students-table"] tr').filter({ hasText: userId });
    await row.waitFor({ state: 'visible', timeout: 10000 });
    return row.textContent() || '';
  }

  /** Navigate back to the course list. */
  async goBack() {
    await this.page.getByRole('button', { name: /back to courses/i }).click();
    await this.waitForCourseList();
  }

  // ── v0.19.0: TA Access ───────────────────────────────────────────────────

  /** Switch to the TA Access tab in a course detail view. */
  async openTAAccessTab() {
    await this.switchToTab('TA Access');
    await this.page.waitForSelector('[data-testid="ta-access-table"]', { state: 'visible', timeout: 10000 });
  }

  /** Grant TA access to a user. */
  async grantTAAccess(email: string, displayName?: string) {
    await this.page.getByTestId('ta-grant-button').click();
    await this.page.waitForSelector('[data-testid="ta-grant-modal"]', { state: 'visible', timeout: 5000 });
    await this.page.getByTestId('ta-email-input').locator('input').fill(email);
    if (displayName) {
      await this.page.getByTestId('ta-name-input').locator('input').fill(displayName);
    }
    await this.page.getByTestId('ta-grant-submit').click();
    await this.page.waitForSelector('[data-testid="ta-grant-modal"]', { state: 'hidden', timeout: 5000 });
  }

  /** Click the connect button for a TA row (opens SSH command modal). */
  async connectTA(email: string) {
    const row = this.page.locator('[data-testid="ta-access-table"] tr').filter({ hasText: email });
    await row.getByRole('button', { name: /connect/i }).click();
    await this.page.waitForSelector('[data-testid="ta-connect-modal"]', { state: 'visible', timeout: 5000 });
  }

  // ── v0.19.0: Shared Materials ────────────────────────────────────────────

  /** Switch to the Materials tab in a course detail view. */
  async openMaterialsTab() {
    await this.switchToTab('Materials');
  }

  /** Create a shared materials volume. */
  async createMaterials(sizeGB: number, mountPath?: string) {
    await this.page.getByTestId('create-materials-button').click();
    await this.page.waitForSelector('[data-testid="create-materials-modal"]', { state: 'visible', timeout: 5000 });
    await this.page.getByTestId('materials-size-input').locator('input').fill(String(sizeGB));
    if (mountPath) {
      await this.page.getByTestId('materials-mount-input').locator('input').fill(mountPath);
    }
    await this.page.getByTestId('create-materials-submit').click();
    await this.page.waitForSelector('[data-testid="create-materials-modal"]', { state: 'hidden', timeout: 5000 });
  }

  /** Mount the shared materials volume on all student instances. */
  async mountMaterials() {
    await this.page.getByTestId('mount-materials-button').click();
    await this.page.waitForTimeout(500);
  }

  // ── v0.19.0: Template enforcement ────────────────────────────────────────

  /** Returns true if the "Enforcement Active" badge is visible on the Templates tab. */
  async isTemplateEnforcementActive(): Promise<boolean> {
    await this.switchToTab('Templates');
    const badge = this.page.getByTestId('enforcement-active-badge');
    return badge.isVisible();
  }
}
