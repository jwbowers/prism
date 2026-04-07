/**
 * Connection Dialog Page Object
 *
 * Page object for the Connection Info dialog in Prism GUI.
 * Displays SSH connection information, web URLs, and connection instructions.
 */

import { Page, Locator } from '@playwright/test';

export class ConnectionDialog {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Get dialog container
   */
  getDialog(): Locator {
    return this.page.getByRole('dialog', { name: 'Connection Information' });
  }

  /**
   * Wait for connection dialog to open
   */
  async waitForDialog() {
    await this.getDialog().waitFor({ state: 'visible', timeout: 5000 });
  }

  /**
   * Get SSH command
   */
  async getSSHCommand(): Promise<string | null> {
    const sshCommand = this.page.locator('[data-testid="ssh-command"], code:has-text("ssh")');
    return await sshCommand.textContent();
  }

  /**
   * Get public IP address
   */
  async getPublicIP(): Promise<string | null> {
    const ipText = this.page.locator('[data-testid="public-ip"]');
    return await ipText.textContent();
  }

  /**
   * Get web URL (for Jupyter, RStudio, etc.)
   */
  async getWebURL(): Promise<string | null> {
    const urlText = this.page.locator('[data-testid="web-url"]');
    return await urlText.textContent();
  }

  /**
   * Copy SSH command to clipboard
   */
  async copySshCommand() {
    const copyButton = this.page.getByRole('button', { name: /copy.*ssh/i });
    await copyButton.click();
  }

  /**
   * Copy web URL to clipboard
   */
  async copyWebURL() {
    const copyButton = this.page.getByRole('button', { name: /copy.*url/i });
    await copyButton.click();
  }

  /**
   * Open web URL in browser
   */
  async openWebURL() {
    const openButton = this.page.getByRole('button', { name: /open.*browser/i });
    await openButton.click();
  }

  /**
   * Close dialog
   * Uses .last() because Cloudscape Modal renders both a header close icon button
   * and a footer "Close" text button — both match /close/i.
   */
  async close() {
    const closeButton = this.page.getByRole('button', { name: /close/i }).last();
    await closeButton.click();
  }

  /**
   * Verify SSH connection info is displayed
   */
  async hasSSHInfo(): Promise<boolean> {
    const sshCommand = await this.getSSHCommand();
    return sshCommand !== null && sshCommand.includes('ssh');
  }

  /**
   * Verify web URL is displayed
   */
  async hasWebURL(): Promise<boolean> {
    const webURL = await this.getWebURL();
    return webURL !== null && (webURL.includes('http') || webURL.includes('https'));
  }

  /**
   * Verify public IP is displayed
   */
  async hasPublicIP(): Promise<boolean> {
    const publicIP = await this.getPublicIP();
    return publicIP !== null && /\d+\.\d+\.\d+\.\d+/.test(publicIP);
  }

  /**
   * Get connection instructions
   */
  async getInstructions(): Promise<string | null> {
    const instructions = this.page.locator('[data-testid="connection-instructions"]');
    return await instructions.textContent();
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
   * Get all displayed ports
   */
  async getPorts(): Promise<string[] | null> {
    const portsSection = this.page.locator('[data-testid="ports-list"]');
    if (await portsSection.isVisible()) {
      const portsText = await portsSection.textContent();
      if (portsText) {
        // Extract port numbers from text
        const portMatches = portsText.match(/\d{2,5}/g);
        return portMatches || [];
      }
    }
    return null;
  }

  /**
   * Verify connection type (SSH, Web, Both)
   */
  async getConnectionType(): Promise<'ssh' | 'web' | 'both' | 'unknown'> {
    const hasSSH = await this.hasSSHInfo();
    const hasWeb = await this.hasWebURL();

    if (hasSSH && hasWeb) return 'both';
    if (hasSSH) return 'ssh';
    if (hasWeb) return 'web';
    return 'unknown';
  }
}
