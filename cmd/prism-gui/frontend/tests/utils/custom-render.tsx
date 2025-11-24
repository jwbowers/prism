/**
 * Custom Render Function for Prism GUI Tests
 *
 * Wraps components with necessary providers and context for testing.
 * Use this instead of @testing-library/react's render() directly.
 */

import { render, RenderOptions } from '@testing-library/react';
import { ReactElement, ReactNode } from 'react';

/**
 * Custom render options extending React Testing Library's RenderOptions
 */
interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  /**
   * Initial Wails mock functions
   */
  wailsMocks?: {
    GetTemplates?: () => Promise<any>;
    GetInstances?: () => Promise<any>;
    LaunchInstance?: (...args: any[]) => Promise<any>;
    [key: string]: any;
  };
}

/**
 * Wrapper component that provides all necessary context
 */
function AllTheProviders({ children }: { children: ReactNode }) {
  return (
    <>
      {children}
    </>
  );
}

/**
 * Custom render function that wraps component with providers
 *
 * @example
 * ```tsx
 * import { renderWithProviders } from 'tests/utils/custom-render';
 * import { createMockTemplates } from 'tests/utils/mock-data-factories';
 *
 * test('renders template list', () => {
 *   renderWithProviders(<TemplateList />, {
 *     wailsMocks: {
 *       GetTemplates: async () => createMockTemplates(),
 *     },
 *   });
 *   // ... assertions
 * });
 * ```
 */
export function renderWithProviders(
  ui: ReactElement,
  options: CustomRenderOptions = {}
) {
  const { wailsMocks, ...renderOptions } = options;

  // Set up Wails mocks if provided
  if (wailsMocks) {
    const mockWails = {
      PrismService: {
        ...wailsMocks,
      },
    };

    // @ts-ignore - window.wails is a test mock
    window.wails = mockWails;
  }

  return render(ui, {
    wrapper: AllTheProviders,
    ...renderOptions,
  });
}

/**
 * Re-export everything from React Testing Library
 */
export * from '@testing-library/react';
export { renderWithProviders as render };
