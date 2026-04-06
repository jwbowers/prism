import { useState, useEffect } from 'react'
import {
  Modal,
  SpaceBetween,
  Box,
  Button,
  FormField,
  Input,
  Alert,
} from '../lib/cloudscape-shim'
import { ValidationError } from '../components/ValidationError'

interface RedeemTokenModalProps {
  visible: boolean
  onDismiss: () => void
  onSubmit: (token: string) => Promise<void>
}

export function RedeemTokenModal({ visible, onDismiss, onSubmit }: RedeemTokenModalProps) {
  const [token, setToken] = useState('')
  const [validationError, setValidationError] = useState('')

  // Reset form when modal opens
  useEffect(() => {
    if (visible) {
      setToken('')
      setValidationError('')
    }
  }, [visible])

  const handleDismiss = () => {
    setValidationError('')
    onDismiss()
  }

  const handleSubmit = async () => {
    try {
      await onSubmit(token)
    } catch (error: any) {
      setValidationError(error.message || 'Failed to redeem token')
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleDismiss}
      header="Redeem Invitation Token"
      data-testid="redeem-token-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={handleDismiss}>Cancel</Button>
            <Button variant="primary" onClick={handleSubmit} data-testid="confirm-redeem-token">
              Redeem
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {validationError && (
          <ValidationError message={validationError} visible={true} />
        )}

        <FormField label="Invitation Token" description="Enter the invitation token you received">
          <Input
            value={token}
            onChange={({ detail }) => setToken(detail.value)}
            placeholder="Enter token..."
            data-testid="invitation-token-input"
          />
        </FormField>

        <Alert type="info">
          The token can be found in your invitation email or shared by the project admin.
        </Alert>
      </SpaceBetween>
    </Modal>
  )
}
