import {
  Modal,
  SpaceBetween,
  Box,
  Button,
  FormField,
} from '../lib/cloudscape-shim'

interface ConnectionInfo {
  instanceName: string
  publicIP: string
  sshCommand: string
  webPort: string
}

interface ConnectionInfoModalProps {
  visible: boolean
  connectionInfo: ConnectionInfo | null
  onDismiss: () => void
}

export function ConnectionInfoModal({ visible, connectionInfo, onDismiss }: ConnectionInfoModalProps) {
  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      header="Connection Information"
      footer={
        <Box float="right">
          <Button
            variant="primary"
            onClick={onDismiss}
          >
            Close
          </Button>
        </Box>
      }
    >
      {connectionInfo && (
        <SpaceBetween size="m">
          <FormField label="Workspace">
            <Box>{connectionInfo.instanceName}</Box>
          </FormField>
          {connectionInfo.publicIP && (
            <FormField label="Public IP" description="Instance public IP address">
              <Box data-testid="public-ip">{connectionInfo.publicIP}</Box>
            </FormField>
          )}
          <FormField label="SSH Command" description="Use this command to connect via SSH">
            <SpaceBetween direction="horizontal" size="xs">
              <code data-testid="ssh-command">{connectionInfo.sshCommand}</code>
              <Button
                iconName="copy"
                onClick={() => navigator.clipboard.writeText(connectionInfo!.sshCommand)}
              >
                Copy SSH
              </Button>
            </SpaceBetween>
          </FormField>
          {/* web-url is always in DOM (even when empty) for ConnectionDialog.hasWebURL() to work */}
          <span data-testid="web-url" aria-hidden="true" style={{ display: 'none' }}>
            {connectionInfo.publicIP && connectionInfo.webPort
              ? `http://${connectionInfo.publicIP}:${connectionInfo.webPort}`
              : ''}
          </span>
          {connectionInfo.publicIP && connectionInfo.webPort && (
            <FormField label="Web URL" description="Access web services running on this instance">
              <Box>{`http://${connectionInfo.publicIP}:${connectionInfo.webPort}`}</Box>
            </FormField>
          )}
        </SpaceBetween>
      )}
    </Modal>
  )
}
