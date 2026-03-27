import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test-setup.ts',
    css: true,
    testTimeout: 30000, // 30 seconds per test (increased from default 5s)
    hookTimeout: 30000, // 30 seconds for hooks
    teardownTimeout: 10000, // 10 seconds for teardown
    silent: false, // Show test names but not all console output
    // Only run Vitest unit tests — exclude Playwright E2E and visual tests
    include: ['src/**/*.test.{ts,tsx}', 'tests/unit/**/*.test.{ts,tsx,js}'],
    exclude: ['tests/e2e/**', 'tests/visual/**', 'node_modules/**'],
  },
});