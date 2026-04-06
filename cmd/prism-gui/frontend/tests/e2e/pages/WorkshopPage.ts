/**
 * WorkshopPage Page Object — v0.18.0
 *
 * Playwright page object for Workshop & Event Management in Prism GUI.
 */

import { Page } from '@playwright/test';
import { BasePage } from './BasePage';

export class WorkshopPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /** Navigate to the Workshops view via the sidebar. */
  async navigateToWorkshops() {
    await this.goto();
    const workshopsLink = this.page.getByRole('button', { name: /^Workshops$/i });
    await workshopsLink.waitFor({ state: 'visible', timeout: 10000 });
    await workshopsLink.click();
    await this.waitForWorkshopList();
  }

  /** Wait for the workshops panel to render. */
  async waitForWorkshopList() {
    await this.page.waitForSelector('[data-testid="workshops-table"]', {
      state: 'visible',
      timeout: 15000
    });
  }

  /**
   * Create a workshop via the Create Workshop button and modal.
   */
  async createWorkshop(data: {
    title: string;
    owner?: string;
    template?: string;
    start?: string;
    end?: string;
    maxParticipants?: string;
    budget?: string;
  }) {
    const createBtn = this.page.getByTestId('create-workshop-button');
    await createBtn.waitFor({ state: 'visible', timeout: 10000 });
    await createBtn.click();

    await this.page.waitForSelector('[data-testid="create-workshop-modal"]', {
      state: 'visible',
      timeout: 10000
    });

    await this.page.getByTestId('workshop-title-input').locator('input').fill(data.title);

    if (data.owner) {
      await this.page.getByLabel(/^owner/i).fill(data.owner);
    }
    if (data.template) {
      await this.page.getByLabel(/^template/i).fill(data.template);
    }
    if (data.start) {
      await this.page.getByLabel(/start.*time/i).fill(data.start);
    }
    if (data.end) {
      await this.page.getByLabel(/end.*time/i).fill(data.end);
    }
    if (data.maxParticipants) {
      await this.page.getByLabel(/max.*participants/i).fill(data.maxParticipants);
    }
    if (data.budget) {
      await this.page.getByLabel(/budget.*participant/i).fill(data.budget);
    }

    // Submit
    await this.page.getByTestId('create-workshop-submit').click();

    // Wait for modal to close
    await this.page.waitForSelector('[data-testid="create-workshop-modal"]', {
      state: 'hidden',
      timeout: 10000
    });
  }

  /**
   * Click on a workshop title to open the dashboard.
   */
  async openWorkshopDashboard(title: string) {
    const titleLink = this.page.getByText(title);
    await titleLink.waitFor({ state: 'visible', timeout: 10000 });
    await titleLink.click();
    // Wait for the dashboard tab panel to appear
    await this.page.waitForSelector('[data-testid="participants-table"]', {
      state: 'visible',
      timeout: 10000
    }).catch(() => {
      // Dashboard may show placeholder if no participants
    });
  }

  /**
   * Navigate to the Dashboard tab.
   */
  async switchToDashboardTab() {
    const dashTab = this.page.getByRole('tab', { name: /dashboard/i });
    await dashTab.waitFor({ state: 'visible', timeout: 5000 });
    await dashTab.click();
  }

  /**
   * Navigate to the Config Templates tab.
   */
  async switchToConfigTab() {
    const configTab = this.page.getByRole('tab', { name: /config/i });
    await configTab.waitFor({ state: 'visible', timeout: 5000 });
    await configTab.click();
    await this.page.waitForSelector('[data-testid="workshop-configs-table"]', {
      state: 'visible',
      timeout: 10000
    });
  }

  /**
   * Click the Provision button for a given workshop row.
   */
  async provisionWorkshop(title: string) {
    const row = this.page.locator('[data-testid="workshops-table"] tr').filter({ hasText: title });
    await row.getByRole('button', { name: /^provision$/i }).click();
  }

  /**
   * Click the End button for a given workshop row (opens confirmation modal).
   */
  async clickEndWorkshop(title: string) {
    const row = this.page.locator('[data-testid="workshops-table"] tr').filter({ hasText: title });
    await row.getByRole('button', { name: /^end$/i }).click();
  }

  /**
   * Confirm the End Workshop modal.
   */
  async confirmEndWorkshop() {
    const modal = this.page.getByRole('dialog');
    await modal.waitFor({ state: 'visible', timeout: 5000 });
    await modal.getByRole('button', { name: /^end workshop$/i }).click();
  }

  /**
   * Click the Delete button for a given workshop row.
   */
  async deleteWorkshop(title: string) {
    const row = this.page.locator('[data-testid="workshops-table"] tr').filter({ hasText: title });
    await row.getByRole('button', { name: /^delete$/i }).click();
  }
}
