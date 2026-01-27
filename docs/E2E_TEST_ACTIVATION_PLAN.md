# E2E Test Activation Plan - Issue #315

**Last Updated**: January 26, 2026
**Status**: Investigation Complete, Action Plan Defined

## Executive Summary

Issue #315 tracks activation of 51 skipped E2E tests. Investigation reveals:

**Current State**:
- **15 tests explicitly skipped** with `test.skip` markers
  - 6 budget-related tests (feature removed)
  - 9 user feature tests (UI not implemented)
- **28 invitation tests** exist (not marked skip, status unknown - daemon won't start)
- **~10 tests actively passing** (project CRUD basics)

**Root Causes**:
1. Budget tracking removed from backend → 6 tests obsolete
2. User management UI incomplete → 9 tests skipped
3. Invitation system not implemented → 28 tests untested

## Detailed Test Breakdown

### Project Workflows (project-workflows.spec.ts)

**Total Tests**: 15
**Active Tests**: 9 (should be passing)
**Skipped Tests**: 6 (all budget-related)

**Skipped**:
1. `should create project with budget limit`
2. `should show budget utilization in project details`
3. `should prevent deleting project with active resources`
4. `should track project spending`
5. `should alert when approaching budget limit`
6. `should prevent operations when budget exceeded`

**Reason**: "Budget feature removed from backend in Phase A2 fixes"

**Recommendation**: Remove these tests in v0.7.7 or defer to v0.11.0

---

### User Workflows (user-workflows.spec.ts)

**Skipped Tests**: 9

**Categories**:

**SSH Key Management (1 test)**:
- `should display existing SSH keys`

**User Provisioning (3 tests)**:
- `should provision user on workspace`
- `should show provisioned workspaces for user`
- `should warn when deleting user with active workspaces`

**User Statistics (1 test)**:
- `should show user statistics`

**User Display & Filtering (4 tests)**:
- `should display UID for each user`
- `should filter users by status`
- `should view user status details`
- `should update user status`

**Reason**: Missing GUI implementations

**Recommendation**: Implement in v0.7.5 (Complete Half-Finished GUI Features)

---

### Invitation Workflows (invitation-workflows.spec.ts)

**Total Tests**: 28
**Skipped with test.skip**: 0
**Status**: Unknown (daemon won't start for E2E tests)

**Reason**: Backend invitation APIs not implemented

**Recommendation**: Implement in v0.8.0 (Authentication & Invitations)

---

## Milestone Assignment

### v0.7.4 - Critical Fixes & Test Activation (Current)

**Scope**:
1. ✅ Document E2E test status (DONE)
2. ✅ Fix daemon startup for E2E tests (DONE - Jan 26, 2026)
3. ⏳ Verify active tests pass
4. ⏳ Fix test infrastructure issues

**Issues**:
- #315: This epic (investigation complete)
- #354: Cloudscape/Playwright compatibility (DONE - ConfirmDialog.ts fixed)
- Daemon startup FIXED - See `/tmp/daemon_startup_fix_summary.md`

---

### v0.7.5 - Complete Half-Finished GUI Features

**Scope**: Implement 9 missing user features to activate 9 skipped tests

**User Features Needed**:

1. **SSH Key Display** (#346 - User Detail View)
   - SSH key section in user detail
   - Key fingerprint display
   - Download/copy functionality
   - **Activates**: 1 test

2. **User Provisioning** (#347 - User Provisioning on Workspaces)
   - Provisioning dialog
   - Workspace selection
   - Provisioned workspaces list
   - Delete warnings
   - **Activates**: 3 tests

3. **User Statistics** (#348 - User Status Management)
   - Statistics cards
   - Per-user usage tracking
   - **Activates**: 1 test

4. **UID Display** (#349 - Edit User Functionality)
   - UID column in user table
   - UID in user details
   - **Activates**: 1 test

5. **User Filtering & Status** (#348 - User Status Management)
   - Status filter dropdown
   - Status detail view
   - Status update functionality
   - **Activates**: 3 tests

**Already Assigned to v0.7.5**:
- #332: Manage Members action
- #333-339: Project/User management (7 issues)
- #346-349: User management features (4 issues)

**Total v0.7.5 Scope**: 13 existing issues + 9 test activations

---

### v0.7.7 - Backlog Cleanup & Decisions

**Scope**: Decide fate of 6 budget-related tests

**Options**:

**Option A: Remove Tests** (Recommended)
- Budget feature was permanently removed
- Tests are obsolete
- Clean up test file
- Document removal decision

**Option B: Keep Skipped, Move to v0.11.0**
- Tag for future budget implementation
- Assign to v0.11.0 milestone
- Keep code but don't maintain

**Option C: Re-implement Budget Now**
- Violates "no new features" constraint
- Out of scope for v0.7.x
- Defer to v0.11.0

**Recommendation**: Option A - Remove in v0.7.7

---

### v0.8.0 - Authentication & Invitations

**Scope**: Implement invitation system (28 tests)

**Backend Work**:
- Invitation APIs (#308)
- Database schema for invitations
- Shared token system

**Frontend Work**:
- Individual invitations UI (#311)
- Bulk invitations enhancement (#312)
- Shared token management (#313)

**Already Assigned to v0.8.0**:
- #310-313: Invitation system issues

---

## Implementation Roadmap

### Phase 1: v0.7.4 (Feb 15, 2026) - Test Infrastructure

**Goals**:
- Fix daemon startup for E2E tests
- Verify 9 active project tests pass
- Document test infrastructure issues
- Create foundation for v0.7.5 work

**Deliverables**:
- Working E2E test environment
- Test status report
- Documented blockers

---

### Phase 2: v0.7.5 (Mar 15, 2026) - GUI Completion

**Goals**:
- Implement 9 missing user features
- Activate 9 skipped user tests
- Achieve GUI feature parity with CLI/TUI

**Deliverables**:
- All user management features complete
- 9 previously-skipped tests passing
- Zero skipped tests in user-workflows.spec.ts

---

### Phase 3: v0.7.7 (May 1, 2026) - Budget Test Resolution

**Goals**:
- Make decision on 6 budget tests
- Clean up test file
- Document decision

**Deliverables**:
- Budget tests removed OR clearly marked for v0.11.0
- Clean test files
- Documentation of decision

---

### Phase 4: v0.8.0 (Jun 1, 2026) - Invitation System

**Goals**:
- Implement complete invitation system
- Activate all 28 invitation tests
- Achieve 100% E2E test coverage

**Deliverables**:
- Backend invitation APIs working
- Frontend invitation UI complete
- All 28 invitation tests passing

---

## Success Metrics

### v0.7.4 Complete
- ✅ E2E test status documented
- ✅ Daemon startup fixed
- ✅ Active tests verified passing
- ✅ Test infrastructure issues resolved

### v0.7.5 Complete
- ✅ All 9 user features implemented
- ✅ 9 skipped user tests activated
- ✅ GUI feature parity achieved
- ✅ user-workflows.spec.ts: 0 skipped tests

### v0.7.7 Complete
- ✅ Budget test decision executed
- ✅ Test files cleaned up
- ✅ Backlog <20 issues

### v0.8.0 Complete
- ✅ Invitation system implemented
- ✅ All 28 invitation tests passing
- ✅ 100% E2E test activation (61 tests)

---

## Current Blockers

### 1. ~~Daemon Startup Failure~~ ✅ FIXED (Jan 26, 2026)
**Problem**: ~~E2E tests fail with "Daemon failed to start within 30 seconds"~~
**Solution**: Fixed PID file cleanup (was deleting `daemon.pid` instead of `prismd.pid`), added registry cleanup, and increased timeout
**Files Fixed**: `tests/e2e/global-setup.js`, `tests/e2e/setup-daemon.js`
**Test Result**: Daemon now starts in 1 second, basic.spec.ts passes (3/3 tests)
**Status**: ✅ RESOLVED

### 2. Missing User Features
**Problem**: 9 user tests skipped due to missing UI
**Impact**: Cannot activate user tests
**Priority**: MEDIUM
**Assignee**: v0.7.5

### 3. No Invitation Backend
**Problem**: 28 invitation tests cannot run
**Impact**: Cannot test invitation workflows
**Priority**: LOW (deferred to v0.8.0)
**Assignee**: v0.8.0

---

## Files Reference

**Test Files**:
- `cmd/prism-gui/frontend/tests/e2e/project-workflows.spec.ts` (15 tests, 6 skipped)
- `cmd/prism-gui/frontend/tests/e2e/user-workflows.spec.ts` (unknown total, 9 skipped)
- `cmd/prism-gui/frontend/tests/e2e/invitation-workflows.spec.ts` (28 tests, 0 skipped)

**Test Infrastructure**:
- `cmd/prism-gui/frontend/tests/e2e/setup-daemon.js` (needs debugging)
- `cmd/prism-gui/frontend/tests/e2e/pages/` (page objects)

**Backend**:
- Project APIs: Working
- User APIs: Mostly working (needs provisioning, SSH keys)
- Invitation APIs: Not implemented

---

## Next Immediate Steps

1. **Fix daemon startup** (v0.7.4 blocker)
2. **Run E2E tests** to verify active tests pass
3. **Generate test report** with pass/fail/skip counts
4. **Start v0.7.5 work** on user features

---

**Document Owner**: v0.7.4 Investigation
**Related Issues**: #315, #332-339, #346-349, #310-313
**Related Milestones**: v0.7.4, v0.7.5, v0.7.7, v0.8.0
