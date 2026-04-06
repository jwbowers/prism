import { useState } from 'react'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  FormField,
  Input,
  Textarea,
  Select,
  Alert,
} from '../lib/cloudscape-shim'

interface CreateBudgetModalProps {
  visible: boolean
  onDismiss: () => void
  onSubmit: (data: {
    name: string
    description: string
    totalAmount: string
    period: string
    alertThreshold: string
  }) => Promise<void>
}

export function CreateBudgetModal({ visible, onDismiss, onSubmit }: CreateBudgetModalProps) {
  const [budgetName, setBudgetName] = useState('')
  const [budgetDescription, setBudgetDescription] = useState('')
  const [totalAmount, setTotalAmount] = useState('')
  const [period, setPeriod] = useState('monthly')
  const [alertThreshold, setAlertThreshold] = useState('80')
  const [validationError, setValidationError] = useState('')

  const handleDismiss = () => {
    setValidationError('')
    onDismiss()
  }

  const handleSubmit = async () => {
    setValidationError('')
    try {
      await onSubmit({
        name: budgetName,
        description: budgetDescription,
        totalAmount,
        period,
        alertThreshold,
      })
      setBudgetName('')
      setBudgetDescription('')
      setTotalAmount('')
      setPeriod('monthly')
      setAlertThreshold('80')
      handleDismiss()
    } catch (error) {
      setValidationError((error as Error).message || 'Failed to create budget')
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      header="Create Budget Pool"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={handleDismiss}>Cancel</Button>
            <Button variant="primary" data-testid="create-budget-submit-button" onClick={handleSubmit}>Create Budget</Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {validationError && (
          <Alert type="error" data-testid="budget-validation-error">
            {validationError}
          </Alert>
        )}
        <FormField label="Budget Name" description="E.g., 'NSF Grant CISE-2024-12345'">
          <Input
            data-testid="budget-name-input"
            value={budgetName}
            onChange={({ detail }) => setBudgetName(detail.value)}
            placeholder="Enter budget pool name"
          />
        </FormField>
        <FormField label="Description" description="Brief description of the funding source">
          <Textarea
            data-testid="budget-description-input"
            value={budgetDescription}
            onChange={({ detail }) => setBudgetDescription(detail.value)}
            placeholder="Describe the budget source..."
          />
        </FormField>
        <FormField label="Total Amount (USD)" description="Total funding available">
          <Input
            data-testid="budget-amount-input"
            value={totalAmount}
            onChange={({ detail }) => setTotalAmount(detail.value)}
            type="number"
            placeholder="50000.00"
          />
        </FormField>
        <FormField label="Budget Period" description="Timeframe for this budget">
          <Select
            data-testid="budget-period-select"
            selectedOption={{ label: period.charAt(0).toUpperCase() + period.slice(1), value: period }}
            onChange={({ detail }) => setPeriod(detail.selectedOption.value!)}
            options={[
              { label: 'Monthly', value: 'monthly' },
              { label: 'Quarterly', value: 'quarterly' },
              { label: 'Yearly', value: 'yearly' },
              { label: 'Project Lifetime', value: 'project' }
            ]}
          />
        </FormField>
        <FormField label="Alert Threshold (%)" description="Alert when spending exceeds this percentage">
          <Input
            data-testid="budget-threshold-input"
            value={alertThreshold}
            onChange={({ detail }) => setAlertThreshold(detail.value)}
            type="number"
            placeholder="80"
          />
        </FormField>
      </SpaceBetween>
    </Modal>
  )
}
