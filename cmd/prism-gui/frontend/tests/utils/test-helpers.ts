/**
 * Test Helper Functions for Prism GUI Tests
 *
 * Common utilities and helpers used across all test types.
 */

import { Page } from '@playwright/test';

/**
 * Wait for element to be visible with timeout
 */
export async function waitForElement(
  page: Page,
  selector: string,
  timeout: number = 5000
): Promise<boolean> {
  try {
    await page.waitForSelector(selector, { state: 'visible', timeout });
    return true;
  } catch {
    return false;
  }
}

/**
 * Wait for API call to complete
 */
export async function waitForApiCall(
  page: Page,
  urlPattern: string | RegExp,
  timeout: number = 10000
): Promise<boolean> {
  try {
    await page.waitForResponse(
      (response) => {
        const url = response.url();
        if (typeof urlPattern === 'string') {
          return url.includes(urlPattern);
        }
        return urlPattern.test(url);
      },
      { timeout }
    );
    return true;
  } catch {
    return false;
  }
}

/**
 * Navigate to a specific tab in the GUI
 */
export async function navigateToTab(page: Page, tabName: string) {
  const tabSelectors: Record<string, string> = {
    templates: '[data-testid="templates-tab"]',
    instances: '[data-testid="instances-tab"]',
    storage: '[data-testid="storage-tab"]',
    settings: '[data-testid="settings-tab"]',
  };

  const selector = tabSelectors[tabName.toLowerCase()];
  if (!selector) {
    throw new Error(`Unknown tab: ${tabName}`);
  }

  await page.click(selector);
  await page.waitForTimeout(500); // Wait for tab transition
}

/**
 * Fill form field by label
 */
export async function fillFormField(
  page: Page,
  label: string,
  value: string
) {
  const input = page.locator(`label:has-text("${label}") + input, input[aria-label="${label}"]`);
  await input.fill(value);
}

/**
 * Select option from dropdown by label
 */
export async function selectOption(
  page: Page,
  label: string,
  value: string
) {
  const select = page.locator(`label:has-text("${label}") + select, select[aria-label="${label}"]`);
  await select.selectOption(value);
}

/**
 * Click button by text or test ID
 */
export async function clickButton(
  page: Page,
  textOrTestId: string
) {
  const button = page.locator(`button:has-text("${textOrTestId}"), [data-testid="${textOrTestId}"]`);
  await button.click();
}

/**
 * Wait for loading state to complete
 */
export async function waitForLoadingComplete(
  page: Page,
  timeout: number = 10000
) {
  try {
    await page.waitForSelector('[data-testid="loading-spinner"]', {
      state: 'hidden',
      timeout,
    });
  } catch {
    // Loading spinner might not exist, that's okay
  }
}

/**
 * Check if element exists (without throwing)
 */
export async function elementExists(
  page: Page,
  selector: string
): Promise<boolean> {
  return (await page.locator(selector).count()) > 0;
}

/**
 * Get text content of element
 */
export async function getTextContent(
  page: Page,
  selector: string
): Promise<string | null> {
  try {
    return await page.locator(selector).textContent();
  } catch {
    return null;
  }
}

/**
 * Wait for multiple conditions
 */
export async function waitForAll(
  page: Page,
  conditions: (() => Promise<any>)[]
): Promise<void> {
  await Promise.all(conditions.map((fn) => fn()));
}

/**
 * Retry action until it succeeds or times out
 */
export async function retryUntilSuccess<T>(
  action: () => Promise<T>,
  options: {
    maxAttempts?: number;
    delayMs?: number;
    onError?: (error: Error) => void;
  } = {}
): Promise<T> {
  const { maxAttempts = 5, delayMs = 1000, onError } = options;

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await action();
    } catch (error) {
      if (attempt === maxAttempts) {
        throw error;
      }
      if (onError && error instanceof Error) {
        onError(error);
      }
      await new Promise((resolve) => setTimeout(resolve, delayMs));
    }
  }

  throw new Error('Retry failed: max attempts reached');
}

/**
 * Take screenshot with timestamp
 */
export async function takeTimestampedScreenshot(
  page: Page,
  name: string
): Promise<void> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  await page.screenshot({
    path: `test-results/screenshots/${name}-${timestamp}.png`,
    fullPage: true,
  });
}

/**
 * Clear all local storage and cookies
 */
export async function clearBrowserState(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
  await page.context().clearCookies();
}

/**
 * Mock API response
 */
export async function mockApiResponse(
  page: Page,
  urlPattern: string | RegExp,
  response: any,
  status: number = 200
): Promise<void> {
  await page.route(urlPattern, (route) => {
    route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(response),
    });
  });
}

/**
 * Wait for daemon to be ready
 */
export async function waitForDaemonReady(
  daemonUrl: string = 'http://localhost:8947',
  maxAttempts: number = 40,
  delayMs: number = 500
): Promise<boolean> {
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      const response = await fetch(`${daemonUrl}/api/v1/health`);
      if (response.ok) {
        return true;
      }
    } catch {
      // Daemon not ready yet
    }
    await new Promise((resolve) => setTimeout(resolve, delayMs));
  }
  return false;
}

/**
 * Format currency for comparison in tests
 */
export function formatCurrency(amount: number): string {
  return `$${amount.toFixed(2)}`;
}

/**
 * Parse instance ID from various formats
 */
export function parseInstanceId(idOrFullString: string): string {
  const match = idOrFullString.match(/i-[a-f0-9]+/);
  return match ? match[0] : idOrFullString;
}

/**
 * Sleep for specified milliseconds
 */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Generate unique test ID
 */
export function generateTestId(prefix: string = 'test'): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}
