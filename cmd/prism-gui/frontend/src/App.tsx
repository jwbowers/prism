// Prism GUI - Bulletproof AWS Integration
import { logger } from './utils/logger';
// Complete error handling, real API integration, professional UX

import React, { useState, useEffect, useRef } from 'react';
import './index.css';
import { toast } from 'sonner';
import { AppLayout as AppLayoutShell } from './components/app-layout';
import { SideNav } from './components/side-nav';
import Terminal from './Terminal';
import { ValidationError } from './components/ValidationError';
import { ProjectDetailView } from './components/ProjectDetailView';
import { InvitationManagementView } from './components/InvitationManagementView';
import { CoursesPanel } from './components/CoursesPanel';
import { WorkshopsPanel } from './components/WorkshopsPanel';
import CapacityBlocksPanel from './components/CapacityBlocksPanel';
import { SSHKeyModal } from './components/SSHKeyModal';
import { ApprovalsView as ApprovalsViewExtracted } from './views/ApprovalsView';
import { PlaceholderView } from './views/placeholder-view';
import { LogsView } from './views/LogsView';
import { WebViewView } from './views/WebViewView';
import { DashboardView as DashboardViewExtracted } from './views/DashboardView';
import { TemplateSelectionView as TemplateSelectionViewExtracted } from './views/TemplateSelectionView';
import { BackupManagementView as BackupManagementViewExtracted } from './views/BackupManagementView';
import { AMIManagementView as AMIManagementViewExtracted } from './views/AMIManagementView';
import { MarketplaceView as MarketplaceViewExtracted } from './views/MarketplaceView';
import { IdleDetectionView as IdleDetectionViewExtracted } from './views/IdleDetectionView';
import { ProfileSelectorView as ProfileSelectorViewExtracted } from './views/ProfileSelectorView';
import { InstanceManagementView as InstanceManagementViewExtracted } from './views/InstanceManagementView';
import { StorageManagementView as StorageManagementViewExtracted } from './views/StorageManagementView';
import { ProjectManagementView as ProjectManagementViewExtracted } from './views/ProjectManagementView';
import { UserManagementView as UserManagementViewExtracted } from './views/UserManagementView';
import { getTemplateName, getTemplateSlug, getTemplateDescription, getTemplateTags } from './lib/template-utils';
import { ApiContext } from './hooks/use-api';
import { DeleteConfirmationModal as DeleteConfirmationModalExtracted } from './modals/DeleteConfirmationModal';
import { HibernateConfirmationModal as HibernateConfirmationModalExtracted } from './modals/HibernateConfirmationModal';
import { IdlePolicyModal as IdlePolicyModalExtracted } from './modals/IdlePolicyModal';

import {
  SideNavigation,
  Container,
  Header,
  SpaceBetween,
  Button,
  Cards,
  StatusIndicator,
  Badge,
  Table,
  Modal,
  Form,
  FormField,
  Input,
  Select,
  Alert,
  Spinner,
  Box,
  ColumnLayout,
  Tabs,
  Wizard,
  ProgressBar,
  Textarea,
  Toggle,
  Checkbox
} from './lib/cloudscape-shim';

// Type definitions
interface Project {
  id: string;
  name: string;
  description?: string;
  budget_limit?: number;
  current_spend?: number;
  owner_id?: string;
  owner_email?: string;
  owner?: string;
  created_at: string;
  updated_at?: string;
  status: string;
  member_count?: number;
  active_instances?: number;
  total_cost?: number;
  budget_status?: {
    total_budget: number;
    spent_amount: number;
    spent_percentage: number;
    alert_count: number;
  };
  last_activity?: string;
}

interface User {
  username: string;
  uid: number;
  display_name?: string;  // Frontend sends this
  full_name?: string;     // Backend returns this
  email: string;
  ssh_keys: number;
  created_at: string;
  provisioned_instances?: string[];
  status?: string;
  enabled?: boolean;       // User account status (enabled/disabled)
}

interface Template {
  Name: string;  // API returns capital N
  Slug: string;  // API returns capital S
  Description?: string;  // API returns capital D
  name?: string;  // Keep lowercase for backward compatibility
  slug?: string;  // Keep lowercase for backward compatibility
  description?: string;  // Keep lowercase for backward compatibility
  category?: string;
  complexity?: string;
  package_manager?: string;
  features?: string[];
  // Additional fields that might come from API
  [key: string]: unknown;
}

interface Instance {
  id: string;
  name: string;
  template: string;
  state: string;
  public_ip?: string;
  instance_type?: string;
  launch_time?: string;
  region?: string;
  username?: string;
  project?: string;
  web_services?: WebService[];
}

// Unified StorageVolume interface matching backend API
interface StorageVolume {
  name: string;
  type: 'workspace' | 'shared' | 'cloud';
  aws_service: 'ebs' | 'efs' | 's3';
  region: string;
  state: string;
  creation_time: string;

  // Size fields (varies by type)
  size_gb?: number;      // EBS
  size_bytes?: number;   // EFS

  // EBS-specific fields
  volume_id?: string;
  volume_type?: string;
  iops?: number;
  throughput?: number;
  attached_to?: string;

  // EFS-specific fields
  filesystem_id?: string;
  mount_targets?: string[];
  performance_mode?: string;
  throughput_mode?: string;

  // Cost
  estimated_cost_gb: number;
}

// Legacy interfaces for backward compatibility
interface EFSVolume {
  name: string;
  filesystem_id: string;
  region: string;
  creation_time: string;
  state: string;
  performance_mode: string;
  throughput_mode: string;
  estimated_cost_gb: number;
  size_bytes: number;
  attached_to?: string;
}

interface EBSVolume {
  name: string;
  volume_id: string;
  region: string;
  creation_time: string;
  state: string;
  volume_type: string;
  size_gb: number;
  estimated_cost_gb: number;
  attached_to?: string;
}

interface InstanceSnapshot {
  snapshot_id: string;
  snapshot_name: string;
  source_instance: string;
  source_instance_id: string;
  source_template: string;
  description: string;
  state: string;
  architecture: string;
  storage_cost_monthly: number;
  created_at: string;
  size_gb?: number;
}

interface BudgetData {
  project_id: string;
  project_name: string;
  total_budget: number;
  spent_amount: number;
  spent_percentage: number;
  remaining: number;
  alert_count: number;
  status: 'ok' | 'warning' | 'critical';
  projected_monthly_spend?: number;
  days_until_exhausted?: number;
  active_alerts?: Array<{
    threshold: number;
    action: string;
    triggered_at: string;
  }>;
}

interface CostBreakdown {
  ec2_compute?: number;
  ebs_storage?: number;
  efs_storage?: number;
  data_transfer: number;
  other?: number;
  total: number;
  instances?: number;
  storage?: number;
}

interface Budget {
  id: string;
  name: string;
  description: string;
  total_amount: number;
  allocated_amount: number;
  spent_amount: number;
  period: string;
  start_date: string;
  end_date?: string;
  alert_threshold: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

interface BudgetAllocation {
  id: string;
  budget_id: string;
  project_id: string;
  allocated_amount: number;
  spent_amount: number;
  alert_threshold?: number;
  notes?: string;
  allocated_at: string;
  allocated_by: string;
}

interface BudgetSummary {
  budget: Budget;
  allocations: BudgetAllocation[];
  project_names: Record<string, string>;
  remaining_amount: number;
  utilization_rate: number;
}

interface CreateBudgetRequest {
  name: string;
  description: string;
  total_amount: number;
  period: string;
  start_date: string;
  end_date?: string;
  alert_threshold: number;
  created_by: string;
}

interface AMI {
  id: string;
  name: string;
  template_name: string;
  region: string;
  state: string;
  architecture: string;
  size_gb: number;
  description?: string;
  created_at: string;
  tags?: Record<string, string>;
}

interface AMIBuild {
  id: string;
  template_name: string;
  status: string;
  progress: number;
  current_step?: string;
  error?: string;
  started_at: string;
  completed_at?: string;
}

interface AMIRegion {
  name: string;
  ami_count: number;
  total_size_gb: number;
  monthly_cost: number;
}

interface RightsizingRecommendation {
  instance_name: string;
  current_type: string;
  recommended_type: string;
  cpu_utilization: number;
  memory_utilization: number;
  current_cost: number;
  recommended_cost: number;
  monthly_savings: number;
  savings_percentage: number;
  confidence: 'high' | 'medium' | 'low';
  reason?: string;
}

interface RightsizingStats {
  total_recommendations: number;
  total_monthly_savings: number;
  average_cpu_utilization: number;
  average_memory_utilization: number;
  over_provisioned_count: number;
  optimized_count: number;
}

interface PolicyStatus {
  enabled: boolean;
  status: string;
  status_icon: string;
  assigned_policies: string[];
  message?: string;
}

interface PolicySet {
  id: string;
  name: string;
  description: string;
  policies: number;
  status: string;
  tags?: Record<string, string>;
}

interface PolicyCheckResult {
  allowed: boolean;
  template_name: string;
  reason: string;
  matched_policies?: string[];
  suggestions?: string[];
}

interface MarketplaceTemplate {
  id: string;
  name: string;
  display_name: string;
  author: string;
  publisher: string;
  category: string;
  description: string;
  rating: number;
  downloads: number;
  verified: boolean;
  featured: boolean;
  version: string;
  tags?: string[];
  badges?: string[];
  created_at: string;
  updated_at: string;
  ami_available?: boolean;
}

interface MarketplaceCategory {
  id: string;
  name: string;
  count: number;
}

interface IdlePolicy {
  id: string;
  name: string;
  idle_minutes: number;
  action: 'hibernate' | 'stop' | 'notify';
  cpu_threshold: number;
  memory_threshold: number;
  network_threshold: number;
  description?: string;
  enabled: boolean;
}

interface IdleSchedule {
  instance_name: string;
  policy_name: string;
  enabled: boolean;
  last_checked: string;
  idle_minutes: number;
  status: string;
}

interface CachedInvitation {
  token: string;
  invitation_id: string;
  project_id: string;
  project_name: string;
  email: string;
  role: string;
  invited_by: string;
  invited_at: string;
  expires_at: string;
  status: 'pending' | 'accepted' | 'declined' | 'expired' | 'revoked';
  message?: string;
  added_at: string;
}

interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  header?: string;
  content: string;
  dismissible?: boolean;
  onDismiss?: () => void;
}

interface ProjectData {
  name: string;
  description?: string;
  budget_limit?: number;
  [key: string]: unknown;
}

interface MemberData {
  user_id: string;
  username?: string;
  role: string;
  joined_at?: string;
  [key: string]: unknown;
}

interface UserData {
  username: string;
  display_name: string;
  email: string;
  [key: string]: unknown;
}

interface SharedTokenConfig {
  name: string;
  role: string;
  redemption_limit?: number;
  expires_at?: string;
  [key: string]: unknown;
}

interface BulkInviteResponse {
  total: number;
  sent: number;
  failed: number;
  skipped?: number;
  errors?: Array<{email: string; error: string}>;
}

// Additional type definitions for API integration
interface Invitation {
  id: string;
  project_id: string;
  project_name: string;
  email: string;
  role: 'viewer' | 'member' | 'admin';
  token: string;
  status: 'pending' | 'accepted' | 'declined' | 'expired';
  invited_by: string;
  invited_at: string;
  expires_at: string;
  responded_at?: string;
  message?: string;
}

interface ProjectDetails extends Project {
  members: ProjectMember[];
  cost_breakdown: CostBreakdown;
}

interface ProjectMember {
  user_id: string;
  username: string;
  role: string;
  joined_at: string;
}


interface SharedInvitationToken {
  token: string;
  project_id: string;
  project_name: string;
  name: string;
  role: 'viewer' | 'member' | 'admin';
  redemption_limit: number;
  redemptions: number;
  redemption_count?: number;
  created_at: string;
  created_by: string;
  expires_at: string;
  revoked: boolean;
  qr_code_url?: string;
  status?: string;
}

// v0.13.0 Governance Interfaces
interface RoleQuota {
  role: string;
  max_instances: number;
  max_instance_type: string;
  max_spend_daily: number;
}
interface GrantPeriod {
  name: string;
  start_date: string;
  end_date: string;
  auto_freeze: boolean;
  frozen_at?: string;
}
interface ApprovalRequest {
  id: string;
  project_id: string;
  requested_by: string;
  type: string;
  status: string;
  details: Record<string, unknown>;
  reason: string;
  reviewed_by?: string;
  review_note?: string;
  created_at: string;
  expires_at: string;
  reviewed_at?: string;
}
interface BudgetShareRequest {
  from_project_id: string;
  to_project_id?: string;
  to_member_id?: string;
  amount: number;
  reason?: string;
  expires_at?: string;
}
interface BudgetShareRecord {
  id: string;
  request: BudgetShareRequest;
  approved_by: string;
  created_at: string;
  expires_at?: string;
}
interface OnboardingTemplate {
  id?: string;
  name: string;
  description?: string;
  templates?: string[];
  budget_limit?: number;
}

// v0.14.0/v0.16.0 University Education System Interfaces
interface ClassMember {
  user_id: string;
  email: string;
  display_name: string;
  role: string;
  budget_spent: number;
  budget_limit: number;
  added_at: string;
  expires_at?: string;
}
interface Course {
  id: string;
  code: string;
  title: string;
  department: string;
  semester: string;
  semester_start: string;
  semester_end: string;
  owner: string;
  status: string; // "pending"|"active"|"closed"|"archived"
  members: ClassMember[];
  approved_templates: string[];
  per_student_budget: number;
  total_budget: number;
  default_template: string;
  auto_provision_on_enroll: boolean;
  created_at: string;
  updated_at: string;
}
interface CourseBudgetSummary {
  total_budget: number;
  per_student_default: number;
  total_spent: number;
  students: StudentBudgetInfo[];
}
interface StudentBudgetInfo {
  user_id: string;
  email: string;
  display_name: string;
  budget_limit: number;
  budget_spent: number;
  remaining: number;
}
interface CourseOverview {
  course_id: string;
  course_code: string;
  total_students: number;
  active_instances: number;
  total_budget_spent: number;
  students: StudentOverviewStatus[];
}
interface StudentOverviewStatus {
  user_id: string;
  email: string;
  display_name: string;
  instances: Instance[];
  budget_spent: number;
  budget_limit: number;
  budget_status: string; // "ok"|"warning"|"exceeded"
}
interface UsageReport {
  course_id: string;
  course_code: string;
  semester: string;
  total_spent: number;
  total_budget: number;
  students: StudentUsageRecord[];
  generated_at: string;
}
interface StudentUsageRecord {
  user_id: string;
  email: string;
  display_name: string;
  total_hours: number;
  total_cost: number;
  instance_count: number;
}
interface CourseAuditEntry {
  timestamp: string;
  course_id: string;
  actor: string;
  target: string;
  action: string;
  detail?: Record<string, unknown>;
}

// v0.19.0 Education Power Features interfaces
interface SharedMaterialsVolume {
  course_id: string;
  efs_id: string;
  mount_path: string;
  state: string;
  size_gb: number;
  created_at: string;
  mounted_instance_count: number;
}
interface WorkspaceResetResult {
  student_id: string;
  backup_snapshot_id?: string;
  backup_download_url?: string;
  status: string;
  backup_expires_at?: string;
}

// v0.18.0 Workshop & Event Management interfaces
interface WorkshopParticipant {
  user_id: string;
  email?: string;
  display_name?: string;
  joined_at: string;
  instance_id?: string;
  instance_name?: string;
  status: string; // "pending"|"provisioned"|"running"|"stopped"
  progress?: number; // 0-100
}

interface WorkshopEvent {
  id: string;
  title: string;
  description?: string;
  owner: string;
  template: string;
  approved_templates?: string[];
  max_participants: number;
  budget_per_participant?: number;
  start_time: string;
  end_time: string;
  early_access_hours?: number;
  status: string; // "draft"|"active"|"ended"|"archived"
  join_token?: string;
  participants?: WorkshopParticipant[];
  created_at: string;
  updated_at: string;
}

interface WorkshopDashboard {
  workshop_id: string;
  title: string;
  total_participants: number;
  active_instances: number;
  stopped_instances: number;
  pending_instances: number;
  total_spent: number;
  time_remaining: string;
  status: string;
  participants: WorkshopParticipant[];
}

interface WorkshopConfig {
  name: string;
  template: string;
  max_participants: number;
  budget_per_participant?: number;
  duration_hours: number;
  early_access_hours?: number;
  description?: string;
  created_at: string;
}

interface AppState {
  activeView: 'dashboard' | 'templates' | 'workspaces' | 'storage' | 'backups' | 'projects' | 'project-detail' | 'users' | 'ami' | 'rightsizing' | 'policy' | 'marketplace' | 'idle' | 'invitations' | 'logs' | 'settings' | 'terminal' | 'webview' | 'budgets' | 'approvals' | 'courses' | 'workshops' | 'capacity-blocks';
  settingsSection: 'general' | 'profiles' | 'users' | 'ami' | 'rightsizing' | 'policy' | 'marketplace' | 'idle' | 'logs';
  templates: Record<string, Template>;
  instances: Instance[];
  efsVolumes: EFSVolume[];
  ebsVolumes: EBSVolume[];
  snapshots: InstanceSnapshot[];
  projects: Project[];
  users: User[];
  budgets: BudgetData[];
  budgetPools: Budget[];
  selectedBudgetId: string | null;
  amis: AMI[];
  amiBuilds: AMIBuild[];
  amiRegions: AMIRegion[];
  rightsizingRecommendations: RightsizingRecommendation[];
  rightsizingStats: RightsizingStats | null;
  policyStatus: PolicyStatus | null;
  policySets: PolicySet[];
  marketplaceTemplates: MarketplaceTemplate[];
  marketplaceCategories: MarketplaceCategory[];
  idlePolicies: IdlePolicy[];
  idleSchedules: IdleSchedule[];
  invitations: CachedInvitation[];
  courses: Course[];
  workshops: WorkshopEvent[];
  selectedTemplate: Template | null;
  selectedProject: Project | null;
  selectedTerminalInstance: string;
  loading: boolean;
  notifications: Notification[]; // kept for migration; will be removed in Phase 8
  connected: boolean;
  error: string | null;
  updateInfo: any | null;
  autoStartEnabled: boolean;
  pendingApprovalsCount: number;
}

// API Response Interfaces
interface HibernationStatus {
  supported: boolean;
  enabled?: boolean;
  state?: string;
  message?: string;
}

interface UserStatus {
  username: string;
  status: string;
  provisioned_instances: string[];
  ssh_keys_count: number;
  last_active?: string;
}

interface UserProvisionResponse {
  success: boolean;
  instance: string;
  username: string;
  message?: string;
}

interface SSHKeyResponse {
  public_key: string;
  private_key: string;
  fingerprint: string;
  generated_at: string;
}

interface SSHKeyConfig {
  key_id: string;
  profile_id: string;
  username: string;
  key_type: string;
  fingerprint: string;
  public_key: string;
  comment: string;
  created_at: string;
  last_used?: string;
  from_profile: string;
  auto_generated: boolean;
}

interface UserSSHKeysResponse {
  username: string;
  keys: SSHKeyConfig[];
}

interface ProjectUsageResponse {
  project_id: string;
  period: string;
  instance_hours: number;
  storage_gb_hours: number;
  data_transfer_gb: number;
  [key: string]: string | number;
}

interface SharedToken {
  token: string;
  project_id: string;
  project_name: string;
  name: string;
  role: string;
  redemption_limit: number;
  redemptions_remaining: number;
  created_by: string;
  created_at: string;
  expires_at?: string;
  status: 'active' | 'expired' | 'exhausted';
}

interface QRCodeData {
  qr_code: string;  // base64 encoded image or URL
  url: string;
}

interface RedeemTokenResponse {
  success: boolean;
  project_id: string;
  project_name: string;
  role: string;
  message?: string;
}

interface WebService {
  name: string;
  port: number;
  url?: string;
  description?: string;
  type?: string;
}

interface UserUpdateRequest {
  email?: string;
  display_name?: string;
  role?: string;
}

// ── v0.20.0 Storage Power types ──────────────────────────────────────────────

interface FileEntry {
  path: string;
  size_bytes: number;
  is_dir: boolean;
  modified_at: string;
  permissions: string;
}

interface CapacityBlock {
  id: string;
  instance_type: string;
  instance_count: number;
  availability_zone: string;
  start_time: string;
  end_time: string;
  duration_hours: number;
  state: string;
  total_cost: number;
}

interface CapacityBlockRequest {
  instance_type: string;
  instance_count: number;
  availability_zone?: string;
  start_time: string;
  duration_hours: number;
}

interface S3Mount {
  instance_name: string;
  bucket_name: string;
  mount_path: string;
  method: string;
  read_only: boolean;
  status: string;
}

interface StorageAnalyticsSummary {
  storage_name: string;
  type: string;
  period: string;
  usage_percent: number;
  total_cost: number;
  daily_cost: number;
  recommendations: string[];
}

// Safe API Service with comprehensive error handling
class SafePrismAPI {
  private baseURL = 'http://localhost:8947';
  private apiKey = '';

  constructor() {
    // API key loading disabled - daemon runs in PRISM_TEST_MODE with auth bypass
  }

  private async safeRequest<T = unknown>(endpoint: string, method = 'GET', body?: unknown): Promise<T> {
    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
        },
        body: body ? JSON.stringify(body) : undefined,
      });

      if (!response.ok) {
        // Try to parse error message from response body
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const errorData = await response.text();
          if (errorData) {
            errorMessage = errorData;
          }
        } catch (e) {
          // If parsing fails, use the default message
        }
        const error: any = new Error(errorMessage);
        error.response = { status: response.status, statusText: response.statusText };
        throw error;
      }

      // Handle HTTP 204 No Content (empty response)
      if (response.status === 204 || response.headers.get('content-length') === '0') {
        return {} as T;
      }

      const data = await response.json();
      return data;
    } catch (error) {
      logger.error(`API request failed for ${endpoint}:`, error);
      throw error;
    }
  }

  async getTemplates(): Promise<Record<string, Template>> {
    try {
      const data = await this.safeRequest<Record<string, Template>>('/api/v1/templates');
      return data || {};
    } catch (error) {
      logger.error('Failed to fetch templates:', error);
      return {};
    }
  }

  async getInstances(): Promise<Instance[]> {
    try {
      const data = await this.safeRequest<{instances: Instance[]}>('/api/v1/instances');
      return Array.isArray(data?.instances) ? data.instances : [];
    } catch (error) {
      logger.error('Failed to fetch instances:', error);
      return [];
    }
  }

  async launchInstance(templateSlug: string, name: string, size: string = 'M', dryRun: boolean = false): Promise<Instance & { approval_pending?: boolean; approval_request_id?: string; message?: string }> {
    const body: Record<string, unknown> = {
      template: templateSlug,
      name,
      size,
    };
    if (dryRun) {
      body.dry_run = true;
    }
    return this.safeRequest('/api/v1/instances', 'POST', body);
  }

  // Comprehensive Instance Management APIs - Using Fixed Backend Endpoints
  async startInstance(identifier: string): Promise<void> {
    await this.safeRequest(`/api/v1/instances/${identifier}/start`, 'POST');
  }

  async stopInstance(identifier: string): Promise<void> {
    await this.safeRequest(`/api/v1/instances/${identifier}/stop`, 'POST');
  }

  async hibernateInstance(identifier: string): Promise<void> {
    await this.safeRequest(`/api/v1/instances/${identifier}/hibernate`, 'POST');
  }

  async resumeInstance(identifier: string): Promise<void> {
    await this.safeRequest(`/api/v1/instances/${identifier}/resume`, 'POST');
  }

  async getConnectionInfo(identifier: string): Promise<string> {
    const data = await this.safeRequest<{connection_info?: string}>(`/api/v1/instances/${identifier}/connect`);
    return data.connection_info || '';
  }

  async getHibernationStatus(identifier: string): Promise<HibernationStatus> {
    return this.safeRequest(`/api/v1/instances/${identifier}/hibernation-status`);
  }

  async deleteInstance(identifier: string): Promise<void> {
    await this.safeRequest(`/api/v1/instances/${identifier}`, 'DELETE');
  }

  // Comprehensive Storage Management APIs

  // Helper functions to convert unified StorageVolume to legacy formats
  private storageVolumeToEFS(vol: StorageVolume): EFSVolume | null {
    if (vol.type !== 'shared' && vol.aws_service !== 'efs') return null;
    return {
      name: vol.name,
      filesystem_id: vol.filesystem_id || '',
      region: vol.region,
      creation_time: vol.creation_time,
      state: vol.state,
      performance_mode: vol.performance_mode || '',
      throughput_mode: vol.throughput_mode || '',
      estimated_cost_gb: vol.estimated_cost_gb,
      size_bytes: vol.size_bytes || 0,
      attached_to: vol.attached_to,
    };
  }

  private storageVolumeToEBS(vol: StorageVolume): EBSVolume | null {
    if (vol.type !== 'workspace' && vol.aws_service !== 'ebs') return null;
    return {
      name: vol.name,
      volume_id: vol.volume_id || '',
      region: vol.region,
      creation_time: vol.creation_time,
      state: vol.state,
      volume_type: vol.volume_type || '',
      size_gb: vol.size_gb || 0,
      estimated_cost_gb: vol.estimated_cost_gb,
      attached_to: vol.attached_to,
    };
  }

  // EFS Volume Management (using unified API)
  async getEFSVolumes(): Promise<EFSVolume[]> {
    try {
      const data: StorageVolume[] = await this.safeRequest('/api/v1/volumes');
      if (!Array.isArray(data)) return [];
      // Convert unified StorageVolume to legacy EFSVolume format
      return data.map(vol => this.storageVolumeToEFS(vol)).filter((v): v is EFSVolume => v !== null);
    } catch (error) {
      logger.error('Failed to fetch shared storage volumes:', error);
      return [];
    }
  }

  async createEFSVolume(name: string, performanceMode: string = 'generalPurpose', throughputMode: string = 'bursting'): Promise<EFSVolume> {
    return this.safeRequest('/api/v1/volumes', 'POST', {
      name,
      performance_mode: performanceMode,
      throughput_mode: throughputMode,
    });
  }

  async deleteEFSVolume(name: string): Promise<void> {
    await this.safeRequest(`/api/v1/volumes/${name}`, 'DELETE');
  }

  async mountEFSVolume(volumeName: string, instance: string, mountPoint?: string): Promise<void> {
    const body: Record<string, string> = { instance };
    if (mountPoint) body.mount_point = mountPoint;
    await this.safeRequest(`/api/v1/volumes/${volumeName}/mount`, 'POST', body);
  }

  async unmountEFSVolume(volumeName: string, instance: string): Promise<void> {
    await this.safeRequest(`/api/v1/volumes/${volumeName}/unmount`, 'POST', { instance });
  }

  async syncEFSVolume(volumeName: string): Promise<EFSVolume> {
    return this.safeRequest(`/api/v1/volumes/${volumeName}/sync`, 'POST');
  }

  async syncAllEFSVolumes(): Promise<EFSVolume[]> {
    return this.safeRequest('/api/v1/volumes/sync', 'POST');
  }

  // EBS Storage Management (using unified API)
  async getEBSVolumes(): Promise<EBSVolume[]> {
    try {
      const data: StorageVolume[] = await this.safeRequest('/api/v1/storage');
      if (!Array.isArray(data)) return [];
      // Convert unified StorageVolume to legacy EBSVolume format
      // Note: /api/v1/storage now returns ALL storage (EBS + EFS), so filter for workspace only
      return data
        .filter(vol => vol.type === 'workspace' || vol.aws_service === 'ebs')
        .map(vol => this.storageVolumeToEBS(vol))
        .filter((v): v is EBSVolume => v !== null);
    } catch (error) {
      logger.error('Failed to fetch workspace storage volumes:', error);
      return [];
    }
  }

  async createEBSVolume(name: string, size: string = 'M', volumeType: string = 'gp3'): Promise<EBSVolume> {
    return this.safeRequest('/api/v1/storage', 'POST', {
      name,
      size,
      volume_type: volumeType,
    });
  }

  async deleteEBSVolume(name: string): Promise<void> {
    await this.safeRequest(`/api/v1/storage/${name}`, 'DELETE');
  }

  async attachEBSVolume(storageName: string, instance: string): Promise<void> {
    await this.safeRequest(`/api/v1/storage/${storageName}/attach`, 'POST', { instance });
  }

  async detachEBSVolume(storageName: string): Promise<void> {
    await this.safeRequest(`/api/v1/storage/${storageName}/detach`, 'POST');
  }

  async syncEBSVolume(storageName: string): Promise<EBSVolume> {
    return this.safeRequest(`/api/v1/storage/${storageName}/sync`, 'POST');
  }

  async syncAllEBSVolumes(): Promise<EBSVolume[]> {
    return this.safeRequest('/api/v1/storage/sync', 'POST');
  }

  // Snapshot/Backup Management APIs
  async getSnapshots(): Promise<InstanceSnapshot[]> {
    try {
      const data = await this.safeRequest<{snapshots: InstanceSnapshot[], count: number}>('/api/v1/snapshots');
      return Array.isArray(data?.snapshots) ? data.snapshots : [];
    } catch (error) {
      logger.error('Failed to fetch snapshots:', error);
      return [];
    }
  }

  async createSnapshot(instanceName: string, snapshotName: string, description?: string): Promise<InstanceSnapshot> {
    return this.safeRequest('/api/v1/snapshots', 'POST', {
      instance_name: instanceName,
      snapshot_name: snapshotName,
      description: description || '',
      no_reboot: true
    });
  }

  async deleteSnapshot(snapshotName: string): Promise<void> {
    await this.safeRequest(`/api/v1/snapshots/${snapshotName}`, 'DELETE');
  }

  async restoreSnapshot(snapshotName: string, instanceName: string): Promise<void> {
    await this.safeRequest(`/api/v1/snapshots/${snapshotName}/restore`, 'POST', {
      instance_name: instanceName
    });
  }

  // Comprehensive Project Management APIs

  // Project Operations
  async getProjects(): Promise<Project[]> {
    try {
      const data = await this.safeRequest<{projects?: Project[]}>('/api/v1/projects');
      return Array.isArray(data?.projects) ? data.projects : [];
    } catch (error) {
      logger.error('Failed to fetch projects:', error);
      return [];
    }
  }

  async createProject(projectData: ProjectData): Promise<Project> {
    return this.safeRequest<Project>('/api/v1/projects', 'POST', projectData);
  }

  async getProject(projectId: string): Promise<Project> {
    return this.safeRequest<Project>(`/api/v1/projects/${projectId}`);
  }

  async updateProject(projectId: string, projectData: Partial<ProjectData>): Promise<Project> {
    return this.safeRequest<Project>(`/api/v1/projects/${projectId}`, 'PUT', projectData);
  }

  async deleteProject(projectId: string): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}`, 'DELETE');
  }

  // Project Members
  async getProjectMembers(projectId: string): Promise<MemberData[]> {
    try {
      const data = await this.safeRequest(`/api/v1/projects/${projectId}/members`);
      return Array.isArray(data) ? data : [];
    } catch (error) {
      logger.error('Failed to fetch project members:', error);
      return [];
    }
  }

  async addProjectMember(projectId: string, memberData: MemberData): Promise<MemberData> {
    return this.safeRequest<MemberData>(`/api/v1/projects/${projectId}/members`, 'POST', memberData);
  }

  async updateProjectMember(projectId: string, userId: string, memberData: Partial<MemberData>): Promise<MemberData> {
    return this.safeRequest<MemberData>(`/api/v1/projects/${projectId}/members/${userId}`, 'PUT', memberData);
  }

  async removeProjectMember(projectId: string, userId: string): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/members/${userId}`, 'DELETE');
  }

  // Budget Management
  async getProjectBudget(projectId: string): Promise<BudgetData> {
    return this.safeRequest<BudgetData>(`/api/v1/projects/${projectId}/budget`);
  }

  // Cost Analysis
  async getProjectCosts(projectId: string, startDate?: string, endDate?: string): Promise<CostBreakdown> {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    const query = params.toString();
    return this.safeRequest<CostBreakdown>(`/api/v1/projects/${projectId}/costs${query ? '?' + query : ''}`);
  }

  // Resource Usage
  async getProjectUsage(projectId: string, period?: string): Promise<ProjectUsageResponse> {
    const query = period ? `?period=${period}` : '';
    return this.safeRequest<ProjectUsageResponse>(`/api/v1/projects/${projectId}/usage${query}`);
  }

  // User Operations
  async getUsers(): Promise<User[]> {
    try {
      const data = await this.safeRequest<User[] | {users?: User[]}>('/api/v1/users');
      // Handle both direct array response and wrapped object response
      if (Array.isArray(data)) return data;
      if (!Array.isArray(data) && Array.isArray((data as {users?: User[]}).users)) return (data as {users: User[]}).users;
      return [];
    } catch (error) {
      logger.error('Failed to fetch users:', error);
      return [];
    }
  }

  async createUser(userData: UserData): Promise<User> {
    console.log('[DEBUG] createUser called with:', userData);
    try {
      const result = await this.safeRequest<User>('/api/v1/users', 'POST', userData);
      console.log('[DEBUG] createUser success:', result);
      return result;
    } catch (error) {
      console.error('[DEBUG] createUser error:', error);
      throw error;
    }
  }

  async deleteUser(username: string): Promise<void> {
    await this.safeRequest(`/api/v1/users/${username}`, 'DELETE');
  }

  async getUserStatus(username: string): Promise<UserStatus> {
    return this.safeRequest(`/api/v1/users/${username}/status`);
  }

  async provisionUser(username: string, instanceName: string): Promise<UserProvisionResponse> {
    return this.safeRequest(`/api/v1/users/${username}/provision`, 'POST', { instance: instanceName });
  }

  async enableUser(username: string): Promise<void> {
    await this.safeRequest(`/api/v1/users/${username}/enable`, 'POST');
  }

  async disableUser(username: string): Promise<void> {
    await this.safeRequest(`/api/v1/users/${username}/disable`, 'POST');
  }

  async generateSSHKey(username: string): Promise<SSHKeyResponse> {
    return this.safeRequest(`/api/v1/users/${username}/ssh-key`, 'POST', {
      username: username,
      key_type: 'ed25519'
    });
  }

  async getUserSSHKeys(username: string): Promise<UserSSHKeysResponse> {
    return this.safeRequest(`/api/v1/users/${username}/ssh-key`);
  }

  async getUser(username: string): Promise<User> {
    return this.safeRequest(`/api/v1/users/${username}`);
  }

  async updateUser(username: string, updates: Partial<UserUpdateRequest>): Promise<User> {
    return this.safeRequest(`/api/v1/users/${username}`, 'PUT', updates);
  }

  // Helper function to calculate budget status
  private calculateBudgetStatus(spentPercent: number): 'ok' | 'warning' | 'critical' {
    if (spentPercent >= 95) return 'critical';
    if (spentPercent >= 80) return 'warning';
    return 'ok';
  }

  // Budget Management APIs
  async getBudgets(): Promise<BudgetData[]> {
    try {
      const projects = await this.getProjects();
      const budgets: BudgetData[] = [];

      // Fetch budget status for each project
      for (const project of projects) {
        try {
          const budgetStatus = await this.safeRequest<BudgetData>(`/api/v1/projects/${project.id}/budget`);

          if (budgetStatus && budgetStatus.total_budget > 0) {
            const remaining = budgetStatus.total_budget - budgetStatus.spent_amount;
            const spentPercent = budgetStatus.spent_percentage * 100;

            budgets.push({
              project_id: project.id,
              project_name: project.name,
              total_budget: budgetStatus.total_budget,
              spent_amount: budgetStatus.spent_amount,
              spent_percentage: budgetStatus.spent_percentage,
              remaining: remaining,
              alert_count: budgetStatus.alert_count || 0,
              status: this.calculateBudgetStatus(spentPercent),
              projected_monthly_spend: budgetStatus.projected_monthly_spend,
              days_until_exhausted: budgetStatus.days_until_exhausted,
              active_alerts: budgetStatus.active_alerts
            });
          }
        } catch (error) {
          logger.error(`Failed to fetch budget for project ${project.id}:`, error);
        }
      }

      return budgets;
    } catch (error) {
      logger.error('Failed to fetch budgets:', error);
      return [];
    }
  }

  async getCostBreakdown(projectId: string, startDate?: string, endDate?: string): Promise<CostBreakdown> {
    try {
      const params = new URLSearchParams();
      if (startDate) params.append('start_date', startDate);
      if (endDate) params.append('end_date', endDate);
      const query = params.toString();

      const data = await this.safeRequest<CostBreakdown>(`/api/v1/projects/${projectId}/costs${query ? '?' + query : ''}`);

      return {
        ec2_compute: data.ec2_compute || 0,
        ebs_storage: data.ebs_storage || 0,
        efs_storage: data.efs_storage || 0,
        data_transfer: data.data_transfer || 0,
        other: data.other || 0,
        total: data.total || 0
      };
    } catch (error) {
      logger.error(`Failed to fetch cost breakdown for project ${projectId}:`, error);
      return {
        ec2_compute: 0,
        ebs_storage: 0,
        efs_storage: 0,
        data_transfer: 0,
        other: 0,
        total: 0
      };
    }
  }

  async setBudget(projectId: string, totalBudget: number, alertThresholds?: number[]): Promise<void> {
    const alerts = alertThresholds?.map(threshold => ({
      threshold,
      action: 'notify',
      enabled: true
    })) || [];

    await this.safeRequest(`/api/v1/projects/${projectId}/budget`, 'PUT', {
      total_budget: totalBudget,
      alert_thresholds: alerts,
      budget_period: 'monthly'
    });
  }

  // Budget Pool Operations (v0.6.0+)
  async getBudgetPools(): Promise<Budget[]> {
    try {
      const data = await this.safeRequest<{budgets?: Budget[]}>('/api/v1/budgets');
      return Array.isArray(data?.budgets) ? data.budgets : [];
    } catch (error) {
      logger.error('Failed to fetch budget pools:', error);
      return [];
    }
  }

  async getBudgetPool(budgetId: string): Promise<Budget> {
    return this.safeRequest<Budget>(`/api/v1/budgets/${budgetId}`);
  }

  async getBudgetSummary(budgetId: string): Promise<BudgetSummary> {
    return this.safeRequest<BudgetSummary>(`/api/v1/budgets/${budgetId}/summary`);
  }

  async createBudgetPool(budgetData: CreateBudgetRequest): Promise<Budget> {
    return this.safeRequest<Budget>('/api/v1/budgets', 'POST', budgetData);
  }

  async updateBudgetPool(budgetId: string, updates: Partial<CreateBudgetRequest>): Promise<Budget> {
    return this.safeRequest<Budget>(`/api/v1/budgets/${budgetId}`, 'PUT', updates);
  }

  async deleteBudgetPool(budgetId: string): Promise<void> {
    await this.safeRequest(`/api/v1/budgets/${budgetId}`, 'DELETE');
  }

  async getBudgetAllocations(budgetId: string): Promise<BudgetAllocation[]> {
    try {
      const data = await this.safeRequest<{allocations?: BudgetAllocation[]}>(`/api/v1/budgets/${budgetId}/allocations`);
      return Array.isArray(data?.allocations) ? data.allocations : [];
    } catch (error) {
      logger.error('Failed to fetch budget allocations:', error);
      return [];
    }
  }

  // Invitation Management APIs (v0.5.11+)
  async getInvitationByToken(token: string): Promise<CachedInvitation> {
    try {
      const data = await this.safeRequest<{invitation: Invitation; project: {name: string}}>(`/api/v1/invitations/${token}`);
      const inv = data.invitation;

      // Map backend Invitation type to frontend CachedInvitation type
      return {
        token: inv.token,
        invitation_id: inv.id,  // Backend uses 'id', frontend uses 'invitation_id'
        project_id: inv.project_id,
        project_name: inv.project_name || data.project?.name || 'Unknown Project',
        email: inv.email,
        role: inv.role,
        invited_by: inv.invited_by,
        invited_at: inv.invited_at,
        expires_at: inv.expires_at,
        status: inv.status,
        message: inv.message,
        added_at: new Date().toISOString()  // Client-side timestamp for when added to cache
      };
    } catch (error) {
      logger.error('Failed to fetch invitation:', error);
      throw error;
    }
  }

  async acceptInvitation(token: string): Promise<void> {
    try {
      await this.safeRequest(`/api/v1/invitations/${token}/accept`, 'POST');
    } catch (error) {
      logger.error('Failed to accept invitation:', error);
      throw error;
    }
  }

  async declineInvitation(token: string, reason?: string): Promise<void> {
    try {
      const body = reason ? { reason } : undefined;
      await this.safeRequest(`/api/v1/invitations/${token}/decline`, 'POST', body);
    } catch (error) {
      logger.error('Failed to decline invitation:', error);
      throw error;
    }
  }

  async sendInvitation(projectId: string, email: string, role: string, message?: string, expiresAt?: string): Promise<Invitation> {
    try {
      const body: Record<string, string> = {
        email,
        role,
        invited_by: 'test-user', // Required by backend API
      };
      if (message) body.message = message;
      if (expiresAt) body.expires_at = expiresAt;

      // Backend returns nested response: {invitation, project, message}
      const data = await this.safeRequest<{invitation: Invitation, project: unknown, message: string}>(`/api/v1/projects/${projectId}/invitations`, 'POST', body);

      // Emit event to notify UI components that a new invitation was created
      window.dispatchEvent(new CustomEvent('invitation-created'));

      return data.invitation;
    } catch (error) {
      logger.error('Failed to send invitation:', error);
      throw error;
    }
  }

  // Bulk Invitation API (Issue #240)
  async bulkInvite(projectId: string, emails: string[], role: string, message?: string): Promise<BulkInviteResponse> {
    try {
      // Backend expects invitations array with BulkInvitationEntry objects
      const body: Record<string, unknown> = {
        invitations: emails.map(email => ({ email })),
        default_role: role,
      };
      if (message) body.default_message = message;

      // Backend returns BulkInvitationResponse with summary and results
      const data = await this.safeRequest<{summary: {total: number; sent: number; failed: number}; results: Array<{email: string; status: string; error?: string}>}>(`/api/v1/projects/${projectId}/invitations/bulk`, 'POST', body);

      // Transform backend response to match frontend BulkInviteResponse interface
      return {
        total: data.summary.total,
        sent: data.summary.sent,
        failed: data.summary.failed,
        errors: data.results.filter(r => r.status === 'failed').map(r => ({
          email: r.email,
          error: r.error || 'Unknown error'
        }))
      };
    } catch (error) {
      logger.error('Failed to send bulk invitations:', error);
      throw error;
    }
  }

  // Shared Token APIs (Issue #241)
  async getSharedTokens(projectId: string): Promise<SharedToken[]> {
    try {
      const data = await this.safeRequest<{tokens?: SharedToken[]}>(`/api/v1/projects/${projectId}/invitations/shared`);
      return Array.isArray(data?.tokens) ? data.tokens : [];
    } catch (error) {
      logger.error('Failed to fetch shared tokens:', error);
      return [];
    }
  }

  async createSharedToken(projectId: string, config: SharedTokenConfig): Promise<SharedInvitationToken> {
    try {
      const data = await this.safeRequest<SharedInvitationToken>(`/api/v1/projects/${projectId}/invitations/shared`, 'POST', config);
      // Emit event to notify UI that shared token was created
      window.dispatchEvent(new CustomEvent('shared-token-created'));
      return data;
    } catch (error) {
      logger.error('Failed to create shared token:', error);
      throw error;
    }
  }

  async extendSharedToken(token: string, expiresIn: string): Promise<void> {
    try {
      await this.safeRequest(`/api/v1/invitations/shared/${token}/extend`, 'PATCH', {
        add_days: parseInt(expiresIn)
      });
    } catch (error) {
      logger.error('Failed to extend shared token:', error);
      throw error;
    }
  }

  async revokeSharedToken(token: string): Promise<void> {
    try {
      await this.safeRequest(`/api/v1/invitations/shared/${token}`, 'DELETE');
    } catch (error) {
      logger.error('Failed to revoke shared token:', error);
      throw error;
    }
  }

  async redeemSharedToken(token: string): Promise<RedeemTokenResponse> {
    try {
      const data = await this.safeRequest<RedeemTokenResponse>(`/api/v1/invitations/shared/redeem`, 'POST', { token });
      return data;
    } catch (error) {
      logger.error('Failed to redeem shared token:', error);
      throw error;
    }
  }

  async getSharedTokenQRCode(token: string, format: 'json' | 'png' = 'json'): Promise<QRCodeData> {
    try {
      const data = await this.safeRequest<QRCodeData>(`/api/v1/invitations/shared/${token}/qr?format=${format}`);
      return data;
    } catch (error) {
      logger.error('Failed to get QR code:', error);
      throw error;
    }
  }

  // Additional Invitation Management APIs
  async getMyInvitations(email?: string): Promise<Invitation[]> {
    try {
      // Use provided email or default for testing
      const userEmail = email || 'test-user@example.com';
      const response = await this.safeRequest(`/api/v1/invitations/my?email=${encodeURIComponent(userEmail)}`);
      // Backend returns {invitations: [...], email: "...", filter: {...}}
      const data = (response as any)?.invitations || response;
      return Array.isArray(data) ? data : [];
    } catch (error) {
      logger.error('Failed to fetch my invitations:', error);
      return [];
    }
  }

  async getProjectInvitations(projectId: string): Promise<Invitation[]> {
    try {
      const data = await this.safeRequest(`/api/v1/projects/${projectId}/invitations`);
      return Array.isArray(data) ? data : [];
    } catch (error) {
      logger.error('Failed to fetch project invitations:', error);
      return [];
    }
  }

  async revokeInvitation(invitationId: string): Promise<void> {
    try {
      await this.safeRequest(`/api/v1/invitations/${invitationId}`, 'DELETE');
    } catch (error) {
      logger.error('Failed to revoke invitation:', error);
      throw error;
    }
  }

  // Project Details API
  async getProjectDetails(projectId: string): Promise<ProjectDetails> {
    const data = await this.safeRequest<any>(`/api/v1/projects/${projectId}`);
    // Map backend types.Project format (budget.total_budget) to ProjectDetails format (budget_limit)
    return {
      ...data,
      budget_limit: data.budget?.total_budget,
      current_spend: data.budget?.spent_amount || 0,
      members: data.members || [],
      cost_breakdown: data.cost_breakdown || { instances: 0, storage: 0, data_transfer: 0, total: 0 }
    };
  }

  // User Provisioning API
  async provisionUserOnWorkspace(username: string, instanceId: string): Promise<void> {
    await this.safeRequest(`/api/v1/users/${username}/provision`, 'POST', { instance_id: instanceId });
  }

  // AMI Management APIs
  async getAMIs(): Promise<AMI[]> {
    try {
      const data = await this.safeRequest('/api/v1/ami/list');

      // Transform backend response to frontend format
      if (!data || !Array.isArray(data)) {
        return [];
      }

      return data.map((ami: Record<string, unknown>) => ({
        id: (ami.id as string) || (ami.ami_id as string) || '',
        name: (ami.name as string) || (ami.id as string) || '',
        template_name: (ami.template_name as string) || (ami.template as string) || 'unknown',
        region: (ami.region as string) || 'us-west-2',
        state: (ami.state as string) || 'available',
        architecture: (ami.architecture as string) || 'x86_64',
        size_gb: (ami.size_gb as number) || (ami.size as number) || 0,
        description: (ami.description as string) || '',
        created_at: (ami.created_at as string) || (ami.creation_date as string) || new Date().toISOString(),
        tags: (ami.tags as Record<string, string>) || {}
      }));
    } catch (error) {
      logger.error('Failed to fetch AMIs:', error);
      return [];
    }
  }

  async getAMIBuilds(): Promise<AMIBuild[]> {
    try {
      // Note: Backend may not have build tracking yet
      // Return empty array for now
      return [];
    } catch (error) {
      logger.error('Failed to fetch AMI builds:', error);
      return [];
    }
  }

  async getAMIRegions(): Promise<AMIRegion[]> {
    try {
      const amis = await this.getAMIs();

      // Group AMIs by region
      const regionMap = new Map<string, { count: number; totalSize: number }>();

      amis.forEach(ami => {
        const existing = regionMap.get(ami.region) || { count: 0, totalSize: 0 };
        regionMap.set(ami.region, {
          count: existing.count + 1,
          totalSize: existing.totalSize + ami.size_gb
        });
      });

      // Convert to array and calculate costs (estimated at $0.05 per GB-month for EBS snapshots)
      return Array.from(regionMap.entries()).map(([name, data]) => ({
        name,
        ami_count: data.count,
        total_size_gb: data.totalSize,
        monthly_cost: data.totalSize * 0.05
      })).sort((a, b) => b.ami_count - a.ami_count);
    } catch (error) {
      logger.error('Failed to calculate AMI regions:', error);
      return [];
    }
  }

  async deleteAMI(amiId: string): Promise<void> {
    await this.safeRequest('/api/v1/ami/delete', 'POST', {
      ami_id: amiId,
      deregister_only: false
    });
  }

  async buildAMI(templateName: string): Promise<{ build_id: string }> {
    const response = await this.safeRequest<{ build_id: string }>('/api/v1/ami/create', 'POST', {
      template_name: templateName
    });
    return response;
  }

  // Rightsizing APIs
  async getRightsizingRecommendations(): Promise<RightsizingRecommendation[]> {
    try {
      const data = await this.safeRequest<{recommendations?: Record<string, unknown>[]}>('/api/v1/rightsizing/recommendations');
      if (!data || !Array.isArray(data.recommendations)) {
        return [];
      }
      return data.recommendations.map((rec: Record<string, unknown>) => ({
        instance_name: (rec.instance_name as string) || (rec.InstanceName as string) || '',
        current_type: (rec.current_type as string) || (rec.CurrentType as string) || '',
        recommended_type: (rec.recommended_type as string) || (rec.RecommendedType as string) || '',
        cpu_utilization: (rec.cpu_utilization as number) || (rec.CPUUtilization as number) || 0,
        memory_utilization: (rec.memory_utilization as number) || (rec.MemoryUtilization as number) || 0,
        current_cost: (rec.current_cost as number) || (rec.CurrentCost as number) || 0,
        recommended_cost: (rec.recommended_cost as number) || (rec.RecommendedCost as number) || 0,
        monthly_savings: (rec.monthly_savings as number) || (rec.MonthlySavings as number) || 0,
        savings_percentage: (rec.savings_percentage as number) || (rec.SavingsPercentage as number) || 0,
        confidence: ((rec.confidence as string) || (rec.Confidence as string) || 'medium') as 'high' | 'medium' | 'low',
        reason: (rec.reason || rec.Reason) as string | undefined
      }));
    } catch (error) {
      logger.error('Failed to fetch rightsizing recommendations:', error);
      return [];
    }
  }

  async getRightsizingStats(): Promise<RightsizingStats | null> {
    try {
      const data = await this.safeRequest<Partial<RightsizingStats>>('/api/v1/rightsizing/stats');
      return {
        total_recommendations: data.total_recommendations || 0,
        total_monthly_savings: data.total_monthly_savings || 0,
        average_cpu_utilization: data.average_cpu_utilization || 0,
        average_memory_utilization: data.average_memory_utilization || 0,
        over_provisioned_count: data.over_provisioned_count || 0,
        optimized_count: data.optimized_count || 0
      };
    } catch (error: unknown) {
      // Silently handle 400/404 - endpoint may not be implemented yet
      const errorMessage = error instanceof Error ? error.message : String(error);
      if (errorMessage.includes('HTTP 400') || errorMessage.includes('HTTP 404')) {
        return null; // Don't log, just return null
      }
      // Only log unexpected errors
      logger.error('Unexpected error fetching rightsizing stats:', error);
      return null;
    }
  }

  async applyRightsizingRecommendation(instanceName: string): Promise<void> {
    await this.safeRequest(`/api/v1/rightsizing/instance/${instanceName}/apply`, 'POST');
  }

  // Policy APIs
  async getPolicyStatus(): Promise<PolicyStatus | null> {
    try {
      const data = await this.safeRequest<Partial<PolicyStatus>>('/api/v1/policies/status');
      return {
        enabled: data.enabled || false,
        status: data.status || 'unknown',
        status_icon: data.status_icon || '',
        assigned_policies: data.assigned_policies || [],
        message: data.message
      };
    } catch (error) {
      logger.error('Failed to fetch policy status:', error);
      return null;
    }
  }

  async getPolicySets(): Promise<PolicySet[]> {
    try {
      const data = await this.safeRequest<{policy_sets?: Record<string, Record<string, unknown>>}>('/api/v1/policies/sets');
      if (!data || !data.policy_sets) {
        return [];
      }
      return Object.entries(data.policy_sets).map(([id, info]: [string, Record<string, unknown>]) => ({
        id,
        name: (info.name as string) || id,
        description: (info.description as string) || '',
        policies: (info.policies as number) || 0,
        status: (info.status as string) || 'active',
        tags: info.tags as Record<string, string> | undefined
      }));
    } catch (error) {
      logger.error('Failed to fetch policy sets:', error);
      return [];
    }
  }

  async setPolicyEnforcement(enabled: boolean): Promise<void> {
    await this.safeRequest('/api/v1/policies/enforcement', 'POST', { enabled });
  }

  async assignPolicySet(policySetId: string): Promise<void> {
    await this.safeRequest('/api/v1/policies/assign', 'POST', { policy_set: policySetId });
  }

  async checkTemplateAccess(templateName: string): Promise<PolicyCheckResult> {
    const data = await this.safeRequest<Partial<PolicyCheckResult>>('/api/v1/policies/check', 'POST', { template_name: templateName });
    return {
      allowed: data.allowed || false,
      template_name: data.template_name || templateName,
      reason: data.reason || '',
      matched_policies: data.matched_policies,
      suggestions: data.suggestions
    };
  }

  // Marketplace APIs
  async getMarketplaceTemplates(query?: string, category?: string): Promise<MarketplaceTemplate[]> {
    try {
      let url = '/api/v1/marketplace/templates?';
      if (query) url += `query=${encodeURIComponent(query)}&`;
      if (category) url += `category=${encodeURIComponent(category)}&`;

      const data = await this.safeRequest<{templates?: Record<string, unknown>[]}>( url);
      if (!data || !Array.isArray(data.templates)) {
        return [];
      }
      return data.templates.map((t: Record<string, unknown>) => ({
        id: (t.id as string) || (t.ID as string) || '',
        name: (t.name as string) || (t.Name as string) || '',
        display_name: (t.display_name as string) || (t.DisplayName as string) || (t.name as string) || '',
        author: (t.author as string) || (t.Author as string) || '',
        publisher: (t.publisher as string) || (t.Publisher as string) || '',
        category: (t.category as string) || (t.Category as string) || '',
        description: (t.description as string) || (t.Description as string) || '',
        rating: (t.rating as number) || (t.Rating as number) || 0,
        downloads: (t.downloads as number) || (t.Downloads as number) || 0,
        verified: (t.verified as boolean) || (t.Verified as boolean) || false,
        featured: (t.featured as boolean) || (t.Featured as boolean) || false,
        version: (t.version as string) || (t.Version as string) || '',
        tags: (t.tags as string[]) || (t.Tags as string[]),
        badges: (t.badges as string[]) || (t.Badges as string[]),
        created_at: (t.created_at as string) || (t.CreatedAt as string) || '',
        updated_at: (t.updated_at as string) || (t.UpdatedAt as string) || '',
        ami_available: (t.ami_available as boolean) || (t.AMIAvailable as boolean) || false
      }));
    } catch (error) {
      logger.error('Failed to fetch marketplace templates:', error);
      return [];
    }
  }

  async getMarketplaceCategories(): Promise<MarketplaceCategory[]> {
    try {
      const data = await this.safeRequest<{categories?: Record<string, unknown>[]}>('/api/v1/marketplace/categories');
      if (!data || !Array.isArray(data.categories)) {
        return [];
      }
      return data.categories.map((c: Record<string, unknown>) => ({
        id: (c.id as string) || (c.ID as string) || '',
        name: (c.name as string) || (c.Name as string) || '',
        count: (c.count as number) || (c.Count as number) || 0
      }));
    } catch (error) {
      logger.error('Failed to fetch marketplace categories:', error);
      return [];
    }
  }

  async installMarketplaceTemplate(templateId: string, localName: string): Promise<void> {
    await this.safeRequest('/api/v1/templates/install-marketplace', 'POST', { marketplace_template_id: templateId, local_name: localName });
  }

  // Idle Detection APIs
  async getIdlePolicies(): Promise<IdlePolicy[]> {
    try {
      const data = await this.safeRequest<{policies?: Record<string, Record<string, unknown>>}>('/api/v1/idle/policies');
      if (!data || !data.policies) {
        return [];
      }
      const policies = Object.entries(data.policies).map(([id, p]: [string, Record<string, unknown>]) => ({
        id,
        name: (p.name as string) || (p.Name as string) || id,
        idle_minutes: (p.idle_minutes as number) || (p.IdleMinutes as number) || 0,
        action: ((p.action as string) || (p.Action as string) || 'notify') as 'hibernate' | 'stop' | 'notify',
        cpu_threshold: (p.cpu_threshold as number) || (p.CPUThreshold as number) || 10,
        memory_threshold: (p.memory_threshold as number) || (p.MemoryThreshold as number) || 10,
        network_threshold: (p.network_threshold as number) || (p.NetworkThreshold as number) || 1,
        description: (p.description as string) || (p.Description as string),
        enabled: p.enabled !== undefined ? (p.enabled as boolean) : (p.Enabled !== undefined ? (p.Enabled as boolean) : true)
      }));
      return policies;
    } catch (error) {
      logger.error('Failed to fetch idle policies:', error);
      return [];
    }
  }

  async getIdleSchedules(): Promise<IdleSchedule[]> {
    try {
      const data = await this.safeRequest<{schedules?: Record<string, unknown>[]}>('/api/v1/idle/schedules');
      if (!data || !Array.isArray(data.schedules)) {
        return [];
      }
      return data.schedules.map((s: Record<string, unknown>) => ({
        instance_name: (s.instance_name as string) || (s.InstanceName as string) || '',
        policy_name: (s.policy_name as string) || (s.PolicyName as string) || '',
        enabled: s.enabled !== undefined ? (s.enabled as boolean) : (s.Enabled !== undefined ? (s.Enabled as boolean) : true),
        last_checked: (s.last_checked as string) || (s.LastChecked as string) || '',
        idle_minutes: (s.idle_minutes as number) || (s.IdleMinutes as number) || 0,
        status: (s.status as string) || (s.Status as string) || ''
      }));
    } catch (error) {
      logger.error('Failed to fetch idle schedules:', error);
      return [];
    }
  }

  // Per-instance Idle Policy APIs (Issue #288)
  async getInstanceIdlePolicies(instanceName: string): Promise<IdlePolicy[]> {
    try {
      const data = await this.safeRequest(`/api/v1/instances/${instanceName}/idle/policies`);
      if (!data || !Array.isArray(data)) return [];
      return (data as Record<string, unknown>[]).map((p) => ({
        id: (p.id as string) || '',
        name: (p.name as string) || (p.Name as string) || '',
        idle_minutes: (p.idle_minutes as number) || 0,
        action: ((p.action as string) || 'notify') as 'hibernate' | 'stop' | 'notify',
        cpu_threshold: (p.cpu_threshold as number) || 10,
        memory_threshold: (p.memory_threshold as number) || 10,
        network_threshold: (p.network_threshold as number) || 1,
        description: (p.description as string) || '',
        enabled: p.enabled !== undefined ? (p.enabled as boolean) : true
      }));
    } catch (error) {
      logger.error(`Failed to fetch idle policies for ${instanceName}:`, error);
      return [];
    }
  }

  async applyIdlePolicy(instanceName: string, policyId: string): Promise<void> {
    await this.safeRequest(`/api/v1/instances/${instanceName}/idle/policies/${policyId}`, 'PUT');
  }

  async removeIdlePolicy(instanceName: string, policyId: string): Promise<void> {
    await this.safeRequest(`/api/v1/instances/${instanceName}/idle/policies/${policyId}`, 'DELETE');
  }

  // Profile Management APIs
  async getProfiles(): Promise<any[]> {
    try {
      const data = await this.safeRequest('/api/v1/profiles');
      return Array.isArray(data) ? data : [];
    } catch (error) {
      logger.error('Failed to fetch profiles:', error);
      return [];
    }
  }

  async createProfile(profile: { name: string; aws_profile: string; region?: string }): Promise<any> {
    return this.safeRequest('/api/v1/profiles', 'POST', profile);
  }

  async updateProfile(profileId: string, updates: { name?: string; aws_profile?: string; region?: string }): Promise<any> {
    return this.safeRequest(`/api/v1/profiles/${profileId}`, 'PUT', updates);
  }

  async deleteProfile(profileId: string): Promise<void> {
    await this.safeRequest(`/api/v1/profiles/${profileId}`, 'DELETE');
  }

  async switchProfile(profileId: string): Promise<any> {
    return this.safeRequest(`/api/v1/profiles/${profileId}/activate`, 'POST');
  }

  async checkForUpdates(): Promise<any> {
    return this.safeRequest('/api/v1/update/check');
  }

  async getAutoStartStatus(): Promise<{ enabled: boolean; method?: string; path?: string }> {
    // Call the Wails backend method (only available in desktop app, not in E2E tests)
    try {
      // Check if running in Wails environment
      if (!(window as any).go?.main?.PrismService) {
        // Not in Wails environment (e.g., E2E tests), return default
        return { enabled: false };
      }
      const response = await (window as any).go.main.PrismService.GetAutoStartStatus();
      return response;
    } catch (error) {
      logger.error('Failed to get auto-start status:', error);
      // Return default instead of throwing to avoid breaking E2E tests
      return { enabled: false };
    }
  }

  async setAutoStart(enable: boolean): Promise<void> {
    // Call the Wails backend method (only available in desktop app, not in E2E tests)
    try {
      // Check if running in Wails environment
      if (!(window as any).go?.main?.PrismService) {
        // Not in Wails environment (e.g., E2E tests), no-op
        logger.debug('setAutoStart called in non-Wails environment, ignoring');
        return;
      }
      await (window as any).go.main.PrismService.SetAutoStart(enable);
    } catch (error) {
      logger.error('Failed to set auto-start:', error);
      // Don't throw to avoid breaking E2E tests
    }
  }

  // v0.13.0 Governance APIs
  async getProjectQuotas(projectId: string): Promise<RoleQuota[]> {
    const data = await this.safeRequest<any>(`/api/v1/projects/${projectId}/quotas`);
    return data?.role_quotas || [];
  }

  async setProjectQuota(projectId: string, quota: RoleQuota): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/quotas`, 'PUT', quota);
  }

  async deleteProjectQuota(projectId: string, role: string): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/quotas/${role}`, 'DELETE');
  }

  async getGrantPeriod(projectId: string): Promise<GrantPeriod | null> {
    const data = await this.safeRequest<any>(`/api/v1/projects/${projectId}/grant-period`);
    return data?.grant_period || null;
  }

  async setGrantPeriod(projectId: string, gp: GrantPeriod): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/grant-period`, 'PUT', gp);
  }

  async deleteGrantPeriod(projectId: string): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/grant-period`, 'DELETE');
  }

  async listApprovals(projectId: string, status?: string): Promise<ApprovalRequest[]> {
    const qs = status ? `?status=${status}` : '';
    const data = await this.safeRequest<any>(`/api/v1/projects/${projectId}/approvals${qs}`);
    return data?.approvals || [];
  }

  async approveRequest(projectId: string, approvalId: string, note: string): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/approvals/${approvalId}/approve`, 'POST', { note });
  }

  async denyRequest(projectId: string, approvalId: string, note: string): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/approvals/${approvalId}/deny`, 'POST', { note });
  }

  async listAllApprovals(status?: string): Promise<ApprovalRequest[]> {
    const qs = status ? `?status=${status}` : '';
    const data = await this.safeRequest<any>(`/api/v1/admin/approvals${qs}`);
    return data?.approvals || [];
  }

  // v0.21.0: Get a single approval request by ID (#495)
  async getApproval(projectId: string, approvalId: string): Promise<ApprovalRequest> {
    return this.safeRequest<ApprovalRequest>(`/api/v1/projects/${projectId}/approvals/${approvalId}`);
  }

  // v0.21.0: Submit approval request for a launch (#495)
  async submitApprovalForLaunch(projectId: string, templateName: string, size: string, reason: string): Promise<ApprovalRequest> {
    const data = await this.safeRequest<any>(`/api/v1/projects/${projectId}/approvals`, 'POST', {
      type: 'expensive_instance',
      details: { template: templateName, size },
      reason,
    });
    return data;
  }

  async shareProjectBudget(projectId: string, req: BudgetShareRequest): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/budget/share`, 'POST', req);
  }

  async listProjectBudgetShares(projectId: string): Promise<BudgetShareRecord[]> {
    const data = await this.safeRequest<any>(`/api/v1/projects/${projectId}/budget/shares`);
    return data?.shares || [];
  }

  async listOnboardingTemplates(projectId: string): Promise<OnboardingTemplate[]> {
    const data = await this.safeRequest<any>(`/api/v1/projects/${projectId}/onboarding-templates`);
    return data?.onboarding_templates || [];
  }

  async addOnboardingTemplate(projectId: string, tmpl: OnboardingTemplate): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/onboarding-templates`, 'POST', tmpl);
  }

  async deleteOnboardingTemplate(projectId: string, nameOrId: string): Promise<void> {
    await this.safeRequest(`/api/v1/projects/${projectId}/onboarding-templates/${encodeURIComponent(nameOrId)}`, 'DELETE');
  }

  async getMonthlyReport(projectId: string, month: string, format: string): Promise<string> {
    const qs = `?month=${month}&format=${format}`;
    const data = await this.safeRequest<any>(`/api/v1/projects/${projectId}/reports/monthly${qs}`);
    return typeof data === 'string' ? data : JSON.stringify(data, null, 2);
  }

  // ── v0.14.0/v0.16.0 University Education System ──────────────────────────

  async getCourses(): Promise<Course[]> {
    const data = await this.safeRequest<any>('/api/v1/courses');
    return (data?.courses || []) as Course[];
  }

  async createCourse(courseData: Partial<Course>): Promise<Course> {
    const data = await this.safeRequest<Course>('/api/v1/courses', 'POST', courseData);
    return data as Course;
  }

  async getCourse(id: string): Promise<Course> {
    const data = await this.safeRequest<Course>(`/api/v1/courses/${encodeURIComponent(id)}`);
    return data as Course;
  }

  async closeCourse(id: string): Promise<void> {
    await this.safeRequest(`/api/v1/courses/${encodeURIComponent(id)}/close`, 'POST');
  }

  async deleteCourse(id: string): Promise<void> {
    await this.safeRequest(`/api/v1/courses/${encodeURIComponent(id)}`, 'DELETE');
  }

  async archiveCourse(id: string): Promise<{ instances_stopped: string[] }> {
    const data = await this.safeRequest<any>(`/api/v1/courses/${encodeURIComponent(id)}/archive`, 'POST');
    return data || { instances_stopped: [] };
  }

  async getCourseMembers(id: string, role?: string): Promise<ClassMember[]> {
    const qs = role ? `?role=${encodeURIComponent(role)}` : '';
    const data = await this.safeRequest<any>(`/api/v1/courses/${encodeURIComponent(id)}/members${qs}`);
    return (data?.members || []) as ClassMember[];
  }

  async enrollCourseMember(id: string, memberData: Partial<ClassMember>): Promise<ClassMember> {
    const data = await this.safeRequest<ClassMember>(`/api/v1/courses/${encodeURIComponent(id)}/members`, 'POST', memberData);
    return data as ClassMember;
  }

  async unenrollCourseMember(id: string, userId: string): Promise<void> {
    await this.safeRequest(`/api/v1/courses/${encodeURIComponent(id)}/members/${encodeURIComponent(userId)}`, 'DELETE');
  }

  async getCourseTemplates(id: string): Promise<{ templates: string[] }> {
    const data = await this.safeRequest<any>(`/api/v1/courses/${encodeURIComponent(id)}/templates`);
    return { templates: data?.approved_templates || [] };
  }

  async addCourseTemplate(id: string, slug: string): Promise<void> {
    await this.safeRequest(`/api/v1/courses/${encodeURIComponent(id)}/templates`, 'POST', { template: slug });
  }

  async removeCourseTemplate(id: string, slug: string): Promise<void> {
    await this.safeRequest(`/api/v1/courses/${encodeURIComponent(id)}/templates/${encodeURIComponent(slug)}`, 'DELETE');
  }

  async getCourseBudget(id: string): Promise<CourseBudgetSummary> {
    const data = await this.safeRequest<CourseBudgetSummary>(`/api/v1/courses/${encodeURIComponent(id)}/budget`);
    return data as CourseBudgetSummary;
  }

  async distributeCourseBudget(id: string, amount: number): Promise<void> {
    await this.safeRequest(`/api/v1/courses/${encodeURIComponent(id)}/budget/distribute`, 'POST', { amount_per_student: amount });
  }

  async debugStudent(courseId: string, studentId: string): Promise<Record<string, unknown>> {
    const data = await this.safeRequest<any>(`/api/v1/courses/${encodeURIComponent(courseId)}/members/${encodeURIComponent(studentId)}/debug`);
    return data || {};
  }

  async resetStudent(courseId: string, studentId: string, reason: string): Promise<void> {
    await this.safeRequest(`/api/v1/courses/${encodeURIComponent(courseId)}/members/${encodeURIComponent(studentId)}/reset`, 'POST', { reason });
  }

  async provisionStudent(courseId: string, studentId: string, data?: Record<string, unknown>): Promise<Record<string, unknown>> {
    const result = await this.safeRequest<any>(`/api/v1/courses/${encodeURIComponent(courseId)}/members/${encodeURIComponent(studentId)}/provision`, 'POST', data || {});
    return result || {};
  }

  async getCourseOverview(id: string): Promise<CourseOverview> {
    const data = await this.safeRequest<CourseOverview>(`/api/v1/courses/${encodeURIComponent(id)}/overview`);
    return data as CourseOverview;
  }

  async getCourseReport(id: string, format?: string): Promise<UsageReport> {
    const qs = format ? `?format=${encodeURIComponent(format)}` : '';
    const data = await this.safeRequest<UsageReport>(`/api/v1/courses/${encodeURIComponent(id)}/report${qs}`);
    return data as UsageReport;
  }

  async getCourseAuditLog(id: string, params?: { student_id?: string; since?: string; limit?: number }): Promise<{ entries: CourseAuditEntry[] }> {
    const qs = new URLSearchParams();
    if (params?.student_id) qs.set('student_id', params.student_id);
    if (params?.since) qs.set('since', params.since);
    if (params?.limit) qs.set('limit', String(params.limit));
    const qstr = qs.toString() ? `?${qs.toString()}` : '';
    const data = await this.safeRequest<any>(`/api/v1/courses/${encodeURIComponent(id)}/audit${qstr}`);
    return { entries: (data?.entries || []) as CourseAuditEntry[] };
  }

  async importCourseRoster(id: string, file: File, format?: string): Promise<{ enrolled: number; errors: string[] }> {
    const qs = format ? `?format=${encodeURIComponent(format)}` : '';
    const url = `/api/v1/courses/${encodeURIComponent(id)}/members/import${qs}`;
    const baseURL = (window as any).__daemonURL || 'http://localhost:8947';
    const resp = await fetch(`${baseURL}${url}`, {
      method: 'POST',
      headers: { 'Content-Type': 'text/csv' },
      body: file,
    });
    if (!resp.ok) throw new Error(`Import failed: ${resp.statusText}`);
    const data = await resp.json();
    return { enrolled: data?.enrolled || 0, errors: data?.errors || [] };
  }

  // ── v0.18.0 Workshop & Event Management ────────────────────────────────────

  async getWorkshops(params?: { owner?: string; status?: string }): Promise<WorkshopEvent[]> {
    const qs = params ? '?' + new URLSearchParams(Object.entries(params).filter(([, v]) => v)).toString() : '';
    const data = await this.safeRequest<any>(`/api/v1/workshops${qs}`);
    return (data?.workshops || []) as WorkshopEvent[];
  }

  async createWorkshop(workshopData: Partial<WorkshopEvent>): Promise<WorkshopEvent> {
    const data = await this.safeRequest<WorkshopEvent>('/api/v1/workshops', 'POST', workshopData);
    return data as WorkshopEvent;
  }

  async getWorkshop(id: string): Promise<WorkshopEvent> {
    const data = await this.safeRequest<WorkshopEvent>(`/api/v1/workshops/${encodeURIComponent(id)}`);
    return data as WorkshopEvent;
  }

  async updateWorkshop(id: string, updates: Partial<WorkshopEvent>): Promise<WorkshopEvent> {
    const data = await this.safeRequest<WorkshopEvent>(`/api/v1/workshops/${encodeURIComponent(id)}`, 'PUT', updates);
    return data as WorkshopEvent;
  }

  async deleteWorkshop(id: string): Promise<void> {
    await this.safeRequest(`/api/v1/workshops/${encodeURIComponent(id)}`, 'DELETE');
  }

  async provisionWorkshop(id: string): Promise<{ provisioned: number; skipped: number; errors: string[] }> {
    const data = await this.safeRequest<any>(`/api/v1/workshops/${encodeURIComponent(id)}/provision`, 'POST');
    return { provisioned: data?.provisioned || 0, skipped: data?.skipped || 0, errors: data?.errors || [] };
  }

  async getWorkshopDashboard(id: string): Promise<WorkshopDashboard> {
    const data = await this.safeRequest<WorkshopDashboard>(`/api/v1/workshops/${encodeURIComponent(id)}/dashboard`);
    return data as WorkshopDashboard;
  }

  async endWorkshop(id: string): Promise<{ stopped: number; errors: string[] }> {
    const data = await this.safeRequest<any>(`/api/v1/workshops/${encodeURIComponent(id)}/end`, 'POST');
    return { stopped: data?.stopped || 0, errors: data?.errors || [] };
  }

  async getWorkshopDownload(id: string): Promise<{ workshop_id: string; participants: any[] }> {
    const data = await this.safeRequest<any>(`/api/v1/workshops/${encodeURIComponent(id)}/download`);
    return { workshop_id: data?.workshop_id || id, participants: data?.participants || [] };
  }

  async getWorkshopConfigs(): Promise<WorkshopConfig[]> {
    const data = await this.safeRequest<any>('/api/v1/workshops/configs');
    return (data?.configs || []) as WorkshopConfig[];
  }

  async saveWorkshopConfig(workshopId: string, configName: string): Promise<WorkshopConfig> {
    const data = await this.safeRequest<WorkshopConfig>(`/api/v1/workshops/${encodeURIComponent(workshopId)}/config`, 'POST', { name: configName });
    return data as WorkshopConfig;
  }

  async createWorkshopFromConfig(configName: string, workshopData: Partial<WorkshopEvent>): Promise<WorkshopEvent> {
    const data = await this.safeRequest<WorkshopEvent>(`/api/v1/workshops/from-config/${encodeURIComponent(configName)}`, 'POST', workshopData);
    return data as WorkshopEvent;
  }

  // ── v0.19.0 TA Access, Shared Materials, Workspace Reset ──
  async listCourseTAAccess(courseId: string): Promise<ClassMember[]> {
    const data = await this.safeRequest<{ ta_members: ClassMember[] }>(`/api/v1/courses/${encodeURIComponent(courseId)}/ta-access`, 'GET');
    return (data?.ta_members || []) as ClassMember[];
  }

  async grantCourseTAAccess(courseId: string, email: string, displayName?: string): Promise<ClassMember> {
    const data = await this.safeRequest<ClassMember>(`/api/v1/courses/${encodeURIComponent(courseId)}/ta-access`, 'POST', { email, display_name: displayName || '' });
    return data as ClassMember;
  }

  async revokeCourseTAAccess(courseId: string, email: string): Promise<void> {
    await this.safeRequest<void>(`/api/v1/courses/${encodeURIComponent(courseId)}/ta-access/${encodeURIComponent(email)}`, 'DELETE');
  }

  async connectCourseTAAccess(courseId: string, studentId: string, reason: string): Promise<{ ssh_command: string }> {
    const data = await this.safeRequest<{ ssh_command: string }>(`/api/v1/courses/${encodeURIComponent(courseId)}/ta-access/connect`, 'POST', { student_id: studentId, reason });
    return data as { ssh_command: string };
  }

  async resetCourseStudentWorkspace(courseId: string, studentId: string, reason: string, backup: boolean): Promise<WorkspaceResetResult> {
    const data = await this.safeRequest<WorkspaceResetResult>(`/api/v1/courses/${encodeURIComponent(courseId)}/ta/reset/${encodeURIComponent(studentId)}`, 'POST', { reason, backup });
    return data as WorkspaceResetResult;
  }

  async getCourseMaterials(courseId: string): Promise<SharedMaterialsVolume | null> {
    const data = await this.safeRequest<{ materials: SharedMaterialsVolume | null }>(`/api/v1/courses/${encodeURIComponent(courseId)}/materials`, 'GET');
    return data?.materials || null;
  }

  async createCourseMaterials(courseId: string, sizeGB: number, mountPath: string): Promise<SharedMaterialsVolume> {
    const data = await this.safeRequest<{ materials: SharedMaterialsVolume }>(`/api/v1/courses/${encodeURIComponent(courseId)}/materials`, 'POST', { size_gb: sizeGB, mount_path: mountPath });
    return data?.materials as SharedMaterialsVolume;
  }

  async mountCourseMaterials(courseId: string): Promise<{ status: string; note: string }> {
    const data = await this.safeRequest<{ status: string; note: string }>(`/api/v1/courses/${encodeURIComponent(courseId)}/materials/mount`, 'POST', {});
    return data as { status: string; note: string };
  }

  // ── v0.20.0 SSM File Operations (#30) ──────────────────────────────────────
  async listInstanceFiles(instanceName: string, path?: string): Promise<FileEntry[]> {
    const url = `/api/v1/instances/${encodeURIComponent(instanceName)}/files${path ? `?path=${encodeURIComponent(path)}` : ''}`;
    const data = await this.safeRequest<FileEntry[]>(url, 'GET');
    return (data || []) as FileEntry[];
  }

  async pushFileToInstance(instanceName: string, localPath: string, remotePath: string): Promise<{ status: string; message: string }> {
    const data = await this.safeRequest<{ status: string; message: string }>(
      `/api/v1/instances/${encodeURIComponent(instanceName)}/files/push`, 'POST',
      { local_path: localPath, remote_path: remotePath });
    return data as { status: string; message: string };
  }

  async pullFileFromInstance(instanceName: string, remotePath: string, localPath: string): Promise<{ status: string; message: string }> {
    const data = await this.safeRequest<{ status: string; message: string }>(
      `/api/v1/instances/${encodeURIComponent(instanceName)}/files/pull`, 'POST',
      { remote_path: remotePath, local_path: localPath });
    return data as { status: string; message: string };
  }

  // ── v0.20.0 EC2 Capacity Blocks (#63) ──────────────────────────────────────
  async getCapacityBlocks(): Promise<CapacityBlock[]> {
    const data = await this.safeRequest<CapacityBlock[]>('/api/v1/capacity-blocks', 'GET');
    return (data || []) as CapacityBlock[];
  }

  async reserveCapacityBlock(req: CapacityBlockRequest): Promise<CapacityBlock> {
    const data = await this.safeRequest<CapacityBlock>('/api/v1/capacity-blocks', 'POST', req);
    return data as CapacityBlock;
  }

  async describeCapacityBlock(id: string): Promise<CapacityBlock> {
    const data = await this.safeRequest<CapacityBlock>(`/api/v1/capacity-blocks/${encodeURIComponent(id)}`, 'GET');
    return data as CapacityBlock;
  }

  async cancelCapacityBlock(id: string): Promise<void> {
    await this.safeRequest<void>(`/api/v1/capacity-blocks/${encodeURIComponent(id)}`, 'DELETE');
  }

  // S3 mount methods (#22c)

  async listInstanceS3Mounts(instanceName: string): Promise<S3Mount[]> {
    const data = await this.safeRequest<S3Mount[]>(`/api/v1/instances/${encodeURIComponent(instanceName)}/s3-mounts`, 'GET');
    return (data as S3Mount[]) || [];
  }

  async mountS3Bucket(instanceName: string, bucket: string, mountPath: string, method = 'mountpoint', readOnly = false): Promise<S3Mount> {
    const data = await this.safeRequest<S3Mount>(`/api/v1/instances/${encodeURIComponent(instanceName)}/s3-mounts`, 'POST', {
      bucket_name: bucket,
      mount_path: mountPath,
      method,
      read_only: readOnly,
    });
    return data as S3Mount;
  }

  async unmountS3Bucket(instanceName: string, mountPath: string): Promise<void> {
    const encoded = encodeURIComponent(mountPath.replace(/^\//, ''));
    await this.safeRequest<void>(`/api/v1/instances/${encodeURIComponent(instanceName)}/s3-mounts/${encoded}`, 'DELETE');
  }

  // Storage analytics methods (#23c)

  async getAllStorageAnalytics(period = 'daily'): Promise<StorageAnalyticsSummary[]> {
    const data = await this.safeRequest<{ resources?: StorageAnalyticsSummary[] }>(`/api/v1/storage/analytics?period=${encodeURIComponent(period)}`, 'GET');
    const result = data as { resources?: StorageAnalyticsSummary[] };
    return result?.resources || [];
  }

  async getStorageAnalytics(name: string, period = 'daily'): Promise<StorageAnalyticsSummary> {
    const data = await this.safeRequest<StorageAnalyticsSummary>(`/api/v1/storage/analytics/${encodeURIComponent(name)}?period=${encodeURIComponent(period)}`, 'GET');
    return data as StorageAnalyticsSummary;
  }
}

// CoursesManagementView — top-level component (not inside App) to prevent re-mount on state change.
// Pattern matches InvitationManagementView and ApprovalsView.
function CoursesManagementView() {
  return <CoursesPanel />;
}

// WorkshopsManagementView — top-level component to prevent re-mount on state change (#13).
function WorkshopsManagementView() {
  return <WorkshopsPanel />;
}

// CapacityBlocksManagementView — top-level component for EC2 Capacity Blocks (v0.20.0 #63).
function CapacityBlocksManagementView() {
  return <CapacityBlocksPanel />;
}

export default function PrismApp() {
  const api = new SafePrismAPI();

  // Make API client available to ProjectDetailView component
  (window as any).__apiClient = api;

  const [state, setState] = useState<AppState>({
    activeView: 'dashboard',
    settingsSection: 'general',
    templates: {},
    instances: [],
    efsVolumes: [],
    ebsVolumes: [],
    snapshots: [],
    projects: [],
    users: [],
    budgets: [],
    budgetPools: [],
    selectedBudgetId: null,
    amis: [],
    amiBuilds: [],
    amiRegions: [],
    rightsizingRecommendations: [],
    rightsizingStats: null,
    policyStatus: null,
    policySets: [],
    marketplaceTemplates: [],
    marketplaceCategories: [],
    idlePolicies: [],
    idleSchedules: [],
    invitations: [],
    courses: [],
    workshops: [],
    selectedTemplate: null,
    selectedProject: null,
    selectedTerminalInstance: '',
    loading: true,
    notifications: [],
    connected: false,
    error: null,
    updateInfo: null,
    autoStartEnabled: false,
    pendingApprovalsCount: 0
  });

  const [launchModalVisible, setLaunchModalVisible] = useState(false);
  const [launchConfig, setLaunchConfig] = useState({
    name: '',
    size: 'M',
    spot: false,
    hibernation: false,
    dryRun: false
  });

  // Delete confirmation modal state
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [deleteModalConfig, setDeleteModalConfig] = useState<{
    type: 'workspace' | 'efs-volume' | 'ebs-volume' | 'project' | 'user' | null;
    name: string;
    requireNameConfirmation: boolean;
    warning?: string;
    onConfirm: () => Promise<void>;
  }>({
    type: null,
    name: '',
    requireNameConfirmation: false,
    onConfirm: async () => {}
  });

  // Hibernate confirmation modal state
  const [hibernateModalVisible, setHibernateModalVisible] = useState(false);
  const [hibernateModalInstance, setHibernateModalInstance] = useState<Instance | null>(null);

  // Idle policy modal state (Issue #288)
  const [idlePolicyModalInstance, setIdlePolicyModalInstance] = useState<string | null>(null);

  // Onboarding wizard state
  const [onboardingVisible, setOnboardingVisible] = useState(false);
  const [onboardingStep, setOnboardingStep] = useState(0);
  const [onboardingComplete, setOnboardingComplete] = useState(() => {
    // Check if user has completed onboarding before
    const completed = localStorage.getItem('prism_onboarding_complete');
    return completed === 'true';
  });

  // First-time user detection (for context-aware dashboard)
  const [isFirstTimeUser, setIsFirstTimeUser] = useState(() => {
    // Check if user has ever launched a workspace
    const hasLaunched = localStorage.getItem('prism_has_launched_workspace');
    return hasLaunched !== 'true';
  });

  // Quick Start Wizard state
  const [quickStartWizardVisible, setQuickStartWizardVisible] = useState(false);
  const [quickStartActiveStepIndex, setQuickStartActiveStepIndex] = useState(0);
  const [quickStartConfig, setQuickStartConfig] = useState({
    selectedTemplate: null as Template | null,
    workspaceName: '',
    size: 'M',
    launchInProgress: false,
    launchedWorkspaceId: null as string | null
  });

  // Bulk selection state for instances
  const [selectedInstances, setSelectedInstances] = useState<Instance[]>([]);

  // Filtering state for instances table
  const [instancesFilterQuery, setInstancesFilterQuery] = useState<{ tokens: Array<{ propertyKey?: string; operator: string; value: string }>; operation: 'and' | 'or' }>({ tokens: [], operation: 'and' });

  // Create Backup modal state
  const [createBackupModalVisible, setCreateBackupModalVisible] = useState(false);
  const [createBackupConfig, setCreateBackupConfig] = useState({
    instanceId: '',
    backupName: '',
    backupType: 'full',
    description: ''
  });
  const [createBackupValidationAttempted, setCreateBackupValidationAttempted] = useState(false);

  // Delete Backup modal state
  const [deleteBackupModalVisible, setDeleteBackupModalVisible] = useState(false);
  const [selectedBackupForDelete, setSelectedBackupForDelete] = useState<InstanceSnapshot | null>(null);

  // Restore Backup modal state
  const [restoreBackupModalVisible, setRestoreBackupModalVisible] = useState(false);
  const [selectedBackupForRestore, setSelectedBackupForRestore] = useState<InstanceSnapshot | null>(null);
  const [restoreInstanceName, setRestoreInstanceName] = useState('');

  // Storage creation modal state
  const [createEFSModalVisible, setCreateEFSModalVisible] = useState(false);
  const [createEBSModalVisible, setCreateEBSModalVisible] = useState(false);
  const [storageVolumeName, setStorageVolumeName] = useState('');
  const [storageVolumeSize, setStorageVolumeSize] = useState('');
  const [storageVolumeNameError, setStorageVolumeNameError] = useState('');
  const [storageVolumeSizeError, setStorageVolumeSizeError] = useState('');

  // Create Project modal state
  const [projectModalVisible, setProjectModalVisible] = useState(false);
  const [projectName, setProjectName] = useState('');
  const [projectDescription, setProjectDescription] = useState('');
  const [projectBudget, setProjectBudget] = useState('');
  const [projectValidationError, setProjectValidationError] = useState('');

  // Project detail view state
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);

  // Create User modal state
  const [userModalVisible, setUserModalVisible] = useState(false);
  const [username, setUsername] = useState('');
  const [userEmail, setUserEmail] = useState('');
  const [userFullName, setUserFullName] = useState('');
  const [userValidationError, setUserValidationError] = useState('');
  const [creatingUser, setCreatingUser] = useState(false);

  // SSH Key modal state
  const [sshKeyModalVisible, setSshKeyModalVisible] = useState(false);
  const [selectedUsername, setSelectedUsername] = useState<string>('');

  // User details modal state
  const [userDetailsModalVisible, setUserDetailsModalVisible] = useState(false);
  const [selectedUserForDetails, setSelectedUserForDetails] = useState<User | null>(null);
  const [userSSHKeys, setUserSSHKeys] = useState<SSHKeyConfig[]>([]);
  const [loadingSSHKeys, setLoadingSSHKeys] = useState(false);

  // User status management state
  const [userStatusModalVisible, setUserStatusModalVisible] = useState(false);
  const [selectedUserForStatus, setSelectedUserForStatus] = useState<User | null>(null);
  const [userStatusDetails, setUserStatusDetails] = useState<UserStatus | null>(null);
  const [loadingUserStatus, setLoadingUserStatus] = useState(false);

  // User provisioning state
  const [provisionModalVisible, setProvisionModalVisible] = useState(false);
  const [selectedUserForProvision, setSelectedUserForProvision] = useState<User | null>(null);
  const [selectedWorkspaceForProvision, setSelectedWorkspaceForProvision] = useState<string>('');
  const [provisioningInProgress, setProvisioningInProgress] = useState(false);

  // Connection info modal state
  const [connectionModalVisible, setConnectionModalVisible] = useState(false);
  const [connectionInfo, setConnectionInfo] = useState<{
    instanceName: string;
    publicIP: string;
    sshCommand: string;
    webPort: string;
  } | null>(null);

  // Create Budget Pool modal state
  const [createBudgetModalVisible, setCreateBudgetModalVisible] = useState(false);
  const [budgetName, setBudgetName] = useState('');
  const [budgetDescription, setBudgetDescription] = useState('');
  const [totalAmount, setTotalAmount] = useState('');
  const [period, setPeriod] = useState('monthly');
  const [alertThreshold, setAlertThreshold] = useState('80');
  const [budgetValidationError, setBudgetValidationError] = useState('');

  // Track users data version to prevent stale data overwrites
  // This prevents initial page load getUsers() from overwriting optimistic updates
  const usersVersionRef = useRef(0);

  // Individual invitation states
  const [sendInvitationModalVisible, setSendInvitationModalVisible] = useState(false);
  const [selectedProjectForInvitation, setSelectedProjectForInvitation] = useState<string>('');
  const [invitationEmail, setInvitationEmail] = useState('');
  const [invitationRole, setInvitationRole] = useState<'viewer' | 'member' | 'admin'>('member');
  const [invitationMessage, setInvitationMessage] = useState('');
  const [invitationValidationError, setInvitationValidationError] = useState('');

  // Invitation token redemption
  const [redeemTokenModalVisible, setRedeemTokenModalVisible] = useState(false);
  const [invitationToken, setInvitationToken] = useState('');
  const [tokenValidationError, setTokenValidationError] = useState('');

  // Project management modals
  const [showEditProjectModal, setShowEditProjectModal] = useState(false);
  const [selectedProjectForEdit, setSelectedProjectForEdit] = useState<Project | null>(null);
  const [editProjectName, setEditProjectName] = useState('');
  const [editProjectDescription, setEditProjectDescription] = useState('');
  const [editProjectStatus, setEditProjectStatus] = useState('');
  const [editProjectSubmitting, setEditProjectSubmitting] = useState(false);
  const [showManageMembersModal, setShowManageMembersModal] = useState(false);
  const [selectedProjectForMembers, setSelectedProjectForMembers] = useState<Project | null>(null);
  const [manageMembersData, setManageMembersData] = useState<MemberData[]>([]);
  const [manageMembersLoading, setManageMembersLoading] = useState(false);
  const [addMemberUsername, setAddMemberUsername] = useState('');
  const [addMemberRole, setAddMemberRole] = useState('member');
  const [showBudgetModal, setShowBudgetModal] = useState(false);
  const [selectedProjectForBudget, setSelectedProjectForBudget] = useState<Project | null>(null);
  const [budgetModalData, setBudgetModalData] = useState<BudgetData | null>(null);
  const [budgetModalLoading, setBudgetModalLoading] = useState(false);
  const [showCostModal, setShowCostModal] = useState(false);
  const [selectedProjectForCosts, setSelectedProjectForCosts] = useState<Project | null>(null);
  const [costModalData, setCostModalData] = useState<CostBreakdown | null>(null);
  const [costModalLoading, setCostModalLoading] = useState(false);
  const [showUsageModal, setShowUsageModal] = useState(false);
  const [selectedProjectForUsage, setSelectedProjectForUsage] = useState<Project | null>(null);
  const [usageModalData, setUsageModalData] = useState<ProjectUsageResponse | null>(null);
  const [usageModalLoading, setUsageModalLoading] = useState(false);

  // User management modals
  const [showEditUserModal, setShowEditUserModal] = useState(false);
  const [selectedUserForEdit, setSelectedUserForEdit] = useState<User | null>(null);
  const [editUserEmail, setEditUserEmail] = useState('');
  const [editUserDisplayName, setEditUserDisplayName] = useState('');
  const [editUserRole, setEditUserRole] = useState('');
  const [editUserSubmitting, setEditUserSubmitting] = useState(false);

  // Helper: add a toast notification
  // Supports two calling conventions:
  //   addNotification('error', 'Header', 'Content')
  //   addNotification({ type: 'success', content: '...' })
  const addNotification = (
    typeOrObj: 'success' | 'error' | 'warning' | 'info' | Partial<{ type: string; header?: string; content: string; dismissible?: boolean }>,
    header?: string,
    content?: string
  ) => {
    const type = typeof typeOrObj === 'string' ? typeOrObj : (typeOrObj.type || 'info');
    const title = typeof typeOrObj === 'string' ? (header || content || '') : (typeOrObj.header || typeOrObj.content || '');
    const desc = typeof typeOrObj === 'string' ? (header && content ? content : undefined) : (typeOrObj.header && typeOrObj.content ? typeOrObj.content : undefined);
    const toastFn = type === 'success' ? toast.success : type === 'error' ? toast.error : type === 'warning' ? toast.warning : toast;
    toastFn(title, desc ? { description: desc } : undefined);
  };

  // Helper: set (replace/update) a toast notification
  const setNotification = (notification: Partial<{ id?: string; type?: string; header?: string; content: string; dismissible?: boolean }>) => {
    const title = notification.header || notification.content || '';
    const desc = notification.header && notification.content ? notification.content : undefined;
    const type = notification.type || 'info';
    const toastFn = type === 'success' ? toast.success : type === 'error' ? toast.error : type === 'warning' ? toast.warning : toast;
    toastFn(title, { ...(desc ? { description: desc } : {}), ...(notification.id ? { id: notification.id } : {}) });
  };

  // Safe data loading with comprehensive error handling
  // Wrapped in useCallback to prevent unnecessary re-renders in dependent useEffect hooks
  const loadApplicationData = React.useCallback(async () => {
    try {
      setState(prev => ({ ...prev, loading: true, error: null }));

      // Capture current users version BEFORE starting async API calls
      // This allows us to detect if optimistic updates occurred during the API call
      const usersVersionBeforeLoad = usersVersionRef.current;

      // Use Promise.allSettled to allow individual API calls to fail without breaking the entire load
      // This is essential for test environments where some endpoints may not have AWS credentials
      // NOTE: Budgets and budget pools are loaded separately via loadBudgetData() to avoid excessive API calls
      const results = await Promise.allSettled([
        api.getTemplates(),
        api.getInstances(),
        api.getEFSVolumes(),
        api.getEBSVolumes(),
        api.getSnapshots(),
        api.getProjects(),
        api.getUsers(),
        api.getAMIs(),
        api.getAMIBuilds(),
        api.getAMIRegions(),
        api.getRightsizingRecommendations(),
        // api.getRightsizingStats() - Removed: requires instance name parameter, called per-instance instead
        api.getPolicyStatus(),
        api.getPolicySets(),
        api.getMarketplaceTemplates(),
        api.getMarketplaceCategories(),
        api.getIdlePolicies(),
        api.getIdleSchedules(),
        api.getMyInvitations(),
        api.getAutoStartStatus(),
        api.getCourses(),
        api.getWorkshops()
      ]);

      // Extract successful results, using empty fallbacks for failed promises
      const rawResults = results.map((result, index) => {
        if (result.status === 'fulfilled') {
          return result.value;
        } else {
          // Return appropriate empty fallback based on expected type
          if (index === 0) return {}; // templates (object)
          if (index === 10) return null; // policyStatus (nullable, adjusted index after removing budgets)
          if (index === 18) return { enabled: false }; // autoStartStatus (object with enabled boolean, adjusted index)
          return []; // everything else (arrays)
        }
      });
      const templatesData = rawResults[0] as Record<string, Template>;
      const instancesData = rawResults[1] as Instance[];
      const efsVolumesData = rawResults[2] as EFSVolume[];
      const ebsVolumesData = rawResults[3] as EBSVolume[];
      const snapshotsData = rawResults[4] as InstanceSnapshot[];
      const projectsData = rawResults[5] as Project[];
      const usersData = rawResults[6] as User[];
      const amisData = rawResults[7] as AMI[];
      const amiBuildsData = rawResults[8] as AMIBuild[];
      const amiRegionsData = rawResults[9] as AMIRegion[];
      const rightsizingRecommendationsData = rawResults[10] as RightsizingRecommendation[];
      const policyStatusData = rawResults[11] as PolicyStatus | null;
      const policySetsData = rawResults[12] as PolicySet[];
      const marketplaceTemplatesData = rawResults[13] as MarketplaceTemplate[];
      const marketplaceCategoriesData = rawResults[14] as MarketplaceCategory[];
      const idlePoliciesData = rawResults[15] as IdlePolicy[];
      const idleSchedulesData = rawResults[16] as IdleSchedule[];
      const invitationsData = rawResults[17] as Invitation[];
      const autoStartStatusData = rawResults[18] as { enabled: boolean } | null;
      const coursesData = rawResults[19] as Course[];
      const workshopsData = rawResults[20] as WorkshopEvent[];

      // Initialize rightsizingStatsData since api.getRightsizingStats() was removed (requires instance name parameter)
      const rightsizingStatsData = null;

      // Convert Invitation[] to CachedInvitation[] format
      const cachedInvitations: CachedInvitation[] = (invitationsData || []).map((inv: Invitation) => ({
        token: inv.token,
        invitation_id: inv.id,
        project_id: inv.project_id,
        project_name: inv.project_name,
        email: inv.email,
        role: inv.role,
        invited_by: inv.invited_by,
        invited_at: inv.invited_at,
        expires_at: inv.expires_at,
        status: inv.status,
        message: inv.message || '',
        added_at: new Date().toISOString()
      }));

      setState(prev => ({
        ...prev,
        templates: templatesData,
        instances: instancesData,
        efsVolumes: efsVolumesData,
        ebsVolumes: ebsVolumesData,
        snapshots: snapshotsData,
        projects: projectsData,
        // Only update users if version hasn't changed (no optimistic updates occurred)
        // This prevents stale API data from overwriting fresh optimistic updates
        users: usersVersionRef.current === usersVersionBeforeLoad ? usersData : prev.users,
        // budgets and budgetPools loaded separately via loadBudgetData()
        amis: amisData,
        amiBuilds: amiBuildsData,
        amiRegions: amiRegionsData,
        rightsizingRecommendations: rightsizingRecommendationsData,
        rightsizingStats: rightsizingStatsData,
        policyStatus: policyStatusData,
        policySets: policySetsData,
        marketplaceTemplates: marketplaceTemplatesData,
        marketplaceCategories: marketplaceCategoriesData,
        idlePolicies: idlePoliciesData,
        idleSchedules: idleSchedulesData,
        invitations: cachedInvitations,
        courses: coursesData || [],
        workshops: workshopsData || [],
        autoStartEnabled: autoStartStatusData?.enabled || false,
        loading: false,
        connected: true,
        error: null
      }));


      // Load pending approvals count (fire-and-forget)
      api.listAllApprovals('pending').then(approvals =>
        setState(prev => ({ ...prev, pendingApprovalsCount: approvals.length }))
      ).catch(() => {});

    } catch (error) {
      logger.error('Failed to load application data:', error);

      toast.error('Connection Error', { description: `Failed to connect to Prism daemon: ${error instanceof Error ? error.message : 'Unknown error'}` });
      setState(prev => ({
        ...prev,
        loading: false,
        connected: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      }));
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty deps: api and setState are stable references that don't change

  // Load budget data separately (only when viewing budgets/projects)
  // This prevents excessive API calls (one per project) during normal operations
  const loadBudgetData = React.useCallback(async () => {
    try {
      const results = await Promise.allSettled([
        api.getBudgets(),
        api.getBudgetPools()
      ]);

      const budgetsData = results[0].status === 'fulfilled' ? results[0].value as BudgetData[] : [];
      const budgetPoolsData = results[1].status === 'fulfilled' ? results[1].value as Budget[] : [];

      setState(prev => ({
        ...prev,
        budgets: budgetsData,
        budgetPools: budgetPoolsData
      }));
    } catch (error) {
      logger.error('Failed to load budget data:', error);
    }
  }, [api]);

  // Load budget data when switching to budgets view or project detail
  // NOTE: Removed 'projects' to avoid N+1 query problem (Issue #457)
  // Projects table doesn't need budget data - only Budgets page and Project Detail need it
  useEffect(() => {
    if (state.activeView === 'budgets' || state.activeView === 'project-detail') {
      loadBudgetData();
    }
  }, [state.activeView, loadBudgetData]);

  // Utility function to get accessible status labels (WCAG 1.1.1)
  const getStatusLabel = (context: string, status: string, additionalInfo?: string): string => {
    const labels: Record<string, Record<string, string>> = {
      instance: {
        running: 'Workspace running',
        stopped: 'Workspace stopped',
        pending: 'Workspace pending',
        stopping: 'Workspace stopping',
        terminated: 'Workspace terminated',
        hibernated: 'Workspace hibernated'
      },
      volume: {
        available: 'Volume available',
        'in-use': 'Volume in use',
        creating: 'Volume creating',
        deleting: 'Volume deleting'
      },
      project: {
        active: 'Project active',
        suspended: 'Project suspended',
        archived: 'Project archived'
      },
      user: {
        active: 'User active',
        inactive: 'User inactive'
      },
      connection: {
        success: 'Connected to daemon',
        error: 'Disconnected from daemon'
      },
      ami: {
        available: 'AMI available',
        pending: 'AMI pending',
        failed: 'AMI failed'
      },
      build: {
        completed: 'Build completed',
        failed: 'Build failed',
        'in-progress': 'Build in progress'
      },
      budget: {
        ok: 'Budget OK',
        warning: 'Budget warning',
        critical: 'Budget critical'
      },
      policy: {
        enabled: 'Policy enabled',
        disabled: 'Policy disabled'
      },
      marketplace: {
        verified: 'Verified publisher',
        community: 'Community template'
      },
      idle: {
        enabled: 'Idle detection enabled',
        disabled: 'Idle detection disabled'
      },
      auth: {
        authenticated: 'Authenticated'
      }
    };
    const label = labels[context]?.[status] || `${context} ${status}`;
    return additionalInfo ? `${label}: ${additionalInfo}` : label;
  };

  // Load data on mount and refresh periodically
  // NOTE: Budget loading on navigation is handled by the separate effect below (line 2145)
  // This effect intentionally uses [] deps to avoid re-triggering on every navigation
  useEffect(() => {
    loadApplicationData();
    const interval = setInterval(loadApplicationData, 30000);
    return () => clearInterval(interval);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only on mount - prevents unnecessary reloads on every navigation

  // Check for updates on mount
  useEffect(() => {
    const checkUpdates = async () => {
      try {
        const updateInfo = await api.checkForUpdates();
        setState(prev => ({ ...prev, updateInfo }));
      } catch (error) {
        logger.error('Failed to check for updates:', error);
      }
    };
    checkUpdates();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Run only on mount

  // Show onboarding for first-time users
  useEffect(() => {
    if (!onboardingComplete && state.connected && !state.loading) {
      // Show onboarding after a short delay to let the UI settle
      const timer = setTimeout(() => {
        setOnboardingVisible(true);
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [onboardingComplete, state.connected, state.loading]);

  // Update first-time user status when user launches workspaces
  useEffect(() => {
    if (state.instances.length > 0) {
      localStorage.setItem('prism_has_launched_workspace', 'true');
      setIsFirstTimeUser(false);
    }
  }, [state.instances]);

  // Keyboard shortcuts for common actions
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      // Skip if user is typing in an input field
      const target = event.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
        return;
      }

      // Cmd/Ctrl + R: Refresh data
      if ((event.metaKey || event.ctrlKey) && event.key === 'r') {
        event.preventDefault();
        loadApplicationData();
        toast.success('Data refreshed');
      }

      // Cmd/Ctrl + K: Focus search/filter
      if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
        event.preventDefault();
        // Focus first search input if available
        const searchInput = document.querySelector('input[type="search"]') as HTMLInputElement;
        if (searchInput) searchInput.focus();
      }

      // Number keys 1-9: Navigate to views
      if (!event.metaKey && !event.ctrlKey && !event.altKey) {
        const viewMap: Record<string, string> = {
          '1': 'dashboard',
          '2': 'templates',
          '3': 'workspaces',
          '4': 'storage',
          '5': 'projects',
          '6': 'users',
          '7': 'settings'
        };
        if (viewMap[event.key]) {
          event.preventDefault();
          setState(prev => ({ ...prev, activeView: viewMap[event.key] as AppState['activeView'] }));
        }
      }

      // ? : Show keyboard shortcuts help
      if (event.key === '?' && !event.shiftKey) {
        toast('Keyboard Shortcuts', { description: 'Cmd/Ctrl+R: Refresh | Cmd/Ctrl+K: Search | 1-7: Navigate views | ?: Help' });
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [state.activeView, loadApplicationData]);

  // Handle navigation events dispatched by the Go backend (e.g. system tray menu clicks)
  useEffect(() => {
    const handler = (e: Event) => {
      const view = (e as CustomEvent<string>).detail;
      setState(prev => ({ ...prev, activeView: view as typeof prev.activeView }));
    };
    window.addEventListener('prism-navigate', handler);
    return () => window.removeEventListener('prism-navigate', handler);
  }, []);

  // Safe template selection
  const handleTemplateSelection = (template: Template) => {
    try {
      setState(prev => ({ ...prev, selectedTemplate: template }));
      setLaunchConfig({ name: '', size: 'M', spot: false, hibernation: false, dryRun: false });
      setLaunchModalVisible(true);
    } catch (error) {
      logger.error('Template selection failed:', error);
    }
  };

  // Handle modal dismissal
  const handleModalDismiss = () => {
    setLaunchModalVisible(false);
    setState(prev => ({ ...prev, selectedTemplate: null }));
  };

  // Safe instance launch
  const handleLaunchInstance = async () => {
    if (!state.selectedTemplate || !launchConfig.name.trim()) {
      return;
    }

    // Capture inputs before closing modal
    const templateSlug = getTemplateSlug(state.selectedTemplate);
    const templateName = getTemplateName(state.selectedTemplate);
    const instanceName = launchConfig.name;
    const instanceSize = launchConfig.size;
    const isDryRun = launchConfig.dryRun || false;

    // Close modal IMMEDIATELY
    handleModalDismiss();

    // Show progress notification
    setState(prev => ({
      ...prev,
      notifications: [
        {
          type: 'info',
          header: 'Launching Workspace',
          content: `Launching ${instanceName}... This may take a few minutes.`,
          dismissible: true,
          id: Date.now().toString()
        },
        ...prev.notifications
      ]
    }));

    // Fire-and-forget
    try {
      const result = await api.launchInstance(templateSlug, instanceName, instanceSize, isDryRun);
      // HTTP 202 approval pending (#495)
      if (result && (result as any).approval_pending) {
        const approvalId = (result as any).approval_request_id || 'unknown';
        setState(prev => ({
          ...prev,
          notifications: [
            {
              type: 'info',
              header: 'Approval Required',
              content: `Launch of ${instanceName} requires PI approval. Request created: ${approvalId}. Check the Approvals panel.`,
              dismissible: true,
              id: Date.now().toString()
            },
            ...prev.notifications
          ]
        }));
        // Refresh pending approvals count
        api.listAllApprovals('pending').then(approvals =>
          setState(prev => ({ ...prev, pendingApprovalsCount: approvals.length }))
        ).catch(() => {});
        return;
      }
      setState(prev => ({
        ...prev,
        notifications: [
          {
            type: 'success',
            header: 'Workspace Launched',
            content: `Successfully launched ${instanceName} using ${templateName}`,
            dismissible: true,
            id: Date.now().toString()
          },
          ...prev.notifications
        ]
      }));
      // Reload data in background (don't block the success notification)
      setTimeout(loadApplicationData, 1000);
    } catch (error) {
      setState(prev => ({
        ...prev,
        notifications: [
          {
            type: 'error',
            header: 'Launch Failed',
            content: `Failed to launch ${instanceName}: ${error instanceof Error ? error.message : 'Unknown error'}`,
            dismissible: true,
            id: Date.now().toString()
          },
          ...prev.notifications
        ]
      }));
    }
  };

  // Handle Create Project
  const handleCreateProject = async () => {
    // Validate
    if (!projectName.trim()) {
      setProjectValidationError('Project name is required');
      return;
    }

    // Note: Budget validation removed - budget field is deprecated in v0.5.10
    // Budget configuration is now managed separately via Budget/Allocation system

    try {
      // Call API to create project - send budget via budget.total_budget (backend format)
      const budgetPayload = projectBudget ? { budget: { total_budget: parseFloat(projectBudget) } } : {};
      const createdProject = await api.createProject({
        name: projectName.trim(),
        description: projectDescription.trim(),
        ...budgetPayload
      });

      // Map backend response (types.Project with budget.total_budget) to frontend
      // ProjectSummary format (with budget_status.total_budget) for the optimistic update
      const rawProject = createdProject as any;
      const projectForState = {
        ...createdProject,
        budget_status: rawProject.budget ? {
          total_budget: rawProject.budget.total_budget,
          spent_amount: rawProject.budget.spent_amount || 0,
          spent_percentage: 0,
          alert_count: 0,
        } : undefined
      } as Project;

      // Optimistic UI update: add project directly to state without re-fetching
      // Prepend new project so it appears at top of list (page 1) - fixes Issue #457
      setState(prev => ({
        ...prev,
        projects: [projectForState, ...prev.projects],
        notifications: [{
          type: 'success',
          header: 'Project Created',
          content: `Project "${projectName}" created successfully`,
          dismissible: true,
          id: Date.now().toString()
        }, ...prev.notifications]
      }));

      setProjectValidationError('');
      setProjectModalVisible(false);
      setProjectName('');
      setProjectDescription('');
      setProjectBudget('');
      // Refresh data from backend to get accurate budget status (e.g. test-mode mock spend)
      setTimeout(loadApplicationData, 500);
    } catch (error: any) {
      setProjectValidationError(`Failed to create project: ${error.message || 'Unknown error'}`);
    }
  };

  // Handle Create User
  const handleCreateUser = async () => {
    // Validate
    if (!username.trim()) {
      setUserValidationError('Username is required');
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (userEmail && !emailRegex.test(userEmail)) {
      setUserValidationError('Please enter a valid email address');
      return;
    }

    setCreatingUser(true);
    try {
      // Call API to create user - it returns the created user object
      const newUser = await api.createUser({
        username: username.trim(),
        email: userEmail.trim(),
        display_name: userFullName.trim()
      });

      // Increment users version to mark data as fresh
      // This prevents stale API responses from overwriting our optimistic update
      usersVersionRef.current++;

      // Optimistic UI update: Add the new user directly to state
      // This eliminates race conditions from concurrent getUsers() calls
      setState(prev => ({
        ...prev,
        users: [...prev.users, newUser], // Add new user to existing list
        notifications: [{
          type: 'success',
          header: 'User Created',
          content: `User "${username}" created successfully`,
          dismissible: true,
          id: Date.now().toString()
        }, ...prev.notifications]
      }));

      setUserValidationError('');
      setUserModalVisible(false);
      setUsername('');
      setUserEmail('');
      setUserFullName('');
    } catch (error: any) {
      // Check for duplicate error - backend returns HTTP 409
      // Error format: "HTTP 409: Conflict" or error.response.status === 409
      const is409 = error.response?.status === 409 ||
                   (error.message && error.message.includes('HTTP 409'));

      if (is409) {
        setUserValidationError('A user with this username already exists');
      } else {
        setUserValidationError(`Failed to create user: ${error.message || 'Unknown error'}`);
      }
    } finally {
      setCreatingUser(false);
    }
  };

  // Handle Create Budget Pool
  const handleCreateBudget = async () => {
    // Validation
    if (!budgetName.trim()) {
      setBudgetValidationError('Budget name is required');
      return;
    }
    if (!totalAmount || parseFloat(totalAmount) <= 0) {
      setBudgetValidationError('Total amount must be greater than 0');
      return;
    }

    try {
      const createdBudget = await api.createBudgetPool({
        name: budgetName.trim(),
        description: budgetDescription.trim(),
        total_amount: parseFloat(totalAmount),
        period: period,
        start_date: new Date().toISOString(),
        alert_threshold: parseFloat(alertThreshold) / 100,
        created_by: 'current-user'  // TODO: Get from auth context
      });

      // Optimistic UI update (don't re-fetch)
      setState(prev => ({
        ...prev,
        budgetPools: [...prev.budgetPools, createdBudget],
        notifications: [{
          type: 'success',
          header: 'Budget Created',
          content: `Budget pool "${budgetName}" created successfully`,
          dismissible: true,
          id: Date.now().toString()
        }, ...prev.notifications]
      }));

      // Reset and close
      setBudgetValidationError('');
      setCreateBudgetModalVisible(false);
      setBudgetName('');
      setBudgetDescription('');
      setTotalAmount('');
      setPeriod('monthly');
      setAlertThreshold('80');
    } catch (error: any) {
      setBudgetValidationError(`Failed to create budget: ${error.message || 'Unknown error'}`);
    }
  };

  // Handle Delete Budget Pool
  // SSH Key generation handler
  const handleGenerateSSHKey = async (username: string): Promise<any> => {
    try {
      const response = await api.generateSSHKey(username);

      // Refresh users list to update SSH key status
      const users = await api.getUsers();

      // Increment users version to mark this data as fresh
      // This prevents stale data from overwriting our updated list
      usersVersionRef.current++;

      // Update state with both users and notification in single atomic operation
      setState(prev => ({
        ...prev,
        users,
        notifications: [{
          type: 'success',
          header: 'SSH Key Generated',
          content: `SSH key pair generated successfully for user "${username}". Download the private key before closing the dialog.`,
          dismissible: true,
          id: Date.now().toString()
        }, ...prev.notifications]
      }));

      return response;
    } catch (error: any) {
      throw new Error(error.message || 'Failed to generate SSH key');
    }
  };

  // Individual Invitation Handlers
  const handleSendInvitation = async () => {
    // Validate
    if (!invitationEmail.trim()) {
      setInvitationValidationError('Email address is required');
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(invitationEmail)) {
      setInvitationValidationError('Please enter a valid email address');
      return;
    }

    if (!selectedProjectForInvitation) {
      setInvitationValidationError('Please select a project');
      return;
    }

    try {
      await api.sendInvitation(selectedProjectForInvitation, invitationEmail.trim(), invitationRole, invitationMessage.trim() || undefined);

      // Show success notification
      setState(prev => ({
        ...prev,
        notifications: [{
          type: 'success',
          header: 'Invitation Sent',
          content: `Invitation sent to ${invitationEmail}`,
          dismissible: true,
          id: Date.now().toString()
        }, ...prev.notifications]
      }));

      // Reset form
      setInvitationValidationError('');
      setSendInvitationModalVisible(false);
      setInvitationEmail('');
      setInvitationMessage('');
      setInvitationRole('member');
    } catch (error: any) {
      setInvitationValidationError(`Failed to send invitation: ${error.message || 'Unknown error'}`);
    }
  };

  const handleRedeemToken = async () => {
    if (!invitationToken.trim()) {
      setTokenValidationError('Invitation token is required');
      return;
    }

    try {
      const invitationData = await api.getInvitationByToken(invitationToken.trim());

      // Show confirmation with invitation details
      const confirmed = confirm(`Accept invitation to project "${invitationData.project_name}" as ${invitationData.role}?`);

      if (confirmed) {
        await api.acceptInvitation(invitationToken.trim());

        setState(prev => ({
          ...prev,
          notifications: [{
            type: 'success',
            header: 'Token Redeemed',
            content: `Successfully joined project "${invitationData.project_name}"`,
            dismissible: true,
            id: Date.now().toString()
          }, ...prev.notifications]
        }));

        setTokenValidationError('');
        setRedeemTokenModalVisible(false);
        setInvitationToken('');
      }
    } catch (error: any) {
      setTokenValidationError(`Failed to redeem token: ${error.message || 'Invalid or expired token'}`);
    }
  };

  // Safe accessors for template data
  // Comprehensive Instance Action Handler
  const handleInstanceAction = async (action: string, instance: Instance) => {
    // Hibernate requires a confirmation dialog with educational content
    if (action === 'hibernate') {
      setHibernateModalInstance(instance);
      setHibernateModalVisible(true);
      return;
    }

    // Lifecycle actions use fire-and-forget (no global loading state)
    const lifecycleActions: Record<string, [string, string]> = {
      start: ['Starting', 'Started'],
      stop: ['Stopping', 'Stopped'],
      hibernate: ['Hibernating', 'Hibernated'],
      resume: ['Resuming', 'Resumed'],
    };

    if (lifecycleActions[action]) {
      const [progressLabel, completeLabel] = lifecycleActions[action];

      // Show progress notification immediately (no loading state)
      setState(prev => ({
        ...prev,
        notifications: [
          ...prev.notifications,
          {
            type: 'info',
            header: `${progressLabel} Workspace`,
            content: `${progressLabel} ${instance.name}...`,
            dismissible: true,
            id: Date.now().toString()
          }
        ]
      }));

      // Fire-and-forget
      try {
        switch (action) {
          case 'start': await api.startInstance(instance.name); break;
          case 'stop': await api.stopInstance(instance.name); break;
          case 'hibernate': await api.hibernateInstance(instance.name); break;
          case 'resume': await api.resumeInstance(instance.name); break;
        }
        await loadApplicationData();
        setState(prev => ({
          ...prev,
          notifications: [
            ...prev.notifications,
            {
              type: 'success',
              header: `Workspace ${completeLabel}`,
              content: `${instance.name} ${completeLabel.toLowerCase()} successfully`,
              dismissible: true,
              id: Date.now().toString()
            }
          ]
        }));
      } catch (error) {
        logger.error(`Failed to ${action} workspace ${instance.name}:`, error);
        setState(prev => ({
          ...prev,
          notifications: [
            ...prev.notifications,
            {
              type: 'error',
              header: 'Action Failed',
              content: `Failed to ${action} ${instance.name}: ${error instanceof Error ? error.message : String(error)}`,
              dismissible: true,
              id: Date.now().toString()
            }
          ]
        }));
      }
      return;
    }

    try {
      setState(prev => ({ ...prev, loading: true }));

      let actionMessage = '';

      switch (action) {
        case 'connect': {
          // Show connection info modal (fire-and-forget style - no loading state)
          const ip = instance.public_ip || '';
          const user = instance.username || 'ubuntu';
          const sshCmd = ip ? `ssh ${user}@${ip}` : `ssh ${user}@<instance-ip>`;
          setState(prev => ({ ...prev, loading: false }));
          setConnectionInfo({
            instanceName: instance.name,
            publicIP: ip,
            sshCommand: sshCmd,
            webPort: ''
          });
          setConnectionModalVisible(true);
          return;
        }
        case 'terminal':
          // Open terminal view with this instance pre-selected
          setState(prev => ({
            ...prev,
            activeView: 'terminal',
            selectedTerminalInstance: instance.name,
            loading: false
          }));
          return; // Don't continue with normal flow
        case 'webservice':
          // Open webview view (user will select specific service)
          setState(prev => ({
            ...prev,
            activeView: 'webview',
            loading: false
          }));
          return; // Don't continue with normal flow
        case 'manage-idle-policy':
          setState(prev => ({ ...prev, loading: false }));
          setIdlePolicyModalInstance(instance.name);
          return;
        case 'delete':
          // Show confirmation modal instead of deleting immediately
          setState(prev => ({ ...prev, loading: false }));
          setDeleteModalConfig({
            type: 'workspace',
            name: instance.name,
            requireNameConfirmation: true,
            onConfirm: async () => {
              try {
                await api.deleteInstance(instance.name);
                setState(prev => ({
                  ...prev,
                  notifications: [
                    ...prev.notifications,
                    {
                      type: 'success',
                      header: 'Workspace Deleted',
                      content: `Successfully deleted workspace ${instance.name}`,
                      dismissible: true,
                      id: Date.now().toString()
                    }
                  ]
                }));
                setDeleteModalVisible(false);
                setTimeout(loadApplicationData, 1000);
              } catch (error) {
                setState(prev => ({
                  ...prev,
                  notifications: [
                    ...prev.notifications,
                    {
                      type: 'error',
                      header: 'Delete Failed',
                      content: `Failed to delete workspace: ${error instanceof Error ? error.message : 'Unknown error'}`,
                      dismissible: true,
                      id: Date.now().toString()
                    }
                  ]
                }));
              }
            }
          });
          setDeleteModalVisible(true);
          return; // Don't continue with normal flow
        default:
          throw new Error(`Unknown action: ${action}`);
      }

      // Add success notification
      setState(prev => ({
        ...prev,
        loading: false,
        notifications: [
          ...prev.notifications,
          {
            type: 'success',
            header: 'Action Successful',
            content: actionMessage,
            dismissible: true,
            id: Date.now().toString()
          }
        ]
      }));

      // Refresh instances after action
      setTimeout(loadApplicationData, 1000);

    } catch (error) {
      logger.error(`Failed to ${action} workspace ${instance.name}:`, error);

      setState(prev => ({
        ...prev,
        loading: false,
        notifications: [
          ...prev.notifications,
          {
            type: 'error',
            header: 'Action Failed',
            content: `Failed to ${action} workspace ${instance.name}: ${error instanceof Error ? error.message : 'Unknown error'}`,
            dismissible: true,
            id: Date.now().toString()
          }
        ]
      }));
    }
  };

  // Bulk action handlers for multiple instances
  const handleBulkAction = async (action: 'start' | 'stop' | 'hibernate' | 'delete') => {
    if (selectedInstances.length === 0) {
      setState(prev => ({
        ...prev,
        notifications: [
          ...prev.notifications,
          {
            type: 'warning',
            header: 'No Workspaces Selected',
            content: 'Please select one or more workspaces to perform bulk actions.',
            dismissible: true,
            id: Date.now().toString()
          }
        ]
      }));
      return;
    }

    // For delete, show confirmation modal
    if (action === 'delete') {
      setDeleteModalConfig({
        type: 'workspace',
        name: `${selectedInstances.length} workspace${selectedInstances.length > 1 ? 's' : ''}`,
        requireNameConfirmation: false,
        onConfirm: async () => {
          await executeBulkAction('delete');
          setDeleteModalVisible(false);
        }
      });
      setDeleteModalVisible(true);
      return;
    }

    // Execute non-delete bulk actions immediately
    await executeBulkAction(action);
  };

  const executeBulkAction = async (action: 'start' | 'stop' | 'hibernate' | 'delete') => {
    try {
      setState(prev => ({ ...prev, loading: true }));

      // Execute action on all selected instances
      const results = await Promise.allSettled(
        selectedInstances.map(async (instance) => {
          switch (action) {
            case 'start':
              return await api.startInstance(instance.name);
            case 'stop':
              return await api.stopInstance(instance.name);
            case 'hibernate':
              return await api.hibernateInstance(instance.name);
            case 'delete':
              return await api.deleteInstance(instance.name);
            default:
              throw new Error(`Unknown action: ${action}`);
          }
        })
      );

      // Count successes and failures
      const successes = results.filter(r => r.status === 'fulfilled').length;
      const failures = results.filter(r => r.status === 'rejected').length;

      // Show notification with results
      setState(prev => ({
        ...prev,
        loading: false,
        notifications: [
          ...prev.notifications,
          {
            type: failures > 0 ? 'warning' : 'success',
            header: `Bulk ${action.charAt(0).toUpperCase() + action.slice(1)} ${failures > 0 ? 'Partially Complete' : 'Complete'}`,
            content: `Successfully ${action}ed ${successes} workspace${successes !== 1 ? 's' : ''}${failures > 0 ? `, failed to ${action} ${failures} workspace${failures !== 1 ? 's' : ''}` : ''}.`,
            dismissible: true,
            id: Date.now().toString()
          }
        ]
      }));

      // Clear selection and refresh data
      setSelectedInstances([]);
      setTimeout(loadApplicationData, 1000);

    } catch (error) {
      logger.error(`Failed to execute bulk ${action}:`, error);

      setState(prev => ({
        ...prev,
        loading: false,
        notifications: [
          ...prev.notifications,
          {
            type: 'error',
            header: 'Bulk Action Failed',
            content: `Failed to ${action} workspaces: ${error instanceof Error ? error.message : 'Unknown error'}`,
            dismissible: true,
            id: Date.now().toString()
          }
        ]
      }));
    }
  };

  // Filter instances based on PropertyFilter query
  const getFilteredInstances = () => {
    if (!instancesFilterQuery.tokens || instancesFilterQuery.tokens.length === 0) {
      return state.instances;
    }

    return state.instances.filter((instance) => {
      return instancesFilterQuery.tokens.every((token: { propertyKey?: string; value: string; operator?: string }) => {
        const { propertyKey, value, operator } = token;

        if (!propertyKey) {
          // Free text search across all fields
          const searchValue = value.toLowerCase();
          return (
            instance.name.toLowerCase().includes(searchValue) ||
            instance.template.toLowerCase().includes(searchValue) ||
            instance.state.toLowerCase().includes(searchValue) ||
            (instance.public_ip && instance.public_ip.toLowerCase().includes(searchValue))
          );
        }

        // Property-specific filtering
        const instanceValue = instance[propertyKey as keyof Instance];
        if (!instanceValue) return false;

        const stringValue = String(instanceValue).toLowerCase();
        const filterValue = value.toLowerCase();

        switch (operator) {
          case '=':
            return stringValue === filterValue;
          case '!=':
            return stringValue !== filterValue;
          case ':':
            return stringValue.includes(filterValue);
          case '!:':
            return !stringValue.includes(filterValue);
          default:
            return stringValue.includes(filterValue);
        }
      });
    });
  };

  // Storage Management View
  // Backup Management View
  // Settings View
  const SettingsView = () => {
    // Settings side navigation items
    const settingsNavItems: Array<{ type: string; text?: string; href?: string; items?: Array<{ type: string; text?: string; href?: string }> }> = [
      { type: "link", text: "General", href: "#general" },
      { type: "link", text: "Profiles", href: "#profiles" },
      { type: "link", text: "Users", href: "#users" },
      { type: "divider" },
      {
        type: "expandable-link-group",
        text: "Advanced",
        href: "#advanced",
        items: [
          { type: "link", text: "AMI Management", href: "#ami" },
          { type: "link", text: "Rightsizing", href: "#rightsizing" },
          { type: "link", text: "Policy Framework", href: "#policy" },
          { type: "link", text: "Template Marketplace", href: "#marketplace" },
          { type: "link", text: "Idle Detection", href: "#idle" },
          { type: "link", text: "Logs Viewer", href: "#logs" }
        ]
      }
    ];

    // Render content based on current section
    const renderSettingsContent = () => {
      switch (state.settingsSection) {
        case 'general':
          return (
            <SpaceBetween size="l">
              <Header
                variant="h1"
                description="Configure Prism preferences and system settings"
                actions={
                  <SpaceBetween direction="horizontal" size="xs">
                    <Button onClick={loadApplicationData} disabled={state.loading}>
                      {state.loading ? <Spinner /> : 'Refresh'}
                    </Button>
                    <Button variant="primary">
                      Save Settings
                    </Button>
                  </SpaceBetween>
                }
              >
                General Settings
              </Header>

      {/* System Status Section */}
      <Container
        header={
          <Header
            variant="h2"
            description="System status and daemon configuration"
          >
            System Status
          </Header>
        }
      >
        <ColumnLayout columns={3} variant="text-grid">
          <SpaceBetween size="m">
            <Box variant="awsui-key-label">Daemon Status</Box>
            <StatusIndicator
              type={state.connected ? 'success' : 'error'}
              iconAriaLabel={getStatusLabel('connection', state.connected ? 'success' : 'error')}
            >
              {state.connected ? 'Connected' : 'Disconnected'}
            </StatusIndicator>
            <Box color="text-body-secondary">
              Prism daemon on port 8947
            </Box>
          </SpaceBetween>
          <SpaceBetween size="m">
            <Box variant="awsui-key-label">API Version</Box>
            <Box fontSize="heading-m">v0.5.1</Box>
            <Box color="text-body-secondary">
              Current Prism version
            </Box>
          </SpaceBetween>
          <SpaceBetween size="m">
            <Box variant="awsui-key-label">Active Resources</Box>
            <Box fontSize="heading-m">{state.instances.length + state.efsVolumes.length + state.ebsVolumes.length}</Box>
            <Box color="text-body-secondary">
              Workspaces, EFS and EBS volumes
            </Box>
          </SpaceBetween>
        </ColumnLayout>
      </Container>

      {/* Update Information Section */}
      <Container
        header={
          <Header
            variant="h2"
            description="Check for and install Prism updates"
          >
            Update Information
          </Header>
        }
      >
        {state.updateInfo ? (
          <SpaceBetween size="m">
            <ColumnLayout columns={3} variant="text-grid">
              <SpaceBetween size="m">
                <Box variant="awsui-key-label">Current Version</Box>
                <Box fontSize="heading-m">{state.updateInfo.current_version}</Box>
              </SpaceBetween>
              <SpaceBetween size="m">
                <Box variant="awsui-key-label">Latest Version</Box>
                <Box fontSize="heading-m">
                  {state.updateInfo.latest_version}
                  {state.updateInfo.is_update_available && (
                    <span style={{ marginLeft: '8px' }}><Badge color="green">New</Badge></span>
                  )}
                </Box>
              </SpaceBetween>
              <SpaceBetween size="m">
                <Box variant="awsui-key-label">Status</Box>
                <StatusIndicator
                  type={state.updateInfo.is_update_available ? 'info' : 'success'}
                >
                  {state.updateInfo.is_update_available ? 'Update Available' : 'Up to Date'}
                </StatusIndicator>
              </SpaceBetween>
            </ColumnLayout>

            {state.updateInfo.is_update_available && (
              <Alert type="info" header="New version available">
                <SpaceBetween size="s">
                  <div><strong>Installation method:</strong> {state.updateInfo.install_method}</div>
                  <div><strong>Update command:</strong> <code>{state.updateInfo.update_command}</code></div>
                  <div><strong>Published:</strong> {new Date(state.updateInfo.published_at).toLocaleDateString()}</div>
                  <div>
                    <a href={state.updateInfo.release_url} target="_blank" rel="noopener noreferrer">
                      View release notes →
                    </a>
                  </div>
                </SpaceBetween>
              </Alert>
            )}
          </SpaceBetween>
        ) : (
          <Alert type="info">Checking for updates...</Alert>
        )}
      </Container>

      {/* Configuration Section */}
      <Container
        header={
          <Header
            variant="h2"
            description="Prism configuration and preferences"
          >
            Configuration
          </Header>
        }
      >
        <SpaceBetween size="l">
          <FormField
            label="Auto-refresh interval"
            description="How often the GUI should refresh data from the daemon"
          >
            <Select
              selectedOption={{ label: "30 seconds", value: "30" }}
              onChange={() => {}}
              options={[
                { label: "15 seconds", value: "15" },
                { label: "30 seconds", value: "30" },
                { label: "1 minute", value: "60" },
                { label: "2 minutes", value: "120" }
              ]}
            />
          </FormField>

          <FormField
            label="Default workspace size"
            description="Default size for new workspaces when launching templates"
          >
            <Select
              selectedOption={{ label: "Medium (M)", value: "M" }}
              onChange={() => {}}
              options={[
                { label: "Small (S)", value: "S" },
                { label: "Medium (M)", value: "M" },
                { label: "Large (L)", value: "L" },
                { label: "Extra Large (XL)", value: "XL" }
              ]}
            />
          </FormField>

          <FormField
            label="Show advanced features"
            description="Display advanced management options like hibernation policies and cost tracking"
          >
            <Select
              selectedOption={{ label: "Enabled", value: "enabled" }}
              onChange={() => {}}
              options={[
                { label: "Enabled", value: "enabled" },
                { label: "Disabled", value: "disabled" }
              ]}
            />
          </FormField>

          <FormField
            label="Start at login"
            description="Automatically start Prism GUI when you log in to your computer"
          >
            <Toggle
              checked={state.autoStartEnabled || false}
              onChange={async ({ detail }) => {
                try {
                  await api.setAutoStart(detail.checked);
                  setState(prev => ({ ...prev, autoStartEnabled: detail.checked }));

                  setNotification({
                    type: 'success',
                    content: `Auto-start ${detail.checked ? 'enabled' : 'disabled'} successfully`,
                    dismissible: true,
                    id: 'auto-start-update'
                  });
                } catch (error) {
                  logger.error('Failed to update auto-start:', error);
                  setNotification({
                    type: 'error',
                    content: `Failed to update auto-start: ${error instanceof Error ? error.message : 'Unknown error'}`,
                    dismissible: true,
                    id: 'auto-start-error'
                  });
                }
              }}
            >
              {state.autoStartEnabled ? 'Enabled' : 'Disabled'}
            </Toggle>
          </FormField>
        </SpaceBetween>
      </Container>

      {/* Profile and Authentication Section */}
      <Container
        header={
          <Header
            variant="h2"
            description="AWS profile and authentication settings"
          >
            AWS Configuration
          </Header>
        }
      >
        <ColumnLayout columns={2}>
          <SpaceBetween size="m">
            <FormField
              label="AWS Profile"
              description="Current AWS profile for authentication"
            >
              <Input
                value="aws"
                readOnly
                placeholder="AWS profile name"
              />
            </FormField>
            <FormField
              label="AWS Region"
              description="Current AWS region for resources"
            >
              <Input
                value="us-west-2"
                readOnly
                placeholder="AWS region"
              />
            </FormField>
          </SpaceBetween>
          <SpaceBetween size="m">
            <Box variant="strong">Authentication Status</Box>
            <StatusIndicator
              type="success"
              iconAriaLabel={getStatusLabel('auth', 'authenticated')}
            >
              Authenticated via AWS profile
            </StatusIndicator>
            <Box color="text-body-secondary">
              Using credentials from AWS profile "aws" in region us-west-2.
              Prism automatically manages authentication for all API calls.
            </Box>
          </SpaceBetween>
        </ColumnLayout>
      </Container>

      {/* Feature Management */}
      <Container
        header={
          <Header
            variant="h2"
            description="Enable or disable Prism features"
          >
            Feature Management
          </Header>
        }
      >
        <SpaceBetween size="m">
          {[
            { name: "Workspace Management", status: "enabled", description: "Launch, manage, and connect to cloud workspaces" },
            { name: "Storage Management", status: "enabled", description: "EFS and EBS volume operations" },
            { name: "Project Management", status: "enabled", description: "Multi-user collaboration and budget tracking" },
            { name: "User Management", status: "enabled", description: "Research users with persistent identity" },
            { name: "Hibernation Policies", status: "enabled", description: "Automated cost optimization through hibernation" },
            { name: "Cost Tracking", status: "partial", description: "Budget monitoring and expense analysis" },
            { name: "Template Marketplace", status: "partial", description: "Community template sharing and discovery" },
            { name: "Scaling Predictions", status: "partial", description: "Resource optimization recommendations" }
          ].map((feature) => (
            <Box key={feature.name}>
              <SpaceBetween direction="horizontal" size="s">
                <span style={{ fontWeight: 'bold', minWidth: '200px', display: 'inline-block' }}>{feature.name}:</span>
                <StatusIndicator
                  type={
                    feature.status === 'enabled' ? 'success' :
                    feature.status === 'partial' ? 'warning' : 'error'
                  }
                  iconAriaLabel={getStatusLabel('policy', feature.status, feature.name)}
                >
                  {feature.status}
                </StatusIndicator>
                <Box color="text-body-secondary">{feature.description}</Box>
              </SpaceBetween>
            </Box>
          ))}
        </SpaceBetween>
      </Container>

      {/* Debug and Troubleshooting */}
      <Container
        header={
          <Header
            variant="h2"
            description="Debug information and troubleshooting tools"
          >
            Debug & Troubleshooting
          </Header>
        }
      >
        <SpaceBetween size="m">
          <Alert type="info">
            <Box variant="strong">Debug Mode</Box>
            <Box variant="p">
              Console logging is enabled. Check browser developer tools for detailed API interactions and error messages.
            </Box>
          </Alert>

          <ColumnLayout columns={2}>
            <SpaceBetween size="s">
              <Box variant="strong">Quick Actions</Box>
              <Button>Test API Connection</Button>
              <Button>Refresh All Data</Button>
              <Button>Clear Notifications</Button>
              <Button>Export Configuration</Button>
            </SpaceBetween>
            <SpaceBetween size="s">
              <Box variant="strong">Troubleshooting</Box>
              <Button variant="link" external>View Documentation</Button>
              <Button variant="link" external>GitHub Issues</Button>
              <Button variant="link" external>Troubleshooting Guide</Button>
            </SpaceBetween>
          </ColumnLayout>
        </SpaceBetween>
      </Container>
            </SpaceBetween>
          );

        case 'profiles':
          return <ProfileSelectorViewExtracted />;

        case 'users':
          return (
            <UserManagementViewExtracted
              users={state.users}
              instances={state.instances}
              loading={state.loading}
              onRefresh={loadApplicationData}
              onCreateUser={() => setUserModalVisible(true)}
              onEditUser={(user) => {
                setSelectedUserForEdit(user);
                setEditUserEmail(user.email || '');
                setEditUserDisplayName(user.display_name || (user as any).full_name || '');
                setEditUserRole('');
                setShowEditUserModal(true);
              }}
              onViewUserDetails={async (user) => {
                setSelectedUserForDetails(user);
                setUserDetailsModalVisible(true);
                setLoadingSSHKeys(true);
                try {
                  const response = await api.getUserSSHKeys(user.username);
                  setUserSSHKeys(response.keys || []);
                } catch (_e) {
                  setUserSSHKeys([]);
                } finally {
                  setLoadingSSHKeys(false);
                }
              }}
              onViewUserStatus={async (user) => {
                setSelectedUserForStatus(user);
                setUserStatusModalVisible(true);
                setLoadingUserStatus(true);
                try {
                  const statusData = await api.getUserStatus(user.username);
                  setUserStatusDetails(statusData);
                } catch (_e) {
                  setUserStatusDetails(null);
                } finally {
                  setLoadingUserStatus(false);
                }
              }}
              onProvisionUser={(username) => {
                const user = state.users.find(u => u.username === username) || null;
                setSelectedUserForProvision(user);
                setProvisionModalVisible(true);
              }}
              onManageSSHKeys={(username) => {
                setSelectedUsername(username);
                setSshKeyModalVisible(true);
              }}
              onDeleteUser={(user) => {
                const hasWorkspaces = (user.provisioned_instances?.length || 0) > 0;
                const workspaceWarning = hasWorkspaces
                  ? `This user has ${user.provisioned_instances!.length} provisioned workspace(s). Deleting the user will remove their access to these workspaces.`
                  : undefined;
                setDeleteModalConfig({
                  type: 'user',
                  name: user.username,
                  requireNameConfirmation: false,
                  warning: workspaceWarning,
                  onConfirm: async () => {
                    try {
                      await api.deleteUser(user.username);
                      usersVersionRef.current++;
                      setState(prev => ({
                        ...prev,
                        users: prev.users.filter(u => u.username !== user.username),
                        notifications: [{
                          type: 'success',
                          header: 'User Deleted',
                          content: `User "${user.username}" deleted successfully`,
                          dismissible: true,
                          id: Date.now().toString()
                        }, ...prev.notifications]
                      }));
                      setDeleteModalVisible(false);
                    } catch (error: any) {
                      setState(prev => ({
                        ...prev,
                        notifications: [{
                          type: 'error',
                          header: 'Delete Failed',
                          content: error.message || 'Failed to delete user',
                          dismissible: true,
                          id: Date.now().toString()
                        }, ...prev.notifications]
                      }));
                    }
                  }
                });
                setDeleteModalVisible(true);
              }}
            />
          );

        case 'ami':
          return (
            <AMIManagementViewExtracted
              amis={state.amis}
              amiRegions={state.amiRegions}
              amiBuilds={state.amiBuilds}
              loading={state.loading}
              onRefresh={loadApplicationData}
            />
          );

        case 'rightsizing':
          return <PlaceholderView title="Rightsizing Recommendations" description="Workspace rightsizing recommendations will help optimize your costs by suggesting better-sized workspaces based on actual usage patterns." />;

        case 'policy':
          return <PlaceholderView title="Policy Management" description="Policy management allows you to configure institutional policies, access controls, and governance rules for your Prism deployment." />;

        case 'marketplace':
          return (
            <MarketplaceViewExtracted
              marketplaceTemplates={state.marketplaceTemplates}
              marketplaceCategories={state.marketplaceCategories}
              loading={state.loading}
              onRefresh={loadApplicationData}
            />
          );

        case 'idle':
          return (
            <IdleDetectionViewExtracted
              idlePolicies={state.idlePolicies}
              idleSchedules={state.idleSchedules}
              loading={state.loading}
              onRefresh={loadApplicationData}
            />
          );

        case 'logs':
          return <LogsView instances={state.instances} loading={state.loading} onRefresh={loadApplicationData} />;

        default:
          return (
            <Alert type="error">
              Unknown settings section: {state.settingsSection}
            </Alert>
          );
      }
    };

    // Return the layout with side navigation
    return (
      <div style={{ display: 'flex', height: '100%' }}>
        <div style={{ width: '280px', borderRight: '1px solid #e9ebed', padding: '20px 0' }}>
          <SideNavigation
            activeHref={`#${state.settingsSection}`}
            header={{ text: "Settings", href: "#general" }}
            items={settingsNavItems}
            onFollow={(e: { detail: { href: string; external?: boolean }; preventDefault: () => void }) => {
              e.preventDefault();
              const href = e.detail.href.replace('#', '');
              setState(prev => ({
                ...prev,
                settingsSection: href as typeof prev.settingsSection
              }));
            }}
          />
        </div>
        <div style={{ flex: 1, padding: '20px', overflow: 'auto' }}>
          {renderSettingsContent()}
        </div>
      </div>
    );
  };


  // Launch Modal


  const LaunchModal = () => (
    <Modal
      onDismiss={handleModalDismiss}
      visible={launchModalVisible}
      header={`Launch ${state.selectedTemplate ? getTemplateName(state.selectedTemplate) : 'Research Environment'}`}
      size="medium"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={handleModalDismiss}>
              Cancel
            </Button>
            <Button
              variant="primary"
              disabled={!launchConfig.name.trim()}
              onClick={handleLaunchInstance}
            >
              Launch Workspace
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <Form>
        <SpaceBetween size="m">
          <FormField
            label="Workspace name"
            description="Choose a descriptive name for your research project"
            errorText={!launchConfig.name.trim() ? "Workspace name is required" : ""}
          >
            <Input
              value={launchConfig.name}
              onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, name: detail.value }))}
              placeholder="my-research-project"
            />
          </FormField>

          <FormField label="Workspace size" description="Choose the right size for your workload">
            <Select
              selectedOption={{ label: "Medium (M) - Recommended", value: "M" }}
              onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, size: detail.selectedOption.value || 'M' }))}
              options={[
                { label: "Small (S) - Light workloads", value: "S" },
                { label: "Medium (M) - Recommended", value: "M" },
                { label: "Large (L) - Heavy compute", value: "L" },
                { label: "Extra Large (XL) - Maximum performance", value: "XL" }
              ]}
              data-testid="instance-size-select"
            />
          </FormField>

          {state.selectedTemplate && (
            <Alert type="info">
              <Box>
                <Box variant="strong">Template: {getTemplateName(state.selectedTemplate)}</Box>
                <Box>Description: {getTemplateDescription(state.selectedTemplate)}</Box>
                {state.selectedTemplate.package_manager && (
                  <Box>Package Manager: {state.selectedTemplate.package_manager}</Box>
                )}
                {state.selectedTemplate.complexity && (
                  <Box>Complexity: {state.selectedTemplate.complexity}</Box>
                )}
              </Box>
            </Alert>
          )}

          <FormField
            label="Instance Options"
            description="Configure advanced instance settings"
          >
            <SpaceBetween size="s">
              <Checkbox
                checked={launchConfig.spot || false}
                onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, spot: detail.checked }))}
              >
                Spot instance - use lower-cost spot pricing
              </Checkbox>
              <Checkbox
                checked={launchConfig.hibernation || false}
                onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, hibernation: detail.checked }))}
              >
                Hibernation - enable instance hibernation support
              </Checkbox>
            </SpaceBetween>
          </FormField>

          <FormField
            label="Validation"
            description="Test your configuration without actually launching resources"
          >
            <Checkbox
              checked={launchConfig.dryRun || false}
              onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, dryRun: detail.checked }))}
            >
              Dry run mode - validate without creating resources
            </Checkbox>
          </FormField>
        </SpaceBetween>
      </Form>
    </Modal>
  );

  // Create Backup Modal
  const CreateBackupModal = () => {
    const handleCreateBackup = async () => {
      try {
        setCreateBackupValidationAttempted(true);

        if (!createBackupConfig.instanceId || !createBackupConfig.backupName) {
          setState(prev => ({
            ...prev,
            notifications: [
              ...prev.notifications,
              {
                type: 'error',
                header: 'Validation Error',
                content: 'Instance and backup name are required',
                dismissible: true,
                id: Date.now().toString()
              }
            ]
          }));
          return;
        }

        setState(prev => ({ ...prev, loading: true }));
        setCreateBackupModalVisible(false);

        await api.createSnapshot(
          createBackupConfig.instanceId,
          createBackupConfig.backupName,
          createBackupConfig.description
        );

        setState(prev => ({
          ...prev,
          loading: false,
          notifications: [
            ...prev.notifications,
            {
              type: 'success',
              header: 'Backup Created',
              content: `Backup "${createBackupConfig.backupName}" is being created. This may take 5-10 minutes.`,
              dismissible: true,
              id: Date.now().toString()
            }
          ]
        }));

        // Reload data to show new backup
        await loadApplicationData();
      } catch (error) {
        setState(prev => ({
          ...prev,
          loading: false,
          notifications: [
            ...prev.notifications,
            {
              type: 'error',
              header: 'Backup Creation Failed',
              content: error instanceof Error ? error.message : 'Unknown error occurred',
              dismissible: true,
              id: Date.now().toString()
            }
          ]
        }));
      }
    };

    const handleDismiss = () => {
      setCreateBackupModalVisible(false);
      setCreateBackupValidationAttempted(false);
      setCreateBackupConfig({
        instanceId: '',
        backupName: '',
        backupType: 'full',
        description: ''
      });
    };

    return (
      <Modal
        onDismiss={handleDismiss}
        visible={createBackupModalVisible}
        header="Create Backup"
        size="medium"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={handleDismiss}>
                Cancel
              </Button>
              <Button
                variant="primary"
                disabled={!createBackupConfig.instanceId || !createBackupConfig.backupName.trim()}
                onClick={handleCreateBackup}
                data-testid="create-backup-submit"
              >
                Create
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <Form>
          <SpaceBetween size="m">
            <FormField
              label="Instance"
              description="Select the workspace instance to backup"
              errorText={createBackupValidationAttempted && !createBackupConfig.instanceId ? "Instance is required" : ""}
            >
              <Select
                data-testid="instance-select"
                selectedOption={
                  createBackupConfig.instanceId
                    ? {
                        label: state.instances.find(i => i.name === createBackupConfig.instanceId)?.name || createBackupConfig.instanceId,
                        value: createBackupConfig.instanceId
                      }
                    : null
                }
                onChange={({ detail }) =>
                  setCreateBackupConfig(prev => ({ ...prev, instanceId: detail.selectedOption.value || '' }))
                }
                options={state.instances.map(instance => ({
                  label: `${instance.name} (${instance.template || 'Unknown template'})`,
                  value: instance.name
                }))}
                placeholder="Select an instance..."
                empty="No instances available"
                ariaLabel="Instance"
              />
            </FormField>

            <FormField
              label="Backup name"
              description="Choose a descriptive name for this backup"
              errorText={createBackupValidationAttempted && !createBackupConfig.backupName.trim() ? "Backup name is required" : ""}
            >
              <Input
                value={createBackupConfig.backupName}
                onChange={({ detail }) =>
                  setCreateBackupConfig(prev => ({ ...prev, backupName: detail.value }))
                }
                placeholder="my-backup-2024-11-16"
              />
            </FormField>

            <FormField
              label="Backup type"
              description="Full backups capture the entire instance state"
            >
              <Select
                selectedOption={{ label: "Full backup", value: "full" }}
                onChange={({ detail }) =>
                  setCreateBackupConfig(prev => ({ ...prev, backupType: detail.selectedOption.value || 'full' }))
                }
                options={[
                  { label: "Full backup", value: "full" },
                  { label: "Incremental backup", value: "incremental" }
                ]}
              />
            </FormField>

            <FormField
              label="Description (optional)"
              description="Add notes about this backup"
            >
              <Input
                value={createBackupConfig.description}
                onChange={({ detail }) =>
                  setCreateBackupConfig(prev => ({ ...prev, description: detail.value }))
                }
                placeholder="Weekly backup before major update"
              />
            </FormField>

            {createBackupConfig.instanceId && (
              <Alert type="info">
                <SpaceBetween size="s">
                  <Box variant="strong">Backup Information</Box>
                  <Box>
                    • <strong>Creation time:</strong> 5-10 minutes depending on instance size
                  </Box>
                  <Box>
                    • <strong>Cost:</strong> ~$0.05/GB/month for snapshot storage
                  </Box>
                  <Box>
                    • <strong>Instance continues running:</strong> No downtime during backup creation
                  </Box>
                </SpaceBetween>
              </Alert>
            )}

            {createBackupValidationAttempted && (!createBackupConfig.instanceId || !createBackupConfig.backupName.trim()) && (
              <div data-testid="validation-error">
                {!createBackupConfig.instanceId ? "Instance is required" : !createBackupConfig.backupName.trim() ? "Backup name is required" : ""}
              </div>
            )}
          </SpaceBetween>
        </Form>
      </Modal>
    );
  };

  // Delete Backup Confirmation Modal
  const DeleteBackupModal = () => {
    if (!selectedBackupForDelete) return null;

    const handleDeleteBackup = async () => {
      try {
        setState(prev => ({ ...prev, loading: true }));
        setDeleteBackupModalVisible(false);

        await api.deleteSnapshot(selectedBackupForDelete.snapshot_name);

        setState(prev => ({
          ...prev,
          loading: false,
          notifications: [
            ...prev.notifications,
            {
              type: 'success',
              header: 'Backup Deleted',
              content: `Backup "${selectedBackupForDelete.snapshot_name}" has been deleted. You will save $${selectedBackupForDelete.storage_cost_monthly.toFixed(2)}/month.`,
              dismissible: true,
              id: Date.now().toString()
            }
          ]
        }));

        // Reload data to update backup list
        await loadApplicationData();
      } catch (error) {
        setState(prev => ({
          ...prev,
          loading: false,
          notifications: [
            ...prev.notifications,
            {
              type: 'error',
              header: 'Delete Failed',
              content: error instanceof Error ? error.message : 'Unknown error occurred',
              dismissible: true,
              id: Date.now().toString()
            }
          ]
        }));
      }
    };

    const handleDismiss = () => {
      setDeleteBackupModalVisible(false);
      setSelectedBackupForDelete(null);
    };

    const sizeGB = selectedBackupForDelete.size_gb || Math.ceil(selectedBackupForDelete.storage_cost_monthly / 0.05);
    const monthlySavings = selectedBackupForDelete.storage_cost_monthly;

    return (
      <Modal
        onDismiss={handleDismiss}
        visible={deleteBackupModalVisible}
        header="Delete Backup Confirmation"
        size="medium"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={handleDismiss}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={handleDeleteBackup}
              >
                Delete
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          <Alert type="warning">
            <Box variant="strong">This action cannot be undone</Box>
          </Alert>

          <Box>
            Are you sure you want to delete backup <strong>"{selectedBackupForDelete.snapshot_name}"</strong>?
          </Box>

          <Container>
            <SpaceBetween size="s">
              <Box variant="h4">💰 Cost Savings</Box>
              <ColumnLayout columns={2} variant="text-grid">
                <SpaceBetween size="xs">
                  <Box variant="awsui-key-label">Storage Size</Box>
                  <Box>{sizeGB} GB</Box>
                </SpaceBetween>
                <SpaceBetween size="xs">
                  <Box variant="awsui-key-label">Monthly Savings</Box>
                  <Box color="text-status-success" fontSize="heading-m">
                    <strong>${monthlySavings.toFixed(2)}/month</strong>
                  </Box>
                </SpaceBetween>
                <SpaceBetween size="xs">
                  <Box variant="awsui-key-label">Annual Savings</Box>
                  <Box>${(monthlySavings * 12).toFixed(2)}/year</Box>
                </SpaceBetween>
                <SpaceBetween size="xs">
                  <Box variant="awsui-key-label">Free Storage</Box>
                  <Box>{sizeGB} GB freed</Box>
                </SpaceBetween>
              </ColumnLayout>
            </SpaceBetween>
          </Container>

          <Box variant="small" color="text-body-secondary">
            <strong>Backup Details:</strong><br/>
            • Source: {selectedBackupForDelete.source_instance}<br/>
            • Template: {selectedBackupForDelete.source_template}<br/>
            • Created: {new Date(selectedBackupForDelete.created_at).toLocaleString()}
          </Box>
        </SpaceBetween>
      </Modal>
    );
  };

  // Restore Backup Modal
  const RestoreBackupModal = () => {
    if (!selectedBackupForRestore) return null;

    const handleRestoreBackup = async () => {
      try {
        if (!restoreInstanceName.trim()) {
          setState(prev => ({
            ...prev,
            notifications: [
              ...prev.notifications,
              {
                type: 'error',
                header: 'Validation Error',
                content: 'New instance name is required',
                dismissible: true,
                id: Date.now().toString()
              }
            ]
          }));
          return;
        }

        setState(prev => ({ ...prev, loading: true }));
        setRestoreBackupModalVisible(false);

        await api.restoreSnapshot(selectedBackupForRestore.snapshot_name, restoreInstanceName);

        setState(prev => ({
          ...prev,
          loading: false,
          notifications: [
            ...prev.notifications,
            {
              type: 'success',
              header: 'Restoring Backup',
              content: `Backup "${selectedBackupForRestore.snapshot_name}" is being restored to instance "${restoreInstanceName}". This may take 10-15 minutes.`,
              dismissible: true,
              id: Date.now().toString()
            }
          ]
        }));

        // Reload data to show new instance
        await loadApplicationData();
      } catch (error) {
        setState(prev => ({
          ...prev,
          loading: false,
          notifications: [
            ...prev.notifications,
            {
              type: 'error',
              header: 'Restore Failed',
              content: error instanceof Error ? error.message : 'Unknown error occurred',
              dismissible: true,
              id: Date.now().toString()
            }
          ]
        }));
      }
    };

    const handleDismiss = () => {
      setRestoreBackupModalVisible(false);
      setSelectedBackupForRestore(null);
      setRestoreInstanceName('');
    };

    const sizeGB = selectedBackupForRestore.size_gb || Math.ceil(selectedBackupForRestore.storage_cost_monthly / 0.05);

    return (
      <Modal
        onDismiss={handleDismiss}
        visible={restoreBackupModalVisible}
        header="Restore Backup"
        size="medium"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={handleDismiss}>
                Cancel
              </Button>
              <Button
                variant="primary"
                disabled={!restoreInstanceName.trim()}
                onClick={handleRestoreBackup}
              >
                Restore
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          <Alert type="info">
            <Box variant="strong">⏱️ Restore Time</Box>
            <Box>
              Restoring this backup may take 10-15 minutes depending on the backup size ({sizeGB} GB).
              The new instance will be created with all data and configuration from the backup.
            </Box>
          </Alert>

          <Box>
            Restore backup <strong>"{selectedBackupForRestore.snapshot_name}"</strong> to a new instance.
          </Box>

          <FormField
            label="New instance name"
            description="Choose a name for the restored instance"
            errorText={!restoreInstanceName.trim() ? "Instance name is required" : ""}
          >
            <Input
              value={restoreInstanceName}
              onChange={({ detail }) => setRestoreInstanceName(detail.value)}
              placeholder="restored-instance"
            />
          </FormField>

          <Container>
            <SpaceBetween size="s">
              <Box variant="h4">📋 Backup Details</Box>
              <ColumnLayout columns={2} variant="text-grid">
                <SpaceBetween size="xs">
                  <Box variant="awsui-key-label">Source Instance</Box>
                  <Box>{selectedBackupForRestore.source_instance}</Box>
                </SpaceBetween>
                <SpaceBetween size="xs">
                  <Box variant="awsui-key-label">Template</Box>
                  <Box>{selectedBackupForRestore.source_template}</Box>
                </SpaceBetween>
                <SpaceBetween size="xs">
                  <Box variant="awsui-key-label">Backup Size</Box>
                  <Box>{sizeGB} GB</Box>
                </SpaceBetween>
                <SpaceBetween size="xs">
                  <Box variant="awsui-key-label">Created</Box>
                  <Box>{new Date(selectedBackupForRestore.created_at).toLocaleDateString()}</Box>
                </SpaceBetween>
              </ColumnLayout>
            </SpaceBetween>
          </Container>

          <Alert type="warning">
            <Box variant="strong">What happens during restore:</Box>
            <ul>
              <li>A new EC2 instance will be launched from this backup</li>
              <li>All files, configurations, and installed software will be restored</li>
              <li>The new instance will have the same specifications as the original</li>
              <li>You can modify the instance after restoration completes</li>
            </ul>
          </Alert>
        </SpaceBetween>
      </Modal>
    );
  };

  // Onboarding Wizard Modal
  const OnboardingWizard = () => {
    const totalSteps = 3;

    const handleNext = () => {
      if (onboardingStep < totalSteps - 1) {
        setOnboardingStep(onboardingStep + 1);
      } else {
        // Complete onboarding
        localStorage.setItem('prism_onboarding_complete', 'true');
        setOnboardingComplete(true);
        setOnboardingVisible(false);
        setOnboardingStep(0);
      }
    };

    const handleBack = () => {
      if (onboardingStep > 0) {
        setOnboardingStep(onboardingStep - 1);
      }
    };

    const handleSkip = () => {
      localStorage.setItem('prism_onboarding_complete', 'true');
      setOnboardingComplete(true);
      setOnboardingVisible(false);
      setOnboardingStep(0);
    };

    return (
      <Modal
        visible={onboardingVisible}
        onDismiss={handleSkip}
        header={`Welcome to Prism - Step ${onboardingStep + 1} of ${totalSteps}`}
        size="large"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              {onboardingStep > 0 && (
                <Button onClick={handleBack}>
                  Back
                </Button>
              )}
              <Button variant="link" onClick={handleSkip}>
                Skip Tour
              </Button>
              <Button variant="primary" onClick={handleNext}>
                {onboardingStep < totalSteps - 1 ? 'Next' : 'Get Started'}
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="l">
          {/* Step 1: AWS Profile Setup */}
          {onboardingStep === 0 && (
            <SpaceBetween size="m">
              <Alert type="info" header="AWS Credentials Configured">
                Prism is already connected to your AWS account using the configured profile.
              </Alert>
              <Box variant="h2">Step 1: AWS Configuration</Box>
              <Box>
                Prism manages cloud workstations in your AWS account. Your current AWS configuration:
              </Box>
              <Container>
                <ColumnLayout columns={2} variant="text-grid">
                  <div>
                    <Box variant="awsui-key-label">AWS Profile</Box>
                    <Box fontWeight="bold">aws</Box>
                    <Box variant="small" color="text-body-secondary">
                      Your AWS credentials profile
                    </Box>
                  </div>
                  <div>
                    <Box variant="awsui-key-label">Region</Box>
                    <Box fontWeight="bold">us-west-2</Box>
                    <Box variant="small" color="text-body-secondary">
                      Resources will be created here
                    </Box>
                  </div>
                </ColumnLayout>
              </Container>
              <Box variant="p" color="text-body-secondary">
                Prism uses your AWS credentials to create and manage cloud workstations.
                You maintain full control over your resources and costs.
              </Box>
            </SpaceBetween>
          )}

          {/* Step 2: Template Discovery Tour */}
          {onboardingStep === 1 && (
            <SpaceBetween size="m">
              <Box variant="h2">Step 2: Choose Your Research Environment</Box>
              <Box>
                Prism provides pre-configured templates for different research workflows.
                Each template includes specialized software, libraries, and tools.
              </Box>
              <ColumnLayout columns={2}>
                <Container header={<Header variant="h3">Popular Templates</Header>}>
                  <SpaceBetween size="s">
                    <Box>
                      <Box variant="strong">Python Machine Learning</Box>
                      <Box variant="small" color="text-body-secondary">
                        Python 3, Jupyter, TensorFlow, PyTorch, scikit-learn
                      </Box>
                    </Box>
                    <Box>
                      <Box variant="strong">R Research Environment</Box>
                      <Box variant="small" color="text-body-secondary">
                        R, RStudio Server, tidyverse, statistical packages
                      </Box>
                    </Box>
                    <Box>
                      <Box variant="strong">Collaborative Workspace</Box>
                      <Box variant="small" color="text-body-secondary">
                        Multi-language support with Python, R, Julia
                      </Box>
                    </Box>
                  </SpaceBetween>
                </Container>
                <Container header={<Header variant="h3">What's Included</Header>}>
                  <SpaceBetween size="s">
                    <Box>✓ Pre-installed software and dependencies</Box>
                    <Box>✓ Optimized workspace sizing for your workload</Box>
                    <Box>✓ Persistent storage for your data</Box>
                    <Box>✓ SSH and remote access configured</Box>
                    <Box>✓ Security best practices applied</Box>
                  </SpaceBetween>
                </Container>
              </ColumnLayout>
              <Alert type="info">
                You can browse all available templates in the <strong>Templates</strong> section after completing this tour.
              </Alert>
            </SpaceBetween>
          )}

          {/* Step 3: Launch Your First Workspace */}
          {onboardingStep === 2 && (
            <SpaceBetween size="m">
              <Box variant="h2">Step 3: Launch Your First Workstation</Box>
              <Box>
                Ready to get started? Here's how to launch your first cloud workstation:
              </Box>
              <Container>
                <SpaceBetween size="m">
                  <div>
                    <Box variant="h4">1. Select a Template</Box>
                    <Box>Choose a template that matches your research needs from the Templates page.</Box>
                  </div>
                  <div>
                    <Box variant="h4">2. Configure Workspace</Box>
                    <Box>Give your workstation a name and select the appropriate size (Small, Medium, Large).</Box>
                  </div>
                  <div>
                    <Box variant="h4">3. Launch & Connect</Box>
                    <Box>Prism creates your workspace in minutes. Connect via SSH or web interface when ready.</Box>
                  </div>
                </SpaceBetween>
              </Container>
              <Alert type="success" header="You're All Set!">
                After clicking "Get Started", explore the dashboard to see your system status,
                browse templates, and launch your first cloud workstation.
              </Alert>
              <Box variant="p" color="text-body-secondary">
                💡 <strong>Tip:</strong> Start with a Medium (M) sized workspace for most workloads.
                You can always stop, resize, or terminate workspaces to manage costs.
              </Box>
            </SpaceBetween>
          )}
        </SpaceBetween>
      </Modal>
    );
  };

  // Quick Start Wizard
  const QuickStartWizard = () => {
    const handleWizardNavigate = (event: { detail: { requestedStepIndex: number; reason: string } }) => {
      setQuickStartActiveStepIndex(event.detail.requestedStepIndex);
    };

    const handleWizardCancel = () => {
      setQuickStartWizardVisible(false);
      setQuickStartActiveStepIndex(0);
      setQuickStartConfig({
        selectedTemplate: null,
        workspaceName: '',
        size: 'M',
        launchInProgress: false,
        launchedWorkspaceId: null
      });
    };

    const handleWizardSubmit = async () => {
      if (!quickStartConfig.selectedTemplate) return;

      setQuickStartConfig(prev => ({ ...prev, launchInProgress: true }));
      setQuickStartActiveStepIndex(3); // Move to progress step

      try {
        const result = await api.launchInstance(
          getTemplateSlug(quickStartConfig.selectedTemplate),
          quickStartConfig.workspaceName,
          quickStartConfig.size
        );

        setQuickStartConfig(prev => ({
          ...prev,
          launchInProgress: false,
          launchedWorkspaceId: result?.id || null
        }));

        addNotification({
          type: 'success',
          content: `Workspace "${quickStartConfig.workspaceName}" launched successfully!`,
          dismissible: true
        });

        // Refresh workspace list
        await loadApplicationData();
      } catch (error) {
        setQuickStartConfig(prev => ({ ...prev, launchInProgress: false }));
        addNotification({
          type: 'error',
          content: `Failed to launch workspace: ${error instanceof Error ? error.message : 'Unknown error'}`,
          dismissible: true
        });
      }
    };

    const getSizeDescription = (size: string): string => {
      const descriptions: Record<string, string> = {
        'S': 'Small - 2 vCPU, 4GB RAM (~$0.08/hour)',
        'M': 'Medium - 4 vCPU, 8GB RAM (~$0.16/hour)',
        'L': 'Large - 8 vCPU, 16GB RAM (~$0.32/hour)',
        'XL': 'Extra Large - 16 vCPU, 32GB RAM (~$0.64/hour)'
      };
      return descriptions[size] || descriptions['M'];
    };

    const getCategoryTemplates = (category: string): Template[] => {
      return Object.values(state.templates).filter(t => {
        const name = getTemplateName(t).toLowerCase();
        const desc = getTemplateDescription(t).toLowerCase();
        switch (category) {
          case 'ml':
            return name.includes('machine learning') || name.includes('ml') || name.includes('python') && desc.includes('tensorflow');
          case 'datascience':
            return name.includes('python') || name.includes('jupyter') || name.includes('data');
          case 'r':
            return name.includes('r ') || name.includes('rstudio');
          case 'bio':
            return name.includes('bio') || name.includes('genomics');
          default:
            return true;
        }
      });
    };

    return (
      <Modal
        visible={quickStartWizardVisible}
        onDismiss={handleWizardCancel}
        size="large"
        header="Quick Start - Launch Workspace"
      >
        <Wizard
          i18nStrings={{
            stepNumberLabel: stepNumber => `Step ${stepNumber}`,
            collapsedStepsLabel: (stepNumber, stepsCount) => `Step ${stepNumber} of ${stepsCount}`,
            skipToButtonLabel: (step) => `Skip to ${step.title}`,
            navigationAriaLabel: "Steps",
            cancelButton: "Cancel",
            previousButton: "Previous",
            nextButton: "Next",
            submitButton: "Launch Workspace",
            optional: "optional"
          }}
          onNavigate={handleWizardNavigate}
          onCancel={handleWizardCancel}
          onSubmit={handleWizardSubmit}
          activeStepIndex={quickStartActiveStepIndex}
          isLoadingNextStep={quickStartConfig.launchInProgress}
          steps={[
            {
              title: "Select Template",
              description: "Choose a pre-configured research environment",
              content: (
                <SpaceBetween size="l">
                  <Alert type="info">
                    Select a template that matches your research needs. Each template includes specialized software and tools.
                  </Alert>

                  <Tabs
                    tabs={[
                      {
                        id: "all",
                        label: "All Templates",
                        content: (
                          <Cards
                            cardDefinition={{
                              header: item => (
                                <Box variant="h3">{getTemplateName(item)}</Box>
                              ),
                              sections: [
                                {
                                  id: "description",
                                  content: item => getTemplateDescription(item)
                                },
                                {
                                  id: "tags",
                                  content: item => (
                                    <SpaceBetween direction="horizontal" size="xs">
                                      {getTemplateTags(item).slice(0, 3).map((tag, idx) => (
                                        <Badge key={idx} color="blue">{tag}</Badge>
                                      ))}
                                    </SpaceBetween>
                                  )
                                }
                              ]
                            }}
                            items={Object.values(state.templates)}
                            selectionType="single"
                            selectedItems={quickStartConfig.selectedTemplate ? [quickStartConfig.selectedTemplate] : []}
                            onSelectionChange={({ detail }) => {
                              if (detail.selectedItems.length > 0) {
                                setQuickStartConfig(prev => ({
                                  ...prev,
                                  selectedTemplate: detail.selectedItems[0]
                                }));
                              }
                            }}
                            cardsPerRow={[{ cards: 1 }, { minWidth: 500, cards: 2 }]}
                            empty={
                              <Box textAlign="center" color="inherit">
                                <b>No templates available</b>
                                <Box padding={{ bottom: "s" }} variant="p" color="inherit">
                                  No research templates found.
                                </Box>
                              </Box>
                            }
                          />
                        )
                      },
                      {
                        id: "ml",
                        label: "ML/AI",
                        content: (
                          <Cards
                            cardDefinition={{
                              header: item => <Box variant="h3">{getTemplateName(item)}</Box>,
                              sections: [{ id: "description", content: item => getTemplateDescription(item) }]
                            }}
                            items={getCategoryTemplates('ml')}
                            selectionType="single"
                            selectedItems={quickStartConfig.selectedTemplate ? [quickStartConfig.selectedTemplate] : []}
                            onSelectionChange={({ detail }) => {
                              if (detail.selectedItems.length > 0) {
                                setQuickStartConfig(prev => ({ ...prev, selectedTemplate: detail.selectedItems[0] }));
                              }
                            }}
                            cardsPerRow={[{ cards: 1 }, { minWidth: 500, cards: 2 }]}
                          />
                        )
                      },
                      {
                        id: "datascience",
                        label: "Data Science",
                        content: (
                          <Cards
                            cardDefinition={{
                              header: item => <Box variant="h3">{getTemplateName(item)}</Box>,
                              sections: [{ id: "description", content: item => getTemplateDescription(item) }]
                            }}
                            items={getCategoryTemplates('datascience')}
                            selectionType="single"
                            selectedItems={quickStartConfig.selectedTemplate ? [quickStartConfig.selectedTemplate] : []}
                            onSelectionChange={({ detail }) => {
                              if (detail.selectedItems.length > 0) {
                                setQuickStartConfig(prev => ({ ...prev, selectedTemplate: detail.selectedItems[0] }));
                              }
                            }}
                            cardsPerRow={[{ cards: 1 }, { minWidth: 500, cards: 2 }]}
                          />
                        )
                      }
                    ]}
                  />
                </SpaceBetween>
              ),
              isOptional: false
            },
            {
              title: "Configure Workspace",
              description: "Set workspace name and size",
              content: (
                <SpaceBetween size="l">
                  <FormField
                    label="Workspace Name"
                    description="Choose a unique name for your workspace"
                    constraintText="Use lowercase letters, numbers, and hyphens only"
                  >
                    <Input
                      value={quickStartConfig.workspaceName}
                      onChange={({ detail }) => setQuickStartConfig(prev => ({ ...prev, workspaceName: detail.value }))}
                      placeholder="my-research-workspace"
                    />
                  </FormField>

                  <FormField
                    label="Workspace Size"
                    description="Choose the compute resources for your workspace"
                  >
                    <Select
                      selectedOption={{ label: getSizeDescription(quickStartConfig.size), value: quickStartConfig.size }}
                      onChange={({ detail }) => setQuickStartConfig(prev => ({ ...prev, size: detail.selectedOption.value || 'M' }))}
                      options={[
                        { label: getSizeDescription('S'), value: 'S' },
                        { label: getSizeDescription('M'), value: 'M' },
                        { label: getSizeDescription('L'), value: 'L' },
                        { label: getSizeDescription('XL'), value: 'XL' }
                      ]}
                    />
                  </FormField>

                  <Alert type="info">
                    💡 <strong>Tip:</strong> Start with Medium size for most workloads. You can always stop and resize later.
                  </Alert>
                </SpaceBetween>
              ),
              isOptional: false
            },
            {
              title: "Review & Launch",
              description: "Review your configuration",
              content: (
                <SpaceBetween size="l">
                  <Container header={<Header variant="h3">Configuration Summary</Header>}>
                    <ColumnLayout columns={2} variant="text-grid">
                      <div>
                        <Box variant="awsui-key-label">Template</Box>
                        <Box>{quickStartConfig.selectedTemplate ? getTemplateName(quickStartConfig.selectedTemplate) : 'None'}</Box>
                      </div>
                      <div>
                        <Box variant="awsui-key-label">Workspace Name</Box>
                        <Box>{quickStartConfig.workspaceName || 'Not set'}</Box>
                      </div>
                      <div>
                        <Box variant="awsui-key-label">Size</Box>
                        <Box>{getSizeDescription(quickStartConfig.size)}</Box>
                      </div>
                      <div>
                        <Box variant="awsui-key-label">Estimated Cost</Box>
                        <Box data-testid="cost-estimate">
                          {quickStartConfig.size === 'S' && '~$0.08/hour (~$58/month)'}
                          {quickStartConfig.size === 'M' && '~$0.16/hour (~$115/month)'}
                          {quickStartConfig.size === 'L' && '~$0.32/hour (~$230/month)'}
                          {quickStartConfig.size === 'XL' && '~$0.64/hour (~$460/month)'}
                        </Box>
                      </div>
                    </ColumnLayout>
                  </Container>

                  <Alert type="warning">
                    <strong>Cost Reminder:</strong> Remember to stop or hibernate your workspace when not in use to save costs.
                  </Alert>

                  {quickStartConfig.selectedTemplate && quickStartConfig.workspaceName && (
                    <Alert type="success">
                      ✅ Ready to launch! Click "Launch Workspace" to proceed.
                    </Alert>
                  )}
                </SpaceBetween>
              ),
              isOptional: false
            },
            {
              title: "Launch Progress",
              description: "Launching your workspace",
              content: (
                <SpaceBetween size="l">
                  {quickStartConfig.launchInProgress && (
                    <Box>
                      <ProgressBar value={50} description="Launching workspace..." />
                      <Box margin={{ top: "m" }} color="text-body-secondary">
                        This typically takes 2-3 minutes. Your workspace is being provisioned with all required software and configurations.
                      </Box>
                    </Box>
                  )}

                  {!quickStartConfig.launchInProgress && quickStartConfig.launchedWorkspaceId && (
                    <Alert type="success" header="Workspace Launched Successfully!">
                      <SpaceBetween size="m">
                        <Box>
                          Your workspace <strong>{quickStartConfig.workspaceName}</strong> is now running and ready to use.
                        </Box>
                        <Box>
                          <strong>Next Steps:</strong>
                          <ul>
                            <li>Connect via SSH or web interface from the Workspaces page</li>
                            <li>Access pre-installed software and tools</li>
                            <li>Remember to stop or hibernate when done to save costs</li>
                          </ul>
                        </Box>
                        <SpaceBetween direction="horizontal" size="s">
                          <Button
                            variant="primary"
                            onClick={() => {
                              setState(prev => ({ ...prev, activeView: 'workspaces' }));
                              handleWizardCancel();
                            }}
                          >
                            View Workspace
                          </Button>
                          <Button onClick={handleWizardCancel}>
                            Close
                          </Button>
                        </SpaceBetween>
                      </SpaceBetween>
                    </Alert>
                  )}

                  {!quickStartConfig.launchInProgress && !quickStartConfig.launchedWorkspaceId && (
                    <Alert type="info">
                      Click "Launch Workspace" to start the deployment process.
                    </Alert>
                  )}
                </SpaceBetween>
              ),
              isOptional: false
            }
          ]}
        />
      </Modal>
    );
  };

  // Create Project Modal
  const CreateProjectModal = () => (
    <Modal
      visible={projectModalVisible}
      onDismiss={() => {
        setProjectModalVisible(false);
        setProjectValidationError('');
      }}
      header="Create New Project"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={() => setProjectModalVisible(false)}>Cancel</Button>
            <Button variant="primary" data-testid="create-project-submit-button" onClick={handleCreateProject}>Create</Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {projectValidationError && (
          <Alert type="error" data-testid="validation-error">
            {projectValidationError}
          </Alert>
        )}

        <FormField label="Project Name" description="Unique identifier for the project">
          <Input
            data-testid="project-name-input"
            value={projectName}
            onChange={({ detail }) => setProjectName(detail.value)}
            placeholder="e.g., ML Research 2024"
          />
        </FormField>

        <FormField label="Description" description="Brief description of the project">
          <Textarea
            data-testid="project-description-input"
            value={projectDescription}
            onChange={({ detail }) => setProjectDescription(detail.value)}
            placeholder="Describe the project purpose..."
          />
        </FormField>

        <FormField label="Budget Limit (optional)" description="Maximum spending limit in USD">
          <Input
            data-testid="project-budget-input"
            type="number"
            value={projectBudget}
            onChange={({ detail }) => setProjectBudget(detail.value)}
            placeholder="1000.00"
          />
        </FormField>
      </SpaceBetween>
    </Modal>
  );

  // Create User Modal
  const CreateUserModal = () => (
    <Modal
      visible={userModalVisible}
      onDismiss={() => {
        setUserModalVisible(false);
        setUserValidationError('');
      }}
      header="Create New User"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={() => setUserModalVisible(false)} disabled={creatingUser}>Cancel</Button>
            <Button variant="primary" onClick={handleCreateUser} disabled={creatingUser} loading={creatingUser}>
              {creatingUser ? 'Creating...' : 'Create'}
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {userValidationError && (
          <ValidationError message={userValidationError} visible={true} />
        )}

        <FormField label="Username" description="Unique username for the user">
          <Input
            data-testid="user-username-input"
            value={username}
            onChange={({ detail }) => setUsername(detail.value)}
            placeholder="e.g., jsmith"
          />
        </FormField>

        <FormField label="Email" description="User's email address">
          <Input
            data-testid="user-email-input"
            type="email"
            value={userEmail}
            onChange={({ detail }) => setUserEmail(detail.value)}
            placeholder="user@example.com"
          />
        </FormField>

        <FormField label="Full Name" description="User's full name">
          <Input
            data-testid="user-fullname-input"
            value={userFullName}
            onChange={({ detail }) => setUserFullName(detail.value)}
            placeholder="John Smith"
          />
        </FormField>
      </SpaceBetween>
    </Modal>
  );

  // Create Budget Pool Modal
  const CreateBudgetModal = () => (
    <Modal
      visible={createBudgetModalVisible}
      onDismiss={() => {
        setCreateBudgetModalVisible(false);
        setBudgetValidationError('');
      }}
      header="Create Budget Pool"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={() => setCreateBudgetModalVisible(false)}>Cancel</Button>
            <Button variant="primary" data-testid="create-budget-submit-button" onClick={handleCreateBudget}>Create Budget</Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {budgetValidationError && (
          <Alert type="error" data-testid="budget-validation-error">
            {budgetValidationError}
          </Alert>
        )}

        <FormField label="Budget Name" description="E.g., 'NSF Grant CISE-2024-12345'">
          <Input
            data-testid="budget-name-input"
            value={budgetName}
            onChange={({ detail }) => setBudgetName(detail.value)}
            placeholder="Enter budget pool name"
          />
        </FormField>

        <FormField label="Description" description="Brief description of the funding source">
          <Textarea
            data-testid="budget-description-input"
            value={budgetDescription}
            onChange={({ detail }) => setBudgetDescription(detail.value)}
            placeholder="Describe the budget source..."
          />
        </FormField>

        <FormField label="Total Amount (USD)" description="Total funding available">
          <Input
            data-testid="budget-amount-input"
            value={totalAmount}
            onChange={({ detail }) => setTotalAmount(detail.value)}
            type="number"
            placeholder="50000.00"
          />
        </FormField>

        <FormField label="Budget Period" description="Timeframe for this budget">
          <Select
            data-testid="budget-period-select"
            selectedOption={{ label: period.charAt(0).toUpperCase() + period.slice(1), value: period }}
            onChange={({ detail }) => setPeriod(detail.selectedOption.value!)}
            options={[
              { label: 'Monthly', value: 'monthly' },
              { label: 'Quarterly', value: 'quarterly' },
              { label: 'Yearly', value: 'yearly' },
              { label: 'Project Lifetime', value: 'project' }
            ]}
          />
        </FormField>

        <FormField label="Alert Threshold (%)" description="Alert when spending exceeds this percentage">
          <Input
            data-testid="budget-threshold-input"
            value={alertThreshold}
            onChange={({ detail }) => setAlertThreshold(detail.value)}
            type="number"
            placeholder="80"
          />
        </FormField>
      </SpaceBetween>
    </Modal>
  );

  // Main render
  return (
    <ApiContext.Provider value={api as any}>
      <a
        href="#main-content"
        style={{ position: 'absolute', top: '-40px', left: 0, background: '#000', color: '#fff',
                 padding: '8px 16px', zIndex: 9999, fontWeight: 600 }}
        onFocus={(e) => { e.currentTarget.style.top = '0'; }}
        onBlur={(e) => { e.currentTarget.style.top = '-40px'; }}
      >
        Skip to main content
      </a>
      <AppLayoutShell
        sidebar={
          <SideNav
            activeView={state.activeView}
            onNavigate={(view) => setState(prev => ({ ...prev, activeView: view }))}
            instanceCount={state.instances.length}
            hasRunningInstances={state.instances.some(i => i.state === 'running')}
            templateCount={Object.keys(state.templates).length}
            pendingInvitations={state.invitations.filter(i => i.status === 'pending').length}
            activeCourses={state.courses.filter(c => c.status === 'active').length}
            activeWorkshops={state.workshops.filter(w => w.status === 'active').length}
            budgetPoolCount={state.budgetPools.length}
            pendingApprovalsCount={state.pendingApprovalsCount}
          />
        }
      >
        <div id="main-content" role="main" tabIndex={-1}>
          {/* Update Notification Banner */}
          {state.updateInfo && state.updateInfo.is_update_available && (
            <Alert
              type="info"
              dismissible
              onDismiss={() => setState(prev => ({ ...prev, updateInfo: prev.updateInfo ? { ...prev.updateInfo, is_update_available: false } : null }))}
              header={`New version available: ${state.updateInfo.latest_version}`}
            >
              <SpaceBetween size="xs">
                <div>You're currently running version {state.updateInfo.current_version}</div>
                <div><strong>Installation method:</strong> {state.updateInfo.install_method}</div>
                <div><strong>Update command:</strong> <code>{state.updateInfo.update_command}</code></div>
                <div>
                  <a href={state.updateInfo.release_url} target="_blank" rel="noopener noreferrer">
                    View release notes
                  </a>
                </div>
              </SpaceBetween>
            </Alert>
          )}
          {state.activeView === 'dashboard' && (
            <DashboardViewExtracted
              instances={state.instances}
              templates={state.templates}
              connected={state.connected}
              loading={state.loading}
              isFirstTimeUser={isFirstTimeUser}
              onNavigate={(view) => setState(prev => ({ ...prev, activeView: view as AppState['activeView'] }))}
              onRefresh={loadApplicationData}
              onShowQuickStart={() => setQuickStartWizardVisible(true)}
              onConnect={(info) => {
                setConnectionInfo(info)
                setConnectionModalVisible(true)
              }}
              onStartInstance={async (instanceName) => {
                try {
                  await api.startInstance(instanceName)
                  toast.success(`Starting workspace "${instanceName}"`)
                  setTimeout(loadApplicationData, 2000)
                } catch (error) {
                  toast.error(`Failed to start workspace: ${error instanceof Error ? error.message : 'Unknown error'}`)
                }
              }}
            />
          )}
          {state.activeView === 'templates' && (
            <TemplateSelectionViewExtracted
              templates={state.templates}
              loading={state.loading}
              selectedTemplate={state.selectedTemplate}
              onRefresh={loadApplicationData}
              onLaunch={() => setLaunchModalVisible(true)}
              onSelectTemplate={handleTemplateSelection}
            />
          )}
          {state.activeView === 'workspaces' && (
            <InstanceManagementViewExtracted
              instances={getFilteredInstances()}
              loading={state.loading}
              filterQuery={instancesFilterQuery}
              onFilterChange={setInstancesFilterQuery}
              selectedInstances={selectedInstances}
              onSelectionChange={setSelectedInstances}
              onRefresh={loadApplicationData}
              onNavigateToTemplates={() => setState(prev => ({ ...prev, activeView: 'templates' }))}
              onConnect={(info) => { setConnectionInfo(info); setConnectionModalVisible(true); }}
              onInstanceAction={handleInstanceAction}
              onBulkAction={handleBulkAction}
            />
          )}
          {state.activeView === 'terminal' && (() => {
              const runningInstances = state.instances.filter(i => i.state === 'running');

              if (runningInstances.length === 0) {
                return (
                  <Container header={<Header variant="h1">SSH Terminal</Header>}>
                    <Alert type="info">
                      No running workspaces available. Launch a workspace to access the SSH terminal.
                    </Alert>
                  </Container>
                );
              }

              return (
                <SpaceBetween size="l">
                  <Container header={<Header variant="h1">SSH Terminal</Header>}>
                    <SpaceBetween size="m">
                      <FormField label="Select Workspace">
                        <Select
                          selectedOption={state.selectedTerminalInstance ? { label: state.selectedTerminalInstance, value: state.selectedTerminalInstance } : null}
                          onChange={({ detail }) => setState({ ...state, selectedTerminalInstance: detail.selectedOption.value || '' })}
                          options={runningInstances.map(i => ({ label: i.name, value: i.name }))}
                          placeholder="Choose a workspace"
                        />
                      </FormField>
                      {state.selectedTerminalInstance && <Terminal instanceName={state.selectedTerminalInstance} />}
                    </SpaceBetween>
                  </Container>
                </SpaceBetween>
              );
            })()}
          {state.activeView === 'webview' && <WebViewView instances={state.instances} />}
          {state.activeView === 'storage' && (
            <StorageManagementViewExtracted
              efsVolumes={state.efsVolumes}
              ebsVolumes={state.ebsVolumes}
              instances={state.instances}
              loading={state.loading}
              onRefresh={loadApplicationData}
              onOpenCreateEFS={() => setCreateEFSModalVisible(true)}
              onOpenCreateEBS={() => setCreateEBSModalVisible(true)}
              onDeleteRequest={(volumeName, type, onConfirm) => {
                setDeleteModalConfig({ type, name: volumeName, requireNameConfirmation: false, onConfirm });
                setDeleteModalVisible(true);
              }}
            />
          )}
          {state.activeView === 'backups' && (
            <BackupManagementViewExtracted
              snapshots={state.snapshots}
              loading={state.loading}
              onRefresh={loadApplicationData}
              onNavigate={(view) => setState(prev => ({ ...prev, activeView: view as AppState['activeView'] }))}
              onCreateBackup={() => {
                setCreateBackupConfig({ instanceId: '', backupName: '', backupType: 'full', description: '' });
                setCreateBackupValidationAttempted(false);
                setCreateBackupModalVisible(true);
              }}
              onDeleteBackup={(item) => {
                setSelectedBackupForDelete(item);
                setDeleteBackupModalVisible(true);
              }}
              onRestoreBackup={(item) => {
                setSelectedBackupForRestore(item);
                setRestoreInstanceName('');
                setRestoreBackupModalVisible(true);
              }}
            />
          )}
          {state.activeView === 'projects' && (
            selectedProjectId ? (
              <ProjectDetailView
                projectId={selectedProjectId}
                onBack={() => setSelectedProjectId(null)}
              />
            ) : (
              <ProjectManagementViewExtracted
                projects={state.projects}
                loading={state.loading}
                onRefresh={loadApplicationData}
                onCreateProject={() => setProjectModalVisible(true)}
                onSelectProject={(id) => setSelectedProjectId(id)}
                onEditProject={(project) => {
                  setSelectedProjectForEdit(project);
                  setEditProjectName(project.name);
                  setEditProjectDescription(project.description || '');
                  setEditProjectStatus(project.status || 'active');
                  setShowEditProjectModal(true);
                }}
                onManageBudget={async (project) => {
                  setSelectedProjectForBudget(project);
                  setBudgetModalLoading(true);
                  setBudgetModalData(null);
                  setShowBudgetModal(true);
                  try {
                    const data = await api.getProjectBudget(project.id);
                    setBudgetModalData(data);
                  } catch (_e) {
                    setBudgetModalData(null);
                  } finally {
                    setBudgetModalLoading(false);
                  }
                }}
                onViewCost={async (project) => {
                  setSelectedProjectForCosts(project);
                  setCostModalLoading(true);
                  setCostModalData(null);
                  setShowCostModal(true);
                  try {
                    const data = await api.getProjectCosts(project.id);
                    setCostModalData(data);
                  } catch (_e) {
                    setCostModalData(null);
                  } finally {
                    setCostModalLoading(false);
                  }
                }}
                onViewUsage={async (project) => {
                  setSelectedProjectForUsage(project);
                  setUsageModalLoading(true);
                  setUsageModalData(null);
                  setShowUsageModal(true);
                  try {
                    const data = await api.getProjectUsage(project.id);
                    setUsageModalData(data);
                  } catch (_e) {
                    setUsageModalData(null);
                  } finally {
                    setUsageModalLoading(false);
                  }
                }}
                onManageMembers={async (project) => {
                  setSelectedProjectForMembers(project);
                  setManageMembersLoading(true);
                  setShowManageMembersModal(true);
                  try {
                    const members = await api.getProjectMembers(project.id);
                    setManageMembersData(members);
                  } catch (_e) {
                    setManageMembersData([]);
                  } finally {
                    setManageMembersLoading(false);
                  }
                }}
                onDeleteProject={(project) => {
                  setDeleteModalConfig({
                    type: 'project',
                    name: project.name,
                    requireNameConfirmation: false,
                    onConfirm: async () => {
                      try {
                        await api.deleteProject(project.id);
                        setState(prev => ({
                          ...prev,
                          projects: prev.projects.filter(p => p.id !== project.id),
                          notifications: [{
                            type: 'success',
                            header: 'Project Deleted',
                            content: `Project "${project.name}" has been successfully deleted.`,
                            dismissible: true,
                            id: Date.now().toString()
                          }, ...prev.notifications]
                        }));
                        setDeleteModalVisible(false);
                      } catch (error: any) {
                        setState(prev => ({
                          ...prev,
                          notifications: [{
                            type: 'error',
                            header: 'Delete Failed',
                            content: `Failed to delete project: ${error.message || 'Unknown error'}`,
                            dismissible: true,
                            id: Date.now().toString()
                          }, ...prev.notifications]
                        }));
                      }
                    }
                  });
                  setDeleteModalVisible(true);
                }}
              />
            )
          )}
          {state.activeView === 'invitations' && <InvitationManagementView />}
          {state.activeView === 'budgets' && <PlaceholderView title="Budget Pool Management" description="Manage budget pools and allocations for your research projects." />}
          {state.activeView === 'approvals' && <ApprovalsViewExtracted />}
          {state.activeView === 'courses' && <CoursesManagementView />}
          {state.activeView === 'workshops' && <WorkshopsManagementView />}
          {state.activeView === 'capacity-blocks' && <CapacityBlocksManagementView />}
          {state.activeView === 'project-detail' && <PlaceholderView title="Project Detail" description="Select a project from the Projects view to see its details." />}
          {state.activeView === 'users' && (
            <UserManagementViewExtracted
              users={state.users}
              instances={state.instances}
              loading={state.loading}
              onRefresh={loadApplicationData}
              onCreateUser={() => setUserModalVisible(true)}
              onEditUser={(user) => {
                setSelectedUserForEdit(user);
                setEditUserEmail(user.email || '');
                setEditUserDisplayName(user.display_name || (user as any).full_name || '');
                setEditUserRole('');
                setShowEditUserModal(true);
              }}
              onViewUserDetails={async (user) => {
                setSelectedUserForDetails(user);
                setUserDetailsModalVisible(true);
                setLoadingSSHKeys(true);
                try {
                  const response = await api.getUserSSHKeys(user.username);
                  setUserSSHKeys(response.keys || []);
                } catch (_e) {
                  setUserSSHKeys([]);
                } finally {
                  setLoadingSSHKeys(false);
                }
              }}
              onViewUserStatus={async (user) => {
                setSelectedUserForStatus(user);
                setUserStatusModalVisible(true);
                setLoadingUserStatus(true);
                try {
                  const statusData = await api.getUserStatus(user.username);
                  setUserStatusDetails(statusData);
                } catch (_e) {
                  setUserStatusDetails(null);
                } finally {
                  setLoadingUserStatus(false);
                }
              }}
              onProvisionUser={(username) => {
                const user = state.users.find(u => u.username === username) || null;
                setSelectedUserForProvision(user);
                setProvisionModalVisible(true);
              }}
              onManageSSHKeys={(username) => {
                setSelectedUsername(username);
                setSshKeyModalVisible(true);
              }}
              onDeleteUser={(user) => {
                const hasWorkspaces = (user.provisioned_instances?.length || 0) > 0;
                const workspaceWarning = hasWorkspaces
                  ? `This user has ${user.provisioned_instances!.length} provisioned workspace(s). Deleting the user will remove their access to these workspaces.`
                  : undefined;
                setDeleteModalConfig({
                  type: 'user',
                  name: user.username,
                  requireNameConfirmation: false,
                  warning: workspaceWarning,
                  onConfirm: async () => {
                    try {
                      await api.deleteUser(user.username);
                      usersVersionRef.current++;
                      setState(prev => ({
                        ...prev,
                        users: prev.users.filter(u => u.username !== user.username),
                        notifications: [{
                          type: 'success',
                          header: 'User Deleted',
                          content: `User "${user.username}" deleted successfully`,
                          dismissible: true,
                          id: Date.now().toString()
                        }, ...prev.notifications]
                      }));
                      setDeleteModalVisible(false);
                    } catch (error: any) {
                      setState(prev => ({
                        ...prev,
                        notifications: [{
                          type: 'error',
                          header: 'Delete Failed',
                          content: error.message || 'Failed to delete user',
                          dismissible: true,
                          id: Date.now().toString()
                        }, ...prev.notifications]
                      }));
                    }
                  }
                });
                setDeleteModalVisible(true);
              }}
            />
          )}
          {state.activeView === 'ami' && (
            <AMIManagementViewExtracted
              amis={state.amis}
              amiRegions={state.amiRegions}
              amiBuilds={state.amiBuilds}
              loading={state.loading}
              onRefresh={loadApplicationData}
            />
          )}
          {state.activeView === 'rightsizing' && <PlaceholderView title="Rightsizing Recommendations" description="Workspace rightsizing recommendations will help optimize your costs by suggesting better-sized workspaces based on actual usage patterns." />}
          {state.activeView === 'policy' && <PlaceholderView title="Policy Management" description="Policy management allows you to configure institutional policies, access controls, and governance rules for your Prism deployment." />}
          {state.activeView === 'marketplace' && (
            <MarketplaceViewExtracted
              marketplaceTemplates={state.marketplaceTemplates}
              marketplaceCategories={state.marketplaceCategories}
              loading={state.loading}
              onRefresh={loadApplicationData}
            />
          )}
          {state.activeView === 'idle' && (
            <IdleDetectionViewExtracted
              idlePolicies={state.idlePolicies}
              idleSchedules={state.idleSchedules}
              loading={state.loading}
              onRefresh={loadApplicationData}
            />
          )}
          {state.activeView === 'logs' && <LogsView instances={state.instances} loading={state.loading} onRefresh={loadApplicationData} />}
          {state.activeView === 'settings' && <SettingsView />}
        </div>
      </AppLayoutShell>
      <LaunchModal />
      <CreateBackupModal />
      <DeleteBackupModal />
      <RestoreBackupModal />
      <DeleteConfirmationModalExtracted
        visible={deleteModalVisible}
        config={deleteModalConfig}
        onDismiss={() => setDeleteModalVisible(false)}
      />
      {hibernateModalVisible && hibernateModalInstance && (
        <HibernateConfirmationModalExtracted
          instance={hibernateModalInstance}
          onDismiss={() => {
            setHibernateModalVisible(false);
            setHibernateModalInstance(null);
          }}
          onRefresh={loadApplicationData}
        />
      )}
      {idlePolicyModalInstance && (
        <IdlePolicyModalExtracted
          instanceName={idlePolicyModalInstance}
          onDismiss={() => setIdlePolicyModalInstance(null)}
        />
      )}
      <OnboardingWizard />
      <QuickStartWizard />
      <CreateProjectModal />
      <CreateUserModal />
      <CreateBudgetModal />
      <SSHKeyModal
        visible={sshKeyModalVisible}
        username={selectedUsername}
        onDismiss={() => {
          setSshKeyModalVisible(false);
          setSelectedUsername('');
        }}
        onGenerate={handleGenerateSSHKey}
      />

      {/* User Details Modal */}
      <Modal
        visible={userDetailsModalVisible}
        onDismiss={() => {
          setUserDetailsModalVisible(false);
          setSelectedUserForDetails(null);
          setUserSSHKeys([]);
        }}
        size="large"
        header={selectedUserForDetails ? `User Details: ${selectedUserForDetails.username}` : 'User Details'}
        data-testid="user-details-modal"
      >
        {selectedUserForDetails && (
          <SpaceBetween size="l">
            {/* User Information */}
            <Container header={<Header variant="h2">User Information</Header>}>
              <ColumnLayout columns={2} variant="text-grid">
                <div>
                  <Box variant="awsui-key-label">Username</Box>
                  <div>{selectedUserForDetails.username}</div>
                </div>
                <div>
                  <Box variant="awsui-key-label">Display Name</Box>
                  <div>{selectedUserForDetails.display_name}</div>
                </div>
                <div>
                  <Box variant="awsui-key-label">Email</Box>
                  <div>{selectedUserForDetails.email}</div>
                </div>
                <div>
                  <Box variant="awsui-key-label">UID</Box>
                  <div>{selectedUserForDetails.uid}</div>
                </div>
                <div>
                  <Box variant="awsui-key-label">Created</Box>
                  <div>{new Date(selectedUserForDetails.created_at).toLocaleString()}</div>
                </div>
                <div>
                  <Box variant="awsui-key-label">SSH Keys</Box>
                  <div>{selectedUserForDetails.ssh_keys || 0}</div>
                </div>
              </ColumnLayout>
            </Container>

            {/* SSH Keys Section */}
            <Container header={<Header variant="h2">SSH Keys</Header>}>
              {loadingSSHKeys ? (
                <Box textAlign="center" padding={{ vertical: 'xl' }}>
                  <Spinner size="large" />
                </Box>
              ) : userSSHKeys.length === 0 ? (
                <Box textAlign="center" color="inherit" padding={{ vertical: 'xl' }}>
                  <Box variant="p" color="text-body-secondary">
                    No SSH keys found for this user
                  </Box>
                </Box>
              ) : (
                <Table
                  columnDefinitions={[
                    {
                      id: "key_type",
                      header: "Type",
                      cell: (item: SSHKeyConfig) => item.key_type,
                      width: 100
                    },
                    {
                      id: "fingerprint",
                      header: "Fingerprint",
                      cell: (item: SSHKeyConfig) => (
                        <Box fontSize="body-s">
                          <span style={{ fontFamily: 'monospace' }}>{item.fingerprint}</span>
                        </Box>
                      )
                    },
                    {
                      id: "comment",
                      header: "Comment",
                      cell: (item: SSHKeyConfig) => item.comment,
                      width: 250
                    },
                    {
                      id: "created_at",
                      header: "Created",
                      cell: (item: SSHKeyConfig) => new Date(item.created_at).toLocaleString(),
                      width: 180
                    },
                    {
                      id: "auto_generated",
                      header: "Auto-Generated",
                      cell: (item: SSHKeyConfig) => item.auto_generated ? "Yes" : "No",
                      width: 120
                    }
                  ]}
                  items={userSSHKeys}
                  variant="embedded"
                  empty={
                    <Box textAlign="center" color="inherit">
                      <Box variant="p" color="text-body-secondary">
                        No SSH keys
                      </Box>
                    </Box>
                  }
                />
              )}
            </Container>

            {/* Provisioned Workspaces Section */}
            <Container header={<Header variant="h2">Provisioned Workspaces</Header>}>
              {selectedUserForDetails.provisioned_instances && selectedUserForDetails.provisioned_instances.length > 0 ? (
                <Table
                  columnDefinitions={[
                    {
                      id: "workspace",
                      header: "Workspace",
                      cell: (item: string) => item
                    }
                  ]}
                  items={selectedUserForDetails.provisioned_instances}
                  variant="embedded"
                  empty={
                    <Box textAlign="center" color="inherit">
                      <Box variant="p" color="text-body-secondary">
                        No provisioned workspaces
                      </Box>
                    </Box>
                  }
                />
              ) : (
                <Box textAlign="center" color="inherit" padding={{ vertical: 'xl' }}>
                  <Box variant="p" color="text-body-secondary">
                    No provisioned workspaces
                  </Box>
                </Box>
              )}
            </Container>
          </SpaceBetween>
        )}
      </Modal>

      {/* User Provision Modal */}
      <Modal
        visible={provisionModalVisible}
        onDismiss={() => {
          setProvisionModalVisible(false);
          setSelectedUserForProvision(null);
          setSelectedWorkspaceForProvision('');
        }}
        size="medium"
        header={selectedUserForProvision ? `Provision ${selectedUserForProvision.username} on Workspace` : 'Provision User'}
        data-testid="provision-modal"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button onClick={() => setProvisionModalVisible(false)} disabled={provisioningInProgress}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={async () => {
                  if (!selectedUserForProvision || !selectedWorkspaceForProvision) return;

                  setProvisioningInProgress(true);
                  try {
                    await api.provisionUser(selectedUserForProvision.username, selectedWorkspaceForProvision);

                    // Update user's provisioned instances
                    setState(prev => ({
                      ...prev,
                      users: prev.users.map(u =>
                        u.username === selectedUserForProvision.username
                          ? { ...u, provisioned_instances: [...(u.provisioned_instances || []), selectedWorkspaceForProvision] }
                          : u
                      ),
                      notifications: [
                        {
                          type: 'success',
                          header: 'User Provisioned',
                          content: `User "${selectedUserForProvision.username}" provisioned on workspace "${selectedWorkspaceForProvision}"`,
                          dismissible: true,
                          id: Date.now().toString()
                        },
                        ...prev.notifications
                      ]
                    }));
                    setProvisionModalVisible(false);
                    setSelectedWorkspaceForProvision('');
                  } catch (error: any) {
                    setState(prev => ({
                      ...prev,
                      notifications: [
                        {
                          type: 'error',
                          header: 'Provisioning Failed',
                          content: error.message || 'Failed to provision user',
                          dismissible: true,
                          id: Date.now().toString()
                        },
                        ...prev.notifications
                      ]
                    }));
                  } finally {
                    setProvisioningInProgress(false);
                  }
                }}
                disabled={!selectedWorkspaceForProvision || provisioningInProgress}
                loading={provisioningInProgress}
                data-testid="provision"
              >
                Provision
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        {selectedUserForProvision && (
          <SpaceBetween size="m">
            <FormField label="Workspace" description="Select a running workspace to provision this user on">
              <Select
                selectedOption={
                  selectedWorkspaceForProvision
                    ? { label: selectedWorkspaceForProvision, value: selectedWorkspaceForProvision }
                    : null
                }
                onChange={({ detail }) => setSelectedWorkspaceForProvision(detail.selectedOption?.value || '')}
                options={state.instances
                  .filter(instance => instance.state === 'running')
                  .map(instance => ({
                    label: instance.name,
                    value: instance.name
                  }))}
                placeholder="Select a workspace"
                empty="No running workspaces available"
                ariaLabel="Workspace"
              />
            </FormField>

            <Alert type="info">
              Provisioning will create a user account for "{selectedUserForProvision.username}" on the selected workspace with the same UID/GID and SSH keys.
            </Alert>
          </SpaceBetween>
        )}
      </Modal>

      {/* User Status Modal */}
      <Modal
        visible={userStatusModalVisible}
        onDismiss={() => {
          setUserStatusModalVisible(false);
          setSelectedUserForStatus(null);
          setUserStatusDetails(null);
        }}
        size="medium"
        header={selectedUserForStatus ? `User Status: ${selectedUserForStatus.username}` : 'User Status'}
        data-testid="user-status-modal"
        footer={
          <Box float="right">
            <Button onClick={() => setUserStatusModalVisible(false)} data-testid="close">
              Close
            </Button>
          </Box>
        }
      >
        {selectedUserForStatus && (
          <SpaceBetween size="m">
            {loadingUserStatus ? (
              <Box textAlign="center" padding={{ vertical: 'xl' }}>
                <Spinner size="large" />
              </Box>
            ) : userStatusDetails ? (
              <Container header={<Header variant="h2">Status Details</Header>}>
                <ColumnLayout columns={2} variant="text-grid">
                  <div>
                    <Box variant="awsui-key-label">Username</Box>
                    <div>{userStatusDetails.username}</div>
                  </div>
                  <div>
                    <Box variant="awsui-key-label">Status</Box>
                    <div>
                      <StatusIndicator type={userStatusDetails.status === 'active' ? 'success' : 'warning'}>
                        {userStatusDetails.status || 'active'}
                      </StatusIndicator>
                    </div>
                  </div>
                  <div>
                    <Box variant="awsui-key-label">SSH Keys</Box>
                    <div>{userStatusDetails.ssh_keys_count || 0}</div>
                  </div>
                  <div>
                    <Box variant="awsui-key-label">Provisioned Workspaces</Box>
                    <div>{userStatusDetails.provisioned_instances?.length || 0}</div>
                  </div>
                  {userStatusDetails.last_active && (
                    <div>
                      <Box variant="awsui-key-label">Last Active</Box>
                      <div>{new Date(userStatusDetails.last_active).toLocaleString()}</div>
                    </div>
                  )}
                </ColumnLayout>
              </Container>
            ) : (
              <Box textAlign="center" color="inherit" padding={{ vertical: 'xl' }}>
                <Box variant="p" color="text-body-secondary">
                  Failed to load user status
                </Box>
              </Box>
            )}
          </SpaceBetween>
        )}
      </Modal>

      {/* Send Invitation Modal */}
      <Modal
        visible={sendInvitationModalVisible}
        onDismiss={() => {
          setSendInvitationModalVisible(false);
          setInvitationValidationError('');
        }}
        header="Send Project Invitation"
        data-testid="send-invitation-modal"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button onClick={() => setSendInvitationModalVisible(false)}>Cancel</Button>
              <Button variant="primary" onClick={handleSendInvitation} data-testid="confirm-send-invitation">
                Send Invitation
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          {invitationValidationError && (
            <ValidationError message={invitationValidationError} visible={true} />
          )}

          <FormField label="Project" description="Select the project to invite user to">
            <Select
              selectedOption={
                selectedProjectForInvitation
                  ? { label: state.projects.find(p => p.id === selectedProjectForInvitation)?.name || '', value: selectedProjectForInvitation }
                  : null
              }
              onChange={({ detail }) => setSelectedProjectForInvitation(detail.selectedOption?.value || '')}
              options={state.projects.map(p => ({ label: p.name, value: p.id }))}
              placeholder="Select a project"
              data-testid="invitation-project-select"
            />
          </FormField>

          <FormField label="Email Address" description="Enter the recipient's email address">
            <Input
              value={invitationEmail}
              onChange={({ detail }) => setInvitationEmail(detail.value)}
              placeholder="user@example.com"
              type="email"
              data-testid="invitation-email-input"
            />
          </FormField>

          <FormField label="Role" description="Select the role for this user">
            <Select
              selectedOption={{ label: invitationRole, value: invitationRole }}
              onChange={({ detail }) => setInvitationRole(detail.selectedOption?.value as 'viewer' | 'member' | 'admin')}
              options={[
                { label: 'viewer', value: 'viewer', description: 'Read-only access' },
                { label: 'member', value: 'member', description: 'Can create and manage resources' },
                { label: 'admin', value: 'admin', description: 'Full project control' }
              ]}
              data-testid="invitation-role-select"
            />
          </FormField>

          <FormField label="Message (optional)" description="Add a personal message to the invitation">
            <Textarea
              value={invitationMessage}
              onChange={({ detail }) => setInvitationMessage(detail.value)}
              placeholder="Welcome to the project! Looking forward to collaborating..."
              data-testid="invitation-message-input"
            />
          </FormField>
        </SpaceBetween>
      </Modal>

      {/* Redeem Token Modal */}
      <Modal
        visible={redeemTokenModalVisible}
        onDismiss={() => {
          setRedeemTokenModalVisible(false);
          setTokenValidationError('');
        }}
        header="Redeem Invitation Token"
        data-testid="redeem-token-modal"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button onClick={() => setRedeemTokenModalVisible(false)}>Cancel</Button>
              <Button variant="primary" onClick={handleRedeemToken} data-testid="confirm-redeem-token">
                Redeem
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          {tokenValidationError && (
            <ValidationError message={tokenValidationError} visible={true} />
          )}

          <FormField label="Invitation Token" description="Enter the invitation token you received">
            <Input
              value={invitationToken}
              onChange={({ detail }) => setInvitationToken(detail.value)}
              placeholder="Enter token..."
              data-testid="invitation-token-input"
            />
          </FormField>

          <Alert type="info">
            The token can be found in your invitation email or shared by the project admin.
          </Alert>
        </SpaceBetween>
      </Modal>

      {/* Create EFS Volume Modal */}
      <Modal
        visible={createEFSModalVisible}
        onDismiss={() => {
          setCreateEFSModalVisible(false);
          setStorageVolumeName('');
          setStorageVolumeNameError('');
        }}
        header="Create EFS Volume"
        size="medium"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => {
                setCreateEFSModalVisible(false);
                setStorageVolumeName('');
                setStorageVolumeNameError('');
              }}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={async () => {
                  // Validate volume name
                  if (!storageVolumeName.trim()) {
                    setStorageVolumeNameError('Volume name is required');
                    return;
                  }

                  const volumeName = storageVolumeName;

                  // Close modal immediately - don't wait for AWS to finish
                  setCreateEFSModalVisible(false);
                  setStorageVolumeName('');
                  setStorageVolumeNameError('');

                  // Show notification that creation is in progress
                  setState(prev => ({
                    ...prev,
                    notifications: [
                      ...prev.notifications,
                      {
                        type: 'info',
                        header: 'Creating EFS Volume',
                        content: `Creating EFS volume "${volumeName}"... This may take 1-3 minutes.`,
                        dismissible: true,
                        id: Date.now().toString()
                      }
                    ]
                  }));

                  // Start creation in background - backend will wait for AWS
                  try {
                    await api.createEFSVolume(volumeName);
                    // Sync volume state from AWS to ensure we have current state
                    await api.syncEFSVolume(volumeName);
                    await loadApplicationData();
                    setState(prev => ({
                      ...prev,
                      notifications: [
                        ...prev.notifications,
                        {
                          type: 'success',
                          header: 'EFS Volume Created',
                          content: `Successfully created EFS volume "${volumeName}"`,
                          dismissible: true,
                          id: Date.now().toString()
                        }
                      ]
                    }));
                  } catch (error) {
                    setState(prev => ({
                      ...prev,
                      notifications: [
                        ...prev.notifications,
                        {
                          type: 'error',
                          header: 'Failed to Create EFS Volume',
                          content: error instanceof Error ? error.message : 'Unknown error occurred',
                          dismissible: true,
                          id: Date.now().toString()
                        }
                      ]
                    }));
                  }
                }}
              >
                Create
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <Form>
          <SpaceBetween size="m">
            {storageVolumeNameError && (
              <Box data-testid="validation-error" color="text-status-error">
                {storageVolumeNameError}
              </Box>
            )}
            <FormField
              label="EFS Volume Name"
              description="Enter a name for your EFS volume"
              errorText={storageVolumeNameError}
            >
              <Input
                value={storageVolumeName}
                onChange={({ detail }) => {
                  setStorageVolumeName(detail.value);
                  setStorageVolumeNameError(''); // Clear error on change
                }}
                placeholder="my-shared-data"
                ariaLabel="EFS Volume Name"
              />
            </FormField>
          </SpaceBetween>
        </Form>
      </Modal>

      {/* Create EBS Volume Modal */}
      <Modal
        visible={createEBSModalVisible}
        onDismiss={() => {
          setCreateEBSModalVisible(false);
          setStorageVolumeName('');
          setStorageVolumeSize('');
          setStorageVolumeNameError('');
          setStorageVolumeSizeError('');
        }}
        header="Create EBS Volume"
        size="medium"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => {
                setCreateEBSModalVisible(false);
                setStorageVolumeName('');
                setStorageVolumeSize('');
                setStorageVolumeNameError('');
                setStorageVolumeSizeError('');
              }}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={async () => {
                  // Validate volume name
                  if (!storageVolumeName.trim()) {
                    setStorageVolumeNameError('Volume name is required');
                    return;
                  }

                  // Validate volume size
                  if (!storageVolumeSize.trim()) {
                    setStorageVolumeSizeError('Volume size is required');
                    return;
                  }

                  const sizeNum = parseInt(storageVolumeSize);
                  if (isNaN(sizeNum) || sizeNum <= 0) {
                    setStorageVolumeSizeError('Volume size must be a positive number');
                    return;
                  }

                  const volumeName = storageVolumeName;
                  const volumeSize = storageVolumeSize;

                  // Close modal immediately - don't wait for AWS to finish
                  setCreateEBSModalVisible(false);
                  setStorageVolumeName('');
                  setStorageVolumeSize('');
                  setStorageVolumeNameError('');
                  setStorageVolumeSizeError('');

                  // Show notification that creation is in progress
                  setState(prev => ({
                    ...prev,
                    notifications: [
                      ...prev.notifications,
                      {
                        type: 'info',
                        header: 'Creating EBS Volume',
                        content: `Creating EBS volume "${volumeName}" (${volumeSize} GB)... This may take 30-120 seconds.`,
                        dismissible: true,
                        id: Date.now().toString()
                      }
                    ]
                  }));

                  // Start creation in background - backend will wait for AWS
                  try {
                    await api.createEBSVolume(volumeName, volumeSize);
                    // Sync volume state from AWS to ensure we have current state
                    await api.syncEBSVolume(volumeName);
                    await loadApplicationData();
                    setState(prev => ({
                      ...prev,
                      notifications: [
                        ...prev.notifications,
                        {
                          type: 'success',
                          header: 'EBS Volume Created',
                          content: `Successfully created EBS volume "${volumeName}"`,
                          dismissible: true,
                          id: Date.now().toString()
                        }
                      ]
                    }));
                  } catch (error) {
                    setState(prev => ({
                      ...prev,
                      notifications: [
                        ...prev.notifications,
                        {
                          type: 'error',
                          header: 'Failed to Create EBS Volume',
                          content: error instanceof Error ? error.message : 'Unknown error occurred',
                          dismissible: true,
                          id: Date.now().toString()
                        }
                      ]
                    }));
                  }
                }}
              >
                Create
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <Form>
          <SpaceBetween size="m">
            {(storageVolumeNameError || storageVolumeSizeError) && (
              <Box data-testid="validation-error" color="text-status-error">
                {storageVolumeNameError || storageVolumeSizeError}
              </Box>
            )}
            <FormField
              label="EBS Volume Name"
              description="Enter a name for your EBS volume"
              errorText={storageVolumeNameError}
            >
              <Input
                value={storageVolumeName}
                onChange={({ detail }) => {
                  setStorageVolumeName(detail.value);
                  setStorageVolumeNameError(''); // Clear error on change
                }}
                placeholder="my-private-data"
                ariaLabel="EBS Volume Name"
              />
            </FormField>
            <FormField
              label="EBS Volume Size (GB)"
              description="Enter the size of the volume in gigabytes"
              errorText={storageVolumeSizeError}
            >
              <Input
                value={storageVolumeSize}
                onChange={({ detail }) => {
                  setStorageVolumeSize(detail.value);
                  setStorageVolumeSizeError(''); // Clear error on change
                }}
                placeholder="100"
                type="number"
                ariaLabel="EBS Volume Size"
              />
            </FormField>
          </SpaceBetween>
        </Form>
      </Modal>

      {/* Connection Info Modal - at App level so it's always mounted regardless of active view */}
      <Modal
        visible={connectionModalVisible}
        onDismiss={() => {
          setConnectionModalVisible(false);
          setConnectionInfo(null);
        }}
        header="Connection Information"
        footer={
          <Box float="right">
            <Button
              variant="primary"
              onClick={() => {
                setConnectionModalVisible(false);
                setConnectionInfo(null);
              }}
            >
              Close
            </Button>
          </Box>
        }
      >
        {connectionInfo && (
          <SpaceBetween size="m">
            <FormField label="Workspace">
              <Box>{connectionInfo.instanceName}</Box>
            </FormField>
            {connectionInfo.publicIP && (
              <FormField label="Public IP" description="Instance public IP address">
                <Box data-testid="public-ip">{connectionInfo.publicIP}</Box>
              </FormField>
            )}
            <FormField label="SSH Command" description="Use this command to connect via SSH">
              <SpaceBetween direction="horizontal" size="xs">
                <code data-testid="ssh-command">{connectionInfo.sshCommand}</code>
                <Button
                  iconName="copy"
                  onClick={() => navigator.clipboard.writeText(connectionInfo!.sshCommand)}
                >
                  Copy SSH
                </Button>
              </SpaceBetween>
            </FormField>
            {/* web-url is always in DOM (even when empty) for ConnectionDialog.hasWebURL() to work */}
            <span data-testid="web-url" aria-hidden="true" style={{ display: 'none' }}>
              {connectionInfo.publicIP && connectionInfo.webPort
                ? `http://${connectionInfo.publicIP}:${connectionInfo.webPort}`
                : ''}
            </span>
            {connectionInfo.publicIP && connectionInfo.webPort && (
              <FormField label="Web URL" description="Access web services running on this instance">
                <Box>{`http://${connectionInfo.publicIP}:${connectionInfo.webPort}`}</Box>
              </FormField>
            )}
          </SpaceBetween>
        )}
      </Modal>

      {/* Edit Project Modal (#336) */}
      <Modal
        visible={showEditProjectModal}
        onDismiss={() => setShowEditProjectModal(false)}
        header="Edit Project"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setShowEditProjectModal(false)}>Cancel</Button>
              <Button
                variant="primary"
                loading={editProjectSubmitting}
                onClick={async () => {
                  if (!selectedProjectForEdit) return;
                  setEditProjectSubmitting(true);
                  try {
                    await api.updateProject(selectedProjectForEdit.id, {
                      name: editProjectName,
                      description: editProjectDescription,
                      status: editProjectStatus
                    });
                    const updatedProjects = await api.getProjects();
                    setState(prev => ({
                      ...prev,
                      projects: updatedProjects,
                      notifications: [
                        {
                          type: 'success',
                          header: 'Project Updated',
                          content: `Project "${editProjectName}" updated successfully.`,
                          dismissible: true,
                          id: Date.now().toString()
                        },
                        ...prev.notifications
                      ]
                    }));
                    setShowEditProjectModal(false);
                  } catch (error: any) {
                    setState(prev => ({
                      ...prev,
                      notifications: [
                        {
                          type: 'error',
                          header: 'Update Failed',
                          content: `Failed to update project: ${error.message || 'Unknown error'}`,
                          dismissible: true,
                          id: Date.now().toString()
                        },
                        ...prev.notifications
                      ]
                    }));
                  } finally {
                    setEditProjectSubmitting(false);
                  }
                }}
              >
                Save Changes
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <Form>
          <SpaceBetween size="m">
            <FormField label="Project Name">
              <Input
                value={editProjectName}
                onChange={({ detail }) => setEditProjectName(detail.value)}
                placeholder="Project name"
              />
            </FormField>
            <FormField label="Description">
              <Textarea
                value={editProjectDescription}
                onChange={({ detail }) => setEditProjectDescription(detail.value)}
                placeholder="Project description"
                rows={3}
              />
            </FormField>
            <FormField label="Status">
              <Select
                selectedOption={{ value: editProjectStatus, label: editProjectStatus }}
                onChange={({ detail }) => setEditProjectStatus(detail.selectedOption.value || 'active')}
                options={[
                  { value: 'active', label: 'Active' },
                  { value: 'paused', label: 'Paused' },
                  { value: 'completed', label: 'Completed' },
                  { value: 'archived', label: 'Archived' }
                ]}
              />
            </FormField>
          </SpaceBetween>
        </Form>
      </Modal>

      {/* Manage Members Modal (#332, #339) */}
      <Modal
        visible={showManageMembersModal}
        onDismiss={() => {
          setShowManageMembersModal(false);
          setAddMemberUsername('');
          setAddMemberRole('member');
        }}
        header={`Manage Members — ${selectedProjectForMembers?.name || ''}`}
        size="large"
        footer={
          <Box float="right">
            <Button variant="primary" onClick={() => setShowManageMembersModal(false)}>Close</Button>
          </Box>
        }
      >
        <SpaceBetween size="l">
          {manageMembersLoading ? (
            <Box textAlign="center"><Spinner /> Loading members...</Box>
          ) : (
            <Table
              columnDefinitions={[
                {
                  id: 'username',
                  header: 'Username',
                  cell: (member: MemberData) => member.username || member.user_id
                },
                {
                  id: 'role',
                  header: 'Role',
                  cell: (member: MemberData) => (
                    <Badge color={member.role === 'admin' ? 'red' : member.role === 'member' ? 'blue' : 'grey'}>
                      {member.role}
                    </Badge>
                  )
                },
                {
                  id: 'joined_at',
                  header: 'Joined',
                  cell: (member: MemberData) => member.joined_at ? new Date(member.joined_at).toLocaleDateString() : '-'
                },
                {
                  id: 'actions',
                  header: 'Actions',
                  cell: (member: MemberData) => (
                    <SpaceBetween direction="horizontal" size="xs">
                      <Select
                        selectedOption={{ value: member.role, label: member.role }}
                        onChange={async ({ detail }) => {
                          if (!selectedProjectForMembers) return;
                          try {
                            await api.updateProjectMember(selectedProjectForMembers.id, member.user_id, { role: detail.selectedOption.value });
                            const updated = await api.getProjectMembers(selectedProjectForMembers.id);
                            setManageMembersData(updated);
                          } catch (error: any) {
                            toast.error('Update Failed', { description: error.message || 'Failed to update role' });
                          }
                        }}
                        options={[
                          { value: 'viewer', label: 'Viewer' },
                          { value: 'member', label: 'Member' },
                          { value: 'admin', label: 'Admin' }
                        ]}
                      />
                      <Button
                        variant="link"
                        onClick={async () => {
                          if (!selectedProjectForMembers) return;
                          try {
                            await api.removeProjectMember(selectedProjectForMembers.id, member.user_id);
                            const updated = await api.getProjectMembers(selectedProjectForMembers.id);
                            setManageMembersData(updated);
                          } catch (error: any) {
                            toast.error('Remove Failed', { description: error.message || 'Failed to remove member' });
                          }
                        }}
                      >
                        Remove
                      </Button>
                    </SpaceBetween>
                  )
                }
              ]}
              items={manageMembersData}
              empty={<Box textAlign="center">No members yet.</Box>}
              header={<Header variant="h3">Current Members</Header>}
            />
          )}
          <Container header={<Header variant="h3">Add Member</Header>}>
            <SpaceBetween direction="horizontal" size="xs">
              <FormField label="Username">
                <Input
                  value={addMemberUsername}
                  onChange={({ detail }) => setAddMemberUsername(detail.value)}
                  placeholder="username"
                />
              </FormField>
              <FormField label="Role">
                <Select
                  selectedOption={{ value: addMemberRole, label: addMemberRole }}
                  onChange={({ detail }) => setAddMemberRole(detail.selectedOption.value || 'member')}
                  options={[
                    { value: 'viewer', label: 'Viewer' },
                    { value: 'member', label: 'Member' },
                    { value: 'admin', label: 'Admin' }
                  ]}
                />
              </FormField>
              <Box padding={{ top: 'xl' }}>
                <Button
                  variant="primary"
                  disabled={!addMemberUsername.trim()}
                  onClick={async () => {
                    if (!selectedProjectForMembers || !addMemberUsername.trim()) return;
                    try {
                      await api.addProjectMember(selectedProjectForMembers.id, { user_id: addMemberUsername, role: addMemberRole });
                      const updated = await api.getProjectMembers(selectedProjectForMembers.id);
                      setManageMembersData(updated);
                      setAddMemberUsername('');
                      setAddMemberRole('member');
                    } catch (error: any) {
                      toast.error('Add Failed', { description: error.message || 'Failed to add member' });
                    }
                  }}
                >
                  Add Member
                </Button>
              </Box>
            </SpaceBetween>
          </Container>
        </SpaceBetween>
      </Modal>

      {/* Budget Analysis Modal (#333) */}
      <Modal
        visible={showBudgetModal}
        onDismiss={() => setShowBudgetModal(false)}
        header={`Budget Analysis — ${selectedProjectForBudget?.name || ''}`}
        footer={<Box float="right"><Button variant="primary" onClick={() => setShowBudgetModal(false)}>Close</Button></Box>}
      >
        {budgetModalLoading ? (
          <Box textAlign="center"><Spinner /> Loading budget data...</Box>
        ) : budgetModalData ? (
          <SpaceBetween size="m">
            <ProgressBar
              value={Math.min((budgetModalData.spent_percentage || 0) * 100, 100)}
              status={
                (budgetModalData.spent_percentage || 0) >= 0.95 ? 'error' :
                (budgetModalData.spent_percentage || 0) >= 0.80 ? 'in-progress' : 'success'
              }
              label="Budget utilization"
              description={`${((budgetModalData.spent_percentage || 0) * 100).toFixed(1)}% used`}
            />
            <ColumnLayout columns={3} variant="text-grid">
              <div>
                <Box variant="awsui-key-label">Budget Limit</Box>
                <Box>${(budgetModalData.total_budget || 0).toFixed(2)}</Box>
              </div>
              <div>
                <Box variant="awsui-key-label">Spent to Date</Box>
                <Box>${(budgetModalData.spent_amount || 0).toFixed(2)}</Box>
              </div>
              <div>
                <Box variant="awsui-key-label">Remaining</Box>
                <Box>${(budgetModalData.remaining || 0).toFixed(2)}</Box>
              </div>
            </ColumnLayout>
            {budgetModalData.projected_monthly_spend !== undefined && (
              <ColumnLayout columns={2} variant="text-grid">
                <div>
                  <Box variant="awsui-key-label">Projected Monthly Spend</Box>
                  <Box>${budgetModalData.projected_monthly_spend.toFixed(2)}</Box>
                </div>
                {budgetModalData.days_until_exhausted !== undefined && (
                  <div>
                    <Box variant="awsui-key-label">Days Until Exhausted</Box>
                    <Box>{budgetModalData.days_until_exhausted} days</Box>
                  </div>
                )}
              </ColumnLayout>
            )}
            {budgetModalData.alert_count > 0 && (
              <Box color="text-status-warning">
                {budgetModalData.alert_count} budget alert(s) active
              </Box>
            )}
          </SpaceBetween>
        ) : (
          <Box color="text-body-secondary">No budget data available for this project.</Box>
        )}
      </Modal>

      {/* Cost Report Modal (#334) */}
      <Modal
        visible={showCostModal}
        onDismiss={() => setShowCostModal(false)}
        header={`Cost Report — ${selectedProjectForCosts?.name || ''}`}
        size="medium"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              {costModalData && (
                <Button
                  onClick={() => {
                    if (!costModalData) return;
                    const rows = [
                      ['Service', 'Amount ($)'],
                      ['Instances', costModalData.instances?.toFixed(2) ?? '0.00'],
                      ['Storage', costModalData.storage?.toFixed(2) ?? '0.00'],
                      ['Data Transfer', costModalData.data_transfer?.toFixed(2) ?? '0.00'],
                      ['Total', costModalData.total?.toFixed(2) ?? '0.00']
                    ];
                    const csv = rows.map(r => r.join(',')).join('\n');
                    const blob = new Blob([csv], { type: 'text/csv' });
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = `cost-report-${selectedProjectForCosts?.name || 'project'}.csv`;
                    a.click();
                    URL.revokeObjectURL(url);
                  }}
                >
                  Export CSV
                </Button>
              )}
              <Button variant="primary" onClick={() => setShowCostModal(false)}>Close</Button>
            </SpaceBetween>
          </Box>
        }
      >
        {costModalLoading ? (
          <Box textAlign="center"><Spinner /> Loading cost data...</Box>
        ) : costModalData ? (
          <SpaceBetween size="m">
            <Table
              columnDefinitions={[
                { id: 'service', header: 'Service', cell: (item: { service: string; amount: number }) => item.service },
                { id: 'amount', header: 'Amount', cell: (item: { service: string; amount: number }) => `$${item.amount.toFixed(2)}` }
              ]}
              items={[
                { service: 'Instances (EC2)', amount: costModalData.instances || 0 },
                { service: 'Storage', amount: costModalData.storage || 0 },
                { service: 'Data Transfer', amount: costModalData.data_transfer || 0 }
              ]}
              footer={
                <Box textAlign="right" fontWeight="bold">
                  Total: ${(costModalData.total || 0).toFixed(2)}
                </Box>
              }
            />
          </SpaceBetween>
        ) : (
          <Box color="text-body-secondary">No cost data available for this project.</Box>
        )}
      </Modal>

      {/* Usage Statistics Modal (#335) */}
      <Modal
        visible={showUsageModal}
        onDismiss={() => setShowUsageModal(false)}
        header={`Usage Statistics — ${selectedProjectForUsage?.name || ''}`}
        footer={<Box float="right"><Button variant="primary" onClick={() => setShowUsageModal(false)}>Close</Button></Box>}
      >
        {usageModalLoading ? (
          <Box textAlign="center"><Spinner /> Loading usage data...</Box>
        ) : usageModalData ? (
          <ColumnLayout columns={2} variant="text-grid">
            <div>
              <Box variant="awsui-key-label">Instance Hours</Box>
              <Box>{(usageModalData.instance_hours || 0).toFixed(1)} hrs</Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Storage (GB-hours)</Box>
              <Box>{(usageModalData.storage_gb_hours || 0).toFixed(1)} GB-hrs</Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Data Transfer</Box>
              <Box>{(usageModalData.data_transfer_gb || 0).toFixed(2)} GB</Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Period</Box>
              <Box>{usageModalData.period || 'current'}</Box>
            </div>
          </ColumnLayout>
        ) : (
          <Box color="text-body-secondary">No usage data available for this project.</Box>
        )}
      </Modal>

      {/* Edit User Modal (#349, #338) */}
      <Modal
        visible={showEditUserModal}
        onDismiss={() => setShowEditUserModal(false)}
        header={`Edit User — ${selectedUserForEdit?.username || ''}`}
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setShowEditUserModal(false)}>Cancel</Button>
              <Button
                variant="primary"
                loading={editUserSubmitting}
                onClick={async () => {
                  if (!selectedUserForEdit) return;
                  setEditUserSubmitting(true);
                  try {
                    const updates: Partial<UserUpdateRequest> = {};
                    if (editUserEmail) updates.email = editUserEmail;
                    if (editUserDisplayName) updates.display_name = editUserDisplayName;
                    if (editUserRole) updates.role = editUserRole;
                    await api.updateUser(selectedUserForEdit.username, updates);
                    const updatedUsers = await api.getUsers();
                    setState(prev => ({
                      ...prev,
                      users: updatedUsers,
                      notifications: [
                        {
                          type: 'success',
                          header: 'User Updated',
                          content: `User "${selectedUserForEdit.username}" updated successfully.`,
                          dismissible: true,
                          id: Date.now().toString()
                        },
                        ...prev.notifications
                      ]
                    }));
                    setShowEditUserModal(false);
                  } catch (error: any) {
                    setState(prev => ({
                      ...prev,
                      notifications: [
                        {
                          type: 'error',
                          header: 'Update Failed',
                          content: `Failed to update user: ${error.message || 'Unknown error'}`,
                          dismissible: true,
                          id: Date.now().toString()
                        },
                        ...prev.notifications
                      ]
                    }));
                  } finally {
                    setEditUserSubmitting(false);
                  }
                }}
              >
                Save Changes
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <Form>
          <SpaceBetween size="m">
            <FormField label="Email">
              <Input
                value={editUserEmail}
                onChange={({ detail }) => setEditUserEmail(detail.value)}
                placeholder="user@example.com"
                type="email"
              />
            </FormField>
            <FormField label="Display Name">
              <Input
                value={editUserDisplayName}
                onChange={({ detail }) => setEditUserDisplayName(detail.value)}
                placeholder="Display name"
              />
            </FormField>
            <FormField label="Role" description="Leave blank to keep current role">
              <Select
                selectedOption={editUserRole ? { value: editUserRole, label: editUserRole } : { value: '', label: 'Keep current role' }}
                onChange={({ detail }) => setEditUserRole(detail.selectedOption.value || '')}
                options={[
                  { value: '', label: 'Keep current role' },
                  { value: 'researcher', label: 'Researcher' },
                  { value: 'admin', label: 'Admin' },
                  { value: 'viewer', label: 'Viewer' }
                ]}
              />
            </FormField>
          </SpaceBetween>
        </Form>
      </Modal>
    </ApiContext.Provider>
  );
}
