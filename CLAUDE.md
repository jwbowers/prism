# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Cloud Workstation Platform — Development Guide

## Project Overview

Command-line tool that provides academic researchers with pre-configured cloud workstations on AWS. Two client interfaces (CLI, GUI) share a single backend daemon.

```
cmd/
├── prism/        # CLI client binary
├── prism-gui/    # GUI client binary (Wails v3 + React/Cloudscape)
└── prismd/       # Backend daemon binary (REST API on :8947)

pkg/
├── api/          # API client interface
├── daemon/       # Daemon core + HTTP handlers
├── aws/          # AWS operations
├── state/        # State management
├── project/      # Project & budget management
├── idle/         # Hibernation & cost optimization
├── profile/      # AWS credential profiles
└── types/        # Shared types

internal/
└── cli/          # CLI application logic
```

## Core Design Principles

These guide every design decision:

- **Default to Success** — `prism workspace launch python-ml my-project` must always work out of the box
- **Transparent Fallbacks** — always tell the user what changed and why ("ARM not available, using x86")
- **Helpful Warnings** — gentle guidance on suboptimal choices; never silent failures
- **Zero Surprises** — show config preview before launch; dry-run mode available
- **Progressive Disclosure** — simple by default, full control available (`--instance-type`, `--spot`, etc.)

## Architecture Decisions

### Multi-Modal Design
- **CLI**: automation, scripting, power users
- **GUI**: mouse-friendly desktop (Wails v3 + Cloudscape)
- **Unified Backend**: both share the daemon REST API on port 8947

### Frontend API Client Pattern (SafePrismAPI)

**CRITICAL**: The frontend uses named methods, never generic HTTP methods.

**✅ Correct:**
```typescript
// cmd/prism-gui/frontend/src/App.tsx
class SafePrismAPI {
  async getTemplates(): Promise<Template[]> { ... }
  async createProfile(profile: ProfileData): Promise<Profile> { ... }
}
const profiles = await api.getProfiles();
```

**❌ Wrong — do not add:**
```typescript
async get(endpoint: string) { ... }   // No
async post(endpoint: string) { ... }  // No
await api.get('/api/v1/profiles');    // No
```

When adding a feature: add specific methods to SafePrismAPI → each calls private `safeRequest()` → components use those methods only.

### API Authentication

- **Production**: `X-API-Key` header required; stored in `~/.prism/state.json`; exempt: `/ping`, `/auth`, `/authenticate`
- **Test mode**: `PRISM_TEST_MODE=true` bypasses all auth (`middleware.go:73-76`); set automatically by `setup-daemon.js`
- GUI frontend disables `loadAPIKey()` in test mode to avoid keychain prompts

### Template Inheritance

Templates can inherit from parents. Merging rules:
- **Packages, Users, Services**: append (child adds to parent)
- **Package Manager**: override (child replaces parent)
- **Ports**: deduplicate

### State Persistence

Daemon persists state to `~/.prism/`:
- `state.json` — instances, volumes, current profile
- `courses.json` — course records
- `prismd.pid` — singleton lock
- `daemon_registry.json` — process registry

## Development Commands

### AWS Credentials for Testing
- **Profile**: `aws` (`AWS_PROFILE=aws`)
- **Region**: `us-west-2`

### Build
```bash
make build                                    # all binaries
go build -o bin/prism ./cmd/prism/
go build -o bin/prismd ./cmd/prismd/
go build -o bin/prism-gui ./cmd/prism-gui/
make cross-compile                            # all platforms
```

### Run
```bash
./bin/prism workspace launch python-ml my-project   # CLI (daemon auto-starts)
./bin/prism-gui                                      # GUI
curl http://localhost:8947/api/v1/ping               # verify daemon
./bin/prism admin daemon status
```

### Test
```bash
make test
go test ./...
go test -tags integration ./test/integration/... -v  # real AWS

cd cmd/prism-gui/frontend
npm run test:unit           # Vitest unit tests
npx playwright test         # all E2E
npx playwright test backup-workflows.spec.ts --debug
```

## Key Implementation Patterns

### Go API Client
```go
client := api.NewClientWithOptions("http://localhost:8947", api.Options{
    AWSProfile: profile.AWSProfile,
    AWSRegion:  profile.Region,
})
```

### Integration Test Fixtures
```go
func TestBackupWorkflow(t *testing.T) {
    registry := fixtures.NewFixtureRegistry(t, client)
    instance, _ := fixtures.CreateTestInstance(t, registry, opts)
    // t.Cleanup() handles teardown automatically
}
```

### E2E Tests — Critical Patterns

**Onboarding modal**: always suppress before navigation:
```typescript
test.beforeEach(async ({ context }) => {
  await context.addInitScript(() => {
    localStorage.setItem('prism_onboarding_complete', 'true');
  });
});
```

**Wait for API + DOM** before asserting — never assert immediately after navigation.

**Test mode**: `PRISM_TEST_MODE=true` is set by `setup-daemon.js`; no API key needed; AWS operations that would block indefinitely must check this env var and return synthetic data.

**Vite dev server**: App.tsx >500 KB triggers BABEL deoptimisation; playwright.config.js sets `NODE_OPTIONS="--max-old-space-size=4096"` to prevent crashes during multi-context test runs.

**Course code limit**: backend enforces ≤20 characters (`pkg/course/types.go`). Test codes must fit: `PREFIX-${Date.now() % 100000}` (5-digit suffix keeps total ≤14 chars for typical prefixes).

**Concurrent worker isolation**: with 2 browser projects (chromium + webkit), `beforeEach` creates courses with the same prefix in parallel. Always store and use the exact code returned by the API — never filter the table by prefix alone.

## Common Debugging

| Problem | Fix |
|---------|-----|
| `TimeoutError: locator.click` in E2E | Onboarding modal blocking — check `addInitScript` |
| HTTP 400/500 on API call that exists | Fields mismatch between frontend and Go struct — read `pkg/*/types.go` |
| `connect ECONNREFUSED :8947` in E2E | Daemon died — check for AWS call hanging in test mode (must return synthetic data) |
| `resolved to 2 elements` strict-mode | Test using prefix match where exact code required — use stored exact code |
| Vite crashes after ~4 browser contexts | `NODE_OPTIONS="--max-old-space-size=4096"` missing from webServer command |
| Course not appearing after `beforeEach` | Race between creation and initial load — call `refreshBtn.click()` + `waitForResponse` |

**Backend type mismatch checklist**:
1. `grep -r "type.*Request" pkg/` — find the Go struct
2. Compare with frontend `safeRequest()` call body
3. Check `omitempty` tags — required fields have none
4. Verify with `curl -X POST http://localhost:8947/api/v1/endpoint -d '...'`

**Key files**: `pkg/project/types.go`, `pkg/invitation/manager.go`, `pkg/types/*.go`, `cmd/prism-gui/frontend/src/App.tsx` (SafePrismAPI), `cmd/prism-gui/frontend/tests/e2e/pages/*.ts`

## Versioning and Changelog

Prism follows **[Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html)** and **[Keep a Changelog](https://keepachangelog.com/en/1.0.0/)**.

**Every release must include a `CHANGELOG.md` entry** with format:
```markdown
## [X.Y.Z] - YYYY-MM-DD
### Added / Changed / Fixed / Removed / Security / Deprecated
- Brief description (#issue)
```

Version numbers must be in sync in two places:
- `pkg/version/version.go` — `Version = "X.Y.Z"`
- `cmd/prism-gui/frontend/package.json` — `"version": "X.Y.Z"`

## Release Quality Gates ⚠️ MANDATORY

No release commit until ALL pass:

1. **Compilation** — `go build ./...` and `npm run build` — zero errors
2. **Go unit tests** — `go test ./...` — zero failures
3. **Frontend unit tests** — `npm run test:unit` — zero failures
4. **Go lint** — `make lint` — zero issues
5. **TypeScript typecheck** — `npm run typecheck` — zero errors
6. **CHANGELOG.md** — entry for the release version exists

Applies to every version bump and every feature PR. Fix failures before releasing — never defer.

**Vitest unit tests**: all `src/*.test.tsx` and `tests/unit/*.test.js` must pass. Stale mocks must be rewritten, not skipped.
