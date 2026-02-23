import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',

  // Long timeouts for real AWS operations (hibernation + storage can take 5+ minutes)
  timeout: 600000, // 10 minutes per test

  // Sequential execution to avoid AWS rate limits and port conflicts
  workers: 1,
  fullyParallel: false,

  globalSetup: './tests/e2e/setup-aws.js',
  globalTeardown: './tests/e2e/teardown-aws.js',

  use: {
    baseURL: 'http://localhost:3000',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 30000,
  },

  projects: [
    {
      // Use chromium only for AWS tests - webkit has resource exhaustion at 2h mark
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  reporter: [
    ['html', { outputFolder: 'playwright-report-aws' }],
    ['json', { outputFile: 'playwright-report-aws.json' }],
    ['list'],
  ],

  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: true,
    timeout: 120 * 1000,
  },
})
