# Phase 4.4: Individual Invitations System - Implementation Plan

**Date**: 2025-11-25
**Epic**: Issue #315 (E2E Test Activation)
**Milestone**: v0.5.16 (Projects & Users - Week 2)
**Complexity**: HIGH - New UI Component (not bug fix)

---

## Executive Summary

Phase 4.4 requires implementing the **InvitationManagementView** component from scratch. Unlike previous phases (SSH Keys, Project Detail) which were debugging existing components, this is a completely new feature requiring:

- ✅ Backend APIs ready (7 methods)
- ✅ Type definitions ready (6 interfaces)
- ✅ Test helpers ready (10+ methods)
- ❌ **UI Component missing** (needs creation)

**Scope**: ~350-500 lines of new React/TypeScript code

---

## Current State Analysis

### ✅ What Exists (Ready to Use)

**API Methods** (src/App.tsx lines 1030-1161):
```typescript
- getInvitationByToken(token): CachedInvitation
- acceptInvitation(token): void
- declineInvitation(token, reason?): void
- sendInvitation(projectId, email, role, message?, expiresAt?): Invitation
- sendBulkInvitations(projectId, emails, role, message?): BulkInvitationResponse
- createSharedToken(projectId, config): SharedInvitationToken
- getMyInvitations(): Invitation[]
- getProjectInvitations(projectId): Invitation[]
```

**Type Definitions** (src/App.tsx lines 339-489):
```typescript
interface CachedInvitation {
  token, invitation_id, project_id, project_name, email, role,
  invited_by, invited_at, expires_at, status, message?, added_at
}

interface Invitation {
  id, project_id, project_name, email, role, token, status,
  invited_by, invited_at, expires_at, responded_at?, message?
}

interface SharedInvitationToken {
  token, project_id, project_name, name, role, redemption_limit,
  redemptions, created_at, created_by, expires_at, revoked, qr_code_url?
}

interface BulkInvitationResponse {
  sent, failed, invitations: Invitation[], errors: {email, error}[]
}
```

**State Management** (src/App.tsx line 513):
```typescript
interface AppState {
  activeView: includes 'invitations'
  invitations: CachedInvitation[]
}
```

**Test Helpers** (tests/e2e/pages/ProjectsPage.ts lines 279-480):
```typescript
- navigateToInvitations()
- switchToIndividualInvitations()
- switchToBulkInvitations()
- switchToSharedTokens()
- getInvitationRows(): Locator
- addInvitationToken(token)
- acceptInvitation(projectName)
- declineInvitation(projectName, reason?)
- sendBulkInvitations(projectId, emails, role, message?)
- createSharedToken(name, limit, expires, role, message?)
- viewQRCode(tokenName)
- revokeSharedToken(tokenName)
- verifyInvitationExists(projectName): boolean
```

### ❌ What's Missing (Needs Implementation)

**InvitationManagementView Component** (NEW FILE):
- File: `src/components/InvitationManagementView.tsx`
- Size estimate: ~350-500 lines
- Requirements:
  1. Tab navigation (Individual / Bulk / Shared Tokens)
  2. Individual Invitations tab
  3. Bulk Invitations tab
  4. Shared Tokens tab
  5. All data-testid attributes for testing

---

## Component Architecture

### InvitationManagementView Structure

```
InvitationManagementView (Main Component)
├── Statistics Section (Header)
│   ├── Total Invitations count
│   ├── Pending count
│   ├── Accepted count
│   └── Expired count
├── Tabs Container
│   ├── Tab: Individual Invitations
│   │   ├── Add Token Input + Button
│   │   ├── Invitations Table
│   │   │   ├── Columns: Project, Role, Invited By, Expires, Status
│   │   │   └── Actions: Accept / Decline buttons
│   │   └── Empty State (no invitations)
│   ├── Tab: Bulk Invitations
│   │   ├── Project Selection dropdown
│   │   ├── Email Addresses textarea (multi-line)
│   │   ├── Role selection dropdown
│   │   ├── Optional Welcome Message textarea
│   │   ├── Send Bulk Invitations button
│   │   └── Results Display (sent/failed counts)
│   └── Tab: Shared Tokens
│       ├── Create Token Button
│       ├── Tokens Table
│       │   ├── Columns: Name, Role, Redemptions, Expires, Status
│       │   └── Actions: View QR, Extend, Revoke
│       └── QR Code Modal (view/copy token URL)
└── Modals
    ├── Accept Invitation Confirmation
    ├── Decline Invitation Confirmation
    ├── Create Shared Token Form
    └── QR Code Display
```

### Required data-testid Attributes

**Individual Invitations Tab**:
- `invitations-table` - Main table
- `invitation-token-input` - Token input field
- `add-invitation-button` - Add button
- `accept-invitation-button-{id}` - Accept button (per row)
- `decline-invitation-button-{id}` - Decline button (per row)

**Bulk Invitations Tab**:
- `bulk-project-select` - Project dropdown
- `bulk-emails-textarea` - Emails input
- `bulk-role-select` - Role dropdown
- `bulk-message-textarea` - Optional message
- `send-bulk-invitations-button` - Send button
- `bulk-results-container` - Results display

**Shared Tokens Tab**:
- `shared-tokens-table` - Main table
- `create-shared-token-button` - Create button
- `view-qr-button-{id}` - View QR button (per row)
- `extend-token-button-{id}` - Extend button (per row)
- `revoke-token-button-{id}` - Revoke button (per row)
- `qr-code-modal` - QR code modal
- `copy-token-url-button` - Copy URL button

---

## Implementation Steps

### Step 1: Create InvitationManagementView Component

**File**: `src/components/InvitationManagementView.tsx`

**Imports Required**:
```typescript
import React, { useState, useEffect } from 'react';
import {
  Container, Header, Tabs, Table, Box, SpaceBetween,
  Button, Input, Textarea, Select, Badge, StatusIndicator,
  Modal, Alert, Grid, ColumnLayout, FormField
} from '@cloudscape-design/components';
```

**Props Interface**:
```typescript
interface InvitationManagementViewProps {
  // No props needed - uses window.__apiClient directly
}
```

**State Management**:
```typescript
- activeTabId: 'individual' | 'bulk' | 'shared'
- invitations: CachedInvitation[]
- sharedTokens: SharedInvitationToken[]
- loading: boolean
- error: string | null
- selectedInvitation: CachedInvitation | null (for modals)
- tokenInput: string (for adding tokens)
- bulkForm: { projectId, emails, role, message }
- bulkResults: { sent, failed, errors }
- qrCodeModal: { visible, token }
```

### Step 2: Implement Individual Invitations Tab

**Table Columns**:
1. Project Name
2. Role (with color-coded badge)
3. Invited By
4. Expires (relative time)
5. Status (with StatusIndicator)
6. Actions (Accept/Decline buttons)

**Functionality**:
- Load invitations via `api.getMyInvitations()`
- Add invitation by token
- Accept invitation (show confirmation modal)
- Decline invitation (show confirmation modal with optional reason)
- Filter by status (pending/accepted/declined/expired)

### Step 3: Implement Bulk Invitations Tab

**Form Fields**:
1. Project Selection (dropdown of available projects)
2. Email Addresses (textarea, one per line)
3. Role Selection (viewer/member/admin dropdown)
4. Optional Welcome Message (textarea)

**Functionality**:
- Validate email format (client-side)
- Call `api.sendBulkInvitations()`
- Display results summary (sent/failed/errors)
- Clear form after successful send

### Step 4: Implement Shared Tokens Tab

**Table Columns**:
1. Token Name
2. Role
3. Redemptions (used/limit)
4. Created By
5. Expires
6. Status (active/expired/revoked)
7. Actions (View QR/Extend/Revoke)

**Functionality**:
- Create shared token (modal with form)
- View QR code (modal with QR image + copy URL button)
- Extend expiration (modal with duration selector)
- Revoke token (confirmation modal)

### Step 5: Wire into App.tsx

**Changes Required**:

1. Import component (add to imports):
```typescript
import { InvitationManagementView } from './components/InvitationManagementView';
```

2. Add render case (around line 9800):
```typescript
{state.activeView === 'invitations' && (
  <InvitationManagementView />
)}
```

3. Add navigation tab (in Tab definitions):
```typescript
{
  label: "Invitations",
  id: "invitations",
  content: null // Content rendered separately
}
```

### Step 6: Update Test Helpers (if needed)

Most test helpers already exist in ProjectsPage.ts. May need minor adjustments for actual data-testid values.

### Step 7: Unskip E2E Tests

**Individual Invitations Tests** (lines 28-84 in invitation-workflows.spec.ts):
- ❌ "should add invitation by token" (line 29)
- ❌ "should display invitation details" (line 43)
- ❌ "should show invitation status badges" (line 55)
- ❌ "should filter by invitation status" (line 66)

**Accept/Decline Tests** (lines 86-140):
- ❌ "should accept invitation with confirmation" (line 87)
- ❌ "should show acceptance confirmation dialog" (line 103)
- ❌ "should add user to project after acceptance" (line 123)
- ❌ "should decline invitation with reason" (line 143)
- ❌ "should show decline confirmation dialog" (line 160)
- ❌ "should allow declining without reason" (line 180)

**Initial Implementation Scope**: Unskip first 4 Individual Invitations tests

---

## Testing Strategy

### Unit Testing (Optional - Time Permitting)
- Component rendering tests
- Form validation tests
- API call mocking

### E2E Testing (Primary Focus)
1. Navigate to Invitations tab
2. Switch between Individual/Bulk/Shared tabs
3. Add invitation by token
4. Display invitation details
5. Accept/decline invitations
6. Send bulk invitations
7. Create/manage shared tokens

---

## Implementation Order (Priority)

1. **✅ Analysis & Planning** (Complete)
2. **⏭️ Individual Invitations Tab** (Core functionality)
   - Create component shell
   - Implement table display
   - Add token input
   - Accept/Decline buttons
3. **⏭️ Wire into App.tsx** (Enable navigation)
4. **⏭️ Unskip First 4 Tests** (Validate core functionality)
5. **⏭️ Bulk Invitations Tab** (Secondary feature)
6. **⏭️ Shared Tokens Tab** (Advanced feature)

---

## Estimated Effort

- **Component Creation**: 2-3 hours
- **Individual Tab**: 1.5 hours
- **Bulk Tab**: 1 hour
- **Shared Tokens Tab**: 1.5 hours
- **Testing & Debugging**: 1-2 hours
- **Documentation**: 0.5 hours

**Total**: ~7-9 hours for complete implementation

**Phase 4.4 Initial Scope**: Focus on Individual Invitations tab (3-4 hours)

---

## Risk Assessment

**Technical Risks**:
- ✅ LOW: Backend APIs already exist and tested
- ✅ LOW: Type definitions complete
- ⚠️ MEDIUM: Complex UI with multiple tabs and modals
- ⚠️ MEDIUM: E2E tests may need API mocking or real invitation creation

**Mitigation**:
- Follow ProjectDetailView pattern (existing reference implementation)
- Use Cloudscape Design System components (consistent styling)
- Implement incrementally (Individual → Bulk → Shared)
- Test each tab independently before integration

---

## Success Criteria

**Phase 4.4 Complete When**:
1. ✅ InvitationManagementView component created
2. ✅ Individual Invitations tab functional
3. ✅ Navigation wired into App.tsx
4. ✅ First 4 E2E tests passing
5. ✅ TypeScript compilation zero errors
6. ✅ Documentation updated

**Future Phases** (Phase 4.5, 4.6):
- Bulk Invitations enhancement
- Shared Tokens system
- Additional E2E test activation

---

## References

**Similar Components**:
- ProjectDetailView.tsx (reference pattern)
- SSHKeyModal.tsx (modal patterns)
- ProjectManagementView (table patterns)

**API Documentation**:
- Lines 1030-1161: Invitation API methods
- Lines 339-489: Type definitions

**Test Specifications**:
- tests/e2e/invitation-workflows.spec.ts: All test cases
- tests/e2e/pages/ProjectsPage.ts: Test helper methods

---

**Status**: ✅ Planning Complete - Ready for Implementation

**Next Step**: Create InvitationManagementView component shell with Individual Invitations tab
