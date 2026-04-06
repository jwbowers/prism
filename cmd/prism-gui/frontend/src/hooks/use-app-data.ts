import { useCallback, useEffect } from 'react'
import React from 'react'
import { toast } from 'sonner'
import { SafePrismAPI } from '../lib/api'
import { logger } from '../utils/logger'
import type {
  Template,
  Instance,
  EFSVolume,
  EBSVolume,
  InstanceSnapshot,
  Project,
  User,
  AMI,
  AMIBuild,
  AMIRegion,
  RightsizingRecommendation,
  PolicyStatus,
  PolicySet,
  MarketplaceTemplate,
  MarketplaceCategory,
  IdlePolicy,
  IdleSchedule,
  Invitation,
  CachedInvitation,
  Course,
  WorkshopEvent,
  BudgetData,
  Budget,
} from '../lib/types'

interface UseAppDataOptions {
  api: SafePrismAPI
  setState: React.Dispatch<React.SetStateAction<any>>
  activeView: string
  usersVersionRef: React.MutableRefObject<number>
}

export function useAppData({ api, setState, activeView, usersVersionRef }: UseAppDataOptions) {
  // Safe data loading with comprehensive error handling
  // Wrapped in useCallback to prevent unnecessary re-renders in dependent useEffect hooks
  const loadApplicationData = useCallback(async () => {
    try {
      setState((prev: any) => ({ ...prev, loading: true, error: null }));

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

      setState((prev: any) => ({
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
      api.listAllApprovals('pending').then((approvals: any[]) =>
        setState((prev: any) => ({ ...prev, pendingApprovalsCount: approvals.length }))
      ).catch(() => {});

    } catch (error) {
      logger.error('Failed to load application data:', error);

      toast.error('Connection Error', { description: `Failed to connect to Prism daemon: ${error instanceof Error ? error.message : 'Unknown error'}` });
      setState((prev: any) => ({
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
  const loadBudgetData = useCallback(async () => {
    try {
      const results = await Promise.allSettled([
        api.getBudgets(),
        api.getBudgetPools()
      ]);

      const budgetsData = results[0].status === 'fulfilled' ? results[0].value as BudgetData[] : [];
      const budgetPoolsData = results[1].status === 'fulfilled' ? results[1].value as Budget[] : [];

      setState((prev: any) => ({
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
    if (activeView === 'budgets' || activeView === 'project-detail') {
      loadBudgetData();
    }
  }, [activeView, loadBudgetData]);

  // Load data on mount and refresh periodically
  // NOTE: Budget loading on navigation is handled by the separate effect above
  // This effect intentionally uses [] deps to avoid re-triggering on every navigation
  useEffect(() => {
    loadApplicationData();
    const interval = setInterval(loadApplicationData, 30000);
    return () => clearInterval(interval);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only on mount - prevents unnecessary reloads on every navigation

  return { loadApplicationData, loadBudgetData }
}
