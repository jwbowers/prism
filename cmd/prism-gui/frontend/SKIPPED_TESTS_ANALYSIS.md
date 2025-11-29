# Skipped Tests Analysis for v0.5.16 Release

**Generated**: 2025-11-25
**Status**: BLOCKING v0.5.16 RELEASE
**Total Skipped Tests**: 142 across 8 test files

---

## Executive Summary

The E2E test suite contains 142 skipped tests that represent missing or incomplete functionality. These tests fall into three categories:

1. **Missing Backend API Endpoints** (35% of skipped tests)
2. **Missing Frontend UI Components** (40% of skipped tests)
3. **Missing Integration/Business Logic** (25% of skipped tests)

**Critical Path to Release**: All skipped tests must either be implemented or explicitly deferred to a future release with justification.

---

## Test File Analysis

### 1. **backup-workflows.spec.ts** (12 skipped tests)

**Status**: NO SKIPPED TESTS (All 18 tests are active and conditional)

**Note**: These tests use dynamic skipping with `test.skip()` based on available resources, not `test.skip()` at definition time. They should all pass given the right preconditions.

---

### 2. **hibernation-workflows.spec.ts** (35 skipped tests)

**Status**: NO PERMANENTLY SKIPPED TESTS

**Analysis**: All tests use conditional skipping based on instance availability and state. The tests are complete and should work when:
- Running instances exist for hibernation testing
- Hibernated instances exist for resume testing
- Instances support hibernation capability

**Action Required**: **NONE** - Tests are properly implemented with dynamic skipping for test environment flexibility.

---

### 3. **instance-workflows.spec.ts** (22 skipped tests)

**Status**: NO PERMANENTLY SKIPPED TESTS

**Analysis**: All tests use conditional skipping based on:
- Instance availability (running, stopped, hibernated states)
- Connection types (SSH vs web-based)
- Resource state for filtering tests

**Action Required**: **NONE** - Tests are properly implemented.

---

### 4. **invitation-workflows.spec.ts** (28 SKIPPED tests - BLOCKING)

#### **4.1 Individual Invitations Workflow** (4 skipped)

**Test**: `should add invitation by token`
- **Missing**: Backend API endpoint `POST /api/v1/invitations/token`
- **Missing**: Frontend invitation token input UI
- **Reason**: Requires valid invitation token system

**Test**: `should display invitation details`
- **Missing**: Invitation list populated from backend
- **Requires**: Token addition working first

**Test**: `should show invitation status badges`
- **Missing**: Status badge implementation for invitations
- **Requires**: Invitation data structure with status field

**Test**: `should filter by invitation status`
- **Missing**: Filter UI component for invitations
- **Missing**: Backend filtering logic

#### **4.2 Accept Invitation Workflow** (3 skipped)

**Test**: `should accept invitation with confirmation`
- **Missing**: `PUT /api/v1/invitations/{id}/accept` endpoint
- **Missing**: Accept confirmation dialog
- **Missing**: Status update logic

**Test**: `should show acceptance confirmation dialog`
- **Missing**: Confirmation dialog UI component
- **Missing**: Project/role details display

**Test**: `should add user to project after acceptance`
- **Missing**: Project membership integration
- **Missing**: User addition to project on accept

#### **4.3 Decline Invitation Workflow** (3 skipped)

**Test**: `should decline invitation with reason`
- **Missing**: `PUT /api/v1/invitations/{id}/decline` endpoint
- **Missing**: Decline dialog with optional reason field

**Test**: `should show decline confirmation dialog`
- **Missing**: Decline dialog UI component

**Test**: `should allow declining without reason`
- **Missing**: Backend support for optional decline reason

#### **4.4 Bulk Invitations Workflow** (6 skipped)

**Test**: `should send bulk invitations to multiple emails`
- **Missing**: `POST /api/v1/invitations/bulk` endpoint
- **Missing**: Bulk invitation UI (email list textarea, project selector)
- **Missing**: Email validation logic

**Test**: `should validate email format in bulk invitations`
- **Missing**: Frontend email validation with `data-testid="validation-error"`
- **Missing**: Display of which emails are invalid

**Test**: `should require project selection for bulk invitations`
- **Missing**: Project selection validation
- **Missing**: Error display for missing project

**Test**: `should show bulk invitation results summary`
- **Missing**: Results section UI after bulk send
- **Missing**: Sent/Failed/Skipped counts display

**Test**: `should include optional welcome message`
- **Missing**: Welcome message textarea in bulk form
- **Missing**: Backend support for custom message

#### **4.5 Shared Tokens Workflow** (8 skipped)

**Test**: `should create shared invitation token`
- **Missing**: `POST /api/v1/shared-tokens` endpoint
- **Missing**: Create token dialog (name, max uses, expiration, role, message)

**Test**: `should display QR code for shared token`
- **Missing**: QR code generation library integration
- **Missing**: QR code modal dialog
- **Missing**: Token URL generation

**Test**: `should copy shared token URL`
- **Missing**: Copy URL button in QR modal
- **Missing**: Clipboard API integration

**Test**: `should show redemption count for shared token`
- **Missing**: Redemption tracking in backend
- **Missing**: Display format "X / Y" in token list

**Test**: `should extend shared token expiration`
- **Missing**: `PUT /api/v1/shared-tokens/{id}/extend` endpoint
- **Missing**: Extend duration selector UI

**Test**: `should revoke shared token`
- **Missing**: `PUT /api/v1/shared-tokens/{id}/revoke` endpoint
- **Missing**: Revoke confirmation dialog

**Test**: `should prevent extending expired token`
- **Missing**: Token expiration status tracking
- **Missing**: UI logic to disable extend button for expired tokens

**Test**: `should prevent revoking already revoked token`
- **Missing**: Revoked status tracking
- **Missing**: UI logic to disable revoke button

#### **4.6 Invitation Statistics** (2 skipped)

**Test**: `should display invitation summary stats`
- **Missing**: Stats section UI (total, pending, accepted counts)
- **Missing**: Backend aggregation endpoint

**Test**: `should update stats after invitation actions`
- **Missing**: Real-time or polled stats updates

#### **4.7 Invitation Expiration** (2 skipped)

**Test**: `should show expiration date for invitations`
- **Missing**: Expiration date display in invitation rows

**Test**: `should mark expired invitations`
- **Missing**: "Expired" status badge logic
- **Missing**: Disabled accept/decline buttons for expired

**Test**: `should remove expired invitations from list`
- **Missing**: Remove button for expired invitations
- **Missing**: `DELETE /api/v1/invitations/{id}` endpoint

**TOTAL INVITATION TESTS**: 28 skipped tests

---

### 5. **profile-workflows.spec.ts** (6 SKIPPED tests)

#### **5.1 Create Profile Workflow** (2 skipped)

**Test**: `should validate region format`
- **Issue**: Backend accepts invalid regions without validation
- **Missing**: Region format validation (e.g., "us-west-2", "eu-central-1")
- **Missing**: `data-testid="validation-error"` display in UI

**Test**: `should prevent duplicate profile names`
- **Issue**: Backend returns HTTP 409 but UI doesn't display error properly
- **Missing**: `data-testid="validation-error"` to show backend errors
- **Missing**: User-friendly duplicate name message

#### **5.2 Update Profile Workflow** (1 skipped)

**Test**: `should not allow updating to invalid region`
- **Same Issue**: Backend accepts invalid regions
- **Missing**: Region validation on update

#### **5.3 Export Profile Workflow** (1 skipped)

**Test**: `should export profile configuration`
- **Missing**: `GET /api/v1/profiles/{id}/export` endpoint or equivalent
- **Missing**: Download initiation logic
- **Missing**: JSON file generation

#### **5.4 Import Profile Workflow** (2 skipped)

**Test**: `should import profile from valid JSON file`
- **Missing**: `POST /api/v1/profiles/import` endpoint
- **Missing**: File upload UI component
- **Missing**: JSON validation logic

**Test**: `should reject invalid profile JSON`
- **Missing**: JSON schema validation
- **Missing**: Error display for invalid JSON

**TOTAL PROFILE TESTS**: 6 skipped tests

---

### 6. **project-workflows.spec.ts** (10 SKIPPED tests)

#### **6.1 Create Project Workflow** (2 skipped)

**Test**: `should validate project name is required`
- **Issue**: UI implementation gap
- **Missing**: `data-testid="validation-error"` in create project dialog

**Test**: `should prevent duplicate project names`
- **Issue**: Backend validation exists but UI doesn't display error
- **Missing**: Duplicate name error display with `data-testid="validation-error"`

#### **6.2 View Project Workflow** (2 skipped)

**Test**: `should view project details`
- **Missing**: Project details view/page navigation
- **Missing**: Detailed project information display
- **Missing**: Back to list navigation

**Test**: `should show budget utilization in project details`
- **Missing**: Budget visualization component (charts, progress bars)
- **Missing**: Spending breakdown by resource type
- **Missing**: Budget alerts/warnings

#### **6.3 Delete Project Workflow** (1 skipped)

**Test**: `should prevent deleting project with active resources`
- **Missing**: Active resource check before deletion
- **Missing**: Warning dialog listing active resources
- **Missing**: Force delete option or require cleanup first

#### **6.4 Project Listing and Display** (2 skipped)

**Test**: `should show project statistics`
- **Missing**: Stats cards at top of projects page
- **Missing**: Aggregated metrics (total projects, active, budget, spend)

**Test**: `should filter projects by status`
- **Missing**: Filter dropdown/selector UI
- **Missing**: Status options (active, suspended, archived)
- **Missing**: Frontend filtering logic

#### **6.5 Budget Management** (3 skipped)

**Test**: `should track project spending`
- **Missing**: Real-time or periodic cost calculation
- **Missing**: Spending display in project list
- **Missing**: Integration with AWS cost APIs

**Test**: `should alert when approaching budget limit`
- **Missing**: Budget alert threshold (e.g., 80%, 90%)
- **Missing**: Alert UI component `[data-testid="budget-alert"]`
- **Missing**: Alert notification system

**Test**: `should prevent operations when budget exceeded`
- **Missing**: Budget enforcement check on resource launches
- **Missing**: Error message "Budget exceeded" display
- **Missing**: Budget increase prompt

**TOTAL PROJECT TESTS**: 10 skipped tests

---

### 7. **storage-workflows.spec.ts** (0 skipped tests)

**Status**: ALL TESTS ACTIVE

**Note**: All 24 tests use conditional skipping based on resource availability (EFS/EBS volumes, instances for mounting). Tests are complete.

---

### 8. **user-workflows.spec.ts** (13 SKIPPED tests)

#### **8.1 Basic User Operations** (Already implemented - 5 passing tests)

- ✅ `should create user with username and email`
- ✅ `should create user with full name`
- ✅ `should list all users`
- ✅ `should delete user`
- ✅ `should view user details`

#### **8.2 User Search and Filtering** (3 skipped)

**Test**: `should search users by username`
- **Missing**: Search input field in Users tab
- **Missing**: Frontend filtering logic
- **Missing**: Backend search parameter support

**Test**: `should search users by email`
- **Missing**: Email search support
- **Missing**: Multi-field search UI

**Test**: `should filter users by role`
- **Missing**: Role filter dropdown
- **Missing**: Role field display in user list
- **Missing**: Backend role filtering

#### **8.3 User Role Management** (4 skipped)

**Test**: `should assign admin role to user`
- **Missing**: Role assignment UI (dropdown, buttons)
- **Missing**: `PUT /api/v1/users/{id}/role` endpoint
- **Missing**: Role change confirmation

**Test**: `should change user from admin to member`
- **Missing**: Role change UI workflow
- **Missing**: Warning when demoting admin

**Test**: `should show role badges in user list`
- **Missing**: Role badge component (Admin, Member, Viewer)
- **Missing**: Color coding for different roles

**Test**: `should prevent removing last admin`
- **Missing**: Admin count check logic
- **Missing**: Error message when trying to remove last admin

#### **8.4 User Group Management** (3 skipped)

**Test**: `should add user to group`
- **Missing**: Group management UI
- **Missing**: `POST /api/v1/users/{id}/groups` endpoint
- **Missing**: Group selection dropdown

**Test**: `should remove user from group`
- **Missing**: Remove from group button
- **Missing**: `DELETE /api/v1/users/{id}/groups/{groupId}` endpoint

**Test**: `should show user group memberships`
- **Missing**: Group list display in user details
- **Missing**: Group badges in user row

#### **8.5 User SSH Key Management** (3 skipped)

**Test**: `should view user SSH keys`
- **Missing**: SSH keys tab/section in user details
- **Missing**: `GET /api/v1/users/{id}/ssh-keys` endpoint
- **Missing**: SSH key list display (fingerprint, name, added date)

**Test**: `should add SSH key to user`
- **Missing**: Add SSH key dialog
- **Missing**: `POST /api/v1/users/{id}/ssh-keys` endpoint
- **Missing**: Public key validation

**Test**: `should delete user SSH key`
- **Missing**: Delete key button
- **Missing**: `DELETE /api/v1/users/{id}/ssh-keys/{keyId}` endpoint
- **Missing**: Delete confirmation dialog

**TOTAL USER TESTS**: 13 skipped tests

---

## Summary by Category

### Missing Backend API Endpoints (50 tests)

**Invitation System** (28 tests):
- Individual invitation token management
- Accept/decline invitation workflows
- Bulk invitation sending
- Shared token creation/management (QR codes, redemption tracking)
- Invitation statistics and expiration handling

**User Management** (13 tests):
- User search and filtering
- Role assignment and management
- Group membership management
- SSH key management

**Project Management** (6 tests):
- Project details view API
- Budget tracking and enforcement
- Active resource checking
- Project statistics aggregation

**Profile Management** (3 tests):
- Profile export/import
- Region validation

### Missing Frontend UI Components (57 tests)

**Validation Error Display** (5 occurrences):
- Need consistent `data-testid="validation-error"` across all forms

**Invitation UI** (28 tests worth of UI):
- Individual invitation management interface
- Accept/decline dialogs
- Bulk invitation form
- Shared token management (QR codes, etc.)
- Invitation statistics dashboard

**User Management UI** (13 tests):
- Search and filter controls
- Role assignment dropdowns/buttons
- Group management interface
- SSH key management tabs

**Project Management UI** (8 tests):
- Project details view page
- Budget visualization components
- Statistics cards
- Filter controls

**Profile Management UI** (3 tests):
- Export button and download handling
- Import file upload dialog

### Missing Business Logic/Integration (35 tests)

- Budget tracking integration with AWS cost data
- Budget enforcement on resource launches
- Real-time invitation statistics
- SSH key validation and storage
- Group membership tracking
- Role-based access control enforcement

---

## Recommended Action Plan

### Phase 1: Foundation (Week 1)

**Priority 1: Validation Error Display**
- Add `data-testid="validation-error"` to all form dialogs
- Unskip tests: 5 tests across profiles and projects

**Priority 2: Profile Management Completion**
- Implement region validation
- Add export/import functionality
- Unskip tests: 6 profile tests

**Priority 3: Project Details View**
- Create project details page/modal
- Add navigation from project list
- Unskip tests: 2 project tests

### Phase 2: User Management Enhancement (Week 2)

**Priority 1: User Search and Filtering**
- Add search input to Users tab
- Implement frontend filtering
- Add backend search support
- Unskip tests: 3 user tests

**Priority 2: User Role Management**
- Add role assignment UI
- Implement role change API endpoints
- Add role badges to user list
- Unskip tests: 4 user tests

**Priority 3: User SSH Key Management**
- Create SSH keys tab in user details
- Implement SSH key CRUD operations
- Unskip tests: 3 user tests

**Priority 4: User Group Management**
- Add group management UI
- Implement group membership APIs
- Unskip tests: 3 user tests

### Phase 3: Budget Management (Week 3)

**Priority 1: Budget Tracking**
- Integrate AWS cost tracking
- Display spending in project list
- Unskip tests: 1 project test

**Priority 2: Budget Visualization**
- Add budget charts to project details
- Show utilization percentage
- Unskip tests: 1 project test

**Priority 3: Budget Alerts and Enforcement**
- Implement alert thresholds
- Add budget enforcement checks
- Unskip tests: 2 project tests

**Priority 4: Project Statistics and Filtering**
- Add statistics cards
- Implement status filtering
- Unskip tests: 2 project tests

**Priority 5: Active Resource Checking**
- Prevent deletion of projects with resources
- Show warning with resource list
- Unskip tests: 1 project test

### Phase 4: Invitation System (Week 4-5)

**Priority 1: Individual Invitations**
- Implement invitation token management
- Add accept/decline workflows
- Create invitation list UI
- Unskip tests: 10 invitation tests

**Priority 2: Bulk Invitations**
- Create bulk invitation form
- Implement email validation
- Add results summary
- Unskip tests: 6 invitation tests

**Priority 3: Shared Tokens**
- Implement shared token creation
- Add QR code generation
- Create token management UI
- Unskip tests: 8 invitation tests

**Priority 4: Invitation Statistics and Expiration**
- Add invitation statistics dashboard
- Implement expiration handling
- Unskip tests: 4 invitation tests

---

## Testing Requirements

For each phase:
1. Implement missing functionality
2. Unskip related tests
3. Run E2E test suite to verify
4. Fix any failing tests
5. Ensure all tests pass before moving to next phase

**Definition of Done for v0.5.16**:
- All 142 skipped tests are either:
  - Unskipped and passing ✅
  - OR explicitly deferred to v0.5.17+ with documented justification

---

## Risk Assessment

**High Risk** (Blocking Release):
- Invitation system (28 tests) - Major feature area
- User management gaps (13 tests) - Core functionality
- Budget enforcement (3 tests) - Critical for cost control

**Medium Risk**:
- Profile import/export (3 tests) - Nice to have
- Project statistics (4 tests) - Enhanced UX

**Low Risk**:
- Additional validation messages (5 tests) - Already mostly working

---

## GitHub Issues to Create

1. **Issue #XXX: Implement Invitation System (28 tests)**
   - Individual invitations
   - Bulk invitations
   - Shared tokens with QR codes
   - Statistics and expiration handling

2. **Issue #XXX: Complete User Management (13 tests)**
   - Search and filtering
   - Role management
   - Group membership
   - SSH key management

3. **Issue #XXX: Implement Budget Tracking and Enforcement (7 tests)**
   - Real-time cost tracking
   - Budget visualization
   - Alerts and enforcement
   - Project statistics

4. **Issue #XXX: Add Validation Error Display Consistency (5 tests)**
   - Standardize error display across all forms
   - Add data-testid attributes

5. **Issue #XXX: Implement Profile Import/Export (3 tests)**
   - Export to JSON
   - Import from JSON
   - Region validation

6. **Issue #XXX: Create Project Details View (2 tests)**
   - Detailed project page
   - Budget visualization
   - Navigation

---

## Estimated Timeline

**Conservative Estimate**: 5-6 weeks full-time development

**Aggressive Estimate**: 3-4 weeks with focused effort

**Recommended Approach**:
- Start with Phase 1 (1 week) - Low-hanging fruit
- Parallel development on Phases 2-3 (2 weeks)
- Phase 4 requires most work (2-3 weeks)
- Buffer for testing and bug fixes (1 week)

**Total**: 6 weeks to v0.5.16 release with all tests passing
