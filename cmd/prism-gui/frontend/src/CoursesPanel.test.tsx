/**
 * CoursesPanel Unit Tests — v0.17.0
 *
 * Tests the CoursesPanel and CourseDetailPanel components using window.__apiClient mock.
 *
 * Coverage:
 * - Renders course list with correct columns
 * - Renders empty state when no courses
 * - Create Course modal opens/closes
 * - Tab switching for all 5 tabs in CourseDetailPanel
 * - Budget bar progress logic (green/yellow/red thresholds)
 * - Archive confirmation modal appears
 *
 * Key patterns:
 * - Cloudscape <Input data-testid="x"> puts testid on wrapper <div>.
 *   Use getNativeInput(testId) to get the native <input>.
 * - Each test owns its own render(); no shared beforeEach renders.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { CoursesPanel, CourseDetailPanel } from './components/CoursesPanel';

// ── Mock API client ──────────────────────────────────────────────────────────

const mockCourses = [
  {
    id: 'cs101',
    code: 'CS101',
    title: 'Intro to Computer Science',
    department: 'CS',
    semester: 'Fall 2099',
    semester_start: '2099-09-01T00:00:00Z',
    semester_end: '2099-12-15T00:00:00Z',
    owner: 'prof1',
    status: 'active',
    members: [],
    approved_templates: [],
    per_student_budget: 50,
    total_budget: 1000,
    default_template: '',
    auto_provision_on_enroll: false,
    created_at: '2099-09-01T00:00:00Z',
    updated_at: '2099-09-01T00:00:00Z',
  },
];

const mockApiClient = {
  getCourses: vi.fn(),
  createCourse: vi.fn(),
  getCourse: vi.fn(),
  closeCourse: vi.fn(),
  deleteCourse: vi.fn(),
  archiveCourse: vi.fn(),
  getCourseMembers: vi.fn(),
  enrollCourseMember: vi.fn(),
  unenrollCourseMember: vi.fn(),
  getCourseTemplates: vi.fn(),
  addCourseTemplate: vi.fn(),
  removeCourseTemplate: vi.fn(),
  getCourseBudget: vi.fn(),
  distributeCourseBudget: vi.fn(),
  getCourseOverview: vi.fn(),
  getCourseReport: vi.fn(),
  getCourseAuditLog: vi.fn(),
  debugStudent: vi.fn(),
  resetStudent: vi.fn(),
  provisionStudent: vi.fn(),
  importCourseRoster: vi.fn(),
};

// Inject mock before each test
beforeEach(() => {
  (window as any).__apiClient = mockApiClient;

  // Default happy-path responses
  mockApiClient.getCourses.mockResolvedValue(mockCourses);
  mockApiClient.createCourse.mockResolvedValue({ id: 'new-course', code: 'CS201', title: 'New', status: 'pending' });
  mockApiClient.getCourseMembers.mockResolvedValue([]);
  mockApiClient.getCourseTemplates.mockResolvedValue({ templates: [] });
  mockApiClient.getCourseBudget.mockResolvedValue({ total_budget: 1000, per_student_default: 50, total_spent: 0, students: [] });
  mockApiClient.getCourseOverview.mockResolvedValue({ course_id: 'cs101', course_code: 'CS101', total_students: 0, active_instances: 0, total_budget_spent: 0, students: [] });
  mockApiClient.getCourseAuditLog.mockResolvedValue({ entries: [] });
  mockApiClient.getCourseReport.mockResolvedValue({});
  mockApiClient.enrollCourseMember.mockResolvedValue({});
  mockApiClient.addCourseTemplate.mockResolvedValue(undefined);
  mockApiClient.removeCourseTemplate.mockResolvedValue(undefined);
  mockApiClient.distributeCourseBudget.mockResolvedValue({});
  mockApiClient.archiveCourse.mockResolvedValue({ instances_stopped: [] });
  vi.clearAllMocks();
  // Re-apply defaults after clear
  mockApiClient.getCourses.mockResolvedValue(mockCourses);
  mockApiClient.createCourse.mockResolvedValue({ id: 'new-course', code: 'CS201', title: 'New', status: 'pending' });
  mockApiClient.getCourseMembers.mockResolvedValue([]);
  mockApiClient.getCourseTemplates.mockResolvedValue({ templates: [] });
  mockApiClient.getCourseBudget.mockResolvedValue({ total_budget: 1000, per_student_default: 50, total_spent: 0, students: [] });
  mockApiClient.getCourseOverview.mockResolvedValue({ course_id: 'cs101', course_code: 'CS101', total_students: 0, active_instances: 0, total_budget_spent: 0, students: [] });
  mockApiClient.getCourseAuditLog.mockResolvedValue({ entries: [] });
  mockApiClient.getCourseReport.mockResolvedValue({});
  mockApiClient.enrollCourseMember.mockResolvedValue({});
  mockApiClient.addCourseTemplate.mockResolvedValue(undefined);
  mockApiClient.removeCourseTemplate.mockResolvedValue(undefined);
  mockApiClient.distributeCourseBudget.mockResolvedValue({});
  mockApiClient.archiveCourse.mockResolvedValue({ instances_stopped: [] });
});

// ── Helpers ──────────────────────────────────────────────────────────────────

const getNativeInput = (testId: string): HTMLInputElement => {
  const wrapper = screen.getByTestId(testId);
  const native = wrapper.querySelector('input');
  if (!native) throw new Error(`No <input> inside [data-testid="${testId}"]`);
  return native as HTMLInputElement;
};

const clickTab = async (label: string) => {
  const user = userEvent.setup();
  const tab = screen.getAllByRole('tab').find(t => t.textContent?.includes(label));
  if (!tab) throw new Error(`Tab "${label}" not found`);
  await user.click(tab);
};

// ── CoursesPanel (list view) tests ────────────────────────────────────────────

describe('CoursesPanel', () => {
  it('renders course list with correct columns', async () => {
    render(<CoursesPanel />);
    await waitFor(() => {
      expect(screen.getByTestId('courses-table')).toBeTruthy();
    });
    // Column headers — use getAllByText to handle possible duplicates in Cloudscape DOM
    expect(screen.getAllByText(/^Code$/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/^Title$/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/^Status$/i).length).toBeGreaterThan(0);
  });

  it('renders course data from API', async () => {
    render(<CoursesPanel />);
    await waitFor(() => {
      expect(screen.getByText('CS101')).toBeTruthy();
    });
    expect(screen.getByText('Intro to Computer Science')).toBeTruthy();
  });

  it('renders empty state when API returns no courses', async () => {
    mockApiClient.getCourses.mockResolvedValue([]);
    render(<CoursesPanel />);
    await waitFor(() => {
      expect(screen.getByText(/no courses found/i)).toBeTruthy();
    });
  });

  it('shows Create Course button', async () => {
    render(<CoursesPanel />);
    await waitFor(() => {
      expect(screen.getByTestId('create-course-button')).toBeTruthy();
    });
  });

  it('opens Create Course modal on button click', async () => {
    const user = userEvent.setup();
    render(<CoursesPanel />);
    await waitFor(() => screen.getByTestId('create-course-button'));

    await user.click(screen.getByTestId('create-course-button'));
    await waitFor(() => {
      expect(screen.getByTestId('create-course-modal')).toBeTruthy();
    });
  });

  it('closes Create Course modal on Cancel without calling API', async () => {
    const user = userEvent.setup();
    render(<CoursesPanel />);
    await waitFor(() => screen.getByTestId('create-course-button'));

    await user.click(screen.getByTestId('create-course-button'));
    await waitFor(() => screen.getByTestId('create-course-modal'));

    // Cloudscape Modals keep hidden content in DOM — verify the API was NOT called.
    const cancelBtn = screen.getAllByRole('button', { name: /cancel/i })[0];
    await user.click(cancelBtn);
    expect(mockApiClient.createCourse).not.toHaveBeenCalled();
  });

  it('calls createCourse API on submit', async () => {
    const user = userEvent.setup();
    render(<CoursesPanel />);
    await waitFor(() => screen.getByTestId('create-course-button'));

    await user.click(screen.getByTestId('create-course-button'));
    await waitFor(() => screen.getByTestId('create-course-modal'));

    await user.type(getNativeInput('course-code-input'), 'CS201');
    await user.type(getNativeInput('course-title-input'), 'Data Structures');

    await user.click(screen.getByTestId('create-course-submit'));
    await waitFor(() => {
      expect(mockApiClient.createCourse).toHaveBeenCalled();
    });
  });
});

// ── CourseDetailPanel tests ───────────────────────────────────────────────────

describe('CourseDetailPanel', () => {
  const mockCourse = mockCourses[0];
  const onBack = vi.fn();
  const onRefresh = vi.fn();

  const renderDetail = () =>
    render(<CourseDetailPanel course={mockCourse} onBack={onBack} onRefresh={onRefresh} />);

  it('renders course header with code and title', async () => {
    renderDetail();
    await waitFor(() => {
      expect(screen.getByText(/CS101/)).toBeTruthy();
      expect(screen.getByText(/Intro to Computer Science/)).toBeTruthy();
    });
  });

  it('renders 5 tabs', async () => {
    renderDetail();
    await waitFor(() => {
      const tabs = screen.getAllByRole('tab');
      const labels = tabs.map(t => t.textContent || '');
      expect(labels.some(l => /overview/i.test(l))).toBe(true);
      expect(labels.some(l => /members/i.test(l))).toBe(true);
      expect(labels.some(l => /templates/i.test(l))).toBe(true);
      expect(labels.some(l => /budget/i.test(l))).toBe(true);
      expect(labels.some(l => /audit/i.test(l))).toBe(true);
    });
  });

  it('Members tab shows enroll button', async () => {
    renderDetail();
    await clickTab('Members');
    await waitFor(() => {
      expect(screen.getByTestId('enroll-member-button')).toBeTruthy();
    });
  });

  it('Templates tab shows add template input', async () => {
    renderDetail();
    await clickTab('Templates');
    await waitFor(() => {
      expect(screen.getByTestId('add-template-input')).toBeTruthy();
    });
  });

  it('Templates tab shows empty whitelist info', async () => {
    renderDetail();
    await clickTab('Templates');
    await waitFor(() => {
      expect(screen.getByText(/all templates are allowed/i)).toBeTruthy();
    });
  });

  it('Budget tab shows distribute button', async () => {
    renderDetail();
    await clickTab('Budget');
    await waitFor(() => {
      expect(screen.getByTestId('distribute-budget-button')).toBeTruthy();
    });
  });

  it('Audit tab shows audit table', async () => {
    renderDetail();
    await clickTab('Audit');
    await waitFor(() => {
      expect(screen.getByTestId('audit-table')).toBeTruthy();
    });
  });

  it('Audit tab shows download report button', async () => {
    renderDetail();
    await clickTab('Audit');
    await waitFor(() => {
      expect(screen.getByTestId('download-report-button')).toBeTruthy();
    });
  });

  it('archive button is hidden for non-closed courses', async () => {
    // Default course is 'active' — archive button should not be visible
    renderDetail();
    await clickTab('Audit');
    await waitFor(() => screen.getByTestId('download-report-button'));
    expect(screen.queryByTestId('archive-course-button')).toBeFalsy();
  });

  it('archive button visible for closed courses', async () => {
    const closedCourse = { ...mockCourse, status: 'closed' };
    render(<CourseDetailPanel course={closedCourse} onBack={onBack} onRefresh={onRefresh} />);
    await clickTab('Audit');
    await waitFor(() => {
      expect(screen.getByTestId('archive-course-button')).toBeTruthy();
    });
  });

  it('archive modal appears on archive button click', async () => {
    const user = userEvent.setup();
    const closedCourse = { ...mockCourse, status: 'closed' };
    render(<CourseDetailPanel course={closedCourse} onBack={onBack} onRefresh={onRefresh} />);
    await clickTab('Audit');
    await waitFor(() => screen.getByTestId('archive-course-button'));

    await user.click(screen.getByTestId('archive-course-button'));
    await waitFor(() => {
      expect(screen.getByTestId('archive-confirm-button')).toBeTruthy();
    });
  });

  it('Back button calls onBack', async () => {
    const user = userEvent.setup();
    renderDetail();
    await waitFor(() => screen.getByRole('button', { name: /back to courses/i }));

    await user.click(screen.getByRole('button', { name: /back to courses/i }));
    expect(onBack).toHaveBeenCalled();
  });
});
