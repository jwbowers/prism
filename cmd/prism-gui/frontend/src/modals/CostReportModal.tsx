import {
  Modal,
  SpaceBetween,
  Box,
  Button,
  Table,
  Spinner,
} from '../lib/cloudscape-shim'
import type { Project, CostBreakdown } from '../lib/types'

interface CostReportModalProps {
  visible: boolean
  project: Project | null
  costData: CostBreakdown | null
  loading: boolean
  onDismiss: () => void
}

export function CostReportModal({ visible, project, costData, loading, onDismiss }: CostReportModalProps) {
  const handleExportCSV = () => {
    if (!costData) return
    const rows = [
      ['Service', 'Amount ($)'],
      ['Instances', costData.instances?.toFixed(2) ?? '0.00'],
      ['Storage', costData.storage?.toFixed(2) ?? '0.00'],
      ['Data Transfer', costData.data_transfer?.toFixed(2) ?? '0.00'],
      ['Total', costData.total?.toFixed(2) ?? '0.00']
    ]
    const csv = rows.map(r => r.join(',')).join('\n')
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `cost-report-${project?.name || 'project'}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      header={`Cost Report — ${project?.name || ''}`}
      size="medium"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            {costData && (
              <Button onClick={handleExportCSV}>
                Export CSV
              </Button>
            )}
            <Button variant="primary" onClick={onDismiss}>Close</Button>
          </SpaceBetween>
        </Box>
      }
    >
      {loading ? (
        <Box textAlign="center"><Spinner /> Loading cost data...</Box>
      ) : costData ? (
        <SpaceBetween size="m">
          <Table
            columnDefinitions={[
              { id: 'service', header: 'Service', cell: (item: { service: string; amount: number }) => item.service },
              { id: 'amount', header: 'Amount', cell: (item: { service: string; amount: number }) => `$${item.amount.toFixed(2)}` }
            ]}
            items={[
              { service: 'Instances (EC2)', amount: costData.instances || 0 },
              { service: 'Storage', amount: costData.storage || 0 },
              { service: 'Data Transfer', amount: costData.data_transfer || 0 }
            ]}
            footer={
              <Box textAlign="right" fontWeight="bold">
                Total: ${(costData.total || 0).toFixed(2)}
              </Box>
            }
          />
        </SpaceBetween>
      ) : (
        <Box color="text-body-secondary">No cost data available for this project.</Box>
      )}
    </Modal>
  )
}
