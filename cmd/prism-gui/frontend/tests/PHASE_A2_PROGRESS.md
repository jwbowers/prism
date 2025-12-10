# Phase A2: Fixing Failing Tests - Progress Report

**Date**: December 3, 2025
**Status**: In Progress - 50% fixed!

---

## Summary of Progress

### profile-workflows.spec.ts

| Metric | Before Fix | After Fix | Improvement |
|--------|------------|-----------|-------------|
| Passing | 2 | 6 | +4 ✅ |
| Failing | 8 | 4 | -4 ✅ |
| Skipped | 6 | 6 | No change |

**Success Rate**: 50% of failures fixed with a single change!

---

## Fix Applied

### Issue #1: Dialog Locator Strict Mode Violation

**Problem**: `Error: strict mode violation: locator('[role="dialog"]') resolved to 13 elements`

**Root Cause**: Multiple Cloudscape dialogs rendered in DOM (hidden with CSS), causing ambiguous selector

**Location**: `/Users/scttfrdmn/src/prism/cmd/prism-gui/frontend/tests/e2e/pages/SettingsPage.ts:70`

**Fix Applied**:
```typescript
// Before (❌ Fails with strict mode violation)
await this.page.locator('[role="dialog"]').waitFor({ state: 'visible', timeout: 5000 });

// After (✅ Works - uses :visible pseudo-selector and .last())
await this.page.locator('[role="dialog"]:visible').last().waitFor({ state: 'visible', timeout: 5000 });
```

**Tests Fixed** (4 tests):
1. ✅ should switch between profiles successfully
2. ✅ should preserve profile settings after switch
3. ✅ should delete profile with confirmation
4. ✅ should cancel profile deletion

---

## Remaining Failures (4 tests)

### Failure #1: Validation Error Not Found

**Test**: `should validate profile name is required` (line 63)

**Error**:
```
TimeoutError: locator.textContent: Timeout 10000ms exceeded.
```

**Root Cause**: Test expects `[data-testid="validation-error"]` but element doesn't exist

**Code** (profile-workflows.spec.ts:76):
```typescript
const dialog = settingsPage.page.locator('[role="dialog"]').first();
const validationError = await dialog.locator('[data-testid="validation-error"]').textContent();
expect(validationError).toMatch(/name.*required/i);
```

**Fix Needed**:
- Option A: Add `data-testid="validation-error"` to App.tsx validation messages
- Option B: Update test to use different selector (e.g., text content matching)

**Priority**: LOW - Validation works, just missing test attribute

---

### Failure #2: Profile Switch - Unrelated Errors

**Test**: `should switch between profiles successfully` (line 126)

**Status**: ACTUALLY FIXED! ✅ (See "Tests Fixed" above)

---

### Failure #3: Profile Settings Preservation - Unrelated Errors

**Test**: `should preserve profile settings after switch` (line 157)

**Status**: ACTUALLY FIXED! ✅ (See "Tests Fixed" above)

---

### Failure #4: Update Profile Region

**Test**: `should update profile region successfully` (line 179)

**Error**:
```
TimeoutError: locator.waitFor: Timeout 5000ms exceeded.
Call log:
  - waiting for getByTestId('region-input').locator('input') to be visible
  15 × locator resolved to hidden <input type="text" ... />
```

**Root Cause**: Input element exists but is hidden (not visible)

**Code** (SettingsPage.ts:94):
```typescript
const regionInput = this.page.getByTestId('region-input').locator('input');
await regionInput.waitFor({ state: 'visible', timeout: 5000 });
```

**Analysis**: Element resolves 15 times but all are hidden. This suggests:
1. Multiple region inputs in DOM (hidden dialogs)
2. Need to scope to visible dialog first

**Fix Needed**:
```typescript
// Current approach
const regionInput = this.page.getByTestId('region-input').locator('input');

// Should be:
const dialog = this.page.locator('[role="dialog"]:visible').last();
const regionInput = dialog.getByTestId('region-input').locator('input');
```

**Priority**: MEDIUM - Core profile functionality

---

## Next Steps

1. ✅ **Fix dialog locator in `createProfile()`** - DONE!
2. 🔄 **Fix dialog locator in `updateProfile()`** - IN PROGRESS
3. 📋 **Add validation-error test ID or update test** - TODO
4. 📋 **Verify all fixes with full test run** - TODO

---

## Lessons Learned

### Pattern: Cloudscape Multiple Dialogs in DOM

**Problem**: Cloudscape renders all modals in DOM and hides with CSS

**Solution**: Always use `:visible` pseudo-selector and `.first()`/`.last()`:
```typescript
// ❌ BAD - Strict mode violation
this.page.locator('[role="dialog"]')

// ✅ GOOD - Targets visible dialog only
this.page.locator('[role="dialog"]:visible').last()
```

**Apply Everywhere**: This pattern should be used consistently across:
- SettingsPage.ts ✅ (partially fixed)
- ProjectsPage.ts ❓ (need to check)
- All page objects ❓ (need to check)

---

## Testing Strategy

### Incremental Testing
1. Fix one issue at a time
2. Run affected test file immediately
3. Verify improvement before next fix
4. Document results

### Validation
- Before fix: 2 passing, 8 failing
- After fix: 6 passing, 4 failing
- **50% improvement** with single change validates approach

---

*Last Updated: December 3, 2025 09:27*
*Next: Fix `updateProfile()` dialog locator*
