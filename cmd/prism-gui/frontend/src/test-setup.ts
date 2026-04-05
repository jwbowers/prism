import { afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';
import '@testing-library/jest-dom/vitest';

// Cleanup after each test case (e.g. clearing jsdom)
afterEach(() => {
  cleanup();
});

// Global React act() warning handler for test environment
(globalThis as Record<string, unknown>).IS_REACT_ACT_ENVIRONMENT = true;

// Mock window.matchMedia
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

// Mock ResizeObserver
(globalThis as typeof globalThis & { ResizeObserver: unknown }).ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

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

// Suppress non-critical test noise
const originalConsoleError = console.error;
const originalConsoleWarn = console.warn;

console.error = (...args) => {
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
  originalConsoleError.apply(console, args);
};

console.warn = (...args) => {
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
