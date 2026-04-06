import { useState, useEffect } from 'react'
import {
  Modal,
  SpaceBetween,
  Box,
  Button,
  Form,
  FormField,
  Input,
  Select,
  Textarea,
} from '../lib/cloudscape-shim'
import type { Project } from '../lib/types'

interface EditProjectModalProps {
  visible: boolean
  project: Project | null
  onDismiss: () => void
  onSubmit: (projectId: string, data: { name: string; description: string; status: string }) => Promise<void>
}

export function EditProjectModal({ visible, project, onDismiss, onSubmit }: EditProjectModalProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Populate form when project changes
  useEffect(() => {
    if (project) {
      setName(project.name)
      setDescription(project.description || '')
      setStatus(project.status)
    }
  }, [project])

  const handleSubmit = async () => {
    if (!project) return
    setSubmitting(true)
    try {
      await onSubmit(project.id, { name, description, status })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      header="Edit Project"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={onDismiss}>Cancel</Button>
            <Button
              variant="primary"
              loading={submitting}
              onClick={handleSubmit}
            >
              Save Changes
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <Form>
        <SpaceBetween size="m">
          <FormField label="Project Name">
            <Input
              value={name}
              onChange={({ detail }) => setName(detail.value)}
              placeholder="Project name"
            />
          </FormField>
          <FormField label="Description">
            <Textarea
              value={description}
              onChange={({ detail }) => setDescription(detail.value)}
              placeholder="Project description"
              rows={3}
            />
          </FormField>
          <FormField label="Status">
            <Select
              selectedOption={{ value: status, label: status }}
              onChange={({ detail }) => setStatus(detail.selectedOption.value || 'active')}
              options={[
                { value: 'active', label: 'Active' },
                { value: 'paused', label: 'Paused' },
                { value: 'completed', label: 'Completed' },
                { value: 'archived', label: 'Archived' }
              ]}
            />
          </FormField>
        </SpaceBetween>
      </Form>
    </Modal>
  )
}
