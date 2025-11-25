# GUI Implementation Roadmap: Projects/Users/Invitations E2E Test Activation

This document provides a step-by-step guide to implement the missing GUI functionality needed to activate all 51 skipped E2E tests for Projects, Users, and Invitations.

## Overview

- **Total E2E Tests Created**: 61
- **Currently Active**: 10
- **Skipped (Need Implementation)**: 51

## Phase 1: Add Missing Test IDs (Quick Win - 2-4 hours)

### Projects Tab (App.tsx - ProjectManagementView)

**Location**: Around line 3491

#### Required Test IDs:

```typescript
// Projects Table
<Table
  data-testid="projects-table"
  columnDefinitions={projectColumns}
  items={state.projects}
  // ... existing props
/>

// Create Project Button
<Button
  variant="primary"
  onClick={() => setProjectModalVisible(true)}
  data-testid="create-project-button"
>
  Create Project
</Button>

// Project Actions Dropdown (for each project row)
<ButtonDropdown
  items={[
    { text: "View Details", id: "view" },
    { text: "Manage Members", id: "members" },
    { text: "Edit Project", id: "edit" },
    { text: "Delete", id: "delete" },
  ]}
  data-testid={`project-actions-${project.id}`}
/>

// Delete Confirmation Modal
<Modal
  visible={deleteProjectModalVisible}
  data-testid="delete-project-modal"
>
  <Button onClick={handleDeleteProject} data-testid="confirm-delete-project">
    Delete
  </Button>
  <Button onClick={() => setDeleteProjectModalVisible(false)}>
    Cancel
  </Button>
</Modal>
```

### Users Tab (App.tsx - UserManagementView)

**Location**: Around line 4145

#### Required Test IDs:

```typescript
// Users Table
<Table
  data-testid="users-table"
  columnDefinitions={userColumns}
  items={state.users}
  // ... existing props
/>

// Create User Button
<Button
  variant="primary"
  onClick={() => setUserModalVisible(true)}
  data-testid="create-user-button"
>
  Create User
</Button>

// User Actions Dropdown
<ButtonDropdown
  items={[
    { text: "View Details", id: "view" },
    { text: "Generate SSH Key", id: "ssh-key" },
    { text: "Provision on Workspace", id: "provision" },
    { text: "Edit User", id: "edit" },
    { text: "Delete", id: "delete" },
  ]}
  data-testid={`user-actions-${user.username}`}
/>
```

### Invitations Tab (App.tsx - InvitationView)

**Location**: Around line 4414

#### Required Test IDs:

```typescript
// Invitations Table
<Table
  data-testid="invitations-table"
  columnDefinitions={invitationColumns}
  items={state.invitations}
  // ... existing props
/>

// Tab Selection
<Tabs
  tabs={[
    { label: "Individual", id: "individual" },
    { label: "Bulk Invitations", id: "bulk" },
    { label: "Shared Tokens", id: "shared" },
  ]}
  activeTabId={invitationTab}
/>

// Bulk Invite Project Select (ALREADY EXISTS)
<Select
  data-testid="bulk-invite-project-select"
  // ... existing props
/>
```

## Phase 2: Implement Validation Error Display (4-6 hours)

### Create Reusable Validation Error Component

**New File**: `src/components/ValidationError.tsx`

```typescript
import { Alert } from '@cloudscape-design/components';

interface ValidationErrorProps {
  message: string;
  visible: boolean;
}

export const ValidationError: React.FC<ValidationErrorProps> = ({ message, visible }) => {
  if (!visible) return null;

  return (
    <Alert
      type="error"
      data-testid="validation-error"
    >
      {message}
    </Alert>
  );
};
```

### Add Validation to Forms

#### Project Creation Form:

```typescript
const [validationError, setValidationError] = useState<string>('');

const handleCreateProject = async () => {
  // Validate
  if (!projectName.trim()) {
    setValidationError('Project name is required');
    return;
  }

  if (projectBudget && projectBudget < 0) {
    setValidationError('Budget must be a positive number');
    return;
  }

  try {
    await api.createProject({
      name: projectName,
      description: projectDescription,
      budget_limit: projectBudget,
    });
    setValidationError('');
    setProjectModalVisible(false);
    await loadApplicationData();
  } catch (error: any) {
    // Display backend errors (e.g., duplicate name)
    if (error.response?.status === 409) {
      setValidationError('A project with this name already exists');
    } else {
      setValidationError(`Failed to create project: ${error.message}`);
    }
  }
};

// In the modal JSX
<Modal visible={projectModalVisible}>
  <ValidationError message={validationError} visible={!!validationError} />
  {/* ... form fields */}
</Modal>
```

#### User Creation Form:

```typescript
const handleCreateUser = async () => {
  // Validate
  if (!username.trim()) {
    setValidationError('Username is required');
    return;
  }

  if (!email.match(/^[^\s@]+@[^\s@]+\.[^\s@]+$/)) {
    setValidationError('Please enter a valid email address');
    return;
  }

  try {
    await api.createUser({ username, email, full_name: fullName });
    setValidationError('');
    setUserModalVisible(false);
    await loadApplicationData();
  } catch (error: any) {
    if (error.response?.status === 409) {
      setValidationError('A user with this username already exists');
    } else {
      setValidationError(`Failed to create user: ${error.message}`);
    }
  }
};
```

## Phase 3: Backend API Integration (8-12 hours)

### Verify/Implement Required API Endpoints

**File**: `cmd/prism-gui/frontend/src/App.tsx` (APIClient class)

#### Projects API (Likely Already Exists):

```typescript
// APIClient methods to verify/add:
async createProject(project: CreateProjectRequest): Promise<Project> {
  return this.safeRequest('/api/v1/projects', 'POST', project);
}

async updateProject(projectId: string, updates: Partial<Project>): Promise<Project> {
  return this.safeRequest(`/api/v1/projects/${projectId}`, 'PUT', updates);
}

async deleteProject(projectId: string): Promise<void> {
  return this.safeRequest(`/api/v1/projects/${projectId}`, 'DELETE');
}

async getProjectDetails(projectId: string): Promise<ProjectDetails> {
  return this.safeRequest(`/api/v1/projects/${projectId}`);
}
```

#### Users API (Likely Already Exists):

```typescript
async getUsers(): Promise<User[]> {
  return this.safeRequest('/api/v1/users');
}

async createUser(user: CreateUserRequest): Promise<User> {
  return this.safeRequest('/api/v1/users', 'POST', user);
}

async deleteUser(username: string): Promise<void> {
  return this.safeRequest(`/api/v1/users/${username}`, 'DELETE');
}

async generateSSHKey(username: string): Promise<SSHKeyResponse> {
  return this.safeRequest(`/api/v1/users/${username}/ssh-keys`, 'POST');
}

async provisionUserOnWorkspace(username: string, instanceId: string): Promise<void> {
  return this.safeRequest(`/api/v1/users/${username}/provision`, 'POST', { instance_id: instanceId });
}
```

#### Invitations API (NEEDS IMPLEMENTATION):

```typescript
// Individual Invitations
async getInvitationByToken(token: string): Promise<Invitation> {
  return this.safeRequest(`/api/v1/invitations/${token}`);
}

async acceptInvitation(token: string): Promise<void> {
  return this.safeRequest(`/api/v1/invitations/${token}/accept`, 'POST');
}

async declineInvitation(token: string, reason?: string): Promise<void> {
  return this.safeRequest(`/api/v1/invitations/${token}/decline`, 'POST', { reason });
}

// Bulk Invitations (ALREADY EXISTS - line 4574)
async bulkInvite(projectId: string, emails: string[], role: string, message?: string): Promise<BulkInviteResult> {
  return this.safeRequest(`/api/v1/projects/${projectId}/invitations/bulk`, 'POST', {
    emails,
    role,
    message,
  });
}

// Shared Tokens
async getSharedTokens(projectId: string): Promise<SharedToken[]> {
  return this.safeRequest(`/api/v1/projects/${projectId}/shared-tokens`);
}

async createSharedToken(projectId: string, config: CreateSharedTokenRequest): Promise<SharedToken> {
  return this.safeRequest(`/api/v1/projects/${projectId}/shared-tokens`, 'POST', config);
}

async extendSharedToken(projectId: string, tokenId: string, duration: string): Promise<void> {
  return this.safeRequest(`/api/v1/projects/${projectId}/shared-tokens/${tokenId}/extend`, 'POST', { duration });
}

async revokeSharedToken(projectId: string, tokenId: string): Promise<void> {
  return this.safeRequest(`/api/v1/projects/${projectId}/shared-tokens/${tokenId}/revoke`, 'POST');
}
```

### Backend Implementation Checklist

**File**: `pkg/daemon/project_handlers.go` and new files

#### Required Backend Handlers:

1. **Invitation System** (`pkg/daemon/invitation_handlers.go` - NEW):
   ```go
   func (d *Daemon) handleGetInvitation(w http.ResponseWriter, r *http.Request)
   func (d *Daemon) handleAcceptInvitation(w http.ResponseWriter, r *http.Request)
   func (d *Daemon) handleDeclineInvitation(w http.ResponseWriter, r *http.Request)
   func (d *Daemon) handleCreateSharedToken(w http.ResponseWriter, r *http.Request)
   func (d *Daemon) handleExtendSharedToken(w http.ResponseWriter, r *http.Request)
   func (d *Daemon) handleRevokeSharedToken(w http.ResponseWriter, r *http.Request)
   ```

2. **User Provisioning** (`pkg/daemon/user_handlers.go` - ENHANCE):
   ```go
   func (d *Daemon) handleGenerateSSHKey(w http.ResponseWriter, r *http.Request)
   func (d *Daemon) handleProvisionUser(w http.ResponseWriter, r *http.Request)
   ```

3. **Database Schema** (`pkg/database/schema.sql` or equivalent):
   ```sql
   CREATE TABLE invitations (
     token TEXT PRIMARY KEY,
     project_id TEXT NOT NULL,
     email TEXT NOT NULL,
     role TEXT NOT NULL,
     invited_by TEXT NOT NULL,
     message TEXT,
     status TEXT NOT NULL DEFAULT 'pending',
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     expires_at TIMESTAMP NOT NULL,
     FOREIGN KEY (project_id) REFERENCES projects(id)
   );

   CREATE TABLE shared_tokens (
     id TEXT PRIMARY KEY,
     project_id TEXT NOT NULL,
     name TEXT NOT NULL,
     role TEXT NOT NULL,
     redemption_limit INTEGER NOT NULL,
     redemption_count INTEGER DEFAULT 0,
     expires_at TIMESTAMP NOT NULL,
     status TEXT NOT NULL DEFAULT 'active',
     welcome_message TEXT,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     FOREIGN KEY (project_id) REFERENCES projects(id)
   );
   ```

## Phase 4: Feature Implementation Priority (16-24 hours total)

### High Priority (Enable Most Tests):

1. **Project CRUD with Validation** (4 hours)
   - ✅ Add test IDs to existing UI
   - ✅ Add validation error display
   - ✅ Wire up API calls with error handling
   - **Tests Activated**: 6-8 tests

2. **User CRUD with Validation** (4 hours)
   - ✅ Add test IDs to existing UI
   - ✅ Add validation error display
   - ✅ Wire up API calls with error handling
   - **Tests Activated**: 5-7 tests

3. **Project Detail View** (3 hours)
   - Create ProjectDetailView component
   - Add navigation from project list
   - Display project info, members, budget usage
   - **Tests Activated**: 2-3 tests

### Medium Priority:

4. **Invitation System - Individual** (6 hours)
   - Implement backend invitation API
   - Add invitation table with status badges
   - Implement accept/decline dialogs
   - **Tests Activated**: 10-12 tests

5. **Invitation System - Bulk** (4 hours)
   - Enhance existing bulk invite UI
   - Add result summary display
   - Add email validation
   - **Tests Activated**: 5-6 tests

6. **SSH Key Management** (4 hours)
   - Add SSH key generation dialog
   - Display SSH keys in user details
   - Implement key download/copy
   - **Tests Activated**: 2-3 tests

### Lower Priority:

7. **Shared Token System** (6 hours)
   - Create shared token UI
   - Implement QR code generation
   - Add extend/revoke functionality
   - **Tests Activated**: 8-10 tests

8. **User Provisioning** (5 hours)
   - Implement provisioning dialog
   - Add workspace selection
   - Track provisioned workspaces
   - **Tests Activated**: 2-3 tests

9. **Statistics and Filtering** (4 hours)
   - Add overview stats cards
   - Implement table filtering
   - Add status indicators
   - **Tests Activated**: 6-8 tests

## Phase 5: Test Activation Strategy

### Step-by-Step Activation:

1. **After Phase 1 (Test IDs)**:
   ```bash
   # Un-skip basic CRUD tests
   # In project-workflows.spec.ts, user-workflows.spec.ts:
   # Remove .skip from tests 1-6

   npx playwright test project-workflows.spec.ts user-workflows.spec.ts --project=chromium
   ```

2. **After Phase 2 (Validation)**:
   ```bash
   # Un-skip validation tests
   # Remove .skip from validation tests

   npx playwright test project-workflows.spec.ts:40 user-workflows.spec.ts:55 --project=chromium
   ```

3. **After Phase 3-4 (Backend + Features)**:
   ```bash
   # Un-skip feature-specific tests progressively
   # Test each feature group separately

   npx playwright test invitation-workflows.spec.ts:20 --project=chromium  # Individual invitations
   npx playwright test invitation-workflows.spec.ts:150 --project=chromium # Bulk invitations
   ```

4. **Final Verification**:
   ```bash
   # Run all tests together
   npx playwright test project-workflows.spec.ts user-workflows.spec.ts invitation-workflows.spec.ts --project=chromium
   ```

## Phase 6: Documentation Updates

1. **Update TESTING.md**:
   - Document how to run the new test suites
   - Add troubleshooting guide
   - Document test data requirements

2. **Update CLAUDE.md**:
   - Add Projects/Users/Invitations to testing section
   - Document E2E test patterns
   - Add examples of common test scenarios

3. **Create API Documentation**:
   - Document invitation system API
   - Add request/response examples
   - Document error codes

## Quick Start: Immediate Next Steps

### Option A: Quick Win (Today - 2 hours)

1. Add all test IDs from Phase 1
2. Run the 10 active tests to verify they pass
3. Commit and push

**Commands**:
```bash
cd /Users/scttfrdmn/src/prism/cmd/prism-gui/frontend
# Edit App.tsx to add test IDs
npx playwright test project-workflows.spec.ts user-workflows.spec.ts --project=chromium --grep-invert "@skip"
```

### Option B: Feature Complete (This Week - 16 hours)

1. Complete Phases 1-2 (Test IDs + Validation)
2. Implement Project/User CRUD (Phase 4, items 1-2)
3. Add Project Detail View (Phase 4, item 3)
4. Un-skip and run 20-25 tests

**Outcome**: ~40% of tests active (25/61)

### Option C: Full Implementation (Next 2 Weeks - 40 hours)

1. Complete all Phases 1-5
2. Full invitation system
3. All 61 tests active

**Outcome**: 100% test coverage for Projects/Users/Invitations

## Success Metrics

- [ ] All 10 currently-active tests passing
- [ ] At least 25 additional tests un-skipped and passing
- [ ] Zero TypeScript compilation errors
- [ ] All new UI components have data-testid attributes
- [ ] Backend APIs documented and tested
- [ ] E2E tests run in CI/CD pipeline

## Files to Modify

### Frontend:
- `cmd/prism-gui/frontend/src/App.tsx` (add test IDs, validation)
- `cmd/prism-gui/frontend/src/components/ValidationError.tsx` (NEW)
- `cmd/prism-gui/frontend/tests/e2e/project-workflows.spec.ts` (un-skip tests)
- `cmd/prism-gui/frontend/tests/e2e/user-workflows.spec.ts` (un-skip tests)
- `cmd/prism-gui/frontend/tests/e2e/invitation-workflows.spec.ts` (un-skip tests)

### Backend:
- `pkg/daemon/invitation_handlers.go` (NEW - invitation system)
- `pkg/daemon/user_handlers.go` (enhance with SSH key, provisioning)
- `pkg/daemon/project_handlers.go` (verify existing endpoints)
- `pkg/database/schema.sql` (add invitation tables)
- `pkg/invitation/` (NEW package for invitation logic)

### Documentation:
- `cmd/prism-gui/frontend/TESTING.md` (update with new tests)
- `docs/CLAUDE.md` (add E2E testing section)
- `docs/API.md` (NEW - document invitation API)

---

**Total Estimated Effort**: 40-50 hours for complete implementation
**Quick Win Path**: 2-4 hours for test IDs + initial validation
**ROI**: 61 comprehensive E2E tests covering critical user workflows
