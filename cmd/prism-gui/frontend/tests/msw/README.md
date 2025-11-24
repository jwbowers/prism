# MSW (Mock Service Worker) Setup

This directory contains MSW configuration for mocking the Prism daemon API in tests.

## Files

- **`handlers.ts`**: API endpoint handlers that return mock responses
- **`server.ts`**: Node.js MSW server for unit/component tests (Vitest + jsdom)
- **`browser.ts`**: Browser MSW worker for E2E tests (Playwright)

## Usage

### Component Tests (Vitest)

```typescript
import { describe, test, expect } from 'vitest';
import { render, screen, waitFor } from 'tests/utils';
import { setupMSW } from 'tests/msw/server';
import { TemplateList } from './TemplateList';

describe('TemplateList', () => {
  setupMSW(); // Enable MSW for all tests in this suite

  test('displays templates from API', async () => {
    render(<TemplateList />);

    await waitFor(() => {
      expect(screen.getByText('Python Machine Learning')).toBeInTheDocument();
      expect(screen.getByText('R Research Environment')).toBeInTheDocument();
    });
  });
});
```

### Custom Handlers for Specific Tests

```typescript
import { server } from 'tests/msw/server';
import { http, HttpResponse } from 'msw';

test('handles API error gracefully', async () => {
  // Override handler for this test
  server.use(
    http.get('http://localhost:8947/api/v1/templates', () => {
      return HttpResponse.json({ error: 'Server error' }, { status: 500 });
    })
  );

  render(<TemplateList />);

  await waitFor(() => {
    expect(screen.getByText('Failed to load templates')).toBeInTheDocument();
  });
});
```

### Testing Empty States

```typescript
import { emptyHandlers } from 'tests/msw/handlers';

test('displays empty state when no templates', async () => {
  server.use(...emptyHandlers);

  render(<TemplateList />);

  await waitFor(() => {
    expect(screen.getByText('No templates available')).toBeInTheDocument();
  });
});
```

### Testing Error Scenarios

```typescript
import { errorHandlers } from 'tests/msw/handlers';

test('displays error message on API failure', async () => {
  server.use(...errorHandlers);

  render(<TemplateList />);

  await waitFor(() => {
    expect(screen.getByText('Failed to fetch templates')).toBeInTheDocument();
  });
});
```

### E2E Tests (Playwright)

For E2E tests, we typically use the **real daemon** for integration testing. However, if you need to mock APIs in the browser:

```typescript
import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  // Enable MSW in the browser
  await page.addInitScript(() => {
    import('/tests/msw/browser.js').then(({ worker }) => worker.start());
  });
});

test('template selection with mocked API', async ({ page }) => {
  await page.goto('/');
  // MSW will intercept API calls and return mock data
});
```

## Handler Types

### Default Handlers (`handlers`)
Standard mock responses for all API endpoints. Used in most tests.

### Error Handlers (`errorHandlers`)
Returns error responses for testing error handling.

### Empty Handlers (`emptyHandlers`)
Returns empty arrays for testing empty states.

## API Endpoints Covered

- **Health**: `GET /api/v1/health`
- **Templates**: `GET /api/v1/templates`, `GET /api/v1/templates/:name`
- **Instances**: `GET /api/v1/instances`, `POST /api/v1/instances/launch`, instance actions
- **Profiles**: CRUD operations for profiles
- **Volumes**: EFS volume management
- **Storage**: EBS storage management
- **Idle Policies**: Idle detection and hibernation policies
- **Backups**: Backup and snapshot management

## Customizing Mock Data

All mock data comes from `tests/utils/mock-data-factories.ts`. To customize:

```typescript
import { createMockTemplate } from 'tests/utils/mock-data-factories';

const customTemplate = createMockTemplate({
  Name: 'Custom Template',
  Complexity: 'advanced',
  EstimatedCostPerHour: { 'x86_64': 1.50 },
});
```

## Best Practices

1. **Use MSW for component tests**: Mock the daemon API to test components in isolation
2. **Use real daemon for E2E tests**: Integration tests should test the real API
3. **Reset handlers after each test**: Ensures test isolation
4. **Override handlers per test**: Use `server.use()` to customize responses for specific tests
5. **Test error scenarios**: Use `errorHandlers` to test error handling
6. **Test empty states**: Use `emptyHandlers` to test empty state UI

## Debugging

To see MSW logs:

```typescript
server.listen({ onUnhandledRequest: 'warn' }); // Warns about unhandled requests
```

To inspect intercepted requests:

```typescript
server.events.on('request:start', ({ request }) => {
  console.log('MSW intercepted:', request.method, request.url);
});
```

## Resources

- [MSW Documentation](https://mswjs.io/)
- [MSW with Vitest](https://mswjs.io/docs/integrations/node)
- [MSW with Playwright](https://mswjs.io/docs/integrations/browser)
