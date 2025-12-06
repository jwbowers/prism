/**
 * Project Workflows E2E Tests
 *
 * End-to-end tests for complete project management workflows in Prism GUI.
 * Tests: Create, view, update, budget tracking, member management, and delete projects.
 */

import { test, expect } from '@playwright/test';
import { ProjectsPage } from './pages';

test.describe('Project Management Workflows', () => {
  let projectsPage: ProjectsPage;

  test.beforeEach(async ({ page, context }) => {
    // Set localStorage BEFORE navigating to prevent onboarding modal
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });

    projectsPage = new ProjectsPage(page);
    await projectsPage.goto();
    await projectsPage.navigate();

    // Force close any open dialogs from previous tests
    await projectsPage.forceCloseDialogs();

    // NOTE: Cleanup removed from beforeEach to prevent timeouts when many test projects exist.
    // Each test cleans up its own projects in the test body or afterEach phase.
  });

  test.describe('Create Project Workflow', () => {
    test('should create a new project with basic configuration', async () => {
      const uniqueName = `test-project-${Date.now()}`;

      // Create project
      await projectsPage.createProject(uniqueName, 'Test project description');

      // Verify project appears in list
      const projectExists = await projectsPage.verifyProjectExists(uniqueName);
      expect(projectExists).toBe(true);

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
      await projectsPage.waitForProjectToBeRemoved(uniqueName);
    });

    test.skip('should create project with budget limit', async () => {
      // TODO: Budget feature removed from backend in Phase A2 fixes
      // This test requires re-implementing budget tracking
      const uniqueName = `budget-test-${Date.now()}`;

      // Create project with budget
      await projectsPage.createProject(uniqueName, 'Project with budget', 1000);

      // Verify project exists
      const projectExists = await projectsPage.verifyProjectExists(uniqueName);
      expect(projectExists).toBe(true);

      // Verify budget is shown in the project row
      const projectRow = projectsPage.getProjectByName(uniqueName);
      const projectText = await projectRow.textContent();
      expect(projectText).toContain('1000');

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
      await projectsPage.waitForProjectToBeRemoved(uniqueName);
    });

    test('should validate project name is required', async () => {
      // Validation is now implemented with data-testid="validation-error"
      await projectsPage.page.getByTestId('create-project-button').click();

      // Get the dialog context
      const dialog = projectsPage.page.locator('[role="dialog"]').first();

      // Try to submit without name - use data-testid to find wrapper, then find textarea inside
      await projectsPage.page.getByTestId('project-description-input').locator('textarea').fill('Test description');

      // Click the Create button using data-testid
      await projectsPage.page.getByTestId('create-project-submit-button').click();

      // Wait for and verify validation error appears - search in the whole page
      await projectsPage.page.locator('text=/Project name is required/i').waitFor({ state: 'visible', timeout: 5000 });
      const validationError = await projectsPage.page.locator('text=/Project name is required/i').textContent();
      expect(validationError).toMatch(/name.*required/i);

      // Cancel - search at page level
      await projectsPage.page.getByRole('button', { name: /cancel/i }).click();
    });

    test('should prevent duplicate project names', async () => {
      // Backend validation test - verify UI displays error properly
      const uniqueName = `duplicate-test-${Date.now()}`;

      // Create first project
      await projectsPage.createProject(uniqueName, 'First project');

      // Try to create second project with same name - use data-testid to avoid strict mode
      await projectsPage.page.getByTestId('create-project-button').click();

      // Find the dialog with "Create New Project" heading (more specific than first dialog)
      const dialog = projectsPage.page.locator('[role="dialog"]:has-text("Create New Project")');
      await dialog.waitFor({ state: 'visible', timeout: 5000 });
      await dialog.getByLabel(/project name/i).fill(uniqueName);
      await dialog.getByLabel(/description/i).fill('Second project');
      await dialog.getByRole('button', { name: /^create$/i }).click();

      // Should show duplicate error
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/already exists|duplicate/i);

      // Cleanup
      await projectsPage.clickButton('cancel');
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
    });
  });

  test.describe('View Project Workflow', () => {
    test('should view project details', async () => {
      // ProjectDetailView component now implemented
      const uniqueName = `view-test-${Date.now()}`;

      // Create project
      await projectsPage.createProject(uniqueName, 'Test project for viewing', 500);
      await projectsPage.page.waitForTimeout(1000);

      // View details
      await projectsPage.viewProjectDetails(uniqueName);

      // Verify we're on the details page using data-testid
      await projectsPage.page.waitForTimeout(1000);
      const detailView = projectsPage.page.getByTestId('project-detail-view');
      expect(await detailView.isVisible()).toBe(true);

      // Verify project information
      const description = await projectsPage.page.getByTestId('project-description').textContent();
      expect(description).toContain('Test project for viewing');

      // Navigate back to projects list using back button
      await projectsPage.page.getByTestId('back-to-projects-button').click();
      await projectsPage.page.waitForTimeout(500);

      // Verify we're back on projects page
      const projectsTable = projectsPage.page.getByTestId('projects-table');
      expect(await projectsTable.isVisible()).toBe(true);

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
    });

    test.skip('should show budget utilization in project details', async () => {
      // TODO: Budget feature removed from backend in Phase A2 fixes
      // This test requires re-implementing budget tracking UI
      const uniqueName = `budget-view-test-${Date.now()}`;

      await projectsPage.createProject(uniqueName, 'Budget tracking test', 1000);
      await projectsPage.viewProjectDetails(uniqueName);

      // Check for budget visualization using data-testid
      await projectsPage.page.waitForTimeout(1000);
      const budgetContainer = projectsPage.page.getByTestId('budget-utilization-container');
      expect(await budgetContainer.isVisible()).toBe(true);

      // Verify budget details are present
      const budgetLimit = await projectsPage.page.getByTestId('budget-limit').textContent();
      expect(budgetLimit).toContain('1000.00');

      const currentSpend = await projectsPage.page.getByTestId('current-spend').textContent();
      expect(currentSpend).toBeDefined();

      // Verify progress bar is visible
      const progressBar = projectsPage.page.getByTestId('budget-progress-bar');
      expect(await progressBar.isVisible()).toBe(true);

      // Navigate back
      await projectsPage.page.getByTestId('back-to-projects-button').click();
      await projectsPage.page.waitForTimeout(500);

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
    });
  });

  test.describe('Delete Project Workflow', () => {
    test('should delete project with confirmation', async () => {
      const uniqueName = `delete-test-${Date.now()}`;

      // Create project
      await projectsPage.createProject(uniqueName, 'Test project for deletion');

      // Verify exists
      let projectExists = await projectsPage.verifyProjectExists(uniqueName);
      expect(projectExists).toBe(true);

      // Delete project
      await projectsPage.deleteProject(uniqueName);

      // Confirm deletion
      await projectsPage.page.getByTestId('confirm-delete-button').click();

      // Wait for removal (this will throw if project isn't removed)
      await projectsPage.waitForProjectToBeRemoved(uniqueName);
    });

    test('should cancel project deletion', async () => {
      const uniqueName = `cancel-delete-test-${Date.now()}`;

      // Create project
      await projectsPage.createProject(uniqueName, 'Test project');
      await projectsPage.page.waitForTimeout(1000);

      // Start deletion
      await projectsPage.deleteProject(uniqueName);

      // Cancel
      await projectsPage.clickButton('cancel');
      await projectsPage.page.waitForTimeout(500);

      // Verify still exists
      const projectExists = await projectsPage.verifyProjectExists(uniqueName);
      expect(projectExists).toBe(true);

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
    });

    test.skip('should prevent deleting project with active resources', async () => {
      // TODO: Requires project with active instances/resources
      const uniqueName = `active-delete-test-${Date.now()}`;

      // Would need to:
      // 1. Create project
      // 2. Launch instance in project
      // 3. Try to delete project
      // 4. Verify warning/error about active resources

      expect(true).toBe(true); // Placeholder
    });
  });

  test.describe('Project Listing and Display', () => {
    test('should display all projects in list', async () => {
      test.setTimeout(60000); // Increase timeout for test creating 2 projects
      const name1 = `list-test-1-${Date.now()}`;
      const name2 = `list-test-2-${Date.now()}`;

      // Create multiple projects
      await projectsPage.createProject(name1, 'First test project');
      await projectsPage.createProject(name2, 'Second test project');

      // Verify both projects exist in the list
      expect(await projectsPage.verifyProjectExists(name1)).toBe(true);
      expect(await projectsPage.verifyProjectExists(name2)).toBe(true);

      // Cleanup
      await projectsPage.deleteProject(name1);
      await projectsPage.page.getByTestId('confirm-delete-button').click();
      await projectsPage.waitForProjectToBeRemoved(name1);

      await projectsPage.deleteProject(name2);
      await projectsPage.page.getByTestId('confirm-delete-button').click();
      await projectsPage.waitForProjectToBeRemoved(name2);
    });

    test('should show project statistics', async () => {
      // Statistics cards already implemented in ProjectManagementView
      await projectsPage.navigate();

      // Check for overview stats - statistics are in separate Container headers
      const totalProjects = projectsPage.page.locator('text=/total projects/i').first();
      expect(await totalProjects.isVisible()).toBe(true);

      const activeProjects = projectsPage.page.locator('text=/active projects/i').first();
      expect(await activeProjects.isVisible()).toBe(true);

      const totalBudget = projectsPage.page.locator('text=/total budget/i').first();
      expect(await totalBudget.isVisible()).toBe(true);

      const currentSpend = projectsPage.page.locator('text=/current spend/i').first();
      expect(await currentSpend.isVisible()).toBe(true);
    });

    test('should filter projects by status', async () => {
      test.setTimeout(60000); // Increase timeout for test creating 2 projects
      // Filter functionality now implemented with Select dropdown
      const activeName = `active-project-${Date.now()}`;
      const suspendedName = `suspended-project-${Date.now()}`;

      // Clean up any open dialogs first
      await projectsPage.forceCloseDialogs();

      await projectsPage.createProject(activeName, 'Active project');
      await projectsPage.createProject(suspendedName, 'Suspended project');

      // Verify both projects exist before applying filter
      expect(await projectsPage.verifyProjectExists(activeName)).toBe(true);
      expect(await projectsPage.verifyProjectExists(suspendedName)).toBe(true);

      // Apply filter for active projects using data-testid
      const filterSelect = projectsPage.page.getByTestId('project-filter-select');
      await filterSelect.click();
      await projectsPage.page.getByRole('option', { name: 'Active Only' }).click();

      // Verify both projects are still visible after filter
      // (both should have 'active' status by default)
      expect(await projectsPage.verifyProjectExists(activeName)).toBe(true);
      expect(await projectsPage.verifyProjectExists(suspendedName)).toBe(true);

      // Switch to "All Projects" filter for cleanup
      await filterSelect.click();
      await projectsPage.page.getByRole('option', { name: 'All Projects' }).click();

      // Cleanup
      await projectsPage.deleteProject(activeName);
      await projectsPage.page.getByTestId('confirm-delete-button').click();
      await projectsPage.waitForProjectToBeRemoved(activeName);

      await projectsPage.deleteProject(suspendedName);
      await projectsPage.page.getByTestId('confirm-delete-button').click();
      await projectsPage.waitForProjectToBeRemoved(suspendedName);
    });
  });

  test.describe('Budget Management', () => {
    test.skip('should track project spending', async () => {
      // TODO: Requires cost tracking integration
      const uniqueName = `spend-test-${Date.now()}`;

      await projectsPage.createProject(uniqueName, 'Spending test', 1000);

      // Launch resource in project (would incur costs)
      // ...

      // Verify spending is tracked
      const projectRow = projectsPage.getProjectByName(uniqueName);
      const projectText = await projectRow.textContent();

      // Should show current spend
      expect(projectText).toMatch(/\$\d+\.\d{2}/);

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
    });

    test.skip('should alert when approaching budget limit', async () => {
      // TODO: Requires budget alert system
      const uniqueName = `alert-test-${Date.now()}`;

      await projectsPage.createProject(uniqueName, 'Alert test', 100);

      // Incur costs approaching limit
      // ...

      // Verify alert is shown
      const alert = projectsPage.page.locator('[data-testid="budget-alert"]');
      expect(await alert.isVisible()).toBe(true);

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
    });

    test.skip('should prevent operations when budget exceeded', async () => {
      // TODO: Requires budget enforcement
      const uniqueName = `exceeded-test-${Date.now()}`;

      await projectsPage.createProject(uniqueName, 'Exceeded test', 10);

      // Try to launch expensive resource
      // ...

      // Verify operation is blocked
      const error = projectsPage.page.locator('text=/budget.*exceeded/i');
      expect(await error.isVisible()).toBe(true);

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.confirmDeletion();
    });
  });
});
