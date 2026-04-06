import { useState } from 'react'
import { toast } from 'sonner'
import {
  Modal,
  SpaceBetween,
  Container,
  Header,
  Box,
  Button,
  Table,
  Badge,
  FormField,
  Input,
  Select,
  Spinner,
} from '../lib/cloudscape-shim'
import type { Project, MemberData } from '../lib/types'
import { useApi } from '../hooks/use-api'

interface ManageMembersModalProps {
  visible: boolean
  project: Project | null
  members: MemberData[]
  loading: boolean
  onDismiss: () => void
  onMembersChange: (members: MemberData[]) => void
}

export function ManageMembersModal({ visible, project, members, loading, onDismiss, onMembersChange }: ManageMembersModalProps) {
  const api = useApi()
  const [addMemberUsername, setAddMemberUsername] = useState('')
  const [addMemberRole, setAddMemberRole] = useState('member')

  const handleDismiss = () => {
    setAddMemberUsername('')
    setAddMemberRole('member')
    onDismiss()
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      header={`Manage Members — ${project?.name || ''}`}
      size="large"
      footer={
        <Box float="right">
          <Button variant="primary" onClick={handleDismiss}>Close</Button>
        </Box>
      }
    >
      <SpaceBetween size="l">
        {loading ? (
          <Box textAlign="center"><Spinner /> Loading members...</Box>
        ) : (
          <Table
            columnDefinitions={[
              {
                id: 'username',
                header: 'Username',
                cell: (member: MemberData) => member.username || member.user_id
              },
              {
                id: 'role',
                header: 'Role',
                cell: (member: MemberData) => (
                  <Badge color={member.role === 'admin' ? 'red' : member.role === 'member' ? 'blue' : 'grey'}>
                    {member.role}
                  </Badge>
                )
              },
              {
                id: 'joined_at',
                header: 'Joined',
                cell: (member: MemberData) => member.joined_at ? new Date(member.joined_at).toLocaleDateString() : '-'
              },
              {
                id: 'actions',
                header: 'Actions',
                cell: (member: MemberData) => (
                  <SpaceBetween direction="horizontal" size="xs">
                    <Select
                      selectedOption={{ value: member.role, label: member.role }}
                      onChange={async ({ detail }) => {
                        if (!project) return
                        try {
                          await api.updateProjectMember(project.id, member.user_id, { role: detail.selectedOption.value })
                          const updated = await api.getProjectMembers(project.id)
                          onMembersChange(updated)
                        } catch (error: any) {
                          toast.error('Update Failed', { description: error.message || 'Failed to update role' })
                        }
                      }}
                      options={[
                        { value: 'viewer', label: 'Viewer' },
                        { value: 'member', label: 'Member' },
                        { value: 'admin', label: 'Admin' }
                      ]}
                    />
                    <Button
                      variant="link"
                      onClick={async () => {
                        if (!project) return
                        try {
                          await api.removeProjectMember(project.id, member.user_id)
                          const updated = await api.getProjectMembers(project.id)
                          onMembersChange(updated)
                        } catch (error: any) {
                          toast.error('Remove Failed', { description: error.message || 'Failed to remove member' })
                        }
                      }}
                    >
                      Remove
                    </Button>
                  </SpaceBetween>
                )
              }
            ]}
            items={members}
            empty={<Box textAlign="center">No members yet.</Box>}
            header={<Header variant="h3">Current Members</Header>}
          />
        )}
        <Container header={<Header variant="h3">Add Member</Header>}>
          <SpaceBetween direction="horizontal" size="xs">
            <FormField label="Username">
              <Input
                value={addMemberUsername}
                onChange={({ detail }) => setAddMemberUsername(detail.value)}
                placeholder="username"
              />
            </FormField>
            <FormField label="Role">
              <Select
                selectedOption={{ value: addMemberRole, label: addMemberRole }}
                onChange={({ detail }) => setAddMemberRole(detail.selectedOption.value || 'member')}
                options={[
                  { value: 'viewer', label: 'Viewer' },
                  { value: 'member', label: 'Member' },
                  { value: 'admin', label: 'Admin' }
                ]}
              />
            </FormField>
            <Box padding={{ top: 'xl' }}>
              <Button
                variant="primary"
                disabled={!addMemberUsername.trim()}
                onClick={async () => {
                  if (!project || !addMemberUsername.trim()) return
                  try {
                    await api.addProjectMember(project.id, { user_id: addMemberUsername, role: addMemberRole })
                    const updated = await api.getProjectMembers(project.id)
                    onMembersChange(updated)
                    setAddMemberUsername('')
                    setAddMemberRole('member')
                  } catch (error: any) {
                    toast.error('Add Failed', { description: error.message || 'Failed to add member' })
                  }
                }}
              >
                Add Member
              </Button>
            </Box>
          </SpaceBetween>
        </Container>
      </SpaceBetween>
    </Modal>
  )
}
