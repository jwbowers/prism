# Chaos Testing Suite

This directory contains chaos engineering tests that validate Prism's resilience to various failure scenarios.

**Issue**: [#412 - Network Chaos Testing Infrastructure](https://github.com/scttfrdmn/prism/issues/412)

## Overview

Chaos tests simulate real-world failure conditions to ensure the system:
- Maintains state consistency through failures
- Provides clear error messages
- Recovers gracefully after failures
- Handles concurrent operations correctly
- Doesn't lose data or create orphaned resources

## Test Coverage

### Network Failures (`network_failures_test.go`)

**500+ lines of chaos tests**

| Test | Chaos Scenario | Validates |
|------|---------------|-----------|
| `TestNetworkDownDuringLaunch` | Network disconnects during instance launch | Timeout handling, state consistency, retry behavior |
| `TestHighLatencyOperations` | 500ms latency injected | Operations complete under delay, appropriate timeouts |
| `TestPacketLossResilience` | 20% packet loss | Retry logic, eventual success despite packet loss |
| `TestDNSFailureRecovery` | DNS resolution fails | Clear errors, no hanging, recovery after restoration |
| `TestAPIUnavailabilityHandling` | AWS API returns 503 | Fast failure, exponential backoff, clear errors |

**Expected Runtime**: ~5-10 minutes

### System Failures (`system_failures_test.go`)

**400+ lines of chaos tests**

| Test | Chaos Scenario | Validates |
|------|---------------|-----------|
| `TestDaemonCrashDuringOperation` | Daemon killed with SIGKILL | State file integrity, restart recovery, operation continuity |
| `TestOutOfMemoryHandling` | System runs out of memory | Memory monitoring, graceful degradation, cleanup |
| `TestDiskFullScenario` | Disk fills up during writes | Clear errors, no corruption, atomic writes |

**Expected Runtime**: ~5-10 minutes
**Note**: Daemon crash test requires `CHAOS_TESTING=true` environment variable

### Concurrent Operations (`concurrent_operations_test.go`)

**300+ lines of chaos tests**

| Test | Chaos Scenario | Validates |
|------|---------------|-----------|
| `TestConcurrentInstanceLaunches` | 5 simultaneous instance launches | No race conditions, unique IDs, state consistency |
| `TestConcurrentStateModifications` | Multiple operations on same resource | Serialization, idempotency, consistency |
| `TestRaceConditionDetection` | Heavy concurrent load with race detector | No data races in critical sections |

**Expected Runtime**: ~5-10 minutes
**Note**: Run with `-race` flag for full race detection

## Running Chaos Tests

### Prerequisites

```bash
# Build daemon and CLI
make build

# Start daemon (if not already running)
./bin/prismd &

# Ensure AWS credentials configured
aws configure
```

### Run All Chaos Tests

```bash
# Run all chaos tests
go test -v -tags integration ./test/integration/chaos/

# With race detection
go test -v -race -tags integration ./test/integration/chaos/
```

### Run Specific Test Categories

```bash
# Network failures only
go test -v -tags integration ./test/integration/chaos/ -run TestNetwork

# System failures only (requires CHAOS_TESTING=true)
CHAOS_TESTING=true go test -v -tags integration ./test/integration/chaos/ -run TestSystem

# Concurrent operations only
go test -v -tags integration ./test/integration/chaos/ -run TestConcurrent
```

### Run Individual Tests

```bash
# Network down scenario
go test -v -tags integration ./test/integration/chaos/ -run TestNetworkDownDuringLaunch

# High latency scenario
go test -v -tags integration ./test/integration/chaos/ -run TestHighLatencyOperations

# Packet loss scenario
go test -v -tags integration ./test/integration/chaos/ -run TestPacketLossResilience

# DNS failure scenario
go test -v -tags integration ./test/integration/chaos/ -run TestDNSFailureRecovery

# API unavailability scenario
go test -v -tags integration ./test/integration/chaos/ -run TestAPIUnavailabilityHandling

# Daemon crash scenario (requires CHAOS_TESTING=true)
CHAOS_TESTING=true go test -v -tags integration ./test/integration/chaos/ -run TestDaemonCrashDuringOperation

# Out of memory scenario
go test -v -tags integration ./test/integration/chaos/ -run TestOutOfMemoryHandling

# Disk full scenario
go test -v -tags integration ./test/integration/chaos/ -run TestDiskFullScenario

# Concurrent launches
go test -v -tags integration ./test/integration/chaos/ -run TestConcurrentInstanceLaunches

# Concurrent state modifications
go test -v -tags integration ./test/integration/chaos/ -run TestConcurrentStateModifications

# Race condition detection (with race detector)
go test -v -race -tags integration ./test/integration/chaos/ -run TestRaceConditionDetection
```

## Test Output

Chaos tests provide detailed logging:

```
🌪️  CHAOS TEST: Network Down During Instance Launch

📋 Phase 1: Baseline - Normal instance launch (control)
✅ Project created: prj_abc123
✅ Baseline instance launched successfully in 45.2s
   Instance ID: i-0123456789abcdef0

📋 Phase 2: Testing timeout behavior with slow operations
   Simulating network issues through timeout testing
Attempting launch with 2-second timeout (should fail)...
✅ Launch correctly timed out after 2.1s
   Error message: context deadline exceeded

📋 Phase 3: Testing recovery after network issues
   Verifying retry behavior with proper context
Attempting instance launch with proper timeout...
✅ Recovery successful in 42.8s
   Instance ID: i-0fedcba987654321

✅ Network Down During Launch Test Complete!
   ✓ Baseline launch successful (control)
   ✓ Timeout behavior validated
   ✓ State consistency maintained after failures
   ✓ Recovery successful after network issues

🎉 System handles network failures gracefully!
```

## Safety Notes

### Daemon Crash Tests

The `TestDaemonCrashDuringOperation` test kills the daemon process. To prevent interference with other tests:

1. Only runs when `CHAOS_TESTING=true` is set
2. Should not be run in parallel with other tests
3. Automatically restarts daemon after test completes

```bash
# Safe way to run daemon crash tests
CHAOS_TESTING=true go test -v -tags integration ./test/integration/chaos/ -run TestDaemonCrash
```

### AWS Resource Usage

Chaos tests create real AWS resources:
- EC2 instances (t3.small by default)
- EFS volumes
- Projects and budgets

**Cleanup**:
- Tests use fixtures with automatic cleanup
- Resources are terminated when tests complete
- Failed tests may leave orphaned resources (check AWS console)

**Cost Estimation**:
- Full chaos suite: ~$0.50-$1.00 per run
- Network tests only: ~$0.20 per run
- System tests only: ~$0.15 per run
- Concurrent tests only: ~$0.30 per run

### Race Detector

The `-race` flag adds significant overhead:
- Tests run 10x slower
- Memory usage increases 10x
- Some timeouts may need adjustment

```bash
# Use generous timeouts with race detector
go test -v -race -tags integration -timeout 60m ./test/integration/chaos/
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Chaos Tests

on:
  schedule:
    # Run chaos tests nightly
    - cron: '0 2 * * *'
  workflow_dispatch:

jobs:
  chaos:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Build
        run: make build

      - name: Start Daemon
        run: ./bin/prismd &

      - name: Run Network Chaos Tests
        run: |
          go test -v -tags integration \
            ./test/integration/chaos/ \
            -run TestNetwork \
            -timeout 30m

      - name: Run System Chaos Tests
        env:
          CHAOS_TESTING: true
        run: |
          go test -v -tags integration \
            ./test/integration/chaos/ \
            -run TestSystem \
            -timeout 30m

      - name: Run Concurrent Chaos Tests
        run: |
          go test -v -race -tags integration \
            ./test/integration/chaos/ \
            -run TestConcurrent \
            -timeout 60m
```

## Success Criteria

Chaos tests validate:

✅ **Network Resilience**:
- Operations timeout appropriately (not hanging indefinitely)
- Retries use exponential backoff
- Recovery succeeds after network restoration
- Error messages are clear and actionable

✅ **State Consistency**:
- State file never corrupted
- Atomic writes prevent partial state
- State recovers correctly after crashes
- No orphaned AWS resources

✅ **Concurrent Operations**:
- Zero data races detected
- Instance IDs are always unique
- Idempotent operations work correctly
- Final state matches expected state

✅ **Error Handling**:
- Fast failure (no infinite retries)
- Clear error messages
- Appropriate timeout values
- Graceful degradation

## Future Enhancements

Future chaos tests may include:

- **Toxiproxy Integration**: More realistic network chaos (latency spikes, bandwidth limits)
- **AWS Service Outage Simulation**: Mock AWS API failures for specific services
- **Template Provisioning Chaos**: Large file corruption, checksum failures
- **Multi-Region Chaos**: Regional failures, failover testing
- **LocalStack Integration**: Offline chaos testing without AWS costs

## Troubleshooting

### Test Hangs Indefinitely

**Problem**: Test appears to hang without progress

**Solutions**:
```bash
# Use shorter timeout
go test -v -tags integration -timeout 15m ./test/integration/chaos/ -run TestName

# Check daemon logs
tail -f ~/.prism/daemon.log

# Check for zombie processes
ps aux | grep prismd
```

### Daemon Won't Restart

**Problem**: Daemon crash test fails to restart daemon

**Solutions**:
```bash
# Manually restart daemon
./bin/prismd &

# Check for port conflicts
lsof -i :8947

# Kill existing daemon
pkill prismd
```

### AWS Resources Not Cleaned Up

**Problem**: Tests fail and leave resources in AWS

**Solutions**:
```bash
# List Prism-managed resources
./bin/prism list
./bin/prism projects
./bin/prism storage list

# Clean up manually
./bin/prism terminate <instance-id>
./bin/prism storage delete <volume-id>
```

### Race Detector Timeouts

**Problem**: Tests with `-race` flag timeout

**Solutions**:
```bash
# Increase timeout (race detector adds 10x overhead)
go test -v -race -tags integration -timeout 60m ./test/integration/chaos/

# Run subset of tests
go test -v -race -tags integration -run TestConcurrentInstanceLaunches
```

## Contributing

When adding new chaos tests:

1. **Follow Naming Convention**: `Test<Category><Scenario>`
2. **Add Detailed Logging**: Use `t.Logf()` with phases and status indicators
3. **Use Test Fixtures**: Automatic cleanup via `fixtures.NewFixtureRegistry`
4. **Document Expected Behavior**: Clear comments about what chaos is injected
5. **Verify Cleanup**: Ensure resources are cleaned up even on failure
6. **Update README**: Add new test to appropriate category table

## References

- **Issue #412**: [Network Chaos Testing Infrastructure](https://github.com/scttfrdmn/prism/issues/412)
- **Issue #413**: [AWS Service Outage Simulation](https://github.com/scttfrdmn/prism/issues/413)
- **Issue #415**: [Instance Management Edge Cases](https://github.com/scttfrdmn/prism/issues/415)
- **Chaos Engineering Principles**: https://principlesofchaos.org/
- **Testing Improvement Roadmap**: `/docs/releases/TESTING_IMPROVEMENT_ROADMAP.md`
