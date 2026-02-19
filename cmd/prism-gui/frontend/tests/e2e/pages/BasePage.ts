/**
 * Base Page Object
 *
 * Common functionality for all page objects in Prism GUI E2E tests.
 * Provides navigation, waiting, and common interaction patterns.
 */

import { Page, Locator, expect } from '@playwright/test';

export class BasePage {
  readonly page: Page;
  private consoleMessages: string[] = [];
  private pageErrors: string[] = [];

  constructor(page: Page) {
    this.page = page;

    // Capture browser console messages
    this.page.on('console', msg => {
      const text = `[BROWSER ${msg.type().toUpperCase()}] ${msg.text()}`;
      this.consoleMessages.push(text);
      console.log(text);
    });

    // Capture JavaScript errors
    this.page.on('pageerror', err => {
      const text = `[PAGE ERROR] ${err.message}\n${err.stack || ''}`;
      this.pageErrors.push(text);
      console.error(text);
    });
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
        // Wait for the dialog to actually close using proper Playwright selector
        await this.page.getByRole('dialog').filter({ hasText: /tour/i }).waitFor({ state: 'hidden', timeout: 5000 });
      }
    } catch {
      // Welcome tour not present, that's okay
    }
  }

  /**
   * Wait for dialog to close
   * Cloudscape-specific: Wait for visible dialogs to become hidden
   */
  async waitForDialogClose(timeout: number = 5000) {
    try {
      // Find all dialogs and check which are visible
      const dialogs = this.page.getByRole('dialog');
      const count = await dialogs.count();

      // Check each dialog to see if it's visible
      for (let i = 0; i < count; i++) {
        const dialog = dialogs.nth(i);
        const isVisible = await dialog.isVisible();
        if (isVisible) {
          // Wait for this visible dialog to be hidden
          await dialog.waitFor({ state: 'hidden', timeout });
          // INCREASED delay for Cloudscape animation to fully complete
          await this.page.waitForTimeout(1000);
          return; // Exit after first visible dialog closes
        }
      }
    } catch {
      // Dialog already closed or timeout - safe to continue
    }
  }

  /**
   * Wait for the application to finish initial data loading
   */
  async waitForAppReady() {
    try {
      console.log('[waitForAppReady] Waiting for React root to mount...');
      // Wait for React root to mount
      await this.page.waitForSelector('#root > *', { state: 'visible', timeout: 15000 });
      console.log('[waitForAppReady] React root mounted successfully');

      console.log('[waitForAppReady] Waiting for Cloudscape side navigation...');
      // Wait for Cloudscape side navigation to be present and visible
      // Cloudscape renders multiple navigation elements (some hidden), so we wait for actual navigation links
      // which will only be present in the visible navigation
      await this.page.waitForSelector('a[href="/dashboard"]', { state: 'visible', timeout: 10000 });
      console.log('[waitForAppReady] Side navigation is visible');

      // Wait for initial API calls to complete
      // The app makes several API calls on load: templates, instances, storage, etc.
      // Wait for the templates API call which is one of the key indicators
      await this.waitForApiCall('/api/v1/templates', 15000).catch(() => {
        console.log('[waitForAppReady] Templates API call did not complete (non-critical)');
      });

      // Wait for React to finish rendering - look for any tab content to be visible
      await this.page.locator('[role="tabpanel"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {
        console.log('[waitForAppReady] Tab content not visible (may be on different page)');
      });

      console.log('[waitForAppReady] App is ready');
    } catch (error) {
      // If rendering fails, log detailed diagnostics
      console.error('[waitForAppReady] React app may not be rendering:', error);

      // Capture page HTML for debugging
      const html = await this.page.content();
      console.log('[waitForAppReady] Page HTML length:', html.length, 'characters');
      console.log('[waitForAppReady] Page title:', await this.page.title());
      console.log('[waitForAppReady] Page URL:', this.page.url());

      // Check for specific elements
      const rootCount = await this.page.locator('#root').count();
      const rootChildren = await this.page.locator('#root > *').count();
      const navCount = await this.page.locator('a[href="/dashboard"]').count();

      console.log('[waitForAppReady] Elements found:');
      console.log('  - #root elements:', rootCount);
      console.log('  - #root children:', rootChildren);
      console.log('  - navigation links (a[href="/dashboard"]):', navCount);

      // Check if there are any errors collected
      if (this.pageErrors.length > 0) {
        console.error('[waitForAppReady] Page errors detected:');
        this.pageErrors.forEach(err => console.error('  -', err));
      }

      await this.page.waitForLoadState('domcontentloaded', { timeout: 10000 });

      // Re-throw if navigation still isn't present - this indicates a serious problem
      // Check for actual navigation links (which will only be present in visible navigation)
      const navExists = await this.page.locator('a[href="/dashboard"]').count();
      if (navExists === 0) {
        throw new Error('React app failed to render: side navigation not present');
      }
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
   * Navigate to a specific tab using Cloudscape SideNavigation links
   */
  async navigateToTab(tabName: string) {
    // First, ensure no dialogs are blocking the navigation
    await this.waitForDialogClose(3000);

    // CRITICAL: Wait for navigation to be fully rendered and interactive
    // This is essential for tests that call navigateToTab() immediately after page load
    // Wait for actual navigation links (Dashboard link) which will only be present in visible navigation
    console.log(`[navigateToTab] Ensuring navigation is ready for "${tabName}"...`);
    await this.page.waitForSelector('a[href="/dashboard"]', {
      state: 'visible',
      timeout: 15000
    }).catch(async (error) => {
      console.error('[navigateToTab] Navigation not visible - attempting recovery...');
      await this.waitForAppReady();
      throw error;
    });

    // Map common tab names to actual SideNavigation link text
    const linkTextMap: Record<string, string> = {
      'dashboard': 'Dashboard',
      'templates': 'Templates',
      'instances': 'My Workspaces',
      'workspaces': 'My Workspaces',
      'storage': 'Storage',
      'backups': 'Backups',
      'projects': 'Projects',
      'settings': 'Settings',
      'profiles': 'Profiles',
      'idle': 'Idle Management',
      'users': 'Users',
      'invitations': 'Invitations'
    };

    const linkText = linkTextMap[tabName.toLowerCase()] || tabName;
    console.log(`[navigateToTab] Looking for link: "${linkText}"`);

    const link = this.page.getByRole('link', { name: new RegExp(linkText, 'i') });

    // Wait for the link to be both visible and enabled before clicking
    await link.waitFor({ state: 'visible', timeout: 10000 });
    console.log(`[navigateToTab] Link "${linkText}" is visible, clicking...`);

    await link.click();
    console.log(`[navigateToTab] Successfully clicked "${linkText}"`);

    // Wait for the DOM to update after tab switch
    await this.page.waitForLoadState('domcontentloaded', { timeout: 3000 }).catch(() => {});
    await this.waitForLoadingComplete();
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
      // No timeout specified - uses test-level timeout to accommodate AWS operations
      await this.page.waitForSelector('[data-testid="loading"], .loading-spinner', {
        state: 'hidden',
      });
    } catch {
      // Loading indicator might not exist, that's okay
    }
  }

  /**
   * Click button by text or role
   * If multiple buttons match, clicks the last visible one (usually in a modal/dialog)
   */
  async clickButton(text: string) {
    // Use exact match for common action buttons to avoid matching resource names
    // like "cancel-delete-test", "delete-test-efs", "create-test-project", etc.
    const lowerText = text.toLowerCase();
    const exactMatchButtons = ['cancel', 'create', 'delete', 'attach', 'detach', 'mount', 'unmount', 'confirm'];

    const options = exactMatchButtons.includes(lowerText)
      ? { name: text.charAt(0).toUpperCase() + text.slice(1), exact: true }
      : { name: new RegExp(text, 'i') };

    const button = this.page.getByRole('button', options);
    const count = await button.count();

    if (count > 1) {
      // Multiple buttons found - click the last one (the one in the topmost dialog)
      await button.last().click();
    } else {
      await button.click();
    }
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
      // Wait for network response or 500ms, whichever comes first
      await this.page.waitForResponse(
        response => response.url().includes('/api/v1/health'),
        { timeout: 500 }
      ).catch(() => {
        // Timeout is expected when daemon not ready yet
      });
    }
    throw new Error('Daemon not ready after maximum attempts');
  }
}
