/**
 * Instance Manager Component Tests
 *
 * Tests the Instance Management functionality within the Prism GUI App.
 * This tests the "Instances" tab including instance table, actions, launch workflow,
 * connection info, and status updates.
 *
 * Coverage:
 * - Instance table with filters
 * - Action buttons (stop, terminate, hibernate, resume, start)
 * - Launch instance workflow
 * - Connection info display
 * - Status updates and polling
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../tests/msw/server';
import App from './App';

// Mock window.wails
const mockWails = {
  PrismService: {
    GetTemplates: vi.fn(),
    GetInstances: vi.fn(),
    LaunchInstance: vi.fn(),
    StopInstance: vi.fn(),
    StartInstance: vi.fn(),
    TerminateInstance: vi.fn(),
    HibernateInstance: vi.fn(),
    ResumeInstance: vi.fn(),
    GetConnectionInfo: vi.fn(),
    GetInstanceStatus: vi.fn(),
    GetProfiles: vi.fn(),
    GetStorageVolumes: vi.fn(),
    GetIdlePolicies: vi.fn(),
    GetProjects: vi.fn(),
    GetUsers: vi.fn(),
  },
};

Object.defineProperty(window, 'wails', {
  value: mockWails,
  writable: true,
});

describe('InstanceManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock responses
    mockWails.PrismService.GetTemplates.mockResolvedValue([]);
    mockWails.PrismService.GetProfiles.mockResolvedValue([]);
    mockWails.PrismService.GetStorageVolumes.mockResolvedValue([]);
    mockWails.PrismService.GetIdlePolicies.mockResolvedValue([]);
    mockWails.PrismService.GetProjects.mockResolvedValue([]);
    mockWails.PrismService.GetUsers.mockResolvedValue([]);

    // Instance-specific mocks
    mockWails.PrismService.GetInstances.mockResolvedValue([
      {
        id: 'i-1234567890abcdef0',
        name: 'ml-research',
        template: 'Python Machine Learning',
        state: 'running',
        public_ip: '54.123.45.67',
        instance_type: 't3.xlarge',
        launch_time: '2025-01-15T10:30:00Z',
        region: 'us-west-2',
        cost_per_hour: 0.48,
        estimated_daily_cost: 11.52,
      },
      {
        id: 'i-abcdef1234567890',
        name: 'data-analysis',
        template: 'R Research Environment',
        state: 'stopped',
        instance_type: 't3.large',
        launch_time: '2025-01-10T08:00:00Z',
        region: 'us-west-2',
        cost_per_hour: 0.24,
        estimated_daily_cost: 0, // stopped
      },
      {
        id: 'i-fedcba0987654321',
        name: 'gpu-training',
        template: 'Deep Learning GPU',
        state: 'hibernated',
        instance_type: 'g4dn.xlarge',
        launch_time: '2025-01-12T14:00:00Z',
        region: 'us-east-1',
        cost_per_hour: 0.526,
        estimated_daily_cost: 0, // hibernated
      },
    ]);
  });

  describe('Instance List Rendering', () => {
    it('should display list of instances', async () => {
      render(<App />);
      const user = userEvent.setup();

      // Navigate to Instances tab
      const instancesTab = screen.getByRole('link', { name: /instances/i });
      await user.click(instancesTab);

      await waitFor(() => {
        expect(screen.getByText('ml-research')).toBeInTheDocument();
        expect(screen.getByText('data-analysis')).toBeInTheDocument();
        expect(screen.getByText('gpu-training')).toBeInTheDocument();
      });
    });

    it('should display instance details (state, type, region, cost)', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        expect(screen.getByText(/running/i)).toBeInTheDocument();
        expect(screen.getByText(/stopped/i)).toBeInTheDocument();
        expect(screen.getByText(/hibernated/i)).toBeInTheDocument();
        expect(screen.getByText(/t3\.xlarge/i)).toBeInTheDocument();
        expect(screen.getByText(/us-west-2/i)).toBeInTheDocument();
        expect(screen.getByText(/\$0\.48.*hour|\$11\.52.*day/i)).toBeInTheDocument();
      });
    });

    it('should display public IP for running instances', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        expect(screen.getByText('54.123.45.67')).toBeInTheDocument();
      });
    });

    it('should handle empty instance list', async () => {
      mockWails.PrismService.GetInstances.mockResolvedValue([]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        expect(screen.getByText(/no.*instances|launch your first instance/i)).toBeInTheDocument();
      });
    });

    it('should show launch time and uptime', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        // Should show relative time like "5 days ago" or absolute time
        expect(screen.getByText(/ago|2025-01-15/i)).toBeInTheDocument();
      });
    });
  });

  describe('Instance Actions - Stop', () => {
    it('should open stop confirmation dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const stopButton = await screen.findByRole('button', { name: /stop.*ml-research/i });
      await user.click(stopButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /stop instance|confirm/i })).toBeInTheDocument();
        expect(screen.getByText(/are you sure.*stop/i)).toBeInTheDocument();
      });
    });

    it('should stop instance on confirmation', async () => {
      mockWails.PrismService.StopInstance.mockResolvedValue({ status: 'stopping' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const stopButton = await screen.findByRole('button', { name: /stop.*ml-research/i });
      await user.click(stopButton);

      const confirmButton = await screen.findByRole('button', { name: /stop|confirm/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockWails.PrismService.StopInstance).toHaveBeenCalledWith('i-1234567890abcdef0');
      });
    });

    it('should show success notification after stopping', async () => {
      mockWails.PrismService.StopInstance.mockResolvedValue({ status: 'stopping' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const stopButton = await screen.findByRole('button', { name: /stop.*ml-research/i });
      await user.click(stopButton);

      await user.click(await screen.findByRole('button', { name: /stop|confirm/i }));

      await waitFor(() => {
        expect(screen.getByText(/stopping.*ml-research|instance.*stopping/i)).toBeInTheDocument();
      });
    });
  });

  describe('Instance Actions - Start', () => {
    it('should start stopped instance', async () => {
      mockWails.PrismService.StartInstance.mockResolvedValue({ status: 'starting' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      // data-analysis is stopped
      const startButton = await screen.findByRole('button', { name: /start.*data-analysis/i });
      await user.click(startButton);

      await waitFor(() => {
        expect(mockWails.PrismService.StartInstance).toHaveBeenCalledWith('i-abcdef1234567890');
      });
    });

    it('should disable start button for running instances', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        // ml-research is running, start should be disabled
        const startButton = screen.queryByRole('button', { name: /start.*ml-research/i });
        if (startButton) {
          expect(startButton).toBeDisabled();
        }
      });
    });
  });

  describe('Instance Actions - Terminate', () => {
    it('should open terminate confirmation with warning', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const terminateButton = await screen.findByRole('button', {
        name: /terminate.*ml-research|delete.*ml-research/i,
      });
      await user.click(terminateButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /terminate|delete/i })).toBeInTheDocument();
        expect(screen.getByText(/permanent|cannot be undone/i)).toBeInTheDocument();
      });
    });

    it('should require typing instance name for confirmation', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const terminateButton = await screen.findByRole('button', {
        name: /terminate.*ml-research/i,
      });
      await user.click(terminateButton);

      // Terminate button should be disabled until name is typed
      const confirmButton = await screen.findByRole('button', { name: /terminate|delete/i });
      expect(confirmButton).toBeDisabled();

      // Type instance name
      const nameInput = screen.getByPlaceholderText(/type.*ml-research|enter.*name/i);
      await user.type(nameInput, 'ml-research');

      await waitFor(() => {
        expect(confirmButton).not.toBeDisabled();
      });
    });

    it('should terminate instance after confirmation', async () => {
      mockWails.PrismService.TerminateInstance.mockResolvedValue({ status: 'terminating' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const terminateButton = await screen.findByRole('button', {
        name: /terminate.*ml-research/i,
      });
      await user.click(terminateButton);

      const nameInput = screen.getByPlaceholderText(/type.*ml-research/i);
      await user.type(nameInput, 'ml-research');

      const confirmButton = await screen.findByRole('button', { name: /terminate|delete/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockWails.PrismService.TerminateInstance).toHaveBeenCalledWith('i-1234567890abcdef0');
      });
    });
  });

  describe('Instance Actions - Hibernate/Resume', () => {
    it('should hibernate running instance', async () => {
      mockWails.PrismService.HibernateInstance.mockResolvedValue({ status: 'hibernating' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const hibernateButton = await screen.findByRole('button', {
        name: /hibernate.*ml-research/i,
      });
      await user.click(hibernateButton);

      await waitFor(() => {
        expect(mockWails.PrismService.HibernateInstance).toHaveBeenCalledWith('i-1234567890abcdef0');
      });
    });

    it('should resume hibernated instance', async () => {
      mockWails.PrismService.ResumeInstance.mockResolvedValue({ status: 'resuming' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      // gpu-training is hibernated
      const resumeButton = await screen.findByRole('button', { name: /resume.*gpu-training/i });
      await user.click(resumeButton);

      await waitFor(() => {
        expect(mockWails.PrismService.ResumeInstance).toHaveBeenCalledWith('i-fedcba0987654321');
      });
    });

    it('should show educational message about hibernation', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const hibernateButton = await screen.findByRole('button', {
        name: /hibernate.*ml-research/i,
      });

      // Hover or click should show info
      await user.hover(hibernateButton);

      await waitFor(() => {
        // Should show tooltip or info about hibernation benefits
        expect(screen.getByText(/faster than.*stop|preserves.*ram/i)).toBeInTheDocument();
      });
    });
  });

  describe('Launch Instance Workflow', () => {
    it('should open launch wizard from templates tab', async () => {
      mockWails.PrismService.GetTemplates.mockResolvedValue([
        {
          Name: 'Python Machine Learning',
          Description: 'ML environment',
          Category: 'Machine Learning',
        },
      ]);

      render(<App />);
      const user = userEvent.setup();

      // Should start on templates tab
      await waitFor(() => {
        expect(screen.getByText('Python Machine Learning')).toBeInTheDocument();
      });

      // Click launch on a template
      const launchButton = await screen.findByRole('button', { name: /launch.*python/i });
      await user.click(launchButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /launch.*instance|configure/i })).toBeInTheDocument();
      });
    });

    it('should require instance name', async () => {
      mockWails.PrismService.GetTemplates.mockResolvedValue([
        { Name: 'Python Machine Learning', Description: 'ML' },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(await screen.findByRole('button', { name: /launch/i }));

      // Try to launch without name
      const launchButton = screen.getByRole('button', { name: /launch|start/i });
      await user.click(launchButton);

      await waitFor(() => {
        expect(screen.getByText(/instance name.*required/i)).toBeInTheDocument();
      });
    });

    it('should validate instance name format', async () => {
      mockWails.PrismService.GetTemplates.mockResolvedValue([
        { Name: 'Python Machine Learning' },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(await screen.findByRole('button', { name: /launch/i }));

      const nameInput = screen.getByLabelText(/instance name/i);
      await user.type(nameInput, 'Invalid Name!'); // spaces and special chars

      await waitFor(() => {
        expect(screen.getByText(/invalid.*name|alphanumeric/i)).toBeInTheDocument();
      });
    });

    it('should launch instance with valid configuration', async () => {
      mockWails.PrismService.GetTemplates.mockResolvedValue([
        { Name: 'Python Machine Learning' },
      ]);
      mockWails.PrismService.LaunchInstance.mockResolvedValue({
        instance_id: 'i-newinstance123',
        status: 'pending',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(await screen.findByRole('button', { name: /launch/i }));

      const nameInput = screen.getByLabelText(/instance name/i);
      await user.type(nameInput, 'test-ml-instance');

      const launchButton = screen.getByRole('button', { name: /launch|start/i });
      await user.click(launchButton);

      await waitFor(() => {
        expect(mockWails.PrismService.LaunchInstance).toHaveBeenCalledWith({
          template: 'Python Machine Learning',
          name: 'test-ml-instance',
        });
      });
    });

    it('should show advanced options (size, spot)', async () => {
      mockWails.PrismService.GetTemplates.mockResolvedValue([
        { Name: 'Python Machine Learning' },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(await screen.findByRole('button', { name: /launch/i }));

      // Click advanced options toggle
      const advancedToggle = screen.getByRole('button', { name: /advanced|show more/i });
      await user.click(advancedToggle);

      await waitFor(() => {
        expect(screen.getByLabelText(/size|instance type/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/spot instance/i)).toBeInTheDocument();
      });
    });

    it('should show launch progress', async () => {
      mockWails.PrismService.GetTemplates.mockResolvedValue([
        { Name: 'Python Machine Learning' },
      ]);
      mockWails.PrismService.LaunchInstance.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve({ instance_id: 'i-new' }), 100))
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(await screen.findByRole('button', { name: /launch/i }));

      const nameInput = screen.getByLabelText(/instance name/i);
      await user.type(nameInput, 'test-instance');

      await user.click(screen.getByRole('button', { name: /launch|start/i }));

      // Should show progress indicator
      expect(screen.getByText(/launching|creating/i)).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getByText(/launched.*successfully/i)).toBeInTheDocument();
      });
    });

    it('should handle launch error', async () => {
      mockWails.PrismService.GetTemplates.mockResolvedValue([
        { Name: 'Python Machine Learning' },
      ]);
      mockWails.PrismService.LaunchInstance.mockRejectedValue(
        new Error('Insufficient capacity in availability zone')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(await screen.findByRole('button', { name: /launch/i }));

      const nameInput = screen.getByLabelText(/instance name/i);
      await user.type(nameInput, 'test-instance');

      await user.click(screen.getByRole('button', { name: /launch|start/i }));

      await waitFor(() => {
        expect(screen.getByText(/insufficient capacity|failed to launch/i)).toBeInTheDocument();
      });
    });
  });

  describe('Connection Info Display', () => {
    it('should show connection info button for running instances', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        // ml-research is running
        expect(screen.getByRole('button', { name: /connect.*ml-research/i })).toBeInTheDocument();
      });
    });

    it('should open connection info dialog', async () => {
      mockWails.PrismService.GetConnectionInfo.mockResolvedValue({
        instance_id: 'i-1234567890abcdef0',
        public_ip: '54.123.45.67',
        ssh_command: 'ssh ubuntu@54.123.45.67',
        ssh_key_path: '~/.ssh/prism-key.pem',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const connectButton = await screen.findByRole('button', { name: /connect.*ml-research/i });
      await user.click(connectButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /connection info/i })).toBeInTheDocument();
        expect(screen.getByText('ssh ubuntu@54.123.45.67')).toBeInTheDocument();
      });
    });

    it('should provide copy button for SSH command', async () => {
      mockWails.PrismService.GetConnectionInfo.mockResolvedValue({
        ssh_command: 'ssh ubuntu@54.123.45.67',
      });

      // Mock clipboard API
      Object.assign(navigator, {
        clipboard: {
          writeText: vi.fn().mockResolvedValue(undefined),
        },
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));
      await user.click(await screen.findByRole('button', { name: /connect.*ml-research/i }));

      const copyButton = await screen.findByRole('button', { name: /copy/i });
      await user.click(copyButton);

      await waitFor(() => {
        expect(navigator.clipboard.writeText).toHaveBeenCalledWith('ssh ubuntu@54.123.45.67');
        expect(screen.getByText(/copied/i)).toBeInTheDocument();
      });
    });

    it('should hide connect button for stopped instances', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        // data-analysis is stopped, should not have connect button
        const connectButton = screen.queryByRole('button', { name: /connect.*data-analysis/i });
        expect(connectButton).toBeNull();
      });
    });
  });

  describe('Status Updates and Polling', () => {
    it('should poll for instance status updates', async () => {
      vi.useFakeTimers();

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      // Initial fetch
      expect(mockWails.PrismService.GetInstances).toHaveBeenCalledTimes(1);

      // Advance time to trigger polling
      vi.advanceTimersByTime(30000); // 30 seconds

      await waitFor(() => {
        expect(mockWails.PrismService.GetInstances).toHaveBeenCalledTimes(2);
      });

      vi.useRealTimers();
    });

    it('should update instance state in realtime', async () => {
      mockWails.PrismService.GetInstances
        .mockResolvedValueOnce([
          {
            id: 'i-test',
            name: 'test-instance',
            state: 'pending',
          },
        ])
        .mockResolvedValueOnce([
          {
            id: 'i-test',
            name: 'test-instance',
            state: 'running',
          },
        ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      // Initially pending
      await waitFor(() => {
        expect(screen.getByText(/pending/i)).toBeInTheDocument();
      });

      // Manually trigger refresh
      const refreshButton = screen.getByRole('button', { name: /refresh/i });
      await user.click(refreshButton);

      // Should now show running
      await waitFor(() => {
        expect(screen.getByText(/running/i)).toBeInTheDocument();
        expect(screen.queryByText(/pending/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('Instance Filtering and Search', () => {
    it('should filter instances by state', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      // Filter to show only running
      const stateFilter = await screen.findByLabelText(/filter by state/i);
      await user.selectOptions(stateFilter, 'running');

      await waitFor(() => {
        expect(screen.getByText('ml-research')).toBeInTheDocument();
        expect(screen.queryByText('data-analysis')).not.toBeInTheDocument();
      });
    });

    it('should search instances by name', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const searchInput = await screen.findByPlaceholderText(/search.*instances|find/i);
      await user.type(searchInput, 'gpu');

      await waitFor(() => {
        expect(screen.getByText('gpu-training')).toBeInTheDocument();
        expect(screen.queryByText('ml-research')).not.toBeInTheDocument();
      });
    });

    it('should filter by template', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const templateFilter = await screen.findByLabelText(/filter by template/i);
      await user.selectOptions(templateFilter, 'Python Machine Learning');

      await waitFor(() => {
        expect(screen.getByText('ml-research')).toBeInTheDocument();
        expect(screen.queryByText('data-analysis')).not.toBeInTheDocument();
      });
    });
  });

  describe('Loading and Error States', () => {
    it('should show loading state while fetching instances', async () => {
      mockWails.PrismService.GetInstances.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve([]), 100))
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      expect(screen.getByText(/loading/i)).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
      });
    });

    it('should handle API error when fetching instances', async () => {
      mockWails.PrismService.GetInstances.mockRejectedValue(
        new Error('Failed to fetch instances')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed.*load instances|error.*instances/i)).toBeInTheDocument();
      });
    });

    it('should handle action errors gracefully', async () => {
      mockWails.PrismService.StopInstance.mockRejectedValue(
        new Error('Instance not in correct state')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const stopButton = await screen.findByRole('button', { name: /stop.*ml-research/i });
      await user.click(stopButton);

      await user.click(await screen.findByRole('button', { name: /stop|confirm/i }));

      await waitFor(() => {
        expect(screen.getByText(/not in correct state|failed to stop/i)).toBeInTheDocument();
      });
    });
  });
});
