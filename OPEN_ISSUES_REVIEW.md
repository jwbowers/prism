# Open Issues Comprehensive Review

**Generated**: 2025-11-25
**Total Open Issues**: 150+
**Ready/High-Priority Issues**: 38
**Current Focus**: v0.5.16 Release - E2E Test Coverage

---

## Executive Summary

Prism has **150+ open issues** spanning multiple development phases. The immediate priority is completing v0.5.16 with full E2E test coverage. However, there are **18 ready issues without milestones** that need triage and prioritization.

### Critical Finding

**v0.5.16 Release is BLOCKED** by 57 skipped E2E tests across 4 feature areas:
1. **Invitations**: 28 skipped tests
2. **Users**: 13 skipped tests
3. **Projects**: 10 skipped tests
4. **Profiles**: 6 skipped tests

**No release until all skipped tests are resolved** (implemented or explicitly deferred).

---

## Issues By Milestone

### 🚨 **NO MILESTONE** (18 ready/high-priority issues)

**Status**: NEEDS TRIAGE - These issues are marked ready but not assigned to a milestone

#### v0.5.16 E2E Test Coverage (11 issues) - **CRITICAL**

| # | Title | Priority | Estimate |
|---|-------|----------|----------|
| 315 | [Epic] Activate Projects/Users/Invitations E2E Tests (51 skipped tests) | High | 40-50h |
| 320 | [GUI] Complete Profile Management (6 skipped E2E tests) | Medium | 12h |
| 307 | [GUI] Phase 2: Implement Validation Error Display for Forms | High | 4-6h |
| 308 | [GUI] Phase 4.1: Implement Project Detail View | High | 3h |
| 309 | [GUI] Phase 4.2: Implement SSH Key Management for Users | Medium | 4h |
| 310 | [GUI] Phase 4.3: Implement Individual Invitations System | High | 6h |
| 311 | [GUI] Phase 4.4: Enhance Bulk Invitations System | High | 4h |
| 312 | [GUI] Phase 4.5: Implement Shared Token System | Medium | 6h |
| 313 | [GUI] Phase 4.6: Implement User Provisioning on Workspaces | Medium | 5h |
| 314 | [GUI] Phase 4.7: Implement Statistics and Filtering | Medium | 4h |
| 305 | [Testing] E2E Test Framework Complete: Projects/Users/Invitations (61 tests) | - | Doc only |

**Total Estimated Effort**: 88-98 hours (~11-12 days)

**Recommendation**: Assign all to **v0.5.16 milestone** - these are blockers for release.

#### CLI Integration Tests (3 issues) - **HIGH PRIORITY**

| # | Title | Priority | Status |
|---|-------|----------|--------|
| 293 | Integration Tests: Template Inheritance & Stacking | High | Ready |
| 292 | Integration Tests: Volume Management CLI | High | Ready |
| 291 | Integration Tests: Profile Management CLI | High | Ready |

**Recommendation**: Assign to **v0.5.16 milestone** or defer to **v0.5.17** if E2E GUI tests take precedence.

#### Production Critical Bugs (2 issues) - **BLOCKER**

| # | Title | Priority | Impact |
|---|-------|----------|--------|
| 130 | 🐛 Daemon Authentication Middleware Issues in Production | Critical | Production broken |
| 129 | 🐛 Template Discovery Broken in Production Without Environment Variables | Critical | Production broken |

**Recommendation**: **FIX IMMEDIATELY** - these break production deployments. Assign to **v0.5.16 milestone** with highest priority.

#### Feature Requests (2 issues) - **CAN DEFER**

| # | Title | Priority | Target |
|---|-------|----------|--------|
| 62 | 🖥️ GUI System Tray/Menu Bar and Auto-Start on Login | High | v0.6.0+ |
| 61 | 🔄 Auto-Update Feature: Version Detection and Update Notifications | High | v0.6.0+ |

**Recommendation**: Defer to **v0.6.0** - nice-to-have features, not release blockers.

---

### 📊 **Phase 0.6.0: Budget Safety Net** (4 issues)

**Status**: Well-defined milestone for budget management features

| # | Title | Priority | Type |
|---|-------|----------|------|
| 41 | [Budget] Personal Budget System for Solo Researchers | Critical | Core |
| 42 | [Budget] Budget Alerts and Email Notifications | Critical | Core |
| 43 | [Budget] Pre-Launch Budget Impact Preview | Critical | Core |
| 44 | [Budget] Budget Forecasting and Planning Tool | High | Enhanced |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - These are the next major feature set after v0.5.16.

---

### 🎓 **Phase 0.7.0: Class Management Basics** (3 issues)

**Status**: Education-focused features for classroom use

| # | Title | Priority |
|---|-------|----------|
| 45 | [Class] Course Creation and Management System | Critical |
| 46 | [Class] Template Whitelisting for Courses | Critical |
| 47 | [Class] Student Bulk Management and Budget Distribution | Critical |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - Defer until after budget features (0.6.0).

---

### 🎓 **Phase 0.7.1: TA Support Tools** (2 issues)

**Status**: Teaching Assistant support features

| # | Title | Priority |
|---|-------|----------|
| 48 | [TA] Debug Access (God Mode) for Student Support | Critical |
| 49 | [TA] Instance Reset Capability for Fresh Starts | Critical |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - Part of class management suite.

---

### 🔐 **Phase 0.8.0: Invitation Security** (1 issue)

| # | Title | Priority |
|---|-------|----------|
| 50 | [Invitation] Wire CLI Commands to Existing Security Code | High |

**Recommendation**: **MERGE INTO v0.5.16** - This is part of the invitation system being built for v0.5.16.

---

### 🖥️ **Phase 0.9.0: DCV Desktop Integration** (1 issue)

| # | Title | Priority |
|---|-------|----------|
| 51 | [DCV] NICE DCV Desktop Workstation Support | High |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - Major feature addition for desktop GUIs.

---

### 🌍 **Phase 6.0: International Support** (1 issue)

| # | Title | Priority |
|---|-------|----------|
| 52 | International Support: AWS Regional Expansion | High |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - Post-1.0 enhancement.

---

### 💾 **Phase 5.5: Advanced Storage** (1 issue)

| # | Title | Priority |
|---|-------|----------|
| 53 | Directory Sync: Dropbox-like Local-Cloud Synchronization | High |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - Advanced feature, not critical path.

---

### 🔐 **Phase 6.0: Authentication** (2 issues)

| # | Title | Priority |
|---|-------|----------|
| 35 | [Security] IAM Profile Validation Pre-Launch | High |
| 36 | [Security] Role-Based Access Control (RBAC) System | High |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - Enterprise auth features.

---

### 📋 **Phase 5.4: Policy Framework** (1 issue)

| # | Title | Priority |
|---|-------|----------|
| 56 | Research IT Features: Institutional Research Enablement and Governance | High |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - Enterprise governance features.

---

### 🏥 **v0.6.0: AWS Health & Quota Management** (4 issues)

| # | Title | Priority |
|---|-------|----------|
| 57 | AWS Quota Management & Intelligent Availability Handling | High |
| 58 | Implement Quota Awareness System | High |
| 59 | Implement Intelligent AZ Failover | High |
| 60 | Implement AWS Health Dashboard Integration | High |

**Recommendation**: **KEEP AS SEPARATE MILESTONE** - Important operational features but not blocking current release.

---

## Other Open Issues (Not Ready/Not Prioritized)

**Count**: 112+ issues
**Status**: Backlog - not tagged as ready or high-priority

### Categories:
- **Enhancement**: 118 issues (largest category)
- **Testing**: 14 issues (integration tests, E2E tests, regression)
- **Bug**: Various bugs and technical debt
- **Documentation**: 2 issues
- **Security**: Auth and RBAC features (phase 6.0+)

---

## Recommended Actions for v0.5.16 Release

### 1. **Create v0.5.16 Milestone** (if not exists)

Assign the following issues:

**CRITICAL (Must Fix):**
- #130 - Daemon Authentication Middleware Issues ⚠️
- #129 - Template Discovery Broken ⚠️

**HIGH PRIORITY (E2E Test Coverage):**
- #315 - Epic: Activate E2E Tests (parent)
- #307 - Validation Error Display
- #308 - Project Detail View
- #310 - Individual Invitations
- #311 - Bulk Invitations
- #312 - Shared Tokens
- #313 - User Provisioning
- #314 - Statistics and Filtering
- #309 - SSH Key Management
- #320 - Profile Management
- #50 - Invitation Security (move from 0.8.0)

**OPTIONAL (If Time Permits):**
- #293 - Template Inheritance Tests
- #292 - Volume Management Tests
- #291 - Profile Management Tests

### 2. **Defer to v0.6.0**

Move these issues OUT of immediate scope:
- #62 - GUI System Tray
- #61 - Auto-Update Feature

### 3. **Keep Separate Milestones**

DO NOT merge these into v0.5.16:
- Phase 0.6.0: Budget Safety Net (4 issues)
- Phase 0.7.0: Class Management (3 issues)
- Phase 0.7.1: TA Support (2 issues)
- Phase 0.9.0: DCV Desktop (1 issue)
- Phase 6.0: Authentication (2 issues)
- v0.6.0: AWS Health/Quota (4 issues)

### 4. **Triage Backlog**

Review the 112+ untagged issues and either:
- Close if no longer relevant
- Tag with appropriate priority/milestone
- Keep in backlog for future consideration

---

## Timeline Estimate for v0.5.16

### Critical Path (In Order):

**Week 1: Fix Production Bugs**
- Fix #130 - Daemon auth middleware (2-4 hours)
- Fix #129 - Template discovery (2-4 hours)
- **Deliverable**: Production deployments working

**Week 2-3: Validation & Project Management**
- #307 - Validation error display (4-6 hours)
- #308 - Project detail view (3 hours)
- #314 - Statistics and filtering (4 hours)
- **Deliverable**: 10-15 E2E tests activated

**Week 4-5: User Management**
- #309 - SSH key management (4 hours)
- #313 - User provisioning (5 hours)
- **Deliverable**: 13 user E2E tests activated

**Week 6-7: Invitation System**
- #310 - Individual invitations (6 hours)
- #311 - Bulk invitations (4 hours)
- #312 - Shared tokens (6 hours)
- #50 - Invitation security (4 hours)
- **Deliverable**: 28 invitation E2E tests activated

**Week 8: Profile Management & Polish**
- #320 - Profile management (12 hours)
- Bug fixes and regression testing
- **Deliverable**: All 57 E2E tests activated and passing

### Total Timeline: **8 weeks** (conservative estimate)

**Aggressive Timeline**: 5-6 weeks with parallel development

---

## Success Criteria for v0.5.16

- [ ] #130 and #129 production bugs fixed
- [ ] All 57 skipped E2E tests are unskipped and passing
- [ ] E2E test suite: 165/165 tests active (100% coverage)
- [ ] Zero TypeScript compilation errors
- [ ] All new features documented
- [ ] Release notes prepared

---

## Risk Assessment

### HIGH RISK (Blocking Release):
- Production bugs (#130, #129) - **Fix immediately**
- Invitation system (28 tests) - **Most complex feature area**
- Time estimate: 8 weeks is aggressive for this scope

### MEDIUM RISK:
- User management SSH keys - **Requires backend work**
- Profile import/export - **File handling complexity**

### LOW RISK:
- Validation error display - **Quick win**
- Project details view - **Straightforward UI work**

---

## Recommendations

### Option A: Full v0.5.16 (8 weeks)
Implement all 57 skipped tests as planned. High quality but long timeline.

### Option B: MVP v0.5.16 (4 weeks)
Focus on critical features only:
- Production bugs (#130, #129)
- Validation (#307)
- Projects (#308, #314) - 10 tests
- Basic Users (#309) - 3 tests
- Defer invitations to v0.5.17

**Deliverable**: ~20 tests activated, stable production

### Option C: Phased Approach (Recommended)
- **v0.5.16 (4 weeks)**: Production bugs + Projects + Users (23 tests)
- **v0.5.17 (4 weeks)**: Invitation system (28 tests)
- **v0.5.18 (2 weeks)**: Profiles + polish (6 tests)

**Benefit**: Smaller, more manageable releases with faster feedback cycles.

---

## Next Steps

1. **User Decision Required**: Choose Option A, B, or C for v0.5.16 scope
2. **Create v0.5.16 Milestone**: Assign chosen issues
3. **Update Issue #315**: Reflect chosen scope and timeline
4. **Begin Implementation**: Start with production bugs first
5. **Weekly Progress Reviews**: Track against timeline

---

## Related Documents

- `/cmd/prism-gui/frontend/SKIPPED_TESTS_ANALYSIS.md` - Detailed test breakdown
- GitHub Issue #315 - Epic tracking
- `/docs/ROADMAP.md` - Overall project roadmap
