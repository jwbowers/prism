# Prism GUI Testing Guide

**Version**: 0.5.16
**Last Updated**: November 16, 2025
**Related Issue**: #297 (GUI Testing Phase 0)

This guide provides comprehensive documentation for testing the Prism GUI application across all test types: backend service layer tests, component tests, E2E tests, and visual regression tests.

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Running Tests](#running-tests)
4. [Writing New Tests](#writing-new-tests)
5. [Test Utilities](#test-utilities)
6. [Debugging](#debugging)
7. [CI/CD Integration](#cicd-integration)
8. [Common Issues](#common-issues)
9. [Best Practices](#best-practices)

---

## Overview

The Prism GUI uses a **three-layer testing strategy** to ensure comprehensive coverage:

### Testing Layers

```
┌─────────────────────────────────────────┐
│  Layer 3: E2E Tests (Playwright)        │
│  - Full workflow testing                │
│  - Real daemon integration              │
│  - Browser automation                   │
└─────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────┐
│  Layer 2: Component Tests (Vitest)      │
│  - React component testing              │
│  - MSW API mocking                      │
│  - User interaction simulation          │
└─────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────┐
│  Layer 1: Backend Service Tests (Go)    │
│  - PrismService API testing             │
│  - Daemon integration                   │
│  - Service layer logic                  │
└─────────────────────────────────────────┘
```

### Test Metrics (Target)

- **Backend Service Tests**: 35+ test functions (~1,300 lines)
- **Component Tests**: 66+ test cases (~2,150 lines)
- **E2E Tests**: 38+ scenarios (~1,700 lines)
- **Visual Regression Tests**: Component + page snapshots (~400 lines)
- **Total**: 171+ tests (~5,500 lines)
- **Coverage Goal**: ≥80% for frontend code, ≥95% pass rate

---

## Prerequisites

### Required Tools

```bash
# Node.js 20+
node --version  # Should be v20+

# npm dependencies (install if needed)
npm install

# Playwright browsers (for E2E tests)
npx playwright install chromium

# Go 1.24+ (for backend tests)
go version  # Should be 1.24+

# Prism daemon binary (for E2E tests)
go build -o ../../bin/prismd ../../../cmd/prismd
```

### Environment Setup

```bash
# Navigate to frontend directory
cd cmd/prism-gui/frontend

# Install dependencies
npm ci

# Verify test setup
npm run typecheck
npm run lint
```

---

## Running Tests

### Quick Reference

```bash
# Run ALL tests
npm run test:all

# Unit/Component tests only
npm test                    # Watch mode
npm run test:unit           # Run once
npm run test:coverage       # With coverage report

# E2E tests only
npm run test:e2e            # All browsers
npm run test:e2e -- --project=chromium  # Chromium only

# Visual regression tests
npm run test:visual         # Percy + Playwright

# Type checking and linting
npm run typecheck
npm run lint
npm run lint:fix
```

### Backend Service Layer Tests (Go)

Backend tests are located in `cmd/prism-gui/gui_test.go` and test the PrismService layer that interfaces with the daemon.

```bash
# From project root
cd /Users/scttfrdmn/src/prism

# Run all GUI backend tests
go test -v ./cmd/prism-gui/

# Run specific test
go test -v ./cmd/prism-gui/ -run TestPrismService

# Run with coverage
go test -v -coverprofile=coverage.out ./cmd/prism-gui/
go tool cover -html=coverage.out
```

**Test Files**:
- `gui_test.go` (938 lines, 18 test functions)
- `gui_profile_test.go` (planned - Phase 1)
- `gui_storage_test.go` (planned - Phase 1)
- `gui_hibernation_test.go` (planned - Phase 1)
- `gui_idle_test.go` (planned - Phase 1)
- `gui_backup_test.go` (planned - Phase 1)

### Component Tests (Vitest)

Component tests use Vitest, React Testing Library, and MSW for API mocking.

```bash
# Run in watch mode (recommended for development)
npm test

# Run all tests once
npm run test:unit

# Run specific test file
npm test -- App.test.tsx

# Run with coverage
npm run test:coverage

# Run tests matching pattern
npm test -- --grep="Template"
```

**Test Files**:
- `src/App.test.tsx` (339 lines)
- `src/App.simple.test.tsx`
- `src/App.behavior.test.tsx`
- `tests/unit/instance-management.test.js`
- `tests/unit/template-selection.test.js`
- `tests/unit/theme-management.test.js`

### E2E Tests (Playwright)

E2E tests use Playwright for browser automation and test complete user workflows.

```bash
# Run all E2E tests
npm run test:e2e

# Run specific browser only
npm run test:e2e -- --project=chromium
npm run test:e2e -- --project=firefox
npm run test:e2e -- --project=webkit

# Run specific test file
npm run test:e2e -- tests/e2e/instance-management.spec.js

# Run in headed mode (see the browser)
npm run test:e2e -- --headed

# Run in debug mode
npm run test:e2e -- --debug

# Run with UI mode (interactive)
npx playwright test --ui

# Generate test report
npx playwright show-report
```

**Test Files**:
- `tests/e2e/instance-management.spec.js`
- `tests/e2e/launch-workflow.spec.js`
- `tests/e2e/daemon-integration.spec.js`
- `tests/e2e/navigation.spec.js`
- `tests/e2e/form-validation.spec.js`
- `tests/e2e/error-boundary.spec.js`
- + 7 more files

**Requirements**:
- Daemon binary must be built: `go build -o ../../bin/prismd ../../../cmd/prismd`
- Vite dev server starts automatically via Playwright config
- Tests run with single worker to avoid port conflicts

### Visual Regression Tests (Percy)

Visual regression tests use Percy for screenshot comparison.

```bash
# Run visual tests with Percy
npm run test:visual

# Note: Requires PERCY_TOKEN environment variable
export PERCY_TOKEN=your_percy_token
npm run test:visual

# Run without Percy (local screenshots only)
npm run test:e2e -- tests/visual/
```

**Test Files**:
- `tests/visual/themes.spec.js`
- More visual tests planned in Phase 4

---

## Writing New Tests

### Component Test Pattern

```typescript
// src/components/__tests__/ProfileManager.test.tsx
import { describe, test, expect } from 'vitest';
import { render, screen, waitFor } from 'tests/utils';
import { setupMSW } from 'tests/msw/server';
import { createMockProfiles } from 'tests/utils/mock-data-factories';
import { ProfileManager } from '../ProfileManager';

describe('ProfileManager', () => {
  setupMSW(); // Enable MSW API mocking

  test('displays list of profiles', async () => {
    render(<ProfileManager />);

    await waitFor(() => {
      expect(screen.getByText('default')).toBeInTheDocument();
      expect(screen.getByText('research-profile')).toBeInTheDocument();
    });
  });

  test('creates new profile', async () => {
    const { user } = render(<ProfileManager />);

    // Click add button
    await user.click(screen.getByRole('button', { name: /add profile/i }));

    // Fill form
    await user.type(screen.getByLabelText(/profile name/i), 'test-profile');
    await user.selectOptions(screen.getByLabelText(/region/i), 'us-west-2');

    // Submit
    await user.click(screen.getByRole('button', { name: /create/i }));

    // Verify success
    await waitFor(() => {
      expect(screen.getByText('test-profile')).toBeInTheDocument();
    });
  });

  test('handles API error gracefully', async () => {
    server.use(
      http.get('http://localhost:8947/api/v1/profiles', () => {
        return HttpResponse.json({ error: 'Server error' }, { status: 500 });
      })
    );

    render(<ProfileManager />);

    await waitFor(() => {
      expect(screen.getByText(/failed to load profiles/i)).toBeInTheDocument();
    });
  });
});
```

### E2E Test Pattern

```typescript
// tests/e2e/profile-workflows.spec.ts
import { test, expect } from '@playwright/test';
import {
  navigateToTab,
  fillFormField,
  clickButton,
  waitForLoadingComplete,
} from '../utils/test-helpers';

test.describe('Profile Management Workflows', () => {
  test('complete profile lifecycle', async ({ page }) => {
    await page.goto('/');

    // Navigate to settings/profiles
    await navigateToTab(page, 'settings');
    await clickButton(page, 'Profiles');

    // Create profile
    await clickButton(page, 'Add Profile');
    await fillFormField(page, 'Profile Name', 'test-profile');
    await fillFormField(page, 'AWS Profile', 'default');
    await fillFormField(page, 'Region', 'us-west-2');
    await clickButton(page, 'Create');

    // Verify profile created
    await expect(page.locator('text=test-profile')).toBeVisible();

    // Switch to profile
    await page.click('[data-profile="test-profile"] button:has-text("Switch")');
    await expect(page.locator('.active-profile')).toContainText('test-profile');

    // Delete profile
    await page.click('[data-profile="test-profile"] button:has-text("Delete")');
    await clickButton(page, 'Confirm');
    await expect(page.locator('text=test-profile')).not.toBeVisible();
  });
});
```

### Backend Service Test Pattern

```go
// cmd/prism-gui/gui_profile_test.go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestProfileManagement(t *testing.T) {
    service := setupTestService(t)
    defer cleanupTestService(t, service)

    t.Run("CreateProfile", func(t *testing.T) {
        profile := ProfileRequest{
            Name:       "test-profile",
            AWSProfile: "default",
            Region:     "us-west-2",
        }

        err := service.CreateProfile(profile)
        assert.NoError(t, err)

        profiles, err := service.GetProfiles()
        assert.NoError(t, err)
        assert.Contains(t, profiles, "test-profile")
    })

    t.Run("DeleteProfile", func(t *testing.T) {
        err := service.DeleteProfile("test-profile")
        assert.NoError(t, err)

        profiles, err := service.GetProfiles()
        assert.NoError(t, err)
        assert.NotContains(t, profiles, "test-profile")
    })
}
```

---

## Test Utilities

### Mock Data Factories

Located in `tests/utils/mock-data-factories.ts`:

```typescript
import {
  createMockTemplate,
  createMockTemplates,
  createMockInstance,
  createMockInstances,
  createMockProfile,
  createMockProfiles,
} from 'tests/utils';

// Create single mock
const template = createMockTemplate({
  Name: 'Custom Template',
  Complexity: 'advanced',
});

// Create collection
const templates = createMockTemplates();

// Customize mock
const instance = createMockInstance({
  name: 'my-test-instance',
  status: 'running',
  cost_per_hour: 1.50,
});
```

### Custom Render Function

Located in `tests/utils/custom-render.tsx`:

```typescript
import { render } from 'tests/utils';

// Render with default mocks
render(<MyComponent />);

// Render with custom Wails mocks
render(<MyComponent />, {
  wailsMocks: {
    GetTemplates: async () => createMockTemplates(),
    LaunchInstance: async () => ({ instance_id: 'i-test123' }),
  },
});
```

### Test Helpers

Located in `tests/utils/test-helpers.ts`:

```typescript
import {
  waitForElement,
  waitForApiCall,
  navigateToTab,
  fillFormField,
  selectOption,
  clickButton,
  retryUntilSuccess,
} from 'tests/utils';

// Wait for element
await waitForElement(page, '.template-card');

// Navigate to tab
await navigateToTab(page, 'instances');

// Fill form
await fillFormField(page, 'Instance Name', 'my-instance');

// Retry with backoff
const result = await retryUntilSuccess(async () => {
  return await fetchData();
}, { maxAttempts: 5, delayMs: 1000 });
```

### Custom Assertions

Located in `tests/utils/assertions.ts`:

```typescript
import {
  assertTemplateCard,
  assertInstanceRow,
  assertNotification,
  assertModalOpen,
  assertFieldError,
} from 'tests/utils';

// Assert template card displays correctly
await assertTemplateCard(page, {
  Name: 'Python Machine Learning',
  Category: 'Machine Learning',
});

// Assert instance row
await assertInstanceRow(page, {
  name: 'my-ml-research',
  status: 'running',
});

// Assert notification
await assertNotification(page, 'Instance launched successfully', 'success');
```

### MSW API Mocking

Located in `tests/msw/`:

```typescript
import { setupMSW } from 'tests/msw/server';
import { server } from 'tests/msw/server';
import { http, HttpResponse } from 'msw';
import { errorHandlers, emptyHandlers } from 'tests/msw/handlers';

describe('MyComponent', () => {
  setupMSW(); // Enable MSW for all tests

  test('with custom response', () => {
    server.use(
      http.get('http://localhost:8947/api/v1/templates', () => {
        return HttpResponse.json([{ Name: 'Custom Template' }]);
      })
    );
    // ... test code
  });

  test('with error response', () => {
    server.use(...errorHandlers);
    // ... test code
  });

  test('with empty response', () => {
    server.use(...emptyHandlers);
    // ... test code
  });
});
```

---

## Debugging

### Debugging Component Tests

```bash
# Run tests in watch mode
npm test

# Run specific test file
npm test -- ProfileManager.test.tsx

# Add console.log statements
test('my test', () => {
  render(<MyComponent />);
  screen.debug(); // Print DOM
  console.log(screen.getByText('Hello').textContent);
});

# Check what queries are available
import { logRoles } from '@testing-library/react';
const { container } = render(<MyComponent />);
logRoles(container);
```

### Debugging E2E Tests

```bash
# Run in headed mode (see browser)
npm run test:e2e -- --headed

# Run in debug mode (step through)
npm run test:e2e -- --debug

# Run with UI mode (interactive)
npx playwright test --ui

# Add debugging statements
test('my test', async ({ page }) => {
  await page.pause(); // Pause execution
  await page.screenshot({ path: 'debug.png' }); // Take screenshot
  console.log(await page.content()); // Print HTML
});

# View test artifacts
npx playwright show-report

# Check test results
cat playwright-report.json | jq .
```

### Debugging Backend Tests

```bash
# Run with verbose output
go test -v ./cmd/prism-gui/

# Run specific test
go test -v ./cmd/prism-gui/ -run TestProfileManagement

# Add print statements
import "fmt"

func TestMyFunction(t *testing.T) {
    fmt.Printf("Debug: value = %v\n", value)
    t.Logf("Debug: value = %v", value) // Better - shows in test output
}

# Run with race detector
go test -race ./cmd/prism-gui/

# Run with coverage
go test -coverprofile=coverage.out ./cmd/prism-gui/
go tool cover -html=coverage.out
```

### Common Debugging Patterns

**Component test not finding element**:
```typescript
// Bad
expect(screen.getByText('Hello')).toBeInTheDocument(); // Fails immediately

// Good
await waitFor(() => {
  expect(screen.getByText('Hello')).toBeInTheDocument();
}); // Waits for async rendering
```

**E2E test timing issues**:
```typescript
// Bad
await page.click('button');
expect(await page.textContent('.result')).toBe('Success'); // Race condition

// Good
await page.click('button');
await expect(page.locator('.result')).toHaveText('Success'); // Auto-waits
```

**MSW handler not being called**:
```typescript
// Check MSW is enabled
setupMSW(); // In beforeAll or describe block

// Check URL matches exactly
server.use(
  http.get('http://localhost:8947/api/v1/templates', () => { // Must match exactly
    return HttpResponse.json([...]);
  })
);

// Enable MSW logging
server.listen({ onUnhandledRequest: 'warn' });
```

---

## CI/CD Integration

### GitHub Actions Workflow

Tests run automatically on every push and pull request via `.github/workflows/ci.yml`:

**Jobs**:
1. **gui-tests**: Unit tests, type checking, linting, coverage
2. **gui-e2e-tests**: E2E tests with Playwright

**Configuration**:
```yaml
gui-tests:
  - Setup Node.js 20
  - Install dependencies (npm ci)
  - Run type checking
  - Run linter
  - Run unit tests
  - Upload coverage to Codecov

gui-e2e-tests:
  - Setup Go 1.24 + Node.js 20
  - Build daemon binary
  - Install Playwright browsers
  - Run E2E tests
  - Upload Playwright report and test results
```

### Local CI Testing

```bash
# Run exactly what CI runs
npm ci  # Clean install
npm run typecheck
npm run lint
npm run test:unit
npm run test:e2e

# Build daemon (required for E2E)
cd ../../..
go build -o bin/prismd ./cmd/prismd
cd cmd/prism-gui/frontend
```

### Viewing CI Artifacts

When tests fail in CI:
1. Go to GitHub Actions run
2. Click on failed job
3. Download artifacts:
   - `playwright-report/` - HTML test report
   - `playwright-results/` - Screenshots and videos
   - `coverage/` - Coverage reports

---

## Common Issues

### Issue: Tests fail with "Daemon not available"

**Symptom**: E2E tests fail with connection errors

**Solution**:
```bash
# Build daemon binary
cd /Users/scttfrdmn/src/prism
go build -o bin/prismd ./cmd/prismd

# Verify daemon binary exists
ls -la bin/prismd

# Check daemon starts
./bin/prismd --version
```

### Issue: "Element not found" in component tests

**Symptom**: `TestingLibraryElementError: Unable to find element`

**Solution**:
```typescript
// Use waitFor for async rendering
await waitFor(() => {
  expect(screen.getByText('Hello')).toBeInTheDocument();
});

// Check what's rendered
screen.debug();

// Use more flexible queries
screen.getByText(/hello/i); // Case insensitive regex
```

### Issue: Playwright browser not installed

**Symptom**: `Error: browserType.launch: Executable doesn't exist`

**Solution**:
```bash
# Install Playwright browsers
npx playwright install chromium

# Or install all browsers
npx playwright install
```

### Issue: MSW handlers not working

**Symptom**: Tests still make real API calls

**Solution**:
```typescript
// Ensure setupMSW is called
describe('MyComponent', () => {
  setupMSW(); // Add this!

  test('...', () => {
    // ...
  });
});

// Check MSW is running
server.listen({ onUnhandledRequest: 'warn' }); // Shows unhandled requests
```

### Issue: Tests pass locally but fail in CI

**Symptom**: Green locally, red in CI

**Common Causes**:
1. **Timing differences**: Add more `waitFor` with longer timeouts
2. **Missing dependencies**: Ensure all deps in package.json
3. **Environment differences**: Check Node.js version matches
4. **Daemon not built**: CI builds daemon automatically
5. **Race conditions**: Use Playwright auto-waiting features

**Solution**:
```typescript
// Increase timeout for CI
await waitFor(() => {
  expect(screen.getByText('Hello')).toBeInTheDocument();
}, { timeout: 10000 }); // 10 seconds

// Use retries in E2E
test.use({ retries: 2 }); // Retry failed tests
```

### Issue: Coverage reports incorrect

**Symptom**: Coverage shows 0% or missing files

**Solution**:
```bash
# Ensure coverage is configured in vitest.config.ts
# Run with coverage flag
npm run test:coverage

# View HTML report
open coverage/index.html
```

---

## Best Practices

### Component Testing

1. ✅ **Use MSW for API mocking** - Avoid mocking fetch directly
2. ✅ **Test user behavior, not implementation** - Use `userEvent` over `fireEvent`
3. ✅ **Use accessible queries** - `getByRole`, `getByLabelText` over `getByTestId`
4. ✅ **Wait for async updates** - Always use `waitFor` for async operations
5. ✅ **Test error states** - Use `errorHandlers` from MSW
6. ✅ **Test empty states** - Use `emptyHandlers` from MSW
7. ❌ **Don't test implementation details** - Avoid testing internal state
8. ❌ **Don't use arbitrary waits** - Use `waitFor`, not `setTimeout`

### E2E Testing

1. ✅ **Use real daemon** - Integration tests should use real APIs
2. ✅ **Use Page Object Model** - Abstract page interactions
3. ✅ **Use data-testid sparingly** - Prefer accessible selectors
4. ✅ **Clean up after tests** - Delete created resources
5. ✅ **Test critical paths** - Focus on common user workflows
6. ✅ **Use auto-waiting** - Playwright waits automatically
7. ❌ **Don't use arbitrary waits** - Use `waitForSelector`, not `waitForTimeout`
8. ❌ **Don't leave test data** - Clean up instances, volumes, etc.

### General

1. ✅ **Keep tests isolated** - Each test should be independent
2. ✅ **Use descriptive test names** - Explain what is being tested
3. ✅ **Test one thing per test** - Single assertion focus
4. ✅ **Use test utilities** - Reuse helper functions
5. ✅ **Keep tests fast** - Mock expensive operations
6. ✅ **Write tests first** - TDD when possible
7. ❌ **Don't skip tests** - Fix or remove, don't skip
8. ❌ **Don't test third-party code** - Focus on your code

---

## Additional Resources

- [Vitest Documentation](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Playwright Documentation](https://playwright.dev/)
- [MSW Documentation](https://mswjs.io/)
- [Percy Visual Testing](https://percy.io/)
- [GUI Test Architecture](../../docs/GUI_TESTING_ARCHITECTURE.md)
- [Phase 0 Implementation](../../docs/PHASE_0_IMPLEMENTATION.md)

---

**Questions or Issues?** Create an issue at https://github.com/scttfrdmn/prism/issues with the `gui` and `testing` labels.
