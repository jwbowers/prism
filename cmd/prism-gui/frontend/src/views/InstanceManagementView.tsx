import {
  SpaceBetween,
  Container,
  Header,
  Button,
  Box,
  Table,
  Link,
  Spinner,
  ButtonDropdown,
  PropertyFilter,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import type { Instance } from '../lib/types'

interface ConnectionInfo {
  instanceName: string
  publicIP: string
  sshCommand: string
  webPort: string
}

type FilterQuery = {
  tokens: Array<{ propertyKey?: string; operator: string; value: string }>
  operation: 'and' | 'or'
}

interface InstanceManagementViewProps {
  instances: Instance[]
  loading: boolean
  filterQuery: FilterQuery
  onFilterChange: (query: FilterQuery) => void
  selectedInstances: Instance[]
  onSelectionChange: (items: Instance[]) => void
  onRefresh: () => void
  onNavigateToTemplates: () => void
  onConnect: (info: ConnectionInfo) => void
  onInstanceAction: (action: string, instance: Instance) => void
  onBulkAction: (action: 'start' | 'stop' | 'hibernate' | 'delete') => void
}

export function InstanceManagementView({
  instances,
  loading,
  filterQuery,
  onFilterChange,
  selectedInstances,
  onSelectionChange,
  onRefresh,
  onNavigateToTemplates,
  onConnect,
  onInstanceAction,
  onBulkAction,
}: InstanceManagementViewProps) {
  return (
    <SpaceBetween size="l">
      <Container
        header={
          <Header
            variant="h1"
            description="Monitor and manage your research computing environments"
            counter={`(${instances.length})`}
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Button
                  onClick={onRefresh}
                  disabled={loading}
                  data-testid="refresh-instances-button"
                >
                  {loading ? <Spinner /> : 'Refresh'}
                </Button>
                <Button variant="primary" onClick={onNavigateToTemplates}>
                  Launch New Workspace
                </Button>
              </SpaceBetween>
            }
          >
            My Workspaces
          </Header>
        }
      >
        <PropertyFilter
          query={filterQuery}
          onChange={({ detail }) => onFilterChange(detail)}
          filteringPlaceholder="Search instances by name or filter by status"
          filteringProperties={[
            { key: 'name', propertyLabel: 'Workspace Name', groupValuesLabel: 'Workspace Name values', operators: [':', '!:', '=', '!='] },
            { key: 'template', propertyLabel: 'Template', groupValuesLabel: 'Template values', operators: [':', '!:', '=', '!='] },
            { key: 'state', propertyLabel: 'Status', groupValuesLabel: 'Status values', operators: ['=', '!='] },
            { key: 'public_ip', propertyLabel: 'Public IP', groupValuesLabel: 'Public IP values', operators: [':', '!:', '=', '!='] }
          ]}
          filteringOptions={[
            { propertyKey: 'state', value: 'running', label: 'Status: Running' },
            { propertyKey: 'state', value: 'stopped', label: 'Status: Stopped' },
            { propertyKey: 'state', value: 'hibernated', label: 'Status: Hibernated' },
            { propertyKey: 'state', value: 'pending', label: 'Status: Pending' }
          ]}
        />

        {selectedInstances.length > 0 && (
          <SpaceBetween direction="horizontal" size="xs">
            <Box variant="awsui-key-label">
              {selectedInstances.length} workspace{selectedInstances.length !== 1 ? 's' : ''} selected
            </Box>
            <Button
              onClick={() => onBulkAction('start')}
              disabled={loading || selectedInstances.every(i => i.state === 'running')}
            >
              Start Selected
            </Button>
            <Button
              onClick={() => onBulkAction('stop')}
              disabled={loading || selectedInstances.every(i => i.state !== 'running')}
            >
              Stop Selected
            </Button>
            <Button
              onClick={() => onBulkAction('hibernate')}
              disabled={loading || selectedInstances.every(i => i.state !== 'running')}
            >
              Hibernate Selected
            </Button>
            <Button
              onClick={() => onBulkAction('delete')}
              disabled={loading}
            >
              Delete Selected
            </Button>
            <Button variant="link" onClick={() => onSelectionChange([])}>
              Clear Selection
            </Button>
          </SpaceBetween>
        )}

        <Table
          data-testid="instances-table"
          selectionType="multi"
          selectedItems={selectedInstances}
          onSelectionChange={({ detail }) => onSelectionChange(detail.selectedItems)}
          columnDefinitions={[
            {
              id: 'name',
              header: 'Workspace Name',
              cell: (item: Instance) => <Link fontSize="body-m" data-testid="instance-name">{item.name}</Link>,
              sortingField: 'name'
            },
            {
              id: 'template',
              header: 'Template',
              cell: (item: Instance) => item.template
            },
            {
              id: 'status',
              header: 'Status',
              cell: (item: Instance) => (
                <div data-testid="instance-status">
                  <span data-testid="status-badge">
                    <StatusIndicator
                      type={
                        item.state === 'running' ? 'success' :
                        item.state === 'stopped' ? 'stopped' :
                        item.state === 'hibernated' ? 'pending' :
                        item.state === 'pending' ? 'in-progress' : 'error'
                      }
                    >
                      {item.state}
                    </StatusIndicator>
                  </span>
                </div>
              )
            },
            {
              id: 'public_ip',
              header: 'Public IP',
              cell: (item: Instance) => item.public_ip || 'Not assigned'
            },
            {
              id: 'dns_hostname',
              header: 'Hostname',
              cell: (item: Instance) => item.dns_hostname ? (
                <span className="font-mono text-xs">{item.dns_hostname}</span>
              ) : <span className="text-muted-foreground">—</span>
            },
            {
              id: 'time_remaining',
              header: 'Time Remaining',
              cell: (item: Instance) => {
                if (!item.ttl && !item.expires_at) return <span className="text-muted-foreground">—</span>
                if (item.expires_at) {
                  const ms = new Date(item.expires_at).getTime() - Date.now()
                  if (ms <= 0) return <StatusIndicator type="error">Expired</StatusIndicator>
                  const hours = Math.floor(ms / 3600000)
                  const mins = Math.floor((ms % 3600000) / 60000)
                  if (hours < 1) return <StatusIndicator type="error">{mins}m</StatusIndicator>
                  if (hours < 2) return <StatusIndicator type="warning">{hours}h {mins}m</StatusIndicator>
                  return <StatusIndicator type="info">{hours}h</StatusIndicator>
                }
                return <span className="text-muted-foreground">{item.ttl}</span>
              },
            },
            {
              id: 'actions',
              header: 'Actions',
              cell: (item: Instance) => (
                <SpaceBetween direction="horizontal" size="xs">
                  {item.state === 'running' && (
                    <Button
                      data-testid={`connect-btn-${item.name}`}
                      iconName="external"
                      variant="inline-link"
                      onClick={() => {
                        const host = item.dns_hostname || item.public_ip || ''
                        const user = item.username || 'ubuntu'
                        onConnect({
                          instanceName: item.name,
                          publicIP: item.public_ip || '',
                          sshCommand: host ? `ssh ${user}@${host}` : `ssh ${user}@<instance-ip>`,
                          webPort: ''
                        })
                      }}
                    >
                      Connect
                    </Button>
                  )}
                  <ButtonDropdown
                    expandToViewport
                    items={[
                      { text: 'Connect', id: 'connect', disabled: item.state !== 'running' },
                      { text: 'Open Terminal', id: 'terminal', disabled: item.state !== 'running', iconName: 'command-prompt' },
                      { text: 'Open Web Service', id: 'webservice', disabled: item.state !== 'running' || !item.web_services || item.web_services.length === 0, iconName: 'external' },
                      { text: 'Stop', id: 'stop', disabled: item.state !== 'running' },
                      { text: 'Start', id: 'start', disabled: item.state === 'running' },
                      { text: 'Hibernate', id: 'hibernate', disabled: item.state !== 'running' },
                      { text: 'Resume', id: 'resume', disabled: item.state !== 'stopped' && item.state !== 'hibernated' },
                      { text: 'Manage Idle Policy', id: 'manage-idle-policy' },
                      ...(item.expires_at ? [{ text: 'Extend Time (+4h)', id: 'extend-ttl' }] : []),
                      { text: 'Delete', id: 'delete', disabled: item.state === 'running' || item.state === 'pending' }
                    ]}
                    onItemClick={({ detail }) => onInstanceAction(detail.id, item)}
                  >
                    Actions
                  </ButtonDropdown>
                </SpaceBetween>
              )
            }
          ]}
          items={instances}
          loadingText="Loading workspaces from AWS"
          loading={loading}
          trackBy="id"
          empty={
            <Box data-testid="empty-instances" textAlign="center" color="inherit">
              <Box variant="strong" textAlign="center" color="inherit">
                No workspaces running
              </Box>
              <Box variant="p" padding={{ bottom: 's' }} color="inherit">
                Launch your first research environment to get started.
              </Box>
              <Button variant="primary" onClick={onNavigateToTemplates}>
                Browse Templates
              </Button>
            </Box>
          }
          sortingDisabled={false}
        />
      </Container>
    </SpaceBetween>
  )
}
