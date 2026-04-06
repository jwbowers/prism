import { useState } from 'react'
import { toast } from 'sonner'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  FormField,
  Input,
  Alert,
  Container,
  ColumnLayout,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'
import type { InstanceSnapshot } from '../lib/types'

interface RestoreBackupModalProps {
  visible: boolean
  backup: InstanceSnapshot | null
  onDismiss: () => void
  onSuccess: () => void
}

export function RestoreBackupModal({ visible, backup, onDismiss, onSuccess }: RestoreBackupModalProps) {
  const api = useApi()
  const [restoreInstanceName, setRestoreInstanceName] = useState('')
  const [loading, setLoading] = useState(false)

  if (!backup) return null

  const sizeGB = backup.size_gb || Math.ceil(backup.storage_cost_monthly / 0.05)

  const handleRestoreBackup = async () => {
    try {
      if (!restoreInstanceName.trim()) {
        toast.error('New instance name is required')
        return
      }

      setLoading(true)
      handleDismiss()

      await api.restoreSnapshot(backup.snapshot_name, restoreInstanceName)

      toast.success(`Backup "${backup.snapshot_name}" is being restored to instance "${restoreInstanceName}". This may take 10-15 minutes.`)
      setLoading(false)

      await onSuccess()
    } catch (error) {
      setLoading(false)
      toast.error(`Restore failed: ${error instanceof Error ? error.message : 'Unknown error occurred'}`)
    }
  }

  const handleDismiss = () => {
    setRestoreInstanceName('')
    onDismiss()
  }

  return (
    <Modal
      onDismiss={handleDismiss}
      visible={visible}
      header="Restore Backup"
      size="medium"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={handleDismiss}>
              Cancel
            </Button>
            <Button
              variant="primary"
              disabled={!restoreInstanceName.trim()}
              onClick={handleRestoreBackup}
              loading={loading}
            >
              Restore
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        <Alert type="info">
          <Box variant="strong">Restore Time</Box>
          <Box>
            Restoring this backup may take 10-15 minutes depending on the backup size ({sizeGB} GB).
            The new instance will be created with all data and configuration from the backup.
          </Box>
        </Alert>

        <Box>
          Restore backup <strong>&quot;{backup.snapshot_name}&quot;</strong> to a new instance.
        </Box>

        <FormField
          label="New instance name"
          description="Choose a name for the restored instance"
          errorText={!restoreInstanceName.trim() ? "Instance name is required" : ""}
        >
          <Input
            value={restoreInstanceName}
            onChange={({ detail }) => setRestoreInstanceName(detail.value)}
            placeholder="restored-instance"
          />
        </FormField>

        <Container>
          <SpaceBetween size="s">
            <Box variant="h4">Backup Details</Box>
            <ColumnLayout columns={2} variant="text-grid">
              <SpaceBetween size="xs">
                <Box variant="awsui-key-label">Source Instance</Box>
                <Box>{backup.source_instance}</Box>
              </SpaceBetween>
              <SpaceBetween size="xs">
                <Box variant="awsui-key-label">Template</Box>
                <Box>{backup.source_template}</Box>
              </SpaceBetween>
              <SpaceBetween size="xs">
                <Box variant="awsui-key-label">Backup Size</Box>
                <Box>{sizeGB} GB</Box>
              </SpaceBetween>
              <SpaceBetween size="xs">
                <Box variant="awsui-key-label">Created</Box>
                <Box>{new Date(backup.created_at).toLocaleDateString()}</Box>
              </SpaceBetween>
            </ColumnLayout>
          </SpaceBetween>
        </Container>

        <Alert type="warning">
          <Box variant="strong">What happens during restore:</Box>
          <ul>
            <li>A new EC2 instance will be launched from this backup</li>
            <li>All files, configurations, and installed software will be restored</li>
            <li>The new instance will have the same specifications as the original</li>
            <li>You can modify the instance after restoration completes</li>
          </ul>
        </Alert>
      </SpaceBetween>
    </Modal>
  )
}
