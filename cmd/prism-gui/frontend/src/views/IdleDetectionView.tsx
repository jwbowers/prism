import { useState } from 'react'
import {
  SpaceBetween,
  Header,
  Container,
  Button,
  Box,
  Table,
  ColumnLayout,
  Link,
  Spinner,
  Alert,
  Badge,
  Tabs,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import type { IdlePolicy, IdleSchedule } from '../lib/types'

interface IdleDetectionViewProps {
  idlePolicies: IdlePolicy[]
  idleSchedules: IdleSchedule[]
  loading: boolean
  onRefresh: () => void
}

function getActionBadgeColor(action: string): string {
  switch (action) {
    case 'hibernate': return 'green'
    case 'stop': return 'blue'
    case 'notify': return 'grey'
    default: return 'grey'
  }
}

export function IdleDetectionView({
  idlePolicies,
  idleSchedules,
  loading,
  onRefresh,
}: IdleDetectionViewProps) {
  const [selectedTab, setSelectedTab] = useState<'policies' | 'schedules'>('policies')
  const [selectedPolicy, setSelectedPolicy] = useState<IdlePolicy | null>(null)

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="Automatic cost optimization through idle detection and hibernation"
        actions={
          <Button onClick={onRefresh} disabled={loading}>
            {loading ? <Spinner /> : 'Refresh'}
          </Button>
        }
      >
        Idle Detection & Hibernation
      </Header>

      <ColumnLayout columns={4} variant="text-grid">
        <Container header={<Header variant="h3">Active Policies</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-info">
            {idlePolicies.filter(p => p.enabled).length}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Total Policies</Header>}>
          <Box fontSize="display-l" fontWeight="bold">
            {idlePolicies.length}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Monitored Workspaces</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-success">
            {idleSchedules.filter(s => s.enabled).length}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Cost Savings</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-success">
            ~40%
          </Box>
          <Box variant="small" color="text-body-secondary">
            Through hibernation
          </Box>
        </Container>
      </ColumnLayout>

      <Tabs
        activeTabId={selectedTab}
        onChange={({ detail }) => setSelectedTab(detail.activeTabId as 'policies' | 'schedules')}
        tabs={[
          {
            id: 'policies',
            label: 'Idle Policies',
            content: (
              <Container>
                <Table
                  data-testid="idle-policies-table"
                  columnDefinitions={[
                    {
                      id: 'name',
                      header: 'Policy Name',
                      cell: (item: IdlePolicy) => <Link onFollow={() => setSelectedPolicy(item)}>{item.name}</Link>,
                      sortingField: 'name'
                    },
                    {
                      id: 'idle_minutes',
                      header: 'Idle Threshold',
                      cell: (item: IdlePolicy) => `${item.idle_minutes} minutes`,
                      sortingField: 'idle_minutes'
                    },
                    {
                      id: 'action',
                      header: 'Action',
                      cell: (item: IdlePolicy) => (
                        <Badge color={getActionBadgeColor(item.action)}>
                          {item.action.toUpperCase()}
                        </Badge>
                      )
                    },
                    {
                      id: 'thresholds',
                      header: 'Thresholds',
                      cell: (item: IdlePolicy) => (
                        <Box variant="small">
                          CPU: {item.cpu_threshold}%, Mem: {item.memory_threshold}%, Net: {item.network_threshold} Mbps
                        </Box>
                      )
                    },
                    {
                      id: 'enabled',
                      header: 'Status',
                      cell: (item: IdlePolicy) => (
                        <StatusIndicator type={item.enabled ? 'success' : 'stopped'}>
                          {item.enabled ? 'Enabled' : 'Disabled'}
                        </StatusIndicator>
                      )
                    }
                  ]}
                  items={idlePolicies}
                  loadingText="Loading idle policies..."
                  loading={loading}
                  trackBy="id"
                  empty={
                    <Box textAlign="center" padding="xl">
                      <Box variant="strong">No idle policies configured</Box>
                      <Box variant="p" color="text-body-secondary">
                        Idle policies automatically hibernate or stop workspaces when they're not being used.
                      </Box>
                    </Box>
                  }
                  sortingDisabled={false}
                />
              </Container>
            )
          },
          {
            id: 'schedules',
            label: 'Workspace Schedules',
            content: (
              <Container>
                <Table
                  columnDefinitions={[
                    {
                      id: 'instance',
                      header: 'Workspace',
                      cell: (item: IdleSchedule) => item.instance_name,
                      sortingField: 'instance_name'
                    },
                    {
                      id: 'policy',
                      header: 'Policy',
                      cell: (item: IdleSchedule) => <Badge>{item.policy_name}</Badge>
                    },
                    {
                      id: 'idle_minutes',
                      header: 'Current Idle Time',
                      cell: (item: IdleSchedule) => `${item.idle_minutes} minutes`,
                      sortingField: 'idle_minutes'
                    },
                    {
                      id: 'status',
                      header: 'Status',
                      cell: (item: IdleSchedule) => item.status || 'Active'
                    },
                    {
                      id: 'last_checked',
                      header: 'Last Checked',
                      cell: (item: IdleSchedule) => item.last_checked ? new Date(item.last_checked).toLocaleString() : 'Never'
                    },
                    {
                      id: 'enabled',
                      header: 'Monitoring',
                      cell: (item: IdleSchedule) => (
                        <StatusIndicator type={item.enabled ? 'success' : 'stopped'}>
                          {item.enabled ? 'Enabled' : 'Disabled'}
                        </StatusIndicator>
                      )
                    }
                  ]}
                  items={idleSchedules}
                  loadingText="Loading workspace schedules..."
                  loading={loading}
                  trackBy="instance_name"
                  empty={
                    <Box textAlign="center" padding="xl">
                      <Box variant="strong">No workspaces being monitored</Box>
                      <Box variant="p" color="text-body-secondary">
                        Start workspaces with idle detection enabled to see them here.
                      </Box>
                    </Box>
                  }
                  sortingDisabled={false}
                />
              </Container>
            )
          }
        ]}
      />

      {selectedPolicy && (
        <Container
          header={
            <Header
              variant="h2"
              actions={<Button onClick={() => setSelectedPolicy(null)}>Close</Button>}
            >
              {selectedPolicy.name}
            </Header>
          }
        >
          <SpaceBetween size="l">
            <ColumnLayout columns={2}>
              <SpaceBetween size="m">
                <div>
                  <Box variant="awsui-key-label">Policy ID</Box>
                  <Box>{selectedPolicy.id}</Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Idle Threshold</Box>
                  <Box fontWeight="bold">{selectedPolicy.idle_minutes} minutes</Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Action</Box>
                  <Badge color={getActionBadgeColor(selectedPolicy.action)}>
                    {selectedPolicy.action.toUpperCase()}
                  </Badge>
                </div>
                <div>
                  <Box variant="awsui-key-label">Status</Box>
                  <StatusIndicator type={selectedPolicy.enabled ? 'success' : 'stopped'}>
                    {selectedPolicy.enabled ? 'Enabled' : 'Disabled'}
                  </StatusIndicator>
                </div>
              </SpaceBetween>
              <SpaceBetween size="m">
                <div>
                  <Box variant="awsui-key-label">CPU Threshold</Box>
                  <Box>{selectedPolicy.cpu_threshold}%</Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Memory Threshold</Box>
                  <Box>{selectedPolicy.memory_threshold}%</Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Network Threshold</Box>
                  <Box>{selectedPolicy.network_threshold} Mbps</Box>
                </div>
              </SpaceBetween>
            </ColumnLayout>

            {selectedPolicy.description && (
              <div>
                <Box variant="awsui-key-label">Description</Box>
                <Box variant="p">{selectedPolicy.description}</Box>
              </div>
            )}

            <Alert type="info">
              <Box variant="strong">How It Works:</Box>
              <Box variant="p">
                This policy monitors workspace activity. When CPU, memory, and network usage all fall below
                the specified thresholds for {selectedPolicy.idle_minutes} consecutive minutes, the system will
                automatically {selectedPolicy.action === 'hibernate' ? 'hibernate (preserve RAM state)' :
                selectedPolicy.action === 'stop' ? 'stop the workspace' : 'send a notification'}.
              </Box>
            </Alert>

            {selectedPolicy.action === 'hibernate' && (
              <Alert type="success">
                <Box variant="strong">Cost Savings with Hibernation:</Box>
                <Box variant="p">
                  Hibernation preserves your RAM state to disk, allowing instant resume while only paying for
                  EBS storage (~$0.10/GB/month). This can save ~40% on compute costs for workspaces that are
                  idle for extended periods.
                </Box>
              </Alert>
            )}
          </SpaceBetween>
        </Container>
      )}

      <Container header={<Header variant="h2">About Idle Detection</Header>}>
        <SpaceBetween size="m">
          <Box variant="p">
            Idle detection monitors your workspaces and automatically hibernates or stops them when they're not
            being used, saving significant compute costs while preserving your work environment.
          </Box>
          <ColumnLayout columns={3}>
            <div>
              <Box variant="strong">Hibernate</Box>
              <Box variant="small" color="text-body-secondary">
                Preserves RAM state to disk. Resume in seconds with your session intact. Best for
                workloads that need quick resumption.
              </Box>
            </div>
            <div>
              <Box variant="strong">Stop</Box>
              <Box variant="small" color="text-body-secondary">
                Fully stops the workspace. Cheaper than hibernation but requires full restart.
                Best for workspaces that don't need quick resumption.
              </Box>
            </div>
            <div>
              <Box variant="strong">Notify</Box>
              <Box variant="small" color="text-body-secondary">
                Sends a notification without taking action. Useful for monitoring patterns
                before enabling automated actions.
              </Box>
            </div>
          </ColumnLayout>
        </SpaceBetween>
      </Container>
    </SpaceBetween>
  )
}
