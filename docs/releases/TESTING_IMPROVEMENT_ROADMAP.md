# Testing Improvement Roadmap: v0.6.1 - v0.6.3

**Created**: December 13, 2025
**Context**: Post-v0.6.0 analysis identified critical testing gaps
**Goal**: Achieve enterprise-grade confidence in Prism's reliability

---

## 📊 Current State (v0.6.0)

### Strengths ✅
- **Critical path coverage**: 8/8 smoke tests passing
- **Multi-modal E2E**: GUI Playwright tests covering 5+ workflows
- **Real AWS testing**: Integration tests use actual AWS resources
- **Build validation**: Pre-push hooks prevent broken code

### Critical Gaps ⚠️
1. **Low coverage**: invitation (33.6%), daemon (14.9%), AMI (1.8%), sleepwake (18.0%)
2. **Recent failures**: 3 packages didn't compile before v0.6.0 (suggests tests not run regularly)
3. **Error path gaps**: Unclear how AWS failures, network issues, state corruption are handled
4. **Enterprise features untested**: Projects, budgets, invitations lack integration tests
5. **Concurrency untested**: Multi-user/multi-project scenarios not validated
6. **No chaos testing**: Haven't validated behavior under AWS outages, network failures

### Risk Assessment
- **Academic/Individual Use**: Moderate-High confidence ✅
- **Production Enterprise**: Moderate confidence ⚠️ (needs improvement)

---

## 🎯 Release Plan Overview

| Release | Focus | Timeline | Confidence Gain |
|---------|-------|----------|----------------|
| **v0.6.1** | Critical Path Hardening | 1-2 weeks | Med-High → High |
| **v0.6.2** | Enterprise Feature Reliability | 2-3 weeks | High → Very High |
| **v0.6.3** | Production Hardening & Chaos | 2-3 weeks | Very High → Production Ready |

**Total Timeline**: 5-8 weeks
**End Goal**: 70%+ coverage, comprehensive error handling, chaos-tested

---

## 🚀 v0.6.1: Critical Path Hardening

**Release Date**: Target December 27, 2025 (2 weeks)
**Focus**: Fix what could break user workflows TODAY
**Success Metric**: Core packages at 60%+ coverage, error recovery validated

### Phase 1.1: Core Package Coverage (Week 1)

**Goal**: Bring critical packages to 60%+ coverage

#### pkg/daemon: 14.9% → 60%+
**Priority**: 🔴 CRITICAL (daemon is the heart of Prism)

**Test Files to Create**:
1. `pkg/daemon/server_test.go` (300+ lines)
   - Server initialization and shutdown
   - Graceful termination
   - Port binding conflicts
   - API key authentication middleware
   - Request routing

2. `pkg/daemon/handlers_test.go` (400+ lines)
   - All REST endpoint handlers
   - Request validation
   - Error response formats
   - Concurrent request handling

3. `pkg/daemon/state_management_test.go` (200+ lines)
   - State persistence (save/load)
   - Concurrent state updates
   - State corruption recovery
   - Migration between versions

**Approach**: Table-driven tests with real HTTP requests to test server

#### pkg/project: 0% → 60%+
**Priority**: 🔴 CRITICAL (enterprise features depend on this)

**Test File to Create**:
1. `pkg/project/manager_test.go` (600+ lines)
   - CreateProject with validation
   - UpdateProject (members, metadata)
   - DeleteProject with cascade
   - ListProjects with filters
   - GetProjectSummary
   - AddMember/RemoveMember
   - Budget allocation tracking
   - State persistence

**Approach**: In-memory storage with mock AWS clients

#### pkg/invitation: 33.6% → 60%+
**Priority**: 🟡 HIGH (v0.6.0 baseline, needs expansion)

**Expand**: `pkg/invitation/manager_test.go`
- Email sending failure scenarios (400+ lines)
- Token expiration edge cases
- Concurrent invitation acceptance
- Bulk invitation with partial failures
- Resend with rate limiting
- Project member sync on acceptance

### Phase 1.2: Error Recovery Testing (Week 2)

**Goal**: Validate graceful degradation and recovery

#### AWS Error Injection Tests
**New File**: `test/integration/aws_error_injection_test.go` (500+ lines)

**Scenarios to Test**:
```go
// 1. AWS Throttling
TestEC2Throttling_RecoversWithBackoff()
TestS3Throttling_RecoversWithBackoff()

// 2. AWS Capacity Issues
TestInsufficientCapacity_FallbackInstance()
TestInsufficientCapacity_FailGracefully()

// 3. AWS Credential Issues
TestExpiredCredentials_ClearErrorMessage()
TestInvalidProfile_ClearErrorMessage()

// 4. Network Failures
TestNetworkTimeout_RetryAndRecover()
TestConnectionDrop_MidOperation()

// 5. AWS Service Outages
TestEC2ServiceUnavailable_QueueOperation()
TestEFSServiceUnavailable_FailGracefully()
```

**Integration with Existing Retry Logic** (pkg/aws/retry.go):
- Validate 5 retry attempts work correctly
- Confirm exponential backoff timing
- Test context cancellation during retries

#### State Corruption Recovery
**New File**: `test/integration/state_corruption_test.go` (300+ lines)

**Scenarios**:
```go
TestCorruptedStateJSON_RecoverWithBackup()
TestPartialWriteState_RollbackToLast()
TestMissingStateFile_InitializeNew()
TestConcurrentStateUpdates_NoCorruption()
```

**Implementation**:
- Atomic writes with temp file + rename
- State file versioning
- Automatic backups (state.json.bak)

### Phase 1.3: CLI Integration Test Suite (Week 2)

**Goal**: Validate all CLI commands work end-to-end

**New Directory**: `test/integration/cli/` with tests per command group

#### Core Command Tests
**File**: `test/integration/cli/core_commands_test.go` (400+ lines)

```go
TestCLI_Launch_FullWorkflow()           // launch → connect → stop → delete
TestCLI_List_AllFilters()               // --state, --template, --project
TestCLI_Info_AllInstanceStates()        // running, stopped, terminated
TestCLI_Connect_SSH_And_SSM()           // both connection types
TestCLI_Stop_And_Start()                // state transitions
TestCLI_Delete_WithConfirmation()       // -y flag, confirmation prompt
```

#### Project Management Tests
**File**: `test/integration/cli/project_commands_test.go` (300+ lines)

```go
TestCLI_ProjectCreate_WithBudget()
TestCLI_ProjectList_Filtering()
TestCLI_ProjectMember_AddRemove()
TestCLI_ProjectInfo_Summary()
```

#### Template Tests
**File**: `test/integration/cli/template_commands_test.go` (250+ lines)

```go
TestCLI_TemplateList_Categories()
TestCLI_TemplateInfo_Complete()
TestCLI_TemplateValidate_AllTemplates()
TestCLI_TemplateDiscovery_CustomPaths()
```

**Test Pattern**:
```go
// Use real daemon + real AWS (cleanup with fixtures)
func TestCLI_Launch_FullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Integration test requires AWS")
    }

    registry := fixtures.NewFixtureRegistry(t, awsClient)

    // Execute CLI commands via exec.Command
    output := runCLI(t, "launch", "python-ml", "test-instance")
    assert.Contains(t, output, "launched successfully")

    // Verify instance exists
    instances := runCLI(t, "list")
    assert.Contains(t, instances, "test-instance")

    // Cleanup handled by fixtures
}
```

### Phase 1.4: Improve CI/CD Test Execution (Week 2)

**Goal**: Ensure tests run on every commit

#### GitHub Actions Workflow Updates
**File**: `.github/workflows/test.yml`

```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Run Unit Tests
        run: go test ./... -short -race -coverprofile=coverage.txt

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.txt

  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          role-to-assume: ${{ secrets.AWS_TEST_ROLE }}
          aws-region: us-west-2

      - name: Run Integration Tests
        run: go test -tags integration ./test/integration/... -v
```

#### Coverage Tracking
**New File**: `scripts/coverage-check.sh`

```bash
#!/bin/bash
# Fail if coverage drops below thresholds

check_coverage() {
    local package=$1
    local threshold=$2
    local coverage=$(go test -cover $package | grep coverage | awk '{print $5}' | tr -d '%')

    if (( $(echo "$coverage < $threshold" | bc -l) )); then
        echo "❌ $package coverage ($coverage%) below threshold ($threshold%)"
        exit 1
    fi
}

check_coverage "./pkg/daemon/..." 60
check_coverage "./pkg/project/..." 60
check_coverage "./pkg/invitation/..." 60
```

---

## 🏢 v0.6.2: Enterprise Feature Reliability

**Release Date**: Target January 10, 2026 (3 weeks)
**Focus**: Multi-user, concurrent operations, enterprise workflows
**Success Metric**: Enterprise features battle-tested, concurrency validated

### Phase 2.1: Integration Test Suite for Enterprise Features (Week 1)

**Goal**: End-to-end testing of Phase 4/5 features

#### Project + Budget Integration Tests
**New File**: `test/integration/enterprise/project_budget_test.go` (600+ lines)

**Scenarios**:
```go
// Budget Allocation Workflows
TestProjectBudget_CreateAndAllocate()
TestProjectBudget_MultiSourceFunding()        // N budgets → 1 project
TestProjectBudget_SharedBudgetPool()         // 1 budget → N projects
TestProjectBudget_Reallocation()
TestProjectBudget_CostTracking_RealInstance()
TestProjectBudget_AlertThresholds()
TestProjectBudget_OverspendPrevention()

// Project Lifecycle with Budget
TestProject_CreateWithDefaultAllocation()
TestProject_LaunchInstance_DeductsFromBudget()
TestProject_DeleteWithActiveAllocations()
```

#### Invitation + Project Integration Tests
**New File**: `test/integration/enterprise/invitation_test.go` (500+ lines)

**Scenarios**:
```go
// Full Invitation Workflows
TestInvitation_SendAndAccept_AddsToProject()
TestInvitation_BulkInvite_University_Class()   // 50+ invitations
TestInvitation_QuotaCheck_PreventOversubscribe()
TestInvitation_ResearchUserProvisioning()      // SSH key, UID/GID, EFS home
TestInvitation_RoleBasedPermissions()
TestInvitation_ExpiredToken_CannotAccept()
TestInvitation_RevokePending_NotifyByEmail()
```

#### Multi-Project Collaboration Tests
**New File**: `test/integration/enterprise/collaboration_test.go` (400+ lines)

**Scenarios**:
```go
// Cross-Project Workflows
TestCollaboration_SharedStorage_MultiProject()
TestCollaboration_UserInMultipleProjects()
TestCollaboration_ProjectMemberPermissions()
TestCollaboration_BudgetVisibility()
```

### Phase 2.2: Concurrency & Race Condition Testing (Week 2)

**Goal**: Validate thread-safe operations

#### Concurrent API Access Tests
**New File**: `test/integration/concurrency/api_concurrent_test.go` (500+ lines)

**Scenarios**:
```go
// Concurrent Operations
TestConcurrent_MultipleUserLaunches()          // 10 users, 50 instances
TestConcurrent_ProjectCreation()               // 20 projects simultaneously
TestConcurrent_BudgetAllocation()              // Race on allocation updates
TestConcurrent_InvitationAcceptance()          // Multiple users accept at once
TestConcurrent_StateUpdates()                  // Daemon state updates

// Load Testing
TestLoad_100ConcurrentLaunches()               // Rate limiting validation
TestLoad_DaemonUnder50Requests()               // Response time < 200ms
```

**Test Pattern**:
```go
func TestConcurrent_MultipleUserLaunches(t *testing.T) {
    var wg sync.WaitGroup
    errors := make(chan error, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            _, err := api.LaunchInstance(ctx, &LaunchRequest{
                Template:    "python-ml",
                Name:        fmt.Sprintf("user-%d-instance", id),
                ProjectID:   "shared-project",
            })
            if err != nil {
                errors <- err
            }
        }(i)
    }

    wg.Wait()
    close(errors)

    for err := range errors {
        t.Errorf("Launch failed: %v", err)
    }
}
```

#### Race Detector Validation
**Script**: `scripts/race-check.sh`

```bash
#!/bin/bash
# Run all tests with race detector

echo "🔍 Running tests with race detector..."

# Unit tests
go test ./pkg/... -race -short

# Integration tests
go test -tags integration ./test/integration/... -race

# Specific race-prone areas
go test ./pkg/daemon/... -race -count=100  # Hammer test
go test ./pkg/project/... -race -count=100
go test ./pkg/state/... -race -count=100

echo "✅ Race detector validation complete"
```

### Phase 2.3: Performance Benchmarking (Week 3)

**Goal**: Establish performance baselines

#### API Endpoint Benchmarks
**New File**: `test/benchmark/api_bench_test.go` (300+ lines)

```go
func BenchmarkAPI_LaunchInstance(b *testing.B)
func BenchmarkAPI_ListInstances_100(b *testing.B)
func BenchmarkAPI_ListInstances_1000(b *testing.B)
func BenchmarkAPI_ProjectList_100Projects(b *testing.B)
func BenchmarkAPI_BudgetSummary(b *testing.B)
```

#### State Management Benchmarks
**New File**: `test/benchmark/state_bench_test.go` (200+ lines)

```go
func BenchmarkState_Save(b *testing.B)
func BenchmarkState_Load(b *testing.B)
func BenchmarkState_ConcurrentUpdates(b *testing.B)
```

**Performance Targets**:
- API response time: < 200ms (95th percentile)
- State save/load: < 10ms
- List operations: < 50ms for 100 items
- Concurrent launches: 10 simultaneous without errors

---

## 💥 v0.6.3: Production Hardening & Chaos Testing

**Release Date**: Target January 31, 2026 (3 weeks)
**Focus**: Edge cases, failure injection, regional testing
**Success Metric**: Survives chaos, handles all edge cases gracefully

### Phase 3.1: Chaos Engineering (Week 1)

**Goal**: Validate resilience under adverse conditions

#### Network Chaos Tests
**New File**: `test/chaos/network_chaos_test.go` (500+ lines)

**Scenarios**:
```go
// Network Failures
TestChaos_NetworkDown_MidLaunch()
TestChaos_NetworkLatency_500ms()
TestChaos_PacketLoss_20Percent()
TestChaos_DNSFailure_Recovers()

// AWS API Chaos
TestChaos_EC2APIUnavailable_5Minutes()
TestChaos_S3APIThrottling_Extreme()
TestChaos_EFSAPITimeout_Repeated()
TestChaos_IAMAPIError_PermissionDenied()

// Daemon Chaos
TestChaos_DaemonKilled_MidOperation()
TestChaos_DaemonOutOfMemory_Graceful()
TestChaos_DaemonDiskFull_StateRecovery()
```

**Implementation Using Chaos Toolkit**:
```yaml
# chaos/network-experiments.yaml
title: Network latency experiment
description: Introduce 500ms latency to AWS API calls
method:
  - type: action
    name: add-network-latency
    provider:
      type: process
      path: tc
      arguments: qdisc add dev eth0 root netem delay 500ms
```

#### AWS Service Outage Simulation
**New File**: `test/chaos/aws_outage_test.go` (400+ lines)

**Scenarios**:
```go
// Regional Outages
TestChaos_USWest2_FullOutage()             // Fallback to us-east-1
TestChaos_EC2OnlyOutage_EFSWorking()
TestChaos_EFSOutage_LaunchWithoutEFS()

// Partial Outages
TestChaos_AZ_OneUnavailable()              // Launch in different AZ
TestChaos_InstanceType_Unavailable()       // Fallback to similar type
```

### Phase 3.2: Edge Case Coverage (Week 2)

**Goal**: Handle unusual but valid scenarios

#### Template Edge Cases
**New File**: `test/edge/template_edge_test.go` (400+ lines)

```go
// Inheritance Edge Cases
TestEdge_CircularInheritance_Detected()
TestEdge_DeepInheritance_10Levels()
TestEdge_InheritFromNonExistent_Error()
TestEdge_DuplicatePackages_Deduped()

// Template Validation Edge Cases
TestEdge_EmptyTemplate_ValidationError()
TestEdge_HugeTemplate_10000Lines()
TestEdge_UnicodeInTemplate_UTF8Handled()
TestEdge_CommentOnlyTemplate_Error()

// Provisioning Edge Cases
TestEdge_ProvisionFile_5GB_Success()
TestEdge_ProvisionFile_S3NotFound_Graceful()
TestEdge_ProvisionFile_Checksum_Mismatch()
```

#### Instance Management Edge Cases
**New File**: `test/edge/instance_edge_test.go` (500+ lines)

```go
// Launch Edge Cases
TestEdge_LaunchWithoutProfile_Error()
TestEdge_Launch_InvalidInstanceType()
TestEdge_Launch_UnsupportedRegion()
TestEdge_Launch_QuotaExceeded_Clear()

// Lifecycle Edge Cases
TestEdge_StopAlreadyStopped_Idempotent()
TestEdge_DeleteAlreadyDeleted_Idempotent()
TestEdge_ConnectToTerminated_Error()
TestEdge_StopDuringLaunch_SafeAbort()

// State Edge Cases
TestEdge_InstanceVanishedFromAWS_StateCleanup()
TestEdge_ManuallyModifiedInstance_Detected()
TestEdge_OrphanedResources_Cleanup()
```

#### Multi-Region Testing
**New File**: `test/edge/regional_edge_test.go` (300+ lines)

```go
// Test in all supported regions
TestRegional_LaunchInAllRegions()          // 8 regions
TestRegional_ARMvsX86_Availability()
TestRegional_EFSSupport_PerRegion()
TestRegional_InstanceTypeAvailability()
```

### Phase 3.3: LocalStack Integration (Week 3)

**Goal**: Fast, offline testing without AWS costs

#### LocalStack Test Suite
**New File**: `test/localstack/setup.go` (200+ lines)

```go
// LocalStack initialization
func SetupLocalStack(t *testing.T) *LocalStackEnv {
    // Start LocalStack with Docker
    // Configure AWS SDK to use LocalStack endpoints
    // Return environment for tests
}
```

**New File**: `test/localstack/instance_test.go` (600+ lines)

**Scenarios**:
```go
// All core workflows against LocalStack
TestLocalStack_Launch_Complete()
TestLocalStack_Storage_EFS()
TestLocalStack_Storage_EBS()
TestLocalStack_Networking_SecurityGroups()
TestLocalStack_IAM_RoleAttachment()
```

**Benefits**:
- ⚡ Fast: No real AWS latency
- 💰 Free: No AWS costs
- 🔁 Repeatable: Deterministic results
- 🔌 Offline: Works without internet

**Integration with CI/CD**:
```yaml
# .github/workflows/test-localstack.yml
- name: Start LocalStack
  run: docker-compose -f test/localstack/docker-compose.yml up -d

- name: Run LocalStack Tests
  run: go test -tags localstack ./test/localstack/... -v
```

---

## 📈 Success Metrics

### Coverage Targets

| Package | v0.6.0 | v0.6.1 Target | v0.6.2 Target | v0.6.3 Target |
|---------|--------|---------------|---------------|---------------|
| pkg/daemon | 14.9% | **60%** | 70% | 75% |
| pkg/project | 0% | **60%** | 70% | 75% |
| pkg/invitation | 33.6% | **60%** | 70% | 75% |
| pkg/ami | 1.8% | 40% | 50% | 60% |
| pkg/sleepwake | 18.0% | 50% | 60% | 70% |
| pkg/budget | 0% | 50% | 60% | 70% |
| **Overall** | ~25% | **55%** | **65%** | **70%** |

### Reliability Metrics

| Metric | v0.6.0 | v0.6.3 Target |
|--------|--------|---------------|
| Smoke test pass rate | 100% | 100% |
| Unit test pass rate | 100% | 100% |
| Integration test pass rate | ~95% | **100%** |
| Race conditions detected | 0 | 0 |
| AWS error recovery rate | Unknown | **95%+** |
| Concurrent operation success | Unknown | **100%** |
| Chaos test survival | Unknown | **90%+** |

### Quality Gates (CI/CD)

**v0.6.1**:
- ✅ All unit tests pass with -race
- ✅ Core packages at 60%+ coverage
- ✅ CLI integration tests pass (20+ scenarios)
- ✅ No new race conditions introduced

**v0.6.2**:
- ✅ All enterprise integration tests pass
- ✅ Concurrent operation tests pass (10+ simultaneous users)
- ✅ Performance benchmarks meet targets
- ✅ 65%+ overall coverage

**v0.6.3**:
- ✅ Chaos tests: 90%+ survival rate
- ✅ Edge case tests: 100% pass rate
- ✅ LocalStack test suite: 100% pass rate
- ✅ Regional tests: All 8 regions validated
- ✅ 70%+ overall coverage

---

## 🔧 Implementation Strategy

### Development Workflow

1. **Test-First Development**:
   - Write failing test
   - Implement feature/fix
   - Validate test passes
   - Run full suite with -race

2. **Integration Test Pattern**:
   ```go
   func TestIntegration_Feature(t *testing.T) {
       if testing.Short() {
           t.Skip("Integration test")
       }
       registry := fixtures.NewFixtureRegistry(t, client)
       // Test implementation
       // Cleanup handled automatically
   }
   ```

3. **CI/CD Integration**:
   - Unit tests: Every commit
   - Integration tests: Main branch only
   - Chaos tests: Weekly schedule
   - Coverage reports: Codecov integration

### Testing Infrastructure

**New Tools to Add**:
1. **Codecov**: Coverage tracking and reporting
2. **Chaos Toolkit**: Network and service chaos injection
3. **LocalStack**: Offline AWS simulation
4. **hey/vegeta**: Load testing CLI APIs

**Scripts to Create**:
- `scripts/test-all.sh`: Run complete test suite
- `scripts/test-coverage.sh`: Generate coverage report
- `scripts/test-race.sh`: Race detector validation
- `scripts/test-chaos.sh`: Chaos engineering suite
- `scripts/test-localstack.sh`: LocalStack test suite

---

## 📦 Deliverables by Release

### v0.6.1 Deliverables
- [ ] pkg/daemon tests (900+ lines, 60%+ coverage)
- [ ] pkg/project tests (600+ lines, 60%+ coverage)
- [ ] pkg/invitation expanded tests (400+ lines, 60%+ coverage)
- [ ] AWS error injection tests (500+ lines)
- [ ] State corruption recovery tests (300+ lines)
- [ ] CLI integration test suite (950+ lines)
- [ ] GitHub Actions workflow updates
- [ ] Coverage tracking scripts
- **Total**: ~4,650 lines of test code

### v0.6.2 Deliverables
- [ ] Project+Budget integration tests (600+ lines)
- [ ] Invitation integration tests (500+ lines)
- [ ] Collaboration integration tests (400+ lines)
- [ ] Concurrency tests (500+ lines)
- [ ] Performance benchmarks (500+ lines)
- [ ] Race detector validation scripts
- **Total**: ~2,500 lines of test code

### v0.6.3 Deliverables
- [ ] Network chaos tests (500+ lines)
- [ ] AWS outage simulation (400+ lines)
- [ ] Template edge cases (400+ lines)
- [ ] Instance edge cases (500+ lines)
- [ ] Regional tests (300+ lines)
- [ ] LocalStack integration (800+ lines)
- [ ] Chaos engineering scripts
- **Total**: ~2,900 lines of test code

---

## 🎯 Final State (Post-v0.6.3)

### Test Suite Statistics
- **Total Test Lines**: ~10,000+ new lines
- **Overall Coverage**: 70%+
- **Test Count**: 300+ new tests
- **Test Categories**:
  - Unit: 200+ tests
  - Integration: 60+ tests
  - E2E: 30+ tests (existing GUI + new CLI)
  - Chaos: 20+ experiments
  - Edge: 40+ scenarios

### Confidence Level
- **Academic/Individual Use**: ✅ Very High
- **Lab/Department Deployment**: ✅ Very High
- **Enterprise Production**: ✅ High (deploy-with-confidence)

### Remaining Gaps (Post-v0.6.3)
- GUI automated testing (Playwright covers main flows)
- Cross-platform validation (macOS/Linux/Windows)
- Long-running stability tests (72+ hours)
- Security penetration testing
- Scalability testing (1000+ instances)

These gaps are acceptable for most deployments and can be addressed in v0.7.x+ releases.

---

## 📝 Notes

### Test Maintenance
- Review and update tests quarterly
- Archive obsolete tests when features change
- Keep fixture cleanup patterns consistent
- Document test patterns in TESTING.md

### Cost Considerations
- Integration tests cost ~$5-10 per full run (AWS usage)
- Run integration tests on main branch only
- Use LocalStack for development iteration
- Implement test resource tagging for cost tracking

### Team Coordination
- Assign 1 developer per release
- Code review all test additions
- Pair program chaos tests (complex)
- Document test failures thoroughly

---

**Next Steps**: Review this roadmap with team, prioritize v0.6.1 work, create GitHub milestone.
