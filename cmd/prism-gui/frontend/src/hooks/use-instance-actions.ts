import type { Instance } from '../lib/types';
import { SafePrismAPI } from '../lib/api';
import { logger } from '../utils/logger';
import React from 'react';

export interface DeleteModalConfig {
  type: 'workspace' | 'efs-volume' | 'ebs-volume' | 'project' | 'user' | null;
  name: string;
  requireNameConfirmation: boolean;
  warning?: string;
  onConfirm: () => Promise<void>;
}

interface UseInstanceActionsOptions {
  api: SafePrismAPI;
  instances: Instance[];
  setState: React.Dispatch<React.SetStateAction<any>>;
  selectedInstances: Instance[];
  setSelectedInstances: React.Dispatch<React.SetStateAction<Instance[]>>;
  instancesFilterQuery: { tokens: Array<{ propertyKey?: string; operator: string; value: string }>; operation: 'and' | 'or' };
  setHibernateModalInstance: React.Dispatch<React.SetStateAction<Instance | null>>;
  setHibernateModalVisible: React.Dispatch<React.SetStateAction<boolean>>;
  setConnectionInfo: React.Dispatch<React.SetStateAction<{ instanceName: string; publicIP: string; sshCommand: string; webPort: string } | null>>;
  setConnectionModalVisible: React.Dispatch<React.SetStateAction<boolean>>;
  setIdlePolicyModalInstance: React.Dispatch<React.SetStateAction<string | null>>;
  setDeleteModalConfig: React.Dispatch<React.SetStateAction<DeleteModalConfig>>;
  setDeleteModalVisible: React.Dispatch<React.SetStateAction<boolean>>;
  loadApplicationData: () => Promise<void>;
}

export function useInstanceActions(options: UseInstanceActionsOptions) {
  const {
    api,
    instances,
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
  } = options;

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
      setState((prev: any) => ({
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
        setState((prev: any) => ({
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
        setState((prev: any) => ({
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
      setState((prev: any) => ({ ...prev, loading: true }));

      let actionMessage = '';

      switch (action) {
        case 'connect': {
          // Show connection info modal (fire-and-forget style - no loading state)
          const ip = instance.public_ip || '';
          const user = instance.username || 'ubuntu';
          const sshCmd = ip ? `ssh ${user}@${ip}` : `ssh ${user}@<instance-ip>`;
          setState((prev: any) => ({ ...prev, loading: false }));
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
          setState((prev: any) => ({
            ...prev,
            activeView: 'terminal',
            selectedTerminalInstance: instance.name,
            loading: false
          }));
          return; // Don't continue with normal flow
        case 'webservice':
          // Open webview view (user will select specific service)
          setState((prev: any) => ({
            ...prev,
            activeView: 'webview',
            loading: false
          }));
          return; // Don't continue with normal flow
        case 'manage-idle-policy':
          setState((prev: any) => ({ ...prev, loading: false }));
          setIdlePolicyModalInstance(instance.name);
          return;
        case 'delete':
          // Show confirmation modal instead of deleting immediately
          setState((prev: any) => ({ ...prev, loading: false }));
          setDeleteModalConfig({
            type: 'workspace',
            name: instance.name,
            requireNameConfirmation: true,
            onConfirm: async () => {
              try {
                await api.deleteInstance(instance.name);
                setState((prev: any) => ({
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
                setState((prev: any) => ({
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
      setState((prev: any) => ({
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

      setState((prev: any) => ({
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

  // Internal helper — only called by handleBulkAction
  const executeBulkAction = async (action: 'start' | 'stop' | 'hibernate' | 'delete') => {
    try {
      setState((prev: any) => ({ ...prev, loading: true }));

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
      setState((prev: any) => ({
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

      setState((prev: any) => ({
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

  // Bulk action handlers for multiple instances
  const handleBulkAction = async (action: 'start' | 'stop' | 'hibernate' | 'delete') => {
    if (selectedInstances.length === 0) {
      setState((prev: any) => ({
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

  // Filter instances based on PropertyFilter query
  const getFilteredInstances = () => {
    if (!instancesFilterQuery.tokens || instancesFilterQuery.tokens.length === 0) {
      return instances;
    }

    return instances.filter((instance) => {
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

  return { handleInstanceAction, handleBulkAction, getFilteredInstances };
}
