/**
 * Storage Manager Component Tests
 *
 * Tests the Storage Management functionality within the Prism GUI App.
 * This tests the "Storage" tab including EFS volumes and EBS storage management.
 *
 * Coverage:
 * - EFS volume table and actions
 * - EBS storage table and actions
 * - Create/delete workflows with modals
 * - Mount/unmount operations
 * - Error handling and loading states
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
    GetStorageVolumes: vi.fn(),
    GetEBSStorages: vi.fn(),
    CreateEFSVolume: vi.fn(),
    CreateEBSStorage: vi.fn(),
    DeleteEFSVolume: vi.fn(),
    DeleteEBSStorage: vi.fn(),
    MountEFSVolume: vi.fn(),
    UnmountEFSVolume: vi.fn(),
    AttachEBSStorage: vi.fn(),
    DetachEBSStorage: vi.fn(),
    GetProfiles: vi.fn(),
    GetIdlePolicies: vi.fn(),
    GetProjects: vi.fn(),
    GetUsers: vi.fn(),
  },
};

Object.defineProperty(window, 'wails', {
  value: mockWails,
  writable: true,
});

describe('StorageManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock responses
    mockWails.PrismService.GetTemplates.mockResolvedValue([]);
    mockWails.PrismService.GetInstances.mockResolvedValue([]);
    mockWails.PrismService.GetProfiles.mockResolvedValue([]);
    mockWails.PrismService.GetIdlePolicies.mockResolvedValue([]);
    mockWails.PrismService.GetProjects.mockResolvedValue([]);
    mockWails.PrismService.GetUsers.mockResolvedValue([]);

    // Storage-specific mocks
    mockWails.PrismService.GetStorageVolumes.mockResolvedValue([
      {
        name: 'shared-data',
        type: 'shared',
        aws_service: 'efs',
        region: 'us-west-2',
        state: 'available',
        filesystem_id: 'fs-1234567890abcdef0',
        size_gb: 50,
        performance_mode: 'generalPurpose',
        throughput_mode: 'bursting',
        mount_targets: ['subnet-12345', 'subnet-67890'],
        created_at: '2025-01-01T00:00:00Z',
      },
      {
        name: 'ml-workspace',
        type: 'workspace',
        aws_service: 'efs',
        region: 'us-west-2',
        state: 'in-use',
        filesystem_id: 'fs-abcdef1234567890',
        size_gb: 100,
        performance_mode: 'maxIO',
        throughput_mode: 'provisioned',
        mount_targets: ['subnet-12345'],
        created_at: '2025-01-10T00:00:00Z',
      },
    ]);

    mockWails.PrismService.GetEBSStorages.mockResolvedValue([
      {
        volume_id: 'vol-1234567890abcdef0',
        name: 'project-storage-L',
        state: 'available',
        size_gb: 500,
        volume_type: 'gp3',
        iops: 3000,
        throughput: 125,
        encrypted: true,
        availability_zone: 'us-west-2a',
        attached_to: null,
        created_at: '2025-01-05T00:00:00Z',
      },
      {
        volume_id: 'vol-abcdef1234567890',
        name: 'data-backup-XL',
        state: 'in-use',
        size_gb: 1000,
        volume_type: 'gp3',
        iops: 3000,
        throughput: 125,
        encrypted: true,
        availability_zone: 'us-west-2b',
        attached_to: 'i-test123',
        created_at: '2025-01-12T00:00:00Z',
      },
    ]);
  });

  describe('EFS Volume List Rendering', () => {
    it('should display list of EFS volumes', async () => {
      render(<App />);
      const user = userEvent.setup();

      // Navigate to Storage tab
      const storageTab = screen.getByRole('link', { name: /storage/i });
      await user.click(storageTab);

      await waitFor(() => {
        expect(screen.getByText('shared-data')).toBeInTheDocument();
        expect(screen.getByText('ml-workspace')).toBeInTheDocument();
      });
    });

    it('should display volume details (size, performance mode, state)', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      await waitFor(() => {
        expect(screen.getByText(/50.*GB/i)).toBeInTheDocument();
        expect(screen.getByText(/100.*GB/i)).toBeInTheDocument();
        expect(screen.getByText(/generalPurpose|maxIO/i)).toBeInTheDocument();
        expect(screen.getByText(/available|in-use/i)).toBeInTheDocument();
      });
    });

    it('should show filesystem IDs', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      await waitFor(() => {
        expect(screen.getByText('fs-1234567890abcdef0')).toBeInTheDocument();
        expect(screen.getByText('fs-abcdef1234567890')).toBeInTheDocument();
      });
    });

    it('should handle empty EFS volume list', async () => {
      mockWails.PrismService.GetStorageVolumes.mockResolvedValue([]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      await waitFor(() => {
        expect(screen.getByText(/no.*volumes|create your first volume/i)).toBeInTheDocument();
      });
    });
  });

  describe('EBS Storage List Rendering', () => {
    it('should display list of EBS volumes', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      // Switch to EBS tab if needed
      const ebsTab = screen.queryByRole('tab', { name: /ebs|block storage/i });
      if (ebsTab) {
        await user.click(ebsTab);
      }

      await waitFor(() => {
        expect(screen.getByText('project-storage-L')).toBeInTheDocument();
        expect(screen.getByText('data-backup-XL')).toBeInTheDocument();
      });
    });

    it('should display EBS volume details (size, type, IOPS)', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      await waitFor(() => {
        expect(screen.getByText(/500.*GB/i)).toBeInTheDocument();
        expect(screen.getByText(/1000.*GB|1.*TB/i)).toBeInTheDocument();
        expect(screen.getByText(/gp3/i)).toBeInTheDocument();
        expect(screen.getByText(/3000.*IOPS/i)).toBeInTheDocument();
      });
    });

    it('should show attachment status', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      await waitFor(() => {
        // available volume
        expect(screen.getByText(/available|not attached/i)).toBeInTheDocument();
        // attached volume
        expect(screen.getByText(/attached.*i-test123|in-use/i)).toBeInTheDocument();
      });
    });
  });

  describe('Create EFS Volume', () => {
    it('should open create EFS volume dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const createButton = await screen.findByRole('button', {
        name: /create.*efs|add.*volume/i,
      });
      await user.click(createButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /create.*efs volume/i })).toBeInTheDocument();
      });
    });

    it('should create EFS volume with valid input', async () => {
      mockWails.PrismService.CreateEFSVolume.mockResolvedValue({
        filesystem_id: 'fs-newvolume123',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));
      await user.click(await screen.findByRole('button', { name: /create.*efs/i }));

      // Fill form
      await user.type(screen.getByLabelText(/volume name/i), 'test-efs-volume');
      await user.selectOptions(screen.getByLabelText(/performance mode/i), 'generalPurpose');
      await user.selectOptions(screen.getByLabelText(/throughput mode/i), 'bursting');

      await user.click(screen.getByRole('button', { name: /create/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.CreateEFSVolume).toHaveBeenCalledWith({
          name: 'test-efs-volume',
          performance_mode: 'generalPurpose',
          throughput_mode: 'bursting',
        });
      });
    });

    it('should validate required fields for EFS', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));
      await user.click(await screen.findByRole('button', { name: /create.*efs/i }));

      // Try to submit without name
      await user.click(screen.getByRole('button', { name: /create/i }));

      await waitFor(() => {
        expect(screen.getByText(/volume name.*required/i)).toBeInTheDocument();
      });
    });

    it('should show success notification after creating EFS', async () => {
      mockWails.PrismService.CreateEFSVolume.mockResolvedValue({
        filesystem_id: 'fs-new123',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));
      await user.click(await screen.findByRole('button', { name: /create.*efs/i }));

      await user.type(screen.getByLabelText(/volume name/i), 'new-volume');
      await user.click(screen.getByRole('button', { name: /create/i }));

      await waitFor(() => {
        expect(screen.getByText(/created.*successfully|volume.*created/i)).toBeInTheDocument();
      });
    });
  });

  describe('Create EBS Storage', () => {
    it('should open create EBS storage dialog', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const createButton = await screen.findByRole('button', {
        name: /create.*ebs|add.*ebs/i,
      });
      await user.click(createButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /create.*ebs/i })).toBeInTheDocument();
      });
    });

    it('should create EBS volume with size selection', async () => {
      mockWails.PrismService.CreateEBSStorage.mockResolvedValue({
        volume_id: 'vol-newstorage123',
      });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));
      await user.click(await screen.findByRole('button', { name: /create.*ebs/i }));

      await user.type(screen.getByLabelText(/storage name/i), 'test-ebs-storage');
      await user.selectOptions(screen.getByLabelText(/size|capacity/i), 'L'); // 500GB
      await user.selectOptions(screen.getByLabelText(/volume type/i), 'gp3');

      await user.click(screen.getByRole('button', { name: /create/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.CreateEBSStorage).toHaveBeenCalledWith({
          name: 'test-ebs-storage',
          size: 'L',
          volume_type: 'gp3',
        });
      });
    });

    it('should show IOPS and throughput options for gp3', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));
      await user.click(await screen.findByRole('button', { name: /create.*ebs/i }));

      // Select gp3
      const volumeTypeSelect = screen.getByLabelText(/volume type/i);
      await user.selectOptions(volumeTypeSelect, 'gp3');

      // IOPS and throughput fields should appear
      await waitFor(() => {
        expect(screen.getByLabelText(/iops/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/throughput/i)).toBeInTheDocument();
      });
    });

    it('should validate EBS size selection', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));
      await user.click(await screen.findByRole('button', { name: /create.*ebs/i }));

      await user.type(screen.getByLabelText(/storage name/i), 'test');
      // Don't select size

      await user.click(screen.getByRole('button', { name: /create/i }));

      await waitFor(() => {
        expect(screen.getByText(/size.*required|select.*size/i)).toBeInTheDocument();
      });
    });
  });

  describe('Delete EFS Volume', () => {
    it('should open delete confirmation for EFS', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const deleteButton = await screen.findByRole('button', {
        name: /delete.*shared-data/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog', { name: /delete.*volume|confirm/i })).toBeInTheDocument();
        expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
      });
    });

    it('should delete EFS volume on confirmation', async () => {
      mockWails.PrismService.DeleteEFSVolume.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const deleteButton = await screen.findByRole('button', {
        name: /delete.*shared-data/i,
      });
      await user.click(deleteButton);

      const confirmButton = await screen.findByRole('button', { name: /delete|confirm/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockWails.PrismService.DeleteEFSVolume).toHaveBeenCalledWith('fs-1234567890abcdef0');
      });
    });

    it('should prevent deleting in-use EFS volume', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      // Try to delete ml-workspace (in-use)
      const deleteButton = screen.queryByRole('button', {
        name: /delete.*ml-workspace/i,
      });

      // Button should be disabled or show warning
      if (deleteButton) {
        await user.click(deleteButton);

        await waitFor(() => {
          expect(screen.getByText(/in use|cannot delete.*in use/i)).toBeInTheDocument();
        });
      }
    });
  });

  describe('Delete EBS Storage', () => {
    it('should delete EBS storage on confirmation', async () => {
      mockWails.PrismService.DeleteEBSStorage.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const deleteButton = await screen.findByRole('button', {
        name: /delete.*project-storage-L/i,
      });
      await user.click(deleteButton);

      const confirmButton = await screen.findByRole('button', { name: /delete|confirm/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockWails.PrismService.DeleteEBSStorage).toHaveBeenCalledWith('vol-1234567890abcdef0');
      });
    });

    it('should prevent deleting attached EBS volume', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      // Try to delete data-backup-XL (attached)
      const deleteButton = screen.queryByRole('button', {
        name: /delete.*data-backup-XL/i,
      });

      if (deleteButton) {
        await user.click(deleteButton);

        await waitFor(() => {
          expect(screen.getByText(/attached|detach.*first/i)).toBeInTheDocument();
        });
      }
    });
  });

  describe('Mount/Unmount EFS Operations', () => {
    it('should mount EFS volume to instance', async () => {
      mockWails.PrismService.MountEFSVolume.mockResolvedValue({ success: true });
      mockWails.PrismService.GetInstances.mockResolvedValue([
        { id: 'i-test123', name: 'my-instance', state: 'running' },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const mountButton = await screen.findByRole('button', {
        name: /mount.*shared-data/i,
      });
      await user.click(mountButton);

      // Select instance
      const instanceSelect = await screen.findByLabelText(/select instance/i);
      await user.selectOptions(instanceSelect, 'i-test123');

      await user.click(screen.getByRole('button', { name: /mount/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.MountEFSVolume).toHaveBeenCalledWith({
          filesystem_id: 'fs-1234567890abcdef0',
          instance_id: 'i-test123',
        });
      });
    });

    it('should unmount EFS volume from instance', async () => {
      mockWails.PrismService.UnmountEFSVolume.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      // ml-workspace is in-use (mounted)
      const unmountButton = await screen.findByRole('button', {
        name: /unmount.*ml-workspace/i,
      });
      await user.click(unmountButton);

      const confirmButton = await screen.findByRole('button', { name: /unmount|confirm/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockWails.PrismService.UnmountEFSVolume).toHaveBeenCalledWith('fs-abcdef1234567890');
      });
    });

    it('should show mount target information', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      await waitFor(() => {
        // shared-data has 2 mount targets
        expect(screen.getByText(/2.*mount targets|subnet-12345/i)).toBeInTheDocument();
      });
    });
  });

  describe('Attach/Detach EBS Operations', () => {
    it('should attach EBS volume to instance', async () => {
      mockWails.PrismService.AttachEBSStorage.mockResolvedValue({ success: true });
      mockWails.PrismService.GetInstances.mockResolvedValue([
        { id: 'i-test456', name: 'my-instance', state: 'running' },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const attachButton = await screen.findByRole('button', {
        name: /attach.*project-storage-L/i,
      });
      await user.click(attachButton);

      // Select instance
      const instanceSelect = await screen.findByLabelText(/select instance/i);
      await user.selectOptions(instanceSelect, 'i-test456');

      // Select device name
      const deviceInput = screen.getByLabelText(/device name/i);
      await user.type(deviceInput, '/dev/sdf');

      await user.click(screen.getByRole('button', { name: /attach/i }));

      await waitFor(() => {
        expect(mockWails.PrismService.AttachEBSStorage).toHaveBeenCalledWith({
          volume_id: 'vol-1234567890abcdef0',
          instance_id: 'i-test456',
          device: '/dev/sdf',
        });
      });
    });

    it('should detach EBS volume from instance', async () => {
      mockWails.PrismService.DetachEBSStorage.mockResolvedValue({ success: true });

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      // data-backup-XL is attached
      const detachButton = await screen.findByRole('button', {
        name: /detach.*data-backup-XL/i,
      });
      await user.click(detachButton);

      const confirmButton = await screen.findByRole('button', { name: /detach|confirm/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockWails.PrismService.DetachEBSStorage).toHaveBeenCalledWith('vol-abcdef1234567890');
      });
    });

    it('should validate device name for attach', async () => {
      mockWails.PrismService.GetInstances.mockResolvedValue([
        { id: 'i-test456', name: 'my-instance', state: 'running' },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const attachButton = await screen.findByRole('button', {
        name: /attach.*project-storage-L/i,
      });
      await user.click(attachButton);

      const instanceSelect = await screen.findByLabelText(/select instance/i);
      await user.selectOptions(instanceSelect, 'i-test456');

      // Invalid device name
      const deviceInput = screen.getByLabelText(/device name/i);
      await user.type(deviceInput, 'invalid');

      await user.click(screen.getByRole('button', { name: /attach/i }));

      await waitFor(() => {
        expect(screen.getByText(/invalid.*device|must start with/i)).toBeInTheDocument();
      });
    });
  });

  describe('Loading and Error States', () => {
    it('should show loading state while fetching storage', async () => {
      mockWails.PrismService.GetStorageVolumes.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve([]), 100))
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      expect(screen.getByText(/loading/i)).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
      });
    });

    it('should handle API error when fetching storage', async () => {
      mockWails.PrismService.GetStorageVolumes.mockRejectedValue(
        new Error('Failed to fetch storage')
      );

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed.*load storage|error.*storage/i)).toBeInTheDocument();
      });
    });

    it('should handle mount operation error', async () => {
      mockWails.PrismService.MountEFSVolume.mockRejectedValue(
        new Error('Mount failed: Instance not in correct state')
      );
      mockWails.PrismService.GetInstances.mockResolvedValue([
        { id: 'i-test123', name: 'my-instance', state: 'running' },
      ]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const mountButton = await screen.findByRole('button', { name: /mount.*shared-data/i });
      await user.click(mountButton);

      const instanceSelect = await screen.findByLabelText(/select instance/i);
      await user.selectOptions(instanceSelect, 'i-test123');

      await user.click(screen.getByRole('button', { name: /mount/i }));

      await waitFor(() => {
        expect(screen.getByText(/mount failed|instance not in correct state/i)).toBeInTheDocument();
      });
    });

    it('should show retry button on error', async () => {
      mockWails.PrismService.GetStorageVolumes.mockRejectedValueOnce(new Error('Failed'))
        .mockResolvedValue([]);

      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed/i)).toBeInTheDocument();
      });

      const retryButton = screen.getByRole('button', { name: /retry/i });
      await user.click(retryButton);

      await waitFor(() => {
        expect(mockWails.PrismService.GetStorageVolumes).toHaveBeenCalledTimes(2);
      });
    });
  });

  describe('Storage Filtering and Search', () => {
    it('should filter storage by type', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      // Filter by workspace type
      const filterSelect = await screen.findByLabelText(/filter by type/i);
      await user.selectOptions(filterSelect, 'workspace');

      await waitFor(() => {
        expect(screen.getByText('ml-workspace')).toBeInTheDocument();
        expect(screen.queryByText('shared-data')).not.toBeInTheDocument();
      });
    });

    it('should search storage by name', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const searchInput = await screen.findByPlaceholderText(/search.*storage|find.*volume/i);
      await user.type(searchInput, 'ml');

      await waitFor(() => {
        expect(screen.getByText('ml-workspace')).toBeInTheDocument();
        expect(screen.queryByText('shared-data')).not.toBeInTheDocument();
      });
    });

    it('should filter by state (available, in-use)', async () => {
      render(<App />);
      const user = userEvent.setup();

      await user.click(screen.getByRole('link', { name: /storage/i }));

      const stateFilter = await screen.findByLabelText(/filter by state|state/i);
      await user.selectOptions(stateFilter, 'available');

      await waitFor(() => {
        expect(screen.getByText('shared-data')).toBeInTheDocument();
        expect(screen.queryByText('ml-workspace')).not.toBeInTheDocument();
      });
    });
  });
});
