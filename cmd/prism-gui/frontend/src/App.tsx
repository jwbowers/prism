// Prism GUI - Bulletproof AWS Integration
import { logger } from './utils/logger';
// Complete error handling, real API integration, professional UX

import React, { useState, useEffect, useRef } from 'react';
import './index.css';
import { toast } from 'sonner';
import { AppLayout as AppLayoutShell } from './components/app-layout';
import { SideNav } from './components/side-nav';
// ValidationError moved to extracted modal components
import { ProjectDetailView } from './components/ProjectDetailView';
import { InvitationManagementView } from './components/InvitationManagementView';
import { CoursesManagementView } from './views/CoursesManagementView';
import { WorkshopsManagementView } from './views/WorkshopsManagementView';
import { CapacityBlocksManagementView } from './views/CapacityBlocksManagementView';
import { TerminalView } from './views/TerminalView';
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
import { SettingsView as SettingsViewExtracted } from './views/SettingsView';
import { ApiContext } from './hooks/use-api';
import { SafePrismAPI } from './lib/api';
import { useInstanceActions } from './hooks/use-instance-actions';
import type { DeleteModalConfig } from './hooks/use-instance-actions';
import { useCrudHandlers } from './hooks/use-crud-handlers';
import { DeleteConfirmationModal as DeleteConfirmationModalExtracted } from './modals/DeleteConfirmationModal';
import { HibernateConfirmationModal as HibernateConfirmationModalExtracted } from './modals/HibernateConfirmationModal';
import { IdlePolicyModal as IdlePolicyModalExtracted } from './modals/IdlePolicyModal';
import { CreateProjectModal as CreateProjectModalExtracted } from './modals/CreateProjectModal';
import { CreateUserModal as CreateUserModalExtracted } from './modals/CreateUserModal';
import { CreateBudgetModal as CreateBudgetModalExtracted } from './modals/CreateBudgetModal';
import { OnboardingWizard as OnboardingWizardExtracted } from './modals/OnboardingWizard';
import { LaunchModal as LaunchModalExtracted } from './modals/LaunchModal';
import { CreateBackupModal as CreateBackupModalExtracted } from './modals/CreateBackupModal';
import { DeleteBackupModal as DeleteBackupModalExtracted } from './modals/DeleteBackupModal';
import { RestoreBackupModal as RestoreBackupModalExtracted } from './modals/RestoreBackupModal'
import { QuickStartWizard as QuickStartWizardExtracted } from './modals/QuickStartWizard';
import { UserDetailsModal as UserDetailsModalExtracted } from './modals/UserDetailsModal';
import { UserProvisionModal as UserProvisionModalExtracted } from './modals/UserProvisionModal';
import { UserStatusModal as UserStatusModalExtracted } from './modals/UserStatusModal';
import { SendInvitationModal as SendInvitationModalExtracted } from './modals/SendInvitationModal';
import { RedeemTokenModal as RedeemTokenModalExtracted } from './modals/RedeemTokenModal';
import { CreateEFSVolumeModal as CreateEFSVolumeModalExtracted } from './modals/CreateEFSVolumeModal';
import { CreateEBSVolumeModal as CreateEBSVolumeModalExtracted } from './modals/CreateEBSVolumeModal';
import { ConnectionInfoModal as ConnectionInfoModalExtracted } from './modals/ConnectionInfoModal';
import { EditProjectModal as EditProjectModalExtracted } from './modals/EditProjectModal';
import { ManageMembersModal as ManageMembersModalExtracted } from './modals/ManageMembersModal';
import { BudgetAnalysisModal as BudgetAnalysisModalExtracted } from './modals/BudgetAnalysisModal';
import { CostReportModal as CostReportModalExtracted } from './modals/CostReportModal';
import { UsageStatisticsModal as UsageStatisticsModalExtracted } from './modals/UsageStatisticsModal';
import { EditUserModal as EditUserModalExtracted } from './modals/EditUserModal';

import {
  SpaceBetween,
  Alert,
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

interface MemberData {
  user_id: string;
  username?: string;
  role: string;
  joined_at?: string;
  [key: string]: unknown;
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

interface UserStatus {
  username: string;
  status: string;
  provisioned_instances: string[];
  ssh_keys_count: number;
  last_active?: string;
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

interface ProjectUsageResponse {
  project_id: string;
  period: string;
  instance_hours: number;
  storage_gb_hours: number;
  data_transfer_gb: number;
  [key: string]: string | number;
}

interface WebService {
  name: string;
  port: number;
  url?: string;
  description?: string;
  type?: string;
}

// SafePrismAPI class is in src/lib/api.ts — used via useApi() hook

export default function PrismApp() {
  const api = new SafePrismAPI();

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
  // launchConfig state moved into LaunchModal component

  // Delete confirmation modal state
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [deleteModalConfig, setDeleteModalConfig] = useState<DeleteModalConfig>({
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
  // onboardingStep state moved into OnboardingWizard component
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

  // Bulk selection state for instances
  const [selectedInstances, setSelectedInstances] = useState<Instance[]>([]);

  // Filtering state for instances table
  const [instancesFilterQuery, setInstancesFilterQuery] = useState<{ tokens: Array<{ propertyKey?: string; operator: string; value: string }>; operation: 'and' | 'or' }>({ tokens: [], operation: 'and' });

  // Create Backup modal state
  const [createBackupModalVisible, setCreateBackupModalVisible] = useState(false);

  // Delete Backup modal state
  const [deleteBackupModalVisible, setDeleteBackupModalVisible] = useState(false);
  const [selectedBackupForDelete, setSelectedBackupForDelete] = useState<InstanceSnapshot | null>(null);

  // Restore Backup modal state
  const [restoreBackupModalVisible, setRestoreBackupModalVisible] = useState(false);
  const [selectedBackupForRestore, setSelectedBackupForRestore] = useState<InstanceSnapshot | null>(null);


  // Storage creation modal state
  const [createEFSModalVisible, setCreateEFSModalVisible] = useState(false);
  const [createEBSModalVisible, setCreateEBSModalVisible] = useState(false);

  // Create Project modal state
  const [projectModalVisible, setProjectModalVisible] = useState(false);

  // Project detail view state
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);

  // Create User modal state
  const [userModalVisible, setUserModalVisible] = useState(false);

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

  // Track users data version to prevent stale data overwrites
  // This prevents initial page load getUsers() from overwriting optimistic updates
  const usersVersionRef = useRef(0);

  // Individual invitation states
  const [sendInvitationModalVisible, setSendInvitationModalVisible] = useState(false);
  const [selectedProjectForInvitation, setSelectedProjectForInvitation] = useState<string>('');

  // Invitation token redemption
  const [redeemTokenModalVisible, setRedeemTokenModalVisible] = useState(false);

  // Project management modals
  const [showEditProjectModal, setShowEditProjectModal] = useState(false);
  const [selectedProjectForEdit, setSelectedProjectForEdit] = useState<Project | null>(null);
  const [showManageMembersModal, setShowManageMembersModal] = useState(false);
  const [selectedProjectForMembers, setSelectedProjectForMembers] = useState<Project | null>(null);
  const [manageMembersData, setManageMembersData] = useState<MemberData[]>([]);
  const [manageMembersLoading, setManageMembersLoading] = useState(false);
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

  // Helper: add a toast notification
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
  // getStatusLabel moved to src/views/SettingsView.tsx

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

  // Instance action handlers extracted to custom hook
  const { handleInstanceAction, handleBulkAction, getFilteredInstances } = useInstanceActions({
    api,
    instances: state.instances,
    setState,
    selectedInstances,
    setSelectedInstances,
    instancesFilterQuery,
    setHibernateModalInstance,
    setHibernateModalVisible,
    setConnectionInfo,
    setConnectionModalVisible,
    setIdlePolicyModalInstance,
    setDeleteModalConfig,
    setDeleteModalVisible,
    loadApplicationData,
  });

  // Safe template selection
  const handleTemplateSelection = (template: Template) => {
    try {
      setState(prev => ({ ...prev, selectedTemplate: template }));
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

  const {
    handleLaunchInstance,
    handleCreateProject,
    handleCreateUser,
    handleCreateBudget,
    handleGenerateSSHKey,
    handleSendInvitation,
    handleRedeemToken,
  } = useCrudHandlers({
    api,
    setState,
    selectedTemplate: state.selectedTemplate,
    handleModalDismiss,
    loadApplicationData,
    usersVersionRef,
    setProjectModalVisible,
    setUserModalVisible,
    setCreateBudgetModalVisible,
    setSendInvitationModalVisible,
    setRedeemTokenModalVisible,
    selectedProjectForInvitation,
  });

  // Storage Management View
  // Backup Management View
  // Settings View — extracted to src/views/SettingsView.tsx

  // Shared UserManagementView — used in both 'users' route and settings sub-view
  const userMgmtView = (
    <UserManagementViewExtracted
      users={state.users}
      instances={state.instances}
      loading={state.loading}
      onRefresh={loadApplicationData}
      onCreateUser={() => setUserModalVisible(true)}
      onEditUser={(user) => {
        setSelectedUserForEdit(user);
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

  // Main render
  return (
    <ApiContext.Provider value={api}>
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
          {state.activeView === 'terminal' && (
            <TerminalView
              instances={state.instances}
              selectedTerminalInstance={state.selectedTerminalInstance || ''}
              onSelectInstance={(name) => setState({ ...state, selectedTerminalInstance: name })}
            />
          )}
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
                setCreateBackupModalVisible(true);
              }}
              onDeleteBackup={(item) => {
                setSelectedBackupForDelete(item);
                setDeleteBackupModalVisible(true);
              }}
              onRestoreBackup={(item) => {
                setSelectedBackupForRestore(item);
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
          {state.activeView === 'users' && userMgmtView}
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
          {state.activeView === 'settings' && (
            <SettingsViewExtracted
              settingsSection={state.settingsSection}
              onSectionChange={(section) => setState(prev => ({ ...prev, settingsSection: section as typeof prev.settingsSection }))}
              connected={state.connected}
              loading={state.loading}
              instanceCount={state.instances.length}
              efsVolumeCount={state.efsVolumes.length}
              ebsVolumeCount={state.ebsVolumes.length}
              updateInfo={state.updateInfo}
              autoStartEnabled={state.autoStartEnabled}
              onRefresh={loadApplicationData}
              onSetAutoStart={async (enabled) => {
                try {
                  await api.setAutoStart(enabled);
                  setState(prev => ({ ...prev, autoStartEnabled: enabled }));
                  setNotification({
                    type: 'success',
                    content: `Auto-start ${enabled ? 'enabled' : 'disabled'} successfully`,
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
              subViews={{
                profiles: <ProfileSelectorViewExtracted />,
                users: userMgmtView,
                ami: (
                  <AMIManagementViewExtracted
                    amis={state.amis}
                    amiRegions={state.amiRegions}
                    amiBuilds={state.amiBuilds}
                    loading={state.loading}
                    onRefresh={loadApplicationData}
                  />
                ),
                marketplace: (
                  <MarketplaceViewExtracted
                    marketplaceTemplates={state.marketplaceTemplates}
                    marketplaceCategories={state.marketplaceCategories}
                    loading={state.loading}
                    onRefresh={loadApplicationData}
                  />
                ),
                idle: (
                  <IdleDetectionViewExtracted
                    idlePolicies={state.idlePolicies}
                    idleSchedules={state.idleSchedules}
                    loading={state.loading}
                    onRefresh={loadApplicationData}
                  />
                ),
                logs: <LogsView instances={state.instances} loading={state.loading} onRefresh={loadApplicationData} />,
              }}
            />
          )}
        </div>
      </AppLayoutShell>
      <LaunchModalExtracted
        visible={launchModalVisible}
        selectedTemplate={state.selectedTemplate}
        onDismiss={handleModalDismiss}
        onLaunch={handleLaunchInstance}
      />
      <CreateBackupModalExtracted
        visible={createBackupModalVisible}
        instances={state.instances}
        onDismiss={() => setCreateBackupModalVisible(false)}
        onSuccess={loadApplicationData}
      />
      <DeleteBackupModalExtracted
        visible={deleteBackupModalVisible}
        backup={selectedBackupForDelete}
        onDismiss={() => { setDeleteBackupModalVisible(false); setSelectedBackupForDelete(null); }}
        onSuccess={loadApplicationData}
      />
      <RestoreBackupModalExtracted
        visible={restoreBackupModalVisible}
        backup={selectedBackupForRestore}
        onDismiss={() => { setRestoreBackupModalVisible(false); setSelectedBackupForRestore(null); }}
        onSuccess={loadApplicationData}
      />
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
      <OnboardingWizardExtracted
        visible={onboardingVisible}
        onComplete={() => {
          setOnboardingComplete(true);
          setOnboardingVisible(false);
        }}
      />
      <QuickStartWizardExtracted
        visible={quickStartWizardVisible}
        templates={state.templates}
        onDismiss={() => setQuickStartWizardVisible(false)}
        onSuccess={loadApplicationData}
        onNavigateToWorkspaces={() => setState(prev => ({ ...prev, activeView: 'workspaces' }))}
      />
      <CreateProjectModalExtracted
        visible={projectModalVisible}
        onDismiss={() => setProjectModalVisible(false)}
        onSubmit={handleCreateProject}
      />
      <CreateUserModalExtracted
        visible={userModalVisible}
        onDismiss={() => setUserModalVisible(false)}
        onSubmit={handleCreateUser}
      />
      <CreateBudgetModalExtracted
        visible={createBudgetModalVisible}
        onDismiss={() => setCreateBudgetModalVisible(false)}
        onSubmit={handleCreateBudget}
      />
      <SSHKeyModal
        visible={sshKeyModalVisible}
        username={selectedUsername}
        onDismiss={() => {
          setSshKeyModalVisible(false);
          setSelectedUsername('');
        }}
        onGenerate={handleGenerateSSHKey}
      />

      <UserDetailsModalExtracted
        visible={userDetailsModalVisible}
        user={selectedUserForDetails}
        sshKeys={userSSHKeys}
        loadingSSHKeys={loadingSSHKeys}
        onDismiss={() => {
          setUserDetailsModalVisible(false);
          setSelectedUserForDetails(null);
          setUserSSHKeys([]);
        }}
      />

      <UserProvisionModalExtracted
        visible={provisionModalVisible}
        user={selectedUserForProvision}
        instances={state.instances}
        onDismiss={() => {
          setProvisionModalVisible(false);
          setSelectedUserForProvision(null);
        }}
        onProvision={async (username, workspaceName) => {
          await api.provisionUser(username, workspaceName);
          setState(prev => ({
            ...prev,
            users: prev.users.map(u =>
              u.username === username
                ? { ...u, provisioned_instances: [...(u.provisioned_instances || []), workspaceName] }
                : u
            ),
            notifications: [
              {
                type: 'success',
                header: 'User Provisioned',
                content: `User "${username}" provisioned on workspace "${workspaceName}"`,
                dismissible: true,
                id: Date.now().toString()
              },
              ...prev.notifications
            ]
          }));
          setProvisionModalVisible(false);
        }}
      />

      <UserStatusModalExtracted
        visible={userStatusModalVisible}
        user={selectedUserForStatus}
        statusDetails={userStatusDetails}
        loading={loadingUserStatus}
        onDismiss={() => {
          setUserStatusModalVisible(false);
          setSelectedUserForStatus(null);
          setUserStatusDetails(null);
        }}
      />

      <SendInvitationModalExtracted
        visible={sendInvitationModalVisible}
        projects={state.projects}
        selectedProjectId={selectedProjectForInvitation}
        onProjectChange={setSelectedProjectForInvitation}
        onDismiss={() => setSendInvitationModalVisible(false)}
        onSubmit={handleSendInvitation}
      />

      <RedeemTokenModalExtracted
        visible={redeemTokenModalVisible}
        onDismiss={() => setRedeemTokenModalVisible(false)}
        onSubmit={handleRedeemToken}
      />

      <CreateEFSVolumeModalExtracted
        visible={createEFSModalVisible}
        onDismiss={() => setCreateEFSModalVisible(false)}
        onSuccess={loadApplicationData}
        onNotify={(notification) => {
          setState(prev => ({
            ...prev,
            notifications: [...prev.notifications, notification as any]
          }));
        }}
      />

      <CreateEBSVolumeModalExtracted
        visible={createEBSModalVisible}
        onDismiss={() => setCreateEBSModalVisible(false)}
        onSuccess={loadApplicationData}
        onNotify={(notification) => {
          setState(prev => ({
            ...prev,
            notifications: [...prev.notifications, notification as any]
          }));
        }}
      />

      {/* Connection Info Modal - at App level so it's always mounted regardless of active view */}
      <ConnectionInfoModalExtracted
        visible={connectionModalVisible}
        connectionInfo={connectionInfo}
        onDismiss={() => {
          setConnectionModalVisible(false);
          setConnectionInfo(null);
        }}
      />

      {/* Edit Project Modal (#336) */}
      <EditProjectModalExtracted
        visible={showEditProjectModal}
        project={selectedProjectForEdit}
        onDismiss={() => setShowEditProjectModal(false)}
        onSubmit={async (projectId, data) => {
          await api.updateProject(projectId, {
            name: data.name,
            description: data.description,
            status: data.status
          });
          const updatedProjects = await api.getProjects();
          setState(prev => ({
            ...prev,
            projects: updatedProjects,
            notifications: [
              {
                type: 'success',
                header: 'Project Updated',
                content: `Project "${data.name}" updated successfully.`,
                dismissible: true,
                id: Date.now().toString()
              },
              ...prev.notifications
            ]
          }));
          setShowEditProjectModal(false);
        }}
      />

      {/* Manage Members Modal (#332, #339) */}
      <ManageMembersModalExtracted
        visible={showManageMembersModal}
        project={selectedProjectForMembers}
        members={manageMembersData}
        loading={manageMembersLoading}
        onDismiss={() => setShowManageMembersModal(false)}
        onMembersChange={setManageMembersData}
      />

      {/* Budget Analysis Modal (#333) */}
      <BudgetAnalysisModalExtracted
        visible={showBudgetModal}
        project={selectedProjectForBudget}
        budgetData={budgetModalData}
        loading={budgetModalLoading}
        onDismiss={() => setShowBudgetModal(false)}
      />

      {/* Cost Report Modal (#334) */}
      <CostReportModalExtracted
        visible={showCostModal}
        project={selectedProjectForCosts}
        costData={costModalData}
        loading={costModalLoading}
        onDismiss={() => setShowCostModal(false)}
      />

      {/* Usage Statistics Modal (#335) */}
      <UsageStatisticsModalExtracted
        visible={showUsageModal}
        project={selectedProjectForUsage}
        usageData={usageModalData}
        loading={usageModalLoading}
        onDismiss={() => setShowUsageModal(false)}
      />

      {/* Edit User Modal (#349, #338) */}
      <EditUserModalExtracted
        visible={showEditUserModal}
        user={selectedUserForEdit}
        onDismiss={() => setShowEditUserModal(false)}
        onSubmit={async (username, updates) => {
          await api.updateUser(username, updates);
          const updatedUsers = await api.getUsers();
          setState(prev => ({
            ...prev,
            users: updatedUsers,
            notifications: [
              {
                type: 'success',
                header: 'User Updated',
                content: `User "${username}" updated successfully.`,
                dismissible: true,
                id: Date.now().toString()
              },
              ...prev.notifications
            ]
          }));
          setShowEditUserModal(false);
        }}
      />
    </ApiContext.Provider>
  );
}
