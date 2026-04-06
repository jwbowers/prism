import React from 'react';
import { useApi } from '../hooks/use-api';
import {
  Container,
  Header,
  SpaceBetween,
  ColumnLayout,
  Box,
  Button,
  ProgressBar,
  Table,
  Badge,
  Spinner,
  Select,
  Modal,
  Form,
  FormField,
  Tabs
} from '../lib/cloudscape-shim';
import { GovernancePanel } from './GovernancePanel';

interface ProjectMember {
  user_id: string;
  username: string;
  role: string;
  joined_at: string;
}

interface CostBreakdown {
  instances: number;
  storage: number;
  data_transfer: number;
  total: number;
}

interface ProjectDetails {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  budget_limit?: number;
  current_spend?: number;
  owner_id?: string;
  members: ProjectMember[];
  cost_breakdown: CostBreakdown;
}

interface ProjectDetailViewProps {
  projectId: string;
  onBack: () => void;
  onEditMember?: (member: ProjectMember) => void;
  onRemoveMember?: (member: ProjectMember) => void;
}

export const ProjectDetailView: React.FC<ProjectDetailViewProps> = ({ projectId, onBack, onEditMember, onRemoveMember }) => {
  const apiClient = useApi();
  const [project, setProject] = React.useState<ProjectDetails | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [editMemberModalVisible, setEditMemberModalVisible] = React.useState(false);
  const [memberToEdit, setMemberToEdit] = React.useState<ProjectMember | null>(null);
  const [editMemberRole, setEditMemberRole] = React.useState('');
  const [activeDetailTab, setActiveDetailTab] = React.useState('info');

  React.useEffect(() => {
    loadProjectDetails();
  }, [projectId]);

  const loadProjectDetails = async () => {
    setLoading(true);
    setError(null);
    try {
      const details = await apiClient.getProjectDetails(projectId);
      setProject(details as unknown as ProjectDetails);
    } catch (err: any) {
      setError(err.message || 'Failed to load project details');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Container>
        <Box textAlign="center" padding={{ vertical: 'xxl' }}>
          <Spinner size="large" />
          <Box variant="p" padding={{ top: 'm' }}>
            Loading project details...
          </Box>
        </Box>
      </Container>
    );
  }

  if (error || !project) {
    return (
      <Container>
        <Box textAlign="center" padding={{ vertical: 'xxl' }}>
          <Box variant="h3" color="text-status-error">
            Failed to load project details
          </Box>
          <Box variant="p" padding={{ top: 's' }}>
            {error || 'Project not found'}
          </Box>
          <Button onClick={onBack} variant="primary">
            Back to Projects
          </Button>
        </Box>
      </Container>
    );
  }

  const budgetUtilization = project.budget_limit
    ? (project.current_spend || 0) / project.budget_limit * 100
    : 0;

  const budgetStatus: 'error' | 'success' | 'in-progress' = budgetUtilization >= 90
    ? 'error'
    : budgetUtilization >= 75
    ? 'in-progress'
    : 'success';

  return (
    <SpaceBetween size="l" data-testid="project-detail-view">
      <Header
        variant="h1"
        actions={
          <Button onClick={onBack} data-testid="back-to-projects-button">
            Back to Projects
          </Button>
        }
      >
        {project.name}
      </Header>

      <Tabs
        data-testid="project-detail-tabs"
        activeTabId={activeDetailTab}
        onChange={({ detail }) => setActiveDetailTab(detail.activeTabId)}
        tabs={[
          {
            id: 'info',
            label: 'Overview',
            content: (
              <SpaceBetween size="l">

      {/* Project Information */}
      <Container header={<Header variant="h2">Project Information</Header>}>
        <ColumnLayout columns={3} variant="text-grid">
          <div>
            <Box variant="awsui-key-label">Description</Box>
            <div data-testid="project-description">
              {project.description || 'No description provided'}
            </div>
          </div>
          <div>
            <Box variant="awsui-key-label">Created</Box>
            <div data-testid="project-created-date">
              {new Date(project.created_at).toLocaleDateString()}
            </div>
          </div>
          <div>
            <Box variant="awsui-key-label">Last Updated</Box>
            <div data-testid="project-updated-date">
              {new Date(project.updated_at).toLocaleDateString()}
            </div>
          </div>
        </ColumnLayout>
      </Container>

      {/* Budget Utilization */}
      {project.budget_limit && (
        <Container
          header={<Header variant="h2">Budget Utilization</Header>}
          data-testid="budget-utilization-container"
        >
          <SpaceBetween size="l">
            <ColumnLayout columns={3} variant="text-grid">
              <div>
                <Box variant="awsui-key-label">Budget Limit</Box>
                <div data-testid="budget-limit">
                  ${project.budget_limit.toFixed(2)}
                </div>
              </div>
              <div>
                <Box variant="awsui-key-label">Current Spend</Box>
                <div data-testid="current-spend">
                  ${(project.current_spend || 0).toFixed(2)}
                </div>
              </div>
              <div>
                <Box variant="awsui-key-label">Remaining</Box>
                <div data-testid="budget-remaining">
                  ${(project.budget_limit - (project.current_spend || 0)).toFixed(2)}
                </div>
              </div>
            </ColumnLayout>

            <ProgressBar
              value={budgetUtilization}
              status={budgetStatus}
              label="Budget utilization"
              description={`${budgetUtilization.toFixed(1)}% of budget used`}
              data-testid="budget-progress-bar"
            />

            {budgetUtilization >= 75 && (
              <Box variant="p" color={budgetUtilization >= 90 ? 'text-status-error' : 'text-status-warning'}>
                {budgetUtilization >= 90
                  ? '⚠️ Budget nearly exhausted! Consider increasing limit or optimizing resources.'
                  : '⚠️ Approaching budget limit. Monitor spending closely.'}
              </Box>
            )}
          </SpaceBetween>
        </Container>
      )}

      {/* Cost Breakdown */}
      {project.cost_breakdown && (
        <Container
          header={<Header variant="h2">Cost Breakdown</Header>}
          data-testid="cost-breakdown-container"
        >
          <ColumnLayout columns={4} variant="text-grid">
            <div>
              <Box variant="awsui-key-label">Instances</Box>
              <div data-testid="cost-instances">
                ${project.cost_breakdown.instances.toFixed(2)}
              </div>
            </div>
            <div>
              <Box variant="awsui-key-label">Storage</Box>
              <div data-testid="cost-storage">
                ${project.cost_breakdown.storage.toFixed(2)}
              </div>
            </div>
            <div>
              <Box variant="awsui-key-label">Data Transfer</Box>
              <div data-testid="cost-data-transfer">
                ${project.cost_breakdown.data_transfer.toFixed(2)}
              </div>
            </div>
            <div>
              <Box variant="awsui-key-label">Total</Box>
              <div data-testid="cost-total">
                ${project.cost_breakdown.total.toFixed(2)}
              </div>
            </div>
          </ColumnLayout>
        </Container>
      )}

      {/* Project Members */}
      <Container
        header={
          <Header
            variant="h2"
            description="Users with access to this project"
          >
            Project Members ({project.members?.length || 0})
          </Header>
        }
        data-testid="project-members-container"
      >
        <Table
          columnDefinitions={[
            {
              id: 'username',
              header: 'Username',
              cell: (item: ProjectMember) => item.username,
              sortingField: 'username'
            },
            {
              id: 'role',
              header: 'Role',
              cell: (item: ProjectMember) => (
                <Badge
                  color={
                    item.role === 'admin' ? 'red' :
                    item.role === 'member' ? 'blue' :
                    'grey'
                  }
                >
                  {item.role}
                </Badge>
              ),
              sortingField: 'role'
            },
            {
              id: 'joined_at',
              header: 'Joined',
              cell: (item: ProjectMember) =>
                new Date(item.joined_at).toLocaleDateString(),
              sortingField: 'joined_at'
            },
            {
              id: 'actions',
              header: 'Actions',
              cell: (item: ProjectMember) => (
                <SpaceBetween direction="horizontal" size="xs">
                  <Button
                    onClick={() => {
                      if (onEditMember) {
                        onEditMember(item);
                      } else {
                        setMemberToEdit(item);
                        setEditMemberRole(item.role);
                        setEditMemberModalVisible(true);
                      }
                    }}
                  >
                    Change Role
                  </Button>
                  <Button
                    variant="link"
                    onClick={() => onRemoveMember && onRemoveMember(item)}
                  >
                    Remove
                  </Button>
                </SpaceBetween>
              )
            }
          ]}
          items={project.members || []}
          empty={
            <Box textAlign="center" color="inherit">
              <b>No members</b>
              <Box variant="p" color="inherit">
                No members have been added to this project yet.
              </Box>
            </Box>
          }
          data-testid="project-members-table"
        />

        {/* Inline edit role modal for when parent doesn't provide onEditMember */}
        <Modal
          visible={editMemberModalVisible}
          onDismiss={() => setEditMemberModalVisible(false)}
          header={`Change Role — ${memberToEdit?.username || ''}`}
          footer={
            <Box float="right">
              <SpaceBetween direction="horizontal" size="xs">
                <Button variant="link" onClick={() => setEditMemberModalVisible(false)}>Cancel</Button>
                <Button
                  variant="primary"
                  onClick={async () => {
                    if (!memberToEdit) return;
                    try {
                      await apiClient.updateProjectMember(projectId, memberToEdit.user_id, { role: editMemberRole });
                      await loadProjectDetails();
                    } catch (err: any) {
                      setError(err.message || 'Failed to update member role');
                    } finally {
                      setEditMemberModalVisible(false);
                    }
                  }}
                >
                  Save
                </Button>
              </SpaceBetween>
            </Box>
          }
        >
          <Form>
            <FormField label="Role">
              <Select
                selectedOption={{ value: editMemberRole, label: editMemberRole }}
                onChange={({ detail }) => setEditMemberRole(detail.selectedOption.value || 'member')}
                options={[
                  { value: 'viewer', label: 'Viewer' },
                  { value: 'member', label: 'Member' },
                  { value: 'admin', label: 'Admin' }
                ]}
              />
            </FormField>
          </Form>
        </Modal>
      </Container>

              </SpaceBetween>
            )
          },
          {
            id: 'governance',
            label: 'Governance',
            content: <GovernancePanel projectId={projectId} />
          }
        ]}
      />
    </SpaceBetween>
  );
};
