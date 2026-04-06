import {
  SpaceBetween,
  Container,
  Header,
  Alert,
  FormField,
  Select,
} from '../lib/cloudscape-shim'
import Terminal from '../Terminal'
import type { Instance } from '../lib/types'

interface TerminalViewProps {
  instances: Instance[]
  selectedTerminalInstance: string
  onSelectInstance: (name: string) => void
}

export function TerminalView({ instances, selectedTerminalInstance, onSelectInstance }: TerminalViewProps) {
  const runningInstances = instances.filter(i => i.state === 'running')

  if (runningInstances.length === 0) {
    return (
      <Container header={<Header variant="h1">SSH Terminal</Header>}>
        <Alert type="info">
          No running workspaces available. Launch a workspace to access the SSH terminal.
        </Alert>
      </Container>
    )
  }

  return (
    <SpaceBetween size="l">
      <Container header={<Header variant="h1">SSH Terminal</Header>}>
        <SpaceBetween size="m">
          <FormField label="Select Workspace">
            <Select
              selectedOption={selectedTerminalInstance ? { label: selectedTerminalInstance, value: selectedTerminalInstance } : null}
              onChange={({ detail }) => onSelectInstance(detail.selectedOption.value || '')}
              options={runningInstances.map(i => ({ label: i.name, value: i.name }))}
              placeholder="Choose a workspace"
            />
          </FormField>
          {selectedTerminalInstance && <Terminal instanceName={selectedTerminalInstance} />}
        </SpaceBetween>
      </Container>
    </SpaceBetween>
  )
}
