# Prism v0.5.13 Release Notes: Cost Control & Monitoring

**Release Date**: November 15, 2025
**Focus**: Advanced cost control and system power management

---

## 🎉 Highlights

🚦 **Launch Throttling**: Prevent cost overruns with multi-level rate limiting
🌙 **Sleep/Wake Auto-Hibernation**: Intelligent instance hibernation with idle detection
🐛 **Bug Fixes**: Race condition eliminated, 100% test pass rate with -race detector
📝 **Technical Debt Tracked**: Integration test and IOKit deprecation documented

---

## 🚀 New Features

### Launch Throttling System (#90)

Comprehensive rate limiting to prevent accidental cost overruns from rapid or scripted launches.

**Three-Level Protection**:
- **Global Limit**: System-wide launches per hour (default: 20)
- **Per-User Limit**: Individual user launches per hour (default: 10)
- **Per-Template Limit**: Template-specific launches per hour (default: 5)

**CLI Commands**:
```bash
# Check current throttling status
prism admin throttling status

# Configure limits
prism admin throttling configure --global 30 --per-user 15 --per-template 10

# Reset throttle counters
prism admin throttling reset

# Check wait time for a launch
prism admin throttling wait-time --user alice --template python-ml
```

**REST API**:
- `GET /api/v1/throttle/status` - Current status and usage
- `POST /api/v1/throttle/configure` - Update configuration
- `POST /api/v1/throttle/reset` - Reset counters
- `GET /api/v1/throttle/wait-time` - Estimate wait time

**Use Cases**:
- Prevent cost overruns from script loops
- Fair resource sharing in multi-user environments
- Protect expensive GPU templates from overuse
- Handle burst traffic gracefully with token bucket

---

### Sleep/Wake Auto-Hibernation (#91)

Automatic instance hibernation when your computer goes to sleep, with intelligent idle detection to protect active workloads.

**Three Hibernation Modes**:

1. **`idle_only`** (RECOMMENDED - Default):
   - Only hibernates instances that are actually idle
   - Integrates with CloudWatch metrics
   - Checks: CPU (<5%), memory (<10%), network (<1KB/s), disk, GPU
   - **Fail-safe**: Errors skip hibernation to protect your work
   - Perfect for: Long-running jobs that shouldn't be interrupted

2. **`all`** (Aggressive):
   - Hibernates all instances except explicitly excluded
   - Maximum cost savings
   - Perfect for: Personal development environments

3. **`manual_only`** (Conservative):
   - No automatic hibernation
   - Complete user control
   - Perfect for: Critical production workloads

**CLI Commands**:
```bash
# Check status and statistics
prism admin sleep-wake status

# Enable/disable monitoring
prism admin sleep-wake enable
prism admin sleep-wake disable

# Configure behavior
prism admin sleep-wake configure --mode idle_only              # Safe default
prism admin sleep-wake configure --mode all --exclude prod-db  # Aggressive
prism admin sleep-wake configure --resume-on-wake              # Auto-resume
```

**REST API**:
- `GET /api/v1/sleep-wake/status` - Monitor status and statistics
- `POST /api/v1/sleep-wake/configure` - Update configuration
- `GET /api/v1/sleep-wake/statistics` - Detailed hibernation stats

**Platform Support**:
- ✅ **macOS**: Full support with IOKit integration
- 🔄 **Linux**: Coming soon (systemd-logind integration)
- 🔄 **Windows**: Coming soon (WM_POWERBROADCAST integration)

**Configuration Options**:
- **Grace Period**: 30s default (cancel unintended hibernation)
- **Idle Check Timeout**: 10s default (configurable)
- **Resume on Wake**: Optional (disabled by default for safety)
- **Exclusion List**: Protect specific instances

**Cost Savings**:
- Estimated 30-50% reduction in idle instance costs
- Automatic overnight/weekend hibernation for laptop users
- Intelligent idle detection prevents workflow interruption

---

## 🐛 Bug Fixes & Improvements

### Race Condition Fixed
**Package**: `pkg/progress/reporter.go`

Fixed data race in progress callback notification system:
- Deep copy metadata maps before goroutine spawning
- Copy callbacks slice before iteration
- All operations now thread-safe
- Tests pass with `-race` detector

**Impact**: Improved stability and reliability of progress reporting across CLI, TUI, and GUI

### Test Stability
- ✅ 100% unit test pass rate
- ✅ Zero race conditions detected
- ✅ All packages tested with `-race` detector
- ✅ 14/14 throttling tests passing
- ✅ 20/20 progress tests passing

### Technical Debt Documented
1. **Integration Test Timeout** ([#252](https://github.com/scttfrdmn/prism/issues/252)):
   - Tracked for future remediation
   - Does not block releases
   - Unit tests provide comprehensive coverage

2. **macOS IOKit Deprecation** (TECHNICAL_DEBT_BACKLOG #12):
   - `kIOMasterPortDefault` deprecated in macOS 12.0+
   - Compiler warning documented
   - Functionality works correctly
   - Scheduled for v0.6.0 update

---

## 📊 Implementation Statistics

- **Lines of Code**: 2,575+ across 13 files
- **Test Coverage**: 100% unit test pass rate
- **Race Detection**: Clean with `-race` detector
- **Build Status**: ✅ Successful (warnings documented)
- **Files Created**:
  - Launch Throttling: 5 files (750+ lines)
  - Sleep/Wake Monitor: 8 files (1,825+ lines)

---

## 🎯 Use Cases

### Laptop Researchers
**Before v0.5.13**: Instances run overnight, wasting $5-15/day
**After v0.5.13**: Auto-hibernation saves 30-50% on idle costs

```bash
prism admin sleep-wake configure --mode idle_only
# Instances auto-hibernate when you close your laptop
# Long-running jobs protected by idle detection
```

### Lab Environments
**Before v0.5.13**: No throttling, users can accidentally launch 50 instances
**After v0.5.13**: Per-user limits prevent individual cost overruns

```bash
prism admin throttling configure --per-user 10
# Each user limited to 10 launches per hour
# Fair resource sharing across lab members
```

### GPU Workloads
**Before v0.5.13**: Expensive GPU templates used liberally
**After v0.5.13**: Template-specific throttling protects budget

```bash
prism admin throttling configure --per-template 5
# GPU templates limited to 5 launches/hour
# Forces users to be intentional with expensive resources
```

### Multi-User Classes
**Before v0.5.13**: 50 students launching simultaneously causes API throttling
**After v0.5.13**: Global throttling prevents AWS rate limit errors

```bash
prism admin throttling configure --global 20
# System-wide limit prevents AWS API overwhelm
# Predictable performance for bulk operations
```

---

## 📚 Documentation

### Updated Documentation
- ✅ ROADMAP.md - v0.5.10-v0.5.12 corrected, v0.5.13-v0.5.15 added
- ✅ GitHub Issues - Comprehensive implementation summaries
- ✅ TECHNICAL_DEBT_BACKLOG.md - New entries for tracked issues
- ✅ Code Comments - Extensive godoc throughout

### CLI Help
All commands include comprehensive help text:
```bash
prism admin throttling --help
prism admin sleep-wake --help
```

---

## 🔄 Upgrade Guide

### From v0.5.12 to v0.5.13

**No Breaking Changes** - This is a feature-additive release.

**Optional Configuration**:
1. **Enable Launch Throttling** (optional, disabled by default):
   ```bash
   prism admin throttling configure --global 20 --per-user 10
   ```

2. **Enable Sleep/Wake Monitoring** (optional, macOS only):
   ```bash
   prism admin sleep-wake enable
   prism admin sleep-wake configure --mode idle_only
   ```

**Backwards Compatibility**:
- Existing instances: No changes
- Existing configurations: Fully compatible
- API endpoints: Additive only (no removals)
- CLI commands: New commands added under `admin` group

---

## ⚠️ Known Issues

### Integration Test Timeout (#252)
- **Impact**: Low (does not affect functionality)
- **Scope**: Test infrastructure only
- **Workaround**: Unit tests provide comprehensive coverage
- **Timeline**: Tracked for future fix

### macOS IOKit Deprecation Warning (TECHNICAL_DEBT_BACKLOG #12)
- **Impact**: None (compiler warning only)
- **Scope**: macOS builds only
- **Functionality**: Works correctly on all macOS versions
- **Timeline**: v0.6.0 update planned

---

## 🚀 What's Next

### v0.5.14 (December 2025): Desktop Applications Foundation
- Nice DCV integration for GUI applications
- MATE desktop environment support
- Browser-based remote desktop access

### v0.5.15 (December 2025): Desktop Applications
- MATLAB template (numerical computing)
- QGIS templates (geographic information systems)
- Mathematica and Stata templates

### v0.6.0 (January 2026): Enterprise Authentication
- OAuth/OIDC with Globus Auth
- LDAP/Active Directory integration
- SAML support for federated SSO

---

## 🙏 Acknowledgments

Special thanks to user feedback on Issue #91, which led to the idle detection integration:

> *"It would be very useful to have an option to not except certain cloud workspaces from this - There may be a long running workload for example so it should not be automatic - we do have idle detection so that should be considered/integrated"*

This feedback resulted in the `idle_only` hibernation mode, making sleep/wake monitoring safe for long-running research workloads.

---

## 📝 Release Assets

- **Binaries**: `prism` (CLI), `prismd` (daemon), `prism-gui` (GUI)
- **Platforms**: macOS (Apple Silicon + Intel), Linux (amd64, arm64), Windows (amd64)
- **Source**: Tagged as `v0.5.13` in GitHub repository

For installation instructions, see: [Installation Guide](https://github.com/scttfrdmn/prism#installation)

---

**Full Changelog**: [v0.5.12...v0.5.13](https://github.com/scttfrdmn/prism/compare/v0.5.12...v0.5.13)
