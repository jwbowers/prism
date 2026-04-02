# Prism Testing Guide

This guide covers the comprehensive testing strategy for Prism, including unit tests, integration tests with Substrate, and code coverage analysis.

## Test Coverage Targets

- **pkg/aws**: 85% (most critical package handling money and cloud resources)
- **pkg/daemon**: 80% (HTTP API server)
- **pkg/api**: 75% (API client library)
- **Overall project**: 75%

## Current Coverage Status

| Package | Current Coverage | Target | Status |
|---------|------------------|---------|---------|
| pkg/aws | 49.5% | 85% | 🟡 In Progress |
| pkg/daemon | 27.8% | 80% | 🟡 In Progress |
| pkg/api | 58.3% | 75% | 🟡 Approaching |
| pkg/state | 76.1% | 75% | ✅ Complete |
| pkg/types | 100% | 75% | ✅ Complete |

## Test Types

### 1. Unit Tests

**Location**: `*_test.go` files alongside source code
**Command**: `go test ./...`

Unit tests cover:
- Helper functions and utilities
- Pricing calculations and discounts
- Template validation
- Error handling
- Business logic without external dependencies

**Key Test Files:**
- `pkg/aws/manager_test.go` - Comprehensive AWS manager tests
- `pkg/daemon/server_test.go` - HTTP handler tests
- `pkg/state/manager_test.go` - State management tests
- `pkg/types/types_test.go` - Type validation tests

### 2. Substrate Integration Tests

**Location**: `pkg/aws/substrate_integration_test.go`
**Build tag**: `substrate`
**Commands**:
```bash
make test-substrate          # In-process (no Docker)
make test-substrate-docker   # Via Docker container
```

Substrate tests provide in-process AWS emulation:
- Real AWS API testing without actual cloud costs
- Sub-100ms startup (no Docker needed for unit tests)
- Complete EBS/EFS volume lifecycle testing
- Error handling with real AWS error types

**Prerequisites for Docker mode:**
- Docker installed
- `make substrate-start` (or Docker Compose)

### 3. Test Coverage Analysis

**Command**: `go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out`

Generates detailed HTML coverage reports showing:
- Line-by-line coverage
- Function coverage
- Package-level summaries
- Uncovered code paths

## Running Tests

### Basic Unit Tests
```bash
# Run all unit tests
go test ./...

# Run with coverage
go test ./... -coverprofile=coverage.out -covermode=atomic

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Substrate Integration Tests

```bash
# In-process (fastest — no Docker needed)
make test-substrate

# Via Docker container
make substrate-start
make test-substrate-docker
make substrate-stop

# Manually
go test -tags=substrate ./pkg/aws/... -v
```

### Individual Package Testing

```bash
# Test specific packages
go test ./pkg/aws -coverprofile=aws_coverage.out
go test ./pkg/daemon -coverprofile=daemon_coverage.out
go test ./pkg/api -coverprofile=api_coverage.out
```

## Substrate Setup

[Substrate](https://github.com/scttfrdmn/substrate) provides in-process AWS emulation:

**Services Emulated:**
- EC2 (instance management)
- EFS (file system volumes)
- EBS (block storage)
- STS (security token service)
- IAM (basic permissions — pending protocol fix substrate#260)
- SSM (run command — pending DateTime fix substrate#261)

**In-process usage** (`pkg/aws/substrate_integration_test.go`):
```go
//go:build substrate

ts := substrate.StartTestServer(t) // stops automatically at test end
cfg := aws.Config{
    Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
    EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
        func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
            return aws.Endpoint{URL: ts.URL, HostnameImmutable: true}, nil
        }),
}
```

**Docker Compose** (`test/substrate/docker-compose.yml`):
```yaml
services:
  substrate:
    image: ghcr.io/scttfrdmn/substrate:latest
    ports:
      - "4566:4566"
```

## Test Categories

### 1. AWS Manager Tests (`pkg/aws/manager_test.go`)

**Pricing Tests:**
- Regional pricing multipliers
- Instance type cost calculations
- Volume pricing (EBS, EFS)
- Discount combinations
- Cost caching logic

**Template Tests:**
- Template validation
- Architecture mapping
- AMI selection by region
- Instance type selection

**Helper Function Tests:**
- Size parsing (XS, S, M, L, XL → GB)
- Performance parameter calculation
- User data manipulation
- Error handling

### 2. Daemon Server Tests (`pkg/daemon/server_test.go`)

**HTTP Handler Tests:**
- Method validation (GET, POST, etc.)
- Request routing
- JSON request/response handling
- Error response formatting
- Path parsing

**API Endpoint Tests:**
- `/api/v1/ping` - Health check
- `/api/v1/status` - Daemon status
- `/api/v1/instances` - Instance operations
- `/api/v1/volumes` - Volume operations
- `/api/v1/storage` - Storage operations

### 3. Substrate Integration Tests (`pkg/aws/substrate_integration_test.go`)

**Currently Passing:**
- `TestSubstrateCreateEBSVolume` - EBS volume create/list/delete
- `TestSubstrateErrorHandling` - Invalid resource error propagation

**Skipped (pending Substrate fixes):**
- `TestSubstrateLaunchInstance` — needs VPC/AMI pre-seeding (substrate SeedDefaults helper)
- `TestSubstrateEBSAttachDetach` — depends on instance launch
- `TestSubstrateIAMInstanceProfile` — substrate#260 (IAM Query/XML protocol)
- `TestSubstrateSSMRunCommand` — substrate#261 (SSM DateTime format)

## Coverage Improvement Strategies

### For AWS Package (Target: 85%)

**Currently Tested (49.5%):**
✅ Pricing calculations and regional multipliers
✅ Template validation and architecture mapping
✅ Helper functions (parsing, validation)
✅ Discount application logic
✅ Billing information handling

**Needs Integration Testing:**
🔄 Instance launch/management operations
🔄 Volume creation/management operations
🔄 AWS API error handling
🔄 Network and security group creation

**Strategy**: Use Substrate integration tests to cover the actual AWS operations that require API calls.

### For Daemon Package (Target: 80%)

**Currently Tested (27.8%):**
✅ HTTP method validation
✅ Request routing and path parsing
✅ JSON error responses
✅ Basic handler functionality

**Needs More Coverage:**
🔄 Complete request/response cycles
🔄 State management integration
🔄 AWS manager integration
🔄 Middleware functionality

**Strategy**: Add comprehensive handler tests with mock dependencies.

## Continuous Integration

**Recommended CI Pipeline:**
1. **Lint**: `golangci-lint run`
2. **Unit Tests**: `go test ./... -coverprofile=coverage.out`
3. **Substrate Tests**: `go test -tags=substrate ./pkg/aws/...`
4. **Coverage Analysis**: Fail if below targets
5. **Build**: Ensure all binaries build successfully

**Environment Variables:**
- `PRISM_TEST_MODE=true` - Bypass API authentication in daemon

## Debugging Tests

### Verbose Output
```bash
go test ./pkg/aws -v  # Verbose test output
go test ./pkg/aws -v -run TestSpecificFunction  # Run specific test
```

### Substrate Debugging
```bash
# View Substrate container logs
make substrate-logs

# Check Substrate health
curl http://localhost:4566/health

# Reset Substrate state between tests
curl -X POST http://localhost:4566/_substrate/reset
```

### Coverage Debugging
```bash
# Show uncovered functions
go tool cover -func=coverage.out | grep -v "100.0%"

# Generate coverage profile for specific package
go test ./pkg/aws -coverprofile=aws.out -covermode=count
go tool cover -func=aws.out
```

## Best Practices

1. **Test Structure**: Use table-driven tests for multiple scenarios
2. **Isolation**: Each test should be independent and clean up after itself
3. **Substrate for AWS**: Use Substrate for AWS API tests — in-process, no Docker required
4. **Coverage**: Focus on critical paths and error conditions
5. **Performance**: Keep unit tests fast (<1s each), integration tests can be slower
6. **Documentation**: Test names should clearly describe what they test

## Future Improvements

1. **Fuzzing**: Add fuzz tests for input validation
2. **Benchmarks**: Add performance benchmarks for critical paths
3. **Property Testing**: Add property-based tests for complex algorithms
4. **Load Testing**: Add load tests for daemon server
5. **End-to-End**: Add full workflow tests with real AWS (optional)
6. **Substrate coverage**: Re-enable skipped tests when substrate#260 and #261 are fixed

## Troubleshooting

**Common Issues:**
- Substrate container not starting: Check Docker daemon and port 4566
- `go test -tags=substrate` failing: Run `go mod tidy` to ensure substrate dep is present
- Coverage reports not generating: Check file permissions and output directory

**Debug Commands:**
```bash
# Check Substrate container status
docker ps | grep substrate

# Test Substrate connectivity
curl -v http://localhost:4566/health

# Validate test build tags
go list -tags=substrate ./pkg/aws
```
