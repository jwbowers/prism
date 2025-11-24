# Phase 1 Audit: Missing GUI Features and Test IDs

**Date:** 2025-11-22
**Test Results:** 15 passed, 55 failed, 30 skipped (out of 100 tests)

## Critical Finding: ProfileSelectorView Component Missing

**Status:** ❌ **COMPONENT DOES NOT EXIST**

The `ProfileSelectorView` component is referenced in Settings (App.tsx:5639) but is **never defined anywhere in the codebase**. This causes **all 16 profile management tests to fail**.

### Expected Features (from profile-workflows.spec.ts):

1. **Create Profile** (4 tests):
   - "Create Profile" button
   - Form with inputs: profile name, aws profile, region
   - Validation: name required, region format, duplicate prevention
   - data-testid="validation-error" for error messages

2. **Switch Profile** (2 tests):
   - Ability to switch between profiles
   - Current profile indicator
   - Settings preserved after switch

3. **Update Profile** (2 tests):
   - Edit profile region
   - Validation for invalid regions

4. **Export Profile** (1 test):
   - Export button triggering download
   - JSON file format: `{name}-profile.json`

5. **Import Profile** (2 tests):
   - Import button with file upload
   - JSON validation

6. **Delete Profile** (3 tests):
   - Delete button with confirmation dialog
   - Cancel functionality
   - Prevent deletion of active profile

7. **Profile Listing** (2 tests):
   - Display all profiles
   - Show current profile indicator

### API Integration Points:
The component will need to integrate with the existing profile API:
- `/api/v1/profiles` (list profiles)
- `/api/v1/profiles` POST (create)
- `/api/v1/profiles/{name}` PUT (update)
- `/api/v1/profiles/{name}` DELETE (delete)
- `/api/v1/profiles/{name}/activate` POST (switch)

## Other Missing Features

### 1. Backup Management - Actions Dropdown (13 tests failing)
**Location:** BackupManagementView (App.tsx)
**Missing:**
- Actions dropdown button for each backup row
- Actions: Restore, Clone, View Details, Delete
- Dialogs for restore/clone/delete workflows
- Cost savings display in delete dialog
- Restore time warning in restore dialog
- Instance name validation in restore form

**Required data-testids:**
- `data-testid="backup-actions-dropdown"`
- `data-testid="restore-action"`
- `data-testid="clone-action"`
- `data-testid="delete-action"`
- `data-testid="restore-dialog"`
- `data-testid="delete-dialog"`

### 2. Instance Management - Launch and Actions (11 tests failing)
**Location:** InstanceManagementView (App.tsx)
**Missing:**
- Launch dialog workflow (basic + advanced)
- Instance name validation
- Template selection validation
- Cost estimate display based on size
- Hibernation option checkbox
- EBS volume attachment during launch
- Instance action buttons (start, stop, terminate, hibernate)
- Search/filter functionality

**Required data-testids:**
- `data-testid="launch-dialog"`
- `data-testid="instance-name-input"`
- `data-testid="template-select"`
- `data-testid="cost-estimate"`
- `data-testid="hibernation-option"`
- `data-testid="search-instances"`
- `data-testid="filter-status"`

### 3. Storage Management - Search and CRUD (15 tests failing)
**Location:** StorageManagementView (App.tsx)
**Missing:**
- Search input for EFS/EBS volumes (not visible)
- Create EFS dialog with name validation
- Create EBS dialog with size validation
- Delete confirmation dialogs
- Attach/detach volume workflows
- Volume size and type display

**Required data-testids:**
- `data-testid="search-volumes"` ✅ (may exist but not visible)
- `data-testid="create-efs-button"`
- `data-testid="create-ebs-button"`
- `data-testid="efs-name-input"`
- `data-testid="ebs-size-input"`
- `data-testid="volume-type"`
- `data-testid="volume-size"`
- `data-testid="attach-button"`
- `data-testid="detach-button"`

### 4. Hibernation Workflows - Already Passing! ✅
**Status:** 4 tests passing
**Reason:** Educational/tooltip tests that don't require complex interactions

## Comparison with Existing Settings Components

### ✅ Components That Exist:
- UserManagementView (App.tsx:3728)
- AMIManagementView (App.tsx:6097)
- MarketplaceView (App.tsx:6393)
- IdleDetectionView (App.tsx:6723)
- LogsView (App.tsx:7038)
- RightsizingView (App.tsx:7293)
- PolicyView (App.tsx:7300)

### ❌ Components Missing:
- **ProfileSelectorView** - CRITICAL (16 tests blocked)

## Summary of Test Failures

| Feature Area | Tests Failed | Primary Issue |
|-------------|--------------|---------------|
| **Profile Management** | 16 | Component doesn't exist |
| **Backup Actions** | 13 | Missing Actions dropdown & dialogs |
| **Instance Launch/Actions** | 11 | Missing launch dialog & action buttons |
| **Storage CRUD/Search** | 15 | Missing search visibility & dialogs |
| **TOTAL** | **55 failures** | - |

## Recommended Implementation Order

### Priority 1: ProfileSelectorView (Blocks 16 tests)
Build complete CRUD component with:
1. Profile list table with current indicator
2. Create Profile dialog (name, aws_profile, region inputs)
3. Edit Profile dialog
4. Delete confirmation dialog with active profile check
5. Switch profile button/action
6. Export/Import functionality
7. Full validation and error handling

### Priority 2: Backup Actions Dropdown (Blocks 13 tests)
Add Actions menu to backup rows with:
1. Dropdown button on each row
2. Restore dialog with instance name input & time warning
3. Clone dialog
4. Delete confirmation with cost savings
5. View Details action

### Priority 3: Instance Launch Dialog (Blocks 11 tests)
Complete launch workflow:
1. Launch button opening dialog
2. Form with name, template, size, options
3. Cost estimate display
4. Hibernation checkbox for supported types
5. EBS volume attachment options
6. Validation for all inputs

### Priority 4: Storage Search & Dialogs (Blocks 15 tests)
Enhance storage management:
1. Make search input visible/functional
2. Create EFS dialog with validation
3. Create EBS dialog with size validation
4. Delete confirmations for both types
5. Attach/detach workflows
6. Display volume size and type in tables

## Next Steps

Following the user's guidance: **"If we need to detour to build out the GUI - then we should include that"**

We will:
1. Build ProfileSelectorView component (highest priority)
2. Add all required data-testids during implementation
3. Test incrementally after each component
4. Continue with other missing features in priority order
5. Ensure all 100 tests pass before final commit
