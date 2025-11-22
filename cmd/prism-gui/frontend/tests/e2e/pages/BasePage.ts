/**
 * Base Page Object
 *
 * Common functionality for all page objects in Prism GUI E2E tests.
 * Provides navigation, waiting, and common interaction patterns.
 */

import { Page, Locator, expect } from '@playwright/test';

export class BasePage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Navigate to the application root
   */
  async goto() {
    await this.page.goto('/');
    await this.waitForPageLoad();
    await this.dismissWelcomeTour();
    await this.waitForAppReady();
  }

  /**
   * Dismiss the welcome tour dialog if it's present
   */
  async dismissWelcomeTour() {
    try {
      const skipButton = this.page.getByRole('button', { name: /skip tour/i });
      if (await skipButton.isVisible({ timeout: 2000 })) {
        await skipButton.click();
        await this.page.waitForTimeout(500); // Wait for dialog to close
      }
    } catch {
      // Welcome tour not present, that's okay
    }
  }

  /**
   * Wait for the application to finish initial data loading
   */
  async waitForAppReady() {
    try {
      // Wait for initial API calls to complete
      // The app makes several API calls on load: templates, instances, storage, etc.
      // Wait for the templates API call which is one of the key indicators
      await this.waitForApiCall('/api/v1/templates', 15000);

      // Wait a bit more for React state to update and render
      await this.page.waitForTimeout(2000);
    } catch {
      // If API call doesn't complete, just wait a reasonable amount
      await this.page.waitForTimeout(5000);
    }
  }

  /**
   * Wait for page to be fully loaded
   */
  async waitForPageLoad() {
    // Use 'load' instead of 'networkidle' because in test mode with failing API calls,
    // the network may never go completely idle due to retries
    await this.page.waitForLoadState('load', { timeout: 10000 });
  }

  /**
   * Navigate to a specific tab using accessible link navigation
   */
  async navigateToTab(tabName: string) {
    const link = this.page.getByRole('link', { name: new RegExp(tabName, 'i') });
    await link.click();
    await this.page.waitForTimeout(500); // Allow tab transition
  }

  /**
   * Wait for API call to complete
   */
  async waitForApiCall(urlPattern: string | RegExp, timeout: number = 10000) {
    await this.page.waitForResponse(
      (response) => {
        const url = response.url();
        if (typeof urlPattern === 'string') {
          return url.includes(urlPattern);
        }
        return urlPattern.test(url);
      },
      { timeout }
    );
  }

  /**
   * Wait for element to be visible
   */
  async waitForElement(locator: Locator, timeout: number = 5000) {
    await locator.waitFor({ state: 'visible', timeout });
  }

  /**
   * Wait for loading to complete (look for loading indicators)
   */
  async waitForLoadingComplete() {
    try {
      // Wait for common loading indicators to disappear
      await this.page.waitForSelector('[data-testid="loading"], .loading-spinner', {
        state: 'hidden',
        timeout: 10000,
      });
    } catch {
      // Loading indicator might not exist, that's okay
    }
  }

  /**
   * Click button by text or role
   */
  async clickButton(text: string) {
    const button = this.page.getByRole('button', { name: new RegExp(text, 'i') });
    await button.click();
  }

  /**
   * Fill input field by label
   */
  async fillInput(label: string, value: string) {
    const input = this.page.getByLabel(new RegExp(label, 'i'));
    await input.fill(value);
  }

  /**
   * Select option from dropdown by label
   */
  async selectOption(label: string, value: string) {
    const select = this.page.getByLabel(new RegExp(label, 'i'));
    await select.selectOption(value);
  }

  /**
   * Check if element exists
   */
  async elementExists(locator: Locator): Promise<boolean> {
    return (await locator.count()) > 0;
  }

  /**
   * Get text content of element
   */
  async getTextContent(locator: Locator): Promise<string | null> {
    try {
      return await locator.textContent();
    } catch {
      return null;
    }
  }

  /**
   * Take screenshot with custom name
   */
  async takeScreenshot(name: string) {
    await this.page.screenshot({
      path: `test-results/screenshots/${name}.png`,
      fullPage: true,
    });
  }

  /**
   * Verify no JavaScript errors occurred
   */
  setupErrorTracking(): string[] {
    const errors: string[] = [];
    this.page.on('pageerror', (error) => {
      errors.push(error.message);
    });
    return errors;
  }

  /**
   * Wait for daemon API to be ready
   */
  async waitForDaemonReady(maxAttempts: number = 20) {
    const daemonUrl = 'http://localhost:8947';
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      try {
        const response = await fetch(`${daemonUrl}/api/v1/health`);
        if (response.ok) {
          return true;
        }
      } catch {
        // Daemon not ready yet
      }
      await this.page.waitForTimeout(500);
    }
    throw new Error('Daemon not ready after maximum attempts');
  }
}
