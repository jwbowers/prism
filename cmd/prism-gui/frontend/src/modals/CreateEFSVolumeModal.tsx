import { useState, useEffect } from 'react'
import {
  Modal,
  SpaceBetween,
  Box,
  Button,
  Form,
  FormField,
  Input,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'
import { toast } from 'sonner'

interface CreateEFSVolumeModalProps {
  visible: boolean
  onDismiss: () => void
  onSuccess: () => Promise<void>
}

export function CreateEFSVolumeModal({ visible, onDismiss, onSuccess }: CreateEFSVolumeModalProps) {
  const api = useApi()
  const [volumeName, setVolumeName] = useState('')
  const [volumeNameError, setVolumeNameError] = useState('')

  // Reset form when modal opens
  useEffect(() => {
    if (visible) {
      setVolumeName('')
      setVolumeNameError('')
    }
  }, [visible])

  const handleDismiss = () => {
    setVolumeName('')
    setVolumeNameError('')
    onDismiss()
  }

  const handleCreate = async () => {
    if (!volumeName.trim()) {
      setVolumeNameError('Volume name is required')
      return
    }

    const name = volumeName

    // Close modal immediately - don't wait for AWS to finish
    handleDismiss()

    // Show notification that creation is in progress
    const toastId = toast.loading(`Creating EFS volume "${name}"... This may take 1-3 minutes.`)

    // Start creation in background - backend will wait for AWS
    try {
      await api.createEFSVolume(name)
      // Sync volume state from AWS to ensure we have current state
      await api.syncEFSVolume(name)
      await onSuccess()
      toast.success('EFS Volume Created', {
        id: toastId,
        description: `Successfully created EFS volume "${name}"`
      })
    } catch (error) {
      toast.error('Failed to Create EFS Volume', {
        id: toastId,
        description: error instanceof Error ? error.message : 'Unknown error occurred'
      })
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      header="Create EFS Volume"
      size="medium"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={handleDismiss}>
              Cancel
            </Button>
            <Button variant="primary" onClick={handleCreate}>
              Create
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <Form>
        <SpaceBetween size="m">
          {volumeNameError && (
            <Box data-testid="validation-error" color="text-status-error">
              {volumeNameError}
            </Box>
          )}
          <FormField
            label="EFS Volume Name"
            description="Enter a name for your EFS volume"
            errorText={volumeNameError}
          >
            <Input
              value={volumeName}
              onChange={({ detail }) => {
                setVolumeName(detail.value)
                setVolumeNameError('') // Clear error on change
              }}
              placeholder="my-shared-data"
              ariaLabel="EFS Volume Name"
            />
          </FormField>
        </SpaceBetween>
      </Form>
    </Modal>
  )
}
