# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 📋 Quick Navigation

**Project Management**:
- 🎯 [GitHub Issues](https://github.com/scttfrdmn/prism/issues) - Current work, bugs, features
- 📊 [GitHub Milestones](https://github.com/scttfrdmn/prism/milestones) - Phase tracking and progress

**Essential Reading**:
- 👥 [USER_SCENARIOS/](USER_SCENARIOS/) - **5 persona walkthroughs (our north star)**
- 🎨 [UX Evaluation](architecture/UX_EVALUATION_AND_RECOMMENDATIONS.md) - Current UX issues and fixes
- 📐 [DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md) - Core design philosophy
- 🗺️ [ROADMAP.md](ROADMAP.md) - Current status and priorities

**Implementation Guides**:
- 🏗️ [Architecture](architecture/) - Technical architecture and system design
- 💻 [Development](development/) - Setup, testing, code quality, release process
- 📚 [User Guides](user-guides/) - End-user documentation
- 👨‍💼 [Admin Guides](admin-guides/) - Administrator and institutional docs

---

## Project Overview

Prism is a command-line tool that provides academic researchers with pre-configured cloud workstations, eliminating the need for manual environment configuration.

**Current Version**: v0.7.5 (Released January 27, 2026)
**Current Focus**: [v0.8.0 Planning](ROADMAP.md) — LocalStack, auto-update, idle policy execution

---

## 🎯 Persona-Driven Development (CRITICAL)

Prism's feature development is guided by [5 persona walkthroughs](USER_SCENARIOS/) that represent real-world research scenarios. These scenarios are our **north star** for prioritization and decision-making.

### Before Implementing ANY Feature:

1. **Ask**: "Does this clearly improve one of the 5 persona workflows?"
2. **If yes**: Validate the feature makes the workflow simpler/faster/clearer
3. **If no**: Question whether it's the right priority

### The 5 Personas:

1. **[Solo Researcher](USER_SCENARIOS/01_SOLO_RESEARCHER_WALKTHROUGH.md)** - Individual research projects
2. **[Lab Environment](USER_SCENARIOS/02_LAB_ENVIRONMENT_WALKTHROUGH.md)** - Team collaboration
3. **[University Class](USER_SCENARIOS/03_UNIVERSITY_CLASS_WALKTHROUGH.md)** - Teaching & coursework
4. **[Conference Workshop](USER_SCENARIOS/04_CONFERENCE_WORKSHOP_WALKTHROUGH.md)** - Workshops & tutorials
5. **[Cross-Institutional Collaboration](USER_SCENARIOS/05_CROSS_INSTITUTIONAL_COLLABORATION_WALKTHROUGH.md)** - Multi-institution projects

These walkthroughs prioritize **usability and clarity of use** over technical sophistication.

---

## Core Design Principles

See [DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md) for full details. Key principles:

- 🎯 **Default to Success**: Templates work out of the box in every supported region
- ⚡ **Optimize by Default**: Templates automatically choose best instance size/type
- 🔍 **Transparent Fallbacks**: Users always know what changed and why
- 💡 **Helpful Warnings**: Gentle guidance when users make suboptimal choices
- 🚫 **Zero Surprises**: Clear communication about what's happening
- 📈 **Progressive Disclosure**: Simple by default, detailed when needed

---

## 🚀 Current Development Status

**Current Milestone**: [v0.7.0: Production Hardening & Enterprise Features](https://github.com/scttfrdmn/prism/milestone/40)

### Completed Phases
- ✅ Phase 1: Distributed Architecture
- ✅ Phase 2: Multi-Modal Access (CLI/TUI/GUI)
- ✅ Phase 3: Cost Optimization & Hibernation
- ✅ Phase 4: Enterprise Features (projects, budgets, collaboration)
- ✅ Phase 5A: Multi-User Foundation
- ✅ Phase 5B: Template Marketplace
- ✅ v0.5.7: Template Provisioning & Test Infrastructure
- ✅ v0.5.8: Quick Start Experience
- ✅ v0.5.9: Navigation Restructure
- ✅ v0.6.0: Test Infrastructure & API Documentation
- ✅ v0.6.1: Pragmatic Testing Strategy
- ✅ v0.6.2: Enterprise Feature Testing & E2E Infrastructure
- ✅ v0.6.3: Homebrew Template Discovery Fix (RELEASED January 14, 2026)

### Current State: v0.7.5 — Stable & Fully Tested (February 2026)

- ✅ **428/428 E2E tests passing**, 0 failed, 0 skipped
- ✅ All enterprise features: projects, budgets, users, invitations, hibernation, storage
- ✅ Zombie resource prevention: pre-run cleanup + post-run EC2 check
- ✅ Dependencies current: Cloudscape 3.0.1208, Wails v3 alpha.72, React 19

### Next: v0.8.0 Planning

Candidates (see [ROADMAP.md](ROADMAP.md)):
- **#417** LocalStack integration — <5 min tests, no real-AWS dependency
- **#419** Auto-update (version detection + notifications)
- **#288** Idle policy backend execution — make hibernation schedules actually run
- **#285** Backup/restore commands
- **#418** AWS quota management + AZ failover

---

## 🏗️ Architecture Overview

### Multi-Modal Access Strategy

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ CLI Client  │  │ TUI Client  │  │ GUI Client  │
│ (cmd/prism) │  │ (prism tui) │  │ (prism-gui) │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
                 ┌─────────────┐
                 │ Backend     │
                 │ Daemon      │
                 │ (prismd)    │
                 └─────────────┘
```

**See** [GUI Architecture](architecture/GUI_ARCHITECTURE.md) and [Daemon API Reference](architecture/DAEMON_API_REFERENCE.md) for details.

---

## 🧪 Development Workflow

### Building

```bash
# Build all components
make build

# Build specific components
go build -o bin/prism ./cmd/prism/      # CLI
go build -o bin/prismd ./cmd/prismd/    # Daemon
go build -o bin/prism-gui ./cmd/prism-gui/ # GUI

# Run tests
make test
```

### Running

```bash
# CLI interface - daemon auto-starts as needed
./bin/prism launch python-ml my-project

# TUI interface - daemon auto-starts as needed
./bin/prism tui

# GUI interface - daemon auto-starts as needed
./bin/prism-gui
```

**See** [Development Setup](development/DEVELOPMENT_SETUP.md) for detailed setup instructions.

---

## 🧭 Key Implementation Guidelines

### 1. Validate Against Personas
Before implementing features, check if it improves one of the [5 persona workflows](USER_SCENARIOS/).

### 2. Follow Design Principles
See [DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md) - especially "Default to Success" and "Progressive Disclosure".

### 3. Maintain Multi-Modal Parity
Features must work across CLI, TUI, and GUI. See [Feature Parity Matrix](ROADMAP.md).

### 4. Focus on Usability First
Current priority is [Phase 5.0 UX Redesign](ROADMAP.md#-current-focus-phase-50---ux-redesign). Usability improvements take precedence over new features.

### 5. Use Existing Documentation
- Architecture questions: [architecture/](architecture/)
- User workflows: [USER_SCENARIOS/](USER_SCENARIOS/)
- Admin features: [admin-guides/](admin-guides/)
- Development: [development/](development/)

---

## 🎯 Quick Reference: Common Tasks

### Adding a New Feature
1. ✅ Does it improve a [persona workflow](USER_SCENARIOS/)?
2. ✅ Does it follow [design principles](DESIGN_PRINCIPLES.md)?
3. ✅ Check [UX evaluation](architecture/UX_EVALUATION_AND_RECOMMENDATIONS.md)
4. ✅ Implement in daemon (pkg/), then expose via API
5. ✅ Add to CLI (internal/cli/), TUI (internal/tui/), GUI (cmd/prism-gui/)
6. ✅ Update [ROADMAP.md](ROADMAP.md) status
7. ✅ Document in [user-guides/](user-guides/) or [admin-guides/](admin-guides/)

### Fixing UX Issues
1. ✅ Check [UX evaluation](architecture/UX_EVALUATION_AND_RECOMMENDATIONS.md) for prioritized fixes
2. ✅ Verify fix improves [persona workflows](USER_SCENARIOS/)
3. ✅ Update [ROADMAP.md](ROADMAP.md) Phase 5.0 checkboxes
4. ✅ Test against success metrics

### Creating a Release
**See [GoReleaser Release Process](development/GORELEASER_RELEASE_PROCESS.md) for complete documentation.**

Quick steps:
1. ✅ Sync version (pkg/version/version.go, cmd/prism-gui/frontend/package.json)
2. ✅ Commit and push all changes
3. ✅ Create annotated git tag: `git tag -a v0.5.9 -m "Release v0.5.9: Navigation Restructure"`
4. ✅ Set GitHub token: `export GITHUB_TOKEN=$(gh auth token)`
5. ✅ Run GoReleaser: `goreleaser release --clean`
6. ✅ Verify release at https://github.com/scttfrdmn/prism/releases

---

## 📊 Success Metrics

**See [ROADMAP.md - Success Metrics](ROADMAP.md#-success-metrics) for current vs target state.**

Key metrics we're tracking:
- ⏱️ Time to first workspace launch
- 🧭 Navigation complexity (number of items)
- 🎯 CLI first-attempt success rate
- 😃 User confusion rate (% of support tickets)
- 🔧 Advanced feature discoverability

---

**For detailed roadmap and current priorities, see [ROADMAP.md](ROADMAP.md)**
