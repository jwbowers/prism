import {
  Modal,
  ColumnLayout,
  Box,
  Button,
  Spinner,
} from '../lib/cloudscape-shim'
import type { Project, ProjectUsageResponse } from '../lib/types'

interface UsageStatisticsModalProps {
  visible: boolean
  project: Project | null
  usageData: ProjectUsageResponse | null
  loading: boolean
  onDismiss: () => void
}

export function UsageStatisticsModal({ visible, project, usageData, loading, onDismiss }: UsageStatisticsModalProps) {
  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      header={`Usage Statistics — ${project?.name || ''}`}
      footer={<Box float="right"><Button variant="primary" onClick={onDismiss}>Close</Button></Box>}
    >
      {loading ? (
        <Box textAlign="center"><Spinner /> Loading usage data...</Box>
      ) : usageData ? (
        <ColumnLayout columns={2} variant="text-grid">
          <div>
            <Box variant="awsui-key-label">Instance Hours</Box>
            <Box>{(usageData.instance_hours || 0).toFixed(1)} hrs</Box>
          </div>
          <div>
            <Box variant="awsui-key-label">Storage (GB-hours)</Box>
            <Box>{(usageData.storage_gb_hours || 0).toFixed(1)} GB-hrs</Box>
          </div>
          <div>
            <Box variant="awsui-key-label">Data Transfer</Box>
            <Box>{(usageData.data_transfer_gb || 0).toFixed(2)} GB</Box>
          </div>
          <div>
            <Box variant="awsui-key-label">Period</Box>
            <Box>{usageData.period || 'current'}</Box>
          </div>
        </ColumnLayout>
      ) : (
        <Box color="text-body-secondary">No usage data available for this project.</Box>
      )}
    </Modal>
  )
}
