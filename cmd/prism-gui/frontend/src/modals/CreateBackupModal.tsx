import { useState } from 'react'
import { toast } from 'sonner'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Form,
  FormField,
  Input,
  Select,
  Alert,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'
import type { Instance } from '../lib/types'

interface CreateBackupModalProps {
  visible: boolean
  instances: Instance[]
  onDismiss: () => void
  onSuccess: () => void
}

export function CreateBackupModal({ visible, instances, onDismiss, onSuccess }: CreateBackupModalProps) {
  const api = useApi()

  const [createBackupConfig, setCreateBackupConfig] = useState({
    instanceId: '',
    backupName: '',
    backupType: 'full',
    description: ''
  })
  const [createBackupValidationAttempted, setCreateBackupValidationAttempted] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleCreateBackup = async () => {
    try {
      setCreateBackupValidationAttempted(true)

      if (!createBackupConfig.instanceId || !createBackupConfig.backupName) {
        toast.error('Instance and backup name are required')
        return
      }

      setLoading(true)
      handleDismiss()

      await api.createSnapshot(
        createBackupConfig.instanceId,
        createBackupConfig.backupName,
        createBackupConfig.description
      )

      toast.success(`Backup "${createBackupConfig.backupName}" is being created. This may take 5-10 minutes.`)
      setLoading(false)

      await onSuccess()
    } catch (error) {
      setLoading(false)
      toast.error(`Backup creation failed: ${error instanceof Error ? error.message : 'Unknown error occurred'}`)
    }
  }

  const handleDismiss = () => {
    setCreateBackupValidationAttempted(false)
    setCreateBackupConfig({
      instanceId: '',
      backupName: '',
      backupType: 'full',
      description: ''
    })
    onDismiss()
  }

  return (
    <Modal
      onDismiss={handleDismiss}
      visible={visible}
      header="Create Backup"
      size="medium"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={handleDismiss}>
              Cancel
            </Button>
            <Button
              variant="primary"
              disabled={!createBackupConfig.instanceId || !createBackupConfig.backupName.trim()}
              onClick={handleCreateBackup}
              loading={loading}
              data-testid="create-backup-submit"
            >
              Create
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <Form>
        <SpaceBetween size="m">
          <FormField
            label="Instance"
            description="Select the workspace instance to backup"
            errorText={createBackupValidationAttempted && !createBackupConfig.instanceId ? "Instance is required" : ""}
          >
            <Select
              data-testid="instance-select"
              selectedOption={
                createBackupConfig.instanceId
                  ? {
                      label: instances.find(i => i.name === createBackupConfig.instanceId)?.name || createBackupConfig.instanceId,
                      value: createBackupConfig.instanceId
                    }
                  : null
              }
              onChange={({ detail }) =>
                setCreateBackupConfig(prev => ({ ...prev, instanceId: detail.selectedOption.value || '' }))
              }
              options={instances.map(instance => ({
                label: `${instance.name} (${instance.template || 'Unknown template'})`,
                value: instance.name
              }))}
              placeholder="Select an instance..."
              empty="No instances available"
              ariaLabel="Instance"
            />
          </FormField>

          <FormField
            label="Backup name"
            description="Choose a descriptive name for this backup"
            errorText={createBackupValidationAttempted && !createBackupConfig.backupName.trim() ? "Backup name is required" : ""}
          >
            <Input
              value={createBackupConfig.backupName}
              onChange={({ detail }) =>
                setCreateBackupConfig(prev => ({ ...prev, backupName: detail.value }))
              }
              placeholder="my-backup-2024-11-16"
            />
          </FormField>

          <FormField
            label="Backup type"
            description="Full backups capture the entire instance state"
          >
            <Select
              selectedOption={{ label: "Full backup", value: "full" }}
              onChange={({ detail }) =>
                setCreateBackupConfig(prev => ({ ...prev, backupType: detail.selectedOption.value || 'full' }))
              }
              options={[
                { label: "Full backup", value: "full" },
                { label: "Incremental backup", value: "incremental" }
              ]}
            />
          </FormField>

          <FormField
            label="Description (optional)"
            description="Add notes about this backup"
          >
            <Input
              value={createBackupConfig.description}
              onChange={({ detail }) =>
                setCreateBackupConfig(prev => ({ ...prev, description: detail.value }))
              }
              placeholder="Weekly backup before major update"
            />
          </FormField>

          {createBackupConfig.instanceId && (
            <Alert type="info">
              <SpaceBetween size="s">
                <Box variant="strong">Backup Information</Box>
                <Box>
                  • <strong>Creation time:</strong> 5-10 minutes depending on instance size
                </Box>
                <Box>
                  • <strong>Cost:</strong> ~$0.05/GB/month for snapshot storage
                </Box>
                <Box>
                  • <strong>Instance continues running:</strong> No downtime during backup creation
                </Box>
              </SpaceBetween>
            </Alert>
          )}

          {createBackupValidationAttempted && (!createBackupConfig.instanceId || !createBackupConfig.backupName.trim()) && (
            <div data-testid="validation-error">
              {!createBackupConfig.instanceId ? "Instance is required" : !createBackupConfig.backupName.trim() ? "Backup name is required" : ""}
            </div>
          )}
        </SpaceBetween>
      </Form>
    </Modal>
  )
}
