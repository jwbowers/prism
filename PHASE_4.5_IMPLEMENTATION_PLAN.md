# Phase 4.5: Accept/Decline Invitation Workflows - Implementation Plan

**Date**: 2025-11-25
**Epic**: Issue #315 (E2E Test Activation)
**Milestone**: v0.5.16 (Projects & Users - Week 2)
**Phase**: 4.5 of 10 (Accept/Decline Workflows)
**Complexity**: MEDIUM - Requires invitation creation setup, modal interaction testing

---

## Executive Summary

Phase 4.5 activates 6 E2E tests for invitation Accept/Decline workflows. Unlike Phase 4.4 which created the UI component, this phase focuses on **testing the existing modal functionality** for accepting and declining invitations.

**Key Challenge**: Tests require **pending invitations** to exist before they can test accept/decline actions.

**Scope**: 6 tests to activate, test setup helpers to create

---

## Current State Analysis

### ✅ What Exists (Already Implemented)

**InvitationManagementView Component** (Phase 4.4 - Complete):
- ✅ Accept invitation modal (lines 761-817)
- ✅ Decline invitation modal (lines 819-878)
- ✅ Modal confirmation dialogs with project details
- ✅ Optional reason field for decline
- ✅ API integration (`acceptInvitation()`, `declineInvitation()`)
- ✅ Status update after accept/decline

**Test Helper Methods** (ProjectsPage.ts):
- ✅ `acceptInvitation(projectName)` (lines 337-348)
- ✅ `declineInvitation(projectName, reason?)` (lines 353-367)
- ✅ `switchToIndividualInvitations()` (line 306)
- ✅ `getInvitationRows()` (line 318)

**Backend API Endpoints**:
- ✅ `POST /api/v1/projects/{id}/invitations` - Send invitation
- ✅ `POST /api/v1/invitations/{token}/accept` - Accept invitation
- ✅ `POST /api/v1/invitations/{token}/decline` - Decline invitation

### ❌ What's Missing (Needs Implementation)

**Test Setup Infrastructure**:
1. Method to create test project for invitations
2. Method to send invitation to test user (get invitation token)
3. Method to add invitation token to Individual Invitations tab
4. Test data cleanup after each test

---

## Tests to Activate

### Accept Invitation Workflow (3 tests)

**Test 1: "should accept invitation with confirmation"** (line 87-101)
- Creates pending invitation
- Clicks Accept button
- Confirms in modal
- Verifies status changes to "Accepted"

**Test 2: "should show acceptance confirmation dialog"** (line 103-121)
- Creates pending invitation
- Clicks Accept button
- Verifies modal displays project details and role
- Tests Cancel button

**Test 3: "should add user to project after acceptance"** (line 123-139)
- Creates pending invitation
- Accepts invitation
- Navigates to project details
- Verifies user appears in members list

### Decline Invitation Workflow (3 tests)

**Test 4: "should decline invitation with reason"** (line 143-158)
- Creates pending invitation
- Clicks Decline button
- Enters decline reason
- Verifies status changes to "Declined"

**Test 5: "should show decline confirmation dialog"** (line 160-178)
- Creates pending invitation
- Clicks Decline button
- Verifies modal shows reason input field
- Tests Cancel button

**Test 6: "should allow declining without reason"** (line 180-194)
- Creates pending invitation
- Clicks Decline button
- Confirms without entering reason
- Verifies status changes to "Declined"

---

## Implementation Strategy

### Approach: API-Based Invitation Setup

**Why**: Tests need real pending invitations to interact with modals.

**Setup Flow**:
1. Create test project via API
2. Send invitation to test user via API
3. Get invitation token from API response
4. Add invitation token via UI (Individual Invitations tab)
5. Run test interactions (Accept/Decline)
6. Cleanup test data

### Test Structure Pattern

```typescript
test('should accept invitation with confirmation', async () => {
  // SETUP: Create test invitation
  const testProjectId = await projectsPage.createTestProject('Test Invitation Project');
  const invitationToken = await projectsPage.sendTestInvitation(
    testProjectId,
    'testuser@example.com',
    'member'
  );

  // Add invitation to Individual Invitations tab
  await projectsPage.navigateToInvitations();
  await projectsPage.switchToIndividualInvitations();
  await projectsPage.addInvitationToken(invitationToken);

  // TEST: Accept invitation
  await projectsPage.acceptInvitation('Test Invitation Project');

  // VERIFY: Status changed
  const invitationRow = projectsPage.page.locator(`tr:has-text("Test Invitation Project")`).first();
  const invitationText = await invitationRow.textContent();
  expect(invitationText).toContain('Accepted');
});
```

---

## Required Test Helper Methods

### Method 1: Create Test Project

**Location**: `tests/e2e/pages/ProjectsPage.ts`

```typescript
/**
 * Create test project via API for invitation testing
 */
async createTestProject(name: string): Promise<string> {
  const response = await this.page.evaluate(async (projectName) => {
    const api = (window as any).__apiClient;
    const project = await api.createProject({
      name: projectName,
      description: 'Test project for invitation workflows',
      owner: 'test-owner',
      budget_limit: 1000,
      budget_period: 'monthly',
      status: 'active'
    });
    return project.id;
  }, name);

  return response;
}
```

### Method 2: Send Test Invitation

**Location**: `tests/e2e/pages/ProjectsPage.ts`

```typescript
/**
 * Send invitation to test user and return invitation token
 */
async sendTestInvitation(
  projectId: string,
  email: string,
  role: 'viewer' | 'member' | 'admin'
): Promise<string> {
  const token = await this.page.evaluate(async (args) => {
    const api = (window as any).__apiClient;
    const invitation = await api.sendInvitation(
      args.projectId,
      args.email,
      args.role,
      'Test invitation message'
    );
    return invitation.token;
  }, { projectId, email, role });

  return token;
}
```

### Method 3: View Project Members (for Test 3)

**Location**: `tests/e2e/pages/ProjectsPage.ts`

```typescript
/**
 * Verify user appears in project members list
 */
async verifyProjectMember(projectName: string, username: string): Promise<boolean> {
  await this.viewProjectDetails(projectName);
  await this.page.waitForTimeout(1000);

  const membersSection = this.page.locator('[data-testid="project-members"]');
  if (await membersSection.isVisible()) {
    const membersText = await membersSection.textContent();
    return membersText?.includes(username) || false;
  }

  return false;
}
```

---

## Implementation Steps

### Step 1: Add Test Helper Methods ✅
1. Add `createTestProject()` to ProjectsPage.ts
2. Add `sendTestInvitation()` to ProjectsPage.ts
3. Add `verifyProjectMember()` to ProjectsPage.ts

### Step 2: Update Test Structure ✅
1. Add `beforeEach` setup for invitation creation
2. Add `afterEach` cleanup for test projects
3. Update test selectors if needed

### Step 3: Unskip Accept Invitation Tests (3 tests) ✅
1. "should accept invitation with confirmation" (line 87)
2. "should show acceptance confirmation dialog" (line 103)
3. "should add user to project after acceptance" (line 123)

### Step 4: Unskip Decline Invitation Tests (3 tests) ✅
1. "should decline invitation with reason" (line 143)
2. "should show decline confirmation dialog" (line 160)
3. "should allow declining without reason" (line 180)

### Step 5: Run E2E Tests ✅
1. Run individual tests to verify functionality
2. Run full invitation-workflows.spec.ts suite
3. Debug any failures

### Step 6: Document Results ✅
1. Create ISSUE_315_PHASE_4.5_SUMMARY.md
2. Update Epic #315 on GitHub
3. Commit changes

---

## Testing Strategy

### Unit Testing (Optional)
- Test helper method functionality
- Mock API responses

### E2E Testing (Primary Focus)
1. **Test 1-2**: Basic Accept workflow with modal verification
2. **Test 3**: Full Accept → Project Membership flow
3. **Test 4-5**: Decline workflow with optional reason
4. **Test 6**: Decline without reason

### Expected Behavior

**Accept Flow**:
1. Click Accept button → Modal opens
2. Modal shows project name, role, optional message
3. Click Accept → API call → Status updates to "Accepted"
4. User added to project members (Test 3)

**Decline Flow**:
1. Click Decline button → Modal opens
2. Modal shows optional reason textarea
3. Enter reason (or leave blank) → Click Decline
4. API call → Status updates to "Declined"

---

## Known Challenges

### Challenge 1: Invitation Token Requirement
**Problem**: Tests need actual invitation tokens from backend
**Solution**: Create test invitations via API before each test

### Challenge 2: Project Membership Verification (Test 3)
**Problem**: Need to verify user was added to project after acceptance
**Solution**: Navigate to project details and check members section

### Challenge 3: Test Data Cleanup
**Problem**: Test projects and invitations accumulate
**Solution**: Add `afterEach` cleanup to delete test projects

### Challenge 4: Timing Issues
**Problem**: Modal animations and API calls may need wait times
**Solution**: Use `waitForTimeout()` and `waitForSelector()` appropriately

---

## Success Criteria

**Phase 4.5 Complete When**:
1. ✅ Test helper methods created for invitation setup
2. ✅ 3 Accept Invitation tests passing
3. ✅ 3 Decline Invitation tests passing
4. ✅ Modal interactions verified (Cancel, Confirm)
5. ✅ Project membership verification working (Test 3)
6. ✅ TypeScript compilation zero errors
7. ✅ Documentation updated

---

## Risk Assessment

**Technical Risks**:
- ⚠️ MEDIUM: Backend API must support invitation creation in test mode
- ⚠️ MEDIUM: Test data cleanup required to prevent state pollution
- ⚠️ LOW: Modals already implemented and functional (Phase 4.4)
- ⚠️ LOW: Test helper methods follow established patterns

**Mitigation**:
- Use `PRISM_TEST_MODE=true` for API testing
- Add `afterEach` cleanup for all test data
- Follow existing test patterns from Phase 4.4
- Test one workflow at a time (Accept first, then Decline)

---

## Estimated Effort

- **Test Helper Methods**: 1 hour
- **Update Test Structure**: 1 hour
- **Unskip Accept Tests**: 1 hour
- **Unskip Decline Tests**: 1 hour
- **Testing & Debugging**: 2 hours
- **Documentation**: 0.5 hours

**Total**: ~6-7 hours for complete Phase 4.5 implementation

---

## References

**Component Implementation**:
- InvitationManagementView.tsx: Accept modal (lines 761-817), Decline modal (lines 819-878)

**Test Files**:
- invitation-workflows.spec.ts: Tests to activate (lines 86-195)
- ProjectsPage.ts: Test helper methods (lines 306-367)

**Backend APIs**:
- pkg/daemon/invitation_handlers.go: API endpoint implementations

**Related Phases**:
- Phase 4.4: InvitationManagementView component creation (Complete)
- Phase 4.6: Bulk Invitations Enhancement (Future)
- Phase 4.7: Shared Tokens System (Future)

---

**Status**: ✅ Planning Complete - Ready for Implementation

**Next Step**: Implement test helper methods for invitation creation setup
