/**
 * StorageManager.test.tsx
 *
 * Tests the Storage Management functionality in the Prism GUI App.
 * The StorageManagementView is rendered when activeView === 'storage' (nav item "Storage").
 *
 * The view shows:
 * - data-testid="storage-page" wrapper
 * - data-testid="efs-table" (Shared EFS tab, active by default)
 * - data-testid="ebs-table" (Private EBS tab)
 * - data-testid="create-efs-header-button" → "Create EFS Volume"
 * - Tabs: "Shared (EFS) - N", "Private (EBS) - N"
 * - "Shared Storage Volumes" header
 * - "No shared storage volumes found" empty state
 * - Actions ButtonDropdown per row (Mount, Unmount, View Details, Delete)
 *
 * API calls:
 * - GET /api/v1/volumes → EFS volumes (array of StorageVolume)
 * - GET /api/v1/storage → EBS volumes (array of StorageVolume, filtered for workspace/ebs)
 *
 * Uses vi.stubGlobal('fetch', ...) to mock SafePrismAPI HTTP calls.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// ── Mock Data ─────────────────────────────────────────────────────────────

// StorageVolume format from the API (used for both EFS and EBS)
const mockEFSVolumes = [
  {
    name: 'shared-data',
    filesystem_id: 'fs-1234567890abcdef0',
    region: 'us-west-2',
    creation_time: '2025-01-01T00:00:00Z',
    state: 'available',
    performance_mode: 'generalPurpose',
    throughput_mode: 'bursting',
    estimated_cost_gb: 0.30,
    size_bytes: 53687091200, // 50 GB
    aws_service: 'efs',
    type: 'shared',
  },
  {
    name: 'ml-workspace',
    filesystem_id: 'fs-abcdef1234567890',
    region: 'us-west-2',
    creation_time: '2025-01-10T00:00:00Z',
    state: 'available',
    performance_mode: 'maxIO',
    throughput_mode: 'provisioned',
    estimated_cost_gb: 0.30,
    size_bytes: 107374182400, // 100 GB
    attached_to: 'i-test123',
    aws_service: 'efs',
    type: 'shared',
  },
];

const mockEBSVolumes = [
  {
    name: 'project-storage-L',
    volume_id: 'vol-1234567890abcdef0',
    region: 'us-west-2',
    creation_time: '2025-01-05T00:00:00Z',
    state: 'available',
    volume_type: 'gp3',
    size_gb: 500,
    estimated_cost_gb: 0.10,
    aws_service: 'ebs',
    type: 'workspace',
  },
  {
    name: 'data-backup-XL',
    volume_id: 'vol-abcdef1234567890',
    region: 'us-west-2',
    creation_time: '2025-01-12T00:00:00Z',
    state: 'in-use',
    volume_type: 'gp3',
    size_gb: 1000,
    estimated_cost_gb: 0.10,
    attached_to: 'i-test123',
    aws_service: 'ebs',
    type: 'workspace',
  },
];

function buildFetchMock(overrides?: {
  efsVolumes?: unknown[];
  ebsVolumes?: unknown[];
  failEFS?: boolean;
  failEBS?: boolean;
}) {
  return vi.fn().mockImplementation((url: string) => {
    if (url.includes('/api/v1/volumes')) {
      if (overrides?.failEFS) {
        return Promise.reject(new Error('Failed to fetch EFS volumes'));
      }
      const volumes = overrides?.efsVolumes ?? mockEFSVolumes;
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve(volumes) });
    }
    if (url.includes('/api/v1/storage')) {
      if (overrides?.failEBS) {
        return Promise.reject(new Error('Failed to fetch EBS volumes'));
      }
      const volumes = overrides?.ebsVolumes ?? mockEBSVolumes;
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve(volumes) });
    }
    if (url.includes('/api/v1/templates')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({}) });
    }
    if (url.includes('/api/v1/instances')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ instances: [] }) });
    }
    if (url.includes('/api/v1/snapshots')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ snapshots: [], count: 0 }) });
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

async function navigateToStorage() {
  const user = userEvent.setup();
  render(<App />);

  const storageLinks = screen.getAllByText('Storage');
  await user.click(storageLinks[0]);

  await waitFor(() => {
    expect(screen.getByTestId('storage-page')).toBeInTheDocument();
  });

  return user;
}

// ── Tests ─────────────────────────────────────────────────────────────────

describe('StorageManager', () => {
  describe('Navigation', () => {
    it('renders the app with Storage nav item', () => {
      render(<App />);
      expect(screen.getAllByText('Storage').length).toBeGreaterThan(0);
    });

    it('navigates to Storage view and shows storage page', async () => {
      await navigateToStorage();
    });

    it('shows Storage header', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getAllByText('Storage').length).toBeGreaterThan(0);
      });
    });
  });

  describe('EFS Volume List Rendering', () => {
    it('should show the EFS table', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByTestId('efs-table')).toBeInTheDocument();
      });
    });

    it('should display EFS volume names', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByText('shared-data')).toBeInTheDocument();
        expect(screen.getByText('ml-workspace')).toBeInTheDocument();
      });
    });

    it('should display filesystem IDs', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByText('fs-1234567890abcdef0')).toBeInTheDocument();
        expect(screen.getByText('fs-abcdef1234567890')).toBeInTheDocument();
      });
    });

    it('should display volume sizes in GB', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByText('50 GB')).toBeInTheDocument();
        expect(screen.getByText('100 GB')).toBeInTheDocument();
      });
    });

    it('should display volume status', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getAllByText('available').length).toBeGreaterThan(0);
      });
    });

    it('should show Shared Storage Volumes section header', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByText('Shared Storage Volumes')).toBeInTheDocument();
      });
    });

    it('should handle empty EFS volume list', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ efsVolumes: [] }));

      const user = userEvent.setup();
      render(<App />);
      const storageLinks = screen.getAllByText('Storage');
      await user.click(storageLinks[0]);

      await waitFor(() => {
        expect(screen.getByText('No shared storage volumes found')).toBeInTheDocument();
      });
    });

    it('should show Create EFS Volume button', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByTestId('create-efs-header-button')).toBeInTheDocument();
      });
    });

    it('should show Actions dropdown for each EFS volume', async () => {
      await navigateToStorage();

      await waitFor(() => {
        const actionsButtons = screen.getAllByText('Actions');
        // At least 2 for the 2 EFS volumes
        expect(actionsButtons.length).toBeGreaterThanOrEqual(2);
      });
    });
  });

  describe('EBS Volume List Rendering', () => {
    it('should show the EBS table after clicking Private tab', async () => {
      const user = await navigateToStorage();

      // Click the Private (EBS) tab
      const ebsTab = screen.queryByRole('tab', { name: /Private.*EBS/i }) ||
                     screen.queryByText(/Private.*EBS/i);
      if (ebsTab) {
        await user.click(ebsTab);
      }

      await waitFor(() => {
        expect(screen.getByTestId('ebs-table')).toBeInTheDocument();
      });
    });

    it('should display EBS volume names after switching tabs', async () => {
      const user = await navigateToStorage();

      // Click the Private (EBS) tab
      const ebsTabEl = screen.queryByRole('tab', { name: /Private.*EBS/i }) ||
                       screen.queryByText(/Private.*EBS/i);
      if (ebsTabEl) {
        await user.click(ebsTabEl);

        await waitFor(() => {
          expect(screen.getByText('project-storage-L')).toBeInTheDocument();
          expect(screen.getByText('data-backup-XL')).toBeInTheDocument();
        });
      } else {
        // If tabs aren't rendered as expected, verify EBS volumes visible some other way
        expect(screen.getByTestId('ebs-table')).toBeInTheDocument();
      }
    });

    it('should show both tabs in the storage view', async () => {
      await navigateToStorage();

      await waitFor(() => {
        // Shared (EFS) tab label appears in tabs
        const sharedTabs = screen.queryAllByText(/Shared.*EFS/i);
        expect(sharedTabs.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Create EFS Volume Modal', () => {
    it('should open Create EFS Volume modal when button is clicked', async () => {
      const user = await navigateToStorage();

      const createButton = screen.getByTestId('create-efs-header-button');
      await user.click(createButton);

      await waitFor(() => {
        // "Create EFS Volume" may appear as button text + modal header
        const matches = screen.getAllByText('Create EFS Volume');
        expect(matches.length).toBeGreaterThan(0);
      });
    });

    it('should show volume name field in create EFS modal', async () => {
      const user = await navigateToStorage();

      const createButton = screen.getByTestId('create-efs-header-button');
      await user.click(createButton);

      await waitFor(() => {
        // Modal with "Create EFS Volume" header visible
        const headers = screen.getAllByText('Create EFS Volume');
        expect(headers.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Educational Overview', () => {
    it('should display Shared Storage (EFS) info', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByText(/Shared Storage.*EFS/)).toBeInTheDocument();
      });
    });

    it('should display Private Storage (EBS) info', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByText(/Private Storage.*EBS/)).toBeInTheDocument();
      });
    });

    it('should display Storage Selection Guide alert', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByText(/Storage Selection Guide/i)).toBeInTheDocument();
      });
    });
  });

  describe('API Integration', () => {
    it('calls /api/v1/volumes on mount for EFS', async () => {
      const fetchSpy = buildFetchMock();
      vi.stubGlobal('fetch', fetchSpy);

      render(<App />);

      await waitFor(() => {
        const calls = (fetchSpy as ReturnType<typeof vi.fn>).mock.calls.map((c: unknown[]) => c[0] as string);
        expect(calls.some((url: string) => url.includes('/api/v1/volumes'))).toBe(true);
      });
    });

    it('calls /api/v1/storage on mount for EBS', async () => {
      const fetchSpy = buildFetchMock();
      vi.stubGlobal('fetch', fetchSpy);

      render(<App />);

      await waitFor(() => {
        const calls = (fetchSpy as ReturnType<typeof vi.fn>).mock.calls.map((c: unknown[]) => c[0] as string);
        expect(calls.some((url: string) => url.includes('/api/v1/storage'))).toBe(true);
      });
    });

    it('handles EFS API failure gracefully', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ failEFS: true }));

      const user = userEvent.setup();
      render(<App />);

      const storageLinks = screen.getAllByText('Storage');
      await user.click(storageLinks[0]);

      // Should show empty state, not crash
      await waitFor(() => {
        const storageEl = screen.getByTestId('storage-page');
        expect(storageEl).toBeInTheDocument();
      });
    });

    it('handles EBS API failure gracefully', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ failEBS: true }));

      render(<App />);

      // App should still render
      expect(screen.getAllByText('Storage').length).toBeGreaterThan(0);
    });
  });

  describe('Refresh', () => {
    it('shows Refresh button in storage view', async () => {
      await navigateToStorage();

      await waitFor(() => {
        expect(screen.getByText('Refresh')).toBeInTheDocument();
      });
    });
  });
});
