/**
 * MSW Server Setup for Node.js Environment (Vitest)
 *
 * This server is used in unit and component tests running in Node.js with jsdom.
 */

import { setupServer } from 'msw/node';
import { beforeAll, afterEach, afterAll } from 'vitest';
import { handlers } from './handlers';

/**
 * MSW server instance for Node.js tests
 */
export const server = setupServer(...handlers);

/**
 * Setup MSW server hooks for tests
 *
 * Call this in your test setup file or individual test suites.
 *
 * @example
 * ```typescript
 * import { setupMSW } from 'tests/msw/server';
 *
 * setupMSW(); // Sets up MSW for all tests in this file
 * ```
 */
export function setupMSW() {
  // Start server before all tests
  beforeAll(() => server.listen({ onUnhandledRequest: 'warn' }));

  // Reset handlers after each test
  afterEach(() => server.resetHandlers());

  // Clean up after all tests
  afterAll(() => server.close());
}
