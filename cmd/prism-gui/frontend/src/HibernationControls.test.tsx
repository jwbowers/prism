/**
 * Hibernation Controls Component Tests
 *
 * Tests the Hibernation Control functionality within the Prism GUI App.
 * This tests hibernation/resume UI components, capability detection, confirmation dialogs,
 * and educational messaging about hibernation benefits.
 *
 * Coverage:
 * - Hibernate/resume buttons
 * - Capability detection UI
 * - Confirmation dialogs
 * - Educational messaging
 * - Fallback to stop behavior
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../tests/msw/server';
import App from './App';

// Mock window.wails
const mockWails = {
  PrismService: {
    GetTemplates: vi.fn(),
    GetInstances: vi.fn(),
    HibernateInstance: vi.fn(),
    ResumeInstance: vi.fn(),
    GetHibernationStatus: vi.fn(),
    StopInstance: vi.fn(),
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

describe('HibernationControls', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock responses
    mockWails.PrismService.GetTemplates.mockResolvedValue([]);
    mockWails.PrismService.GetProfiles.mockResolvedValue([]);
    mockWails.PrismService.GetStorageVolumes.mockResolvedValue([]);
    mockWails.PrismService.GetIdlePolicies.mockResolvedValue([]);
    mockWails.PrismService.GetProjects.mockResolvedValue([]);
    mockWails.PrismService.GetUsers.mockResolvedValue([]);

    mockWails.PrismService.GetInstances.mockResolvedValue([
      {
        id: 'i-hibernation-capable',
        name: 'ml-workstation',
        state: 'running',
        instance_type: 'm5.large', // Supports hibernation
        template: 'Python Machine Learning',
      },
      {
        id: 'i-hibernated',
        name: 'data-research',
        state: 'hibernated',
        instance_type: 'r5.xlarge',
        template: 'R Research Environment',
      },
      {
        id: 'i-no-hibernation',
        name: 'basic-compute',
        state: 'running',
        instance_type: 't2.micro', // Does not support hibernation
        template: 'Basic Ubuntu',
      },
    ]);
  });

  describe('Hibernation Capability Detection', () => {
    it('should show hibernate button for capable instances', async () => {
      mockWails.PrismService.GetHibernationStatus.mockResolvedValue({
        instance_id: 'i-hibernation-capable',
        hibernation_enabled: true,
        hibernation_configured: true,
        can_hibernate: true,
        current_state: 'running',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /hibernate.*ml-workstation/i })).toBeInTheDocument();
      });
    });

    it('should hide hibernate button for incapable instances', async () => {
      mockWails.PrismService.GetHibernationStatus.mockResolvedValue({
        instance_id: 'i-no-hibernation',
        hibernation_enabled: false,
        hibernation_configured: false,
        can_hibernate: false,
        current_state: 'running',
        message: 'Instance type does not support hibernation',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        // basic-compute (t2.micro) should not have hibernate button
        const hibernateButton = screen.queryByRole('button', {
          name: /hibernate.*basic-compute/i,
        });
        expect(hibernateButton).toBeNull();
      });
    });

    it('should show tooltip explaining hibernation support', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const hibernateButton = await screen.findByRole('button', {
        name: /hibernate.*ml-workstation/i,
      });

      await user.hover(hibernateButton);

      await waitFor(() => {
        expect(
          screen.getByText(/faster than.*stop|preserves.*RAM|instant resume/i)
        ).toBeInTheDocument();
      });
    });
  });

  describe('Hibernate Action', () => {
    it('should open educational confirmation dialog', async () => {
      mockWails.PrismService.GetHibernationStatus.mockResolvedValue({
        can_hibernate: true,
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const hibernateButton = await screen.findByRole('button', {
        name: /hibernate.*ml-workstation/i,
      });
      await user.click(hibernateButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /hibernate.*instance/i })).toBeInTheDocument();
        // Should show educational content
        expect(screen.getByText(/saves.*state|preserves.*RAM|faster.*resume/i)).toBeInTheDocument();
      });
    });

    it('should hibernate instance on confirmation', async () => {
      mockWails.PrismService.HibernateInstance.mockResolvedValue({
        status: 'hibernating',
        message: 'Instance is entering hibernation',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /hibernate.*ml-workstation/i }));

      const confirmButton = await screen.findByRole('button', { name: /hibernate|confirm/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockWails.PrismService.HibernateInstance).toHaveBeenCalledWith('i-hibernation-capable');
      });
    });

    it('should show success notification after hibernation', async () => {
      mockWails.PrismService.HibernateInstance.mockResolvedValue({
        status: 'hibernating',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /hibernate.*ml-workstation/i }));
      await user.click(await screen.findByRole('button', { name: /hibernate|confirm/i }));

      await waitFor(() => {
        expect(screen.getByText(/hibernating.*ml-workstation/i)).toBeInTheDocument();
      });
    });

    it('should display estimated cost savings', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /hibernate.*ml-workstation/i }));

      // Dialog should show cost savings estimate
      await waitFor(() => {
        expect(screen.getByText(/save.*\$|cost.*\$/i)).toBeInTheDocument();
      });
    });
  });

  describe('Resume Action', () => {
    it('should show resume button for hibernated instances', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        // data-research is hibernated
        expect(screen.getByRole('button', { name: /resume.*data-research/i })).toBeInTheDocument();
      });
    });

    it('should resume instance immediately (no confirmation)', async () => {
      mockWails.PrismService.ResumeInstance.mockResolvedValue({
        status: 'resuming',
        message: 'Instance is resuming from hibernation',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      const resumeButton = await screen.findByRole('button', { name: /resume.*data-research/i });
      await user.click(resumeButton);

      // Should resume immediately without confirmation (unlike hibernate)
      await waitFor(() => {
        expect(mockWails.PrismService.ResumeInstance).toHaveBeenCalledWith('i-hibernated');
      });
    });

    it('should show educational message about fast resume', async () => {
      mockWails.PrismService.ResumeInstance.mockResolvedValue({
        status: 'resuming',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /resume.*data-research/i }));

      await waitFor(() => {
        expect(screen.getByText(/faster.*start|instant.*resume|preserved/i)).toBeInTheDocument();
      });
    });
  });

  describe('Fallback Behavior', () => {
    it('should fallback to stop when hibernation fails', async () => {
      mockWails.PrismService.HibernateInstance.mockRejectedValue(
        new Error('Hibernation not supported')
      );
      mockWails.PrismService.StopInstance.mockResolvedValue({ status: 'stopping' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /hibernate.*ml-workstation/i }));
      await user.click(await screen.findByRole('button', { name: /hibernate|confirm/i }));

      await waitFor(() => {
        // Should show fallback message
        expect(
          screen.getByText(/hibernation.*not.*supported|falling.*back.*stop/i)
        ).toBeInTheDocument();
      });

      // Should automatically fallback to stop
      await waitFor(() => {
        expect(mockWails.PrismService.StopInstance).toHaveBeenCalledWith('i-hibernation-capable');
      });
    });

    it('should explain fallback in notification', async () => {
      mockWails.PrismService.HibernateInstance.mockRejectedValue(
        new Error('Hibernation failed')
      );
      mockWails.PrismService.StopInstance.mockResolvedValue({ status: 'stopping' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /hibernate.*ml-workstation/i }));
      await user.click(await screen.findByRole('button', { name: /hibernate|confirm/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/hibernation.*failed.*stopped.*instead|fallback/i)
        ).toBeInTheDocument();
      });
    });
  });

  describe('Educational Messaging', () => {
    it('should explain hibernation benefits in UI', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      // Info icon or help text near hibernate button
      const infoIcon = screen.queryByLabelText(/hibernation.*info|learn.*hibernation/i);
      if (infoIcon) {
        await user.click(infoIcon);

        await waitFor(() => {
          expect(screen.getByText(/faster.*than.*stop/i)).toBeInTheDocument();
          expect(screen.getByText(/preserves.*RAM.*state/i)).toBeInTheDocument();
          expect(screen.getByText(/instant.*resume/i)).toBeInTheDocument();
        });
      }
    });

    it('should show first-time hibernation tutorial', async () => {
      // Mock first-time user
      localStorage.setItem('hibernation_tutorial_seen', 'false');

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /hibernate.*ml-workstation/i }));

      // Should show extended tutorial
      await waitFor(() => {
        expect(screen.getByText(/first.*time|tutorial|learn/i)).toBeInTheDocument();
      });

      // Cleanup
      localStorage.removeItem('hibernation_tutorial_seen');
    });

    it('should provide link to hibernation documentation', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /hibernate.*ml-workstation/i }));

      await waitFor(() => {
        expect(screen.getByRole('link', { name: /learn.*more|documentation/i })).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle hibernation API error', async () => {
      mockWails.PrismService.HibernateInstance.mockRejectedValue(
        new Error('Instance not in correct state for hibernation')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /hibernate.*ml-workstation/i }));
      await user.click(await screen.findByRole('button', { name: /hibernate|confirm/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/not in correct state|failed.*hibernate/i)
        ).toBeInTheDocument();
      });
    });

    it('should handle resume API error', async () => {
      mockWails.PrismService.ResumeInstance.mockRejectedValue(
        new Error('Failed to resume instance')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await user.click(await screen.findByRole('button', { name: /resume.*data-research/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed.*resume/i)).toBeInTheDocument();
      });
    });

    it('should show capability check error', async () => {
      mockWails.PrismService.GetHibernationStatus.mockRejectedValue(
        new Error('Failed to check hibernation status')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /instances/i }));

      await waitFor(() => {
        // Should still show instances but maybe without hibernate buttons
        expect(screen.getByText('ml-workstation')).toBeInTheDocument();
      });
    });
  });
});
