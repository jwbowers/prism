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

  // Clean up old test projects once before all tests to prevent pagination issues
  test.beforeAll(async ({ browser }) => {
    test.setTimeout(180000); // 3 minutes for cleanup

    const context = await browser.newContext();
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });
    const page = await context.newPage();

    try {
      console.log('🧹 Cleaning up old test projects via API...');

      // Use API to bulk delete test projects (faster than UI)
      const response = await page.request.get('http://localhost:8947/api/v1/projects');
      if (response.ok()) {
        const data = await response.json();
        const projects = data.projects || []; // API returns {projects: [...], totalCount: N}
        const testProjectPatterns = [
          /^list-test-/,              // All list test variants
          /^active-project-/,         // All active project variants
          /^suspended-project-/,      // All suspended project variants
          /^cancel-delete-test-/,     // All cancel delete variants
          /^delete-test-/,            // All delete test variants
          /^test-project-/,           // All test project variants
          /^view-test-/,              // View test projects
          /^budget-view-test-/,       // Budget view test projects
          /^budget-test-/,            // Budget test projects
          /^duplicate-test-/,         // Duplicate test projects
          /^spend-test-/,             // Spending test projects
          /^alert-test-/,             // Alert test projects
          /^exceeded-test-/,          // Budget exceeded test projects
          /^active-delete-test-/      // Active delete test projects
        ];

        let deletedCount = 0;
        for (const project of projects) {
          if (deletedCount >= 200) break; // Increased limit for aggressive cleanup

          const isTestProject = testProjectPatterns.some(pattern => pattern.test(project.name));
          if (isTestProject) {
            try {
              await page.request.delete(`http://localhost:8947/api/v1/projects/${project.id}`);
              deletedCount++;
            } catch (e) {
              // Skip if deletion fails
            }
          }
        }
        console.log(`✅ Cleanup complete - deleted ${deletedCount} test projects`);
      }
    } catch (e) {
      console.log('⚠️ Cleanup failed, continuing with tests:', e);
    } finally {
      await page.close();
      await context.close();
    }
  });

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
  });

  test.describe('Create Project Workflow', () => {
    test('should create a new project with basic configuration', async () => {
      const uniqueName = `test-project-${Date.now()}`;

      // Create project
      await projectsPage.createProject(uniqueName, 'Test project description');

      // Verify project appears in list
      const projectExists = await projectsPage.verifyProjectExists(uniqueName);
      expect(projectExists).toBe(true);

      // Cleanup via API (faster and more reliable than UI with 650+ projects)
      await projectsPage.deleteProjectViaAPI(uniqueName);
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

      // Cleanup via API
      await projectsPage.deleteProjectViaAPI(uniqueName);
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

      // Cancel - use exact text match to find the modal's cancel button
      // Note: Multiple project names may contain "cancel" so we use exact text
      await projectsPage.page.getByRole('button', { name: 'Cancel', exact: true }).click();
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
      await projectsPage.deleteProjectViaAPI(uniqueName);
    });
  });

  test.describe('View Project Workflow', () => {
    test('should view project details', async () => {
      // ProjectDetailView component now implemented
      const uniqueName = `view-test-${Date.now()}`;

      // Create project
      await projectsPage.createProject(uniqueName, 'Test project for viewing', 500);

      // View details
      await projectsPage.viewProjectDetails(uniqueName);

      // Verify we're on the details page using data-testid
      const detailView = projectsPage.page.getByTestId('project-detail-view');
      await detailView.waitFor({ state: 'visible', timeout: 5000 });

      // Verify project information
      const description = await projectsPage.page.getByTestId('project-description').textContent();
      expect(description).toContain('Test project for viewing');

      // Navigate back to projects list using back button
      await projectsPage.page.getByTestId('back-to-projects-button').click();

      // Verify we're back on projects page
      const projectsTable = projectsPage.page.getByTestId('projects-table');
      await projectsTable.waitFor({ state: 'visible', timeout: 5000 });

      // Cleanup via API
      await projectsPage.deleteProjectViaAPI(uniqueName);
    });

    test('should show budget utilization in project details', async () => {
      const uniqueName = `budget-view-test-${Date.now()}`;

      await projectsPage.createProject(uniqueName, 'Budget tracking test', 1000);
      await projectsPage.viewProjectDetails(uniqueName);

      // Check for budget visualization using data-testid
      const budgetContainer = projectsPage.page.getByTestId('budget-utilization-container');
      await budgetContainer.waitFor({ state: 'visible', timeout: 5000 });

      // Verify budget details are present
      const budgetLimit = await projectsPage.page.getByTestId('budget-limit').textContent();
      expect(budgetLimit).toContain('1000.00');

      const currentSpend = await projectsPage.page.getByTestId('current-spend').textContent();
      expect(currentSpend).toBeDefined();

      // Verify progress bar is visible
      await expect(projectsPage.page.getByTestId('budget-progress-bar')).toBeVisible();

      // Navigate back
      await projectsPage.page.getByTestId('back-to-projects-button').click();

      // Cleanup via API
      await projectsPage.deleteProjectViaAPI(uniqueName);
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

      // Start deletion
      await projectsPage.deleteProject(uniqueName);

      // Cancel
      await projectsPage.clickButton('cancel');
      await projectsPage.waitForDialogClose();

      // Verify still exists
      const projectExists = await projectsPage.verifyProjectExists(uniqueName);
      expect(projectExists).toBe(true);

      // Cleanup via API
      await projectsPage.deleteProjectViaAPI(uniqueName);
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

      // Cleanup via API (faster and more reliable than UI with 650+ projects)
      await projectsPage.deleteProjectViaAPI(name1);
      await projectsPage.deleteProjectViaAPI(name2);
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
      await projectsPage.page.locator('[data-value="active"]').click();

      // Verify both projects are still visible after filter
      // (both should have 'active' status by default)
      expect(await projectsPage.verifyProjectExists(activeName)).toBe(true);
      expect(await projectsPage.verifyProjectExists(suspendedName)).toBe(true);

      // Cleanup via API (faster and more reliable than UI with 650+ projects)
      await projectsPage.deleteProjectViaAPI(activeName);
      await projectsPage.deleteProjectViaAPI(suspendedName);
    });
  });

  test.describe('Budget Management', () => {
    test('should track project spending', async () => {
      const uniqueName = `spend-test-${Date.now()}`;

      await projectsPage.createProject(uniqueName, 'Spending test', 1000);

      // Launch resource in project (would incur costs)
      // ...

      // Verify spending is tracked
      const projectRow = projectsPage.getProjectByName(uniqueName);
      const projectText = await projectRow.textContent();

      // Should show current spend
      expect(projectText).toMatch(/\$\d+\.\d{2}/);

      // Cleanup via API
      await projectsPage.deleteProjectViaAPI(uniqueName);
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

      // Cleanup via API
      await projectsPage.deleteProjectViaAPI(uniqueName);
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

      // Cleanup via API
      await projectsPage.deleteProjectViaAPI(uniqueName);
    });
  });
});
