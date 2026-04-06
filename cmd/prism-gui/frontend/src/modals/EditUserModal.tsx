import { useState, useEffect } from 'react'
import {
  Modal,
  SpaceBetween,
  Box,
  Button,
  Form,
  FormField,
  Input,
  Select,
} from '../lib/cloudscape-shim'
import type { User } from '../lib/types'

interface EditUserModalProps {
  visible: boolean
  user: User | null
  onDismiss: () => void
  onSubmit: (username: string, updates: { email?: string; display_name?: string; role?: string }) => Promise<void>
}

export function EditUserModal({ visible, user, onDismiss, onSubmit }: EditUserModalProps) {
  const [email, setEmail] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [role, setRole] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Populate form when user changes
  useEffect(() => {
    if (user) {
      setEmail(user.email || '')
      setDisplayName(user.display_name || user.full_name || '')
      setRole('')
    }
  }, [user])

  const handleSubmit = async () => {
    if (!user) return
    setSubmitting(true)
    try {
      const updates: { email?: string; display_name?: string; role?: string } = {}
      if (email) updates.email = email
      if (displayName) updates.display_name = displayName
      if (role) updates.role = role
      await onSubmit(user.username, updates)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      header={`Edit User — ${user?.username || ''}`}
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={onDismiss}>Cancel</Button>
            <Button
              variant="primary"
              loading={submitting}
              onClick={handleSubmit}
            >
              Save Changes
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <Form>
        <SpaceBetween size="m">
          <FormField label="Email">
            <Input
              value={email}
              onChange={({ detail }) => setEmail(detail.value)}
              placeholder="user@example.com"
              type="email"
            />
          </FormField>
          <FormField label="Display Name">
            <Input
              value={displayName}
              onChange={({ detail }) => setDisplayName(detail.value)}
              placeholder="Display name"
            />
          </FormField>
          <FormField label="Role" description="Leave blank to keep current role">
            <Select
              selectedOption={role ? { value: role, label: role } : { value: '', label: 'Keep current role' }}
              onChange={({ detail }) => setRole(detail.selectedOption.value || '')}
              options={[
                { value: '', label: 'Keep current role' },
                { value: 'researcher', label: 'Researcher' },
                { value: 'admin', label: 'Admin' },
                { value: 'viewer', label: 'Viewer' }
              ]}
            />
          </FormField>
        </SpaceBetween>
      </Form>
    </Modal>
  )
}
