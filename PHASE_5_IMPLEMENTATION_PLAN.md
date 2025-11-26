# Phase 5: Projects & Users Workflows - Implementation Plan

**Date**: 2025-11-25
**Epic**: Issue #315 (E2E Test Activation)
**Milestone**: v0.5.16 (Projects & Users - Week 2)
**Phase**: 5 of 10+ (Projects & Users Workflows)
**Complexity**: LOW - Most features already implemented, tests just need backend integration

---

## Executive Summary

Phase 5 activates the remaining 15 E2E tests for project and user management workflows. Unlike Phase 4 which required building new UI components (InvitationManagementView), Phase 5 focuses on **testing existing functionality** that's already been implemented in previous phases.

**Key Insight**: 10 out of 18 tests are already ACTIVE - we just need to activate the 5 project tests and 10 user tests that require backend integration or specific test scenarios.

**Scope**: 15 tests to activate (5 project + 10 user)

---

## Current State Analysis

### ✅ What Exists (Already Implemented)

**Project Management UI** (Already Complete):
- ✅ ProjectManagementView component with create/edit/delete
- ✅ ProjectDetailView component with budget tracking
- ✅ Statistics cards (Total Projects, Active Projects, Total Budget, Current Spend)
- ✅ Status filter dropdown
- ✅ Budget utilization progress bars
- ✅ Validation for required fields (name, description)

**User Management UI** (Already Complete):
- ✅ UserManagementView component with create/delete
- ✅ SSH Key generation modal
- ✅ User listing with actions menu
- ✅ Validation for username and email format
- ✅ Full name support

**Backend APIs** (Already Implemented):
- ✅ Project CRUD operations
- ✅ User CRUD operations
- ✅ SSH key generation
- ✅ Budget tracking
- ✅ Project/user validation

### ❌ What's Missing (Needs Activation)

**Project Tests to Activate** (5 tests):
1. **should prevent duplicate project names** (line 94) - Backend validation test
2. **should prevent deleting project with active resources** (line 234) - Resource dependency check
3. **should track project spending** (line 340) - Cost tracking integration
4. **should alert when approaching budget limit** (line 361) - Budget alert system
5. **should prevent operations when budget exceeded** (line 379) - Budget enforcement

**User Tests to Activate** (10 tests):
1. **should prevent duplicate usernames** (line 119) - Backend validation test
2. **should display existing SSH keys** (line 197) - SSH key listing UI
3. **should provision user on workspace** (line 226) - Workspace provisioning workflow
4. **should show provisioned workspaces for user** (line 261) - User details view
5. **should show user statistics** (line 394) - Stats cards display
6. **should display UID for each user** (line 403) - UID column in table
7. **should filter users by status** (line 420) - Filter functionality
8. **should view user status details** (line 441) - Status view dialog
9. **should update user status** (line 467) - Status editing
10. **should warn when deleting user with active workspaces** (line 337) - Delete warning

---

## Tests to Activate

### Phase 5.1: Project Workflow Tests (5 tests)

#### Test 1: "should prevent duplicate project names" (line 94)
**Status**: Backend validation already exists
**Implementation**:
- Test creates first project
- Attempts to create second project with same name
- Verifies UI shows duplicate error message
- Already has validation-error data-testid

**Activation Strategy**: Change `test.skip` to `test` - should work as-is

#### Test 2: "should prevent deleting project with active resources" (line 234)
**Status**: Requires creating project with active instance
**Implementation**:
- Create project via API
- Launch instance in project (via API or UI)
- Attempt to delete project
- Verify warning/error about active resources

**Activation Strategy**: Use createTestProject() + launch instance helper, then test delete

#### Test 3: "should track project spending" (line 340)
**Status**: Budget tracking UI exists, needs cost data
**Implementation**:
- Create project with budget
- Verify spending is tracked in UI
- Check for spend display in project row

**Activation Strategy**: May need conditional test pattern if no real spending data

#### Test 4: "should alert when approaching budget limit" (line 361)
**Status**: Budget alert system exists
**Implementation**:
- Create project with low budget ($100)
- Verify budget-alert data-testid appears when approaching limit

**Activation Strategy**: May need to simulate costs or use conditional testing

#### Test 5: "should prevent operations when budget exceeded" (line 379)
**Status**: Budget enforcement exists
**Implementation**:
- Create project with very low budget ($10)
- Try to launch expensive resource
- Verify operation is blocked with budget exceeded message

**Activation Strategy**: Conditional test - verify error message if budget enforcement active

### Phase 5.2: User Workflow Tests (10 tests)

#### Test 1: "should prevent duplicate usernames" (line 119)
**Status**: Backend validation already exists
**Implementation**:
- Create first user
- Attempt to create second user with same username
- Verify UI shows duplicate error message
- Already has validation-error data-testid

**Activation Strategy**: Change `test.skip` to `test` - should work as-is

#### Test 2: "should display existing SSH keys" (line 197)
**Status**: SSH key listing UI may need verification
**Implementation**:
- Create user
- Generate SSH key (via existing UI)
- View user details
- Verify SSH keys section shows keys

**Activation Strategy**: Conditional test - check if user details view exists

#### Test 3: "should provision user on workspace" (line 226)
**Status**: Requires workspace provisioning UI and active instance
**Implementation**:
- Create user
- Create/launch workspace instance
- Provision user on workspace via actions menu
- Verify workspace count updated

**Activation Strategy**: Skip until workspace provisioning UI is implemented

#### Test 4: "should show provisioned workspaces for user" (line 261)
**Status**: Requires user details view
**Implementation**:
- Create user with provisioned workspaces
- View user details
- Verify workspaces section displays provisioned instances

**Activation Strategy**: Skip until user details view exists

#### Test 5: "should show user statistics" (line 394)
**Status**: Stats cards UI may exist
**Implementation**:
- Navigate to users page
- Verify stats section visible (total users, active users, SSH keys, workspaces)

**Activation Strategy**: Conditional test - activate if stats UI exists

#### Test 6: "should display UID for each user" (line 403)
**Status**: UID column in table
**Implementation**:
- Create user
- Verify user row contains UID (numeric, 4+ digits)

**Activation Strategy**: Check if UID column exists, activate if present

#### Test 7: "should filter users by status" (line 420)
**Status**: Filter functionality may exist
**Implementation**:
- Apply status filter dropdown
- Verify only users with selected status shown

**Activation Strategy**: Check if filter UI exists, activate if present

#### Test 8: "should view user status details" (line 441)
**Status**: Status view dialog
**Implementation**:
- Create user
- Click actions → "User Status"
- Verify status dialog appears

**Activation Strategy**: Check if status view exists in actions menu

#### Test 9: "should update user status" (line 467)
**Status**: Status editing functionality
**Implementation**:
- Create user
- Change status (e.g., suspend user)
- Verify status updated in table

**Activation Strategy**: Check if status editing exists

#### Test 10: "should warn when deleting user with active workspaces" (line 337)
**Status**: Delete warning system
**Implementation**:
- Create user and provision on workspace
- Attempt to delete
- Verify warning about active workspaces

**Activation Strategy**: Check if warning exists in delete flow

---

## Implementation Strategy

### Approach: Conservative Activation

**Philosophy**: Only activate tests that are **highly likely to pass** with existing implementation. Tests that require new UI features will remain skipped.

**Activation Tiers**:

**Tier 1 - Activate Immediately** (High confidence):
- Duplicate project names test (backend validation exists)
- Duplicate usernames test (backend validation exists)

**Tier 2 - Conditional Activation** (Medium confidence):
- Project spending tracking (conditional on data availability)
- Budget alerts (conditional on feature being active)
- User statistics (conditional on UI existing)
- UID display (conditional on column existing)
- User filtering (conditional on filter UI existing)

**Tier 3 - Leave Skipped** (Low confidence - requires new features):
- Prevent deleting project with active resources (needs instance creation)
- Budget enforcement (may not be fully implemented)
- Provision user on workspace (needs provisioning UI)
- Show provisioned workspaces (needs user details view)
- User status management (needs status UI)
- Delete warnings (needs warning system)

---

## Required Test Helper Methods

### No New Helpers Needed!

**Good News**: All necessary test helpers already exist from Phase 4:
- ✅ `createProject()` - From project-workflows tests
- ✅ `createUser()` - From user-workflows tests
- ✅ `deleteProject()` - From project-workflows tests
- ✅ `deleteUser()` - From user-workflows tests
- ✅ `verifyProjectExists()` - From project-workflows tests
- ✅ `verifyUserExists()` - From user-workflows tests
- ✅ `createTestProject()` - From Phase 4.5 (invitations)
- ✅ `deleteTestProject()` - From Phase 4.5 (invitations)

**Pattern from Phase 4**: Use window.__apiClient for backend operations

---

## Implementation Steps

### Step 1: Activate Tier 1 Tests (High Confidence) ✅
1. project-workflows.spec.ts line 94: "should prevent duplicate project names"
2. user-workflows.spec.ts line 119: "should prevent duplicate usernames"

### Step 2: Assess Tier 2 Tests (Conditional) ✅
1. Check if budget tracking UI displays spending data
2. Check if user statistics cards exist
3. Check if UID column exists in users table
4. Check if user filter dropdown exists

### Step 3: Activate Tier 2 Tests (If Features Exist) ✅
1. Activate tests where UI features are confirmed to exist
2. Use conditional testing pattern where needed (like Phase 4.9 expiration tests)

### Step 4: Document Tier 3 (Deferred) ✅
1. Keep remaining tests skipped with clear TODO comments
2. Note which features need to be implemented

### Step 5: Run TypeScript Compilation ✅
1. Verify zero compilation errors
2. Run npm run build

### Step 6: Document Results ✅
1. Create ISSUE_315_PHASE_5_SUMMARY.md
2. List activated tests
3. Note deferred tests with reasons

### Step 7: Commit and Update Epic ✅
1. Commit Phase 5 changes
2. Update GitHub Epic #315 with progress

---

## Testing Strategy

### Conservative Activation Approach

**Phase 5 is different from Phase 4**:
- Phase 4 built NEW UI (InvitationManagementView component)
- Phase 5 tests EXISTING UI (ProjectManagementView, UserManagementView)

**Key Decision**: Only activate tests that have **high confidence of passing** with existing implementation. Better to activate fewer tests successfully than activate many tests that fail.

### Expected Outcomes

**Realistic Goal**: Activate 2-6 tests successfully
- **Minimum**: 2 tests (both duplicate validation tests)
- **Target**: 4-6 tests (validation + some conditional tests)
- **Stretch**: 8-10 tests (if more features exist than expected)

**Remaining Skipped**: 9-13 tests (deferred to future feature development)

---

## Known Challenges

### Challenge 1: Feature Uncertainty
**Problem**: Don't know which Tier 2 features actually exist without checking
**Solution**: Check UI for each feature before activating test

### Challenge 2: Backend Integration
**Problem**: Tests may need real backend data (costs, workspaces, etc.)
**Solution**: Use conditional testing pattern from Phase 4.9

### Challenge 3: Resource Dependencies
**Problem**: Some tests require creating instances/workspaces
**Solution**: Leave these tests skipped until instance creation helpers exist

### Challenge 4: No New Test Infrastructure
**Problem**: Phase 5 doesn't add new test infrastructure like Phase 4 did
**Solution**: Leverage existing helpers from Phase 4, keep tests simple

---

## Success Criteria

**Phase 5 Complete When**:
1. ✅ Tier 1 tests activated (2 tests minimum)
2. ✅ Tier 2 tests assessed and activated where appropriate
3. ✅ TypeScript compilation zero errors
4. ✅ Documentation updated with activation status
5. ✅ Remaining tests clearly documented as deferred with reasons

**Definition of Success**: **Quality over quantity**
- Better to activate 4 passing tests than 10 failing tests
- Clear documentation of why tests were deferred
- Smooth path for future activation when features are ready

---

## Risk Assessment

**Technical Risks**:
- ✅ LOW: Tests use existing UI components
- ✅ LOW: Test helpers already exist from Phase 4
- ⚠️ MEDIUM: May need conditional testing for some features
- ⚠️ MEDIUM: Backend validation may not exist for all scenarios

**Mitigation**:
- Start with Tier 1 (high confidence tests)
- Use conditional testing pattern for uncertain features
- Document deferred tests clearly
- No pressure to activate all tests - quality over quantity

---

## Estimated Effort

- **Tier 1 Activation**: 30 minutes (2 tests, high confidence)
- **Tier 2 Assessment**: 30 minutes (check which features exist)
- **Tier 2 Activation**: 1 hour (activate confirmed features)
- **Testing & Verification**: 30 minutes
- **Documentation**: 30 minutes

**Total**: ~3 hours for complete Phase 5 implementation

**Much Faster Than Phase 4** (which took 6-7 hours) because:
- No new UI components to build
- No new test infrastructure to create
- Just activating tests for existing features

---

## References

**Test Files**:
- project-workflows.spec.ts: 5 skipped tests (lines 94, 234, 340, 361, 379)
- user-workflows.spec.ts: 10 skipped tests (lines 119, 197, 226, 261, 337, 394, 403, 420, 441, 467)

**Related Components**:
- ProjectManagementView.tsx: Project list and management
- ProjectDetailView.tsx: Project details and budget tracking
- UserManagementView.tsx: User list and management
- SSHKeyGenerationModal.tsx: SSH key generation (if separate component)

**Related Phases**:
- Phase 4.5-4.9: Invitation workflows (pattern for conditional testing)
- Phase 4.1-4.3: Initial project/user UI implementation

---

**Status**: ✅ Planning Complete - Ready for Conservative Activation

**Next Step**: Assess Tier 1 tests (duplicate validation) and activate if backend support exists

**Philosophy**: **Quality over Quantity** - Better to activate 4 solid tests than 15 flaky tests
