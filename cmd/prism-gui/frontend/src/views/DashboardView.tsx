import {
  SpaceBetween,
  Header,
  Container,
  Button,
  Box,
  TextContent,
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

      <Container>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
          <StatusIndicator type={connected ? 'success' : 'error'}>
            {connected ? 'Connected' : 'Disconnected'}
          </StatusIndicator>
          <Box color="text-body-secondary">
            {Object.keys(templates).length} template{Object.keys(templates).length !== 1 ? 's' : ''}
          </Box>
          <Box color="text-body-secondary">
            {instances.filter(i => i.state === 'running').length} running · {instances.length} total workspace{instances.length !== 1 ? 's' : ''}
          </Box>
          <div style={{ marginLeft: 'auto', display: 'flex', gap: '0.5rem' }}>
            <Button variant="primary" iconName="add-plus" onClick={onShowQuickStart}>
              New Workspace
            </Button>
            <Button onClick={onRefresh} disabled={loading}>
              {loading ? <Spinner size="normal" /> : 'Refresh'}
            </Button>
          </div>
        </div>
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
    </SpaceBetween>
  )
}
