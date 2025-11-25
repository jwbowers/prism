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

    // Clean up any test projects from previous runs
    try {
      await projectsPage.cleanupTestProjects(/^test-project-|^budget-test-|^view-test-|^delete-test-|^list-test-/);
    } catch (error) {
      console.log('Project cleanup failed:', error);
    }
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
      await projectsPage.clickButton('delete');
      await projectsPage.waitForProjectToBeRemoved(uniqueName);
    });

    test('should create project with budget limit', async () => {
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
      await projectsPage.clickButton('delete');
      await projectsPage.waitForProjectToBeRemoved(uniqueName);
    });

    test.skip('should validate project name is required', async () => {
      // TODO: UI implementation gap - need validation with data-testid="validation-error"
      await projectsPage.page.getByRole('button', { name: /create project/i }).click();

      // Try to submit without name
      await projectsPage.fillInput('description', 'Test description');
      await projectsPage.clickButton('create');

      // Should show validation error
      const dialog = projectsPage.page.locator('[role="dialog"]').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/name.*required/i);

      // Cancel
      await projectsPage.clickButton('cancel');
    });

    test.skip('should prevent duplicate project names', async () => {
      // TODO: Backend validation - verify UI displays error properly
      const uniqueName = `duplicate-test-${Date.now()}`;

      // Create first project
      await projectsPage.createProject(uniqueName, 'First project');

      // Try to create second project with same name
      await projectsPage.page.getByRole('button', { name: /create project/i }).click();
      await projectsPage.fillInput('project name', uniqueName);
      await projectsPage.fillInput('description', 'Second project');
      await projectsPage.clickButton('create');

      // Should show duplicate error
      const dialog = projectsPage.page.locator('[role="dialog"]').first();
      const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
      expect(validationError).toMatch(/already exists|duplicate/i);

      // Cleanup
      await projectsPage.clickButton('cancel');
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.clickButton('delete');
    });
  });

  test.describe('View Project Workflow', () => {
    test.skip('should view project details', async () => {
      // TODO: Requires project details view to be navigable
      const uniqueName = `view-test-${Date.now()}`;

      // Create project
      await projectsPage.createProject(uniqueName, 'Test project for viewing', 500);
      await projectsPage.page.waitForTimeout(1000);

      // View details
      await projectsPage.viewProjectDetails(uniqueName);

      // Verify we're on the details page
      await projectsPage.page.waitForTimeout(1000);
      const pageContent = await projectsPage.page.textContent('body');
      expect(pageContent).toContain(uniqueName);
      expect(pageContent).toContain('Test project for viewing');

      // Navigate back to projects list
      await projectsPage.navigate();

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.clickButton('delete');
    });

    test.skip('should show budget utilization in project details', async () => {
      // TODO: Requires budget tracking UI
      const uniqueName = `budget-view-test-${Date.now()}`;

      await projectsPage.createProject(uniqueName, 'Budget tracking test', 1000);
      await projectsPage.viewProjectDetails(uniqueName);

      // Check for budget visualization
      const budgetSection = projectsPage.page.locator('text=/budget.*utilization/i');
      expect(await budgetSection.isVisible()).toBe(true);

      // Navigate back
      await projectsPage.navigate();

      // Cleanup
      await projectsPage.deleteProject(uniqueName);
      await projectsPage.clickButton('delete');
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
      await projectsPage.clickButton('delete');

      // Wait for removal
      await projectsPage.waitForProjectToBeRemoved(uniqueName);

      // Verify removed
      projectExists = await projectsPage.verifyProjectExists(uniqueName);
      expect(projectExists).toBe(false);
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
      await projectsPage.clickButton('delete');
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
      const name1 = `list-test-1-${Date.now()}`;
      const name2 = `list-test-2-${Date.now()}`;

      // Get initial count
      const initialCount = await projectsPage.getProjectCount();

      // Create multiple projects
      await projectsPage.createProject(name1, 'First test project');
      await projectsPage.createProject(name2, 'Second test project');

      // Verify count increased
      const newCount = await projectsPage.getProjectCount();
      expect(newCount).toBe(initialCount + 2);

      // Cleanup
      await projectsPage.deleteProject(name1);
      await projectsPage.clickButton('delete');
      await projectsPage.waitForProjectToBeRemoved(name1);

      await projectsPage.deleteProject(name2);
      await projectsPage.clickButton('delete');
      await projectsPage.waitForProjectToBeRemoved(name2);
    });

    test.skip('should show project statistics', async () => {
      // TODO: Verify stats cards at top of page
      await projectsPage.navigate();

      // Check for overview stats
      const statsSection = projectsPage.page.locator('text=/total projects|active projects|total budget|current spend/i').first();
      expect(await statsSection.isVisible()).toBe(true);
    });

    test.skip('should filter projects by status', async () => {
      // TODO: Requires filter/search functionality
      const activeName = `active-project-${Date.now()}`;
      const suspendedName = `suspended-project-${Date.now()}`;

      await projectsPage.createProject(activeName, 'Active project');
      await projectsPage.createProject(suspendedName, 'Suspended project');

      // Apply filter for active projects
      // ... filtering logic

      // Verify only active shown
      // ... verification

      // Cleanup
      await projectsPage.deleteProject(activeName);
      await projectsPage.clickButton('delete');
      await projectsPage.deleteProject(suspendedName);
      await projectsPage.clickButton('delete');
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
      await projectsPage.clickButton('delete');
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
      await projectsPage.clickButton('delete');
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
      await projectsPage.clickButton('delete');
    });
  });
});
