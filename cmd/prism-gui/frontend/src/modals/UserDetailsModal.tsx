import {
  Modal,
  SpaceBetween,
  Container,
  Header,
  ColumnLayout,
  Box,
  Table,
  Spinner,
} from '../lib/cloudscape-shim'
import type { User, SSHKeyConfig } from '../lib/types'

interface UserDetailsModalProps {
  visible: boolean
  user: User | null
  sshKeys: SSHKeyConfig[]
  loadingSSHKeys: boolean
  onDismiss: () => void
}

export function UserDetailsModal({ visible, user, sshKeys, loadingSSHKeys, onDismiss }: UserDetailsModalProps) {
  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      size="large"
      header={user ? `User Details: ${user.username}` : 'User Details'}
      data-testid="user-details-modal"
    >
      {user && (
        <SpaceBetween size="l">
          {/* User Information */}
          <Container header={<Header variant="h2">User Information</Header>}>
            <ColumnLayout columns={2} variant="text-grid">
              <div>
                <Box variant="awsui-key-label">Username</Box>
                <div>{user.username}</div>
              </div>
              <div>
                <Box variant="awsui-key-label">Display Name</Box>
                <div>{user.display_name}</div>
              </div>
              <div>
                <Box variant="awsui-key-label">Email</Box>
                <div>{user.email}</div>
              </div>
              <div>
                <Box variant="awsui-key-label">UID</Box>
                <div>{user.uid}</div>
              </div>
              <div>
                <Box variant="awsui-key-label">Created</Box>
                <div>{new Date(user.created_at).toLocaleString()}</div>
              </div>
              <div>
                <Box variant="awsui-key-label">SSH Keys</Box>
                <div>{user.ssh_keys || 0}</div>
              </div>
            </ColumnLayout>
          </Container>

          {/* SSH Keys Section */}
          <Container header={<Header variant="h2">SSH Keys</Header>}>
            {loadingSSHKeys ? (
              <Box textAlign="center" padding={{ vertical: 'xl' }}>
                <Spinner size="large" />
              </Box>
            ) : sshKeys.length === 0 ? (
              <Box textAlign="center" color="inherit" padding={{ vertical: 'xl' }}>
                <Box variant="p" color="text-body-secondary">
                  No SSH keys found for this user
                </Box>
              </Box>
            ) : (
              <Table
                columnDefinitions={[
                  {
                    id: "key_type",
                    header: "Type",
                    cell: (item: SSHKeyConfig) => item.key_type,
                    width: 100
                  },
                  {
                    id: "fingerprint",
                    header: "Fingerprint",
                    cell: (item: SSHKeyConfig) => (
                      <Box fontSize="body-s">
                        <span style={{ fontFamily: 'monospace' }}>{item.fingerprint}</span>
                      </Box>
                    )
                  },
                  {
                    id: "comment",
                    header: "Comment",
                    cell: (item: SSHKeyConfig) => item.comment,
                    width: 250
                  },
                  {
                    id: "created_at",
                    header: "Created",
                    cell: (item: SSHKeyConfig) => new Date(item.created_at).toLocaleString(),
                    width: 180
                  },
                  {
                    id: "auto_generated",
                    header: "Auto-Generated",
                    cell: (item: SSHKeyConfig) => item.auto_generated ? "Yes" : "No",
                    width: 120
                  }
                ]}
                items={sshKeys}
                variant="embedded"
                empty={
                  <Box textAlign="center" color="inherit">
                    <Box variant="p" color="text-body-secondary">
                      No SSH keys
                    </Box>
                  </Box>
                }
              />
            )}
          </Container>

          {/* Provisioned Workspaces Section */}
          <Container header={<Header variant="h2">Provisioned Workspaces</Header>}>
            {user.provisioned_instances && user.provisioned_instances.length > 0 ? (
              <Table
                columnDefinitions={[
                  {
                    id: "workspace",
                    header: "Workspace",
                    cell: (item: string) => item
                  }
                ]}
                items={user.provisioned_instances}
                variant="embedded"
                empty={
                  <Box textAlign="center" color="inherit">
                    <Box variant="p" color="text-body-secondary">
                      No provisioned workspaces
                    </Box>
                  </Box>
                }
              />
            ) : (
              <Box textAlign="center" color="inherit" padding={{ vertical: 'xl' }}>
                <Box variant="p" color="text-body-secondary">
                  No provisioned workspaces
                </Box>
              </Box>
            )}
          </Container>
        </SpaceBetween>
      )}
    </Modal>
  )
}
