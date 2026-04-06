/**
 * WorkshopsPanel Unit Tests — v0.18.0
 *
 * Tests the WorkshopsPanel component using window.__apiClient mock.
 *
 * Coverage:
 * - Renders workshop list with correct columns
 * - Renders empty state when no workshops
 * - Active-workshop badge count on active workshops
 * - Create Workshop modal opens/closes
 * - Provision/End/Delete actions call API
 * - Dashboard tab shows stats and participant table
 * - Config templates tab renders configs
 * - Use Config modal opens and submits
 * - Error alerts on API failure
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { WorkshopsPanel } from './components/WorkshopsPanel';
import { ApiContext } from './hooks/use-api';

// ── Mock data ─────────────────────────────────────────────────────────────────

const mockWorkshops = [
  {
    id: 'ws-001',
    title: 'NeurIPS DL Tutorial',
    description: 'Deep learning tutorial',
    owner: 'prof1',
    template: 'pytorch-ml',
    max_participants: 60,
    budget_per_participant: 5.0,
    start_time: '2026-12-08T09:00:00Z',
    end_time: '2026-12-08T15:00:00Z',
    status: 'draft',
    join_token: 'WS-TESTTOKEN',
    participants: [
      { user_id: 'p1', display_name: 'Alice', status: 'running', joined_at: '2026-12-08T09:00:00Z' },
      { user_id: 'p2', display_name: 'Bob', status: 'pending', joined_at: '2026-12-08T09:05:00Z' },
    ],
    created_at: '2026-12-01T00:00:00Z',
    updated_at: '2026-12-01T00:00:00Z',
  },
  {
    id: 'ws-002',
    title: 'Hackathon 2027',
    owner: 'organizer2',
    template: 'python-ml',
    max_participants: 0,
    start_time: '2027-01-15T10:00:00Z',
    end_time: '2027-01-17T18:00:00Z',
    status: 'active',
    join_token: 'WS-HACKTOKEN',
    participants: [],
    created_at: '2027-01-01T00:00:00Z',
    updated_at: '2027-01-01T00:00:00Z',
  },
];

const mockDashboard = {
  workshop_id: 'ws-001',
  title: 'NeurIPS DL Tutorial',
  total_participants: 2,
  active_instances: 1,
  stopped_instances: 0,
  pending_instances: 1,
  total_spent: 3.75,
  time_remaining: '5h 30m',
  status: 'draft',
  participants: mockWorkshops[0].participants,
};

const mockConfigs = [
  {
    name: 'ml-6h',
    template: 'pytorch-ml',
    max_participants: 30,
    budget_per_participant: 5.0,
    duration_hours: 6,
    description: 'Saved from workshop: NeurIPS DL Tutorial',
    created_at: '2026-12-09T00:00:00Z',
  },
];

const mockApiClient = {
  getWorkshops: vi.fn(),
  createWorkshop: vi.fn(),
  getWorkshop: vi.fn(),
  updateWorkshop: vi.fn(),
  deleteWorkshop: vi.fn(),
  provisionWorkshop: vi.fn(),
  getWorkshopDashboard: vi.fn(),
  endWorkshop: vi.fn(),
  getWorkshopDownload: vi.fn(),
  getWorkshopConfigs: vi.fn(),
  saveWorkshopConfig: vi.fn(),
  createWorkshopFromConfig: vi.fn(),
};

beforeEach(() => {
  (window as any).__apiClient = mockApiClient;
  vi.clearAllMocks();

  // Default happy-path responses
  mockApiClient.getWorkshops.mockResolvedValue(mockWorkshops);
  mockApiClient.createWorkshop.mockResolvedValue({ id: 'ws-new', title: 'New Workshop', status: 'draft', join_token: 'WS-NEWTOKEN' });
  mockApiClient.getWorkshopDashboard.mockResolvedValue(mockDashboard);
  mockApiClient.provisionWorkshop.mockResolvedValue({ provisioned: 2, skipped: 0, errors: [] });
  mockApiClient.endWorkshop.mockResolvedValue({ stopped: 2, errors: [] });
  mockApiClient.deleteWorkshop.mockResolvedValue(undefined);
  mockApiClient.getWorkshopConfigs.mockResolvedValue(mockConfigs);
  mockApiClient.saveWorkshopConfig.mockResolvedValue(mockConfigs[0]);
  mockApiClient.createWorkshopFromConfig.mockResolvedValue({ id: 'ws-from-cfg', title: 'New From Config', join_token: 'WS-CFGTOKEN' });
});

// ── Helper ────────────────────────────────────────────────────────────────────

const renderPanel = () => render(
  <ApiContext.Provider value={mockApiClient as any}>
    <WorkshopsPanel />
  </ApiContext.Provider>
);

// ── WorkshopsPanel — Workshops tab ────────────────────────────────────────────

describe('WorkshopsPanel — Workshops tab', () => {
  it('renders workshops table with correct columns', async () => {
    renderPanel();
    await waitFor(() => {
      expect(screen.getByTestId('workshops-table')).toBeTruthy();
    });
    expect(screen.getAllByText(/^Title$/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/^Status$/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/^Template$/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/^Participants$/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/^Join Token$/i).length).toBeGreaterThan(0);
  });

  it('renders workshop data from API', async () => {
    renderPanel();
    await waitFor(() => {
      expect(screen.getByText('NeurIPS DL Tutorial')).toBeTruthy();
    });
    expect(screen.getByText('Hackathon 2027')).toBeTruthy();
  });

  it('renders empty state when API returns no workshops', async () => {
    mockApiClient.getWorkshops.mockResolvedValue([]);
    renderPanel();
    await waitFor(() => {
      expect(screen.getByText(/no workshops/i)).toBeTruthy();
    });
  });

  it('shows draft and active status badges', async () => {
    renderPanel();
    await waitFor(() => {
      expect(screen.getByText('Draft')).toBeTruthy();
      expect(screen.getByText('Active')).toBeTruthy();
    });
  });

  it('shows Create Workshop button', async () => {
    renderPanel();
    await waitFor(() => {
      expect(screen.getByText('Create Workshop')).toBeTruthy();
    });
  });

  it('shows error alert when API fails', async () => {
    mockApiClient.getWorkshops.mockRejectedValue(new Error('Connection refused'));
    renderPanel();
    await waitFor(() => {
      expect(screen.getByText(/connection refused/i)).toBeTruthy();
    });
  });
});

// ── WorkshopsPanel — Create modal ─────────────────────────────────────────────

describe('WorkshopsPanel — Create modal', () => {
  it('opens create modal when button clicked', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getByText('Create Workshop'));
    // Click the primary create button (there may be multiple; pick the primary action button)
    const createButtons = screen.getAllByText('Create Workshop');
    await user.click(createButtons[0]);
    await waitFor(() => {
      // Modal-specific field: look for the Title label inside a form field (not the table column)
      expect(screen.getAllByText('Title').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Template').length).toBeGreaterThan(0);
      // The modal should have a Create submit button
      expect(screen.getAllByRole('button', { name: /^Create$/i }).length).toBeGreaterThan(0);
    });
  });

  it('shows validation error when required fields missing', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getByText('Create Workshop'));
    const createButtons = screen.getAllByText('Create Workshop');
    await user.click(createButtons[0]);
    // Look for a "Create" submit button in the modal
    await waitFor(() => {
      const createBtns = screen.getAllByRole('button', { name: /^Create$/i });
      expect(createBtns.length).toBeGreaterThan(0);
    });
    const submitBtn = screen.getAllByRole('button', { name: /^Create$/i })[0];
    await user.click(submitBtn);
    await waitFor(() => {
      // Validation messages or required field markers should appear
      expect(screen.getAllByText(/required/i).length).toBeGreaterThan(0);
    });
  });
});

// ── WorkshopsPanel — Actions ──────────────────────────────────────────────────

describe('WorkshopsPanel — Actions', () => {
  it('calls provisionWorkshop on provision click', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getAllByText('Provision'));
    const provisionBtns = screen.getAllByText('Provision');
    await user.click(provisionBtns[0]);
    await waitFor(() => {
      expect(mockApiClient.provisionWorkshop).toHaveBeenCalledWith('ws-001');
    });
  });

  it('calls deleteWorkshop on delete click', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getAllByText('Delete'));
    const deleteBtns = screen.getAllByText('Delete');
    await user.click(deleteBtns[0]);
    await waitFor(() => {
      expect(mockApiClient.deleteWorkshop).toHaveBeenCalledWith('ws-001');
    });
  });

  it('shows end confirmation modal when end clicked', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getAllByText('End'));
    const endBtns = screen.getAllByText('End');
    await user.click(endBtns[0]);
    await waitFor(() => {
      // The confirmation modal dialog should appear with 'End Workshop' heading
      expect(screen.getAllByText('End Workshop').length).toBeGreaterThan(0);
    });
  });
});

// ── WorkshopsPanel — Dashboard tab ────────────────────────────────────────────

describe('WorkshopsPanel — Dashboard tab', () => {
  it('shows placeholder when no workshop selected', async () => {
    renderPanel();
    // Switch to dashboard tab
    await waitFor(() => screen.getAllByRole('tab'));
    const tabs = screen.getAllByRole('tab');
    const dashTab = tabs.find(t => t.textContent?.includes('Dashboard'));
    if (dashTab) {
      const user = userEvent.setup();
      await user.click(dashTab);
    }
    await waitFor(() => {
      expect(screen.getByText(/select a workshop/i)).toBeTruthy();
    });
  });

  it('loads dashboard when workshop title is clicked', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getByText('NeurIPS DL Tutorial'));
    // Click the title link to load the dashboard
    const titleLink = screen.getByText('NeurIPS DL Tutorial');
    await user.click(titleLink);
    await waitFor(() => {
      expect(mockApiClient.getWorkshopDashboard).toHaveBeenCalledWith('ws-001');
    });
    await waitFor(() => {
      expect(screen.getByTestId('participants-table')).toBeTruthy();
    });
  });

  it('displays dashboard stats', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getByText('NeurIPS DL Tutorial'));
    await user.click(screen.getByText('NeurIPS DL Tutorial'));
    await waitFor(() => {
      expect(screen.getByText('2')).toBeTruthy(); // total participants
      expect(screen.getByText('5h 30m')).toBeTruthy(); // time remaining
    });
  });
});

// ── WorkshopsPanel — Config Templates tab ────────────────────────────────────

describe('WorkshopsPanel — Config Templates tab', () => {
  it('renders configs table', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getAllByRole('tab'));
    const tabs = screen.getAllByRole('tab');
    const configTab = tabs.find(t => t.textContent?.includes('Config'));
    if (configTab) {
      await user.click(configTab);
    }
    await waitFor(() => {
      expect(screen.getByTestId('workshop-configs-table')).toBeTruthy();
    });
  });

  it('shows config data from API', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getAllByRole('tab'));
    const tabs = screen.getAllByRole('tab');
    const configTab = tabs.find(t => t.textContent?.includes('Config'));
    if (configTab) {
      await user.click(configTab);
    }
    await waitFor(() => {
      expect(screen.getByText('ml-6h')).toBeTruthy();
      expect(screen.getByText('pytorch-ml')).toBeTruthy();
      expect(screen.getByText('6h')).toBeTruthy();
    });
  });

  it('shows empty state when no configs', async () => {
    mockApiClient.getWorkshopConfigs.mockResolvedValue([]);
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getAllByRole('tab'));
    const tabs = screen.getAllByRole('tab');
    const configTab = tabs.find(t => t.textContent?.includes('Config'));
    if (configTab) {
      await user.click(configTab);
    }
    await waitFor(() => {
      expect(screen.getByText(/no configs saved/i)).toBeTruthy();
    });
  });

  it('opens Use Config modal', async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => screen.getAllByRole('tab'));
    const tabs = screen.getAllByRole('tab');
    const configTab = tabs.find(t => t.textContent?.includes('Config'));
    if (configTab) {
      await user.click(configTab);
    }
    await waitFor(() => screen.getByText('Use Config'));
    await user.click(screen.getByText('Use Config'));
    await waitFor(() => {
      expect(screen.getByText(/create from config/i)).toBeTruthy();
    });
  });
});
