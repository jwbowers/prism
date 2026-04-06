import {
  Modal,
  SpaceBetween,
  Container,
  Header,
  ColumnLayout,
  Box,
  Button,
  Spinner,
  StatusIndicator,
} from '../lib/cloudscape-shim'
import type { User, UserStatus } from '../lib/types'

interface UserStatusModalProps {
  visible: boolean
  user: User | null
  statusDetails: UserStatus | null
  loading: boolean
  onDismiss: () => void
}

export function UserStatusModal({ visible, user, statusDetails, loading, onDismiss }: UserStatusModalProps) {
  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      size="medium"
      header={user ? `User Status: ${user.username}` : 'User Status'}
      data-testid="user-status-modal"
      footer={
        <Box float="right">
          <Button onClick={onDismiss} data-testid="close">
            Close
          </Button>
        </Box>
      }
    >
      {user && (
        <SpaceBetween size="m">
          {loading ? (
            <Box textAlign="center" padding={{ vertical: 'xl' }}>
              <Spinner size="large" />
            </Box>
          ) : statusDetails ? (
            <Container header={<Header variant="h2">Status Details</Header>}>
              <ColumnLayout columns={2} variant="text-grid">
                <div>
                  <Box variant="awsui-key-label">Username</Box>
                  <div>{statusDetails.username}</div>
                </div>
                <div>
                  <Box variant="awsui-key-label">Status</Box>
                  <div>
                    <StatusIndicator type={statusDetails.status === 'active' ? 'success' : 'warning'}>
                      {statusDetails.status || 'active'}
                    </StatusIndicator>
                  </div>
                </div>
                <div>
                  <Box variant="awsui-key-label">SSH Keys</Box>
                  <div>{statusDetails.ssh_keys_count || 0}</div>
                </div>
                <div>
                  <Box variant="awsui-key-label">Provisioned Workspaces</Box>
                  <div>{statusDetails.provisioned_instances?.length || 0}</div>
                </div>
                {statusDetails.last_active && (
                  <div>
                    <Box variant="awsui-key-label">Last Active</Box>
                    <div>{new Date(statusDetails.last_active).toLocaleString()}</div>
                  </div>
                )}
              </ColumnLayout>
            </Container>
          ) : (
            <Box textAlign="center" color="inherit" padding={{ vertical: 'xl' }}>
              <Box variant="p" color="text-body-secondary">
                Failed to load user status
              </Box>
            </Box>
          )}
        </SpaceBetween>
      )}
    </Modal>
  )
}
