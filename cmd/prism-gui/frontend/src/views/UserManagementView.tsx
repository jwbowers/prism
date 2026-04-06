import { useState } from 'react'
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
  ColumnLayout,
  FormField,
  Select,
  Badge,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import { useApi } from '../hooks/use-api'
import { toast } from 'sonner'
import type { User, Instance } from '../lib/types'

interface UserManagementViewProps {
  users: User[]
  instances: Instance[]
  loading: boolean
  onRefresh: () => void
  onCreateUser: () => void
  onEditUser: (user: User) => void
  onViewUserDetails: (user: User) => void
  onViewUserStatus: (user: User) => void
  onProvisionUser: (username: string) => void
  onManageSSHKeys: (username: string) => void
  onDeleteUser: (user: User) => void
}

export function UserManagementView({
  users,
  instances,
  loading,
  onRefresh,
  onCreateUser,
  onEditUser,
  onViewUserDetails,
  onViewUserStatus,
  onProvisionUser,
  onManageSSHKeys,
  onDeleteUser,
}: UserManagementViewProps) {
  const api = useApi()
  const [userStatusFilter, setUserStatusFilter] = useState('all')

  const getFilteredUsers = () => {
    if (userStatusFilter === 'all') {
      return users
    }
    return users.filter(user => {
      if (user.enabled === false) {
        return userStatusFilter === 'inactive'
      }
      const userStatus = user.status?.toLowerCase() || 'active'
      return userStatus === userStatusFilter.toLowerCase()
    })
  }

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="Manage research users with persistent identity across Prism workspaces"
        counter={`(${users.length} users)`}
        actions={
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={onRefresh} disabled={loading}>
              {loading ? <Spinner /> : 'Refresh'}
            </Button>
            <Button
              variant="primary"
              data-testid="create-user-button"
              onClick={onCreateUser}
            >
              Create User
            </Button>
          </SpaceBetween>
        }
      >
        User Management
      </Header>

      {/* User Overview Stats */}
      <ColumnLayout columns={4} variant="text-grid">
        <Container header={<Header variant="h3">Total Users</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-info">
            {users.length}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Active Users</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-success">
            {users.filter(u => u.status !== 'inactive').length}
          </Box>
        </Container>
        <Container header={<Header variant="h3">SSH Keys Generated</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-body-secondary">
            {users.reduce((sum, u) => sum + (u.ssh_keys || 0), 0)}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Provisioned Workspaces</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-warning">
            {users.reduce((sum, u) => sum + (u.provisioned_instances?.length || 0), 0)}
          </Box>
        </Container>
      </ColumnLayout>

      {/* Status Filter */}
      <Container>
        <FormField label="Filter by Status">
          <Select
            selectedOption={{ label: userStatusFilter === 'all' ? 'All Users' : userStatusFilter === 'active' ? 'Active' : 'Inactive', value: userStatusFilter }}
            onChange={({ detail }) => setUserStatusFilter(detail.selectedOption.value || 'all')}
            options={[
              { label: 'All Users', value: 'all' },
              { label: 'Active', value: 'active' },
              { label: 'Inactive', value: 'inactive' }
            ]}
            selectedAriaLabel="Selected"
          />
        </FormField>
      </Container>

      {/* Users Table */}
      <Container
        header={
          <Header
            variant="h2"
            description="Research users with persistent identity and SSH key management"
            counter={`(${getFilteredUsers().length})`}
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Button>Export Users</Button>
                <Button variant="primary">Create User</Button>
              </SpaceBetween>
            }
          >
            Research Users
          </Header>
        }
      >
        <Table
          data-testid="users-table"
          columnDefinitions={[
            {
              id: "username",
              header: "Username",
              cell: (item: User) => <Link fontSize="body-m">{item.username}</Link>,
              sortingField: "username"
            },
            {
              id: "uid",
              header: "UID",
              cell: (item: User) => item.uid.toString(),
              sortingField: "uid"
            },
            {
              id: "full_name",
              header: "Full Name",
              cell: (item: User) => item.full_name || item.display_name || 'Not provided',
              sortingField: "full_name"
            },
            {
              id: "email",
              header: "Email",
              cell: (item: User) => item.email || 'Not provided',
              sortingField: "email"
            },
            {
              id: "ssh_keys",
              header: "SSH Keys",
              cell: (item: User) => {
                const keyCount = item.ssh_keys || 0
                return (
                  <SpaceBetween direction="horizontal" size="xs">
                    <StatusIndicator
                      type={keyCount > 0 ? 'success' : 'warning'}
                    >
                      {keyCount}
                    </StatusIndicator>
                    {keyCount === 0 && (
                      <Badge color="grey">No keys</Badge>
                    )}
                  </SpaceBetween>
                )
              }
            },
            {
              id: "workspaces",
              header: "Workspaces",
              cell: (item: User) => {
                const count = item.provisioned_instances?.length || 0
                return (
                  <span data-testid={`workspace-count-${item.username}`}>
                    {count > 0 ? count.toString() : 'None'}
                  </span>
                )
              }
            },
            {
              id: "status",
              header: "Status",
              cell: (item: User) => {
                const isEnabled = item.enabled !== false
                const displayStatus = !isEnabled ? 'Suspended' : (item.status || 'Active')
                const statusType = !isEnabled ? 'error' : (
                  item.status === 'active' || !item.status ? 'success' : 'warning'
                )
                return (
                  <StatusIndicator
                    type={statusType}
                  >
                    {displayStatus}
                  </StatusIndicator>
                )
              },
              sortingField: "status"
            },
            {
              id: "created",
              header: "Created",
              cell: (item: User) => new Date(item.created_at).toLocaleDateString(),
              sortingField: "created_at"
            },
            {
              id: "actions",
              header: "Actions",
              cell: (item: User) => (
                <ButtonDropdown
                  data-testid={`user-actions-${item.username}`}
                  expandToViewport
                  items={[
                    { text: "View Details", id: "view" },
                    { text: "Generate SSH Key", id: "ssh-key", disabled: (item.ssh_keys || 0) > 0 },
                    { text: "Provision on Workspace", id: "provision" },
                    { text: "User Status", id: "status" },
                    ...(item.enabled !== false
                      ? [{ text: "Disable User", id: "disable" }]
                      : [{ text: "Enable User", id: "enable" }]),
                    { text: "Edit User", id: "edit" },
                    { text: "Delete User", id: "delete" }
                  ]}
                  onItemClick={async (detail) => {
                    if (detail.detail.id === 'view') {
                      onViewUserDetails(item)
                    } else if (detail.detail.id === 'status') {
                      onViewUserStatus(item)
                    } else if (detail.detail.id === 'provision') {
                      onProvisionUser(item.username)
                    } else if (detail.detail.id === 'ssh-key') {
                      onManageSSHKeys(item.username)
                    } else if (detail.detail.id === 'delete') {
                      onDeleteUser(item)
                    } else if (detail.detail.id === 'enable') {
                      try {
                        await api.enableUser(item.username)
                        onRefresh()
                      } catch (error) {
                        toast.error((error as Error).message || 'Failed to enable user')
                      }
                    } else if (detail.detail.id === 'disable') {
                      try {
                        await api.disableUser(item.username)
                        onRefresh()
                      } catch (error) {
                        toast.error((error as Error).message || 'Failed to disable user')
                      }
                    } else if (detail.detail.id === 'edit') {
                      onEditUser(item)
                    }
                  }}
                >
                  Actions
                </ButtonDropdown>
              )
            }
          ]}
          items={getFilteredUsers()}
          loadingText="Loading users..."
          empty={
            <Box textAlign="center" color="text-body-secondary">
              <Box variant="strong" textAlign="center" color="text-body-secondary">
                No users found
              </Box>
              <Box variant="p" padding={{ bottom: 's' }} color="text-body-secondary">
                Create your first research user to enable persistent identity across workspaces.
              </Box>
              <Button variant="primary">Create User</Button>
            </Box>
          }
          header={
            <Header
              counter={`(${users.length})`}
              description="Research users with persistent UID/GID mapping and SSH key management"
            >
              All Users
            </Header>
          }
          pagination={<Box>Showing all {users.length} users</Box>}
        />
      </Container>

      {/* User Analytics and SSH Key Management */}
      <Container
        header={
          <Header
            variant="h2"
            description="User analytics and SSH key management"
          >
            User Analytics
          </Header>
        }
      >
        <ColumnLayout columns={2}>
          <SpaceBetween size="m">
            <Header variant="h3">SSH Key Status</Header>
            {users.length > 0 ? (
              users.map((user) => {
                const keyCount = user.ssh_keys || 0
                return (
                  <Box key={user.username}>
                    <SpaceBetween direction="horizontal" size="s">
                      <Box fontWeight="bold">{user.username}:</Box>
                      <StatusIndicator
                        type={keyCount > 0 ? 'success' : 'warning'}
                      >
                        {keyCount > 0 ? `${keyCount} SSH keys` : 'No SSH keys'}
                      </StatusIndicator>
                      {keyCount === 0 && (
                        <Button variant="link">Generate Key</Button>
                      )}
                    </SpaceBetween>
                  </Box>
                )
              })
            ) : (
              <Box color="text-body-secondary">No users to display</Box>
            )}
          </SpaceBetween>

          <SpaceBetween size="m">
            <Header variant="h3">Workspace Provisioning</Header>
            <Box color="text-body-secondary">
              User provisioning across workspaces and EFS home directory management.
              Persistent identity ensures same UID/GID mapping across all environments.
            </Box>
            {users.length > 0 && (
              <SpaceBetween size="s">
                <Box variant="strong">Available for Provisioning:</Box>
                {instances.length > 0 ? (
                  instances.filter(i => i.state === 'running').map(instance => (
                    <Box key={instance.id}>
                      <StatusIndicator
                        type="success"
                      >
                        {instance.name}
                      </StatusIndicator>
                    </Box>
                  ))
                ) : (
                  <Box color="text-body-secondary">No running workspaces available</Box>
                )}
              </SpaceBetween>
            )}
          </SpaceBetween>
        </ColumnLayout>
      </Container>
    </SpaceBetween>
  )
}
