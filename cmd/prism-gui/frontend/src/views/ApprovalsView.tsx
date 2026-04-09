import { useState, useEffect } from 'react'
import {
  SpaceBetween,
  Header,
  Alert,
  Container,
  Select,
  Button,
  Spinner,
  Table,
  Badge,
  Box,
  Modal,
  Form,
  FormField,
  Textarea,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import { useApi } from '../hooks/use-api'
import type { ApprovalRequest } from '../lib/types'

export function ApprovalsView() {
  const api = useApi()
  const [approvals, setApprovals] = useState<ApprovalRequest[]>([])
  const [approvalsLoading, setApprovalsLoading] = useState(true)
  const [approvalsError, setApprovalsError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState('pending')
  const [reviewModalVisible, setReviewModalVisible] = useState(false)
  const [reviewingApproval, setReviewingApproval] = useState<ApprovalRequest | null>(null)
  const [reviewAction, setReviewAction] = useState<'approve' | 'deny'>('approve')
  const [reviewNote, setReviewNote] = useState('')

  const loadApprovals = async () => {
    setApprovalsLoading(true)
    setApprovalsError(null)
    try {
      const result = await api.listAllApprovals(statusFilter || undefined)
      setApprovals(result)
    } catch (e) {
      setApprovalsError((e as Error).message || 'Failed to load approvals')
    } finally {
      setApprovalsLoading(false)
    }
  }

  useEffect(() => { loadApprovals() }, [statusFilter]) // eslint-disable-line react-hooks/exhaustive-deps

  const openReview = (item: ApprovalRequest, action: 'approve' | 'deny') => {
    setReviewingApproval(item)
    setReviewAction(action)
    setReviewNote('')
    setReviewModalVisible(true)
  }

  const submitReview = async () => {
    if (!reviewingApproval) return
    setReviewModalVisible(false)
    try {
      if (reviewAction === 'approve') {
        await api.approveRequest(reviewingApproval.project_id, reviewingApproval.id, reviewNote)
      } else {
        await api.denyRequest(reviewingApproval.project_id, reviewingApproval.id, reviewNote)
      }
      loadApprovals()
    } catch (e) {
      setApprovalsError((e as Error).message || 'Failed to submit review')
    }
  }

  const statusColor = (s: string) => {
    switch (s) {
      case 'pending': return 'in-progress' as const
      case 'approved': return 'success' as const
      case 'denied': return 'error' as const
      default: return 'stopped' as const
    }
  }

  const typeColor = (t: string): 'blue' | 'red' | 'grey' | 'green' => {
    if (t.includes('gpu') || t.includes('expensive')) return 'red'
    if (t.includes('emergency')) return 'red'
    if (t.includes('budget')) return 'blue'
    return 'grey'
  }

  return (
    <SpaceBetween size="l" data-testid="approvals-view">
      <Header
        variant="h1"
        description="Review and manage approval requests across all projects"
        actions={
          <Button onClick={loadApprovals} disabled={approvalsLoading}>
            {approvalsLoading ? <Spinner /> : 'Refresh'}
          </Button>
        }
      >
        Approval Requests
      </Header>
      {!approvalsLoading && approvalsError && (
        <Alert type="error" dismissible onDismiss={() => setApprovalsError(null)}>{approvalsError}</Alert>
      )}
      <Container
        header={
          <Header
            variant="h2"
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Select
                  data-testid="approvals-status-filter"
                  selectedOption={{ value: statusFilter, label: statusFilter || 'All' }}
                  onChange={({ detail }) => setStatusFilter(detail.selectedOption.value || '')}
                  options={[
                    { value: '', label: 'All' },
                    { value: 'pending', label: 'Pending' },
                    { value: 'approved', label: 'Approved' },
                    { value: 'denied', label: 'Denied' },
                    { value: 'expired', label: 'Expired' }
                  ]}
                />
                <Button onClick={loadApprovals} iconName="refresh" data-testid="approvals-refresh-button">Refresh</Button>
              </SpaceBetween>
            }
          >
            {statusFilter ? `${statusFilter.charAt(0).toUpperCase() + statusFilter.slice(1)} Requests` : 'All Requests'}
          </Header>
        }
      >
        <Table
          data-testid="approvals-table"
          loading={approvalsLoading}
          loadingText="Loading approvals..."
          columnDefinitions={[
            {
              id: 'type',
              header: 'Type',
              cell: (item: ApprovalRequest) => (
                <Badge color={typeColor(item.type)}>{item.type}</Badge>
              )
            },
            {
              id: 'requested_by',
              header: 'Requester',
              cell: (item: ApprovalRequest) => item.requested_by
            },
            {
              id: 'project_id',
              header: 'Project',
              cell: (item: ApprovalRequest) => item.project_id
            },
            {
              id: 'reason',
              header: 'Reason',
              cell: (item: ApprovalRequest) => item.reason.length > 60 ? item.reason.slice(0, 57) + '...' : item.reason
            },
            {
              id: 'created_at',
              header: 'Requested',
              cell: (item: ApprovalRequest) => new Date(item.created_at).toLocaleDateString()
            },
            {
              id: 'expires_at',
              header: 'Expires',
              cell: (item: ApprovalRequest) => new Date(item.expires_at).toLocaleDateString()
            },
            {
              id: 'status',
              header: 'Status',
              cell: (item: ApprovalRequest) => (
                <StatusIndicator type={statusColor(item.status)}>{item.status}</StatusIndicator>
              )
            },
            {
              id: 'actions',
              header: 'Actions',
              cell: (item: ApprovalRequest) => item.status === 'pending' ? (
                <SpaceBetween direction="horizontal" size="xs">
                  <Button
                    variant="primary"
                    data-testid={`approve-btn-${item.id}`}
                    onClick={() => openReview(item, 'approve')}
                  >
                    Approve
                  </Button>
                  <Button
                    variant="link"
                    data-testid={`deny-btn-${item.id}`}
                    onClick={() => openReview(item, 'deny')}
                  >
                    Deny
                  </Button>
                </SpaceBetween>
              ) : null
            }
          ]}
          items={approvals}
          empty={
            <Box textAlign="center" padding={{ vertical: 'l' }}>
              <Box variant="strong">No approval requests</Box>
              <Box variant="p" color="text-body-secondary">
                {statusFilter
                  ? `No ${statusFilter} requests found. Try changing the filter.`
                  : 'Approval requests will appear here when users request resources that require review.'}
              </Box>
            </Box>
          }
        />
      </Container>

      <Modal
        visible={reviewModalVisible}
        onDismiss={() => setReviewModalVisible(false)}
        header={reviewAction === 'approve' ? 'Approve Request' : 'Deny Request'}
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setReviewModalVisible(false)}>Cancel</Button>
              <Button
                variant="primary"
                data-testid="submit-review-button"
                onClick={submitReview}
              >
                {reviewAction === 'approve' ? 'Approve' : 'Deny'}
              </Button>
            </SpaceBetween>
          </Box>
        }
      >
        <Form>
          <FormField label="Review note (optional)">
            <Textarea
              value={reviewNote}
              onChange={({ detail }) => setReviewNote(detail.value)}
              placeholder="Add a note explaining your decision..."
              data-testid="review-note-input"
            />
          </FormField>
          {reviewingApproval && (
            <SpaceBetween size="s">
              <div><strong>Type:</strong> {reviewingApproval.type}</div>
              <div><strong>Requester:</strong> {reviewingApproval.requested_by}</div>
              <div><strong>Reason:</strong> {reviewingApproval.reason}</div>
            </SpaceBetween>
          )}
        </Form>
      </Modal>
    </SpaceBetween>
  )
}
