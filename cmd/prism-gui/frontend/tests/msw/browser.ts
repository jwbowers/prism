/**
 * MSW Browser Setup for E2E Tests (Playwright)
 *
 * This worker is used in E2E tests running in a real browser.
 * Note: For E2E tests, we typically use real daemon integration,
 * so this is mainly for specialized browser-based mocking scenarios.
 */

import { setupWorker } from 'msw/browser';
import { handlers } from './handlers';

/**
 * MSW worker instance for browser tests
 */
export const worker = setupWorker(...handlers);

/**
 * Start MSW worker in browser
 *
 * Call this in your E2E test setup if you want to mock APIs in browser.
 *
 * @example
 * ```typescript
 * // In E2E test or global setup
 * await import('tests/msw/browser').then(({ worker }) => worker.start());
 * ```
 */
export async function startMSWWorker() {
  await worker.start({
    onUnhandledRequest: 'warn',
    serviceWorker: {
      url: '/mockServiceWorker.js',
    },
  });
}
