import { toast } from 'sonner'
import {
  SpaceBetween,
  Header,
  Container,
  Button,
  Box,
  Table,
  ColumnLayout,
  Link,
  ButtonDropdown,
  Spinner,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import type { InstanceSnapshot } from '../lib/types'

interface BackupManagementViewProps {
  snapshots: InstanceSnapshot[]
  loading: boolean
  onRefresh: () => void
  onNavigate: (view: string) => void
  onCreateBackup: () => void
  onDeleteBackup: (item: InstanceSnapshot) => void
  onRestoreBackup: (item: InstanceSnapshot) => void
}

function snapshotStatusType(state: string) {
  if (state === 'available') return 'success' as const
  if (state === 'creating' || state === 'pending') return 'in-progress' as const
  if (state === 'deleting') return 'warning' as const
  return 'error' as const
}

export function BackupManagementView({
  snapshots,
  loading,
  onRefresh,
  onNavigate,
  onCreateBackup,
  onDeleteBackup,
  onRestoreBackup,
}: BackupManagementViewProps) {
  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="Create and manage backups (snapshots) of your research workspaces for disaster recovery and reproducibility"
        actions={
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={onRefresh} disabled={loading}>
              {loading ? <Spinner /> : 'Refresh'}
            </Button>
          </SpaceBetween>
        }
      >
        Backups
      </Header>

      <Container>
        <SpaceBetween size="m">
          <Box variant="h3">💾 Instance Snapshots & Backups</Box>
          <Box color="text-body-secondary">
            Instance snapshots (AMI backups) capture the complete state of your workspace, including installed software, configurations, and data.
            Use snapshots for disaster recovery, creating reproducible research environments, or cloning workspaces.
          </Box>
          <ColumnLayout columns={3} variant="text-grid">
            <SpaceBetween size="s">
              <Box variant="awsui-key-label">💰 Cost</Box>
              <Box color="text-body-secondary">~$0.05/GB/month for EBS snapshot storage</Box>
            </SpaceBetween>
            <SpaceBetween size="s">
              <Box variant="awsui-key-label">⏱️ Creation Time</Box>
              <Box color="text-body-secondary">5-10 minutes (depending on instance size)</Box>
            </SpaceBetween>
            <SpaceBetween size="s">
              <Box variant="awsui-key-label">🔄 Restore Time</Box>
              <Box color="text-body-secondary">10-15 minutes to launch from snapshot</Box>
            </SpaceBetween>
          </ColumnLayout>
        </SpaceBetween>
      </Container>

      <Container
        header={
          <Header
            variant="h2"
            description="Instance snapshots available for restore or clone operations"
            counter={`(${snapshots.length})`}
            actions={
              <Button variant="primary" onClick={onCreateBackup}>
                Create Backup
              </Button>
            }
          >
            Available Backups
          </Header>
        }
      >
        <Table
          data-testid="backups-table"
          columnDefinitions={[
            {
              id: "name",
              header: "Backup Name",
              cell: (item: InstanceSnapshot) => (
                <div data-testid="backup-name">
                  <Link fontSize="body-m">{item.snapshot_name}</Link>
                </div>
              ),
              sortingField: "snapshot_name"
            },
            {
              id: "instance",
              header: "Source Instance",
              cell: (item: InstanceSnapshot) => item.source_instance,
              sortingField: "source_instance"
            },
            {
              id: "template",
              header: "Template",
              cell: (item: InstanceSnapshot) => item.source_template || 'N/A'
            },
            {
              id: "size",
              header: "Size",
              cell: (item: InstanceSnapshot) => {
                const sizeGB = item.size_gb || Math.ceil(item.storage_cost_monthly / 0.05)
                return `${sizeGB} GB`
              }
            },
            {
              id: "status",
              header: "Status",
              cell: (item: InstanceSnapshot) => (
                <div data-testid="status-badge">
                  <StatusIndicator type={snapshotStatusType(item.state)}>
                    {item.state}
                  </StatusIndicator>
                </div>
              )
            },
            {
              id: "created",
              header: "Created",
              cell: (item: InstanceSnapshot) => {
                const date = new Date(item.created_at)
                return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
              }
            },
            {
              id: "cost",
              header: "Monthly Cost",
              cell: (item: InstanceSnapshot) => `$${item.storage_cost_monthly.toFixed(2)}`
            },
            {
              id: "actions",
              header: "Actions",
              cell: (item: InstanceSnapshot) => (
                <ButtonDropdown
                  expandToViewport
                  items={[
                    { text: 'Restore to New Instance', id: 'restore', disabled: item.state !== 'available' },
                    { text: 'Clone Instance', id: 'clone', disabled: item.state !== 'available' },
                    { text: 'View Details', id: 'details' },
                    { text: 'Delete', id: 'delete', disabled: item.state !== 'available' }
                  ]}
                  onItemClick={({ detail }) => {
                    if (detail.id === 'delete') {
                      onDeleteBackup(item)
                    } else if (detail.id === 'restore') {
                      onRestoreBackup(item)
                    } else if (detail.id === 'clone') {
                      toast.info('Clone Instance', {
                        description: `Restoring backup "${item.snapshot_name}" will create a new instance. This may take 10-15 minutes.`
                      })
                    }
                  }}
                >
                  Actions
                </ButtonDropdown>
              )
            }
          ]}
          items={snapshots}
          loadingText="Loading backups from AWS"
          loading={loading}
          trackBy="snapshot_id"
          empty={
            <Box data-testid="empty-backups" textAlign="center" color="inherit" padding={{ vertical: 'xl' }}>
              <SpaceBetween size="m">
                <Box variant="strong" textAlign="center" color="inherit">No backups found</Box>
                <Box variant="p" color="text-body-secondary">
                  Create backups of your workspaces to enable disaster recovery, reproducibility, and environment cloning.
                  Backups capture the complete state of your instance including all installed software and data.
                </Box>
                <Box textAlign="center">
                  <Button variant="primary" onClick={() => onNavigate('workspaces')}>
                    Go to Workspaces
                  </Button>
                </Box>
              </SpaceBetween>
            </Box>
          }
          sortingDisabled={false}
        />
      </Container>

      <Container header={<Header variant="h3">📊 Backup Storage Summary</Header>}>
        <ColumnLayout columns={4} variant="text-grid">
          <SpaceBetween size="s">
            <Box variant="awsui-key-label">Total Backups</Box>
            <Box fontSize="display-l" fontWeight="bold" color="text-status-info">
              {snapshots.length}
            </Box>
          </SpaceBetween>
          <SpaceBetween size="s">
            <Box variant="awsui-key-label">Total Storage Size</Box>
            <Box fontSize="display-l" fontWeight="bold" color="text-body-secondary">
              {snapshots.reduce((sum, s) => {
                const sizeGB = s.size_gb || Math.ceil(s.storage_cost_monthly / 0.05)
                return sum + sizeGB
              }, 0)} GB
            </Box>
          </SpaceBetween>
          <SpaceBetween size="s">
            <Box variant="awsui-key-label">Available Backups</Box>
            <Box fontSize="display-l" fontWeight="bold" color="text-status-success">
              {snapshots.filter(s => s.state === 'available').length}
            </Box>
          </SpaceBetween>
          <SpaceBetween size="s">
            <Box variant="awsui-key-label">Monthly Storage Cost</Box>
            <Box fontSize="display-l" fontWeight="bold" color="text-status-warning">
              ${snapshots.reduce((sum, s) => sum + s.storage_cost_monthly, 0).toFixed(2)}
            </Box>
          </SpaceBetween>
        </ColumnLayout>
      </Container>
    </SpaceBetween>
  )
}
