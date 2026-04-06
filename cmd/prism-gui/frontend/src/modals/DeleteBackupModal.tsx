import { useState } from 'react'
import { toast } from 'sonner'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Alert,
  Container,
  ColumnLayout,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'
import type { InstanceSnapshot } from '../lib/types'

interface DeleteBackupModalProps {
  visible: boolean
  backup: InstanceSnapshot | null
  onDismiss: () => void
  onSuccess: () => void
}

export function DeleteBackupModal({ visible, backup, onDismiss, onSuccess }: DeleteBackupModalProps) {
  const api = useApi()
  const [loading, setLoading] = useState(false)

  if (!backup) return null

  const sizeGB = backup.size_gb || Math.ceil(backup.storage_cost_monthly / 0.05)
  const monthlySavings = backup.storage_cost_monthly

  const handleDeleteBackup = async () => {
    try {
      setLoading(true)
      onDismiss()

      await api.deleteSnapshot(backup.snapshot_name)

      toast.success(`Backup "${backup.snapshot_name}" has been deleted. You will save $${backup.storage_cost_monthly.toFixed(2)}/month.`)
      setLoading(false)

      await onSuccess()
    } catch (error) {
      setLoading(false)
      toast.error(`Delete failed: ${error instanceof Error ? error.message : 'Unknown error occurred'}`)
    }
  }

  return (
    <Modal
      onDismiss={onDismiss}
      visible={visible}
      header="Delete Backup Confirmation"
      size="medium"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={onDismiss}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleDeleteBackup}
              loading={loading}
            >
              Delete
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        <Alert type="warning">
          <Box variant="strong">This action cannot be undone</Box>
        </Alert>

        <Box>
          Are you sure you want to delete backup <strong>&quot;{backup.snapshot_name}&quot;</strong>?
        </Box>

        <Container>
          <SpaceBetween size="s">
            <Box variant="h4">Cost Savings</Box>
            <ColumnLayout columns={2} variant="text-grid">
              <SpaceBetween size="xs">
                <Box variant="awsui-key-label">Storage Size</Box>
                <Box>{sizeGB} GB</Box>
              </SpaceBetween>
              <SpaceBetween size="xs">
                <Box variant="awsui-key-label">Monthly Savings</Box>
                <Box color="text-status-success" fontSize="heading-m">
                  <strong>${monthlySavings.toFixed(2)}/month</strong>
                </Box>
              </SpaceBetween>
              <SpaceBetween size="xs">
                <Box variant="awsui-key-label">Annual Savings</Box>
                <Box>${(monthlySavings * 12).toFixed(2)}/year</Box>
              </SpaceBetween>
              <SpaceBetween size="xs">
                <Box variant="awsui-key-label">Free Storage</Box>
                <Box>{sizeGB} GB freed</Box>
              </SpaceBetween>
            </ColumnLayout>
          </SpaceBetween>
        </Container>

        <Box variant="small" color="text-body-secondary">
          <strong>Backup Details:</strong><br/>
          • Source: {backup.source_instance}<br/>
          • Template: {backup.source_template}<br/>
          • Created: {new Date(backup.created_at).toLocaleString()}
        </Box>
      </SpaceBetween>
    </Modal>
  )
}
