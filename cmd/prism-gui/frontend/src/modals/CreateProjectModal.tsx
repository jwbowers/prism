import { useState } from 'react'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  FormField,
  Input,
  Textarea,
  Alert,
} from '../lib/cloudscape-shim'

interface CreateProjectModalProps {
  visible: boolean
  onDismiss: () => void
  onSubmit: (data: { name: string; description: string; budget: string }) => Promise<void>
}

export function CreateProjectModal({ visible, onDismiss, onSubmit }: CreateProjectModalProps) {
  const [projectName, setProjectName] = useState('')
  const [projectDescription, setProjectDescription] = useState('')
  const [projectBudget, setProjectBudget] = useState('')
  const [validationError, setValidationError] = useState('')

  const handleDismiss = () => {
    setValidationError('')
    onDismiss()
  }

  const handleSubmit = async () => {
    setValidationError('')
    try {
      await onSubmit({ name: projectName, description: projectDescription, budget: projectBudget })
      setProjectName('')
      setProjectDescription('')
      setProjectBudget('')
      handleDismiss()
    } catch (error) {
      setValidationError((error as Error).message || 'Failed to create project')
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      header="Create New Project"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={handleDismiss}>Cancel</Button>
            <Button variant="primary" data-testid="create-project-submit-button" onClick={handleSubmit}>Create</Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {validationError && (
          <Alert type="error" data-testid="validation-error">
            {validationError}
          </Alert>
        )}
        <FormField label="Project Name" description="Unique identifier for the project">
          <Input
            data-testid="project-name-input"
            value={projectName}
            onChange={({ detail }) => setProjectName(detail.value)}
            placeholder="e.g., ML Research 2024"
          />
        </FormField>
        <FormField label="Description" description="Brief description of the project">
          <Textarea
            data-testid="project-description-input"
            value={projectDescription}
            onChange={({ detail }) => setProjectDescription(detail.value)}
            placeholder="Describe the project purpose..."
          />
        </FormField>
        <FormField label="Budget Limit (optional)" description="Maximum spending limit in USD">
          <Input
            data-testid="project-budget-input"
            type="number"
            value={projectBudget}
            onChange={({ detail }) => setProjectBudget(detail.value)}
            placeholder="1000.00"
          />
        </FormField>
      </SpaceBetween>
    </Modal>
  )
}
