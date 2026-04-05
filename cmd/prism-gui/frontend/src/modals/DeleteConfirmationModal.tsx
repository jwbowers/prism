import { useState, useEffect } from 'react'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Alert,
  FormField,
  Input,
} from '../lib/cloudscape-shim'

export interface DeleteModalConfig {
  type: 'workspace' | 'efs-volume' | 'ebs-volume' | 'project' | 'user' | null
  name: string
  requireNameConfirmation: boolean
  warning?: string
  onConfirm: () => Promise<void>
}

interface DeleteConfirmationModalProps {
  visible: boolean
  config: DeleteModalConfig
  onDismiss: () => void
}

export function DeleteConfirmationModal({ visible, config, onDismiss }: DeleteConfirmationModalProps) {
  const [confirmationText, setConfirmationText] = useState('')

  // Reset confirmation text whenever the modal opens with a new config
  useEffect(() => {
    if (visible) setConfirmationText('')
  }, [visible, config.name])

  const handleDismiss = () => {
    setConfirmationText('')
    onDismiss()
  }

  const getDeleteMessage = () => {
    switch (config.type) {
      case 'workspace':
        return `You are about to permanently delete the workspace "${config.name}". This action cannot be undone.`
      case 'efs-volume':
        return `You are about to permanently delete the EFS volume "${config.name}". All data on this volume will be lost. This action cannot be undone.`
      case 'ebs-volume':
        return `You are about to permanently delete the EBS volume "${config.name}". All data on this volume will be lost. This action cannot be undone.`
      case 'project':
        return `You are about to permanently delete the project "${config.name}". This action cannot be undone.`
      case 'user':
        return `You are about to permanently delete the user "${config.name}". This action cannot be undone.`
      default:
        return 'This action cannot be undone.'
    }
  }

  const isConfirmationValid = config.requireNameConfirmation
    ? confirmationText === config.name
    : true

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      header={`Delete ${config.type?.replace('-', ' ') || 'Resource'}?`}
      size="medium"
      data-testid="delete-confirmation-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={handleDismiss}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={config.onConfirm}
              disabled={!isConfirmationValid}
              data-testid="confirm-delete-button"
            >
              Delete
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        <Alert type="warning" header="Warning: This action is permanent">
          {getDeleteMessage()}
        </Alert>

        {config.warning && (
          <Alert type="error" header="Additional Warning">
            {config.warning}
          </Alert>
        )}

        {config.requireNameConfirmation && (
          <FormField
            label={`Type "${config.name}" to confirm deletion`}
            description="This extra step helps prevent accidental deletions"
            errorText={
              confirmationText.length > 0 && confirmationText !== config.name
                ? `Name must match exactly: "${config.name}"`
                : ""
            }
          >
            <Input
              value={confirmationText}
              onChange={({ detail }) => setConfirmationText(detail.value)}
              placeholder={config.name}
              ariaRequired
              invalid={confirmationText.length > 0 && confirmationText !== config.name}
            />
          </FormField>
        )}

        <Box variant="p" color="text-body-secondary">
          {config.requireNameConfirmation
            ? 'Enter the exact name above to enable the delete button.'
            : 'Click Delete to confirm this action.'}
        </Box>
      </SpaceBetween>
    </Modal>
  )
}
