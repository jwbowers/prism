import { useState, useEffect } from 'react'
import { toast } from 'sonner'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Alert,
  Table,
  Spinner,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'
import type { IdlePolicy } from '../lib/types'

interface IdlePolicyModalProps {
  instanceName: string
  onDismiss: () => void
}

export function IdlePolicyModal({ instanceName, onDismiss }: IdlePolicyModalProps) {
  const api = useApi()
  const [availablePolicies, setAvailablePolicies] = useState<IdlePolicy[]>([])
  const [appliedPolicies, setAppliedPolicies] = useState<IdlePolicy[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    Promise.all([
      api.getIdlePolicies(),
      api.getInstanceIdlePolicies(instanceName)
    ]).then(([all, applied]) => {
      setAvailablePolicies(all)
      setAppliedPolicies(applied)
    }).catch(err => {
      console.error('Failed to load idle policies:', err)
    }).finally(() => setLoading(false))
  }, [instanceName])

  const appliedIds = new Set(appliedPolicies.map(p => p.id))

  const handleApply = async (policyId: string, policyName: string) => {
    onDismiss()
    toast(`Applying Idle Policy`, { description: `Applying "${policyName}" to ${instanceName}...` })
    try {
      await api.applyIdlePolicy(instanceName, policyId)
      toast.success('Idle Policy Applied', { description: `"${policyName}" applied to ${instanceName}` })
    } catch (error) {
      toast.error('Failed to Apply Policy', { description: `${error instanceof Error ? error.message : String(error)}` })
    }
  }

  const handleRemove = async (policyId: string, policyName: string) => {
    onDismiss()
    toast('Removing Idle Policy', { description: `Removing "${policyName}" from ${instanceName}...` })
    try {
      await api.removeIdlePolicy(instanceName, policyId)
      toast.success('Idle Policy Removed', { description: `"${policyName}" removed from ${instanceName}` })
    } catch (error) {
      toast.error('Failed to Remove Policy', { description: `${error instanceof Error ? error.message : String(error)}` })
    }
  }

  return (
    <Modal
      visible={true}
      onDismiss={onDismiss}
      header={`Manage Idle Policy — ${instanceName}`}
      size="medium"
      data-testid="idle-policy-modal"
      footer={
        <Box float="right">
          <Button variant="link" onClick={onDismiss}>Close</Button>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {appliedPolicies.length > 0 && (
          <Alert type="info" header="Currently Applied">
            {appliedPolicies.map(p => (
              <SpaceBetween key={p.id} direction="horizontal" size="xs">
                <Box variant="span"><strong>{p.name}</strong> — {p.description}</Box>
                <Button
                  variant="link"
                  onClick={() => handleRemove(p.id, p.name)}
                  data-testid={`remove-idle-policy-${p.id}`}
                >
                  Remove
                </Button>
              </SpaceBetween>
            ))}
          </Alert>
        )}
        {loading ? (
          <Box textAlign="center"><Spinner /></Box>
        ) : (
          <Table
            data-testid="idle-policy-templates-table"
            columnDefinitions={[
              { id: 'name', header: 'Policy', cell: (item: IdlePolicy) => item.name },
              { id: 'description', header: 'Description', cell: (item: IdlePolicy) => item.description || '' },
              { id: 'savings', header: 'Est. Savings', cell: (item: IdlePolicy) => `${(item as unknown as Record<string, unknown>).estimated_savings_percent ?? '—'}%` },
              {
                id: 'action',
                header: '',
                cell: (item: IdlePolicy) => appliedIds.has(item.id) ? (
                  <Button
                    variant="link"
                    onClick={() => handleRemove(item.id, item.name)}
                    data-testid={`remove-idle-policy-${item.id}`}
                  >
                    Remove
                  </Button>
                ) : (
                  <Button
                    variant="primary"
                    onClick={() => handleApply(item.id, item.name)}
                    data-testid={`apply-idle-policy-${item.id}`}
                  >
                    Apply
                  </Button>
                )
              }
            ]}
            items={availablePolicies}
            empty={<Box textAlign="center">No idle policy templates available</Box>}
          />
        )}
      </SpaceBetween>
    </Modal>
  )
}
