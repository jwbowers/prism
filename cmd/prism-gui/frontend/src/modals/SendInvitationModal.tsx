import { useState, useEffect } from 'react'
import {
  Modal,
  SpaceBetween,
  Box,
  Button,
  FormField,
  Input,
  Select,
  Textarea,
} from '../lib/cloudscape-shim'
import { ValidationError } from '../components/ValidationError'
import type { Project } from '../lib/types'

interface SendInvitationModalProps {
  visible: boolean
  projects: Project[]
  selectedProjectId: string
  onProjectChange: (projectId: string) => void
  onDismiss: () => void
  onSubmit: (data: { email: string; role: 'viewer' | 'member' | 'admin'; message: string }) => Promise<void>
}

export function SendInvitationModal({ visible, projects, selectedProjectId, onProjectChange, onDismiss, onSubmit }: SendInvitationModalProps) {
  const [email, setEmail] = useState('')
  const [role, setRole] = useState<'viewer' | 'member' | 'admin'>('member')
  const [message, setMessage] = useState('')
  const [validationError, setValidationError] = useState('')

  // Reset form when modal opens
  useEffect(() => {
    if (visible) {
      setEmail('')
      setRole('member')
      setMessage('')
      setValidationError('')
    }
  }, [visible])

  const handleDismiss = () => {
    setValidationError('')
    onDismiss()
  }

  const handleSubmit = async () => {
    try {
      await onSubmit({ email, role, message })
    } catch (error: any) {
      setValidationError(error.message || 'Failed to send invitation')
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      header="Send Project Invitation"
      data-testid="send-invitation-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={handleDismiss}>Cancel</Button>
            <Button variant="primary" onClick={handleSubmit} data-testid="confirm-send-invitation">
              Send Invitation
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {validationError && (
          <ValidationError message={validationError} visible={true} />
        )}

        <FormField label="Project" description="Select the project to invite user to">
          <Select
            selectedOption={
              selectedProjectId
                ? { label: projects.find(p => p.id === selectedProjectId)?.name || '', value: selectedProjectId }
                : null
            }
            onChange={({ detail }) => onProjectChange(detail.selectedOption?.value || '')}
            options={projects.map(p => ({ label: p.name, value: p.id }))}
            placeholder="Select a project"
            data-testid="invitation-project-select"
          />
        </FormField>

        <FormField label="Email Address" description="Enter the recipient's email address">
          <Input
            value={email}
            onChange={({ detail }) => setEmail(detail.value)}
            placeholder="user@example.com"
            type="email"
            data-testid="invitation-email-input"
          />
        </FormField>

        <FormField label="Role" description="Select the role for this user">
          <Select
            selectedOption={{ label: role, value: role }}
            onChange={({ detail }) => setRole(detail.selectedOption?.value as 'viewer' | 'member' | 'admin')}
            options={[
              { label: 'viewer', value: 'viewer', description: 'Read-only access' },
              { label: 'member', value: 'member', description: 'Can create and manage resources' },
              { label: 'admin', value: 'admin', description: 'Full project control' }
            ]}
            data-testid="invitation-role-select"
          />
        </FormField>

        <FormField label="Message (optional)" description="Add a personal message to the invitation">
          <Textarea
            value={message}
            onChange={({ detail }) => setMessage(detail.value)}
            placeholder="Welcome to the project! Looking forward to collaborating..."
            data-testid="invitation-message-input"
          />
        </FormField>
      </SpaceBetween>
    </Modal>
  )
}
