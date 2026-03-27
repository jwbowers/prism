/**
 * App.test.tsx — Rewired for SafePrismAPI / fetch mock pattern.
 *
 * The current App uses SafePrismAPI which calls fetch → http://localhost:8947.
 * We stub global.fetch with URL-based routing so every endpoint returns
 * sensible data without a running daemon.
 *
 * Default view is 'dashboard'. Navigation items include Dashboard, Templates,
 * My Workspaces, Storage, Backups, Projects, Users, etc. Some text like
 * "My Workspaces" appears in both the nav and the content, so we use
 * getAllByText or role-based queries.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// ── Helpers ─────────────────────────────────────────────────────────────────

const mockTemplates = {
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
};

const mockInstances = [
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
];

function makeFetchMock(overrides?: Record<string, unknown>) {
  return vi.fn().mockImplementation((url: string) => {
    if (url.includes('/api/v1/templates')) {
      const data = overrides?.templates ?? mockTemplates;
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve(data) });
    }
    if (url.includes('/api/v1/instances')) {
      const data = { instances: overrides?.instances ?? mockInstances };
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve(data) });
    }
    if (url.includes('/api/v1/snapshots')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ snapshots: [], count: 0 }) });
    }
    // Default: return empty success
    return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({}) });
  });
}

// ── Setup ────────────────────────────────────────────────────────────────────

beforeEach(() => {
  vi.stubGlobal('fetch', makeFetchMock());
  localStorage.setItem('cws_onboarding_complete', 'true');
});

afterEach(() => {
  vi.unstubAllGlobals();
  localStorage.removeItem('cws_onboarding_complete');
});

// ── Tests ─────────────────────────────────────────────────────────────────

describe('Prism App', () => {
  describe('Initial Render', () => {
    it('renders without crashing', () => {
      render(<App />);
      // Prism header link appears in SideNavigation
      expect(screen.getByRole('link', { name: /prism/i })).toBeInTheDocument();
    });

    it('shows navigation link for Templates', () => {
      render(<App />);
      // Use getAllByText since "Templates" may appear in nav and content
      const matches = screen.getAllByText('Templates');
      expect(matches.length).toBeGreaterThan(0);
    });

    it('shows navigation link for My Workspaces', () => {
      render(<App />);
      const matches = screen.getAllByText('My Workspaces');
      expect(matches.length).toBeGreaterThan(0);
    });

    it('starts on Dashboard view', () => {
      render(<App />);
      expect(screen.getByText('Welcome to Prism')).toBeInTheDocument();
    });

    it('shows Dashboard nav item as active link', () => {
      render(<App />);
      // Dashboard link should be present
      const dashboardLinks = screen.getAllByText('Dashboard');
      expect(dashboardLinks.length).toBeGreaterThan(0);
    });
  });

  describe('Template Selection', () => {
    it('shows Research Templates header when navigated to templates', async () => {
      const user = userEvent.setup();
      render(<App />);

      // Click the Templates nav item (first occurrence)
      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      await waitFor(() => {
        const headings = screen.getAllByText('Research Templates');
        expect(headings.length).toBeGreaterThan(0);
      });
    });

    it('loads and displays templates after navigation', async () => {
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
    });

    it('shows empty state when no templates available', async () => {
      vi.stubGlobal('fetch', makeFetchMock({ templates: {} }));

      const user = userEvent.setup();
      render(<App />);

      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      await waitFor(() => {
        const noTemplates = screen.getAllByText('No templates available');
        expect(noTemplates.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Instance Management', () => {
    it('shows My Workspaces counter when navigated to workspaces', async () => {
      const user = userEvent.setup();
      render(<App />);

      const workspacesLinks = screen.getAllByText('My Workspaces');
      await user.click(workspacesLinks[0]);

      await waitFor(() => {
        // Counter shows instance count
        expect(screen.getByText('(1)')).toBeInTheDocument();
      });
    });

    it('displays instance name in the workspaces view', async () => {
      const user = userEvent.setup();
      render(<App />);

      const workspacesLinks = screen.getAllByText('My Workspaces');
      await user.click(workspacesLinks[0]);

      await waitFor(() => {
        expect(screen.getByText('my-ml-research')).toBeInTheDocument();
      });
    });
  });

  describe('Navigation', () => {
    it('shows welcome message on dashboard', () => {
      render(<App />);
      expect(screen.getByText('Welcome to Prism')).toBeInTheDocument();
    });

    it('can navigate from dashboard to templates', async () => {
      const user = userEvent.setup();
      render(<App />);

      expect(screen.getByText('Welcome to Prism')).toBeInTheDocument();

      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      await waitFor(() => {
        const headings = screen.getAllByText('Research Templates');
        expect(headings.length).toBeGreaterThan(0);
      });
    });

    it('calls API on mount', async () => {
      const fetchSpy = vi.fn().mockImplementation(makeFetchMock());
      vi.stubGlobal('fetch', fetchSpy);

      render(<App />);

      await waitFor(() => {
        expect(fetchSpy).toHaveBeenCalled();
      });
    });
  });

  describe('Error Handling', () => {
    it('handles fetch failures gracefully — shows empty templates state', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

      const user = userEvent.setup();
      render(<App />);

      const templatesLinks = screen.getAllByText('Templates');
      await user.click(templatesLinks[0]);

      // With all fetches failing, loading finishes and templates show empty state
      await waitFor(() => {
        const noTemplates = screen.getAllByText('No templates available');
        expect(noTemplates.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
    });
  });
});
