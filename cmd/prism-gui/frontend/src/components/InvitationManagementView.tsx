/**
 * InvitationManagementView Component
 *
 * Manages project invitations with three modes:
 * 1. Individual Invitations - View and respond to received invitations
 * 2. Bulk Invitations - Send invitations to multiple users at once
 * 3. Shared Tokens - Create reusable invitation links with QR codes
 *
 * Part of Phase 4.4: Individual Invitations System (Epic #315)
 */

import React, { useState, useEffect } from 'react';
import {
  Container,
  Header,
  Tabs,
  Table,
  Box,
  SpaceBetween,
  Button,
  Input,
  Textarea,
  Select,
  Badge,
  StatusIndicator,
  Modal,
  Alert,
  Grid,
  ColumnLayout,
  FormField,
} from '@cloudscape-design/components';

// ==================== TYPE DEFINITIONS ====================

interface CachedInvitation {
  token: string;
  invitation_id: string;
  project_id: string;
  project_name: string;
  email: string;
  role: string;
  invited_by: string;
  invited_at: string;
  expires_at: string;
  status: 'pending' | 'accepted' | 'declined' | 'expired' | 'revoked';
  message?: string;
  added_at: string;
}

interface SharedInvitationToken {
  token: string;
  project_id: string;
  project_name: string;
  name: string;
  role: 'viewer' | 'member' | 'admin';
  redemption_limit: number;
  redemptions: number;
  created_at: string;
  created_by: string;
  expires_at: string;
  revoked: boolean;
  qr_code_url?: string;
}

interface BulkInvitationResult {
  sent: number;
  failed: number;
  errors?: Array<{email: string; error: string}>;
}

// ==================== MAIN COMPONENT ====================

export const InvitationManagementView: React.FC = () => {
  // Tab state
  const [activeTabId, setActiveTabId] = useState('individual');

  // Data state
  const [invitations, setInvitations] = useState<CachedInvitation[]>([]);
  const [sharedTokens, setSharedTokens] = useState<SharedInvitationToken[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Individual invitations state
  const [tokenInput, setTokenInput] = useState('');
  const [selectedInvitation, setSelectedInvitation] = useState<CachedInvitation | null>(null);
  const [acceptModalVisible, setAcceptModalVisible] = useState(false);
  const [declineModalVisible, setDeclineModalVisible] = useState(false);
  const [declineReason, setDeclineReason] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  // Bulk invitations state
  const [bulkProjectId, setBulkProjectId] = useState('');
  const [bulkEmails, setBulkEmails] = useState('');
  const [bulkRole, setBulkRole] = useState<'viewer' | 'member' | 'admin'>('member');
  const [bulkMessage, setBulkMessage] = useState('');
  const [bulkResults, setBulkResults] = useState<BulkInvitationResult | null>(null);
  const [bulkLoading, setBulkLoading] = useState(false);
  const [bulkError, setBulkError] = useState<string | null>(null);

  // Shared tokens state
  const [createTokenModalVisible, setCreateTokenModalVisible] = useState(false);
  const [qrModalVisible, setQRModalVisible] = useState(false);
  const [selectedToken, setSelectedToken] = useState<SharedInvitationToken | null>(null);
  const [tokenName, setTokenName] = useState('');
  const [tokenRole, setTokenRole] = useState<'viewer' | 'member' | 'admin'>('member');
  const [tokenLimit, setTokenLimit] = useState('10');
  const [tokenExpires, setTokenExpires] = useState<'1d' | '7d' | '30d' | '90d'>('7d');
  const [tokenMessage, setTokenMessage] = useState('');

  // Access API client from window context
  const api = (window as any).__apiClient;

  // ==================== DATA LOADING ====================

  useEffect(() => {
    loadInvitations();
    // Shared tokens would be loaded separately per project
  }, []);

  const loadInvitations = async () => {
    if (!api) return;

    setLoading(true);
    setError(null);

    try {
      const data = await api.getMyInvitations();
      setInvitations(data || []);
    } catch (err: any) {
      console.error('Failed to load invitations:', err);
      setError(err.message || 'Failed to load invitations');
    } finally {
      setLoading(false);
    }
  };

  // ==================== INDIVIDUAL INVITATIONS HANDLERS ====================

  const handleAddToken = async () => {
    if (!api || !tokenInput.trim()) return;

    setLoading(true);
    setError(null);

    try {
      const invitation = await api.getInvitationByToken(tokenInput.trim());

      // Add to local state if not already present
      if (!invitations.find(inv => inv.token === invitation.token)) {
        setInvitations(prev => [...prev, invitation]);
      }

      setTokenInput('');
    } catch (err: any) {
      console.error('Failed to add invitation:', err);
      setError(err.message || 'Failed to add invitation');
    } finally {
      setLoading(false);
    }
  };

  const handleAcceptClick = (invitation: CachedInvitation) => {
    setSelectedInvitation(invitation);
    setAcceptModalVisible(true);
  };

  const handleAcceptConfirm = async () => {
    if (!api || !selectedInvitation) return;

    setLoading(true);
    setError(null);

    try {
      await api.acceptInvitation(selectedInvitation.token);

      // Update local state
      setInvitations(prev => prev.map(inv =>
        inv.token === selectedInvitation.token
          ? { ...inv, status: 'accepted' as const }
          : inv
      ));

      setAcceptModalVisible(false);
      setSelectedInvitation(null);
    } catch (err: any) {
      console.error('Failed to accept invitation:', err);
      setError(err.message || 'Failed to accept invitation');
    } finally {
      setLoading(false);
    }
  };

  const handleDeclineClick = (invitation: CachedInvitation) => {
    setSelectedInvitation(invitation);
    setDeclineModalVisible(true);
    setDeclineReason('');
  };

  const handleDeclineConfirm = async () => {
    if (!api || !selectedInvitation) return;

    setLoading(true);
    setError(null);

    try {
      await api.declineInvitation(
        selectedInvitation.token,
        declineReason.trim() || undefined
      );

      // Update local state
      setInvitations(prev => prev.map(inv =>
        inv.token === selectedInvitation.token
          ? { ...inv, status: 'declined' as const }
          : inv
      ));

      setDeclineModalVisible(false);
      setSelectedInvitation(null);
      setDeclineReason('');
    } catch (err: any) {
      console.error('Failed to decline invitation:', err);
      setError(err.message || 'Failed to decline invitation');
    } finally {
      setLoading(false);
    }
  };

  // ==================== BULK INVITATIONS HANDLERS ====================

  const handleBulkSend = async () => {
    if (!api || !bulkProjectId || !bulkEmails.trim()) {
      setBulkError('Please select a project and enter at least one email address');
      return;
    }

    setBulkLoading(true);
    setBulkError(null);
    setBulkResults(null);

    try {
      // Parse emails (one per line)
      const emails = bulkEmails
        .split('\n')
        .map(email => email.trim())
        .filter(email => email.length > 0);

      if (emails.length === 0) {
        setBulkError('Please enter at least one email address');
        return;
      }

      const result = await api.sendBulkInvitations(
        bulkProjectId,
        emails,
        bulkRole,
        bulkMessage.trim() || undefined
      );

      setBulkResults(result);

      // Clear form on success
      if (result.sent > 0) {
        setBulkEmails('');
        setBulkMessage('');
      }
    } catch (err: any) {
      console.error('Failed to send bulk invitations:', err);
      setBulkError(err.message || 'Failed to send invitations');
    } finally {
      setBulkLoading(false);
    }
  };

  // ==================== SHARED TOKENS HANDLERS ====================

  const handleCreateToken = async () => {
    if (!api || !tokenName.trim() || !bulkProjectId) {
      setError('Please enter a token name and select a project');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const config = {
        name: tokenName.trim(),
        role: tokenRole,
        redemption_limit: parseInt(tokenLimit) || 10,
        expires_in: tokenExpires,
        message: tokenMessage.trim() || undefined
      };

      const token = await api.createSharedToken(bulkProjectId, config);

      // Add to local state
      setSharedTokens(prev => [...prev, token]);

      // Reset form
      setTokenName('');
      setTokenMessage('');
      setCreateTokenModalVisible(false);
    } catch (err: any) {
      console.error('Failed to create shared token:', err);
      setError(err.message || 'Failed to create token');
    } finally {
      setLoading(false);
    }
  };

  const handleViewQRCode = (token: SharedInvitationToken) => {
    setSelectedToken(token);
    setQRModalVisible(true);
  };

  const handleCopyTokenURL = () => {
    if (!selectedToken) return;

    // Copy token URL to clipboard
    const url = `${window.location.origin}/invitations/redeem?token=${selectedToken.token}`;
    navigator.clipboard.writeText(url);
  };

  // ==================== HELPER FUNCTIONS ====================

  const getStatusBadge = (status: string) => {
    const statusMap = {
      pending: { color: 'blue' as const, label: 'Pending' },
      accepted: { color: 'green' as const, label: 'Accepted' },
      declined: { color: 'red' as const, label: 'Declined' },
      expired: { color: 'grey' as const, label: 'Expired' },
      revoked: { color: 'grey' as const, label: 'Revoked' },
    };

    const config = statusMap[status as keyof typeof statusMap] || statusMap.pending;
    return <Badge color={config.color}>{config.label}</Badge>;
  };

  const getRoleBadge = (role: string) => {
    const roleColors = {
      admin: 'red' as const,
      member: 'blue' as const,
      viewer: 'grey' as const,
    };

    return <Badge color={roleColors[role as keyof typeof roleColors] || 'grey'}>{role}</Badge>;
  };

  const getRelativeTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = date.getTime() - now.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays < 0) return 'Expired';
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Tomorrow';
    return `${diffDays} days`;
  };

  // Filter invitations based on status
  const filteredInvitations = statusFilter === 'all'
    ? invitations
    : invitations.filter(inv => inv.status === statusFilter);

  // ==================== STATISTICS CALCULATION ====================

  const stats = {
    total: invitations.length,
    pending: invitations.filter(inv => inv.status === 'pending').length,
    accepted: invitations.filter(inv => inv.status === 'accepted').length,
    declined: invitations.filter(inv => inv.status === 'declined').length,
    expired: invitations.filter(inv => inv.status === 'expired').length,
  };

  // ==================== INDIVIDUAL INVITATIONS TAB ====================

  const individualInvitationsTab = (
    <Container
      header={
        <Header
          variant="h2"
          description="View and respond to project invitations you've received"
        >
          My Invitations
        </Header>
      }
    >
      <SpaceBetween size="l">
        {/* Statistics */}
        <ColumnLayout columns={5} variant="text-grid">
          <div>
            <Box variant="awsui-key-label">Total</Box>
            <div>{stats.total}</div>
          </div>
          <div>
            <Box variant="awsui-key-label">Pending</Box>
            <div>{stats.pending}</div>
          </div>
          <div>
            <Box variant="awsui-key-label">Accepted</Box>
            <div>{stats.accepted}</div>
          </div>
          <div>
            <Box variant="awsui-key-label">Declined</Box>
            <div>{stats.declined}</div>
          </div>
          <div>
            <Box variant="awsui-key-label">Expired</Box>
            <div>{stats.expired}</div>
          </div>
        </ColumnLayout>

        {/* Add Invitation by Token */}
        <SpaceBetween size="s">
          <FormField
            label="Add Invitation by Token"
            description="Enter an invitation token you received"
          >
            <Grid gridDefinition={[{ colspan: 10 }, { colspan: 2 }]}>
              <Input
                value={tokenInput}
                onChange={({ detail }) => setTokenInput(detail.value)}
                placeholder="Enter invitation token"
                data-testid="invitation-token-input"
                disabled={loading}
              />
              <Button
                onClick={handleAddToken}
                loading={loading}
                disabled={!tokenInput.trim()}
                data-testid="add-invitation-button"
              >
                Add Invitation
              </Button>
            </Grid>
          </FormField>
        </SpaceBetween>

        {/* Status Filter */}
        <FormField label="Filter by Status">
          <Select
            selectedOption={{ label: statusFilter === 'all' ? 'All' : statusFilter, value: statusFilter }}
            onChange={({ detail }) => setStatusFilter(detail.selectedOption.value || 'all')}
            options={[
              { label: 'All', value: 'all' },
              { label: 'Pending', value: 'pending' },
              { label: 'Accepted', value: 'accepted' },
              { label: 'Declined', value: 'declined' },
              { label: 'Expired', value: 'expired' },
            ]}
            data-testid="invitation-status-filter"
          />
        </FormField>

        {/* Error Display */}
        {error && (
          <Alert type="error" dismissible onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Invitations Table */}
        <Table
          columnDefinitions={[
            {
              id: 'project',
              header: 'Project',
              cell: (item: CachedInvitation) => item.project_name,
              sortingField: 'project_name',
            },
            {
              id: 'role',
              header: 'Role',
              cell: (item: CachedInvitation) => getRoleBadge(item.role),
              sortingField: 'role',
            },
            {
              id: 'invited_by',
              header: 'Invited By',
              cell: (item: CachedInvitation) => item.invited_by,
              sortingField: 'invited_by',
            },
            {
              id: 'expires',
              header: 'Expires',
              cell: (item: CachedInvitation) => getRelativeTime(item.expires_at),
              sortingField: 'expires_at',
            },
            {
              id: 'status',
              header: 'Status',
              cell: (item: CachedInvitation) => getStatusBadge(item.status),
              sortingField: 'status',
            },
            {
              id: 'actions',
              header: 'Actions',
              cell: (item: CachedInvitation) => (
                <SpaceBetween direction="horizontal" size="xs">
                  <Button
                    variant="primary"
                    iconName="check"
                    onClick={() => handleAcceptClick(item)}
                    disabled={item.status !== 'pending' || loading}
                    data-testid={`accept-invitation-button-${item.invitation_id}`}
                  >
                    Accept
                  </Button>
                  <Button
                    variant="normal"
                    iconName="close"
                    onClick={() => handleDeclineClick(item)}
                    disabled={item.status !== 'pending' || loading}
                    data-testid={`decline-invitation-button-${item.invitation_id}`}
                  >
                    Decline
                  </Button>
                </SpaceBetween>
              ),
            },
          ]}
          items={filteredInvitations}
          loading={loading}
          loadingText="Loading invitations..."
          empty={
            <Box textAlign="center" color="inherit">
              <b>No invitations</b>
              <Box padding={{ bottom: 's' }} variant="p" color="inherit">
                {statusFilter === 'all'
                  ? "You don't have any project invitations yet."
                  : `No ${statusFilter} invitations found.`}
              </Box>
            </Box>
          }
          data-testid="invitations-table"
        />
      </SpaceBetween>
    </Container>
  );

  // ==================== BULK INVITATIONS TAB ====================

  const bulkInvitationsTab = (
    <Container
      header={
        <Header
          variant="h2"
          description="Send invitations to multiple users at once"
        >
          Bulk Invitations
        </Header>
      }
    >
      <SpaceBetween size="l">
        <Alert type="info">
          Enter one email address per line. All users will receive an invitation to join the selected project with the specified role.
        </Alert>

        <FormField label="Project" constraintText="Required">
          <Select
            selectedOption={bulkProjectId ? { label: bulkProjectId, value: bulkProjectId } : null}
            onChange={({ detail }) => setBulkProjectId(detail.selectedOption.value || '')}
            options={[
              { label: 'Select a project', value: '' },
              // TODO: Load from actual projects list
            ]}
            placeholder="Select a project"
            data-testid="bulk-project-select"
          />
        </FormField>

        <FormField
          label="Email Addresses"
          description="One email per line"
          constraintText="Required"
        >
          <Textarea
            value={bulkEmails}
            onChange={({ detail }) => setBulkEmails(detail.value)}
            placeholder={'researcher1@example.com\nresearcher2@example.com\nresearcher3@example.com'}
            rows={6}
            data-testid="bulk-emails-textarea"
            disabled={bulkLoading}
          />
        </FormField>

        <FormField label="Role">
          <Select
            selectedOption={{ label: bulkRole, value: bulkRole }}
            onChange={({ detail }) => setBulkRole(detail.selectedOption.value as any)}
            options={[
              { label: 'viewer', value: 'viewer' },
              { label: 'member', value: 'member' },
              { label: 'admin', value: 'admin' },
            ]}
            data-testid="bulk-role-select"
          />
        </FormField>

        <FormField
          label="Welcome Message"
          description="Optional message included with invitations"
        >
          <Textarea
            value={bulkMessage}
            onChange={({ detail }) => setBulkMessage(detail.value)}
            placeholder="Enter an optional welcome message..."
            rows={3}
            data-testid="bulk-message-textarea"
            disabled={bulkLoading}
          />
        </FormField>

        {bulkError && (
          <Alert type="error" dismissible onDismiss={() => setBulkError(null)}>
            {bulkError}
          </Alert>
        )}

        {bulkResults && (
          <Alert
            type={bulkResults.failed === 0 ? 'success' : 'warning'}
            dismissible
            onDismiss={() => setBulkResults(null)}
            data-testid="bulk-results-container"
          >
            <SpaceBetween size="xs">
              <div>
                <strong>Bulk Invitation Results:</strong>
              </div>
              <div>✓ Sent: {bulkResults.sent}</div>
              {bulkResults.failed > 0 && (
                <>
                  <div>✗ Failed: {bulkResults.failed}</div>
                  {bulkResults.errors && bulkResults.errors.length > 0 && (
                    <Box variant="p">
                      <strong>Errors:</strong>
                      <ul>
                        {bulkResults.errors.map((err, idx) => (
                          <li key={idx}>{err.email}: {err.error}</li>
                        ))}
                      </ul>
                    </Box>
                  )}
                </>
              )}
            </SpaceBetween>
          </Alert>
        )}

        <Button
          variant="primary"
          onClick={handleBulkSend}
          loading={bulkLoading}
          disabled={!bulkProjectId || !bulkEmails.trim()}
          data-testid="send-bulk-invitations-button"
        >
          Send Bulk Invitations
        </Button>
      </SpaceBetween>
    </Container>
  );

  // ==================== SHARED TOKENS TAB ====================

  const sharedTokensTab = (
    <Container
      header={
        <Header
          variant="h2"
          description="Create reusable invitation links with QR codes"
          actions={
            <Button
              variant="primary"
              iconName="add-plus"
              onClick={() => setCreateTokenModalVisible(true)}
              data-testid="create-shared-token-button"
            >
              Create Shared Token
            </Button>
          }
        >
          Shared Invitation Tokens
        </Header>
      }
    >
      <Table
        columnDefinitions={[
          {
            id: 'name',
            header: 'Name',
            cell: (item: SharedInvitationToken) => item.name,
            sortingField: 'name',
          },
          {
            id: 'project',
            header: 'Project',
            cell: (item: SharedInvitationToken) => item.project_name,
            sortingField: 'project_name',
          },
          {
            id: 'role',
            header: 'Role',
            cell: (item: SharedInvitationToken) => getRoleBadge(item.role),
            sortingField: 'role',
          },
          {
            id: 'redemptions',
            header: 'Redemptions',
            cell: (item: SharedInvitationToken) => `${item.redemptions} / ${item.redemption_limit}`,
          },
          {
            id: 'expires',
            header: 'Expires',
            cell: (item: SharedInvitationToken) => getRelativeTime(item.expires_at),
            sortingField: 'expires_at',
          },
          {
            id: 'status',
            header: 'Status',
            cell: (item: SharedInvitationToken) =>
              item.revoked ? getStatusBadge('revoked') :
              new Date(item.expires_at) < new Date() ? getStatusBadge('expired') :
              getStatusBadge('pending'),
          },
          {
            id: 'actions',
            header: 'Actions',
            cell: (item: SharedInvitationToken) => (
              <SpaceBetween direction="horizontal" size="xs">
                <Button
                  iconName="view-full"
                  onClick={() => handleViewQRCode(item)}
                  data-testid={`view-qr-button-${item.token}`}
                >
                  View QR
                </Button>
              </SpaceBetween>
            ),
          },
        ]}
        items={sharedTokens}
        loading={loading}
        loadingText="Loading shared tokens..."
        empty={
          <Box textAlign="center" color="inherit">
            <b>No shared tokens</b>
            <Box padding={{ bottom: 's' }} variant="p" color="inherit">
              Create a shared invitation token to generate a reusable link with QR code.
            </Box>
          </Box>
        }
        data-testid="shared-tokens-table"
      />
    </Container>
  );

  // ==================== MODALS ====================

  const acceptInvitationModal = (
    <Modal
      onDismiss={() => {
        setAcceptModalVisible(false);
        setSelectedInvitation(null);
      }}
      visible={acceptModalVisible}
      header="Accept Invitation"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button
              variant="link"
              onClick={() => {
                setAcceptModalVisible(false);
                setSelectedInvitation(null);
              }}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleAcceptConfirm}
              loading={loading}
            >
              Accept
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      {selectedInvitation && (
        <SpaceBetween size="m">
          <Box>
            Are you sure you want to accept this invitation?
          </Box>
          <ColumnLayout columns={2} variant="text-grid">
            <div>
              <Box variant="awsui-key-label">Project</Box>
              <div>{selectedInvitation.project_name}</div>
            </div>
            <div>
              <Box variant="awsui-key-label">Role</Box>
              <div>{selectedInvitation.role}</div>
            </div>
          </ColumnLayout>
          {selectedInvitation.message && (
            <Box>
              <Box variant="awsui-key-label">Message</Box>
              <Box variant="p">{selectedInvitation.message}</Box>
            </Box>
          )}
        </SpaceBetween>
      )}
    </Modal>
  );

  const declineInvitationModal = (
    <Modal
      onDismiss={() => {
        setDeclineModalVisible(false);
        setSelectedInvitation(null);
        setDeclineReason('');
      }}
      visible={declineModalVisible}
      header="Decline Invitation"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button
              variant="link"
              onClick={() => {
                setDeclineModalVisible(false);
                setSelectedInvitation(null);
                setDeclineReason('');
              }}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleDeclineConfirm}
              loading={loading}
            >
              Decline
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      {selectedInvitation && (
        <SpaceBetween size="m">
          <Box>
            Are you sure you want to decline this invitation to <strong>{selectedInvitation.project_name}</strong>?
          </Box>
          <FormField
            label="Reason (Optional)"
            description="Provide an optional reason for declining"
          >
            <Textarea
              value={declineReason}
              onChange={({ detail }) => setDeclineReason(detail.value)}
              placeholder="Enter reason for declining..."
              rows={3}
              data-testid="decline-reason-textarea"
            />
          </FormField>
        </SpaceBetween>
      )}
    </Modal>
  );

  const qrCodeModal = (
    <Modal
      onDismiss={() => {
        setQRModalVisible(false);
        setSelectedToken(null);
      }}
      visible={qrModalVisible}
      header="Shared Invitation Token"
      size="medium"
      data-testid="qr-code-modal"
    >
      {selectedToken && (
        <SpaceBetween size="l">
          <ColumnLayout columns={2} variant="text-grid">
            <div>
              <Box variant="awsui-key-label">Name</Box>
              <div>{selectedToken.name}</div>
            </div>
            <div>
              <Box variant="awsui-key-label">Project</Box>
              <div>{selectedToken.project_name}</div>
            </div>
            <div>
              <Box variant="awsui-key-label">Role</Box>
              <div>{getRoleBadge(selectedToken.role)}</div>
            </div>
            <div>
              <Box variant="awsui-key-label">Redemptions</Box>
              <div>{selectedToken.redemptions} / {selectedToken.redemption_limit}</div>
            </div>
          </ColumnLayout>

          {selectedToken.qr_code_url && (
            <Box textAlign="center">
              <img
                src={selectedToken.qr_code_url}
                alt="QR Code"
                style={{ maxWidth: '300px', height: 'auto' }}
              />
            </Box>
          )}

          <FormField label="Invitation URL">
            <Input
              value={`${window.location.origin}/invitations/redeem?token=${selectedToken.token}`}
              readOnly
            />
          </FormField>

          <Button
            variant="primary"
            iconName="copy"
            onClick={handleCopyTokenURL}
            data-testid="copy-token-url-button"
          >
            Copy URL
          </Button>
        </SpaceBetween>
      )}
    </Modal>
  );

  // ==================== MAIN RENDER ====================

  return (
    <SpaceBetween size="l">
      <Tabs
        activeTabId={activeTabId}
        onChange={({ detail }) => setActiveTabId(detail.activeTabId)}
        tabs={[
          {
            label: 'Individual',
            id: 'individual',
            content: individualInvitationsTab,
          },
          {
            label: 'Bulk',
            id: 'bulk',
            content: bulkInvitationsTab,
          },
          {
            label: 'Shared Tokens',
            id: 'shared',
            content: sharedTokensTab,
          },
        ]}
      />

      {/* Modals */}
      {acceptInvitationModal}
      {declineInvitationModal}
      {qrCodeModal}
    </SpaceBetween>
  );
};

export default InvitationManagementView;
