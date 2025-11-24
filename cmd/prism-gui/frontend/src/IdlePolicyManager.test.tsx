/**
 * Idle Policy Manager Component Tests
 *
 * Tests the Idle Detection Policy Management functionality within the Prism GUI App.
 * This tests idle policy CRUD, policy application, history display, and savings reports.
 *
 * Coverage:
 * - Policy list and creation
 * - Policy application to instances
 * - Idle history display
 * - Savings reports
 * - Validation and presets
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// Mock window.wails
const mockWails = {
  PrismService: {
    GetTemplates: vi.fn(),
    GetInstances: vi.fn(),
    GetIdlePolicies: vi.fn(),
    CreateIdlePolicy: vi.fn(),
    UpdateIdlePolicy: vi.fn(),
    DeleteIdlePolicy: vi.fn(),
    ApplyIdlePolicy: vi.fn(),
    GetIdleHistory: vi.fn(),
    GetIdlePolicyPresets: vi.fn(),
    GetProfiles: vi.fn(),
    GetStorageVolumes: vi.fn(),
    GetProjects: vi.fn(),
    GetUsers: vi.fn(),
  },
};

Object.defineProperty(window, 'wails', {
  value: mockWails,
  writable: true,
});

describe('IdlePolicyManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    mockWails.PrismService.GetTemplates.mockResolvedValue([]);
    mockWails.PrismService.GetInstances.mockResolvedValue([]);
    mockWails.PrismService.GetProfiles.mockResolvedValue([]);
    mockWails.PrismService.GetStorageVolumes.mockResolvedValue([]);
    mockWails.PrismService.GetProjects.mockResolvedValue([]);
    mockWails.PrismService.GetUsers.mockResolvedValue([]);

    mockWails.PrismService.GetIdlePolicies.mockResolvedValue([
      {
        id: 'pol-gpu',
        name: 'gpu',
        description: 'GPU instance idle policy',
        idle_minutes: 15,
        action: 'stop',
        cpu_threshold: 10,
        memory_threshold: 20,
        network_threshold: 1000,
        disk_threshold: 1000,
        gpu_threshold: 10,
        enabled: true,
      },
      {
        id: 'pol-batch',
        name: 'batch',
        description: 'Batch processing idle policy',
        idle_minutes: 60,
        action: 'hibernate',
        cpu_threshold: 5,
        memory_threshold: 10,
        network_threshold: 500,
        disk_threshold: 500,
        enabled: true,
      },
    ]);
  });

  describe('Policy List Rendering', () => {
    it('should display list of idle policies', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        expect(screen.getByText('gpu')).toBeInTheDocument();
        expect(screen.getByText('batch')).toBeInTheDocument();
      });
    });

    it('should display policy details (idle time, action, thresholds)', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        expect(screen.getByText(/15.*minutes|15.*min/i)).toBeInTheDocument();
        expect(screen.getByText(/60.*minutes|1.*hour/i)).toBeInTheDocument();
        expect(screen.getByText(/stop/i)).toBeInTheDocument();
        expect(screen.getByText(/hibernate/i)).toBeInTheDocument();
      });
    });

    it('should show enabled/disabled status', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        expect(screen.getAllByText(/enabled|active/i)).toHaveLength(2); // both policies enabled
      });
    });
  });

  describe('Create Policy', () => {
    it('should open create policy dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const createButton = await screen.findByRole('button', { name: /create.*policy|add.*policy/i });
      await user.click(createButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /create.*policy/i })).toBeInTheDocument();
      });
    });

    it('should create policy with valid input', async () => {
      mockWails.PrismService.CreateIdlePolicy.mockResolvedValue({ id: 'pol-new', success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /create.*policy/i }));

      await user.type(screen.getByLabelText(/policy name/i), 'cost-optimized');
      await user.type(screen.getByLabelText(/idle.*minutes/i), '10');
      await user.selectOptions(screen.getByLabelText(/action/i), 'hibernate');
      await user.type(screen.getByLabelText(/cpu.*threshold/i), '5');

      await user.click(screen.getByRole('button', { name: /create|save/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.CreateIdlePolicy).toHaveBeenCalledWith({
          name: 'cost-optimized',
          idle_minutes: 10,
          action: 'hibernate',
          cpu_threshold: 5,
        });
      });
    });

    it('should validate idle minutes minimum', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /create.*policy/i }));

      await user.type(screen.getByLabelText(/idle.*minutes/i), '2'); // Less than minimum

      await waitFor(() => {
        expect(screen.getByText(/minimum.*5.*minutes/i)).toBeInTheDocument();
      });
    });
  });

  describe('Update Policy', () => {
    it('should edit existing policy', async () => {
      mockWails.PrismService.UpdateIdlePolicy.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const editButton = await screen.findByRole('button', { name: /edit.*gpu/i });
      await user.click(editButton);

      const idleMinutesInput = screen.getByLabelText(/idle.*minutes/i);
      await user.clear(idleMinutesInput);
      await user.type(idleMinutesInput, '20');

      await user.click(screen.getByRole('button', { name: /save|update/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.UpdateIdlePolicy).toHaveBeenCalledWith('pol-gpu', {
          idle_minutes: 20,
        });
      });
    });
  });

  describe('Delete Policy', () => {
    it('should delete policy on confirmation', async () => {
      mockWails.PrismService.DeleteIdlePolicy.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const deleteButton = await screen.findByRole('button', { name: /delete.*gpu/i });
      await user.click(deleteButton);

      await user.click(await screen.findByRole('button', { name: /delete|confirm/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.DeleteIdlePolicy).toHaveBeenCalledWith('pol-gpu');
      });
    });
  });

  describe('Apply Policy to Instance', () => {
    it('should apply policy to specific instance', async () => {
      mockWails.PrismService.GetInstances.mockResolvedValue([
        { id: 'i-test', name: 'my-instance', state: 'running' },
      ]);
      mockWails.PrismService.ApplyIdlePolicy.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const applyButton = await screen.findByRole('button', { name: /apply.*gpu/i });
      await user.click(applyButton);

      const instanceSelect = await screen.findByLabelText(/select instance/i);
      await user.selectOptions(instanceSelect, 'i-test');

      await user.click(screen.getByRole('button', { name: /apply/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.ApplyIdlePolicy).toHaveBeenCalledWith({
          policy_id: 'pol-gpu',
          instance_id: 'i-test',
        });
      });
    });
  });

  describe('Idle History', () => {
    it('should display idle history', async () => {
      mockWails.PrismService.GetIdleHistory.mockResolvedValue([
        {
          instance_id: 'i-test',
          instance_name: 'ml-research',
          action: 'hibernate',
          reason: 'Idle for 15 minutes (CPU < 10%)',
          timestamp: '2025-01-15T14:30:00Z',
          cost_saved: 2.40,
        },
        {
          instance_id: 'i-test2',
          instance_name: 'data-analysis',
          action: 'stop',
          reason: 'Idle for 60 minutes',
          timestamp: '2025-01-14T08:00:00Z',
          cost_saved: 5.76,
        },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      // Navigate to idle history tab
      const historyTab = screen.queryByRole('tab', { name: /history/i });
      if (historyTab) {
        await user.click(historyTab);
      }

      await waitFor(() => {
        expect(screen.getByText('ml-research')).toBeInTheDocument();
        expect(screen.getByText('data-analysis')).toBeInTheDocument();
        expect(screen.getByText(/\$2\.40/)).toBeInTheDocument();
        expect(screen.getByText(/\$5\.76/)).toBeInTheDocument();
      });
    });

    it('should show total cost savings', async () => {
      mockWails.PrismService.GetIdleHistory.mockResolvedValue([
        { cost_saved: 2.40 },
        { cost_saved: 5.76 },
        { cost_saved: 1.20 },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        // Total: $9.36
        expect(screen.getByText(/total.*\$9\.36|savings.*\$9\.36/i)).toBeInTheDocument();
      });
    });
  });

  describe('Policy Presets', () => {
    it('should load policy from preset', async () => {
      mockWails.PrismService.GetIdlePolicyPresets.mockResolvedValue([
        {
          name: 'cost-optimized',
          description: 'Aggressive cost savings',
          idle_minutes: 10,
          action: 'hibernate',
          cpu_threshold: 5,
        },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /create.*policy/i }));

      // Select preset
      const presetSelect = screen.getByLabelText(/preset|template/i);
      await user.selectOptions(presetSelect, 'cost-optimized');

      await waitFor(() => {
        const idleMinutesInput = screen.getByLabelText(/idle.*minutes/i) as HTMLInputElement;
        expect(idleMinutesInput.value).toBe('10');
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle fetch error', async () => {
      mockWails.PrismService.GetIdlePolicies.mockRejectedValue(new Error('Failed to fetch'));

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed.*load.*policies/i)).toBeInTheDocument();
      });
    });
  });
});
