# Prism Milestone Organization

**Last Updated**: January 24, 2026

## Overview

All open issues have been organized into prioritized, achievable release milestones following the "no new features until backlog cleared" principle.

## Milestone Structure

### Near-Term (v0.7.4-v0.7.7): Fix → Complete → Clean
**Focus**: Existing bugs, half-finished features, backlog cleanup
**Timeline**: Feb-May 2026 (4 months)
**Goal**: <20 open issues by May 1, 2026

| Milestone | Issues | Due Date | Type |
|-----------|--------|----------|------|
| v0.7.4 - Critical Fixes & Test Activation | 15 | Feb 15, 2026 | Bug fixes, test activation |
| v0.7.5 - Complete Half-Finished GUI Features | 13 | Mar 15, 2026 | Complete existing work |
| v0.7.6 - Security & Credential Fixes | 5 | Apr 15, 2026 | Security fixes |
| v0.7.7 - Backlog Cleanup & Decisions | 21 | May 1, 2026 | Final cleanup |

### Future Releases (v0.8.0+): New Features (GATED)
**Prerequisites**: All v0.7.x complete, backlog <20 issues, all tests passing

| Milestone | Issues | Due Date | Focus |
|-----------|--------|----------|-------|
| v0.8.0 - Authentication & Invitations | 10 | Jun 1, 2026 | OAuth/OIDC, RBAC, invitations |
| v0.9.0 - DCV Desktop Support | 15 | Sep 30, 2026 | MATLAB, QGIS, desktop apps |
| v0.10.0 - Template Marketplace & Community | 9 | Dec 31, 2026 | Community templates, plugins |
| v0.11.0 - Advanced Cost & Budget | 35 | Mar 31, 2027 | Cost optimization, forecasting |
| v0.12.0 - AWS Services Integration | 11 | Jun 30, 2027 | SageMaker, EMR, Braket |
| v0.13.0 - Education & Collaboration | 37 | Sep 30, 2027 | Class/workshop features |

## Release Gates

**v0.8.0+ Requirements**:
- ✅ All tests passing (v0.7.4)
- ✅ All GUI features complete (v0.7.5)
- ✅ All security bugs fixed (v0.7.6)
- ✅ Backlog <20 issues (v0.7.7)

**No new features until gates are met!**

## Next Immediate Actions

### 1. v0.7.4 Work (Starting NOW)
Priority bugs to fix:
- #455: TestHandleAMICheckFreshness timeout (NEW)
- #315: Activate 51 skipped E2E tests (BLOCKING)
- #354: Cloudscape/Playwright compatibility (BLOCKING)

### 2. Milestone Cleanup
Close completed milestones with 0 open issues:
- v0.7.0, v0.7.1, v0.7.2
- v0.6.1, v0.6.2, v0.6.3
- Phase milestones (consolidate into version milestones)

### 3. Ongoing Triage
Monitor v0.7.7 milestone - decide to implement or close each issue

## Success Metrics

**Current State**:
- ~100 open issues
- Broken/skipped tests
- Half-finished GUI features
- Unorganized backlog

**Target State (May 1, 2026)**:
- <20 open issues
- All tests passing
- Zero broken features
- Clear roadmap for v0.8+

## Version Numbering

**Pattern**: v0.x.y (never v1.x)
- Increment minor version (0.7 → 0.8 → 0.9 → 0.10 → 0.11...)
- Keep major version at 0 indefinitely
- Patch versions (0.7.1, 0.7.2) for hotfixes

## Contributing

When creating new issues:
1. Assign to appropriate milestone (or leave unassigned for triage)
2. Add relevant labels (bug, enhancement, testing, etc.)
3. Check if it fits v0.7.x (fix/complete) or v0.8+ (new feature)
4. Respect the "no new features" gate for v0.8+
