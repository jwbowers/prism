import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  // Test directory
  testDir: './tests/e2e',
  
  // Run tests in files in parallel
  fullyParallel: true,
  
  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,
  
  // Retry on CI only
  retries: process.env.CI ? 2 : 0,
  
  // Use single worker for daemon integration tests to avoid port conflicts
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

  // Configure projects for Wails WebView engines only (desktop only - CloudWorkstation is not mobile)
  // - Chromium: Windows (WebView2)
  // - Webkit: macOS (WKWebView), Linux (webkit2gtk)
  // Firefox not needed as Wails doesn't use it
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    }
  ],

  // Start the Vite dev server for the frontend
  // Note: reuseExistingServer set to false for reliability
  // This ensures tests always have a fresh server and work in all environments
  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: false,
    timeout: 120 * 1000,
  }
})