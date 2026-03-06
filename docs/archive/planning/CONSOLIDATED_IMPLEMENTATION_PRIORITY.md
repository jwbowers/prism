# Consolidated Implementation Priority List
**Date**: October 7, 2025
**Status**: Post-Phase 1 Analysis - 22/145 placeholders eliminated (15%)
**Goal**: Complete ALL remaining unimplemented features - Zero placeholders

---

## Executive Summary

**Current State**:
- ✅ Phase 1 Critical (22 placeholders) - COMPLETE
- 🔄 Phase 2 High Priority (38 placeholders) - 0% complete
- 🔄 Phase 3 Medium Priority (48 placeholders) - 0% complete
- 🔄 Phase 5 Planned Features (v0.5.3-0.5.5) - Not started

**Total Remaining Work**: ~123 placeholders + Phase 5 planned features + 100+ AWS integration tests

---

## Priority 1: IMMEDIATE - User-Facing Feature Gaps (10-15 hours)

These affect core user workflows and should be fixed immediately:

### 1.1 Template System Completions (HIGH IMPACT) ⭐⭐⭐
**Files**: pkg/templates/executor.go, pkg/templates/marketplace_validator.go
**Issues**:
- Line 260 in executor.go: "systems Manager executor not implemented (placeholder)"
- Dependency validation incomplete
- Port opening placeholder

**Impact**: Templates cannot execute via Systems Manager, dependency validation incomplete
**Effort**: 4 hours
**Priority**: IMMEDIATE - blocks template execution feature

### 1.2 Budget Command Flag Parsing (MEDIUM IMPACT) ⭐⭐
**File**: internal/cli/app.go:1158
**Issue**: Budget command flags (--monthly-limit, --daily-limit, --alert) not parsed

**Impact**: Legacy CLI budget commands non-functional (Cobra migration needed)
**Effort**: 2 hours
**Priority**: HIGH - budget feature gap

### 1.3 User Management Completions (MEDIUM IMPACT) ⭐⭐
**Files**: pkg/usermgmt/types.go:504, internal/cli/research_user_cobra.go
**Issues**:
- Authentication returns unimplemented (line 504)
- Update user API integration missing

**Impact**: Research user authentication non-functional, update operations incomplete
**Effort**: 3 hours
**Priority**: HIGH - security feature gap

### 1.4 Instance Launch Flag Integration (LOW-MEDIUM IMPACT) ⭐
**File**: internal/cli/instance_impl.go:268
**Issue**: Cobra flag integration note

**Impact**: CLI migration incomplete, affects launch command consistency
**Effort**: 2 hours
**Priority**: MEDIUM - migration task

---

## Priority 2: HIGH - Infrastructure Completions (15-20 hours)

Core infrastructure that affects system reliability and functionality:

### 2.1 Connection Reliability Enhancements (MEDIUM IMPACT) ⭐⭐
**Files**: pkg/connection/reliability.go, daemon_client.go, manager.go
**Issues**:
- Line 291: "In production, you'd maintain a sliding window" (connection history)
- Line 238: "In production, use proper URL parsing"
- Line 224: "For now, port availability is sufficient"

**Impact**: Connection tracking simplified, URL parsing basic, service detection incomplete
**Effort**: 4 hours
**Priority**: MEDIUM - affects reliability features

### 2.2 Budget Tracker Notifications (MEDIUM IMPACT) ⭐⭐
**File**: pkg/project/budget_tracker.go:818, 831, 843
**Issues**: Email, Slack, webhook alerts currently just log

**Impact**: Budget alerts don't actually notify users - logs only
**Effort**: 6 hours (requires email/Slack/webhook integration)
**Priority**: MEDIUM - reduces budget alert effectiveness

### 2.3 Storage System Completions (LOW-MEDIUM IMPACT) ⭐
**Files**: pkg/storage/s3_manager.go, pkg/storage/fsx_manager.go (likely)
**Issues**: Additional storage optimizations and FSx mount commands

**Impact**: Storage features work but with simplified implementations
**Effort**: 5 hours
**Priority**: LOW-MEDIUM - nice to have optimizations

---

## Priority 3: MEDIUM - AWS Integration Gaps (20-25 hours)

AWS-specific features that enhance functionality but don't block core workflows:

### 3.1 AMI System Completions (MEDIUM IMPACT) ⭐⭐
**Files**: pkg/aws/manager.go, ami_integration.go, ami_resolver.go, ami_cache.go
**Issues**:
- Line 3431: AMI override functionality
- Lines 3487-3488: Placeholder implementation
- Lines 219-220: AMI-based launching
- Line 328-329: AMI ID extraction simulation
- Lines 374-386: Placeholder AWS integration methods
- Line 271: Sophisticated sorting placeholder

**Impact**: AMI features partially functional, some edge cases unhandled
**Effort**: 8 hours
**Priority**: MEDIUM - AMI system enhancements

### 3.2 Idle Detection Cost Integration (LOW IMPACT) ⭐
**File**: pkg/daemon/idle_handlers.go:154
**Issue**: Cost tracking integration deferred

**Impact**: Idle cost tracking not integrated (marked as deferred)
**Effort**: 3 hours (once cost tracking implemented)
**Priority**: LOW - depends on cost tracking system

### 3.3 AWS Compliance Handlers (LOW IMPACT) ⭐
**File**: pkg/daemon/aws_compliance_handlers.go
**Issues**: Likely has placeholders for compliance checks

**Impact**: Compliance features simplified
**Effort**: 4 hours
**Priority**: LOW - nice to have for institutional deployments

---

## Priority 4: MARKETPLACE - DynamoDB Integration (25-30 hours)

Large architectural work for marketplace backend:

### 4.1 Marketplace DynamoDB Backend (HIGH COMPLEXITY) ⭐⭐
**File**: pkg/marketplace/registry.go
**Issues**: 15 placeholders - ALL marketplace operations use mock data
- Lines 33-34: "In production, this would query DynamoDB"
- Line 61: "In production, this would fetch from DynamoDB"
- Line 99: "For now, return based on recent downloads"
- Line 191: "In production, this would update DynamoDB"
- Plus 11 more DynamoDB integration points

**Impact**: MEDIUM - Marketplace functional with mock data, but not production-ready
**Effort**: 25 hours (requires DynamoDB schema design + implementation)
**Priority**: MEDIUM - Can defer to later phase if needed

**Note**: This is a Phase 6 feature - marketplace can work with local file storage for now

---

## Priority 5: PHASE 5 ROADMAP - Planned Features (60-100+ hours)

Features explicitly planned in CLAUDE.md roadmap but not yet implemented:

### 5.1 v0.5.3: Advanced Storage Integration (20-25 hours) ⭐⭐⭐
**Status**: 🔄 PLANNED (December 2025)
**Features**:
- FSx Integration: High-performance filesystem support
- S3 Mount Points: Direct S3 access from instances
- Storage Analytics: Usage patterns and cost optimization

**Impact**: HIGH - Research workflows need high-performance storage
**Effort**: 25 hours (AWS FSx SDK, S3 mount integration, analytics)
**Priority**: HIGH - Planned for Q4 2025

### 5.2 v0.5.4: Policy Framework Enhancement (15-20 hours) ⭐⭐
**Status**: 🔄 PLANNED (January 2026)
**Features**:
- Advanced Policies: Template access, resource limits, compliance rules
- Audit Logging: Comprehensive activity tracking and reporting
- Compliance Dashboards: NIST 800-171, SOC 2, institutional requirements

**Impact**: MEDIUM-HIGH - Institutional deployments need governance
**Effort**: 20 hours (policy engine expansion, audit system, compliance reporting)
**Priority**: MEDIUM-HIGH - Planned for Q1 2026

### 5.3 v0.5.5: AWS Research Services Integration (25-30 hours) ⭐⭐⭐
**Status**: 🔄 PLANNED (February 2026)
**Features**:
- EMR Studio: Big data analytics and Spark-based research
- Amazon Braket: Quantum computing research access
- SageMaker Integration: ML workflow integration (pending AWS partnership)

**Impact**: HIGH - Native AWS research tool integration expands use cases
**Effort**: 30 hours (EMR Studio integration, Braket API, SageMaker feasibility study)
**Priority**: MEDIUM-HIGH - Strategic feature for AWS ecosystem

### 5.4 Commercial Software Templates (10-15 hours) ⭐⭐⭐
**Status**: v0.5.2 planned feature
**Features**:
- Direct AMI Reference System
- AMI Auto-Discovery
- BYOL License Integration
- Commercial Template Schema
- Initial Templates: MATLAB, ArcGIS, Mathematica, Stata

**Impact**: VERY HIGH - Enables academic commercial software licensing
**Effort**: 15 hours (AMI discovery, license server config)
**Priority**: HIGH - Unblocks major academic use cases

### 5.5 Directory Sync System (15-20 hours) ⭐⭐⭐
**Status**: v0.5.5 planned feature
**Features**:
- EFS-Backed Bidirectional Sync
- Research-Optimized Rules
- Conflict Resolution
- Multi-Instance Support

**Impact**: HIGH - Google Drive/Dropbox-like experience for research
**Effort**: 20 hours (bidirectional sync, conflict resolution, EFS integration)
**Priority**: HIGH - Major UX improvement

---

## Priority 6: TESTING - AWS Integration Tests (80-120 hours)

Comprehensive AWS testing to validate all implementations:

### 6.1 CLI AWS Tests (35 tests, 40 hours) ⭐⭐⭐
**Coverage**:
- Launch instances (all template types)
- Instance lifecycle (start/stop/hibernate/resume)
- Storage operations (EFS/EBS/S3)
- Project management with AWS tags
- Budget tracking with Cost Explorer
- Research users with IAM/SSH
- Policy enforcement
- Template operations
- Marketplace with S3
- AMI operations with EC2

**Impact**: CRITICAL - Validates all CLI functionality against real AWS
**Effort**: 40 hours
**Priority**: HIGH - Required for production readiness

### 6.2 TUI AWS Tests (35 tests, 40 hours) ⭐⭐
**Coverage**:
- Dashboard functionality
- Instance management screens
- Template selection and validation
- Storage management (EFS/EBS)
- Settings configuration
- Profile management with AWS credentials

**Impact**: HIGH - Validates TUI functionality against real AWS
**Effort**: 40 hours
**Priority**: MEDIUM-HIGH - TUI must have feature parity

### 6.3 GUI AWS Tests (30 tests, 35 hours) ⭐⭐
**Coverage**:
- System tray operations
- Tabbed interface
- Instance management via GUI
- Template selection
- Storage operations

**Impact**: HIGH - Validates GUI functionality against real AWS
**Effort**: 35 hours
**Priority**: MEDIUM-HIGH - GUI must have feature parity

---

## Priority 7: LOW - Documentation & Cleanup (20-30 hours)

Final polish and documentation:

### 7.1 Security & Profile Enhancements (10 placeholders) ⭐
**Files**: pkg/security/, pkg/profile/
**Issues**: context.TODO() in tests, simplified implementations

**Impact**: LOW - Security works with simpler implementations
**Effort**: 8 hours
**Priority**: LOW - nice to have improvements

### 7.2 Web Services (3 placeholders) ⭐
**Files**: pkg/web/terminal.go, proxy.go
**Issues**: WebSocket library notes, simple REST API placeholders

**Impact**: LOW - Web features work with basic implementation
**Effort**: 3 hours
**Priority**: LOW - existing implementation sufficient

### 7.3 Test Context Cleanup (20+ instances) ⭐
**Files**: Various test files
**Issues**: context.TODO() in tests, "For now" test expectations

**Impact**: LOW - Tests run successfully
**Effort**: 4 hours
**Priority**: LOW - cleanup task

### 7.4 CLI/TUI/GUI Polish (9-12 placeholders) ⭐
**Files**: internal/cli/, internal/tui/, cmd/prism-gui/
**Issues**: Hardcoded policies, default ports, visualization placeholders

**Impact**: LOW - UIs functional with limitations
**Effort**: 6 hours
**Priority**: LOW - polish tasks

---

## Recommended Implementation Order

### Week 1-2: User-Facing Gaps + High Priority Infrastructure (30 hours)
1. ✅ **Template System Completions** (4h) - Implement Systems Manager executor
2. ✅ **Budget Command Flags** (2h) - Complete flag parsing
3. ✅ **User Management** (3h) - Implement authentication + update API
4. ✅ **Connection Reliability** (4h) - Sliding window, URL parsing improvements
5. ✅ **Budget Notifications** (6h) - Email/Slack/webhook integration
6. ✅ **AMI System** (8h) - Complete AMI override, launching, resolution
7. ✅ **Instance Launch Flags** (2h) - Cobra flag integration

### Week 3-4: Phase 5 High-Impact Features (50 hours)
8. ✅ **Commercial Software Templates** (15h) - AMI discovery, BYOL, MATLAB/ArcGIS/etc
9. ✅ **Advanced Storage (v0.5.3)** (25h) - FSx integration, S3 mounts, analytics
10. ✅ **Directory Sync (v0.5.5)** (20h) - EFS-backed bidirectional sync

### Week 5-6: AWS Research Services + Policy Framework (45 hours)
11. ✅ **Policy Framework Enhancement (v0.5.4)** (20h) - Advanced policies, audit logging
12. ✅ **AWS Research Services (v0.5.5)** (30h) - EMR Studio, Braket, SageMaker

### Week 7-10: AWS Integration Testing (115 hours)
13. ✅ **CLI AWS Tests** (40h) - All 35 test scenarios
14. ✅ **TUI AWS Tests** (40h) - All 35 test scenarios
15. ✅ **GUI AWS Tests** (35h) - All 30 test scenarios

### Week 11: Marketplace DynamoDB (Optional - 25 hours)
16. **Marketplace DynamoDB Backend** (25h) - IF time permits, otherwise defer to Phase 6

### Week 12: Polish & Documentation (15 hours)
17. ✅ **Security/Profile Polish** (8h) - Clean up placeholders
18. ✅ **Test Context Cleanup** (4h) - Remove context.TODO()
19. ✅ **Final Documentation** (3h) - Update all guides

---

## Total Effort Estimate

### Immediate Priority (Weeks 1-6): 125 hours
- User-facing gaps: 15 hours
- Infrastructure completions: 15 hours
- Phase 5 high-impact features: 95 hours

### Testing Phase (Weeks 7-10): 115 hours
- CLI/TUI/GUI AWS integration tests: 115 hours

### Optional & Polish (Weeks 11-12): 40 hours
- Marketplace DynamoDB: 25 hours (optional)
- Documentation & cleanup: 15 hours

**GRAND TOTAL**: ~280 hours of focused development

---

## Success Criteria

✅ Zero TODO markers in production code
✅ Zero placeholder implementations ("For now", "In production, this would...")
✅ Zero "not implemented" error messages
✅ All Phase 5.3-5.5 features implemented
✅ Commercial software template support complete
✅ 100+ AWS integration tests passing (CLI + TUI + GUI)
✅ Complete feature parity across CLI/TUI/GUI
✅ Production-ready code quality
✅ Comprehensive documentation

---

## Critical Path Dependencies

**Block 1**: Template System → Commercial Software Templates
- Template executor must work before commercial templates

**Block 2**: AMI System → Commercial Software Templates
- AMI discovery/override needed for commercial software

**Block 3**: Storage Completions → Directory Sync
- Storage foundation needed for sync system

**Block 4**: Policy Framework → Institutional Deployments
- Advanced policies enable institutional use cases

**Block 5**: All Implementations → AWS Integration Tests
- Cannot test what isn't implemented

---

## Notes

- **NO PLACEHOLDERS ALLOWED**: Every feature must be fully implemented
- **AWS Testing Required**: All implementations tested against real AWS (AWS_PROFILE=aws, AWS_REGION=us-west-2)
- **Feature Parity**: TUI and GUI must replicate all CLI functionality
- **Phase 5 Alignment**: Implement v0.5.3-0.5.5 roadmap features
- **Quality Over Speed**: No time pressure - completeness is the goal

---

*Last Updated: October 7, 2025*
*Progress: 22/145 placeholders eliminated (15%) + Phase 5 planning complete*
