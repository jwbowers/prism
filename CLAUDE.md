# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Cloud Workstation Platform - Claude Development Context

## Project Overview

This is a command-line tool that provides academic researchers with pre-configured cloud workstations, eliminating the need for manual environment configuration.

## Core Design Principles

These principles guide every design decision and feature implementation:

### 🎯 **Default to Success**
Every template must work out of the box in every supported region. No configuration should be required for basic usage.
- `prism workspace launch python-ml my-project` should always work
- Smart fallbacks handle regional/architecture limitations transparently
- Templates include battle-tested defaults for their specific use cases

### ⚡ **Optimize by Default**
Templates automatically choose the best instance size and type for their intended workload.
- ML templates default to GPU instances
- R templates default to memory-optimized configurations
- Cost-performance ratio optimized for academic budgets
- ARM instances preferred when available (better price/performance)

### 🔍 **Transparent Fallbacks**
When the ideal configuration isn't available, users always know what changed and why.
- Clear communication: "ARM GPU not available in us-west-1, using x86 GPU instead"
- Fallback chains documented and predictable
- No silent degradation of performance or capabilities

### 💡 **Helpful Warnings**
Gentle guidance when users make suboptimal choices, with clear alternatives offered.
- Warning when choosing CPU instance for ML workload
- Memory warnings for data-intensive R work
- Cost alerts for expensive configurations
- Educational not prescriptive approach

### 🚫 **Zero Surprises**
Users should never be surprised by what they get - clear communication about what's happening.
- Detailed configuration preview before launch
- Real-time progress reporting during operations
- Clear cost estimates and architecture information
- Dry-run mode for validation without commitment

### 📈 **Progressive Disclosure**
Simple by default, detailed when needed. Power users can access advanced features without cluttering basic workflows.
- Basic: `prism workspace launch template-name project-name`
- Intermediate: `prism workspace launch template-name project-name --size L`
- Advanced: `prism workspace launch template-name project-name --instance-type c5.2xlarge --spot`
- Expert: Full template customization and regional optimization

## Current Status

**Production-Ready Enterprise Platform** (Phase 5 - November 2025)
- Multi-modal access: CLI, TUI, GUI with full feature parity
- Distributed architecture: Backend daemon (prismd) + client interfaces
- AWS-native: Cloudscape Design System for GUI

**For detailed features, roadmap, and architecture**:
- [Project Roadmap](docs/TECHNICAL_DEBT_BACKLOG.md)
- [Architecture Docs](docs/architecture/)
- [User Guides](docs/)

**Multi-Modal Access Strategy**:
```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ CLI Client  │  │ TUI Client  │  │ GUI Client  │
│ (cmd/prism) │  │ (prism tui) │  │(cmd/prism-gui)│
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
                 ┌─────────────┐
                 │ Backend     │
                 │ Daemon      │
                 │(prismd:8947)│
                 └─────────────┘
```

**Current Architecture**:
```
cmd/
├── prism/        # CLI client binary
├── prism-gui/    # GUI client binary (Wails v3-based)
└── prismd/       # Backend daemon binary

pkg/
├── api/          # API client interface
├── daemon/       # Daemon core logic
├── aws/          # AWS operations
├── state/        # State management
├── project/      # Project & budget management (Phase 4)
├── idle/         # Hibernation & cost optimization (Phase 3)
├── profile/      # Enhanced profile system
└── types/        # Shared types & project models

internal/
├── cli/          # CLI application logic
├── tui/          # TUI application (BubbleTea-based)
└── gui/          # (GUI logic is in cmd/prism-gui/)
```

**Phase 4 Enterprise Components**:
```
pkg/project/
├── manager.go         # Project lifecycle & member management
├── budget_tracker.go  # Real-time cost tracking & alerts
├── cost_calculator.go # AWS pricing engine & hibernation savings
└── types.go          # Request/response types & filters

pkg/daemon/
└── project_handlers.go # REST API endpoints (/api/v1/projects)

pkg/types/
└── project.go         # Enterprise data models & budget types
```

**Feature Parity Matrix**:
| Feature | CLI | TUI | GUI | Status |
|---------|-----|-----|-----|---------|
| Templates | ✅ | ✅ | ✅ | Complete |
| Instance Management | ✅ | ✅ | ✅ | Complete |
| Storage (EFS/EBS) | ✅ | ✅ | ✅ | Complete |
| Advanced Launch | ✅ | ✅¹ | ✅ | Complete |
| Profile Management | ✅ | ✅ | ✅ | Complete |
| Daemon Control | ✅ | ✅ | ✅ | Complete |

¹ *TUI provides CLI command guidance for launch operations*

## Architecture Decisions

### Multi-Modal Design Philosophy
- **CLI**: Power users, automation, scripting - maximum efficiency
- **TUI**: Interactive terminal users, remote access - keyboard-first navigation
- **GUI**: Desktop users, visual management - mouse-friendly interface
- **Unified Backend**: All interfaces share same daemon API and state

### API Architecture
- **REST API**: HTTP endpoints on port 8947 (CWS on phone keypad)
- **Options Pattern**: Modern `api.NewClientWithOptions()` with configuration
- **Profile Integration**: Integrated AWS credential and region management
- **Graceful Operations**: Proper shutdown, error handling, progress reporting

### Frontend API Client Pattern (SafePrismAPI)

**CRITICAL**: The frontend follows a **specific method pattern** for API calls - NOT generic HTTP methods.

**✅ CORRECT Pattern - Specific Methods:**
```typescript
// SafePrismAPI class (cmd/prism-gui/frontend/src/App.tsx)
class SafePrismAPI {
  // Feature-specific methods that call safeRequest internally
  async getTemplates(): Promise<Template[]> { ... }
  async getInstances(): Promise<Instance[]> { ... }
  async getProfiles(): Promise<Profile[]> { ... }
  async createProfile(profile: ProfileData): Promise<Profile> { ... }
  async updateProfile(id: string, updates: Partial<ProfileData>): Promise<Profile> { ... }
  async deleteProfile(id: string): Promise<void> { ... }
  async switchProfile(id: string): Promise<void> { ... }
}

// Components use specific methods
const profiles = await api.getProfiles();
await api.createProfile({ name, aws_profile, region });
```

**❌ INCORRECT Pattern - Generic HTTP Methods:**
```typescript
// DO NOT ADD these to SafePrismAPI:
async get(endpoint: string) { ... }    // ❌ No
async post(endpoint: string) { ... }   // ❌ No
async put(endpoint: string) { ... }    // ❌ No
async delete(endpoint: string) { ... } // ❌ No

// DO NOT use generic calls in components:
await api.get('/api/v1/profiles');     // ❌ No
await api.post('/api/v1/profiles', data); // ❌ No
```

**Why This Pattern?**
1. **Architectural Consistency**: All features (Templates, Instances, Storage, Profiles) use specific methods
2. **Type Safety**: Specific methods provide better TypeScript typing
3. **Maintainability**: Changes to API structure are isolated to method implementations
4. **Discoverability**: IDE autocomplete shows available operations

**When Adding New Features:**
1. Add specific methods to SafePrismAPI for the feature (e.g., `getProfiles()`, `createProfile()`)
2. Each method calls the private `safeRequest()` helper internally
3. Components use these specific methods, never direct HTTP calls
4. Follow the same error handling pattern as existing methods

### API Authentication
The daemon uses header-based authentication with test mode bypass:

**Production Mode:**
- API key required via `X-API-Key` header
- Key stored in daemon state configuration (`~/.prism/state.json`)
- Constant-time comparison prevents timing attacks (pkg/daemon/middleware.go:94)
- Applies to all endpoints except `/api/v1/ping`, `/api/v1/auth`, `/api/v1/authenticate`

**Test Mode (E2E Tests):**
- Set `PRISM_TEST_MODE=true` environment variable
- Authentication middleware bypasses all key checks (middleware.go:73-76)
- GUI frontend does NOT send API key in test mode (App.tsx loadAPIKey disabled)
- E2E test setup script (`tests/e2e/setup-daemon.js`) automatically sets test mode

**Implementation Reference:**
```go
// pkg/daemon/middleware.go
if os.Getenv("PRISM_TEST_MODE") == "true" {
    next(w, r)  // Skip authentication
    return
}
```

**Testing Pattern:**
- E2E tests use setup-daemon.js to launch daemon with PRISM_TEST_MODE
- No API key configuration needed for tests
- Frontend disables loadAPIKey() to avoid keychain prompts

### Streamlined User Experience
- **Auto-Start Daemon**: All interfaces automatically start daemon as needed - no manual setup required
- **Zero Keychain Prompts**: Basic profiles work without macOS keychain password requests
- **Intelligent Binary Discovery**: Auto-locates daemon binary in development and production environments
- **Profile System Unified**: Single enhanced profile manager eliminates configuration conflicts

### Templates (Inheritance Architecture)

**✅ IMPLEMENTED: Template Inheritance System**

Prism now supports template stacking and inheritance, allowing templates to build upon each other:

```bash
# Base template provides foundation
# templates/base-rocky9.yml: Rocky Linux 9 + DNF + system tools + rocky user

# Stacked template inherits and extends  
# templates/rocky9-conda-stack.yml:
#   inherits: ["Rocky Linux 9 Base"]
#   package_manager: "conda"  # Override parent's DNF
#   adds: conda packages, datascientist user, jupyter service

# Launch stacked template
prism workspace launch "Rocky Linux 9 + Conda Stack" my-analysis
# ↳ Gets: rocky user + datascientist user, system packages + conda packages, ports 22 + 8888
```

**Inheritance Merging Rules**:
- **Packages**: Append (base system packages + child conda packages)
- **Users**: Append (base rocky user + child datascientist user)  
- **Services**: Append (base services + child jupyter service)
- **Package Manager**: Override (child conda overrides parent DNF)
- **Ports**: Deduplicate (base 22 + child 8888 = [22, 8888])

**Available Templates**:
- `Rocky Linux 9 Base`: Foundation with DNF, system tools, rocky user
- `Rocky Linux 9 + Conda Stack`: Inherits base + adds conda ML packages
- `Python Machine Learning (Simplified)`: Conda + Jupyter + ML packages  
- `R Research Environment (Simplified)`: Conda + RStudio + tidyverse
- `Basic Ubuntu (APT)`: Ubuntu + APT package management
- `Web Development (APT)`: Ubuntu + web development tools

**Future Multi-Stack Architecture**:
```bash  
# Planned: Complex inheritance chains
prism workspace launch gpu-ml-workstation my-training
# ↳ Inherits: Base OS → GPU Drivers → Conda ML → Desktop GUI

# Power users can override at launch
prism workspace launch "Rocky Linux 9 + Conda Stack" my-project --with spack
```

**Design Benefits**:
- **Composition Over Duplication**: Inherit and extend vs copy/paste
- **Maintainable Library**: Base template updates propagate to children
- **Clear Relationships**: Explicit parent-child dependencies
- **Flexible Override**: Change any aspect while preserving inheritance

### Desktop Applications (Nice DCV)

**Future Feature (v0.6.1+)**: Desktop GUI applications (MATLAB, QGIS, Mathematica) via AWS Nice DCV

**Reference Implementation**: [Lens project](https://github.com/scttfrdmn/lens) has complete DCV implementation

**Documentation**: [docs/architecture/NICE_DCV_ARCHITECTURE.md](docs/architecture/NICE_DCV_ARCHITECTURE.md)

**Template Pattern**:
```yaml
# Desktop template (planned)
name: "MATLAB Workstation"
connection_type: "desktop"
desktop:
  environment: "mate"
  dcv_port: 8443
ports:
  - 8443
```

### State Management
Enhanced state management with profile integration:
```json
{
  "instances": {
    "my-instance": {
      "id": "i-1234567890abcdef0",
      "name": "my-instance", 
      "template": "r-research",
      "public_ip": "54.123.45.67",
      "state": "running",
      "launch_time": "2024-06-15T10:30:00Z",
      "estimated_daily_cost": 2.40,
      "attached_volumes": ["shared-data"],
      "attached_ebs_volumes": ["project-storage-L"]
    }
  },
  "volumes": {
    "shared-data": {
      "filesystem_id": "fs-1234567890abcdef0",
      "state": "available",
      "creation_time": "2024-06-15T10:00:00Z"
    }
  },
  "current_profile": {
    "name": "research-profile",
    "aws_profile": "my-aws-profile", 
    "region": "us-west-2"
  }
}
```

## Development Principles

1. **Multi-modal first**: Every feature must work across CLI, TUI, and GUI
2. **API-driven**: All interfaces use the same backend API
3. **Profile-aware**: Integrated AWS credential and region management
4. **Real-time sync**: Changes reflect across all interfaces automatically
5. **Professional quality**: Zero compilation errors, comprehensive testing

## Future Phases (Post-Phase 2)

- **Phase 3**: Advanced research features (multi-package managers, hibernation, snapshots) ✅ COMPLETE
- **Phase 4**: Collaboration & scale (multi-user, template marketplace, enterprise features) ✅ COMPLETE
- **Phase 5**: AWS-native research ecosystem expansion (advanced storage, networking, research services)

## Development Commands

### AWS Credentials for Testing
- **AWS Profile**: `aws` (`AWS_PROFILE=aws`)
- **AWS Region**: `us-west-2`

Use these for all testing against real AWS — CLI integration tests, E2E tests, and template validation launches.

### Building and Testing
```bash
# Build all components
make build
# Builds: prism (CLI), prismd (daemon), prism-gui (GUI)

# Build specific components
go build -o bin/prism ./cmd/prism/        # CLI
go build -o bin/prismd ./cmd/prismd/      # Daemon
go build -o bin/prism-gui ./cmd/prism-gui/ # GUI

# Run tests
make test

# Cross-compile for all platforms
make cross-compile

# Clean build artifacts
make clean
```

### Running Different Interfaces
```bash
# CLI interface (traditional) - daemon auto-starts as needed
./bin/prism workspace launch python-ml my-project

# TUI interface (interactive terminal) - daemon auto-starts as needed
./bin/prism tui
# Navigation: 1=Dashboard, 2=Instances, 3=Templates, 4=Storage, 5=Settings, 6=Profiles

# GUI interface (desktop application) - daemon auto-starts as needed
./bin/prism-gui
# System tray integration with professional tabbed interface

# Manual daemon control (optional)
./bin/prism admin daemon start    # Manually start daemon
./bin/prism admin daemon stop     # Stop daemon
./bin/prism admin daemon status   # Check daemon status
```

### Development Workflow
```bash
# Test CLI functionality (daemon auto-starts)
./bin/prism templates
./bin/prism workspace list

# Test TUI functionality (daemon auto-starts if needed)
./bin/prism tui

# Test GUI functionality (daemon auto-starts if needed)
./bin/prism-gui

# Optional: Manual daemon control for development
./bin/prismd &                        # Start daemon manually (for debugging)
./bin/prism admin daemon stop         # Graceful shutdown
./bin/prism admin daemon status       # Check status
```

## Key Implementation Details

### API Client Pattern (All Interfaces)
```go
// Modern API client initialization
client := api.NewClientWithOptions("http://localhost:8947", client.Options{
    AWSProfile: profile.AWSProfile,
    AWSRegion:  profile.Region,
})
```

### Profile System Integration
```go
// Enhanced profile management
currentProfile, err := profile.GetCurrentProfile()
if err != nil {
    // Handle gracefully with defaults
}

// Apply to API client
apiClient := api.NewClientWithOptions(daemonURL, client.Options{
    AWSProfile: currentProfile.AWSProfile,
    AWSRegion:  currentProfile.Region,
})
```

### Cross-Interface State Synchronization
- All interfaces use same daemon backend (port 8947)
- Real-time updates via polling and WebSocket (future)
- Shared profile and configuration system
- Consistent error handling and user feedback

### GUI Specific (cmd/prism-gui/main.go)
- **Wails v3 Framework**: Cross-platform web-based native GUI with React frontend
- **Cloudscape Design System**: AWS-native professional UI components
- **Tabbed Interface**: Templates, Instances, Storage, Settings
- **Professional Dialogs**: Connection info, confirmations, progress
- **Real-time Updates**: Automatic refresh with visual indicators

### TUI Specific (internal/tui/)
- **BubbleTea Framework**: Professional terminal interface
- **Page Navigation**: Keyboard-driven (1-6 keys for pages)
- **Real-time Updates**: 30-second refresh intervals
- **Professional Styling**: Consistent theming, loading states
- **Action Dialogs**: Instance management with confirmations

## Testing Strategy

### Unit and Integration Tests
```bash
# Run all tests
make test
go test ./...

# Run specific package tests
go test ./pkg/templates/...
go test ./pkg/research/...
```

### Integration Testing with Fixtures

**Prism uses a fixture pattern for integration tests** that create real AWS resources with automatic cleanup.

Key benefits:
- ✅ Automatic cleanup via `t.Cleanup()` - no orphaned resources
- ✅ Proper dependency ordering (backups → instances → EBS → EFS)
- ✅ Production-ready pattern used across all integration tests

**Quick Example**:
```go
func TestBackupWorkflow(t *testing.T) {
    registry := fixtures.NewFixtureRegistry(t, client)

    instance, _ := fixtures.CreateTestInstance(t, registry, opts)
    backup, _ := fixtures.CreateTestBackup(t, registry, opts)

    // Test your logic...
    // Cleanup happens automatically!
}
```

**Running**:
```bash
go test -tags integration ./test/integration/... -v
```

**📚 Complete documentation**: See **[docs/TESTING.md](docs/TESTING.md)** for:
- Detailed fixture usage patterns
- E2E testing guide (Playwright)
- CLI integration tests
- Configuration and troubleshooting

### E2E Testing (Playwright)
```bash
cd cmd/prism-gui/frontend

# Run all E2E tests
npx playwright test

# Run specific test suite
npx playwright test backup-workflows.spec.ts
npx playwright test instance-workflows.spec.ts

# Debug mode
npx playwright test --debug
```

**Critical E2E Pattern**: All tests MUST set onboarding modal flag before navigation:
```typescript
test.beforeEach(async ({ page, context }) => {
  // Set localStorage BEFORE navigating
  await context.addInitScript(() => {
    localStorage.setItem('cws_onboarding_complete', 'true');
  });
  // ... rest of setup
});
```

**Test Configuration**:
- AWS Profile: `aws`
- AWS Region: `us-west-2`
- Daemon auto-starts via `setup-daemon.js`
- Tests run against real AWS integration

**📚 For detailed E2E patterns, configuration, and troubleshooting**: See [docs/TESTING.md](docs/TESTING.md)

### Build Testing
```bash
# Ensure zero compilation errors
make build

# Cross-platform compilation
make cross-compile
```

## Common Development Issues & Debugging

### 1. Profile and AWS Credential Issues
**Problem**: API client can't find AWS credentials or uses wrong region

**Solution**: Ensure profile integration in API client:
```go
// Get current profile
currentProfile, err := profile.GetCurrentProfile()

// Create API client with profile
client := api.NewClientWithOptions(daemonURL, api.Options{
    AWSProfile: currentProfile.AWSProfile,
    AWSRegion:  currentProfile.Region,
})
```

**Testing**: Use `aws` profile and `us-west-2` region for tests

### 2. Daemon Connection Issues
**Problem**: CLI/TUI/GUI can't connect to daemon

**Debug Steps**:
```bash
# Check daemon status
./bin/prism admin daemon status

# Manually start daemon for debugging
./bin/prismd &

# Check daemon logs
# (logs location varies by OS)

# Verify daemon is listening
curl http://localhost:8947/api/v1/ping
```

### 3. Template Validation Errors
**Problem**: Template fails to load or validate

**Debug**:
```bash
# Validate all templates
./bin/prism templates validate

# Check specific template
./bin/prism templates info "template-name"
```

**Common Issues**:
- Invalid package manager in template inheritance
- Missing required fields (name, base_os, etc.)
- Invalid port numbers or user specifications

### 4. E2E Test Failures
**Problem**: GUI tests fail with "TimeoutError: locator.click"

**Cause**: Onboarding modal blocking interactions

**Fix**: Always use `context.addInitScript()` BEFORE page navigation (see Testing Strategy section)

### 5. Cross-Platform Build Issues
**Problem**: Build fails on specific platform

**Solution**:
```bash
# Clean and rebuild
make clean
make build

# Check for platform-specific code
grep -r "// +build" .

# Test cross-compilation
make cross-compile
```

### 6. API Compatibility Issues
**Problem**: Daemon API changes break client interfaces

**Best Practice**:
- Maintain backward compatibility in REST API
- Version API endpoints when making breaking changes
- Update all three interfaces (CLI/TUI/GUI) together
- Test cross-interface compatibility

### 7. API Request/Response Signature Mismatches ⚠️ CRITICAL
**Problem**: Tests or frontend fail with HTTP 400/500 errors even though backend code exists

**Common Symptoms**:
- "HTTP 400: Bad Request" on API calls that should work
- "HTTP 500: Internal Server Error" from validation failures
- Test data setup fails silently
- "No data" shown in UI despite setup code running

**Root Cause**: Frontend/tests sending fields that don't exist in backend types, or missing required fields

**Debugging Process**:
```bash
# 1. Check the actual error in test output
npx playwright test tests/e2e/your-test.spec.ts --reporter=list 2>&1 | grep "Error:"

# 2. Find the backend type definition
grep -r "type.*Request" pkg/

# 3. Compare frontend API call to backend type
# Frontend: cmd/prism-gui/frontend/src/App.tsx or tests/e2e/pages/*.ts
# Backend: pkg/*/types.go or pkg/*/manager.go
```

**Example Case** (Issue #322, Commit d0852b674):

Frontend was sending:
```typescript
// ❌ WRONG - fields don't exist in backend
await api.createProject({
  name: 'Test',
  budget_limit: 1000,      // Not in CreateProjectRequest
  budget_period: 'monthly', // Not in CreateProjectRequest
  status: 'active'          // Not in CreateProjectRequest
});
```

Backend expects:
```go
// pkg/project/types.go:11
type CreateProjectRequest struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Owner       string   `json:"owner"`
    // No budget_limit, budget_period, or status fields!
}
```

**Solution Pattern**:
1. ✅ Read backend type definition (`pkg/*/types.go`)
2. ✅ Match frontend API call exactly to backend struct
3. ✅ Include ALL required fields (non-omitempty without defaults)
4. ✅ Exclude fields that don't exist in backend
5. ✅ Test with curl to verify API works: `curl -X POST http://localhost:8947/api/v1/endpoint -d '{"field":"value"}'`

**Prevention**:
- Always check backend type definitions before writing frontend API calls
- Use TypeScript interfaces that match backend Go structs
- Add API integration tests that validate request/response schemas
- Don't assume field names - verify against source code

**Key Files to Check**:
- Backend types: `pkg/project/types.go`, `pkg/invitation/manager.go`, `pkg/types/*.go`
- Frontend API: `cmd/prism-gui/frontend/src/App.tsx` (SafePrismAPI class)
- Test helpers: `cmd/prism-gui/frontend/tests/e2e/pages/*.ts`

## Development Best Practices

1. **Multi-modal first**: Every feature must work across CLI, TUI, and GUI
2. **API-driven**: All interfaces use the same backend API
3. **Profile-aware**: Integrated AWS credential and region management
4. **Real-time sync**: Changes reflect across all interfaces automatically
5. **Professional quality**: Zero compilation errors, comprehensive testing
6. **Test before commit**: Run `make test` and verify E2E tests pass