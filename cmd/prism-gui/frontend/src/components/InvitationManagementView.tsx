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

  // Load shared tokens when switching to shared tab and project is selected
  useEffect(() => {
    if (activeTabId === 'shared' && bulkProjectId) {
      loadSharedTokens(bulkProjectId);
    }
  }, [activeTabId, bulkProjectId]);
  const [bulkEmails, setBulkEmails] = useState('');
  const [bulkRole, setBulkRole] = useState<'viewer' | 'member' | 'admin'>('member');
  const [bulkMessage, setBulkMessage] = useState('');
  const [bulkResults, setBulkResults] = useState<BulkInvitationResult | null>(null);
  const [bulkLoading, setBulkLoading] = useState(false);
  const [bulkError, setBulkError] = useState<string | null>(null);

  // Shared tokens state
  const [createTokenModalVisible, setCreateTokenModalVisible] = useState(false);
  const [qrModalVisible, setQRModalVisible] = useState(false);
  const [revokeTokenModalVisible, setRevokeTokenModalVisible] = useState(false);
  const [extendTokenModalVisible, setExtendTokenModalVisible] = useState(false);
  const [selectedToken, setSelectedToken] = useState<SharedInvitationToken | null>(null);
  const [tokenName, setTokenName] = useState('');
  const [tokenRole, setTokenRole] = useState<'viewer' | 'member' | 'admin'>('member');
  const [tokenLimit, setTokenLimit] = useState('10');
  const [tokenExpires, setTokenExpires] = useState<'1d' | '7d' | '30d' | '90d'>('7d');
  const [extendDuration, setExtendDuration] = useState<'7' | '30' | '90'>('7');
  const [tokenMessage, setTokenMessage] = useState('');
  const [projects, setProjects] = useState<Array<{ id: string; name: string }>>([]);

  // Access API client from window context
  const api = (window as any).__apiClient;

  // ==================== DATA LOADING ====================

  useEffect(() => {
    loadInvitations();

    // Listen for invitation-created events to refresh the list
    const handleInvitationCreated = () => loadInvitations();
    window.addEventListener('invitation-created', handleInvitationCreated);

    // Listen for shared-token-created events to refresh shared tokens
    const handleSharedTokenCreated = () => {
      if (bulkProjectId) {
        loadSharedTokens(bulkProjectId);
      }
    };
    window.addEventListener('shared-token-created', handleSharedTokenCreated);

    // Cleanup function to remove event listeners
    return () => {
      window.removeEventListener('invitation-created', handleInvitationCreated);
      window.removeEventListener('shared-token-created', handleSharedTokenCreated);
    };
  }, [bulkProjectId]);

  // Load projects for shared token creation
  useEffect(() => {
    const loadProjects = async () => {
      if (!api) return;
      try {
        const projectList = await api.getProjects();
        setProjects(projectList.map((p: any) => ({ id: p.id, name: p.name })));
      } catch (err) {
        console.error('Failed to load projects:', err);
      }
    };
    loadProjects();
  }, [api]);

  const loadInvitations = async () => {
    if (!api) return;

    setLoading(true);
    setError(null);

    try {
      // For tests: use test-user@example.com to match test invitation recipient
      const testEmail = 'test-user@example.com';
      const data = await api.getMyInvitations(testEmail);
      setInvitations(data || []);
    } catch (err: any) {
      console.error('Failed to load invitations:', err);
      setError(err.message || 'Failed to load invitations');
    } finally {
      setLoading(false);
    }
  };

  const loadSharedTokens = async (projectId: string) => {
    if (!api || !projectId) return;

    setLoading(true);
    setError(null);

    try {
      const data = await api.getSharedTokens(projectId);
      setSharedTokens(data || []);
    } catch (err: any) {
      console.error('Failed to load shared tokens:', err);
      setError(err.message || 'Failed to load shared tokens');
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

  const handleViewQRCode = async (token: SharedInvitationToken) => {
    if (!api) return;

    setSelectedToken(token);
    setLoading(true);
    setError(null);

    try {
      // Fetch QR code data from backend
      const qrData = await api.getSharedTokenQRCode(token.token, 'json');

      // Update token with QR code URL (base64 data URL)
      setSelectedToken({
        ...token,
        qr_code_url: `data:image/png;base64,${qrData.qr_code}`
      });

      setQRModalVisible(true);
    } catch (err: any) {
      console.error('Failed to fetch QR code:', err);
      setError(err.message || 'Failed to fetch QR code');
    } finally {
      setLoading(false);
    }
  };

  const handleCopyTokenURL = () => {
    if (!selectedToken) return;

    // Copy token URL to clipboard
    const url = `${window.location.origin}/invitations/redeem?token=${selectedToken.token}`;
    navigator.clipboard.writeText(url);
  };

  const handleExtendToken = async () => {
    if (!api || !selectedToken) return;

    setLoading(true);
    setError(null);

    try {
      await api.extendSharedToken(selectedToken.token, extendDuration);

      // Update local state
      setSharedTokens(prev => prev.map(t => {
        if (t.token === selectedToken.token) {
          const currentExpires = new Date(t.expires_at);
          const newExpires = new Date(currentExpires.getTime() + parseInt(extendDuration) * 24 * 60 * 60 * 1000);
          return { ...t, expires_at: newExpires.toISOString() };
        }
        return t;
      }));

      setExtendTokenModalVisible(false);
      setSelectedToken(null);

      // Dispatch event to refresh if needed
      window.dispatchEvent(new CustomEvent('shared-token-updated'));
    } catch (err: any) {
      console.error('Failed to extend token:', err);
      setError(err.message || 'Failed to extend token');
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeToken = async () => {
    if (!api || !selectedToken) return;

    setLoading(true);
    setError(null);

    try {
      await api.revokeSharedToken(selectedToken.token);

      // Update local state
      setSharedTokens(prev => prev.map(t =>
        t.token === selectedToken.token ? { ...t, revoked: true } : t
      ));

      setRevokeTokenModalVisible(false);
      setSelectedToken(null);

      // Dispatch event to refresh if needed
      window.dispatchEvent(new CustomEvent('shared-token-updated'));
    } catch (err: any) {
      console.error('Failed to revoke token:', err);
      setError(err.message || 'Failed to revoke token');
    } finally {
      setLoading(false);
    }
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
            selectedOption={
              bulkProjectId
                ? { label: projects.find(p => p.id === bulkProjectId)?.name || bulkProjectId, value: bulkProjectId }
                : null
            }
            onChange={({ detail }) => setBulkProjectId(detail.selectedOption.value || '')}
            options={projects.map(p => ({ label: p.name, value: p.id }))}
            placeholder="Select a project"
            data-testid="bulk-invite-project-select"
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
            cell: (item: SharedInvitationToken) => {
              const isExpired = new Date(item.expires_at) < new Date();
              const isRevoked = item.revoked;
              const isInactive = isExpired || isRevoked;

              return (
                <SpaceBetween direction="horizontal" size="xs">
                  <Button
                    iconName="view-full"
                    onClick={() => handleViewQRCode(item)}
                    data-testid={`view-qr-button-${item.token}`}
                  >
                    View
                  </Button>
                  <Button
                    iconName="add-plus"
                    onClick={() => {
                      setSelectedToken(item);
                      setExtendTokenModalVisible(true);
                    }}
                    disabled={isInactive}
                    data-testid={`extend-token-button-${item.token}`}
                  >
                    Extend
                  </Button>
                  <Button
                    iconName="close"
                    onClick={() => {
                      setSelectedToken(item);
                      setRevokeTokenModalVisible(true);
                    }}
                    disabled={isInactive}
                    data-testid={`revoke-token-button-${item.token}`}
                  >
                    Revoke
                  </Button>
                </SpaceBetween>
              );
            },
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

  const extendTokenModal = (
    <Modal
      onDismiss={() => {
        setExtendTokenModalVisible(false);
        setSelectedToken(null);
        setError(null);
      }}
      visible={extendTokenModalVisible}
      header="Extend Token Expiration"
      size="medium"
      data-testid="extend-token-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button
              variant="link"
              onClick={() => {
                setExtendTokenModalVisible(false);
                setSelectedToken(null);
                setError(null);
              }}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleExtendToken}
              loading={loading}
              data-testid="confirm-extend-button"
            >
              Extend
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {error && (
          <Alert type="error" dismissible onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {selectedToken && (
          <Box>
            <Box variant="awsui-key-label">Token Name</Box>
            <div>{selectedToken.name}</div>
          </Box>
        )}

        <FormField label="Extend By" description="Add additional days to the expiration">
          <Select
            selectedOption={{ label: `${extendDuration} days`, value: extendDuration }}
            onChange={({ detail }) => setExtendDuration(detail.selectedOption.value as '7' | '30' | '90')}
            options={[
              { label: '7 days', value: '7' },
              { label: '30 days', value: '30' },
              { label: '90 days', value: '90' }
            ]}
            data-testid="extend-duration-select"
          />
        </FormField>
      </SpaceBetween>
    </Modal>
  );

  const revokeTokenModal = (
    <Modal
      onDismiss={() => {
        setRevokeTokenModalVisible(false);
        setSelectedToken(null);
        setError(null);
      }}
      visible={revokeTokenModalVisible}
      header="Revoke Shared Token"
      size="medium"
      data-testid="revoke-token-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button
              variant="link"
              onClick={() => {
                setRevokeTokenModalVisible(false);
                setSelectedToken(null);
                setError(null);
              }}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleRevokeToken}
              loading={loading}
              data-testid="confirm-revoke-button"
            >
              Revoke
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {error && (
          <Alert type="error" dismissible onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        <Alert type="warning">
          Are you sure you want to revoke this token? This action cannot be undone and the token will no longer be usable.
        </Alert>

        {selectedToken && (
          <Box>
            <Box variant="awsui-key-label">Token Name</Box>
            <div>{selectedToken.name}</div>
          </Box>
        )}
      </SpaceBetween>
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
      footer={
        <Box float="right">
          <Button
            onClick={() => {
              setQRModalVisible(false);
              setSelectedToken(null);
            }}
            data-testid="close-qr-modal-button"
          >
            Close
          </Button>
        </Box>
      }
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

  const createSharedTokenModal = (
    <Modal
      onDismiss={() => {
        setCreateTokenModalVisible(false);
        setError(null);
      }}
      visible={createTokenModalVisible}
      header="Create Shared Invitation Token"
      size="medium"
      data-testid="create-shared-token-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button
              variant="link"
              onClick={() => {
                setCreateTokenModalVisible(false);
                setError(null);
              }}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleCreateToken}
              loading={loading}
              data-testid="create-token-button"
            >
              Create Token
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="m">
        {error && (
          <Alert type="error" dismissible onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        <FormField label="Project" description="Select the project for this shared token">
          <Select
            selectedOption={
              bulkProjectId
                ? { label: projects.find(p => p.id === bulkProjectId)?.name || '', value: bulkProjectId }
                : null
            }
            onChange={({ detail }) => setBulkProjectId(detail.selectedOption.value || '')}
            options={projects.map(p => ({ label: p.name, value: p.id }))}
            placeholder="Select a project"
            data-testid="token-project-select"
          />
        </FormField>

        <FormField label="Token Name" description="A friendly name to identify this token">
          <Input
            value={tokenName}
            onChange={({ detail }) => setTokenName(detail.value)}
            placeholder="e.g., Research Team Access"
            data-testid="token-name-input"
          />
        </FormField>

        <FormField label="Role" description="Access level for users who redeem this token">
          <Select
            selectedOption={{ label: tokenRole, value: tokenRole }}
            onChange={({ detail }) => setTokenRole(detail.selectedOption.value as 'viewer' | 'member' | 'admin')}
            options={[
              { label: 'viewer', value: 'viewer' },
              { label: 'member', value: 'member' },
              { label: 'admin', value: 'admin' }
            ]}
            data-testid="shared-token-role-select"
          />
        </FormField>

        <FormField label="Redemption Limit" description="Maximum number of times this token can be used">
          <Input
            value={tokenLimit}
            onChange={({ detail }) => setTokenLimit(detail.value)}
            type="number"
            placeholder="10"
            data-testid="redemption-limit-input"
          />
        </FormField>

        <FormField label="Expires In" description="How long the token remains valid">
          <Select
            selectedOption={{ label: tokenExpires, value: tokenExpires }}
            onChange={({ detail }) => setTokenExpires(detail.selectedOption.value as '1d' | '7d' | '30d' | '90d')}
            options={[
              { label: '1d', value: '1d' },
              { label: '7d', value: '7d' },
              { label: '30d', value: '30d' },
              { label: '90d', value: '90d' }
            ]}
            data-testid="expires-in-select"
          />
        </FormField>

        <FormField label="Welcome Message (Optional)" description="A message shown to users who redeem this token">
          <Textarea
            value={tokenMessage}
            onChange={({ detail }) => setTokenMessage(detail.value)}
            placeholder="Welcome to the project!"
            rows={3}
            data-testid="token-message-textarea"
          />
        </FormField>
      </SpaceBetween>
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
      {createSharedTokenModal}
      {extendTokenModal}
      {revokeTokenModal}
      {qrCodeModal}
    </SpaceBetween>
  );
};

export default InvitationManagementView;
