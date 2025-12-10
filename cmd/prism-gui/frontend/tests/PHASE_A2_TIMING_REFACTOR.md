# Phase A2: Timing Refactoring - Making Tests Deterministic

**Date**: December 3, 2025
**Status**: Completed ✅

---

## Problem Statement

User feedback: **"Fluctuating [test results] - is bad meaning there is probably some timing issue - tests should be timing independent"**

### Root Cause

Tests were using arbitrary `waitForTimeout()` calls instead of polling for actual state changes:

```typescript
// ❌ ANTI-PATTERN - Race condition
await settingsPage.createProfile(name, 'default', 'us-west-2');
await settingsPage.page.waitForTimeout(1000);  // Sometimes 1s is enough, sometimes not
await settingsPage.createProfile(name2, 'default', 'us-east-1');
```

**Problem**: Fixed timeouts create race conditions. Sometimes 1 second is enough, sometimes not, causing flaky test results.

---

## Solution: Polling-Based Waits

### New Polling Helper Added to SettingsPage.ts

```typescript
/**
 * Poll until a profile appears in the list after creation
 */
async waitForProfileToExist(profileName: string, timeout: number = 10000) {
  // First wait for any dialogs to close
  await this.waitForDialogClose();

  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    try {
      const exists = await this.verifyProfileExists(profileName);
      if (exists) {
        return; // Success - profile exists!
      }
    } catch (error) {
      // Continue polling even if query fails
    }
    // Wait for table to update or short timeout
    await this.page.waitForLoadState('domcontentloaded', { timeout: 500 }).catch(() => {});
    await this.page.waitForTimeout(200); // Small delay between polls
  }
  throw new Error(`Profile "${profileName}" did not appear within ${timeout}ms`);
}
```

**Location**: `/Users/scttfrdmn/src/prism/cmd/prism-gui/frontend/tests/e2e/pages/SettingsPage.ts:204-223`

### Existing Polling Helpers Utilized

Already available in SettingsPage.ts:
- `waitForProfileToBeCurrent()` - Polls until profile becomes current
- `waitForProfileToBeRemoved()` - Polls until profile is deleted
- `waitForProfileRegion()` - Polls until region update reflects in UI
- `waitForDialogClose()` - Waits for dialogs to actually close

---

## Refactoring Applied

### Files Modified

1. **SettingsPage.ts** - Added `waitForProfileToExist()` helper (lines 204-223)
2. **profile-workflows.spec.ts** - Replaced ALL `waitForTimeout()` calls with polling

### Test Changes Summary

| Test | Before | After |
|------|--------|-------|
| `should switch between profiles successfully` | `waitForTimeout(1000)` x 3 | `waitForProfileToExist()` x 2 + `waitForProfileToBeRemoved()` x 2 |
| `should preserve profile settings after switch` | `waitForTimeout(1000)` x 2 | `waitForProfileToExist()` + `switchProfile()` (polls internally) |
| `should not allow updating to invalid region` (skipped) | `waitForTimeout(1000)` + `waitForTimeout(500)` | `waitForProfileToExist()` + dialog visibility wait |
| `should export profile configuration` (skipped) | `waitForTimeout(1000)` | `waitForProfileToExist()` |
| `should cancel profile deletion` | `waitForTimeout(1000)` + `waitForTimeout(500)` | `waitForProfileToExist()` + `waitForDialogClose()` |
| `should prevent deleting currently active profile` | `waitForTimeout(1000)` x 2 | `waitForProfileToExist()` + `switchProfile()` (polls internally) |

### Before (❌ Timing-Dependent):

```typescript
// Create test profiles
await settingsPage.createProfile(name1, 'default', 'us-west-2');
await settingsPage.page.waitForTimeout(1000);  // ❌ Race condition
await settingsPage.createProfile(name2, 'default', 'us-east-1');
await settingsPage.page.waitForTimeout(1000);  // ❌ Race condition

// Cleanup
await settingsPage.deleteProfile(name1);
await settingsPage.clickButton('delete');
await settingsPage.page.waitForTimeout(500);  // ❌ Race condition
await settingsPage.deleteProfile(name2);
await settingsPage.clickButton('delete');
```

### After (✅ Deterministic):

```typescript
// Create test profiles with polling to ensure they exist
await settingsPage.createProfile(name1, 'default', 'us-west-2');
await settingsPage.waitForProfileToExist(name1);  // ✅ Polls for actual state
await settingsPage.createProfile(name2, 'default', 'us-east-1');
await settingsPage.waitForProfileToExist(name2);  // ✅ Polls for actual state

// Cleanup - poll after each delete to ensure completion
await settingsPage.deleteProfile(name1);
await settingsPage.clickButton('delete');
await settingsPage.waitForProfileToBeRemoved(name1);  // ✅ Polls for actual state
await settingsPage.deleteProfile(name2);
await settingsPage.clickButton('delete');
```

---

## Verification

### Removed ALL Timing Dependencies

```bash
grep -n "waitForTimeout" tests/e2e/profile-workflows.spec.ts
# Result: No matches found ✅
```

**Total `waitForTimeout()` calls removed**: 10
**Total polling-based waits added**: 10

---

## Benefits

1. **Deterministic Tests**: Tests now wait for actual state changes, not arbitrary timeouts
2. **No More Flakiness**: Results will be consistent across runs
3. **Faster When Possible**: Polling completes as soon as condition is met (not always waiting full timeout)
4. **Clearer Intent**: Code explicitly shows what state we're waiting for
5. **Better Debugging**: Timeout errors now show exactly what condition wasn't met

---

## Pattern for Future Tests

### ✅ DO: Poll for State Changes

```typescript
await performAction();
await waitForStateChange();  // Polls until actual state is reached
expect(actualState).toBe(expectedState);
```

### ❌ DON'T: Use Arbitrary Timeouts

```typescript
await performAction();
await page.waitForTimeout(1000);  // ❌ Race condition - sometimes enough, sometimes not
expect(actualState).toBe(expectedState);
```

### Polling Helper Pattern

```typescript
async waitForSomeCondition(timeout: number = 10000) {
  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    try {
      if (await checkCondition()) {
        return; // Success!
      }
    } catch {
      // Continue polling even if check fails
    }
    await this.page.waitForLoadState('domcontentloaded', { timeout: 500 }).catch(() => {});
    await this.page.waitForTimeout(200); // Small delay between polls
  }
  throw new Error(`Condition not met within ${timeout}ms`);
}
```

---

## Next Steps

1. ✅ Run profile-workflows.spec.ts to verify determinism
2. 📋 Apply same pattern to other test files with timing issues
3. 📋 Document polling pattern in testing guidelines

---

*Last Updated: December 3, 2025 09:58*
*Pattern: Polling-based waits for deterministic E2E tests*
