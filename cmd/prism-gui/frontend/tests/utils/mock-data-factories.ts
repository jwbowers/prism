/**
 * Mock Data Factories for Prism GUI Tests
 *
 * Provides consistent mock data for testing across all test types.
 * Use these factories to create test data instead of inline objects.
 */

import type { Template, Instance, Profile, Volume, EBSStorage } from '../../src/types';

/**
 * Template Mock Data Factory
 */
export const createMockTemplate = (overrides?: Partial<Template>): Template => ({
  Name: 'Python Machine Learning',
  Description: 'Complete ML environment with TensorFlow, PyTorch, and Jupyter',
  Category: 'Machine Learning',
  Domain: 'ml',
  Complexity: 'moderate',
  Icon: '🤖',
  Popular: true,
  EstimatedLaunchTime: 2,
  EstimatedCostPerHour: { 'x86_64': 0.48, 'arm64': 0.38 },
  ValidationStatus: 'validated',
  Tags: ['python', 'ml', 'jupyter', 'tensorflow', 'pytorch'],
  ...overrides,
});

export const createMockTemplates = (): Template[] => [
  createMockTemplate({
    Name: 'Python Machine Learning',
    Domain: 'ml',
    Complexity: 'moderate',
    Icon: '🤖',
  }),
  createMockTemplate({
    Name: 'R Research Environment',
    Description: 'Statistical computing with R, RStudio, and tidyverse packages',
    Category: 'Data Science',
    Domain: 'datascience',
    Complexity: 'simple',
    Icon: '📊',
    EstimatedLaunchTime: 3,
    EstimatedCostPerHour: { 'x86_64': 0.24, 'arm64': 0.19 },
  }),
  createMockTemplate({
    Name: 'Web Development',
    Description: 'Node.js, React, TypeScript development environment',
    Category: 'Development',
    Domain: 'webdev',
    Complexity: 'simple',
    Icon: '🌐',
    EstimatedLaunchTime: 2,
    EstimatedCostPerHour: { 'x86_64': 0.17, 'arm64': 0.13 },
  }),
];

/**
 * Instance Mock Data Factory
 */
export const createMockInstance = (overrides?: Partial<Instance>): Instance => ({
  id: 'i-1234567890abcdef0',
  name: 'my-ml-research',
  template: 'Python Machine Learning',
  status: 'running',
  public_ip: '54.123.45.67',
  private_ip: '10.0.1.42',
  instance_type: 't3.medium',
  architecture: 'x86_64',
  cost_per_hour: 0.48,
  launch_time: '2025-09-28T10:30:00Z',
  region: 'us-west-2',
  availability_zone: 'us-west-2a',
  hibernation_enabled: true,
  spot_instance: false,
  ...overrides,
});

export const createMockInstances = (): Instance[] => [
  createMockInstance({
    name: 'my-ml-research',
    template: 'Python Machine Learning',
    status: 'running',
  }),
  createMockInstance({
    id: 'i-0987654321fedcba0',
    name: 'my-r-analysis',
    template: 'R Research Environment',
    status: 'stopped',
    public_ip: undefined,
    cost_per_hour: 0.24,
  }),
  createMockInstance({
    id: 'i-abcdef1234567890',
    name: 'web-dev-environment',
    template: 'Web Development',
    status: 'hibernating',
    public_ip: undefined,
    cost_per_hour: 0.17,
    hibernation_enabled: true,
  }),
];

/**
 * Profile Mock Data Factory
 */
export const createMockProfile = (overrides?: Partial<Profile>): Profile => ({
  name: 'default',
  aws_profile: 'default',
  region: 'us-west-2',
  is_default: true,
  created_at: '2025-09-01T00:00:00Z',
  ...overrides,
});

export const createMockProfiles = (): Profile[] => [
  createMockProfile({
    name: 'default',
    region: 'us-west-2',
    is_default: true,
  }),
  createMockProfile({
    name: 'research-profile',
    aws_profile: 'research',
    region: 'us-east-1',
    is_default: false,
  }),
  createMockProfile({
    name: 'production',
    aws_profile: 'prod',
    region: 'eu-west-1',
    is_default: false,
  }),
];

/**
 * Volume (EFS) Mock Data Factory
 */
export const createMockVolume = (overrides?: Partial<Volume>): Volume => ({
  filesystem_id: 'fs-1234567890abcdef0',
  name: 'shared-data',
  state: 'available',
  size_gb: 100,
  performance_mode: 'generalPurpose',
  throughput_mode: 'bursting',
  creation_time: '2025-09-15T10:00:00Z',
  mount_targets: [],
  tags: { Name: 'shared-data', Project: 'research' },
  ...overrides,
});

export const createMockVolumes = (): Volume[] => [
  createMockVolume({
    name: 'shared-data',
    size_gb: 100,
  }),
  createMockVolume({
    filesystem_id: 'fs-0987654321fedcba0',
    name: 'project-storage',
    size_gb: 500,
    state: 'in-use',
  }),
];

/**
 * EBS Storage Mock Data Factory
 */
export const createMockEBSStorage = (overrides?: Partial<EBSStorage>): EBSStorage => ({
  volume_id: 'vol-1234567890abcdef0',
  name: 'project-storage-L',
  state: 'available',
  size_gb: 100,
  volume_type: 'gp3',
  iops: 3000,
  throughput: 125,
  encrypted: true,
  availability_zone: 'us-west-2a',
  creation_time: '2025-09-20T12:00:00Z',
  attached_to: undefined,
  device: undefined,
  ...overrides,
});

export const createMockEBSStorages = (): EBSStorage[] => [
  createMockEBSStorage({
    name: 'project-storage-L',
    size_gb: 100,
    state: 'available',
  }),
  createMockEBSStorage({
    volume_id: 'vol-0987654321fedcba0',
    name: 'large-dataset',
    size_gb: 1000,
    volume_type: 'gp3',
    state: 'in-use',
    attached_to: 'i-1234567890abcdef0',
    device: '/dev/sdf',
  }),
];

/**
 * API Response Mock Data Factories
 */
export const createMockHealthResponse = () => ({
  status: 'healthy',
  version: '0.5.16',
  uptime: 3600,
});

export const createMockLaunchResponse = () => ({
  instance_id: 'i-new1234567890abcd',
  status: 'launching',
  estimated_ready_time: 120,
});

export const createMockConnectionInfo = (overrides?: Partial<any>) => ({
  instance_id: 'i-1234567890abcdef0',
  instance_name: 'my-ml-research',
  public_ip: '54.123.45.67',
  ssh_command: 'ssh ec2-user@54.123.45.67',
  ports: [
    { port: 22, protocol: 'tcp', description: 'SSH' },
    { port: 8888, protocol: 'tcp', description: 'Jupyter' },
  ],
  web_urls: [
    { name: 'Jupyter', url: 'http://54.123.45.67:8888' },
  ],
  ...overrides,
});

/**
 * Hibernation Mock Data
 */
export const createMockHibernationStatus = (overrides?: Partial<any>) => ({
  instance_id: 'i-1234567890abcdef0',
  hibernation_enabled: true,
  hibernation_configured: true,
  can_hibernate: true,
  current_state: 'running',
  ...overrides,
});

/**
 * Idle Policy Mock Data
 */
export const createMockIdlePolicy = (overrides?: Partial<any>) => ({
  id: 'policy-123',
  name: 'gpu',
  description: 'GPU instance idle detection',
  idle_minutes: 15,
  action: 'stop',
  cpu_threshold: 10,
  memory_threshold: 20,
  network_threshold: 1,
  disk_threshold: 5,
  gpu_threshold: 10,
  enabled: true,
  ...overrides,
});

export const createMockIdlePolicies = () => [
  createMockIdlePolicy({
    name: 'gpu',
    idle_minutes: 15,
    action: 'stop',
  }),
  createMockIdlePolicy({
    id: 'policy-456',
    name: 'batch',
    description: 'Long-running batch jobs',
    idle_minutes: 60,
    action: 'hibernate',
  }),
  createMockIdlePolicy({
    id: 'policy-789',
    name: 'cost-optimized',
    description: 'Maximum cost savings',
    idle_minutes: 10,
    action: 'hibernate',
  }),
];

/**
 * Error Response Mock Data
 */
export const createMockError = (message: string = 'An error occurred', code: number = 500) => ({
  error: message,
  code,
  timestamp: new Date().toISOString(),
});
