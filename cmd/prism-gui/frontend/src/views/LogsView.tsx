import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import {
  SpaceBetween,
  Header,
  Container,
  Select,
  Button,
  Box,
  FormField,
  Alert,
  ColumnLayout,
  Spinner,
} from '../lib/cloudscape-shim'
import type { Instance } from '../lib/types'

interface LogsViewProps {
  instances: Instance[]
  loading?: boolean
  onRefresh?: () => void
}

export function LogsView({ instances, loading, onRefresh }: LogsViewProps) {
  const [selectedInstance, setSelectedInstance] = useState<string>('')
  const [logType, setLogType] = useState<string>('console')
  const [logLines, setLogLines] = useState<string[]>([])
  const [loadingLogs, setLoadingLogs] = useState(false)

  const logTypes = [
    { label: 'Console Output', value: 'console' },
    { label: 'Cloud-Init Log', value: 'cloud-init' },
    { label: 'System Log', value: 'system' },
    { label: 'Application Log', value: 'application' }
  ]

  const runningInstances = instances.filter(i => i.state === 'running' || i.state === 'stopped')

  const fetchLogs = useCallback(async () => {
    if (!selectedInstance) return

    setLoadingLogs(true)
    try {
      // Mock log fetching - in real implementation would call API
      const mockLogs = [
        `[${new Date().toISOString()}] Workspace ${selectedInstance} logs (${logType})`,
        `[INFO] Workspace started successfully`,
        `[INFO] Loading configuration...`,
        `[INFO] Mounting EFS volumes...`,
        `[INFO] Starting services...`,
        `[INFO] Prism template: ${instances.find(i => i.name === selectedInstance)?.template || 'unknown'}`,
        `[INFO] All services running`,
        `[DEBUG] Memory usage: 1.2GB / 8GB`,
        `[DEBUG] CPU usage: 5%`,
        `[INFO] Workspace ready for use`,
        `[INFO] SSH access: ssh ${instances.find(i => i.name === selectedInstance)?.public_ip || 'N/A'}`,
        `--- End of ${logType} log ---`
      ]

      setLogLines(mockLogs)
    } catch (error) {
      toast.error(`Failed to fetch logs: ${error}`)
      setLogLines([`Error fetching logs: ${error}`])
    } finally {
      setLoadingLogs(false)
    }
  }, [selectedInstance, logType, instances])

  useEffect(() => {
    if (selectedInstance) {
      fetchLogs()
    }
  }, [selectedInstance, logType, fetchLogs])

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="View workspace console output and system logs"
        actions={
          onRefresh ? (
            <Button onClick={onRefresh} disabled={loading}>
              {loading ? <Spinner /> : 'Refresh'}
            </Button>
          ) : undefined
        }
      >
        Workspace Logs Viewer
      </Header>

      <Container>
        <SpaceBetween size="m">
          <FormField
            label="Workspace"
            description="Select a workspace to view its logs"
          >
            <Select
              selectedOption={selectedInstance ?
                { label: selectedInstance, value: selectedInstance } : null}
              onChange={({ detail }) => {
                setSelectedInstance(detail.selectedOption?.value || '')
                setLogLines([])
              }}
              options={runningInstances.map(i => ({
                label: `${i.name} (${i.state})`,
                value: i.name
              }))}
              placeholder="Choose a workspace"
              selectedAriaLabel="Selected workspace"
              disabled={runningInstances.length === 0}
            />
          </FormField>

          {selectedInstance && (
            <FormField
              label="Log Type"
              description="Select the type of log to view"
            >
              <Select
                selectedOption={logType ?
                  (logTypes.find(t => t.value === logType) || null) : null}
                onChange={({ detail }) => {
                  setLogType(detail.selectedOption?.value || 'console')
                  setLogLines([])
                }}
                options={logTypes}
                selectedAriaLabel="Selected log type"
              />
            </FormField>
          )}

          {selectedInstance && (
            <Button
              onClick={fetchLogs}
              loading={loadingLogs}
              disabled={loadingLogs}
            >
              Refresh Logs
            </Button>
          )}
        </SpaceBetween>
      </Container>

      {selectedInstance ? (
        <Container
          header={
            <Header
              variant="h2"
              description={`Viewing ${logType} logs for ${selectedInstance}`}
            >
              Log Output
            </Header>
          }
        >
          {loadingLogs ? (
            <Box textAlign="center" padding="xl">
              <Spinner size="large" />
              <Box variant="p">Loading logs...</Box>
            </Box>
          ) : logLines.length > 0 ? (
            <Box padding="s" variant="code">
              <pre style={{
                fontFamily: 'monospace',
                fontSize: '12px',
                lineHeight: '1.5',
                margin: 0,
                padding: '8px',
                backgroundColor: '#232f3e',
                color: '#d4d4d4',
                borderRadius: '4px',
                maxHeight: '600px',
                overflow: 'auto',
                whiteSpace: 'pre-wrap',
                wordWrap: 'break-word'
              }}>
                {logLines.join('\n')}
              </pre>
            </Box>
          ) : (
            <Box textAlign="center" padding="xl">
              <Box variant="strong">No logs available</Box>
              <Box variant="p" color="text-body-secondary">
                Select a log type and click "Refresh Logs" to view output.
              </Box>
            </Box>
          )}

          {logLines.length > 0 && (
            <Box padding={{ top: 'm' }}>
              <SpaceBetween direction="horizontal" size="xs">
                <Button iconName="copy" onClick={() => {
                  navigator.clipboard.writeText(logLines.join('\n'))
                  toast.success('Logs copied to clipboard')
                }}>
                  Copy to Clipboard
                </Button>
                <Button iconName="download" onClick={() => {
                  const blob = new Blob([logLines.join('\n')], { type: 'text/plain' })
                  const url = URL.createObjectURL(blob)
                  const a = document.createElement('a')
                  a.href = url
                  a.download = `${selectedInstance}-${logType}-${new Date().toISOString().split('T')[0]}.log`
                  a.click()
                  URL.revokeObjectURL(url)
                  toast.success('Log file downloaded')
                }}>
                  Download Log File
                </Button>
              </SpaceBetween>
            </Box>
          )}
        </Container>
      ) : (
        <Container>
          <Box textAlign="center" padding="xl">
            <Box variant="strong">Select a Workspace</Box>
            <Box variant="p" color="text-body-secondary">
              {runningInstances.length === 0
                ? 'No running or stopped workspaces available. Start a workspace to view its logs.'
                : 'Choose a workspace from the dropdown above to view its logs.'}
            </Box>
          </Box>
        </Container>
      )}

      <Container header={<Header variant="h2">About Log Viewing</Header>}>
        <SpaceBetween size="m">
          <Box variant="p">
            View real-time console output and system logs from your Prism workspaces.
            Logs are useful for troubleshooting startup issues, monitoring application output,
            and understanding workspace behavior.
          </Box>
          <ColumnLayout columns={4}>
            <div>
              <Box variant="strong">Console Output</Box>
              <Box variant="small" color="text-body-secondary">
                System boot messages and console output
              </Box>
            </div>
            <div>
              <Box variant="strong">Cloud-Init</Box>
              <Box variant="small" color="text-body-secondary">
                Prism provisioning logs
              </Box>
            </div>
            <div>
              <Box variant="strong">System Log</Box>
              <Box variant="small" color="text-body-secondary">
                Operating system events and services
              </Box>
            </div>
            <div>
              <Box variant="strong">Application Log</Box>
              <Box variant="small" color="text-body-secondary">
                Application-specific output
              </Box>
            </div>
          </ColumnLayout>
          <Alert type="info">
            <Box variant="strong">Note:</Box> Log viewing is read-only. To interact with your workspace,
            use SSH: <Box variant="code">
              ssh {selectedInstance && instances.find(i => i.name === selectedInstance)?.public_ip || 'instance-ip'}
            </Box>
          </Alert>
        </SpaceBetween>
      </Container>
    </SpaceBetween>
  )
}
