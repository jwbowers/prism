/**
 * Custom Assertions for Prism GUI Tests
 *
 * Domain-specific assertions that make tests more readable and maintainable.
 */

import { expect, Page } from '@playwright/test';
import type { Template, Instance } from '../../src/types';

/**
 * Assert template card is displayed correctly
 */
export async function assertTemplateCard(
  page: Page,
  template: Partial<Template>
) {
  if (template.Name) {
    await expect(page.locator(`text=${template.Name}`)).toBeVisible();
  }
  if (template.Description) {
    await expect(page.locator(`text=${template.Description}`)).toBeVisible();
  }
  if (template.Icon) {
    await expect(page.locator(`text=${template.Icon}`)).toBeVisible();
  }
  if (template.Category) {
    await expect(page.locator(`text=${template.Category}`)).toBeVisible();
  }
}

/**
 * Assert instance row is displayed correctly
 */
export async function assertInstanceRow(
  page: Page,
  instance: Partial<Instance>
) {
  if (instance.name) {
    await expect(page.locator(`text=${instance.name}`)).toBeVisible();
  }
  if (instance.id) {
    await expect(page.locator(`text=${instance.id}`)).toBeVisible();
  }
  if (instance.status) {
    const statusText = instance.status.charAt(0).toUpperCase() + instance.status.slice(1);
    await expect(page.locator(`text=${statusText}`)).toBeVisible();
  }
  if (instance.public_ip) {
    await expect(page.locator(`text=${instance.public_ip}`)).toBeVisible();
  }
}

/**
 * Assert notification is displayed
 */
export async function assertNotification(
  page: Page,
  message: string,
  type: 'success' | 'error' | 'info' | 'warning' = 'info'
) {
  const notification = page.locator(`[data-testid="notification-${type}"]`);
  await expect(notification).toBeVisible();
  await expect(notification).toContainText(message);
}

/**
 * Assert modal dialog is open
 */
export async function assertModalOpen(
  page: Page,
  title?: string
) {
  const modal = page.locator('[role="dialog"], [data-testid="modal"]');
  await expect(modal).toBeVisible();

  if (title) {
    await expect(modal.locator(`text=${title}`)).toBeVisible();
  }
}

/**
 * Assert modal dialog is closed
 */
export async function assertModalClosed(page: Page) {
  const modal = page.locator('[role="dialog"], [data-testid="modal"]');
  await expect(modal).not.toBeVisible();
}

/**
 * Assert form field has error
 */
export async function assertFieldError(
  page: Page,
  fieldLabel: string,
  errorMessage?: string
) {
  const field = page.locator(`label:has-text("${fieldLabel}")`).locator('..');
  await expect(field.locator('[role="alert"], .error-message')).toBeVisible();

  if (errorMessage) {
    await expect(field.locator(`text=${errorMessage}`)).toBeVisible();
  }
}

/**
 * Assert button is enabled/disabled
 */
export async function assertButtonState(
  page: Page,
  buttonText: string,
  enabled: boolean
) {
  const button = page.locator(`button:has-text("${buttonText}")`);
  if (enabled) {
    await expect(button).toBeEnabled();
  } else {
    await expect(button).toBeDisabled();
  }
}

/**
 * Assert table has expected number of rows
 */
export async function assertTableRowCount(
  page: Page,
  expectedCount: number,
  tableSelector: string = 'table'
) {
  const rows = page.locator(`${tableSelector} tbody tr`);
  await expect(rows).toHaveCount(expectedCount);
}

/**
 * Assert element has specific CSS class
 */
export async function assertHasClass(
  page: Page,
  selector: string,
  className: string
) {
  const element = page.locator(selector);
  await expect(element).toHaveClass(new RegExp(className));
}

/**
 * Assert cost is displayed correctly
 */
export async function assertCostDisplay(
  page: Page,
  costPerHour: number,
  size?: string
) {
  const costText = `$${costPerHour.toFixed(2)}/hour`;
  await expect(page.locator(`text=${costText}`)).toBeVisible();

  if (size) {
    await expect(page.locator(`text=${size}`)).toBeVisible();
  }
}

/**
 * Assert API error is displayed
 */
export async function assertApiError(
  page: Page,
  errorMessage: string
) {
  await expect(page.locator('[data-testid="error-message"], .error-alert')).toBeVisible();
  await expect(page.locator(`text=${errorMessage}`)).toBeVisible();
}

/**
 * Assert loading state
 */
export async function assertLoading(page: Page, isLoading: boolean) {
  const spinner = page.locator('[data-testid="loading-spinner"], .loading-spinner');
  if (isLoading) {
    await expect(spinner).toBeVisible();
  } else {
    await expect(spinner).not.toBeVisible();
  }
}

/**
 * Assert empty state is displayed
 */
export async function assertEmptyState(
  page: Page,
  message?: string
) {
  const emptyState = page.locator('[data-testid="empty-state"], .empty-state');
  await expect(emptyState).toBeVisible();

  if (message) {
    await expect(emptyState).toContainText(message);
  }
}

/**
 * Assert instance status badge
 */
export async function assertInstanceStatus(
  page: Page,
  status: 'running' | 'stopped' | 'hibernating' | 'launching' | 'terminating'
) {
  const statusBadge = page.locator(`[data-status="${status}"]`);
  await expect(statusBadge).toBeVisible();
}

/**
 * Assert connection info is displayed
 */
export async function assertConnectionInfo(
  page: Page,
  connectionInfo: {
    ssh?: string;
    webUrl?: string;
    port?: number;
  }
) {
  if (connectionInfo.ssh) {
    await expect(page.locator(`text=${connectionInfo.ssh}`)).toBeVisible();
  }
  if (connectionInfo.webUrl) {
    await expect(page.locator(`a[href="${connectionInfo.webUrl}"]`)).toBeVisible();
  }
  if (connectionInfo.port) {
    await expect(page.locator(`text=${connectionInfo.port}`)).toBeVisible();
  }
}

/**
 * Assert profile is active/selected
 */
export async function assertActiveProfile(
  page: Page,
  profileName: string
) {
  const profileRow = page.locator(`[data-profile="${profileName}"]`);
  await expect(profileRow).toHaveClass(/active|selected/);
}

/**
 * Assert storage volume details
 */
export async function assertStorageVolume(
  page: Page,
  volume: {
    name?: string;
    size?: number;
    state?: string;
    type?: string;
  }
) {
  if (volume.name) {
    await expect(page.locator(`text=${volume.name}`)).toBeVisible();
  }
  if (volume.size) {
    await expect(page.locator(`text=${volume.size} GB`)).toBeVisible();
  }
  if (volume.state) {
    await expect(page.locator(`text=${volume.state}`)).toBeVisible();
  }
  if (volume.type) {
    await expect(page.locator(`text=${volume.type}`)).toBeVisible();
  }
}

/**
 * Assert hibernation capability
 */
export async function assertHibernationCapability(
  page: Page,
  canHibernate: boolean
) {
  const hibernateButton = page.locator('button:has-text("Hibernate")');
  if (canHibernate) {
    await expect(hibernateButton).toBeEnabled();
  } else {
    await expect(hibernateButton).toBeDisabled();
  }
}

/**
 * Assert idle policy details
 */
export async function assertIdlePolicy(
  page: Page,
  policy: {
    name?: string;
    idleMinutes?: number;
    action?: string;
  }
) {
  if (policy.name) {
    await expect(page.locator(`text=${policy.name}`)).toBeVisible();
  }
  if (policy.idleMinutes) {
    await expect(page.locator(`text=${policy.idleMinutes} min`)).toBeVisible();
  }
  if (policy.action) {
    await expect(page.locator(`text=${policy.action}`)).toBeVisible();
  }
}
