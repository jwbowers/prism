/**
 * WorkshopsPanel — v0.18.0
 *
 * 3-tab workshop management panel:
 *   Workshops | Dashboard | Config Templates
 *
 * Rendered by WorkshopsManagementView in App.tsx when activeView === 'workshops'.
 * Uses window.__apiClient (set in App.tsx).
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Tabs,
  Header,
  SpaceBetween,
  Button,
  Table,
  Modal,
  Form,
  FormField,
  Input,
  Box,
  Spinner,
  Badge,
  Alert,
  ColumnLayout,
  ProgressBar,
  DatePicker,
  StatusIndicator
} from '../lib/cloudscape-shim';

// ── Local types (mirror App.tsx interfaces) ──────────────────────────────────

interface WorkshopParticipant {
  user_id: string;
  email?: string;
  display_name?: string;
  joined_at: string;
  instance_id?: string;
  instance_name?: string;
  status: string;
  progress?: number;
}

interface WorkshopEvent {
  id: string;
  title: string;
  description?: string;
  owner: string;
  template: string;
  max_participants: number;
  budget_per_participant?: number;
  start_time: string;
  end_time: string;
  early_access_hours?: number;
  status: string;
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
  description?: string;
  created_at: string;
}

// ── API helper ────────────────────────────────────────────────────────────────

function getAPI(): any {
  return (window as any).__apiClient;
}

// ── Status badge ──────────────────────────────────────────────────────────────

function statusBadge(status: string) {
  switch (status) {
    case 'active':   return <Badge color="green">Active</Badge>;
    case 'draft':    return <Badge color="blue">Draft</Badge>;
    case 'ended':    return <Badge color="grey">Ended</Badge>;
    case 'archived': return <Badge color="grey">Archived</Badge>;
    default:         return <Badge color="grey">{status}</Badge>;
  }
}

// ── WorkshopsPanel ────────────────────────────────────────────────────────────

export function WorkshopsPanel() {
  const [activeTab, setActiveTab] = useState('workshops');
  const [selectedWorkshop, setSelectedWorkshop] = useState<WorkshopEvent | null>(null);
  const [dashboard, setDashboard] = useState<WorkshopDashboard | null>(null);
  const [dashboardLoading, setDashboardLoading] = useState(false);

  const handleSelectWorkshop = useCallback(async (ws: WorkshopEvent) => {
    setSelectedWorkshop(ws);
    setActiveTab('dashboard');
    setDashboardLoading(true);
    try {
      const api = getAPI();
      const dash = await api.getWorkshopDashboard(ws.id);
      setDashboard(dash);
    } catch (err) {
      console.error('Failed to load dashboard:', err);
    } finally {
      setDashboardLoading(false);
    }
  }, []);

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="Manage time-bounded workshops, tutorials, and hackathons"
        actions={
          <Button
            variant="primary"
            onClick={() => setActiveTab('workshops')}
          >
            Workshops
          </Button>
        }
      >
        Workshop &amp; Event Management
      </Header>

      <Tabs
        activeTabId={activeTab}
        onChange={({ detail }) => setActiveTab(detail.activeTabId)}
        tabs={[
          {
            id: 'workshops',
            label: 'Workshops',
            content: (
              <WorkshopsListTab
                onSelectWorkshop={handleSelectWorkshop}
              />
            ),
          },
          {
            id: 'dashboard',
            label: selectedWorkshop ? `Dashboard — ${selectedWorkshop.title}` : 'Dashboard',
            content: (
              <DashboardTab
                workshop={selectedWorkshop}
                dashboard={dashboard}
                loading={dashboardLoading}
                onRefresh={selectedWorkshop ? () => handleSelectWorkshop(selectedWorkshop) : undefined}
              />
            ),
          },
          {
            id: 'configs',
            label: 'Config Templates',
            content: <ConfigsTab />,
          },
        ]}
      />
    </SpaceBetween>
  );
}

// ── Workshops List Tab ────────────────────────────────────────────────────────

function WorkshopsListTab({ onSelectWorkshop }: { onSelectWorkshop: (ws: WorkshopEvent) => void }) {
  const [workshops, setWorkshops] = useState<WorkshopEvent[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showEndConfirm, setShowEndConfirm] = useState(false);
  const [selectedItems, setSelectedItems] = useState<WorkshopEvent[]>([]);
  const [notification, setNotification] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const api = getAPI();
      const data = await api.getWorkshops();
      setWorkshops(data);
    } catch (err: any) {
      setError(err?.message || 'Failed to load workshops');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleDelete = async (id: string) => {
    try {
      await getAPI().deleteWorkshop(id);
      setNotification('Workshop deleted.');
      await load();
    } catch (err: any) {
      setError(err?.message || 'Failed to delete workshop');
    }
  };

  const handleProvision = async (id: string) => {
    try {
      const result = await getAPI().provisionWorkshop(id);
      setNotification(`Provisioned ${result.provisioned} workspaces (${result.skipped} skipped).`);
      await load();
    } catch (err: any) {
      setError(err?.message || 'Failed to provision workshop');
    }
  };

  const handleEnd = async () => {
    if (!selectedItems[0]) return;
    try {
      const result = await getAPI().endWorkshop(selectedItems[0].id);
      setShowEndConfirm(false);
      setNotification(`Workshop ended. Stopped ${result.stopped} instances.`);
      await load();
    } catch (err: any) {
      setError(err?.message || 'Failed to end workshop');
    }
  };

  return (
    <SpaceBetween size="m">
      {notification && (
        <Alert type="success" onDismiss={() => setNotification(null)}>
          {notification}
        </Alert>
      )}
      {error && (
        <Alert type="error" onDismiss={() => setError(null)}>
          {error}
        </Alert>
      )}

      <Table
        data-testid="workshops-table"
        loading={loading}
        loadingText="Loading workshops..."
        items={workshops}
        selectionType="single"
        selectedItems={selectedItems}
        onSelectionChange={({ detail }) => setSelectedItems(detail.selectedItems as WorkshopEvent[])}
        columnDefinitions={[
          {
            id: 'title',
            header: 'Title',
            cell: (item) => (
              <Button variant="link" onClick={() => onSelectWorkshop(item)}>
                {item.title}
              </Button>
            ),
            sortingField: 'title',
          },
          {
            id: 'status',
            header: 'Status',
            cell: (item) => statusBadge(item.status),
          },
          {
            id: 'template',
            header: 'Template',
            cell: (item) => item.template,
          },
          {
            id: 'start',
            header: 'Start',
            cell: (item) => item.start_time?.substring(0, 16) || '—',
          },
          {
            id: 'end',
            header: 'Ends At',
            cell: (item) => item.end_time?.substring(0, 16) || '—',
          },
          {
            id: 'participants',
            header: 'Participants',
            cell: (item) => {
              const count = item.participants?.length ?? 0;
              const max = item.max_participants;
              return max > 0 ? `${count} / ${max}` : String(count);
            },
          },
          {
            id: 'join_token',
            header: 'Join Token',
            cell: (item) => item.join_token || '—',
          },
          {
            id: 'actions',
            header: 'Actions',
            cell: (item) => (
              <SpaceBetween direction="horizontal" size="xs">
                <Button
                  variant="inline-link"
                  onClick={() => handleProvision(item.id)}
                  disabled={item.status === 'ended' || item.status === 'archived'}
                >
                  Provision
                </Button>
                <Button
                  variant="inline-link"
                  onClick={() => { setSelectedItems([item]); setShowEndConfirm(true); }}
                  disabled={item.status === 'ended' || item.status === 'archived'}
                >
                  End
                </Button>
                <Button
                  variant="inline-link"
                  onClick={() => handleDelete(item.id)}
                >
                  Delete
                </Button>
              </SpaceBetween>
            ),
          },
        ]}
        header={
          <Header
            counter={workshops.length > 0 ? `(${workshops.length})` : undefined}
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Button onClick={load} iconName="refresh">Refresh</Button>
                <Button data-testid="create-workshop-button" variant="primary" onClick={() => setShowCreate(true)}>
                  Create Workshop
                </Button>
              </SpaceBetween>
            }
          >
            Workshops
          </Header>
        }
        empty={
          <Box textAlign="center" color="inherit">
            <b>No workshops</b>
            <Box padding={{ bottom: 's' }} variant="p" color="inherit">
              Create a workshop to get started.
            </Box>
            <Button onClick={() => setShowCreate(true)}>Create Workshop</Button>
          </Box>
        }
      />

      {/* End confirmation modal */}
      <Modal
        visible={showEndConfirm}
        onDismiss={() => setShowEndConfirm(false)}
        header="End Workshop"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setShowEndConfirm(false)}>Cancel</Button>
              <Button variant="primary" onClick={handleEnd}>End Workshop</Button>
            </SpaceBetween>
          </Box>
        }
      >
        <p>
          Ending <strong>{selectedItems[0]?.title}</strong> will stop all participant instances.
          This action cannot be undone. Proceed?
        </p>
      </Modal>

      {showCreate && (
        <CreateWorkshopModal
          onDismiss={() => setShowCreate(false)}
          onCreated={async () => { setShowCreate(false); await load(); }}
        />
      )}
    </SpaceBetween>
  );
}

// ── Create Workshop Modal ─────────────────────────────────────────────────────

function CreateWorkshopModal({
  onDismiss,
  onCreated,
}: {
  onDismiss: () => void;
  onCreated: () => void;
}) {
  const [title, setTitle] = useState('');
  const [template, setTemplate] = useState('');
  const [owner, setOwner] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [maxPax, setMaxPax] = useState('');
  const [budget, setBudget] = useState('');
  const [earlyAccess, setEarlyAccess] = useState('');
  const [description, setDescription] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreate = async () => {
    setError(null);
    if (!title || !template || !owner || !startDate || !endDate) {
      setError('Title, template, owner, start time, and end time are required.');
      return;
    }
    setSaving(true);
    try {
      await getAPI().createWorkshop({
        title,
        template,
        owner,
        start_time: new Date(startDate).toISOString(),
        end_time: new Date(endDate).toISOString(),
        max_participants: maxPax ? parseInt(maxPax, 10) : 0,
        budget_per_participant: budget ? parseFloat(budget) : undefined,
        early_access_hours: earlyAccess ? parseInt(earlyAccess, 10) : undefined,
        description: description || undefined,
      });
      onCreated();
    } catch (err: any) {
      setError(err?.message || 'Failed to create workshop');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal
      visible
      onDismiss={onDismiss}
      header="Create Workshop"
      size="large"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={onDismiss}>Cancel</Button>
            <Button variant="primary" loading={saving} onClick={handleCreate}>
              Create
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {error && <Alert type="error">{error}</Alert>}
        <Form>
          <SpaceBetween size="m">
            <FormField label="Title" constraintText="Required">
              <Input value={title} onChange={({ detail }) => setTitle(detail.value)} placeholder="NeurIPS DL Tutorial" />
            </FormField>
            <FormField label="Template" constraintText="Required">
              <Input value={template} onChange={({ detail }) => setTemplate(detail.value)} placeholder="pytorch-ml" />
            </FormField>
            <FormField label="Organizer (Owner)" constraintText="Required">
              <Input value={owner} onChange={({ detail }) => setOwner(detail.value)} placeholder="organizer-user-id" />
            </FormField>
            <ColumnLayout columns={2}>
              <FormField label="Start Time" constraintText="Required (YYYY-MM-DD)">
                <DatePicker
                  value={startDate}
                  onChange={({ detail }) => setStartDate(detail.value)}
                  placeholder="2026-12-08"
                  openCalendarAriaLabel={(selectedDate) =>
                    'Choose start date' + (selectedDate ? `, selected date is ${selectedDate}` : '')
                  }
                />
              </FormField>
              <FormField label="End Time" constraintText="Required (YYYY-MM-DD)">
                <DatePicker
                  value={endDate}
                  onChange={({ detail }) => setEndDate(detail.value)}
                  placeholder="2026-12-09"
                  openCalendarAriaLabel={(selectedDate) =>
                    'Choose end date' + (selectedDate ? `, selected date is ${selectedDate}` : '')
                  }
                />
              </FormField>
            </ColumnLayout>
            <ColumnLayout columns={3}>
              <FormField label="Max Participants" constraintText="0 = unlimited">
                <Input value={maxPax} onChange={({ detail }) => setMaxPax(detail.value)} type="number" placeholder="60" />
              </FormField>
              <FormField label="Budget per Participant ($)">
                <Input value={budget} onChange={({ detail }) => setBudget(detail.value)} type="number" placeholder="5.00" />
              </FormField>
              <FormField label="Early Access (hours)">
                <Input value={earlyAccess} onChange={({ detail }) => setEarlyAccess(detail.value)} type="number" placeholder="24" />
              </FormField>
            </ColumnLayout>
            <FormField label="Description">
              <Input value={description} onChange={({ detail }) => setDescription(detail.value)} placeholder="Optional description" />
            </FormField>
          </SpaceBetween>
        </Form>
      </SpaceBetween>
    </Modal>
  );
}

// ── Dashboard Tab ─────────────────────────────────────────────────────────────

function DashboardTab({
  workshop,
  dashboard,
  loading,
  onRefresh,
}: {
  workshop: WorkshopEvent | null;
  dashboard: WorkshopDashboard | null;
  loading: boolean;
  onRefresh?: () => void;
}) {
  if (!workshop) {
    return (
      <Box textAlign="center" padding="xl">
        <StatusIndicator type="info">
          Select a workshop from the Workshops tab to view its dashboard.
        </StatusIndicator>
      </Box>
    );
  }

  if (loading) {
    return (
      <Box textAlign="center" padding="xl">
        <Spinner size="large" />
      </Box>
    );
  }

  if (!dashboard) {
    return (
      <Alert type="error">
        Failed to load dashboard for {workshop.title}.
      </Alert>
    );
  }

  const total = dashboard.total_participants || 1; // avoid /0

  return (
    <SpaceBetween size="l">
      <Header
        variant="h2"
        actions={onRefresh && <Button onClick={onRefresh} iconName="refresh">Refresh</Button>}
      >
        {dashboard.title} {statusBadge(dashboard.status)}
      </Header>

      <ColumnLayout columns={4} borders="vertical">
        <SpaceBetween size="xs">
          <Box variant="awsui-key-label">Participants</Box>
          <Box variant="h2">{dashboard.total_participants}</Box>
        </SpaceBetween>
        <SpaceBetween size="xs">
          <Box variant="awsui-key-label">Active Instances</Box>
          <Box variant="h2"><Badge color="green">{dashboard.active_instances}</Badge></Box>
        </SpaceBetween>
        <SpaceBetween size="xs">
          <Box variant="awsui-key-label">Time Remaining</Box>
          <Box variant="h2">{dashboard.time_remaining}</Box>
        </SpaceBetween>
        <SpaceBetween size="xs">
          <Box variant="awsui-key-label">Total Spent</Box>
          <Box variant="h2">${(dashboard.total_spent || 0).toFixed(2)}</Box>
        </SpaceBetween>
      </ColumnLayout>

      <ProgressBar
        value={Math.round((dashboard.active_instances / total) * 100)}
        label="Active instances"
        description={`${dashboard.active_instances} active, ${dashboard.stopped_instances} stopped, ${dashboard.pending_instances} pending`}
      />

      <Table
        data-testid="participants-table"
        items={dashboard.participants || []}
        columnDefinitions={[
          { id: 'user_id', header: 'User ID', cell: (p) => p.user_id },
          { id: 'display_name', header: 'Name', cell: (p) => p.display_name || '—' },
          {
            id: 'status',
            header: 'Status',
            cell: (p) => {
              switch (p.status) {
                case 'running':     return <StatusIndicator type="success">Running</StatusIndicator>;
                case 'stopped':     return <StatusIndicator type="stopped">Stopped</StatusIndicator>;
                case 'provisioned': return <StatusIndicator type="in-progress">Provisioned</StatusIndicator>;
                default:            return <StatusIndicator type="pending">Pending</StatusIndicator>;
              }
            },
          },
          { id: 'instance', header: 'Instance', cell: (p) => p.instance_name || '—' },
          {
            id: 'progress',
            header: 'Progress',
            cell: (p) =>
              p.progress !== undefined ? (
                <ProgressBar value={p.progress} />
              ) : '—',
          },
        ]}
        header={<Header counter={`(${dashboard.participants?.length ?? 0})`}>Participants</Header>}
        empty={<Box textAlign="center">No participants yet.</Box>}
      />
    </SpaceBetween>
  );
}

// ── Config Templates Tab ──────────────────────────────────────────────────────

function ConfigsTab() {
  const [configs, setConfigs] = useState<WorkshopConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notification, setNotification] = useState<string | null>(null);
  const [showUse, setShowUse] = useState(false);
  const [selectedConfig, setSelectedConfig] = useState<WorkshopConfig | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getAPI().getWorkshopConfigs();
      setConfigs(data);
    } catch (err: any) {
      setError(err?.message || 'Failed to load configs');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  return (
    <SpaceBetween size="m">
      {notification && (
        <Alert type="success" onDismiss={() => setNotification(null)}>
          {notification}
        </Alert>
      )}
      {error && (
        <Alert type="error" onDismiss={() => setError(null)}>
          {error}
        </Alert>
      )}

      <Table
        data-testid="workshop-configs-table"
        loading={loading}
        loadingText="Loading configs..."
        items={configs}
        columnDefinitions={[
          { id: 'name', header: 'Name', cell: (c) => <strong>{c.name}</strong> },
          { id: 'template', header: 'Template', cell: (c) => c.template },
          { id: 'duration', header: 'Duration', cell: (c) => `${c.duration_hours}h` },
          { id: 'max_pax', header: 'Max Participants', cell: (c) => c.max_participants || 'Unlimited' },
          {
            id: 'budget',
            header: 'Budget / Participant',
            cell: (c) => c.budget_per_participant ? `$${c.budget_per_participant.toFixed(2)}` : '—',
          },
          {
            id: 'actions',
            header: 'Actions',
            cell: (c) => (
              <Button
                variant="inline-link"
                onClick={() => { setSelectedConfig(c); setShowUse(true); }}
              >
                Use Config
              </Button>
            ),
          },
        ]}
        header={
          <Header
            counter={configs.length > 0 ? `(${configs.length})` : undefined}
            actions={<Button onClick={load} iconName="refresh">Refresh</Button>}
          >
            Saved Workshop Configs
          </Header>
        }
        empty={
          <Box textAlign="center" color="inherit">
            <b>No configs saved</b>
            <Box padding={{ bottom: 's' }} variant="p" color="inherit">
              Use <em>prism workshop config save</em> to save a workshop as a reusable config.
            </Box>
          </Box>
        }
      />

      {showUse && selectedConfig && (
        <UseConfigModal
          config={selectedConfig}
          onDismiss={() => setShowUse(false)}
          onCreated={() => { setShowUse(false); setNotification(`Workshop created from config "${selectedConfig.name}".`); }}
        />
      )}
    </SpaceBetween>
  );
}

// ── Use Config Modal ──────────────────────────────────────────────────────────

function UseConfigModal({
  config,
  onDismiss,
  onCreated,
}: {
  config: WorkshopConfig;
  onDismiss: () => void;
  onCreated: () => void;
}) {
  const [title, setTitle] = useState('');
  const [owner, setOwner] = useState('');
  const [startDate, setStartDate] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreate = async () => {
    setError(null);
    if (!title || !owner || !startDate) {
      setError('Title, owner, and start time are required.');
      return;
    }
    setSaving(true);
    try {
      const start = new Date(startDate);
      const end = new Date(start.getTime() + config.duration_hours * 3600 * 1000);
      await getAPI().createWorkshopFromConfig(config.name, {
        title,
        owner,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
      });
      onCreated();
    } catch (err: any) {
      setError(err?.message || 'Failed to create workshop from config');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal
      visible
      onDismiss={onDismiss}
      header={`Create from Config: ${config.name}`}
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={onDismiss}>Cancel</Button>
            <Button variant="primary" loading={saving} onClick={handleCreate}>Create</Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {error && <Alert type="error">{error}</Alert>}
        <Alert type="info">
          Template: <strong>{config.template}</strong> · Duration: <strong>{config.duration_hours}h</strong>
          {config.max_participants > 0 && ` · Max ${config.max_participants} participants`}
        </Alert>
        <Form>
          <SpaceBetween size="m">
            <FormField label="Workshop Title" constraintText="Required">
              <Input value={title} onChange={({ detail }) => setTitle(detail.value)} placeholder="Spring ML Workshop" />
            </FormField>
            <FormField label="Organizer (Owner)" constraintText="Required">
              <Input value={owner} onChange={({ detail }) => setOwner(detail.value)} placeholder="organizer-user-id" />
            </FormField>
            <FormField label="Start Date" constraintText="Required">
              <DatePicker
                value={startDate}
                onChange={({ detail }) => setStartDate(detail.value)}
                placeholder="2027-01-10"
                openCalendarAriaLabel={(selectedDate) =>
                  'Choose start date' + (selectedDate ? `, selected date is ${selectedDate}` : '')
                }
              />
            </FormField>
          </SpaceBetween>
        </Form>
      </SpaceBetween>
    </Modal>
  );
}
