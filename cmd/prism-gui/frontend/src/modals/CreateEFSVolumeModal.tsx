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

interface CreateEFSVolumeModalProps {
  visible: boolean
  onDismiss: () => void
  onSuccess: () => Promise<void>
  onNotify: (notification: { type: string; header: string; content: string; dismissible: boolean; id: string }) => void
}

export function CreateEFSVolumeModal({ visible, onDismiss, onSuccess, onNotify }: CreateEFSVolumeModalProps) {
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
    onNotify({
      type: 'info',
      header: 'Creating EFS Volume',
      content: `Creating EFS volume "${name}"... This may take 1-3 minutes.`,
      dismissible: true,
      id: Date.now().toString()
    })

    // Start creation in background - backend will wait for AWS
    try {
      await api.createEFSVolume(name)
      // Sync volume state from AWS to ensure we have current state
      await api.syncEFSVolume(name)
      await onSuccess()
      onNotify({
        type: 'success',
        header: 'EFS Volume Created',
        content: `Successfully created EFS volume "${name}"`,
        dismissible: true,
        id: Date.now().toString()
      })
    } catch (error) {
      onNotify({
        type: 'error',
        header: 'Failed to Create EFS Volume',
        content: error instanceof Error ? error.message : 'Unknown error occurred',
        dismissible: true,
        id: Date.now().toString()
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
