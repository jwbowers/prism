/**
 * Templates Page Object
 *
 * Page object for the Templates tab in Prism GUI.
 * Handles template browsing, filtering, and selection.
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class TemplatesPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to Templates tab
   */
  async navigate() {
    await this.navigateToTab('templates');
    await this.waitForLoadingComplete();
  }

  /**
   * Get all template cards
   */
  getTemplateCards(): Locator {
    // Cloudscape uses Cards component - look for card containers
    return this.page.locator('[data-testid="template-card"], .awsui-cards-card');
  }

  /**
   * Get template card by name
   */
  getTemplateByName(name: string): Locator {
    return this.page.locator(`[data-testid="template-card"]:has-text("${name}")`);
  }

  /**
   * Select a template by name
   */
  async selectTemplate(name: string) {
    const template = this.getTemplateByName(name);
    await template.click();
  }

  /**
   * Search for templates
   */
  async searchTemplates(query: string) {
    const searchInput = this.page.getByPlaceholder(/search.*templates/i);
    await searchInput.fill(query);
    // Wait for the template cards to update after search filter
    await this.page.waitForLoadState('domcontentloaded');
  }

  /**
   * Filter templates by category
   */
  async filterByCategory(category: string) {
    const filterSelect = this.page.getByLabel(/category/i);
    await filterSelect.selectOption(category);
    await this.waitForLoadingComplete();
  }

  /**
   * Get template count
   */
  async getTemplateCount(): Promise<number> {
    return await this.getTemplateCards().count();
  }

  /**
   * Verify template information is displayed
   */
  async verifyTemplateInfo(name: string): Promise<boolean> {
    const template = this.getTemplateByName(name);
    const text = await this.getTextContent(template);
    return text !== null && text.includes(name);
  }

  /**
   * Click "Launch" button on a template
   */
  async clickLaunchOnTemplate(name: string) {
    const template = this.getTemplateByName(name);
    const launchButton = template.getByRole('button', { name: /launch/i });
    await launchButton.click();
  }

  /**
   * Get template details (description, packages, etc.)
   */
  async getTemplateDetails(name: string): Promise<string | null> {
    const template = this.getTemplateByName(name);
    return await this.getTextContent(template);
  }

  /**
   * Verify templates are loaded from daemon
   */
  async verifyTemplatesLoaded(): Promise<boolean> {
    const count = await this.getTemplateCount();
    return count > 0;
  }
}
