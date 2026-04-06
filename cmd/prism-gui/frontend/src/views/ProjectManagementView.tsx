import { useState, useMemo } from 'react'
import {
  SpaceBetween,
  Header,
  Container,
  Button,
  Table,
  Spinner,
  Box,
  ColumnLayout,
  Link,
  ButtonDropdown,
  Select,
  Badge,
  Pagination,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import { useApi } from '../hooks/use-api'
import type { Project } from '../lib/types'

interface ProjectManagementViewProps {
  projects: Project[]
  loading: boolean
  onRefresh: () => void
  onCreateProject: () => void
  onSelectProject: (id: string) => void
  onEditProject: (project: Project) => void
  onManageBudget: (project: Project) => void
  onViewCost: (project: Project) => void
  onViewUsage: (project: Project) => void
  onManageMembers: (project: Project) => void
  onDeleteProject: (project: Project) => void
}

export function ProjectManagementView({
  projects,
  loading,
  onRefresh,
  onCreateProject,
  onSelectProject,
  onEditProject,
  onManageBudget,
  onViewCost,
  onViewUsage,
  onManageMembers,
  onDeleteProject,
}: ProjectManagementViewProps) {
  const api = useApi()

  const [projectFilter, setProjectFilter] = useState<string>('all');

  const [projectsCurrentPage, setProjectsCurrentPage] = useState(1);
  const projectsPageSize = 20;

  const [projectsSortingColumn, setProjectsSortingColumn] = useState({});

  const filteredProjects = useMemo(() => {
    let list = projects;
    if (projectFilter === 'active') list = list.filter(p => p.status === 'active');
    if (projectFilter === 'suspended') list = list.filter(p => p.status === 'suspended');

    return [...list].sort((a, b) => {
      const dateA = a.created_at ? new Date(a.created_at).getTime() : 0;
      const dateB = b.created_at ? new Date(b.created_at).getTime() : 0;
      const timeA = isNaN(dateA) ? 0 : dateA;
      const timeB = isNaN(dateB) ? 0 : dateB;

      if (timeB !== timeA) {
        return timeB - timeA;
      }

      return (a.id || '').localeCompare(b.id || '');
    });
  }, [projects, projectFilter]);

  const projectsTotalPages = Math.max(1, Math.ceil(filteredProjects.length / projectsPageSize));
  // Clamp current page to valid range (handles filter reducing total pages)
  const clampedPage = Math.min(projectsCurrentPage, projectsTotalPages);

  const paginatedProjects = useMemo(() => {
    const startIndex = (clampedPage - 1) * projectsPageSize;
    const endIndex = startIndex + projectsPageSize;
    return filteredProjects.slice(startIndex, endIndex);
  }, [filteredProjects, clampedPage]);

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="Manage research projects, budgets, and collaboration"
        counter={`(${projects.length} projects)`}
        actions={
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={onRefresh} disabled={loading}>
              {loading ? <Spinner /> : 'Refresh'}
            </Button>
            <Button
              variant="primary"
              data-testid="create-project-button"
              onClick={onCreateProject}
            >
              Create Project
            </Button>
          </SpaceBetween>
        }
      >
        Project Management
      </Header>

      <ColumnLayout columns={4} variant="text-grid">
        <Container header={<Header variant="h3">Total Projects</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-info">
            {projects.length}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Active Projects</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-success">
            {projects.filter(p => p.status === 'active').length}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Total Budget</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-body-secondary">
            ${projects.reduce((sum, p) => sum + (p.budget_limit || 0), 0).toFixed(2)}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Current Spend</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-warning">
            ${projects.reduce((sum, p) => sum + (p.current_spend || 0), 0).toFixed(2)}
          </Box>
        </Container>
      </ColumnLayout>

      <Container
        header={
          <Header
            variant="h2"
            description="Research projects with budget tracking and member management"
            counter={`(${filteredProjects.length})`}
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Select
                  selectedOption={{ label: projectFilter === 'all' ? 'All Projects' : projectFilter === 'active' ? 'Active Only' : 'Suspended', value: projectFilter }}
                  onChange={({ detail }) => { setProjectFilter(detail.selectedOption.value!); setProjectsCurrentPage(1); }}
                  options={[
                    { label: 'All Projects', value: 'all' },
                    { label: 'Active Only', value: 'active' },
                    { label: 'Suspended', value: 'suspended' }
                  ]}
                  data-testid="project-filter-select"
                />
                <Button>Export Data</Button>
                <Button variant="primary">Create Project</Button>
              </SpaceBetween>
            }
          >
            Projects
          </Header>
        }
      >
        <Table
          data-testid="projects-table"
          columnDefinitions={[
            {
              id: "name",
              header: "Project Name",
              cell: (item: Project) => (
                <Link
                  fontSize="body-m"
                  onFollow={() => {
                    onSelectProject(item.id);
                  }}
                >
                  {item.name}
                </Link>
              ),
              sortingField: "name"
            },
            {
              id: "description",
              header: "Description",
              cell: (item: Project) => item.description || 'No description',
              sortingField: "description"
            },
            {
              id: "owner",
              header: "Owner",
              cell: (item: Project) => item.owner_email || 'Unknown',
              sortingField: "owner_email"
            },
            {
              id: "budget",
              header: "Budget",
              cell: (item: Project) => {
                const budget = item.budget_status?.total_budget || (item as unknown as Record<string, number>).budget_limit || 0;
                return budget > 0 ? `$${budget.toFixed(2)}` : '-';
              },
              sortingField: "budget_status"
            },
            {
              id: "spend",
              header: "Current Spend",
              cell: (item: Project) => {
                const itemRecord = item as unknown as Record<string, number>;
                const spend = item.budget_status?.spent_amount || itemRecord.current_spend || 0;
                const limit = item.budget_status?.total_budget || itemRecord.budget_limit || 0;
                const percentage = limit > 0 ? (spend / limit) * 100 : 0;
                const colorType = percentage > 80 ? 'error' : percentage > 60 ? 'warning' : 'success';

                return (
                  <SpaceBetween direction="horizontal" size="xs">
                    <span {...(percentage > 80 ? { 'data-testid': 'budget-alert' } : {})}>
                      <StatusIndicator type={colorType}>
                        ${spend.toFixed(2)}
                      </StatusIndicator>
                    </span>
                    {limit > 0 && (
                      <Badge color={colorType === 'error' ? 'red' : colorType === 'warning' ? 'blue' : 'green'}>
                        {percentage.toFixed(1)}%
                      </Badge>
                    )}
                    {percentage >= 100 && (
                      <Box color="text-status-error" fontSize="body-s">Budget exceeded</Box>
                    )}
                  </SpaceBetween>
                );
              }
            },
            {
              id: "members",
              header: "Members",
              cell: (item: Project) => item.member_count || 1,
              sortingField: "member_count"
            },
            {
              id: "status",
              header: "Status",
              cell: (item: Project) => (
                <StatusIndicator
                  type={
                    item.status === 'active' ? 'success' :
                    item.status === 'suspended' ? 'warning' : 'error'
                  }
                >
                  {item.status || 'active'}
                </StatusIndicator>
              ),
              sortingField: "status"
            },
            {
              id: "created",
              header: "Created",
              cell: (item: Project) => new Date(item.created_at).toLocaleDateString(),
              sortingField: "created_at"
            },
            {
              id: "actions",
              header: "Actions",
              cell: (item: Project) => (
                <ButtonDropdown
                  data-testid={`project-actions-${item.id}`}
                  expandToViewport
                  items={[
                    { text: "View Details", id: "view" },
                    { text: "Manage Members", id: "members" },
                    { text: "Budget Analysis", id: "budget" },
                    { text: "Cost Report", id: "costs" },
                    { text: "Usage Statistics", id: "usage" },
                    { text: "Edit Project", id: "edit" },
                    { text: item.status === 'paused' ? "Resume Project" : "Suspend Project", id: "suspend" },
                    { text: "Delete", id: "delete" }
                  ]}
                  onItemClick={async (detail) => {
                    if (detail.detail.id === 'view') {
                      onSelectProject(item.id);
                    } else if (detail.detail.id === 'delete') {
                      onDeleteProject(item);
                    } else if (detail.detail.id === 'edit') {
                      onEditProject(item);
                    } else if (detail.detail.id === 'suspend') {
                      const newStatus = item.status === 'paused' ? 'active' : 'paused';
                      try {
                        await api.updateProject(item.id, { status: newStatus });
                        onRefresh();
                      } catch {
                        // caller can add notification handling if needed
                      }
                    } else if (detail.detail.id === 'members') {
                      onManageMembers(item);
                    } else if (detail.detail.id === 'budget') {
                      onManageBudget(item);
                    } else if (detail.detail.id === 'costs') {
                      onViewCost(item);
                    } else if (detail.detail.id === 'usage') {
                      onViewUsage(item);
                    }
                  }}
                >
                  Actions
                </ButtonDropdown>
              )
            }
          ]}
          items={paginatedProjects}
          trackBy="id"
          sortingColumn={projectsSortingColumn}
          onSortingChange={({ detail }) => setProjectsSortingColumn(detail.sortingColumn)}
          sortingDisabled={true}
          loadingText="Loading projects..."
          empty={
            <Box textAlign="center" color="text-body-secondary">
              <Box variant="strong" textAlign="center" color="text-body-secondary">
                No projects found
              </Box>
              <Box variant="p" padding={{ bottom: 's' }} color="text-body-secondary">
                Create your first research project to get started.
              </Box>
              <Button variant="primary">Create Project</Button>
            </Box>
          }
          header={
            <Header
              counter={`(${filteredProjects.length})`}
              description="Research projects with comprehensive budget and collaboration management"
            >
              All Projects
            </Header>
          }
          pagination={
            <Pagination
              currentPageIndex={clampedPage}
              pagesCount={projectsTotalPages}
              onChange={({ detail }) => setProjectsCurrentPage(detail.currentPageIndex)}
              ariaLabels={{
                nextPageLabel: 'Next page',
                previousPageLabel: 'Previous page',
                pageLabel: pageNumber => `Page ${pageNumber}`
              }}
            />
          }
        />
      </Container>

      <Container
        header={
          <Header
            variant="h2"
            description="Project analytics and budget insights"
          >
            Project Analytics
          </Header>
        }
      >
        <ColumnLayout columns={2}>
          <SpaceBetween size="m">
            <Header variant="h3">Budget Utilization</Header>
            {projects.length > 0 ? (
              projects.map((project) => {
                const spend = project.current_spend || 0;
                const limit = project.budget_limit || 0;
                const percentage = limit > 0 ? (spend / limit) * 100 : 0;

                return (
                  <Box key={project.id}>
                    <SpaceBetween direction="horizontal" size="s">
                      <Box fontWeight="bold">{project.name}:</Box>
                      <StatusIndicator
                        type={percentage > 80 ? 'error' : percentage > 60 ? 'warning' : 'success'}
                      >
                        ${spend.toFixed(2)} / ${limit.toFixed(2)} ({percentage.toFixed(1)}%)
                      </StatusIndicator>
                    </SpaceBetween>
                  </Box>
                );
              })
            ) : (
              <Box color="text-body-secondary">No projects to display</Box>
            )}
          </SpaceBetween>

          <SpaceBetween size="m">
            <Header variant="h3">Recent Activity</Header>
            <Box color="text-body-secondary">
              Project activity and cost tracking metrics will be displayed here.
              Connect projects to instances and storage for detailed analytics.
            </Box>
          </SpaceBetween>
        </ColumnLayout>
      </Container>
    </SpaceBetween>
  );
}
