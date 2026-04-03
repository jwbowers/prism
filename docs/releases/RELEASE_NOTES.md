# Release Notes

For the complete changelog with all details, see [CHANGELOG.md on GitHub](https://github.com/scttfrdmn/prism/blob/main/CHANGELOG.md).

---

## Version history

| Version | Date | Highlights |
|---------|------|------------|
| [v0.30.0](#v0300) | 2026-04-02 | Removed TUI; CLI + GUI cover all use cases |
| [v0.29.2](#v0292) | 2026-04-02 | WAI-ARIA accessibility fixes across GUI and landing page |
| [v0.29.1](#v0291) | 2026-04-02 | Project deletion now blocks when active instances exist |
| [v0.29.0](#v0290) | 2026-04-02 | File operations (`workspace files push/pull/list`), Substrate integration tests |
| [v0.28.1](#v0281) | 2026-04-02 | AMI, Marketplace, Idle Detection E2E tests |
| [v0.28.0](#v0280) | 2026-04-01 | Cloudscape 3.0.1255 upgrade, Windows GUID, semver constraints |

---

## v0.30.0

**Released**: 2026-04-02

### Removed
- **TUI removed** — the `prism tui` command and `internal/tui/` package (~22,000 lines) have been retired. The TUI had fallen behind the CLI and GUI in feature coverage (missing file operations, governance, courses, and all features added after v0.26.0), was excluded from CI, and added maintenance overhead. All use cases are covered by the CLI and GUI. Removes `charmbracelet/bubbletea`, `bubbles`, and `lipgloss` dependencies.

---

## v0.29.2

**Released**: 2026-04-02

### Fixed
- WAI-ARIA accessibility across GUI and landing page (#568–577)
- `Terminal.tsx`: `role="status"`, `aria-live`, `aria-hidden` on status bar and pulsing dot; `role="application"` + `aria-label` on terminal container
- `SSHKeyModal.tsx`: `aria-label` and `aria-readonly` on key textareas
- `WebView.tsx`: `role="status"` + `aria-label` on loading overlay
- `App.tsx`: skip-navigation link; `tabIndex={-1}` on `#main-content`
- Landing page: `<main>` landmark, `<section aria-labelledby>`, `<article>` cards, corrected heading hierarchy, `aria-hidden` on decorative emoji
- Landing page CSS: `:focus-visible` outlines on all hero buttons

---

## v0.29.1

**Released**: 2026-04-02

### Fixed
- Project deletion now correctly blocks when running instances belong to the project (#539). Previously `getActiveInstancesForProject` was a permanent stub returning empty — deletion safety was silently bypassed.

---

## v0.29.0

**Released**: 2026-04-02

### Added
- **File operations**: `prism workspace files push/pull/list` — transfers files to/from running instances via S3 relay (#30a)
- Substrate v0.48.0 — all 6 Substrate integration tests pass (replaces LocalStack, no Docker required)

### Fixed
- All 268 pre-existing TypeScript typecheck errors resolved (`npm run typecheck` now clean)

---

## v0.28.1

**Released**: 2026-04-02

### Added
- E2E test coverage for AMI Management, Template Marketplace, and Idle Detection
- `AMIPage`, `MarketplacePage`, `IdlePage` Playwright page objects

### Fixed
- `WorkshopsPanel.tsx`: added missing `data-testid` that was causing E2E timeout cascade
- `course_handlers.go`: EFS creation now returns deterministic fake ID in test mode (was hanging indefinitely)
- `playwright.config.js`: added `NODE_OPTIONS="--max-old-space-size=4096"` to prevent Vite crash on large App.tsx

---

## v0.28.0

**Released**: 2026-04-01

### Added / Changed
- Cloudscape Design System upgrade to 3.0.1255
- Windows GUID support
- Semver constraint enforcement
- Policy framework fixes
- Research user migration improvements
