# Comprehensive Integration Testing Strategy for Prism v0.5.16+

## Executive Summary

**Goal**: Create meticulous, repeatable integration tests for CLI and GUI using real AWS that serve as regression and acceptance tests before each release.

**Current Status**: ✅ Good foundation, ⚠️ Major gaps in coverage
**Priority**: **HIGHEST** - Blocking for release confidence

---

## Current State Analysis

### ✅ What We Have
1. **Test Harness** (`test/integration/helpers.go`):
   - Daemon lifecycle management
   - API client integration
   - Resource tracking and cleanup
   - Test isolation (temporary state directories)
   - Timeout management

2. **Implemented Tests**:
   - `TestSoloResearcherPersona` - Complete end-to-end workflow
   - `TestS3TransferIntegration` - S3 file operations
   - Basic instance lifecycle (launch, stop, hibernate, delete)
   - Cost tracking validation

3. **Infrastructure**:
   - AWS profile configuration
   - Daemon auto-start
   - Polling-based state verification
   - Automatic cleanup

### ⚠️ Critical Gaps

#### 1. **NO CLI BINARY TESTING**
**Problem**: Tests use API client directly, never execute actual `prism` CLI commands
**Impact**: User-facing CLI could be broken even if API tests pass
**Risk**: HIGH - CLI is primary user interface

#### 2. **NO GUI TESTING**
**Problem**: Zero GUI test coverage
**Impact**: GUI features completely untested with real AWS
**Risk**: HIGH - GUI is second primary interface

#### 3. **LIMITED FEATURE COVERAGE**
**Missing**:
- ❌ Template validation (only 1 of 5+ templates tested)
- ❌ Storage creation/attachment (EFS, EBS)
- ❌ Research user workflows
- ❌ Project management
- ❌ Budget tracking and alerts
- ❌ Multi-user collaboration
- ❌ Hibernation policy configuration
- ❌ Profile management
- ❌ AMI building

#### 4. **NO CONNECTIVITY VERIFICATION**
**Problem**: Tests launch instances but never verify connectivity
**Missing**:
- ❌ SSH connectivity
- ❌ Web service accessibility (Jupyter, RStudio, etc.)
- ❌ Port forwarding verification
- ❌ Actual user workflow simulation

#### 5. **NO SERVICE VALIDATION**
**Problem**: Doesn't verify services actually start and are usable
**Missing**:
- ❌ Jupyter notebook loads
- ❌ RStudio interface accessible
- ❌ VS Code server running
- ❌ Conda environments functional

#### 6. **INCOMPLETE PERSONA COVERAGE**
- ✅ Solo Researcher (implemented)
- ❌ Lab Environment (planned, not implemented)
- ❌ University Class (planned, not implemented)
- ❌ Conference Workshop (planned, not implemented)
- ❌ Cross-Institutional (planned, not implemented)

---

## Comprehensive Testing Requirements

### Phase 1: CLI Integration Tests (v0.5.17 - 2 weeks)
**Priority**: CRITICAL - Must validate actual user commands

#### 1.1 CLI Test Harness
**File**: `test/integration/cli_tests.go`
**Approach**: Execute real `prism` binary, parse output

```go
type CLIContext struct {
    PrismBin   string // Path to prism binary
    ConfigDir  string // Isolated config for test
    OutputLog  *bytes.Buffer
}

func (c *CLIContext) RunCommand(args ...string) (string, error) {
    cmd := exec.Command(c.PrismBin, args...)
    // Capture output, check exit code
}
```

#### 1.2 Core CLI Commands to Test
**Launch & Lifecycle**:
- [ ] `prism workspace launch <template> <name> --size M` - All templates
- [ ] `prism workspace list` - Verify output formatting
- [ ] `prism workspace stop <name>` - Stop instance
- [ ] `prism workspace start <name>` - Start instance
- [ ] `prism workspace hibernate <name>` - Hibernate (if supported)
- [ ] `prism workspace resume <name>` - Resume from hibernation
- [ ] `prism workspace delete <name>` - Delete instance
- [ ] `prism workspace connect <name>` - Get SSH command

**Storage**:
- [ ] `prism storage create <name>` - Create EFS
- [ ] `prism storage attach <vol> <instance>` - Attach EFS
- [ ] `prism storage detach <vol> <instance>` - Detach EFS
- [ ] `prism storage delete <name>` - Delete EFS
- [ ] `prism volume create <name> --size 100` - Create EBS
- [ ] `prism volume attach <vol> <instance>` - Attach EBS
- [ ] `prism volume detach <vol> <instance>` - Detach EBS
- [ ] `prism volume delete <name>` - Delete EBS

**Templates**:
- [ ] `prism templates` - List templates
- [ ] `prism templates info <template>` - Template details
- [ ] `prism templates validate` - Validate all templates

**Research Users**:
- [ ] `prism user create <username>` - Create research user
- [ ] `prism user list` - List research users
- [ ] `prism user provision <user> <instance>` - Provision on instance
- [ ] `prism user ssh-key <user>` - Get SSH key
- [ ] `prism user delete <user>` - Delete research user

**Projects & Budgets**:
- [ ] `prism project create <name>` - Create project
- [ ] `prism project list` - List projects
- [ ] `prism project budget set <project> <amount>` - Set budget
- [ ] `prism project invite <project> <email>` - Invite member
- [ ] `prism project costs <project>` - Get cost breakdown

**Profiles**:
- [ ] `prism profile list` - List AWS profiles
- [ ] `prism profile set <profile>` - Switch profile
- [ ] `prism profile current` - Show current profile

**Daemon**:
- [ ] `prism daemon status` - Check daemon status
- [ ] `prism daemon start` - Start daemon
- [ ] `prism daemon stop` - Stop daemon
- [ ] `prism daemon logs` - View daemon logs

**Estimated Effort**: 3-4 days implementation + 1-2 days debugging

---

### Phase 2: Template Validation Tests (v0.5.17 - 1 week)
**Priority**: HIGH - Ensure all templates actually work

#### 2.1 Template Test Matrix
Test EVERY template with real AWS launch:

**Basic Templates**:
- [ ] `Ubuntu Workstation` - Verify APT, basic tools
- [ ] `Rocky Linux 9 Base` - Verify DNF, system tools
- [ ] `Rocky Linux 9 + Conda Stack` - Verify Conda, ML packages

**ML/Data Science Templates**:
- [ ] `Python Machine Learning` - Verify Jupyter, TensorFlow, PyTorch
- [ ] `R Research Environment` - Verify RStudio, tidyverse, ggplot2
- [ ] `Julia Data Science` - Verify Julia REPL, packages

**Bioinformatics Templates** (if available):
- [ ] Verify bioinformatics tools (BLAST, BWA, SAMtools, etc.)

**Desktop Templates** (when implemented):
- [ ] `MATLAB Workstation` - Verify MATLAB GUI via Nice DCV
- [ ] `QGIS Desktop` - Verify QGIS GUI via Nice DCV

#### 2.2 Template Validation Checklist
For EACH template, verify:
1. **Launch Success**: Instance reaches running state
2. **User Creation**: Template-specified users exist
3. **Package Installation**: All declared packages installed
4. **Port Accessibility**: Declared ports are open
5. **Service Startup**: Web services (Jupyter, RStudio) start
6. **Environment Variables**: Template env vars set correctly
7. **File Permissions**: Correct ownership and permissions
8. **Resource Sizing**: Correct instance type for size (S/M/L)

**Estimated Effort**: 3-4 days (1-2 templates per day with thorough validation)

---

### Phase 3: Storage Integration Tests (v0.5.17 - 1 week)
**Priority**: HIGH - Storage is core functionality

#### 3.1 EFS Tests
**File**: `test/integration/efs_tests.go`

- [ ] **Create EFS Volume**
  - Verify EFS filesystem created in AWS
  - Verify mount target in correct VPC/subnet
  - Verify security groups allow NFS traffic

- [ ] **Attach EFS to Instance**
  - Verify EFS mounted at correct path
  - Verify write permissions
  - Create test file, verify persistence

- [ ] **Multi-Instance EFS Sharing**
  - Attach same EFS to 2+ instances
  - Write from instance A, read from instance B
  - Verify concurrent access

- [ ] **Detach EFS**
  - Verify cleanly unmounted
  - Verify instance still functional

- [ ] **Delete EFS**
  - Verify deletion in AWS
  - Verify mount targets cleaned up

#### 3.2 EBS Tests
**File**: `test/integration/ebs_tests.go`

- [ ] **Create EBS Volume**
  - Verify volume created with correct size/type
  - Verify in correct AZ

- [ ] **Attach EBS to Instance**
  - Verify attached as block device
  - Format filesystem (if needed)
  - Mount and verify writable

- [ ] **Resize EBS Volume**
  - Increase volume size
  - Verify expansion successful

- [ ] **Detach EBS**
  - Verify cleanly unmounted
  - Verify data persists after detach

- [ ] **Snapshot EBS**
  - Create snapshot
  - Create new volume from snapshot
  - Verify data integrity

- [ ] **Delete EBS**
  - Verify deletion in AWS

**Estimated Effort**: 4-5 days (EFS and EBS comprehensive testing)

---

### Phase 4: Connectivity & Service Validation (v0.5.18 - 1.5 weeks)
**Priority**: HIGH - Must verify services actually work

#### 4.1 SSH Connectivity Tests
**File**: `test/integration/connectivity_tests.go`

- [ ] **SSH Connection**
  - Parse SSH connection info
  - Establish SSH connection
  - Execute remote command
  - Verify output

- [ ] **SSH Key Management**
  - Verify SSH keys deployed correctly
  - Test password-less authentication
  - Verify key permissions (0600)

- [ ] **Multi-User SSH**
  - SSH as template user
  - SSH as research user
  - Verify user isolation

#### 4.2 Web Service Validation
**File**: `test/integration/services_tests.go`

**Jupyter Notebook**:
- [ ] Launch Python ML template
- [ ] Verify Jupyter port (8888) accessible
- [ ] HTTP GET to Jupyter URL
- [ ] Verify Jupyter UI loads
- [ ] Execute simple notebook via API
- [ ] Verify Python kernel responsive

**RStudio Server**:
- [ ] Launch R Research template
- [ ] Verify RStudio port (8787) accessible
- [ ] HTTP GET to RStudio URL
- [ ] Verify RStudio login page
- [ ] (Optional) Authenticate and execute R code

**VS Code Server**:
- [ ] Launch template with VS Code
- [ ] Verify VS Code port accessible
- [ ] HTTP GET to VS Code URL
- [ ] Verify VS Code UI loads

**Custom Web Services**:
- [ ] Test user-defined web services
- [ ] Verify port forwarding
- [ ] Test multiple services per instance

**Estimated Effort**: 5-6 days (multiple service types, thorough validation)

---

### Phase 5: Multi-User & Collaboration Tests (v0.5.18 - 1 week)
**Priority**: MEDIUM-HIGH - Key differentiator

#### 5.1 Research User Workflows
**File**: `test/integration/research_user_tests.go`

- [ ] **Create Research Users**
  - Create 3 research users
  - Verify UID/GID consistency
  - Verify home directories created

- [ ] **Provision Users on Instances**
  - Provision user on existing instance
  - Verify user account created remotely
  - Verify SSH keys deployed

- [ ] **User SSH Access**
  - SSH as research user
  - Verify home directory accessible
  - Verify proper shell environment

- [ ] **EFS Home Directories**
  - Verify EFS-backed home dirs
  - Write file, terminate instance, launch new, verify persistence

- [ ] **Multi-User Collaboration**
  - Multiple users on same instance
  - Shared EFS volume
  - Verify file permissions and isolation

#### 5.2 Lab Environment Scenario
**File**: `test/integration/lab_environment_test.go`

Based on `02_LAB_ENVIRONMENT_WALKTHROUGH.md`:
- [ ] Launch 3 workspaces for team
- [ ] Create shared EFS volume
- [ ] Attach EFS to all workspaces
- [ ] Create 3 research users
- [ ] Provision users across all workspaces
- [ ] Test cross-instance file sharing
- [ ] Verify user isolation
- [ ] Cleanup all resources

**Estimated Effort**: 4-5 days

---

### Phase 6: Project & Budget Management Tests (v0.5.18 - 1 week)
**Priority**: MEDIUM - Enterprise feature

#### 6.1 Project Management
**File**: `test/integration/project_tests.go`

- [ ] **Create Project**
  - Create project with metadata
  - Verify project ID generated
  - Verify stored in database

- [ ] **Project Members**
  - Add member as Owner
  - Add member as Admin
  - Add member as Viewer
  - Verify role permissions

- [ ] **Project Invitations**
  - Send email invitation
  - Accept invitation
  - Verify member added

- [ ] **Shared Tokens**
  - Create shared token
  - Redeem token
  - Verify member added via token

#### 6.2 Budget Tracking
**File**: `test/integration/budget_tests.go`

- [ ] **Set Project Budget**
  - Set budget amount
  - Verify budget stored
  - Verify alerts configured

- [ ] **Cost Tracking**
  - Launch instances in project
  - Verify costs accumulate
  - Verify cost breakdown

- [ ] **Budget Alerts**
  - Exceed warning threshold (80%)
  - Verify alert triggered
  - Exceed critical threshold (95%)
  - Verify actions taken

- [ ] **Cost Optimization**
  - Test hibernation savings tracking
  - Verify projected costs
  - Verify days until exhausted

**Estimated Effort**: 4-5 days

---

### Phase 7: GUI Integration Tests (v0.5.19 - 2 weeks)
**Priority**: HIGH - Second primary interface

#### 7.1 GUI Test Approach
**Challenge**: GUI testing is complex
**Solution**: Test via daemon API (GUI talks to daemon)

**File**: `test/integration/gui_tests.go`

#### 7.2 GUI Test Scenarios
Simulate GUI workflows via API:

**Dashboard**:
- [ ] Load dashboard data
- [ ] Verify all instance cards display
- [ ] Verify status indicators correct

**Template Selection**:
- [ ] Load template list
- [ ] Filter by category
- [ ] Search templates
- [ ] Get template details

**Instance Launch**:
- [ ] Launch via API (simulating GUI)
- [ ] Monitor launch progress
- [ ] Verify success notification

**Instance Management**:
- [ ] Stop/start/hibernate/resume
- [ ] Verify state transitions
- [ ] Test bulk operations

**Storage Management**:
- [ ] Create EFS/EBS via GUI workflow
- [ ] Attach to instances
- [ ] Verify in GUI list

**Project Management**:
- [ ] Create project via GUI
- [ ] Invite members
- [ ] Set budgets
- [ ] View cost analytics

**Settings**:
- [ ] Profile switching
- [ ] Configuration updates
- [ ] Verify persistence

#### 7.3 Optional: Headless Browser Testing
For true end-to-end GUI testing:
- Use Playwright or Selenium
- Launch actual GUI application
- Automate clicks and form fills
- Verify visual elements

**Estimated Effort**: 7-10 days (API testing) or 14+ days (full browser automation)

---

### Phase 8: End-to-End Persona Tests (v0.5.20 - 2 weeks)
**Priority**: MEDIUM - Comprehensive validation

Implement all remaining persona tests from USER_SCENARIOS:

#### 8.1 Lab Environment Persona
**File**: `test/integration/personas_test.go`
- Multi-user setup
- Shared storage
- Team collaboration
- Cost tracking per user

#### 8.2 University Class Persona
- Bulk launch (25 instances)
- Student access management
- Uniform policies
- Cost per student

#### 8.3 Conference Workshop Persona
- Rapid deployment
- Time-limited workspaces
- Auto-termination
- Public access

#### 8.4 Cross-Institutional Persona
- Multi-profile setup
- Cross-account EFS
- Budget tracking per institution
- Data sharing workflows

**Estimated Effort**: 8-10 days (all 4 personas)

---

## Testing Infrastructure Improvements

### 1. Enhanced Test Harness

#### CLI Test Harness
```go
// test/integration/cli_harness.go
type CLITestContext struct {
    *TestContext // Embed existing context
    PrismBin     string
    ConfigDir    string
}

func (c *CLITestContext) Prism(args ...string) *CLIResult {
    // Execute prism command, capture output
}

type CLIResult struct {
    Stdout   string
    Stderr   string
    ExitCode int
}

func (r *CLIResult) AssertSuccess(t *testing.T)
func (r *CLIResult) AssertContains(t *testing.T, substring string)
func (r *CLIResult) ParseJSON(t *testing.T, v interface{})
```

#### Connectivity Test Harness
```go
// test/integration/connectivity_harness.go
type ConnectivityTester struct {
    Instance *types.Instance
    SSHKey   string
}

func (c *ConnectivityTester) SSHConnect() (*ssh.Client, error)
func (c *ConnectivityTester) SSHExec(command string) (string, error)
func (c *ConnectivityTester) HTTPGet(port int, path string) (*http.Response, error)
func (c *ConnectivityTester) VerifyPort(port int) error
```

#### Service Validation Harness
```go
// test/integration/service_harness.go
type ServiceValidator struct {
    Instance *types.Instance
}

func (s *ServiceValidator) VerifyJupyter() error
func (s *ServiceValidator) VerifyRStudio() error
func (s *ServiceValidator) VerifyVSCode() error
func (s *ServiceValidator) VerifyCustomService(port int) error
```

### 2. Test Data Management

**Test Fixtures**:
```
test/integration/fixtures/
├── templates/           # Test template definitions
├── projects/            # Test project configurations
├── users/               # Test user definitions
└── expected_outputs/    # Expected CLI outputs for validation
```

### 3. Parallel Test Execution

**Strategy**: Run independent tests in parallel to reduce total time
```go
func TestAllTemplates(t *testing.T) {
    templates := []string{"python-ml", "r-research", "ubuntu-ws"}

    for _, template := range templates {
        template := template // Capture for goroutine
        t.Run(template, func(t *testing.T) {
            t.Parallel() // Run in parallel
            // Test template
        })
    }
}
```

**Estimated savings**: 3x-5x faster test execution

### 4. Cost Management

**Problem**: Integration tests cost money (EC2, EFS, EBS, data transfer)

**Solutions**:
1. **Resource Cleanup**: Aggressive cleanup after every test
2. **Spot Instances**: Use spot for non-critical tests
3. **Small Instance Types**: Default to t3.micro/t3.small
4. **Time Limits**: Max test duration 30 minutes
5. **Nightly Runs**: Run full suite nightly, not on every commit
6. **PR Testing**: Run subset of critical tests on PRs

**Estimated Cost**: $5-15/day for nightly full suite

### 5. Test Reporting

**HTML Report Generation**:
```go
// test/integration/reporter.go
type TestReport struct {
    StartTime    time.Time
    Duration     time.Duration
    TestsPassed  int
    TestsFailed  int
    TestsSkipped int
    AWSCosts     float64
    Details      []TestDetail
}

func GenerateHTMLReport(report *TestReport) string
```

**Output**: `test-reports/integration-YYYY-MM-DD-HHMMSS.html`

---

## Test Execution & CI/CD Integration

### Local Development
```bash
# Run all integration tests (requires AWS)
make test-integration

# Run specific test
go test -v ./test/integration -run TestSoloResearcherPersona

# Run with timeout
go test -timeout 30m -v ./test/integration

# Skip integration tests
go test -short ./...
```

### GitHub Actions Workflow
```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests

on:
  schedule:
    - cron: '0 2 * * *'  # Nightly at 2 AM
  workflow_dispatch:      # Manual trigger

jobs:
  integration:
    runs-on: ubuntu-latest
    timeout-minutes: 120

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Configure AWS
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-west-2

      - name: Build binaries
        run: make build

      - name: Run integration tests
        run: |
          go test -v -timeout 120m ./test/integration \
            -tags=integration \
            > test-output.log 2>&1

      - name: Upload test report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: integration-test-report
          path: test-reports/

      - name: Cleanup AWS resources
        if: always()
        run: ./scripts/cleanup-test-resources.sh
```

### Pre-Release Checklist
```markdown
## v0.5.17 Release Checklist

### Integration Tests (MUST PASS)
- [ ] TestSoloResearcherPersona - PASS
- [ ] All CLI commands tested - PASS
- [ ] All templates validated - PASS (5/5)
- [ ] EFS storage tests - PASS
- [ ] EBS volume tests - PASS
- [ ] SSH connectivity verified - PASS
- [ ] Jupyter service validated - PASS
- [ ] RStudio service validated - PASS
- [ ] Research user workflows - PASS
- [ ] Project management - PASS
- [ ] Budget tracking - PASS
- [ ] GUI API workflows - PASS

### Manual Verification
- [ ] GUI launches and connects to daemon
- [ ] All templates appear in GUI
- [ ] Instance cards display correctly
- [ ] Cost analytics visible

### Performance
- [ ] Launch time < 5 minutes
- [ ] Dashboard load < 3 seconds
- [ ] No memory leaks (daemon runs 24h+)

**Release Approved**: ☐ YES  ☐ NO
```

---

## Implementation Plan

### Sprint 1: CLI & Template Tests (2 weeks)
**v0.5.17 Target**
- Week 1: CLI test harness + core commands
- Week 2: Template validation matrix

**Deliverables**:
- ✅ CLI test harness implemented
- ✅ All CLI commands tested
- ✅ All templates validated with real AWS
- ✅ Test reports generated

### Sprint 2: Storage & Connectivity (2 weeks)
**v0.5.18 Target**
- Week 1: EFS/EBS comprehensive tests
- Week 2: SSH + service validation

**Deliverables**:
- ✅ Storage tests complete
- ✅ SSH connectivity verified
- ✅ Web services validated

### Sprint 3: Multi-User & Projects (2 weeks)
**v0.5.18 Target**
- Week 1: Research user workflows
- Week 2: Project/budget management

**Deliverables**:
- ✅ Multi-user tests complete
- ✅ Lab environment scenario working
- ✅ Budget tracking validated

### Sprint 4: GUI & Personas (2 weeks)
**v0.5.19 Target**
- Week 1: GUI API testing
- Week 2: Remaining persona tests

**Deliverables**:
- ✅ GUI workflows tested
- ✅ All persona tests implemented
- ✅ CI/CD integration complete

### Sprint 5: Polish & Automation (1 week)
**v0.5.20 Target**
- Test stabilization
- Performance optimization
- Documentation

**Deliverables**:
- ✅ 100% test pass rate
- ✅ HTML reporting
- ✅ Pre-release checklist automated

---

## Success Metrics

### Coverage Goals
- **CLI Commands**: 100% of user-facing commands tested
- **Templates**: 100% of templates validated
- **Features**: 90%+ of documented features tested
- **Personas**: 100% of user scenarios tested (5/5)
- **API Endpoints**: 95%+ endpoint coverage

### Quality Goals
- **Test Pass Rate**: 100% (no flaky tests)
- **Test Execution Time**: < 2 hours for full suite
- **AWS Cost**: < $20/day for nightly runs
- **Documentation**: Every test documented with purpose

### Release Confidence
- **Pre-Release**: No releases without 100% integration test pass
- **Regression**: Zero regressions in tested functionality
- **User Confidence**: Tests mirror actual user workflows

---

## Appendix A: Test Naming Convention

```
TestSoloResearcherPersona         # Persona-based test
TestCLI_Launch                    # CLI command test
TestCLI_Launch_PythonML           # CLI with specific template
TestTemplate_PythonML             # Template validation
TestTemplate_PythonML_Jupyter     # Template service validation
TestEFS_CreateAttachDelete        # Storage test
TestResearchUser_Provision        # Research user test
TestProject_BudgetTracking        # Project management test
TestGUI_LaunchWorkflow            # GUI workflow test
```

---

## Appendix B: Quick Start Guide

### Running Your First Integration Test

1. **Prerequisites**:
```bash
# Build binaries
make build

# Verify AWS credentials
aws sts get-caller-identity --profile aws

# Set test profile
export PRISM_TEST_PROFILE=aws
```

2. **Run Single Test**:
```bash
go test -v ./test/integration \
  -run TestSoloResearcherPersona \
  -timeout 30m
```

3. **Review Results**:
```bash
# Check test output
cat test-output.log

# Review HTML report (if generated)
open test-reports/latest.html
```

4. **Cleanup** (if test fails):
```bash
# List test resources
./bin/prism workspace list --profile aws

# Delete test instances
./bin/prism workspace delete test-* --profile aws
```

---

## Appendix C: AWS Resource Naming

All test resources must follow naming convention:
```
test-<test-name>-<timestamp>
```

Examples:
- `test-solo-researcher-1699564800`
- `test-efs-shared-1699564801`
- `test-python-ml-template-1699564802`

This allows:
1. Easy identification of test resources
2. Automated cleanup scripts
3. Cost tracking per test
4. Resource leak detection

---

## Appendix D: Cleanup Automation

**Script**: `scripts/cleanup-test-resources.sh`

```bash
#!/bin/bash
# Cleanup test resources older than 24 hours

PROFILE="${PRISM_TEST_PROFILE:-aws}"
CUTOFF=$(date -v-24H +%s)

# List all test instances
./bin/prism workspace list --profile "$PROFILE" | grep "^test-" | while read name; do
    # Extract timestamp from name
    timestamp=$(echo "$name" | grep -oE '[0-9]{10}$')

    if [ "$timestamp" -lt "$CUTOFF" ]; then
        echo "Deleting old test instance: $name"
        ./bin/prism workspace delete "$name" --profile "$PROFILE"
    fi
done

# Cleanup test EFS volumes
# Cleanup test EBS volumes
# etc.
```

Run after every test suite execution to prevent resource leaks.

---

**Document Version**: 1.0
**Last Updated**: 2025-11-13
**Next Review**: v0.5.17 Sprint Planning
