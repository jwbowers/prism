/**
 * GovernancePanel — v0.13.0
 *
 * 5-tab governance panel for projects:
 *   Quotas | Grant Period | Budget Sharing | Onboarding Templates | Monthly Report
 *
 * Accessed via ProjectDetailView's Governance tab.
 * Uses window.__apiClient (set in App.tsx).
 */

import React from 'react';
import {
  Tabs,
  Container,
  Header,
  SpaceBetween,
  Button,
  Table,
  Modal,
  Form,
  FormField,
  Input,
  Select,
  Toggle,
  Box,
  Spinner,
  Badge,
  Alert,
  Textarea,
  ColumnLayout
} from '../lib/cloudscape-shim';

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

interface BudgetShareRequest {
  from_project_id: string;
  to_project_id?: string;
  to_member_id?: string;
  amount: number;
  reason?: string;
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

export interface GovernancePanelProps {
  projectId: string;
}

export const GovernancePanel: React.FC<GovernancePanelProps> = ({ projectId }) => {
  const apiClient = (window as any).__apiClient;

  // ── Quotas ────────────────────────────────────────────────────────────────
  const [quotas, setQuotas] = React.useState<RoleQuota[]>([]);
  const [quotasLoading, setQuotasLoading] = React.useState(true);
  const [quotaError, setQuotaError] = React.useState<string | null>(null);
  const [quotaModalVisible, setQuotaModalVisible] = React.useState(false);
  const [quotaRole, setQuotaRole] = React.useState('member');
  const [quotaMaxInstances, setQuotaMaxInstances] = React.useState('');
  const [quotaMaxInstanceType, setQuotaMaxInstanceType] = React.useState('');
  const [quotaMaxSpendDaily, setQuotaMaxSpendDaily] = React.useState('');

  const loadQuotas = async () => {
    setQuotasLoading(true);
    try {
      const result = await apiClient.getProjectQuotas(projectId);
      setQuotas(result);
    } catch (e: any) {
      setQuotaError(e.message || 'Failed to load quotas');
    } finally {
      setQuotasLoading(false);
    }
  };

  React.useEffect(() => { loadQuotas(); }, [projectId]); // eslint-disable-line react-hooks/exhaustive-deps

  const openQuotaModal = () => {
    setQuotaRole('member');
    setQuotaMaxInstances('');
    setQuotaMaxInstanceType('');
    setQuotaMaxSpendDaily('');
    setQuotaModalVisible(true);
  };

  const saveQuota = async () => {
    setQuotaModalVisible(false);
    try {
      await apiClient.setProjectQuota(projectId, {
        role: quotaRole,
        max_instances: parseInt(quotaMaxInstances) || -1,
        max_instance_type: quotaMaxInstanceType,
        max_spend_daily: parseFloat(quotaMaxSpendDaily) || 0
      });
      loadQuotas();
    } catch (e: any) {
      setQuotaError(e.message || 'Failed to set quota');
    }
  };

  const deleteQuota = async (role: string) => {
    try {
      await apiClient.deleteProjectQuota(projectId, role);
      loadQuotas();
    } catch (e: any) {
      setQuotaError(e.message || 'Failed to delete quota');
    }
  };

  // ── Grant Period ──────────────────────────────────────────────────────────
  const [grantPeriod, setGrantPeriod] = React.useState<GrantPeriod | null>(null);
  const [grantLoading, setGrantLoading] = React.useState(true);
  const [grantError, setGrantError] = React.useState<string | null>(null);
  const [grantModalVisible, setGrantModalVisible] = React.useState(false);
  const [grantName, setGrantName] = React.useState('');
  const [grantStartDate, setGrantStartDate] = React.useState('');
  const [grantEndDate, setGrantEndDate] = React.useState('');
  const [grantAutoFreeze, setGrantAutoFreeze] = React.useState(false);
  const [grantDeleteConfirm, setGrantDeleteConfirm] = React.useState(false);

  const loadGrantPeriod = async () => {
    setGrantLoading(true);
    try {
      const result = await apiClient.getGrantPeriod(projectId);
      setGrantPeriod(result);
    } catch (e: any) {
      setGrantError(e.message || 'Failed to load grant period');
    } finally {
      setGrantLoading(false);
    }
  };

  React.useEffect(() => { loadGrantPeriod(); }, [projectId]); // eslint-disable-line react-hooks/exhaustive-deps

  const openGrantModal = (existing?: GrantPeriod) => {
    setGrantName(existing?.name || '');
    setGrantStartDate(existing?.start_date ? existing.start_date.slice(0, 10) : '');
    setGrantEndDate(existing?.end_date ? existing.end_date.slice(0, 10) : '');
    setGrantAutoFreeze(existing?.auto_freeze ?? false);
    setGrantModalVisible(true);
  };

  const saveGrantPeriod = async () => {
    setGrantModalVisible(false);
    try {
      // Backend expects RFC3339 timestamps; append T00:00:00Z if only a date was entered
      const toRFC3339 = (d: string) => d.length === 10 ? `${d}T00:00:00Z` : d;
      await apiClient.setGrantPeriod(projectId, {
        name: grantName,
        start_date: toRFC3339(grantStartDate),
        end_date: toRFC3339(grantEndDate),
        auto_freeze: grantAutoFreeze
      });
      loadGrantPeriod();
    } catch (e: any) {
      setGrantError(e.message || 'Failed to save grant period');
    }
  };

  const handleDeleteGrantPeriod = async () => {
    setGrantDeleteConfirm(false);
    try {
      await apiClient.deleteGrantPeriod(projectId);
      setGrantPeriod(null);
    } catch (e: any) {
      setGrantError(e.message || 'Failed to delete grant period');
    }
  };

  // ── Budget Sharing ────────────────────────────────────────────────────────
  const [shares, setShares] = React.useState<BudgetShareRecord[]>([]);
  const [sharesLoading, setSharesLoading] = React.useState(true);
  const [shareError, setShareError] = React.useState<string | null>(null);
  const [shareModalVisible, setShareModalVisible] = React.useState(false);
  const [shareToProjectId, setShareToProjectId] = React.useState('');
  const [shareToMemberId, setShareToMemberId] = React.useState('');
  const [shareAmount, setShareAmount] = React.useState('');
  const [shareReason, setShareReason] = React.useState('');

  const loadShares = async () => {
    setSharesLoading(true);
    try {
      const result = await apiClient.listProjectBudgetShares(projectId);
      setShares(result);
    } catch (e: any) {
      setShareError(e.message || 'Failed to load budget shares');
    } finally {
      setSharesLoading(false);
    }
  };

  React.useEffect(() => { loadShares(); }, [projectId]); // eslint-disable-line react-hooks/exhaustive-deps

  const shareBudget = async () => {
    setShareModalVisible(false);
    try {
      await apiClient.shareProjectBudget(projectId, {
        from_project_id: projectId,
        to_project_id: shareToProjectId || undefined,
        to_member_id: shareToMemberId || undefined,
        amount: parseFloat(shareAmount) || 0,
        reason: shareReason || undefined
      });
      loadShares();
    } catch (e: any) {
      setShareError(e.message || 'Failed to share budget');
    }
  };

  // ── Onboarding Templates ──────────────────────────────────────────────────
  const [onboardingTemplates, setOnboardingTemplates] = React.useState<OnboardingTemplate[]>([]);
  const [onboardingLoading, setOnboardingLoading] = React.useState(true);
  const [onboardingError, setOnboardingError] = React.useState<string | null>(null);
  const [onboardingModalVisible, setOnboardingModalVisible] = React.useState(false);
  const [onboardingName, setOnboardingName] = React.useState('');
  const [onboardingDescription, setOnboardingDescription] = React.useState('');

  const loadOnboardingTemplates = async () => {
    setOnboardingLoading(true);
    try {
      const result = await apiClient.listOnboardingTemplates(projectId);
      setOnboardingTemplates(result);
    } catch (e: any) {
      setOnboardingError(e.message || 'Failed to load onboarding templates');
    } finally {
      setOnboardingLoading(false);
    }
  };

  React.useEffect(() => { loadOnboardingTemplates(); }, [projectId]); // eslint-disable-line react-hooks/exhaustive-deps

  const addOnboardingTemplate = async () => {
    setOnboardingModalVisible(false);
    try {
      await apiClient.addOnboardingTemplate(projectId, { name: onboardingName, description: onboardingDescription });
      loadOnboardingTemplates();
    } catch (e: any) {
      setOnboardingError(e.message || 'Failed to add onboarding template');
    }
  };

  const deleteOnboardingTemplate = async (nameOrId: string) => {
    try {
      await apiClient.deleteOnboardingTemplate(projectId, nameOrId);
      loadOnboardingTemplates();
    } catch (e: any) {
      setOnboardingError(e.message || 'Failed to delete onboarding template');
    }
  };

  // ── Monthly Report ────────────────────────────────────────────────────────
  const [reportMonth, setReportMonth] = React.useState(() => {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
  });
  const [reportFormat, setReportFormat] = React.useState('text');
  const [reportOutput, setReportOutput] = React.useState<string | null>(null);
  const [reportLoading, setReportLoading] = React.useState(false);
  const [reportError, setReportError] = React.useState<string | null>(null);

  const generateReport = async () => {
    setReportLoading(true);
    setReportError(null);
    setReportOutput(null);
    try {
      const result = await apiClient.getMonthlyReport(projectId, reportMonth, reportFormat);
      setReportOutput(result);
    } catch (e: any) {
      setReportError(e.message || 'Failed to generate report');
    } finally {
      setReportLoading(false);
    }
  };

  const roleColor = (role: string): 'blue' | 'red' | 'grey' | 'green' => {
    switch (role) {
      case 'owner': return 'red';
      case 'admin': return 'blue';
      case 'member': return 'green';
      default: return 'grey';
    }
  };

  return (
    <SpaceBetween size="m" data-testid="governance-panel">
      <Tabs
        data-testid="governance-tabs"
        tabs={[
          // ── Tab 1: Quotas ────────────────────────────────────────────────
          {
            id: 'quotas',
            label: 'Quotas',
            content: (
              <SpaceBetween size="m">
                {quotaError && (
                  <Alert type="error" dismissible onDismiss={() => setQuotaError(null)}>{quotaError}</Alert>
                )}
                <Container
                  header={
                    <Header
                      variant="h3"
                      actions={
                        <Button
                          data-testid="set-quota-button"
                          onClick={openQuotaModal}
                        >
                          Set Quota
                        </Button>
                      }
                    >
                      Role Quotas
                    </Header>
                  }
                >
                  {quotasLoading ? (
                    <Box textAlign="center" padding="l"><Spinner /></Box>
                  ) : (
                    <Table
                      data-testid="quotas-table"
                      columnDefinitions={[
                        {
                          id: 'role',
                          header: 'Role',
                          cell: (item: RoleQuota) => <Badge color={roleColor(item.role)}>{item.role}</Badge>
                        },
                        {
                          id: 'max_instances',
                          header: 'Max Instances',
                          cell: (item: RoleQuota) => item.max_instances === -1 ? 'Unlimited' : String(item.max_instances)
                        },
                        {
                          id: 'max_instance_type',
                          header: 'Max Instance Type',
                          cell: (item: RoleQuota) => item.max_instance_type || 'Unlimited'
                        },
                        {
                          id: 'max_spend_daily',
                          header: 'Max Spend/Day ($)',
                          cell: (item: RoleQuota) => item.max_spend_daily === 0 ? 'Unlimited' : `$${item.max_spend_daily.toFixed(2)}`
                        },
                        {
                          id: 'actions',
                          header: 'Actions',
                          cell: (item: RoleQuota) => (
                            <Button
                              variant="link"
                              data-testid={`delete-quota-${item.role}`}
                              onClick={() => deleteQuota(item.role)}
                            >
                              Delete
                            </Button>
                          )
                        }
                      ]}
                      items={quotas}
                      empty={
                        <Box textAlign="center" color="inherit">
                          <b>No quotas configured</b>
                          <Box variant="p" color="inherit">Set quotas to limit resource usage per role.</Box>
                        </Box>
                      }
                    />
                  )}
                </Container>

                {/* Set Quota Modal */}
                <Modal
                  visible={quotaModalVisible}
                  onDismiss={() => setQuotaModalVisible(false)}
                  header="Set Role Quota"
                  footer={
                    <Box float="right">
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setQuotaModalVisible(false)}>Cancel</Button>
                        <Button variant="primary" data-testid="save-quota-button" onClick={saveQuota}>Save</Button>
                      </SpaceBetween>
                    </Box>
                  }
                >
                  <Form>
                    <SpaceBetween size="m">
                      <FormField label="Role">
                        <Select
                          data-testid="quota-role-select"
                          selectedOption={{ value: quotaRole, label: quotaRole }}
                          onChange={({ detail }) => setQuotaRole(detail.selectedOption.value || 'member')}
                          options={[
                            { value: 'owner', label: 'Owner' },
                            { value: 'admin', label: 'Admin' },
                            { value: 'member', label: 'Member' },
                            { value: 'viewer', label: 'Viewer' }
                          ]}
                        />
                      </FormField>
                      <FormField label="Max Instances (-1 = unlimited)">
                        <Input
                          data-testid="quota-max-instances-input"
                          value={quotaMaxInstances}
                          onChange={({ detail }) => setQuotaMaxInstances(detail.value)}
                          placeholder="-1"
                          type="number"
                        />
                      </FormField>
                      <FormField label="Max Instance Type (prefix, e.g. t3; empty = unlimited)">
                        <Input
                          data-testid="quota-max-instance-type-input"
                          value={quotaMaxInstanceType}
                          onChange={({ detail }) => setQuotaMaxInstanceType(detail.value)}
                          placeholder="t3"
                        />
                      </FormField>
                      <FormField label="Max Daily Spend USD (0 = unlimited)">
                        <Input
                          data-testid="quota-max-spend-daily-input"
                          value={quotaMaxSpendDaily}
                          onChange={({ detail }) => setQuotaMaxSpendDaily(detail.value)}
                          placeholder="0"
                          type="number"
                        />
                      </FormField>
                    </SpaceBetween>
                  </Form>
                </Modal>
              </SpaceBetween>
            )
          },

          // ── Tab 2: Grant Period ──────────────────────────────────────────
          {
            id: 'grant-period',
            label: 'Grant Period',
            content: (
              <SpaceBetween size="m">
                {grantError && (
                  <Alert type="error" dismissible onDismiss={() => setGrantError(null)}>{grantError}</Alert>
                )}
                <Container header={<Header variant="h3">Grant Period</Header>}>
                  {grantLoading ? (
                    <Box textAlign="center" padding="l"><Spinner /></Box>
                  ) : grantPeriod ? (
                    <SpaceBetween size="m">
                      <ColumnLayout columns={2} variant="text-grid" data-testid="grant-period-details">
                        <div>
                          <Box variant="awsui-key-label">Name</Box>
                          <div data-testid="grant-period-name">{grantPeriod.name}</div>
                        </div>
                        <div>
                          <Box variant="awsui-key-label">Start Date</Box>
                          <div data-testid="grant-period-start">{grantPeriod.start_date.slice(0, 10)}</div>
                        </div>
                        <div>
                          <Box variant="awsui-key-label">End Date</Box>
                          <div data-testid="grant-period-end">{grantPeriod.end_date.slice(0, 10)}</div>
                        </div>
                        <div>
                          <Box variant="awsui-key-label">Auto-Freeze</Box>
                          <Badge color={grantPeriod.auto_freeze ? 'blue' : 'grey'} data-testid="grant-auto-freeze-badge">
                            {grantPeriod.auto_freeze ? 'Enabled' : 'Disabled'}
                          </Badge>
                        </div>
                        {grantPeriod.frozen_at && (
                          <div>
                            <Box variant="awsui-key-label">Frozen At</Box>
                            <Badge color="red" data-testid="grant-frozen-badge">
                              {new Date(grantPeriod.frozen_at).toLocaleDateString()}
                            </Badge>
                          </div>
                        )}
                      </ColumnLayout>
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button
                          data-testid="edit-grant-period-button"
                          onClick={() => openGrantModal(grantPeriod)}
                        >
                          Edit
                        </Button>
                        {!grantDeleteConfirm ? (
                          <Button
                            variant="link"
                            data-testid="delete-grant-period-button"
                            onClick={() => setGrantDeleteConfirm(true)}
                          >
                            Delete
                          </Button>
                        ) : (
                          <SpaceBetween direction="horizontal" size="xs">
                            <Button
                              variant="primary"
                              data-testid="confirm-delete-grant-period-button"
                              onClick={handleDeleteGrantPeriod}
                            >
                              Confirm Delete
                            </Button>
                            <Button variant="link" onClick={() => setGrantDeleteConfirm(false)}>Cancel</Button>
                          </SpaceBetween>
                        )}
                      </SpaceBetween>
                    </SpaceBetween>
                  ) : (
                    <SpaceBetween size="m">
                      <Alert type="info" data-testid="no-grant-period-alert">
                        No grant period configured for this project.
                      </Alert>
                      <Button
                        data-testid="configure-grant-period-button"
                        onClick={() => openGrantModal()}
                      >
                        Configure Grant Period
                      </Button>
                    </SpaceBetween>
                  )}
                </Container>

                {/* Grant Period Modal */}
                <Modal
                  visible={grantModalVisible}
                  onDismiss={() => setGrantModalVisible(false)}
                  header={grantPeriod ? 'Edit Grant Period' : 'Configure Grant Period'}
                  footer={
                    <Box float="right">
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setGrantModalVisible(false)}>Cancel</Button>
                        <Button variant="primary" data-testid="save-grant-period-button" onClick={saveGrantPeriod}>Save</Button>
                      </SpaceBetween>
                    </Box>
                  }
                >
                  <Form>
                    <SpaceBetween size="m">
                      <FormField label="Grant Period Name">
                        <Input
                          data-testid="grant-period-name-input"
                          value={grantName}
                          onChange={({ detail }) => setGrantName(detail.value)}
                          placeholder="e.g. NSF Year 1"
                        />
                      </FormField>
                      <FormField label="Start Date (YYYY-MM-DD)">
                        <Input
                          data-testid="grant-period-start-input"
                          value={grantStartDate}
                          onChange={({ detail }) => setGrantStartDate(detail.value)}
                          placeholder="2024-01-01"
                        />
                      </FormField>
                      <FormField label="End Date (YYYY-MM-DD)">
                        <Input
                          data-testid="grant-period-end-input"
                          value={grantEndDate}
                          onChange={({ detail }) => setGrantEndDate(detail.value)}
                          placeholder="2024-12-31"
                        />
                      </FormField>
                      <FormField label="Auto-Freeze when period ends">
                        <Toggle
                          data-testid="grant-auto-freeze-toggle"
                          checked={grantAutoFreeze}
                          onChange={({ detail }) => setGrantAutoFreeze(detail.checked)}
                        >
                          Auto-Freeze
                        </Toggle>
                      </FormField>
                    </SpaceBetween>
                  </Form>
                </Modal>
              </SpaceBetween>
            )
          },

          // ── Tab 3: Budget Sharing ────────────────────────────────────────
          {
            id: 'budget-sharing',
            label: 'Budget Sharing',
            content: (
              <SpaceBetween size="m">
                {shareError && (
                  <Alert type="error" dismissible onDismiss={() => setShareError(null)}>{shareError}</Alert>
                )}
                <Container
                  header={
                    <Header
                      variant="h3"
                      actions={
                        <Button
                          data-testid="share-budget-button"
                          onClick={() => {
                            setShareToProjectId('');
                            setShareToMemberId('');
                            setShareAmount('');
                            setShareReason('');
                            setShareModalVisible(true);
                          }}
                        >
                          Share Budget
                        </Button>
                      }
                    >
                      Budget Shares
                    </Header>
                  }
                >
                  {sharesLoading ? (
                    <Box textAlign="center" padding="l"><Spinner /></Box>
                  ) : (
                    <Table
                      data-testid="budget-shares-table"
                      columnDefinitions={[
                        {
                          id: 'to',
                          header: 'To Project / Member',
                          cell: (item: BudgetShareRecord) =>
                            item.request.to_project_id || item.request.to_member_id || '—'
                        },
                        {
                          id: 'amount',
                          header: 'Amount',
                          cell: (item: BudgetShareRecord) => `$${item.request.amount.toFixed(2)}`
                        },
                        {
                          id: 'reason',
                          header: 'Reason',
                          cell: (item: BudgetShareRecord) => item.request.reason || '—'
                        },
                        {
                          id: 'created_at',
                          header: 'Created',
                          cell: (item: BudgetShareRecord) => new Date(item.created_at).toLocaleDateString()
                        },
                        {
                          id: 'expires_at',
                          header: 'Expires',
                          cell: (item: BudgetShareRecord) => item.expires_at ? new Date(item.expires_at).toLocaleDateString() : '—'
                        }
                      ]}
                      items={shares}
                      empty={
                        <Box textAlign="center" color="inherit">
                          <b>No budget shares</b>
                          <Box variant="p" color="inherit">No budget has been shared from this project.</Box>
                        </Box>
                      }
                    />
                  )}
                </Container>

                {/* Share Budget Modal */}
                <Modal
                  visible={shareModalVisible}
                  onDismiss={() => setShareModalVisible(false)}
                  header="Share Budget"
                  footer={
                    <Box float="right">
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setShareModalVisible(false)}>Cancel</Button>
                        <Button variant="primary" data-testid="confirm-share-budget-button" onClick={shareBudget}>Share</Button>
                      </SpaceBetween>
                    </Box>
                  }
                >
                  <Form>
                    <SpaceBetween size="m">
                      <FormField label="To Project ID (optional)">
                        <Input
                          data-testid="share-to-project-input"
                          value={shareToProjectId}
                          onChange={({ detail }) => setShareToProjectId(detail.value)}
                          placeholder="project-id"
                        />
                      </FormField>
                      <FormField label="To Member ID (optional)">
                        <Input
                          data-testid="share-to-member-input"
                          value={shareToMemberId}
                          onChange={({ detail }) => setShareToMemberId(detail.value)}
                          placeholder="user-id"
                        />
                      </FormField>
                      <FormField label="Amount (USD)" constraintText="Required">
                        <Input
                          data-testid="share-amount-input"
                          value={shareAmount}
                          onChange={({ detail }) => setShareAmount(detail.value)}
                          placeholder="100.00"
                          type="number"
                        />
                      </FormField>
                      <FormField label="Reason (optional)">
                        <Input
                          data-testid="share-reason-input"
                          value={shareReason}
                          onChange={({ detail }) => setShareReason(detail.value)}
                          placeholder="Research allocation"
                        />
                      </FormField>
                    </SpaceBetween>
                  </Form>
                </Modal>
              </SpaceBetween>
            )
          },

          // ── Tab 4: Onboarding Templates ──────────────────────────────────
          {
            id: 'onboarding-templates',
            label: 'Onboarding Templates',
            content: (
              <SpaceBetween size="m">
                {onboardingError && (
                  <Alert type="error" dismissible onDismiss={() => setOnboardingError(null)}>{onboardingError}</Alert>
                )}
                <Container
                  header={
                    <Header
                      variant="h3"
                      actions={
                        <Button
                          data-testid="add-onboarding-template-button"
                          onClick={() => {
                            setOnboardingName('');
                            setOnboardingDescription('');
                            setOnboardingModalVisible(true);
                          }}
                        >
                          Add Template
                        </Button>
                      }
                    >
                      Onboarding Templates
                    </Header>
                  }
                >
                  {onboardingLoading ? (
                    <Box textAlign="center" padding="l"><Spinner /></Box>
                  ) : (
                    <Table
                      data-testid="onboarding-templates-table"
                      columnDefinitions={[
                        {
                          id: 'name',
                          header: 'Name',
                          cell: (item: OnboardingTemplate) => item.name
                        },
                        {
                          id: 'description',
                          header: 'Description',
                          cell: (item: OnboardingTemplate) => item.description || '—'
                        },
                        {
                          id: 'actions',
                          header: 'Actions',
                          cell: (item: OnboardingTemplate) => (
                            <Button
                              variant="link"
                              data-testid={`delete-onboarding-template-${item.name}`}
                              onClick={() => deleteOnboardingTemplate(item.id || item.name)}
                            >
                              Delete
                            </Button>
                          )
                        }
                      ]}
                      items={onboardingTemplates}
                      empty={
                        <Box textAlign="center" color="inherit">
                          <b>No onboarding templates</b>
                          <Box variant="p" color="inherit">Add templates to automatically provision resources for new members.</Box>
                        </Box>
                      }
                    />
                  )}
                </Container>

                {/* Add Onboarding Template Modal */}
                <Modal
                  visible={onboardingModalVisible}
                  onDismiss={() => setOnboardingModalVisible(false)}
                  header="Add Onboarding Template"
                  footer={
                    <Box float="right">
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setOnboardingModalVisible(false)}>Cancel</Button>
                        <Button variant="primary" data-testid="save-onboarding-template-button" onClick={addOnboardingTemplate}>Save</Button>
                      </SpaceBetween>
                    </Box>
                  }
                >
                  <Form>
                    <SpaceBetween size="m">
                      <FormField label="Template Name" constraintText="Required">
                        <Input
                          data-testid="onboarding-template-name-input"
                          value={onboardingName}
                          onChange={({ detail }) => setOnboardingName(detail.value)}
                          placeholder="My Onboarding Template"
                        />
                      </FormField>
                      <FormField label="Description (optional)">
                        <Textarea
                          data-testid="onboarding-template-description-input"
                          value={onboardingDescription}
                          onChange={({ detail }) => setOnboardingDescription(detail.value)}
                          placeholder="Describe what this template provisions..."
                          rows={3}
                        />
                      </FormField>
                    </SpaceBetween>
                  </Form>
                </Modal>
              </SpaceBetween>
            )
          },

          // ── Tab 5: Monthly Report ────────────────────────────────────────
          {
            id: 'monthly-report',
            label: 'Monthly Report',
            content: (
              <SpaceBetween size="m">
                {reportError && (
                  <Alert type="error" dismissible onDismiss={() => setReportError(null)}>{reportError}</Alert>
                )}
                <Container header={<Header variant="h3">Generate Monthly Report</Header>}>
                  <SpaceBetween size="m">
                    <SpaceBetween direction="horizontal" size="m">
                      <FormField label="Month (YYYY-MM)">
                        <Input
                          data-testid="report-month-input"
                          value={reportMonth}
                          onChange={({ detail }) => setReportMonth(detail.value)}
                          placeholder="2024-01"
                        />
                      </FormField>
                      <FormField label="Format">
                        <Select
                          data-testid="report-format-select"
                          selectedOption={{ value: reportFormat, label: reportFormat }}
                          onChange={({ detail }) => setReportFormat(detail.selectedOption.value || 'text')}
                          options={[
                            { value: 'text', label: 'Text' },
                            { value: 'csv', label: 'CSV' },
                            { value: 'json', label: 'JSON' }
                          ]}
                        />
                      </FormField>
                      <FormField label=" ">
                        <Button
                          variant="primary"
                          data-testid="generate-report-button"
                          onClick={generateReport}
                          loading={reportLoading}
                        >
                          Generate
                        </Button>
                      </FormField>
                    </SpaceBetween>

                    {reportLoading && (
                      <Box textAlign="center" padding="l"><Spinner /></Box>
                    )}

                    {reportOutput !== null && !reportLoading && (
                      <Container header={<Header variant="h3">Report Output</Header>}>
                        <pre
                          data-testid="monthly-report-output"
                          style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', maxHeight: '500px', overflowY: 'auto', fontFamily: 'monospace', fontSize: '12px' }}
                        >
                          {reportOutput}
                        </pre>
                      </Container>
                    )}
                  </SpaceBetween>
                </Container>
              </SpaceBetween>
            )
          }
        ]}
      />
    </SpaceBetween>
  );
};
