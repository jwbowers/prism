/**
 * GovernancePanel Unit Tests — v0.13.0
 *
 * Tests the GovernancePanel component using window.__apiClient mock.
 * GovernancePanel accesses the API via (window as any).__apiClient
 * (set by App.tsx) rather than importing SafePrismAPI directly.
 *
 * Coverage:
 * - All 5 tabs render without crashing
 * - Quotas: table renders, Set Quota modal opens/submits/cancels
 * - Grant Period: no-period state, configure modal, delete confirm
 * - Budget Sharing: table renders, Share Budget modal
 * - Onboarding Templates: table renders, Add Template modal
 * - Monthly Report: generate button calls API, output renders
 * - Error states: alert shown when API call fails
 *
 * Key patterns:
 * - Cloudscape <Input data-testid="x"> puts testid on wrapper <div>,
 *   not the native <input>. Use getNativeInput(testId) helper.
 * - Cloudscape <Textarea data-testid="x"> same — use getNativeTextarea(testId).
 * - Each test owns its own render() call; no shared beforeEach renders.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { GovernancePanel } from './components/GovernancePanel';
import { ApiContext } from './hooks/use-api';

// ── Mock API client ──────────────────────────────────────────────────────────

const mockApiClient = {
  getProjectQuotas: vi.fn(),
  setProjectQuota: vi.fn(),
  deleteProjectQuota: vi.fn(),
  getGrantPeriod: vi.fn(),
  setGrantPeriod: vi.fn(),
  deleteGrantPeriod: vi.fn(),
  listProjectBudgetShares: vi.fn(),
  shareProjectBudget: vi.fn(),
  listOnboardingTemplates: vi.fn(),
  addOnboardingTemplate: vi.fn(),
  deleteOnboardingTemplate: vi.fn(),
  getMonthlyReport: vi.fn(),
};

const PROJECT_ID = 'test-project-123';

// ── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Cloudscape <Input data-testid="x"> renders testid on a wrapper <div>.
 * This helper finds the actual <input> inside that wrapper.
 */
const getNativeInput = (testId: string): HTMLInputElement => {
  const wrapper = screen.getByTestId(testId);
  const native = wrapper.querySelector('input');
  if (!native) throw new Error(`No <input> inside [data-testid="${testId}"]`);
  return native as HTMLInputElement;
};

/**
 * Cloudscape <Textarea data-testid="x"> same pattern.
 */
const getNativeTextarea = (testId: string): HTMLTextAreaElement => {
  const wrapper = screen.getByTestId(testId);
  const native = wrapper.querySelector('textarea');
  if (!native) throw new Error(`No <textarea> inside [data-testid="${testId}"]`);
  return native as HTMLTextAreaElement;
};

/** Click a governance sub-tab by partial label text. */
const clickTab = async (label: string) => {
  const user = userEvent.setup();
  const tab = screen.getAllByRole('tab').find(t => t.textContent?.includes(label));
  if (!tab) throw new Error(`Tab "${label}" not found`);
  await user.click(tab);
};

const renderPanel = () => render(
  <ApiContext.Provider value={mockApiClient as any}>
    <GovernancePanel projectId={PROJECT_ID} />
  </ApiContext.Provider>
);

// ── Setup ────────────────────────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks();

  (window as any).__apiClient = mockApiClient;

  // Default empty/null responses
  mockApiClient.getProjectQuotas.mockResolvedValue([]);
  mockApiClient.getGrantPeriod.mockResolvedValue(null);
  mockApiClient.listProjectBudgetShares.mockResolvedValue([]);
  mockApiClient.listOnboardingTemplates.mockResolvedValue([]);
  mockApiClient.getMonthlyReport.mockResolvedValue('');
  mockApiClient.setProjectQuota.mockResolvedValue(undefined);
  mockApiClient.deleteProjectQuota.mockResolvedValue(undefined);
  mockApiClient.setGrantPeriod.mockResolvedValue(undefined);
  mockApiClient.deleteGrantPeriod.mockResolvedValue(undefined);
  mockApiClient.shareProjectBudget.mockResolvedValue(undefined);
  mockApiClient.addOnboardingTemplate.mockResolvedValue(undefined);
  mockApiClient.deleteOnboardingTemplate.mockResolvedValue(undefined);
});

// ── Tests ────────────────────────────────────────────────────────────────────

describe('GovernancePanel', () => {

  // ── Initial Render ─────────────────────────────────────────────────────────

  describe('Initial Render', () => {
    it('renders the governance panel', async () => {
      renderPanel();
      await waitFor(() => {
        expect(screen.getByTestId('governance-panel')).toBeInTheDocument();
      });
    });

    it('renders all 5 tab labels', async () => {
      renderPanel();
      await waitFor(() => {
        expect(screen.getByRole('tab', { name: /quotas/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /grant period/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /budget sharing/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /onboarding templates/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /monthly report/i })).toBeInTheDocument();
      });
    });

    it('calls all four load APIs on mount', async () => {
      renderPanel();
      await waitFor(() => {
        expect(mockApiClient.getProjectQuotas).toHaveBeenCalledWith(PROJECT_ID);
        expect(mockApiClient.getGrantPeriod).toHaveBeenCalledWith(PROJECT_ID);
        expect(mockApiClient.listProjectBudgetShares).toHaveBeenCalledWith(PROJECT_ID);
        expect(mockApiClient.listOnboardingTemplates).toHaveBeenCalledWith(PROJECT_ID);
      });
    });

    it('reloads data when projectId changes', async () => {
      const { rerender } = renderPanel();
      await waitFor(() => {
        expect(mockApiClient.getProjectQuotas).toHaveBeenCalledWith(PROJECT_ID);
      });
      rerender(
        <ApiContext.Provider value={mockApiClient as any}>
          <GovernancePanel projectId="different-project" />
        </ApiContext.Provider>
      );
      await waitFor(() => {
        expect(mockApiClient.getProjectQuotas).toHaveBeenCalledWith('different-project');
      });
    });
  });

  // ── Quotas Tab ─────────────────────────────────────────────────────────────

  describe('Quotas Tab', () => {
    it('shows empty quotas table when API returns empty list', async () => {
      renderPanel();
      await waitFor(() => {
        expect(screen.getByTestId('quotas-table')).toBeInTheDocument();
      });
    });

    it('renders quota rows returned by API', async () => {
      mockApiClient.getProjectQuotas.mockResolvedValue([
        { role: 'admin', max_instances: 10, max_instance_type: '', max_spend_daily: 0 },
        { role: 'viewer', max_instances: 1, max_instance_type: 't3', max_spend_daily: 5 },
      ]);
      renderPanel();
      // Use getAllByText to handle possible duplicates; confirm both roles appear
      await waitFor(() => {
        expect(screen.getAllByText('admin').length).toBeGreaterThan(0);
        expect(screen.getAllByText('viewer').length).toBeGreaterThan(0);
        expect(screen.getByText('10')).toBeInTheDocument();
        expect(screen.getByText('t3')).toBeInTheDocument();
      });
    });

    it('opens Set Quota modal when button clicked', async () => {
      const user = userEvent.setup();
      renderPanel();

      await waitFor(() => screen.getByTestId('set-quota-button'));
      await user.click(screen.getByTestId('set-quota-button'));

      await waitFor(() => {
        expect(screen.getByTestId('save-quota-button')).toBeInTheDocument();
        expect(screen.getByTestId('quota-role-select')).toBeInTheDocument();
        expect(screen.getByTestId('quota-max-instances-input')).toBeInTheDocument();
        expect(screen.getByTestId('quota-max-spend-daily-input')).toBeInTheDocument();
      });
    });

    it('cancels Set Quota modal without calling API', async () => {
      const user = userEvent.setup();
      renderPanel();

      await waitFor(() => screen.getByTestId('set-quota-button'));
      await user.click(screen.getByTestId('set-quota-button'));

      await waitFor(() => screen.getByTestId('save-quota-button'));
      await user.click(screen.getByRole('button', { name: /cancel/i }));

      expect(mockApiClient.setProjectQuota).not.toHaveBeenCalled();
    });

    it('calls setProjectQuota on save and reloads', async () => {
      const user = userEvent.setup();
      renderPanel();

      await waitFor(() => screen.getByTestId('set-quota-button'));
      await user.click(screen.getByTestId('set-quota-button'));

      // Cloudscape Input puts testid on wrapper; find the native <input> inside
      await waitFor(() => screen.getByTestId('quota-max-instances-input'));
      const instancesInput = getNativeInput('quota-max-instances-input');
      await user.click(instancesInput);
      await user.type(instancesInput, '5');

      const spendInput = getNativeInput('quota-max-spend-daily-input');
      await user.click(spendInput);
      await user.type(spendInput, '20');

      await user.click(screen.getByTestId('save-quota-button'));

      await waitFor(() => {
        expect(mockApiClient.setProjectQuota).toHaveBeenCalledWith(
          PROJECT_ID,
          expect.objectContaining({ max_instances: 5, max_spend_daily: 20 })
        );
        expect(mockApiClient.getProjectQuotas).toHaveBeenCalledTimes(2);
      });
    });

    it('shows error alert when getProjectQuotas fails', async () => {
      mockApiClient.getProjectQuotas.mockRejectedValue(new Error('quota fetch failed'));
      renderPanel();
      await waitFor(() => {
        expect(screen.getByText(/quota fetch failed/i)).toBeInTheDocument();
      });
    });

    it('calls deleteProjectQuota when Delete clicked', async () => {
      mockApiClient.getProjectQuotas.mockResolvedValue([
        { role: 'viewer', max_instances: 1, max_instance_type: '', max_spend_daily: 0 },
      ]);
      const user = userEvent.setup();
      renderPanel();

      await waitFor(() => screen.getByTestId('delete-quota-viewer'));
      await user.click(screen.getByTestId('delete-quota-viewer'));

      await waitFor(() => {
        expect(mockApiClient.deleteProjectQuota).toHaveBeenCalledWith(PROJECT_ID, 'viewer');
      });
    });
  });

  // ── Grant Period Tab ───────────────────────────────────────────────────────
  // Note: each test renders its own panel to avoid double-render pollution.

  describe('Grant Period Tab', () => {
    it('shows no-period alert when API returns null', async () => {
      // mockApiClient.getGrantPeriod already returns null (from beforeEach)
      renderPanel();
      await clickTab('Grant Period');
      await waitFor(() => {
        expect(screen.getByTestId('no-grant-period-alert')).toBeInTheDocument();
      });
    });

    it('shows Configure button when no grant period', async () => {
      renderPanel();
      await clickTab('Grant Period');
      await waitFor(() => {
        expect(screen.getByTestId('configure-grant-period-button')).toBeInTheDocument();
      });
    });

    it('opens grant period modal when Configure clicked', async () => {
      const user = userEvent.setup();
      renderPanel();
      await clickTab('Grant Period');

      await waitFor(() => screen.getByTestId('configure-grant-period-button'));
      await user.click(screen.getByTestId('configure-grant-period-button'));

      await waitFor(() => {
        expect(screen.getByTestId('grant-period-name-input')).toBeInTheDocument();
        expect(screen.getByTestId('grant-period-start-input')).toBeInTheDocument();
        expect(screen.getByTestId('grant-period-end-input')).toBeInTheDocument();
        expect(screen.getByTestId('save-grant-period-button')).toBeInTheDocument();
      });
    });

    it('calls setGrantPeriod with correct data on save', async () => {
      const user = userEvent.setup();
      renderPanel();
      await clickTab('Grant Period');

      await waitFor(() => screen.getByTestId('configure-grant-period-button'));
      await user.click(screen.getByTestId('configure-grant-period-button'));

      await waitFor(() => screen.getByTestId('grant-period-name-input'));
      // Use native input inside Cloudscape Input wrapper
      const nameInput = getNativeInput('grant-period-name-input');
      await user.click(nameInput);
      await user.type(nameInput, 'NSF Year 1');

      const startInput = getNativeInput('grant-period-start-input');
      await user.click(startInput);
      await user.type(startInput, '2024-01-01');

      const endInput = getNativeInput('grant-period-end-input');
      await user.click(endInput);
      await user.type(endInput, '2024-12-31');

      await user.click(screen.getByTestId('save-grant-period-button'));

      await waitFor(() => {
        expect(mockApiClient.setGrantPeriod).toHaveBeenCalledWith(
          PROJECT_ID,
          expect.objectContaining({ name: 'NSF Year 1' })
        );
      });
    });

    it('renders grant period details when API returns existing data', async () => {
      mockApiClient.getGrantPeriod.mockResolvedValue({
        name: 'NSF Year 2',
        start_date: '2025-01-01T00:00:00Z',
        end_date: '2025-12-31T00:00:00Z',
        auto_freeze: true,
      });

      renderPanel();
      await clickTab('Grant Period');

      await waitFor(() => {
        expect(screen.getByTestId('grant-period-details')).toBeInTheDocument();
        expect(screen.getByTestId('grant-period-name')).toHaveTextContent('NSF Year 2');
      });
    });

    it('shows Edit and Delete buttons when grant period exists', async () => {
      mockApiClient.getGrantPeriod.mockResolvedValue({
        name: 'Existing Grant',
        start_date: '2025-01-01T00:00:00Z',
        end_date: '2025-12-31T00:00:00Z',
        auto_freeze: false,
      });

      renderPanel();
      await clickTab('Grant Period');

      await waitFor(() => {
        expect(screen.getByTestId('edit-grant-period-button')).toBeInTheDocument();
        expect(screen.getByTestId('delete-grant-period-button')).toBeInTheDocument();
      });
    });

    it('calls deleteGrantPeriod after confirm-delete clicked', async () => {
      mockApiClient.getGrantPeriod.mockResolvedValue({
        name: 'To Delete',
        start_date: '2025-01-01T00:00:00Z',
        end_date: '2025-12-31T00:00:00Z',
        auto_freeze: false,
      });

      const user = userEvent.setup();
      renderPanel();
      await clickTab('Grant Period');

      await waitFor(() => screen.getByTestId('delete-grant-period-button'));
      await user.click(screen.getByTestId('delete-grant-period-button'));

      await waitFor(() => screen.getByTestId('confirm-delete-grant-period-button'));
      await user.click(screen.getByTestId('confirm-delete-grant-period-button'));

      await waitFor(() => {
        expect(mockApiClient.deleteGrantPeriod).toHaveBeenCalledWith(PROJECT_ID);
      });
    });
  });

  // ── Budget Sharing Tab ─────────────────────────────────────────────────────

  describe('Budget Sharing Tab', () => {
    it('shows empty budget shares table by default', async () => {
      renderPanel();
      await clickTab('Budget Sharing');
      await waitFor(() => {
        expect(screen.getByTestId('budget-shares-table')).toBeInTheDocument();
      });
    });

    it('renders share rows when API returns data', async () => {
      mockApiClient.listProjectBudgetShares.mockResolvedValue([
        {
          id: 'share-1',
          request: { from_project_id: PROJECT_ID, to_member_id: 'alice', amount: 100, reason: 'allocation' },
          approved_by: 'admin',
          created_at: '2025-01-01T00:00:00Z',
        },
      ]);

      renderPanel();
      await clickTab('Budget Sharing');

      await waitFor(() => {
        expect(screen.getByText('alice')).toBeInTheDocument();
        expect(screen.getByText('$100.00')).toBeInTheDocument();
        expect(screen.getByText('allocation')).toBeInTheDocument();
      });
    });

    it('opens Share Budget modal', async () => {
      const user = userEvent.setup();
      renderPanel();
      await clickTab('Budget Sharing');

      await waitFor(() => screen.getByTestId('share-budget-button'));
      await user.click(screen.getByTestId('share-budget-button'));

      await waitFor(() => {
        expect(screen.getByTestId('confirm-share-budget-button')).toBeInTheDocument();
        expect(screen.getByTestId('share-amount-input')).toBeInTheDocument();
      });
    });

    it('calls shareProjectBudget with correct data on confirm', async () => {
      const user = userEvent.setup();
      renderPanel();
      await clickTab('Budget Sharing');

      await waitFor(() => screen.getByTestId('share-budget-button'));
      await user.click(screen.getByTestId('share-budget-button'));

      await waitFor(() => screen.getByTestId('share-to-member-input'));
      // Use native inputs inside Cloudscape Input wrappers
      const memberInput = getNativeInput('share-to-member-input');
      await user.click(memberInput);
      await user.type(memberInput, 'bob');

      const amountInput = getNativeInput('share-amount-input');
      await user.click(amountInput);
      await user.type(amountInput, '50');

      const reasonInput = getNativeInput('share-reason-input');
      await user.click(reasonInput);
      await user.type(reasonInput, 'extra funds');

      await user.click(screen.getByTestId('confirm-share-budget-button'));

      await waitFor(() => {
        expect(mockApiClient.shareProjectBudget).toHaveBeenCalledWith(
          PROJECT_ID,
          expect.objectContaining({ to_member_id: 'bob', amount: 50, reason: 'extra funds' })
        );
      });
    });
  });

  // ── Onboarding Templates Tab ───────────────────────────────────────────────

  describe('Onboarding Templates Tab', () => {
    it('shows empty table by default', async () => {
      renderPanel();
      await clickTab('Onboarding Templates');
      await waitFor(() => {
        expect(screen.getByTestId('onboarding-templates-table')).toBeInTheDocument();
      });
    });

    it('renders template rows when API returns data', async () => {
      mockApiClient.listOnboardingTemplates.mockResolvedValue([
        { id: 'tmpl-1', name: 'ML Starter', description: 'Python ML environment' },
        { id: 'tmpl-2', name: 'R Basics', description: 'R research setup' },
      ]);

      renderPanel();
      await clickTab('Onboarding Templates');

      await waitFor(() => {
        expect(screen.getByText('ML Starter')).toBeInTheDocument();
        expect(screen.getByText('R Basics')).toBeInTheDocument();
        expect(screen.getByText('Python ML environment')).toBeInTheDocument();
      });
    });

    it('opens Add Template modal', async () => {
      const user = userEvent.setup();
      renderPanel();
      await clickTab('Onboarding Templates');

      await waitFor(() => screen.getByTestId('add-onboarding-template-button'));
      await user.click(screen.getByTestId('add-onboarding-template-button'));

      await waitFor(() => {
        expect(screen.getByTestId('onboarding-template-name-input')).toBeInTheDocument();
        expect(screen.getByTestId('onboarding-template-description-input')).toBeInTheDocument();
        expect(screen.getByTestId('save-onboarding-template-button')).toBeInTheDocument();
      });
    });

    it('calls addOnboardingTemplate with correct data on save', async () => {
      const user = userEvent.setup();
      renderPanel();
      await clickTab('Onboarding Templates');

      await waitFor(() => screen.getByTestId('add-onboarding-template-button'));
      await user.click(screen.getByTestId('add-onboarding-template-button'));

      await waitFor(() => screen.getByTestId('onboarding-template-name-input'));
      // Cloudscape Input → native input; Cloudscape Textarea → native textarea
      const nameInput = getNativeInput('onboarding-template-name-input');
      await user.click(nameInput);
      await user.type(nameInput, 'New Template');

      const descTextarea = getNativeTextarea('onboarding-template-description-input');
      await user.click(descTextarea);
      await user.type(descTextarea, 'A description');

      await user.click(screen.getByTestId('save-onboarding-template-button'));

      await waitFor(() => {
        expect(mockApiClient.addOnboardingTemplate).toHaveBeenCalledWith(
          PROJECT_ID,
          expect.objectContaining({ name: 'New Template', description: 'A description' })
        );
      });
    });

    it('calls deleteOnboardingTemplate when Delete clicked', async () => {
      mockApiClient.listOnboardingTemplates.mockResolvedValue([
        { id: 'tmpl-1', name: 'ML Starter', description: 'desc' },
      ]);

      const user = userEvent.setup();
      renderPanel();
      await clickTab('Onboarding Templates');

      await waitFor(() => screen.getByTestId('delete-onboarding-template-ML Starter'));
      await user.click(screen.getByTestId('delete-onboarding-template-ML Starter'));

      await waitFor(() => {
        expect(mockApiClient.deleteOnboardingTemplate).toHaveBeenCalledWith(PROJECT_ID, 'tmpl-1');
      });
    });
  });

  // ── Monthly Report Tab ─────────────────────────────────────────────────────

  describe('Monthly Report Tab', () => {
    it('renders report controls', async () => {
      renderPanel();
      await clickTab('Monthly Report');
      await waitFor(() => {
        expect(screen.getByTestId('report-month-input')).toBeInTheDocument();
        expect(screen.getByTestId('report-format-select')).toBeInTheDocument();
        expect(screen.getByTestId('generate-report-button')).toBeInTheDocument();
      });
    });

    it('pre-fills current month in the native input', async () => {
      renderPanel();
      await clickTab('Monthly Report');

      const now = new Date();
      const expected = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;

      await waitFor(() => {
        // Cloudscape Input puts testid on wrapper; find native <input> inside
        const input = getNativeInput('report-month-input');
        expect(input.value).toBe(expected);
      });
    });

    it('calls getMonthlyReport and shows output on Generate', async () => {
      mockApiClient.getMonthlyReport.mockResolvedValue('Project: test\nTotal: $0.00');

      const user = userEvent.setup();
      renderPanel();
      await clickTab('Monthly Report');

      await waitFor(() => screen.getByTestId('generate-report-button'));
      await user.click(screen.getByTestId('generate-report-button'));

      await waitFor(() => {
        expect(mockApiClient.getMonthlyReport).toHaveBeenCalledWith(
          PROJECT_ID,
          expect.any(String),
          'text'
        );
        expect(screen.getByTestId('monthly-report-output')).toBeInTheDocument();
        expect(screen.getByTestId('monthly-report-output')).toHaveTextContent('Total: $0.00');
      });
    });

    it('shows error alert when getMonthlyReport fails', async () => {
      mockApiClient.getMonthlyReport.mockRejectedValue(new Error('report generation failed'));

      const user = userEvent.setup();
      renderPanel();
      await clickTab('Monthly Report');

      await waitFor(() => screen.getByTestId('generate-report-button'));
      await user.click(screen.getByTestId('generate-report-button'));

      await waitFor(() => {
        expect(screen.getByText(/report generation failed/i)).toBeInTheDocument();
      });
    });
  });
});
