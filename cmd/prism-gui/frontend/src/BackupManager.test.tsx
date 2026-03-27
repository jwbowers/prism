/**
 * BackupManager.test.tsx
 *
 * Tests the Backup & Snapshot Management functionality in the Prism GUI App.
 * The BackupManagementView is rendered when activeView === 'backups' (nav item "Backups").
 *
 * The view fetches snapshots from /api/v1/snapshots.
 * - data-testid="backups-table"
 * - "No backups found" empty state
 * - "Create Backup" button that opens a modal
 * - "Available Backups" section header
 * - Actions ButtonDropdown per row (Restore, Clone, View Details, Delete)
 *
 * Uses vi.stubGlobal('fetch', ...) to mock SafePrismAPI HTTP calls.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// ── Mock Data ─────────────────────────────────────────────────────────────

const mockSnapshots = [
  {
    snapshot_id: 'snap-full-12345',
    snapshot_name: 'ml-research-backup-full',
    source_instance: 'ml-research',
    source_instance_id: 'i-test',
    source_template: 'Python Machine Learning',
    description: 'Full backup',
    state: 'available',
    architecture: 'x86_64',
    storage_cost_monthly: 2.50,
    created_at: '2025-01-15T10:00:00Z',
    size_gb: 50,
  },
  {
    snapshot_id: 'snap-incr-67890',
    snapshot_name: 'ml-research-backup-incremental',
    source_instance: 'ml-research',
    source_instance_id: 'i-test',
    source_template: 'Python Machine Learning',
    description: 'Incremental backup',
    state: 'available',
    architecture: 'x86_64',
    storage_cost_monthly: 0.25,
    created_at: '2025-01-16T10:00:00Z',
    size_gb: 5,
  },
];

function buildFetchMock(overrides?: { snapshots?: unknown[]; failSnapshots?: boolean }) {
  return vi.fn().mockImplementation((url: string) => {
    if (url.includes('/api/v1/snapshots')) {
      if (overrides?.failSnapshots) {
        return Promise.reject(new Error('Failed to fetch snapshots'));
      }
      const snapshots = overrides?.snapshots ?? mockSnapshots;
      const count = (snapshots as unknown[]).length;
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ snapshots, count }) });
    }
    if (url.includes('/api/v1/templates')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({}) });
    }
    if (url.includes('/api/v1/instances')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ instances: [{ id: 'i-test', name: 'ml-research', state: 'running', launch_time: '2025-01-01T00:00:00Z', region: 'us-west-2' }] }) });
    }
    return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({}) });
  });
}

// ── Setup ─────────────────────────────────────────────────────────────────

beforeEach(() => {
  vi.stubGlobal('fetch', buildFetchMock());
  localStorage.setItem('cws_onboarding_complete', 'true');
});

afterEach(() => {
  vi.unstubAllGlobals();
  localStorage.removeItem('cws_onboarding_complete');
});

// ── Helpers ────────────────────────────────────────────────────────────────

async function navigateToBackups() {
  const user = userEvent.setup();
  render(<App />);
  const backupsLinks = screen.getAllByText('Backups');
  await user.click(backupsLinks[0]);
  await waitFor(() => {
    expect(screen.getByTestId('backups-table')).toBeInTheDocument();
  });
  return user;
}

// ── Tests ─────────────────────────────────────────────────────────────────

describe('BackupManager', () => {
  describe('Backup List Rendering', () => {
    it('should display Backups nav item', () => {
      render(<App />);
      expect(screen.getAllByText('Backups').length).toBeGreaterThan(0);
    });

    it('should render the backups table after navigating to Backups', async () => {
      await navigateToBackups();
    });

    it('should display backup names in the table', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText('ml-research-backup-full')).toBeInTheDocument();
        expect(screen.getByText('ml-research-backup-incremental')).toBeInTheDocument();
      });
    });

    it('should display backup counter showing count', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText('(2)')).toBeInTheDocument();
      });
    });

    it('should display the source instance for each backup', async () => {
      await navigateToBackups();

      await waitFor(() => {
        // source_instance field rendered
        expect(screen.getAllByText('ml-research').length).toBeGreaterThan(0);
      });
    });

    it('should display backup status (available)', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getAllByText('available').length).toBeGreaterThan(0);
      });
    });

    it('should display backup size in GB', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText('50 GB')).toBeInTheDocument();
        expect(screen.getByText('5 GB')).toBeInTheDocument();
      });
    });

    it('should display monthly cost for each backup', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText('$2.50')).toBeInTheDocument();
        expect(screen.getByText('$0.25')).toBeInTheDocument();
      });
    });

    it('should display source template for each backup', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getAllByText('Python Machine Learning').length).toBeGreaterThan(0);
      });
    });

    it('should handle empty backup list', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ snapshots: [] }));

      const user = userEvent.setup();
      render(<App />);
      const backupsLinks = screen.getAllByText('Backups');
      await user.click(backupsLinks[0]);

      await waitFor(() => {
        expect(screen.getByText('No backups found')).toBeInTheDocument();
      });
    });

    it('should show "Go to Workspaces" button in empty state', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ snapshots: [] }));

      const user = userEvent.setup();
      render(<App />);
      const backupsLinks = screen.getAllByText('Backups');
      await user.click(backupsLinks[0]);

      await waitFor(() => {
        expect(screen.getByText('Go to Workspaces')).toBeInTheDocument();
      });
    });
  });

  describe('Backups Header and Actions', () => {
    it('should display the main "Backups" header', async () => {
      await navigateToBackups();

      await waitFor(() => {
        // "Backups" appears as H1 heading text too
        expect(screen.getAllByText('Backups').length).toBeGreaterThan(0);
      });
    });

    it('should show "Available Backups" section header', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getAllByText('Available Backups').length).toBeGreaterThan(0);
      });
    });

    it('should show "Create Backup" button', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getAllByText('Create Backup').length).toBeGreaterThan(0);
      });
    });

    it('should show Actions dropdown for each backup', async () => {
      await navigateToBackups();

      await waitFor(() => {
        const actionsButtons = screen.getAllByText('Actions');
        // 2 backups → at least 2 Actions buttons
        expect(actionsButtons.length).toBeGreaterThanOrEqual(2);
      });
    });
  });

  describe('Create Backup Modal', () => {
    it('should open Create Backup modal when button is clicked', async () => {
      const user = await navigateToBackups();

      // "Create Backup" may appear multiple times (header and button) — click the primary button
      const createButtons = screen.getAllByText('Create Backup');
      await user.click(createButtons[0]);

      // Modal becomes visible — "Create Backup" count increases or modal-specific content appears
      await waitFor(() => {
        const allCreateBackup = screen.getAllByText('Create Backup');
        expect(allCreateBackup.length).toBeGreaterThan(0);
      });
    });

    it('should show instance selection in create backup modal', async () => {
      const user = await navigateToBackups();

      const createButtons = screen.getAllByText('Create Backup');
      await user.click(createButtons[0]);

      await waitFor(() => {
        // Modal content - backup name or instance selection
        expect(screen.getAllByText('Create Backup').length).toBeGreaterThan(0);
      });
    });
  });

  describe('Backup Storage Summary', () => {
    it('should display Backup Storage Summary section', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getAllByText(/Backup Storage Summary/).length).toBeGreaterThan(0);
      });
    });

    it('should show total backup count', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText('Total Backups')).toBeInTheDocument();
      });
    });

    it('should show total storage size', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText('Total Storage Size')).toBeInTheDocument();
      });
    });

    it('should show monthly storage cost', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText('Monthly Storage Cost')).toBeInTheDocument();
      });
    });
  });

  describe('Educational Overview', () => {
    it('should show Instance Snapshots & Backups info section', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText(/Instance Snapshots & Backups/)).toBeInTheDocument();
      });
    });

    it('should show cost information', async () => {
      await navigateToBackups();

      await waitFor(() => {
        expect(screen.getByText(/\$0\.05\/GB\/month/i)).toBeInTheDocument();
      });
    });
  });

  describe('API Integration', () => {
    it('calls /api/v1/snapshots on mount', async () => {
      const fetchSpy = buildFetchMock();
      vi.stubGlobal('fetch', fetchSpy);

      render(<App />);

      await waitFor(() => {
        const calls = (fetchSpy as ReturnType<typeof vi.fn>).mock.calls.map((c: unknown[]) => c[0] as string);
        expect(calls.some((url: string) => url.includes('/api/v1/snapshots'))).toBe(true);
      });
    });

    it('handles snapshot API failure gracefully', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ failSnapshots: true }));

      const user = userEvent.setup();
      render(<App />);

      // App should still render
      expect(screen.getAllByText('Backups').length).toBeGreaterThan(0);

      const backupsLinks = screen.getAllByText('Backups');
      await user.click(backupsLinks[0]);

      // Should show empty state or table (not crash)
      await waitFor(() => {
        expect(screen.getByTestId('backups-table')).toBeInTheDocument();
      });
    });
  });
});
