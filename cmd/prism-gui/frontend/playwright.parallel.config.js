/**
 * Parallel Execution Config for Playwright
 *
 * Splits the test suite into 4 tiers by safety level, enabling faster feedback
 * by running safe tiers with increased worker counts and intra-file parallelism.
 *
 * Usage:
 *   npm run test:e2e:fast    → Tiers A+B, workers:4, fullyParallel  (~12 min vs ~60 min)
 *   npm run test:e2e:serial  → Tiers C+D, workers:1  (same as default, safe)
 *
 * To run both simultaneously in separate terminals:
 *   Terminal 1: npm run test:e2e:fast
 *   Terminal 2: npm run test:e2e:serial
 *
 * Tier breakdown:
 *
 *   Tier A (ui):      basic, navigation, error-boundary, form-validation, settings
 *                     → Zero mutations. fullyParallel safe. ~25s vs ~75s sequential.
 *
 *   Tier B (read):    backup, hibernation, instance, budget workflows
 *                     → testMode bypasses all AWS state — daemon never mutates on mock-*
 *                       instances. Each test has isolated browser context. fullyParallel safe.
 *                       Hibernation alone: 18 tests, ceil(18/4)×23s=115s vs 18×23s=414s.
 *
 *   Tier C (write):   profile, project, user, invitation workflows
 *                     → Stateful. Broad cleanup patterns collide across workers. workers:1 only.
 *
 *   Tier D (storage): storage-workflows
 *                     → Hardcoded AWS resource names. Concurrent EBS/EFS ops crash daemon.
 *                       workers:1 only.
 *
 * Safety: fullyParallel:true is harmless for Tiers C+D because test:e2e:serial
 * hardcodes --workers=1, and the config default is also workers:1.
 */

import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true, // Intra-file parallelism for Tiers A+B; Tiers C+D use --workers=1
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: parseInt(process.env.PLAYWRIGHT_WORKERS || '1'),

  reporter: [
    ['html', { outputFolder: 'playwright-report-parallel' }],
    ['json', { outputFile: 'playwright-report-parallel.json' }],
    process.env.CI ? ['github'] : ['list']
  ],

  globalSetup: './tests/e2e/global-setup.js',
  globalTeardown: './tests/e2e/global-teardown.js',

  use: {
    baseURL: 'http://localhost:3000',
    viewport: { width: 1280, height: 720 },
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'on-first-retry',
    actionTimeout: 10000
  },

  timeout: 180000,

  projects: [
    // ── Tier A: UI-Only ──────────────────────────────────────────────────────
    // Pure navigation and render tests. Zero mutations. Safe at workers:3.
    {
      name: 'ui-chromium',
      testMatch: /(basic|navigation|error-boundary|form-validation|settings)\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'ui-webkit',
      testMatch: /(basic|navigation|error-boundary|form-validation|settings)\.spec\.ts/,
      use: { ...devices['Desktop Safari'] },
    },

    // ── Tier B: Read Workflows ───────────────────────────────────────────────
    // Read-heavy tests with conditional skips and dry-run modes. No persistent
    // mutations. Safe at workers:2.
    {
      name: 'read-chromium',
      testMatch: /(backup|hibernation|instance|budget)-workflows\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'read-webkit',
      testMatch: /(backup|hibernation|instance|budget)-workflows\.spec\.ts/,
      use: { ...devices['Desktop Safari'] },
    },

    // ── Tier C: Write Workflows ──────────────────────────────────────────────
    // Stateful tests using timestamped resource names. Broad beforeAll/afterEach
    // cleanup patterns would collide across parallel workers. Must run serially.
    {
      name: 'write-chromium',
      testMatch: /(profile|project|user|invitation)-workflows\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'write-webkit',
      testMatch: /(profile|project|user|invitation)-workflows\.spec\.ts/,
      use: { ...devices['Desktop Safari'] },
    },

    // ── Tier D: Storage ──────────────────────────────────────────────────────
    // Hardcoded AWS resource names (test-setup-efs, test-setup-ebs, etc.).
    // Concurrent EBS/EFS operations can crash the daemon. Must run serially
    // and isolated from other storage-touching operations.
    {
      name: 'storage-chromium',
      testMatch: /storage-workflows\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'storage-webkit',
      testMatch: /storage-workflows\.spec\.ts/,
      use: { ...devices['Desktop Safari'] },
    },
  ],

  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: true,
    timeout: 120 * 1000,
  }
})
