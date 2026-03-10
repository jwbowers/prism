/**
 * Governance Page Object — v0.13.0
 *
 * Page object for governance interactions in Prism GUI.
 * Extends ProjectsPage to provide governance-specific helper methods.
 */

import { Page, expect } from '@playwright/test';
import { ProjectsPage } from './ProjectsPage';

export class GovernancePage extends ProjectsPage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to the Governance tab of a specific project.
   * Uses the Actions menu → View Details pattern (same as viewProjectDetails),
   * then clicks the Governance tab.
   */
  async navigateToGovernance(projectName: string) {
    await this.navigate();

    // Poll for project to appear in table (handles async UI updates)
    const projectExists = await this.verifyProjectExists(projectName);
    if (!projectExists) {
      throw new Error(`Project "${projectName}" not found in projects table`);
    }

    // Open detail view via Actions → View Details (same pattern as viewProjectDetails)
    const projectRow = this.getProjectByName(projectName);
    const actionsButton = projectRow.getByRole('button', { name: /actions/i });
    await actionsButton.waitFor({ state: 'visible', timeout: 5000 });
    await actionsButton.click();

    const viewDetailsOption = this.page.getByRole('menuitem', { name: /view details/i });
    await viewDetailsOption.waitFor({ state: 'visible', timeout: 5000 });
    await viewDetailsOption.click();

    // Wait for project detail view and tabs to appear
    await this.page.waitForSelector('[data-testid="project-detail-tabs"]', {
      state: 'visible',
      timeout: 15000
    });

    // Click the Governance tab
    const governanceTab = this.page.getByRole('tab', { name: /governance/i });
    await governanceTab.waitFor({ state: 'visible', timeout: 10000 });
    await governanceTab.click();

    // Wait for governance panel to render
    await this.page.waitForSelector('[data-testid="governance-panel"]', {
      state: 'visible',
      timeout: 10000
    });
  }

  /**
   * Switch to a governance sub-tab by label text.
   */
  async switchToGovernanceTab(label: string) {
    const tab = this.page.getByRole('tab', { name: new RegExp(label, 'i') }).last();
    await tab.waitFor({ state: 'visible', timeout: 10000 });
    await tab.click();
    await this.page.waitForTimeout(300);
  }

  /**
   * Add a role quota.
   */
  async addQuota(role: string, maxInstances: number, maxSpendDaily: number) {
    const btn = this.page.getByTestId('set-quota-button');
    await btn.waitFor({ state: 'visible', timeout: 10000 });
    await btn.click();

    // Select role
    const roleSelect = this.page.getByTestId('quota-role-select');
    await roleSelect.click();
    await this.page.locator(`[data-value="${role}"]`).click();

    // Fill max instances (Cloudscape Input wraps native <input> in a <div>)
    await this.page.getByTestId('quota-max-instances-input').locator('input').fill(String(maxInstances));

    // Fill max spend daily
    await this.page.getByTestId('quota-max-spend-daily-input').locator('input').fill(String(maxSpendDaily));

    // Save
    await this.page.getByTestId('save-quota-button').click();

    // Wait for modal to close
    await this.page.waitForSelector('[data-testid="save-quota-button"]', { state: 'hidden', timeout: 5000 }).catch(() => {});
    await this.page.waitForTimeout(500);
  }

  /**
   * Configure a grant period.
   */
  async setGrantPeriod(name: string, startDate: string, endDate: string, autoFreeze: boolean) {
    // Click Configure or Edit button
    const configBtn = this.page.getByTestId('configure-grant-period-button');
    const editBtn = this.page.getByTestId('edit-grant-period-button');

    if (await configBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await configBtn.click();
    } else {
      await editBtn.waitFor({ state: 'visible', timeout: 5000 });
      await editBtn.click();
    }

    // Fill form (Cloudscape Input wraps native <input> in a <div>)
    await this.page.getByTestId('grant-period-name-input').locator('input').fill(name);
    await this.page.getByTestId('grant-period-start-input').locator('input').fill(startDate);
    await this.page.getByTestId('grant-period-end-input').locator('input').fill(endDate);

    const toggle = this.page.getByTestId('grant-auto-freeze-toggle');
    const isChecked = await toggle.isChecked().catch(() => false);
    if (isChecked !== autoFreeze) {
      await toggle.click();
    }

    // Save
    await this.page.getByTestId('save-grant-period-button').click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Delete the grant period.
   */
  async deleteGrantPeriod() {
    await this.page.getByTestId('delete-grant-period-button').click();
    await this.page.getByTestId('confirm-delete-grant-period-button').click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Share budget from this project.
   */
  async shareBudget(toProjectId: string, toMemberId: string, amount: number, reason: string) {
    await this.page.getByTestId('share-budget-button').click();

    // (Cloudscape Input wraps native <input>/<textarea> in a <div>)
    if (toProjectId) {
      await this.page.getByTestId('share-to-project-input').locator('input').fill(toProjectId);
    }
    if (toMemberId) {
      await this.page.getByTestId('share-to-member-input').locator('input').fill(toMemberId);
    }
    await this.page.getByTestId('share-amount-input').locator('input').fill(String(amount));
    if (reason) {
      await this.page.getByTestId('share-reason-input').locator('input').fill(reason);
    }

    await this.page.getByTestId('confirm-share-budget-button').click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Add an onboarding template.
   */
  async addOnboardingTemplate(name: string, description: string) {
    await this.page.getByTestId('add-onboarding-template-button').click();

    // (Cloudscape Input/Textarea wraps native element in a <div>)
    await this.page.getByTestId('onboarding-template-name-input').locator('input').fill(name);
    if (description) {
      await this.page.getByTestId('onboarding-template-description-input').locator('textarea').fill(description);
    }

    await this.page.getByTestId('save-onboarding-template-button').click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Delete an onboarding template by name.
   */
  async deleteOnboardingTemplate(name: string) {
    const deleteBtn = this.page.getByTestId(`delete-onboarding-template-${name}`);
    await deleteBtn.waitFor({ state: 'visible', timeout: 5000 });
    await deleteBtn.click();
    await this.page.waitForTimeout(500);
  }

  /**
   * Generate a monthly report.
   */
  async generateMonthlyReport(month: string, format: string) {
    // Cloudscape Input wraps native <input> in a <div>
    await this.page.getByTestId('report-month-input').locator('input').fill(month);

    const formatSelect = this.page.getByTestId('report-format-select');
    await formatSelect.click();
    await this.page.locator(`[data-value="${format}"]`).click();

    await this.page.getByTestId('generate-report-button').click();
  }

  /**
   * Wait for the monthly report output to appear.
   */
  async waitForReportOutput() {
    await this.page.waitForSelector('[data-testid="monthly-report-output"]', {
      state: 'visible',
      timeout: 30000
    });
    return this.page.getByTestId('monthly-report-output').textContent();
  }

  /**
   * Navigate to the Approvals view via the sidebar.
   */
  async navigateToApprovals() {
    await this.navigateToTab('approvals');
    await this.page.waitForSelector('[data-testid="approvals-view"]', {
      state: 'visible',
      timeout: 10000
    });
  }
}
