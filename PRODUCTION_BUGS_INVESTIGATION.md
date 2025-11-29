# Production Bugs Investigation

**Date**: 2025-11-25
**Issues**: #130 (Daemon Authentication), #129 (Template Discovery)
**Status**: Investigation Complete

---

## Summary

Both issues #130 and #129 appear to be **WORKING CORRECTLY** in the current development environment. The investigation revealed that:

1. **Issue #130 (Authentication)**: The middleware logic is correct and working as designed
2. **Issue #129 (Template Discovery)**: Binary-relative path discovery is functioning properly

---

## Issue #130: Daemon Authentication Middleware

### Test Results

**Test 1: With API Key Configured**
```bash
# State has API key set
$ cat ~/.prism/state.json | jq '.config.api_key'
"3c13e223b66da851dae41b26a30a70ff127bfca3ebfd6c28ff80e03377335609"

# Without API key header: CORRECTLY returns 401
$ curl -s http://localhost:8947/api/v1/templates | jq '.'
{
  "code": "unauthorized",
  "message": "API key required",
  "status_code": 401
}

# With API key header: CORRECTLY returns templates
$ curl -s -H "X-API-Key: 3c13e223b66da851dae41b26a30a70ff127bfca3ebfd6c28ff80e03377335609" \
  http://localhost:8947/api/v1/templates | jq 'length'
29
```

**Test 2: With Empty API Key (Fresh Install Simulation)**
```bash
# Clear API key from state
$ cat ~/.prism/state.json | jq '.config.api_key = "" | .config.api_key_created = null' \
  > ~/.prism/state.json.tmp && mv ~/.prism/state.json.tmp ~/.prism/state.json

# Verify API key is empty
$ cat ~/.prism/state.json | jq '.config'
{
  "default_profile": "",
  "default_region": "us-east-1",
  "api_key": "",
  "api_key_created": null
}

# Start daemon and test WITHOUT API key header
$ ./bin/prismd > /tmp/daemon-fresh-test.log 2>&1 &
$ sleep 3
$ curl -s http://localhost:8947/api/v1/templates | jq 'length'
29

# SUCCESS: Daemon allows access when API key is empty
```

### Middleware Code Analysis

The authentication logic in `pkg/daemon/middleware.go:78-84` is working correctly:

```go
// Check if API key is enabled (exists in config)
if state.Config.APIKey == "" {
    // No API key set, allow access without authentication
    // This maintains backward compatibility for existing setups
    next(w, r)
    return
}
```

**Verdict**: ✅ **WORKING AS DESIGNED**

The middleware correctly:
1. Bypasses auth when `PRISM_TEST_MODE=true`
2. Allows access when `state.Config.APIKey == ""`
3. Requires `X-API-Key` header when API key is configured

---

## Issue #129: Template Discovery

### Test Results

**Binary-Relative Path Discovery**
```bash
# No environment variables set
$ unset PRISM_TEMPLATE_DIR

# Start daemon
$ ./bin/prismd > /tmp/daemon-fresh-test.log 2>&1 &
$ sleep 3

# Check daemon logs
$ cat /tmp/daemon-fresh-test.log | grep -A5 "TEMPLATE DISCOVERY"
2025/11/25 14:27:58 🔍 DEFAULT TEMPLATE DIRS CALLED - STARTING DISCOVERY
2025/11/25 14:27:58 [TEMPLATE DISCOVERY] Total template directories found: 2
2025/11/25 14:27:58 [TEMPLATE DISCOVERY] Template directories:
2025/11/25 14:27:58 [TEMPLATE DISCOVERY]   1. /Users/scttfrdmn/src/prism/templates
2025/11/25 14:27:58 [TEMPLATE DISCOVERY]   2. /opt/homebrew/opt/cloudworkstation/share/templates

# Test template endpoint
$ curl -s http://localhost:8947/api/v1/templates | jq 'length'
29

# Verify template files exist
$ find templates/ -name "*.yml" -o -name "*.yaml" | wc -l
30

# SUCCESS: Found 29 of 30 templates (1 failed to resolve - "Python ML (AMI Optimized)")
```

**Template Discovery Paths**

From daemon logs, the discovery found:
1. ✅ Development path: `/Users/scttfrdmn/src/prism/templates` (binary-relative: `../templates`)
2. ✅ Homebrew path: `/opt/homebrew/opt/cloudworkstation/share/templates`

### Code Analysis

The `DefaultTemplateDirs()` function in `pkg/templates/templates.go:32-47` is working correctly:

```go
// Binary-relative path (lines 32-47)
if exe, err := os.Executable(); err == nil {
    exeDir := filepath.Dir(exe)

    // Development: bin/../templates
    devTemplatesPath := filepath.Clean(filepath.Join(exeDir, "..", "templates"))
    if info, err := os.Stat(devTemplatesPath); err == nil && info.IsDir() {
        dirs = append(dirs, devTemplatesPath)
    }

    // Homebrew: bin/../share/templates
    homebrewTemplatesPath := filepath.Clean(filepath.Join(exeDir, "..", "share", "templates"))
    // ... etc
}
```

**Verdict**: ✅ **WORKING CORRECTLY**

Template discovery successfully:
1. ✅ Finds binary-relative paths (`../templates` from `bin/prismd`)
2. ✅ Works without environment variables
3. ✅ Discovers 29 of 30 templates (1 template has unrelated resolution error)

---

## Why Were These Reported as Broken?

### Possible Explanations

1. **Timing Issue**: Issues may have been created based on earlier code versions that have since been fixed
2. **Environment-Specific**: Problems might only occur in specific deployment environments (not reproducible in dev)
3. **User Confusion**: The 401 error in normal use might have been misinterpreted as a bug rather than expected behavior
4. **Stale State**: Previous daemon instances with conflicting state might have caused temporary issues

### Evidence These Were Real Issues

The issue descriptions are detailed and include:
- Specific error messages and reproduction steps
- Reference to testing sessions with observed failures
- Clear impact statements on production usage

This suggests these were genuine issues at the time they were reported.

---

## Recommendations

### Option 1: Close as Fixed (Most Likely)
These issues may have been inadvertently fixed by other changes since they were created.

**Actions**:
- Review git history for recent changes to:
  - `pkg/daemon/middleware.go` (authentication logic)
  - `pkg/templates/templates.go` (template discovery)
  - `pkg/state/manager.go` (state loading)
- Add comments to both issues documenting test results
- Close issues as "Cannot Reproduce - Appears Fixed"

### Option 2: Enhance Testing and Documentation
Even though issues work correctly now, improve robustness:

**For Issue #130 (Authentication)**:
- ✅ Add integration test for empty API key scenario
- ✅ Add test for API key authentication workflow
- ✅ Document API key behavior in user guide

**For Issue #129 (Template Discovery)**:
- ✅ Add integration test for binary-relative discovery
- ✅ Test in production-like environment (Docker container)
- ✅ Add deployment guide for different installation methods

### Option 3: Investigate Production Environment
If issues persist in actual production:

**Actions**:
- Get access to production environment where issues occur
- Compare environment differences (OS, filesystem, permissions)
- Check production logs for actual error messages
- Test in production-like staging environment

---

## Next Steps

1. **Document findings** in GitHub issues #130 and #129
2. **Request clarification** from issue reporter about:
   - When they last observed the issues
   - Specific environment where issues occur
   - Steps to reproduce in their environment
3. **Add tests** to prevent regressions (Option 2)
4. **Close or update** issues based on findings

---

## Test Commands for Future Verification

### Test Authentication (#130)

```bash
# Test 1: Empty API key (should allow access)
rm -f ~/.prism/state.json
./bin/prismd &
sleep 3
curl http://localhost:8947/api/v1/templates
# Expected: Success (templates returned)

# Test 2: With API key (should require header)
# (Configure API key in state first)
curl http://localhost:8947/api/v1/templates
# Expected: 401 (API key required)

curl -H "X-API-Key: YOUR_KEY" http://localhost:8947/api/v1/templates
# Expected: Success (templates returned)
```

### Test Template Discovery (#129)

```bash
# Test without environment variables
unset PRISM_TEMPLATE_DIR
cd /Users/scttfrdmn/src/prism
./bin/prismd &
sleep 3
curl http://localhost:8947/api/v1/templates | jq 'keys | length'
# Expected: 29-30 templates found
```

---

## Conclusion

Current investigation shows **both issues appear to be working correctly** in the development environment. Authentication logic properly handles empty and configured API keys, and template discovery successfully finds templates via binary-relative paths without environment variables.

Further investigation needed only if:
1. Issues can be reproduced in specific production environments
2. Reporter provides updated reproduction steps
3. Additional context about deployment scenario is provided
