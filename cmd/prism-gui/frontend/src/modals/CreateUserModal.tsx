import { useState } from 'react'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  FormField,
  Input,
} from '../lib/cloudscape-shim'
import { ValidationError } from '../components/ValidationError'

interface CreateUserModalProps {
  visible: boolean
  onDismiss: () => void
  onSubmit: (data: { username: string; email: string; fullName: string }) => Promise<void>
}

export function CreateUserModal({ visible, onDismiss, onSubmit }: CreateUserModalProps) {
  const [username, setUsername] = useState('')
  const [userEmail, setUserEmail] = useState('')
  const [userFullName, setUserFullName] = useState('')
  const [validationError, setValidationError] = useState('')
  const [creating, setCreating] = useState(false)

  const handleDismiss = () => {
    setValidationError('')
    onDismiss()
  }

  const handleSubmit = async () => {
    setValidationError('')
    setCreating(true)
    try {
      await onSubmit({ username, email: userEmail, fullName: userFullName })
      setUsername('')
      setUserEmail('')
      setUserFullName('')
      handleDismiss()
    } catch (error) {
      setValidationError((error as Error).message || 'Failed to create user')
    } finally {
      setCreating(false)
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      header="Create New User"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={handleDismiss} disabled={creating}>Cancel</Button>
            <Button variant="primary" onClick={handleSubmit} disabled={creating} loading={creating}>
              {creating ? 'Creating...' : 'Create'}
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {validationError && (
          <ValidationError message={validationError} visible={true} />
        )}
        <FormField label="Username" description="Unique username for the user">
          <Input
            data-testid="user-username-input"
            value={username}
            onChange={({ detail }) => setUsername(detail.value)}
            placeholder="e.g., jsmith"
          />
        </FormField>
        <FormField label="Email" description="User's email address">
          <Input
            data-testid="user-email-input"
            type="email"
            value={userEmail}
            onChange={({ detail }) => setUserEmail(detail.value)}
            placeholder="user@example.com"
          />
        </FormField>
        <FormField label="Full Name" description="User's full name">
          <Input
            data-testid="user-fullname-input"
            value={userFullName}
            onChange={({ detail }) => setUserFullName(detail.value)}
            placeholder="John Smith"
          />
        </FormField>
      </SpaceBetween>
    </Modal>
  )
}
