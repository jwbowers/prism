/**
 * Confirm Dialog Page Object
 *
 * Page object for confirmation dialogs in Prism GUI.
 * Handles delete confirmations, action confirmations, etc.
 */

import { Page, Locator } from '@playwright/test';

export class ConfirmDialog {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Get dialog container
   */
  getDialog(): Locator {
    return this.page.locator('[role="dialog"]');
  }

  /**
   * Wait for confirmation dialog to open
   */
  async waitForDialog() {
    await this.getDialog().waitFor({ state: 'visible', timeout: 5000 });
  }

  /**
   * Get dialog message
   */
  async getMessage(): Promise<string | null> {
    const messageText = this.getDialog().locator('.awsui-modal-body, .modal-body');
    return await messageText.textContent();
  }

  /**
   * Click Confirm button
   */
  async clickConfirm() {
    const confirmButton = this.page.getByRole('button', { name: /confirm|yes|delete/i });
    await confirmButton.click();
  }

  /**
   * Click Cancel button
   */
  async clickCancel() {
    const cancelButton = this.page.getByRole('button', { name: /cancel|no/i });
    await cancelButton.click();
  }

  /**
   * Confirm action (wait for dialog and click confirm)
   */
  async confirm() {
    await this.waitForDialog();
    await this.clickConfirm();
  }

  /**
   * Cancel action (wait for dialog and click cancel)
   */
  async cancel() {
    await this.waitForDialog();
    await this.clickCancel();
  }

  /**
   * Verify dialog contains specific text
   */
  async containsText(text: string): Promise<boolean> {
    const message = await this.getMessage();
    return message !== null && message.includes(text);
  }

  /**
   * Verify dialog is open
   */
  async isOpen(): Promise<boolean> {
    return await this.getDialog().isVisible();
  }

  /**
   * Verify dialog is closed
   */
  async isClosed(): Promise<boolean> {
    return !(await this.getDialog().isVisible());
  }

  /**
   * Confirm delete action
   */
  async confirmDelete() {
    await this.waitForDialog();
    const deleteButton = this.page.getByRole('button', { name: /delete/i });
    await deleteButton.click();
  }

  /**
   * Confirm terminate action (for instances)
   */
  async confirmTerminate() {
    await this.waitForDialog();
    const terminateButton = this.page.getByRole('button', { name: /terminate/i });
    await terminateButton.click();
  }

  /**
   * Confirm hibernate action (for instances)
   */
  async confirmHibernate() {
    await this.waitForDialog();
    const hibernateButton = this.page.getByRole('button', { name: /hibernate/i });
    await hibernateButton.click();
  }

  /**
   * Verify educational message is shown (for hibernation)
   */
  async hasEducationalMessage(): Promise<boolean> {
    const message = await this.getMessage();
    if (!message) return false;

    // Check for educational keywords
    const educationalKeywords = [
      'save',
      'faster',
      'preserves',
      'cost',
      'state',
      'resume',
      'instant',
    ];

    return educationalKeywords.some((keyword) =>
      message.toLowerCase().includes(keyword)
    );
  }

  /**
   * Verify cost savings information is shown
   */
  async hasCostSavings(): Promise<boolean> {
    const message = await this.getMessage();
    return message !== null && /\$[\d.]+/.test(message);
  }
}
