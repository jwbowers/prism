import {
  SpaceBetween,
  Header,
  Container,
  Button,
  Box,
  TextContent,
  ColumnLayout,
  Link,
  Spinner,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import type { Instance, Template } from '../lib/types'

interface ConnectionInfo {
  instanceName: string
  publicIP: string
  sshCommand: string
  webPort: string
}

interface DashboardViewProps {
  instances: Instance[]
  templates: Record<string, Template>
  connected: boolean
  loading: boolean
  isFirstTimeUser: boolean
  onNavigate: (view: string) => void
  onRefresh: () => void
  onShowQuickStart: () => void
  onConnect: (info: ConnectionInfo) => void
  onStartInstance: (instanceName: string) => void
}

interface RecentWorkspacesProps {
  instances: Instance[]
  onNavigate: (view: string) => void
  onShowQuickStart: () => void
  onConnect: (info: ConnectionInfo) => void
  onStartInstance: (instanceName: string) => void
}

function RecentWorkspaces({ instances, onNavigate, onShowQuickStart, onConnect, onStartInstance }: RecentWorkspacesProps) {
  const recentWorkspaces = instances.slice(0, 3)

  return (
    <Container header={<Header variant="h2">Recent Workspaces</Header>}>
      <SpaceBetween size="m">
        {recentWorkspaces.length === 0 ? (
          <Box textAlign="center" padding={{ vertical: 'l' }}>
            <TextContent>
              <Box variant="p" color="text-body-secondary">
                No workspaces yet. Launch your first workspace to get started.
              </Box>
            </TextContent>
            <Button variant="primary" iconName="add-plus" onClick={onShowQuickStart}>
              Launch Workspace
            </Button>
          </Box>
        ) : (
          <>
            {recentWorkspaces.map((instance) => (
              <Container key={instance.name}>
                <SpaceBetween size="s">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                      <Box variant="h3">{instance.name}</Box>
                      <Box variant="small" color="text-body-secondary">
                        Template: {instance.template} | Type: {instance.instance_type || 'N/A'}
                      </Box>
                    </div>
                    <StatusIndicator type={
                      instance.state === 'running' ? 'success' :
                      instance.state === 'stopped' ? 'stopped' :
                      instance.state === 'pending' ? 'in-progress' :
                      'error'
                    }>
                      {instance.state}
                    </StatusIndicator>
                  </div>
                  <SpaceBetween direction="horizontal" size="xs">
                    {instance.state === 'running' && (
                      <Button
                        iconName="external"
                        onClick={() => onConnect({
                          instanceName: instance.name,
                          publicIP: instance.public_ip || '',
                          sshCommand: `ssh -i ~/.ssh/your-key.pem ubuntu@${instance.public_ip}`,
                          webPort: ''
                        })}
                      >
                        Connect
                      </Button>
                    )}
                    {instance.state === 'stopped' && (
                      <Button onClick={() => onStartInstance(instance.name)}>
                        Start
                      </Button>
                    )}
                    <Button variant="normal" onClick={() => onNavigate('workspaces')}>
                      Manage
                    </Button>
                  </SpaceBetween>
                </SpaceBetween>
              </Container>
            ))}
            {instances.length > 3 && (
              <Box textAlign="center">
                <Link onFollow={() => onNavigate('workspaces')}>
                  View all {instances.length} workspaces
                </Link>
              </Box>
            )}
          </>
        )}
      </SpaceBetween>
    </Container>
  )
}

export function DashboardView({
  instances,
  templates,
  connected,
  loading,
  isFirstTimeUser,
  onNavigate,
  onRefresh,
  onShowQuickStart,
  onConnect,
  onStartInstance,
}: DashboardViewProps) {
  return (
    <SpaceBetween size="l">
      <Container>
        <SpaceBetween size="l">
          <Box textAlign="center" padding={{ top: 'xl', bottom: 'l' }}>
            <SpaceBetween size="m">
              <TextContent>
                <h1>Welcome to Prism</h1>
                <Box variant="p" fontSize="heading-m" color="text-body-secondary">
                  {isFirstTimeUser
                    ? 'Launch your research workspace in seconds'
                    : 'Manage your research workspaces'}
                </Box>
              </TextContent>
              {isFirstTimeUser && (
                <>
                  <Button variant="primary" iconName="add-plus" onClick={onShowQuickStart}>
                    Quick Start - Launch Workspace
                  </Button>
                  <Box color="text-body-secondary">
                    Pre-configured environments for ML, Data Science, Bioinformatics, and more
                  </Box>
                </>
              )}
              {!isFirstTimeUser && (
                <SpaceBetween direction="horizontal" size="s">
                  <Button variant="primary" iconName="add-plus" onClick={onShowQuickStart}>
                    New Workspace
                  </Button>
                  <Button variant="normal" iconName="view-full" onClick={() => onNavigate('workspaces')}>
                    View All Workspaces
                  </Button>
                </SpaceBetween>
              )}
            </SpaceBetween>
          </Box>
        </SpaceBetween>
      </Container>

      {!isFirstTimeUser && (
        <RecentWorkspaces
          instances={instances}
          onNavigate={onNavigate}
          onShowQuickStart={onShowQuickStart}
          onConnect={onConnect}
          onStartInstance={onStartInstance}
        />
      )}

      <Header
        variant="h1"
        description="Prism research computing platform - manage your cloud environments"
        actions={
          <Button onClick={onRefresh} disabled={loading}>
            {loading ? <Spinner size="normal" /> : 'Refresh'}
          </Button>
        }
      >
        Dashboard
      </Header>

      <ColumnLayout columns={3} variant="text-grid">
        <Container header={<Header variant="h2">Research Templates</Header>}>
          <SpaceBetween size="s">
            <Box>
              <Box variant="awsui-key-label">Available Templates</Box>
              <Box fontSize="display-l" fontWeight="bold" color={connected ? 'text-status-success' : 'text-status-error'}>
                {Object.keys(templates).length}
              </Box>
            </Box>
            <Button variant="primary" onClick={() => onNavigate('templates')}>
              Browse Templates
            </Button>
          </SpaceBetween>
        </Container>

        <Container header={<Header variant="h2">Active Workspaces</Header>}>
          <SpaceBetween size="s">
            <Box>
              <Box variant="awsui-key-label">Running Workspaces</Box>
              <Box fontSize="display-l" fontWeight="bold" color="text-status-success">
                {instances.filter(i => i.state === 'running').length}
              </Box>
            </Box>
            <Button onClick={() => onNavigate('workspaces')}>
              Manage Workspaces
            </Button>
          </SpaceBetween>
        </Container>

        <Container header={<Header variant="h2">System Status</Header>}>
          <SpaceBetween size="s">
            <Box>
              <Box variant="awsui-key-label">Connection</Box>
              <StatusIndicator type={connected ? 'success' : 'error'}>
                {connected ? 'Connected' : 'Disconnected'}
              </StatusIndicator>
            </Box>
            <Button onClick={onRefresh} disabled={loading}>
              {loading ? 'Checking...' : 'Test Connection'}
            </Button>
          </SpaceBetween>
        </Container>
      </ColumnLayout>

      <Container header={<Header variant="h2">Quick Actions</Header>}>
        <SpaceBetween direction="horizontal" size="s">
          <Button
            variant="primary"
            onClick={() => onNavigate('templates')}
            disabled={Object.keys(templates).length === 0}
          >
            Launch New Workspace
          </Button>
          <Button
            onClick={() => onNavigate('workspaces')}
            disabled={instances.length === 0}
          >
            View Workspaces ({instances.length})
          </Button>
          <Button onClick={() => onNavigate('storage')}>
            Storage Management
          </Button>
        </SpaceBetween>
      </Container>
    </SpaceBetween>
  )
}
