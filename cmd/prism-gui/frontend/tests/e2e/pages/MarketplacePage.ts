/**
 * Marketplace Page Object
 *
 * Page object for the Template Marketplace section in Prism GUI
 * (Settings > Advanced > Template Marketplace).
 */

import { Page, Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class MarketplacePage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Navigate to Template Marketplace via Settings > Advanced > Template Marketplace
   */
  async navigate() {
    await this.navigateToSettingsAdvanced('Template Marketplace');
    await this.waitForMarketplaceView();
  }

  /**
   * Wait for the Template Marketplace view to be visible
   */
  async waitForMarketplaceView() {
    await this.page.waitForSelector('text=Template Marketplace', { state: 'visible', timeout: 10000 });
  }

  /**
   * Get the search input field
   */
  getSearchInput(): Locator {
    return this.page.getByPlaceholder(/search templates/i);
  }

  /**
   * Get the category filter select
   */
  getCategoryFilter(): Locator {
    return this.page.locator('text=Category').first();
  }

  /**
   * Search for templates
   */
  async search(query: string) {
    const input = this.getSearchInput();
    await input.fill(query);
  }

  /**
   * Wait for marketplace templates API call
   */
  async waitForMarketplaceData() {
    await this.waitForApiCall('/api/v1/marketplace/templates').catch(() => {
      // Non-critical: endpoint may not exist in all test environments
    });
  }
}
