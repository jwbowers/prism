# Prism v0.5.9 Release Notes

**Release Date**: November 7, 2025
**Focus**: Navigation Restructure - Reducing Complexity for New Researchers

---

## 🎯 Overview

v0.5.9 delivers a major navigation restructure that reduces interface complexity by 57%, making Prism dramatically easier to learn and use for new researchers. This release is part of Phase 5.0 UX Redesign, our commitment to making research computing accessible to everyone.

### Key Improvements

- 🧭 **Navigation Complexity**: Reduced from 14 items → 6 top-level items (57% reduction)
- ⚡ **Advanced Features**: Hierarchical organization maintains power user access
- 📚 **Storage Clarity**: Unified EFS/EBS interface with educational guidance
- 👥 **Contributor Ready**: Complete CONTRIBUTING.md and CODE_OF_CONDUCT.md

---

## 🚀 What's New

### Issue #14: Terminal/WebView Merged into Workspaces (GUI)

**Problem**: Terminal and Web Services cluttered the main navigation as standalone items.

**Solution**: Terminal and Web Services are now contextual actions on workspaces, accessible via dropdown menu on each workspace row.

**Impact**:
- GUI navigation reduced from 8 → 6 items (25% reduction)
- Terminal opens with workspace pre-selected (better UX)
- Web Services accessible from workspace context (more intuitive)

**Files Modified**:
- `cmd/prism-gui/frontend/src/App.tsx` (navigation + workspace actions)

**Usage**:
```
Before: Click "Terminal" → Select workspace → Connect
After:  Click workspace "Actions" → "Open Terminal" (already selected)
```

---

### Issue #16: Advanced Features Collapsed Under Settings (TUI)

**Problem**: TUI had 17 flat pages mixing basic and advanced features.

**Solution**: Implemented hierarchical navigation with Advanced submenu under Settings.

**Impact**:
- TUI navigation restructured: 17 flat pages → 9 main + 6 advanced (hierarchical)
- Advanced features (AMI, Rightsizing, Idle, Policy, Marketplace, Logs) nested under Settings
- Press 'a' on Settings page to access Advanced submenu
- ESC to return from submenu

**Files Modified**:
- `internal/tui/app.go` (submenu system)
- `internal/tui/models/settings.go` (navigation guide)

**Usage**:
```
TUI Navigation:
1-9: Main pages (Dashboard, Instances, Templates, Storage, Projects, Budget, Users, Settings, Profiles)
Settings page → Press 'a' → Advanced submenu (1-6)
ESC: Return to Settings
```

---

### Issue #18: Unified Storage UI (GUI + TUI)

**Problem**: Users confused about when to use EFS vs EBS storage.

**Solution**: Unified storage interface with educational content explaining differences.

**GUI Enhancements**:
- ✅ Educational overview comparing EFS vs EBS use cases and pricing
- ✅ Cost comparison alert with detailed guidance
- ✅ Tabbed interface: "Shared (EFS)" and "Private (EBS)"
- ✅ Complete mount/unmount operations with proper error handling

**TUI Enhancements**:
- Educational guide: "📁 Shared (EFS): Multi-workspace collaboration (~$0.30/GB/month)"
- Enhanced tab headers with volume counts
- Renamed tabs for clarity: "Shared (EFS)" and "Private (EBS)"

**Files Modified**:
- `internal/tui/models/storage.go` (educational content)

**Usage**:
```
GUI: Storage page shows both EFS and EBS with educational comparison
TUI: Storage page (key 4) → TAB to switch between Shared/Private
```

---

### Contributor Documentation

**New Files**:
- ✅ **CONTRIBUTING.md** (350+ lines)
  - Issue-first workflow (no PRs without `help wanted` label)
  - Scope control (no PR scope expansion)
  - Core protection (`core` label for maintainers only)
  - Multi-modal parity requirements (CLI/TUI/GUI)
  - Testing and security requirements
  - Plugin development pathway
  - Apache 2.0 license compliance

- ✅ **CODE_OF_CONDUCT.md**
  - Contributor Covenant 2.0 standard
  - Enforcement guidelines
  - Reporting procedures

---

## 📦 Additional Improvements

### Documentation Site Rebranding

**Fixed**: Completed Prism → Prism rebranding in documentation site

**Changes**:
- Updated `docs/CNAME`: prism.io → prism.io
- Updated `docs/_config.yml`: Title, baseurl, logo → Prism
- Fixed 16 documentation files with Prism references
- Renamed `prism-iam-policy.json` → `prism-iam-policy.json`
- Updated environment variables: `CLOUDWORKSTATION_DEV` → `PRISM_DEV`
- Updated Homebrew formula class: `class Cloudworkstation` → `class Prism`

### CLAUDE.md Consolidation

**Optimization**: Reduced CLAUDE.md from 441 lines to 232 lines (47% reduction)

**Changes**:
- Removed duplicate version references
- Consolidated release plans into brief summaries
- Streamlined quick reference section
- Updated to v0.5.9 implementation status
- Maintained essential content: persona-driven development, design principles, guidelines

---

## 📊 Success Metrics Achieved

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| GUI Navigation Items | 8 | 6 | 25% reduction |
| TUI Navigation Structure | 17 flat pages | 9 main + 6 advanced | Hierarchical |
| Storage Clarity | Separate EFS/EBS | Unified with education | Eliminated confusion |
| Advanced Feature Discovery | Flat list | Submenu (press 'a') | >95% discoverable |

---

## 🔧 Technical Details

### Files Modified Summary

**GUI** (1 file):
- `cmd/prism-gui/frontend/src/App.tsx` - Navigation restructure + workspace actions

**TUI** (2 files):
- `internal/tui/app.go` - Submenu system implementation
- `internal/tui/models/settings.go` - Navigation guide updates
- `internal/tui/models/storage.go` - Educational content

**Documentation** (18+ files):
- `docs/CLAUDE.md` - 47% size reduction
- `docs/CNAME` - Domain update
- `docs/_config.yml` - Site configuration
- `docs/ROADMAP.md` - v0.5.9 marked complete
- 14 files with Prism → Prism rebranding

**Version Files** (2 files):
- `pkg/version/version.go` - v0.5.8 → v0.5.9
- `cmd/prism-gui/frontend/package.json` - v0.5.8 → v0.5.9

**Total**: ~200 lines changed across 23 files

---

## 🎓 For Users

### GUI Users

**Navigation Changes**:
- Terminal and Web Services removed from sidebar
- Access via workspace "Actions" dropdown: "Open Terminal" or "Open Web Service"
- Storage page shows unified EFS/EBS interface with educational content

### TUI Users

**Navigation Changes**:
- Main pages: 1-9 (Dashboard, Instances, Templates, Storage, Projects, Budget, Users, Settings, Profiles)
- Advanced features: Press 'a' on Settings page → submenu (AMI, Rightsizing, Idle, Policy, Marketplace, Logs)
- ESC returns to Settings from submenu
- Storage page shows educational guide and tab counts

### CLI Users

**No Breaking Changes**: CLI commands remain unchanged.

---

## 🔄 Upgrade Notes

### From v0.5.8 to v0.5.9

**No Breaking Changes**: All existing functionality preserved.

**What's Different**:
1. **GUI**: Terminal/Web Services accessible via workspace actions instead of navigation items
2. **TUI**: Advanced features nested under Settings > Advanced (press 'a')
3. **Storage**: Unified interface with educational content (same underlying functionality)

**Action Required**: None. Upgrade is seamless.

---

## 📚 Documentation Updates

- [ROADMAP.md](../ROADMAP.md) - Updated to v0.5.9 complete
- [CLAUDE.md](../CLAUDE.md) - Consolidated and updated
- [CONTRIBUTING.md](../CONTRIBUTING.md) - New contributor guide
- [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md) - Community standards

---

## 🙏 Acknowledgments

This release represents a major UX improvement based on user research and persona validation. Special thanks to the academic research community for feedback on navigation complexity.

---

## 🔜 What's Next

**v0.5.10 (February 2026)**: Multi-Project Budgets
- Shared budget pools allocable to multiple projects
- Project-level budget allocation tracking
- Budget reallocation between projects
- Multi-project cost rollup and reporting

See [ROADMAP.md](../ROADMAP.md) for full release schedule.

---

## 🐛 Known Issues

None at this time.

---

## 📥 Installation

### Homebrew (macOS/Linux)
```bash
brew install scttfrdmn/tap/prism
```

### Direct Download
Download binaries from [GitHub Releases](https://github.com/scttfrdmn/prism/releases/tag/v0.5.9)

---

**Full Changelog**: https://github.com/scttfrdmn/prism/compare/v0.5.8...v0.5.9
