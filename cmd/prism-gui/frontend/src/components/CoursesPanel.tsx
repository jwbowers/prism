/**
 * CoursesPanel — v0.17.0
 *
 * 5-tab course management panel:
 *   Overview | Members | Templates | Budget | Audit & Reports
 *
 * Rendered by CoursesManagementView in App.tsx when activeView === 'courses'.
 * Uses window.__apiClient (set in App.tsx).
 */

import React from 'react';
import {
  Tabs,
  Container,
  Header,
  SpaceBetween,
  Button,
  Table,
  Modal,
  Form,
  FormField,
  Input,
  Select,
  Box,
  Spinner,
  Badge,
  Alert,
  ColumnLayout,
  ProgressBar,
  DatePicker
} from '@cloudscape-design/components';

// ── Local types (mirror App.tsx interfaces) ──────────────────────────────────

interface ClassMember {
  user_id: string;
  email: string;
  display_name: string;
  role: string;
  budget_spent: number;
  budget_limit: number;
  added_at: string;
}

interface Course {
  id: string;
  code: string;
  title: string;
  department: string;
  semester: string;
  semester_start: string;
  semester_end: string;
  owner: string;
  status: string;
  members: ClassMember[];
  approved_templates: string[];
  per_student_budget: number;
  total_budget: number;
  default_template: string;
  auto_provision_on_enroll: boolean;
  created_at: string;
  updated_at: string;
  shared_materials_efs_id?: string;
  shared_materials_mount_path?: string;
  shared_materials_size_gb?: number;
}

interface CourseBudgetSummary {
  total_budget: number;
  per_student_default: number;
  total_spent: number;
  students: StudentBudgetInfo[];
}

interface StudentBudgetInfo {
  user_id: string;
  email: string;
  display_name: string;
  budget_limit: number;
  budget_spent: number;
  remaining: number;
}

interface CourseOverview {
  course_id: string;
  course_code: string;
  total_students: number;
  active_instances: number;
  total_budget_spent: number;
  students: StudentOverviewStatus[];
}

interface StudentOverviewStatus {
  user_id: string;
  email: string;
  display_name: string;
  budget_spent: number;
  budget_limit: number;
  budget_status: string;
}

interface CourseAuditEntry {
  timestamp: string;
  course_id: string;
  actor: string;
  target: string;
  action: string;
}

// v0.19.0 types

interface SharedMaterialsVolume {
  course_id: string;
  efs_id: string;
  size_gb: number;
  mount_path: string;
  state: string;
  created_at: string;
  mounted_instance_count: number;
}

interface WorkspaceResetResult {
  student_id: string;
  backup_snapshot_id?: string;
  backup_download_url?: string;
  backup_expires_at?: string;
  status: string;
}

// ── CourseDetailPanel ────────────────────────────────────────────────────────

interface CourseDetailPanelProps {
  course: Course;
  onBack: () => void;
  onRefresh: () => void;
}

export const CourseDetailPanel: React.FC<CourseDetailPanelProps> = ({ course, onBack, onRefresh }) => {
  const api = (window as any).__apiClient;

  // ── Members tab ─────────────────────────────────────────────────────────

  const [members, setMembers] = React.useState<ClassMember[]>([]);
  const [membersLoading, setMembersLoading] = React.useState(true);
  const [memberError, setMemberError] = React.useState<string | null>(null);
  const [enrollModalVisible, setEnrollModalVisible] = React.useState(false);
  const [enrollEmail, setEnrollEmail] = React.useState('');
  const [enrollUserId, setEnrollUserId] = React.useState('');
  const [enrollName, setEnrollName] = React.useState('');
  const [enrollRole, setEnrollRole] = React.useState({ label: 'Student', value: 'student' });
  const [enrollBudget, setEnrollBudget] = React.useState('');
  const [enrollSaving, setEnrollSaving] = React.useState(false);

  const loadMembers = async () => {
    setMembersLoading(true);
    try {
      const result = await api.getCourseMembers(course.id);
      setMembers(result);
    } catch (e: any) {
      setMemberError(e.message || 'Failed to load members');
    } finally {
      setMembersLoading(false);
    }
  };

  const enrollMember = async () => {
    setEnrollSaving(true);
    try {
      await api.enrollCourseMember(course.id, {
        user_id: enrollUserId,
        email: enrollEmail,
        display_name: enrollName,
        role: enrollRole.value,
        budget_limit: enrollBudget ? parseFloat(enrollBudget) : undefined,
      });
      setEnrollModalVisible(false);
      setEnrollEmail(''); setEnrollUserId(''); setEnrollName('');
      setEnrollBudget('');
      await loadMembers();
    } catch (e: any) {
      setMemberError(e.message || 'Enroll failed');
    } finally {
      setEnrollSaving(false);
    }
  };

  const unenrollMember = async (userId: string) => {
    try {
      await api.unenrollCourseMember(course.id, userId);
      await loadMembers();
    } catch (e: any) {
      setMemberError(e.message || 'Unenroll failed');
    }
  };

  // ── Templates tab ────────────────────────────────────────────────────────

  const [templates, setTemplates] = React.useState<string[]>([]);
  const [templatesLoading, setTemplatesLoading] = React.useState(true);
  const [templateError, setTemplateError] = React.useState<string | null>(null);
  const [addTemplateSlug, setAddTemplateSlug] = React.useState('');
  const [addTemplateSaving, setAddTemplateSaving] = React.useState(false);

  const loadTemplates = async () => {
    setTemplatesLoading(true);
    try {
      const result = await api.getCourseTemplates(course.id);
      setTemplates(result.templates || []);
    } catch (e: any) {
      setTemplateError(e.message || 'Failed to load templates');
    } finally {
      setTemplatesLoading(false);
    }
  };

  const addTemplate = async () => {
    if (!addTemplateSlug.trim()) return;
    setAddTemplateSaving(true);
    try {
      await api.addCourseTemplate(course.id, addTemplateSlug.trim());
      setAddTemplateSlug('');
      await loadTemplates();
    } catch (e: any) {
      setTemplateError(e.message || 'Failed to add template');
    } finally {
      setAddTemplateSaving(false);
    }
  };

  const removeTemplate = async (slug: string) => {
    try {
      await api.removeCourseTemplate(course.id, slug);
      await loadTemplates();
    } catch (e: any) {
      setTemplateError(e.message || 'Failed to remove template');
    }
  };

  // ── Budget tab ───────────────────────────────────────────────────────────

  const [budget, setBudget] = React.useState<CourseBudgetSummary | null>(null);
  const [budgetLoading, setBudgetLoading] = React.useState(true);
  const [budgetError, setBudgetError] = React.useState<string | null>(null);
  const [distributeModalVisible, setDistributeModalVisible] = React.useState(false);
  const [distributeAmount, setDistributeAmount] = React.useState('');
  const [distributeSaving, setDistributeSaving] = React.useState(false);

  const loadBudget = async () => {
    setBudgetLoading(true);
    try {
      const result = await api.getCourseBudget(course.id);
      setBudget(result);
    } catch (e: any) {
      setBudgetError(e.message || 'Failed to load budget');
    } finally {
      setBudgetLoading(false);
    }
  };

  const distributeBudget = async () => {
    const amount = parseFloat(distributeAmount);
    if (isNaN(amount) || amount <= 0) return;
    setDistributeSaving(true);
    try {
      await api.distributeCourseBudget(course.id, amount);
      setDistributeModalVisible(false);
      setDistributeAmount('');
      await loadBudget();
    } catch (e: any) {
      setBudgetError(e.message || 'Failed to distribute budget');
    } finally {
      setDistributeSaving(false);
    }
  };

  // ── Audit & Reports tab ──────────────────────────────────────────────────

  const [auditEntries, setAuditEntries] = React.useState<CourseAuditEntry[]>([]);
  const [auditLoading, setAuditLoading] = React.useState(true);
  const [auditError, setAuditError] = React.useState<string | null>(null);
  const [auditStudentFilter, setAuditStudentFilter] = React.useState('');
  const [archiveModalVisible, setArchiveModalVisible] = React.useState(false);
  const [archiving, setArchiving] = React.useState(false);

  const loadAudit = async () => {
    setAuditLoading(true);
    try {
      const params: Record<string, string | number> = {};
      if (auditStudentFilter) params.student_id = auditStudentFilter;
      const result = await api.getCourseAuditLog(course.id, params);
      setAuditEntries(result.entries || []);
    } catch (e: any) {
      setAuditError(e.message || 'Failed to load audit log');
    } finally {
      setAuditLoading(false);
    }
  };

  const downloadReport = async () => {
    try {
      const result = await api.getCourseReport(course.id, 'csv');
      const csv = typeof result === 'string' ? result : JSON.stringify(result, null, 2);
      const blob = new Blob([csv], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${course.code}-report.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e: any) {
      setAuditError(e.message || 'Failed to download report');
    }
  };

  const archiveCourse = async () => {
    setArchiving(true);
    setArchiveModalVisible(false);
    try {
      await api.archiveCourse(course.id);
      onRefresh();
      onBack();
    } catch (e: any) {
      setAuditError(e.message || 'Failed to archive course');
    } finally {
      setArchiving(false);
    }
  };

  // ── Overview tab ─────────────────────────────────────────────────────────

  const [overview, setOverview] = React.useState<CourseOverview | null>(null);
  const [overviewLoading, setOverviewLoading] = React.useState(true);
  const [overviewError, setOverviewError] = React.useState<string | null>(null);

  const loadOverview = async () => {
    setOverviewLoading(true);
    try {
      const result = await api.getCourseOverview(course.id);
      setOverview(result);
    } catch (e: any) {
      setOverviewError(e.message || 'Failed to load overview');
    } finally {
      setOverviewLoading(false);
    }
  };

  // ── TA Access tab (#48, #160) ─────────────────────────────────────────────

  const [taList, setTAList] = React.useState<ClassMember[]>([]);
  const [taLoading, setTALoading] = React.useState(false);
  const [taError, setTAError] = React.useState<string | null>(null);
  const [grantEmail, setGrantEmail] = React.useState('');
  const [grantName, setGrantName] = React.useState('');
  const [grantSaving, setGrantSaving] = React.useState(false);
  const [connectStudentId, setConnectStudentId] = React.useState('');
  const [connectReason, setConnectReason] = React.useState('');
  const [connectSshCommand, setConnectSshCommand] = React.useState('');
  const [connectLoading, setConnectLoading] = React.useState(false);
  const [connectModalVisible, setConnectModalVisible] = React.useState(false);

  const loadTAList = async () => {
    setTALoading(true);
    try {
      const result = await api.listCourseTAAccess(course.id);
      setTAList(result || []);
    } catch (e: any) {
      setTAError(e.message || 'Failed to load TAs');
    } finally {
      setTALoading(false);
    }
  };

  const grantTAAccess = async () => {
    if (!grantEmail) return;
    setGrantSaving(true);
    try {
      await api.grantCourseTAAccess(course.id, grantEmail, grantName);
      setGrantEmail(''); setGrantName('');
      await loadTAList();
    } catch (e: any) {
      setTAError(e.message || 'Failed to grant TA access');
    } finally {
      setGrantSaving(false);
    }
  };

  const revokeTAAccess = async (email: string) => {
    try {
      await api.revokeCourseTAAccess(course.id, email);
      await loadTAList();
    } catch (e: any) {
      setTAError(e.message || 'Failed to revoke TA access');
    }
  };

  const connectTA = async () => {
    if (!connectStudentId || !connectReason) return;
    setConnectLoading(true);
    try {
      const result = await api.connectCourseTAAccess(course.id, connectStudentId, connectReason);
      setConnectSshCommand(result.ssh_command || '');
    } catch (e: any) {
      setTAError(e.message || 'Failed to get SSH command');
    } finally {
      setConnectLoading(false);
    }
  };

  // ── Materials tab (#167) ─────────────────────────────────────────────────

  const [materials, setMaterials] = React.useState<SharedMaterialsVolume | null>(null);
  const [materialsLoading, setMaterialsLoading] = React.useState(false);
  const [materialsError, setMaterialsError] = React.useState<string | null>(null);
  const [createMaterialsSizeGB, setCreateMaterialsSizeGB] = React.useState('50');
  const [createMaterialsMount, setCreateMaterialsMount] = React.useState('/mnt/course-materials');
  const [createMaterialsSaving, setCreateMaterialsSaving] = React.useState(false);
  const [mountingMaterials, setMountingMaterials] = React.useState(false);
  const [mountResult, setMountResult] = React.useState<string | null>(null);

  const loadMaterials = async () => {
    setMaterialsLoading(true);
    try {
      const result = await api.getCourseMaterials(course.id);
      setMaterials(result);
    } catch (e: any) {
      setMaterialsError(e.message || 'Failed to load materials');
    } finally {
      setMaterialsLoading(false);
    }
  };

  const createMaterials = async () => {
    setCreateMaterialsSaving(true);
    try {
      const result = await api.createCourseMaterials(course.id, parseInt(createMaterialsSizeGB, 10), createMaterialsMount);
      setMaterials(result);
    } catch (e: any) {
      setMaterialsError(e.message || 'Failed to create materials volume');
    } finally {
      setCreateMaterialsSaving(false);
    }
  };

  const mountMaterials = async () => {
    setMountingMaterials(true);
    try {
      const result = await api.mountCourseMaterials(course.id);
      setMountResult(result.status || 'mount_scheduled');
    } catch (e: any) {
      setMaterialsError(e.message || 'Failed to schedule mount');
    } finally {
      setMountingMaterials(false);
    }
  };

  // ── Reset Workspace (#49, #164) ──────────────────────────────────────────

  const [resetStudentId, setResetStudentId] = React.useState('');
  const [resetReason, setResetReason] = React.useState('');
  const [resetBackup, setResetBackup] = React.useState(true);
  const [resetSaving, setResetSaving] = React.useState(false);
  const [resetResult, setResetResult] = React.useState<WorkspaceResetResult | null>(null);
  const [resetModalVisible, setResetModalVisible] = React.useState(false);

  const openResetModal = (studentId: string) => {
    setResetStudentId(studentId);
    setResetReason('');
    setResetBackup(true);
    setResetResult(null);
    setResetModalVisible(true);
  };

  const resetWorkspace = async () => {
    if (!resetStudentId || !resetReason) return;
    setResetSaving(true);
    try {
      const result = await api.resetCourseStudentWorkspace(course.id, resetStudentId, resetReason, resetBackup);
      setResetResult(result);
    } catch (e: any) {
      setOverviewError(e.message || 'Reset failed');
      setResetModalVisible(false);
    } finally {
      setResetSaving(false);
    }
  };

  // Load all tabs on mount
  React.useEffect(() => {
    loadOverview();
    loadMembers();
    loadTemplates();
    loadBudget();
    loadAudit();
    loadTAList();
    loadMaterials();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [course.id]);

  const budgetStatusColor = (status: string) => {
    if (status === 'exceeded') return 'red';
    if (status === 'warning') return 'blue';
    return 'green';
  };

  const budgetBarColor = (spent: number, limit: number) => {
    if (limit <= 0) return 'in-progress';
    const pct = spent / limit;
    if (pct >= 1) return 'error';
    if (pct >= 0.8) return 'in-progress';
    return 'success';
  };

  return (
    <SpaceBetween size="l" data-testid="course-detail-panel">
      <Container
        header={
          <Header
            variant="h2"
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Button onClick={onBack} variant="link">← Back to courses</Button>
                <Button onClick={onRefresh} iconName="refresh">Refresh</Button>
              </SpaceBetween>
            }
          >
            {course.code} — {course.title}
          </Header>
        }
      >
        <ColumnLayout columns={4} variant="text-grid">
          <div><Box variant="awsui-key-label">Status</Box><Badge color={course.status === 'active' ? 'green' : course.status === 'archived' ? 'grey' : 'blue'}>{course.status}</Badge></div>
          <div><Box variant="awsui-key-label">Semester</Box><Box>{course.semester}</Box></div>
          <div><Box variant="awsui-key-label">Owner</Box><Box>{course.owner}</Box></div>
          <div><Box variant="awsui-key-label">Department</Box><Box>{course.department || '—'}</Box></div>
        </ColumnLayout>
      </Container>

      <Tabs
        data-testid="course-tabs"
        tabs={[
          {
            id: 'overview',
            label: 'Overview',
            content: (
              <Container header={<Header variant="h3">Course Overview</Header>}>
                {overviewError && <Alert type="error" dismissible onDismiss={() => setOverviewError(null)}>{overviewError}</Alert>}
                {overviewLoading ? <Spinner /> : overview ? (
                  <SpaceBetween size="m">
                    <ColumnLayout columns={3} variant="text-grid">
                      <div><Box variant="awsui-key-label">Total Students</Box><Box variant="awsui-value-large">{overview.total_students}</Box></div>
                      <div><Box variant="awsui-key-label">Active Instances</Box><Box variant="awsui-value-large">{overview.active_instances}</Box></div>
                      <div><Box variant="awsui-key-label">Total Spent</Box><Box variant="awsui-value-large">${overview.total_budget_spent?.toFixed(2) || '0.00'}</Box></div>
                    </ColumnLayout>
                    <Table
                      data-testid="overview-students-table"
                      header={<Header variant="h3">Students</Header>}
                      loading={overviewLoading}
                      items={overview.students || []}
                      columnDefinitions={[
                        { id: 'name', header: 'Name', cell: (s: StudentOverviewStatus) => s.display_name },
                        { id: 'email', header: 'Email', cell: (s: StudentOverviewStatus) => s.email },
                        { id: 'spent', header: 'Spent', cell: (s: StudentOverviewStatus) => `$${s.budget_spent?.toFixed(2) || '0.00'}` },
                        { id: 'limit', header: 'Limit', cell: (s: StudentOverviewStatus) => `$${s.budget_limit?.toFixed(2) || '0.00'}` },
                        { id: 'status', header: 'Budget Status', cell: (s: StudentOverviewStatus) => <Badge color={budgetStatusColor(s.budget_status)}>{s.budget_status}</Badge> },
                        { id: 'actions', header: 'Actions', cell: (s: StudentOverviewStatus) => (
                          <Button
                            variant="normal"
                            onClick={() => openResetModal(s.user_id || s.email)}
                            data-testid={`reset-workspace-${s.user_id}`}
                          >
                            Reset Workspace
                          </Button>
                        )},
                      ]}
                    />
                  </SpaceBetween>
                ) : <Box color="text-body-secondary">No overview data available.</Box>}
              </Container>
            )
          },
          {
            id: 'members',
            label: 'Members',
            content: (
              <Container
                header={
                  <Header
                    variant="h3"
                    actions={
                      <Button onClick={() => setEnrollModalVisible(true)} iconName="add-plus" data-testid="enroll-member-button">
                        Enroll Member
                      </Button>
                    }
                  >
                    Enrolled Members
                  </Header>
                }
              >
                {memberError && <Alert type="error" dismissible onDismiss={() => setMemberError(null)}>{memberError}</Alert>}
                <Table
                  data-testid="members-table"
                  loading={membersLoading}
                  loadingText="Loading members..."
                  items={members}
                  empty={<Box textAlign="center" color="text-body-secondary">No members enrolled.</Box>}
                  columnDefinitions={[
                    { id: 'name', header: 'Name', cell: (m: ClassMember) => m.display_name || m.user_id },
                    { id: 'email', header: 'Email', cell: (m: ClassMember) => m.email },
                    { id: 'role', header: 'Role', cell: (m: ClassMember) => <Badge color={m.role === 'ta' ? 'blue' : m.role === 'instructor' ? 'green' : 'grey'}>{m.role}</Badge> },
                    { id: 'spent', header: 'Budget Spent', cell: (m: ClassMember) => `$${m.budget_spent?.toFixed(2) || '0.00'}` },
                    { id: 'limit', header: 'Budget Limit', cell: (m: ClassMember) => m.budget_limit ? `$${m.budget_limit.toFixed(2)}` : 'Unlimited' },
                    {
                      id: 'actions',
                      header: 'Actions',
                      cell: (m: ClassMember) => (
                        <Button variant="link" onClick={() => unenrollMember(m.user_id)}>Unenroll</Button>
                      )
                    },
                  ]}
                />

                <Modal
                  visible={enrollModalVisible}
                  onDismiss={() => setEnrollModalVisible(false)}
                  header="Enroll Member"
                  footer={
                    <Box float="right">
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setEnrollModalVisible(false)}>Cancel</Button>
                        <Button variant="primary" loading={enrollSaving} onClick={enrollMember} data-testid="enroll-submit-button">Enroll</Button>
                      </SpaceBetween>
                    </Box>
                  }
                >
                  <Form>
                    <SpaceBetween size="m">
                      <FormField label="Email">
                        <Input value={enrollEmail} onChange={({ detail }) => setEnrollEmail(detail.value)} placeholder="student@uni.edu" data-testid="enroll-email-input" />
                      </FormField>
                      <FormField label="User ID">
                        <Input value={enrollUserId} onChange={({ detail }) => setEnrollUserId(detail.value)} placeholder="student123" />
                      </FormField>
                      <FormField label="Display Name">
                        <Input value={enrollName} onChange={({ detail }) => setEnrollName(detail.value)} placeholder="Alice Smith" />
                      </FormField>
                      <FormField label="Role">
                        <Select
                          selectedOption={enrollRole}
                          onChange={({ detail }) => setEnrollRole(detail.selectedOption as any)}
                          options={[
                            { label: 'Student', value: 'student' },
                            { label: 'TA', value: 'ta' },
                            { label: 'Instructor', value: 'instructor' },
                          ]}
                        />
                      </FormField>
                      <FormField label="Budget Limit (USD, optional)">
                        <Input value={enrollBudget} onChange={({ detail }) => setEnrollBudget(detail.value)} placeholder="e.g. 50.00" inputMode="decimal" />
                      </FormField>
                    </SpaceBetween>
                  </Form>
                </Modal>
              </Container>
            )
          },
          {
            id: 'templates',
            label: 'Templates',
            content: (
              <Container header={
                <Header
                  variant="h3"
                  actions={
                    templates.length > 0
                      ? <Badge color="blue" data-testid="enforcement-active-badge">Enforcement Active</Badge>
                      : <Badge color="grey" data-testid="enforcement-unrestricted-badge">Unrestricted</Badge>
                  }
                >
                  Approved Templates Whitelist
                </Header>
              }>
                {templateError && <Alert type="error" dismissible onDismiss={() => setTemplateError(null)}>{templateError}</Alert>}
                {templatesLoading ? <Spinner /> : (
                  <SpaceBetween size="m">
                    {templates.length === 0 ? (
                      <Alert type="info">No whitelist set — all templates are allowed for this course.</Alert>
                    ) : (
                      <Table
                        data-testid="templates-table"
                        items={templates.map(slug => ({ slug }))}
                        columnDefinitions={[
                          { id: 'slug', header: 'Template Slug', cell: (row: { slug: string }) => row.slug },
                          {
                            id: 'actions',
                            header: 'Actions',
                            cell: (row: { slug: string }) => (
                              <Button variant="link" onClick={() => removeTemplate(row.slug)}>Remove</Button>
                            )
                          },
                        ]}
                      />
                    )}
                    <SpaceBetween direction="horizontal" size="xs">
                      <Input
                        value={addTemplateSlug}
                        onChange={({ detail }) => setAddTemplateSlug(detail.value)}
                        placeholder="e.g. python-ml"
                        data-testid="add-template-input"
                      />
                      <Button onClick={addTemplate} loading={addTemplateSaving} data-testid="add-template-button">Add Template</Button>
                    </SpaceBetween>
                  </SpaceBetween>
                )}
              </Container>
            )
          },
          {
            id: 'budget',
            label: 'Budget',
            content: (
              <Container
                header={
                  <Header
                    variant="h3"
                    actions={
                      <Button onClick={() => setDistributeModalVisible(true)} data-testid="distribute-budget-button">Distribute Budget</Button>
                    }
                  >
                    Budget Summary
                  </Header>
                }
              >
                {budgetError && <Alert type="error" dismissible onDismiss={() => setBudgetError(null)}>{budgetError}</Alert>}
                {budgetLoading ? <Spinner /> : budget ? (
                  <SpaceBetween size="m">
                    <ColumnLayout columns={3} variant="text-grid">
                      <div><Box variant="awsui-key-label">Total Budget</Box><Box>${budget.total_budget?.toFixed(2) || '0.00'}</Box></div>
                      <div><Box variant="awsui-key-label">Total Spent</Box><Box>${budget.total_spent?.toFixed(2) || '0.00'}</Box></div>
                      <div><Box variant="awsui-key-label">Per-Student Default</Box><Box>${budget.per_student_default?.toFixed(2) || '0.00'}</Box></div>
                    </ColumnLayout>
                    <Table
                      data-testid="budget-students-table"
                      items={budget.students || []}
                      columnDefinitions={[
                        { id: 'name', header: 'Student', cell: (s: StudentBudgetInfo) => s.display_name || s.user_id },
                        { id: 'email', header: 'Email', cell: (s: StudentBudgetInfo) => s.email },
                        {
                          id: 'progress',
                          header: 'Budget Used',
                          cell: (s: StudentBudgetInfo) => (
                            <ProgressBar
                              value={s.budget_limit > 0 ? (s.budget_spent / s.budget_limit) * 100 : 0}
                              additionalInfo={`$${s.budget_spent?.toFixed(2)} / $${s.budget_limit?.toFixed(2)}`}
                              status={budgetBarColor(s.budget_spent, s.budget_limit)}
                            />
                          )
                        },
                        { id: 'remaining', header: 'Remaining', cell: (s: StudentBudgetInfo) => `$${s.remaining?.toFixed(2) || '0.00'}` },
                      ]}
                    />
                  </SpaceBetween>
                ) : <Box color="text-body-secondary">No budget data.</Box>}

                <Modal
                  visible={distributeModalVisible}
                  onDismiss={() => setDistributeModalVisible(false)}
                  header="Distribute Budget"
                  footer={
                    <Box float="right">
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setDistributeModalVisible(false)}>Cancel</Button>
                        <Button variant="primary" loading={distributeSaving} onClick={distributeBudget} data-testid="distribute-submit-button">Set Budget</Button>
                      </SpaceBetween>
                    </Box>
                  }
                >
                  <FormField label="Amount per student (USD)">
                    <Input
                      value={distributeAmount}
                      onChange={({ detail }) => setDistributeAmount(detail.value)}
                      placeholder="e.g. 50.00"
                      inputMode="decimal"
                      data-testid="distribute-amount-input"
                    />
                  </FormField>
                </Modal>
              </Container>
            )
          },
          {
            id: 'audit',
            label: 'Audit & Reports',
            content: (
              <Container header={<Header variant="h3">Audit Log & Reports</Header>}>
                {auditError && <Alert type="error" dismissible onDismiss={() => setAuditError(null)}>{auditError}</Alert>}
                <SpaceBetween size="m">
                  <SpaceBetween direction="horizontal" size="xs">
                    <Input
                      value={auditStudentFilter}
                      onChange={({ detail }) => setAuditStudentFilter(detail.value)}
                      placeholder="Filter by student ID"
                      data-testid="audit-student-filter"
                    />
                    <Button onClick={loadAudit} iconName="search">Filter</Button>
                    <Button onClick={downloadReport} iconName="download" data-testid="download-report-button">Download CSV Report</Button>
                    {course.status === 'closed' && (
                      <Button
                        variant="normal"
                        onClick={() => setArchiveModalVisible(true)}
                        loading={archiving}
                        data-testid="archive-course-button"
                      >
                        Archive Course
                      </Button>
                    )}
                  </SpaceBetween>
                  <Table
                    data-testid="audit-table"
                    loading={auditLoading}
                    loadingText="Loading audit log..."
                    items={auditEntries}
                    empty={<Box textAlign="center" color="text-body-secondary">No audit entries found.</Box>}
                    columnDefinitions={[
                      { id: 'time', header: 'Time', cell: (e: CourseAuditEntry) => new Date(e.timestamp).toLocaleString() },
                      { id: 'actor', header: 'Actor', cell: (e: CourseAuditEntry) => e.actor },
                      { id: 'action', header: 'Action', cell: (e: CourseAuditEntry) => e.action },
                      { id: 'target', header: 'Target', cell: (e: CourseAuditEntry) => e.target || '—' },
                    ]}
                  />
                </SpaceBetween>

                <Modal
                  visible={archiveModalVisible}
                  onDismiss={() => setArchiveModalVisible(false)}
                  header="Archive Course"
                  footer={
                    <Box float="right">
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setArchiveModalVisible(false)}>Cancel</Button>
                        <Button variant="primary" onClick={archiveCourse} loading={archiving} data-testid="archive-confirm-button">Archive</Button>
                      </SpaceBetween>
                    </Box>
                  }
                >
                  <Box>Archive course <strong>{course.code}</strong>? All student instances will be stopped. This cannot be undone.</Box>
                </Modal>
              </Container>
            )
          },
          {
            id: 'ta-access',
            label: 'TA Access',
            content: (
              <Container header={<Header variant="h3">TA Access Management</Header>}>
                {taError && <Alert type="error" dismissible onDismiss={() => setTAError(null)}>{taError}</Alert>}
                <SpaceBetween size="m">
                  <SpaceBetween direction="horizontal" size="xs">
                    <Input
                      value={grantEmail}
                      onChange={({ detail }) => setGrantEmail(detail.value)}
                      placeholder="ta@uni.edu"
                      data-testid="ta-grant-email-input"
                    />
                    <Input
                      value={grantName}
                      onChange={({ detail }) => setGrantName(detail.value)}
                      placeholder="Display name (optional)"
                    />
                    <Button onClick={grantTAAccess} loading={grantSaving} data-testid="ta-grant-button">Grant TA Access</Button>
                  </SpaceBetween>

                  <Table
                    data-testid="ta-access-table"
                    loading={taLoading}
                    loadingText="Loading TAs..."
                    items={taList}
                    empty={<Box textAlign="center" color="text-body-secondary">No TAs configured. Grant access above.</Box>}
                    columnDefinitions={[
                      { id: 'email', header: 'Email', cell: (m: ClassMember) => m.email },
                      { id: 'name', header: 'Name', cell: (m: ClassMember) => m.display_name || '—' },
                      { id: 'added', header: 'Added', cell: (m: ClassMember) => new Date(m.added_at).toLocaleDateString() },
                      {
                        id: 'actions', header: 'Actions',
                        cell: (m: ClassMember) => (
                          <SpaceBetween direction="horizontal" size="xs">
                            <Button variant="link" onClick={() => { setConnectStudentId(''); setConnectReason(''); setConnectSshCommand(''); setConnectModalVisible(true); }}>Connect to Student</Button>
                            <Button variant="link" onClick={() => revokeTAAccess(m.email)}>Revoke</Button>
                          </SpaceBetween>
                        )
                      },
                    ]}
                  />

                  <Modal
                    visible={connectModalVisible}
                    onDismiss={() => setConnectModalVisible(false)}
                    header="TA Connect to Student Instance"
                    data-testid="ta-connect-modal"
                    footer={
                      <Box float="right">
                        <SpaceBetween direction="horizontal" size="xs">
                          <Button variant="link" onClick={() => setConnectModalVisible(false)}>Close</Button>
                          <Button variant="primary" loading={connectLoading} onClick={connectTA} data-testid="ta-connect-submit">Get SSH Command</Button>
                        </SpaceBetween>
                      </Box>
                    }
                  >
                    <Form>
                      <SpaceBetween size="m">
                        <FormField label="Student Email or ID">
                          <Input value={connectStudentId} onChange={({ detail }) => setConnectStudentId(detail.value)} placeholder="student@uni.edu" data-testid="ta-connect-student-input" />
                        </FormField>
                        <FormField label="Reason (required — recorded in audit log)">
                          <Input value={connectReason} onChange={({ detail }) => setConnectReason(detail.value)} placeholder="e.g. office hours debugging" data-testid="ta-connect-reason-input" />
                        </FormField>
                        {connectSshCommand && (
                          <Alert type="success" data-testid="ta-connect-result">
                            <Box variant="code">{connectSshCommand}</Box>
                          </Alert>
                        )}
                      </SpaceBetween>
                    </Form>
                  </Modal>
                </SpaceBetween>
              </Container>
            )
          },
          {
            id: 'materials',
            label: 'Materials',
            content: (
              <Container header={<Header variant="h3">Shared Course Materials (EFS)</Header>}>
                {materialsError && <Alert type="error" dismissible onDismiss={() => setMaterialsError(null)}>{materialsError}</Alert>}
                {mountResult && <Alert type="success" dismissible onDismiss={() => setMountResult(null)}>Mount scheduled: {mountResult}</Alert>}
                {materialsLoading ? <Spinner /> : materials ? (
                  <SpaceBetween size="m">
                    <ColumnLayout columns={4} variant="text-grid">
                      <div><Box variant="awsui-key-label">EFS ID</Box><Box data-testid="materials-efs-id">{materials.efs_id}</Box></div>
                      <div><Box variant="awsui-key-label">Size</Box><Box>{materials.size_gb} GB</Box></div>
                      <div><Box variant="awsui-key-label">Mount Path</Box><Box>{materials.mount_path}</Box></div>
                      <div><Box variant="awsui-key-label">State</Box><Badge color={materials.state === 'available' ? 'green' : materials.state === 'creating' ? 'blue' : 'red'} data-testid="materials-state-badge">{materials.state}</Badge></div>
                    </ColumnLayout>
                    <Button onClick={mountMaterials} loading={mountingMaterials} data-testid="mount-materials-button">
                      Mount on All Student Instances
                    </Button>
                  </SpaceBetween>
                ) : (
                  <SpaceBetween size="m">
                    <Box color="text-body-secondary">No shared materials volume. Create one below.</Box>
                    <SpaceBetween direction="horizontal" size="xs">
                      <FormField label="Size (GB)">
                        <Input
                          value={createMaterialsSizeGB}
                          onChange={({ detail }) => setCreateMaterialsSizeGB(detail.value)}
                          data-testid="materials-size-input"
                          type="number"
                        />
                      </FormField>
                      <FormField label="Mount Path">
                        <Input
                          value={createMaterialsMount}
                          onChange={({ detail }) => setCreateMaterialsMount(detail.value)}
                          data-testid="materials-mount-input"
                        />
                      </FormField>
                      <FormField label=" ">
                        <Button onClick={createMaterials} loading={createMaterialsSaving} data-testid="create-materials-button">
                          Create Materials Volume
                        </Button>
                      </FormField>
                    </SpaceBetween>
                  </SpaceBetween>
                )}

                <Modal
                  visible={resetModalVisible}
                  onDismiss={() => setResetModalVisible(false)}
                  header="Reset Student Workspace"
                  data-testid="reset-workspace-modal"
                  footer={
                    <Box float="right">
                      <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setResetModalVisible(false)}>Cancel</Button>
                        <Button variant="primary" onClick={resetWorkspace} loading={resetSaving} data-testid="reset-workspace-confirm">Reset</Button>
                      </SpaceBetween>
                    </Box>
                  }
                >
                  <Form>
                    <SpaceBetween size="m">
                      <FormField label="Student">
                        <Box>{resetStudentId}</Box>
                      </FormField>
                      <FormField label="Reason (required — recorded in audit log)">
                        <Input value={resetReason} onChange={({ detail }) => setResetReason(detail.value)} placeholder="e.g. broken environment" data-testid="reset-reason-input" />
                      </FormField>
                      <FormField label="Create backup before reset">
                        <Button
                          variant={resetBackup ? 'primary' : 'normal'}
                          onClick={() => setResetBackup(!resetBackup)}
                          data-testid="reset-backup-toggle"
                        >
                          {resetBackup ? 'Backup: ON' : 'Backup: OFF'}
                        </Button>
                      </FormField>
                      {resetResult && (
                        <Alert type="success" data-testid="reset-result">
                          Reset scheduled. {resetResult.backup_download_url ? `Backup URL: ${resetResult.backup_download_url}` : ''}
                        </Alert>
                      )}
                    </SpaceBetween>
                  </Form>
                </Modal>
              </Container>
            )
          }
        ]}
      />
    </SpaceBetween>
  );
};

// ── CoursesPanel — course list view ─────────────────────────────────────────

export const CoursesPanel: React.FC = () => {
  const api = (window as any).__apiClient;

  const [courses, setCourses] = React.useState<Course[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [selectedCourse, setSelectedCourse] = React.useState<Course | null>(null);
  const [createModalVisible, setCreateModalVisible] = React.useState(false);
  const [saving, setSaving] = React.useState(false);

  // Create form state
  const [newCode, setNewCode] = React.useState('');
  const [newTitle, setNewTitle] = React.useState('');
  const [newDept, setNewDept] = React.useState('');
  const [newSemester, setNewSemester] = React.useState('');
  const [newStart, setNewStart] = React.useState('');
  const [newEnd, setNewEnd] = React.useState('');
  const [newOwner, setNewOwner] = React.useState('');
  const [newBudget, setNewBudget] = React.useState('');

  const loadCourses = async () => {
    setLoading(true);
    try {
      const result = await api.getCourses();
      setCourses(result || []);
    } catch (e: any) {
      setError(e.message || 'Failed to load courses');
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => { loadCourses(); }, []);

  const createCourse = async () => {
    setSaving(true);
    try {
      await api.createCourse({
        code: newCode,
        title: newTitle,
        department: newDept,
        semester: newSemester,
        semester_start: newStart ? `${newStart}T00:00:00Z` : undefined,
        semester_end: newEnd ? `${newEnd}T00:00:00Z` : undefined,
        owner: newOwner,
        per_student_budget: newBudget ? parseFloat(newBudget) : 0,
      });
      setCreateModalVisible(false);
      setNewCode(''); setNewTitle(''); setNewDept(''); setNewSemester('');
      setNewStart(''); setNewEnd(''); setNewOwner(''); setNewBudget('');
      await loadCourses();
    } catch (e: any) {
      setError(e.message || 'Failed to create course');
    } finally {
      setSaving(false);
    }
  };

  const statusColor = (status: string) => {
    switch (status) {
      case 'active': return 'green';
      case 'archived': return 'grey';
      case 'closed': return 'red';
      default: return 'blue';
    }
  };

  if (selectedCourse) {
    return (
      <CourseDetailPanel
        course={selectedCourse}
        onBack={() => setSelectedCourse(null)}
        onRefresh={loadCourses}
      />
    );
  }

  return (
    <SpaceBetween size="l" data-testid="courses-panel">
      {error && <Alert type="error" dismissible onDismiss={() => setError(null)}>{error}</Alert>}

      <Table
        data-testid="courses-table"
        loading={loading}
        loadingText="Loading courses..."
        header={
          <Header
            variant="h2"
            counter={`(${courses.length})`}
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Button onClick={loadCourses} iconName="refresh" data-testid="refresh-courses-button">Refresh</Button>
                <Button variant="primary" onClick={() => setCreateModalVisible(true)} iconName="add-plus" data-testid="create-course-button">Create Course</Button>
              </SpaceBetween>
            }
          >
            Courses
          </Header>
        }
        empty={
          <Box textAlign="center" color="text-body-secondary" padding="xl">
            <Box variant="strong">No courses found.</Box>
            <Box>Create a course to get started.</Box>
          </Box>
        }
        items={courses}
        onRowClick={({ detail }) => setSelectedCourse(detail.item)}
        columnDefinitions={[
          { id: 'code', header: 'Code', cell: (c: Course) => c.code },
          { id: 'title', header: 'Title', cell: (c: Course) => c.title },
          { id: 'semester', header: 'Semester', cell: (c: Course) => c.semester },
          { id: 'status', header: 'Status', cell: (c: Course) => <Badge color={statusColor(c.status)}>{c.status}</Badge> },
          { id: 'students', header: 'Students', cell: (c: Course) => c.members?.length ?? 0 },
          { id: 'budget', header: 'Budget/Student', cell: (c: Course) => c.per_student_budget ? `$${c.per_student_budget.toFixed(2)}` : '—' },
          {
            id: 'actions',
            header: 'Actions',
            cell: (c: Course) => (
              <Button variant="link" onClick={() => setSelectedCourse(c)}>View</Button>
            )
          },
        ]}
      />

      <Modal
        visible={createModalVisible}
        onDismiss={() => setCreateModalVisible(false)}
        header="Create Course"
        size="large"
        data-testid="create-course-modal"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setCreateModalVisible(false)}>Cancel</Button>
              <Button variant="primary" loading={saving} onClick={createCourse} data-testid="create-course-submit">Create</Button>
            </SpaceBetween>
          </Box>
        }
      >
        <Form>
          <SpaceBetween size="m">
            <ColumnLayout columns={2}>
              <FormField label="Course Code" constraintText="e.g. CS101">
                <Input value={newCode} onChange={({ detail }) => setNewCode(detail.value)} placeholder="CS101" data-testid="course-code-input" />
              </FormField>
              <FormField label="Title">
                <Input value={newTitle} onChange={({ detail }) => setNewTitle(detail.value)} placeholder="Introduction to Computer Science" data-testid="course-title-input" />
              </FormField>
              <FormField label="Department">
                <Input value={newDept} onChange={({ detail }) => setNewDept(detail.value)} placeholder="Computer Science" />
              </FormField>
              <FormField label="Semester">
                <Input value={newSemester} onChange={({ detail }) => setNewSemester(detail.value)} placeholder="Fall 2099" />
              </FormField>
              <FormField label="Semester Start">
                <DatePicker value={newStart} onChange={({ detail }) => setNewStart(detail.value)} placeholder="YYYY-MM-DD" />
              </FormField>
              <FormField label="Semester End">
                <DatePicker value={newEnd} onChange={({ detail }) => setNewEnd(detail.value)} placeholder="YYYY-MM-DD" />
              </FormField>
              <FormField label="Owner (User ID)">
                <Input value={newOwner} onChange={({ detail }) => setNewOwner(detail.value)} placeholder="prof-user-id" />
              </FormField>
              <FormField label="Per-Student Budget (USD)">
                <Input value={newBudget} onChange={({ detail }) => setNewBudget(detail.value)} placeholder="50.00" inputMode="decimal" />
              </FormField>
            </ColumnLayout>
          </SpaceBetween>
        </Form>
      </Modal>
    </SpaceBetween>
  );
};
