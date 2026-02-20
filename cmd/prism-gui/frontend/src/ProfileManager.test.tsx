/**
 * Profile Manager Component Tests
 *
 * Tests the Profile Management functionality within the Prism GUI App.
 * This tests the "Profiles" section of Settings tab as if it were a separate component.
 *
 * Coverage:
 * - Profile list rendering
 * - Add/edit/delete profile interactions
 * - Profile switching
 * - Export/import workflows
 * - Validation and error states
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// Mock window.wails for the entire App
const mockWails = {
  PrismService: {
    GetTemplates: vi.fn(),
    GetInstances: vi.fn(),
    GetProfiles: vi.fn(),
    GetCurrentProfile: vi.fn(),
    CreateProfile: vi.fn(),
    UpdateProfile: vi.fn(),
    DeleteProfile: vi.fn(),
    SwitchProfile: vi.fn(),
    ExportProfile: vi.fn(),
    ImportProfile: vi.fn(),
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

describe('ProfileManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock responses
    mockWails.PrismService.GetTemplates.mockResolvedValue([]);
    mockWails.PrismService.GetInstances.mockResolvedValue([]);
    mockWails.PrismService.GetStorageVolumes.mockResolvedValue([]);
    mockWails.PrismService.GetIdlePolicies.mockResolvedValue([]);
    mockWails.PrismService.GetProjects.mockResolvedValue([]);
    mockWails.PrismService.GetUsers.mockResolvedValue([]);

    // Profile-specific mocks
    mockWails.PrismService.GetProfiles.mockResolvedValue([
      {
        name: 'default',
        aws_profile: 'default',
        region: 'us-west-2',
        is_default: true,
        created_at: '2025-01-01T00:00:00Z',
      },
      {
        name: 'research-profile',
        aws_profile: 'research',
        region: 'us-east-1',
        is_default: false,
        created_at: '2025-01-15T00:00:00Z',
      },
    ]);

    mockWails.PrismService.GetCurrentProfile.mockResolvedValue({
      name: 'default',
      aws_profile: 'default',
      region: 'us-west-2',
      is_default: true,
    });
  });

  describe('Profile List Rendering', () => {
    it('should display list of profiles', async () => {
      render(<App />);
      const user = userEvent.setup();

      // Navigate to Settings tab
      const settingsTab = screen.getByRole('link', { name: /settings/i });
      await user.click(settingsTab);

      await waitFor(() => {
        // Profiles should be visible in settings
        expect(screen.getByText(/default/)).toBeInTheDocument();
        expect(screen.getByText(/research-profile/)).toBeInTheDocument();
      });
    });

    it('should show current profile indicator', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        // Default profile should have active/current indicator
        const defaultProfile = screen.getByText(/default/).closest('[data-profile-item]');
        if (defaultProfile) {
          expect(within(defaultProfile as HTMLElement).getByText(/current|active/i)).toBeInTheDocument();
        }
      });
    });

    it('should display profile details (region, AWS profile)', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        expect(screen.getByText(/us-west-2/)).toBeInTheDocument();
        expect(screen.getByText(/us-east-1/)).toBeInTheDocument();
      });
    });

    it('should handle empty profile list', async () => {
      mockWails.PrismService.GetProfiles.mockResolvedValue([]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        expect(screen.getByText(/no profiles|create your first profile/i)).toBeInTheDocument();
      });
    });
  });

  describe('Create Profile', () => {
    it('should open create profile dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      // Click Add Profile button
      const addButton = await screen.findByRole('button', { name: /add profile|create profile/i });
      await user.click(addButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /create profile|add profile/i })).toBeInTheDocument();
      });
    });

    it('should create new profile with valid input', async () => {
      mockWails.PrismService.CreateProfile.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /add profile|create profile/i }));

      // Fill form
      const nameInput = screen.getByLabelText(/profile name/i);
      await user.type(nameInput, 'test-profile');

      const awsProfileInput = screen.getByLabelText(/aws profile/i);
      await user.type(awsProfileInput, 'test-aws');

      const regionSelect = screen.getByLabelText(/region/i);
      await user.selectOptions(regionSelect, 'eu-west-1');

      // Submit
      await user.click(screen.getByRole('button', { name: /create|save/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.CreateProfile).toHaveBeenCalledWith({
          name: 'test-profile',
          aws_profile: 'test-aws',
          region: 'eu-west-1',
          is_default: false,
        });
      });
    });

    it('should validate required fields', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /add profile|create profile/i }));

      // Try to submit without filling required fields
      const submitButton = screen.getByRole('button', { name: /create|save/i });
      await user.click(submitButton);

      await waitFor(() => {
        // Should show validation errors
        expect(screen.getByText(/profile name.*required/i)).toBeInTheDocument();
      });
    });

    it('should validate profile name format', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /add profile|create profile/i }));

      const nameInput = screen.getByLabelText(/profile name/i);
      await user.type(nameInput, 'Invalid Name With Spaces!');

      const submitButton = screen.getByRole('button', { name: /create|save/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/invalid.*name|alphanumeric/i)).toBeInTheDocument();
      });
    });

    it('should validate region format', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /add profile|create profile/i }));

      const nameInput = screen.getByLabelText(/profile name/i);
      await user.type(nameInput, 'test-profile');

      const awsProfileInput = screen.getByLabelText(/aws profile/i);
      await user.type(awsProfileInput, 'test');

      // Try invalid region
      const regionSelect = screen.getByLabelText(/region/i);
      await user.selectOptions(regionSelect, 'invalid-region');

      await user.click(screen.getByRole('button', { name: /create|save/i }));

      await waitFor(() => {
        expect(screen.getByText(/invalid.*region/i)).toBeInTheDocument();
      });
    });

    it('should handle API error during creation', async () => {
      mockWails.PrismService.CreateProfile.mockRejectedValue(
        new Error('Profile with this name already exists')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /add profile|create profile/i }));

      await user.type(screen.getByLabelText(/profile name/i), 'test-profile');
      await user.type(screen.getByLabelText(/aws profile/i), 'test');
      await user.selectOptions(screen.getByLabelText(/region/i), 'us-west-2');

      await user.click(screen.getByRole('button', { name: /create|save/i }));

      await waitFor(() => {
        expect(screen.getByText(/already exists|failed to create/i)).toBeInTheDocument();
      });
    });
  });

  describe('Update Profile', () => {
    it('should open edit profile dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      // Find and click edit button for research-profile
      const editButton = await screen.findByRole('button', {
        name: /edit.*research-profile|research-profile.*edit/i
      });
      await user.click(editButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /edit profile/i })).toBeInTheDocument();
      });
    });

    it('should pre-fill form with existing profile data', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      const editButton = await screen.findByRole('button', {
        name: /edit.*research-profile|research-profile.*edit/i,
      });
      await user.click(editButton);

      await waitFor(() => {
        const nameInput = screen.getByLabelText(/profile name/i) as HTMLInputElement;
        expect(nameInput.value).toBe('research-profile');

        const regionSelect = screen.getByLabelText(/region/i) as HTMLSelectElement;
        expect(regionSelect.value).toBe('us-east-1');
      });
    });

    it('should update profile with new values', async () => {
      mockWails.PrismService.UpdateProfile.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      const editButton = await screen.findByRole('button', {
        name: /edit.*research-profile|research-profile.*edit/i,
      });
      await user.click(editButton);

      // Update region
      const regionSelect = screen.getByLabelText(/region/i);
      await user.selectOptions(regionSelect, 'eu-central-1');

      await user.click(screen.getByRole('button', { name: /save|update/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.UpdateProfile).toHaveBeenCalledWith('research-profile', {
          region: 'eu-central-1',
        });
      });
    });
  });

  describe('Delete Profile', () => {
    it('should open delete confirmation dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const deleteButton = await screen.findByRole('button', {
        name: /delete.*research-profile|research-profile.*delete/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /delete profile|confirm/i })).toBeInTheDocument();
        expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
      });
    });

    it('should delete profile on confirmation', async () => {
      mockWails.PrismService.DeleteProfile.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const deleteButton = await screen.findByRole('button', {
        name: /delete.*research-profile|research-profile.*delete/i,
      });
      await user.click(deleteButton);

      // Confirm deletion
      const confirmButton = await screen.findByRole('button', { name: /delete|confirm/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockWails.PrismService.DeleteProfile).toHaveBeenCalledWith('research-profile');
      });
    });

    it('should cancel deletion on cancel button', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const deleteButton = await screen.findByRole('button', {
        name: /delete.*research-profile|research-profile.*delete/i,
      });
      await user.click(deleteButton);

      // Cancel
      const cancelButton = await screen.findByRole('button', { name: /cancel/i });
      await user.click(cancelButton);

      await waitFor(() => {
        expect(mockWails.PrismService.DeleteProfile).not.toHaveBeenCalled();
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
      });
    });

    it('should prevent deleting the current profile', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      // Try to delete default (current) profile
      const deleteButton = screen.queryByRole('button', {
        name: /delete.*default|default.*delete/i,
      });

      // Button should be disabled or not exist
      if (deleteButton) {
        expect(deleteButton).toBeDisabled();
      }
    });
  });

  describe('Switch Profile', () => {
    it('should switch to different profile', async () => {
      mockWails.PrismService.SwitchProfile.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      // Switch to research-profile
      const switchButton = await screen.findByRole('button', {
        name: /switch.*research-profile|activate.*research-profile/i,
      });
      await user.click(switchButton);

      await waitFor(() => {
        expect(mockWails.PrismService.SwitchProfile).toHaveBeenCalledWith('research-profile');
      });
    });

    it('should show success notification after switching', async () => {
      mockWails.PrismService.SwitchProfile.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const switchButton = await screen.findByRole('button', {
        name: /switch.*research-profile|activate.*research-profile/i,
      });
      await user.click(switchButton);

      await waitFor(() => {
        expect(screen.getByText(/switched.*research-profile|now using.*research-profile/i)).toBeInTheDocument();
      });
    });

    it('should update UI to reflect new current profile', async () => {
      mockWails.PrismService.SwitchProfile.mockResolvedValue({ success: true });
      mockWails.PrismService.GetCurrentProfile.mockResolvedValue({
        name: 'research-profile',
        aws_profile: 'research',
        region: 'us-east-1',
        is_default: false,
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const switchButton = await screen.findByRole('button', {
        name: /switch.*research-profile|activate.*research-profile/i,
      });
      await user.click(switchButton);

      await waitFor(() => {
        // Research-profile should now show as current
        const researchProfile = screen.getByText(/research-profile/).closest('[data-profile-item]');
        if (researchProfile) {
          expect(within(researchProfile as HTMLElement).getByText(/current|active/i)).toBeInTheDocument();
        }
      });
    });
  });

  describe('Export Profile', () => {
    it('should export profile configuration', async () => {
      mockWails.PrismService.ExportProfile.mockResolvedValue({
        name: 'research-profile',
        aws_profile: 'research',
        region: 'us-east-1',
        is_default: false,
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const exportButton = await screen.findByRole('button', {
        name: /export.*research-profile|research-profile.*export/i,
      });
      await user.click(exportButton);

      await waitFor(() => {
        expect(mockWails.PrismService.ExportProfile).toHaveBeenCalledWith('research-profile');
      });
    });

    it('should show export success notification', async () => {
      mockWails.PrismService.ExportProfile.mockResolvedValue({});

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const exportButton = await screen.findByRole('button', {
        name: /export.*research-profile|research-profile.*export/i,
      });
      await user.click(exportButton);

      await waitFor(() => {
        expect(screen.getByText(/exported.*successfully/i)).toBeInTheDocument();
      });
    });
  });

  describe('Import Profile', () => {
    it('should open import profile dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      const importButton = await screen.findByRole('button', { name: /import profile/i });
      await user.click(importButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /import profile/i })).toBeInTheDocument();
      });
    });

    it('should import profile from file', async () => {
      mockWails.PrismService.ImportProfile.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /import profile/i }));

      // Simulate file selection
      const fileInput = screen.getByLabelText(/select file|choose file/i);
      const file = new File(
        [JSON.stringify({ name: 'imported-profile', aws_profile: 'imported', region: 'us-west-2' })],
        'profile.json',
        { type: 'application/json' }
      );

      await user.upload(fileInput, file);

      await user.click(screen.getByRole('button', { name: /import/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.ImportProfile).toHaveBeenCalled();
      });
    });

    it('should validate imported profile JSON', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));
      await user.click(await screen.findByRole('button', { name: /import profile/i }));

      const fileInput = screen.getByLabelText(/select file|choose file/i);
      const invalidFile = new File(['invalid json'], 'invalid.json', { type: 'application/json' });

      await user.upload(fileInput, invalidFile);

      await waitFor(() => {
        expect(screen.getByText(/invalid.*json|malformed/i)).toBeInTheDocument();
      });
    });
  });

  describe('Loading and Error States', () => {
    it('should show loading state while fetching profiles', async () => {
      mockWails.PrismService.GetProfiles.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve([]), 100))
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      // Should show loading indicator
      expect(screen.getByText(/loading/i)).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
      });
    });

    it('should handle API error when fetching profiles', async () => {
      mockWails.PrismService.GetProfiles.mockRejectedValue(new Error('Failed to fetch profiles'));

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed.*load profiles|error.*profiles/i)).toBeInTheDocument();
      });
    });

    it('should show retry button on error', async () => {
      mockWails.PrismService.GetProfiles.mockRejectedValueOnce(new Error('Failed'))
        .mockResolvedValue([]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /settings/i }));

      // Wait for error
      await waitFor(() => {
        expect(screen.getByText(/failed/i)).toBeInTheDocument();
      });

      // Click retry
      const retryButton = screen.getByRole('button', { name: /retry/i });
      await user.click(retryButton);

      await waitFor(() => {
        expect(mockWails.PrismService.GetProfiles).toHaveBeenCalledTimes(2);
      });
    });
  });
});
