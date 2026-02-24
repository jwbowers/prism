import { afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';
import '@testing-library/jest-dom/vitest';
import * as lodash from 'lodash';

// Make lodash available globally for Cloudscape components
(globalThis as any)._ = lodash;

// Cleanup after each test case (e.g. clearing jsdom)
afterEach(() => {
  cleanup();
});

// Global React act() warning handler for test environment
(globalThis as Record<string, unknown>).IS_REACT_ACT_ENVIRONMENT = true;

// Mock window.matchMedia for Cloudscape components
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => {},
  }),
});

// Mock ResizeObserver for Cloudscape components
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock CSS.supports and CSS.escape for Cloudscape
Object.defineProperty(window, 'CSS', {
  value: {
    supports: () => false,
    escape: (str: string) => {
      // Simple CSS.escape polyfill for tests
      // Escapes special characters that need escaping in CSS selectors
      return str.replace(/[!"#$%&'()*+,.\/:;<=>?@[\\\]^`{|}~]/g, '\\$&');
    },
  },
  writable: true
});

// Mock window.getComputedStyle
Object.defineProperty(window, 'getComputedStyle', {
  value: () => ({
    getPropertyValue: () => '',
  }),
});

// Mock scroll APIs
Object.defineProperty(Element.prototype, 'scrollIntoView', {
  value: () => {},
});

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};

  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString();
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
});

// Suppress CSS parsing errors and other test noise
const originalConsoleError = console.error;
const originalConsoleWarn = console.warn;

console.error = (...args) => {
  // Suppress specific errors that don't affect functionality
  if (
    args.length > 0 &&
    typeof args[0] === 'string' &&
    (args[0].includes('\\8 and \\9 are not allowed in strict mode') ||
     args[0].includes('CSS parsing error') ||
     args[0].includes('Could not parse CSS stylesheet') ||
     args[0].includes('nwsapi') ||
     args[0].includes('Not implemented: HTMLCanvasElement'))
  ) {
    return;
  }
  // Suppress React key warnings from Cloudscape internals (React 19 + Cloudscape compatibility issue)
  // These fire from Cloudscape's ForwardRef components (SpaceBetween, ColumnLayout, Grid) and are
  // not from our code. They don't affect functionality. Tracked: cloudscape-design/components#3xxx
  if (args.length > 0 && typeof args[0] === 'string' && args[0].includes('Each child in a list')) {
    return;
  }
  originalConsoleError.apply(console, args);
};

console.warn = (...args) => {
  // Suppress non-critical warnings
  if (
    args.length > 0 &&
    typeof args[0] === 'string' &&
    (args[0].includes('Could not parse CSS stylesheet') ||
     args[0].includes('Not implemented: HTMLCanvasElement'))
  ) {
    return;
  }
  originalConsoleWarn.apply(console, args);
};