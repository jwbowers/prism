# Prism Testing Guide

This guide covers all testing approaches used in the Prism project, from unit tests to E2E tests with real AWS resources.

## Testing Philosophy

Prism uses a **multi-layer testing strategy** to ensure reliability:

1. **Unit Tests** - Fast, isolated tests of individual functions and packages
2. **Integration Tests** - Tests that interact with the daemon API and AWS
3. **E2E Tests** - Browser-based tests of the GUI with real AWS resources
4. **CLI Tests** - Direct command-line interface testing

**Key Principle**: *"The deadline and the exact % are less important than what the tests are testing and what the tests surface."*

Tests validate **real user workflows**, not just code coverage. Integration tests verify that Prism actually delivers on its core value propositions:
- Templates provision correctly (Jupyter works, not just "instance running")
- Hibernation saves money (idle detection triggers, costs stop)
- Budgets prevent overspend (launches actually blocked)
- Teams can collaborate (RBAC enforced, shared resources work)
- Data persists safely (EFS/EBS survive instance lifecycle)
- System recovers from failures (daemon crash, AWS errors)

---

## Integration Testing Strategy

Prism's integration testing follows a **4-phase workflow-driven approach**:

### Phase 1: Critical Workflows (PRIORITY)
**Goal**: Validate core value propositions actually work

**Tests** (< 15 minutes, run on-demand):
- [#396](https://github.com/scttfrdmn/prism/issues/396) - Template Provisioning End-to-End
- [#397](https://github.com/scttfrdmn/prism/issues/397) - Idle Detection & Hibernation Flow
- [#398](https://github.com/scttfrdmn/prism/issues/398) - Budget Enforcement & Cost Tracking
- [#399](https://github.com/scttfrdmn/prism/issues/399) - Multi-User Collaboration Workflows

**Run with**: `make test-workflows`

### Phase 2: Persona-Driven Workflows
**Goal**: Validate all 5 user personas can accomplish their goals

**Tests** (< 45 minutes, run on-demand):
- [#400](https://github.com/scttfrdmn/prism/issues/400) - Solo Researcher Persona (Dr. Sarah Chen)
- [#401](https://github.com/scttfrdmn/prism/issues/401) - Lab Environment Persona (Prof. Martinez)
- [#402](https://github.com/scttfrdmn/prism/issues/402) - University Class Persona (Prof. Thompson)
- [#403](https://github.com/scttfrdmn/prism/issues/403) - Conference Workshop Persona (Dr. Patel)
- [#404](https://github.com/scttfrdmn/prism/issues/404) - Cross-Institutional Persona (Dr. Kim)

**Run with**: `make test-personas`

### Phase 3: Failure Recovery & Resilience
**Goal**: Validate graceful handling of failures

**Tests** (< 20 minutes, run on-demand):
- [#405](https://github.com/scttfrdmn/prism/issues/405) - Daemon Failure Recovery
- [#406](https://github.com/scttfrdmn/prism/issues/406) - AWS API Error Handling
- [#407](https://github.com/scttfrdmn/prism/issues/407) - Instance Crash Recovery

**Run with**: `make test-resilience`

### Phase 4: Long-Running Workflows
**Goal**: Validate extended time periods

**Tests** (hours to days, run manually):
- Week-long instance lifecycle with hibernation cycles
- Multi-day cost accumulation and forecast accuracy
- Monthly budget rollover and reporting

**Run with**: `make test-long-running`

### Execution Strategy

**On-Demand**: All integration tests run manually or on-demand (not on every commit)

**AWS Account**: Use `aws` profile with real AWS resources (no mocking)

**Test Organization**:
```
test/integration/
├── phase1_workflows/       # Critical workflows (< 15 min)
├── phase2_personas/        # User scenarios (< 45 min)
├── phase3_resilience/      # Failure recovery (< 20 min)
└── phase4_long_running/    # Extended workflows (hours/days)
```

**Success Criteria**:
- Tests catch real bugs (provisioning failures, hibernation not triggering, budget not enforced)
- Tests validate user experience (SSH verification, actual Jupyter accessibility)
- Zero resource leaks (fixture pattern cleanup)
- 95%+ pass rate (allow occasional AWS transient failures)

---

## Unit Tests

**Location**: Throughout the codebase alongside source files

```bash
# Run all unit tests
make test
go test ./...

# Run specific package tests
go test ./pkg/templates/...
go test ./pkg/daemon/...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

**Best Practices**:
- Keep unit tests fast (< 100ms each)
- Mock external dependencies (AWS, filesystem)
- Test edge cases and error conditions
- Use table-driven tests for multiple scenarios

## Integration Tests with Fixtures (Go)

**Location**: `test/fixtures/` and `test/integration/`

Prism uses a **fixture pattern** for integration tests that create real AWS resources. This pattern ensures automatic cleanup and proper resource dependency management.

### Why Fixtures?

Traditional integration tests often leave orphaned AWS resources when tests fail. The fixture pattern solves this with:
- ✅ Automatic cleanup via Go's `t.Cleanup()` mechanism
- ✅ No manual cleanup scripts or forgotten resources
- ✅ Proper dependency ordering (backups → instances → EBS → EFS)
- ✅ Consistent patterns across all integration tests

### Architecture

```
test/
├── fixtures/
│   ├── registry.go         # FixtureRegistry - cleanup orchestration
│   ├── instances.go         # Instance & backup fixtures
│   ├── storage.go           # EFS & EBS fixtures
│   └── README.md            # Detailed implementation docs
└── integration/
    ├── fixtures_example_test.go  # Example patterns
    ├── cli_profile_test.go       # CLI command tests
    ├── cli_storage_test.go       # Storage CLI tests
    └── cli_template_test.go      # Template CLI tests
```

### Quick Start Example

```go
//go:build integration
// +build integration

package integration

import (
    "testing"
    "github.com/scttfrdmn/prism/pkg/api/client"
    "github.com/scttfrdmn/prism/test/fixtures"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/assert"
)

func TestBackupWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // 1. Initialize API client
    client := client.NewClientWithOptions("http://localhost:8947", client.Options{
        AWSProfile: "aws",
        Region:     "us-west-2",
    })

    // 2. Create fixture registry - cleanup is automatic
    registry := fixtures.NewFixtureRegistry(t, client)

    // 3. Create real AWS instance - automatically registers for cleanup
    instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
        Template: "Ubuntu Basic",
        Name:     "test-instance",
        Size:     "S",
    })
    require.NoError(t, err)
    assert.Equal(t, "running", instance.State)

    // 4. Create real AWS backup - automatically registers for cleanup
    backup, err := fixtures.CreateTestBackup(t, registry, fixtures.CreateTestBackupOptions{
        InstanceID:  instance.Name,
        Name:        "test-backup",
        Description: "Integration test backup",
    })
    require.NoError(t, err)
    assert.Equal(t, "available", backup.State)

    // 5. Test your business logic...

    // 6. Cleanup happens automatically when test completes!
    //    Resources are deleted in correct order: backups → instances → EBS → EFS
}
```

### Running Integration Tests

```bash
# Run all integration tests
go test -tags integration ./test/integration/... -v

# Run specific test
go test -tags integration ./test/integration/ -run TestBackupWorkflow -v

# Skip slow tests
go test -tags integration -short ./test/integration/...
```

### Available Fixture Functions

**Instance Management**:
```go
instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
    Template: "Ubuntu Basic",  // Template name
    Name:     "test-instance", // Unique instance name
    Size:     "S",             // S, M, L, or XL
})
```

**Backup Management**:
```go
backup, err := fixtures.CreateTestBackup(t, registry, fixtures.CreateTestBackupOptions{
    InstanceID:  instance.Name,          // Source instance
    Name:        "test-backup",          // Unique backup name
    Description: "Integration test backup",
})
```

**EFS Volume Management**:
```go
volume, err := fixtures.CreateTestVolume(t, registry, fixtures.CreateTestVolumeOptions{
    Name:            "test-volume",      // Unique volume name
    PerformanceMode: "generalPurpose",   // or "maxIO"
})
```

**EBS Storage Management**:
```go
storage, err := fixtures.CreateTestEBSStorage(t, registry, fixtures.CreateTestEBSStorageOptions{
    Name:       "test-ebs",   // Unique storage name
    SizeGB:     10,           // Size in gigabytes
    VolumeType: "gp3",        // gp3, gp2, io1, etc.
})
```

### Advanced Patterns

**Multi-Resource Test Environments**:
```go
func TestCompleteEnvironment(t *testing.T) {
    client := client.NewClientWithOptions("http://localhost:8947", client.Options{
        AWSProfile: "aws",
        Region:     "us-west-2",
    })
    registry := fixtures.NewFixtureRegistry(t, client)
    ctx := context.Background()

    // Create storage first
    efsVolume, err := fixtures.CreateTestVolume(t, registry, fixtures.CreateTestVolumeOptions{
        Name: "test-efs",
    })
    require.NoError(t, err)

    ebsStorage, err := fixtures.CreateTestEBSStorage(t, registry, fixtures.CreateTestEBSStorageOptions{
        Name:   "test-ebs",
        SizeGB: 10,
    })
    require.NoError(t, err)

    // Create instance
    instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
        Template: "Ubuntu Basic",
        Name:     "test-instance",
    })
    require.NoError(t, err)

    // Attach volumes to instance
    err = client.AttachVolume(ctx, efsVolume.Name, instance.Name)
    require.NoError(t, err)

    err = client.AttachStorage(ctx, ebsStorage.Name, instance.Name)
    require.NoError(t, err)

    // Create backup of configured instance
    backup, err := fixtures.CreateTestBackup(t, registry, fixtures.CreateTestBackupOptions{
        InstanceID: instance.Name,
        Name:       "test-backup",
    })
    require.NoError(t, err)

    // Test your workflow...

    // Automatic cleanup in correct order:
    // 1. Backups
    // 2. Instances
    // 3. EBS storage
    // 4. EFS volumes
}
```

**Manual Cleanup** (optional, useful for mid-test cleanup):
```go
func TestManualCleanup(t *testing.T) {
    client := client.NewClientWithOptions("http://localhost:8947", client.Options{
        AWSProfile: "aws",
        Region:     "us-west-2",
    })
    registry := fixtures.NewFixtureRegistry(t, client)

    // Create resources
    instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
        Template: "Ubuntu Basic",
        Name:     "manual-cleanup-test",
    })
    require.NoError(t, err)

    // Do some testing...

    // Manually trigger cleanup if needed
    registry.Cleanup()

    // Verify cleanup happened
    time.Sleep(5 * time.Second)
    ctx := context.Background()
    _, err = client.GetInstance(ctx, instance.Name)
    assert.Error(t, err) // Should be deleted

    // Note: t.Cleanup() still runs but registry tracks cleanup was called
}
```

### Configuration Requirements

**AWS Configuration**:
- Profile: `aws` (default for tests)
- Region: `us-west-2` (default for tests)
- Valid AWS credentials with EC2/EFS/EBS permissions

**Daemon Requirements**:
- Daemon must be running on `http://localhost:8947`
- Tests auto-start daemon if not running (see test helpers)

**Environment Variables**:
```bash
# Optional: Override defaults
export AWS_PROFILE=aws
export AWS_REGION=us-west-2
```

### Troubleshooting

**Problem: Tests fail with "daemon not responding"**
```bash
# Start daemon manually for debugging
./bin/prismd &

# Check daemon status
./bin/prism admin daemon status
```

**Problem: Tests leave orphaned AWS resources**
```bash
# This should never happen with fixtures!
# If it does, check that tests use FixtureRegistry correctly

# Manual cleanup (last resort)
./bin/prism list
./bin/prism delete instance-name
```

**Problem: Tests timeout waiting for resources**
- Check AWS region availability (some instance types not available in all regions)
- Verify AWS credentials are valid
- Check AWS service quotas (EC2 instance limits, EBS volume limits)

## CLI Integration Tests

**Location**: `test/integration/cli_*_test.go`

These tests verify CLI commands work correctly by executing them directly.

```go
func TestProfileManagement(t *testing.T) {
    ctx := NewCLITestContext(t)

    // Test profile creation
    output, err := ctx.RunCommand("profile", "add", "test-profile",
        "--aws-profile", "my-aws-profile",
        "--region", "us-west-2")
    require.NoError(t, err)
    assert.Contains(t, output, "Profile created")
}
```

**Running CLI tests**:
```bash
go test -tags integration ./test/integration/cli_*_test.go -v
```

## E2E Tests (Playwright)

**Location**: `cmd/prism-gui/frontend/tests/e2e/`

Browser-based end-to-end tests of the GUI using Playwright with real AWS integration.

### Running E2E Tests

```bash
cd cmd/prism-gui/frontend

# Run all E2E tests
npx playwright test

# Run specific workflow
npx playwright test backup-workflows.spec.ts
npx playwright test hibernation-workflows.spec.ts
npx playwright test instance-workflows.spec.ts
npx playwright test profile-workflows.spec.ts
npx playwright test storage-workflows.spec.ts

# Run with UI
npx playwright test --ui

# Debug specific test
npx playwright test --debug
```

### E2E Test Structure

```
cmd/prism-gui/frontend/tests/
├── e2e/
│   ├── fixtures.js                      # Test fixtures and setup
│   ├── setup-daemon.js                  # Daemon auto-start
│   ├── backup-workflows.spec.ts         # Backup management tests
│   ├── hibernation-workflows.spec.ts    # Hibernation tests
│   ├── instance-workflows.spec.ts       # Instance management tests
│   ├── profile-workflows.spec.ts        # Profile management tests
│   └── storage-workflows.spec.ts        # Storage management tests
└── msw/
    └── handlers.ts                      # Mock service worker handlers
```

### Critical E2E Pattern

**IMPORTANT**: All E2E tests MUST set the onboarding modal flag before navigation to prevent blocking interactions:

```typescript
import { test, expect } from '@playwright/test';

test.describe('Your Test Suite', () => {
  test.beforeEach(async ({ page, context }) => {
    // CRITICAL: Set localStorage BEFORE navigating
    await context.addInitScript(() => {
      localStorage.setItem('cws_onboarding_complete', 'true');
    });

    // Set AWS config
    await context.addInitScript(() => {
      localStorage.setItem('cws_aws_profile', 'aws');
      localStorage.setItem('cws_aws_region', 'us-west-2');
    });

    // Now navigate
    await page.goto('http://localhost:8080');
    await page.waitForLoadState('networkidle');
  });

  test('should do something', async ({ page }) => {
    // Your test code...
  });
});
```

### E2E Configuration

**Playwright Config** (`playwright.config.ts`):
- Browser: Chromium
- Base URL: `http://localhost:8080`
- Timeout: 30 seconds per test
- Retries: 2 on CI, 0 locally

**Test Requirements**:
- GUI frontend running on port 8080
- Daemon running on port 8947
- AWS credentials configured (`aws` profile)
- AWS region: `us-west-2`

## GUI Backend Tests (Go)

**Location**: `cmd/prism-gui/*.go`

Backend tests for GUI-specific handlers and Wails integration.

```bash
go test ./cmd/prism-gui/... -v
```

**Example**:
```go
func TestBackupManagement(t *testing.T) {
    // Test GUI backend API
    app := NewTestApp(t)

    backups, err := app.ListBackups()
    require.NoError(t, err)
    assert.IsType(t, []Backup{}, backups)
}
```

## Test Configuration

### AWS Configuration

All tests use consistent AWS configuration:

**Profile**: `aws`
**Region**: `us-west-2`

Set up AWS credentials:
```bash
# Configure AWS CLI
aws configure --profile aws
# AWS Access Key ID: [your-key]
# AWS Secret Access Key: [your-secret]
# Default region name: us-west-2
# Default output format: json

# Verify configuration
aws sts get-caller-identity --profile aws
```

### Daemon Configuration

Tests expect daemon on `http://localhost:8947`:

```bash
# Start daemon (usually auto-started by tests)
./bin/prismd &

# Check status
./bin/prism admin daemon status

# Stop daemon
./bin/prism admin daemon stop
```

### Test Data Cleanup

**Go Integration Tests**: Automatic via `FixtureRegistry`

**E2E Tests**: Manual cleanup via `cleanupTestResources()` in `fixtures.js`

**Manual Cleanup** (emergency):
```bash
# List all resources
./bin/prism list

# Delete specific resources
./bin/prism delete instance-name
./bin/prism storage delete storage-name
./bin/prism volume delete volume-name
```

## Best Practices

### 1. Always Use Fixtures for AWS Resources

❌ **Bad** - Manual cleanup prone to leaks:
```go
func TestManualCleanup(t *testing.T) {
    client := getClient()
    instance, _ := client.LaunchInstance(ctx, launchReq)

    // Test stuff...

    // Manual cleanup - FAILS if test panics!
    client.DeleteInstance(ctx, instance.Name)
}
```

✅ **Good** - Automatic cleanup:
```go
func TestAutoCleanup(t *testing.T) {
    client := getClient()
    registry := fixtures.NewFixtureRegistry(t, client)

    instance, _ := fixtures.CreateTestInstance(t, registry, opts)

    // Test stuff...

    // Cleanup happens automatically via t.Cleanup()
}
```

### 2. Use Build Tags for Integration Tests

```go
//go:build integration
// +build integration

package integration
```

This allows skipping expensive tests:
```bash
go test ./...              # Skips integration tests
go test -tags integration  # Runs integration tests
```

### 3. Respect test.Short()

```go
func TestExpensiveOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long-running test in short mode")
    }
    // ... expensive test
}
```

Run with: `go test -short ./...`

### 4. Use Descriptive Test Names

✅ Good:
- `TestBackupWorkflow_CreatesBackupAndVerifiesAvailability`
- `TestInstanceLaunch_WithCustomSize_UsesCorrectInstanceType`

❌ Bad:
- `TestBackup`
- `TestIt`

### 5. Table-Driven Tests for Multiple Scenarios

```go
func TestInstanceSizing(t *testing.T) {
    tests := []struct {
        name         string
        size         string
        expectedType string
    }{
        {"small size", "S", "t3.small"},
        {"medium size", "M", "t3.medium"},
        {"large size", "L", "t3.large"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic...
        })
    }
}
```

## Continuous Integration

### GitHub Actions

```yaml
name: Tests
on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make test

  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go test -tags integration ./test/integration/...
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npx playwright test
```

## Related Documentation

- **Fixture Implementation Details**: `test/fixtures/README.md`
- **E2E Test Examples**: `test/integration/fixtures_example_test.go`
- **Development Guide**: `CLAUDE.md`
- **Architecture**: `docs/architecture/`

---

**Last Updated**: 2025-11-23
**Status**: ✅ All testing infrastructure production-ready
