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
  }

  /** Wait for the courses table to render. */
  async waitForCourseList() {
    await this.page.waitForSelector('[data-testid="courses-panel"]', {
      state: 'visible',
      timeout: 15000
    });
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

    await this.page.getByTestId('course-code-input').fill(data.code);
    await this.page.getByTestId('course-title-input').fill(data.title);

    if (data.department) {
      await this.page.getByLabel(/department/i).fill(data.department);
    }
    if (data.semester) {
      await this.page.getByLabel(/semester/i).fill(data.semester);
    }
    if (data.owner) {
      await this.page.getByLabel(/owner/i).fill(data.owner);
    }
    if (data.budget) {
      await this.page.getByLabel(/budget/i).fill(data.budget);
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
    await this.page.getByTestId('enroll-email-input').fill(email);
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
    await this.page.getByTestId('add-template-input').fill(slug);
    await this.page.getByTestId('add-template-button').click();
    await this.page.waitForTimeout(300);
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
    await this.page.getByTestId('distribute-amount-input').fill(String(amount));
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
}
