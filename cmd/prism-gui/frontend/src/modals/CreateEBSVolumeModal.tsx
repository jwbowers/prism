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

interface CreateEBSVolumeModalProps {
  visible: boolean
  onDismiss: () => void
  onSuccess: () => Promise<void>
  onNotify: (notification: { type: string; header: string; content: string; dismissible: boolean; id: string }) => void
}

export function CreateEBSVolumeModal({ visible, onDismiss, onSuccess, onNotify }: CreateEBSVolumeModalProps) {
  const api = useApi()
  const [volumeName, setVolumeName] = useState('')
  const [volumeSize, setVolumeSize] = useState('')
  const [volumeNameError, setVolumeNameError] = useState('')
  const [volumeSizeError, setVolumeSizeError] = useState('')

  // Reset form when modal opens
  useEffect(() => {
    if (visible) {
      setVolumeName('')
      setVolumeSize('')
      setVolumeNameError('')
      setVolumeSizeError('')
    }
  }, [visible])

  const handleDismiss = () => {
    setVolumeName('')
    setVolumeSize('')
    setVolumeNameError('')
    setVolumeSizeError('')
    onDismiss()
  }

  const handleCreate = async () => {
    // Validate volume name
    if (!volumeName.trim()) {
      setVolumeNameError('Volume name is required')
      return
    }

    // Validate volume size
    if (!volumeSize.trim()) {
      setVolumeSizeError('Volume size is required')
      return
    }

    const sizeNum = parseInt(volumeSize)
    if (isNaN(sizeNum) || sizeNum <= 0) {
      setVolumeSizeError('Volume size must be a positive number')
      return
    }

    const name = volumeName
    const size = volumeSize

    // Close modal immediately - don't wait for AWS to finish
    handleDismiss()

    // Show notification that creation is in progress
    onNotify({
      type: 'info',
      header: 'Creating EBS Volume',
      content: `Creating EBS volume "${name}" (${size} GB)... This may take 30-120 seconds.`,
      dismissible: true,
      id: Date.now().toString()
    })

    // Start creation in background - backend will wait for AWS
    try {
      await api.createEBSVolume(name, size)
      // Sync volume state from AWS to ensure we have current state
      await api.syncEBSVolume(name)
      await onSuccess()
      onNotify({
        type: 'success',
        header: 'EBS Volume Created',
        content: `Successfully created EBS volume "${name}"`,
        dismissible: true,
        id: Date.now().toString()
      })
    } catch (error) {
      onNotify({
        type: 'error',
        header: 'Failed to Create EBS Volume',
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
      header="Create EBS Volume"
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
          {(volumeNameError || volumeSizeError) && (
            <Box data-testid="validation-error" color="text-status-error">
              {volumeNameError || volumeSizeError}
            </Box>
          )}
          <FormField
            label="EBS Volume Name"
            description="Enter a name for your EBS volume"
            errorText={volumeNameError}
          >
            <Input
              value={volumeName}
              onChange={({ detail }) => {
                setVolumeName(detail.value)
                setVolumeNameError('') // Clear error on change
              }}
              placeholder="my-private-data"
              ariaLabel="EBS Volume Name"
            />
          </FormField>
          <FormField
            label="EBS Volume Size (GB)"
            description="Enter the size of the volume in gigabytes"
            errorText={volumeSizeError}
          >
            <Input
              value={volumeSize}
              onChange={({ detail }) => {
                setVolumeSize(detail.value)
                setVolumeSizeError('') // Clear error on change
              }}
              placeholder="100"
              type="number"
              ariaLabel="EBS Volume Size"
            />
          </FormField>
        </SpaceBetween>
      </Form>
    </Modal>
  )
}
