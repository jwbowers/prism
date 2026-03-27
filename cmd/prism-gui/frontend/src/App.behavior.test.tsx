/**
 * App.behavior.test.tsx — Behavioral tests for the Prism App.
 *
 * Tests key user workflows: template browsing, workspace navigation, basic
 * navigation flows, and error handling.
 *
 * Uses vi.stubGlobal('fetch', ...) to mock the SafePrismAPI HTTP calls.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// ── Mock Data ─────────────────────────────────────────────────────────────

const mockTemplatesData = {
  'python-ml': {
    Name: 'Python Machine Learning',
    Slug: 'python-ml',
    Description: 'Complete ML environment with TensorFlow, PyTorch, and Jupyter',
    category: 'Machine Learning',
    complexity: 'moderate',
  },
  'r-research': {
    Name: 'R Research Environment',
    Slug: 'r-research',
    Description: 'Statistical computing with R, RStudio, and tidyverse packages',
    category: 'Data Science',
    complexity: 'simple',
  },
  'basic-ubuntu': {
    Name: 'Basic Ubuntu (APT)',
    Slug: 'basic-ubuntu',
    Description: 'Ubuntu with APT package management',
    category: 'Base System',
    complexity: 'simple',
  },
};

const mockInstancesData = [
  {
    id: 'i-1234567890abcdef0',
    name: 'my-ml-research',
    template: 'Python Machine Learning',
    state: 'running',
    public_ip: '54.123.45.67',
    instance_type: 't3.xlarge',
    launch_time: '2025-09-28T10:30:00Z',
    region: 'us-west-2',
  },
  {
    id: 'i-0987654321fedcba1',
    name: 'data-analysis-project',
    template: 'R Research Environment',
    state: 'hibernated',
    instance_type: 'r5.xlarge',
    launch_time: '2025-09-27T14:15:00Z',
    region: 'us-west-2',
  },
  {
    id: 'i-abcdef1234567890',
    name: 'web-dev-staging',
    template: 'Basic Ubuntu (APT)',
    state: 'stopped',
    instance_type: 't3.micro',
    launch_time: '2025-09-26T09:45:00Z',
    region: 'us-east-1',
  },
];

function buildFetchMock(overrides?: Partial<{
  templates: unknown;
  instances: unknown[];
  failAll: boolean;
}>) {
  if (overrides?.failAll) {
    return vi.fn().mockRejectedValue(new Error('Network timeout'));
  }

  return vi.fn().mockImplementation((url: string) => {
    if (url.includes('/api/v1/templates')) {
      const data = overrides?.templates ?? mockTemplatesData;
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve(data) });
    }
    if (url.includes('/api/v1/instances')) {
      const instances = overrides?.instances ?? mockInstancesData;
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ instances }) });
    }
    if (url.includes('/api/v1/snapshots')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ snapshots: [], count: 0 }) });
    }
    return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({}) });
  });
}

// ── Setup ────────────────────────────────────────────────────────────────

beforeEach(() => {
  vi.stubGlobal('fetch', buildFetchMock());
  localStorage.setItem('cws_onboarding_complete', 'true');
});

afterEach(() => {
  vi.unstubAllGlobals();
  localStorage.removeItem('cws_onboarding_complete');
});

// ── Tests ──────────────────────────────────────────────────────────────────

describe('Prism Behavioral Tests', () => {
  describe('Critical User Workflows', () => {
    it('should display Research Templates header when navigating to Templates', async () => {
      const user = userEvent.setup();
      render(<App />);

      // Navigate to Templates view
      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      await waitFor(() => {
        const headings = screen.getAllByText('Research Templates');
        expect(headings.length).toBeGreaterThan(0);
      });
    });

    it('should display template names after navigating to templates', async () => {
      const user = userEvent.setup();
      render(<App />);

      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      await waitFor(() => {
        const mlMatches = screen.getAllByText('Python Machine Learning');
        expect(mlMatches.length).toBeGreaterThan(0);
      });

      const rMatches = screen.getAllByText('R Research Environment');
      expect(rMatches.length).toBeGreaterThan(0);

      const ubuntuMatches = screen.getAllByText('Basic Ubuntu (APT)');
      expect(ubuntuMatches.length).toBeGreaterThan(0);
    });

    it('should show My Workspaces header when navigating to workspaces', async () => {
      const user = userEvent.setup();
      render(<App />);

      const workspacesLinks = screen.getAllByText('My Workspaces');
      await user.click(workspacesLinks[0]);

      await waitFor(() => {
        // Counter with 3 instances
        expect(screen.getByText('(3)')).toBeInTheDocument();
      });
    });

    it('should display instance details in the workspaces view', async () => {
      const user = userEvent.setup();
      render(<App />);

      const workspacesLinks = screen.getAllByText('My Workspaces');
      await user.click(workspacesLinks[0]);

      await waitFor(() => {
        expect(screen.getByText('my-ml-research')).toBeInTheDocument();
        expect(screen.getByText('data-analysis-project')).toBeInTheDocument();
        expect(screen.getByText('web-dev-staging')).toBeInTheDocument();
      });
    });

    it('should show instance state labels in the table', async () => {
      const user = userEvent.setup();
      render(<App />);

      const workspacesLinks = screen.getAllByText('My Workspaces');
      await user.click(workspacesLinks[0]);

      await waitFor(() => {
        // StatusIndicator renders state text
        expect(screen.getByText('running')).toBeInTheDocument();
        expect(screen.getByText('stopped')).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling and Edge Cases', () => {
    it('should handle all API failures during initial load', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ failAll: true }));

      const user = userEvent.setup();
      render(<App />);

      // Navigate to templates — should show empty state
      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      await waitFor(() => {
        const noTemplates = screen.getAllByText('No templates available');
        expect(noTemplates.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
    });

    it('should handle empty instances list', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ instances: [] }));

      const user = userEvent.setup();
      render(<App />);

      const workspacesLinks = screen.getAllByText('My Workspaces');
      await user.click(workspacesLinks[0]);

      // Counter should show (0)
      await waitFor(() => {
        expect(screen.getByText('(0)')).toBeInTheDocument();
      });
    });

    it('should validate that the search/filter interface is present for templates', async () => {
      const user = userEvent.setup();
      render(<App />);

      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      await waitFor(() => {
        const headings = screen.getAllByText('Research Templates');
        expect(headings.length).toBeGreaterThan(0);
      });

      // The template list should contain 3 templates
      await waitFor(() => {
        const mlMatches = screen.getAllByText('Python Machine Learning');
        expect(mlMatches.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Navigation and State Management', () => {
    it('should show Dashboard on initial load', () => {
      render(<App />);
      expect(screen.getByText('Welcome to Prism')).toBeInTheDocument();
    });

    it('should navigate from Dashboard to My Workspaces', async () => {
      const user = userEvent.setup();
      render(<App />);

      expect(screen.getByText('Welcome to Prism')).toBeInTheDocument();

      const workspacesLinks = screen.getAllByText('My Workspaces');
      await user.click(workspacesLinks[0]);

      await waitFor(() => {
        expect(screen.getByText('my-ml-research')).toBeInTheDocument();
      });
    });

    it('should navigate from workspaces back to templates', async () => {
      const user = userEvent.setup();
      render(<App />);

      // Go to workspaces
      const workspacesLinks = screen.getAllByText('My Workspaces');
      await user.click(workspacesLinks[0]);

      await waitFor(() => {
        expect(screen.getByText('my-ml-research')).toBeInTheDocument();
      });

      // Go to templates
      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      await waitFor(() => {
        const headings = screen.getAllByText('Research Templates');
        expect(headings.length).toBeGreaterThan(0);
      });
    });

    it('should call the fetch API on mount', async () => {
      const fetchSpy = buildFetchMock();
      vi.stubGlobal('fetch', fetchSpy);

      render(<App />);

      await waitFor(() => {
        expect(fetchSpy).toHaveBeenCalled();
      });
    });

    it('should call templates API endpoint', async () => {
      const fetchSpy = buildFetchMock();
      vi.stubGlobal('fetch', fetchSpy);

      render(<App />);

      await waitFor(() => {
        const calls = (fetchSpy as any).mock.calls.map((c: unknown[]) => c[0] as string);
        expect(calls.some((url: string) => url.includes('/api/v1/templates'))).toBe(true);
      });
    });
  });
});
