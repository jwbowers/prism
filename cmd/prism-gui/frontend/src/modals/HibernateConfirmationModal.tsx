import { toast } from 'sonner'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Alert,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'
import type { Instance } from '../lib/types'

interface HibernateConfirmationModalProps {
  instance: Instance
  onDismiss: () => void
  onRefresh: () => void
}

export function HibernateConfirmationModal({ instance, onDismiss, onRefresh }: HibernateConfirmationModalProps) {
  const api = useApi()

  const handleConfirm = async () => {
    onDismiss()
    toast.info(`Hibernating workspace "${instance.name}"...`)
    try {
      await api.hibernateInstance(instance.name)
      await onRefresh()
      toast.success(`Workspace "${instance.name}" hibernated successfully`)
    } catch (error) {
      toast.error(`Failed to hibernate "${instance.name}": ${error instanceof Error ? error.message : String(error)}`)
    }
  }

  return (
    <Modal
      visible={true}
      onDismiss={onDismiss}
      header="Hibernate Workspace?"
      size="medium"
      data-testid="hibernate-confirmation-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={onDismiss}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleConfirm}
              data-testid="confirm-hibernate-button"
            >
              Hibernate
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        <Alert type="info" header="Cost Optimization">
          Hibernating preserves your workspace state for instant resume. You save approximately $0.90/hour in compute costs — only storage charges apply while hibernated (typically 80% cheaper than keeping it running).
        </Alert>
        <Box variant="p">
          Workspace <strong>{instance.name}</strong> will be hibernated. The instance state (RAM contents and running processes) is saved to EBS storage so you can resume exactly where you left off.
        </Box>
        <Box variant="p" color="text-body-secondary">
          Resuming from hibernation is faster than a full stop/start because the state is fully preserved.
        </Box>
      </SpaceBetween>
    </Modal>
  )
}
