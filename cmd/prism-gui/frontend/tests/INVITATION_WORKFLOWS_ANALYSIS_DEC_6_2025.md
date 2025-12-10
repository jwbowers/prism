# Invitation Workflows Test Analysis - December 6, 2025

## Test Summary

**Total Tests**: 28 tests  
**Passed**: 11 tests (39%)  
**Failed**: 17 tests (61%)  
**Duration**: 7.9 minutes

## Failure Categories

### Category 1: Missing Backend Endpoints (Shared Tokens)
**Tests Failing**: 7 tests (all "Shared Tokens Workflow" tests)

All shared token tests fail with the same error:
```
TimeoutError: locator.waitFor: Timeout 5000ms exceeded.
Call log:
  - waiting for locator('[data-testid="create-shared-token-modal"]') to be visible
```

**Root Cause**: The shared token UI/API is not implemented. Tests try to:
1. Create shared invitation tokens
2. Display QR codes for tokens
3. Copy token URLs
4. Show redemption counts
5. Extend token expiration
6. Revoke tokens

**Failed Tests**:
- `should create shared invitation token` (line 473)
- `should display QR code for shared token` (line 486)
- `should copy shared token URL` (line 514)
- `should show redemption count for shared token` (line 536)
- `should extend shared token expiration` (line 549)
- `should revoke shared token` (line 580)
- `should prevent revoking already revoked token` (line 619)

**Backend Implementation Needed**:
- API endpoints for creating/managing shared tokens
- QR code generation
- Token redemption tracking
- Token expiration/revocation logic

---

### Category 2: UI Element Selector Mismatches (Bulk Invitations)
**Tests Failing**: 4 tests

Error pattern:
```
Error: locator.selectOption: Error: Element is not a <select> element
```

Tests are trying to use `selectOption()` on an element that's not a `<select>` dropdown, likely a Cloudscape component.

**Failed Tests**:
- `should send bulk invitations to multiple emails` (line 355)
- `should validate email format in bulk invitations` (line 377)
- `should show bulk invitation results summary` (line 431)
- `should include optional welcome message` (line 452)

**Root Cause**: Frontend uses Cloudscape UI components (not standard HTML `<select>`), but tests expect standard select dropdowns.

**Fix Needed**: Update test selectors to work with Cloudscape Select component API.

---

### Category 3: Invitation Interaction Failures (Accept/Decline)
**Tests Failing**: 4 tests

Pattern 1 - TimeoutError waiting for invitation rows:
```
TimeoutError: locator.waitFor: Timeout 10000ms exceeded.
Call log:
  - waiting for locator('tr:has-text("Project Name")').first() to be visible
```

Pattern 2 - Expectation failures (UI elements not showing as expected):
```
Error: expect(received).toBe(expected)
Expected: true
Received: false
```

**Failed Tests**:
- `should add user to project after acceptance` (line 209)
- `should show decline confirmation dialog` (line 298)
- `should allow declining without reason` (line 329)
- `should update stats after invitation actions` (line 653)

**Root Cause**: Either:
1. Invitations not being created in the backend
2. UI not displaying invitation rows correctly
3. Race condition between invitation creation and UI updates

---

### Category 4: Invitation Display/Expiration
**Tests Failing**: 2 tests

**Failed Tests**:
- `should display invitation details` (line 74)
- `should show expiration date for invitations` (line 681)

**Root Cause**: Similar to Category 3 - invitations not appearing in UI after creation.

---

## Tests That ARE Passing (11 tests)

These tests prove that SOME invitation functionality works:

1. **Basic Invitation UI**:
   - Modal opens when "Add Invitation" clicked
   - Cancel button works
   - Form displays correctly

2. **Status Badge Display**:
   - Pending invitations show correct status
   - Accepted invitations show correct status

3. **Invitation Filtering**:
   - Status filtering works (All, Pending, Accepted, Declined)

4. **Invitation List**:
   - List displays when navigating to Invitations tab

---

## Root Cause Analysis

### Primary Issues:

1. **Shared Tokens Feature Not Implemented** (7 tests)
   - No backend API
   - No frontend UI
   - No modal dialog

2. **Frontend-Backend Mismatch** (4 tests - bulk invitations)
   - Tests written for standard HTML selects
   - Frontend uses Cloudscape components

3. **Invitation Creation/Display Pipeline Broken** (6 tests)
   - Invitations may not be persisting to backend
   - UI not refreshing after invitation actions
   - Race conditions in test expectations

---

## Recommended Fix Priority

### HIGH Priority - Quick Wins:

**A. Fix Bulk Invitation Selectors** (4 tests)
- Impact: Medium (improves 4 tests)
- Effort: Low (update test selectors)
- File: `cmd/prism-gui/frontend/tests/e2e/invitation-workflows.spec.ts`
- Action: Replace `selectOption()` with Cloudscape Select component interaction

**B. Investigate Invitation Creation Pipeline** (6 tests)
- Impact: High (fixes 6 tests)
- Effort: Medium (debug backend API + frontend state)
- Files to check:
  - Backend: Invitation API endpoints
  - Frontend: `SafePrismAPI` invitation methods
  - Test: `ProjectsPage.sendTestInvitation()`

### MEDIUM Priority:

**C. Implement Shared Tokens Feature** (7 tests)
- Impact: High (enables 7 tests)
- Effort: HIGH (full feature implementation)
- Components needed:
  - Backend API (`/api/v1/invitations/tokens` endpoints)
  - Frontend UI (modal, QR code display, token management)
  - QR code generation library
  - Token redemption tracking
- Decision: This may be a v0.6.x feature, not v0.5.x

---

## Next Steps

1. **Immediate**: Fix bulk invitation selectors (HIGH-A)
2. **Short-term**: Debug invitation creation pipeline (HIGH-B)
3. **Long-term**: Decide if shared tokens should be implemented now or deferred

## Files to Investigate

### Backend:
- `pkg/daemon/*_handlers.go` - Invitation API endpoints
- Check if invitation endpoints exist and work correctly

### Frontend:
- `cmd/prism-gui/frontend/src/App.tsx` - SafePrismAPI invitation methods
- Check if invitation API calls match backend expectations

### Tests:
- `cmd/prism-gui/frontend/tests/e2e/invitation-workflows.spec.ts` - Test implementation
- `cmd/prism-gui/frontend/tests/e2e/pages/ProjectsPage.ts` - Test helper methods

---

## Testing Verification

After fixes, re-run:
```bash
cd cmd/prism-gui/frontend
npx playwright test tests/e2e/invitation-workflows.spec.ts --project=chromium
```

Expected improvement:
- Fix A (selectors): +4 tests passing (15/28 = 54%)
- Fix B (creation): +6 tests passing (21/28 = 75%)
- Total without shared tokens: 21/28 passing (75%)
- With shared tokens: 28/28 passing (100%)
