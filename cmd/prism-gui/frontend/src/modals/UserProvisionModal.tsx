import { useState } from 'react'
import {
  Modal,
  SpaceBetween,
  Box,
  Button,
  FormField,
  Select,
  Alert,
} from '../lib/cloudscape-shim'
import type { User, Instance } from '../lib/types'

interface UserProvisionModalProps {
  visible: boolean
  user: User | null
  instances: Instance[]
  onDismiss: () => void
  onProvision: (username: string, workspaceName: string) => Promise<void>
}

export function UserProvisionModal({ visible, user, instances, onDismiss, onProvision }: UserProvisionModalProps) {
  const [selectedWorkspace, setSelectedWorkspace] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleDismiss = () => {
    setSelectedWorkspace('')
    onDismiss()
  }

  const handleProvision = async () => {
    if (!user || !selectedWorkspace) return
    setSubmitting(true)
    try {
      await onProvision(user.username, selectedWorkspace)
      setSelectedWorkspace('')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      size="medium"
      header={user ? `Provision ${user.username} on Workspace` : 'Provision User'}
      data-testid="provision-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={handleDismiss} disabled={submitting}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleProvision}
              disabled={!selectedWorkspace || submitting}
              loading={submitting}
              data-testid="provision"
            >
              Provision
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      {user && (
        <SpaceBetween size="m">
          <FormField label="Workspace" description="Select a running workspace to provision this user on">
            <Select
              selectedOption={
                selectedWorkspace
                  ? { label: selectedWorkspace, value: selectedWorkspace }
                  : null
              }
              onChange={({ detail }) => setSelectedWorkspace(detail.selectedOption?.value || '')}
              options={instances
                .filter(instance => instance.state === 'running')
                .map(instance => ({
                  label: instance.name,
                  value: instance.name
                }))}
              placeholder="Select a workspace"
              empty="No running workspaces available"
              ariaLabel="Workspace"
            />
          </FormField>

          <Alert type="info">
            Provisioning will create a user account for "{user.username}" on the selected workspace with the same UID/GID and SSH keys.
          </Alert>
        </SpaceBetween>
      )}
    </Modal>
  )
}
