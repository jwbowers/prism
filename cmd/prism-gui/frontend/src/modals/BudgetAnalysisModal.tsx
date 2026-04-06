import {
  Modal,
  SpaceBetween,
  ColumnLayout,
  Box,
  Button,
  ProgressBar,
  Spinner,
} from '../lib/cloudscape-shim'
import type { Project, BudgetData } from '../lib/types'

interface BudgetAnalysisModalProps {
  visible: boolean
  project: Project | null
  budgetData: BudgetData | null
  loading: boolean
  onDismiss: () => void
}

export function BudgetAnalysisModal({ visible, project, budgetData, loading, onDismiss }: BudgetAnalysisModalProps) {
  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      header={`Budget Analysis — ${project?.name || ''}`}
      footer={<Box float="right"><Button variant="primary" onClick={onDismiss}>Close</Button></Box>}
    >
      {loading ? (
        <Box textAlign="center"><Spinner /> Loading budget data...</Box>
      ) : budgetData ? (
        <SpaceBetween size="m">
          <ProgressBar
            value={Math.min((budgetData.spent_percentage || 0) * 100, 100)}
            status={
              (budgetData.spent_percentage || 0) >= 0.95 ? 'error' :
              (budgetData.spent_percentage || 0) >= 0.80 ? 'in-progress' : 'success'
            }
            label="Budget utilization"
            description={`${((budgetData.spent_percentage || 0) * 100).toFixed(1)}% used`}
          />
          <ColumnLayout columns={3} variant="text-grid">
            <div>
              <Box variant="awsui-key-label">Budget Limit</Box>
              <Box>${(budgetData.total_budget || 0).toFixed(2)}</Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Spent to Date</Box>
              <Box>${(budgetData.spent_amount || 0).toFixed(2)}</Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Remaining</Box>
              <Box>${(budgetData.remaining || 0).toFixed(2)}</Box>
            </div>
          </ColumnLayout>
          {budgetData.projected_monthly_spend !== undefined && (
            <ColumnLayout columns={2} variant="text-grid">
              <div>
                <Box variant="awsui-key-label">Projected Monthly Spend</Box>
                <Box>${budgetData.projected_monthly_spend.toFixed(2)}</Box>
              </div>
              {budgetData.days_until_exhausted !== undefined && (
                <div>
                  <Box variant="awsui-key-label">Days Until Exhausted</Box>
                  <Box>{budgetData.days_until_exhausted} days</Box>
                </div>
              )}
            </ColumnLayout>
          )}
          {budgetData.alert_count > 0 && (
            <Box color="text-status-warning">
              {budgetData.alert_count} budget alert(s) active
            </Box>
          )}
        </SpaceBetween>
      ) : (
        <Box color="text-body-secondary">No budget data available for this project.</Box>
      )}
    </Modal>
  )
}
