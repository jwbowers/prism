import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  // Test directory
  testDir: './tests/e2e',

  // File-level parallelism: tests within a file run serially by default.
  // Individual tiers can be run with increased --workers via npm scripts.
  // Tier A (ui) and Tier B (read) are safe at workers:3.
  // Tier C (write) and Tier D (storage) must stay at workers:1.
  fullyParallel: false,

  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Default: serial. Individual tier scripts pass --workers explicitly.
  workers: 1,

  // Reporter to use
  reporter: [
    ['html'],
    ['json', { outputFile: 'playwright-report.json' }],
    process.env.CI ? ['github'] : ['list']
  ],

  // Global setup and teardown
  globalSetup: './tests/e2e/global-setup.js',

  // Global test configuration
  use: {
    // For Wails apps, we test against the served frontend
    // The frontend will connect to the daemon API at localhost:8947
    baseURL: 'http://localhost:3000', // Vite dev server

    // Browser context options
    viewport: { width: 1280, height: 720 },

    // Capture screenshot on failure
    screenshot: 'only-on-failure',

    // Record video on failure
    video: 'retain-on-failure',

    // Collect trace on failure
    trace: 'on-first-retry',

    // Timeout for individual actions
    actionTimeout: 10000
  },

  // Global timeout for each test (includes setup, test body, and cleanup)
  timeout: 180000, // 3 minutes per test - allows for AWS operations (30-180s) + test execution + cleanup

  // ─────────────────────────────────────────────────────────────────────────
  // Execution Tiers
  //
  // Tests are split into 4 tiers by safety level. Run all tiers together
  // (default) or target individual tiers via --project flags for faster
  // feedback. Each tier has chromium + webkit variants.
  //
  //   Tier A – ui        : Pure read-only UI tests. Zero mutations. workers:3 safe.
  //   Tier B – read      : Read-heavy workflows, conditional skips, dry-run modes. workers:2 safe.
  //   Tier C – write     : Stateful tests with broad cleanup patterns. workers:1 only.
  //   Tier D – storage   : Hardcoded AWS resource names, crash risk at concurrency. workers:1 only.
  //
  // npm scripts:
  //   test:e2e           → full suite, workers:1, via run-single.js lock
  //   test:e2e:fast      → Tiers A+B, workers:3 (~30 min)
  //   test:e2e:serial    → Tiers C+D, workers:1 (~60 min)
  //   test:e2e:chromium  → all tiers, chromium only, workers:1
  //   test:e2e:webkit    → all tiers, webkit only, workers:1
  // ─────────────────────────────────────────────────────────────────────────
  projects: [
    // ── Tier A: UI-Only ──────────────────────────────────────────────────
    // Pure navigation and render tests. No API mutations. Safe at workers:3.
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

    // ── Tier B: Read Workflows ───────────────────────────────────────────
    // Read-heavy tests with conditional skips and dry-run modes. No persistent
    // mutations against the daemon. Safe at workers:2.
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

    // ── Tier C: Write Workflows ──────────────────────────────────────────
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

    // ── Tier D: Storage ──────────────────────────────────────────────────
    // Uses hardcoded AWS resource names (test-setup-efs, test-setup-ebs, etc.).
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

  // Start the Vite dev server for the frontend
  // reuseExistingServer: true allows reusing a running dev server (e.g., from a previous test run
  // that didn't clean up). The daemon is always restarted fresh by global-setup.js, so tests
  // remain properly isolated even when the Vite server is reused.
  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: true,
    timeout: 120 * 1000,
  }
})
