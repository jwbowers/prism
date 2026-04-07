// Prism GUI - Bulletproof AWS Integration
import { logger } from './utils/logger';
// Complete error handling, real API integration, professional UX

import { useState, useEffect, useRef } from 'react';
import './index.css';
import { toast } from 'sonner';
import { AppLayout as AppLayoutShell } from './components/app-layout';
import { SideNav } from './components/side-nav';
// ValidationError moved to extracted modal components
import { ProjectDetailView } from './components/ProjectDetailView';
import { InvitationManagementView } from './components/InvitationManagementView';
import { BudgetPoolManagementView } from './views/BudgetPoolManagementView';
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
import { useAppData } from './hooks/use-app-data';
import { useInstanceActions } from './hooks/use-instance-actions';
import type { DeleteModalConfig } from './hooks/use-instance-actions';
import { useCrudHandlers } from './hooks/use-crud-handlers';
import { DeleteConfirmationModal as DeleteConfirmationModalExtracted } from './modals/DeleteConfirmationModal';
import { HibernateConfirmationModal as HibernateConfirmationModalExtracted } from './modals/HibernateConfirmationModal';
import { IdlePolicyModal as IdlePolicyModalExtracted } from './modals/IdlePolicyModal';
import { CreateProjectModal as CreateProjectModalExtracted } from './modals/CreateProjectModal';
import { CreateUserModal as CreateUserModalExtracted } from './modals/CreateUserModal';
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

import type {
  Project,
  User,
  Template,
  Instance,
  InstanceSnapshot,
  BudgetData,
  CostBreakdown,
  MemberData,
  AppState,
  UserStatus,
  SSHKeyConfig,
  ProjectUsageResponse,
} from './lib/types';

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

  // Track users data version to prevent stale data overwrites
  // This prevents initial page load getUsers() from overwriting optimistic updates
  const usersVersionRef = useRef(0);

  const { loadApplicationData } = useAppData({
    api,
    setState,
    activeView: state.activeView,
    usersVersionRef,
  });

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

  // Utility function to get accessible status labels (WCAG 1.1.1)
  // getStatusLabel moved to src/views/SettingsView.tsx

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
          {state.activeView === 'budgets' && <BudgetPoolManagementView budgetPools={state.budgetPools} loading={state.loading} onRefresh={loadApplicationData} />}
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
