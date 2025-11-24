/**
 * Backup Manager Component Tests
 *
 * Tests the Backup & Snapshot Management functionality within the Prism GUI App.
 * This tests backup CRUD, restore workflows, snapshot management, and validation.
 *
 * Coverage:
 * - Backup list and creation
 * - Restore workflow
 * - Snapshot management
 * - Validation and errors
 * - Incremental backups
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
    GetBackups: vi.fn(),
    CreateBackup: vi.fn(),
    DeleteBackup: vi.fn(),
    RestoreBackup: vi.fn(),
    CloneFromBackup: vi.fn(),
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

describe('BackupManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    mockWails.PrismService.GetTemplates.mockResolvedValue([]);
    mockWails.PrismService.GetProfiles.mockResolvedValue([]);
    mockWails.PrismService.GetStorageVolumes.mockResolvedValue([]);
    mockWails.PrismService.GetIdlePolicies.mockResolvedValue([]);
    mockWails.PrismService.GetProjects.mockResolvedValue([]);
    mockWails.PrismService.GetUsers.mockResolvedValue([]);

    mockWails.PrismService.GetInstances.mockResolvedValue([
      { id: 'i-test', name: 'ml-research', state: 'running' },
    ]);

    mockWails.PrismService.GetBackups.mockResolvedValue([
      {
        id: 'snap-full-12345',
        instance_id: 'i-test',
        name: 'ml-research-backup-full',
        created_at: '2025-01-15T10:00:00Z',
        size_gb: 50,
        status: 'available',
        type: 'full',
        cost_gb: 0.05,
      },
      {
        id: 'snap-incr-67890',
        instance_id: 'i-test',
        name: 'ml-research-backup-incremental',
        created_at: '2025-01-16T10:00:00Z',
        size_gb: 5,
        status: 'available',
        type: 'incremental',
        cost_gb: 0.05,
      },
    ]);
  });

  describe('Backup List Rendering', () => {
    it('should display list of backups', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      await waitFor(() => {
        expect(screen.getByText('ml-research-backup-full')).toBeInTheDocument();
        expect(screen.getByText('ml-research-backup-incremental')).toBeInTheDocument();
      });
    });

    it('should display backup details (size, type, status, date)', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      await waitFor(() => {
        expect(screen.getByText(/50.*GB/i)).toBeInTheDocument();
        expect(screen.getByText(/5.*GB/i)).toBeInTheDocument();
        expect(screen.getByText(/full/i)).toBeInTheDocument();
        expect(screen.getByText(/incremental/i)).toBeInTheDocument();
        expect(screen.getByText(/available/i)).toBeInTheDocument();
      });
    });

    it('should show monthly cost for each backup', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      await waitFor(() => {
        // 50GB * $0.05 = $2.50/month
        expect(screen.getByText(/\$2\.50.*month|\$2\.50\/mo/i)).toBeInTheDocument();
        // 5GB * $0.05 = $0.25/month
        expect(screen.getByText(/\$0\.25.*month|\$0\.25\/mo/i)).toBeInTheDocument();
      });
    });

    it('should handle empty backup list', async () => {
      mockWails.PrismService.GetBackups.mockResolvedValue([]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      await waitFor(() => {
        expect(screen.getByText(/no.*backups|create your first backup/i)).toBeInTheDocument();
      });
    });
  });

  describe('Create Backup', () => {
    it('should open create backup dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      const createButton = await screen.findByRole('button', { name: /create.*backup/i });
      await user.click(createButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /create.*backup/i })).toBeInTheDocument();
      });
    });

    it('should create full backup', async () => {
      mockWails.PrismService.CreateBackup.mockResolvedValue({
        id: 'snap-new',
        status: 'creating',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));
      await user.click(await screen.findByRole('button', { name: /create.*backup/i }));

      const instanceSelect = screen.getByLabelText(/select instance/i);
      await user.selectOptions(instanceSelect, 'i-test');

      await user.type(screen.getByLabelText(/backup name/i), 'my-full-backup');
      await user.selectOptions(screen.getByLabelText(/backup type/i), 'full');

      await user.click(screen.getByRole('button', { name: /create|start/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.CreateBackup).toHaveBeenCalledWith({
          instance_id: 'i-test',
          name: 'my-full-backup',
          type: 'full',
        });
      });
    });

    it('should create incremental backup', async () => {
      mockWails.PrismService.CreateBackup.mockResolvedValue({ id: 'snap-incr' });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));
      await user.click(await screen.findByRole('button', { name: /create.*backup/i }));

      await user.selectOptions(screen.getByLabelText(/select instance/i), 'i-test');
      await user.type(screen.getByLabelText(/backup name/i), 'my-incremental');
      await user.selectOptions(screen.getByLabelText(/backup type/i), 'incremental');

      await user.click(screen.getByRole('button', { name: /create|start/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.CreateBackup).toHaveBeenCalledWith({
          instance_id: 'i-test',
          name: 'my-incremental',
          type: 'incremental',
        });
      });
    });

    it('should show estimated backup size', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));
      await user.click(await screen.findByRole('button', { name: /create.*backup/i }));

      await user.selectOptions(screen.getByLabelText(/select instance/i), 'i-test');

      // Should show size estimate
      await waitFor(() => {
        expect(screen.getByText(/estimated.*size|approximately.*GB/i)).toBeInTheDocument();
      });
    });

    it('should validate backup name', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));
      await user.click(await screen.findByRole('button', { name: /create.*backup/i }));

      await user.selectOptions(screen.getByLabelText(/select instance/i), 'i-test');
      // Don't enter name

      await user.click(screen.getByRole('button', { name: /create|start/i }));

      await waitFor(() => {
        expect(screen.getByText(/backup name.*required/i)).toBeInTheDocument();
      });
    });
  });

  describe('Delete Backup', () => {
    it('should delete backup on confirmation', async () => {
      mockWails.PrismService.DeleteBackup.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      const deleteButton = await screen.findByRole('button', {
        name: /delete.*ml-research-backup-full/i,
      });
      await user.click(deleteButton);

      await user.click(await screen.findByRole('button', { name: /delete|confirm/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.DeleteBackup).toHaveBeenCalledWith('snap-full-12345');
      });
    });

    it('should show cost savings after deletion', async () => {
      mockWails.PrismService.DeleteBackup.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      const deleteButton = await screen.findByRole('button', {
        name: /delete.*ml-research-backup-full/i,
      });
      await user.click(deleteButton);

      // Should show how much will be saved
      await waitFor(() => {
        expect(screen.getByText(/save.*\$2\.50|free.*50.*GB/i)).toBeInTheDocument();
      });
    });
  });

  describe('Restore Backup', () => {
    it('should open restore dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      const restoreButton = await screen.findByRole('button', {
        name: /restore.*ml-research-backup-full/i,
      });
      await user.click(restoreButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /restore.*backup/i })).toBeInTheDocument();
      });
    });

    it('should restore backup to new instance', async () => {
      mockWails.PrismService.RestoreBackup.mockResolvedValue({
        instance_id: 'i-restored',
        status: 'pending',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));
      await user.click(await screen.findByRole('button', { name: /restore.*ml-research-backup-full/i }));

      await user.type(screen.getByLabelText(/new instance name/i), 'restored-ml-research');

      await user.click(screen.getByRole('button', { name: /restore|start/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.RestoreBackup).toHaveBeenCalledWith({
          backup_id: 'snap-full-12345',
          instance_name: 'restored-ml-research',
        });
      });
    });

    it('should warn about restore time', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));
      await user.click(await screen.findByRole('button', { name: /restore/i }));

      await waitFor(() => {
        expect(screen.getByText(/may take.*minutes|restore.*time/i)).toBeInTheDocument();
      });
    });
  });

  describe('Clone from Backup', () => {
    it('should clone instance from backup', async () => {
      mockWails.PrismService.CloneFromBackup.mockResolvedValue({
        instance_id: 'i-clone',
        status: 'pending',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      const cloneButton = await screen.findByRole('button', {
        name: /clone.*ml-research-backup-full/i,
      });
      await user.click(cloneButton);

      await user.type(screen.getByLabelText(/clone.*name/i), 'ml-research-clone');

      await user.click(screen.getByRole('button', { name: /clone|create/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.CloneFromBackup).toHaveBeenCalledWith({
          backup_id: 'snap-full-12345',
          instance_name: 'ml-research-clone',
        });
      });
    });
  });

  describe('Backup Filtering', () => {
    it('should filter by backup type', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      const filterSelect = await screen.findByLabelText(/filter by type/i);
      await user.selectOptions(filterSelect, 'full');

      await waitFor(() => {
        expect(screen.getByText('ml-research-backup-full')).toBeInTheDocument();
        expect(screen.queryByText('ml-research-backup-incremental')).not.toBeInTheDocument();
      });
    });

    it('should search backups by name', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      const searchInput = await screen.findByPlaceholderText(/search.*backups/i);
      await user.type(searchInput, 'incremental');

      await waitFor(() => {
        expect(screen.getByText('ml-research-backup-incremental')).toBeInTheDocument();
        expect(screen.queryByText('ml-research-backup-full')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle fetch error', async () => {
      mockWails.PrismService.GetBackups.mockRejectedValue(new Error('Failed to fetch backups'));

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed.*load.*backups/i)).toBeInTheDocument();
      });
    });

    it('should handle backup creation error', async () => {
      mockWails.PrismService.CreateBackup.mockRejectedValue(
        new Error('Insufficient permissions')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));
      await user.click(await screen.findByRole('button', { name: /create.*backup/i }));

      await user.selectOptions(screen.getByLabelText(/select instance/i), 'i-test');
      await user.type(screen.getByLabelText(/backup name/i), 'test');
      await user.click(screen.getByRole('button', { name: /create/i }));

      await waitFor(() => {
        expect(screen.getByText(/insufficient permissions|failed to create/i)).toBeInTheDocument();
      });
    });

    it('should handle restore error', async () => {
      mockWails.PrismService.RestoreBackup.mockRejectedValue(
        new Error('Backup is corrupted')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /backups/i }));
      await user.click(await screen.findByRole('button', { name: /restore/i }));

      await user.type(screen.getByLabelText(/new instance name/i), 'restored');
      await user.click(screen.getByRole('button', { name: /restore/i }));

      await waitFor(() => {
        expect(screen.getByText(/corrupted|failed to restore/i)).toBeInTheDocument();
      });
    });
  });
});
