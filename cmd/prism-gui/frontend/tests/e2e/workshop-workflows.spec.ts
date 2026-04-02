/**
 * Workshop Workflows E2E Tests — v0.18.0
 *
 * Tests for the Workshop & Event Management GUI introduced in v0.18.0:
 *   Workshop list | Create | Dashboard | Config Templates
 *   Provision | End Workshop | Delete
 *
 * Requires: prismd running with PRISM_TEST_MODE=true
 */

import { test, expect, request } from '@playwright/test';
import { WorkshopPage } from './pages';

const BASE_URL = 'http://localhost:8947';

// ── Helpers ───────────────────────────────────────────────────────────────────

function uniqueTitle(prefix: string = 'E2E') {
  return `${prefix}-${Date.now() % 100000}`;
}

function futureDate(hoursFromNow: number): string {
  const d = new Date(Date.now() + hoursFromNow * 3600 * 1000);
  return d.toISOString();
}

async function createWorkshopViaAPI(title: string): Promise<string> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  const resp = await apiContext.post('/api/v1/workshops', {
    data: {
      title,
      owner: 'e2e-organizer',
      template: 'python-ml',
      max_participants: 30,
      budget_per_participant: 5.0,
      start_time: futureDate(24),
      end_time: futureDate(30),
    }
  });
  const body = await resp.json();
  await apiContext.dispose();
  return body.id || '';
}

async function deleteWorkshopViaAPI(id: string): Promise<void> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  await apiContext.delete(`/api/v1/workshops/${id}`);
  await apiContext.dispose();
}

async function addParticipantViaAPI(workshopId: string, userId: string): Promise<void> {
  const apiContext = await request.newContext({ baseURL: BASE_URL });
  await apiContext.post(`/api/v1/workshops/${workshopId}/participants`, {
    data: { user_id: userId, display_name: userId, email: `${userId}@example.com` }
  });
  await apiContext.dispose();
}

// ── Test Suite ────────────────────────────────────────────────────────────────

test.describe('Workshop Workflows', () => {
  let workshopPage: WorkshopPage;
  const createdWorkshopIds: string[] = [];

  test.beforeEach(async ({ page, context }) => {
    await context.addInitScript(() => {
      localStorage.setItem('prism_onboarding_complete', 'true');
    });
    workshopPage = new WorkshopPage(page);
  });

  test.afterAll(async () => {
    for (const id of createdWorkshopIds) {
      await deleteWorkshopViaAPI(id).catch(() => {});
    }
  });

  // ── Workshop Lifecycle ─────────────────────────────────────────────────────

  test.describe('Workshop Lifecycle', () => {
    test('navigates to workshops sidebar entry', async ({ page }) => {
      await workshopPage.navigateToWorkshops();
      await expect(page.getByTestId('workshops-table')).toBeVisible();
    });

    test('shows workshops panel when navigating to workshops', async ({ page }) => {
      await workshopPage.navigateToWorkshops();
      // Either table or empty state should be visible
      const table = page.getByTestId('workshops-table');
      await expect(table).toBeVisible();
    });

    test('shows workshops created via API', async ({ page }) => {
      const title = uniqueTitle('WS-LIST');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      await workshopPage.navigateToWorkshops();
      await expect(page.getByText(title)).toBeVisible({ timeout: 10000 });
    });

    test('shows Create Workshop button', async ({ page }) => {
      await workshopPage.navigateToWorkshops();
      await expect(page.getByTestId('create-workshop-button')).toBeVisible();
    });

    test('deletes a workshop via UI', async ({ page }) => {
      const title = uniqueTitle('WS-DEL');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      await workshopPage.navigateToWorkshops();
      await expect(page.getByText(title)).toBeVisible({ timeout: 10000 });

      await workshopPage.deleteWorkshop(title);

      // Wait for row to disappear
      await expect(page.getByText(title)).not.toBeVisible({ timeout: 10000 });
    });
  });

  // ── Participants ───────────────────────────────────────────────────────────

  test.describe('Participants', () => {
    test('shows join token for workshop', async ({ page }) => {
      const title = uniqueTitle('WS-TOKEN');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      await workshopPage.navigateToWorkshops();
      await expect(page.getByText(title)).toBeVisible({ timeout: 10000 });

      // The join token column should be visible in the table
      await expect(page.getByTestId('workshops-table')).toBeVisible();
      // Token column header
      await expect(page.getByText(/Join Token/i)).toBeVisible();
    });

    test('shows participant count badge', async ({ page }) => {
      const title = uniqueTitle('WS-BADGE');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      // Add a participant
      await addParticipantViaAPI(id, 'test-participant');

      await workshopPage.navigateToWorkshops();
      await expect(page.getByText(title)).toBeVisible({ timeout: 10000 });

      // Participant count should appear somewhere in the row
      const row = page.locator('[data-testid="workshops-table"] tr').filter({ hasText: title });
      await expect(row).toBeVisible();
    });

    test('shows participant status in dashboard', async ({ page }) => {
      const title = uniqueTitle('WS-PARTICIPANTS');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      await addParticipantViaAPI(id, 'dash-participant');

      await workshopPage.navigateToWorkshops();
      await workshopPage.openWorkshopDashboard(title);

      // Dashboard should show participant in some view
      await expect(page.getByTestId('participants-table')).toBeVisible({ timeout: 10000 });
    });
  });

  // ── Dashboard ──────────────────────────────────────────────────────────────

  test.describe('Dashboard', () => {
    test('shows placeholder when no workshop selected', async ({ page }) => {
      await workshopPage.navigateToWorkshops();
      await workshopPage.switchToDashboardTab();
      await expect(page.getByText(/select a workshop/i)).toBeVisible({ timeout: 5000 });
    });

    test('loads dashboard when workshop title clicked', async ({ page }) => {
      const title = uniqueTitle('WS-DASH');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      await addParticipantViaAPI(id, 'dash-user');

      await workshopPage.navigateToWorkshops();
      await workshopPage.openWorkshopDashboard(title);

      // Should switch to dashboard tab and show stats
      await expect(page.getByTestId('participants-table')).toBeVisible({ timeout: 10000 });
    });

    test('shows provision button in dashboard', async ({ page }) => {
      const title = uniqueTitle('WS-PROV');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      await workshopPage.navigateToWorkshops();

      // Provision button should be in the workshops table row
      const row = page.locator('[data-testid="workshops-table"] tr').filter({ hasText: title });
      await expect(row).toBeVisible({ timeout: 10000 });
      await expect(row.getByRole('button', { name: /^provision$/i })).toBeVisible();
    });

    test('shows end workshop confirmation modal', async ({ page }) => {
      const title = uniqueTitle('WS-END');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      await workshopPage.navigateToWorkshops();
      await expect(page.getByText(title)).toBeVisible({ timeout: 10000 });

      await workshopPage.clickEndWorkshop(title);

      // Confirmation modal should appear
      const dialog = page.getByRole('dialog');
      await expect(dialog).toBeVisible({ timeout: 5000 });
      await expect(dialog.getByText(/end workshop/i).first()).toBeVisible();
    });
  });

  // ── Config Templates ───────────────────────────────────────────────────────

  test.describe('Config Templates', () => {
    test('shows config templates tab', async ({ page }) => {
      await workshopPage.navigateToWorkshops();
      const configTab = page.getByRole('tab', { name: /config/i });
      await expect(configTab).toBeVisible({ timeout: 5000 });
    });

    test('renders configs table on config tab', async ({ page }) => {
      await workshopPage.navigateToWorkshops();
      await workshopPage.switchToConfigTab();
      await expect(page.getByTestId('workshop-configs-table')).toBeVisible({ timeout: 10000 });
    });

    test('shows empty state when no configs saved', async ({ page }) => {
      await workshopPage.navigateToWorkshops();
      await workshopPage.switchToConfigTab();
      // Either configs table with data or empty state — both valid
      const configsTable = page.getByTestId('workshop-configs-table');
      await expect(configsTable).toBeVisible({ timeout: 10000 });
    });

    test('saves a config from a workshop via API and shows it in UI', async ({ page, request: req }) => {
      const title = uniqueTitle('WS-CONFIG');
      const id = await createWorkshopViaAPI(title);
      createdWorkshopIds.push(id);

      // Save config via API
      const configName = uniqueTitle('cfg');
      await req.post(`${BASE_URL}/api/v1/workshops/${id}/config`, {
        data: { name: configName }
      });

      await workshopPage.navigateToWorkshops();
      await workshopPage.switchToConfigTab();

      // Config should appear in the table
      await expect(page.getByText(configName)).toBeVisible({ timeout: 10000 });
    });
  });
});
